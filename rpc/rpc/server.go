package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"rpc/codec"
	"strings"
	"sync"
	"time"
)

const (
	Connected        = "200 Connected to Go RPC"
	DefaultRPCPath   = "/_rpc_"
	DefaultDebugPath = "/debug/rpc"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int
	CodecType      codec.Type
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 15,
}

type Server struct {
	serviceMap sync.Map
}

var DefaultServer = NewServer()

func NewServer() *Server {
	return &Server{}
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

func Register(rcvr any) error {
	return DefaultServer.Register(rcvr)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking:", req.RemoteAddr, ": ", err.Error())
	}
	io.WriteString(conn, fmt.Sprintf("HTTP/1.0 %s\n\n", Connected))
	s.ServeConn(conn)
}

func (s *Server) HandleHTTP() {
	http.Handle(DefaultRPCPath, s)
	http.Handle(DefaultDebugPath, debugHTTP{s})
	log.Println("rpc server debug path:", DefaultDebugPath)
}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) Register(rcvr any) error {
	service := newService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(service.name, service); dup {
		return fmt.Errorf("rpc server: service already defined: %s", service.name)
	}
	return nil
}

func (s *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = fmt.Errorf("rpc server: service/method request ill-formed: %s", serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	serviceInterface, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = fmt.Errorf("rpc server: can't find service %s", serviceName)
		return
	}
	svc = serviceInterface.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = fmt.Errorf("rpc server: can't find method %s", methodName)
	}
	return
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.CodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn), opt.HandleTimeout)
}

var invalidRequest = struct{}{}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	mtype        *methodType
	svc          *service
}

func (s *Server) serveCodec(cc codec.Codec, timeout time.Duration) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // io error or EOF, close the connection
			}
			req.h.Error = err.Error()
			s.writeResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg, timeout)
	}
	wg.Wait()
	cc.Close()
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	called := make(chan struct{}, 1)
	writen := make(chan struct{}, 1)
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			s.writeResponse(cc, req.h, invalidRequest, sending)
			return
		}
		s.writeResponse(cc, req.h, req.replyv.Interface(), sending)
		writen <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-writen
		return
	}
	select {
	case <-called:
		<-writen
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request timeout: expect within %s", timeout)
		s.writeResponse(cc, req.h, invalidRequest, sending)
	}
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return nil, err
	}

	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReply()

	var argvi any
	if req.argv.Kind() == reflect.Ptr {
		argvi = req.argv.Interface()
	} else {
		argvi = req.argv.Addr().Interface()
	}
	if err := cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv error:", err)
	}
	return req, err
}

func (s *Server) writeResponse(cc codec.Codec, h *codec.Header, reply any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(h, reply); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

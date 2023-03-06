package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          any
	Reply         any
	Error         error
	Done          chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	codec codec.Codec
	opt   *Option

	sending sync.Mutex
	header  codec.Header

	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool //user has called Close
	shutdown bool //server has told us to stop
}

var _ io.Closer = (*Client)(nil)

var (
	ErrShutdown = errors.New("connection is shut down")
)

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.CodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: code error:", err)
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error:", err)
		_ = conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", DefaultRPCPath))
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == Connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = fmt.Errorf("rpc client: unexpected HTTP response: %s", resp.Status)
	}
	return nil, err
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		codec:   cc,
		seq:     1, //0 means invalid
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return ErrShutdown
	}
	c.closing = true
	c.mu.Unlock()
	return c.codec.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

func (c *Client) Go(serviceMethod string, args, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	if reflect.TypeOf(reply).Kind() != reflect.Ptr {
		log.Panic("rpc client: reply must be a pointer")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

func (c *Client) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	call := c.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		c.removeCall(call.Seq)
		return fmt.Errorf("rpc client: call timeout: service method %s", serviceMethod)
	case <-call.Done:
		return call.Error
	}
}

func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	if err := c.codec.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown || c.closing {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}
func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = c.codec.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.codec.ReadBody(nil)
		case h.Error != "":
			call.Error = errors.New(h.Error)
			err = c.codec.ReadBody(nil)
			call.done()
		default:
			err := c.codec.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body error: " + err.Error())
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

type newClientFunc func(conn net.Conn, opt *Option) (*Client, error)

func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp, unix or other transport protocol
		return Dial(protocol, addr, opts...)
	}
}

func Dial(network, address string, opts ...*Option) (*Client, error) {
	return dial(NewClient, network, address, opts...)
}

func DialHTTP(network, address string, opts ...*Option) (*Client, error) {
	return dial(NewHTTPClient, network, address, opts...)
}

func dial(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()

	if opt.ConnectTimeout == 0 {
		return f(conn, opt)
	}

	type clientAndError struct {
		client *Client
		err    error
	}
	ch := make(chan clientAndError, 1)
	go func() {
		client, err := f(conn, opt)
		// time.Sleep(time.Hour) //模拟网络延迟
		ch <- clientAndError{client, err}
	}()
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case res := <-ch:
		return res.client, res.err
	}
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) > 1 {
		return nil, errors.New("too many options")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

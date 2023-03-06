package xclient

import (
	"log"
	"net/http"
	"rpc"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout = time.Minute * 5
	DefaultPath    = rpc.DefaultRPCPath + "/registry"
)

type Registry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr    string
	regtime time.Time
}

var DefaultRegistry = NewRegistry(DefaultTimeout)

func HandleHTTP() {
	DefaultRegistry.HandleHTTP()
}

func NewRegistry(timeout time.Duration) *Registry {
	return &Registry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Rpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Rpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.registServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Registry) HandleHTTP() {
	http.Handle(DefaultPath, r)
}

func (r *Registry) registServer(addr string) {
	r.mu.Lock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{
			Addr:    addr,
			regtime: time.Now(),
		}
	} else {
		s.regtime = time.Now()
	}
	r.mu.Unlock()
}

func (r *Registry) aliveServers() (servers []string) {
	r.mu.Lock()
	for addr, s := range r.servers {
		if time.Since(s.regtime) < r.timeout || r.timeout <= 0 {
			servers = append(servers, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	r.mu.Unlock()
	return
}

func Heartbeat(registryAddr, serverAddr string, duration time.Duration) {
	if duration == 0 {
		duration = DefaultTimeout * 9 / 10
	}
	err := sendHeartbeat(registryAddr, serverAddr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registryAddr, serverAddr)
		}
	}()
}

func sendHeartbeat(registryAddr, serverAddr string) error {
	log.Printf("rpc registry: heartbeating to %s\n", registryAddr)
	client := &http.Client{}
	req, err := http.NewRequest("POST", registryAddr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Rpc-Server", serverAddr)
	_, err = client.Do(req)
	return err
}

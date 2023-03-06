package xclient

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

var (
	_              Discovery = (*MultiServerDiscovery)(nil)
	ErrNoAvilable            = errors.New("no available server")
	ErrModeInvalid           = errors.New("invalid select mode")
)

type MultiServerDiscovery struct {
	rand    *rand.Rand   //generate random number
	mu      sync.RWMutex //protect following
	servers []string
	index   int // for round robin, index of server
}

func NewMultiServerDiscovery(servers []string) (d *MultiServerDiscovery) {
	d = &MultiServerDiscovery{
		rand:    rand.New(rand.NewSource(rand.Int63())),
		servers: servers,
	}
	if len(servers) > 0 {
		d.index = d.rand.Intn(len(d.servers))
	}
	return
}

func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

func (d *MultiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	d.servers = servers
	d.index = d.index % len(d.servers)
	d.mu.Unlock()
	return nil
}

func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	n := len(d.servers)
	if n == 0 {
		return "", ErrNoAvilable
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.rand.Intn(n)], nil
	case RoundRobinSelect:
		server := d.servers[d.index]
		d.index = (d.index + 1) % n
		return server, nil
	default:
		return "", ErrModeInvalid
	}
}

func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.RLock()
	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	d.mu.RUnlock()
	return servers, nil
}

type RegistryDiscovery struct {
	*MultiServerDiscovery
	registryAddr string
	timeout      time.Duration
	lastUpdate   time.Time
}

const DefaultUpdateTimeout = time.Minute

func NewRegistryDiscovery(registryAddr string, timeout time.Duration) *RegistryDiscovery {
	if timeout <= 0 {
		log.Fatal("rpc discovery: invalid update timeout", timeout)
	}
	return &RegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(nil),
		registryAddr:         registryAddr,
		timeout:              timeout,
	}
}

func (rd *RegistryDiscovery) Update(servers []string) error {
	rd.lastUpdate = time.Now()
	return rd.MultiServerDiscovery.Update(servers)
}

func (rd *RegistryDiscovery) Refresh() error {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	if time.Since(rd.lastUpdate) < rd.timeout {
		return nil
	}

	log.Println("rpc discovery: update servers from registry", rd.registryAddr)
	resp, err := http.Get(rd.registryAddr)
	if err != nil {
		log.Println("rpc discovery: registry err:", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-Rpc-Servers"), ",")
	rd.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if len(server) > 0 {
			rd.servers = append(rd.servers, server)
		}
	}
	rd.MultiServerDiscovery.index = rd.MultiServerDiscovery.index % len(rd.servers)
	rd.lastUpdate = time.Now()
	return nil
}

func (rd *RegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := rd.Refresh(); err != nil {
		return "", err
	}
	return rd.MultiServerDiscovery.Get(mode)
}

func (rd *RegistryDiscovery) GetAll() ([]string, error) {
	if err := rd.Refresh(); err != nil {
		return nil, err
	}
	return rd.MultiServerDiscovery.GetAll()
}

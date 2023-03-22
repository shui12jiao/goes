package xclient

import (
	"context"
	"log"
	"net"
	"net/http"
	"rpc"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type Num int
type Args struct{ A, B int }
type Reply string

func (n *Num) Add(args *Args, reply *Reply) error {
	*reply = Reply(strconv.Itoa(args.A + args.B))
	return nil
}

func startRegistry(t *testing.T) string {
	t.Helper()
	go http.ListenAndServe(":1234", DefaultRegistry)
	return "http://localhost:1234/_rpc_/registry"
}

func startRPCServer(t *testing.T, registryAddr string) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go rpc.Accept(l)

	rpc.Register(new(Num))
	Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)

	return l.Addr().String()
}

func TestCall(t *testing.T) {
	registryAddr := startRegistry(t)
	startRPCServer(t, registryAddr)
	startRPCServer(t, registryAddr)
	startRPCServer(t, registryAddr)
	d := NewRegistryDiscovery(registryAddr, DefaultUpdateTimeout)

	xc := NewXClient(d, RoundRobinSelect, nil)
	require.NotNil(t, xc)
	defer func() { xc.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{A: i, B: i * i}
			var reply Reply
			err := xc.Call(context.Background(), "Num.Add", args, &reply)
			require.NoError(t, err)
			require.Equal(t, Reply(strconv.Itoa(args.A+args.B)), reply)
			log.Printf("call[%d] %d + %d = %s", i, args.A, args.B, reply)
		}(i)
	}
	wg.Wait()
}

func TestBroadcast(t *testing.T) {
	registryAddr := startRegistry(t)
	startRPCServer(t, registryAddr)
	startRPCServer(t, registryAddr)
	startRPCServer(t, registryAddr)
	d := NewRegistryDiscovery(registryAddr, DefaultUpdateTimeout)

	xc := NewXClient(d, RoundRobinSelect, nil)
	require.NotNil(t, xc)
	defer func() { xc.Close() }()

	var reply Reply
	err := xc.Broadcast(context.Background(), "Num.Add", &Args{A: 10, B: 20}, &reply)
	require.NoError(t, err)
	require.Equal(t, Reply(strconv.Itoa(10+20)), reply)
}

package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
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

func startRPCServer(t *testing.T) string {
	t.Helper()
	Register(new(Num))
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go Accept(l)
	return l.Addr().String()
}

func startHTTPServer(t *testing.T) string {
	t.Helper()
	Register(new(Num))
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	HandleHTTP()
	go http.Serve(l, nil)
	return l.Addr().String()
}

func clientCall(t *testing.T, name string, client *Client, wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{A: i, B: i * i}
			var reply Reply
			err := client.Call(context.Background(), "Num.Add", args, &reply)
			require.NoError(t, err)
			require.Equal(t, Reply(strconv.Itoa(args.A+args.B)), reply)
			log.Printf("%s client call[%d] %d + %d = %s", name, i, args.A, args.B, reply)
		}(i)
	}
}

func TestServer(t *testing.T) {
	rpcAddr := startRPCServer(t)
	httpAddr := startHTTPServer(t)

	// test rpc
	log.Println("rpcAddr:", rpcAddr)
	client, err := Dial("tcp", rpcAddr)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer func() { _ = client.Close() }()

	var wg sync.WaitGroup
	clientCall(t, "RPC", client, &wg, 55)
	// wg.Wait()

	// test http
	log.Println("httpAddr:", httpAddr)
	clientHTTP, err := DialHTTP("tcp", httpAddr)
	require.NoError(t, err)
	require.NotNil(t, clientHTTP)

	defer func() { _ = clientHTTP.Close() }()
	clientCall(t, "HTTP", clientHTTP, &wg, 5)
	wg.Wait()
}

func TestRPCClients(t *testing.T) {
	addr := startRPCServer(t)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			client, err := Dial("tcp", addr)
			require.NoError(t, err)
			require.NotNil(t, client)
			var wgc sync.WaitGroup
			clientCall(t, fmt.Sprintf("RPC<%d>", i), client, &wgc, 2)
			wgc.Wait()
			_ = client.Close()
		}(i)
	}
	wg.Wait()
}

func TestHTTPClients(t *testing.T) {
	addr := startHTTPServer(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			client, err := DialHTTP("tcp", addr)
			require.NoError(t, err)
			require.NotNil(t, client)
			var wgc sync.WaitGroup
			clientCall(t, fmt.Sprintf("RPC<%d>", i), client, &wgc, 5)
			wgc.Wait()
			_ = client.Close()
		}(i)
	}
	wg.Wait()
}

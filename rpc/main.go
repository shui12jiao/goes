package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"rpc"
	"strconv"
	"sync"
	"time"
)

type Num int
type Args struct{ A, B int }
type Reply string

func (n *Num) Sum(args *Args, reply *Reply) error {
	*reply = Reply(strconv.Itoa(args.A + args.B))
	return nil
}

func startServer(httpUse bool) string {
	rpc.Register(new(Num))

	l, err := net.Listen("tcp", ":50001")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	if httpUse {
		rpc.HandleHTTP()
		go http.Serve(l, nil)
	} else {
		go rpc.Accept(l)
	}
	return l.Addr().String()
}

func call(addr string, http bool) {
	opt := rpc.DefaultOption
	// opt.ConnectTimeout = 5 * time.Second //

	var client *rpc.Client
	var err error
	if http {
		client, err = rpc.DialHTTP("tcp", addr, opt)
	} else {
		client, err = rpc.Dial("tcp", addr, opt)
	}
	if err != nil {
		log.Fatal("network error:", err)
	}
	defer func() { _ = client.Close() }()

	time.Sleep(time.Millisecond * 100)

	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{A: i, B: i * i}
			var reply Reply
			if err := client.Call(context.Background(), "Num.Sum", args, &reply); err != nil {
				log.Fatal("call Num.Sum error:", err)
			}
			log.Printf("%d + %d = %s", args.A, args.B, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	httpUse := false
	addr := startServer(httpUse)
	go call(addr, httpUse)
	go call(addr, httpUse)
	go call(addr, httpUse)
	time.Sleep(time.Minute)
}

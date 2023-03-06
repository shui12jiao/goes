package main

import (
	"context"
	"log"
	"net"
	"rpc"
	"strconv"
	"sync"
	"time"
)

func startServer(addr chan string) {
	rpc.Register(new(Num))

	l, err := net.Listen("tcp", ":50001")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	rpc.Accept(l)
	// rpc.HandleHTTP()
	// http.Serve(l, nil)
}

type Num int
type Args struct{ A, B int }
type Reply string

func (n *Num) Sum(args *Args, reply *Reply) error {
	*reply = Reply(strconv.Itoa(args.A + args.B))
	return nil
}

func call(addr chan string) {
	opt := rpc.DefaultOption
	opt.ConnectTimeout = 5 * time.Second

	// client, err := rpc.DialHTTP("tcp", <-addr)
	client, err := rpc.Dial("tcp", <-addr, opt)
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
	ch := make(chan string)
	go call(ch)
	startServer(ch)
}

package xclient

import (
	"context"
	"io"
	"reflect"
	"rpc"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *rpc.Option
	mu      sync.Mutex
	clients map[string]*rpc.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *rpc.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*rpc.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply any) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var e error
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply any
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(ctx, rpcAddr, serviceMethod, args, clonedReply)

			mutex.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // cancel unfinished calls if any error occurs
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mutex.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	cancel()
	return e
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(ctx, rpcAddr, serviceMethod, args, reply)
}

func (xc *XClient) dial(rpcAddr string) (client *rpc.Client, err error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	client, ok := xc.clients[rpcAddr]
	if ok {
		if client.IsAvailable() {
			return client, nil
		} else {
			client.Close()
			delete(xc.clients, rpcAddr)
			client = nil
		}
	}
	client, err = rpc.XDial(rpcAddr, xc.opt)
	if err != nil {
		return nil, err
	}
	xc.clients[rpcAddr] = client
	return client, nil
}

func (xc *XClient) call(ctx context.Context, rpcAddr string, serviceMethod string, args, reply any) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

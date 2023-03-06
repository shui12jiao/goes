package main

import (
	"dc"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Alice": "123",
	"Bob":   "456",
	"Cindy": "789",
	"David": "000",
	"Eric":  "111",
	"Frank": "222",
}

func createGroup() *dc.Group {
	return dc.NewGroup("test", 2<<10, dc.GetterFunc(func(key string) ([]byte, error) {
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, group *dc.Group) {
	pool := dc.NewHTTPPool(addr)
	pool.Set(addrs...)
	group.RegisterPeerPicker(pool)
	fmt.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

func startAPIServer(apiAddr string, group *dc.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	fmt.Println("api server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var api bool
	var port int
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.Parse()

	apiAddr := "http://localhost:50500"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}

	startCacheServer(addrMap[port], []string(addrs), g)
}

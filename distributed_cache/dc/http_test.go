package dc

import (
	"dc/pb"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"google.golang.org/protobuf/proto"
)

func BenchmarkHTTP(b *testing.B) {
	go http.ListenAndServe(":50001", http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := ByteView{b: []byte("hello world")}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(val.ByteSlice())
	})))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// GetHTTP("test", "test")
		val, err := GetHTTP("test", "test")
		if err != nil {
			b.Fatal(err)
		}
		if string(val) != "hello world" {
			b.Fatal("not equal")
		}
	}
}

func GetHTTP(group, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", "http://localhost:50001/", url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	bytes, _ := io.ReadAll(res.Body)
	return bytes, nil
}

func BenchmarkProtobuf(b *testing.B) {
	go http.ListenAndServe(":50002", http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view := ByteView{b: []byte("hello world")}
		val, _ := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(val))
	})))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res := &pb.Response{}
		err := GetProtobuf(&pb.Request{Group: "test", Key: "test"}, res)
		if err != nil {
			b.Fatal(err)
		}
		if string(res.GetValue()) != "hello world" {
			b.Fatal("not equal")
		}
	}
}

func GetProtobuf(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		"http://localhost:50002/",
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	bytes, _ := io.ReadAll(res.Body)
	return proto.Unmarshal(bytes, out)
}

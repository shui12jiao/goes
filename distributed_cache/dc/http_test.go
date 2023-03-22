package dc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func BenchmarkHTTP(b *testing.B) {
	go http.ListenAndServe(":50001", http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := ByteView{b: []byte("hello world")}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(val.ByteSlice())
	})))
	time.Sleep(time.Millisecond * 100)
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

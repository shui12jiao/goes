package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"wf"
)

func Logger() wf.HandlerFunc {
	return func(c *wf.Context) {
		// Start timer
		t := time.Now()
		log.Printf("time: %v", t)
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Request.RequestURI, time.Since(t))
	}
}

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := wf.New().Use(wf.Recovery())
	r.GET("/", func(c *wf.Context) {
		c.String(http.StatusOK, "<h1>Hello World</h1>", nil)
	})
	r.GET("/panic", func(ctx *wf.Context) {
		names := []string{"wf"}
		ctx.String(http.StatusOK, names[100])
	})

	r.Run(":50500")
}

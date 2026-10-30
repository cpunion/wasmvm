//go:build ignore

// This external server is shared by own and reference HTTP benchmark processes.
// Run: go run benchmarks/components/http_server.go AUTHORITY_FILE
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: http_server AUTHORITY_FILE")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[1], []byte(listener.Addr().String()), 0600); err != nil {
		panic(err)
	}
	var connections, requests, invalid atomic.Uint64
	payload := []byte(strings.Repeat("0123456789abcdef", 256))
	server := &http.Server{
		ConnState: func(_ net.Conn, state http.ConnState) {
			if state == http.StateNew {
				connections.Add(1)
			}
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/stats" {
				_ = json.NewEncoder(w).Encode(map[string]uint64{"connections": connections.Load(), "requests": requests.Load(), "invalid": invalid.Load()})
				return
			}
			requests.Add(1)
			data, err := io.ReadAll(io.LimitReader(r.Body, int64(len(payload)+1)))
			if err != nil || r.Method != "POST" || r.URL.Path != "/echo" || !bytes.Equal(data, payload) || r.Trailer.Get("X-Request-Trailer") != "done" {
				invalid.Add(1)
				http.Error(w, "invalid request, body or trailer", 400)
				return
			}
			w.Header().Set("Trailer", "X-Reply-Trailer")
			_, _ = w.Write(data)
			w.Header().Set("X-Reply-Trailer", "done")
		}),
	}
	panic(server.Serve(listener))
}

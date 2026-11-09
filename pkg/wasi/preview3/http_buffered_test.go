package preview3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPBufferedCompleteGuestBodiesAndIndependentSnapshots(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, fixture := range []string{"world-rust-http", "world-rust-composed-http"} {
			w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 4096, MaxWriteBytes: 4096}})
			in := instantiate(t, compileFixture(t, backend, fixture), w.Imports())
			var retained []*BufferedResponse
			for _, payload := range [][]byte{nil, []byte("\x00\xff\x80body"), bytes.Repeat([]byte("\x00\xff\x80bytes"), 8192)} {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				response, err := w.HTTP().CallIncomingBuffered(ctx, in, httptest.NewRequest("POST", "http://localhost/", bytes.NewReader(payload)), uint64(max(1, len(payload))), component.CallOptions{})
				cancel()
				if err != nil || response == nil {
					t.Fatal(fixture, err)
				}
				if response.StatusCode != 201 || response.Header.Get("X-Guest") != "rust-p3" || response.Trailer.Get("X-Guest-Trailer") != "done" || !bytes.Equal(response.Body, payload) {
					t.Fatal(fixture, response.StatusCode, len(response.Body), len(payload), response.Header, response.Trailer)
				}
				retained = append(retained, response)
			}
			if string(retained[1].Body) != "\x00\xff\x80body" || len(retained[0].Body) != 0 {
				t.Fatal("later guest execution changed a retained response")
			}
			retained[2].Body[0] = 42
			if retained[1].Body[0] != 0 {
				t.Fatal("response bodies alias")
			}
		}
	})
}

func TestHTTPBufferedExternalResponseAndLengthFailures(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, mode := range []string{"complete", "empty", "limit", "declared-limit", "length", "panic"} {
			p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, _ *http.Request) (*http.Response, error) {
				if mode == "panic" {
					panic("handler")
				}
				body, length := "external", int64(-1)
				header := http.Header{"X-External": []string{"yes"}}
				if mode == "empty" {
					body, length = "", 0
				}
				if mode == "length" {
					length = 3
					header.Set("Content-Length", "3")
				}
				if mode == "declared-limit" {
					length = 100
					header.Set("Content-Length", "100")
				}
				return &http.Response{StatusCode: 203, Header: header, Trailer: http.Header{"X-End": []string{"tail"}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: length}, nil
			}})
			in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
			limit := uint64(64)
			if mode == "limit" || mode == "declared-limit" {
				limit = 4
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			response, err := p.CallIncomingBuffered(ctx, in, httptest.NewRequest("GET", "http://localhost/", nil), limit, component.CallOptions{})
			cancel()
			if mode == "complete" || mode == "empty" {
				want := "external"
				if mode == "empty" {
					want = ""
				}
				if err != nil || response == nil || string(response.Body) != want || response.StatusCode != 203 || response.Header.Get("X-External") != "yes" || response.Trailer.Get("X-End") != "tail" {
					t.Fatal(mode, response, err)
				}
			} else {
				if err == nil || response != nil {
					t.Fatal(mode, response, err)
				}
				if (mode == "limit" || mode == "declared-limit") && !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(mode, err)
				}
			}
		}
	})
}

type bufferedBlockedReader struct {
	*httpBlockedReader
	started chan struct{}
	once    sync.Once
}

func (r *bufferedBlockedReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.started) })
	return r.httpBlockedReader.Read(p)
}

func TestHTTPBufferedCancellationJoinsAndValidatesArguments(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		reader := &bufferedBlockedReader{httpBlockedReader: newHTTPBlockedReader(), started: make(chan struct{})}
		p := httpOwner(t, HTTPConfig{Handler: func(context.Context, *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: reader, ContentLength: -1}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
		request := httptest.NewRequest("GET", "http://localhost/", nil)
		for _, limit := range []uint64{0, ^uint64(0)} {
			if _, err := p.CallIncomingBuffered(nil, in, request, limit, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
		}
		if _, err := p.CallIncomingBuffered(nil, nil, request, 64, component.CallOptions{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			select {
			case <-reader.started:
				cancel()
			case <-ctx.Done():
			}
		}()
		response, err := p.CallIncomingBuffered(ctx, in, request, 64, component.CallOptions{})
		if response != nil || !errors.Is(err, context.Canceled) {
			t.Fatal(response, err)
		}
		select {
		case <-reader.done:
		case <-time.After(time.Second):
			t.Fatal("cancellation did not close the external reader")
		}
		p.mu.Lock()
		requests := p.requests
		p.mu.Unlock()
		if requests != 0 {
			t.Fatal("completed buffered call retained admission", requests)
		}
	})
}

func TestHTTPBufferedMeteredGuestLimitAndSharedWorld(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxIdleWorkers: 1}})
		module := compileFixture(t, backend, "http-server")
		in := instantiate(t, module, w.Imports())
		response, err := w.HTTP().CallIncomingBuffered(nil, in, httptest.NewRequest("GET", "http://localhost/", nil), 5, component.CallOptions{Metered: true, Fuel: 10000})
		if err != nil || response == nil || string(response.Body) != "hello" || response.FuelRemaining == 0 || response.FuelRemaining >= 10000 {
			t.Fatal(response, err)
		}
		limited := instantiate(t, module, w.Imports())
		if response, err := w.HTTP().CallIncomingBuffered(nil, limited, httptest.NewRequest("GET", "http://localhost/", nil), 4, component.CallOptions{}); response != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(response, err)
		}
		echo := compileFixture(t, backend, "world-rust-http")
		instances := []*component.Instance{instantiate(t, echo, w.Imports()), instantiate(t, echo, w.Imports())}
		var jobs sync.WaitGroup
		for index, instance := range instances {
			jobs.Go(func() {
				for round := range 3 {
					payload := strings.Repeat(string([]byte{byte(128 + index*16 + round)}), 1024+index*31+round)
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					response, err := w.HTTP().CallIncomingBuffered(ctx, instance, httptest.NewRequest("POST", "http://localhost/", strings.NewReader(payload)), uint64(len(payload)), component.CallOptions{})
					cancel()
					if err != nil || response == nil || string(response.Body) != payload || response.Trailer.Get("X-Guest-Trailer") != "done" {
						t.Error(index, round, response, err)
						return
					}
				}
			})
		}
		jobs.Wait()
	})
}

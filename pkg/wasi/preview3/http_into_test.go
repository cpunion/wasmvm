package preview3

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPCallerOutputRetainsBoundsAndExplicitAliases(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 4096, MaxWriteBytes: 4096}})
		module := compileFixture(t, backend, "world-rust-http")
		in := instantiate(t, module, w.Imports())
		for _, size := range []int{0, 16, 4096, 65536} {
			whole := bytes.Repeat([]byte{0xa5}, max(1, size)+2)
			dst := whole[1 : len(whole)-1]
			var first *BufferedResponse
			for round := 0; round < 2; round++ {
				payload := bytes.Repeat([]byte{byte(0x70 + round)}, size)
				request := httptest.NewRequest("POST", "http://localhost/", bytes.NewReader(payload))
				response, err := w.HTTP().CallIncomingInto(nil, in, request, dst, component.CallOptions{})
				if err != nil || response == nil || !bytes.Equal(response.Body, payload) || response.StatusCode != 201 || response.Header.Get("X-Guest") != "rust-p3" || response.Trailer.Get("X-Guest-Trailer") != "done" {
					t.Fatal(backend, size, round, response, err)
				}
				if whole[0] != 0xa5 || whole[len(whole)-1] != 0xa5 {
					t.Fatal("write exceeded caller destination")
				}
				if size != 0 && (&response.Body[0] != &dst[0] || first != nil && first.Body[0] != payload[0]) {
					t.Fatal("response did not retain the explicit caller-owned view")
				}
				first = response
			}
		}
		limited := instantiate(t, module, w.Imports())
		guarded := []byte{11, 0, 0, 0, 0, 19}
		if response, err := w.HTTP().CallIncomingInto(nil, limited, httptest.NewRequest("POST", "http://localhost/", bytes.NewBufferString("too large")), guarded[1:5], component.CallOptions{}); response != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(response, err)
		}
		if guarded[0] != 11 || guarded[5] != 19 {
			t.Fatal("failure exceeded output limit")
		}
		if _, err := w.HTTP().CallIncomingInto(nil, in, httptest.NewRequest("GET", "http://localhost/", nil), nil, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("empty destination", err)
		}
		if _, err := w.HTTP().CallIncomingInto(nil, nil, httptest.NewRequest("GET", "http://localhost/", nil), make([]byte, 1), component.CallOptions{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal("nil instance", err)
		}
		cancelled, stop := context.WithCancel(context.Background())
		stop()
		before := append([]byte(nil), guarded...)
		if _, err := w.HTTP().CallIncomingInto(cancelled, in, httptest.NewRequest("GET", "http://localhost/", nil), guarded[1:5], component.CallOptions{}); !errors.Is(err, context.Canceled) || !bytes.Equal(guarded, before) {
			t.Fatal("canceled entry changed caller storage", err)
		}
	})
}

func TestHTTPCallerOutputSharesMeteredGuestJoin(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{})
		in := instantiate(t, compileFixture(t, backend, "http-server"), w.Imports())
		dst := make([]byte, 5)
		response, err := w.HTTP().CallIncomingInto(nil, in, httptest.NewRequest("GET", "http://localhost/", nil), dst, component.CallOptions{Metered: true, Fuel: 10000})
		if err != nil || response == nil || string(dst) != "hello" || response.FuelRemaining == 0 || response.FuelRemaining >= 10000 {
			t.Fatal(response, err)
		}
		w.HTTP().mu.Lock()
		requests := w.HTTP().requests
		w.HTTP().mu.Unlock()
		if requests != 0 {
			t.Fatal("caller-owned output retained admission", requests)
		}
	})
}

func TestHTTPCallerOutputCancellationClosesExternalReader(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		reader := &bufferedBlockedReader{httpBlockedReader: newHTTPBlockedReader(), started: make(chan struct{})}
		p := httpOwner(t, HTTPConfig{Handler: func(context.Context, *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 203, Header: http.Header{}, Trailer: http.Header{}, Body: reader, ContentLength: -1}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go func() {
			select {
			case <-reader.started:
				cancel()
			case <-ctx.Done():
			}
		}()
		guarded := bytes.Repeat([]byte{0xa5}, 66)
		response, err := p.CallIncomingInto(ctx, in, httptest.NewRequest("GET", "http://localhost/", nil), guarded[1:65], component.CallOptions{})
		if response != nil || !errors.Is(err, context.Canceled) || guarded[0] != 0xa5 || guarded[65] != 0xa5 {
			t.Fatal("cancellation result or output boundary", response, err)
		}
		select {
		case <-reader.done:
		case <-time.After(time.Second):
			t.Fatal("canceled caller-buffer read retained its external reader")
		}
		p.mu.Lock()
		requests := p.requests
		p.mu.Unlock()
		if requests != 0 {
			t.Fatal("cancellation retained request admission", requests)
		}
	})
}

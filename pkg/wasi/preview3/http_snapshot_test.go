package preview3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSnapshotRequestBodyAndReplay(t *testing.T) {
	payload := "\x00\xff\x80data"
	r, err := NewSnapshotRequest(context.Background(), "POST", "http://localhost/", payload)
	if err != nil || r.ContentLength != int64(len(payload)) {
		t.Fatal(r, err)
	}
	var empty [0]byte
	if n, err := r.Body.Read(empty[:]); n != 0 || err != nil {
		t.Fatal(n, err)
	}
	first := make([]byte, 2)
	if n, err := r.Body.Read(first); n != 2 || err != nil || string(first) != payload[:2] {
		t.Fatal(n, err, first)
	}
	replay, err := r.GetBody()
	if err != nil {
		t.Fatal(err)
	}
	defer replay.Close()
	if err := r.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Body.Read(first); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	data, err := io.ReadAll(replay)
	if err != nil || string(data) != payload {
		t.Fatal(string(data), err)
	}
	if err := r.Body.Close(); err != nil || r.Body.(*snapshotReader).data != "" {
		t.Fatal("closed reader retained input", err)
	}
	for _, tc := range []struct {
		ctx         context.Context
		method, url string
	}{{nil, "POST", "http://localhost"}, {context.Background(), "bad method", "http://localhost"}, {context.Background(), "GET", ":"}} {
		if _, err := NewSnapshotRequest(tc.ctx, tc.method, tc.url, payload); err == nil {
			t.Fatal("invalid request accepted", tc)
		}
	}
}

func TestHTTPSnapshotAccountingReadAheadAndContention(t *testing.T) {
	p := httpOwner(t, HTTPConfig{})
	p.network.io.config.MaxReadBytes = 4
	for _, expected := range []uint64{5, 3, 7} {
		b, err := p.newBody()
		if err != nil {
			t.Fatal(err)
		}
		r := &snapshotReader{data: "abcde"}
		b.reader, b.expected, b.sizeFailure = r, &expected, "HTTP-request-body-size"
		if _, _, ready := b.readSnapshot(io.NopCloser(strings.NewReader("external"))); ready {
			t.Fatal("arbitrary reader used snapshot path")
		}
		b.readerMu.Lock()
		_, _, ready := b.readSnapshot(r)
		b.readerMu.Unlock()
		if ready {
			t.Fatal("contended read used immediate path")
		}
		r.mu.Lock()
		_, _, ready = b.readSnapshot(r)
		r.mu.Unlock()
		if ready {
			t.Fatal("contended input used immediate path")
		}
		var nilReader *snapshotReader
		if _, _, ready := b.readSnapshot(nilReader); ready {
			t.Fatal("nil reader admitted")
		}
		data, failed, ready := b.readSnapshot(r)
		if !ready || data != "abcd" || (failed != nil) != (expected == 3) {
			t.Fatal(expected, data, failed, ready)
		}
		if expected != 3 {
			data, failed, ready = b.readSnapshot(r)
			if !ready || data != "e" || failed == nil {
				t.Fatal(data, failed, ready)
			}
			if expected == 5 && failed != io.EOF || expected == 7 && failed == io.EOF {
				t.Fatal(expected, failed)
			}
		}
		if !r.closed || !b.bodyDone {
			t.Fatal("terminal read did not close body")
		}
		b.release()
	}
	// A trailer publication panic must not escape the prompt callback.
	b, _ := p.newBody()
	b.reader = &snapshotReader{}
	b.trailerSnapshot = func() (*host.HTTPFields, error) { panic("trailer") }
	if data, err, ready := b.readSnapshot(b.reader); !ready || data != "" || err == nil || !b.bodyDone {
		t.Fatal("trailer panic escaped or retained body", data, err, ready)
	}
	b.release()
}

func TestSnapshotFinalDataClosesBodyAndPublishesTrailers(t *testing.T) {
	p := httpOwner(t, HTTPConfig{})
	p.network.io.config.MaxReadBytes = 64
	b, err := p.newBody()
	if err != nil {
		t.Fatal(err)
	}
	defer b.release()
	r := &snapshotReader{data: "final bytes"}
	b.reader = r
	public := &snapshotReader{data: r.data}
	buffer := make([]byte, 32)
	if n, failed := public.Read(buffer); n != len(r.data) || failed != nil {
		t.Fatal("ordinary Reader EOF contract changed", n, failed)
	}
	if _, failed := public.Read(buffer); failed != io.EOF {
		t.Fatal(failed)
	}
	public.Close()
	publications := 0
	fields, err := host.NewHTTPFields(p.fields, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	b.trailerSnapshot = func() (*host.HTTPFields, error) { publications++; return fields, nil }
	data, failed, ready := b.readSnapshot(r)
	if !ready || data != "final bytes" || failed != io.EOF || !r.closed || !b.bodyDone || publications != 1 || b.trailersValue != fields {
		t.Fatal("final immutable bytes did not finish body/trailers", data, failed, ready, publications)
	}
}

func TestHTTPSnapshotCancellationAndLimit(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{})
		module := compileFixture(t, backend, "world-rust-http")
		for _, canceled := range []bool{false, true} {
			in := instantiate(t, module, w.Imports())
			ctx, cancel := context.WithCancel(context.Background())
			r, _ := NewSnapshotRequest(ctx, "POST", "http://localhost/", "payload")
			if canceled {
				cancel()
			}
			response, err := w.HTTP().CallIncomingBuffered(ctx, in, r, 4, component.CallOptions{})
			cancel()
			want := wasm.ErrLimit
			if canceled {
				want = context.Canceled
			}
			if response != nil || !errors.Is(err, want) {
				t.Fatal(canceled, response, err)
			}
		}
	})
}

func TestHTTPSnapshotCompleteGuestAndRetainedResponse(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 4096, MaxWriteBytes: 4096}})
		in := instantiate(t, compileFixture(t, backend, "world-rust-http"), w.Imports())
		var retained []byte
		for _, payload := range []string{"", "\x00\xff\x80body", strings.Repeat("\x00\xff\x80bytes", 8192)} {
			r, err := NewSnapshotRequest(context.Background(), "POST", "http://localhost/", payload)
			if err != nil {
				t.Fatal(err)
			}
			r.Trailer = http.Header{"X-Incoming": []string{"done"}}
			response, err := w.HTTP().CallIncomingBuffered(context.Background(), in, r, uint64(max(1, len(payload))), component.CallOptions{})
			if err != nil || response == nil {
				t.Fatal(err)
			}
			if !bytes.Equal(response.Body, []byte(payload)) || response.StatusCode != 201 || response.Header.Get("X-Guest") != "rust-p3" || response.Trailer.Get("X-Guest-Trailer") != "done" {
				t.Fatal("incomplete response", response)
			}
			if !r.Body.(*snapshotReader).closed {
				t.Fatal("consumed request body remained open")
			}
			if payload == "\x00\xff\x80body" {
				retained = response.Body
			}
		}
		if string(retained) != "\x00\xff\x80body" {
			t.Fatal("later request changed retained bytes")
		}
		r, _ := NewSnapshotRequest(context.Background(), "POST", "http://localhost/", "streaming")
		response, err := w.HTTP().CallIncoming(context.Background(), in, r, component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(response.Body)
		if err != nil || string(data) != "streaming" || response.Trailer.Get("X-Guest-Trailer") != "done" {
			t.Fatal(string(data), err, response.Trailer)
		}
		if err := response.Body.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestHTTPSnapshotReadinessWithBusyIOWorkers(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxStreams: 1}})
		in := instantiate(t, compileFixture(t, backend, "world-rust-http"), w.Imports())
		release := make(chan struct{})
		defer close(release)
		for range 2 {
			if err := w.IO().worker(func() { <-release }); err != nil {
				t.Fatal(err)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		const payload = "ready without an I/O worker"
		request, err := NewSnapshotRequest(ctx, "POST", "http://localhost/", payload)
		if err != nil {
			t.Fatal(err)
		}
		response, err := w.HTTP().CallIncomingBuffered(ctx, in, request, uint64(len(payload)), component.CallOptions{})
		if err != nil || response == nil {
			t.Fatal(response, err)
		}
		if response.StatusCode != 201 || string(response.Body) != payload || response.Trailer.Get("X-Guest-Trailer") != "done" {
			t.Fatal("incomplete response", response)
		}
	})
}

func TestHTTPSnapshotInlineCloseBodyPoolAndShutdown(t *testing.T) {
	p := httpOwner(t, HTTPConfig{})
	b1, err := p.newBody()
	if err != nil {
		t.Fatal(err)
	}
	r := &snapshotReader{data: "inline-close"}
	b1.reader = r
	b1.release()
	if !r.closed {
		t.Fatal("snapshotReader was not closed inline on release")
	}
	p.mu.Lock()
	bodiesCount, freeCount := len(p.bodies), len(p.freeBodies)
	p.mu.Unlock()
	if bodiesCount != 0 || freeCount != 1 {
		t.Fatalf("unexpected body tracking after inline close: bodies=%d free=%d", bodiesCount, freeCount)
	}
	b2, err := p.newBody()
	if err != nil {
		t.Fatal(err)
	}
	if b2 != b1 {
		t.Fatal("expected recycled httpBody instance on subsequent newBody")
	}
	b2.reader = http.NoBody
	b2.unpooled = true
	b2.release()
	p.mu.Lock()
	bodiesCount, freeCount = len(p.bodies), len(p.freeBodies)
	p.mu.Unlock()
	if bodiesCount != 0 || freeCount != 0 {
		t.Fatalf("unpooled body leaked or recycled: bodies=%d free=%d", bodiesCount, freeCount)
	}
	b3, err := p.newBody()
	if err != nil {
		t.Fatal(err)
	}
	b3.release()
	p.Close()
	if err := p.WaitClosed(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	freeCount = len(p.freeBodies)
	p.mu.Unlock()
	if freeCount != 0 || p.checkedHandler.Load() != nil {
		t.Fatalf("shutdown retained pool/cache state: free=%d checkedHandler=%v", freeCount, p.checkedHandler.Load())
	}
}

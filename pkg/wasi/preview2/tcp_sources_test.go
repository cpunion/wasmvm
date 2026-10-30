//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func tcpBackendSource(t *testing.T, a tcpAPI, process *IO, value component.Value, input bool) *tcpSource {
	t.Helper()
	typ := process.outputType
	if input {
		typ = process.inputType
	}
	id, err := a.in.ResourceRep(nil, typ, value.Resource)
	if err != nil {
		t.Fatal(err)
	}
	process.mu.Lock()
	record := process.records[id]
	process.mu.Unlock()
	if input {
		return record.source.input.(*tcpSource)
	}
	return record.source.output.(*tcpSource)
}
func TestTCPBackendBackpressureReadinessAndClosedStreams(t *testing.T) {
	p, _, process := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	c := a.connect()
	reader, writer := tcpBackendSource(t, a, process, c.si, true), tcpBackendSource(t, a, process, c.co, false)
	if n, err := reader.Read(context.Background(), make([]byte, 8)); n != 0 || !errors.Is(err, ErrWouldBlock) {
		t.Fatal(n, err)
	}
	ready, changed := reader.Readiness()
	if ready || changed == nil {
		t.Fatal(ready)
	}
	if n, err := writer.Write(context.Background(), []byte("abc")); n != 3 || err != nil {
		t.Fatal(n, err)
	}
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("readiness lost notification")
	}
	data := make([]byte, 8)
	if n, err := reader.Read(context.Background(), data); n != 3 || err != nil || string(data[:n]) != "abc" {
		t.Fatal(n, err)
	}
	a.ok("set-send-buffer-size", c.client, component.Value{Bits: 4096})
	a.ok("set-receive-buffer-size", c.child, component.Value{Bits: 4096})
	payload := make([]byte, 65536)
	blocked := false
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for i := 0; i < 512; i++ {
		n, err := writer.Write(ctx, payload)
		if n < 0 || n > len(payload) {
			t.Fatal(n)
		}
		if errors.Is(err, ErrWouldBlock) {
			blocked = true
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if !blocked {
		t.Fatal("socket never applied backpressure")
	}
	ready, changed = writer.Readiness()
	// Acks may race the readiness snapshot. Drain input until pending capacity
	// is observable, without assuming a fixed kernel buffer size or chunk count.
	if !ready {
		for {
			select {
			case <-changed:
				ready = true
			default:
			}
			if ready {
				break
			}
			_, err := reader.Read(ctx, payload)
			if err != nil && !errors.Is(err, ErrWouldBlock) {
				t.Fatal(err)
			}
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
		}
	}
	if err := writer.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	cancelled, stop := context.WithCancel(context.Background())
	stop()
	if _, err := writer.Write(cancelled, []byte("x")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := writer.Flush(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	a.ok("shutdown", c.client, component.Value{Bits: 2})
	if _, err := writer.Write(context.Background(), []byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := writer.Flush(context.Background()); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if ready, _ := writer.Readiness(); !ready {
		t.Fatal("closed source not ready")
	}
	a.close(c)
	var cause *NetworkError
	if !errors.As(tcpStreamError(unix.ECONNRESET), &cause) || cause.Code != NetworkConnectionReset {
		t.Fatal(cause)
	}
	if socketError(&NetworkError{Code: NetworkAccessDenied, Cause: unix.ECONNREFUSED}) != NetworkAccessDenied {
		t.Fatal("error downcast lost structured policy result")
	}
}

//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// io.Closer permits implementations that return an error on a second close.
// A pending HTTP Read must still be interruptible by the first close.
type httpCloseProbe struct {
	closed chan struct{}
	once   sync.Once
	closes atomic.Int32
}

func (r *httpCloseProbe) Read([]byte) (int, error) { <-r.closed; return 0, io.ErrClosedPipe }
func (r *httpCloseProbe) Close() error {
	if r.closes.Add(1) != 1 {
		return errors.New("HTTP input closed more than once")
	}
	r.once.Do(func() { close(r.closed) })
	return nil
}

func TestHTTPTransportReaderClosesOnceOnTimeoutAndShutdown(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		p, _, _ := httpOwner(t, HTTPConfig{})
		body, err := p.newBody(false)
		if err != nil {
			t.Fatal(err)
		}
		reader := &httpCloseProbe{closed: make(chan struct{})}
		duration := time.Second
		if timeout {
			duration = time.Millisecond
		}
		if err := body.pump(reader, nil, duration, duration); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if timeout {
			select {
			case <-body.done:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
		p.Close()
		err = p.WaitClosed(ctx)
		cancel()
		if err != nil || reader.closes.Load() != 1 {
			t.Fatal("transport input ownership violated", timeout, reader.closes.Load(), err)
		}
	}
}

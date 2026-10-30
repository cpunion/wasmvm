package preview2

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPBodyBufferWrapBackpressureAndFinish(t *testing.T) {
	for _, capacity := range []uint32{0, 1<<20 + 1} {
		if _, err := newHTTPBodyBuffer(capacity); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	b, err := newHTTPBodyBuffer(5)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	out := make([]byte, 8)
	if n, err := b.read(ctx, out, false); n != 0 || !errors.Is(err, ErrWouldBlock) {
		t.Fatal(n, err)
	}
	if ready, _ := b.readiness(true); ready {
		t.Fatal("empty read ready")
	}
	ready, changed := b.readiness(false)
	if !ready {
		t.Fatal("empty buffer not writable")
	}
	if n, err := b.write(ctx, []byte("abcdef")); n != 5 || err != nil {
		t.Fatal(n, err)
	}
	select {
	case <-changed:
	default:
		t.Fatal("lost write notification")
	}
	if ready, _ := b.readiness(false); ready {
		t.Fatal("full buffer writable")
	}
	if n, err := b.write(ctx, []byte("x")); n != 0 || !errors.Is(err, ErrWouldBlock) {
		t.Fatal(n, err)
	}
	if n, err := b.read(ctx, out[:3], false); n != 3 || err != nil || string(out[:n]) != "abc" {
		t.Fatal(n, err)
	}
	if n, err := b.write(ctx, []byte("fg")); n != 2 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := b.read(ctx, out, false); n != 4 || err != nil || string(out[:n]) != "defg" {
		t.Fatal(n, err, string(out[:n]))
	}
	if n, err := b.write(ctx, nil); n != 0 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := b.read(ctx, nil, false); n != 0 || err != nil {
		t.Fatal(n, err)
	}
	if err := b.finish(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := b.read(ctx, out, false); !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if _, err := b.write(ctx, []byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := b.finish(nil); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}
func TestHTTPBodyBufferErrorsCancellationDiscardAndClose(t *testing.T) {
	ctx := context.Background()
	b, _ := newHTTPBodyBuffer(4)
	failure := errors.New("network body failed")
	b.write(ctx, []byte("data"))
	if err := b.finish(failure); err != nil {
		t.Fatal(err)
	}
	out := make([]byte, 8)
	if n, err := b.read(ctx, out, false); n != 4 || err != nil {
		t.Fatal(n, err)
	}
	if _, err := b.read(ctx, out, false); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	b.close()
	b.close()
	b, _ = newHTTPBodyBuffer(4)
	b.write(ctx, []byte("full"))
	b.discard()
	b.discard()
	if n, err := b.write(ctx, make([]byte, 1024)); n != 1024 || err != nil {
		t.Fatal(n, err)
	}
	b.mu.Lock()
	retained, written := b.count, b.written
	b.mu.Unlock()
	if retained != 0 || written != 1028 {
		t.Fatal(retained, written)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := b.write(cancelled, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := b.read(cancelled, out, true); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	b, _ = newHTTPBodyBuffer(4)
	done := make(chan error, 1)
	go func() { _, err := b.read(ctx, make([]byte, 1), true); done <- err }()
	b.close()
	select {
	case err := <-done:
		if !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("close stranded reader")
	}
	b, _ = newHTTPBodyBuffer(4)
	waiting, stop := context.WithCancel(ctx)
	go func() { _, err := b.read(waiting, make([]byte, 1), true); done <- err }()
	stop()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel stranded reader")
	}
}
func TestHTTPBodyBufferConcurrentTransport(t *testing.T) {
	b, _ := newHTTPBodyBuffer(17)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 1000; i++ {
			packet := []byte{byte(i), byte(i >> 8), 0}
			for len(packet) > 0 {
				n, err := b.write(ctx, packet)
				packet = packet[n:]
				if err != nil && !errors.Is(err, ErrWouldBlock) {
					done <- err
					return
				}
				if len(packet) > 0 {
					ready, changed := b.readiness(false)
					if ready {
						continue
					}
					select {
					case <-changed:
					case <-ctx.Done():
						done <- ctx.Err()
						return
					}
				}
			}
		}
		done <- b.finish(nil)
	}()
	output := make([]byte, 0, 3000)
	chunk := make([]byte, 13)
	for {
		n, err := b.read(ctx, chunk, true)
		output = append(output, chunk[:n]...)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(output) != 3000 {
		t.Fatal(len(output))
	}
	for i := 0; i < 1000; i++ {
		if output[i*3] != byte(i) || output[i*3+1] != byte(i>>8) || output[i*3+2] != 0 {
			t.Fatal("reordered stream", i)
		}
	}
}

func TestHTTPBodyBufferGrowthPreservesQueuedBytes(t *testing.T) {
	b, _ := newHTTPBodyBuffer(16384)
	ctx := context.Background()
	if n, err := b.write(ctx, []byte("abc")); n != 3 || err != nil {
		t.Fatal(n, err)
	}
	first := make([]byte, 2)
	if n, err := b.read(ctx, first, false); n != 2 || err != nil || string(first) != "ab" {
		t.Fatal(n, err)
	}
	tail := make([]byte, 10000)
	for i := range tail {
		tail[i] = byte(i)
	}
	if n, err := b.write(ctx, tail); n != len(tail) || err != nil {
		t.Fatal(n, err)
	}
	all := make([]byte, 10001)
	if n, err := b.read(ctx, all, false); n != len(all) || err != nil {
		t.Fatal(n, err)
	}
	if all[0] != 'c' {
		t.Fatal("lost queued prefix")
	}
	for i := range tail {
		if all[i+1] != tail[i] {
			t.Fatal("lost data while growing", i)
		}
	}
}

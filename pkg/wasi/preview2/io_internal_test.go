package preview2

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func testIOStream(t *testing.T, config IOConfig, input bool) (*IO, *ioStream) {
	t.Helper()
	p, err := NewIO(config)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(p.ctx)
	source := p.stdout
	if input {
		source = p.stdin
	}
	r := &ioStream{owner: p, id: 1, source: source, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
	p.records[1] = r
	p.next = 2
	source.refs++
	t.Cleanup(func() {
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
	})
	return p, r
}

type notifyOutput struct {
	*OutputBuffer
	blocked chan struct{}
}

func (b notifyOutput) Readiness() (bool, <-chan struct{}) {
	ready, changed := b.OutputBuffer.Readiness()
	if !ready {
		select {
		case b.blocked <- struct{}{}:
		default:
		}
	}
	return ready, changed
}
func TestIOOutputBackpressure(t *testing.T) {
	buffer := NewOutputBuffer(4)
	blocked := make(chan struct{}, 1)
	_, r := testIOStream(t, IOConfig{Stdout: notifyOutput{buffer, blocked}, MaxWriteBytes: 4}, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- r.blockingWrite(ctx, "abcdefgh") }()
	select {
	case <-blocked:
	case <-ctx.Done():
		t.Fatal("output never blocked")
	}
	if got := string(buffer.Drain()); got != "abcd" {
		t.Fatal(got)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("output did not resume")
	}
	if got := string(buffer.Drain()); got != "efgh" {
		t.Fatal(got)
	}
	if err := r.write("x"); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("unchecked write", err)
	}
	if _, err := r.checkWrite(); err != nil {
		t.Fatal(err)
	}
	if err := r.write("abcde"); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("over-permit write", err)
	}
}
func TestIOQueuedWriteAndFlushBackpressure(t *testing.T) {
	buffer := NewOutputBuffer(2)
	blocked := make(chan struct{}, 1)
	_, r := testIOStream(t, IOConfig{Stdout: notifyOutput{buffer, blocked}, MaxWriteBytes: 4}, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if n, err := r.checkWrite(); err != nil || n != 4 {
		t.Fatal(n, err)
	}
	if err := r.write("abcd"); err != nil {
		t.Fatal(err)
	}
	if err := r.flush(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked:
	case <-ctx.Done():
		t.Fatal("queued write never blocked")
	}
	if n, err := r.checkWrite(); err != nil || n != 0 {
		t.Fatal("flush permit", n, err)
	}
	if ready, _ := (outputPoll{r}).Readiness(); ready {
		t.Fatal("flush was ready early")
	}
	if got := string(buffer.Drain()); got != "ab" {
		t.Fatal(got)
	}
	if err := r.waitFlush(ctx); err != nil {
		t.Fatal(err)
	}
	if got := string(buffer.Drain()); got != "cd" {
		t.Fatal(got)
	}
	if ready, _ := (outputPoll{r}).Readiness(); !ready {
		t.Fatal("completed flush not ready")
	}
}
func TestIOPendingOutputCancellationAndClose(t *testing.T) {
	for _, queued := range []bool{false, true} {
		for _, closeProvider := range []bool{false, true} {
			buffer := NewOutputBuffer(1)
			blocked := make(chan struct{}, 1)
			p, r := testIOStream(t, IOConfig{Stdout: notifyOutput{buffer, blocked}, MaxWriteBytes: 8}, false)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			done := make(chan error, 1)
			go func() {
				if queued {
					_, err := r.checkWrite()
					if err == nil {
						err = r.write("abcdef")
					}
					if err == nil {
						err = r.flush()
					}
					if err == nil {
						err = r.waitFlush(ctx)
					}
					done <- err
				} else {
					done <- r.blockingWrite(ctx, "abcdef")
				}
			}()
			select {
			case <-blocked:
			case <-ctx.Done():
				t.Fatal("pending output not reached")
			}
			want := context.Canceled
			if closeProvider {
				p.Close()
				want = wasm.ErrClosed
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatal(queued, closeProvider, err)
				}
			case <-time.After(time.Second):
				t.Fatal("output did not cancel")
			}
			cancel()
		}
	}
}

type gatedInput struct {
	mu      sync.Mutex
	data    []byte
	closed  bool
	changed chan struct{}
	entered chan struct{}
}

func (g *gatedInput) Read(ctx context.Context, out []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, io.EOF
	}
	if len(out) == 0 {
		return 0, nil
	}
	if len(g.data) == 0 {
		return 0, ErrWouldBlock
	}
	n := copy(out, g.data)
	g.data = g.data[n:]
	return n, nil
}
func (g *gatedInput) Readiness() (bool, <-chan struct{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case g.entered <- struct{}{}:
	default:
	}
	return g.closed || len(g.data) != 0, g.changed
}
func (g *gatedInput) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.closed {
		g.closed = true
		close(g.changed)
		g.changed = make(chan struct{})
	}
	return nil
}
func (g *gatedInput) send(data string) {
	g.mu.Lock()
	g.data = append(g.data, data...)
	close(g.changed)
	g.changed = make(chan struct{})
	g.mu.Unlock()
}
func TestIOInputReadinessAndLimits(t *testing.T) {
	source := &gatedInput{changed: make(chan struct{}), entered: make(chan struct{}, 1)}
	_, r := testIOStream(t, IOConfig{Stdin: source, MaxReadBytes: 3}, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if data, err := r.read(ctx, 8, false); err != nil || len(data) != 0 {
		t.Fatal(data, err)
	}
	if data, err := r.read(ctx, 0, true); err != nil || len(data) != 0 {
		t.Fatal(data, err)
	}
	done := make(chan error, 1)
	go func() {
		data, err := r.read(ctx, ^uint64(0), true)
		if err == nil && string(data) != "abc" {
			err = errors.New("wrong bounded input")
		}
		done <- err
	}()
	select {
	case <-source.entered:
	case <-ctx.Done():
		t.Fatal("input not waiting")
	}
	source.send("abcdef")
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("input did not resume")
	}
	if data, err := r.read(ctx, 9, false); err != nil || string(data) != "def" {
		t.Fatal(data, err)
	}
	source.Close()
	if _, err := r.read(ctx, 1, true); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if _, err := r.read(ctx, 0, false); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}

type badCountInput struct{ BytesInput }

func (b *badCountInput) Read(context.Context, []byte) (int, error) { return -1, nil }

type badCountOutput struct {
	discardOutput
	panicWrite bool
}

func (b *badCountOutput) Write(context.Context, []byte) (int, error) {
	if b.panicWrite {
		panic("source write failed")
	}
	return -1, nil
}
func TestIOInvalidSourceCountsAndWorkerPanic(t *testing.T) {
	_, input := testIOStream(t, IOConfig{Stdin: &badCountInput{}}, true)
	if _, err := input.read(context.Background(), 1, false); err == nil {
		t.Fatal("invalid input count")
	}
	for _, panics := range []bool{false, true} {
		p, output := testIOStream(t, IOConfig{Stdout: &badCountOutput{panicWrite: panics}}, false)
		if _, err := output.checkWrite(); err != nil {
			t.Fatal(err)
		}
		if err := output.write("x"); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := output.flush(); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
		if err := output.waitFlush(ctx); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
		cancel()
		output.mu.Lock()
		failure := output.failure
		output.mu.Unlock()
		if failure == nil || panics && !strings.Contains(failure.Error(), "panicked") {
			t.Fatal(failure)
		}
		p.Close()
	}
}

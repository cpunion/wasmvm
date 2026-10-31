package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func closeIO(t *testing.T, p *IO) {
	t.Helper()
	p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.WaitClosed(ctx); err != nil {
		t.Fatal(err)
	}
}
func TestNativeStdioStreamsAndTerminalOwnership(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, drop := range []uint64{0, 1} {
			p, err := NewIO(IOConfig{Stdin: NewBytesInput([]byte("abc")), MaxReadBytes: 1})
			if err != nil {
				t.Fatal(err)
			}
			in := instantiate(t, compileFixture(t, backend, "stdin"), p.Imports())
			v := call(t, in, "run", component.Value{Bits: drop})
			want := uint64('a' + 'b' + 'c')
			if drop == 1 {
				want = 0
			}
			if v.Bits != want {
				t.Fatal(drop, v)
			}
			p.mu.Lock()
			remaining := len(p.operations)
			p.mu.Unlock()
			if remaining != 0 {
				t.Fatal("stdin retained operation", remaining)
			}
			in.Close()
			closeIO(t, p)
		}
		for _, name := range []string{"stdout", "stderr"} {
			for _, mode := range []uint64{0, 1, 2} {
				output := NewOutputBuffer(16)
				p, _ := NewIO(IOConfig{Stdout: output, Stderr: output, MaxWriteBytes: 1})
				in := instantiate(t, compileFixture(t, backend, name), p.Imports())
				v := call(t, in, "run", component.Value{Bits: mode})
				want := uint64(0)
				data := "abcdef"
				if mode == 1 {
					want = 1
					data = ""
				}
				if mode == 2 {
					data = ""
				}
				if v.Bits != want || string(output.Drain()) != data {
					t.Fatal(name, mode, v)
				}
				p.mu.Lock()
				remaining := len(p.operations)
				p.mu.Unlock()
				if remaining != 0 {
					t.Fatal("output retained operation", remaining)
				}
				in.Close()
				closeIO(t, p)
			}
		}
		for _, present := range []bool{false, true} {
			p, _ := NewIO(IOConfig{StdinTerminal: present, StdoutTerminal: present, StderrTerminal: present})
			for _, name := range []string{"stdin", "stdout", "stderr"} {
				// Reuse one Imports snapshot across generative Stores. Type views must come
				// from the actual active callback rather than a provider-global capture.
				m := compileFixture(t, backend, "terminal-"+name)
				imports := p.Imports()
				for range 2 {
					in := instantiate(t, m, imports)
					v := call(t, in, "run")
					want := uint64(0)
					if present {
						want = 1
					}
					if v.Bits != want {
						t.Fatal(name, v)
					}
					in.Close()
				}
			}
			closeIO(t, p)
		}
	})
}

type failingOutput struct {
	failure                                         error
	panicWrite, oversized, noProgress, flushFailure bool
}

func (f *failingOutput) Write(_ context.Context, data []byte) (int, error) {
	if f.panicWrite {
		panic("write")
	}
	if f.oversized {
		return len(data) + 1, nil
	}
	if f.noProgress {
		return 0, nil
	}
	if f.flushFailure {
		return len(data), nil
	}
	return 0, f.failure
}
func (f *failingOutput) Flush(context.Context) error      { return f.failure }
func (*failingOutput) Readiness() (bool, <-chan struct{}) { return true, nil }
func (*failingOutput) Close() error                       { return nil }
func TestOutputErrorsAreFutureResults(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		m := compileFixture(t, backend, "stdout")
		for _, item := range []struct {
			source *failingOutput
			want   uint64
		}{
			{&failingOutput{failure: io.ErrUnexpectedEOF}, 256}, {&failingOutput{failure: io.ErrClosedPipe}, 258}, {&failingOutput{failure: syscall.EILSEQ}, 257},
			{&failingOutput{panicWrite: true}, 256}, {&failingOutput{oversized: true}, 256}, {&failingOutput{noProgress: true}, 256}, {&failingOutput{failure: io.ErrUnexpectedEOF, flushFailure: true}, 256},
		} {
			p, _ := NewIO(IOConfig{Stdout: item.source, MaxWriteBytes: 1})
			in := instantiate(t, m, p.Imports())
			v := call(t, in, "run", component.Value{})
			if v.Bits != item.want {
				t.Fatal(item.source, v)
			}
			in.Close()
			closeIO(t, p)
		}
	})
}
func TestOutputBackpressureAndInstanceCancellation(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		m := compileFixture(t, backend, "stdout")
		output := NewOutputBuffer(1)
		p, _ := NewIO(IOConfig{Stdout: output, MaxWriteBytes: 1})
		defer closeIO(t, p)
		in := instantiate(t, m, p.Imports())
		f, _ := in.Function("run")
		done := make(chan error, 1)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go func() { _, err := f.Call(ctx, []component.Value{{}}, component.CallOptions{}); done <- err }()
		var data []byte
		for {
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
				data = append(data, output.Drain()...)
				if string(data) != "abcdef" {
					t.Fatal(string(data))
				}
				goto completed
			case <-time.After(time.Millisecond):
				data = append(data, output.Drain()...)
			case <-ctx.Done():
				t.Fatal("stream did not progress", ctx.Err())
			}
		}
	completed:
		output = NewOutputBuffer(1)
		blocked, _ := NewIO(IOConfig{Stdout: output, MaxWriteBytes: 1})
		defer closeIO(t, blocked)
		in = instantiate(t, m, blocked.Imports())
		f, _ = in.Function("run")
		go func() { _, err := f.Call(nil, []component.Value{{}}, component.CallOptions{}); done <- err }()
		deadline := time.Now().Add(time.Second)
		for {
			blocked.mu.Lock()
			active := blocked.activeWorkers
			blocked.mu.Unlock()
			if active > 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("output did not start")
			}
			time.Sleep(time.Millisecond)
		}
		in.Close()
		select {
		case err := <-done:
			if !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("close stranded output")
		}
		if err := in.WaitClosed(ctx); err != nil {
			t.Fatal(err)
		}
	})
}

type deferredInput struct {
	mu                 sync.Mutex
	ready, closed      bool
	changed            chan struct{}
	data               []byte
	failure            error
	panicRead, invalid bool
}

func (s *deferredInput) Read(ctx context.Context, out []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if s.panicRead {
		panic("input")
	}
	if s.invalid {
		return len(out) + 1, nil
	}
	if !s.ready {
		return 0, ErrWouldBlock
	}
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	if len(s.data) > 0 {
		n := copy(out, s.data)
		s.data = s.data[n:]
		return n, nil
	}
	return 0, s.failure
}
func (s *deferredInput) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ready || s.closed, s.changed
}
func (s *deferredInput) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if !s.ready {
		s.ready = true
		close(s.changed)
	}
	return nil
}
func TestInputReadinessErrorsAndProviderShutdown(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		m := compileFixture(t, backend, "stdin")
		for _, mode := range []string{"ready", "error", "panic", "invalid", "no-progress", "close"} {
			source := &deferredInput{changed: make(chan struct{}), failure: io.EOF}
			want := uint64(0)
			switch mode {
			case "ready":
				source.data = []byte("z")
				want = 'z'
			case "error":
				source.ready = true
				source.failure = io.ErrClosedPipe
				want = 1<<24 | 2<<16
			case "panic":
				source.ready = true
				source.panicRead = true
				want = 1 << 24
			case "invalid":
				source.ready = true
				source.invalid = true
				want = 1 << 24
			case "no-progress":
				source.ready = true
				source.failure = nil
				want = 1 << 24
			}
			p, _ := NewIO(IOConfig{Stdin: source})
			in := instantiate(t, m, p.Imports())
			f, _ := in.Function("run")
			done := make(chan error, 1)
			var got uint64
			go func() {
				r, err := f.Call(nil, []component.Value{{}}, component.CallOptions{})
				if err == nil {
					got = r.Value.Bits
				}
				done <- err
			}()
			if mode == "ready" {
				source.mu.Lock()
				source.ready = true
				close(source.changed)
				source.mu.Unlock()
			}
			if mode == "close" {
				p.Close()
			}
			select {
			case err := <-done:
				if mode == "close" {
					if !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
						t.Fatal(mode, err)
					}
				} else if err != nil || got != want {
					t.Fatal(mode, got, want, err)
				}
			case <-time.After(time.Second):
				t.Fatal(mode, "blocked")
			}
			in.Close()
			closeIO(t, p)
		}
	})
}
func TestIOLimitsAndSignatureValidation(t *testing.T) {
	if _, err := NewIO(IOConfig{MaxReadBytes: 1 << 28}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewIO(IOConfig{MaxWriteBytes: 1 << 28}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	p, _ := NewIO(IOConfig{MaxStreams: 1})
	defer closeIO(t, p)
	op, err := p.operation()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.operation(); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("operation limit", err)
	}
	op.finish(0)
	op.finish(1)
	m := compileFixture(t, "interpreter", "stdout-wrong-result")
	if _, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
	p.Close()
	if _, err := p.operation(); !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestDefaultIOZeroReadsAndPublicationQuotas(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := NewIO(IOConfig{})
		defer closeIO(t, p)
		for _, name := range []string{"stdin", "stdin-zero-read", "stdout", "stderr"} {
			in := instantiate(t, compileFixture(t, backend, name), p.Imports())
			if v := call(t, in, "run", component.Value{}); v.Bits != 0 {
				t.Fatal(name, v)
			}
		}
		for _, item := range []struct {
			name             string
			resources, tasks uint32
		}{
			{"stdin", 1, 10}, {"stdin", 2, 10}, {"stdout", 2, 10}, {"stdout", 4, 1},
		} {
			cfg := component.DefaultConfig()
			cfg.Core.Backend = backend
			cfg.MaxResources = item.resources
			cfg.MaxTasks = item.tasks
			engine, err := component.NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile("testdata/" + item.name + ".wasm")
			if err != nil {
				t.Fatal(err)
			}
			compiled, err := engine.Compile(nil, data)
			if err != nil {
				t.Fatal(err)
			}
			in := instantiate(t, compiled, p.Imports())
			f, _ := in.Function("run")
			_, err = f.Call(nil, []component.Value{{}}, component.CallOptions{})
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(item, "unbounded publication", err)
			}
			p.mu.Lock()
			remaining := len(p.operations)
			p.mu.Unlock()
			if remaining != 0 {
				t.Fatal(item, "publication retained operation", remaining)
			}
			in.Close()
			compiled.Close()
			engine.Close()
		}
		bad := compileFixture(t, backend, "stdin-wrong-stream")
		if _, err := bad.Instantiate(nil, p.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) {
			t.Fatal(err)
		}
	})
}
func TestInputPendingProviderCloseAndWorkerBounds(t *testing.T) {
	source := &deferredInput{changed: make(chan struct{}), failure: io.EOF}
	p, _ := NewIO(IOConfig{Stdin: source, MaxStreams: 1})
	in := instantiate(t, compileFixture(t, "interpreter", "stdin"), p.Imports())
	f, _ := in.Function("run")
	done := make(chan error, 1)
	go func() { _, err := f.Call(nil, []component.Value{{}}, component.CallOptions{}); done <- err }()
	deadline := time.Now().Add(time.Second)
	for {
		p.mu.Lock()
		active := p.activeWorkers
		p.mu.Unlock()
		if active > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("input never started")
		}
		time.Sleep(time.Millisecond)
	}
	p.Close()
	select {
	case err := <-done:
		if !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("provider close stranded input")
	}
	in.Close()
	closeIO(t, p)
	bounded, _ := NewIO(IOConfig{MaxStreams: 1})
	block := make(chan struct{})
	for range 2 {
		if err := bounded.worker(func() { <-block }); err != nil {
			t.Fatal(err)
		}
	}
	if err := bounded.worker(func() {}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("worker limit", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := bounded.WaitClosed(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	close(block)
	closeIO(t, bounded)
}
func TestIOSourceCloseFailureAndPanicCleanup(t *testing.T) {
	for _, panicked := range []bool{false, true} {
		source := &closeFailureOutput{panicked: panicked}
		p, _ := NewIO(IOConfig{Stdout: source})
		p.Close()
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err == nil {
			t.Fatal("lost close failure")
		}
		if source.closed != 1 {
			t.Fatal("closed source more than once")
		}
	}
}

type closeFailureOutput struct {
	discardOutput
	panicked bool
	closed   int
}

func (s *closeFailureOutput) Close() error {
	s.closed++
	if s.panicked {
		panic("close")
	}
	return io.ErrClosedPipe
}
func TestIOFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/io.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Version   string                       `json:"wasiVersion"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Fixtures  map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Version != "0.3.2" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 11 {
		t.Fatal("changed provenance")
	}
	check := func(path, want string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed fixture", path)
		}
	}
	check("../../../scripts/build_preview3_io_fixtures.py", pins.Generator)
	for name, files := range pins.Fixtures {
		for ext, hash := range files {
			check("testdata/"+name+"."+ext, hash)
		}
	}
}

func TestOutputRequiresTypedResultFuture(t *testing.T) {
	p, _ := NewIO(IOConfig{})
	defer closeIO(t, p)
	for _, name := range []string{"stdout-wrong-future", "stdout-wrong-direction"} {
		m := compileFixture(t, "interpreter", name)
		if _, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) {
			t.Fatal(name, err)
		}
	}
}

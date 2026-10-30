package preview2

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type countedOutput struct {
	discardOutput
	mu         sync.Mutex
	closes     int
	failure    error
	panicClose bool
}

func (s *countedOutput) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closes++
	if s.panicClose {
		panic("source close failed")
	}
	return s.failure
}
func (s *countedOutput) count() int { s.mu.Lock(); defer s.mu.Unlock(); return s.closes }

func TestIOOwnedFactoriesAndFailedPublication(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := component.DefaultConfig()
		config.Core.Backend = backend
		e, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, fixture(t, "io-streams"))
		if err != nil {
			t.Fatal(err)
		}
		process, err := NewIO(IOConfig{MaxStreams: 1})
		if err != nil {
			t.Fatal(err)
		}
		first, overflow := &countedOutput{}, &countedOutput{}
		var saved *component.Caller
		imports := process.Imports()
		imports.Instances["wasi:cli/stdout@0.2.12"].Functions["get-stdout"] = func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			saved = c
			value, err := process.NewOutput(c, first)
			if err != nil {
				return nil, err
			}
			if _, err := process.NewOutput(c, overflow); !errors.Is(err, wasm.ErrLimit) {
				return nil, errors.New("source quota not enforced")
			}
			return &value, nil
		}
		in, err := m.Instantiate(nil, imports, component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, err := in.Function("write-batch")
		if err != nil {
			t.Fatal(err)
		}
		result, err := f.Call(nil, []component.Value{{Bits: 0}}, component.CallOptions{})
		if err != nil || result.Value.Bits != 0 {
			t.Fatal(result, err)
		}
		if first.count() != 1 || overflow.count() != 1 {
			t.Fatal("source ownership", first.count(), overflow.count())
		}
		if value, err := process.NewInput(saved, NewBytesInput([]byte("x"))); value.Resource != nil || !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(value, err)
		}
		stale := &countedOutput{}
		if _, err := process.NewOutput(saved, stale); !errors.Is(err, wasm.ErrClosed) || stale.count() != 1 {
			t.Fatal(err, stale.count())
		}
		process.next = 1 << 32
		if _, err := process.publish(nil, process.stdout); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		in.Close()
		process.Close()
		if err := process.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		rejected := &countedOutput{}
		if _, err := process.NewOutput(nil, rejected); !errors.Is(err, wasm.ErrClosed) || rejected.count() != 0 {
			t.Fatal("closed owner took ownership", err, rejected.count())
		}
		rejected.Close()
		m.Close()
		e.Close()
	}
}

func TestIOInputFactoryAndCleanupFailures(t *testing.T) {
	e, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "io-streams"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	output := NewOutputBuffer(1024)
	process, err := NewIO(IOConfig{Stdout: output})
	if err != nil {
		t.Fatal(err)
	}
	imports := process.Imports()
	imports.Instances["wasi:cli/stdin@0.2.12"].Functions["get-stdin"] = func(c *component.Caller, _ []component.Value) (*component.Value, error) {
		value, err := process.NewInput(c, NewBytesInput([]byte("abcdefgh")))
		return &value, err
	}
	in, err := m.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	f, err := in.Function("exercise")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := f.Call(nil, nil, component.CallOptions{}); err != nil || v.Value.Bits != 42 {
		t.Fatal(v, err)
	}
	in.Close()
	process.Close()
	if err := process.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	for _, panics := range []bool{false, true} {
		sentinel := errors.New("source close failure")
		source := &countedOutput{failure: sentinel, panicClose: panics}
		p, err := NewIO(IOConfig{Stdout: source})
		if err != nil {
			t.Fatal(err)
		}
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = p.WaitClosed(ctx)
		cancel()
		if err == nil || !panics && !errors.Is(err, sentinel) {
			t.Fatal(panics, err)
		}
		if source.count() != 1 {
			t.Fatal(source.count())
		}
	}
}
func TestIOConfigurationAndEmptyOwners(t *testing.T) {
	for _, p := range []*IO{nil, {}} {
		p.Close()
		if p.Poller() != nil {
			t.Fatal("empty poll domain")
		}
		if len(p.Imports().Instances) != 0 {
			t.Fatal("empty provider")
		}
		if err := p.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
		if _, err := p.NewInput(nil, NewBytesInput(nil)); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
		if _, err := p.NewOutput(nil, &discardOutput{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
	}
	for _, config := range []IOConfig{{MaxReadBytes: 1 << 28}, {MaxWriteBytes: 1 << 28}, {Errors: IOErrorConfig{MaxDebugBytes: 1 << 28}}, {Poll: PollConfig{MaxPollItems: 65535}}} {
		if _, err := NewIO(config); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	p, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Poller() == nil {
		t.Fatal("missing poll domain")
	}
	if _, err := p.NewInput(nil, nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if _, err := p.NewOutput(nil, nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if _, err := p.record(nil, nil, true); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	p.Close()
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if err := p.begin(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := p.publish(nil, p.stdin); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestIOVirtualSourceLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := NewBytesInput([]byte("x"))
	if _, err := input.Read(ctx, make([]byte, 1)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	input.Close()
	if _, err := input.Read(context.Background(), nil); err == nil {
		t.Fatal("closed input")
	}
	zero := &OutputBuffer{}
	if _, err := zero.Write(context.Background(), []byte("zero")); err != nil {
		t.Fatal(err)
	}
	if got := string(zero.Drain()); got != "zero" {
		t.Fatal(got)
	}
	zero.Close()
	buffer := NewOutputBuffer(0)
	if _, err := buffer.Write(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := buffer.Flush(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	buffer.Close()
	if _, err := buffer.Write(context.Background(), nil); err == nil {
		t.Fatal("closed output")
	}
	if err := buffer.Flush(context.Background()); err == nil {
		t.Fatal("closed flush")
	}
	sink := &discardOutput{}
	if _, err := sink.Write(context.Background(), []byte("x")); err != nil {
		t.Fatal(err)
	}
	if ready, _ := sink.Readiness(); !ready {
		t.Fatal("discard not ready")
	}
	if err := sink.Flush(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	sink.Close()
	if _, err := sink.Write(context.Background(), nil); err == nil {
		t.Fatal("closed discard")
	}
}

package vm

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestThreadsSharingGrowthWaitAndClose(t *testing.T) {
	input, err := os.ReadFile("testdata/threads.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Profile = "core3"
		cfg.Features = wasm.Threads
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		parent, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer parent.Close()
		memory, err := parent.NewMemory(MemoryType{Min: 1, Max: 3, HasMax: true, Shared: true})
		if err != nil {
			t.Fatal(err)
		}
		capability, err := memory.ShareMemory()
		if err != nil {
			t.Fatal(err)
		}
		defer capability.Close()
		type child struct {
			s      *Store
			in     *Instance
			memory Extern
		}
		newChild := func() child {
			s, err := e.NewStore(DefaultStoreConfig())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { s.Close(); s.WaitClosed(nil) })
			ex, err := s.ImportMemory(capability)
			if err != nil {
				t.Fatal(err)
			}
			again, err := s.ImportMemory(capability)
			if err != nil || again.external.Memory != ex.external.Memory {
				t.Fatal("import alias", err)
			}
			l := NewLinker(s)
			if err := l.Define("env", "memory", ex); err != nil {
				t.Fatal(err)
			}
			return child{s, instantiate(t, l, m), ex}
		}
		a, b, c := newChild(), newChild(), newChild()
		call := func(ch child, name string, args ...Value) CallResult {
			t.Helper()
			r, err := export(t, ch.in, name).Call(nil, args, CallOptions{})
			if err != nil {
				t.Fatal(name, err)
			}
			return r
		}
		// Creator ownership can end while imported leases and the host capability live.
		parent.Close()
		if err := parent.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		call(c, "store", I32(0), I32(0))
		var workers sync.WaitGroup
		errorsOut := make(chan error, 2)
		for _, ch := range []child{a, b} {
			fn := export(t, ch.in, "add")
			workers.Add(1)
			go func() {
				defer workers.Done()
				for i := 0; i < 200; i++ {
					if _, err := fn.Call(nil, []Value{I32(0), I32(1)}, CallOptions{}); err != nil {
						errorsOut <- err
						return
					}
				}
			}()
		}
		workers.Wait()
		close(errorsOut)
		for err := range errorsOut {
			t.Fatal(err)
		}
		if bits(t, call(c, "load", I32(0)).Values[0]) != 400 {
			t.Fatal("lost atomic update")
		}
		if bits(t, call(c, "grow", I32(1)).Values[0]) != 1 || bits(t, call(a, "size").Values[0]) != 2 {
			t.Fatal("shared growth")
		}
		v := call(a, "mixed", I32(65536)).Values[0]
		lo, hi, err := v.Bits()
		if err != nil || lo != 11 || hi != 22 {
			t.Fatal(lo, hi, err)
		}
		v = call(b, "vector", I32(65536)).Values[0]
		lo, hi, _ = v.Bits()
		if lo != 11 || hi != 22 {
			t.Fatal("shared SIMD", lo, hi)
		}
		call(c, "fill", I32(65536), I32(7), I32(16))
		var raw [16]byte
		if err := b.memory.ReadMemory(65536, raw[:]); err != nil || raw[0] != 7 || raw[15] != 7 {
			t.Fatal(raw, err)
		}
		usage, err := a.s.Usage()
		if err != nil || usage.Core.SharedMemoryPages != 2 || usage.Core.ReservedSharedMemoryPages != 3 || usage.Core.MemoryPages != 3 {
			t.Fatal(usage, err)
		}
		call(c, "store", I32(0), I32(0))
		call(c, "store", I32(4), I32(0))
		waitFor := func(n uint32) {
			t.Helper()
			deadline := time.Now().Add(5 * time.Second)
			for e.domain.Waiters() != n {
				if time.Now().After(deadline) {
					t.Fatal("waiter count", e.domain.Waiters(), n)
				}
				time.Sleep(100 * time.Microsecond)
			}
		}
		type outcome struct {
			r   CallResult
			err error
		}
		first, second := make(chan outcome, 1), make(chan outcome, 1)
		wait32, wait64 := export(t, a.in, "wait32"), export(t, b.in, "wait64")
		go func() {
			r, err := wait32.Call(nil, []Value{I32(0), I32(0), I64(^uint64(0))}, CallOptions{})
			first <- outcome{r, err}
		}()
		waitFor(1)
		go func() {
			r, err := wait64.Call(nil, []Value{I32(0), I64(0), I64(^uint64(0))}, CallOptions{})
			second <- outcome{r, err}
		}()
		waitFor(2)
		if bits(t, call(c, "notify", I32(0), I32(0)).Values[0]) != 0 {
			t.Fatal("zero notification")
		}
		if bits(t, call(c, "notify", I32(0), I32(1)).Values[0]) != 1 {
			t.Fatal("notify count")
		}
		got := <-first
		if got.err != nil || bits(t, got.r.Values[0]) != 0 {
			t.Fatal(got)
		}
		waitFor(1)
		b.s.Close()
		got = <-second
		if !errors.Is(got.err, context.Canceled) {
			t.Fatal(got.err)
		}
		if err := b.s.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		waitFor(0)
		if bits(t, call(a, "wait32", I32(0), I32(1), I64(0)).Values[0]) != 1 || bits(t, call(a, "wait32", I32(0), I32(0), I64(1)).Values[0]) != 2 {
			t.Fatal("wait outcomes")
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := wait32.Call(ctx, []Value{I32(0), I32(0), I64(^uint64(0))}, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		low := DefaultStoreConfig()
		low.Limits.MaxMemoryPages = 2
		small, err := e.NewStore(low)
		if err != nil {
			t.Fatal(err)
		}
		defer small.Close()
		if _, err := small.ImportMemory(capability); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("reservation bypass", err)
		}
		other, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer other.Close()
		foreign, err := other.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer foreign.Close()
		if _, err := foreign.ImportMemory(capability); !errors.Is(err, ErrReference) {
			t.Fatal("cross-domain import", err)
		}
		capability.Close()
		if _, err := small.ImportMemory(capability); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		call(c, "store", I32(0), I32(42))
		if bits(t, call(a, "load", I32(0)).Values[0]) != 42 {
			t.Fatal("capability close retired imports")
		}
	})
}

func TestThreadsPrivateMemoryMemory64AndFeatureBoundaries(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Profile = "core3"
		without, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer without.Close()
		input, _ := os.ReadFile("testdata/atomic-fence.wasm")
		if err := without.Validate(nil, input); err == nil {
			t.Fatal("Threads implied by core3")
		}
		s0, err := without.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer s0.Close()
		if _, err := s0.NewMemory(MemoryType{Min: 1, Max: 1, HasMax: true, Shared: true}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
		cfg.Features = wasm.Threads
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		for _, name := range []string{"atomic-private", "atomic-fence", "threads64"} {
			input, err := os.ReadFile("testdata/" + name + ".wasm")
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, input)
			if err != nil {
				t.Fatal(name, err)
			}
			defer m.Close()
			s, err := e.NewStore(DefaultStoreConfig())
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			in := instantiate(t, NewLinker(s), m)
			if name == "atomic-fence" {
				if _, err := export(t, in, "fence").Call(nil, nil, CallOptions{Metered: true, Fuel: 1}); err != nil {
					t.Fatal(err)
				}
				continue
			}
			if name == "atomic-private" {
				for _, fn := range []string{"load", "notify"} {
					r, err := export(t, in, fn).Call(nil, nil, CallOptions{})
					if err != nil || bits(t, r.Values[0]) != 0 {
						t.Fatal(r, err)
					}
				}
				_, err := export(t, in, "wait").Call(nil, nil, CallOptions{})
				var trap *wasm.Trap
				if !errors.As(err, &trap) || trap.Code != wasm.TrapWaitUnshared {
					t.Fatal(err)
				}
				continue
			}
			for _, metered := range []bool{false, true} {
				r, err := export(t, in, "run").Call(nil, []Value{I64(65528)}, CallOptions{Metered: metered, Fuel: 20})
				if err != nil || bits(t, r.Values[0]) != 99 {
					t.Fatal(r, err)
				}
				r, err = export(t, in, "load").Call(nil, []Value{I64(65528)}, CallOptions{})
				if err != nil || bits(t, r.Values[0]) != 104 {
					t.Fatal(r, err)
				}
			}
			for _, at := range []uint64{1, 1 << 32, ^uint64(0)} {
				if _, err := export(t, in, "run").Call(nil, []Value{I64(at)}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(at, err)
				}
			}
		}
	})
}

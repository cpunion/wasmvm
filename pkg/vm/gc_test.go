package vm

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"os"
	"sync"
	"testing"
	"time"
)

func TestGCForcedCollectionCyclesHostRootsAndFailedInstances(t *testing.T) {
	input, err := os.ReadFile("testdata/gc.wasm")
	if err != nil {
		t.Fatal(err)
	}
	failed, err := os.ReadFile("testdata/gc-failed.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		if err := base.Validate(nil, input); err == nil {
			t.Fatal("GC enabled by default")
		}
		cfg := base.config
		cfg.Features = wasm.GC
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
		sc := DefaultStoreConfig()
		sc.CollectEveryAllocation = true
		sc.Limits.MaxHeapObjects = 64
		sc.Limits.MaxHeapBytes = 8192
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		in := instantiate(t, NewLinker(s), m)
		call := func(name string, args ...Value) CallResult {
			t.Helper()
			r, err := export(t, in, name).Call(nil, args, CallOptions{})
			if err != nil {
				t.Fatal(name, err)
			}
			return r
		}
		valueOf := func(v Value) uint64 { return bits(t, call("value", v).Values[0]) }
		null, _ := NullRef(wasm.ExternRef)
		payload, err := s.NewExternRef("payload survives internal objects")
		if err != nil {
			t.Fatal(err)
		}
		nested := call("nested").Values[0]
		child := call("next", nested).Values[0]
		if valueOf(child) != 88 {
			t.Fatal("constructor temporary root")
		}
		child.Release()
		nested.Release()
		a := call("new", I32(11), payload).Values[0]
		payload.Release()
		b := call("new", I32(22), null).Values[0]
		call("set-next", a, b)
		call("set-next", b, a)
		array := call("new-array", I32(3), a).Values[0]
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		if valueOf(a) != 11 || valueOf(b) != 22 {
			t.Fatal("cycle lost")
		}
		result := call("payload", a).Values[0]
		object, err := s.ExternRef(result)
		if err != nil || object != "payload survives internal objects" {
			t.Fatal(object, err)
		}
		result.Release()
		fn := call("function", a).Values[0]
		answer, err := s.CallRef(nil, fn, nil, CallOptions{})
		if err != nil || bits(t, answer.Values[0]) != 42 {
			t.Fatal("function field", answer, err)
		}
		fn.Release()
		if bits(t, call("vector").Values[0]) != 11 {
			t.Fatal("vector field")
		}
		subtype := call("subtype").Values[0]
		if bits(t, call("base-value", subtype).Values[0]) != 42 {
			t.Fatal("struct subtyping")
		}
		subtype.Release()
		hostType, err := export(t, in, "host").FunctionType()
		if err != nil {
			t.Fatal(err)
		}
		host, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(c *Caller, args []Value) ([]Value, error) {
			if err := c.Collect(); err != nil {
				return nil, err
			}
			return []Value{I32(uint32(bits(t, args[0])) + 2)}, nil
		}, hostType.Params[0])
		if err != nil {
			t.Fatal(err)
		}
		hostRef, err := host.Reference()
		if err != nil {
			t.Fatal(err)
		}
		if bits(t, call("host", hostRef, I32(40)).Values[0]) != 42 {
			t.Fatal("nominal host function binding")
		}
		hostRef.Release()
		plus, err := export(t, in, "plus").Reference()
		if err != nil {
			t.Fatal(err)
		}
		if bits(t, call("host", plus, I32(41)).Values[0]) != 42 {
			t.Fatal("function subtype call")
		}
		plus.Release()
		external := call("externalize", a).Values[0]
		opaque, err := s.ExternRef(external)
		if err != nil {
			t.Fatal(err)
		}
		copied, err := s.NewExternRef(opaque)
		if err != nil {
			t.Fatal(err)
		}
		external.Release()
		restored := call("internalize", copied).Values[0]
		if valueOf(restored) != 11 {
			t.Fatal("externalized object identity")
		}
		copied.Release()
		restored.Release()
		table := export(t, in, "table")
		if err := table.TableSet(0, b); err != nil {
			t.Fatal(err)
		}
		call("save", a)
		a.Release()
		b.Release()
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		a = call("load").Values[0]
		b = call("next", a).Values[0]
		if valueOf(a) != 11 || valueOf(b) != 22 {
			t.Fatal("global/table roots")
		}
		fromArray := call("get-array", array, I32(2)).Values[0]
		if valueOf(fromArray) != 11 {
			t.Fatal("array root")
		}
		fromArray.Release()
		before, _ := s.Usage()
		if bits(t, call("stress", I32(200)).Values[0]) != 1 {
			t.Fatal("automatic collection")
		}
		externalRoot := call("externalize", a).Values[0]
		concurrentCtx, cancelConcurrent := context.WithTimeout(context.Background(), 20*time.Millisecond)
		var wg sync.WaitGroup
		wg.Add(1)
		var concurrentErr error
		go func() {
			defer wg.Done()
			_, concurrentErr = export(t, in, "spin").Call(concurrentCtx, nil, CallOptions{})
		}()
		for range 500 {
			copy, err := s.Root(a)
			if err != nil {
				t.Fatal(err)
			}
			copy.Release()
			opaque, err := s.ExternRef(externalRoot)
			if err != nil {
				t.Fatal(err)
			}
			copy, err = s.NewExternRef(opaque)
			if err != nil {
				t.Fatal(err)
			}
			copy.Release()
		}
		wg.Wait()
		cancelConcurrent()
		externalRoot.Release()
		if !errors.Is(concurrentErr, context.DeadlineExceeded) {
			t.Fatal("concurrent root/guest execution", concurrentErr)
		}

		a.Release()
		b.Release()
		array.Release()
		typ, err := export(t, in, "new").FunctionType()
		if err != nil {
			t.Fatal(err)
		}
		nodeNull, _ := NullRef(typ.Results[0].WithNull())
		call("save", nodeNull)
		if err := table.TableSet(0, nodeNull); err != nil {
			t.Fatal(err)
		}
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		after, _ := s.Usage()
		if after.Core.HeapObjects != 1 || after.Core.HeapObjects >= before.Core.HeapObjects || after.ExternRefs != 0 {
			t.Fatal("unreachable cycle not reclaimed", before, after)
		}
		rollbackInput, err := os.ReadFile("testdata/gc-rollback.wasm")
		if err != nil {
			t.Fatal(err)
		}
		rollback, err := e.Compile(nil, rollbackInput)
		if err != nil {
			t.Fatal(err)
		}
		defer rollback.Close()
		usage, _ := s.Usage()
		for range 3 {
			if _, err := NewLinker(s).Instantiate(nil, rollback, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("constant heap allocation quota", err)
			}
			next, _ := s.Usage()
			if next != usage {
				t.Fatal("failed allocation retained heap/objects", usage, next)
			}
		}

		var escaped Value
		capture, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{typ.Results[0]}}, func(c *Caller, args []Value) ([]Value, error) {
			var err error
			escaped, err = c.Store().Root(args[0])
			if err != nil {
				return nil, err
			}
			return nil, c.Collect()
		})
		if err != nil {
			t.Fatal(err)
		}
		l := NewLinker(s)
		if err := l.Define("h", "escape", capture); err != nil {
			t.Fatal(err)
		}
		broken, err := e.Compile(nil, failed)
		if err != nil {
			t.Fatal(err)
		}
		defer broken.Close()
		if _, err := l.Instantiate(nil, broken, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("start did not trap", err)
		}
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		if valueOf(escaped) != 77 {
			t.Fatal("failed-instance object escaped incorrectly")
		}
		fn = call("function", escaped).Values[0]
		answer, err = s.CallRef(nil, fn, nil, CallOptions{})
		if err != nil || bits(t, answer.Values[0]) != 77 {
			t.Fatal("heap field lost failed instance", answer, err)
		}
		fn.Release()
		escaped.Release()
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		final, _ := s.Usage()
		if final.Core.Instances != 1 || final.Core.HeapObjects != 1 {
			t.Fatal("failed instance retained", final)
		}
		if _, err := export(t, in, "large").Call(nil, []Value{I32(^uint32(0))}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("heap quota ignored", err)
		}
		if _, err := export(t, in, "stress").Call(nil, []Value{I32(1000)}, CallOptions{Metered: true, Fuel: 8}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("allocation fuel", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err := export(t, in, "spin").Call(ctx, nil, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("allocation cancellation", err)
		}
		stale := call("new", I32(1), null).Values[0]
		stale.Release()
		if _, err := export(t, in, "value").Call(nil, []Value{stale}, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal("stale heap handle", err)
		}
		foreign, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer foreign.Close()
		foreignIn := instantiate(t, NewLinker(foreign), m)
		fr, err := export(t, foreignIn, "new").Call(nil, []Value{I32(1), null}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer fr.Values[0].Release()
		if _, err := export(t, in, "value").Call(nil, fr.Values, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal("foreign heap handle", err)
		}
	})
}

func TestGCCompositionWithEveryImplementedExtension(t *testing.T) {
	input, err := os.ReadFile("testdata/gc-composition.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		for _, disabled := range []wasm.Features{wasm.GC, wasm.Memory64, wasm.MultiMemory, wasm.TailCall, wasm.RelaxedSIMD, wasm.ExtendedConst} {
			cfg.Features = wasm.SupportedFeatures &^ disabled
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if err := e.Validate(nil, input); err == nil {
				t.Fatal("extension implicitly enabled", disabled)
			}
			e.Close()
		}
		cfg.Features = wasm.SupportedFeatures
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
		sc := DefaultStoreConfig()
		sc.Execution.MaxFrames = 1
		sc.CollectEveryAllocation = true
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		in := instantiate(t, NewLinker(s), m)
		for _, metered := range []bool{false, true} {
			r, err := export(t, in, "run").Call(nil, nil, CallOptions{Metered: metered, Fuel: 100})
			if err != nil || bits(t, r.Values[0]) != 60 {
				t.Fatal(r, err)
			}
		}
	})
}

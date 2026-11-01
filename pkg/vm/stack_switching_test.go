package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestStackSwitchingRootsLifetimesSignalsAndLimits(t *testing.T) {
	input, err := os.ReadFile("testdata/stack-switching.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Profile, cfg.Features = "core3", wasm.StackSwitching|wasm.CustomDescriptors
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
		setup := func(sc StoreConfig) (*Store, *Instance) {
			t.Helper()
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { s.Close(); s.WaitClosed(nil) })
			l := NewLinker(s)
			collect, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { return nil, c.Collect() })
			if err != nil {
				t.Fatal(err)
			}
			echo, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}, func(c *Caller, args []Value) ([]Value, error) { return args, c.Collect() })
			if err != nil {
				t.Fatal(err)
			}
			if err := l.Define("host", "collect", collect); err != nil {
				t.Fatal(err)
			}
			if err := l.Define("host", "echo", echo); err != nil {
				t.Fatal(err)
			}
			return s, instantiate(t, l, m)
		}
		for _, metered := range []bool{false, true} {
			sc := DefaultStoreConfig()
			sc.CollectEveryAllocation = true
			sc.Limits.MaxHeapBytes = 64 << 10
			s, in := setup(sc)
			options := CallOptions{Metered: metered, Fuel: 10000}
			call := func(name string, args ...Value) []Value {
				t.Helper()
				r, err := export(t, in, name).Call(nil, args, options)
				if err != nil {
					t.Fatal(name, err)
				}
				return r.Values
			}
			payload, err := s.NewExternRef("saved guest frames")
			if err != nil {
				t.Fatal(err)
			}
			captured := call("capture", payload)
			if len(captured) != 2 {
				t.Fatal(captured)
			}
			captured[0].Release()
			payload.Release()
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			// Invalid injection must leave the one-shot continuation available.
			if _, err := export(t, in, "null-inject").Call(nil, []Value{captured[1]}, options); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			null, _ := NullRef(wasm.ExternRef)
			result := call("finish", captured[1], null)[0]
			if got, err := s.ExternRef(result); err != nil || got != "saved guest frames" {
				t.Fatal(got, err)
			}
			_, err = export(t, in, "finish").Call(nil, []Value{captured[1], null}, options)
			var trap *wasm.Trap
			if !errors.As(err, &trap) || trap.Code != wasm.TrapConsumedContinuation {
				t.Fatal("reused continuation", err)
			}
			result.Release()
			captured[1].Release()
			for _, name := range []string{"new-bound", "new-host-bound"} {
				payload, err := s.NewExternRef(name)
				if err != nil {
					t.Fatal(err)
				}
				cont := call(name, payload)[0]
				payload.Release()
				if err := s.Collect(); err != nil {
					t.Fatal(err)
				}
				v := call("finish-bound", cont)[0]
				if got, err := s.ExternRef(v); err != nil || got != name {
					t.Fatal(got, err)
				}
				v.Release()
				cont.Release()
			}
			payload, err = s.NewExternRef("unhandled effect payload")
			if err != nil {
				t.Fatal(err)
			}
			_, err = export(t, in, "unhandled").Call(nil, []Value{payload}, options)
			var signal *Suspension
			if !errors.As(err, &signal) || !errors.Is(err, wasm.ErrSuspension) || errors.Is(err, wasm.ErrTrap) || signal.Switch() {
				t.Fatal(err)
			}
			payload.Release()
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			tag, values, err := signal.Payload()
			if err != nil || len(values) != 1 || !tag.SameTag(export(t, in, "yield")) {
				t.Fatal(tag, values, err)
			}
			if got, err := s.ExternRef(values[0]); err != nil || got != "unhandled effect payload" {
				t.Fatal(got, err)
			}
			tagType, err := tag.TagType()
			if err != nil || len(tagType.Results) != 1 || tagType.Results[0] != wasm.ExternRef {
				t.Fatal(tagType, err)
			}
			values[0].Release()
			signal.Release()
			if _, _, err := signal.Payload(); !errors.Is(err, ErrReference) {
				t.Fatal("released signal", err)
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err := s.Usage()
			if err != nil || usage.Core.HeapObjects != 0 || usage.Core.HeapBytes != 0 {
				t.Fatal("stack snapshot leak", usage, err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			_, err = export(t, in, "spin").Call(ctx, nil, CallOptions{})
			cancel()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("continuation polling", err)
			}
		}
		// A two-frame snapshot fits capture's three active frames, but not
		// restoration through the additional finish-deep caller.
		sc := DefaultStoreConfig()
		sc.CollectEveryAllocation = true
		sc.Execution.MaxFrames = 3
		s, in := setup(sc)
		null, _ := NullRef(wasm.ExternRef)
		r, err := export(t, in, "capture").Call(nil, []Value{null}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := export(t, in, "finish-deep").Call(nil, []Value{r.Values[1], null}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("restored frame quota", err)
		}
		r.Values[1].Release()
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		for _, bytes := range []uint64{0, 159} {
			sc := DefaultStoreConfig()
			sc.Limits.MaxHeapBytes = bytes
			_, in := setup(sc)
			if _, err := export(t, in, "new-bound").Call(nil, []Value{null}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("snapshot byte quota", bytes, err)
			}
		}
		var empty *Suspension
		empty.Release()
		if empty.Switch() {
			t.Fatal("nil signal")
		}
		if _, _, err := empty.Payload(); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
	})
}

func TestStackSwitchingHostFeatureGates(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		cfg := DefaultEngineConfig()
		cfg.Profile = "core3"
		if enabled {
			cfg.Features |= wasm.StackSwitching
		}
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		s, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		sig := FunctionType{Params: []wasm.ValueType{wasm.ContRef}, Results: []wasm.ValueType{wasm.ContRef}}
		fn, err := s.NewFunction(sig, func(_ *Caller, args []Value) ([]Value, error) { return args, nil })
		if !enabled {
			if !errors.Is(err, wasm.ErrUnsupported) {
				t.Fatal("continuation reference gate", err)
			}
			if _, err := s.NewTag(FunctionType{Results: []wasm.ValueType{wasm.I32}}); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("result-bearing tag without stack switching", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		null, err := NullRef(wasm.ContRef)
		if err != nil {
			t.Fatal(err)
		}
		result, err := fn.Call(nil, []Value{null}, CallOptions{})
		if err != nil || len(result.Values) != 1 || !result.Values[0].IsNull() {
			t.Fatal("host continuation reference", result, err)
		}
		tag, err := s.NewTag(FunctionType{Results: []wasm.ValueType{wasm.I32}})
		if err != nil {
			t.Fatal(err)
		}
		signature, err := tag.TagType()
		if err != nil || len(signature.Results) != 1 || signature.Results[0] != wasm.I32 {
			t.Fatal(signature, err)
		}
		if _, err := s.NewException(tag, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("control tag accepted as an exception", err)
		}
	}
}

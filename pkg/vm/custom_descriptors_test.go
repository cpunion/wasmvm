package vm

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCustomDescriptorsForcedGCRootsExceptionsAndLimits(t *testing.T) {
	input, err := os.ReadFile("testdata/custom-descriptors.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Profile = "core3"
		without, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer without.Close()
		if err := without.Validate(nil, input); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal("descriptors enabled in Core 3.0 by default", err)
		}
		blocked, err := without.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer blocked.Close()
		host, err := blocked.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
		if err != nil {
			t.Fatal(err)
		}
		ref, err := host.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer ref.Release()
		if _, err := blocked.NewTable(TableType{Element: ref.Type().WithNull().WithExact(true)}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("exact host type enabled without its feature", err)
		}
		cfg.Features |= wasm.CustomDescriptors
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
		for _, metered := range []bool{false, true} {
			sc := DefaultStoreConfig()
			sc.CollectEveryAllocation = true
			sc.Limits.MaxHeapObjects, sc.Limits.MaxHeapBytes = 4, 1024
			if metered {
				sc.Limits.MaxHeapObjects, sc.Limits.MaxHeapBytes = 10, 320
			}
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			collect, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { return nil, c.Collect() })
			if err != nil {
				t.Fatal(err)
			}
			l := NewLinker(s)
			if err := l.Define("host", "collect", collect); err != nil {
				t.Fatal(err)
			}
			in := instantiate(t, l, m)
			options := CallOptions{Metered: metered, Fuel: 1000}
			call := func(name string, args ...Value) Value {
				t.Helper()
				r, err := export(t, in, name).Call(nil, args, options)
				if err != nil || len(r.Values) != 1 {
					t.Fatal(name, err, r)
				}
				return r.Values[0]
			}
			payload, err := s.NewExternRef("descriptor payload")
			if err != nil {
				t.Fatal(err)
			}
			object := call("new", payload)
			payload.Release()
			if err := s.Collect(); err != nil || bits(t, call("check", object)) != 1 {
				t.Fatal("descriptor hidden edge lost", err)
			}
			held := call("payload", object)
			if got, err := s.ExternRef(held); err != nil || got != "descriptor payload" {
				t.Fatal(got, err)
			}
			held.Release()
			null, _ := NullRef(wasm.ExternRef)
			extra := call("new", null)
			if _, err := export(t, in, "new").Call(nil, []Value{null}, options); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("descriptor objects escaped object/byte quotas", err)
			}
			extra.Release()
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err := s.Usage()
			if err != nil || usage.Core.HeapObjects != 2 || usage.Core.HeapBytes != 160 {
				t.Fatal("descriptor hidden-edge accounting", usage, err)
			}
			echo := call("echo-exception", object)
			object.Release()
			if bits(t, call("check", echo)) != 1 {
				t.Fatal("descriptor roots lost across exception/host collection")
			}
			echo.Release()
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err = s.Usage()
			if err != nil || usage.Core.HeapObjects != 0 || usage.Core.HeapBytes != 0 {
				t.Fatal("descriptor cycle retained", usage, err)
			}
		}
	})
}

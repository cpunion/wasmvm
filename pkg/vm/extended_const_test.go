package vm

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExtendedConstPublicAPI(t *testing.T) {
	input, err := os.ReadFile("testdata/extended-const.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		if err := base.Validate(nil, input); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("extension enabled by default", err)
		}
		config := base.config
		config.Features = wasm.Features(1 << 63)
		if _, err := NewEngineWithConfig(config); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
		config.Features = wasm.ExtendedConst
		e, err := NewEngineWithConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		info, err := e.Inspect(nil, input)
		if err != nil || info.Features != wasm.ExtendedConst || info.Profile != "core2" {
			t.Fatal(info, err)
		}
		compiled, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer compiled.Close()
		info, err = compiled.Info()
		if err != nil || info.Features != wasm.ExtendedConst {
			t.Fatal(info, err)
		}
		// Each instance resolves its own immutable imports; compilation cannot
		// cache values or offsets from a previously linked Store.
		for _, n := range []uint32{3, 5} {
			s, err := e.NewStore(DefaultStoreConfig())
			if err != nil {
				t.Fatal(err)
			}
			l := NewLinker(s)
			for _, g := range []struct {
				name  string
				value Value
			}{{"i32", I32(n)}, {"i64", I64(uint64(n))}} {
				x, err := s.NewGlobal(GlobalType{Value: g.value.Type()}, g.value)
				if err != nil {
					t.Fatal(err)
				}
				if err := l.Define("base", g.name, x); err != nil {
					t.Fatal(err)
				}
			}
			in := instantiate(t, l, compiled)
			y, err := export(t, in, "y").GlobalGet()
			if err != nil || bits(t, y) != uint64(n)*20-17 {
				t.Fatal(y, err)
			}
			for _, metered := range []bool{false, true} {
				got, err := export(t, in, "run").Call(nil, nil, CallOptions{Metered: metered, Fuel: 100})
				if err != nil || bits(t, got.Values[0]) != uint64(n)*20+6 {
					t.Fatal(got, err)
				}
			}
			s.Close()
			if err := s.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestSuccessfulPublicationStillCollectsReleasedExternRefs(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		v, err := s.NewExternRef("released before publication")
		if err != nil {
			t.Fatal(err)
		}
		v.Release()
		compiled, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer compiled.Close()
		instantiate(t, l, compiled)
		u, err := s.Usage()
		if err != nil || u.ExternRefs != 0 || u.Core.Instances != 1 {
			t.Fatal(u, err)
		}
	})
}

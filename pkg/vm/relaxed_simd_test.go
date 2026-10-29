package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRelaxedSIMDCompositionFuelAndCancellation(t *testing.T) {
	input, err := os.ReadFile("testdata/relaxed-simd.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		for _, disabled := range []wasm.Features{0, wasm.ExtendedConst, wasm.TailCall, wasm.MultiMemory, wasm.RelaxedSIMD} {
			cfg := base.config
			cfg.Features = wasm.SupportedFeatures &^ disabled
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, input)
			if disabled != 0 {
				if err == nil {
					t.Fatal("extension implicitly enabled", disabled)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			sc := DefaultStoreConfig()
			sc.Execution.MaxFrames = 1
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			in := instantiate(t, NewLinker(s), m)
			for _, metered := range []bool{false, true} {
				r, err := export(t, in, "run").Call(nil, nil, CallOptions{Metered: metered, Fuel: 100})
				if err != nil || bits(t, r.Values[0]) != 112 {
					t.Fatal(metered, r, err)
				}
				r, err = export(t, in, "repeat").Call(nil, []Value{I32(1000)}, CallOptions{Metered: metered, Fuel: 20000})
				if err != nil || bits(t, r.Values[0]) != 28000 {
					t.Fatal(metered, r, err)
				}
			}
			if _, err := export(t, in, "repeat").Call(nil, []Value{I32(1000)}, CallOptions{Metered: true, Fuel: 10}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("fuel ignored", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			_, err = export(t, in, "repeat").Call(ctx, []Value{I32(0)}, CallOptions{})
			cancel()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("cancellation ignored", err)
			}
			if _, err := export(t, in, "run").Call(nil, nil, CallOptions{}); err != nil {
				t.Fatal("workspace not reusable after cancellation", err)
			}
		}
	})
}

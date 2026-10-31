package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestStackfulSuspensionAndBackpressure(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, name := range []string{"async-stackful", "async-stackful-sync-lower"} {
				in := asyncInstance(t, backend, name, Imports{})
				f, _ := in.Function("churn")
				for _, n := range []uint64{0, 1, 17, 128} {
					v, err := f.Call(nil, []Value{{Bits: n}}, CallOptions{})
					if err != nil || v.Value == nil || v.Value.Bits != n {
						t.Fatal(name, n, v, err)
					}
				}
			}
			in := asyncInstance(t, backend, "async-stackful", Imports{})
			cancel, _ := in.Function("cancel")
			for _, async := range []uint64{0, 1} {
				v, err := cancel.Call(nil, []Value{{Bits: async}}, CallOptions{})
				want := uint64(2)
				if async != 0 {
					want = uint64(^uint32(0))
				}
				if err != nil || v.Value == nil || v.Value.Bits != want {
					t.Fatal(async, v, err)
				}
			}
			pressure := asyncInstance(t, backend, "async-backpressure", Imports{})
			run, _ := pressure.Function("pressure")
			for _, cancel := range []uint64{0, 1, 0, 1} {
				v, err := run.Call(nil, []Value{{Bits: cancel}}, CallOptions{})
				want := uint64(49)
				if cancel != 0 {
					want = 3
				}
				if err != nil || v.Value == nil || v.Value.Bits != want {
					t.Fatal("backpressure", cancel, v, err)
				}
			}
			for _, mode := range []uint64{0, 2} {
				failed := asyncInstance(t, backend, "async-backpressure", Imports{})
				control, _ := failed.Function("control")
				if _, err := control.Call(nil, []Value{{Bits: mode}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("counter bounds", mode, err)
				}
				if _, err := control.Call(nil, []Value{{Bits: 1}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("failed control reentered", err)
				}
			}

		})
	}
}

package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestFutureStreamGuestProtocol(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, name := range []string{"async-stream-protocol", "async-future-protocol", "async-stream-memory64", "async-future-memory64"} {
				in := asyncInstance(t, backend, name, Imports{})
				run, _ := in.Function("run")
				for _, n := range []uint64{0, 42, 0xffffffff} {
					v, err := run.Call(nil, []Value{{Bits: n}}, CallOptions{})
					if err != nil || v.Value == nil || v.Value.Bits != n {
						t.Fatal(name, n, v, err)
					}
					if in.tasks != nil {
						t.Fatal("retained protocol state")
					}
				}
			}
			for _, name := range []string{"async-stream-strings", "async-future-strings", "async-stream-sync-copy", "async-future-sync-copy"} {
				in := asyncInstance(t, backend, name, Imports{Functions: map[string]HostFunction{"pass": func(_ *Caller, args []Value) (*Value, error) { return &args[0], nil }}})
				run, _ := in.Function("run")
				for range 3 {
					v, err := run.Call(nil, nil, CallOptions{})
					if err != nil || v.Value == nil || v.Value.Text != "café雪" {
						t.Fatal(name, v, err)
					}
				}
			}

			for _, name := range []string{"async-stream-edges", "async-future-edges"} {
				for _, mode := range []uint64{1, 2, 3, 4, 5, 6, 8, 9, 10, 11, 12} {
					in := asyncInstance(t, backend, name, Imports{})
					f, _ := in.Function("edges")
					v, err := f.Call(nil, []Value{{Bits: mode}}, CallOptions{})
					if mode == 8 || mode == 9 || mode == 10 {
						want := uint64(42)
						if mode == 8 {
							want = 1
						}
						if err != nil || v.Value == nil || v.Value.Bits != want {
							t.Fatal(name, mode, v, err)
						}
					} else if !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("invalid protocol", name, mode, err)
					}
					if in.tasks != nil {
						t.Fatal("failed protocol retained state")
					}
				}
			}
			unit := asyncInstance(t, backend, "async-stream-unit", Imports{})
			run, _ := unit.Function("run")
			if v, err := run.Call(nil, []Value{{Bits: 42}}, CallOptions{}); err != nil || v.Value == nil || v.Value.Bits != 42 {
				t.Fatal("unit stream", v, err)
			}

			zero := asyncInstance(t, backend, "async-stream-zero", Imports{})
			zf, _ := zero.Function("zero")
			for _, mode := range []uint64{0, 1, 2, 3} {
				v, err := zf.Call(nil, []Value{{Bits: mode}}, CallOptions{})
				if err != nil || v.Value == nil || v.Value.Bits != 0 {
					t.Fatal("zero readiness", mode, v, err)
				}
			}

			bytes := asyncInstance(t, backend, "async-stream-bytes", Imports{})
			bf, _ := bytes.Function("run")
			for _, n := range []uint64{0, 42, 255} {
				v, err := bf.Call(nil, []Value{{Bits: n}}, CallOptions{})
				if err != nil || v.Value == nil || v.Value.Bits != n {
					t.Fatal("byte buffer", v, err)
				}
			}

		})
	}
}

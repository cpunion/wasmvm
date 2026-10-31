package component

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLowerComposition(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"async-lower", "async-lower64", "async-lower-sync-callee"} {
			in := asyncInstance(t, backend, name, Imports{})
			f, _ := in.Function("churn")
			for _, count := range []uint64{0, 1, 17, 1000} {
				v, err := f.Call(nil, []Value{{Bits: count}}, CallOptions{})
				if err != nil || v.Value.Bits != count {
					t.Fatal(backend, name, count, v, err)
				}
			}
		}
	}
}

func TestAsyncLowerArgumentABI(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			calls := 0
			sum := func(_ *Caller, args []Value) (*Value, error) {
				calls++
				var n uint64
				for _, arg := range args {
					n += arg.Bits
				}
				return &Value{Bits: n}, nil
			}
			imports := Imports{AsyncFunctions: map[string]HostFunction{"a4": sum, "a5": sum, "mixed": func(_ *Caller, args []Value) (*Value, error) {
				return &Value{Bits: args[0].Bits + args[1].Bits + uint64(math.Float32frombits(uint32(args[2].Bits))) + uint64(math.Float64frombits(args[3].Bits))}, nil
			}, "unit": func(*Caller, []Value) (*Value, error) { return nil, nil }}}
			in := asyncInstance(t, backend, "async-lower-arity", imports)
			for _, name := range []string{"a4", "a5", "mixed", "unit"} {
				f, _ := in.Function(name)
				var args []Value
				want := uint64(2)
				switch name {
				case "a4":
					args = []Value{{Bits: 1}, {Bits: 2}, {Bits: 3}, {Bits: 4}}
					want = 10
				case "a5":
					args = []Value{{Bits: 1}, {Bits: 2}, {Bits: 3}, {Bits: 4}, {Bits: 5}}
					want = 15
				case "mixed":
					args = []Value{{Bits: 1}, {Bits: 1 << 40}, {Bits: uint64(math.Float32bits(2.5))}, {Bits: math.Float64bits(3.5)}}
					want = 1<<40 + 6
				}
				v, err := f.Call(nil, args, CallOptions{})
				if err != nil || v.Value.Bits != want {
					t.Fatal(name, v, err)
				}
			}
			for _, name := range []string{"bad-args", "bad-result"} {
				for _, ptr := range []uint64{2, 65536} {
					fresh := asyncInstance(t, backend, "async-lower-arity", imports)
					f, _ := fresh.Function(name)
					before := calls
					if _, err := f.Call(nil, []Value{{Bits: ptr}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal(name, err)
					}
					want := before
					if name == "bad-result" {
						want++
					}
					if calls != want {
						t.Fatal("wrong effect ordering", name, calls, want)
					}
				}
			}
		})
	}
}

func TestAsyncLowerStringEncodings(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"async-lower-strings", "async-lower-strings-utf16", "async-lower-strings-compact", "async-lower-strings64"} {
			in := asyncInstance(t, backend, name, Imports{})
			f, _ := in.Function("echo")
			for _, text := range []string{"", "abc", "café", "return λ 😃"} {
				v, err := f.Call(nil, []Value{{Text: text}}, CallOptions{})
				if err != nil || v.Value.Text != text {
					t.Fatal(backend, name, v, err)
				}
			}
		}
	}
}

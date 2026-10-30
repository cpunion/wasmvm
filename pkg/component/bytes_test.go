package component

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestByteTransfersOwnMemoryAndPreserveEveryByte(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"bytes", "bytes64", "bytes-host", "bytes-host64"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				config := DefaultConfig()
				config.Core.Backend = backend
				e, err := NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer e.Close()
				m, err := e.Compile(nil, fixture(t, name))
				if err != nil {
					t.Fatal(err)
				}
				defer m.Close()
				callbacks := 0
				in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{
					"echo": func(_ *Caller, args []Value) (*Value, error) {
						callbacks++
						if args[0].Elements != nil {
							t.Fatal("host byte input expanded into value slots")
						}
						return &args[0], nil
					},
				}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				f, err := in.Function("echo")
				if err != nil {
					t.Fatal(err)
				}
				for _, n := range []int{0, 1, 256, 4096} {
					data := make([]byte, n)
					for i := range data {
						data[i] = byte(i)
					}
					for _, packed := range []bool{false, true} {
						v := ByteList(data)
						if !packed {
							v = Value{Elements: make([]Value, n)}
							for i, b := range data {
								v.Elements[i].Bits = uint64(b)
							}
						}
						got, err := f.Call(nil, []Value{v}, CallOptions{Metered: true, Fuel: 10000})
						if err != nil || got.Value == nil {
							t.Fatal(got, err)
						}
						s, err := got.Value.ByteString()
						if err != nil || !bytes.Equal([]byte(s), data) || got.Value.Elements != nil {
							t.Fatal(n, packed, err)
						}
						// Reusing/resetting guest allocation cannot mutate a retained result.
						if _, err := f.Call(nil, []Value{ByteList(bytes.Repeat([]byte{99}, n))}, CallOptions{}); err != nil {
							t.Fatal(err)
						}
						if got.Value.Text != string(data) {
							t.Fatal("result aliases reset guest memory")
						}
					}
				}
				before := callbacks
				if _, err := f.Call(nil, []Value{{Text: "x", Elements: []Value{}}}, CallOptions{}); !errors.Is(err, wasm.ErrInvalid) || callbacks != before {
					t.Fatal("mixed input reached guest", callbacks, err)
				}
				if _, err := f.Call(nil, []Value{ByteList([]byte{255})}, CallOptions{}); err != nil {
					t.Fatal("rejection poisoned instance", err)
				}
			})
		}
	}
}

package component

import (
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestCollectionCallsThroughBothCoreBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"map", "fixed3", "fixed17"} {
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
				in, err := m.Instantiate(nil, Imports{Instances: map[string]Imports{"host": {Functions: map[string]HostFunction{
					"echo": func(_ *Caller, args []Value) (*Value, error) { return &args[0], nil },
				}}}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				f, err := in.Function("echo")
				if err != nil {
					t.Fatal(err)
				}
				value := Value{}
				if name == "map" {
					value.Elements = []Value{{Elements: []Value{{Text: "同"}, {Bits: 1}}}, {Elements: []Value{{Text: "同"}, {Bits: 2}}}}
				} else {
					n := 3
					if name == "fixed17" {
						n = 17
					}
					value.Elements = make([]Value, n)
					for i := range value.Elements {
						value.Elements[i].Bits = uint64(i)
					}
				}
				got, err := f.Call(nil, []Value{value}, CallOptions{})
				if err != nil || !reflect.DeepEqual(got.Value, &value) {
					t.Fatal(got, err)
				}
			})
		}
	}
}

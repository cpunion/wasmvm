package preview2

import (
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"os"
	"testing"
)

func BenchmarkEnvironmentArgumentCount(b *testing.B) {
	input, err := os.ReadFile("testdata/environment.wasm")
	if err != nil {
		b.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		b.Run(backend, func(b *testing.B) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			e, err := component.NewEngine(config)
			if err != nil {
				b.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, input)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			env, err := NewEnvironment(EnvironmentConfig{Args: []string{"guest", "", "界🙂"}, Env: []string{"A=1=2", "B=second"}})
			if err != nil {
				b.Fatal(err)
			}
			in, err := m.Instantiate(nil, env.Imports(), component.CallOptions{})
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			f, err := in.Function("argument-count")
			if err != nil {
				b.Fatal(err)
			}
			if v, err := f.Call(nil, nil, component.CallOptions{}); err != nil || v.Value.Bits != 3 {
				b.Fatal(v, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := f.Call(nil, nil, component.CallOptions{})
				if err != nil || v.Value.Bits != 3 {
					b.Fatal(v, err)
				}
			}
		})
	}
}

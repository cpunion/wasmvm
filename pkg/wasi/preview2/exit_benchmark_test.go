package preview2

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

// Each nonreturning call consumes a fresh instance. Batch preparation and
// export lookup are excluded; Store cancellation and reclamation are included.
func BenchmarkExitCall(b *testing.B) {
	input, err := os.ReadFile("testdata/exit.wasm")
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
			imports := ExitImports()
			args := []component.Value{{Bits: 42}}
			b.ReportAllocs()
			b.ResetTimer()
			b.StopTimer()
			for done := 0; done < b.N; {
				count := min(256, b.N-done)
				instances := make([]*component.Instance, count)
				functions := make([]*component.Function, count)
				errorsOut := make([]error, count)
				for i := range instances {
					instances[i], err = m.Instantiate(nil, imports, component.CallOptions{})
					if err != nil {
						b.Fatal(err)
					}
					functions[i], err = instances[i].Function("with-code")
					if err != nil {
						b.Fatal(err)
					}
				}
				b.StartTimer()
				for i, f := range functions {
					_, errorsOut[i] = f.Call(nil, args, component.CallOptions{})
				}
				b.StopTimer()
				for i, err := range errorsOut {
					var exit *ExitError
					if !errors.As(err, &exit) || exit.Code != 42 {
						b.Fatal(err)
					}
					if err := instances[i].WaitClosed(nil); err != nil {
						b.Fatal(err)
					}
					instances[i].Close()
				}
				done += count
			}
		})
	}
}

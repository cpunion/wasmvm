package component

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func FuzzComponentExecution(f *testing.F) {
	for _, path := range []string{"../../internal/component/decode/testdata/composition.wasm", "testdata/strings.wasm"} {
		bytes, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(bytes)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		config := DefaultConfig()
		config.Decode.MaxBytes, config.Decode.MaxNodes, config.Decode.MaxNesting = 1<<16, 1<<14, 16
		config.Validate.MaxNodes, config.Validate.MaxDepth = 1<<14, 16
		config.Core.Features = wasm.Core3Features
		config.Store.Limits.MaxInstances, config.Store.Limits.MaxMemoryPages = 8, 16
		config.Store.Limits.MaxHeapObjects, config.Store.Limits.MaxHeapBytes = 256, 1<<20
		config.Store.Limits.MaxTableElements = 4096
		config.Store.Execution.MaxFrames, config.Store.Execution.MaxSlots = 32, 4096
		config.MaxValues, config.MaxBytes, config.MaxNesting = 1024, 1<<16, 16
		e, err := NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		m, err := e.Compile(ctx, input)
		if err != nil {
			if m != nil {
				t.Fatal("failed compilation published code")
			}
			return
		}
		defer m.Close()
		in, err := m.Instantiate(ctx, Imports{}, vm.CallOptions{Metered: true, Fuel: 1000})
		if err != nil && in != nil {
			t.Fatal("failed instantiation published an instance")
		}
		if in != nil {
			in.Close()
			if err := in.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		}
	})
}

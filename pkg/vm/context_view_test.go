package vm

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
	"time"
)

func TestCallerContextsRetainMetadataAndExpireIndependently(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		engine, store, linker := setupWithEngine(t, newEngine)
		type key struct{}
		var retained []context.Context
		host, err := store.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			current := c.Context()
			deadline, ok := current.Deadline()
			if current.Err() != nil || current.Value(key{}) != len(retained) || !ok || time.Until(deadline) <= 0 {
				t.Fatal("context metadata", deadline, ok, current.Err())
			}
			for _, old := range retained {
				if !errors.Is(old.Err(), context.Canceled) {
					t.Fatal("old operation reactivated")
				}
			}
			retained = append(retained, current)
			child, cancel := context.WithCancel(current)
			defer cancel()
			if child.Err() != nil || child.Value(key{}) != current.Value(key{}) {
				t.Fatal("derived child metadata")
			}
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err = linker.Define("h", "inspect", host); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "inspect", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Index: 1}}}
		fn := export(t, instantiate(t, linker, compileDefinition(t, engine, d)), "run")
		for i := 0; i < 3; i++ {
			parent, stop := context.WithDeadline(context.WithValue(context.Background(), key{}, i), time.Now().Add(time.Minute))
			_, err := fn.Call(parent, nil, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			current := retained[i]
			if !errors.Is(current.Err(), context.Canceled) || !errors.Is(context.Cause(current), context.Canceled) || parent.Err() != nil {
				t.Fatal("operation expiry changed parent or cause")
			}
			stop()
		}
		cause := errors.New("caller cancelled")
		parent, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		if _, err := fn.Call(parent, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal("missed parent cancellation", err)
		}
		if len(retained) != 3 {
			t.Fatal("cancelled host callback ran")
		}
		if _, err := fn.Call(nil, nil, CallOptions{Metered: true, Fuel: 0}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("fuel bypassed", err)
		}
	})
}

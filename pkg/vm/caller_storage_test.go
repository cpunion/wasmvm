package vm

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCallbackCapabilitiesStayExpiredAcrossBatchGrowth(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, store, _ := setupWithEngine(t, newEngine)
		var retained []Caller
		var host Extern
		host, err := store.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error {
			for i := range retained {
				if _, err := retained[i].Call(host, nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("an earlier capability became usable", i, err)
				}
			}
			retained = append(retained, *c)
			if len(retained) == 10 {
				runtime.GC()
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			_, err := store.WithCaller(nil, CallOptions{}, func(root *Caller) error {
				for range 20 {
					if _, err := root.Call(host, nil); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			for i := range retained {
				if _, err := retained[i].Call(host, nil); !errors.Is(err, wasm.ErrClosed) || retained[i].Context().Err() != context.Canceled {
					t.Fatal("retained capability or context revived", i, err)
				}
			}
		}
	})
}

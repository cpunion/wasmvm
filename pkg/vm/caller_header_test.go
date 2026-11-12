package vm

import (
	"errors"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestCallerHeaderReplacementStillDrainsCallback(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		fn, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error { *c = Caller{}; return nil })
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fn.Call(nil, nil, CallOptions{}); err != nil {
			t.Fatal(err)
		}
		var empty Caller
		if err := empty.Abort(nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("empty caller accepted abort", err)
		}
	})
}

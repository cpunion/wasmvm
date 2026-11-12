package vm

import (
	"errors"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestCopiedCoreCallerExpiresBeforeSessionEnds(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		var copied Caller
		host, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error { copied = *c; return nil })
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.WithCaller(nil, CallOptions{}, func(session *Caller) error {
			if _, err := session.Call(host, nil); err != nil {
				return err
			}
			root, err := copied.NewExternRef("expired copied callback")
			root.Release()
			if !errors.Is(err, wasm.ErrClosed) {
				t.Fatalf("copied callback outlived host return while session remained active: %v", err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

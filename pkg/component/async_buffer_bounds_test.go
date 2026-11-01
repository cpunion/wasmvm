package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTypedEmptyAsyncBufferBounds(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, fixture := range []string{"async-stream-protocol", "async-stream-memory64"} {
			t.Run(backend+"/"+fixture, func(t *testing.T) {
				in := asyncInstance(t, backend, fixture, Imports{})
				run, _ := in.Function("run")
				memoryIndex := uint32(0)
				p := &validate.CanonicalFunction{Definition: &model.Canonical{Op: 0x15}, Memory: &memoryIndex, Memory64: fixture == "async-stream-memory64", Type: &validate.Type{Value: &model.Type{Kind: model.Stream, Element: &model.Type{Kind: model.U32}}}}
				_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
					for _, ptr := range []uint64{0, 4, 65536, 1, 65540, 1 << 32} {
						_, err := in.guestBuffer(c.core, run.scope, p, ptr, 0)
						if ptr <= 65536 && ptr&3 == 0 {
							if err != nil {
								t.Fatal("valid empty endpoint", ptr, err)
							}
						} else if !errors.Is(err, wasm.ErrTrap) {
							t.Fatal("typed empty buffer bypassed pointer checks", ptr, err)
						}
					}
					p.Type = &validate.Type{Value: &model.Type{Kind: model.Stream}}
					if _, err := in.guestBuffer(c.core, run.scope, p, ^uint64(0), 1<<28-1); err != nil {
						t.Fatal("unit buffer used an irrelevant pointer", err)
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

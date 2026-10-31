package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLowerPlans(t *testing.T) {
	for _, name := range []string{"async-lower", "async-lower64", "async-lower-sync-callee", "async-lower-arity", "async-lower-strings", "async-lower-strings-utf16", "async-lower-strings-compact", "async-lower-strings64", "async-lower-resources", "async-lower-resources64", "async-lower-resource-sync", "async-lower-resource-sync64"} {
		bindingSchema(t, name)
	}
	input, err := os.ReadFile("../../../pkg/component/testdata/async-lower-arity.wasm")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"memory", "callback", "post", "sync-type", "signature", "duplicate"} {
		t.Run(mutation, func(t *testing.T) {
			b, err := decode.Decode(input, decode.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			var lower *model.Canonical
			var typ *model.TypeDef
			for _, d := range b.Definitions {
				if d.Type != nil && d.Type.Kind == model.AsyncFunction && typ == nil {
					typ = d.Type
				}
				if d.Canonical != nil && d.Canonical.Op == 1 && lower == nil {
					lower = d.Canonical
				}
			}
			switch mutation {
			case "memory":
				lower.Options = []model.CanonicalOption{{Kind: 6}}
			case "callback":
				lower.Options = append(lower.Options, model.CanonicalOption{Kind: 7})
			case "post":
				lower.Options = append(lower.Options, model.CanonicalOption{Kind: 5})
			case "sync-type":
				typ.Kind = model.Function
			case "signature":
				lower.Options = lower.Options[:1]
			case "duplicate":
				lower.Options = append(lower.Options, model.CanonicalOption{Kind: 6})
			}
			if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(mutation, err)
			}
		})
	}
}

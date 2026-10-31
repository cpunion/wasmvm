package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestErrorContextCanonicalPlans(t *testing.T) {
	for _, name := range []string{"error-context", "error-context-utf16", "error-context-compact", "error-context64", "error-context-host", "error-context-components", "error-context-cross"} {
		bindingSchema(t, name)
	}
	input, err := os.ReadFile("../../../pkg/component/testdata/error-context.wasm")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"memory", "shared", "duplicate", "async", "callback", "post", "encoding", "realloc", "debug-memory", "debug-realloc"} {
		t.Run(mutation, func(t *testing.T) {
			b, err := decode.Decode(input, decode.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			var newContext, debug *model.Canonical
			for _, d := range b.Definitions {
				if d.Canonical != nil {
					switch d.Canonical.Op {
					case 0x1c:
						newContext = d.Canonical
					case 0x1d:
						debug = d.Canonical
					}
				}
			}
			switch mutation {
			case "memory":
				newContext.Options = nil
			case "shared":
				b.Definitions[0].CoreModule.Memories[0].Shared = true
				b.Definitions[0].CoreModule.Memories[0].Max = 1
				b.Definitions[0].CoreModule.Memories[0].HasMax = true
			case "duplicate":
				newContext.Options = append(newContext.Options, newContext.Options[0])
			case "async":
				newContext.Options = append(newContext.Options, model.CanonicalOption{Kind: 6})
			case "callback":
				newContext.Options = append(newContext.Options, model.CanonicalOption{Kind: 7})
			case "post":
				debug.Options = append(debug.Options, model.CanonicalOption{Kind: 5})
			case "encoding":
				newContext.Options = append(newContext.Options, model.CanonicalOption{Kind: 0}, model.CanonicalOption{Kind: 1})
			case "realloc":
				debug.Options[1].Index = 1
			case "debug-memory":
				debug.Options = []model.CanonicalOption{{Kind: 4}}
			case "debug-realloc":
				debug.Options = debug.Options[:1]
			}
			if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(mutation, err)
			}
		})
	}
}

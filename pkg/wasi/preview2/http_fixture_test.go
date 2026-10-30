package preview2

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

// These consumers exercise component type/canonical compilation. They do not
// establish HTTP provider execution or full-world support.
func TestHTTPFixtureTypeValidation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			engine, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			names := []string{"http-fields", "http-fields-api"}
			for _, name := range []string{"tuple-order", "value-width", "header-error", "clone-result", "lookup-result", "request-option", "request-path", "http-error"} {
				names = append(names, "http-contract-"+name)
			}
			for _, name := range names {
				module, err := engine.Compile(nil, fixture(t, name))
				if err != nil {
					t.Fatal(name, err)
				}
				module.Close()
			}
		})
	}
}

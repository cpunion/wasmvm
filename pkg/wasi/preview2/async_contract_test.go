//go:build linux || darwin

package preview2

import (
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestStableWorldRejectsAsyncExportTypes(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"async-cli", "async-http"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				w, err := NewWorld(WorldConfig{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { w.Close(); _ = w.WaitClosed(nil) }()
				cfg := component.DefaultConfig()
				cfg.Core.Backend = backend
				e, err := component.NewEngine(cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer e.Close()
				bytes, err := os.ReadFile("../../component/testdata/" + name + ".wasm")
				if err != nil {
					t.Fatal(err)
				}
				m, err := e.Compile(nil, bytes)
				if err != nil {
					t.Fatal(err)
				}
				defer m.Close()
				in, err := m.Instantiate(nil, w.Imports(), component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				if name == "async-cli" {
					_, err = RunCommand(nil, in, component.CallOptions{})
				} else {
					_, err = w.HTTP().CallIncoming(nil, in, httptest.NewRequest("GET", "http://localhost/", nil), component.CallOptions{})
				}
				if !errors.Is(err, vm.ErrType) {
					t.Fatal("stable synchronous world accepted async export", err)
				}
			})
		}
	}
}

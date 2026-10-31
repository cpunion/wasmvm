package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestErrorContextFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/error-context.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string `json:"componentModelCommit"`
		Producer  string `json:"producer"`
		Generator string `json:"generatorSHA256"`
		Helper    string `json:"helperSHA256"`
		Fixtures  map[string]struct {
			WAT       string `json:"wat"`
			Wasm      string `json:"wasm"`
			Validated bool   `json:"producerValidated"`
			Rejection string `json:"producerRejection"`
		} `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 9 {
		t.Fatal("wrong fixture profile", pins)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("fixture changed", path)
		}
	}
	check("../../scripts/build_error_context_fixtures.py", pins.Generator)
	check("../../scripts/build_resource_value_fixtures.py", pins.Helper)
	for name, fixture := range pins.Fixtures {
		check(filepath.Join("testdata", name+".wasm"), fixture.Wasm)
		if fixture.WAT != "" {
			check(filepath.Join("testdata", name+".wat"), fixture.WAT)
		}
		if name == "error-context64" || name == "error-context-values" {
			if fixture.Validated || fixture.Rejection == "" {
				t.Fatal("lost producer boundary", name)
			}
		} else if !fixture.Validated {
			t.Fatal("missing independent validation", name)
		}
	}
}

func TestErrorContextInvalidHandlesAndOutputRange(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, fixtureName := range []string{"error-context", "error-context64"} {
			for _, mode := range []string{"new-pointer", "drop", "message-handle", "message-pointer"} {
				cfg := DefaultConfig()
				cfg.Core.Backend = backend
				e, _ := NewEngine(cfg)
				m, err := e.Compile(nil, fixture(t, fixtureName))
				if err != nil {
					t.Fatal(err)
				}
				in, err := m.Instantiate(nil, Imports{}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				bad := uint64(0xffffffff)
				if fixtureName == "error-context64" {
					bad = 1 << 32
				}
				var f *Function
				var args []Value
				switch mode {
				case "new-pointer":
					f, _ = in.Function("raw-new")
					args = []Value{{Bits: bad}, {Bits: 1}}
				case "drop":
					f, _ = in.Function("raw-drop")
					args = []Value{{}}
				case "message-handle":
					f, _ = in.Function("raw-message")
					args = []Value{{}, {Bits: 32}}
				case "message-pointer":
					makeRaw, _ := in.Function("raw-new")
					v, err := makeRaw.Call(nil, []Value{{}, {Bits: 1}}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					f, _ = in.Function("raw-message")
					args = []Value{*v.Value, {Bits: bad}}
				}
				if _, err := f.Call(nil, args, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(backend, fixtureName, mode, err)
				}
				if mode == "message-pointer" {
					// Inspect the private Core counter: component entry is poisoned
					// after the trap, but an invalid output must precede realloc.
					n := f.scope
					raw := n.items[model.CoreFunctionSort][1].core
					_, err := in.store.WithCaller(nil, CallOptions{}, func(c *vm.Caller) error {
						v, err := c.Call(raw, nil)
						if err != nil {
							return err
						}
						bits, _, err := v.Values[0].Bits()
						if err == nil && bits != 0 {
							t.Error("invalid result pointer called realloc", bits)
						}
						return err
					})
					if err != nil {
						t.Fatal(err)
					}
				}
				in.Close()
				m.Close()
				e.Close()
			}
		}
	}
}

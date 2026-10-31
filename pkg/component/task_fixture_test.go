package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestAsyncLiftFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/async-lift.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string                       `json:"componentModelCommit"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Resource  string                       `json:"resourceSourceSHA256"`
		Fixtures  map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 14 {
		t.Fatal("wrong provenance")
	}
	check := func(path, want string) {
		t.Helper()
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(bytes)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("input changed", path)
		}
	}
	check("../../scripts/build_async_lift_fixtures.py", pins.Generator)
	check("testdata/ownership.wat", pins.Resource)
	for name, hashes := range pins.Fixtures {
		for extension, want := range hashes {
			check(filepath.Join("testdata", name+"."+extension), want)
		}
	}
}

func TestStackfulTaskReentryRestoresContext(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		initial := asyncInstance(t, backend, "async-init", Imports{})
		entry, _ := initial.Function("run")
		if v, err := entry.Call(nil, []Value{{Bits: 3}}, CallOptions{}); err != nil || v.Value.Bits != 42 {
			t.Fatal("initialization TLS escaped", v, err)
		}
		tls := asyncInstance(t, backend, "async-tls", Imports{Functions: map[string]HostFunction{"text": func(*Caller, []Value) (*Value, error) { return &Value{Text: "tls λ"}, nil }}})
		entry, _ = tls.Function("run")
		for range 5 {
			if v, err := entry.Call(nil, nil, CallOptions{}); err != nil || v.Value.Text != "tls λ" {
				t.Fatal("realloc TLS not isolated", v, err)
			}
		}
		var f *Function
		depth := 0
		var saved *Caller
		in := asyncInstance(t, backend, "async-reentry", Imports{Functions: map[string]HostFunction{"hook": func(c *Caller, _ []Value) (*Value, error) {
			saved = c
			if depth == 0 {
				depth++
				v, err := c.Call(f, []Value{{Bits: 3}})
				depth--
				if err != nil || v.Bits != 45 {
					t.Fatal(v, err)
				}
			}
			return nil, nil
		}}})
		f, _ = in.Function("stackful")
		for range 5 {
			v, err := f.Call(nil, []Value{{Bits: 7}}, CallOptions{})
			if err != nil || v.Value.Bits != 49 {
				t.Fatal("parent TLS not restored", v, err)
			}
			if _, err := saved.Call(f, []Value{{}}); err == nil {
				t.Fatal("expired host caller escaped")
			}
		}
	}
}

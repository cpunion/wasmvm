package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAsyncLowerFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/async-lower.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string `json:"componentModelCommit"`
		Producer  string `json:"producer"`
		Generator string `json:"generatorSHA256"`
		Memory    string `json:"memorySourceSHA256"`
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
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 12 {
		t.Fatal("fixture provenance changed")
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
	check("../../scripts/build_async_lower_fixtures.py", pins.Generator)
	check("testdata/async-values.wat", pins.Memory)
	for name, fixture := range pins.Fixtures {
		check(filepath.Join("testdata", name+".wat"), fixture.WAT)
		check(filepath.Join("testdata", name+".wasm"), fixture.Wasm)
		boundary := name == "async-lower64" || name == "async-lower-strings64" || name == "async-lower-resources64"
		if fixture.Validated == boundary || boundary && fixture.Rejection == "" {
			t.Fatal("wrong producer outcome", name)
		}
	}
}

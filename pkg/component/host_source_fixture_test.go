package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestHostSourceFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/host-sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string                       `json:"componentModelCommit"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Fixtures  map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 8 {
		t.Fatal("provenance changed")
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed fixture", path)
		}
	}
	check("../../scripts/build_host_source_fixtures.py", pins.Generator)
	for name, hashes := range pins.Fixtures {
		for ext, want := range hashes {
			check("testdata/"+name+"."+ext, want)
		}
	}
}

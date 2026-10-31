package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestPendingFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/pending-subtasks.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string                       `json:"componentModelCommit"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Sources   map[string]string            `json:"sources"`
		Fixtures  map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 11 || len(pins.Sources) != 2 {
		t.Fatal("provenance changed")
	}
	check := func(name, want string) {
		t.Helper()
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed fixture", name)
		}
	}
	check("../../scripts/build_pending_subtask_fixtures.py", pins.Generator)
	for name, want := range pins.Sources {
		check("testdata/"+name, want)
	}
	for name, hashes := range pins.Fixtures {
		for ext, want := range hashes {
			check("testdata/"+name+"."+ext, want)
		}
	}
}

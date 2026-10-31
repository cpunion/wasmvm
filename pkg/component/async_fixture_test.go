package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestFutureStreamFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/future-streams.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit     string                       `json:"componentModelCommit"`
		Producer   string                       `json:"producer"`
		Generator  string                       `json:"generatorSHA256"`
		Forward    []string                     `json:"authoredForward"`
		Rejections map[string]string            `json:"producerRejections"`
		Fixtures   map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 16 || len(pins.Forward) != 2 || len(pins.Rejections) != 2 {
		t.Fatal("provenance changed")
	}
	check := func(path, want string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed input", path)
		}
	}
	check("../../scripts/build_future_stream_fixtures.py", pins.Generator)
	for name, hashes := range pins.Fixtures {
		for ext, want := range hashes {
			check("testdata/"+name+"."+ext, want)
		}
	}
	for name, reason := range pins.Rejections {
		if !strings.HasSuffix(name, "-memory64") || !strings.Contains(reason, "type mismatch for export `read`") {
			t.Fatal("producer rejection lost", name)
		}
	}
}

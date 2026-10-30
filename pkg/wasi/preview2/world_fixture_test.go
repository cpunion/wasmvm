package preview2

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorldProducerManifest(t *testing.T) {
	data, err := os.ReadFile("testdata/world-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		WASIVersion string            `json:"wasiVersion"`
		Sources     map[string]string `json:"sourcesSha256"`
		Binaries    map[string]string `json:"binarySha256"`
		Unstable    []string          `json:"unstableFeatures"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.WASIVersion != "0.2.12" || len(manifest.Unstable) != 0 || len(manifest.Binaries) != 11 {
		t.Fatal("wrong stable fixture profile", manifest)
	}
	for _, group := range []struct {
		root  string
		files map[string]string
	}{{"testdata", manifest.Binaries}, {"../../../testdata/wasi-preview2/guests", manifest.Sources}} {
		for name, expected := range group.files {
			data, err := os.ReadFile(filepath.Join(group.root, name))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) != expected {
				t.Fatal("producer input/output checksum changed", name)
			}
		}
	}
}

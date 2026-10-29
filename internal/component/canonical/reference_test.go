package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPinnedReferenceLayouts(t *testing.T) {
	root := "../../../testdata/component-canonical"
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision string
		Cases    int
		SHA256   map[string]string
	}
	if err := json.Unmarshal(b, &manifest); err != nil || manifest.Revision != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || manifest.Cases != 1024 {
		t.Fatal("reference manifest", err)
	}
	for file, want := range manifest.SHA256 {
		b, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("reference input changed", file)
		}
	}
	b, err = os.ReadFile(filepath.Join(root, "layout-oracle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Type        *model.Type
		Memory64    bool
		Size, Align uint64
		Flat        []string
	}
	if err := json.Unmarshal(b, &cases); err != nil || len(cases) != manifest.Cases {
		t.Fatal("reference case inventory", err)
	}
	flatTypes := map[string]wasm.ValueType{"i32": wasm.I32, "i64": wasm.I64, "f32": wasm.F32, "f64": wasm.F64}
	for index, row := range cases {
		got, err := NewBuilder(row.Memory64, DefaultLimits()).Build(row.Type)
		if err != nil || got.Size != row.Size || got.Align != row.Align || got.FlatCount != uint64(len(row.Flat)) {
			t.Fatal(index, got, row.Size, row.Align, row.Flat, err)
		}
		if got.FlatCount <= MaxFlatParams {
			for i, tpe := range row.Flat {
				if got.Flat[i] != flatTypes[tpe] {
					t.Fatal("flattened type differs from reference", index, i, row.Flat, got.Flat)
				}
			}
		} else if got.Flat != nil {
			t.Fatal("unbounded flattened vector", index)
		}
	}
}

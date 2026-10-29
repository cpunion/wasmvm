package execute

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExtendedConstProposal(t *testing.T) {
	root := "../../../testdata/extended-const"
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision string
		SHA256   map[string]string
	}
	if err := json.Unmarshal(raw, &manifest); err != nil || manifest.Revision != "8d4f6aa2b00a8e7c0174410028625c6a176db8a1" || len(manifest.SHA256) != 198 {
		t.Fatal("proposal inventory changed", err)
	}
	for name, want := range manifest.SHA256 {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
			t.Fatalf("proposal checksum: %s: %v", name, err)
		}
	}
	tiers := []int{0}
	if native.Available {
		tiers = append(tiers, 1, 2)
	}
	for _, tier := range tiers {
		t.Run(fmt.Sprint(tier), func(t *testing.T) {
			commands, assertions, text := 0, 0, 0
			for _, name := range []string{"global", "data", "elem"} {
				t.Run(name, func(t *testing.T) {
					path := filepath.Join(root, name, "script.json")
					raw, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var script struct{ Commands []specCommand }
					if err := json.Unmarshal(raw, &script); err != nil {
						t.Fatal(err)
					}
					r := newSpec(t, context.Background(), tier)
					r.features = wasm.ExtendedConst
					for _, cmd := range script.Commands {
						commands++
						if cmd.Type == "assert_malformed" && cmd.ModuleType == "text" {
							text++
							continue
						}
						if strings.HasPrefix(cmd.Type, "assert_") {
							assertions++
						}
						r.runCommand(t, cmd, path)
					}
				})
			}
			if commands != 290 || assertions != 215 || text != 3 {
				t.Fatalf("commands=%d binary assertions=%d text assertions=%d", commands, assertions, text)
			}
			t.Logf("Executed 215 binary assertions in 3 complete proposal scripts; 3 text assertions classified separately")
		})
	}
}

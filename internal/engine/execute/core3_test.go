package execute

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompleteCore3Spec(t *testing.T) {
	root := "../../../artifacts/core3/fixtures"
	manifestBytes, err := os.ReadFile("../../../testdata/core3/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision                                                          string
		ScriptCount, CommandCount, BinaryAssertions, TextParserAssertions int
		Scripts                                                           []struct{ Name string }
		SHA256                                                            map[string]string
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Revision != "aa4dcf44b8de3f3f3270ca067422172518ecc341" || manifest.ScriptCount != 259 || len(manifest.Scripts) != 259 || manifest.CommandCount != 65344 || manifest.BinaryAssertions != 61395 || manifest.TextParserAssertions != 1239 {
		t.Fatal("Core 3.0 inventory changed")
	}
	for name, want := range manifest.SHA256 {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
			t.Fatalf("prepare pinned Core 3.0 inputs with python3 scripts/core3_corpus.py: %s: %v", name, err)
		}
	}
	tiers := []int{0}
	if native.Available {
		tiers = append(tiers, 1, 2)
	}
	for _, tier := range tiers {
		t.Run(fmt.Sprint(tier), func(t *testing.T) {
			reportPath := ""
			if base := os.Getenv("WASMVM_CONFORMANCE_REPORT"); base != "" {
				reportPath = filepath.Join(filepath.Dir(base), fmt.Sprintf("core3-%d.json", tier))
				if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			commands, assertions, text := 0, 0, 0
			for _, s := range manifest.Scripts {
				t.Run(s.Name, func(t *testing.T) {
					path := filepath.Join(root, s.Name, "script.json")
					raw, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var script struct{ Commands []specCommand }
					if err := json.Unmarshal(raw, &script); err != nil {
						t.Fatal(err)
					}
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					r := newSpec(t, ctx, tier)
					r.features = wasm.Core3Features
					for _, cmd := range script.Commands {
						commands++
						if cmd.ModuleType == "text" {
							if cmd.Type != "assert_malformed" && cmd.Type != "assert_invalid" {
								t.Fatal("unclassified executable text", cmd.Type)
							}
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
			if commands != 65344 || assertions != 61395 || text != 1239 {
				t.Fatalf("commands=%d binary=%d text=%d", commands, assertions, text)
			}
			t.Logf("Executed %d binary assertions in 259 complete Core 3.0 scripts; %d text-parser assertions classified separately", assertions, text)
			if reportPath != "" && !t.Failed() {
				report := struct {
					Passed                                                                    bool
					Profile, Revision, ManifestSHA256, Compiler, RuntimeVersion, GOOS, GOARCH string
					Features                                                                  wasm.Features
					Tier, Scripts, Commands, BinaryAssertions, TextParserAssertions           int
				}{true, "core3", manifest.Revision, fmt.Sprintf("%x", sha256.Sum256(manifestBytes)), runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH, wasm.Core3Features, tier, 259, commands, assertions, text}
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, append(data, '\n'), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

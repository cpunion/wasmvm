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

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type proposalCorpus struct {
	name, revision                    string
	features                          wasm.Features
	scripts                           []string
	files, commands, assertions, text int
}

func TestExtendedConstProposal(t *testing.T) {
	runProposal(t, proposalCorpus{"extended-const", "8d4f6aa2b00a8e7c0174410028625c6a176db8a1", wasm.ExtendedConst, []string{"global", "data", "elem"}, 198, 290, 215, 3})
}
func TestTailCallProposal(t *testing.T) {
	runProposal(t, proposalCorpus{"tail-call", "a6003d06aefef41e20a3e36fe2e500062555c895", wasm.TailCall, []string{"return_call", "return_call_indirect"}, 46, 119, 102, 11})
}

func runProposal(t *testing.T, corpus proposalCorpus) {
	root := filepath.Join("../../../testdata", corpus.name)
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision string
		SHA256   map[string]string
	}
	if err := json.Unmarshal(raw, &manifest); err != nil || manifest.Revision != corpus.revision || len(manifest.SHA256) != corpus.files {
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
			reportPath := ""
			if base := os.Getenv("WASMVM_CONFORMANCE_REPORT"); base != "" {
				reportPath = filepath.Join(filepath.Dir(base), "proposals", fmt.Sprintf("%s-%d.json", corpus.name, tier))
				if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			commands, assertions, text := 0, 0, 0
			for _, name := range corpus.scripts {
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
					r.features = corpus.features
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
			if commands != corpus.commands || assertions != corpus.assertions || text != corpus.text {
				t.Fatalf("commands=%d binary assertions=%d text assertions=%d", commands, assertions, text)
			}
			t.Logf("Executed %d binary assertions in %d complete proposal scripts; %d text assertions classified separately", assertions, len(corpus.scripts), text)
			if reportPath != "" && !t.Failed() {
				report := struct {
					Passed                                                                     bool
					Proposal, Revision, ManifestSHA256, Compiler, RuntimeVersion, GOOS, GOARCH string
					Features                                                                   wasm.Features
					Tier, Scripts, BinaryAssertions, TextParserAssertions, Commands            int
				}{true, corpus.name, corpus.revision, fmt.Sprintf("%x", sha256.Sum256(raw)), runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH, corpus.features, tier, len(corpus.scripts), assertions, text, commands}
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

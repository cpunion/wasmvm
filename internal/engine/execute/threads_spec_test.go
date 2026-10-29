package execute

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type specThread struct {
	runner *specRunner
	done   chan struct{}
}

func (r *specRunner) close() {
	if r.closed {
		return
	}
	r.closed = true
	r.cancel()
	for _, child := range r.children {
		<-child.done
		child.runner.close()
	}
	r.store.Close()
	for _, lease := range r.modules {
		lease.Close()
	}
}
func (r *specRunner) startThread(t *testing.T, cmd specCommand, path string) {
	child := newSpecOwned(t, r.context, r.tier, false)
	child.progress = r.progress
	child.features = r.features
	child.store.Domain = r.store.Domain
	if cmd.SharedModule != "" {
		original := r.instance(cmd.SharedModule)
		if original == nil {
			t.Fatal("unknown shared module", cmd.SharedModule)
		}
		snapshot := &state.Instance{Owner: child.store, Exports: make(map[string]state.Extern)}
		for name, ex := range original.Exports {
			if ex.Kind != 2 || !ex.Memory.Shared() {
				t.Fatal("unsupported shared script capability", name)
			}
			imported, err := child.store.ImportShared(ex.Memory.Memory)
			if err != nil {
				t.Fatal(err)
			}
			snapshot.Exports[name] = imported
		}
		child.named[cmd.SharedModule] = snapshot
	}
	th := &specThread{runner: child, done: make(chan struct{})}
	r.children = append(r.children, th)
	r.threads[cmd.Name] = th
	go func() {
		defer close(th.done)
		for _, c := range cmd.Commands {
			if c.ModuleType == "text" {
				if child.progress != nil {
					child.progress.record(c)
				}
				continue
			}
			child.runCommand(t, c, path)
		}
	}()
}
func (r *specRunner) waitThread(t *testing.T, name string) {
	th := r.threads[name]
	if th == nil {
		t.Fatal("unknown script thread", name)
	}
	select {
	case <-th.done:
	case <-r.context.Done():
		t.Fatal(r.context.Err())
	}
}

func TestCoreThreadsConformance(t *testing.T) {
	root := "../../../testdata/threads"
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision string
		Scripts  map[string]struct{ Commands, BinaryAssertions, TextAssertions int }
		SHA256   map[string]string
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Revision != "cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0" || len(manifest.Scripts) != 17 || len(manifest.SHA256) != 495 {
		t.Fatal("Threads inventory changed")
	}
	for name, want := range manifest.SHA256 {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != want {
			t.Fatal(name, err)
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
				reportPath = filepath.Join(filepath.Dir(base), "proposals", fmt.Sprintf("threads-%d.json", tier))
				if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			names := make([]string, 0, len(manifest.Scripts))
			for name := range manifest.Scripts {
				names = append(names, name)
			}
			sort.Strings(names)
			total, binary, text := 0, 0, 0
			for _, name := range names {
				t.Run(name, func(t *testing.T) {
					path := filepath.Join(root, name, "script.json")
					b, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var script struct{ Commands []specCommand }
					if err := json.Unmarshal(b, &script); err != nil {
						t.Fatal(err)
					}
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					r := newSpec(t, ctx, tier)
					r.features = wasm.Threads
					r.progress = &specProgress{}
					defer func() { r.close(); a, b, c, _ := r.progress.snapshot(); total += a; binary += b; text += c }()
					for _, c := range script.Commands {
						if c.ModuleType == "text" {
							r.progress.record(c)
							continue
						}
						r.runCommand(t, c, path)
					}
				})
			}
			if total != 946 || binary != 579 || text != 6 {
				t.Fatal(total, binary, text)
			}
			t.Logf("Executed all %d binary assertions and %d commands, including nested guest threads; %d parser assertions separate", binary, total, text)
			if reportPath != "" && !t.Failed() {
				report := struct {
					Passed                                                                     bool
					Proposal, Revision, ManifestSHA256, Compiler, RuntimeVersion, GOOS, GOARCH string
					Features                                                                   wasm.Features
					Tier, Scripts, Commands, BinaryAssertions, TextParserAssertions            int
				}{true, "threads", manifest.Revision, fmt.Sprintf("%x", sha256.Sum256(raw)), runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH, wasm.Threads, tier, 17, total, binary, text}
				b, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, append(b, '\n'), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// A shared progress counter records actual parent/child execution, rather than
// counting nested commands that might never run after a failed thread start.
type specProgress struct {
	mu                                  sync.Mutex
	commands, assertions, text, threads int
}

func (p *specProgress) record(c specCommand) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.commands++
	if c.ModuleType == "text" {
		p.text++
	} else if strings.HasPrefix(c.Type, "assert_") {
		p.assertions++
	}
	if c.Type == "thread" {
		p.threads++
	}
}
func (p *specProgress) snapshot() (int, int, int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.commands, p.assertions, p.text, p.threads
}

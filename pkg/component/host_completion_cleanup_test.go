package component

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCompletionExternalCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"success", "resolver-error", "resolver-panic", "cleanup-panic", "both-panic", "cancel"} {
			closed := atomic.Int32{}
			var completion *HostCompletion
			in := asyncInstance(t, backend, "async-host-completion", Imports{Functions: map[string]HostFunction{"notify": func(*Caller, []Value) (*Value, error) { return nil, nil }}, AsyncCompletions: map[string]AsyncHostFunction{"add": func(c *Caller, _ []Value) (*HostCompletion, error) {
				completion = NewHostCompletionWithCleanup(func() {
					closed.Add(1)
					if err := completion.Resolve(func(*Caller) (*Value, error) { return nil, nil }); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal("cleanup held completion mutex or completion stayed writable", err)
					}
					if mode == "cleanup-panic" || mode == "both-panic" {
						panic("cleanup")
					}
				})
				if mode != "cancel" {
					completion.Resolve(func(*Caller) (*Value, error) {
						if mode == "resolver-error" {
							return nil, wasm.ErrTrap
						}
						if mode == "resolver-panic" || mode == "both-panic" {
							panic("resolver")
						}
						return &Value{Bits: 42}, nil
					})
				}
				return completion, nil
			}}})
			f, _ := in.Function("direct")
			func() {
				defer func() {
					p := recover()
					if (mode == "resolver-panic" || mode == "both-panic") && p != "resolver" || mode == "cleanup-panic" && p != "cleanup" {
						t.Fatal("wrong cleanup panic", mode, p)
					}
					if p != nil && mode != "resolver-panic" && mode != "cleanup-panic" && mode != "both-panic" {
						panic(p)
					}
				}()
				ctx := context.Background()
				cancel := func() {}
				if mode == "cancel" {
					ctx, cancel = context.WithTimeout(ctx, 20*time.Millisecond)
				}
				defer cancel()
				_, err := f.Call(ctx, []Value{{Bits: 35}}, CallOptions{})
				if mode == "success" && err != nil || mode != "success" && err == nil {
					t.Fatal(mode, err)
				}
			}()
			if closed.Load() != 1 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal(mode, "retained completion", closed.Load())
			}
			completion.close()
			if closed.Load() != 1 {
				t.Fatal("cleanup repeated")
			}
		}
	}
}

func TestHostCompletionUnpublishedFileLease(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"success", "abandon", "cleanup-panic"} {
			abandon := mode == "abandon"
			directory := t.TempDir()
			var mu sync.Mutex
			var file *os.File
			refs := 0
			closed := false
			release := func() {
				mu.Lock()
				defer mu.Unlock()
				refs--
				if refs == 0 && file != nil {
					file.Close()
				}
			}
			typ := NewHostResourceType(func(*Caller, uint32) error { release(); return nil })
			worker := make(chan error, 1)
			allow := make(chan struct{})
			imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, AsyncFunctions: map[string]HostFunction{
				"pass":   func(_ *Caller, args []Value) (*Value, error) { return &args[0], nil },
				"borrow": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
				"blob":   func(_ *Caller, args []Value) (*Value, error) { return &args[0], nil },
			}, AsyncCompletions: map[string]AsyncHostFunction{"make": func(*Caller, []Value) (*HostCompletion, error) {
				completion := NewHostCompletionWithCleanup(func() {
					mu.Lock()
					closed = true
					mu.Unlock()
					release()
					if mode == "cleanup-panic" {
						panic("file cleanup")
					}
				})
				mu.Lock()
				refs = 1
				mu.Unlock()
				go func() {
					<-allow
					opened, err := os.CreateTemp(directory, "completion-")
					if err != nil {
						worker <- err
						return
					}
					mu.Lock()
					file = opened
					wasClosed := closed
					if wasClosed {
						opened.Close()
					}
					mu.Unlock()
					err = completion.Resolve(func(c *Caller) (*Value, error) {
						mu.Lock()
						refs++
						mu.Unlock()
						v, err := c.NewResource(typ, 1)
						if err != nil {
							release()
							return nil, err
						}
						return &v, nil
					})
					if wasClosed && !errors.Is(err, wasm.ErrClosed) {
						worker <- fmt.Errorf("late publication: %v", err)
						return
					}
					worker <- nil
				}()
				return completion, nil
			}}}
			in := asyncInstance(t, backend, "async-host-resources", imports)
			makeResource, _ := in.Function("make")
			ctx := context.Background()
			cancel := func() {}
			if abandon {
				ctx, cancel = context.WithTimeout(ctx, 20*time.Millisecond)
			} else {
				close(allow)
			}
			var value CallResult
			var err error
			var panicked any
			func() {
				defer func() { panicked = recover() }()
				value, err = makeResource.Call(ctx, nil, CallOptions{})
			}()
			if mode == "cleanup-panic" && panicked != "file cleanup" || mode != "cleanup-panic" && panicked != nil {
				t.Fatal(mode, "unexpected panic", panicked)
			}
			cancel()
			if abandon {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
				close(allow)
			} else if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
				mu.Lock()
				_, err = file.Stat()
				held := refs
				mu.Unlock()
				if err != nil || held != 1 {
					t.Fatal("published result lost its resource lease", held, err)
				}
				if _, err := in.DropResource(nil, value.Value.Resource, CallOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case err := <-worker:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("external result worker stranded")
			}
			mu.Lock()
			_, err = file.Stat()
			held := refs
			mu.Unlock()
			if !errors.Is(err, os.ErrClosed) || held != 0 {
				t.Fatal("unpublished/published result leaked", abandon, held, err)
			}
		}
	}
}

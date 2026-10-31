package component

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSyncReentryIgnoresAsyncBackpressureAndIsolatesTasks(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			var outer, inner, makeResource, dropLocal, dropped *Function
			calls := 0
			in := asyncInstance(t, backend, "sync-reentry", Imports{Functions: map[string]HostFunction{"hook": func(c *Caller, _ []Value) (*Value, error) {
				calls++
				n := outer.scope
				task := n.task
				exclusive := n.exclusive
				if !n.entered || task == nil || task.runtime == nil || exclusive != task.runtime || n.backpressure != 1 {
					t.Fatal("missing parent task/exclusive state")
				}
				check := func() {
					t.Helper()
					if n.task != task || !n.entered || n.exclusive != exclusive || n.backpressure != 1 || n.task.storage[0] != 123 {
						t.Fatal("nested sync operation changed async task state")
					}
				}
				for range 2 {
					v, err := c.Call(inner, nil)
					if err != nil || v.Bits != 42 {
						t.Fatal("sync reentry under explicit/implicit backpressure", v, err)
					}
					check()
				}
				v, err := c.Call(makeResource, []Value{{Bits: 9}})
				if err != nil {
					return nil, err
				}
				check()
				if err := c.DropResource(v.Resource); err != nil {
					return nil, err
				}
				check()
				if _, err := c.Call(dropLocal, nil); err != nil {
					return nil, err
				}
				check()
				return &Value{Bits: 42}, nil
			}}})
			outer, _ = in.Function("outer")
			inner, _ = in.Function("inner")
			makeResource, _ = in.Function("make")
			dropLocal, _ = in.Function("drop-local")
			dropped, _ = in.Function("dropped")
			for i := 1; i <= 3; i++ {
				v, err := outer.Call(nil, nil, CallOptions{Metered: true, Fuel: 10000})
				if err != nil || v.Value.Bits != 42 {
					t.Fatal(v, err)
				}
				if n := outer.scope; n.entered || n.task != nil || n.exclusive != nil || n.backpressure != 0 {
					t.Fatal("outer task state was not restored")
				}
				v, err = dropped.Call(nil, nil, CallOptions{})
				if err != nil || v.Value.Bits != uint64(i*14) {
					t.Fatal("destructor result", v, err)
				}
			}
			if calls != 3 {
				t.Fatal(calls)
			}
		})
	}
}
func TestSyncReentryFailureRestoresOuterTask(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"trap", "sync-wait", "panic", "cancel"} {
			var outer, target *Function
			ctx, cancel := context.WithCancel(context.Background())
			in := asyncInstance(t, backend, "sync-reentry", Imports{Functions: map[string]HostFunction{"hook": func(c *Caller, _ []Value) (*Value, error) {
				task := outer.scope.task
				if mode == "panic" {
					panic("host")
				}
				if mode == "cancel" {
					cancel()
				}
				_, err := c.Call(target, nil)
				if outer.scope.task != task || !outer.scope.entered {
					t.Fatal(mode, "nested error lost outer task")
				}
				if err == nil {
					t.Fatal(mode, "nested failure accepted")
				}
				if mode == "sync-wait" && !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("sync task inherited async blocking", err)
				}
				return nil, err
			}}})
			outer, _ = in.Function("outer")
			name := "fail"
			if mode == "sync-wait" {
				name = "sync-wait"
			}
			target, _ = in.Function(name)
			func() {
				defer func() {
					if p := recover(); p != nil && mode != "panic" {
						panic(p)
					} else if mode == "panic" && p != "host" {
						t.Fatal("original panic", p)
					}
				}()
				if _, err := outer.Call(ctx, nil, CallOptions{}); err == nil {
					t.Fatal(mode, "missing failure")
				}
			}()
			cancel()
			if n := outer.scope; n.task != nil || n.entered || n.exclusive != nil || in.tasks != nil {
				t.Fatal(mode, "failed outer retained task state")
			}
		}
	}
}
func TestSyncReentryFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/sync-reentry.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string                       `json:"componentModelCommit"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Files     map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Files) != 3 {
		t.Fatal("changed pins")
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed input", path)
		}
	}
	check("../../scripts/build_sync_reentry_fixtures.py", pins.Generator)
	for name, files := range pins.Files {
		for ext, hash := range files {
			check("testdata/"+name+"."+ext, hash)
		}
	}
}

func TestSyncReentryComposedGuest(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"sync-reentry-composed", "sync-reentry-destructor"} {
			in := asyncInstance(t, backend, name, Imports{})
			run, _ := in.Function("run")
			for i := 1; i <= 3; i++ {
				v, err := run.Call(nil, nil, CallOptions{})
				if err != nil || v.Value.Bits != uint64(i*5) {
					t.Fatal(backend, v, err)
				}
			}
		}
	}
}

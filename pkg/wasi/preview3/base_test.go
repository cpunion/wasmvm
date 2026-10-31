package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func compileFixture(t *testing.T, backend, name string) *component.Compiled {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	e, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	data, err := os.ReadFile("testdata/" + name + ".wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, data)
	if err != nil {
		t.Fatal(name, err)
	}
	t.Cleanup(m.Close)
	return m
}
func instantiate(t *testing.T, m *component.Compiled, imports component.Imports) *component.Instance {
	t.Helper()
	in, err := m.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func call(t *testing.T, in *component.Instance, name string, args ...component.Value) *component.Value {
	t.Helper()
	f, err := in.Function(name)
	if err != nil {
		t.Fatal(err)
	}
	v, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 10000})
	if err != nil {
		t.Fatal(name, err)
	}
	return v.Value
}
func backends(t *testing.T, fn func(*testing.T, string)) {
	t.Helper()
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) { fn(t, backend) })
	}
}
func TestBaseProvidersExecuteVersionedOperations(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		cwd := "/sandbox"
		cfg := DefaultEnvironmentConfig()
		cfg.Args = []string{"guest", "界"}
		cfg.Env = []string{"A=1=2", "A=again"}
		cfg.InitialCWD = &cwd
		env, err := NewEnvironment(cfg)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Args[0] = "changed"
		cwd = "changed"
		in := instantiate(t, compileFixture(t, backend, "environment"), env.Imports())
		if v := call(t, in, "arguments"); len(v.Elements) != 2 || v.Elements[0].Text != "guest" || v.Elements[1].Text != "界" {
			t.Fatal(v)
		}
		if v := call(t, in, "variables"); len(v.Elements) != 2 || v.Elements[0].Elements[1].Text != "1=2" {
			t.Fatal(v)
		}
		if v := call(t, in, "cwd"); v.Case != 1 || v.Elements[0].Text != "/sandbox" {
			t.Fatal(v)
		}
		if call(t, in, "argument-count").Bits != 2 || call(t, in, "environment-count").Bits != 2 {
			t.Fatal("guest list lowering")
		}
		empty, _ := NewEnvironment(EnvironmentConfig{})
		other := instantiate(t, compileFixture(t, backend, "environment"), empty.Imports())
		if call(t, other, "cwd").Case != 0 {
			t.Fatal("ambient cwd")
		}
		random, err := NewRandom(RandomConfig{MaxBytes: 8})
		if err != nil {
			t.Fatal(err)
		}
		in = instantiate(t, compileFixture(t, backend, "random"), random.Imports())
		for _, name := range []string{"secure-bytes", "insecure-bytes"} {
			if v := call(t, in, name, component.Value{}); len(v.Text) != 0 {
				t.Fatal(v)
			}
			for _, n := range []uint64{1, 8, ^uint64(0)} {
				v := call(t, in, name, component.Value{Bits: n})
				if uint64(len(v.Text)) != min(n, 8) {
					t.Fatal(name, n, v)
				}
			}
		}
		call(t, in, "secure-u64")
		call(t, in, "insecure-u64")
		seed := *call(t, in, "seed")
		again := call(t, in, "seed")
		if len(seed.Elements) != 2 || seed.Elements[0].Bits != again.Elements[0].Bits || seed.Elements[1].Bits != again.Elements[1].Bits {
			t.Fatal("unstable seed")
		}
		for _, name := range []string{"success", "failure", "with-code"} {
			in = instantiate(t, compileFixture(t, backend, "exit"), ExitImports())
			f, err := in.Function(name)
			if err != nil {
				t.Fatal(err)
			}
			var args []component.Value
			code := uint32(0)
			if name == "failure" {
				code = 1
			}
			if name == "with-code" {
				args = []component.Value{{Bits: 255}}
				code = 255
			}
			_, err = f.Call(nil, args, component.CallOptions{})
			var exit *ExitError
			if !errors.As(err, &exit) || exit.Code != code {
				t.Fatal(name, err)
			}
		}
	})
}
func TestBaseProviderValidation(t *testing.T) {
	if _, err := NewEnvironment(EnvironmentConfig{Env: []string{"invalid"}}); err == nil {
		t.Fatal("bad env")
	}
	if _, err := NewRandom(RandomConfig{MaxBytes: 1}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name    string
		imports component.Imports
	}{
		{"wrong-arguments", func() component.Imports { p, _ := NewEnvironment(EnvironmentConfig{}); return p.Imports() }()},
		{"random-wrong-seed", func() component.Imports { p, _ := NewRandom(RandomConfig{}); return p.Imports() }()},
		{"exit-wrong-code", ExitImports()},
	} {
		m := compileFixture(t, "interpreter", item.name)
		if _, err := m.Instantiate(nil, item.imports, component.CallOptions{}); !errors.Is(err, vm.ErrType) {
			t.Fatal(item.name, err)
		}
	}
}
func TestSystemClockSignedRangeAndContracts(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		clock, err := NewSystemClock(SystemClockConfig{Now: func(context.Context) (Instant, error) { return Instant{-2, 999999999}, nil }, Resolution: func(context.Context) (uint64, error) { return 1000, nil }})
		if err != nil {
			t.Fatal(err)
		}
		m := compileFixture(t, backend, "system-clock")
		in := instantiate(t, m, clock.Imports())
		v := call(t, in, "instant")
		if int64(v.Elements[0].Bits) != -2 || v.Elements[1].Bits != 999999999 || call(t, in, "resolution").Bits != 1000 {
			t.Fatal(v)
		}
		system, _ := NewSystemClock(SystemClockConfig{})
		in = instantiate(t, m, system.Imports())
		v = call(t, in, "instant")
		if v.Elements[1].Bits >= 1e9 || call(t, in, "resolution").Bits == 0 {
			t.Fatal(v)
		}
		for _, mode := range []string{"nanoseconds", "zero-resolution", "error"} {
			config := clock.config
			switch mode {
			case "nanoseconds":
				config.Now = func(context.Context) (Instant, error) { return Instant{Nanoseconds: 1e9}, nil }
			case "zero-resolution":
				config.Resolution = func(context.Context) (uint64, error) { return 0, nil }
			case "error":
				config.Now = func(context.Context) (Instant, error) { return Instant{}, vm.ErrLink }
			}
			broken, _ := NewSystemClock(config)
			in = instantiate(t, m, broken.Imports())
			name := "instant"
			if mode == "zero-resolution" {
				name = "resolution"
			}
			f, _ := in.Function(name)
			if _, err := f.Call(nil, nil, component.CallOptions{}); err == nil {
				t.Fatal(mode)
			}
		}
		for _, name := range []string{"system-clock-wrong-sign", "system-clock-wrong-resolution"} {
			bad := compileFixture(t, backend, name)
			if _, err := bad.Instantiate(nil, clock.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) && !errors.Is(err, vm.ErrLink) {
				t.Fatal(name, err)
			}
		}
	})
	if _, err := NewSystemClock(SystemClockConfig{Now: func(context.Context) (Instant, error) { return Instant{}, nil }}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}
func waitActive(t *testing.T, m *MonotonicClock) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		m.mu.Lock()
		active := m.active
		m.mu.Unlock()
		if active > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("wait never started")
		}
		time.Sleep(time.Millisecond)
	}
}
func TestMonotonicClockSuspensionCancellationAndClose(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		clock, _ := NewMonotonicClock(MonotonicClockConfig{MaxWaits: 1})
		defer clock.Close()
		m := compileFixture(t, backend, "monotonic-clock")
		in := instantiate(t, m, clock.Imports())
		before := call(t, in, "now").Bits
		if call(t, in, "resolution").Bits != 1000 {
			t.Fatal("resolution")
		}
		after := call(t, in, "wait-for", component.Value{Bits: 2e6}).Bits
		if after < before+2e6 {
			t.Fatal(before, after)
		}
		if call(t, in, "wait-until", component.Value{Bits: after + 2e6}).Bits < after+2e6 {
			t.Fatal("early absolute wait")
		}
		call(t, in, "wait-for", component.Value{})
		call(t, in, "wait-until", component.Value{})
		bad := compileFixture(t, backend, "monotonic-clock-wrong-async")
		if _, err := bad.Instantiate(nil, clock.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) && !errors.Is(err, vm.ErrLink) {
			t.Fatal(err)
		}
		wait, _ := in.Function("wait-for")
		pending := make(chan error, 1)
		go func() {
			_, err := wait.Call(nil, []component.Value{{Bits: ^uint64(0)}}, component.CallOptions{})
			pending <- err
		}()
		waitActive(t, clock)
		other := instantiate(t, m, clock.Imports())
		second, _ := other.Function("wait-for")
		if _, err := second.Call(nil, []component.Value{{Bits: 1e9}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("unbounded provider waits", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		if err := clock.WaitClosed(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		clock.Close()
		clock.Close()
		if err := <-pending; !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("provider close", err)
		}
		if err := clock.WaitClosed(context.Background()); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"now", "wait-for"} {
			closedInstance := instantiate(t, m, clock.Imports())
			f, _ := closedInstance.Function(name)
			var args []component.Value
			if name == "wait-for" {
				args = []component.Value{{}}
			}
			if _, err := f.Call(nil, args, component.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("closed clock admitted operation", name, err)
			}
		}
		for _, name := range []string{"monotonic-clock-wrong-now", "monotonic-clock-wrong-wait"} {
			bad := compileFixture(t, backend, name)
			if _, err := bad.Instantiate(nil, clock.Imports(), component.CallOptions{}); !errors.Is(err, vm.ErrType) {
				t.Fatal(name, err)
			}
		}

		cancelled, _ := NewMonotonicClock(MonotonicClockConfig{})
		defer cancelled.Close()
		in = instantiate(t, m, cancelled.Imports())
		wait, _ = in.Function("wait-until")
		ctx, stop := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer stop()
		if _, err := wait.Call(ctx, []component.Value{{Bits: ^uint64(0)}}, component.CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("call cancellation", err)
		}
		cancelled.Close()
		if err := cancelled.WaitClosed(context.Background()); err != nil {
			t.Fatal(err)
		}
	})
}
func TestBaseFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/base.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Version   string                       `json:"wasiVersion"`
		WIT       string                       `json:"witArchiveSHA256"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Inputs    map[string]string            `json:"inputs"`
		Fixtures  map[string]map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Version != "0.3.2" || pins.WIT != "2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311" || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 20 {
		t.Fatal("changed producer")
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
	check("../../../scripts/build_preview3_base_fixtures.py", pins.Generator)
	for path, hash := range pins.Inputs {
		check("../../../"+path, hash)
	}
	for name, files := range pins.Fixtures {
		for ext, hash := range files {
			check("testdata/"+name+"."+ext, hash)
		}
	}
}

func TestVersionedCommandExportAndExit(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, mode := range []string{"success", "failure", "trap", "sync", "parameter", "payload", "exit"} {
			in := instantiate(t, compileFixture(t, backend, "command-"+mode), ExitImports())
			code, err := RunCommand(nil, in, component.CallOptions{})
			switch mode {
			case "sync", "parameter", "payload":
				if !errors.Is(err, vm.ErrType) {
					t.Fatal(mode, err)
				}
			case "trap":
				if !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(mode, err)
				}
			default:
				want := uint32(0)
				if mode == "failure" {
					want = 1
				}
				if mode == "exit" {
					want = 255
				}
				if err != nil || code != want {
					t.Fatal(mode, code, err)
				}
			}
		}
	})
}

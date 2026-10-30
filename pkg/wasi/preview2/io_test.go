package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestIOStreamsComponentBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			e, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "io-streams"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			output := NewOutputBuffer(1 << 20)
			process, err := NewIO(IOConfig{Stdin: NewBytesInput([]byte("abcdefgh")), Stdout: output})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				process.Close()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if err := process.WaitClosed(ctx); err != nil {
					t.Error(err)
				}
			}()
			in, err := m.Instantiate(nil, process.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			f, err := in.Function("exercise")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			result, err := f.Call(ctx, nil, component.CallOptions{Metered: true, Fuel: 1000000})
			if err != nil || result.Value.Bits != 42 {
				t.Fatal(result, err)
			}
			if got := string(output.Drain()); got != "XYZ\x00\x00\x00\x00\x00!efg" {
				t.Fatalf("output %q", got)
			}
			process.mu.Lock()
			remaining := len(process.records)
			process.mu.Unlock()
			if remaining != 0 {
				t.Fatal("stream drop retained records", remaining)
			}
			f, err = in.Function("write-batch")
			if err != nil {
				t.Fatal(err)
			}
			result, err = f.Call(ctx, []component.Value{{Bits: 3}}, component.CallOptions{Metered: true, Fuel: 1000000})
			if err != nil || result.Value.Bits != 3 {
				t.Fatal(result, err)
			}
			if got := len(output.Drain()); got != 3*4096 {
				t.Fatal("write bytes", got)
			}
		})
	}
}

type failingOutput struct{ cause error }

func (f failingOutput) Readiness() (bool, <-chan struct{})         { return true, nil }
func (f failingOutput) Write(context.Context, []byte) (int, error) { return 0, f.cause }
func (f failingOutput) Flush(context.Context) error                { return f.cause }
func (f failingOutput) Close() error                               { return nil }

func TestIOErrorResourceFromFailedStream(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := component.DefaultConfig()
		config.Core.Backend = backend
		e, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, fixture(t, "io-streams"))
		if err != nil {
			t.Fatal(err)
		}
		process, err := NewIO(IOConfig{Stdout: failingOutput{errors.New("forced output failure")}})
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, process.Imports(), component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, err := in.Function("error-probe")
		if err != nil {
			t.Fatal(err)
		}
		result, err := f.Call(nil, nil, component.CallOptions{})
		if err != nil || result.Value.Bits != 1 {
			t.Fatal(result, err)
		}
		process.errors.mu.Lock()
		live := len(process.errors.entries)
		process.errors.mu.Unlock()
		if live != 0 {
			t.Fatal("error resource retained", live)
		}
		in.Close()
		process.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := process.WaitClosed(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
		m.Close()
		e.Close()
	}
}

func TestIOGuestProtocolTraps(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := component.DefaultConfig()
		config.Core.Backend = backend
		e, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, fixture(t, "io-streams"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"write-unchecked", "write-overpermit", "blocking-too-large", "parent-drop"} {
			process, err := NewIO(IOConfig{})
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, process.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			f, err := in.Function(name)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err = f.Call(ctx, nil, component.CallOptions{Metered: true, Fuel: 1000000})
			want := wasm.ErrTrap
			if name == "blocking-too-large" {
				want = wasm.ErrLimit
			}
			if !errors.Is(err, want) {
				t.Fatal(backend, name, err)
			}
			in.Close()
			process.Close()
			if err := process.WaitClosed(ctx); err != nil {
				t.Fatal(err)
			}
			cancel()
		}
		m.Close()
		e.Close()
	}
}

func TestIOTerminalCapabilitiesAndContractChecks(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := component.DefaultConfig()
		config.Core.Backend = backend
		e, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, fixture(t, "io-streams"))
		if err != nil {
			t.Fatal(err)
		}
		for _, present := range []bool{false, true} {
			p, err := NewIO(IOConfig{StdinTerminal: present, StdoutTerminal: present, StderrTerminal: present})
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, p.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			f, err := in.Function("terminal-probe")
			if err != nil {
				t.Fatal(err)
			}
			result, err := f.Call(nil, nil, component.CallOptions{})
			expected := uint64(0)
			if present {
				expected = 3
			}
			if err != nil || result.Value.Bits != expected {
				t.Fatal(result, err)
			}
			in.Close()
			p.Close()
			if err := p.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		}
		m.Close()
		for _, name := range []string{"read-width", "write-list", "subscribe-result", "error-case", "error-label", "skip-result", "extra-argument", "debug-result"} {
			p, err := NewIO(IOConfig{})
			if err != nil {
				t.Fatal(err)
			}
			invalid, err := e.Compile(nil, fixture(t, "io-contract-"+name))
			if err != nil {
				t.Fatal(name, err)
			}
			in, err := invalid.Instantiate(nil, p.Imports(), component.CallOptions{})
			if in != nil || err == nil {
				t.Fatal(backend, name, in, err)
			}
			invalid.Close()
			p.Close()
			if err := p.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		}
		e.Close()
	}
}

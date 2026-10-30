package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestWallClockBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			cfg := component.DefaultConfig()
			cfg.Core.Backend = backend
			e, err := component.NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "wall-clock"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, sample := range []Datetime{{}, {Seconds: 1, Nanoseconds: 999999999}, {Seconds: 1<<63 + 7, Nanoseconds: 1}, {Seconds: ^uint64(0), Nanoseconds: 999999999}} {
				now := sample
				clock, err := NewWallClock(WallClockConfig{Now: func(context.Context) (Datetime, error) { return now, nil }, Resolution: func(context.Context) (Datetime, error) { return Datetime{Seconds: 2, Nanoseconds: 3}, nil }})
				if err != nil {
					t.Fatal(err)
				}
				in, err := m.Instantiate(nil, clock.Imports(), component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				f, err := in.Function("now")
				if err != nil {
					t.Fatal(err)
				}
				first, err := f.Call(nil, nil, component.CallOptions{Metered: true, Fuel: 1000})
				if err != nil || len(first.Value.Elements) != 2 || first.Value.Elements[0].Bits != sample.Seconds || first.Value.Elements[1].Bits != uint64(sample.Nanoseconds) {
					t.Fatal(first, err)
				}
				now = Datetime{Seconds: 42, Nanoseconds: 4}
				again, err := f.Call(nil, nil, component.CallOptions{})
				if err != nil || again.Value.Elements[0].Bits != 42 || first.Value.Elements[0].Bits != sample.Seconds {
					t.Fatal(again, first, err)
				}
				f, err = in.Function("resolution")
				if err != nil {
					t.Fatal(err)
				}
				result, err := f.Call(nil, nil, component.CallOptions{})
				if err != nil || result.Value.Elements[0].Bits != 2 || result.Value.Elements[1].Bits != 3 {
					t.Fatal(result, err)
				}
				in.Close()
			}
			clock, err := NewWallClock(WallClockConfig{})
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, clock.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			f, err := in.Function("now")
			if err != nil {
				t.Fatal(err)
			}
			before := uint64(time.Now().Unix())
			result, err := f.Call(nil, nil, component.CallOptions{})
			after := uint64(time.Now().Unix())
			if err != nil || result.Value.Elements[0].Bits < before || result.Value.Elements[0].Bits > after || result.Value.Elements[1].Bits >= 1000000000 || result.Value.Elements[1].Bits%1000 != 0 {
				t.Fatal(result, err, before, after)
			}
			f, err = in.Function("resolution")
			if err != nil {
				t.Fatal(err)
			}
			result, err = f.Call(nil, nil, component.CallOptions{})
			if err != nil || result.Value.Elements[0].Bits != 0 || result.Value.Elements[1].Bits != 1000 {
				t.Fatal(result, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err = f.Call(ctx, nil, component.CallOptions{}); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		})
	}
}

func TestWallClockContractsAndInvalidSources(t *testing.T) {
	now := func(context.Context) (Datetime, error) { return Datetime{Seconds: 42}, nil }
	for _, cfg := range []WallClockConfig{{Now: now}, {Resolution: now}} {
		if _, err := NewWallClock(cfg); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, w := range []*WallClock{nil, {}} {
		if len(w.Imports().Instances) != 0 {
			t.Fatal("empty clock exports a provider")
		}
	}
	e, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	w, err := NewWallClock(WallClockConfig{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wrong-kind", "wrong-names", "wrong-order", "wrong-width", "wrong-count", "wrong-params"} {
		m, err := e.Compile(nil, fixture(t, "wall-clock-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		if in, err := m.Instantiate(nil, w.Imports(), component.CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal(name, in, err)
		}
		m.Close()
	}
	sentinel := errors.New("host clock failure")
	for _, resolution := range []bool{false, true} {
		for _, sample := range []struct {
			v   Datetime
			err error
		}{{Datetime{Nanoseconds: 1000000000}, nil}, {Datetime{}, sentinel}, {Datetime{}, nil}} {
			if !resolution && sample.v == (Datetime{}) && sample.err == nil {
				continue
			}
			cfg := WallClockConfig{Now: now, Resolution: now}
			bad := func(context.Context) (Datetime, error) { return sample.v, sample.err }
			if resolution {
				cfg.Resolution = bad
			} else {
				cfg.Now = bad
			}
			clock, err := NewWallClock(cfg)
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, fixture(t, "wall-clock"))
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, clock.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			name := "now"
			if resolution {
				name = "resolution"
			}
			f, err := in.Function(name)
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.Call(nil, nil, component.CallOptions{})
			want := sample.err
			if want == nil {
				want = wasm.ErrInvalid
			}
			if !errors.Is(err, want) {
				t.Fatal(name, sample, err)
			}
			in.Close()
			m.Close()
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := w.read(ctx, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := systemWallNow(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := systemWallResolution(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestWallClockCancellationAndClose(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, closeStore := range []bool{false, true} {
			cfg := component.DefaultConfig()
			cfg.Core.Backend = backend
			e, err := component.NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, fixture(t, "wall-clock"))
			if err != nil {
				t.Fatal(err)
			}
			started := make(chan struct{})
			w, err := NewWallClock(WallClockConfig{Now: func(ctx context.Context) (Datetime, error) {
				close(started)
				<-ctx.Done()
				return Datetime{Seconds: 42}, nil
			}, Resolution: systemWallResolution})
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, w.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			f, err := in.Function("now")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { _, err := f.Call(ctx, nil, component.CallOptions{}); done <- err }()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("source not entered")
			}
			if closeStore {
				in.Close()
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("source not canceled")
			}
			cancel()
			in.Close()
			wait, cancelWait := context.WithTimeout(context.Background(), time.Second)
			if err := in.WaitClosed(wait); err != nil {
				t.Fatal(err)
			}
			cancelWait()
			m.Close()
			e.Close()
		}
	}
}

package component

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPreparedResultUsesOneAdmissionAndExpiresBothCallers(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in := asyncInstance(t, backend, "host-result-stream", Imports{})
		run, _ := in.Function("run")
		var preparation, consumption *Caller
		prepared, consumed := 0, 0
		fuel, err := run.CallWithResultPrepared(nil, CallOptions{Metered: true, Fuel: 1000}, func(c *Caller) ([]Value, error) {
			prepared++
			preparation = c
			if !c.FunctionInfo().Async() || c.FunctionInfo().ParameterCount() != 0 {
				t.Fatal("preparation has the wrong function signature")
			}
			if _, err := in.WithCaller(nil, CallOptions{}, func(*Caller) error { return nil }); !errors.Is(err, vm.ErrBusy) {
				t.Fatal("preparation did not own the Store", err)
			}
			return nil, nil
		}, func(c *Caller, value *Value) error {
			consumed++
			consumption = c
			if _, err := preparation.ReadAsync(nil, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("preparation Caller survived into guest consumption", err)
			}
			r, err := c.ReadAsync(value.Resource, 1)
			if err != nil || len(r.Values) != 1 || r.Values[0].Bits != 42 {
				t.Fatal(r, err)
			}
			return c.DropAsync(value.Resource)
		})
		if err != nil || fuel == 0 || fuel >= 1000 || prepared != 1 || consumed != 1 {
			t.Fatal(fuel, prepared, consumed, err)
		}
		if _, err := consumption.ReadAsync(nil, 1); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		entered, release := make(chan struct{}), make(chan struct{})
		var releaseOnce sync.Once
		defer releaseOnce.Do(func() { close(release) })
		done := make(chan error, 1)
		go func() {
			_, err := in.WithCaller(context.Background(), CallOptions{}, func(*Caller) error {
				close(entered)
				<-release
				return nil
			})
			done <- err
		}()
		<-entered
		_, err = run.CallWithResultPrepared(nil, CallOptions{}, func(*Caller) ([]Value, error) {
			t.Error("rejected admission invoked preparation")
			return nil, nil
		}, func(*Caller, *Value) error { return nil })
		if !errors.Is(err, vm.ErrBusy) {
			t.Fatal(err)
		}
		releaseOnce.Do(func() { close(release) })
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPreparedResultFailureReclaimsNewCapabilities(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, failure := range []string{"error", "panic", "arity"} {
			in := asyncInstance(t, backend, "host-result-stream", Imports{})
			run, _ := in.Function("run")
			closed := 0
			invoke := func() (err error) {
				defer func() {
					if recover() != nil {
						err = wasm.ErrTrap
					}
				}()
				_, err = run.CallWithResultPrepared(nil, CallOptions{}, func(c *Caller) ([]Value, error) {
					_, err := c.NewSource(ByteStreamType(), AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
						return nil, vm.ErrLink
					}, Close: func() { closed++ }})
					if err != nil {
						return nil, err
					}
					if failure == "panic" {
						panic("prepare")
					}
					if failure == "arity" {
						return []Value{{Bits: 1}}, nil
					}
					return nil, vm.ErrLink
				}, func(*Caller, *Value) error {
					t.Error("failed preparation reached result consumer")
					return nil
				})
				return err
			}
			if err := invoke(); err == nil || closed != 1 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal(failure, err, closed, in.tasks, in.heldValues, in.heldBytes)
			}
			probe := canonical.NewResourceTable(in.resourceBudget)
			for range in.config.MaxResources {
				if _, err := probe.AddOpaque(1); err != nil {
					t.Fatal("preparation failure retained a handle", err)
				}
			}
			probe.Close()
		}
	}
	if _, err := new(Function).CallWithResultPrepared(nil, CallOptions{}, func(*Caller) ([]Value, error) { return nil, nil }, func(*Caller, *Value) error { return nil }); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if _, err := new(Function).CallWithResultPrepared(nil, CallOptions{}, nil, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
}

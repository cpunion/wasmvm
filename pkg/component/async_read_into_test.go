package component

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestReadAsyncIntoGuestWritesAndMemoryWidths(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, name := range []string{"host-result-bytes", "host-result-bytes-memory64"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				in := asyncInstance(t, backend, name, Imports{})
				run, _ := in.Function("run")
				var retained []byte
				for round := range 2 {
					var data []byte
					var requests []uint32
					var escaped *Caller
					fuel, err := run.CallWithResult(nil, nil, CallOptions{Metered: true, Fuel: 1000}, func(c *Caller, v *Value) error {
						escaped = c
						for {
							var chunk []byte
							r, err := c.ReadAsyncInto(nil, v.Resource, 16, func(n uint32) ([]byte, error) {
								if _, err := c.Call(nil, nil); !errors.Is(err, vm.ErrBusy) {
									t.Fatal("destination reentered caller", err)
								}
								requests = append(requests, n)
								grown := make([]byte, n)
								copy(grown, chunk)
								chunk = grown
								runtime.GC()
								return chunk, nil
							})
							if err != nil || r.Bytes != "" || r.Values != nil || uint32(len(chunk)) != r.Count {
								t.Fatal(r, err)
							}
							data = append(data, chunk...)
							if r.Closed {
								break
							}
						}

						return c.DropAsync(v.Resource)
					})
					if err != nil || fuel == 0 || fuel >= 1000 || string(data) != "\x80\xff" || !reflect.DeepEqual(requests, []uint32{1, 1}) {
						t.Fatal(fuel, data, requests, err)
					}
					if _, err := escaped.ReadAsyncInto(nil, nil, 1, func(uint32) ([]byte, error) { t.Fatal("expired callback invoked"); return nil, nil }); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal(err)
					}
					if round == 0 {
						retained = data
						retained[0] = 42
					} else if retained[0] != 42 {
						t.Fatal("later execution changed caller-owned output")
					}
				}
				if in.tasks != nil || in.heldBytes != 0 || in.heldValues != 0 {
					t.Fatal("read retained task or budget")
				}
			})
		}
	}
}

func TestReadAsyncIntoHostBatchesAndEmptyReads(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, legacy := range []bool{false, true} {
			in := asyncInstance(t, backend, "host-result-stream", Imports{})
			closed := 0
			value, err := in.NewSource(nil, ByteStreamType(), AsyncSource{Close: func() { closed++ }, Read: func(_ *Caller, n uint32) (*ReadCompletion, error) {
				r := NewReadCompletion()
				r.Resolve(func(*Caller) (ReadResult, error) {
					if n == 0 {
						return ReadResult{}, nil
					}
					if legacy {
						return ReadResult{Values: []Value{{Bits: 0}, {Bits: 255}, {Bits: 128}}, Closed: true}, nil
					}
					return ReadResult{Bytes: "\x00\xff\x80", Closed: true}, nil
				})
				return r, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			destination := func(uint32) ([]byte, error) { t.Fatal("empty read requested storage"); return nil, nil }
			r, err := in.ReadAsyncInto(nil, value.Resource, 0, destination, CallOptions{})
			if err != nil || r.Count != 0 || r.Closed {
				t.Fatal(r, err)
			}
			var data []byte
			r, err = in.ReadAsyncInto(nil, value.Resource, 16, func(n uint32) ([]byte, error) { data = make([]byte, n); return data, nil }, CallOptions{})
			if err != nil || r.Count != 3 || !r.Closed || r.Bytes != "" || string(data) != "\x00\xff\x80" || closed != 1 {
				t.Fatal(r, data, closed, err)
			}
			if _, err := in.DropAsync(nil, value.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestReadAsyncIntoRejectsOwnershipTypesAndLimits(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		in, typ := sourceType(t, backend, "host-source-stream-u8")
		other := asyncInstance(t, backend, "host-result-stream", Imports{})
		source, err := in.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
			t.Fatal("invalid read started source")
			return nil, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		dst := func(uint32) ([]byte, error) { t.Fatal("invalid read requested storage"); return nil, nil }
		if _, err := other.ReadAsyncInto(nil, source.Resource, 1, dst, CallOptions{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
		if _, err := in.ReadAsyncInto(nil, source.Resource, 1, nil, CallOptions{}); !errors.Is(err, vm.ErrType) {
			t.Fatal(err)
		}
		if _, err := in.ReadAsyncInto(nil, source.Resource, 1<<28, dst, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if _, err := in.ReadAsyncInto(nil, nil, 1, dst, CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		in.DropAsync(nil, source.Resource, CallOptions{})
		for _, fixture := range []string{"host-source-future-u32", "host-source-stream-unit", "host-source-stream-u32"} {
			instance, typ := sourceType(t, backend, fixture)
			v, err := instance.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { t.Fatal("wrong type consumed"); return nil, nil }})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := instance.ReadAsyncInto(nil, v.Resource, 1, dst, CallOptions{}); !errors.Is(err, vm.ErrType) {
				t.Fatal(fixture, err)
			}
			if _, err := instance.DropAsync(nil, v.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestReadAsyncIntoFailureCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, mode := range []string{"short", "error", "panic", "cancel", "quota"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				in := asyncInstance(t, backend, "host-result-bytes", Imports{})
				run, _ := in.Function("run")
				sentinel := errors.New("destination rejected data")
				invoke := func() (err error) {
					defer func() {
						if p := recover(); p != nil {
							if p != "destination" {
								t.Fatal(p)
							}
							err = sentinel
						}
					}()
					_, err = run.CallWithResult(nil, nil, CallOptions{}, func(c *Caller, v *Value) error {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						previous := in.config.MaxBytes
						if mode == "quota" {
							in.config.MaxBytes = 1
						}
						defer func() { in.config.MaxBytes = previous }()
						_, err := c.ReadAsyncInto(ctx, v.Resource, 16, func(n uint32) ([]byte, error) {
							if mode == "quota" {
								t.Fatal("source quota failure invoked destination")
							}
							switch mode {
							case "short":
								return nil, nil
							case "error":
								return nil, sentinel
							case "panic":
								panic("destination")
							case "cancel":
								cancel()
							}
							return make([]byte, n), nil
						})
						return err
					})
					return err
				}
				err := invoke()
				want := error(sentinel)
				switch mode {
				case "short":
					want = vm.ErrType
				case "cancel":
					want = context.Canceled
				case "quota":
					want = wasm.ErrLimit
				}
				if !errors.Is(err, want) {
					t.Fatal(mode, err)
				}
				if in.tasks != nil || in.heldBytes != 0 || in.heldValues != 0 {
					t.Fatal("failure retained work/budget")
				}
			})
		}
	}
}

func TestReadAsyncIntoCancellationRejectsLateCompletion(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		in := asyncInstance(t, backend, "host-result-stream", Imports{})
		started := make(chan *ReadCompletion, 1)
		closed := 0
		value, err := in.NewSource(nil, ByteStreamType(), AsyncSource{Close: func() { closed++ }, Read: func(*Caller, uint32) (*ReadCompletion, error) { p := NewReadCompletion(); started <- p; return p, nil }})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := in.ReadAsyncInto(ctx, value.Resource, 16, func(uint32) ([]byte, error) { t.Error("cancelled read retained destination"); return nil, nil }, CallOptions{})
			done <- err
		}()
		pending := <-started
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		_ = pending.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "late", Closed: true}, nil })
		if closed != 1 || in.tasks != nil || in.heldBytes != 0 || in.heldValues != 0 {
			t.Fatal("late completion retained work", closed)
		}
		if _, err := in.DropAsync(nil, value.Resource, CallOptions{}); err != nil {
			t.Fatal(err)
		}
	}
}

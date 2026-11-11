package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestReadyByteSourceAndDeferredFallback(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			in := asyncInstance(t, backend, "host-result-stream", Imports{})
			probes, starts, closes := 0, 0, 0
			v, err := in.NewSource(nil, ByteStreamType(), AsyncSource{
				ReadBytes: func(limit uint32) (ReadResult, bool, error) {
					probes++
					if limit == 0 {
						return ReadResult{}, true, nil
					}
					if probes == 2 {
						return ReadResult{Bytes: "\x00\xff", Count: 2}, true, nil
					}
					return ReadResult{}, false, nil
				},
				Read: func(*Caller, uint32) (*ReadCompletion, error) {
					starts++
					p := NewReadCompletion()
					p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "\x80b", Closed: true}, nil })
					return p, nil
				}, Close: func() { closes++ },
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"", "\x00\xff", "\x80b"} {
				limit := uint32(len(want))
				r, err := in.ReadAsync(nil, v.Resource, limit, CallOptions{})
				if err != nil || r.Bytes != want || r.Closed != (want == "\x80b") {
					t.Fatal(r, err)
				}
				if in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal("completed bytes retained value reservations")
				}
			}
			if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil || probes != 3 || starts != 1 || closes != 1 {
				t.Fatal(probes, starts, closes, err)
			}
		})
	}
}

func TestReadyByteSourceValidationFailureAndCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, mode := range []string{"bounds", "count", "values", "empty", "not-ready-data", "error", "panic", "close-panic", "byte-limit", "value-limit", "destination"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				in := asyncInstance(t, backend, "host-result-stream", Imports{})
				closes := 0
				v, err := in.NewSource(nil, ByteStreamType(), AsyncSource{
					ReadBytes: func(uint32) (ReadResult, bool, error) {
						r := ReadResult{Bytes: "ok", Closed: true}
						switch mode {
						case "bounds":
							r.Bytes = "oversized"
						case "count":
							r.Count = 1
						case "values":
							r.Values = []Value{{Bits: 1}}
						case "empty":
							r = ReadResult{}
						case "not-ready-data":
							return r, false, nil
						case "error":
							return ReadResult{}, false, vm.ErrLink
						case "panic":
							panic(mode)
						}
						return r, true, nil
					}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
						t.Fatal("invalid probe reached fallback")
						return nil, nil
					},
					Close: func() {
						closes++
						if mode == "close-panic" {
							panic(mode)
						}
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if mode == "byte-limit" {
					in.config.MaxBytes = 1
				}
				if mode == "value-limit" {
					in.config.MaxValues = 1
				}
				invoke := func() (err error) {
					defer func() {
						if recover() != nil {
							err = wasm.ErrTrap
						}
					}()
					if mode == "destination" {
						_, err = in.ReadAsyncInto(nil, v.Resource, 2, func(uint32) ([]byte, error) { return nil, vm.ErrLink }, CallOptions{})
					} else {
						_, err = in.ReadAsync(nil, v.Resource, 2, CallOptions{})
					}
					return err
				}
				if err := invoke(); err == nil || closes != 1 || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal(mode, err, closes, in.heldValues, in.heldBytes)
				}
			})
		}
	}
}

func TestReadyByteSourceAdmissionAndSchema(t *testing.T) {
	in := asyncInstance(t, "interpreter", "host-result-stream", Imports{})
	_, numeric := sourceType(t, "interpreter", "host-source-stream-u32")
	probe := func(uint32) (ReadResult, bool, error) {
		t.Fatal("rejected read started a producer")
		return ReadResult{}, false, nil
	}
	fallback := func(*Caller, uint32) (*ReadCompletion, error) {
		t.Fatal("rejected read started fallback")
		return nil, nil
	}
	if _, err := in.NewSource(nil, numeric, AsyncSource{ReadBytes: probe, Read: fallback}); !errors.Is(err, vm.ErrType) {
		t.Fatal("non-byte schema admitted a byte probe", err)
	}
	if _, err := in.NewSource(nil, ByteStreamType(), AsyncSource{ReadBytes: probe}); !errors.Is(err, vm.ErrType) {
		t.Fatal("missing fallback accepted", err)
	}
	for _, mode := range []string{"cancel", "tasks", "reentry"} {
		v, err := in.NewSource(nil, ByteStreamType(), AsyncSource{ReadBytes: probe, Read: fallback})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		oldTasks, oldReentry := in.config.MaxTasks, in.config.Store.MaxReentry
		want := wasm.ErrLimit
		switch mode {
		case "cancel":
			cancel()
			want = context.Canceled
		case "tasks":
			in.config.MaxTasks = 0
		case "reentry":
			in.config.Store.MaxReentry = 0
		}
		_, err = in.ReadAsync(ctx, v.Resource, 2, CallOptions{})
		cancel()
		in.config.MaxTasks, in.config.Store.MaxReentry = oldTasks, oldReentry
		if !errors.Is(err, want) {
			t.Fatal(mode, err)
		}
		if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil {
			t.Fatal(err)
		}
	}
}

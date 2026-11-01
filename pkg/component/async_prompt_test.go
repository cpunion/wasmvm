package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPromptSourceAndWorkRejectBlockingReentry(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			starts, finishes := 0, 0
			guard := func(c *Caller) {
				if _, err := c.Call(nil, nil); !errors.Is(err, vm.ErrType) {
					t.Fatal("prompt guest reentry", err)
				}
				if _, err := c.ReadAsync(nil, 1); !errors.Is(err, vm.ErrType) {
					t.Fatal("prompt blocking read", err)
				}
			}
			_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				value, err := c.NewSource(typ, AsyncSource{Prompt: true, Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
					guard(c)
					result := NewReadCompletion()
					result.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "ready", Closed: true}, nil })
					return result, nil
				}})
				if err != nil {
					return err
				}
				return c.Schedule(AsyncWork{Prompt: true, Owned: []*Resource{value.Resource}, Start: func(c *Caller, refs []*Resource) (*HostCompletion, error) {
					guard(c)
					starts++
					return c.AwaitRead(nil, refs[0], 16, func(_ *Caller, r ReadResult, err error) (*Value, error) {
						if err != nil || r.Bytes != "ready" || !r.Closed {
							t.Fatal(r, err)
						}
						return nil, nil
					})
				}, Done: func(c *Caller, cancelled bool) error {
					guard(c)
					finishes++
					if cancelled {
						t.Fatal("cancelled ready work")
					}
					return nil
				}})
			})
			if err != nil || starts != 1 || finishes != 1 || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal(starts, finishes, in.heldValues, in.heldBytes, err)
			}
		})
	}
}

func TestCheckedReadsRejectBeforeConsumingHostBatches(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, reject := range []bool{false, true} {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			denied := errors.New("batch exceeds the declared size")
			observed := false
			_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				value, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
					result := NewReadCompletion()
					result.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "payload", Closed: true}, nil })
					return result, nil
				}})
				if err != nil {
					return err
				}
				return c.Schedule(AsyncWork{Owned: []*Resource{value.Resource}, Start: func(c *Caller, refs []*Resource) (*HostCompletion, error) {
					return c.AwaitReadChecked(nil, refs[0], 16, func(r ReadResult) error {
						if r.Bytes != "payload" || r.Count != 7 {
							t.Fatal("incorrect validation batch", r)
						}
						if reject {
							return denied
						}
						return nil
					}, func(_ *Caller, r ReadResult, err error) (*Value, error) {
						observed = true
						if reject {
							if !errors.Is(err, denied) || r.Count != 0 {
								t.Fatal(r, err)
							}
						} else if err != nil || r.Bytes != "payload" || r.Count != 7 || !r.Closed {
							t.Fatal(r, err)
						}
						return nil, nil
					})
				}})
			})
			if err != nil || !observed || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal(backend, reject, observed, err)
			}
		}
	}
}

func TestForwardReadCompletionAndValidation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		in, typ := sourceType(t, backend, "host-source-stream-u8")
		var forwarded Value
		_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			original, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
				p := NewReadCompletion()
				p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "forwarded", Closed: true}, nil })
				return p, nil
			}})
			if err != nil {
				return err
			}
			if _, err := c.AwaitSourceRead(nil, original.Resource, 1, nil); !errors.Is(err, vm.ErrType) {
				t.Fatal("nil transform", err)
			}
			if _, err := c.AwaitReadChecked(nil, original.Resource, 1, nil, func(*Caller, ReadResult, error) (*Value, error) { return nil, nil }); !errors.Is(err, vm.ErrType) {
				t.Fatal("nil check", err)
			}
			forwarded, err = c.NewSource(typ, AsyncSource{Owned: []*Resource{original.Resource}, ReadOwned: func(c *Caller, refs []*Resource, limit uint32) (*ReadCompletion, error) {
				return c.AwaitSourceRead(nil, refs[0], limit, func(_ *Caller, r ReadResult, err error) (ReadResult, error) { return r, err })
			}})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		r, err := in.ReadAsync(nil, forwarded.Resource, 16, CallOptions{})
		if err != nil || r.Bytes != "forwarded" || !r.Closed {
			t.Fatal(r, err)
		}
		if _, err := in.DropAsync(nil, forwarded.Resource, CallOptions{}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestHostBufferAccumulatesBeforeEventDelivery(t *testing.T) {
	in, _ := sourceType(t, "interpreter", "host-source-stream-u8")
	_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
		for _, kind := range []model.Kind{model.U8, model.String} {
			element := &model.Type{Kind: kind}
			batch, err := in.bufferType(element)
			if err != nil {
				return err
			}
			b := &guestBuffer{in: in, scope: &namespace{}, typ: element, batch: batch, length: 8}
			chunks := []Value{{Text: "ab"}, {Text: "cd"}}
			if kind == model.String {
				chunks = []Value{{Elements: []Value{{Text: "first"}}}, {Elements: []Value{{Text: "second"}}}}
			}
			for _, chunk := range chunks {
				codec, err := b.store(c.core, chunk)
				if err != nil {
					return err
				}
				if err := finishResources(codec, true, nil); err != nil {
					return err
				}
				b.progress++
			}
			if kind == model.U8 && b.data.Text != "abcd" || kind == model.String && (len(b.data.Elements) != 2 || b.data.Elements[0].Text != "first" || b.data.Elements[1].Text != "second") {
				t.Fatal("overwritten pending host buffer", b.data)
			}
			values, bytes := in.config.MaxValues, in.config.MaxBytes
			in.config.MaxValues, in.config.MaxBytes = 1, 1
			_, failed := b.store(c.core, chunks[0])
			in.config.MaxValues, in.config.MaxBytes = values, bytes
			if !errors.Is(failed, wasm.ErrLimit) {
				t.Fatal("cumulative quota bypass", failed)
			}
		}
		// The budget can still be fully allocated after both success and rejection.
		table := canonical.NewResourceTable(in.resourceBudget)
		defer table.Close()
		for range in.config.MaxResources {
			if _, err := table.AddOpaque(1); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

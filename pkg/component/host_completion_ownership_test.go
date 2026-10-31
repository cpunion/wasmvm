package component

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCompletionOwnershipAndBudgets(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			drops := uint32(0)
			rt := NewHostResourceType(func(*Caller, uint32) error { drops++; return nil })
			var fail bool
			var kept Value
			delayed := func(c *Caller, args []Value, resolve func(*Caller) (*Value, error)) (*HostCompletion, error) {
				p := NewHostCompletion()
				go func() { p.Resolve(resolve) }()
				return p, nil
			}
			imports := Imports{Resources: map[string]*HostResourceType{"r": rt}, AsyncCompletions: map[string]AsyncHostFunction{
				"pass": func(c *Caller, args []Value) (*HostCompletion, error) {
					return delayed(c, args, func(fresh *Caller) (*Value, error) {
						if _, err := fresh.ResourceRep(rt, args[0].Resource); err != nil {
							return nil, err
						}
						if fail {
							return nil, vm.ErrLink
						}
						return &args[0], nil
					})
				},
				"borrow": func(c *Caller, args []Value) (*HostCompletion, error) {
					kept = args[0]
					return delayed(c, args, func(fresh *Caller) (*Value, error) {
						rep, err := fresh.ResourceRep(rt, args[0].Resource)
						if err != nil {
							return nil, err
						}
						if err := fresh.DropResource(args[0].Resource); !errors.Is(err, vm.ErrReference) {
							return nil, vm.ErrType
						}
						return &Value{Bits: uint64(rep)}, nil
					})
				},
				"make": func(c *Caller, args []Value) (*HostCompletion, error) {
					return delayed(c, args, func(fresh *Caller) (*Value, error) {
						v, err := fresh.NewResource(rt, 42)
						if err != nil {
							return nil, err
						}
						if fail {
							return nil, vm.ErrLink
						}
						return &v, nil
					})
				},
				"blob": func(c *Caller, args []Value) (*HostCompletion, error) {
					return delayed(c, args, func(*Caller) (*Value, error) { return &args[0], nil })
				},
			}}
			in := asyncInstance(t, backend, "async-host-resources", imports)
			pass, _ := in.Function("pass")
			borrow, _ := in.Function("borrow")
			makeFn, _ := in.Function("make")
			blob, _ := in.Function("blob")
			input, err := in.NewResource(nil, rt, 42)
			if err != nil {
				t.Fatal(err)
			}
			v, err := borrow.Call(nil, []Value{input}, CallOptions{})
			if err != nil || v.Value.Bits != 42 {
				t.Fatal(v, err)
			}
			if _, err := in.ResourceRep(nil, rt, kept.Resource); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("escaped borrow", err)
			}
			v, err = pass.Call(nil, []Value{input}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.ResourceRep(nil, rt, input.Resource); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("old owner", err)
			}
			fail = true
			if _, err := pass.Call(nil, []Value{*v.Value}, CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal(err)
			}
			if drops != 1 {
				t.Fatal("failed move cleanup", drops)
			}
			if _, err := makeFn.Call(nil, nil, CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal(err)
			}
			if drops != 2 {
				t.Fatal("failed delayed result cleanup", drops)
			}
			fail = false
			v, err = makeFn.Call(nil, nil, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.DropResource(nil, v.Value.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			if drops != 3 {
				t.Fatal(drops)
			}
			in.config.MaxValues = 12
			in.config.MaxBytes = 8
			for _, size := range []int{4, 5, 4} {
				value, err := blob.Call(nil, []Value{{Text: string(make([]byte, size))}}, CallOptions{})
				if (size <= 4) != (err == nil) {
					t.Fatal(size, value, err)
				}
				if in.heldBytes != 0 || in.heldValues != 0 || in.tasks != nil {
					t.Fatal("retained budget")
				}
			}
			// Three parked host arguments share a budget rather than each receiving a
			// full per-call allowance. Failure must cancel and release the earlier two.
			pending := make([]*HostCompletion, 0, 3)
			limits := asyncInstance(t, backend, "async-host-completion", Imports{
				Functions: map[string]HostFunction{"notify": func(*Caller, []Value) (*Value, error) { return nil, nil }},
				AsyncCompletions: map[string]AsyncHostFunction{"add": func(*Caller, []Value) (*HostCompletion, error) {
					p := NewHostCompletion()
					pending = append(pending, p)
					return p, nil
				}},
			})
			limits.config.MaxValues = 2
			multi, _ := limits.Function("multi")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err = multi.Call(ctx, []Value{{Bits: 1}}, CallOptions{})
			cancel()
			if !errors.Is(err, wasm.ErrLimit) || len(pending) != 2 {
				t.Fatal("aggregate pending values", len(pending), err)
			}
			if limits.heldValues != 0 || limits.heldBytes != 0 || limits.tasks != nil {
				t.Fatal("failed pending budget")
			}
			for _, p := range pending {
				if err := p.Cancelled(); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("abandoned completion", err)
				}
			}
		})
	}
}

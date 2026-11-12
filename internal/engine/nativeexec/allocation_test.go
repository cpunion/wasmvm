package nativeexec

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type allocationPanicContext struct{ context.Context }

func (allocationPanicContext) Err() error { panic("allocation entry context") }

func TestNativeAllocationReservationCleanupBeforeEntry(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, panics := range []bool{false, true} {
		d := &module.Module{Types: []module.FuncType{{}, {}}, TypeDefs: []module.TypeDef{{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32}}}, {Kind: module.FunctionKind}}, Functions: []module.Function{{Type: 1, Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xfb0000}, {Op: 0x1a}, {Op: 0xb}}}}}
		checked, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.GC)
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Compile(checked, lower.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		m, err := slots.Compile(ir, slots.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		e, fn, _ := prepare(t, m)
		if !e.Suspend() {
			t.Fatal("initial suspension")
		}
		ctx, cancel := context.WithCancel(context.Background())
		var entryContext context.Context = ctx
		if panics {
			entryContext = allocationPanicContext{context.Background()}
		}
		if err := e.Begin(entryContext, false); err != nil {
			t.Fatal(err)
		}
		cancel()
		var caught any
		func() {
			defer func() { caught = recover() }()
			_, fuel, err := e.Run(fn, 0, 1, 0, 17)
			if panics || !errors.Is(err, context.Canceled) || fuel != 17 {
				t.Fatal("pre-entry exit", panics, fuel, err)
			}
		}()
		if panics && caught != "allocation entry context" || !panics && caught != nil {
			t.Fatal("unexpected panic", panics, caught)
		}
		s := fn.Instance.Owner
		if batches, count := s.NativeAllocationStats(); batches != 0 || count != 0 || s.Usage().HeapObjects != 0 {
			t.Fatal("objects published without native entry", batches, count, s.Usage())
		}
		if view, err := s.PrepareNativeAllocations(nativeabi.HeapAllocationBatch); err != nil || view == 0 {
			t.Fatal("pre-entry exit leaked reservation", panics, view, err)
		}
		if err := s.FinishNativeAllocations(fn.Instance.Types); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNativeAllocationPublicationOnTrapAndCancellation(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, canceled := range []bool{false, true} {
		body := module.Expr{}
		for range nativeabi.HeapAllocationBatch - 1 {
			body = append(body, module.Instr{Op: 0x41, Bits: 42}, module.Instr{Op: 0xfb0000}, module.Instr{Op: 0x1a})
		}
		if canceled {
			body = append(body, module.Instr{Op: 3, Block: -64}, module.Instr{Op: 0xc}, module.Instr{Op: 0xb})
		} else {
			body = append(body, module.Instr{Op: 0})
		}
		body = append(body, module.Instr{Op: 0xb})
		d := &module.Module{Types: []module.FuncType{{}, {}}, TypeDefs: []module.TypeDef{{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32}}}, {Kind: module.FunctionKind}}, Functions: []module.Function{{Type: 1, Body: body}}}
		checked, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.GC)
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Compile(checked, lower.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		m, err := slots.Compile(ir, slots.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		e, fn, lease := prepare(t, m)
		if lease.entries[fn.Index-lease.first].allocationBatch != nativeabi.HeapAllocationBatch {
			t.Fatal("constructor was not lowered")
		}
		if canceled {
			if !e.Suspend() {
				t.Fatal("new workspace did not suspend")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := e.Begin(ctx, false); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { _, _, err := e.Run(fn, 0, 1, 0, 0); done <- err }()
			deadline := time.Now().Add(5 * time.Second)
			for {
				if _, err := e.native.State(); errors.Is(err, native.ErrBusy) {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("native allocation entry did not start")
				}
				runtime.Gosched()
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatal("wrong cancellation exit", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("native allocation entry missed cancellation")
			}
		} else if _, _, err := e.Run(fn, 0, 1, 0, 0); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("wrong trap exit", err)
		}
		s := fn.Instance.Owner
		_, count := s.NativeAllocationStats()
		if count > nativeabi.HeapAllocationBatch-1 || !canceled && count != nativeabi.HeapAllocationBatch-1 || s.Usage().HeapObjects != count {
			t.Fatal("exit lost committed allocations", canceled, count, s.Usage())
		}
		for index := uint64(1); index <= nativeabi.HeapAllocationBatch; index++ {
			ref := uint64(1<<63|1<<32) | index
			if s.IsLiveReference(ref) {
				v, err := s.GC(context.Background(), fn.Instance, &module.Instr{Op: 0xfb0002}, []value.Value{{Lo: ref}})
				if err != nil || v.Lo != 42 {
					t.Fatal("committed object not initialized", index, v, err)
				}
			}
		}
		if _, err := s.PrepareNativeAllocations(nativeabi.HeapAllocationBatch); err != nil {
			t.Fatal("exit leaked reservation", err)
		}
		if err := s.FinishNativeAllocations(fn.Instance.Types); err != nil {
			t.Fatal(err)
		}
		s.CollectReferences(nil, nil, nil)
		if s.Usage().HeapObjects != 0 {
			t.Fatal("unrooted exit allocations survived")
		}
	}
}

package codegen

import (
	"encoding/binary"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestVectorConstantRegisterBudgetAndEveryContinuation(t *testing.T) {
	body := module.Expr{{Op: 0x20}}
	for i := uint64(0); i < 20; i++ {
		literal := module.Instr{Op: 0xfd000c}
		binary.LittleEndian.PutUint64(literal.Vector[:8], 0x9e3779b97f4a7c15*(i+1))
		binary.LittleEndian.PutUint64(literal.Vector[8:], 0xfedcba9876543210^(i*0x0101010101010101))
		body = append(body, literal, module.Instr{Op: 0xfd0051})
	}
	body = append(body, module.Instr{Op: 0x21}, module.Instr{Op: 0x20}, module.Instr{Op: 0xb})
	f := functionModule(t, functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.V128}, Results: []wasm.ValueType{wasm.V128}}, 0, body)).Functions[0]
	for _, arch := range []string{"amd64", "arm64"} {
		e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true}
		e.prepareVectorConstants()
		want := 3
		if arch == "arm64" {
			want = 14
		}
		if len(e.vectorConstants) != want {
			t.Fatal(arch, len(e.vectorConstants), want)
		}
	}
	if !native.Available {
		return
	}
	_, fl, fc := functionContext(t, f, true)
	_, sl, sc := functionContext(t, f, false)
	for pc := range f.Code {
		for i := range f.Slots {
			v := value.Value{Lo: 0xdeadbeef12345678 + uint64(i), Hi: 0xfedcba9876543210 - uint64(i)}
			writeSlot(t, fc, 16, slots.Slot(i), v)
			writeSlot(t, sc, 16, slots.Slot(i), v)
		}
		for _, c := range []*native.Context{fc, sc} {
			if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000}); err != nil {
				t.Fatal(err)
			}
		}
		if exit, err := fc.Enter(fl, 0); err != nil || exit != native.Return {
			t.Fatal(pc, exit, err)
		}
		if exit, err := sc.Enter(sl, 0); err != nil || exit != native.Return {
			t.Fatal(pc, exit, err)
		}
		for i := range f.Slots {
			if got, want := readSlot(t, fc, 16, slots.Slot(i)), readSlot(t, sc, 16, slots.Slot(i)); got != want {
				t.Fatal(pc, i, got, want)
			}
		}
	}
}

package nativeexec

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestIntegerModuleIncludesEveryBodyAndColdOperation(t *testing.T) {
	pure := &slots.Function{Slots: []wasm.ValueType{wasm.I32}, Code: []slots.Instruction{
		{Op: 0x41}, {Op: 0xad}, {Op: 0xfc000a}, {Op: 0xfe001e},
		{Op: slots.ControlBase + uint32(ir.Return)},
	}}
	if !integerModule(&slots.Module{Functions: []*slots.Function{pure}}) {
		t.Fatal("integer, bulk and atomic operations need no FP transition")
	}
	if integerModule(nil) || integerModule(&slots.Module{Functions: []*slots.Function{nil}}) {
		t.Fatal("missing proof admitted")
	}
	// An otherwise unused body still belongs to the linked image. Scalar
	// types in signatures or storage reject even if the body is empty.
	for _, typ := range []wasm.ValueType{wasm.F32, wasm.F64, wasm.V128} {
		for place := 0; place < 3; place++ {
			other := &slots.Function{}
			switch place {
			case 0:
				other.Slots = []wasm.ValueType{typ}
			case 1:
				other.Signature.Params = []wasm.ValueType{typ}
			case 2:
				other.Signature.Results = []wasm.ValueType{typ}
			}
			if integerModule(&slots.Module{Functions: []*slots.Function{pure, other}}) {
				t.Fatal("mixed image admitted", typ, place)
			}
		}
	}
	for _, op := range []uint32{0x43, 0x44, 0x2a, 0x39, 0x5b, 0x8b, 0xa8, 0xbb,
		0xfd000c, 0xfd006f, 0xfc0000, 0xfb0000, 0xe0, 0x19, 0xfe0004, 0xfe004f, ^uint32(0)} {
		cold := &slots.Function{Code: []slots.Instruction{{Op: slots.ControlBase + uint32(ir.Return)}, {Op: op}}}
		if integerModule(&slots.Module{Functions: []*slots.Function{pure, cold}}) {
			t.Fatalf("FP, SIMD or unknown cold operation %x admitted", op)
		}
	}
}

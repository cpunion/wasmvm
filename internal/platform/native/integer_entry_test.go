package native

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Integer source operations can use SIMD instructions internally (for example,
// ARM popcount). Check actual generated code under a poisoned environment,
// including baseline and detected selectors, rather than relying on types alone.
func TestGeneratedIntegerSelectorsPreserveFloatingState(t *testing.T) {
	if !Available {
		return
	}
	for _, config := range []codegen.Config{{Unmetered: true}, {
		Unmetered: true, X86Features: codegen.X86Features(CPUFeatures()), ARM64Features: codegen.ARM64Features(ARM64CPUFeatures()),
	}} {
		for op := uint32(0x45); op <= 0x8a; op++ {
			if op > 0x5a && op < 0x67 {
				continue
			}
			t.Run(fmt.Sprintf("op=%x/x86=%x/arm=%x", op, config.X86Features, config.ARM64Features), func(t *testing.T) {
				input := wasm.I32
				if op >= 0x50 && op <= 0x5a || op >= 0x79 {
					input = wasm.I64
				}
				output := input
				if op <= 0x5a {
					output = wasm.I32
				}
				b := slots.Slot(1)
				if op == 0x45 || op == 0x50 || op >= 0x67 && op <= 0x69 || op >= 0x79 && op <= 0x7b {
					b = slots.NoSlot
				}
				f := &slots.Function{LocalCount: 2, Slots: []wasm.ValueType{input, input, output},
					Signature: module.FuncType{Params: []wasm.ValueType{input, input}, Results: []wasm.ValueType{output}},
					Sources:   []ir.Source{{Wasm: module.Instr{Op: op}}, {Wasm: module.Instr{Op: 0xb}}},
					Args:      []slots.Slot{2}, Code: []slots.Instruction{
						{Op: op, Dst: 2, A: 0, B: b, C: slots.NoSlot},
						{Op: slots.ControlBase + uint32(ir.Return), Source: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
					}}
				compiled, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, f, codegen.DefaultLimits(), config)
				if err != nil {
					t.Fatal(err)
				}
				code, err := PublishIntegerCode(compiled.Entry.Code, []uint32{0}, 1<<20)
				if err != nil {
					t.Fatal(err)
				}
				defer code.Close()
				lease, err := code.Acquire()
				if err != nil {
					t.Fatal(err)
				}
				defer lease.Close()
				ctx, err := NewContext(48, 48, 1)
				if err != nil {
					t.Fatal(err)
				}
				defer ctx.Close()
				ctx.record.Depth = 1
				words := unsafe.Slice((*uint64)(ctx.arena), 6)
				words[0], words[2] = 0x12345678, 3
				entry, err := lease.begin(0)
				if err != nil {
					t.Fatal(err)
				}
				defer lease.end()
				if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), Return) {
					t.Fatal("generated integer selector changed host FP state")
				}
			})
		}
	}
}

func TestIntegerEntryPreservesAmbientFloatingStateAndAdmission(t *testing.T) {
	if !Available {
		return
	}
	// This read-only diagnostic canary snapshots the ambient environment;
	// it is deliberately not an input to the compiler's integer proof.
	var body []byte
	if runtime.GOARCH == "amd64" {
		body = []byte{0x0f, 0xae, 0x5f, 0x60, 0xd9, 0x7f, 0x68, 0x8b, 0x47, 0x70, 0xc3}
	} else {
		body = armWords(0xd53b4401, 0xf9003001, 0xd53b4421, 0xf9003401, 0xb9407000, 0xd65f03c0)
	}
	code, err := PublishIntegerCode(body, []uint32{0}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer code.Close()
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	ctx, err := NewContext(0, 16, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	ctx.record.Payload[0] = uint64(Return)
	if status, err := ctx.Enter(lease, 0); status != Return || err != nil {
		t.Fatal(status, err)
	}
	entry, err := lease.begin(0)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.end()
	for status := Exit(0); status <= Move; status++ {
		ctx.record.Payload[0] = uint64(status)
		if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), status) {
			t.Fatal("ambient FP environment changed", status)
		}
		if ctx.record.ResultLo == 0 || runtime.GOARCH == "amd64" && ctx.record.ResultLo&0xe000 == 0 {
			t.Fatal("integer entry normalized the diagnostic host state")
		}
	}
	ctx.record.Version = 0
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid version admitted")
	}
	ctx.record.Version = ABIVersion
	ctx.record.Used = ctx.record.Capacity + 1
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid arena admitted")
	}
	ctx.record.Used = 0
	ctx.record.Depth = ctx.record.MaxDepth + 1
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), 0) {
		t.Fatal("invalid depth admitted")
	}
	ctx.record.Depth = 0
	ctx.Cancel()
	if !probeIntegerFP(entry, unsafe.Pointer(ctx.record), Poll) {
		t.Fatal("cancelled entry admitted")
	}
}

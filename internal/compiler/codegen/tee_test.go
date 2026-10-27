package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTeeForwardingRetainsIndependentFrames(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, tc := range []struct {
		typ, result wasm.ValueType
		op          uint32
	}{{wasm.I32, wasm.I32, 0x49}, {wasm.I32, wasm.I32, 0x73},
		{wasm.I64, wasm.I64, 0x85}, {wasm.F32, wasm.F32, 0x92},
		{wasm.F64, wasm.F64, 0xa0}, {wasm.V128, wasm.V128, 0xfd0051}} {
		for _, mode := range []struct{ clobber, boundary bool }{{}, {clobber: true}, {boundary: true}} {
			clobber := mode.clobber
			body := module.Expr{{Op: 0x41}, {Op: 0x28}, {Op: 0x1a},
				{Op: 0x20}, {Op: 0x22}}
			if clobber {
				body = append(body, module.Instr{Op: 0x20, A: 1}, module.Instr{Op: 0x21})
			}
			body = append(body, module.Instr{Op: 0x20, A: 1}, module.Instr{Op: tc.op},
				module.Instr{Op: 0x41}, module.Instr{Op: 0x28}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
			m := functionFixture(module.FuncType{Params: []wasm.ValueType{tc.typ, tc.typ}, Results: []wasm.ValueType{tc.result}}, 0, body)
			m.Memories = []module.Limits{{Min: 1}}
			f := functionModule(t, m).Functions[0]
			padding := 0
			if mode.boundary {
				for pc, in := range f.Code {
					if in.Op == 0x22 {
						padding = 255 - pc
						break
					}
				}
				source := uint32(len(f.Sources))
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 1}})
				prefix := make([]slots.Instruction, padding)
				for i := range prefix {
					prefix[i] = slots.Instruction{Op: 1, Source: source, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
				}
				f.Code = append(prefix, f.Code...)
				for i := range f.Edges {
					f.Edges[i].PC += uint32(padding)
				}
				for i := range f.BlockPC {
					f.BlockPC[i] += uint32(padding)
				}
			}
			_, fast, fc := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
			_, slow, sc := functionContext(t, f, false)
			bindMemory(t, fc, 65536, 65536)
			bindMemory(t, sc, 65536, 65536)
			for pc := range f.Code {
				if pc > 0 && pc < padding {
					continue // Padding only positions the tee at the poll boundary.
				}
				for i, typ := range f.Slots {
					v := value.Value{Lo: 0x123456789abcdef0 + uint64(i)*7919}
					if typ == wasm.I32 || typ == wasm.F32 {
						v.Lo = uint64(uint32(v.Lo))
					} else if typ == wasm.V128 {
						v.Hi = ^v.Lo
					}
					writeSlot(t, fc, 16, slots.Slot(i), v)
					writeSlot(t, sc, 16, slots.Slot(i), v)
				}
				s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000}
				if err := fc.SetState(s); err != nil {
					t.Fatal(err)
				}
				if err := sc.SetState(s); err != nil {
					t.Fatal(err)
				}
				got, err := fc.Enter(fast, 0)
				want, refErr := sc.Enter(slow, 0)
				if err != nil || refErr != nil || got != want {
					t.Fatal(tc, clobber, pc, got, want, err, refErr)
				}
				for i := range f.Slots {
					if got, want := readSlot(t, fc, 16, slots.Slot(i)), readSlot(t, sc, 16, slots.Slot(i)); got != want {
						t.Fatal(tc, clobber, pc, i, got, want)
					}
				}
			}
		}
	}
}

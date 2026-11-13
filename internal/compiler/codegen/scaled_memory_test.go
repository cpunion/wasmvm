package codegen

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func scaledMemoryFixture(op, arithmetic uint32, factor uint64, offset uint32) *slots.Function {
	typ := wasm.I32
	if op == 0x29 || op >= 0x30 {
		typ = wasm.I64
	} else if op == 0x2a {
		typ = wasm.F32
	} else if op == 0x2b {
		typ = wasm.F64
	}
	f := &slots.Function{LocalCount: 1, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, typ}}
	emit := func(op uint32, dst, a, b slots.Slot, bits uint64) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: bits, B: offset, Offset: uint64(source) + 100}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, 255)
	emit(0x71, 1, 0, 2, 0)
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, factor)
	emit(arithmetic, 3, 1, 2, 0)
	emit(op, 4, 3, slots.NoSlot, 0)
	// Both intermediates remain observable on cold entry, but normal execution
	// overwrites them after the nontrapping load. Return every original word.
	emit(0x41, 3, slots.NoSlot, slots.NoSlot, 19)
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, 23)
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Slots))
	for i, typ := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
		f.Signature.Results = append(f.Signature.Results, typ)
	}
	return f
}

func TestScaledMemoryProofAndSelectionBoundaries(t *testing.T) {
	for _, mode := range []string{"multiply", "shift", "modulo-shift", "overflow", "zero", "one", "non-power", "large", "live-product", "live-literal", "remapped", "unproven", "region-end", "arm64", "metered", "stack-index"} {
		f := scaledMemoryFixture(0x28, 0x6c, 4, 0)
		want := true
		e := functionEmitter{f: f, a: &assembler{arch: "amd64"}, unmetered: true, usesMemory: true, memoryMinimum: 65536,
			registers: map[slots.Slot]byte{0: 3, 1: 12, 2: 13, 3: 14, 4: 15}}
		switch mode {
		case "shift", "modulo-shift":
			f.Code[3].Op, f.Sources[3].Wasm.Op = 0x74, 0x74
			f.Sources[2].Wasm.Bits = 2
			if mode == "modulo-shift" {
				f.Sources[2].Wasm.Bits = 34
			}
		case "overflow":
			f.Sources[0].Wasm.Bits, want = 0xffffffff, false
		case "zero", "one", "non-power", "large":
			f.Sources[2].Wasm.Bits = map[string]uint64{"zero": 0, "one": 1, "non-power": 3, "large": 16}[mode]
			want = false
		case "live-product", "live-literal":
			pc := 5
			if mode == "live-literal" {
				pc = 6
			}
			f.Code[pc].Op, f.Code[pc].Dst, want = 1, slots.NoSlot, false
		case "arm64":
			e.a.arch, want = "arm64", false
		case "metered":
			e.unmetered, want = false, false
		case "stack-index":
			delete(e.registers, 1)
		}
		r := e.memoryRegions()[0]
		e.regionEnd, e.regionLimit = r.end+1, r.end+1
		e.memoryScales, e.provenPCs = r.scales, r.proven
		switch mode {
		case "remapped":
			e.addresses, want = map[int]memoryAddress{4: {local: 0}}, false
		case "unproven":
			e.provenPCs, want = nil, false
		case "region-end":
			e.regionLimit, want = 4, false
		}
		end, selected := e.combineScaledMemory(2)
		if selected != want || selected && (end != 4 || e.directMemoryOperands != 1) {
			t.Fatal(mode, selected, end, e.directMemoryOperands, want)
		}
	}
}

func TestScaledMemoryEveryColdFrameAndLoadWidth(t *testing.T) {
	for op := uint32(0x28); op <= 0x35; op++ {
		for _, arithmetic := range []uint32{0x6c, 0x74} {
			for shift := uint64(1); shift <= 3; shift++ {
				factor := uint64(1) << shift
				if arithmetic == 0x74 {
					factor = shift + 32
				}
				for _, offset := range []uint32{0, 127, 128, 4095} {
					f := scaledMemoryFixture(op, arithmetic, factor, offset)
					for _, arch := range []string{"amd64", "arm64"} {
						if _, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1}); err != nil {
							t.Fatal(arch, op, arithmetic, shift, offset, err)
						}
					}
					if !native.Available {
						continue
					}
					_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
					_, refLease, ref := configuredContext(t, f, Config{Unmetered: true})
					memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
					initial := make([]byte, 65536)
					for i := range initial {
						initial[i] = byte(i*37 + 193)
					}
					memory.Write(0, initial)
					refMemory.Write(0, initial)
					for pc := range f.Code {
						for _, raw := range []uint64{0, 255, 0xdeadbeef00000042, ^uint64(0)} {
							for slot := range f.Slots {
								bits := value.Value{Lo: raw ^ uint64(slot)}
								if f.Slots[slot] == wasm.F32 {
									bits.Lo = uint64(uint32(bits.Lo))
								}
								writeSlot(t, ctx, 16, slots.Slot(slot), bits)
								writeSlot(t, ref, 16, slots.Slot(slot), bits)
							}
							state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 10000}
							wantState := state
							got, err := ctx.RunState(lease, 0, &state)
							want, refErr := ref.RunState(refLease, 0, &wantState)
							if err != nil || refErr != nil || got != want || state != wantState {
								t.Fatal(op, arithmetic, shift, offset, pc, raw, got, want, state, wantState, err, refErr)
							}
							for slot := range f.Slots {
								if a, b := readSlot(t, ctx, 16, slots.Slot(slot)), readSlot(t, ref, 16, slots.Slot(slot)); a != b {
									t.Fatal("scaled cold frame", op, arithmetic, shift, offset, pc, raw, slot, a, b)
								}
							}
						}
					}
					got, want := make([]byte, 65536), make([]byte, 65536)
					memory.Read(0, got)
					refMemory.Read(0, want)
					if !bytes.Equal(got, want) || !bytes.Equal(got, initial) {
						t.Fatal("scaled load changed memory")
					}
				}
			}
		}
	}
}

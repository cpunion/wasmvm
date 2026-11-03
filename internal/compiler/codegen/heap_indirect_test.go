package codegen

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeHeapIndirectHandlesReplacementAndColdEntries(t *testing.T) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.AnyRef, wasm.I32, wasm.AnyRef, wasm.AnyRef, wasm.AnyRef, wasm.AnyRef, wasm.I32}, LocalCount: 7}
	get := func(a, dst slots.Slot) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0xfb000b, Offset: 100 + uint64(len(f.Code))}})
		in := slots.Instruction{Op: 0xfb000b, Source: source, A: a, B: 1, C: slots.NoSlot, Dst: dst, Extra: uint32(len(f.Args)), Count: 2}
		f.Args = append(f.Args, a, 1)
		f.Code = append(f.Code, in)
	}
	move := func(src, dst slots.Slot) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(src)}})
		f.Code = append(f.Code, slots.Instruction{Op: 0x20, Source: source, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: dst})
	}
	get(0, 2)
	get(2, 3)
	get(0, 2)
	get(2, 3)
	move(4, 0)
	get(0, 5)
	get(2, 5)
	move(6, 1)
	get(2, 5)
	get(3, 5)
	f.Sources = append(f.Sources, ir.Source{})
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: uint32(len(f.Sources) - 1), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Extra: uint32(len(f.Args)), Count: 1})
	f.Args = append(f.Args, 5)
	for _, fast := range []bool{false, true} {
		config := Config{Unmetered: fast, HeapTypes: []module.TypeDef{{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.AnyRef}}}}}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil {
				t.Fatal(err)
			}
			want := uint32(0)
			if fast && arch == "arm64" {
				want = 1
			}
			if c.HeapCaches != want {
				t.Fatal("heap cache policy", arch, fast, c.HeapCaches)
			}
		}
		if !native.Available {
			continue
		}
		_, lease, ctx := configuredContext(t, f, config)
		storage, err := native.NewMemory(512, 512)
		if err != nil {
			t.Fatal(err)
		}
		defer storage.Close()
		buffer, _ := storage.Buffer()
		address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
		words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
		first, second, third := uint64(0x8000000100000001), uint64(0x8000000100000002), uint64(0x8000000100000003)
		initial := []value.Value{{Lo: first}, {}, {Lo: second}, {Lo: first}, {Lo: third}, {Lo: first}, {}}
		enter := func(pc int, frame []value.Value) native.Exit {
			for i, v := range frame {
				writeSlot(t, ctx, 16, slots.Slot(i), v)
			}
			s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64}
			s.Payload[2] = address
			if err := ctx.SetState(s); err != nil {
				t.Fatal(err)
			}
			exit, err := ctx.Enter(lease, 0)
			if err != nil {
				t.Fatal(err)
			}
			return exit
		}
		simulate := func(pc, end int, frame []value.Value, fields map[uint64]uint64) []value.Value {
			out := append([]value.Value(nil), frame...)
			for _, in := range f.Code[pc:end] {
				if in.Op == 0xfb000b {
					out[in.Dst] = value.Value{Lo: fields[out[in.A].Lo]}
				}
				if in.Op == 0x20 {
					out[in.Dst] = out[f.Sources[in.Source].Wasm.A]
				}
			}
			return out
		}
		for layout := 0; layout < 2; layout++ {
			fields := map[uint64]uint64{first: second, second: first, third: first}
			if layout == 1 {
				fields[first], fields[third] = third, second
			}
			words[0], words[1] = 4, address+32
			for i, ref := range []uint64{first, second, third} {
				entry, payload := 8+i*4, 256+((i+layout)%3)*32
				words[entry], words[entry+1], words[entry+2] = ref, address+uint64(payload), 1
				words[payload/8] = fields[ref]
			}
			for pc := range f.Code {
				want := simulate(pc, len(f.Code), initial, fields)
				if exit := enter(pc, initial); exit != native.Return {
					t.Fatal("cold entry", runtime.GOARCH, fast, layout, pc, exit)
				}
				for i, v := range want {
					if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
						t.Fatal("cold frame", fast, layout, pc, i, got, v)
					}
				}
			}
			for _, bad := range []struct {
				ref, index uint64
				exit       native.Exit
				trap       wasm.TrapCode
				pc         int
			}{
				{0, 0, native.Trap, wasm.TrapNullReference, 5},
				{first + 1<<32, 0, native.BadState, 0, 5},
				{first &^ (1 << 63), 0, native.BadState, 0, 5},
				{third + 1, 0, native.BadState, 0, 5},
				{third, 1, native.Trap, wasm.TrapArrayBounds, 8},
				{third, 0xffffffff, native.Trap, wasm.TrapArrayBounds, 8},
			} {
				input := append([]value.Value(nil), initial...)
				input[4].Lo, input[6].Lo = bad.ref, bad.index
				want := simulate(0, bad.pc, input, fields)
				if exit := enter(0, input); exit != bad.exit {
					t.Fatal("warm invalid input", fast, bad, exit)
				}
				state, _ := ctx.State()
				if bad.trap != 0 && (state.Payload[0] != uint64(bad.trap) || state.Source != 100+uint64(bad.pc)) {
					t.Fatal("trap source", bad, state)
				}
				// BadState rejects an invalid internal frame and is terminal.
				// Typed traps publish the complete logical frame for inspection.
				if bad.exit == native.Trap {
					for i, v := range want {
						if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
							t.Fatal("trap frame", fast, layout, bad, i, got, v)
						}
					}
				}
			}
		}
	}
}

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

func indirectCaller() (*slots.Function, Config) {
	a, b := leafFixture(0x6a, nil), leafFixture(0x73, nil)
	f := leafCaller(a, 2, 5)
	f.Code[5].Op, f.Code[5].Count = 0x11, 3
	f.Sources[5].Wasm.Op, f.Sources[5].Wasm.A = 0x11, 0
	f.Sources[5].Inputs = append(append([]wasm.ValueType(nil), a.Signature.Params...), wasm.I32)
	f.Args = []slots.Slot{0, 1, 4, 2, 0, 1, 2, 3, 4}
	f.Code[6].Extra = 4
	return f, Config{Unmetered: true, FunctionCount: 9,
		DirectCallees: map[uint32]*slots.Function{7: a, 8: b}, IndirectCallees: map[uint32][]uint32{0: {7, 8}}}
}

func TestIndirectExpansionReadsLiveTargetsAndPreservesColdEntries(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, constants := range []int{0, 5, 8} {
		for _, table := range []uint32{0, 3000} {
			f, config := indirectCaller()
			// Exercise full, partial and exhausted metadata-register capacity.
			ret := f.Code[len(f.Code)-1]
			f.Code = f.Code[:len(f.Code)-1]
			for i := 0; i < constants; i++ {
				dst, source := len(f.Slots), len(f.Sources)
				f.Slots = append(f.Slots, wasm.I64)
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x42, Bits: 0x123456789abcdef0 + uint64(i)}})
				f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: uint32(source), Dst: slots.Slot(dst), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
			}
			f.Code = append(f.Code, ret)
			f.Sources[5].Wasm.B = table
			compiled, lease, ctx := configuredContext(t, f, config)
			if compiled.InlinedCalls != 1 || !compiled.UsesRuntimeData {
				t.Fatal("indirect call not expanded")
			}
			data, err := native.NewMemory(uint64(9+2*(table+1))*8, uint64(9+2*(table+1))*8)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(data.Close)
			buffer, _ := data.Buffer()
			words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
			target, _ := native.NewMemory(8, 8)
			t.Cleanup(target.Close)
			contents, _ := target.Buffer()
			entry := (*uint64)(unsafe.Pointer(unsafe.SliceData(contents)))
			words[7], words[8] = 1007, 1008
			words[9+table*2], words[10+table*2] = 1, uint64(uintptr(unsafe.Pointer(entry)))
			for _, ref := range []uint64{1007, 1008, 0, 1009} {
				*entry = ref
				for _, index := range []uint64{0, 1, 0xffffffff} {
					for pc := range f.Code {
						initial := []value.Value{{Lo: 42}, {Lo: 17}, {Lo: 3}, {Lo: 5}, {Lo: index}}
						for i, v := range initial {
							writeSlot(t, ctx, 16, slots.Slot(i), v)
						}
						s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
						s.Payload[2] = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
						if err := ctx.SetState(s); err != nil {
							t.Fatal(err)
						}
						want := native.Return
						if pc <= 5 {
							if index != 0 || ref != 1007 && ref != 1008 {
								want = native.Call
							} else if ref == 1007 {
								initial[2].Lo = 59
							} else {
								initial[2].Lo = 42 ^ 17
							}
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != want {
							t.Fatal(table, ref, index, pc, exit, err)
						}
						for i, v := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
								t.Fatal(table, ref, index, pc, i, got, v)
							}
						}
					}
				}
			}
			if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: 5}); err != nil {
				t.Fatal(err)
			}
			if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Call {
				t.Fatal("missing view must retain checked call", exit, err)
			}
		}
	}
}

func TestIndirectExpansionRejectsIneligibleTargets(t *testing.T) {
	for _, kind := range []string{"count", "offset", "identity", "missing", "result-count", "result-type", "pressure", "metered"} {
		f, config := indirectCaller()
		switch kind {
		case "count":
			config.FunctionCount = 0
		case "offset":
			f.Sources[5].Wasm.B = ^uint32(0)
		case "identity":
			config.FunctionCount = 7
		case "missing":
			config.DirectCallees = nil
		case "result-count":
			for _, c := range config.DirectCallees {
				c.Signature.Results = nil
			}
		case "result-type":
			for _, c := range config.DirectCallees {
				c.Signature.Results = []wasm.ValueType{wasm.I64}
			}
		case "pressure":
			for _, c := range config.DirectCallees {
				c.Signature.Params = nil
			}
		case "metered":
			config.Unmetered = false
		}
		c, err := CompileFunctionWithConfig(runtime.GOARCH, f, DefaultLimits(), config)
		if err != nil || c.InlinedCalls != 0 {
			t.Fatal(kind, c, err)
		}
	}
}

package codegen

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestComparisonBranchFlagsAndColdEntries(t *testing.T) {
	for op := uint32(0x45); op <= 0x5a; op++ {
		for copies := 0; copies < 4; copies++ {
			for _, reverse := range []bool{false, true} {
				typ := wasm.I32
				if integerWide(op) {
					typ = wasm.I64
				}
				yes, no := uint32(2), uint32(3)
				if reverse {
					yes, no = no, yes
				}
				f := &slots.Function{Slots: []wasm.ValueType{typ, typ, wasm.I32, wasm.I32, wasm.I32}, Sources: []ir.Source{{}}, Args: []slots.Slot{3, 4},
					Moves: []slots.Move{{Dst: 3, Src: 2}, {Dst: 4, Src: 2}}, Edges: []slots.Edge{{PC: yes, Count: uint32(copies & 1)}, {PC: no, Start: 1, Count: uint32(copies >> 1)}},
					Code: []slots.Instruction{
						{Op: op, Dst: 2, A: 0, B: 1, C: slots.NoSlot},
						{Op: slots.ControlBase + uint32(ir.If), Dst: slots.NoSlot, A: 2, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
						{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
						{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1, Extra: 1},
					}}
				for _, arch := range []string{"arm64", "amd64"} {
					c, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
					if err != nil || (c.ForwardedPredicates == 1) != (arch == "arm64") {
						t.Fatal(arch, op, err)
					}
				}
				if !native.Available {
					continue
				}
				compiled, lease, ctx := functionContext(t, f, true)
				for _, a := range []uint64{0, 1, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)} {
					for _, b := range []uint64{0, 7, 1 << 63, ^uint64(0)} {
						cond, trap := numeric.Eval(op, a, b)
						if trap != wasm.NoTrap {
							t.Fatal(trap)
						}
						out, exit, _ := runFunction(t, compiled, lease, ctx, []value.Value{{Lo: a}, {Lo: b}, {Lo: 99}, {Lo: 11}, {Lo: 22}}, 0, false, 0)
						if exit != native.Return || len(out) != 1 {
							t.Fatal(exit, out)
						}
						selected, w3, w4 := no, uint64(11), uint64(22)
						if cond != 0 {
							selected = yes
							if copies&1 != 0 {
								w3 = cond
							}
						} else if copies&2 != 0 {
							w4 = cond
						}
						want := w3
						if selected == 3 {
							want = w4
						}
						if out[0].Lo != want || readSlot(t, ctx, 16, 2).Lo != cond || readSlot(t, ctx, 16, 3).Lo != w3 || readSlot(t, ctx, 16, 4).Lo != w4 {
							t.Fatal(op, copies, reverse, a, b, out, want)
						}
					}
				}
				// Direct control entry uses the retained boolean, deliberately
				// independent of the operands and comparison flags from entry checks.
				for _, cond := range []uint64{0, 1, 1 << 32, 0xffffffff} {
					writeSlot(t, ctx, 16, 2, value.Value{Lo: cond})
					writeSlot(t, ctx, 16, 3, value.Value{Lo: 11})
					writeSlot(t, ctx, 16, 4, value.Value{Lo: 22})
					if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: 1}); err != nil {
						t.Fatal(err)
					}
					if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
						t.Fatal(exit, err)
					}
					s, _ := ctx.State()
					selected := no
					if uint32(cond) != 0 {
						selected = yes
					}
					if s.Payload[0] != uint64(selected) {
						t.Fatal("cold control inferred flags", op, cond, s)
					}
				}
				ctx.Close()
				lease.Close()
			}
		}
	}
}

func TestComparisonFlagsPollPublishesBoolean(t *testing.T) {
	if !native.Available {
		return
	}
	for _, padding := range []int{0, 254} {
		f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32}, Sources: []ir.Source{{}}, Args: []slots.Slot{2}, Edges: []slots.Edge{{PC: 3}, {PC: 3}}, Code: []slots.Instruction{
			{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x49, Dst: 2, A: 0, B: 1, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.If), Dst: slots.NoSlot, A: 2, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
			{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
		}}
		if padding != 0 {
			prefix := make([]slots.Instruction, padding)
			for i := range prefix {
				prefix[i] = slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
			}
			f.Code = append(prefix, f.Code...)
			for i := range f.Edges {
				f.Edges[i].PC += uint32(padding)
			}
		}
		compiled, err := CompileUnmeteredFunction(runtime.GOARCH, f, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		// Publish cancellation after the adapter check but before the comparison.
		// The comparison reaches its original branch poll, including PC 256.
		var prefix []byte
		if runtime.GOARCH == "arm64" {
			for _, word := range []uint32{0xd2800029, 0x9100000a | uint32(nativeabi.ControlOffset)<<10, 0xc89ffd49} {
				prefix = binary.LittleEndian.AppendUint32(prefix, word)
			}
		} else {
			prefix = []byte{0x48, 0xc7, 0x87}
			prefix = binary.LittleEndian.AppendUint32(prefix, uint32(nativeabi.ControlOffset))
			prefix = binary.LittleEndian.AppendUint32(prefix, 1)
		}
		code, err := native.Publish(append(prefix, compiled.Entry.Code...), []uint32{0}, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		defer code.Close()
		lease, err := code.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		defer lease.Close()
		for _, a := range []uint64{0, 9} {
			ctx, err := native.NewContext(64, 64, 1)
			if err != nil {
				t.Fatal(err)
			}
			defer ctx.Close()
			writeSlot(t, ctx, 16, 0, value.Value{Lo: a})
			writeSlot(t, ctx, 16, 1, value.Value{Lo: 7})
			writeSlot(t, ctx, 16, 2, value.Value{Lo: 99})
			if err = ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(1 + padding)}); err != nil {
				t.Fatal(err)
			}
			if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Poll {
				t.Fatal(exit, err)
			}
			s, err := ctx.State()
			if err != nil || s.Continuation != uint64(2+padding) {
				t.Fatal(s, err)
			}
			want := uint64(0)
			if a < 7 {
				want = 1
			}
			if got := readSlot(t, ctx, 16, 2).Lo; got != want {
				t.Fatal(got, want)
			}
		}
	}
}

func TestComparisonBranchAliasedInputs(t *testing.T) {
	for op := uint32(0x45); op <= 0x5a; op++ {
		for _, dst := range []slots.Slot{0, 1} {
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I32, wasm.I32}, Sources: []ir.Source{{}}, Args: []slots.Slot{2, 3}, Edges: []slots.Edge{{PC: 2}, {PC: 3}},
				Code: []slots.Instruction{
					{Op: op, Dst: dst, A: 0, B: 1, C: slots.NoSlot},
					{Op: slots.ControlBase + uint32(ir.If), Dst: slots.NoSlot, A: dst, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
					{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
					{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1, Extra: 1},
				}}
			for _, arch := range []string{"amd64", "arm64"} {
				compiled, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
				if err != nil || (compiled.ForwardedPredicates == 1) != (arch == "arm64") {
					t.Fatal(op, dst, arch, err)
				}
			}
			if !native.Available {
				continue
			}
			compiled, lease, ctx := functionContext(t, f, true)
			for _, a := range []uint64{0, 1, 9, 0xffffffff, 1 << 63, ^uint64(0)} {
				for _, b := range []uint64{0, 7, 1 << 32, 1 << 63} {
					condition, trap := numeric.Eval(op, a, b)
					if trap != wasm.NoTrap {
						t.Fatal(trap)
					}
					out, exit, _ := runFunction(t, compiled, lease, ctx, []value.Value{{Lo: a}, {Lo: b}, {Lo: 11}, {Lo: 22}}, 0, false, 0)
					want := uint64(22)
					if condition != 0 {
						want = 11
					}
					if exit != native.Return || out[0].Lo != want || readSlot(t, ctx, 16, dst).Lo != condition {
						t.Fatal(op, dst, a, b, out, condition)
					}
				}
			}
			ctx.Close()
			lease.Close()
		}
	}
}

package codegen

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func predicateLoopFixture(op uint32, alias int, pressure, entryRead bool) (*slots.Function, int) {
	typ, xor, add := wasm.I32, uint32(0x73), uint32(0x6a)
	if op >= 0x50 {
		typ, xor, add = wasm.I64, 0x85, 0x7c
	}
	f := &slots.Function{Slots: []wasm.ValueType{typ, typ, typ, wasm.I32, typ, typ}, Args: []slots.Slot{3}, Signature: module.FuncType{Results: []wasm.ValueType{wasm.I32}}}
	emit := func(op uint32, dst, a, b slots.Slot, local uint32) {
		s := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Offset: uint64(1000 + s)}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: s, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	emit(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	if entryRead {
		emit(0x6a, 5, 5, 3, 0)
	}
	emit(xor, 0, 0, 1, 0)
	if pressure {
		// More simultaneously preserved words than either integer bank can
		// hold. The comparison operands have lower scores than these updates.
		for range 44 {
			s := slots.Slot(len(f.Slots))
			f.Slots = append(f.Slots, typ)
			for range 3 {
				emit(add, s, s, 1, 0)
			}
		}
	}
	dst, a, b := slots.Slot(3), slots.Slot(0), slots.Slot(2)
	if alias != 0 {
		source := a
		if alias == 2 {
			source = b
		}
		emit(0x21, slots.NoSlot, source, slots.NoSlot, 4)
		dst = 4
		if alias == 1 {
			a = 4
		} else {
			b = 4
		}
		f.Args[0] = dst
	}
	if op == 0x45 || op == 0x50 {
		b = slots.NoSlot
	}
	emit(op, dst, a, b, 0)
	branch := len(f.Code)
	emit(slots.ControlBase+uint32(ir.If), slots.NoSlot, dst, slots.NoSlot, 0)
	f.Code[branch].Count = 2
	f.Edges = []slots.Edge{{PC: 1}, {PC: uint32(branch + 1)}}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = 1
	f.LocalCount = uint32(len(f.Slots))
	return f, branch
}

// The shared semantic evaluator is independent of native comparison encoding.
// Select a two-state cursor with opposite results for every signed/unsigned op.
func predicateLoopInputs(op uint32, back bool) (initial, mask, rhs uint64) {
	values := []uint64{0, 1, 2, 0x7fffffff, 0x80000000, 0xffffffff, 1 << 63, ^uint64(0)}
	for _, rhs := range values {
		for _, take := range values {
			x, _ := numeric.Eval(op, take, rhs)
			if (x != 0) != back {
				continue
			}
			for _, leave := range values {
				y, _ := numeric.Eval(op, leave, rhs)
				if (y != 0) != back {
					return leave, take ^ leave, rhs
				}
			}
		}
	}
	panic("comparison has no opposite test inputs")
}

func modelPredicateLoop(f *slots.Function, initial []value.Value, start, branch int, poll bool) ([]value.Value, native.Exit, uint64) {
	words := append([]value.Value(nil), initial...)
	for pc, steps := start, 0; steps < len(f.Code)*8; steps++ {
		if poll && pc == branch {
			return words, native.Poll, uint64(pc)
		}
		in := f.Code[pc]
		a, b := uint64(0), uint64(0)
		if in.A != slots.NoSlot {
			a = words[in.A].Lo
		}
		if in.B != slots.NoSlot {
			b = words[in.B].Lo
		}
		switch {
		case in.Op == 1:
		case in.Op == 0x21:
			words[f.Sources[in.Source].Wasm.A].Lo = a
		case in.Op == 0x73:
			words[in.Dst].Lo = uint64(uint32(a) ^ uint32(b))
		case in.Op == 0x85:
			words[in.Dst].Lo = a ^ b
		case in.Op == 0x6a:
			words[in.Dst].Lo = uint64(uint32(a) + uint32(b))
		case in.Op == 0x7c:
			words[in.Dst].Lo = a + b
		case in.Op >= 0x45 && in.Op <= 0x5a:
			words[in.Dst].Lo, _ = numeric.Eval(in.Op, a, b)
		case in.Op == slots.ControlBase+uint32(ir.If):
			edge := 1
			if uint32(a) != 0 {
				edge = 0
			}
			pc = int(f.Edges[edge].PC)
			continue
		case in.Op == slots.ControlBase+uint32(ir.Return):
			return words, native.Return, uint64(pc)
		default:
			panic("unmodeled predicate-loop instruction")
		}
		pc++
	}
	panic("predicate-loop fixture did not terminate")
}

func TestLoopPredicateCompleteFramesColdPCsAndPoll(t *testing.T) {
	for op := uint32(0x45); op <= 0x5a; op++ {
		for _, back := range []bool{false, true} {
			for _, mode := range []struct {
				name                        string
				alias                       int
				pressure, header, entryRead bool
			}{{name: "resident"}, {name: "left-alias", alias: 1}, {name: "right-alias", alias: 2}, {name: "pressure", pressure: true}, {name: "header", header: true}, {name: "entry-read", entryRead: true}} {
				t.Run(fmt.Sprintf("%x/%t/%s", op, back, mode.name), func(t *testing.T) {
					f, branch := predicateLoopFixture(op, mode.alias, mode.pressure, mode.entryRead)
					if mode.header {
						f.Edges[0].PC = 0
					}
					if !back {
						f.Edges[0], f.Edges[1] = f.Edges[1], f.Edges[0]
					}
					for _, arch := range []string{"amd64", "arm64"} {
						c, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
						if err != nil {
							t.Fatal(arch, err)
						}
						if (c.DeferredPredicates == 1) != (!mode.header && !mode.entryRead) {
							t.Fatal(arch, "wrong predicate selection", c.DeferredPredicates)
						}
					}
					if !native.Available {
						return
					}
					c, lease, ctx := functionContext(t, f, true)
					code, err := native.Publish(forcePrivateLoopPoll(t, c, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(code.Close)
					polled, err := code.Acquire()
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(polled.Close)
					cursor, mask, rhs := predicateLoopInputs(op, back)
					initial := make([]value.Value, len(f.Slots))
					for s := range initial {
						initial[s] = value.Value{Lo: 0x1234567800000000 | uint64(s+1), Hi: 0x89abcdef00000000 | uint64(s+1)}
					}
					initial[0].Lo, initial[1].Lo, initial[2].Lo = cursor, mask, rhs
					if op < 0x50 {
						// Untouched and copied upper bits must remain bit-exact, even
						// though the comparison consumes only the low 32 bits.
						initial[0].Lo |= 0x1234567800000000
						initial[2].Lo |= 0x8765432100000000
					}
					check := func(want []value.Value, exit native.Exit, pc uint64, entry *native.Lease) {
						t.Helper()
						got, err := ctx.Enter(entry, 0)
						state, stateErr := ctx.State()
						if err != nil || stateErr != nil || got != exit || state.Continuation != pc || state.Fuel != 42 || state.Budget != 37 {
							t.Fatal("exit", got, exit, pc, state, err, stateErr)
						}
						for s, word := range want {
							if got := readSlot(t, ctx, 16, slots.Slot(s)); got != word {
								t.Fatal("frame", s, got, word)
							}
						}
					}
					for start := range f.Code {
						for _, poll := range []bool{false, true} {
							if poll && start != 0 {
								continue
							}
							for s, word := range initial {
								writeSlot(t, ctx, 16, slots.Slot(s), word)
							}
							if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start), Fuel: 42, Budget: 37}); err != nil {
								t.Fatal(err)
							}
							want, exit, pc := modelPredicateLoop(f, initial, start, branch, poll)
							entry := lease
							if poll {
								entry = polled
							}
							check(want, exit, pc, entry)
							if poll {
								// Resume the actual published snapshot through the original
								// branch PC, rather than rerunning the omitted comparison.
								want, exit, pc = modelPredicateLoop(f, want, branch, branch, false)
								check(want, exit, pc, lease)
							}
						}
					}
				})
			}
		}
	}
}

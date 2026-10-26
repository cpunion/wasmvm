package codegen

import (
	"encoding/binary"
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct {
	MaxCodeBytes              uint64
	MaxInstructions, MaxSlots uint32
}

func DefaultLimits() Limits { return Limits{64 << 20, 1 << 20, 1 << 20} }

// CompiledFunction is immutable. A continuation is an original slot-code PC,
// never an interior machine address. ReferenceSlots conservatively describes
// every reference-bearing slot in the current frame (root-map ID zero).
type CompiledFunction struct {
	Entry                      nativeabi.Entry
	Function                   *slots.Function
	Offsets                    []uint32
	ReferenceSlots             []uint32
	Helpers, Calls, EdgeCopies uint32
}

// CompileFunction consumes trusted, immutable output of slot lowering. It emits
// complete control flow, integer operations, local/constant/select operations
// and explicit runtime exits for the remaining Core 2.0 semantics. A driver
// must service RuntimeHelper/Call; the compiler never inserts an interpreter.
// Payload[3] enables semantic fuel; Budget bounds native instructions per slice.
// RuntimeHelper/Call publish the instruction PC in Payload[0] and the next PC
// in Continuation. Return publishes its PC for multi-result marshalling.
// Move publishes a large edge-copy plan in Payload[0] and its destination PC
// in Continuation; the driver copies it in bounded, cancelable chunks.
func CompileFunction(arch string, f *slots.Function, limits Limits) (out *CompiledFunction, err error) {
	defer func() {
		if p := recover(); p != nil {
			if e, ok := p.(*wasm.Error); ok {
				out = nil
				err = e
			} else {
				panic(p)
			}
		}
	}()
	if f == nil || len(f.Code) == 0 || len(f.Slots) < int(f.LocalCount) {
		return nil, wasm.ErrInvalid
	}
	if uint64(len(f.Code)) > uint64(limits.MaxInstructions) || uint64(len(f.Slots)) > uint64(limits.MaxSlots) || uint64(len(f.Slots)) > math.MaxInt32/16 {
		return nil, wasm.ErrLimit
	}
	// Each continuation/table edge needs at least one word of code or metadata.
	// Bound label arrays before allocating them, even for enormous br_tables.
	if uint64(len(f.Code)) > limits.MaxCodeBytes/4 || uint64(len(f.Edges)) > limits.MaxCodeBytes/4 {
		return nil, wasm.ErrLimit
	}
	if err := checkFunction(f); err != nil {
		return nil, err
	}
	empty, err := nativeabi.Wrap(arch, nil, limits.MaxCodeBytes)
	if err != nil {
		return nil, err
	}
	a := assembler{arch: arch, labels: map[int]int{}, bounded: true, limit: limits.MaxCodeBytes - uint64(len(empty.Code)), longBranches: true}
	e := functionEmitter{a: &a, f: f, exit: a.label(), bad: a.label(), zero: a.label(), overflow: a.label()}
	pcs := make([]int, len(f.Code))
	for i := range pcs {
		pcs[i] = a.label()
	}
	e.header(uint64(len(f.Slots))*16, len(pcs))
	a.dispatch(pcs)
	out = &CompiledFunction{Function: f, Offsets: make([]uint32, len(pcs))}
	for i, t := range f.Slots {
		if t.IsReference() {
			out.ReferenceSlots = append(out.ReferenceSlots, uint32(i))
		}
	}
	for pc, in := range f.Code {
		a.mark(pcs[pc])
		src := f.Sources[in.Source]
		isControl := in.Op >= slots.ControlBase
		e.boundary(uint32(pc), src, in.Op == 0x10 || in.Op == 0x11 || isControl)
		switch op := in.Op; {
		case isControl:
			kind := ir.TermKind(op - slots.ControlBase)
			switch kind {
			case ir.Return:
				e.leave(nativeabi.Return, uint64(pc), uint64(pc))
			case ir.Trap:
				e.trap(wasm.TrapUnreachable, uint64(src.Wasm.Offset))
			case ir.Jump:
				e.edge(in.Extra, pcs)
			case ir.If:
				other := a.label()
				e.slotLoad(in.A, 0, false, false)
				e.test(false)
				e.branch(other, "eq")
				e.edge(in.Extra, pcs)
				a.mark(other)
				e.edge(in.Extra+1, pcs)
			case ir.Table:
				targets := make([]int, in.Count)
				for i := range targets {
					targets[i] = a.label()
				}
				e.slotLoad(in.A, 0, false, false)
				e.constant(1, uint64(in.Count-1))
				e.compare(false)
				valid := a.label()
				e.branch(valid, "lo")
				e.constant(0, uint64(in.Count-1))
				a.mark(valid)
				a.dispatch(targets)
				for i, l := range targets {
					a.mark(l)
					e.edge(in.Extra+uint32(i), pcs)
				}
			}
		case op == 1 || op == 0x1a:
		case op == 0x20:
			e.move(in.Dst, slots.Slot(src.Wasm.A))
		case op == 0x21 || op == 0x22:
			e.move(slots.Slot(src.Wasm.A), in.A)
			if in.Dst != slots.NoSlot {
				e.move(in.Dst, in.A)
			}
		case op == 0x1b || op == 0x1c:
			other, done := a.label(), a.label()
			e.slotLoad(in.C, 0, false, false)
			e.test(false)
			e.branch(other, "eq")
			e.move(in.Dst, in.A)
			a.branch(done, false, 0)
			a.mark(other)
			e.move(in.Dst, in.B)
			a.mark(done)
		case op >= 0x41 && op <= 0x44 || op == 0xd0 || op == 0xfd000c:
			lo, hi := src.Wasm.Bits, uint64(0)
			if op == 0xd0 {
				lo = 0
			}
			if op == 0xfd000c {
				lo = binary.LittleEndian.Uint64(src.Wasm.Vector[:8])
				hi = binary.LittleEndian.Uint64(src.Wasm.Vector[8:])
			}
			e.constant(0, lo)
			e.slotStore(in.Dst, 0, false)
			e.constant(0, hi)
			e.slotStore(in.Dst, 0, true)
		case isInteger(op) || op == 0xd1:
			if op == 0xd1 {
				op = 0x50
			}
			wide := integerWide(op)
			// Only division/remainder can trap; publish their source before entry.
			if op >= 0x6d && op <= 0x70 || op >= 0x7f && op <= 0x82 {
				e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			}
			e.slotLoad(in.A, 0, false, wide)
			e.slotLoad(in.B, 1, false, wide)
			if arch == "amd64" {
				integerAMD64Core(&a, op, wide, e.zero, e.overflow)
			} else {
				integerARM64Core(&a, op, wide, e.zero, e.overflow)
			}
			e.slotStore(in.Dst, 0, false)
			e.constant(0, 0)
			e.slotStore(in.Dst, 0, true)
		case op == 0x10 || op == 0x11:
			out.Calls++
			e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			e.leave(nativeabi.Call, uint64(pc), uint64(pc+1))
		case helperOpcode(op):
			out.Helpers++
			e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			e.leave(nativeabi.RuntimeHelper, uint64(pc), uint64(pc+1))
		default:
			return nil, wasm.ErrUnsupported
		}
	}
	// A validated function never falls through its final terminator.
	a.mark(e.bad)
	e.status(nativeabi.BadState)
	a.mark(e.zero)
	e.trapCode(wasm.TrapIntegerDivideByZero)
	a.mark(e.overflow)
	e.trapCode(wasm.TrapIntegerOverflow)
	a.mark(e.exit)
	out.EdgeCopies = e.edgeCopies
	body, err := a.finish()
	if err != nil {
		return nil, err
	}
	out.Entry, err = nativeabi.Wrap(arch, body, limits.MaxCodeBytes)
	if err != nil {
		return nil, err
	}
	for i, l := range pcs {
		out.Offsets[i] = out.Entry.BodyOffset + uint32(a.labels[l])
	}
	return out, nil
}

func isInteger(op uint32) bool {
	return op >= 0x45 && op <= 0x5a || op >= 0x67 && op <= 0x8a || op == 0xa7 || op == 0xac || op == 0xad || op >= 0xc0 && op <= 0xc4
}
func integerWide(op uint32) bool {
	return op >= 0x50 && op <= 0x5a || op >= 0x79 && op <= 0x8a || op == 0xac || op == 0xad || op >= 0xc2
}
func helperOpcode(op uint32) bool {
	return op >= 0x23 && op <= 0x26 || op >= 0x28 && op <= 0x40 || op >= 0x5b && op <= 0x66 || op >= 0x8b && op <= 0xbf || op == 0xd2 || op >= 0xfc0000 && op <= 0xfc0011 || op>>16 == 0xfd
}

func checkFunction(f *slots.Function) error {
	valid := func(s slots.Slot) bool { return s != slots.NoSlot && uint64(s) < uint64(len(f.Slots)) }
	for _, s := range f.Args {
		if !valid(s) {
			return wasm.ErrInvalid
		}
	}
	for _, m := range f.Moves {
		if !valid(m.Dst) || !valid(m.Src) {
			return wasm.ErrInvalid
		}
	}
	for _, e := range f.Edges {
		if uint64(e.PC) >= uint64(len(f.Code)) || uint64(e.Start)+uint64(e.Count) > uint64(len(f.Moves)) {
			return wasm.ErrInvalid
		}
	}
	for _, in := range f.Code {
		if uint64(in.Source) >= uint64(len(f.Sources)) {
			return wasm.ErrInvalid
		}
		for _, s := range []slots.Slot{in.A, in.B, in.C, in.Dst} {
			if s != slots.NoSlot && !valid(s) {
				return wasm.ErrInvalid
			}
		}
		if in.Op >= 0x20 && in.Op <= 0x22 && f.Sources[in.Source].Wasm.A >= f.LocalCount {
			return wasm.ErrInvalid
		}
		if in.Op >= slots.ControlBase {
			if in.Op > slots.ControlBase+uint32(ir.Trap) {
				return wasm.ErrInvalid
			}
			kind := ir.TermKind(in.Op - slots.ControlBase)
			switch kind {
			case ir.Return:
				if uint64(in.Extra)+uint64(in.Count) > uint64(len(f.Args)) {
					return wasm.ErrInvalid
				}
			case ir.Trap:
			case ir.Jump, ir.If, ir.Table:
				if in.Count == 0 || kind == ir.Jump && in.Count != 1 || kind == ir.If && in.Count != 2 || uint64(in.Extra)+uint64(in.Count) > uint64(len(f.Edges)) {
					return wasm.ErrInvalid
				}
			default:
				return wasm.ErrInvalid
			}
		}
		if in.Op == 0x10 || in.Op == 0x11 {
			if uint64(in.Extra)+uint64(in.Count)+uint64(len(f.Sources[in.Source].Outputs)) > uint64(len(f.Args)) {
				return wasm.ErrInvalid
			}
		}
	}
	return nil
}

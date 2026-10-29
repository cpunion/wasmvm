package codegen

import (
	"encoding/binary"
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// rdi/x0 is the context, rsi/x12 is the current frame, and rax/rcx or x9/x10
// carry operands. The unmetered tier retains selected scalar slots in registers across control
// transfers and spills them only on exits.
type functionEmitter struct {
	a                         *assembler
	f                         *slots.Function
	exit, bad, zero, overflow int
	edgeCopies                uint32
	unmetered                 bool
	x86Features               X86Features
	usesMemory                bool
	memoryMinimum             uint64
	regionLimit               int
	regionEnd                 int
	provenPCs                 map[int]bool
	guardPCs                  map[int]memoryGuard
	addresses                 map[int]memoryAddress
	elidedPCs                 map[int]bool
	boundsElided              uint32
	constantRegisters         map[uint64]byte
	runtimeRegisters          map[uint32]byte
	registers                 map[slots.Slot]byte
	floatRegisters            map[slots.Slot]byte
	vectorRegisters           map[slots.Slot]byte
	vectorConstants           map[[2]uint64]byte
	vectorAliases             map[slots.Slot]slots.Slot
	aliasedVectorReads        uint32
	registerSlots             []slots.Slot
	pollExits                 []pollExit
	trapExits                 []nativeTrapExit
	callExits                 []nativeCallExit
	memoryFallbacks           []memoryFallback
	loopBackedges             map[int]loopBackedge
}

type memoryFallback struct{ label, pc int }

type nativeCallExit struct {
	label, pc int
	source    uint64
}

type pollExit struct {
	label int
	pc    uint32
}

func (e *functionEmitter) integer(in slots.Instruction) {
	op := in.Op
	if op == 0xd1 {
		op = 0x50
	}
	wide := integerWide(op)
	if e.registerInteger(op, in.Dst, in.A, in.B) {
		return
	}
	if op >= 0x6d && op <= 0x70 || op >= 0x7f && op <= 0x82 {
		e.contextConst(nativeabi.SourceOffset, uint64(e.f.Sources[in.Source].Wasm.Offset))
	}
	e.slotLoad(in.A, 0, false, wide)
	e.slotLoad(in.B, 1, false, wide)
	if e.a.arch == "amd64" {
		integerAMD64Core(e.a, op, wide, e.zero, e.overflow)
	} else {
		integerARM64Core(e.a, op, wide, e.zero, e.overflow)
	}
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(in.Dst, 0, true)
	}
}

func (e *functionEmitter) constant(reg byte, v uint64) {
	a := e.a
	if a.arch == "amd64" {
		if v <= math.MaxUint32 {
			a.bytes(0xb8 + reg)
			a.word(uint32(v))
		} else {
			a.xconst(reg, v)
		}
	} else {
		a.aconst(uint32(reg)+9, v, v > math.MaxUint32)
	}
}
func (e *functionEmitter) contextLoad(reg byte, offset uintptr) {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x8b, 0x87+reg*8)
		e.a.word(uint32(offset))
	} else {
		e.a.word(0xf9400000 | uint32(reg+9) | uint32(offset/8)<<10)
	}
}
func (e *functionEmitter) contextStore(reg byte, offset uintptr) {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x89, 0x87+reg*8)
		e.a.word(uint32(offset))
	} else {
		e.a.word(0xf9000000 | uint32(reg+9) | uint32(offset/8)<<10)
	}
}
func (e *functionEmitter) contextConst(offset uintptr, v uint64) {
	e.constant(0, v)
	e.contextStore(0, offset)
}
func (e *functionEmitter) branch(label int, condition string) {
	var cc byte
	if e.a.arch == "amd64" {
		switch condition {
		case "eq":
			cc = 4
		case "ne":
			cc = 5
		case "lo":
			cc = 2
		case "hs":
			cc = 3
		}
	} else {
		switch condition {
		case "eq":
			cc = 0
		case "ne":
			cc = 1
		case "lo":
			cc = 3
		case "hs":
			cc = 2
		}
	}
	e.a.branch(label, true, cc)
}
func (e *functionEmitter) test(wide bool) {
	if e.a.arch == "amd64" {
		e.a.xop(wide, 0x85, 0xc0)
	} else {
		e.a.arm(wide, 0x7100013f)
	}
}
func (e *functionEmitter) compare(wide bool) {
	if e.a.arch == "amd64" {
		e.a.xop(wide, 0x39, 0xc8)
	} else {
		e.a.arm(wide, 0x6b0a013f)
	}
}
func (e *functionEmitter) decrement() {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x83, 0xe8, 1)
	} else {
		e.a.word(0xd1000529)
	}
}
func (e *functionEmitter) slotLoad(slot slots.Slot, reg byte, high, wide bool) {
	if slot == slots.NoSlot {
		e.constant(reg, 0)
		return
	}
	if e.unmetered && e.f.Slots[slot] != wasm.V128 {
		if high {
			e.constant(reg, 0)
			return
		}
		if r, ok := e.floatRegisters[slot]; ok {
			e.floatToInteger(e.operandRegister(reg), r, e.f.Slots[slot] == wasm.F64)
			if !wide {
				e.registerMove(e.operandRegister(reg), e.operandRegister(reg), false)
			}
			return
		}
		if r, ok := e.registers[slot]; ok {
			e.registerMove(e.operandRegister(reg), r, wide)
			return
		}
	}
	offset := uint32(slot) * 16
	if high {
		offset += 8
	}
	a := e.a
	if a.arch == "amd64" {
		a.xop(wide, 0x8b, 0x86+reg*8)
		a.word(offset)
		return
	}
	ins, scale := uint32(0xb9400000), uint32(4)
	if wide {
		ins, scale = 0xf9400000, 8
	}
	base := uint32(12)
	if offset/scale >= 4096 {
		a.aconst(15, uint64(offset), false)
		a.word(0x8b0f018f)
		base = 15
		offset = 0
	}
	a.word(ins | uint32(reg+9) | base<<5 | (offset/scale)<<10)
}
func (e *functionEmitter) slotStore(slot slots.Slot, reg byte, high bool) {
	if slot == slots.NoSlot {
		return
	}
	if e.unmetered && e.f.Slots[slot] != wasm.V128 {
		if high {
			return
		}
		if r, ok := e.floatRegisters[slot]; ok {
			e.floatFromInteger(r, e.operandRegister(reg), e.f.Slots[slot] == wasm.F64)
			return
		}
		if r, ok := e.registers[slot]; ok {
			e.registerMove(r, e.operandRegister(reg), true)
			return
		}
	}
	offset := uint32(slot) * 16
	if high {
		offset += 8
	}
	a := e.a
	if a.arch == "amd64" {
		a.bytes(0x48, 0x89, 0x86+reg*8)
		a.word(offset)
		return
	}
	base := uint32(12)
	if offset/8 >= 4096 {
		a.aconst(15, uint64(offset), false)
		a.word(0x8b0f018f)
		base = 15
		offset = 0
	}
	a.word(0xf9000000 | uint32(reg+9) | base<<5 | (offset/8)<<10)
}
func (e *functionEmitter) move(dst, src slots.Slot) {
	if dst == slots.NoSlot || dst == src {
		return
	}
	if e.f.Slots[dst] == wasm.V128 {
		source := src
		if alias, ok := e.vectorAliases[src]; ok {
			source = alias
		}
		r := e.vectorOperand(src, 0)
		e.vectorStore(dst, r)
		if e.vectorAliases != nil && source != dst {
			if _, cached := e.vectorRegisters[source]; cached {
				e.vectorAliases[dst] = source
			}
		}
		return
	}
	if d, ok := e.floatRegisters[dst]; ok {
		if s, ok := e.floatRegisters[src]; ok {
			e.floatMove(d, s, e.f.Slots[dst] == wasm.F64)
			return
		}
	}
	if d, ok := e.registers[dst]; ok {
		if s, ok := e.registers[src]; ok {
			e.registerMove(d, s, true)
			return
		}
	}
	e.slotLoad(src, 0, false, true)
	if !e.unmetered || e.f.Slots[dst] == wasm.V128 {
		e.slotLoad(src, 1, true, true)
	}
	e.slotStore(dst, 0, false)
	e.slotStore(dst, 1, true)
}
func (e *functionEmitter) status(status nativeabi.Exit) {
	e.spillRegisters()
	e.rawStatus(status)
}
func (e *functionEmitter) rawStatus(status nativeabi.Exit) {
	if e.a.arch == "amd64" {
		e.a.bytes(0xb8)
		e.a.word(uint32(status))
	} else {
		e.a.aconst(0, uint64(status), false)
	}
	e.a.branch(e.exit, false, 0)
}
func (e *functionEmitter) leave(status nativeabi.Exit, payload, pc uint64) {
	e.contextConst(nativeabi.ContinuationOffset, pc)
	e.contextConst(nativeabi.PayloadOffset, payload)
	e.status(status)
}
func (e *functionEmitter) trapCode(trap wasm.TrapCode) {
	e.contextConst(nativeabi.PayloadOffset, uint64(trap))
	e.status(nativeabi.Trap)
}
func (e *functionEmitter) trap(trap wasm.TrapCode, source uint64) {
	e.contextConst(nativeabi.SourceOffset, source)
	e.trapCode(trap)
}

func (e *functionEmitter) header(bytes uint64, instructions int) {
	a := e.a
	if a.arch == "amd64" {
		a.bytes(0x48, 0x8b, 0x77, byte(nativeabi.ArenaOffset), 0x48, 0x8b, 0x57, byte(nativeabi.UsedOffset))
		e.contextLoad(0, nativeabi.CapacityOffset)
		a.bytes(0x48, 0x29, 0xd0)
		e.constant(1, bytes)
		e.compare(true)
		e.branch(e.bad, "lo")
		a.bytes(0x48, 0x89, 0xd0, 0x83, 0xe0, 15)
		e.branch(e.bad, "ne")
		a.bytes(0x48, 0x01, 0xd6)
	} else {
		a.word(0xf940040c)
		a.word(0xf940100d)
		e.contextLoad(0, nativeabi.CapacityOffset)
		a.word(0xcb0d0129)
		e.constant(1, bytes)
		e.compare(true)
		e.branch(e.bad, "lo")
		a.word(0x92400da9)
		e.test(true)
		e.branch(e.bad, "ne")
		a.word(0x8b0d018c)
	}
	e.contextConst(nativeabi.RootMapOffset, 0)
	e.contextLoad(0, nativeabi.ContinuationOffset)
	e.constant(1, uint64(instructions))
	e.compare(true)
	e.branch(e.bad, "hs")
}

func (e *functionEmitter) boundary(pc uint32, src ir.Source, control bool) {
	a := e.a
	ready := a.label()
	e.contextLoad(0, nativeabi.BudgetOffset)
	e.test(true)
	e.branch(ready, "ne")
	e.leave(nativeabi.Poll, 0, uint64(pc))
	a.mark(ready)
	e.decrement()
	e.contextStore(0, nativeabi.BudgetOffset)
	if control {
		e.interrupt(pc)
	}
	if src.Fuel() == 0 {
		return
	}
	done, charged := a.label(), a.label()
	e.contextLoad(0, nativeabi.PayloadOffset+24)
	e.test(true)
	e.branch(done, "eq")
	e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
	e.contextLoad(0, nativeabi.FuelOffset)
	e.test(true)
	e.branch(charged, "ne")
	e.contextConst(nativeabi.ContinuationOffset, uint64(pc))
	e.status(nativeabi.FuelExhausted)
	a.mark(charged)
	e.decrement()
	e.contextStore(0, nativeabi.FuelOffset)
	a.mark(done)
}

func (e *functionEmitter) edge(index uint32, pcs []int, next int) {
	edge := e.f.Edges[index]
	if edge.Count > 256 {
		e.edgeCopies++
		e.leave(nativeabi.Move, uint64(index), uint64(edge.PC))
		return
	}
	for _, m := range e.f.Moves[edge.Start : edge.Start+edge.Count] {
		e.move(m.Dst, m.Src)
	}
	if int(edge.PC) != next {
		e.a.branch(pcs[edge.PC], false, 0)
	}
}

// Interrupt polling is distinct from a scheduling round trip. A straight-line
// sequence has at most 256 slot instructions between polls; every real control
// transfer also polls, so arbitrary forward/backward edges cannot evade it.
// A no-copy jump to the next PC is ordinary fallthrough and cannot skip a poll.
func (e *functionEmitter) fallthroughEdge(pc int) bool {
	in := e.f.Code[pc]
	if in.Op != slots.ControlBase+uint32(ir.Jump) {
		return false
	}
	edge := e.f.Edges[in.Extra]
	return edge.PC == uint32(pc+1) && edge.Count == 0
}

func (e *functionEmitter) interrupt(pc uint32) {
	// An aligned machine load is atomic on both targets. Only a nonzero
	// observation needs acquire ordering; the cold exit supplies that fence.
	// Keep register spills and continuation publication outside the hot region.
	exit := e.a.label()
	if e.a.arch == "arm64" {
		e.contextLoad(0, nativeabi.ControlOffset)
		e.a.compareBranch(exit, e.operandRegister(0), true, true)
	} else {
		e.a.bytes(0x48, 0x83, 0xbf) // cmpq $0, control(%rdi), one aligned 64-bit load.
		e.a.word(uint32(nativeabi.ControlOffset))
		e.a.bytes(0)
		e.branch(exit, "ne")
	}
	e.pollExits = append(e.pollExits, pollExit{exit, pc})
}

func (e *functionEmitter) emitPollExits() {
	for _, exit := range e.pollExits {
		e.a.mark(exit.label)
		if e.a.arch == "arm64" {
			e.a.word(0xd50339bf) // dmb ishld: acquire after the atomic load.
		}
		e.leave(nativeabi.Poll, 0, uint64(exit.pc))
	}
}

func (e *functionEmitter) literal(in slots.Instruction) {
	src := e.f.Sources[in.Source]
	op := in.Op
	lo, hi := src.Wasm.Bits, uint64(0)
	if op == 0xd0 {
		lo = 0
	}
	if op == 0xfd000c {
		lo = binary.LittleEndian.Uint64(src.Wasm.Vector[:8])
		hi = binary.LittleEndian.Uint64(src.Wasm.Vector[8:])
		if r, ok := e.vectorConstants[[2]uint64{lo, hi}]; ok {
			e.vectorStore(in.Dst, r)
		} else {
			e.vectorConstant(0, lo, hi)
			e.vectorStore(in.Dst, 0)
		}
		return
	}
	if r, ok := e.floatRegisters[in.Dst]; ok && e.a.arch == "arm64" {
		wide := e.f.Slots[in.Dst] == wasm.F64
		if immediate, ok := floatImmediate(lo, wide); ok {
			ins := uint32(0x1e201000) | uint32(immediate)<<13 | uint32(r)
			if wide {
				ins |= 1 << 22
			}
			e.a.word(ins)
			return
		}
	}
	if op == 0x41 || op == 0x42 {
		if op == 0x41 {
			lo = uint64(uint32(lo))
		}
		if r, ok := e.constantRegisters[lo]; ok {
			if dst, ok := e.registers[in.Dst]; ok {
				e.registerMove(dst, r, true)
			} else {
				e.registerMove(e.operandRegister(0), r, true)
				e.slotStore(in.Dst, 0, false)
			}
			return
		}
	}
	if r, ok := e.registers[in.Dst]; ok {
		if e.a.arch == "arm64" {
			e.a.aconst(uint32(r), lo, lo > math.MaxUint32)
		} else {
			e.a.xconst(r, lo)
		}
	} else {
		e.constant(0, lo)
		e.slotStore(in.Dst, 0, false)
	}
	if !e.unmetered || e.f.Slots[in.Dst] == wasm.V128 {
		e.constant(0, hi)
		e.slotStore(in.Dst, 0, true)
	}
}

// Keep source publication and register spills off successful conversion paths.
// Each trap site retains its original instruction offset.
type nativeTrapExit struct {
	label  int
	code   wasm.TrapCode
	source uint64
}

func (e *functionEmitter) trapExit(code wasm.TrapCode, source uint64) int {
	label := e.a.label()
	e.trapExits = append(e.trapExits, nativeTrapExit{label, code, source})
	return label
}

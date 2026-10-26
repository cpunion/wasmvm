package codegen

import (
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
	registers                 map[slots.Slot]byte
	registerSlots             []slots.Slot
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
	if dst == slots.NoSlot {
		return
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

func (e *functionEmitter) edge(index uint32, pcs []int) {
	edge := e.f.Edges[index]
	if edge.Count > 256 {
		e.edgeCopies++
		e.leave(nativeabi.Move, uint64(index), uint64(edge.PC))
		return
	}
	for _, m := range e.f.Moves[edge.Start : edge.Start+edge.Count] {
		e.move(m.Dst, m.Src)
	}
	e.a.branch(pcs[edge.PC], false, 0)
}

// Interrupt polling is distinct from a scheduling round trip. A straight-line
// sequence has at most 256 slot instructions between polls; every control
// transfer also polls, so arbitrary forward/backward edges cannot evade it.
func (e *functionEmitter) interrupt(pc uint32) {
	a := e.a
	ok := a.label()
	if a.arch == "amd64" {
		e.contextLoad(0, nativeabi.ControlOffset)
	} else {
		a.word(0x9100000f | uint32(nativeabi.ControlOffset)<<10)
		a.word(0xc8dffde9) // ldar x9,[x15]
	}
	e.test(true)
	e.branch(ok, "eq")
	e.leave(nativeabi.Poll, 0, uint64(pc))
	a.mark(ok)
}

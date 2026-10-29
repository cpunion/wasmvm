package codegen

import (
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Assign fixed registers to the most frequently referenced scalar slots. The
// assignment is identical at every CFG edge and continuation. This first local
// allocator needs neither edge shuffles nor speculative liveness assumptions.
// Operands, integer-core scratch, the frame base and context are excluded.
func (e *functionEmitter) prepareRegisters(plans ...map[int]memoryRegion) {
	var regions map[int]memoryRegion
	if len(plans) != 0 {
		regions = plans[0]
	}
	e.allocateRegisters(Config{}, regions)
}

// Mutable slots and leaf expressions share the target inventory. Planned
// immutable constants and runtime metadata reserve disjoint ARM caller registers.
func (e *functionEmitter) integerRegisterPool() []byte {
	pool := []byte{3, 12, 13, 14, 15, 9, 10, 11}
	if e.usesMemory {
		pool = pool[:6]
	}
	if e.a.arch == "arm64" {
		pool = []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 1, 2, 3, 4, 5, 6, 7, 8}
	}
	var used [32]bool
	for _, r := range e.constantRegisters {
		used[r] = true
	}
	for _, r := range e.runtimeRegisters {
		used[r] = true
	}
	free := pool[:0]
	for _, r := range pool {
		if !used[r] {
			free = append(free, r)
		}
	}
	return free
}

func (e *functionEmitter) spareCallerRegisters() []byte {
	var used [32]bool
	for _, r := range e.registers {
		used[r] = true
	}
	for _, r := range e.constantRegisters {
		used[r] = true
	}
	for _, r := range e.runtimeRegisters {
		used[r] = true
	}
	var free []byte
	for r := byte(1); r <= 8; r++ {
		if !used[r] {
			free = append(free, r)
		}
	}
	return free
}

func (e *functionEmitter) operandRegister(reg byte) byte {
	if e.a.arch == "arm64" {
		return reg + 9
	}
	return reg
}

func (e *functionEmitter) registerMove(dst, src byte, wide bool) {
	if dst == src && wide {
		return
	}
	if e.a.arch == "amd64" {
		rex := byte(0x40) | (src>>3)<<2 | dst>>3
		if wide {
			rex |= 8
		}
		if rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x89, 0xc0|(src&7)<<3|dst&7)
	} else {
		e.a.arm(wide, 0x2a0003e0|uint32(src)<<16|uint32(dst))
	}
}

func (e *functionEmitter) spillRegisters() {
	registers, floats := e.registers, e.floatRegisters
	e.registers, e.floatRegisters = nil, nil
	// Stable order makes emitted code independent of Go map iteration order.
	for _, s := range e.registerSlots {
		if r, ok := e.vectorRegisters[s]; ok {
			e.vectorMemory(s, r, true)
			continue
		}
		if r, ok := registers[s]; ok {
			e.registerMove(e.operandRegister(0), r, true)
		} else {
			e.floatToInteger(e.operandRegister(0), floats[s], e.f.Slots[s] == wasm.F64)
		}
		e.slotStore(s, 0, false)
	}
	e.registers, e.floatRegisters = registers, floats
}

// Common integer operations can name the allocated registers directly. The
// remaining instructions retain the tested constrained-register lowering.
func (e *functionEmitter) canRegisterInteger(op uint32, dst, left, right slots.Slot) bool {
	d, dok := e.registers[dst]
	l, lok := e.registers[left]
	r, rok := e.registers[right]
	if !dok || !lok {
		return false
	}
	if op >= 0x45 && op <= 0x5a {
		return rok || op == 0x45 || op == 0x50
	}
	if !rok || op < 0x67 || op > 0x8a {
		return false
	}
	kind := op - 0x67
	if integerWide(op) {
		kind = op - 0x79
	}
	return (kind >= 3 && kind <= 5 || kind >= 10 && kind <= 17) && (e.a.arch == "arm64" || kind != 4 || d != r || d == l)
}

func (e *functionEmitter) registerInteger(op uint32, dst, left, right slots.Slot) bool {
	if !e.canRegisterInteger(op, dst, left, right) {
		return false
	}
	d, l, r := e.registers[dst], e.registers[left], e.registers[right]
	if op >= 0x45 && op <= 0x5a {
		e.registerCompare(d, l, r, op)
		return true
	}
	wide := integerWide(op)
	kind := op - 0x67
	if wide {
		kind = op - 0x79
	}

	a := e.a
	if a.arch == "arm64" {
		var ins uint32
		switch kind {
		case 3:
			ins = 0x0b000000 // add
		case 4:
			ins = 0x4b000000 // sub
		case 5:
			ins = 0x1b007c00 // mul
		case 10:
			ins = 0x0a000000 // and
		case 11:
			ins = 0x2a000000 // orr
		case 12:
			ins = 0x4a000000 // eor
		case 13:
			ins = 0x1ac02000 // lslv
		case 14:
			ins = 0x1ac02800 // asrv
		case 15:
			ins = 0x1ac02400 // lsrv
		case 16: // rotl uses a scratch count; never destroys a cached operand.
			a.arm(wide, 0x4b0003ea|uint32(r)<<16)
			r = 10
			ins = 0x1ac02c00
		case 17:
			ins = 0x1ac02c00 // rorv
		default:
			return false
		}
		a.arm(wide, ins|uint32(r)<<16|uint32(l)<<5|uint32(d))
		return true
	}
	// Commutative operations can start with either operand. Variable shifts
	// capture their count in CL before replacing an aliased destination.
	// Right-aliased subtraction retains the constrained-register fallback.
	if d == r && d != l && kind < 13 {
		l, r = r, l
	}
	if kind >= 13 && kind <= 17 {
		e.registerMove(1, r, false) // Variable shifts and rotates use CL.
		e.registerMove(d, l, wide)
		rex := byte(0x40) | d>>3
		if wide {
			rex |= 8
		}
		if rex != 0x40 {
			a.bytes(rex)
		}
		ext := [...]byte{4, 7, 5, 0, 1}[kind-13]
		a.bytes(0xd3, 0xc0|ext<<3|d&7)
		return true
	}
	var opcode byte
	switch kind {
	case 3:
		opcode = 0x01
	case 4:
		opcode = 0x29
	case 5:
		opcode = 0xaf
	case 10:
		opcode = 0x21
	case 11:
		opcode = 0x09
	case 12:
		opcode = 0x31
	default:
		return false
	}
	e.registerMove(d, l, wide)
	src, dest := r, d
	if kind == 5 {
		src, dest = d, r
	}
	rex := byte(0x40) | (src>>3)<<2 | dest>>3
	if wide {
		rex |= 8
	}
	if rex != 0x40 {
		a.bytes(rex)
	}
	if kind == 5 {
		a.bytes(0x0f)
	}
	a.bytes(opcode, 0xc0|(src&7)<<3|dest&7)
	return true
}

// Immediate shift selection is valid only along a proven constant fallthrough.
// The original slot PC keeps the generic instruction for independent entry.
func (e *functionEmitter) canRegisterImmediate(op uint32, dst, left slots.Slot, bits uint64) bool {
	_, dok := e.registers[dst]
	_, lok := e.registers[left]
	_, arithmetic := x86ImmediateGroup(op, bits)
	return dok && lok && (op >= 0x74 && op <= 0x78 || op >= 0x86 && op <= 0x8a || e.a.arch == "amd64" && arithmetic)
}

func (e *functionEmitter) registerImmediate(op uint32, dst, left slots.Slot, value uint64) bool {
	d, dok := e.registers[dst]
	l, lok := e.registers[left]
	if !dok || !lok {
		return false
	}
	if e.a.arch == "amd64" && e.x86ArithmeticImmediate(op, d, l, value) {
		return true
	}
	wide := integerWide(op)
	if !wide {
		value = uint64(uint32(value))
	}
	if op == 0x6c || op == 0x7e {
		if value != 0 && value&(value-1) == 0 {
			shift := uint32(0x74)
			if wide {
				shift = 0x86
			}
			return e.registerImmediate(shift, dst, left, uint64(bits.TrailingZeros64(value)))
		}
	}
	if e.a.arch == "arm64" {
		ins, ok := uint32(0), false
		switch op {
		case 0x6a, 0x7c, 0x6b, 0x7d:
			if value < 4096 {
				ins, ok = 0x11000000|uint32(value)<<10, true
				if op == 0x6b || op == 0x7d {
					ins |= 1 << 30
				}
			}
		case 0x71, 0x83:
			if value != 0 && value != ^uint64(0) && value&(value+1) == 0 {
				ins, ok = 0x53000000|uint32(bits.Len64(value)-1)<<10, true
				if wide {
					ins |= 1 << 22
				}
			}
		case 0x6c, 0x7e:
			if value > 2 && (value-1)&(value-2) == 0 {
				ins, ok = 0x0b000000|uint32(l)<<16|uint32(bits.TrailingZeros64(value-1))<<10, true
			}
		}
		if ok {
			e.a.arm(wide, ins|uint32(l)<<5|uint32(d))
			return true
		}
	}
	if !(op >= 0x74 && op <= 0x78 || op >= 0x86 && op <= 0x8a) {
		return false
	}
	first, width := uint32(0x74), uint32(32)
	if wide {
		first, width = 0x86, 64
	}
	kind, count := op-first, uint32(value)&(width-1)
	a := e.a
	if a.arch == "amd64" {
		e.registerMove(d, l, wide)
		rex := byte(0x40) | d>>3
		if wide {
			rex |= 8
		}
		if rex != 0x40 {
			a.bytes(rex)
		}
		ext := [...]byte{4, 7, 5, 0, 1}[kind]
		a.bytes(0xc1, 0xc0|ext<<3|d&7, byte(count))
		return true
	}
	var word uint32
	if kind >= 3 {
		if kind == 3 {
			count = (width - count) & (width - 1)
		}
		word = 0x13800000 | uint32(l)<<16 | count<<10 // EXTR/ROR
	} else {
		immr, imms := count, width-1
		word = 0x53000000 // UBFM: LSL/LSR
		if kind == 0 {
			immr, imms = (width-count)&(width-1), width-1-count
		} else if kind == 1 {
			word = 0x13000000 // SBFM: ASR
		}
		word |= immr<<16 | imms<<10
	}
	if wide {
		word |= 1<<31 | 1<<22
	}
	a.word(word | uint32(l)<<5 | uint32(d))
	return true
}

func (e *functionEmitter) registerCompare(dst, left, right byte, op uint32) {
	a := e.a
	wide := integerWide(op)
	index := op - 0x45
	if wide {
		index = op - 0x50
	}
	if a.arch == "arm64" {
		if index == 0 {
			a.arm(wide, 0x7100001f|uint32(left)<<5)
		} else {
			a.arm(wide, 0x6b00001f|uint32(right)<<16|uint32(left)<<5)
		}
		cc := uint32(armIntegerCondition(op))
		a.word(0x1a9f07e0 | ((cc ^ 1) << 12) | uint32(dst))
		return
	}
	rex := byte(0x40) | left>>3
	if wide {
		rex |= 8
	}
	if index != 0 {
		rex |= (right >> 3) << 2
	}
	if rex != 0x40 {
		a.bytes(rex)
	}
	if index == 0 {
		a.bytes(0x83, 0xf8|left&7, 0)
	} else {
		a.bytes(0x39, 0xc0|(right&7)<<3|left&7)
	}
	cc := [...]byte{4, 4, 5, 12, 2, 15, 7, 14, 6, 13, 3}[index]
	// REX selects low byte registers, including r12b-r15b, rather than AH-DH.
	a.bytes(0x40|dst>>3, 0x0f, 0x90+cc, 0xc0|dst&7)
	a.bytes(0x40|(dst>>3)<<2|dst>>3, 0x0f, 0xb6, 0xc0|(dst&7)<<3|dst&7)
}

func armIntegerCondition(op uint32) byte {
	index := op - 0x45
	if integerWide(op) {
		index = op - 0x50
	}
	return [...]byte{0, 0, 1, 11, 3, 12, 8, 13, 9, 10, 2}[index]
}

package codegen

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Assign fixed registers to the most frequently referenced scalar slots. The
// assignment is identical at every CFG edge and continuation. This first local
// allocator needs neither edge shuffles nor speculative liveness assumptions.
// Operands, integer-core scratch, the frame base and context are excluded.
func (e *functionEmitter) prepareRegisters() {
	weights := make([]uint64, len(e.f.Slots))
	use := func(s slots.Slot) {
		if s != slots.NoSlot && e.f.Slots[s] != wasm.V128 {
			weights[s]++
		}
	}
	for _, in := range e.f.Code {
		use(in.A)
		use(in.B)
		use(in.C)
		use(in.Dst)
		if in.Op >= 0x20 && in.Op <= 0x22 {
			use(slots.Slot(e.f.Sources[in.Source].Wasm.A))
		}
	}
	for _, m := range e.f.Moves {
		use(m.Src)
		use(m.Dst)
	}
	var order []slots.Slot
	for s, n := range weights {
		if n != 0 {
			order = append(order, slots.Slot(s))
		}
	}
	sort.Slice(order, func(i, j int) bool {
		if weights[order[i]] != weights[order[j]] {
			return weights[order[i]] > weights[order[j]]
		}
		return order[i] < order[j]
	})
	available := []byte{3, 12, 13, 14, 15, 9, 10, 11} // rbx, r12-r15, r9-r11
	if e.a.arch == "arm64" {
		available = []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28}
	}
	if len(order) > len(available) {
		order = order[:len(available)]
	}
	// Load before installing the map: slotLoad must read the backing frame.
	for i, s := range order {
		e.slotLoad(s, 0, false, true)
		e.registerMove(available[i], e.operandRegister(0), true)
	}
	e.registerSlots = order
	e.registers = make(map[slots.Slot]byte, len(order))
	for i, s := range order {
		e.registers[s] = available[i]
	}
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
	registers := e.registers
	if len(registers) == 0 {
		return
	}
	e.registers = nil
	// Stable order makes emitted code independent of Go map iteration order.
	for _, s := range e.registerSlots {
		if r, ok := registers[s]; ok {
			e.registerMove(e.operandRegister(0), r, true)
			e.slotStore(slots.Slot(s), 0, false)
		}
	}
	e.registers = registers
}

// Common integer operations can name the allocated registers directly. The
// remaining instructions retain the tested constrained-register lowering.
func (e *functionEmitter) registerInteger(op uint32, dst, left, right slots.Slot) bool {
	d, dok := e.registers[dst]
	l, lok := e.registers[left]
	r, rok := e.registers[right]
	if !dok || !lok || !rok {
		return false
	}
	wide := integerWide(op)
	kind := op - 0x67
	if wide {
		kind = op - 0x79
	}
	if op < 0x67 || op > 0x8a {
		return false
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
	// Two-address x86 operations must preserve the right operand if it also
	// names the destination. Slot lowering normally avoids this overlap, but
	// this emitter also accepts independently constructed trusted slot IR.
	if d == r && d != l {
		return false
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

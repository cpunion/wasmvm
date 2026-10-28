package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// combineMultiplyAdd selects an integer MADD along an uninterrupted path. It
// preserves the intermediate product as sum-addend, outside the result's
// dependency chain. This keeps every slot observable at later exits without a
// second multiply. Independent entries use the original instruction lowering.
// Constant operands can use the immutable pool directly; temporary constant
// slots are omitted only when overwritten before any subsequent observation.
func (e *functionEmitter) combineMultiplyAdd(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "arm64" || pc+1 >= e.codeLimit() {
		return 0, false
	}
	start := pc
	var prefix *slots.Instruction
	var pooled byte
	if in := &e.f.Code[pc]; in.Op == 0x41 || in.Op == 0x42 {
		var ok bool
		pooled, ok = e.literalRegister(*in)
		if !ok || pc+2 >= e.codeLimit() {
			return 0, false
		}
		prefix = in
		pc++
	}
	mul := e.f.Code[pc]
	if mul.Op != 0x6c && mul.Op != 0x7e {
		return 0, false
	}
	end := pc + 1
	constant := e.f.Code[end].Op == 0x41 || e.f.Code[end].Op == 0x42
	if constant {
		if e.f.Code[end].Dst == mul.Dst {
			return 0, false
		}
		end++
	}
	if end >= e.codeLimit() || start/256 != end/256 {
		return 0, false
	}
	add := e.f.Code[end]
	if add.Op != mul.Op-2 {
		return 0, false
	}
	addend := add.B
	if add.A != mul.Dst {
		if add.B != mul.Dst {
			return 0, false
		}
		addend = add.A
	}
	if addend == mul.Dst {
		return 0, false
	}
	l, lok := e.registers[mul.A]
	r, rok := e.registers[mul.B]
	p, pok := e.registers[mul.Dst]
	c, cok := e.registers[addend]
	d, dok := e.registers[add.Dst]
	if !lok || !rok || !pok || !cok || !dok {
		return 0, false
	}
	if prefix != nil {
		if prefix.Dst != mul.A && prefix.Dst != mul.B {
			return 0, false
		}
		if prefix.Dst == mul.A {
			l = pooled
		}
		if prefix.Dst == mul.B {
			r = pooled
		}
	}
	result := add.Dst
	originalEnd := end
	if end+1 < e.codeLimit() && (end+1)/256 == pc/256 {
		next := e.f.Code[end+1]
		local := slots.Slot(e.f.Sources[next.Source].Wasm.A)
		if reg, ok := e.registers[local]; ok && (next.Op == 0x21 || next.Op == 0x22) && next.A == add.Dst {
			result, d = local, reg
			end++
		}
	}

	wide := integerWide(mul.Op)
	keepProduct := mul.Dst != result && mul.Dst != add.Dst && !e.overwrittenBeforeExit(originalEnd+1, mul.Dst)
	if prefix != nil && prefix.Dst != mul.Dst && !e.overwrittenBeforeExit(pc+1, prefix.Dst) {
		e.literal(*prefix)
	}
	if constant {
		in := e.f.Code[pc+1]
		cached, ok := e.literalRegister(in)
		if ok && in.Dst == addend {
			c = cached
		}
		// A pool operand remains available without writing its temporary slot.
		// Keep the slot only if its value can be observed after the combination.
		materialize := !ok || in.Dst != addend || in.Dst != add.Dst && !e.overwrittenBeforeExit(originalEnd+1, in.Dst)
		// The constant may replace a multiplicand's reusable slot. Save only
		// the clobbered inputs; x13/x14 are excluded from the register cache.
		if materialize && in.Dst == mul.A && l == e.registers[mul.A] {
			e.registerMove(13, l, wide)
			l = 13
		}
		if materialize && in.Dst == mul.B && r == e.registers[mul.B] {
			e.registerMove(14, r, wide)
			r = 14
		}
		if materialize {
			e.literal(in)
		}
	}
	// Name the result register directly to avoid extending its dependency chain.
	// Preserve an aliased addend only when the product must also remain live.
	if keepProduct && d == c {
		e.registerMove(10, c, wide)
		c = 10
	}
	e.a.arm(wide, 0x1b000000|uint32(r)<<16|uint32(c)<<10|uint32(l)<<5|uint32(d))
	if keepProduct {
		e.a.arm(wide, 0x4b000000|uint32(c)<<16|uint32(d)<<5|uint32(p)) // product = sum - addend
	}
	if end != originalEnd {
		if !e.overwrittenBeforeExit(end+1, add.Dst) {
			e.move(add.Dst, result)
		}
		e.move(e.f.Code[end].Dst, result)
	}
	return end, true
}

// A temporary is dead only if normal fallthrough overwrites it before any
// reader, poll, trap or helper exit. Entry at another PC makes no assumption
// about its value. The 256-instruction poll interval bounds this scan.
func (e *functionEmitter) overwrittenBeforeExit(pc int, slot slots.Slot) bool {
	for ; pc < e.codeLimit() && pc%256 != 0; pc++ {
		if e.elidedPCs[pc] {
			continue
		}
		in := e.f.Code[pc]
		op := in.Op
		if !plainOpcode(op) || scalarCanTrap(op) || isMemoryAccess(op) && !e.provenPCs[pc] {
			return false
		}
		local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
		address := in.A
		if remap, ok := e.addresses[pc]; ok {
			address = remap.local
		}
		if address == slot || in.B == slot || in.C == slot || op == 0x20 && local == slot {
			return false
		}
		if in.Dst == slot || (op == 0x21 || op == 0x22) && local == slot {
			return true
		}
	}
	return false
}

// Recognize the complete pattern without emitting code. Copy forwarding uses
// the same selection so a shorter group cannot hide a multiply-add candidate.
func (e *functionEmitter) multiplyAddTail(pc int) (int, slots.Slot, bool, bool) {
	if pc+1 >= e.codeLimit() {
		return 0, slots.NoSlot, false, false
	}
	mul := e.f.Code[pc]
	if mul.Op != 0x6c && mul.Op != 0x7e {
		return 0, slots.NoSlot, false, false
	}
	end := pc + 1
	constant := e.f.Code[end].Op == 0x41 || e.f.Code[end].Op == 0x42
	if constant {
		if e.f.Code[end].Dst == mul.Dst {
			return 0, slots.NoSlot, false, false
		}
		end++
	}
	if end >= e.codeLimit() || pc/256 != end/256 {
		return 0, slots.NoSlot, false, false
	}
	add := e.f.Code[end]
	if add.Op != mul.Op-2 {
		return 0, slots.NoSlot, false, false
	}
	addend := add.B
	if add.A != mul.Dst {
		if add.B != mul.Dst {
			return 0, slots.NoSlot, false, false
		}
		addend = add.A
	}
	if addend == mul.Dst {
		return 0, slots.NoSlot, false, false
	}
	return end, addend, constant, true
}

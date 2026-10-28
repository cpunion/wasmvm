package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type memoryGuard struct {
	address       slots.Slot
	before, after uint64
}

type memoryAddress struct {
	local  slots.Slot
	offset uint64
}

type addressRoot struct {
	local   slots.Slot
	version uint32
}

type affineAddress struct {
	root             addressRoot
	bias             uint64
	affine, constant bool
	bits             uint32
}

type addressGroup struct {
	firstBias, end uint64
	sites          []int
}

// Regions are bounded by 256 instructions. Local versions prevent stale facts
// after mutation; temporary recycling cannot change an address's local anchor.
// A guard reads its first access's existing operand, introducing no hidden use.
func (e *functionEmitter) memoryGuards(start, end int, proven map[int]bool) (map[int]memoryGuard, map[int]memoryAddress, uint32) {
	values := map[slots.Slot]affineAddress{}
	versions := map[slots.Slot]uint32{}
	index := map[addressRoot]int{}
	var groups []addressGroup
	sites := map[int]memoryAddress{}
	local := func(s slots.Slot) bool { return s < slots.Slot(e.f.LocalCount) && e.f.Slots[s] == wasm.I32 }
	get := func(s slots.Slot) affineAddress {
		v, ok := values[s]
		if !ok && local(s) {
			v = affineAddress{root: addressRoot{s, versions[s]}, affine: true}
		}
		if v.affine && versions[v.root.local] != v.root.version {
			v.affine = false
		}
		return v
	}
	write := func(s slots.Slot, v affineAddress) {
		if s == slots.NoSlot {
			return
		}
		if local(s) {
			versions[s]++
			v.root, v.bias, v.affine = addressRoot{s, versions[s]}, 0, true
		}
		values[s] = v
	}
	copyValue := func(dst, src slots.Slot) {
		if dst != src {
			write(dst, get(src))
		}
	}
	for pc := start; pc < end; pc++ {
		in := e.f.Code[pc]
		src := e.f.Sources[in.Source].Wasm
		a, b := get(in.A), get(in.B)
		if isMemoryAccess(in.Op) && !proven[pc] && a.affine {
			_, size := memoryAccess(in.Op)
			n := a.bias + uint64(src.B) + uint64(size)
			gi, found := index[a.root]
			if !found {
				gi = len(groups)
				index[a.root] = gi
				groups = append(groups, addressGroup{firstBias: a.bias})
			}
			g := &groups[gi]
			g.end = max(g.end, n)
			g.sites = append(g.sites, pc)
			sites[pc] = memoryAddress{a.root.local, a.bias + uint64(src.B)}
		}
		switch in.Op {
		case 0x20:
			copyValue(in.Dst, slots.Slot(src.A))
		case 0x21, 0x22:
			s := slots.Slot(src.A)
			copyValue(s, in.A)
			copyValue(in.Dst, s)
		case 0x41:
			write(in.Dst, affineAddress{constant: true, bits: uint32(src.Bits)})
		case 0x6a:
			v := affineAddress{constant: a.constant && b.constant, bits: a.bits + b.bits}
			if !a.affine || !b.constant {
				a, b = b, a
			}
			if a.affine && b.constant && a.bias+uint64(b.bits) <= math.MaxUint32 {
				v.root, v.bias, v.affine = a.root, a.bias+uint64(b.bits), true
			}
			write(in.Dst, v)
		default:
			write(in.Dst, affineAddress{})
		}
	}
	guards := map[int]memoryGuard{}
	addresses := map[int]memoryAddress{}
	var guarded uint32
	for _, g := range groups {
		if len(g.sites) < 3 {
			continue
		}
		after := g.end - g.firstBias
		if after > 1<<32 {
			continue
		}
		before := g.firstBias
		if before > (1<<32)-after {
			continue
		}
		first := g.sites[0]
		guards[first] = memoryGuard{e.f.Code[first].A, before, after}
		for _, pc := range g.sites {
			proven[pc] = true
			addresses[pc] = sites[pc]
		}
		guarded += uint32(len(g.sites))
	}
	return guards, addresses, guarded
}

// Bounds exclude wrapping so guarded accesses can fold their address arithmetic.
// Failure takes the exact checked continuation, preserving earlier stores and
// the original trap order, including valid addresses produced by i32 wrapping.
func (e *functionEmitter) emitMemoryGuard(g memoryGuard, cold int) {
	if e.a.arch == "arm64" && g.before+g.after <= 4096 {
		e.compactMemoryGuardARM64(g, cold)
		return
	}
	e.slotLoad(g.address, 0, false, false)
	if g.before != 0 {
		e.constant(1, g.before)
		e.compare(false)
		e.branch(cold, "lo")
	}
	e.constant(1, (1<<32)-g.after)
	e.compare(false)
	hi := byte(8)
	if e.a.arch == "amd64" {
		hi = 7
	}
	e.a.branch(cold, true, hi)
	if e.a.arch == "arm64" {
		e.memoryOffsetARM64(g.after)
		e.a.word(0xeb11013f) // cmp x9, x17 (bound memory length)
	} else {
		if g.after <= math.MaxInt32 {
			e.a.bytes(0x48, 0x05)
			e.a.word(uint32(g.after))
		} else {
			e.constant(1, g.after)
			e.a.bytes(0x48, 0x01, 0xc8)
		}
		e.a.bytes(0x4c, 0x39, 0xd8) // cmp rax, r11
	}
	e.a.branch(cold, true, hi)
}

// Subtracting the first bias modulo i32 makes one unsigned interval reject
// both a wrapped subtraction and a span end beyond memory32. CMN with span-1
// permits an exact 2^32 end. W9 zero-extends before the final 64-bit length check;
// the original mutable operand is unchanged on either checked fallback.
func (e *functionEmitter) compactMemoryGuardARM64(g memoryGuard, cold int) {
	r, cached := e.registers[g.address]
	if !cached {
		e.slotLoad(g.address, 0, false, false)
		r = 9
	}
	if g.before != 0 {
		e.a.word(0x51000009 | uint32(g.before)<<10 | uint32(r)<<5) // sub w9, address, bias
	} else if cached {
		e.registerMove(9, r, false)
	}
	span := uint32(g.before + g.after)
	if span > 1 {
		e.a.word(0x3100013f | (span-1)<<10) // cmn w9, span-1
		e.a.branch(cold, true, 2)           // carry means root+span > 2^32
	}
	add := span << 10
	if span == 4096 {
		add = 1<<22 | 1<<10
	}
	e.a.word(0x91000129 | add) // add x9, root, span (no narrowing at 2^32)
	e.a.word(0xeb11013f)       // cmp x9, bound memory length
	e.a.branch(cold, true, 8)
}

// All physical slots are observable at exits. A backwards killed set permits
// deleting temporary writes only until an overwrite without intervening reads.
// Guards and possible traps restore full canonical-state requirements.
func (e *functionEmitter) deadAddressTemps(start, end int, r memoryRegion) map[int]bool {
	killed, elided := map[slots.Slot]bool{}, map[int]bool{}
	read := func(s slots.Slot) { delete(killed, s) }
	write := func(s slots.Slot) {
		if s != slots.NoSlot {
			killed[s] = true
		}
	}
	for pc := end - 1; pc >= start; pc-- {
		in := e.f.Code[pc]
		if (in.Op == 0x41 || in.Op == 0x6a) && in.Dst >= slots.Slot(e.f.LocalCount) && in.Dst != slots.NoSlot && e.f.Slots[in.Dst] == wasm.I32 && killed[in.Dst] {
			elided[pc] = true
			continue
		}
		write(in.Dst)
		local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
		if in.Op == 0x21 || in.Op == 0x22 {
			write(local)
		}
		address := in.A
		if remap, ok := r.addresses[pc]; ok {
			address = remap.local
		}
		read(address)
		read(in.B)
		read(in.C)
		if in.Op == 0x20 {
			read(local)
		}
		_, guard := r.guards[pc]
		if guard || scalarCanTrap(in.Op) || isMemoryAccess(in.Op) && !r.proven[pc] {
			clear(killed)
		}
	}
	return elided
}

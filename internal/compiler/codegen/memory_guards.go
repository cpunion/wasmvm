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
	firstBias, minimum, end uint64
	sites                   []int
}

// Regions are bounded by 256 instructions. Local versions prevent stale facts
// after mutation; temporary recycling cannot change an address's local anchor.
// A guard reads its first access's existing operand, introducing no hidden use.
func (e *functionEmitter) memoryGuards(start, end int, proven map[int]bool) (map[int]memoryGuard, uint32) {
	values := map[slots.Slot]affineAddress{}
	versions := map[slots.Slot]uint32{}
	index := map[addressRoot]int{}
	var groups []addressGroup
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
				groups = append(groups, addressGroup{firstBias: a.bias, minimum: a.bias})
			}
			g := &groups[gi]
			g.minimum, g.end = min(g.minimum, a.bias), max(g.end, n)
			g.sites = append(g.sites, pc)
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
	var guarded uint32
	for _, g := range groups {
		if len(g.sites) < 3 {
			continue
		}
		after := g.end - g.firstBias
		if after > 1<<32 {
			continue
		}
		before := g.firstBias - g.minimum
		if before > (1<<32)-after {
			continue
		}
		first := g.sites[0]
		guards[first] = memoryGuard{e.f.Code[first].A, before, after}
		for _, pc := range g.sites {
			proven[pc] = true
		}
		guarded += uint32(len(g.sites))
	}
	return guards, guarded
}

// Relative bounds prove every index stays in one unsigned-i32 interval even
// when all addresses wrapped before the first access. Failure takes the exact
// checked continuation, preserving earlier stores and the original trap order.
func (e *functionEmitter) emitMemoryGuard(g memoryGuard, cold int) {
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

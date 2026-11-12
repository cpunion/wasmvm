package codegen

import (
	"encoding/binary"
	"math/bits"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Each 256-byte block covers 64 arm64 words. The second word first marks labels,
// then stores the number of deletions preceding the block. The metadata contains no pointers.
type branchLayoutBlock struct{ removed, prefix uint64 }

// Compact near conditional branches with a word bitmap and block prefix sums.
// All relocation and continuation queries take constant time without maps or
// binary searches; metadata occupies one sixteenth of the reserved code size.
func (a *assembler) compact() error {
	if a.arch != "arm64" {
		return wasm.ErrInvalid
	}
	blocks := make([]branchLayoutBlock, len(a.code)/256+1)
	for _, encoded := range a.labels {
		if encoded == 0 {
			continue
		}
		at := encoded - 1
		if at >= 0 && at <= len(a.code) && at%4 == 0 {
			blocks[at/256].prefix |= uint64(1) << ((at / 4) % 64)
		}
	}
	for _, r := range a.conditionalBranches {
		target, ok := a.labelAt(r.label)
		if !ok {
			return wasm.ErrInvalid
		}
		at := r.position + 4
		mask := uint64(1) << ((at / 4) % 64)
		if blocks[at/256].prefix&mask != 0 {
			return wasm.ErrInvalid
		}
		delta := int64(target) - int64(r.position)
		if delta%4 == 0 && delta >= -(1<<20) && delta < 1<<20 {
			blocks[at/256].removed |= mask
			binary.LittleEndian.PutUint32(a.code[r.position:], r.instruction)
		}
	}
	var removed uint64
	for i := range blocks {
		blocks[i].prefix = removed
		removed += uint64(bits.OnesCount64(blocks[i].removed))
	}
	position := func(old int) int {
		if old <= 0 {
			return old
		}
		if old >= len(a.code) {
			return old - 4*int(removed)
		}
		block := blocks[old/256]
		shift := (old % 256) / 4
		if old%4 != 0 {
			shift++
		}
		mask := (uint64(1) << shift) - 1
		return old - 4*int(block.prefix+uint64(bits.OnesCount64(block.removed&mask)))
	}
	erased := func(at int) bool {
		return at >= 0 && at < len(a.code) && at%4 == 0 && blocks[at/256].removed&(uint64(1)<<((at/4)%64)) != 0
	}
	for i, r := range a.fixups {
		if erased(r.position) {
			r.position -= 4
			r.conditional = true
		}
		r.position = position(r.position)
		a.fixups[i] = r
	}
	remaining := a.conditionalBranches[:0]
	for _, r := range a.conditionalBranches {
		if !erased(r.position + 4) {
			r.position = position(r.position)
			remaining = append(remaining, r)
		}
	}
	a.conditionalBranches = remaining
	for label, old := range a.labels {
		if old != 0 {
			a.labels[label] = position(old-1) + 1
		}
	}
	for i := range a.tables {
		a.tables[i].position = position(a.tables[i].position)
	}
	for i := range a.addresses {
		a.addresses[i].position = position(a.addresses[i].position)
	}
	old, code := a.code, a.code[:0]
	start := 0
	for i, block := range blocks {
		for mask := block.removed; mask != 0; mask &= mask - 1 {
			at := i*256 + 4*bits.TrailingZeros64(mask)
			code = append(code, old[start:at]...)
			start = at + 4
		}
	}
	a.code = append(code, old[start:]...)
	a.compactBranches = false
	if a.bounded && uint64(len(a.code)) > a.limit {
		return wasm.ErrLimit
	}
	return nil
}

package codegen

import (
	"encoding/binary"
	"sort"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Compact known near conditional branches before applying any relocations.
// Removing words can only decrease branch distances. One bounded pass suffices:
// branches that become near afterwards may safely keep their long fallback.
// Labels, ADRs, tables and continuation offsets share the same position map.
func (a *assembler) compact() error {
	if a.arch != "arm64" {
		return wasm.ErrInvalid
	}
	labels := make(map[int]bool, len(a.labels))
	for _, position := range a.labels {
		labels[position] = true
	}
	removed := make([]int, 0, len(a.conditionalBranches))
	short := make(map[int]conditionalRelocation)
	for _, r := range a.conditionalBranches {
		target, ok := a.labels[r.label]
		if !ok {
			return wasm.ErrInvalid
		}
		// No label may enter the second word of a reserved branch pair.
		if labels[r.position+4] {
			return wasm.ErrInvalid
		}
		delta := int64(target) - int64(r.position)
		if delta%4 == 0 && delta >= -(1<<20) && delta < 1<<20 {
			removed = append(removed, r.position+4)
			short[r.position+4] = r
			binary.LittleEndian.PutUint32(a.code[r.position:], r.instruction)
		}
	}
	sort.Ints(removed)
	position := func(old int) int { return old - 4*sort.SearchInts(removed, old) }
	for i, r := range a.fixups {
		if branch, ok := short[r.position]; ok {
			r.position, r.conditional = branch.position, true
		}
		r.position = position(r.position)
		a.fixups[i] = r
	}
	remaining := a.conditionalBranches[:0]
	for _, r := range a.conditionalBranches {
		if _, ok := short[r.position+4]; !ok {
			r.position = position(r.position)
			remaining = append(remaining, r)
		}
	}
	a.conditionalBranches = remaining
	for label, old := range a.labels {
		a.labels[label] = position(old)
	}
	for i := range a.tables {
		a.tables[i].position = position(a.tables[i].position)
	}
	for i := range a.addresses {
		a.addresses[i].position = position(a.addresses[i].position)
	}
	old, code := a.code, a.code[:0]
	start := 0
	for _, at := range removed {
		code = append(code, old[start:at]...)
		start = at + 4
	}
	a.code = append(code, old[start:]...)
	a.compactBranches = false
	if a.bounded && uint64(len(a.code)) > a.limit {
		return wasm.ErrLimit
	}
	return nil
}

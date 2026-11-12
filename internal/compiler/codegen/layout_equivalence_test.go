package codegen

import (
	"bytes"
	"testing"
)

func TestBitmapLayoutMatchesIndependentSearchRelocations(t *testing.T) {
	for seed := 0; seed < 64; seed++ {
		a := assembler{arch: "arm64", labels: map[int]int{}, longBranches: true, compactBranches: true}
		var labels [8]int
		for i := range labels {
			labels[i] = a.label()
		}
		a.mark(labels[0])
		state := uint32(seed + 1)
		for i := 0; i < 32; i++ {
			if i%4 == 0 {
				a.mark(labels[i/4])
			}
			state = state*1664525 + 1013904223
			target := labels[int(state>>8)%len(labels)]
			if state&1 != 0 {
				a.branch(target, true, byte(state>>16)%14)
			} else {
				a.compareBranch(target, byte(state>>16)%29, state&2 != 0, state&4 != 0)
			}
			a.addresses = append(a.addresses, addressRelocation{len(a.code), target})
			a.word(0)
			a.tables = append(a.tables, tableRelocation{len(a.code), target, labels[0]})
			a.word(0)
			a.word(0xd503201f)
		}
		// Relocation discovery order does not define physical instruction order.
		if seed%2 != 0 {
			for l, r := 0, len(a.conditionalBranches)-1; l < r; l, r = l+1, r-1 {
				a.conditionalBranches[l], a.conditionalBranches[r] = a.conditionalBranches[r], a.conditionalBranches[l]
			}
		}
		b := a
		b.code = append([]byte(nil), a.code...)
		b.labels = map[int]int{}
		for k, v := range a.labels {
			b.labels[k] = v
		}
		b.fixups = append([]relocation(nil), a.fixups...)
		b.conditionalBranches = append([]conditionalRelocation(nil), a.conditionalBranches...)
		b.addresses = append([]addressRelocation(nil), a.addresses...)
		b.tables = append([]tableRelocation(nil), a.tables...)
		got, err := a.finish()
		if err != nil {
			t.Fatal(seed, err)
		}
		if err := b.referenceCompact(); err != nil {
			t.Fatal(seed, err)
		}
		want, err := b.finish()
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("changed instruction/relocation bytes", seed, err)
		}
		for label, want := range b.labels {
			if a.labels[label] != want {
				t.Fatal("changed continuation", seed, label, a.labels[label], want)
			}
		}
	}
}

package slots

import "testing"

func TestPrefixHintsKeepCopiesSourcesAndConstantWords(t *testing.T) {
	f := &Function{Code: []Instruction{
		{Op: 0x20, Source: 10, Dst: 0, Extra: 0},
		{Op: 0x20, Source: 11, Dst: 2, Extra: 1},
		{Op: 0x20, Source: 12, Dst: NoSlot, Extra: uint32(NoSlot)},
		{Op: 0x41, Source: 13, Dst: 1, Extra: 0xffffffff},
		{Op: 0x6a, Source: 14, Dst: 2, A: 0, B: 1},
		{Op: 0x42, Source: 15, Dst: 1, Extra: 0xabcdef01, Count: 0x87654321},
		{Op: 0x44, Source: 16, Dst: 1, Extra: 0x12345678, Count: 0x7ff80001},
		{Op: 0x20, Source: 17, Dst: 0, Extra: 0},
	}}
	before := append([]Instruction(nil), f.Code...)
	b := budget{nodes: uint64(len(f.Code))}
	cachePrefixes(f, &b)
	if b.nodes != 0 || len(f.Code) != len(before) {
		t.Fatal("prefix scan changed its bounded work or source count")
	}
	for pc, original := range before {
		want := original
		if pc == 0 || pc == 3 {
			want.Count = 1
		}
		if f.Code[pc] != want {
			t.Fatal("incorrect hint or changed source metadata", pc, f.Code[pc], want)
		}
	}
}

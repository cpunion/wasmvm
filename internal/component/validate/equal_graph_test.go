package validate

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func equalityChain(size int, cycle bool) *model.Type {
	nodes := make([]model.Type, size)
	for i := range nodes {
		nodes[i].Kind = model.List
		if i != 0 {
			nodes[i-1].Element = &nodes[i]
		}
	}
	if cycle {
		nodes[size-1].Element = &nodes[0]
	} else {
		nodes[size-1].Kind = model.U8
	}
	return &nodes[0]
}

func TestEqualityResolvedGraphBudgetAndCycles(t *testing.T) {
	for _, size := range []int{1, 16, 17, 31, 32, 33, 64, 129} {
		for _, cycle := range []bool{false, true} {
			t.Run(fmt.Sprintf("nodes-%d/cycle-%v", size, cycle), func(t *testing.T) {
				a, b := equalityChain(size, cycle), equalityChain(size, cycle)
				for _, budget := range []uint32{0, uint32(size - 1), uint32(size)} {
					ok, err := EqualValue(a, b, budget, func(uint64, uint64) bool {
						t.Fatal("resource-free graph consulted nominal identities")
						return false
					})
					if budget < uint32(size) {
						if ok || !errors.Is(err, wasm.ErrLimit) {
							t.Fatal(budget, ok, err)
						}
					} else if !ok || err != nil {
						t.Fatal(budget, ok, err)
					}
				}
				if !cycle {
					leaf := b
					for leaf.Element != nil {
						leaf = leaf.Element
					}
					leaf.Kind = model.U16
					// A shape mismatch precedes the node charge, including at
					// the exact budget boundary after a large shared prefix.
					if ok, err := EqualValue(a, b, uint32(size-1), nil); ok || err != nil {
						t.Fatal(ok, err)
					}
				}
			})
		}
	}
}

func TestEqualityNominalPredicateBeforeMemoizedPairs(t *testing.T) {
	a := &model.Type{Kind: model.Own, ResourceID: 1}
	b := &model.Type{Kind: model.Own, ResourceID: 2}
	left := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: a}, {Type: a}}}
	right := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: b}, {Type: b}}}
	for _, rejectSecond := range []bool{false, true} {
		calls := 0
		ok, err := EqualValue(left, right, 4, func(x, y uint64) bool {
			calls++
			if x != 1 || y != 2 {
				t.Fatal(x, y)
			}
			return !rejectSecond || calls != 2
		})
		if err != nil || ok == rejectSecond || calls != 2 {
			t.Fatal(rejectSecond, ok, err, calls)
		}
	}
	// An identical schema pointer still needs its current nominal binding.
	for _, accept := range []bool{true, false, true} {
		calls := 0
		ok, err := EqualValue(a, a, 1, func(uint64, uint64) bool { calls++; return accept })
		if err != nil || ok != accept || calls != 1 {
			t.Fatal(accept, ok, err, calls)
		}
	}
	if ok, err := EqualValue(a, a, 0, func(uint64, uint64) bool { return true }); ok || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(ok, err)
	}
	if ok, err := EqualValue(a, a, 0, nil); !ok || err != nil {
		t.Fatal(ok, err)
	}
}

func BenchmarkEqualityResolvedGraphs(b *testing.B) {
	for _, size := range []int{2, 16, 17, 64} {
		b.Run(fmt.Sprintf("nodes-%d", size), func(b *testing.B) {
			a, z := equalityChain(size, false), equalityChain(size, false)
			resolve := func(x, y uint64) bool { return x == y }
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if ok, err := EqualValue(a, z, uint32(size), resolve); !ok || err != nil {
					b.Fatal(ok, err)
				}
			}
		})
	}
}

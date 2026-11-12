package validate

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func checkComparisonPlan(t *testing.T, a, b *model.Type, budget uint32) {
	t.Helper()
	var plans ValueComparisons
	plan := plans.Get(a, b, budget)
	if plan == nil {
		t.Fatal("small program was not compiled")
	}
	for _, reject := range []int{0, 1, 2, 3, 10} {
		var wantCalls, gotCalls [][2]uint64
		want, wantErr := EqualValue(a, b, budget, func(a, b uint64) bool {
			wantCalls = append(wantCalls, [2]uint64{a, b})
			return len(wantCalls) != reject
		})
		got, gotErr := plan.Match(func(a, b uint64) bool {
			gotCalls = append(gotCalls, [2]uint64{a, b})
			return len(gotCalls) != reject
		})
		if got != want || (gotErr == nil) != (wantErr == nil) || errors.Is(gotErr, wasm.ErrLimit) != errors.Is(wantErr, wasm.ErrLimit) || !reflect.DeepEqual(gotCalls, wantCalls) {
			t.Fatalf("budget=%d reject=%d: got %v/%v/%v, want %v/%v/%v", budget, reject, got, gotErr, gotCalls, want, wantErr, wantCalls)
		}
	}
	if plans.Get(a, b, budget) != plan {
		t.Fatal("repeated pair did not reuse immutable program")
	}
}

func TestValueComparisonProgramsPreserveNominalOrderAndLimits(t *testing.T) {
	for _, size := range []int{1, 16, 32, 33, 64, 129} {
		for _, cycle := range []bool{false, true} {
			a, b := equalityChain(size, cycle), equalityChain(size, cycle)
			for _, budget := range []uint32{0, uint32(size - 1), uint32(size), uint32(size + 1)} {
				checkComparisonPlan(t, a, b, budget)
			}
		}
	}
	a, b := &model.Type{Kind: model.Own, ResourceID: 1}, &model.Type{Kind: model.Own, ResourceID: 2}
	left := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Name: "a", Type: a}, {Name: "b", Type: a}}}
	right := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Name: "a", Type: b}, {Name: "b", Type: b}}}
	for budget := uint32(0); budget <= 6; budget++ {
		checkComparisonPlan(t, a, a, budget)
		checkComparisonPlan(t, left, right, budget)
	}
	// A nominal failure before a later structural mismatch/limit must win.
	right.Fields[1].Name = "other"
	for budget := uint32(0); budget <= 6; budget++ {
		checkComparisonPlan(t, left, right, budget)
	}
	checkComparisonPlan(t, nil, nil, 0)
	checkComparisonPlan(t, a, nil, 0)
	var plans ValueComparisons
	p := plans.Get(a, b, 1)
	if !p.HasResources() {
		t.Fatal("nominal obligation was discarded")
	}
	if ok, err := p.Match(nil); ok || err != nil {
		t.Fatal("unresolved resource matched", ok, err)
	}
}

func TestValueComparisonRandomSharedGraphs(t *testing.T) {
	r := rand.New(rand.NewSource(2026111213))
	for round := 0; round < 100; round++ {
		a, b := make([]model.Type, 8), make([]model.Type, 8)
		for i := range a {
			a[i] = model.Type{Kind: model.Kind(0x63 + r.Intn(29)), Length: uint32(r.Intn(3))}
			if r.Intn(3) == 0 {
				a[i].ResourceID = uint64(1 + r.Intn(4))
			}
			b[i] = a[i]
			if r.Intn(3) == 0 {
				b[i].ResourceID++
			}
			if r.Intn(4) == 0 {
				a[i].Labels, b[i].Labels = []string{"first", "second"}, []string{"first", "second"}
			}
			j := r.Intn(len(a))
			a[i].Element, b[i].Element = &a[j], &b[j]
			if r.Intn(2) == 0 {
				j = r.Intn(len(a))
				a[i].Key, b[i].Key = &a[j], &b[j]
			}
		}
		if round%3 == 0 {
			b[r.Intn(len(b))].Length++
		}
		for _, budget := range []uint32{0, 1, 3, 8, 20} {
			checkComparisonPlan(t, &a[0], &b[0], budget)
		}
	}
}

func TestValueComparisonBoundsFreshErrorsAndClear(t *testing.T) {
	var plans ValueComparisons
	typ := &model.Type{Kind: model.U8}
	limited := plans.Get(typ, typ, 0)
	_, firstErr := limited.Match(nil)
	if _, err := limited.Match(nil); !errors.Is(firstErr, wasm.ErrLimit) || !errors.Is(err, wasm.ErrLimit) || err == firstErr {
		t.Fatal("comparison reused a cached error object", firstErr, err)
	}
	if p := plans.Get(typ, typ, 1); p == limited {
		t.Fatal("different comparison budgets shared a program")
	}
	for plans.Len() < maxValueComparisons {
		a := &model.Type{Kind: model.U8}
		if plans.Get(a, a, 1) == nil {
			t.Fatal("capacity rejected early")
		}
	}
	if plans.Get(&model.Type{Kind: model.U8}, typ, 1) != nil {
		t.Fatal("exceeded program bound")
	}
	if plans.Get(typ, typ, 0) != limited {
		t.Fatal("capacity discarded an existing program")
	}
	plans.Clear()
	if plans.Len() != 0 || plans.last != nil || plans.nominal != 0 {
		t.Fatal("clear retained type roots")
	}
	resource := &model.Type{Kind: model.Own, ResourceID: 1}
	large := &model.Type{Kind: model.Tuple, Fields: make([]model.Field, maxComparisonNominals)}
	for i := range large.Fields {
		large.Fields[i].Type = resource
	}
	if plans.Get(large, large, 10000) == nil || plans.nominal != maxComparisonNominals {
		t.Fatal("exact nominal bound")
	}
	if plans.Get(resource, resource, 1) != nil {
		t.Fatal("exceeded nominal retention bound")
	}
	if plans.Get(typ, typ, 1) == nil {
		t.Fatal("resource-free plan consumed nominal storage")
	}
}

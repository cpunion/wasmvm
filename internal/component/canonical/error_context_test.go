package canonical

import (
	"errors"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestErrorContextTaggedHandlesAndSharedBudget(t *testing.T) {
	b := NewResourceBudgetWithErrors(4, 5)
	a, other := NewResourceTable(b), NewResourceTable(b)
	v, err := NewErrorContext("four")
	if err != nil {
		t.Fatal(err)
	}
	rt := &ResourceType{}
	resource, _ := a.New(rt, 42)
	x, err := a.AddErrorContext(v)
	if err != nil {
		t.Fatal(err)
	}
	copy := *v.Resource
	y, err := other.AddErrorContext(Value{Resource: &copy})
	if err != nil || b.errorBytes != 4 {
		t.Fatal(y, b.errorBytes, err)
	}
	if _, err := a.ErrorContext(resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := a.Rep(rt, x); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := a.DropErrorContext(resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	tooMuch, _ := NewErrorContext("xx")
	if _, err := a.AddErrorContext(tooMuch); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	got, err := a.ErrorContext(x)
	if err != nil || !SameErrorContext(got.Resource, v.Resource) {
		t.Fatal(got, err)
	}
	if err := a.DropErrorContext(x); err != nil || b.errorBytes != 4 {
		t.Fatal(err, b.errorBytes)
	}
	if err := a.DropErrorContext(x); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	other.Close()
	other.Close()
	if b.errorBytes != 0 || b.live.Load() != 1 {
		t.Fatal(b.errorBytes, b.live.Load())
	}
	if _, err := other.ErrorContext(y); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := other.AddErrorContext(v); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	a.Close()
	if b.live.Load() != 0 {
		t.Fatal(b.live.Load())
	}
	if message, err := ErrorContextMessage(got); err != nil || message != "four" {
		t.Fatal(message, err)
	}
}

func TestErrorContextLimitsAndHostShape(t *testing.T) {
	if _, err := NewErrorContext("\xff"); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	for _, v := range []Value{{}, {Resource: &Reference{}}, {Resource: NewReference(nil, nil, 0, nil, nil)}} {
		if _, err := ErrorContextMessage(v); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		if _, err := NewResources(1).AddErrorContext(v); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	v, _ := NewErrorContext("abc")
	for _, limit := range []uint32{0, 1} {
		b := NewResourceBudgetWithErrors(limit, 3)
		r := NewResourceTable(b)
		if limit == 1 {
			r.New(&ResourceType{}, 0)
		}
		if _, err := r.AddErrorContext(v); !errors.Is(err, wasm.ErrLimit) || b.errorBytes != 0 {
			t.Fatal(err, b.errorBytes)
		}
		r.Close()
	}
	r := NewResources(2)
	r.next = 1 << 28
	if _, err := r.AddErrorContext(v); !errors.Is(err, wasm.ErrLimit) || r.budget.errorBytes != 0 {
		t.Fatal(err)
	}
	r = NewResources(1)
	r.budget.errors = map[*errorContext]uint32{v.Resource.errorContext: ^uint32(0)}
	if _, err := r.AddErrorContext(v); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	r = NewResourceTable(NewResourceBudgetWithErrors(1, 0))
	empty, _ := NewErrorContext("")
	if _, err := r.AddErrorContext(empty); err != nil {
		t.Fatal(err)
	}
	r.Close()
	c := valueCodec(false)
	typ := &model.Type{Kind: model.ErrorContext}
	cloned, err := c.Clone(typ, v)
	if err != nil || !SameErrorContext(cloned.Resource, v.Resource) {
		t.Fatal(cloned, err)
	}
	if _, err := c.LowerFlat(typ, v); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := c.LiftFlat(typ, []Slot{{Type: wasm.I32, Bits: 1}}); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := c.Load(typ, 0); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if err := c.Store(typ, 0, v); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if SameErrorContext(nil, v.Resource) || SameErrorContext(&Reference{}, v.Resource) {
		t.Fatal("non-context equality")
	}
}

func TestErrorContextConcurrentNamespaces(t *testing.T) {
	b := NewResourceBudgetWithErrors(8, 3)
	v, _ := NewErrorContext("abc")
	var wg sync.WaitGroup
	for range 6 {
		wg.Go(func() {
			r := NewResourceTable(b)
			defer r.Close()
			for range 100 {
				id, err := r.AddErrorContext(v)
				if err != nil {
					t.Error(err)
					return
				}
				if err := r.DropErrorContext(id); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
	if b.live.Load() != 0 || b.errorBytes != 0 {
		t.Fatal(b.live.Load(), b.errorBytes)
	}
}

func TestErrorContextHostCarrierReassignment(t *testing.T) {
	r := NewResourceTable(NewResourceBudgetWithErrors(2, 4))
	v, _ := NewErrorContext("keep")
	replacement, _ := NewErrorContext("other")
	id, err := r.AddErrorContext(v)
	if err != nil {
		t.Fatal(err)
	}
	c := valueCodec(false)
	cloned, err := c.Clone(&model.Type{Kind: model.ErrorContext}, v)
	if err != nil {
		t.Fatal(err)
	}
	*v.Resource = *replacement.Resource
	for _, value := range []Value{cloned} {
		message, err := ErrorContextMessage(value)
		if err != nil || message != "keep" {
			t.Fatal(message, err)
		}
	}
	got, err := r.ErrorContext(id)
	if err != nil {
		t.Fatal(err)
	}
	message, err := ErrorContextMessage(got)
	if err != nil || message != "keep" {
		t.Fatal(message, err)
	}
	*got.Resource = Reference{}
	again, err := r.ErrorContext(id)
	if err != nil {
		t.Fatal(err)
	}
	message, err = ErrorContextMessage(again)
	if err != nil || message != "keep" {
		t.Fatal(message, err)
	}
	if err := r.DropErrorContext(id); err != nil || r.budget.errorBytes != 0 {
		t.Fatal(err, r.budget.errorBytes)
	}
}

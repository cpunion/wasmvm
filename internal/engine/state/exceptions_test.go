package state

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExceptionQuotasDomainsAndRetirement(t *testing.T) {
	limits := DefaultLimits()
	limits.MaxTags = 1
	limits.MaxHeapObjects = 1
	s := New(limits)
	defer s.Close()
	sig := module.FuncType{Params: []wasm.ValueType{wasm.ExnRef}}
	tag, err := s.NewTag(sig)
	if err != nil {
		t.Fatal(err)
	}
	sig.Params[0] = wasm.I32
	if tag.Tag.Signature.Params[0] != wasm.ExnRef {
		t.Fatal("aliased signature")
	}
	if _, err := s.NewTag(module.FuncType{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.NewTag(module.FuncType{Results: []wasm.ValueType{wasm.I32}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.NewException(nil, nil); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := s.NewException(tag.Tag, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.NewException(tag.Tag, []value.Value{{Lo: 123}}); !errors.Is(err, ErrReference) || len(s.temporary) != 0 {
		t.Fatal(err)
	}
	ex, err := s.NewException(tag.Tag, []value.Value{{}})
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(ex, wasm.ErrException) || ex.Error() != wasm.ErrException.Error() {
		t.Fatal(ex)
	}
	if s.TestReference(ex.Ref, wasm.AnyRef) || s.TestReference(I31(1), wasm.ExnRef) || !s.TestReference(ex.Ref, wasm.ExnRef.NonNull()) || s.TestReference(ex.Ref, wasm.NoExnRef) {
		t.Fatal("exception heap domain")
	}
	if got, ok := s.ExceptionError(ex); !ok || got != ex {
		t.Fatal(got, ok)
	}
	other := New(DefaultLimits())
	defer other.Close()
	if _, ok := other.ExceptionError(ex); ok {
		t.Fatal("foreign exception")
	}
	if _, err := other.NewException(tag.Tag, []value.Value{{}}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	s.CollectEveryAllocation = true
	next, err := s.NewException(tag.Tag, []value.Value{{Lo: ex.Ref}}, func(visit func(wasm.ValueType, value.Value)) { visit(wasm.ExnRef, value.Value{Lo: ex.Ref}) })
	if !errors.Is(err, wasm.ErrLimit) || next != nil {
		t.Fatal("rooted allocation exceeded quota", err)
	}
	s.CollectReferences(nil, nil, nil)
	if _, ok := s.ExceptionError(ex); ok {
		t.Fatal("stale exception")
	}
	if _, _, err := s.Exception(ex.Ref); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	next, err = s.NewException(tag.Tag, []value.Value{{}})
	if err != nil || next.Ref == ex.Ref {
		t.Fatal(next, err)
	}
	s.Close()
	if _, err := s.NewTag(module.FuncType{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.NewException(tag.Tag, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

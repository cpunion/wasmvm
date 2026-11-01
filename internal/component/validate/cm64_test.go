package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCM64WorstCaseBorrowSize(t *testing.T) {
	for _, representation := range []wasm.ValueType{wasm.I32, wasm.I64, 0} {
		s := scope(t)
		if representation == 0 {
			if _, err := s.Abstract(); err != nil {
				t.Fatal(err)
			}
		} else {
			define(t, s, &model.TypeDef{Kind: model.Resource, Representation: representation})
		}
		define(t, s, &model.TypeDef{Kind: model.Borrow, Resource: 0})
		def := &model.TypeDef{Kind: model.FixedList, Element: idx(1), Length: 1 << 25}
		_, err := s.Define(def, nil)
		if representation == wasm.I32 {
			if err != nil {
				t.Fatal("known i32 representation was widened", err)
			}
		} else if !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("wide or abstract borrow exceeded the worst-case type bound", err)
		}
	}
}

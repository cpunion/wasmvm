package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLiftPlans(t *testing.T) {
	for _, name := range []string{"async-lift", "async-context64", "async-values", "async-values-utf16", "async-values-compact", "async-values64", "async-host", "async-resources"} {
		bindingSchema(t, name)
	}
	input, err := os.ReadFile("../../../pkg/component/testdata/async-lift.wasm")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"sync-type", "callback-only", "post", "callback-index", "callback-signature", "context-type", "context-slot", "context-width", "return-type", "return-realloc", "return-post", "return-async", "return-callback"} {
		t.Run(mutation, func(t *testing.T) {
			binary, err := decode.Decode(input, decode.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			var lift, ret, get, set *model.Canonical
			var async *model.TypeDef
			for _, d := range binary.Definitions {
				if d.Type != nil && d.Type.Kind == model.AsyncFunction && async == nil {
					async = d.Type
				}
				if d.Canonical == nil {
					continue
				}
				p := d.Canonical
				switch p.Op {
				case 0:
					if lift == nil {
						lift = p
					}
				case 9:
					if ret == nil {
						ret = p
					}
				case 0x0a:
					get = p
				case 0x0b:
					set = p
				}
			}
			switch mutation {
			case "sync-type":
				async.Kind = model.Function
			case "callback-only":
				lift.Options = lift.Options[1:]
			case "post":
				lift.Options = append(lift.Options, model.CanonicalOption{Kind: 5})
			case "callback-index":
				lift.Options[1].Index = 9999
			case "callback-signature":
				lift.Options[1].Index = 0
			case "context-type":
				get.CoreValue = wasm.F32
			case "context-slot":
				get.Index = 2
			case "context-width":
				set.CoreValue = wasm.I64
			case "return-type":
				ret.Results = &model.TypeRef{Indexed: true, Index: 9999}
			case "return-realloc":
				ret.Options = []model.CanonicalOption{{Kind: 4}}
			case "return-post":
				ret.Options = []model.CanonicalOption{{Kind: 5}}
			case "return-async":
				ret.Options = []model.CanonicalOption{{Kind: 6}}
			case "return-callback":
				ret.Options = []model.CanonicalOption{{Kind: 7}}
			}
			if _, err := Check(binary, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(mutation, err)
			}
		})
	}
}

func TestTaskReturnValidationAndBounds(t *testing.T) {
	for _, borrow := range []bool{false, true} {
		kind := model.Own
		if borrow {
			kind = model.Borrow
		}
		b := &model.Binary{Definitions: []model.Definition{
			{Kind: 7, Type: &model.TypeDef{Kind: model.Resource, Representation: wasm.I32}},
			{Kind: 7, Type: &model.TypeDef{Kind: kind}},
			{Kind: 8, Canonical: &model.Canonical{Op: 9, Results: &model.TypeRef{Indexed: true, Index: 1}}},
		}}
		_, err := Check(b, DefaultLimits())
		if borrow && !errors.Is(err, wasm.ErrInvalid) || !borrow && err != nil {
			t.Fatal(borrow, err)
		}
		if !borrow {
			passed := false
			for budget := uint32(1); budget < 90; budget++ {
				_, err := Check(b, Limits{MaxNodes: budget, MaxDepth: 128})
				if err == nil {
					passed = true
				} else if !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(budget, err)
				}
			}
			if !passed {
				t.Fatal("no successful budget")
			}
		}
	}
	input, err := os.ReadFile("../../../pkg/component/testdata/async-values.wasm")
	if err != nil {
		t.Fatal(err)
	}
	b, err := decode.Decode(input, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range b.Definitions {
		if d.Canonical != nil && d.Canonical.Op == 9 {
			d.Canonical.Options = nil
			break
		}
	}
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("string return without memory", err)
	}
}

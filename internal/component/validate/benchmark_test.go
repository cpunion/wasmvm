package validate

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
)

func BenchmarkResolveTypeDeclarations(b *testing.B) {
	definitions := []*model.TypeDef{{Kind: model.Record, Fields: []model.NamedType{
		{Name: "number", Type: model.TypeRef{Primitive: model.U32}},
		{Name: "text", Type: model.TypeRef{Primitive: model.String}},
	}}}
	for i := uint32(0); i < 20; i++ {
		definitions = append(definitions, &model.TypeDef{Kind: model.List, Element: &model.TypeRef{Indexed: true, Index: i}})
	}
	definitions = append(definitions, &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{
		{Name: "input", Type: model.TypeRef{Indexed: true, Index: 20}},
	}, Result: &model.TypeRef{Indexed: true, Index: 0}})
	// Time fresh resolution, budgets, metadata ownership and memory64 layouts.
	// Input construction, binary decoding and process/build costs are separate.
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewTypePool(DefaultLimits())
		s, err := p.Scope(nil, true)
		if err != nil {
			b.Fatal(err)
		}
		for _, d := range definitions {
			if _, err := s.Define(d, nil); err != nil {
				b.Fatal(err)
			}
		}
	}
}

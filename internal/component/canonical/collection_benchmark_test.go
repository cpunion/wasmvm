package canonical

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
)

func BenchmarkCollectionConversion(b *testing.B) {
	for _, name := range []string{"map", "fixed-list"} {
		b.Run(name, func(b *testing.B) {
			typ := &model.Type{Kind: model.Map, Key: &model.Type{Kind: model.String}, Element: &model.Type{Kind: model.U64}}
			value := Value{Elements: []Value{{Elements: []Value{{Text: "key"}, {Bits: 42}}}, {Elements: []Value{{Text: "key"}, {Bits: 43}}}}}
			if name == "fixed-list" {
				typ = &model.Type{Kind: model.FixedList, Element: &model.Type{Kind: model.U64}, Length: 8}
				value.Elements = make([]Value, 8)
			}
			transfer, next := textTransfer(UTF8, false)
			builder := NewBuilder(false, DefaultLimits())
			if _, err := builder.Build(typ); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				*next, transfer.used = 64, 0
				codec := Codec{Transfer: transfer, Layouts: builder, MaxValues: 1024}
				slots, err := codec.LowerFlat(typ, value)
				if err != nil {
					b.Fatal(err)
				}
				got, err := codec.LiftFlat(typ, slots)
				if err != nil || len(got.Elements) != len(value.Elements) {
					b.Fatal(got, err)
				}
			}
		})
	}
}

package canonical

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
)

// BenchmarkFlatRecord excludes immutable layout construction and exercises the
// direct canonical value path without guest memory or resource operations.
func BenchmarkFlatRecord(b *testing.B) {
	tpe := &model.Type{Kind: model.Record, Fields: []model.Field{{Name: "a", Type: &model.Type{Kind: model.U32}}, {Name: "b", Type: &model.Type{Kind: model.F64}}}}
	v := Value{Elements: []Value{{Bits: 42}, {Bits: 0x3ff0000000000000}}}
	layouts := NewBuilder(false, DefaultLimits())
	if _, err := layouts.Build(tpe); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transfer := Transfer{MaxBytes: 1024}
		c := Codec{Transfer: &transfer, Layouts: layouts, MaxValues: 16}
		slots, err := c.LowerFlat(tpe, v)
		if err != nil {
			b.Fatal(err)
		}
		got, err := c.LiftFlat(tpe, slots)
		if err != nil || got.Elements[0].Bits != 42 || got.Elements[1].Bits != v.Elements[1].Bits {
			b.Fatal(got, err)
		}
	}
}

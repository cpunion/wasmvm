package canonical

import "testing"

// This probe isolates table allocation/lookup/drop from Core dispatch and guest
// destructors. It is deliberately separate from the component lifetime probe.
func BenchmarkResourceTable(b *testing.B) {
	table := NewResources(1)
	typ := &ResourceType{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id, err := table.New(typ, 7)
		if err != nil {
			b.Fatal(err)
		}
		rep, err := table.Rep(typ, id)
		if err != nil || rep != 7 {
			b.Fatal(rep, err)
		}
		if err := table.Drop(nil, typ, id); err != nil {
			b.Fatal(err)
		}
	}
}

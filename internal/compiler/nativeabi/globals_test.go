package nativeabi

import "testing"

func TestGlobalPointerOffset(t *testing.T) {
	for _, test := range []struct {
		functions, tables, index, offset uint32
		ok                               bool
	}{{0, 0, 0, 0, true}, {3, 2, 1, 64, true}, {0, 0, 268435455, 2147483640, true},
		{268435456, 0, 0, 2147483648, false}, {0, 134217728, 0, 2147483648, false}, {0, 0, 268435456, 2147483648, false}} {
		offset, ok := GlobalPointerOffset(test.functions, test.tables, test.index)
		if offset != test.offset || ok != test.ok {
			t.Fatal(test, offset, ok)
		}
	}
}

func TestTablePairOffset(t *testing.T) {
	for _, test := range []struct {
		functions, index, offset uint32
		ok                       bool
	}{
		{0, 0, 0, true}, {3, 2, 56, true}, {268435454, 0, 2147483632, true},
		{268435455, 0, 2147483640, false}, {0, 134217727, 2147483632, true}, {0, 134217728, 2147483648, false},
	} {
		offset, ok := TablePairOffset(test.functions, test.index)
		if offset != test.offset || ok != test.ok {
			t.Fatal(test, offset, ok)
		}
	}
}

package decode

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompactImportsFlattenIntoOriginalIndexOrder(t *testing.T) {
	flat := []byte{4, 1, 'm', 1, 'a', 0, 0, 1, 'm', 1, 'b', 3, 0x7f, 1, 1, 'm', 1, 'c', 0, 0, 1, 'm', 1, 'd', 0, 0}
	compact := []byte{2, 1, 'm', 0, 0x7f, 2, 1, 'a', 0, 0, 1, 'b', 3, 0x7f, 1, 1, 'm', 0, 0x7e, 0, 0, 2, 1, 'c', 1, 'd'}
	want, err := Decode(fixture(section(2, flat...)), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	input := fixture(section(2, compact...))
	if _, err := Decode(input, DefaultLimits()); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal("compact imports enabled by default", err)
	}
	got, err := DecodeWithFeatures(input, DefaultLimits(), wasm.CompactImports)
	if err != nil || !reflect.DeepEqual(got.Imports, want.Imports) {
		t.Fatal(got, want, err)
	}
}

func TestCompactImportsRejectMalformedGroupsAndBoundExpansion(t *testing.T) {
	for _, body := range [][]byte{
		{1, 0, 1, 'x', 0x7f, 0}, {1, 0, 1, 'x', 0x7e, 0, 0, 0},
		{1, 0, 0, 0x7f}, {1, 0, 0, 0x7e}, {1, 0, 0, 0x7e, 0xff},
		{1, 0, 0, 0x7f, 1, 0, 0xff}, {1, 0, 0, 0x7f, 1, 0},
		{1, 0, 0, 0x7f, 0xff, 0xff, 0xff, 0xff, 0x0f},
	} {
		if _, err := DecodeWithFeatures(fixture(section(2, body...)), DefaultLimits(), wasm.CompactImports); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(body, err)
		}
	}
	// Empty groups do not create phantom imports, including a shared descriptor.
	for _, body := range [][]byte{{1, 0, 0, 0x7f, 0}, {1, 0, 0, 0x7e, 0, 0, 0}} {
		m, err := DecodeWithFeatures(fixture(section(2, body...)), DefaultLimits(), wasm.CompactImports)
		if err != nil || len(m.Imports) != 0 {
			t.Fatal(m, err)
		}
	}
	input := fixture(section(2, 1, 0, 0, 0x7e, 3, 0x7f, 0, 4, 0, 0, 0, 0))
	limits := DefaultLimits()
	limits.MaxNodes = 4
	if _, err := DecodeWithFeatures(input, limits, wasm.CompactImports); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("expanded imports bypass the node quota", err)
	}
}

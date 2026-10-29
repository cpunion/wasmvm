package decode

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestEmbeddedCoreFragmentsUseTheSameGrammar(t *testing.T) {
	m := new(module.Module)
	r := binary.New([]byte{0x60, 1, 0x7f, 1, 0x7e})
	f := Fragment{Reader: &r, Types: m, Limits: DefaultLimits()}
	if n, err := f.TypeGroup(); err != nil || n != 3 || len(m.Types) != 1 || m.Types[0].Results[0] != wasm.I64 || r.Remaining() != 0 {
		t.Fatal(n, m, err)
	}
	r = binary.New([]byte{0x4e, 2, 0x5f, 1, 0x63, 1, 1, 0x5e, 0x78, 0})
	m = new(module.Module)
	f.Types = m
	f.Features = wasm.SupportedFeatures
	if _, err := f.TypeGroup(); err != nil || len(m.Types) != 2 || m.TypeDefs[0].Kind != module.StructKind || m.TypeDefs[1].Fields[0].Packed != 8 {
		t.Fatal(m, err)
	}
	if m.TypeDefs[0].Fields[0].Value != wasm.IndexedReference(wasm.ArrayRef, 1, true) {
		t.Fatal("unresolved recursive family", m.TypeDefs[0].Fields[0])
	}
	for _, row := range []struct {
		input []byte
		kind  byte
	}{
		{[]byte{0, 0}, 0}, {[]byte{1, 0x70, 1, 0, 1}, 1}, {[]byte{2, 3, 1, 2}, 2}, {[]byte{3, 0x7f, 1}, 3}, {[]byte{4, 0, 0}, 4},
	} {
		r = binary.New(row.input)
		desc, _, err := f.ExternType()
		if err != nil || desc.Kind != row.kind || r.Remaining() != 0 {
			t.Fatal(row, desc, err)
		}
	}
	r = binary.New([]byte{0x7c})
	if v, _, err := f.ValueType(); err != nil || v != wasm.F64 {
		t.Fatal(v, err)
	}
	r = binary.New([]byte("\x00asm\x01\x00\x00\x00"))
	if m, n, err := f.Module(); err != nil || m == nil || n != 0 {
		t.Fatal(m, n, err)
	}
}

func TestEmbeddedCoreFragmentFailuresAreClassified(t *testing.T) {
	m := new(module.Module)
	for _, row := range []struct {
		input    []byte
		limits   Limits
		features wasm.Features
		want     error
	}{
		{[]byte{0x60, 0, 0}, DefaultLimits(), 1 << 63, wasm.ErrUnsupported},
		{[]byte{0x60, 1, 0x7f, 0}, Limits{}, 0, wasm.ErrLimit},
		{[]byte{0x60, 0x80}, DefaultLimits(), 0, wasm.ErrMalformed},
		{[]byte{0x60, 1, 0xff, 0}, DefaultLimits(), 0, wasm.ErrMalformed},
	} {
		r := binary.New(row.input)
		f := Fragment{Reader: &r, Types: m, Limits: row.limits, Features: row.features}
		if _, err := f.TypeGroup(); !errors.Is(err, row.want) {
			t.Fatal(row, err)
		}
	}
	r := binary.New([]byte{0xff})
	f := Fragment{Reader: &r, Types: m, Limits: DefaultLimits()}
	if _, _, err := f.ExternType(); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(err)
	}
	r = binary.New(nil)
	if out, _, err := f.Module(); out != nil || !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(out, err)
	}
}

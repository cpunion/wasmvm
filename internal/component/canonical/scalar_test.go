package canonical

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type testMemory []byte

func (m testMemory) Read(at uint64, b []byte) error {
	if at > uint64(len(m)) || uint64(len(b)) > uint64(len(m))-at {
		return wasm.ErrTrap
	}
	copy(b, m[at:])
	return nil
}
func (m testMemory) Write(at uint64, b []byte) error {
	if at > uint64(len(m)) || uint64(len(b)) > uint64(len(m))-at {
		return wasm.ErrTrap
	}
	copy(m[at:], b)
	return nil
}

func TestScalarLiftingNormalizesCanonicalValues(t *testing.T) {
	for _, row := range []struct {
		kind    model.Kind
		flat    wasm.ValueType
		in, out uint64
	}{
		{model.Bool, wasm.I32, 0, 0}, {model.Bool, wasm.I32, 256, 1},
		{model.U8, wasm.I32, 0x1234ff, 255}, {model.U16, wasm.I32, 0xffffffff, 65535},
		{model.S8, wasm.I32, 0xff, math.MaxUint64}, {model.S16, wasm.I32, 0xffff8000, uint64(math.MaxUint64 - 32767)},
		{model.S32, wasm.I32, 0xffffffff, math.MaxUint64}, {model.S64, wasm.I64, math.MaxUint64, math.MaxUint64},
		{model.U32, wasm.I32, 0xffffffff, 0xffffffff}, {model.U64, wasm.I64, math.MaxUint64, math.MaxUint64},
		{model.F32, wasm.F32, 0xff800001, 0x7fc00000}, {model.F32, wasm.F32, 0x3f800000, 0x3f800000},
		{model.F64, wasm.F64, 0xfff0000000000001, 0x7ff8000000000000}, {model.F64, wasm.F64, 0x7ff0000000000000, 0x7ff0000000000000},
		{model.Char, wasm.I32, 0x10ffff, 0x10ffff},
	} {
		v, err := LiftScalar(row.kind, Slot{row.flat, row.in})
		if err != nil || v.Bits != row.out {
			t.Fatal(row, v, err)
		}
		flat, err := LowerScalar(v)
		if err != nil {
			t.Fatal(row, err)
		}
		got, err := LiftScalar(row.kind, flat)
		if err != nil || got != v {
			t.Fatal(row, got, v, err)
		}
	}
	for _, kind := range []model.Kind{model.Bool, model.S8, model.U8, model.S16, model.U16, model.S32, model.U32, model.S64, model.U64, model.F32, model.F64, model.Char} {
		v := Scalar{Kind: kind, Bits: 1}
		memory := make(testMemory, 32)
		if err := StoreScalar(memory, 8, v); err != nil {
			t.Fatal(kind, err)
		}
		got, err := LoadScalar(memory, 8, kind)
		if err != nil || got != v {
			t.Fatal(kind, got, err)
		}
	}
	if got, err := LiftScalar(model.Bool, Slot{wasm.I32, 0xffffffff}); err != nil || got.Bits != 1 {
		t.Fatal(got, err)
	}
}

func TestScalarInvalidValuesAndMemoryRanges(t *testing.T) {
	for _, slot := range []Slot{{wasm.I64, 1}, {wasm.I32, 1 << 32}} {
		if _, err := LiftScalar(model.U32, slot); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := LiftScalar(model.List, Slot{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	for _, bits := range []uint64{0xd800, 0xdfff, 0x110000} {
		if _, err := LiftScalar(model.Char, Slot{wasm.I32, bits}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(bits, err)
		}
	}
	for _, v := range []Scalar{{model.U8, 256}, {model.S8, 128}, {model.S16, uint64(math.MaxUint64 - 32768)}, {model.Bool, 2}, {model.List, 0}} {
		if _, err := LowerScalar(v); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(v, err)
		}
	}
	if _, err := LowerScalar(Scalar{model.Char, 0xd800}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	memory := make(testMemory, 8)
	for _, at := range []uint64{1, 8, math.MaxUint64} {
		if _, err := LoadScalar(memory, at, model.U64); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(at, err)
		}
		if err := StoreScalar(memory, at, Scalar{model.U64, 1}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(at, err)
		}
	}
	if _, err := LoadScalar(memory, 0, model.List); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := StoreScalar(memory, 0, Scalar{model.Bool, 3}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

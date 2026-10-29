package canonical

import (
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Scalar stores signed integers as sign-extended two's-complement bits, floats
// as IEEE bits, booleans as 0/1 and characters as Unicode scalar values.
type Scalar struct {
	Kind model.Kind
	Bits uint64
}
type Slot struct {
	Type wasm.ValueType
	Bits uint64
}

// Memory copies bytes across a checked memory capability. Implementations must
// reject out-of-bounds ranges; the codec never retains a guest-memory slice.
type Memory interface {
	Read(uint64, []byte) error
	Write(uint64, []byte) error
}

func scalarInfo(kind model.Kind) (size uint64, flat wasm.ValueType, signed bool, err error) {
	switch kind {
	case model.Bool, model.U8:
		return 1, wasm.I32, false, nil
	case model.S8:
		return 1, wasm.I32, true, nil
	case model.U16:
		return 2, wasm.I32, false, nil
	case model.S16:
		return 2, wasm.I32, true, nil
	case model.U32, model.Char:
		return 4, wasm.I32, false, nil
	case model.S32:
		return 4, wasm.I32, true, nil
	case model.U64:
		return 8, wasm.I64, false, nil
	case model.S64:
		return 8, wasm.I64, true, nil
	case model.F32:
		return 4, wasm.F32, false, nil
	case model.F64:
		return 8, wasm.F64, false, nil
	}
	return 0, 0, false, invalid("non-scalar type")
}

func scalarTrap(detail string) error { return fmt.Errorf("%w: canonical %s", wasm.ErrTrap, detail) }

// LiftScalar follows the pinned ABI's truncation and boolean rules. In
// particular, nonzero bools are true, while out-of-range characters trap.
func LiftScalar(kind model.Kind, slot Slot) (Scalar, error) {
	size, flat, signed, err := scalarInfo(kind)
	if err != nil {
		return Scalar{}, err
	}
	if flat != slot.Type || size < 8 && slot.Bits > math.MaxUint32 {
		return Scalar{}, invalid("core scalar representation")
	}
	bits := slot.Bits
	if size < 8 {
		bits &= (uint64(1) << (size * 8)) - 1
	}
	if signed && size < 8 {
		bits = uint64(int64(bits<<(64-size*8)) >> (64 - size*8))
	}
	switch kind {
	case model.Bool:
		bits = 0
		if slot.Bits != 0 {
			bits = 1
		}
	case model.Char:
		if bits > utf8.MaxRune || bits >= 0xd800 && bits <= 0xdfff {
			return Scalar{}, scalarTrap("invalid character")
		}
	case model.F32:
		if bits&0x7f800000 == 0x7f800000 && bits&0x007fffff != 0 {
			bits = 0x7fc00000
		}
	case model.F64:
		if bits&0x7ff0000000000000 == 0x7ff0000000000000 && bits&0x000fffffffffffff != 0 {
			bits = 0x7ff8000000000000
		}
	}
	return Scalar{Kind: kind, Bits: bits}, nil
}

func LowerScalar(value Scalar) (Slot, error) {
	size, flat, signed, err := scalarInfo(value.Kind)
	if err != nil {
		return Slot{}, err
	}
	bits := value.Bits
	if signed && size < 8 {
		width := size * 8
		if int64(bits) < -(int64(1)<<(width-1)) || int64(bits) > (int64(1)<<(width-1))-1 {
			return Slot{}, invalid("signed scalar range")
		}
	} else if size < 8 && bits>>(size*8) != 0 {
		return Slot{}, invalid("unsigned scalar range")
	}
	if value.Kind == model.Bool && bits > 1 {
		return Slot{}, invalid("host boolean")
	}
	if flat == wasm.I32 {
		bits = uint64(uint32(bits))
	}
	// Deterministic canonical NaNs are one allowed lowering choice.
	normalized, err := LiftScalar(value.Kind, Slot{flat, bits})
	if err != nil {
		return Slot{}, err
	}
	bits = normalized.Bits
	if flat == wasm.I32 {
		bits = uint64(uint32(bits))
	}
	return Slot{flat, bits}, nil
}

func LoadScalar(memory Memory, at uint64, kind model.Kind) (Scalar, error) {
	size, flat, _, err := scalarInfo(kind)
	if err != nil {
		return Scalar{}, err
	}
	if at%size != 0 || at > math.MaxUint64-size {
		return Scalar{}, scalarTrap("unaligned or overflowing scalar")
	}
	var buf [8]byte
	if err := memory.Read(at, buf[:size]); err != nil {
		return Scalar{}, err
	}
	bits := binary.LittleEndian.Uint64(buf[:])
	return LiftScalar(kind, Slot{flat, bits})
}

func StoreScalar(memory Memory, at uint64, value Scalar) error {
	slot, err := LowerScalar(value)
	if err != nil {
		return err
	}
	size, _, _, _ := scalarInfo(value.Kind)
	if at%size != 0 || at > math.MaxUint64-size {
		return scalarTrap("unaligned or overflowing scalar")
	}
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], slot.Bits)
	return memory.Write(at, buf[:size])
}

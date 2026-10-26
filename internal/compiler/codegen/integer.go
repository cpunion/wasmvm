// Package codegen selects and emits owned native instructions in pure Go.
package codegen

import (
	"encoding/binary"
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Integer emits a C-ABI entry for one Core 2.0 integer operation. Operands use
// ResultLo/ResultHi; successful output replaces ResultLo. Trap exits write the
// Wasm trap code to Payload[0]. This instruction-selection entry is independent
// of whole-function lowering, fuel, roots and continuation dispatch; it does
// not establish support for complete native Wasm functions.
func Integer(arch string, op uint32, maxBytes uint64) (nativeabi.Entry, error) {
	if !(op >= 0x45 && op <= 0x5a || op >= 0x67 && op <= 0x8a ||
		op == 0xa7 || op == 0xac || op == 0xad || op >= 0xc0 && op <= 0xc4) {
		return nativeabi.Entry{}, wasm.ErrUnsupported
	}
	wide := op >= 0x50 && op <= 0x5a || op >= 0x79 && op <= 0x8a || op == 0xac || op == 0xad || op >= 0xc2
	a := assembler{arch: arch, labels: map[int]int{}}
	switch arch {
	case "amd64":
		integerAMD64(&a, op, wide)
	case "arm64":
		integerARM64(&a, op, wide)
	default:
		return nativeabi.Entry{}, wasm.ErrUnsupported
	}
	body, err := a.finish()
	if err != nil {
		return nativeabi.Entry{}, err
	}
	return nativeabi.Wrap(arch, body, maxBytes)
}

type relocation struct {
	position, label int
	conditional     bool
}
type assembler struct {
	arch   string
	code   []byte
	labels map[int]int
	fixups []relocation
	next   int
}

func (a *assembler) bytes(b ...byte) { a.code = append(a.code, b...) }
func (a *assembler) word(w uint32)   { a.code = binary.LittleEndian.AppendUint32(a.code, w) }
func (a *assembler) imm64(w uint64)  { a.code = binary.LittleEndian.AppendUint64(a.code, w) }
func (a *assembler) label() int      { a.next++; return a.next }
func (a *assembler) mark(label int)  { a.labels[label] = len(a.code) }
func (a *assembler) branch(label int, conditional bool, cc byte) {
	if a.arch == "amd64" {
		if conditional {
			a.bytes(0x0f, 0x80+cc)
		} else {
			a.bytes(0xe9)
		}
		a.fixups = append(a.fixups, relocation{len(a.code), label, conditional})
		a.word(0)
	} else {
		a.fixups = append(a.fixups, relocation{len(a.code), label, conditional})
		if conditional {
			a.word(0x54000000 | uint32(cc))
		} else {
			a.word(0x14000000)
		}
	}
}
func (a *assembler) finish() ([]byte, error) {
	for _, r := range a.fixups {
		target, ok := a.labels[r.label]
		if !ok {
			return nil, wasm.ErrInvalid
		}
		delta := int64(target) - int64(r.position)
		if a.arch == "amd64" {
			delta -= 4
			if delta < math.MinInt32 || delta > math.MaxInt32 {
				return nil, wasm.ErrLimit
			}
			binary.LittleEndian.PutUint32(a.code[r.position:], uint32(delta))
		} else {
			if delta%4 != 0 {
				return nil, wasm.ErrInvalid
			}
			delta /= 4
			bits := uint(26)
			if r.conditional {
				bits = 19
			}
			if delta < -(1<<(bits-1)) || delta >= 1<<(bits-1) {
				return nil, wasm.ErrLimit
			}
			u := uint32(delta) & ((1 << bits) - 1)
			mask := uint32((1 << bits) - 1)
			if r.conditional {
				u <<= 5
				mask <<= 5
			}
			old := binary.LittleEndian.Uint32(a.code[r.position:])
			binary.LittleEndian.PutUint32(a.code[r.position:], old&^mask|u)
		}
	}
	return a.code, nil
}

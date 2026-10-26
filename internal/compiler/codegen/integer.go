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
	arch         string
	code         []byte
	labels       map[int]int
	fixups       []relocation
	next         int
	bounded      bool
	limit        uint64
	longBranches bool
	tables       []tableRelocation
	addresses    []addressRelocation
}

func (a *assembler) reserve(n uint64) {
	if a.bounded && (uint64(len(a.code)) > a.limit || n > a.limit-uint64(len(a.code))) {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Detail: "native code bytes"})
	}
}
func (a *assembler) bytes(b ...byte) { a.reserve(uint64(len(b))); a.code = append(a.code, b...) }
func (a *assembler) word(w uint32) {
	a.reserve(4)
	a.code = binary.LittleEndian.AppendUint32(a.code, w)
}
func (a *assembler) imm64(w uint64) {
	a.reserve(8)
	a.code = binary.LittleEndian.AppendUint64(a.code, w)
}
func (a *assembler) label() int     { a.next++; return a.next }
func (a *assembler) mark(label int) { a.labels[label] = len(a.code) }
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
		if conditional && a.longBranches {
			a.word(0x54000040 | uint32(cc^1)) // Inverted condition skips the long branch.
			conditional = false
		}
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
	for _, r := range a.tables {
		target, ok := a.labels[r.label]
		base, bok := a.labels[r.base]
		if !ok || !bok {
			return nil, wasm.ErrInvalid
		}
		delta := int64(target) - int64(base)
		if delta < math.MinInt32 || delta > math.MaxInt32 {
			return nil, wasm.ErrLimit
		}
		binary.LittleEndian.PutUint32(a.code[r.position:], uint32(delta))
	}
	for _, r := range a.addresses {
		target, ok := a.labels[r.label]
		if !ok {
			return nil, wasm.ErrInvalid
		}
		delta := int64(target) - int64(r.position)
		if delta < -(1<<20) || delta >= 1<<20 {
			return nil, wasm.ErrLimit
		}
		u := uint32(delta) & ((1 << 21) - 1)
		binary.LittleEndian.PutUint32(a.code[r.position:], 0x1000000d|(u&3)<<29|(u>>2)<<5) // ADR x13
	}
	return a.code, nil
}

type tableRelocation struct{ position, label, base int }
type addressRelocation struct{ position, label int }

// dispatch emits a relative jump table. A checked index is in rax/x9; no table
// entry contains an absolute address, and all targets are compiler-owned labels.
func (a *assembler) dispatch(labels []int) {
	base := a.label()
	if a.arch == "amd64" {
		a.bytes(0x48, 0x8d, 0x15) // lea rdx,[rip+table]
		a.fixups = append(a.fixups, relocation{len(a.code), base, false})
		a.word(0)
		a.bytes(0x48, 0x63, 0x0c, 0x82, 0x48, 0x01, 0xd1, 0xff, 0xe1)
	} else {
		a.addresses = append(a.addresses, addressRelocation{len(a.code), base})
		a.word(0)
		a.word(0xb8a979ae)
		a.word(0x8b0e01ad)
		a.word(0xd61f01a0)
	}
	a.mark(base)
	for _, label := range labels {
		a.tables = append(a.tables, tableRelocation{len(a.code), label, base})
		a.word(0)
	}
}

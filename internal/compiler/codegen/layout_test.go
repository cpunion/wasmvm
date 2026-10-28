package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompactedRelocationsAndExactLimit(t *testing.T) {
	for _, limit := range []uint64{19, 20, 24} {
		a := assembler{arch: "arm64", labels: map[int]int{}, longBranches: true, compactBranches: true, bounded: true, limit: limit}
		start, target, base := a.label(), a.label(), a.label()
		a.mark(start)
		a.branch(target, true, 10)
		a.addresses = append(a.addresses, addressRelocation{len(a.code), target})
		a.word(0)
		a.mark(base)
		a.tables = append(a.tables, tableRelocation{len(a.code), target, base})
		a.word(0)
		a.branch(start, false, 0)
		a.mark(target)
		a.word(0xd65f03c0)
		code, err := a.finish()
		if limit == 19 {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
			continue
		}
		if err != nil || len(code) != 20 || a.labels[target] != 16 || a.labels[base] != 8 {
			t.Fatal(len(code), a.labels, err)
		}
		branch := binary.LittleEndian.Uint32(code)
		if branch&15 != 10 || int(int32(branch<<8)>>13)*4 != 16 {
			t.Fatal("conditional relocation", branch)
		}
		adr := binary.LittleEndian.Uint32(code[4:])
		imm := (adr >> 29 & 3) | (adr>>5&0x7ffff)<<2
		if 4+int(int32(imm<<11)>>11) != 16 {
			t.Fatal("address relocation", adr)
		}
		if binary.LittleEndian.Uint32(code[8:]) != 8 {
			t.Fatal("table relocation")
		}
		jump := binary.LittleEndian.Uint32(code[12:])
		if 12+int(int32(jump<<6)>>6)*4 != 0 {
			t.Fatal("backward relocation", jump)
		}
		before := append([]byte(nil), code...)
		if _, err = a.finish(); err != nil || !bytes.Equal(before, a.code) {
			t.Fatal("non-idempotent compaction", err)
		}
	}
}

func TestCompactionNearAndLongBranches(t *testing.T) {
	for _, distance := range []int{(1 << 20) - 4, 1 << 20, (1 << 20) + 16} {
		a := assembler{arch: "arm64", labels: map[int]int{}, longBranches: true, compactBranches: true}
		far, near := a.label(), a.label()
		a.compareBranch(far, 28, true, true)
		a.compareBranch(near, 27, false, false)
		a.mark(near)
		a.bytes(make([]byte, distance-len(a.code))...)
		a.mark(far)
		a.word(0xd65f03c0)
		code, err := a.finish()
		if err != nil {
			t.Fatal(err)
		}
		first := binary.LittleEndian.Uint32(code)
		at := int(int32(first<<8)>>13) * 4
		if first&(1<<24) != 0 {
			if at != a.labels[far] {
				t.Fatal("short destination", distance, at, a.labels)
			}
		} else {
			jump := binary.LittleEndian.Uint32(code[4:])
			if at != 8 || 4+int(int32(jump<<6)>>6)*4 != a.labels[far] {
				t.Fatal("long destination", distance, at, a.labels)
			}
		}
		before := append([]byte(nil), code...)
		if _, err = a.finish(); err != nil || !bytes.Equal(before, a.code) {
			t.Fatal(err)
		}
	}
}

func TestCompactionRejectsInvalidLabels(t *testing.T) {
	for _, kind := range []string{"missing", "interior", "target"} {
		a := assembler{arch: "arm64", labels: map[int]int{}, longBranches: true, compactBranches: true}
		a.branch(1, true, 0)
		if kind == "interior" {
			a.labels[1] = 4
		}
		if kind == "target" {
			a.arch = "amd64"
		}
		if _, err := a.finish(); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(kind, err)
		}
	}
}

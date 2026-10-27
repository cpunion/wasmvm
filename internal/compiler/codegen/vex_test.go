package codegen

import (
	"bytes"
	"testing"
)

func TestVEX128Encoding(t *testing.T) {
	// Fixed bytes cross-checked against LLVM's independent x86 assembler.
	for _, test := range []struct {
		mapByte, prefix, opcode, d, l, r byte
		wide                             bool
		want                             []byte
	}{
		{1, 0, 0x58, 0, 1, 2, false, []byte{0xc4, 0xe1, 0x70, 0x58, 0xc2}},
		{1, 1, 0x58, 8, 9, 10, false, []byte{0xc4, 0x41, 0x31, 0x58, 0xc2}},
		{2, 1, 0x00, 0, 1, 2, false, []byte{0xc4, 0xe2, 0x71, 0x00, 0xc2}},
		{2, 1, 0x47, 0, 1, 2, true, []byte{0xc4, 0xe2, 0xf1, 0x47, 0xc2}},
		{2, 1, 0x23, 0, 0, 1, false, []byte{0xc4, 0xe2, 0x79, 0x23, 0xc1}},
		{3, 1, 0x4b, 8, 9, 10, false, []byte{0xc4, 0x43, 0x31, 0x4b, 0xc2}},
	} {
		a := assembler{arch: "amd64"}
		a.xvex(test.mapByte, test.prefix, test.opcode, test.d, test.l, test.r, test.wide)
		if !bytes.Equal(a.code, test.want) {
			t.Fatalf("%+v: %x", test, a.code)
		}
	}
	a := assembler{arch: "amd64"}
	a.xvexShift(0x73, 10, 9, 3, 8)
	if !bytes.Equal(a.code, []byte{0xc4, 0xc1, 0x29, 0x73, 0xd9, 8}) {
		t.Fatalf("immediate shift: %x", a.code)
	}
}

func TestEVEX128Encoding(t *testing.T) {
	// Independent LLVM 22 assembler output; all forms use k0 and 128 bits.
	for _, test := range []struct {
		opcode, d, l, r byte
		wide            bool
		want            []byte
	}{
		{0x54, 5, 0, 3, false, []byte{0x62, 0xf2, 0x7d, 8, 0x54, 0xeb}},
		{0x54, 13, 0, 10, false, []byte{0x62, 0x52, 0x7d, 8, 0x54, 0xea}},
		{0x46, 9, 3, 1, true, []byte{0x62, 0x72, 0xe5, 8, 0x46, 0xc9}},
		{0x46, 8, 9, 10, true, []byte{0x62, 0x52, 0xb5, 8, 0x46, 0xc2}},
	} {
		a := assembler{arch: "amd64"}
		a.xevex(2, 1, test.opcode, test.d, test.l, test.r, test.wide)
		if !bytes.Equal(a.code, test.want) {
			t.Fatalf("%+v: %x", test, a.code)
		}
	}
}

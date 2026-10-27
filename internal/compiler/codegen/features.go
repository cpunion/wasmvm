package codegen

// X86Features is an explicit, validated instruction-selection input. Zero
// selects baseline SSE2 even when the compiler runs on a more capable host.
type X86Features uint32

const (
	X86SSSE3 X86Features = 1 << iota
	X86SSE41
	X86AVX2
	X86AVX512VL     // AVX512F+VL and complete OS register-state support.
	X86AVX512BITALG // AVX512F+VL+BITALG and complete OS register-state support.
)

const x86KnownFeatures = X86SSSE3 | X86SSE41 | X86AVX2 | X86AVX512VL | X86AVX512BITALG

// EVEX.128 retains the same XMM0-XMM15 slot contract. No mask or upper vector
// register is used; the reserved vvvv field of unary operations uses left=0.
func (a *assembler) xevex(mapByte, prefix, opcode, dst, left, right byte, wide bool) {
	second := byte(0xf0) | mapByte
	if dst >= 8 {
		second &^= 0x80
	}
	if right >= 8 {
		second &^= 0x20
	}
	third := (^left&15)<<3 | 4 | prefix
	if wide {
		third |= 0x80
	}
	a.bytes(0x62, second, third, 8, opcode, 0xc0|(dst&7)<<3|(right&7))
}

// VEX.128 uses the same 128-bit register/slot contract as the baseline. The
// native platform checks AVX, OSXSAVE and XCR0 before advertising this profile.
func (a *assembler) xvex(mapByte, prefix, opcode, dst, left, right byte, wide bool) {
	second := byte(0xe0) | mapByte
	if dst >= 8 {
		second &^= 0x80
	}
	if right >= 8 {
		second &^= 0x20
	}
	third := (^left&15)<<3 | prefix
	if wide {
		third |= 0x80
	}
	a.bytes(0xc4, second, third, opcode, 0xc0|(dst&7)<<3|(right&7))
}

// Packed immediate shifts encode the result in VEX.vvvv and the operation in
// ModRM.reg. This differs from the ordinary three-register encoding above.
func (a *assembler) xvexShift(opcode, dst, src, kind, count byte) {
	a.xvex(1, 1, opcode, kind, dst, src, false)
	a.bytes(count)
}

func (a *assembler) xvectorExtended(mapByte, opcode, dst, src byte) {
	a.bytes(0x66)
	if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
		a.bytes(rex)
	}
	a.bytes(0x0f, mapByte, opcode, 0xc0|(dst&7)<<3|(src&7))
}

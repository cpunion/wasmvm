package codegen

// X86Features is an explicit, validated instruction-selection input. Zero
// selects baseline SSE2 even when the compiler runs on a more capable host.
type X86Features uint32

const (
	X86SSSE3 X86Features = 1 << iota
	X86SSE41
)

func (a *assembler) xvectorExtended(mapByte, opcode, dst, src byte) {
	a.bytes(0x66)
	if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
		a.bytes(rex)
	}
	a.bytes(0x0f, mapByte, opcode, 0xc0|(dst&7)<<3|(src&7))
}

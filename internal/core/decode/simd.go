package decode

// Core 2.0 SIMD opcode inventory, from the binary grammar at
// WebAssembly/spec fffc6e12fa454e475455a7b58d3b5dc343980c10,
// document/core/binary/instructions.rst. Reserved holes must remain invalid.
var simdOpcodes = [4]uint64{
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xf7e27f9bfbffffff,
	0xffffbffbffe27f9b,
}

package numeric

import "math/bits"

// Integer primitives have one canonical implementation shared by Eval and the
// instruction dispatcher. Small bodies permit inlining without a second opcode
// switch; all operands and results remain exact Wasm bits.
func I32Eqz(a uint64) uint64        { return boolean(uint32(a) == 0) }
func I32Eq(a, b uint64) uint64      { return boolean(uint32(a) == uint32(b)) }
func I32Ne(a, b uint64) uint64      { return boolean(uint32(a) != uint32(b)) }
func I32LtS(a, b uint64) uint64     { return boolean(int32(a) < int32(b)) }
func I32LtU(a, b uint64) uint64     { return boolean(uint32(a) < uint32(b)) }
func I32GtS(a, b uint64) uint64     { return boolean(int32(a) > int32(b)) }
func I32GtU(a, b uint64) uint64     { return boolean(uint32(a) > uint32(b)) }
func I32LeS(a, b uint64) uint64     { return boolean(int32(a) <= int32(b)) }
func I32LeU(a, b uint64) uint64     { return boolean(uint32(a) <= uint32(b)) }
func I32GeS(a, b uint64) uint64     { return boolean(int32(a) >= int32(b)) }
func I32GeU(a, b uint64) uint64     { return boolean(uint32(a) >= uint32(b)) }
func I64Eqz(a uint64) uint64        { return boolean(uint64(a) == 0) }
func I64Eq(a, b uint64) uint64      { return boolean(uint64(a) == uint64(b)) }
func I64Ne(a, b uint64) uint64      { return boolean(uint64(a) != uint64(b)) }
func I64LtS(a, b uint64) uint64     { return boolean(int64(a) < int64(b)) }
func I64LtU(a, b uint64) uint64     { return boolean(uint64(a) < uint64(b)) }
func I64GtS(a, b uint64) uint64     { return boolean(int64(a) > int64(b)) }
func I64GtU(a, b uint64) uint64     { return boolean(uint64(a) > uint64(b)) }
func I64LeS(a, b uint64) uint64     { return boolean(int64(a) <= int64(b)) }
func I64LeU(a, b uint64) uint64     { return boolean(uint64(a) <= uint64(b)) }
func I64GeS(a, b uint64) uint64     { return boolean(int64(a) >= int64(b)) }
func I64GeU(a, b uint64) uint64     { return boolean(uint64(a) >= uint64(b)) }
func I32Clz(a uint64) uint64        { return uint64(bits.LeadingZeros32(uint32(a))) }
func I32Ctz(a uint64) uint64        { return uint64(bits.TrailingZeros32(uint32(a))) }
func I32Popcnt(a uint64) uint64     { return uint64(bits.OnesCount32(uint32(a))) }
func I32Add(a, b uint64) uint64     { return uint64(uint32(a) + uint32(b)) }
func I32Sub(a, b uint64) uint64     { return uint64(uint32(a) - uint32(b)) }
func I32Mul(a, b uint64) uint64     { return uint64(uint32(a) * uint32(b)) }
func I32And(a, b uint64) uint64     { return uint64(uint32(a) & uint32(b)) }
func I32Or(a, b uint64) uint64      { return uint64(uint32(a) | uint32(b)) }
func I32Xor(a, b uint64) uint64     { return uint64(uint32(a) ^ uint32(b)) }
func I32Shl(a, b uint64) uint64     { return uint64(uint32(a) << (b & 31)) }
func I32ShrS(a, b uint64) uint64    { return uint64(uint32(int32(a) >> (b & 31))) }
func I32ShrU(a, b uint64) uint64    { return uint64(uint32(a) >> (b & 31)) }
func I32Rotl(a, b uint64) uint64    { return uint64(bits.RotateLeft32(uint32(a), int(b&31))) }
func I32Rotr(a, b uint64) uint64    { return uint64(bits.RotateLeft32(uint32(a), -int(b&31))) }
func I64Clz(a uint64) uint64        { return uint64(bits.LeadingZeros64(uint64(a))) }
func I64Ctz(a uint64) uint64        { return uint64(bits.TrailingZeros64(uint64(a))) }
func I64Popcnt(a uint64) uint64     { return uint64(bits.OnesCount64(uint64(a))) }
func I64Add(a, b uint64) uint64     { return uint64(uint64(a) + uint64(b)) }
func I64Sub(a, b uint64) uint64     { return uint64(uint64(a) - uint64(b)) }
func I64Mul(a, b uint64) uint64     { return uint64(uint64(a) * uint64(b)) }
func I64And(a, b uint64) uint64     { return uint64(uint64(a) & uint64(b)) }
func I64Or(a, b uint64) uint64      { return uint64(uint64(a) | uint64(b)) }
func I64Xor(a, b uint64) uint64     { return uint64(uint64(a) ^ uint64(b)) }
func I64Shl(a, b uint64) uint64     { return uint64(uint64(a) << (b & 63)) }
func I64ShrS(a, b uint64) uint64    { return uint64(uint64(int64(a) >> (b & 63))) }
func I64ShrU(a, b uint64) uint64    { return uint64(uint64(a) >> (b & 63)) }
func I64Rotl(a, b uint64) uint64    { return uint64(bits.RotateLeft64(uint64(a), int(b&63))) }
func I64Rotr(a, b uint64) uint64    { return uint64(bits.RotateLeft64(uint64(a), -int(b&63))) }
func I32WrapI64(a uint64) uint64    { return uint64(uint32(a)) }
func I64ExtendI32S(a uint64) uint64 { return uint64(int64(int32(a))) }
func I64ExtendI32U(a uint64) uint64 { return uint64(uint32(a)) }
func I32Extend8S(a uint64) uint64   { return uint64(uint32(int8(a))) }
func I32Extend16S(a uint64) uint64  { return uint64(uint32(int16(a))) }
func I64Extend8S(a uint64) uint64   { return uint64(uint64(int8(a))) }
func I64Extend16S(a uint64) uint64  { return uint64(uint64(int16(a))) }
func I64Extend32S(a uint64) uint64  { return uint64(uint64(int32(a))) }

package numeric

import "math/bits"

// Integer primitives have one canonical implementation shared by Eval and the
// instruction dispatcher. Small bodies permit inlining without a second opcode
// switch; all operands and results remain exact Wasm bits. Consumer-side
// instantiation exposes the same bodies to LLGo without requiring LTO.
func I32Eqz[T ~uint64](a T) uint64        { return boolean(uint32(a) == 0) }
func I32Eq[T ~uint64](a, b T) uint64      { return boolean(uint32(a) == uint32(b)) }
func I32Ne[T ~uint64](a, b T) uint64      { return boolean(uint32(a) != uint32(b)) }
func I32LtS[T ~uint64](a, b T) uint64     { return boolean(int32(a) < int32(b)) }
func I32LtU[T ~uint64](a, b T) uint64     { return boolean(uint32(a) < uint32(b)) }
func I32GtS[T ~uint64](a, b T) uint64     { return boolean(int32(a) > int32(b)) }
func I32GtU[T ~uint64](a, b T) uint64     { return boolean(uint32(a) > uint32(b)) }
func I32LeS[T ~uint64](a, b T) uint64     { return boolean(int32(a) <= int32(b)) }
func I32LeU[T ~uint64](a, b T) uint64     { return boolean(uint32(a) <= uint32(b)) }
func I32GeS[T ~uint64](a, b T) uint64     { return boolean(int32(a) >= int32(b)) }
func I32GeU[T ~uint64](a, b T) uint64     { return boolean(uint32(a) >= uint32(b)) }
func I64Eqz[T ~uint64](a T) uint64        { return boolean(uint64(a) == 0) }
func I64Eq[T ~uint64](a, b T) uint64      { return boolean(uint64(a) == uint64(b)) }
func I64Ne[T ~uint64](a, b T) uint64      { return boolean(uint64(a) != uint64(b)) }
func I64LtS[T ~uint64](a, b T) uint64     { return boolean(int64(a) < int64(b)) }
func I64LtU[T ~uint64](a, b T) uint64     { return boolean(uint64(a) < uint64(b)) }
func I64GtS[T ~uint64](a, b T) uint64     { return boolean(int64(a) > int64(b)) }
func I64GtU[T ~uint64](a, b T) uint64     { return boolean(uint64(a) > uint64(b)) }
func I64LeS[T ~uint64](a, b T) uint64     { return boolean(int64(a) <= int64(b)) }
func I64LeU[T ~uint64](a, b T) uint64     { return boolean(uint64(a) <= uint64(b)) }
func I64GeS[T ~uint64](a, b T) uint64     { return boolean(int64(a) >= int64(b)) }
func I64GeU[T ~uint64](a, b T) uint64     { return boolean(uint64(a) >= uint64(b)) }
func I32Clz[T ~uint64](a T) uint64        { return uint64(bits.LeadingZeros32(uint32(a))) }
func I32Ctz[T ~uint64](a T) uint64        { return uint64(bits.TrailingZeros32(uint32(a))) }
func I32Popcnt[T ~uint64](a T) uint64     { return uint64(bits.OnesCount32(uint32(a))) }
func I32Add[T ~uint64](a, b T) uint64     { return uint64(uint32(a) + uint32(b)) }
func I32Sub[T ~uint64](a, b T) uint64     { return uint64(uint32(a) - uint32(b)) }
func I32Mul[T ~uint64](a, b T) uint64     { return uint64(uint32(a) * uint32(b)) }
func I32And[T ~uint64](a, b T) uint64     { return uint64(uint32(a) & uint32(b)) }
func I32Or[T ~uint64](a, b T) uint64      { return uint64(uint32(a) | uint32(b)) }
func I32Xor[T ~uint64](a, b T) uint64     { return uint64(uint32(a) ^ uint32(b)) }
func I32Shl[T ~uint64](a, b T) uint64     { return uint64(uint32(a) << (b & 31)) }
func I32ShrS[T ~uint64](a, b T) uint64    { return uint64(uint32(int32(a) >> (b & 31))) }
func I32ShrU[T ~uint64](a, b T) uint64    { return uint64(uint32(a) >> (b & 31)) }
func I32Rotl[T ~uint64](a, b T) uint64    { return uint64(bits.RotateLeft32(uint32(a), int(b&31))) }
func I32Rotr[T ~uint64](a, b T) uint64    { return uint64(bits.RotateLeft32(uint32(a), -int(b&31))) }
func I64Clz[T ~uint64](a T) uint64        { return uint64(bits.LeadingZeros64(uint64(a))) }
func I64Ctz[T ~uint64](a T) uint64        { return uint64(bits.TrailingZeros64(uint64(a))) }
func I64Popcnt[T ~uint64](a T) uint64     { return uint64(bits.OnesCount64(uint64(a))) }
func I64Add[T ~uint64](a, b T) uint64     { return uint64(uint64(a) + uint64(b)) }
func I64Sub[T ~uint64](a, b T) uint64     { return uint64(uint64(a) - uint64(b)) }
func I64Mul[T ~uint64](a, b T) uint64     { return uint64(uint64(a) * uint64(b)) }
func I64And[T ~uint64](a, b T) uint64     { return uint64(uint64(a) & uint64(b)) }
func I64Or[T ~uint64](a, b T) uint64      { return uint64(uint64(a) | uint64(b)) }
func I64Xor[T ~uint64](a, b T) uint64     { return uint64(uint64(a) ^ uint64(b)) }
func I64Shl[T ~uint64](a, b T) uint64     { return uint64(uint64(a) << (b & 63)) }
func I64ShrS[T ~uint64](a, b T) uint64    { return uint64(uint64(int64(a) >> (b & 63))) }
func I64ShrU[T ~uint64](a, b T) uint64    { return uint64(uint64(a) >> (b & 63)) }
func I64Rotl[T ~uint64](a, b T) uint64    { return uint64(bits.RotateLeft64(uint64(a), int(b&63))) }
func I64Rotr[T ~uint64](a, b T) uint64    { return uint64(bits.RotateLeft64(uint64(a), -int(b&63))) }
func I32WrapI64[T ~uint64](a T) uint64    { return uint64(uint32(a)) }
func I64ExtendI32S[T ~uint64](a T) uint64 { return uint64(int64(int32(a))) }
func I64ExtendI32U[T ~uint64](a T) uint64 { return uint64(uint32(a)) }
func I32Extend8S[T ~uint64](a T) uint64   { return uint64(uint32(int8(a))) }
func I32Extend16S[T ~uint64](a T) uint64  { return uint64(uint32(int16(a))) }
func I64Extend8S[T ~uint64](a T) uint64   { return uint64(uint64(int8(a))) }
func I64Extend16S[T ~uint64](a T) uint64  { return uint64(uint64(int16(a))) }
func I64Extend32S[T ~uint64](a T) uint64  { return uint64(uint64(int32(a))) }

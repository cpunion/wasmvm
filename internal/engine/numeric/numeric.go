// Package numeric implements the complete Core 2.0 scalar numeric semantics.
// Operands/results are exact Wasm bits. Guest traps are explicit return codes;
// the implementation never relies on host divide/conversion faults.
package numeric

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
	"math/bits"
)

// Eval evaluates a validated numeric opcode. The caller supplies two operands
// in Wasm stack order; b is ignored for unary instructions. Unknown opcodes are
// compiler bugs and panic rather than becoming a guest trap.
func Eval(op uint32, a, b uint64) (uint64, wasm.TrapCode) {
	switch op {
	case 0x45:
		return boolean(uint32(a) == 0), wasm.NoTrap
	case 0x46:
		return boolean(uint32(a) == uint32(b)), wasm.NoTrap
	case 0x47:
		return boolean(uint32(a) != uint32(b)), wasm.NoTrap
	case 0x48:
		return boolean(int32(a) < int32(b)), wasm.NoTrap
	case 0x49:
		return boolean(uint32(a) < uint32(b)), wasm.NoTrap
	case 0x4a:
		return boolean(int32(a) > int32(b)), wasm.NoTrap
	case 0x4b:
		return boolean(uint32(a) > uint32(b)), wasm.NoTrap
	case 0x4c:
		return boolean(int32(a) <= int32(b)), wasm.NoTrap
	case 0x4d:
		return boolean(uint32(a) <= uint32(b)), wasm.NoTrap
	case 0x4e:
		return boolean(int32(a) >= int32(b)), wasm.NoTrap
	case 0x4f:
		return boolean(uint32(a) >= uint32(b)), wasm.NoTrap
	case 0x50:
		return boolean(uint64(a) == 0), wasm.NoTrap
	case 0x51:
		return boolean(uint64(a) == uint64(b)), wasm.NoTrap
	case 0x52:
		return boolean(uint64(a) != uint64(b)), wasm.NoTrap
	case 0x53:
		return boolean(int64(a) < int64(b)), wasm.NoTrap
	case 0x54:
		return boolean(uint64(a) < uint64(b)), wasm.NoTrap
	case 0x55:
		return boolean(int64(a) > int64(b)), wasm.NoTrap
	case 0x56:
		return boolean(uint64(a) > uint64(b)), wasm.NoTrap
	case 0x57:
		return boolean(int64(a) <= int64(b)), wasm.NoTrap
	case 0x58:
		return boolean(uint64(a) <= uint64(b)), wasm.NoTrap
	case 0x59:
		return boolean(int64(a) >= int64(b)), wasm.NoTrap
	case 0x5a:
		return boolean(uint64(a) >= uint64(b)), wasm.NoTrap
	case 0x5b:
		return boolean(math.Float32frombits(uint32(a)) == math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x5c:
		return boolean(math.Float32frombits(uint32(a)) != math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x5d:
		return boolean(math.Float32frombits(uint32(a)) < math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x5e:
		return boolean(math.Float32frombits(uint32(a)) > math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x5f:
		return boolean(math.Float32frombits(uint32(a)) <= math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x60:
		return boolean(math.Float32frombits(uint32(a)) >= math.Float32frombits(uint32(b))), wasm.NoTrap
	case 0x61:
		return boolean(math.Float64frombits(a) == math.Float64frombits(b)), wasm.NoTrap
	case 0x62:
		return boolean(math.Float64frombits(a) != math.Float64frombits(b)), wasm.NoTrap
	case 0x63:
		return boolean(math.Float64frombits(a) < math.Float64frombits(b)), wasm.NoTrap
	case 0x64:
		return boolean(math.Float64frombits(a) > math.Float64frombits(b)), wasm.NoTrap
	case 0x65:
		return boolean(math.Float64frombits(a) <= math.Float64frombits(b)), wasm.NoTrap
	case 0x66:
		return boolean(math.Float64frombits(a) >= math.Float64frombits(b)), wasm.NoTrap
	case 0x67:
		return uint64(bits.LeadingZeros32(uint32(a))), wasm.NoTrap
	case 0x68:
		return uint64(bits.TrailingZeros32(uint32(a))), wasm.NoTrap
	case 0x69:
		return uint64(bits.OnesCount32(uint32(a))), wasm.NoTrap
	case 0x6a:
		return uint64(uint32(a) + uint32(b)), wasm.NoTrap
	case 0x6b:
		return uint64(uint32(a) - uint32(b)), wasm.NoTrap
	case 0x6c:
		return uint64(uint32(a) * uint32(b)), wasm.NoTrap
	case 0x6d:
		if uint32(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		if int32(a) == math.MinInt32 && int32(b) == -1 {
			return 0, wasm.TrapIntegerOverflow
		}
		return uint64(uint32(int32(a) / int32(b))), wasm.NoTrap
	case 0x6e:
		if uint32(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		return uint64(uint32(uint32(a) / uint32(b))), wasm.NoTrap
	case 0x6f:
		if uint32(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		if int32(a) == math.MinInt32 && int32(b) == -1 {
			return 0, wasm.NoTrap
		}
		return uint64(uint32(int32(a) % int32(b))), wasm.NoTrap
	case 0x70:
		if uint32(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		return uint64(uint32(uint32(a) % uint32(b))), wasm.NoTrap
	case 0x71:
		return uint64(uint32(a) & uint32(b)), wasm.NoTrap
	case 0x72:
		return uint64(uint32(a) | uint32(b)), wasm.NoTrap
	case 0x73:
		return uint64(uint32(a) ^ uint32(b)), wasm.NoTrap
	case 0x74:
		return uint64(uint32(a) << (b & 31)), wasm.NoTrap
	case 0x75:
		return uint64(uint32(int32(a) >> (b & 31))), wasm.NoTrap
	case 0x76:
		return uint64(uint32(a) >> (b & 31)), wasm.NoTrap
	case 0x77:
		return uint64(bits.RotateLeft32(uint32(a), int(b&31))), wasm.NoTrap
	case 0x78:
		return uint64(bits.RotateLeft32(uint32(a), -int(b&31))), wasm.NoTrap
	case 0x79:
		return uint64(bits.LeadingZeros64(uint64(a))), wasm.NoTrap
	case 0x7a:
		return uint64(bits.TrailingZeros64(uint64(a))), wasm.NoTrap
	case 0x7b:
		return uint64(bits.OnesCount64(uint64(a))), wasm.NoTrap
	case 0x7c:
		return uint64(uint64(a) + uint64(b)), wasm.NoTrap
	case 0x7d:
		return uint64(uint64(a) - uint64(b)), wasm.NoTrap
	case 0x7e:
		return uint64(uint64(a) * uint64(b)), wasm.NoTrap
	case 0x7f:
		if uint64(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		if int64(a) == math.MinInt64 && int64(b) == -1 {
			return 0, wasm.TrapIntegerOverflow
		}
		return uint64(uint64(int64(a) / int64(b))), wasm.NoTrap
	case 0x80:
		if uint64(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		return uint64(uint64(uint64(a) / uint64(b))), wasm.NoTrap
	case 0x81:
		if uint64(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		if int64(a) == math.MinInt64 && int64(b) == -1 {
			return 0, wasm.NoTrap
		}
		return uint64(uint64(int64(a) % int64(b))), wasm.NoTrap
	case 0x82:
		if uint64(b) == 0 {
			return 0, wasm.TrapIntegerDivideByZero
		}
		return uint64(uint64(uint64(a) % uint64(b))), wasm.NoTrap
	case 0x83:
		return uint64(uint64(a) & uint64(b)), wasm.NoTrap
	case 0x84:
		return uint64(uint64(a) | uint64(b)), wasm.NoTrap
	case 0x85:
		return uint64(uint64(a) ^ uint64(b)), wasm.NoTrap
	case 0x86:
		return uint64(uint64(a) << (b & 63)), wasm.NoTrap
	case 0x87:
		return uint64(uint64(int64(a) >> (b & 63))), wasm.NoTrap
	case 0x88:
		return uint64(uint64(a) >> (b & 63)), wasm.NoTrap
	case 0x89:
		return uint64(bits.RotateLeft64(uint64(a), int(b&63))), wasm.NoTrap
	case 0x8a:
		return uint64(bits.RotateLeft64(uint64(a), -int(b&63))), wasm.NoTrap
	case 0x8b:
		return a & 0x7fffffff, wasm.NoTrap
	case 0x8c:
		return (a ^ 0x80000000) & 0xffffffff, wasm.NoTrap
	case 0x8d:
		return uint64(result32(float32(math.Ceil(float64(math.Float32frombits(uint32(a))))))), wasm.NoTrap
	case 0x8e:
		return uint64(result32(float32(math.Floor(float64(math.Float32frombits(uint32(a))))))), wasm.NoTrap
	case 0x8f:
		return uint64(result32(float32(math.Trunc(float64(math.Float32frombits(uint32(a))))))), wasm.NoTrap
	case 0x90:
		return uint64(result32(float32(math.RoundToEven(float64(math.Float32frombits(uint32(a))))))), wasm.NoTrap
	case 0x91:
		return uint64(result32(float32(math.Sqrt(float64(math.Float32frombits(uint32(a))))))), wasm.NoTrap
	case 0x92:
		return uint64(result32(math.Float32frombits(uint32(a)) + math.Float32frombits(uint32(b)))), wasm.NoTrap
	case 0x93:
		return uint64(result32(math.Float32frombits(uint32(a)) - math.Float32frombits(uint32(b)))), wasm.NoTrap
	case 0x94:
		return uint64(result32(math.Float32frombits(uint32(a)) * math.Float32frombits(uint32(b)))), wasm.NoTrap
	case 0x95:
		return uint64(result32(math.Float32frombits(uint32(a)) / math.Float32frombits(uint32(b)))), wasm.NoTrap
	case 0x96:
		return uint64(result32(float32(minimum(float64(math.Float32frombits(uint32(a))), float64(math.Float32frombits(uint32(b))))))), wasm.NoTrap
	case 0x97:
		return uint64(result32(float32(maximum(float64(math.Float32frombits(uint32(a))), float64(math.Float32frombits(uint32(b))))))), wasm.NoTrap
	case 0x98:
		return (a & 0x7fffffff) | (b & 0x80000000), wasm.NoTrap
	case 0x99:
		return a & 0x7fffffffffffffff, wasm.NoTrap
	case 0x9a:
		return (a ^ 0x8000000000000000) & 0xffffffffffffffff, wasm.NoTrap
	case 0x9b:
		return uint64(result64(math.Ceil(float64(math.Float64frombits(a))))), wasm.NoTrap
	case 0x9c:
		return uint64(result64(math.Floor(float64(math.Float64frombits(a))))), wasm.NoTrap
	case 0x9d:
		return uint64(result64(math.Trunc(float64(math.Float64frombits(a))))), wasm.NoTrap
	case 0x9e:
		return uint64(result64(math.RoundToEven(float64(math.Float64frombits(a))))), wasm.NoTrap
	case 0x9f:
		return uint64(result64(math.Sqrt(float64(math.Float64frombits(a))))), wasm.NoTrap
	case 0xa0:
		return uint64(result64(math.Float64frombits(a) + math.Float64frombits(b))), wasm.NoTrap
	case 0xa1:
		return uint64(result64(math.Float64frombits(a) - math.Float64frombits(b))), wasm.NoTrap
	case 0xa2:
		return uint64(result64(math.Float64frombits(a) * math.Float64frombits(b))), wasm.NoTrap
	case 0xa3:
		return uint64(result64(math.Float64frombits(a) / math.Float64frombits(b))), wasm.NoTrap
	case 0xa4:
		return uint64(result64(minimum(float64(math.Float64frombits(a)), float64(math.Float64frombits(b))))), wasm.NoTrap
	case 0xa5:
		return uint64(result64(maximum(float64(math.Float64frombits(a)), float64(math.Float64frombits(b))))), wasm.NoTrap
	case 0xa6:
		return (a & 0x7fffffffffffffff) | (b & 0x8000000000000000), wasm.NoTrap
	case 0xa7:
		return uint64(uint32(a)), wasm.NoTrap
	case 0xa8:
		return truncate(float64(math.Float32frombits(uint32(a))), 32, true, false)
	case 0xa9:
		return truncate(float64(math.Float32frombits(uint32(a))), 32, false, false)
	case 0xaa:
		return truncate(math.Float64frombits(a), 32, true, false)
	case 0xab:
		return truncate(math.Float64frombits(a), 32, false, false)
	case 0xae:
		return truncate(float64(math.Float32frombits(uint32(a))), 64, true, false)
	case 0xaf:
		return truncate(float64(math.Float32frombits(uint32(a))), 64, false, false)
	case 0xb0:
		return truncate(math.Float64frombits(a), 64, true, false)
	case 0xb1:
		return truncate(math.Float64frombits(a), 64, false, false)
	case 0xac:
		return uint64(int64(int32(a))), wasm.NoTrap
	case 0xad:
		return uint64(uint32(a)), wasm.NoTrap
	case 0xb2:
		return uint64(result32(float32(int32(a)))), wasm.NoTrap
	case 0xb3:
		return uint64(result32(float32(uint32(a)))), wasm.NoTrap
	case 0xb4:
		return uint64(result32(float32(int64(a)))), wasm.NoTrap
	case 0xb5:
		return uint64(result32(float32(uint64(a)))), wasm.NoTrap
	case 0xb7:
		return uint64(result64(float64(int32(a)))), wasm.NoTrap
	case 0xb8:
		return uint64(result64(float64(uint32(a)))), wasm.NoTrap
	case 0xb9:
		return uint64(result64(float64(int64(a)))), wasm.NoTrap
	case 0xba:
		return uint64(result64(float64(uint64(a)))), wasm.NoTrap
	case 0xb6:
		return uint64(result32(float32(math.Float64frombits(a)))), wasm.NoTrap
	case 0xbb:
		return result64(float64(math.Float32frombits(uint32(a)))), wasm.NoTrap
	case 0xbc:
		return uint64(uint32(a)), wasm.NoTrap
	case 0xbe:
		return uint64(uint32(a)), wasm.NoTrap
	case 0xbd:
		return a, wasm.NoTrap
	case 0xbf:
		return a, wasm.NoTrap
	case 0xc0:
		return uint64(uint32(int8(a))), wasm.NoTrap
	case 0xc1:
		return uint64(uint32(int16(a))), wasm.NoTrap
	case 0xc2:
		return uint64(uint64(int8(a))), wasm.NoTrap
	case 0xc3:
		return uint64(uint64(int16(a))), wasm.NoTrap
	case 0xc4:
		return uint64(uint64(int32(a))), wasm.NoTrap
	case 0xfc0000:
		return truncate(float64(math.Float32frombits(uint32(a))), 32, true, true)
	case 0xfc0001:
		return truncate(float64(math.Float32frombits(uint32(a))), 32, false, true)
	case 0xfc0002:
		return truncate(math.Float64frombits(a), 32, true, true)
	case 0xfc0003:
		return truncate(math.Float64frombits(a), 32, false, true)
	case 0xfc0004:
		return truncate(float64(math.Float32frombits(uint32(a))), 64, true, true)
	case 0xfc0005:
		return truncate(float64(math.Float32frombits(uint32(a))), 64, false, true)
	case 0xfc0006:
		return truncate(math.Float64frombits(a), 64, true, true)
	case 0xfc0007:
		return truncate(math.Float64frombits(a), 64, false, true)
	default:
		panic("numeric: nonnumeric opcode")
	}
}
func boolean(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

// Truncate first, then check exact power-of-two boundaries before converting.
// This accepts negative fractions that truncate to unsigned zero, and avoids
// host-dependent behavior for NaNs, infinities and out-of-range conversions.
func truncate(x float64, width int, signed, saturating bool) (uint64, wasm.TrapCode) {
	if math.IsNaN(x) {
		if saturating {
			return 0, wasm.NoTrap
		}
		return 0, wasm.TrapInvalidConversion
	}
	x = math.Trunc(x)
	upper := math.Ldexp(1, width)
	lower := float64(0)
	maximum := uint64(math.MaxUint64)
	minimum := uint64(0)
	if width == 32 {
		maximum = math.MaxUint32
	}
	if signed {
		upper /= 2
		lower = -upper
		maximum >>= 1
		minimum = maximum + 1
	}
	if x < lower {
		if saturating {
			return minimum, wasm.NoTrap
		}
		return 0, wasm.TrapIntegerOverflow
	}
	if x >= upper {
		if saturating {
			return maximum, wasm.NoTrap
		}
		return 0, wasm.TrapIntegerOverflow
	}
	var result uint64
	if signed {
		result = uint64(int64(x))
	} else {
		result = uint64(x)
	}
	if width == 32 {
		result = uint64(uint32(result))
	}
	return result, wasm.NoTrap
}

// Go math.Min/Max give an extremal infinity precedence over NaN. Wasm must
// propagate NaN first, including those pairs; canonical quiet NaN is permitted.
func minimum(a, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.Float64frombits(0x7ff8000000000000)
	}
	return math.Min(a, b)
}
func maximum(a, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.Float64frombits(0x7ff8000000000000)
	}
	return math.Max(a, b)
}

// Go/LLGo may use different math implementations, whose NaN payloads and
// quieting are not part of Go's API contract. Arithmetic Wasm results require
// quiet NaNs, and finite/canonical inputs require canonical NaNs. Normalize
// arithmetic results only; abs/neg/copysign and reinterpret preserve raw bits.
func result32(x float32) uint32 {
	b := math.Float32bits(x)
	if b&0x7fffffff > 0x7f800000 {
		return 0x7fc00000
	}
	return b
}
func result64(x float64) uint64 {
	b := math.Float64bits(x)
	if b&0x7fffffffffffffff > 0x7ff0000000000000 {
		return 0x7ff8000000000000
	}
	return b
}

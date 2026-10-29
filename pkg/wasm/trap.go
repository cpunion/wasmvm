package wasm

// TrapCode identifies a guest execution trap independently of host errors and
// compiler/resource-limit failures. NoTrap is the successful numeric outcome.
type TrapCode uint8

const (
	NoTrap TrapCode = iota
	TrapUnreachable
	TrapMemoryBounds
	TrapTableBounds
	TrapUninitializedElement
	TrapIndirectCallType
	TrapIntegerOverflow
	TrapIntegerDivideByZero
	TrapInvalidConversion
	TrapNullReference
	TrapArrayBounds
	TrapCastFailure
)

func (c TrapCode) String() string {
	switch c {
	case NoTrap:
		return "no trap"
	case TrapUnreachable:
		return "unreachable"
	case TrapMemoryBounds:
		return "out of bounds memory access"
	case TrapTableBounds:
		return "out of bounds table access"
	case TrapUninitializedElement:
		return "uninitialized element"
	case TrapIndirectCallType:
		return "indirect call type mismatch"
	case TrapIntegerOverflow:
		return "integer overflow"
	case TrapIntegerDivideByZero:
		return "integer divide by zero"
	case TrapInvalidConversion:
		return "invalid conversion to integer"
	case TrapArrayBounds:
		return "out of bounds array access"
	case TrapCastFailure:
		return "cast failure"
	case TrapNullReference:
		return "null reference"
	default:
		return "unknown trap"
	}
}

package wasm

import (
	"errors"
	"fmt"
)

var (
	ErrMalformed   = errors.New("malformed WebAssembly binary")
	ErrInvalid     = errors.New("invalid WebAssembly module")
	ErrLimit       = errors.New("WebAssembly resource limit exceeded")
	ErrUnsupported = errors.New("unsupported WebAssembly feature")
	ErrTrap        = errors.New("WebAssembly execution trap")
)

// Error identifies a frontend failure at an absolute byte offset. Kind is one
// of the error categories above and is available through errors.Is.
type Error struct {
	Kind   error
	Offset uint64
	Detail string
}

func (e *Error) Error() string {
	return fmt.Sprintf("wasm byte %#x: %v: %s", e.Offset, e.Kind, e.Detail)
}
func (e *Error) Unwrap() error { return e.Kind }

// Trap reports a guest failure without conflating it with resource limits or
// host errors. Offset identifies the original Wasm instruction when available.
type Trap struct {
	Code     TrapCode
	Function uint32
	Offset   uint64
}

func (e *Trap) Error() string {
	return fmt.Sprintf("wasm function %d byte %#x: %s", e.Function, e.Offset, e.Code)
}
func (e *Trap) Unwrap() error { return ErrTrap }

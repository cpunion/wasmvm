package vm

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ScalarHostFunction receives integer and floating-point bit patterns. The
// stack has max(parameter count, result count) entries. Parameters occupy its
// beginning on entry; the callback writes results into its beginning on return.
// Each invocation owns an independent stack. Retaining it does not retain guest
// references, and later writes cannot change values already returned to Wasm.
// I32/F32 occupy the low 32 bits; I64/F64 occupy all 64 bits.
type ScalarHostFunction func(*Caller, []uint64) error

// BorrowedScalarHostFunction has ScalarHostFunction's bit-pattern convention,
// but its stack is borrowed only until callback return. It must not retain the
// stack or access it concurrently. Reentry receives separate storage. The Caller
// retains the usual expiration and concurrency checks.
type BorrowedScalarHostFunction func(*Caller, []uint64) error

// BorrowedSuspendableScalarHostFunction writes immediate results into its
// borrowed stack, or returns a HostSuspension using the ordinary continuation
// protocol. Neither the stack nor Caller may survive callback return. Pending
// work copies any needed scalar inputs and resumes with a fresh Caller.
type BorrowedSuspendableScalarHostFunction func(*Caller, []uint64) (*HostSuspension, error)

// NewScalarLeafFunction binds an internal Caller-free leaf with at most two
// scalar inputs and one scalar result.
func (c *Caller) NewScalarLeafFunction(signature FunctionType, leaf *state.ScalarLeaf) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	if !leaf.Valid() || len(signature.Params) > 2 || len(signature.Results) > 1 || !scalarSignature(signature) {
		return Extern{}, ErrType
	}
	return c.op.store.newHostFunction(signature, leaf)
}

func (s *Store) NewBorrowedSuspendableScalarFunction(signature FunctionType, callback BorrowedSuspendableScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	return s.newBorrowedSuspendableScalarFunction(signature, callback, referenceType...)
}

func (s *Store) newBorrowedSuspendableScalarFunction(signature FunctionType, callback BorrowedSuspendableScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if callback == nil || !scalarSignature(signature) {
		return Extern{}, ErrType
	}
	return s.newHostFunction(signature, callback, referenceType...)
}

func (c *Caller) NewBorrowedSuspendableScalarFunction(signature FunctionType, callback BorrowedSuspendableScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	return c.op.store.newBorrowedSuspendableScalarFunction(signature, callback, referenceType...)
}

// NewBorrowedScalarFunction avoids an owned parameter/result stack for callbacks
// that do not retain it. Scalar signatures and Caller contracts are unchanged.
func (s *Store) NewBorrowedScalarFunction(signature FunctionType, callback BorrowedScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	return s.newBorrowedScalarFunction(signature, callback, referenceType...)
}

func (s *Store) newBorrowedScalarFunction(signature FunctionType, callback BorrowedScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if callback == nil || !scalarSignature(signature) {
		return Extern{}, ErrType
	}
	return s.newHostFunction(signature, callback, referenceType...)
}

// NewBorrowedScalarFunction defines a borrowed-stack callback in this Store.
func (c *Caller) NewBorrowedScalarFunction(signature FunctionType, callback BorrowedScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	return c.op.store.newBorrowedScalarFunction(signature, callback, referenceType...)
}

// NewScalarFunction defines a synchronous host function with only I32, I64,
// F32 and F64 parameters/results. It has the same Caller, reentry, cancellation
// and cleanup contracts as NewFunction, without public reference-value encoding.
func (s *Store) NewScalarFunction(signature FunctionType, callback ScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	return s.newScalarFunction(signature, callback, referenceType...)
}

func (s *Store) newScalarFunction(signature FunctionType, callback ScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if callback == nil || !scalarSignature(signature) {
		return Extern{}, ErrType
	}
	return s.newHostFunction(signature, callback, referenceType...)
}

func scalarSignature(signature FunctionType) bool {
	for _, types := range [][]wasm.ValueType{signature.Params, signature.Results} {
		for _, typ := range types {
			switch typ {
			case wasm.I32, wasm.I64, wasm.F32, wasm.F64:
			default:
				return false
			}
		}
	}
	return true
}

// NewScalarFunction defines a function in the current callback's Store.
func (c *Caller) NewScalarFunction(signature FunctionType, callback ScalarHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	return c.op.store.newScalarFunction(signature, callback, referenceType...)
}

func scalarStack(args []value.Value, signature FunctionType) []uint64 {
	stack := make([]uint64, max(len(signature.Params), len(signature.Results)))
	for i, arg := range args {
		bits := arg.Lo
		if typ := signature.Params[i]; typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		stack[i] = bits
	}
	return stack
}

// The execution's host carriers are pointer-free and have capacity for both
// parameters and results. Pack Lo words forward, then unpack results backward:
// source and destination overlap. Scalar-only signatures contain no guest roots
// that collection or nested reentry could misinterpret during the borrow.
func borrowedScalarStack(args []value.Value, signature FunctionType) []uint64 {
	n := max(len(signature.Params), len(signature.Results))
	if n == 0 {
		return nil
	}
	stack := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(args[:n]))), n)
	for i, typ := range signature.Params {
		bits := args[i].Lo
		if typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		stack[i] = bits
	}
	clear(stack[len(signature.Params):])
	return stack[:n:n]
}

func scalarResults(storage []value.Value, stack []uint64, types []wasm.ValueType) []value.Value {
	results := storage[:len(types)]
	for i := len(types) - 1; i >= 0; i-- {
		typ := types[i]
		bits := stack[i]
		if typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		results[i] = value.Value{Lo: bits}
	}
	return results
}

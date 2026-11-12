package vm

import (
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

func scalarResults(storage []value.Value, stack []uint64, types []wasm.ValueType) []value.Value {
	results := storage[:len(types)]
	for i, typ := range types {
		bits := stack[i]
		if typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		results[i] = value.Value{Lo: bits}
	}
	return results
}

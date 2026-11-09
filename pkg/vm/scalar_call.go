package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ScalarFunc binds a numeric-only function and reusable private argument storage.
// Calls use the same execution, fuel, cancellation and lifetime rules as Call.
// Concurrent calls on this binding or its Store return ErrBusy. Host callbacks
// use Caller.Call for controlled reentry.
type ScalarFunc struct {
	mu        sync.Mutex
	function  Extern
	signature FunctionType
	arguments []value.Value
}

// Scalar binds an existing function with only I32, I64, F32 and F64 parameters
// and results. References and V128 continue to use the typed Value API.
func (ex Extern) Scalar() (*ScalarFunc, error) {
	signature, err := ex.FunctionType()
	if err != nil {
		return nil, err
	}
	if !scalarSignature(signature) {
		return nil, ErrType
	}
	return &ScalarFunc{function: ex, signature: signature, arguments: make([]value.Value, len(signature.Params))}, nil
}

// Call receives raw numeric bits in stack, whose length must equal the larger
// of parameter/result counts. Parameters occupy its prefix on entry; results
// replace its prefix only on success. Other words remain unchanged. I32/F32 use
// the low 32 bits, I64/F64 all 64 bits. The caller owns stack and must not mutate
// it during the call. Fuel remaining is returned separately from numeric data.
func (fn *ScalarFunc) Call(ctx context.Context, stack []uint64, options CallOptions) (uint64, error) {
	if fn == nil || fn.function.store == nil {
		return 0, ErrReference
	}
	if !fn.mu.TryLock() {
		return 0, ErrBusy
	}
	defer fn.mu.Unlock()
	op, err := fn.function.store.begin(ctx, options)
	if err != nil {
		return 0, err
	}
	defer op.finish()
	if len(stack) != max(len(fn.signature.Params), len(fn.signature.Results)) {
		return 0, ErrType
	}
	for i, typ := range fn.signature.Params {
		bits := stack[i]
		if typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		fn.arguments[i] = value.Value{Lo: bits}
	}
	result, err := op.call(fn.function.external.Function, fn.arguments)
	if err != nil {
		return op.fuel, err
	}
	for i, typ := range fn.signature.Results {
		bits := result[i].Lo
		if typ == wasm.I32 || typ == wasm.F32 {
			bits = uint64(uint32(bits))
		}
		stack[i] = bits
	}
	return op.fuel, nil
}

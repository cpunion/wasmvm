package vm

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ScalarStorage supplies a fixed number of integer cells to accessor functions.
// Both backends access these cells without a host Caller. The native backend
// copies them into owned native memory and publishes writes before returning
// to any host code, including on cancellation, fuel exhaustion and traps.
type ScalarStorage struct{ core state.ScalarStorage }

// NewScalarStorage creates a storage binding. bind runs on the Store's serialized
// execution thread; it must not reenter the VM or change the binding until that
// entry returns. It returns exactly size words, or false to trap accesses. A
// later entry may bind different words, for example after a task switch.
// Accessor functions sharing storage must use the same ScalarStorage object.
func NewScalarStorage(size uint32, bind func() ([]uint64, bool)) (*ScalarStorage, error) {
	if size == 0 || bind == nil {
		return nil, ErrType
	}
	return &ScalarStorage{core: state.ScalarStorage{Words: size, Bind: bind}}, nil
}

// NewScalarStorageFunction defines an I32/I64 getter (no parameters, one result)
// or setter (one parameter, no results). Setters zero-extend I32 into the cell.
// Storage must be used by one Store at a time, just like host callback state.
func (s *Store) NewScalarStorageFunction(storage *ScalarStorage, index uint32, typ wasm.ValueType, write bool, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	return s.newScalarStorageFunction(storage, index, typ, write, referenceType...)
}

// NewScalarStorageFunction defines an accessor in the active callback's Store.
func (c *Caller) NewScalarStorageFunction(storage *ScalarStorage, index uint32, typ wasm.ValueType, write bool, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	return c.op.store.newScalarStorageFunction(storage, index, typ, write, referenceType...)
}

func (s *Store) newScalarStorageFunction(storage *ScalarStorage, index uint32, typ wasm.ValueType, write bool, referenceType ...wasm.ValueType) (Extern, error) {
	if storage == nil || index >= storage.core.Words || typ != wasm.I32 && typ != wasm.I64 {
		return Extern{}, ErrType
	}
	signature := FunctionType{Results: []wasm.ValueType{typ}}
	if write {
		signature.Params, signature.Results = signature.Results, nil
	}
	return s.newHostFunction(signature, &state.ScalarAccess{Storage: &storage.core, Index: index, Type: typ, Write: write}, referenceType...)
}

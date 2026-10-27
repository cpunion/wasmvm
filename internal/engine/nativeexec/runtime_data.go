package nativeexec

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// The common frame machine retains every active/suspended instance and
// serializes table access, growth and reclamation. This execution owns only
// the native metadata array; table storage belongs to the Store. No Go address
// crosses the ABI. Every entry refreshes table pointers after intervening host
// calls/helpers. Immutable function identities are copied only on instance change.
func (e *Execution) bindRuntimeData(in *state.Instance) (uint64, error) {
	words := uint64(len(in.Functions)) + uint64(len(in.Tables))*2
	if words > uint64(^uint(0)>>1)/8 {
		return 0, wasm.ErrLimit
	}
	if words > uint64(len(e.runtimeWords)) {
		data, err := native.NewMemory(words*8, words*8)
		if err != nil {
			return 0, err
		}
		buffer, err := data.Buffer()
		if err != nil {
			data.Close()
			return 0, err
		}
		if e.runtimeData != nil {
			e.runtimeData.Close()
		}
		e.runtimeData, e.runtimeInstance = data, nil
		e.runtimeWords = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), int(words))
	}
	if e.runtimeInstance != in {
		copy(e.runtimeWords, in.Functions)
		e.runtimeInstance = in
	}
	for i, table := range in.Tables {
		address := table.NativeAddress()
		if len(table.Values) != 0 && address == 0 {
			return 0, wasm.ErrInvalid
		}
		offset := len(in.Functions) + 2*i
		e.runtimeWords[offset], e.runtimeWords[offset+1] = uint64(len(table.Values)), address
	}
	return uint64(uintptr(unsafe.Pointer(unsafe.SliceData(e.runtimeWords)))), nil
}

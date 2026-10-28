package nativeexec

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type globalBinding struct {
	global *state.Global
	word   int
}

func (e *Execution) resetRuntimeGlobals() {
	clear(e.runtimeGlobals)
	e.runtimeGlobals = e.runtimeGlobals[:0]
	e.runtimeFunction = nil
}

// The common frame machine retains every active/suspended instance and
// serializes access, growth and reclamation. This execution owns the metadata
// and numeric cells; table storage belongs to the Store. No Go address
// crosses the ABI. Every entry refreshes table pointers after intervening host
// calls/helpers. Immutable function identities are copied only on instance change.
func (e *Execution) bindRuntimeData(in *state.Instance) (uint64, error) {
	return e.bindRuntimeGlobals(in, nil, nil)
}

func (e *Execution) bindRuntimeGlobals(in *state.Instance, f *slots.Function, indices []uint32) (uint64, error) {
	descriptors := uint64(len(in.Functions)) + uint64(len(in.Tables))*2
	words := descriptors
	if len(indices) != 0 {
		words += uint64(len(in.Globals)) + uint64(len(indices))*2
	}
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
		e.resetRuntimeGlobals()
		e.runtimeWords = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), int(words))
	}
	changed := e.runtimeInstance != in
	if changed {
		copy(e.runtimeWords, in.Functions)
		e.runtimeInstance = in
	}
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(e.runtimeWords))))
	if changed || e.runtimeFunction != f {
		e.resetRuntimeGlobals()
		cells := map[*state.Global]int{}
		for _, index := range indices {
			if uint64(index) >= uint64(len(in.Globals)) || in.Globals[index] == nil || in.Globals[index].Type.Value.IsReference() {
				return 0, wasm.ErrInvalid
			}
			g := in.Globals[index]
			word, exists := cells[g]
			if !exists {
				word = int(descriptors) + len(in.Globals) + len(cells)*2
				cells[g] = word
				e.runtimeGlobals = append(e.runtimeGlobals, globalBinding{g, word})
			}
			e.runtimeWords[int(descriptors)+int(index)] = address + uint64(word)*8
		}
		e.runtimeFunction = f
	}
	for _, binding := range e.runtimeGlobals {
		e.runtimeWords[binding.word] = binding.global.Value.Lo
		e.runtimeWords[binding.word+1] = binding.global.Value.Hi
	}
	for i, table := range in.Tables {
		address := table.NativeAddress()
		if len(table.Values) != 0 && address == 0 {
			return 0, wasm.ErrInvalid
		}
		offset := len(in.Functions) + 2*i
		e.runtimeWords[offset], e.runtimeWords[offset+1] = uint64(len(table.Values)), address
	}
	return address, nil
}

// Publish even on traps, fuel exhaustion and cancellation. The driver must see
// writes preceding the exit before any helper, host callback or guest call.
func (e *Execution) flushRuntimeGlobals() {
	for _, binding := range e.runtimeGlobals {
		binding.global.Value.Lo = e.runtimeWords[binding.word]
		binding.global.Value.Hi = e.runtimeWords[binding.word+1]
	}
}

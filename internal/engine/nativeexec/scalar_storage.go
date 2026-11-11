package nativeexec

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type scalarBinding struct {
	storage *state.ScalarStorage
	word    int
	values  []uint64
}

type scalarAccess struct {
	index, binding int
	access         *state.ScalarAccess
}

func (e *Execution) resetScalarStorage() {
	clear(e.scalarBindings)
	clear(e.scalarAccesses)
	e.scalarBindings, e.scalarAccesses = e.scalarBindings[:0], e.scalarAccesses[:0]
	e.scalarInstance = nil
}

// Every alias shares one native cell and one validity word. Bind once per
// storage per entry, then write back before any host callback or helper. Only
// native-owned addresses enter machine code; task slices stay on this side.
func (e *Execution) bindScalarStorage(in *state.Instance) (uint64, error) {
	changed := e.scalarInstance != in
	if changed {
		e.resetScalarStorage()
		words := uint64(len(in.Functions)) * 2
		bindings := map[*state.ScalarStorage]int{}
		if in.Owner != nil {
			for i, ref := range in.Functions {
				fn, ok := in.Owner.Function(ref)
				if !ok {
					return 0, wasm.ErrInvalid
				}
				access, ok := fn.Host.(*state.ScalarAccess)
				if !ok {
					continue
				}
				binding, exists := bindings[access.Storage]
				if !exists {
					if words+1+uint64(access.Storage.Words) > uint64(^uint(0)>>1)/8 {
						return 0, wasm.ErrLimit
					}
					binding = len(e.scalarBindings)
					bindings[access.Storage] = binding
					e.scalarBindings = append(e.scalarBindings, scalarBinding{storage: access.Storage, word: int(words)})
					words += 1 + uint64(access.Storage.Words)
				}
				e.scalarAccesses = append(e.scalarAccesses, scalarAccess{i, binding, access})
			}
		}
		if len(e.scalarBindings) != 0 && words > uint64(len(e.scalarWords)) {
			data, err := native.NewMemory(words*8, words*8)
			if err != nil {
				return 0, err
			}
			buffer, err := data.Buffer()
			if err != nil {
				data.Close()
				return 0, err
			}
			if e.scalarData != nil {
				e.scalarData.Close()
			}
			e.scalarData = data
			e.scalarWords = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), int(words))
		}
		e.scalarInstance = in
	}
	if len(e.scalarBindings) == 0 {
		return 0, nil
	}
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(e.scalarWords))))
	if changed {
		clear(e.scalarWords)
		for _, a := range e.scalarAccesses {
			word := e.scalarBindings[a.binding].word
			e.scalarWords[a.index*2] = address + uint64(word+1+int(a.access.Index))*8
			e.scalarWords[a.index*2+1] = address + uint64(word)*8
		}
	}
	for i := range e.scalarBindings {
		b := &e.scalarBindings[i]
		values, valid := b.storage.Bind()
		b.values = nil
		e.scalarWords[b.word] = 0
		if valid && len(values) == int(b.storage.Words) {
			b.values = values
			e.scalarWords[b.word] = 1
			copy(e.scalarWords[b.word+1:], values)
		}
	}
	return address, nil
}

func (e *Execution) flushScalarStorage() {
	for _, b := range e.scalarBindings {
		copy(b.values, e.scalarWords[b.word+1:b.word+1+len(b.values)])
	}
}

package state

import "github.com/cpunion/wasmvm/pkg/wasm"

// ScalarStorage binds serialized host integer storage for one execution entry.
// The binding stays valid until that entry returns to the driver.
type ScalarStorage struct {
	Words uint32
	Bind  func() ([]uint64, bool)
}

type ScalarAccess struct {
	Storage *ScalarStorage
	Index   uint32
	Type    wasm.ValueType
	Write   bool
}

func (a *ScalarAccess) Access(bits uint64) (uint64, error) {
	words, valid := a.Storage.Bind()
	if !valid || len(words) != int(a.Storage.Words) || uint64(a.Index) >= uint64(len(words)) {
		return 0, wasm.ErrTrap
	}
	if a.Write {
		if a.Type == wasm.I32 {
			bits = uint64(uint32(bits))
		}
		words[a.Index] = bits
		return 0, nil
	}
	bits = words[a.Index]
	if a.Type == wasm.I32 {
		bits = uint64(uint32(bits))
	}
	return bits, nil
}

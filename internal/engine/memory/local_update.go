package memory

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// UpdateLocalI32 checks an address once for an ordinary load and store that
// access the same bytes. Memories accessible from other Stores keep their
// separately synchronized accesses. The caller must hold the serial Store
// lease throughout the operation.
func (m *Memory) UpdateLocalI32(address, offset uint64, op uint32, rhs uint64) (old, next uint64, trap wasm.TrapCode, applied bool) {
	if m.Shareable() {
		return 0, 0, wasm.NoTrap, false
	}
	bytes, ok := m.access(address, offset, 4)
	if !ok {
		return 0, 0, wasm.TrapMemoryBounds, true
	}
	old = uint64(binary.LittleEndian.Uint32(bytes))
	switch op {
	case 0x6a:
		next = numeric.I32Add(old, rhs)
	case 0x6b:
		next = numeric.I32Sub(old, rhs)
	case 0x6c:
		next = numeric.I32Mul(old, rhs)
	case 0x71:
		next = numeric.I32And(old, rhs)
	case 0x72:
		next = numeric.I32Or(old, rhs)
	case 0x73:
		next = numeric.I32Xor(old, rhs)
	default:
		return 0, 0, wasm.NoTrap, false
	}
	binary.LittleEndian.PutUint32(bytes, uint32(next))
	return old, next, wasm.NoTrap, true
}

package vm

import (
	"github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"sync"
)

// SharedMemory is an explicit, independently retained capability. Close stops
// future imports and releases this root; existing Store imports remain valid.
// A shared backing can only be imported into Stores of its original Engine.
type SharedMemory struct {
	mu     sync.Mutex
	engine *Engine
	memory *memory.Memory
}

func (ex Extern) ShareMemory() (*SharedMemory, error) {
	op, err := ex.access(MemoryKind)
	if err != nil {
		return nil, err
	}
	defer op.finish()
	m := ex.external.Memory.Memory
	if !m.Shared() {
		return nil, ErrType
	}
	if !m.Retain() {
		return nil, ErrReference
	}
	return &SharedMemory{engine: ex.store.engine, memory: m}, nil
}
func (m *SharedMemory) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.memory != nil {
		m.memory.Release()
		m.memory = nil
	}
}
func (s *Store) ImportMemory(m *SharedMemory) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if s.engine.config.Features&wasm.Threads == 0 {
		return Extern{}, wasm.ErrUnsupported
	}
	if m == nil {
		return Extern{}, ErrReference
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.memory == nil {
		return Extern{}, wasm.ErrClosed
	}
	if m.engine != s.engine {
		return Extern{}, ErrReference
	}
	ex, err := s.state.ImportShared(m.memory)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}

package state

import (
	"github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ImportShared creates a Store-owned binding to one domain-owned backing.
// Aliases deduplicate within a Store. Every binding reserves the backing's
// host maximum, so another Store's growth cannot exceed this Store's quota.
func (s *Store) ImportShared(m *memory.Memory) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if m == nil || !m.Shared() || m.Domain() != s.Domain {
		return Extern{}, ErrReference
	}
	for _, existing := range s.memories {
		if existing.Memory == m {
			existing.OwnerInstance = nil
			return Extern{Owner: s, Kind: 2, Memory: existing}, nil
		}
	}
	if s.usage.Memories >= uint64(s.limits.MaxMemories) || m.ReservationPages() > uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages-s.reservedShared {
		return Extern{}, limit("shared memory reservation")
	}
	if !m.Retain() {
		return Extern{}, ErrReference
	}
	owned := &Memory{Owner: s, Memory: m}
	s.memories = append(s.memories, owned)
	s.usage.Memories++
	s.reservedShared += m.ReservationPages()
	return Extern{Owner: s, Kind: 2, Memory: owned}, nil
}

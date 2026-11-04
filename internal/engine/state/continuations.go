package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ContinuationBody owns a bounded, pointer-free guest stack snapshot plus its
// ordinary host metadata. The frame machine supplies tracing without a cyclic
// dependency from Store/GC back to an execution backend.
type ContinuationBody interface {
	HeapBytes() uint64
	Visit(func(wasm.ValueType, value.Value), func(*Instance))
}

type Suspension struct {
	Owner  *Store
	Tag    *Tag
	Args   []value.Value
	Switch bool
}

func (*Suspension) Error() string { return wasm.ErrSuspension.Error() }
func (*Suspension) Unwrap() error { return wasm.ErrSuspension }

func (s *Store) ContinuationSignature(typ uint32) (module.FuncType, bool) {
	if typ == 0 || uint64(typ) >= uint64(len(s.registry.Types)) || s.registry.Types[typ].Definition.Kind != module.ContinuationKind {
		return module.FuncType{}, false
	}
	return s.registry.Types[s.registry.Types[typ].Definition.Continuation].Function, true
}

func (s *Store) NewContinuation(typ uint32, body ContinuationBody) (uint64, error) {
	if s.closed {
		return 0, wasm.ErrClosed
	}
	if body == nil || typ == 0 || uint64(typ) >= uint64(len(s.registry.Types)) || s.registry.Types[typ].Definition.Kind != module.ContinuationKind {
		return 0, wasm.ErrInvalid
	}
	ref, err := s.allocateObject(typ, 0, body.HeapBytes())
	if err != nil {
		return 0, err
	}
	s.objectExtras(ref).Continuation = body
	return ref, nil
}

// Take invalidates the one-shot body before guest execution can observe it.
// The consumed handle remains valid as a value; invoking it again traps.
func (s *Store) Continuation(ref uint64, take bool) (ContinuationBody, error) {
	if ref == 0 {
		return nil, &wasm.Trap{Code: wasm.TrapNullReference}
	}
	o, ok := s.heapObject(ref)
	if !ok || s.registry.Types[o.Type].Definition.Kind != module.ContinuationKind {
		return nil, ErrReference
	}
	extra := s.existingObjectExtras(o)
	if extra == nil || extra.Continuation == nil {
		return nil, &wasm.Trap{Code: wasm.TrapConsumedContinuation}
	}
	body := extra.Continuation
	if take {
		bytes := body.HeapBytes()
		s.usage.HeapBytes -= bytes
		extra.Bytes -= bytes
		extra.Continuation = nil
	}
	return body, nil
}

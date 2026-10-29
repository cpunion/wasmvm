package vm

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"math"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Value carries exact numeric bits or an opaque Store-owned reference. Copies
// of a reference Value share one root; Root creates an independent root.
// Release invalidates all copies of that root. Null references need no release.
type Value struct {
	kind wasm.ValueType
	bits value.Value
	ref  *reference
}
type reference struct {
	store    *Store
	kind     wasm.ValueType
	id       uint64
	released bool
}

func rawReference(id uint64) value.Value { return value.Value{Lo: id} }
func I32(v uint32) Value                 { return Value{kind: wasm.I32, bits: value.Value{Lo: uint64(v)}} }
func I64(v uint64) Value                 { return Value{kind: wasm.I64, bits: value.Value{Lo: v}} }
func F32(v float32) Value                { return F32Bits(math.Float32bits(v)) }
func F64(v float64) Value                { return F64Bits(math.Float64bits(v)) }
func F32Bits(v uint32) Value             { return Value{kind: wasm.F32, bits: value.Value{Lo: uint64(v)}} }
func F64Bits(v uint64) Value             { return Value{kind: wasm.F64, bits: value.Value{Lo: v}} }
func V128(lo, hi uint64) Value           { return Value{kind: wasm.V128, bits: value.Value{Lo: lo, Hi: hi}} }
func NullRef(t wasm.ValueType) (Value, error) {
	if !t.IsReference() || !t.Nullable() {
		return Value{}, ErrType
	}
	return Value{kind: t}, nil
}
func (v Value) Type() wasm.ValueType { return v.kind }
func (v Value) IsNull() bool         { return v.kind.IsReference() && v.ref == nil }

// Bits exposes only numeric/vector bits; it cannot manufacture references.
func (v Value) Bits() (lo, hi uint64, err error) {
	if !v.kind.Valid() || v.kind.IsReference() {
		return 0, 0, ErrType
	}
	return v.bits.Lo, v.bits.Hi, nil
}
func (v Value) Release() {
	if v.ref == nil {
		return
	}
	s := v.ref.store
	s.mu.Lock()
	s.releaseLocked(v.ref)
	s.mu.Unlock()
}
func (s *Store) releaseLocked(r *reference) { r.released = true; delete(s.roots, r) }
func (s *Store) rawLocked(v Value, t wasm.ValueType) (value.Value, error) {
	if !t.Valid() {
		return value.Value{}, ErrType
	}
	if !t.IsReference() {
		if v.kind != t {
			return value.Value{}, ErrType
		}
		return v.bits, nil
	}
	if !v.kind.IsReference() || v.kind.Storage() != t.Storage() {
		return value.Value{}, ErrType
	}
	if v.ref == nil {
		if !t.Nullable() {
			return value.Value{}, ErrType
		}
		// Null still carries a static heap type; preserve its import/call contract.
		if !s.state.ReferenceSubtype(v.kind, t) {
			return value.Value{}, ErrType
		}
		return value.Value{}, nil
	}
	if v.ref.store != s || v.ref.released || s.closing {
		return value.Value{}, ErrReference
	}
	// Existing roots pin their identity until release. Same-type copies and
	// abstract host access must not inspect guest state while it is executing.
	if t == v.kind || t == wasm.FuncRef || t == wasm.ExternRef || t == wasm.ExnRef {
		return value.Value{Lo: v.ref.id}, nil
	}
	if !s.state.IsLiveReference(v.ref.id) {
		return value.Value{}, ErrReference
	}
	if !s.state.TestReference(v.ref.id, t) {
		return value.Value{}, ErrType
	}

	return value.Value{Lo: v.ref.id}, nil
}
func (s *Store) rootLocked(t wasm.ValueType, raw value.Value) (Value, error) {
	v := Value{kind: t, bits: raw}
	if !t.IsReference() {
		return v, nil
	}
	v.bits = value.Value{}
	if raw.Lo == 0 {
		return v, nil
	}
	if uint64(len(s.roots)) >= uint64(s.config.MaxHostRoots) {
		return Value{}, resourceLimit("host roots")
	}
	r := &reference{store: s, kind: t, id: raw.Lo}
	s.roots[r] = struct{}{}
	v.ref = r
	return v, nil
}

// Root duplicates a valid reference root. It is safe while a guest is running;
// collection and reference-table mutations share the Store mutex.
func (s *Store) Root(v Value) (Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return Value{}, wasm.ErrClosed
	}
	raw, err := s.rawLocked(v, v.kind)
	if err != nil {
		return Value{}, err
	}
	return s.rootLocked(v.kind, raw)
}
func (s *Store) NewExternRef(object any) (Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return Value{}, wasm.ErrClosed
	}
	if v, ok := object.(Value); ok && v.ref != nil && v.kind.IsExternalReference() && state.IsExternalizedGuest(v.ref.id) {
		raw, err := s.rawLocked(v, wasm.ExternRef)
		if err != nil {
			return Value{}, err
		}
		return s.rootLocked(wasm.ExternRef, raw)
	}
	if uint64(len(s.externs)) >= uint64(s.config.MaxExternRefs) || s.nextExtern >= state.MaxExternalIdentity {
		return Value{}, resourceLimit("external references")
	}
	id := s.nextExtern + 1
	v, err := s.rootLocked(wasm.ExternRef, value.Value{Lo: id})
	if err != nil {
		return Value{}, err
	}
	s.nextExtern = id
	s.externs[id] = object
	return v, nil
}

// ExternRef returns the rooted Go object. The table stores ordinary Go
// references; object addresses are never retained solely as integer bits.
func (s *Store) ExternRef(v Value) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return nil, wasm.ErrClosed
	}
	raw, err := s.rawLocked(v, wasm.ExternRef)
	if err != nil {
		return nil, err
	}
	if raw.Lo == 0 {
		return nil, nil
	}
	if state.IsExternalizedGuest(raw.Lo) {
		return v, nil
	}
	obj, ok := s.externs[raw.Lo]
	if !ok {
		return nil, ErrReference
	}
	return obj, nil
}

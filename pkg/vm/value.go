package vm

import (
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
	if !t.IsReference() {
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
	if v.kind != t || !t.Valid() {
		return value.Value{}, ErrType
	}
	if !t.IsReference() {
		return v.bits, nil
	}
	if v.ref == nil {
		return value.Value{}, nil
	}
	if v.ref.store != s || v.ref.released || s.closing {
		return value.Value{}, ErrReference
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
	if uint64(len(s.externs)) >= uint64(s.config.MaxExternRefs) || s.nextExtern == math.MaxUint64 {
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
	obj, ok := s.externs[raw.Lo]
	if !ok {
		return nil, ErrReference
	}
	return obj, nil
}

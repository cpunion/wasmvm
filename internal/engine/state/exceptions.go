package state

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Tag identity is independent of its canonical payload signature.
type Tag struct {
	Owner         *Store
	OwnerInstance *Instance
	Signature     module.FuncType
	TypeID        uint32
}

// Exception is an internal borrowed exception handle. The embedding driver must
// root it before releasing the failed execution's lease.
type Exception struct {
	Owner *Store
	Ref   uint64
}

func (*Exception) Error() string { return wasm.ErrException.Error() }
func (*Exception) Unwrap() error { return wasm.ErrException }
func (s *Store) NewTag(t module.FuncType, referenceType ...wasm.ValueType) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if len(t.Results) != 0 {
		return Extern{}, wasm.ErrInvalid
	}
	if s.usage.Tags >= uint64(s.limits.MaxTags) {
		return Extern{}, limit("tags")
	}
	t, id, err := s.hostSignature(t, referenceType)
	if err != nil {
		return Extern{}, err
	}
	return s.newTag(t, id, nil)
}

func (s *Store) newTag(t module.FuncType, id uint32, owner *Instance) (Extern, error) {
	if s.usage.Tags >= uint64(s.limits.MaxTags) {
		return Extern{}, limit("tags")
	}
	tag := &Tag{Owner: s, Signature: t, TypeID: id, OwnerInstance: owner}
	s.tags = append(s.tags, tag)
	s.usage.Tags++
	return Extern{Owner: s, Kind: 4, Tag: tag}, nil
}
func (s *Store) NewException(tag *Tag, args []value.Value, roots ...func(func(wasm.ValueType, value.Value))) (*Exception, error) {
	if s.closed {
		return nil, wasm.ErrClosed
	}
	if tag == nil || tag.Owner != s {
		return nil, ErrReference
	}
	if len(args) != len(tag.Signature.Params) {
		return nil, wasm.ErrInvalid
	}
	mark := len(s.temporary)
	allocating := len(s.allocating)
	if tag.OwnerInstance != nil {
		s.allocating = append(s.allocating, tag.OwnerInstance)
	}
	defer func() {
		clear(s.temporary[mark:])
		s.temporary = s.temporary[:mark]
		clear(s.allocating[allocating:])
		s.allocating = s.allocating[:allocating]
	}()
	for i, t := range tag.Signature.Params {
		if t.IsReference() {
			if !s.TestReference(args[i].Lo, t) {
				return nil, ErrReference
			}
			s.temporary = append(s.temporary, ReferenceValue{t, args[i]})
		}
	}
	if s.HeapSafepoint == nil {
		s.HeapSafepoint = func() {
			var xs []ReferenceValue
			if len(roots) != 0 {
				roots[0](func(t wasm.ValueType, v value.Value) { xs = append(xs, ReferenceValue{t, v}) })
			}
			s.CollectReferences(xs, append([]*Instance{tag.OwnerInstance}, s.instances...), nil)
		}
		defer func() { s.HeapSafepoint = nil }()
	}
	ref, err := s.allocateObject(tag.TypeID, uint32(len(args)))
	if err != nil {
		return nil, err
	}
	o, _ := s.heapObject(ref)
	o.Tag = tag
	for i, v := range args {
		o.set(uint32(i), v)
	}
	return &Exception{Owner: s, Ref: ref}, nil
}
func (s *Store) Exception(ref uint64) (*Tag, []value.Value, error) {
	o, ok := s.heapObject(ref)
	if !ok || o.Tag == nil {
		return nil, nil, ErrReference
	}
	args := make([]value.Value, len(o.Tag.Signature.Params))
	for i := range args {
		args[i] = o.get(uint32(i))
	}
	return o.Tag, args, nil
}
func (s *Store) ExceptionError(err error) (*Exception, bool) {
	var ex *Exception
	if !errors.As(err, &ex) || ex == nil || ex.Owner != s || !s.TestReference(ex.Ref, wasm.ExnRef.NonNull()) {
		return nil, false
	}
	return ex, true
}

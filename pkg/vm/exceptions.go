package vm

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Exception owns a host root for an exception reference. It can be returned
// from a HostFunction to throw into the guest. Release invalidates all copies
// of its Reference; Store.Root creates an independently owned reference.
type Exception struct{ reference Value }

func (*Exception) Error() string      { return wasm.ErrException.Error() }
func (*Exception) Unwrap() error      { return wasm.ErrException }
func (e *Exception) Reference() Value { return e.reference }
func (e *Exception) Release()         { e.reference.Release() }

func (s *Store) NewTag(signature FunctionType, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if s.engine.config.Features&wasm.ExceptionHandling == 0 {
		return Extern{}, wasm.ErrUnsupported
	}
	for _, t := range append(append([]wasm.ValueType(nil), signature.Params...), referenceType...) {
		if err := s.referenceFeature(t); err != nil {
			return Extern{}, err
		}
	}
	ex, err := s.state.NewTag(signature, referenceType...)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}
func (ex Extern) TagType() (FunctionType, error) {
	op, err := ex.access(TagKind)
	if err != nil {
		return FunctionType{}, err
	}
	defer op.finish()
	t := ex.external.Tag
	if t == nil || t.Owner != ex.store.state {
		return FunctionType{}, ErrReference
	}
	return FunctionType{Params: append([]wasm.ValueType(nil), t.Signature.Params...)}, nil
}

// SameTag compares Store-local tag identities, including re-exported imports.
func (ex Extern) SameTag(other Extern) bool {
	return ex.store != nil && ex.store == other.store && ex.Kind() == TagKind && other.Kind() == TagKind && ex.external.Tag != nil && ex.external.Tag == other.external.Tag
}
func (op *operation) newException(tag Extern, args []Value) (*Exception, error) {
	s := op.store
	if tag.store != s || tag.external.Tag == nil || tag.external.Tag.Owner != s.state {
		return nil, ErrReference
	}
	if tag.Kind() != TagKind {
		return nil, ErrType
	}
	s.mu.Lock()
	raw, err := s.decodeValuesLocked(args, tag.external.Tag.Signature.Params)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	ex, err := s.state.NewException(tag.external.Tag, raw)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	v, err := s.rootLocked(wasm.ExnRef.NonNull(), rawReference(ex.Ref))
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &Exception{reference: v}, nil
}
func (s *Store) NewException(tag Extern, args []Value) (*Exception, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return nil, err
	}
	defer op.finish()
	return op.newException(tag, args)
}

// NewException creates a callback-scoped exception. Root its Reference to keep
// it beyond the callback. Returning the exception throws before temporary roots drain.
func (c *Caller) NewException(tag Extern, args []Value) (*Exception, error) {
	if err := c.enter(); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	ex, err := c.op.newException(tag, args)
	if err == nil {
		c.temporary = append(c.temporary, ex.reference)
	}
	return ex, err
}

// Exception creates an independently rooted error from a guest exception value.
func (s *Store) Exception(v Value) (*Exception, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return nil, err
	}
	defer op.finish()
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := s.rawLocked(v, wasm.ExnRef.NonNull())
	if err != nil {
		return nil, err
	}
	root, err := s.rootLocked(wasm.ExnRef.NonNull(), raw)
	if err != nil {
		return nil, err
	}
	return &Exception{reference: root}, nil
}
func (e *Exception) Payload() (Extern, []Value, error) {
	if e == nil || e.reference.ref == nil {
		return Extern{}, nil, ErrReference
	}
	s := e.reference.ref.store
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, nil, err
	}
	defer op.finish()
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := s.rawLocked(e.reference, wasm.ExnRef.NonNull())
	if err != nil {
		return Extern{}, nil, err
	}
	tag, args, err := s.state.Exception(raw.Lo)
	if err != nil {
		return Extern{}, nil, err
	}
	values, err := s.encodeValuesLocked(args, tag.Signature.Params)
	return Extern{store: s, external: state.Extern{Owner: s.state, Kind: 4, Tag: tag}}, values, err
}
func (op *operation) exceptionError(err error) error {
	if ex, ok := op.store.state.ExceptionError(err); ok {
		s := op.store
		s.mu.Lock()
		v, e := s.rootLocked(wasm.ExnRef.NonNull(), rawReference(ex.Ref))
		s.mu.Unlock()
		if e != nil {
			return e
		}
		return &Exception{reference: v}
	}
	return err
}

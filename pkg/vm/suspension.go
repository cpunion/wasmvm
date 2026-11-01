package vm

import "github.com/cpunion/wasmvm/pkg/wasm"

// Suspension reports a control tag that reached the host without a matching
// guest prompt. It is terminal at this host boundary, distinct from a trap or
// an exception. Release drops its private payload root; Payload creates owned
// argument values with the same lifetime contract as Exception.Payload.
type Suspension struct {
	payload   *Exception
	switching bool
}

func (*Suspension) Error() string  { return wasm.ErrSuspension.Error() }
func (*Suspension) Unwrap() error  { return wasm.ErrSuspension }
func (s *Suspension) Switch() bool { return s != nil && s.switching }
func (s *Suspension) Release() {
	if s != nil && s.payload != nil {
		s.payload.Release()
	}
}
func (s *Suspension) Payload() (Extern, []Value, error) {
	if s == nil || s.payload == nil {
		return Extern{}, nil, ErrReference
	}
	return s.payload.Payload()
}

package vm

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (s *Store) NewMemory(typ MemoryType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if typ.Index64 && s.engine.config.Features&wasm.Memory64 == 0 {
		return Extern{}, wasm.ErrUnsupported
	}
	ex, err := s.state.NewMemory(typ)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}
func (s *Store) NewTable(typ TableType, initial ...Value) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if typ.Limits.Index64 && s.engine.config.Features&wasm.Memory64 == 0 {
		return Extern{}, wasm.ErrUnsupported
	}
	if err := s.referenceFeature(typ.Element); err != nil {
		return Extern{}, err
	}
	if len(initial) > 1 {
		return Extern{}, ErrType
	}
	var bits []uint64
	if len(initial) == 1 {
		s.mu.Lock()
		raw, e := s.rawLocked(initial[0], typ.Element)
		s.mu.Unlock()
		if e != nil {
			return Extern{}, e
		}
		bits = []uint64{raw.Lo}
	}
	ex, err := s.state.NewTable(typ, bits...)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}
func (s *Store) NewGlobal(typ GlobalType, initial Value) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	s.mu.Lock()
	raw, err := s.rawLocked(initial, typ.Value)
	s.mu.Unlock()
	if err != nil {
		return Extern{}, err
	}
	if err := s.referenceFeature(typ.Value); err != nil {
		return Extern{}, err
	}
	ex, err := s.state.NewGlobal(typ, raw)
	if err != nil {
		return Extern{}, err
	}
	return Extern{store: s, external: ex}, nil
}
func (ex Extern) access(kind ExternKind) (*operation, error) {
	if ex.store == nil {
		return nil, ErrReference
	}
	if ex.Kind() != kind {
		return nil, ErrType
	}
	return ex.store.begin(nil, CallOptions{})
}
func (ex Extern) ReadMemory(offset uint64, dst []byte) error {
	op, err := ex.access(MemoryKind)
	if err != nil {
		return err
	}
	defer op.finish()
	if !ex.external.Memory.Read(offset, dst) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}
func (ex Extern) WriteMemory(offset uint64, src []byte) error {
	op, err := ex.access(MemoryKind)
	if err != nil {
		return err
	}
	defer op.finish()
	if !ex.external.Memory.Write(offset, src) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}
func (ex Extern) MemoryPages() (uint64, error) {
	op, err := ex.access(MemoryKind)
	if err != nil {
		return 0, err
	}
	defer op.finish()
	return ex.external.Memory.Pages(), nil
}
func (ex Extern) GrowMemory(delta uint64) (old uint64, ok bool, err error) {
	op, err := ex.access(MemoryKind)
	if err != nil {
		return 0, false, err
	}
	defer op.finish()
	old, ok = ex.store.state.GrowMemory(ex.external.Memory, delta)
	return
}
func (ex Extern) TableSize() (uint64, error) {
	op, err := ex.access(TableKind)
	if err != nil {
		return 0, err
	}
	defer op.finish()
	return uint64(len(ex.external.Table.Values)), nil
}
func (ex Extern) TableGet(index uint64) (Value, error) {
	op, err := ex.access(TableKind)
	if err != nil {
		return Value{}, err
	}
	defer op.finish()
	t := ex.external.Table
	if uint64(index) >= uint64(len(t.Values)) {
		return Value{}, &wasm.Trap{Code: wasm.TrapTableBounds}
	}
	s := ex.store
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rootLocked(t.Type.Element, rawReference(t.Values[index]))
}
func (ex Extern) TableSet(index uint64, v Value) error {
	op, err := ex.access(TableKind)
	if err != nil {
		return err
	}
	defer op.finish()
	t := ex.external.Table
	if uint64(index) >= uint64(len(t.Values)) {
		return &wasm.Trap{Code: wasm.TrapTableBounds}
	}
	s := ex.store
	s.mu.Lock()
	raw, err := s.rawLocked(v, t.Type.Element)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	t.Values[index] = raw.Lo
	return nil
}
func (ex Extern) GrowTable(delta uint64, initial Value) (old uint64, ok bool, err error) {
	op, err := ex.access(TableKind)
	if err != nil {
		return 0, false, err
	}
	defer op.finish()
	s := ex.store
	s.mu.Lock()
	raw, err := s.rawLocked(initial, ex.external.Table.Type.Element)
	s.mu.Unlock()
	if err != nil {
		return 0, false, err
	}
	old, ok = s.state.GrowTable(ex.external.Table, delta, raw.Lo)
	return
}
func (ex Extern) GlobalGet() (Value, error) {
	op, err := ex.access(GlobalKind)
	if err != nil {
		return Value{}, err
	}
	defer op.finish()
	s := ex.store
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rootLocked(ex.external.Global.Type.Value, ex.external.Global.Value)
}
func (ex Extern) GlobalSet(v Value) error {
	op, err := ex.access(GlobalKind)
	if err != nil {
		return err
	}
	defer op.finish()
	g := ex.external.Global
	if !g.Type.Mutable {
		return ErrType
	}
	s := ex.store
	s.mu.Lock()
	raw, err := s.rawLocked(v, g.Type.Value)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	g.Value = raw
	return nil
}

func (s *Store) referenceFeature(typ wasm.ValueType) error {
	if typ.IsGCReference() && s.engine.config.Features&wasm.GC == 0 {
		return wasm.ErrUnsupported
	}
	if typ > 255 && typ.IsReference() && s.engine.config.Features&(wasm.FunctionReferences|wasm.GC) == 0 {
		return wasm.ErrUnsupported
	}
	return nil
}

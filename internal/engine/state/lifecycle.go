package state

import (
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Publish makes a successfully initialized instance a Store-lifetime root.
// The caller must finish its start function before publication.
func (s *Store) Publish(in *Instance) error {
	if s.closed {
		return wasm.ErrClosed
	}
	if in == nil || in.Owner != s || in.disposed {
		return ErrReference
	}
	in.published = true
	return nil
}

// HasUnpublished reports whether any instance can still become unreachable.
// Published instances and host-owned objects remain Store-lifetime roots.
func (s *Store) HasUnpublished() bool {
	for _, in := range s.instances {
		if !in.published {
			return true
		}
	}
	return false
}

// CollectFailed reclaims unpublished instances unreachable from published
// instances, host-owned objects, explicit function roots, or active instances.
// It runs only at a serialized safepoint. The driver must supply every active
// or suspended frame's instance and every host-held function reference. Root
// discovery is iterative, so cycles and deep instance chains cannot recurse
// through the host stack. Imported effects are never rolled back.
func (s *Store) CollectFailed(functionRoots []uint64, active []*Instance) {
	if s.closed {
		return
	}
	marked := make(map[*Instance]bool)
	queue := make([]*Instance, 0, len(active))
	mark := func(in *Instance) {
		if in != nil && in.Owner == s && !in.disposed && !marked[in] {
			marked[in] = true
			queue = append(queue, in)
		}
	}
	markFunction := func(ref uint64) {
		if f, ok := s.Function(ref); ok {
			mark(f.Instance)
		}
	}
	visit := func(t wasm.ValueType, v value.Value) {
		if t.IsFunctionReference() {
			markFunction(v.Lo)
		}
	}
	for _, in := range s.instances {
		if in.published {
			mark(in)
		}
	}
	for _, in := range active {
		mark(in)
	}
	for _, ref := range functionRoots {
		markFunction(ref)
	}
	for _, table := range s.tables {
		if table.OwnerInstance == nil && table.Type.Element.IsFunctionReference() {
			for _, ref := range table.Values {
				markFunction(ref)
			}
		}
	}
	for _, global := range s.globals {
		if global.OwnerInstance == nil {
			visit(global.Type.Value, global.Value)
		}
	}
	for head := 0; head < len(queue); head++ {
		in := queue[head]
		for _, ref := range in.Functions {
			markFunction(ref)
		}
		for _, table := range in.Tables {
			mark(table.OwnerInstance)
		}
		for _, memory := range in.Memories {
			mark(memory.OwnerInstance)
		}
		for _, global := range in.Globals {
			mark(global.OwnerInstance)
		}
		in.RetainedValues(visit)
	}
	kept := s.instances[:0]
	for _, in := range s.instances {
		if marked[in] {
			kept = append(kept, in)
		} else {
			s.dispose(in)
		}
	}
	clear(s.instances[len(kept):])
	s.instances = kept
	// Never reuse an interior identity. Removing a suffix is safe because no
	// roots or executing frames reference any of the retired functions.
	for len(s.functions) > 1 && s.functions[len(s.functions)-1] == nil {
		s.functions = s.functions[:len(s.functions)-1]
	}
	s.sweepObjects()
}

func (s *Store) dispose(in *Instance) {
	for _, ref := range in.Functions {
		if f, ok := s.Function(ref); ok && f.Instance == in {
			s.functions[ref] = nil
			s.usage.Functions--
		}
	}
	for i := range in.Data {
		in.DropData(uint32(i))
	}
	for i := range in.Elements {
		in.DropElement(uint32(i))
	}
	s.usage.Instances--
	in.disposed = true
	in.Compiled = nil
	in.Types = nil
	in.Functions = nil
	in.Tables = nil
	in.Memories = nil
	in.Globals = nil
	in.Elements = nil
	in.Data = nil
	in.Exports = nil
}

func (s *Store) sweepObjects() {
	tables := s.tables[:0]
	for _, t := range s.tables {
		if s.closed || t.OwnerInstance != nil && t.OwnerInstance.disposed {
			s.usage.Tables--
			s.usage.TableElements -= uint64(len(t.Values))
			t.release()
			t.maximum = 0
			t.Owner = nil
			t.OwnerInstance = nil
		} else {
			tables = append(tables, t)
		}
	}
	clear(s.tables[len(tables):])
	s.tables = tables
	memories := s.memories[:0]
	for _, m := range s.memories {
		if s.closed || m.OwnerInstance != nil && m.OwnerInstance.disposed {
			s.usage.Memories--
			s.usage.MemoryPages -= uint64(m.Pages())
			m.Release()
			m.Owner = nil
			m.OwnerInstance = nil
		} else {
			memories = append(memories, m)
		}
	}
	clear(s.memories[len(memories):])
	s.memories = memories
	globals := s.globals[:0]
	for _, g := range s.globals {
		if s.closed || g.OwnerInstance != nil && g.OwnerInstance.disposed {
			s.usage.Globals--
			g.Value = value.Value{}
			g.Owner = nil
			g.OwnerInstance = nil
		} else {
			globals = append(globals, g)
		}
	}
	clear(s.globals[len(globals):])
	s.globals = globals
}

// RetainedValues visits host-object and surviving instance reference roots.
// The driver adds execution roots and explicit host roots before collection.
func (s *Store) RetainedValues(visit func(wasm.ValueType, value.Value)) {
	for _, in := range s.instances {
		in.RetainedValues(visit)
	}
	for _, t := range s.tables {
		if t.OwnerInstance == nil {
			for _, ref := range t.Values {
				visit(t.Type.Element, value.Value{Lo: ref})
			}
		}
	}
	for _, g := range s.globals {
		if g.OwnerInstance == nil && g.Type.Value.IsReference() {
			visit(g.Type.Value, g.Value)
		}
	}
}

// Close releases all backing storage and host callback objects. This internal
// operation requires the embedding driver to have drained execution leases;
// the public nonblocking close request is a separate synchronization layer.
func (s *Store) Close() {
	if s.closed {
		return
	}
	s.closed = true
	for _, in := range s.instances {
		s.dispose(in)
	}
	s.sweepObjects()
	clear(s.functions)
	s.functions = nil
	s.instances = nil
	s.tables = nil
	s.memories = nil
	s.globals = nil
	s.types = nil
	s.typeDefinitions = nil
	s.usage = Usage{}
}

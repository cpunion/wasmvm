package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
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
	roots := make([]ReferenceValue, len(functionRoots))
	for i, ref := range functionRoots {
		roots[i] = ReferenceValue{wasm.FuncRef, value.Value{Lo: ref}}
	}
	s.CollectReferences(roots, active, nil)
}

// CollectReferences traverses the joint object/instance graph. External roots
// discovered through internalized references are handed back to the host.
func (s *Store) CollectReferences(roots []ReferenceValue, active []*Instance, external func(uint64)) {
	if s.closed {
		return
	}
	marked := make(map[*Instance]bool)
	queue := make([]*Instance, 0, len(active))
	var objects []*heapObject
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
		if v.Lo == 0 {
			return
		}
		if t.IsFunctionReference() {
			markFunction(v.Lo)
			return
		}
		if v.Lo&heapTag != 0 {
			if o, ok := s.heapObject(v.Lo); ok && !o.Mark {
				o.Mark = true
				objects = append(objects, o)
			}
		} else if v.Lo&i31Tag != 0 {
		} else if v.Lo&externalTag != 0 {
			if external != nil {
				external(v.Lo &^ externalTag)
			}
		} else if t.IsExternalReference() && external != nil {
			external(v.Lo)
		}
	}
	for _, in := range s.instances {
		if in.published || len(s.allocating) != 0 {
			mark(in)
		}
	}
	for _, in := range active {
		mark(in)
	}
	for _, in := range s.allocating {
		mark(in)
	}
	for _, r := range roots {
		visit(r.Type, r.Value)
	}
	for _, r := range s.temporary {
		visit(r.Type, r.Value)
	}
	for _, t := range s.tables {
		if t.OwnerInstance == nil {
			for _, ref := range t.Values {
				visit(t.Type.Element, value.Value{Lo: ref})
			}
		}
	}
	for _, g := range s.globals {
		if g != nil && g.OwnerInstance == nil {
			visit(g.Type.Value, g.Value)
		}
	}
	ih, oh := 0, 0
	for ih < len(queue) || oh < len(objects) {
		for ih < len(queue) {
			in := queue[ih]
			ih++
			for _, ref := range in.Functions {
				markFunction(ref)
			}
			for _, t := range in.Tags {
				mark(t.OwnerInstance)
			}
			for _, t := range in.Tables {
				mark(t.OwnerInstance)
			}
			for _, m := range in.Memories {
				mark(m.OwnerInstance)
			}
			for _, g := range in.Globals {
				if g != nil {
					mark(g.OwnerInstance)
				}
			}
			in.RetainedValues(visit)
		}
		for oh < len(objects) {
			o := objects[oh]
			oh++

			if o.Tag != nil {
				mark(o.Tag.OwnerInstance)
				for i, t := range o.Tag.Signature.Params {
					if t.IsReference() {
						visit(t, o.get(uint32(i)))
					}
				}
				continue
			}
			d := s.registry.Types[o.Type].Definition
			if d.Kind == module.ArrayKind {
				if d.Fields[0].Value.IsReference() {
					for i := uint32(0); i < o.Length; i++ {
						visit(d.Fields[0].Value, o.get(i))
					}
				}
			} else {
				for i, f := range d.Fields {
					if f.Value.IsReference() {
						visit(f.Value, o.get(uint32(i)))
					}
				}
			}
		}
	}
	for i := 1; i < len(s.heap); i++ {
		o := s.heap[i].Object
		if o == nil {
			continue
		}
		if !o.Mark {
			s.releaseObject(uint32(i))
		} else {
			o.Mark = false
		}
	}
	s.heapThreshold = max(1<<20, s.usage.HeapBytes*2)

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
	in.Tags = nil
	in.Elements = nil
	in.Data = nil
	in.Exports = nil
}

func (s *Store) sweepObjects() {
	tags := s.tags[:0]
	for _, t := range s.tags {
		if s.closed || t.OwnerInstance != nil && t.OwnerInstance.disposed {
			s.usage.Tags--
			t.Owner = nil
			t.OwnerInstance = nil
			t.Signature = module.FuncType{}
			t.TypeID = 0
		} else {
			tags = append(tags, t)
		}
	}
	clear(s.tags[len(tags):])
	s.tags = tags
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
		if g != nil && g.OwnerInstance == nil && g.Type.Value.IsReference() {
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
	s.tags = nil
	s.registry = module.TypeRegistry{}
	for i := 1; i < len(s.heap); i++ {
		s.releaseObject(uint32(i))
	}
	s.heap = nil
	s.heapLayouts = nil
	s.heapFree = nil
	s.temporary = nil
	s.allocating = nil
	s.HeapSafepoint = nil
	s.usage = Usage{}
}

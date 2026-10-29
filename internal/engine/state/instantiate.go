package state

import (
	"fmt"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Resolver func(module, name string) (Extern, bool)

// Allocate links all imports and constructs instance state without applying
// active segments or calling start. Failure rolls back every new allocation
// and canonical type; imported objects are never modified in this phase.
func (s *Store) Allocate(compiled *slots.Module, resolve Resolver) (instance *Instance, err error) {
	if s.closed {
		return nil, wasm.ErrClosed
	}
	if compiled == nil || compiled.IR == nil || compiled.IR.Validated == nil {
		return nil, fmt.Errorf("%w: missing compiled module", wasm.ErrInvalid)
	}
	definition := compiled.IR.Validated.Definition()
	oldUsage, oldFunctions, oldTypes := s.usage, len(s.functions), len(s.types)
	oldTables, oldMemories, oldGlobals := len(s.tables), len(s.memories), len(s.globals)
	defer func() {
		if err != nil {
			s.usage = oldUsage
			clear(s.functions[oldFunctions:])
			s.functions = s.functions[:oldFunctions]
			for _, table := range s.tables[oldTables:] {
				table.release()
			}
			clear(s.tables[oldTables:])
			s.tables = s.tables[:oldTables]
			for _, m := range s.memories[oldMemories:] {
				m.Release()
			}
			clear(s.memories[oldMemories:])
			s.memories = s.memories[:oldMemories]
			clear(s.globals[oldGlobals:])
			s.globals = s.globals[:oldGlobals]
			for key, id := range s.types {
				if int(id) > oldTypes {
					delete(s.types, key)
				}
			}
		}
	}()
	instance = &Instance{Owner: s, Compiled: compiled, Exports: make(map[string]Extern)}
	for _, im := range definition.Imports {
		var ex Extern
		var found bool
		if resolve != nil {
			ex, found = resolve(im.Module, im.Name)
		}
		if !found || !s.matches(ex, im, definition) {
			return nil, fmt.Errorf("%w: %s.%s", ErrLink, im.Module, im.Name)
		}
		switch im.Kind {
		case 0:
			instance.Functions = append(instance.Functions, ex.Function)
		case 1:
			instance.Tables = append(instance.Tables, ex.Table)
		case 2:
			instance.Memories = append(instance.Memories, ex.Memory)
		case 3:
			instance.Globals = append(instance.Globals, ex.Global)
		}
	}
	if s.usage.Instances >= uint64(s.limits.MaxInstances) {
		return nil, limit("instances")
	}
	if uint64(len(definition.Functions)) > uint64(s.limits.MaxFunctions)-uint64(len(s.functions)-1) {
		return nil, limit("functions")
	}
	instance.Types = make([]uint32, len(definition.Types))
	for i, t := range definition.Types {
		id := s.TypeID(t)
		if id == 0 {
			return nil, limit("function types")
		}
		instance.Types[i] = id
	}
	for i, f := range definition.Functions {
		id := uint64(len(s.functions))
		index := uint32(len(instance.Functions))
		instance.Functions = append(instance.Functions, id)
		s.functions = append(s.functions, &Function{Owner: s, Signature: definition.Types[f.Type], TypeID: instance.Types[f.Type], Instance: instance, Code: compiled.Functions[i], Index: index})
		s.usage.Functions++
	}
	for _, t := range definition.Tables {
		ex, e := s.NewTable(t)
		if e != nil {
			return nil, e
		}
		ex.Table.OwnerInstance = instance
		instance.Tables = append(instance.Tables, ex.Table)
	}
	for _, m := range definition.Memories {
		ex, e := s.NewMemory(m)
		if e != nil {
			return nil, e
		}
		ex.Memory.OwnerInstance = instance
		instance.Memories = append(instance.Memories, ex.Memory)
	}
	for _, g := range definition.Globals {
		ex, e := s.NewGlobal(g.Type, constant(instance, g.Init))
		if e != nil {
			return nil, e
		}
		ex.Global.OwnerInstance = instance
		instance.Globals = append(instance.Globals, ex.Global)
	}
	instance.Elements = make([][]uint64, len(definition.Elements))
	for i, segment := range definition.Elements {
		n := uint64(len(segment.Functions)) + uint64(len(segment.Expressions))
		if n > s.limits.MaxSegmentElements-s.usage.SegmentElements {
			return nil, limit("element segments")
		}
		values := make([]uint64, 0, int(n))
		for _, f := range segment.Functions {
			values = append(values, instance.Functions[f])
		}
		for _, expr := range segment.Expressions {
			values = append(values, constant(instance, expr).Lo)
		}
		instance.Elements[i] = values
		s.usage.SegmentElements += n
	}
	instance.Data = make([][]byte, len(definition.Data))
	for i, segment := range definition.Data {
		n := uint64(len(segment.Bytes))
		if n > s.limits.MaxSegmentBytes-s.usage.SegmentBytes {
			return nil, limit("data segments")
		}
		instance.Data[i] = segment.Bytes
		s.usage.SegmentBytes += n
	}
	for _, ex := range definition.Exports {
		external := Extern{Owner: s, Kind: ex.Kind}
		switch ex.Kind {
		case 0:
			external.Function = instance.Functions[ex.Index]
		case 1:
			external.Table = instance.Tables[ex.Index]
		case 2:
			external.Memory = instance.Memories[ex.Index]
		case 3:
			external.Global = instance.Globals[ex.Index]
		}
		instance.Exports[ex.Name] = external
	}
	s.instances = append(s.instances, instance)
	s.usage.Instances++
	return instance, nil
}

// Initialize applies active segments in the specification's order: active
// elements, declarative drops, then active data. Earlier imported-state effects
// survive a later trap. The caller must execute start using its normal budget.
// After start succeeds the caller publishes the instance. On failure, it calls
// CollectFailed; references that escaped into imported state preserve owners.
func (in *Instance) Initialize() error {
	if in.disposed || in.Owner.closed {
		return wasm.ErrClosed
	}
	definition := in.Compiled.IR.Validated.Definition()
	for i, segment := range definition.Elements {
		if segment.Mode != module.Active {
			continue
		}
		offset := uint64(uint32(constant(in, segment.Offset).Lo))
		if !in.Tables[segment.Table].Init(offset, 0, uint64(len(in.Elements[i])), in.Elements[i]) {
			return &wasm.Trap{Code: wasm.TrapTableBounds}
		}
		in.DropElement(uint32(i))
	}
	for i, segment := range definition.Elements {
		if segment.Mode == module.Declarative {
			in.DropElement(uint32(i))
		}
	}
	for i, segment := range definition.Data {
		if segment.Mode != module.Active {
			continue
		}
		offset := uint64(uint32(constant(in, segment.Offset).Lo))
		if !in.Memories[segment.Memory].Init(offset, 0, uint64(len(in.Data[i])), in.Data[i]) {
			return &wasm.Trap{Code: wasm.TrapMemoryBounds}
		}
		in.DropData(uint32(i))
	}
	return nil
}

// RetainedValues visits typed reference roots in instance globals, tables and
// passive elements. Frame/host roots are supplied by their respective owners.
func (in *Instance) RetainedValues(visit func(wasm.ValueType, value.Value)) {
	for _, g := range in.Globals {
		if g.Type.Value.IsReference() {
			visit(g.Type.Value, g.Value)
		}
	}
	for _, t := range in.Tables {
		for _, ref := range t.Values {
			visit(t.Type.Element, value.Value{Lo: ref})
		}
	}
	for i, segment := range in.Elements {
		for _, ref := range segment {
			visit(in.Compiled.IR.Validated.Definition().Elements[i].Type, value.Value{Lo: ref})
		}
	}
}

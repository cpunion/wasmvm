package state

import (
	"context"
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
func (s *Store) Allocate(compiled *slots.Module, resolve Resolver) (*Instance, error) {
	return s.AllocateContext(context.Background(), compiled, resolve)
}
func (s *Store) AllocateContext(ctx context.Context, compiled *slots.Module, resolve Resolver) (instance *Instance, err error) {
	if s.closed {
		return nil, wasm.ErrClosed
	}
	if compiled == nil || compiled.IR == nil || compiled.IR.Validated == nil {
		return nil, fmt.Errorf("%w: missing compiled module", wasm.ErrInvalid)
	}
	definition := compiled.IR.Validated.Definition()
	var local [16]value.Value
	scratch := constantScratch(compiled, local[:])
	oldSerial, allocatingMark := s.heapSerial, len(s.allocating)
	oldUsage, oldFunctions, oldTypes := s.usage, len(s.functions), s.registry.Count()
	oldTables, oldMemories, oldGlobals := len(s.tables), len(s.memories), len(s.globals)
	defer func() {
		clear(s.allocating[allocatingMark:])
		s.allocating = s.allocating[:allocatingMark]
		if failure := recover(); failure != nil {
			switch e := failure.(type) {
			case *wasm.Error:
				instance, err = nil, e
			case constantFailure:
				instance, err = nil, e.err
			default:
				panic(failure)
			}
		}
		if err != nil {
			s.rollbackHeap(oldSerial)
			oldUsage.HeapBytes, oldUsage.HeapObjects = s.usage.HeapBytes, s.usage.HeapObjects
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
			s.registry.Rollback(oldTypes)
			for id := range s.heapLayouts {
				if int(id) > oldTypes {
					delete(s.heapLayouts, id)
				}
			}
		}
	}()
	instance = &Instance{initializerContext: ctx, Owner: s, Compiled: compiled, Exports: make(map[string]Extern)}
	s.allocating = append(s.allocating, instance)
	var ok bool
	instance.Types, ok = s.registry.InternModule(definition, s.limits.MaxTypes)
	if !ok {
		return nil, limit("canonical types")
	}

	for _, im := range definition.Imports {
		var ex Extern
		var found bool
		if resolve != nil {
			ex, found = resolve(im.Module, im.Name)
		}
		if !found || !s.matches(ex, im, instance) {
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

	for i, f := range definition.Functions {
		id := uint64(len(s.functions))
		index := uint32(len(instance.Functions))
		instance.Functions = append(instance.Functions, id)
		s.functions = append(s.functions, &Function{Owner: s, Signature: s.registry.Types[instance.Types[f.Type]].Function, TypeID: instance.Types[f.Type], Instance: instance, Code: compiled.Functions[i], Index: index})
		s.usage.Functions++
	}
	for i, t := range definition.Tables {
		t.Element = canonicalType(t.Element, instance.Types)
		var initial []uint64
		if len(definition.TableInit) != 0 && definition.TableInit[i] != nil {
			initial = []uint64{constant(instance, definition.TableInit[i], scratch).Lo}
		}
		ex, e := s.NewTable(t, initial...)
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
	// The validated private globals have one instance lifetime. Allocate their
	// cells and pointer views together instead of one host allocation per cell.
	count := len(definition.Globals)
	if uint64(count) > uint64(s.limits.MaxGlobals)-s.usage.Globals {
		return nil, limit("globals")
	}
	globals := make([]Global, count)
	storeBase, instanceBase := len(s.globals), len(instance.Globals)
	s.globals = append(s.globals, make([]*Global, count)...)
	instance.Globals = append(instance.Globals, make([]*Global, count)...)

	for i := range definition.Globals {
		typ := definition.Globals[i].Type
		typ.Value = canonicalType(typ.Value, instance.Types)
		globals[i] = Global{OwnerInstance: instance, Owner: s, Type: typ}
		s.globals[storeBase+i], instance.Globals[instanceBase+i] = &globals[i], &globals[i]
	}
	for i := range definition.Globals {
		globals[i].Value = constant(instance, definition.Globals[i].Init, scratch)
	}

	s.usage.Globals += uint64(count)
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
			values = append(values, constant(instance, expr, scratch).Lo)
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
	instance.initializerContext = nil
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
	var local [16]value.Value
	scratch := constantScratch(in.Compiled, local[:])
	for i, segment := range definition.Elements {
		if segment.Mode != module.Active {
			continue
		}
		offset := in.Tables[segment.Table].Type.Limits.Address(constant(in, segment.Offset, scratch).Lo)
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
		offset := in.Memories[segment.Memory].Limits().Address(constant(in, segment.Offset, scratch).Lo)
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
		if g != nil && g.Type.Value.IsReference() {
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
			visit(canonicalType(in.Compiled.IR.Validated.Definition().Elements[i].Type, in.Types), value.Value{Lo: ref})
		}
	}
}

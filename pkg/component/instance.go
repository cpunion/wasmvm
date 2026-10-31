package component

import (
	"context"
	"fmt"
	"sync"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type HostFunction func(*Caller, []Value) (*Value, error)

// ValueFactory creates one imported value after all static bindings and provider
// checks succeed, before guest initialization. Its Caller expires on return.
type ValueFactory func(*Caller) (Value, error)

type Imports struct {
	Functions      map[string]HostFunction
	FunctionChecks map[string]FunctionCheck
	Instances      map[string]Imports
	Values         map[string]Value
	ValueFactories map[string]ValueFactory
	ValueChecks    map[string]func(ValueType) error
	Resources      map[string]*HostResourceType
	CoreModules    map[string]*CoreModule
	Components     map[string]*Compiled
}

type Instance struct {
	mu                sync.Mutex
	closed            bool
	store             *vm.Store
	config            Config
	exports           map[string]item
	layouts           [2]*canonical.Builder
	initializing      *canonical.Codec
	instances, depth  uint32
	callDepth         uint32
	work              uint64
	bindings          map[any]Value
	resourceTables    []*canonical.Resources
	resourceBudget    *canonical.ResourceBudget
	hostResources     *canonical.Resources
	hostTypes         map[*hostResourceDefinition]*resourceType
	importResourceIDs map[uint64]*resourceType
	functionChecks    []pendingFunctionCheck
	initialValues     []*boundValue
	initialCreated    []*Resource
	engine            *Engine
	coreBindings      *validate.CoreBindings
	hostModules       map[*CoreModule]*coreModuleLease
	components        map[*Compiled]*componentLease
	componentBindings *validate.ComponentBindings
}

func (in *Instance) Close() {
	in.mu.Lock()
	in.closed = true
	modules := in.hostModules
	in.hostModules = nil
	components := in.components
	in.components = nil
	in.mu.Unlock()
	in.store.Close()
	for _, table := range in.resourceTables {
		table.Close()
	}
	for _, lease := range modules {
		lease.Close()
	}
	for _, lease := range components {
		lease.Close()
	}
}

func (in *Instance) WaitClosed(ctx context.Context) error { return in.store.WaitClosed(ctx) }

func (in *Instance) Function(path ...string) (*Function, error) {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.closed {
		return nil, wasm.ErrClosed
	}
	exports := in.exports
	for i, name := range path {
		e, ok, _ := lookupName(exports, name, nil)
		if !ok {
			return nil, vm.ErrLink
		}
		if i == len(path)-1 && e.function != nil {
			return e.function, nil
		}
		if e.instance == nil {
			return nil, vm.ErrType
		}
		exports = e.instance
	}
	return nil, vm.ErrLink
}

func (in *Instance) Value(ctx context.Context, path ...string) (Value, error) {
	var result Value
	var typ *model.Type
	_, err := in.store.WithCallerCleanup(ctx, vm.CallOptions{}, func(c *vm.Caller) error {
		exports := in.exports
		for i, name := range path {
			e, ok, _ := lookupName(exports, name, nil)
			if !ok {
				return vm.ErrLink
			}
			if i == len(path)-1 && e.sort == model.ValueSort {
				v := e.value
				if v.consumed {
					return wasm.ErrTrap
				}
				codec := in.codec(c, v.scope, nil)
				copy, err := codec.Clone(v.typ, v.value)
				if err == nil && v.resource {
					err = codec.AcceptResult(v.typ, &copy)
				}
				if err = codec.Finish(err); err != nil {
					return err
				}
				v.consumed = v.resource
				result = copy
				typ = v.typ
				return nil
			}
			if e.instance == nil {
				return vm.ErrType
			}
			exports = e.instance
		}
		return vm.ErrLink
	}, func(c *vm.Caller) error {
		var failed error
		canonical.VisitReferences(typ, &result, func(ref *Resource) {
			if err := in.cleanupResource(c, ref); failed == nil {
				failed = err
			}
		})
		return failed
	})
	if err != nil {
		return Value{}, err
	}
	return result, err
}

// Aliases and projected instance views share a value occurrence. Its original
// type and scope retain the nominal identities used when it was produced.
type boundValue struct {
	typ       *model.Type
	expected  *validate.Type
	scope     *namespace
	value     Value
	resource  bool
	consumed  bool
	validated bool
	factory   ValueFactory
	check     func(ValueType) error
}

type item struct {
	sort      model.Sort
	core      vm.Extern
	module    *vm.CompiledModule
	function  *Function
	instance  map[string]item
	component *closure
	value     *boundValue
	resource  *resourceType
}

type closure struct {
	plan   *validate.Component
	parent *namespace
	cores  map[*module.Module]*vm.CompiledModule
}

type namespace struct {
	parent    *namespace
	plan      *validate.Component
	resources map[uint64]*resourceType
	host      *namespace
	handles   *canonical.Resources
	items     map[model.Sort][]item
	exports   map[string]item
	failed    bool
	blocked   bool
	entered   bool
}

func (n *namespace) at(index model.SortIndex) item {
	return n.items[index.Sort][index.Index]
}

func (n *namespace) add(e item) {
	if e.sort == model.TypeSort && e.resource != nil {
		t := n.plan.TypeAt(uint32(len(n.items[model.TypeSort])))
		n.resources[t.Resource] = e.resource
	}
	n.items[e.sort] = append(n.items[e.sort], e)
}

func (c *Compiled) Instantiate(ctx context.Context, imports Imports, options vm.CallOptions) (*Instance, error) {
	if c == nil {
		return nil, vm.ErrReference
	}
	ctx = contextOrBackground(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lease, err := c.acquire(c.engine)
	if err != nil {
		return nil, err
	}
	plan, cores := lease.plan, lease.cores
	store, err := c.engine.core.NewStore(c.engine.config.Store)
	if err != nil {
		lease.Close()
		return nil, err
	}
	in := &Instance{engine: c.engine, store: store, config: c.engine.config, bindings: map[any]Value{}, resourceBudget: canonical.NewResourceBudget(c.engine.config.MaxResources)}
	in.components = map[*Compiled]*componentLease{c: lease}
	in.hostTypes = map[*hostResourceDefinition]*resourceType{}
	in.importResourceIDs = map[uint64]*resourceType{}
	in.hostResources = canonical.NewResourceTable(in.resourceBudget)
	in.resourceTables = append(in.resourceTables, in.hostResources)
	for i := range in.layouts {
		in.layouts[i] = canonical.NewBuilder(i == 1, canonical.Limits{MaxNodes: in.config.Validate.MaxNodes, MaxDepth: in.config.MaxNesting, MaxBytes: in.config.MaxBytes})
	}
	in.initializing = &canonical.Codec{Transfer: &canonical.Transfer{Context: contextOrBackground(ctx), MaxBytes: in.config.MaxBytes}, Layouts: in.layouts[0], MaxValues: in.config.MaxValues}
	published := false
	defer func() {
		for _, v := range in.initialValues {
			v.expected, v.factory, v.check = nil, nil, nil
		}
		in.initializing = nil
		in.bindings = nil
		in.importResourceIDs = nil
		in.functionChecks = nil
		in.initialValues = nil
		in.initialCreated = nil
		in.coreBindings = nil
		in.componentBindings = nil
		if !published {
			in.Close()
		}
	}()
	bindings, err := in.bind(plan.Type.Imports, imports)
	if err != nil {
		return nil, err
	}
	for _, c := range in.functionChecks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := c.check(FunctionInfo{typ: c.typ, resources: in.importResourceIDs}); err != nil {
			return nil, fmt.Errorf("component function %q: %w", c.name, err)
		}
	}
	for _, v := range in.initialValues {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if v.check != nil {
			if err := v.check(ValueType{v.typ, in.importResourceIDs}); err != nil {
				return nil, err
			}
		}
	}
	_, err = store.WithCallerCleanup(ctx, options, func(caller *vm.Caller) (err error) {
		importsScope := &namespace{resources: in.importResourceIDs}
		in.initializing.Resources = &resourceCodec{in: in, n: importsScope, core: caller}
		complete := false
		defer func() { err = finishResources(in.initializing, complete, err) }()
		if err := in.initializeValues(caller, importsScope); err != nil {
			return err
		}
		n, err := in.instantiate(caller, &closure{plan: plan, cores: cores}, bindings, cores)
		if err == nil {
			in.exports = n.exports
		}
		complete = err == nil
		return err
	}, in.cleanupInitialValues)
	if err != nil {
		return nil, err
	}
	published = true
	return in, nil
}

func (in *Instance) bind(types []validate.Extern, imports Imports) (map[string]item, error) {
	bindings := map[string]item{}
	for _, expected := range types {
		if err := in.charge(1); err != nil {
			return nil, err
		}
		e := item{sort: expected.Sort}
		name := expected.Name.Text
		switch expected.Sort {
		case model.ComponentSort:
			host := imports.Components[name]
			if host == nil {
				return nil, fmt.Errorf("%w: missing component %q", vm.ErrLink, name)
			}
			lease := in.components[host]
			if lease == nil {
				var err error
				lease, err = host.acquire(in.engine)
				if err != nil {
					return nil, err
				}
				in.components[host] = lease
			}
			if in.componentBindings == nil {
				var err error
				in.componentBindings, err = expected.Type.ComponentBindings()
				if err != nil {
					return nil, err
				}
			}
			ok, err := in.componentBindings.Check(expected.Type, lease.plan.Type)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%w: component %q", vm.ErrType, name)
			}
			e.component = &closure{plan: lease.plan, cores: lease.cores}
		case model.CoreModuleSort:
			host := imports.CoreModules[name]
			if host == nil {
				return nil, fmt.Errorf("%w: missing Core module %q", vm.ErrLink, name)
			}
			lease := in.hostModules[host]
			if lease == nil {
				var err error
				lease, err = host.acquire(in.engine)
				if err != nil {
					return nil, err
				}
				if in.hostModules == nil {
					in.hostModules = map[*CoreModule]*coreModuleLease{}
				}
				in.hostModules[host] = lease
			}
			if in.coreBindings == nil {
				var err error
				in.coreBindings, err = expected.Type.CoreBindings()
				if err != nil {
					return nil, err
				}
			}
			ok, err := in.coreBindings.Check(expected.Type, lease.schema)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%w: Core module %q", vm.ErrType, name)
			}
			e.module = lease.compiled
		case model.TypeSort:
			if expected.Type.Kind == model.Resource {
				host := imports.Resources[name]
				if host == nil || host.definition == nil {
					return nil, vm.ErrLink
				}
				rt := in.hostTypes[host.definition]
				if rt == nil {
					rt = &resourceType{instance: in, identity: host.definition.identity, representation: wasm.I32, host: host.definition}
					in.hostTypes[host.definition] = rt
				}
				if prior := in.importResourceIDs[expected.Type.Resource]; prior != nil && prior != rt {
					return nil, vm.ErrType
				}
				in.importResourceIDs[expected.Type.Resource] = rt
				e.resource = rt
			}

		case model.FunctionSort:
			fn := imports.Functions[name]
			if fn == nil {
				return nil, fmt.Errorf("%w: missing component function %q", vm.ErrLink, name)
			}
			if check := imports.FunctionChecks[name]; check != nil {
				in.functionChecks = append(in.functionChecks, pendingFunctionCheck{name: name, typ: expected.Type, check: check})
			}
			e.function = &Function{instance: in, typ: expected.Type, host: fn}
		case model.InstanceSort:
			args, ok, err := lookupName(imports.Instances, name, in.charge)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%w: missing component instance %q", vm.ErrLink, name)
			}
			e.instance, err = in.bind(expected.Type.Exports, args)
			if err != nil {
				return nil, err
			}
		case model.ValueSort:
			v, ok := imports.Values[name]
			factory := imports.ValueFactories[name]
			if ok && factory != nil {
				return nil, fmt.Errorf("%w: ambiguous component value %q", vm.ErrLink, name)
			}
			if !ok && factory == nil {
				return nil, fmt.Errorf("%w: missing component value %q", vm.ErrLink, name)
			}
			e.value = &boundValue{typ: expected.Type.Value, expected: expected.Type, resource: expected.Type.ContainsResource(), factory: factory, check: imports.ValueChecks[name]}
			in.initialValues = append(in.initialValues, e.value)
			if factory != nil || e.value.resource {
				// Resource authority can only be checked after all nominal imports
				// bind. Factories run later, inside the shared Core session.
				e.value.value = v
				break
			}
			var err error
			e.value.value, err = in.initializing.Clone(expected.Type.Value, v)
			if err != nil {
				return nil, err
			}
			if literal, ok := expected.Type.LiteralValue(); ok {
				literal, err = in.initializing.Clone(expected.Type.Value, literal)
				if err != nil {
					return nil, err
				}
				if !equalValue(e.value.value, literal) {
					return nil, vm.ErrType
				}
			}
			key := expected.Type.ValueKey()
			if earlier, ok := in.bindings[key]; ok && !equalValue(earlier, e.value.value) {
				return nil, vm.ErrType
			}
			in.bindings[key] = e.value.value
		default:
			return nil, unsupported("host import sort")
		}
		bindings[name] = e
	}
	return bindings, nil
}

func equalValue(a, b Value) bool {
	if a.Bits != b.Bits || a.Case != b.Case || a.Text != b.Text || !equalResource(a.Resource, b.Resource) || len(a.Elements) != len(b.Elements) {
		return false
	}
	for i := range a.Elements {
		if !equalValue(a.Elements[i], b.Elements[i]) {
			return false
		}
	}
	return true
}

func equalResource(a, b *Resource) bool {
	if a == b {
		return true
	}
	at, ai, ax, _, _, ae := canonical.ReferenceInfo(a)
	bt, bi, bx, _, _, be := canonical.ReferenceInfo(b)
	return ae == nil && be == nil && at == bt && ai == bi && ax == bx
}

func (in *Instance) charge(n uint64) error {
	if n > uint64(in.config.Validate.MaxNodes)-in.work {
		return wasm.ErrLimit
	}
	in.work += n
	return nil
}

func (in *Instance) instantiate(caller *vm.Caller, component *closure, args map[string]item, cores map[*module.Module]*vm.CompiledModule) (*namespace, error) {
	if component.cores != nil {
		cores = component.cores
	}
	if in.instances >= in.config.Store.Limits.MaxInstances || in.depth >= in.config.MaxNesting {
		return nil, wasm.ErrLimit
	}
	if err := in.charge(uint64(len(component.plan.Definition().Definitions))); err != nil {
		return nil, err
	}
	in.instances++
	in.depth++
	defer func() { in.depth-- }()
	n := &namespace{parent: component.parent, plan: component.plan, resources: map[uint64]*resourceType{}, items: map[model.Sort][]item{}, exports: map[string]item{}}
	// The index-space plan is needed only while definitions are published.
	defer func() { n.plan = nil }()
	children, canonicalIndex, valuesIndex, instancesIndex := 0, 0, 0, 0
	importTypes := map[string]*validate.Type{}
	for _, e := range component.plan.Type.Imports {
		importTypes[e.Name.Text] = e.Type
	}
	exportTypes := map[string]*validate.Type{}
	for _, e := range component.plan.Type.Exports {
		exportTypes[e.Name.Text] = e.Type
	}
	for _, d := range component.plan.Definition().Definitions {
		if err := caller.Context().Err(); err != nil {
			return nil, err
		}
		switch d.Kind {
		case 1:
			n.add(item{sort: model.CoreModuleSort, module: cores[d.CoreModule]})
		case 2:
			exports := map[string]item{}
			if d.Instance.Instantiate {
				m := n.at(model.SortIndex{Sort: model.CoreModuleSort, Index: d.Instance.Index}).module
				l := vm.NewLinker(in.store)
				for _, a := range d.Instance.Arguments {
					objects := map[string]vm.Extern{}
					for name, e := range n.at(a.Item).instance {
						objects[name] = e.core
					}
					if err := l.DefineModule(a.Name, objects); err != nil {
						return nil, err
					}
				}
				core, err := caller.Instantiate(l, m)
				if err != nil {
					return nil, err
				}
				info, err := m.Info()
				if err != nil {
					return nil, err
				}
				for _, e := range info.Exports {
					ex, err := caller.Export(core, e.Name)
					if err != nil {
						return nil, err
					}
					exports[e.Name] = item{sort: model.CoreFunctionSort + model.Sort(e.Kind), core: ex}
				}
			} else {
				for _, e := range d.Instance.Exports {
					exports[e.Name.Text] = n.at(e.Item)
				}
			}
			n.add(item{sort: model.CoreInstanceSort, instance: exports})
		case 3:
			// Core type metadata is already checked by the shared Core validator.
		case 7:
			e := item{sort: model.TypeSort}
			if d.Type.Kind == model.Resource {
				e.resource = &resourceType{instance: in, identity: &canonical.ResourceType{}, owner: n, representation: d.Type.Representation}
				if d.Type.Destructor != nil {
					destructor := n.items[model.CoreFunctionSort][*d.Type.Destructor].core
					e.resource.destructor = &destructor
				}
			}
			n.add(e)
		case 4:
			child := component.plan.Children[children]
			children++
			n.add(item{sort: model.ComponentSort, component: &closure{plan: child, parent: n, cores: cores}})
		case 5:
			exports := map[string]item{}
			if d.Instance.Instantiate {
				child := n.at(model.SortIndex{Sort: model.ComponentSort, Index: d.Instance.Index}).component
				bindings := map[string]item{}
				for _, a := range d.Instance.Arguments {
					bindings[a.Name] = n.at(a.Item)
				}
				result, err := in.instantiate(caller, child, bindings, cores)
				if err != nil {
					return nil, err
				}
				exports = result.exports
			} else {
				for _, e := range d.Instance.Exports {
					exports[e.Name.Text] = n.at(e.Item)
				}
			}
			view, err := in.project(item{sort: model.InstanceSort, instance: exports}, component.plan.Instances[instancesIndex].Type)
			instancesIndex++
			if err != nil {
				return nil, err
			}
			if err := in.bindResources(n, view, component.plan.Instances[instancesIndex-1].Type); err != nil {
				return nil, err
			}
			n.add(view)
		case 6:
			a := d.Alias
			if a.Kind == 2 {
				at := n
				for range a.Count {
					at = at.parent
				}
				if a.Sort == model.CoreModuleSort || a.Sort == model.ComponentSort || a.Sort == model.TypeSort {
					n.add(at.at(model.SortIndex{Sort: a.Sort, Index: a.Index}))
				}
			} else {
				sort := model.InstanceSort
				if a.Kind == 1 {
					sort = model.CoreInstanceSort
				}
				if a.Sort != model.CoreTypeSort {
					n.add(n.items[sort][a.Instance].instance[a.Name])
				}
			}
		case 8:
			plan := component.plan.Canonical[canonicalIndex]
			canonicalIndex++
			if err := in.canonical(caller, n, plan); err != nil {
				return nil, err
			}
		case 9:
			f := n.at(model.SortIndex{Sort: model.FunctionSort, Index: d.Start.Function}).function
			arguments := make([]Value, len(d.Start.Arguments))
			for i, index := range d.Start.Arguments {
				arguments[i] = n.at(model.SortIndex{Sort: model.ValueSort, Index: index}).value.value
			}
			v, err := f.call(caller, arguments)
			if err != nil {
				return nil, err
			}
			if v != nil {
				bound := &boundValue{typ: f.typ.Result, scope: f.scope, value: *v, resource: f.typ.ResultContainsResource(), validated: true}
				// Track the returned capability before a bounded aggregate copy can
				// fail; unpublished ownership must still be reclaimed.
				in.initialValues = append(in.initialValues, bound)
				previous := in.initializing.Resources
				in.initializing.Resources = in.codec(caller, f.scope, nil).Resources
				copy, err := in.initializing.Clone(f.typ.Result, *v)
				in.initializing.Resources = previous
				if err != nil {
					return nil, err
				}
				bound.value = copy
				n.add(item{sort: model.ValueSort, value: bound})
			}
		case 10:
			e := args[d.Import.Name.Text]
			if err := in.bindResources(n, e, importTypes[d.Import.Name.Text]); err != nil {
				return nil, err
			}
			n.add(e)
		case 11:
			e := n.at(d.Export.Item)
			e, err := in.project(e, exportTypes[d.Export.Name.Text])
			if err != nil {
				return nil, err
			}
			n.exports[d.Export.Name.Text] = e
			n.add(e)
		case 12:
			v := component.plan.Values[valuesIndex]
			valuesIndex++
			copy, err := in.initializing.Clone(v.Type.Value, v.Value)
			if err != nil {
				return nil, err
			}
			n.add(item{sort: model.ValueSort, value: &boundValue{typ: v.Type.Value, value: copy}})
		}
	}
	return n, nil
}

func (in *Instance) project(e item, typ *validate.Type) (item, error) {
	if err := in.charge(1); err != nil {
		return item{}, err
	}
	if e.sort != model.InstanceSort {
		return e, nil
	}
	exports := map[string]item{}
	for _, expected := range typ.Exports {
		actual, ok := e.instance[expected.Name.Text]
		if !ok || actual.sort != expected.Sort {
			return item{}, vm.ErrType
		}
		view, err := in.project(actual, expected.Type)
		if err != nil {
			return item{}, err
		}
		exports[expected.Name.Text] = view
	}
	e.instance = exports
	return e, nil
}

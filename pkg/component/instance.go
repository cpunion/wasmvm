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

type Imports struct {
	Functions      map[string]HostFunction
	FunctionChecks map[string]FunctionCheck
	Instances      map[string]Imports
	Values         map[string]Value
	Resources      map[string]*HostResourceType
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
}

func (in *Instance) Close() {
	in.mu.Lock()
	in.closed = true
	in.mu.Unlock()
	in.store.Close()
	for _, table := range in.resourceTables {
		table.Close()
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
		e, ok := exports[name]
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
	_, err := in.store.WithCaller(ctx, vm.CallOptions{}, func(c *vm.Caller) error {
		exports := in.exports
		for i, name := range path {
			e, ok := exports[name]
			if !ok {
				return vm.ErrLink
			}
			if i == len(path)-1 && e.sort == model.ValueSort {
				var err error
				result, err = in.codec(c, nil, nil).Clone(e.valueType, e.value)
				return err
			}
			if e.instance == nil {
				return vm.ErrType
			}
			exports = e.instance
		}
		return vm.ErrLink
	})
	return result, err
}

type item struct {
	sort      model.Sort
	core      vm.Extern
	module    *vm.CompiledModule
	function  *Function
	instance  map[string]item
	component *closure
	value     Value
	valueType *model.Type
	resource  *resourceType
}

type closure struct {
	plan   *validate.Component
	parent *namespace
}

type namespace struct {
	parent    *namespace
	plan      *validate.Component
	resources map[uint64]*resourceType
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
	c.mu.Lock()
	plan, cores := c.plan, c.cores
	c.mu.Unlock()
	if plan == nil {
		return nil, wasm.ErrClosed
	}
	ctx = contextOrBackground(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := c.engine.core.NewStore(c.engine.config.Store)
	if err != nil {
		return nil, err
	}
	in := &Instance{store: store, config: c.engine.config, bindings: map[any]Value{}, resourceBudget: canonical.NewResourceBudget(c.engine.config.MaxResources)}
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
		in.initializing = nil
		in.bindings = nil
		in.importResourceIDs = nil
		in.functionChecks = nil
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
	_, err = store.WithCaller(ctx, options, func(caller *vm.Caller) error {
		n, err := in.instantiate(caller, &closure{plan: plan}, bindings, cores)
		if err == nil {
			in.exports = n.exports
		}
		return err
	})
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
			} else if expected.Type.Value == nil {
				return nil, unsupported("host nominal type binding")
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
			args, ok := imports.Instances[name]
			if !ok {
				return nil, fmt.Errorf("%w: missing component instance %q", vm.ErrLink, name)
			}
			var err error
			e.instance, err = in.bind(expected.Type.Exports, args)
			if err != nil {
				return nil, err
			}
		case model.ValueSort:
			v, ok := imports.Values[name]
			if !ok {
				return nil, fmt.Errorf("%w: missing component value %q", vm.ErrLink, name)
			}
			var err error
			e.value, err = in.initializing.Clone(expected.Type.Value, v)
			if err != nil {
				return nil, err
			}
			e.valueType = expected.Type.Value
			if literal, ok := expected.Type.LiteralValue(); ok {
				literal, err = in.initializing.Clone(expected.Type.Value, literal)
				if err != nil {
					return nil, err
				}
				if !equalValue(e.value, literal) {
					return nil, vm.ErrType
				}
			}
			key := expected.Type.ValueKey()
			if earlier, ok := in.bindings[key]; ok && !equalValue(earlier, e.value) {
				return nil, vm.ErrType
			}
			in.bindings[key] = e.value
		default:
			return nil, unsupported("host import sort")
		}
		bindings[name] = e
	}
	return bindings, nil
}

func equalValue(a, b Value) bool {
	if a.Bits != b.Bits || a.Case != b.Case || a.Text != b.Text || len(a.Elements) != len(b.Elements) {
		return false
	}
	for i := range a.Elements {
		if !equalValue(a.Elements[i], b.Elements[i]) {
			return false
		}
	}
	return true
}

func (in *Instance) charge(n uint64) error {
	if n > uint64(in.config.Validate.MaxNodes)-in.work {
		return wasm.ErrLimit
	}
	in.work += n
	return nil
}

func (in *Instance) instantiate(caller *vm.Caller, component *closure, args map[string]item, cores map[*module.Module]*vm.CompiledModule) (*namespace, error) {
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
			n.add(item{sort: model.ComponentSort, component: &closure{plan: child, parent: n}})
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
				arguments[i] = n.at(model.SortIndex{Sort: model.ValueSort, Index: index}).value
			}
			v, err := f.call(caller, arguments)
			if err != nil {
				return nil, err
			}
			if v != nil {
				copy, err := in.initializing.Clone(f.typ.Result, *v)
				if err != nil {
					return nil, err
				}
				n.add(item{sort: model.ValueSort, value: copy, valueType: f.typ.Result})
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
			n.add(item{sort: model.ValueSort, value: copy, valueType: v.Type.Value})
		}
	}
	return n, nil
}

func (in *Instance) project(e item, typ *validate.Type) (item, error) {
	if err := in.charge(1); err != nil {
		return item{}, err
	}
	if e.sort != model.InstanceSort {
		if e.sort == model.ValueSort {
			e.valueType = typ.Value
		}
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

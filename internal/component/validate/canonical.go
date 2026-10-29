package validate

import (
	"fmt"
	"math"
	"slices"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type CanonicalFunction struct {
	Definition                  *model.Canonical
	Type                        *Type
	ABI                         canonical.FunctionLayout
	Encoding                    canonical.Encoding
	Memory, Realloc, PostReturn *uint32
	Memory64                    bool
	Signature                   module.FuncType
	Async                       bool
}

func (c *componentContext) canonical(def *model.Canonical) error {
	if def == nil {
		return invalid("missing canonical definition")
	}
	s, p := c.result.scope, c.result.scope.pool
	if err := p.charge(1 + uint64(len(def.Options))); err != nil {
		return err
	}
	if def.Op >= 2 && def.Op <= 4 {
		return c.resourceBuiltin(def)
	}
	if def.Op == 0x0e || def.Op == 0x15 || def.Op == 0x20 || def.Op == 0x21 || def.Op == 0x27 {
		return c.asyncBuiltin(def)
	}
	if def.Op > 1 {
		return fmt.Errorf("%w: canonical operation %#x", wasm.ErrUnsupported, def.Op)
	}
	lifting := def.Op == 0
	var t *Type
	var err error
	if lifting {
		t, err = s.Reference(model.ExternType{Sort: model.FunctionSort, Index: def.Extra})
	} else {
		if uint64(def.Index) >= uint64(len(s.items[model.FunctionSort])) {
			return invalid("canonical lower function index")
		}
		t = s.items[model.FunctionSort][def.Index]
	}
	if err != nil {
		return err
	}
	if t.Kind == model.AsyncFunction {
		return fmt.Errorf("%w: asynchronous canonical signature", wasm.ErrUnsupported)
	}
	info := &CanonicalFunction{Definition: def, Type: t}
	seen := byte(0)
	for _, o := range def.Options {
		kind := o.Kind
		if kind <= 2 {
			kind = 0
		}
		if kind > 7 {
			return invalid("unknown canonical option")
		}
		if seen&(1<<kind) != 0 {
			return invalid("duplicate canonical option")
		}
		seen |= 1 << kind
		index := o.Index
		switch o.Kind {
		case 0, 1, 2:
			info.Encoding = canonical.Encoding(o.Kind)
		case 3:
			memory, err := c.coreAt(model.SortIndex{Sort: model.CoreMemorySort, Index: index})
			if err != nil {
				return err
			}
			if memory.Descriptor.Memory.Shared {
				return invalid("canonical memory is shared")
			}
			info.Memory, info.Memory64 = &index, memory.Descriptor.Memory.Index64
		case 4, 5:
			if _, err := c.coreAt(model.SortIndex{Sort: model.CoreFunctionSort, Index: index}); err != nil {
				return err
			}
			if o.Kind == 4 {
				info.Realloc = &index
			} else {
				info.PostReturn = &index
			}
		case 6, 7:
			return fmt.Errorf("%w: asynchronous canonical option", wasm.ErrUnsupported)
		}
	}
	if info.Realloc != nil && info.Memory == nil {
		return invalid("canonical realloc requires memory")
	}
	if !lifting && info.PostReturn != nil {
		return invalid("post-return is only valid for lift")
	}
	address := wasm.I32
	b := p.layouts32
	if info.Memory64 {
		address, b = wasm.I64, p.layouts
	} else if b == nil {
		b = canonical.NewBuilder(false, canonical.Limits{MaxNodes: p.limits.MaxNodes, MaxDepth: p.limits.MaxDepth, MaxBytes: math.MaxUint64})
		p.layouts32 = b
	}
	if info.Realloc != nil {
		f := c.functions[*info.Realloc]
		if !slices.Equal(f.Params, []wasm.ValueType{address, address, address, address}) || !slices.Equal(f.Results, []wasm.ValueType{address}) {
			return invalid("canonical realloc signature")
		}
	}
	if err := p.charge(uint64(len(t.Parameters))); err != nil {
		return err
	}
	params := make([]*model.Type, len(t.Parameters))
	parameterPointers := false
	for i, f := range t.Parameters {
		params[i] = f.Type
		needed, err := c.hasPointers(f.Type)
		if err != nil {
			return err
		}
		parameterPointers = parameterPointers || needed
	}
	resultPointers, err := c.hasPointers(t.Result)
	if err != nil {
		return err
	}
	info.ABI, err = b.Function(params, t.Result, lifting)
	if err != nil {
		return err
	}
	if (parameterPointers || resultPointers || info.ABI.ParameterMemory != nil || info.ABI.ResultMemory != nil) && info.Memory == nil {
		return invalid("canonical memory is required")
	}
	if (lifting && (parameterPointers || info.ABI.ParameterMemory != nil) || !lifting && resultPointers) && info.Realloc == nil {
		return invalid("canonical realloc is required")
	}
	info.Signature = module.FuncType{Params: info.ABI.Params, Results: info.ABI.Results}
	if info.PostReturn != nil {
		post := c.functions[*info.PostReturn]
		if !slices.Equal(post.Params, info.Signature.Results) || len(post.Results) != 0 {
			return invalid("canonical post-return signature")
		}
	}
	if lifting {
		if uint64(def.Index) >= uint64(len(c.functions)) {
			return invalid("canonical lift Core function index")
		}
		actual := c.functions[def.Index]
		if !slices.Equal(actual.Params, info.Signature.Params) || !slices.Equal(actual.Results, info.Signature.Results) {
			return invalid("canonical lift Core function signature")
		}
		if err := p.charge(1); err != nil {
			return err
		}
		s.items[model.FunctionSort] = append(s.items[model.FunctionSort], t)
	} else {
		if err := c.addSignature(info.Signature); err != nil {
			return err
		}
	}
	c.result.Canonical = append(c.result.Canonical, info)
	return nil
}

// These operations have static plans only. Their task/handle/thread execution
// remains an async-runtime capability, independent of successful validation.
func (c *componentContext) asyncBuiltin(def *model.Canonical) error {
	info := &CanonicalFunction{Definition: def, Async: true}
	f := module.FuncType{}
	s := c.result.scope
	switch def.Op {
	case 0x0e, 0x15:
		t, err := s.At(def.Index)
		if err != nil {
			return err
		}
		kind := model.Stream
		if def.Op == 0x15 {
			kind = model.Future
		}
		if t.Kind != kind {
			return invalid("stream/future constructor type")
		}
		info.Type = t
		f.Results = []wasm.ValueType{wasm.I64}
	case 0x20, 0x21:
		e, err := c.coreAt(model.SortIndex{Sort: model.CoreMemorySort, Index: def.Index})
		if err != nil {
			return err
		}
		if e.Descriptor.Memory.Shared {
			return invalid("async canonical memory is shared")
		}
		index := def.Index
		info.Memory = &index
		info.Memory64 = e.Descriptor.Memory.Index64
		f.Params = []wasm.ValueType{wasm.I32, e.Descriptor.Memory.AddressType()}
		f.Results = []wasm.ValueType{wasm.I32}
	case 0x27:
		if uint64(def.Index) >= uint64(len(s.core.ids)) {
			return invalid("thread closure type index")
		}
		id := s.core.ids[def.Index]
		if id == 0 || s.pool.registry.Types[id].Definition.Kind != module.FunctionKind {
			return invalid("thread closure requires function type")
		}
		closure := s.pool.registry.Types[id].Function
		if len(closure.Params) != 1 || len(closure.Results) != 0 || (closure.Params[0] != wasm.I32 && closure.Params[0] != wasm.I64) {
			return invalid("thread closure signature")
		}
		e, err := c.coreAt(model.SortIndex{Sort: model.CoreTableSort, Index: def.Extra})
		if err != nil {
			return err
		}
		if e.Descriptor.Table.Element != wasm.FuncRef {
			return invalid("thread table element type")
		}
		f.Params = []wasm.ValueType{e.Descriptor.Table.Limits.AddressType(), closure.Params[0]}
		f.Results = []wasm.ValueType{wasm.I32}
	}
	info.Signature = f
	if err := c.addSignature(f); err != nil {
		return err
	}
	c.result.Canonical = append(c.result.Canonical, info)
	return nil
}

func (c *componentContext) hasPointers(t *model.Type) (bool, error) {
	if t == nil {
		return false, nil
	}
	if known, ok := c.pointers[t]; ok {
		return known, nil
	}
	if err := c.result.scope.pool.charge(1 + uint64(len(t.Fields)) + uint64(len(t.Cases))); err != nil {
		return false, err
	}
	needed := t.Kind == model.String || t.Kind == model.List || t.Kind == model.Map
	if !needed && t.Kind != model.Future && t.Kind != model.Stream {
		for _, fs := range [][]model.Field{t.Fields, t.Cases} {
			for _, f := range fs {
				child, err := c.hasPointers(f.Type)
				if err != nil {
					return false, err
				}
				needed = needed || child
			}
		}
		child, err := c.hasPointers(t.Element)
		if err != nil {
			return false, err
		}
		needed = needed || child
	}
	c.pointers[t] = needed
	return needed, nil
}

func (c *componentContext) addSignature(signature module.FuncType) error {
	p := c.result.scope.pool
	if err := p.charge(1 + uint64(len(signature.Params)) + uint64(len(signature.Results))); err != nil {
		return err
	}
	m := &module.Module{Types: []module.FuncType{signature}}
	ids := make([]uint32, 1)
	if !p.registry.InternGroup(m, 0, 1, ids, p.limits.MaxNodes) {
		return limit("canonical function types")
	}
	return c.addCore(CoreExtern{Descriptor: module.Import{Kind: 0, Type: ids[0]}})
}

func (c *componentContext) resourceBuiltin(def *model.Canonical) error {
	t, err := c.result.scope.At(def.Index)
	if err != nil {
		return err
	}
	if t.Kind != model.Resource {
		return invalid("canonical resource type index")
	}
	representation, local := c.representations[t.Resource]
	if def.Op != 3 && !local {
		return invalid("resource representation belongs to another component")
	}
	f := module.FuncType{Params: []wasm.ValueType{wasm.I32}}
	if def.Op == 2 {
		f.Params[0] = representation
		f.Results = []wasm.ValueType{wasm.I32}
	} else if def.Op == 4 {
		f.Results = []wasm.ValueType{representation}
	}
	if err := c.addSignature(f); err != nil {
		return err
	}
	c.result.Canonical = append(c.result.Canonical, &CanonicalFunction{Definition: def, Type: t, Signature: f})
	return nil
}

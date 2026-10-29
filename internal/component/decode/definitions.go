package decode

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
)

func (d *decoder) coreSort() model.Sort {
	b := d.byte()
	switch b {
	case 0, 1, 2, 3, 4, 0x10, 0x11, 0x12:
		return model.Sort(b) | 0x100
	}
	d.fail("unknown Core sort")
	return 0
}
func (d *decoder) sort() model.Sort {
	b := d.byte()
	if b == 0 {
		return d.coreSort()
	}
	if b > 5 {
		d.fail("unknown component sort")
	}
	return model.Sort(b)
}
func (d *decoder) sortIndex() model.SortIndex { return model.SortIndex{Sort: d.sort(), Index: d.u32()} }
func (d *decoder) externName() model.Name {
	kind := d.byte()
	if kind > 2 {
		d.fail("invalid external name discriminator")
	}
	n := model.Name{Text: d.name()}
	if kind == 2 {
		n.Attributes = make([]model.Attribute, d.count())
		seen := byte(0)
		for i := range n.Attributes {
			kind := d.byte()
			if kind > 2 {
				d.fail("unknown external name attribute")
			}
			if seen&(1<<kind) != 0 {
				d.fail("duplicate external name attribute")
			}
			seen |= 1 << kind
			n.Attributes[i] = model.Attribute{Kind: kind, Value: d.name()}
		}
	}
	return n
}

func (d *decoder) externType() model.ExternType {
	sort := d.sort()
	t := model.ExternType{Sort: sort}
	switch sort {
	case model.CoreModuleSort, model.FunctionSort, model.ComponentSort, model.InstanceSort:
		t.Index = d.u32()
	case model.TypeSort:
		t.Bound = d.byte()
		if t.Bound == 0 {
			t.Index = d.u32()
		} else if t.Bound != 1 {
			d.fail("unknown resource type bound")
		}
	case model.ValueSort:
		t.Bound = d.byte()
		if t.Bound == 0 {
			t.Index = d.u32()
		} else if t.Bound == 1 {
			v := d.valueType()
			t.Value = &v
		} else {
			d.fail("unknown value bound")
		}
	default:
		d.fail("invalid external type sort")
	}
	return t
}
func (d *decoder) importType() *model.Import {
	return &model.Import{Name: d.externName(), Type: d.externType()}
}

func (d *decoder) instance(core bool) *model.InstanceExpr {
	flag := d.byte()
	if flag > 1 {
		d.fail("unknown instance expression")
	}
	x := &model.InstanceExpr{Instantiate: flag == 0}
	if x.Instantiate {
		x.Index = d.u32()
		x.Arguments = make([]model.Argument, d.count())
		for i := range x.Arguments {
			a := &x.Arguments[i]
			a.Name = d.name()
			if core {
				d.expect(0x12)
				a.Item = model.SortIndex{Sort: model.CoreInstanceSort, Index: d.u32()}
			} else {
				a.Item = d.sortIndex()
			}
		}
	} else {
		x.Exports = make([]model.Export, d.count())
		for i := range x.Exports {
			e := &x.Exports[i]
			if core {
				e.Name = model.Name{Text: d.name()}
				e.Item = model.SortIndex{Sort: d.coreSort(), Index: d.u32()}
			} else {
				e.Name = d.externName()
				e.Item = d.sortIndex()
			}
		}
	}
	return x
}

func (d *decoder) alias() *model.Alias {
	x := &model.Alias{Sort: d.sort(), Kind: d.byte()}
	switch x.Kind {
	case 0, 1:
		x.Instance = d.u32()
		x.Name = d.name()
	case 2:
		if x.Sort != model.CoreTypeSort && x.Sort != model.CoreModuleSort && x.Sort != model.TypeSort && x.Sort != model.ComponentSort {
			d.fail("invalid outer alias sort")
		}
		x.Count = d.u32()
		x.Index = d.u32()
	default:
		d.fail("unknown alias target")
	}
	if x.Sort == model.CoreTypeSort {
		d.core.Types = append(d.core.Types, module.FuncType{})
		d.core.TypeDefs = append(d.core.TypeDefs, module.TypeDef{})
	}
	return x
}

func (d *decoder) coreType() *model.CoreTypeDef {
	d.enter()
	defer func() { d.depth-- }()
	saved := d.r
	b := d.byte()
	if b != 0x50 {
		if b == 0 {
			d.expect(0x50)
			d.r = saved
			d.expect(0)
		} else {
			d.r = saved
		}
		start, count := d.coreGroup()
		return &model.CoreTypeDef{Start: start, Count: count}
	}
	parent := d.core
	x := &model.CoreTypeDef{Module: true, Start: uint32(len(parent.Types)), Count: 1, Types: new(module.Module)}
	parent.Types = append(parent.Types, module.FuncType{})
	parent.TypeDefs = append(parent.TypeDefs, module.TypeDef{})
	parent.TypeGroups = append(parent.TypeGroups, module.TypeGroup{Start: x.Start, Count: 1})
	d.core = x.Types
	defer func() { d.core = parent }()
	x.Declarations = make([]model.CoreDeclaration, d.count())
	for i := range x.Declarations {
		decl := &x.Declarations[i]
		decl.Kind = d.byte()
		switch decl.Kind {
		case 0:
			decl.Module = d.name()
			decl.Name = d.name()
			decl.Type = d.coreExtern()
		case 1:
			decl.Start, decl.Count = d.coreGroup()
		case 2:
			d.expect(0x10)
			d.expect(1)
			decl.Alias = &model.Alias{Sort: model.CoreTypeSort, Kind: 2, Count: d.u32(), Index: d.u32()}
			// An outer type alias occupies this lexical Core type index even
			// before semantic resolution binds its actual shape.
			d.core.Types = append(d.core.Types, module.FuncType{})
			d.core.TypeDefs = append(d.core.TypeDefs, module.TypeDef{})
		case 3:
			decl.Name = d.name()
			decl.Type = d.coreExtern()
		default:
			d.fail("unknown Core module declaration")
		}
	}
	return x
}

func (d *decoder) valueType() model.TypeRef {
	v := must(d.r.S33())
	if v >= 0 {
		return model.TypeRef{Indexed: true, Index: uint32(v)}
	}
	if v < -128 {
		d.fail("unknown component primitive value type")
	}
	kind := model.Kind(byte(v) & 0x7f)
	if kind < model.String || kind > model.Bool {
		if kind != model.ErrorContext {
			d.fail("unknown component primitive value type")
		}
	}
	return model.TypeRef{Primitive: kind}
}

func (d *decoder) maybeValue() *model.TypeRef {
	if !d.optional() {
		return nil
	}
	v := d.valueType()
	return &v
}
func (d *decoder) namedTypes() []model.NamedType {
	fields := make([]model.NamedType, d.count())
	for i := range fields {
		fields[i] = model.NamedType{Name: d.name(), Type: d.valueType()}
	}
	return fields
}
func (d *decoder) result() *model.TypeRef {
	flag := d.byte()
	if flag == 0 {
		v := d.valueType()
		return &v
	}
	if flag != 1 {
		d.fail("invalid function result encoding")
	}
	d.expect(0)
	return nil
}

func (d *decoder) typeDef() *model.TypeDef {
	d.enter()
	defer func() { d.depth-- }()
	d.charge(1)
	x := &model.TypeDef{Kind: model.Kind(d.byte())}
	switch x.Kind {
	case model.Bool, model.S8, model.U8, model.S16, model.U16, model.S32, model.U32, model.S64, model.U64, model.F32, model.F64, model.Char, model.String, model.ErrorContext:
	case model.Record:
		x.Fields = d.namedTypes()
	case model.Variant:
		x.Cases = make([]model.TypeCase, d.count())
		for i := range x.Cases {
			x.Cases[i] = model.TypeCase{Name: d.name(), Type: d.maybeValue()}
			d.expect(0)
		}
	case model.List, model.Option:
		v := d.valueType()
		x.Element = &v
	case model.FixedList:
		v := d.valueType()
		x.Element = &v
		x.Length = d.u32()
	case model.Map:
		k, v := d.valueType(), d.valueType()
		x.Key, x.Element = &k, &v
	case model.Tuple:
		x.Tuple = make([]model.TypeRef, d.count())
		for i := range x.Tuple {
			x.Tuple[i] = d.valueType()
		}
	case model.Flags, model.Enum:
		x.Labels = make([]string, d.count())
		for i := range x.Labels {
			x.Labels[i] = d.name()
		}
	case model.Result:
		x.Cases = []model.TypeCase{{Type: d.maybeValue()}, {Type: d.maybeValue()}}
	case model.Own, model.Borrow:
		x.Resource = d.u32()
	case model.Stream, model.Future:
		x.Element = d.maybeValue()
	case model.Function, model.AsyncFunction:
		x.Parameters = d.namedTypes()
		x.Result = d.result()
	case model.Resource:
		x.Representation = d.coreValue()
		if d.optional() {
			index := d.u32()
			x.Destructor = &index
		}
	case model.Component, model.Instance:
		parent := d.core
		x.CoreTypes = new(module.Module)
		d.core = x.CoreTypes
		defer func() { d.core = parent }()
		x.Declarations = make([]model.TypeDeclaration, d.count())
		for i := range x.Declarations {
			decl := &x.Declarations[i]
			decl.Kind = d.byte()
			switch decl.Kind {
			case 0:
				decl.CoreType = d.coreType()
			case 1:
				decl.Type = d.typeDef()
			case 2:
				decl.Alias = d.alias()
			case 3:
				if x.Kind != model.Component {
					d.fail("import declaration inside instance type")
				}
				decl.Import = d.importType()
			case 4:
				decl.Export = d.importType()
			default:
				d.fail("unknown component type declaration")
			}
		}
	default:
		d.fail("unknown component type opcode")
	}
	return x
}

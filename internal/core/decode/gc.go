package decode

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (d *decoder) typeSection(m *module.Module) {
	count := d.count()
	if d.features&wasm.GC == 0 {
		m.Types = make([]module.FuncType, count)
		for i := range m.Types {
			d.expect(0x60)
			m.Types[i] = module.FuncType{Params: d.types(), Results: d.types()}
		}
		return
	}
	for range count {
		d.typeGroup(m)
	}
	resolveTypeReferences(m)
}

func (d *decoder) typeGroup(m *module.Module) {
	start := uint32(len(m.Types))
	saved := d.r
	b := d.byte()
	n := uint32(1)
	if b == 0x4e {
		n = d.count()
	} else {
		d.r = saved
	}
	m.TypeGroups = append(m.TypeGroups, module.TypeGroup{Start: start, Count: n})
	for range n {
		b = d.byte()
		def := module.TypeDef{}
		if b == 0x50 || b == 0x4f {
			def.Open = b == 0x50
			def.Supers = d.indexes()
			b = d.byte()
		}
		def.Kind = b
		var f module.FuncType
		switch b {
		case module.FunctionKind:
			f = module.FuncType{Params: d.types(), Results: d.types()}
		case module.StructKind:
			def.Fields = make([]module.FieldType, d.count())
			for i := range def.Fields {
				def.Fields[i] = d.fieldType()
			}
		case module.ArrayKind:
			d.charge(1)
			def.Fields = []module.FieldType{d.fieldType()}
		default:
			d.fail("unknown composite type")
		}
		m.Types = append(m.Types, f)
		m.TypeDefs = append(m.TypeDefs, def)
	}
}

func resolveTypeReferences(m *module.Module) {
	resolveTypeReferencesFrom(m, 0)
}

func resolveTypeReferencesFrom(m *module.Module, start int) {
	// Forward references inside a recursive group receive their physical family
	// after all member shapes have been decoded. This never expands a recursion.
	resolve := func(t wasm.ValueType) wasm.ValueType {
		if i, ok := t.TypeIndex(); ok {
			return m.Reference(i, t.Nullable())
		}
		return t
	}
	for i := start; i < len(m.Types); i++ {
		for j, t := range m.Types[i].Params {
			m.Types[i].Params[j] = resolve(t)
		}
		for j, t := range m.Types[i].Results {
			m.Types[i].Results[j] = resolve(t)
		}
		for j := range m.TypeDefs[i].Fields {
			f := &m.TypeDefs[i].Fields[j]
			f.Value = resolve(f.Value)
		}
	}
}
func (d *decoder) fieldType() module.FieldType {
	saved := d.r
	b := d.byte()
	f := module.FieldType{}
	if b == 0x78 || b == 0x77 {
		f.Value = wasm.I32
		f.Packed = 8
		if b == 0x77 {
			f.Packed = 16
		}
	} else {
		d.r = saved
		f.Value = d.valueType()
	}
	mut := d.byte()
	if mut > 1 {
		d.fail("invalid field mutability")
	}
	f.Mutable = mut == 1
	return f
}
func (d *decoder) gcInstruction(in *module.Instr) {
	sub := d.u32()
	in.Op = 0xfb0000 | sub
	switch {
	case sub == 0 || sub == 1 || sub == 6 || sub == 7 || sub >= 11 && sub <= 14 || sub == 16:
		in.A = d.u32()
	case sub >= 2 && sub <= 5 || sub >= 8 && sub <= 10 || sub >= 17 && sub <= 19:
		in.A, in.B = d.u32(), d.u32()
		if sub == 9 || sub == 18 {
			d.needsDataCount = d.needsDataCount || d.inCode
		}
	case sub >= 20 && sub <= 23:
		t := d.heapReference(sub == 21 || sub == 23)
		in.Bits = uint64(t)
	case sub == 24 || sub == 25:
		flags := d.byte()
		if flags > 3 {
			d.fail("invalid cast branch flags")
		}
		in.A = d.u32()
		in.Types = []wasm.ValueType{d.heapReference(flags&1 != 0), d.heapReference(flags&2 != 0)}
	case sub == 15 || sub >= 26 && sub <= 30:
	default:
		d.failAt(in.Offset, "unknown GC opcode")
	}
}

package canonical

import (
	"math"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// LiftFlat converts one direct value. Function-level parameter/result spilling
// is separate; a value whose flattening exceeds 16 slots requires memory.
func (c *Codec) LiftFlat(t *model.Type, slots []Slot) (Value, error) {
	l, err := c.Layouts.Build(t)
	if err != nil {
		return Value{}, err
	}
	if l.FlatCount > MaxFlatParams || uint64(len(slots)) != l.FlatCount {
		return Value{}, invalid("flat value arity")
	}
	for i, want := range l.Flat {
		if slots[i].Type != want || (want == wasm.I32 || want == wasm.F32) && slots[i].Bits > math.MaxUint32 {
			return Value{}, invalid("flat value representation")
		}
	}
	at := 0
	return c.liftFlat(t, slots, &at, 0)
}

func (c *Codec) liftFlat(t *model.Type, slots []Slot, at *int, depth uint32) (Value, error) {
	if err := c.visit(depth); err != nil {
		return Value{}, err
	}
	if scalarKind(t.Kind) {
		v, err := LiftScalar(t.Kind, slots[*at])
		*at++
		return Value{Bits: v.Bits}, err
	}
	switch t.Kind {
	case model.String, model.List, model.Map:
		ptr, n := slots[*at].Bits, slots[*at+1].Bits
		*at += 2
		if t.Kind == model.String {
			s, err := c.Transfer.LiftString(ptr, n)
			return Value{Text: s}, err
		}
		element := t.Element
		if t.Kind == model.Map {
			layout, _ := c.Layouts.Build(t)
			element = layout.Entry
		}
		return c.loadList(element, ptr, n, depth)
	case model.FixedList:
		if uint64(t.Length) > c.MaxValues-c.values {
			return Value{}, exhausted()
		}
		if err := c.Transfer.charge(uint64(t.Length) * 64); err != nil {
			return Value{}, err
		}
		value := Value{Elements: make([]Value, t.Length)}
		for i := range value.Elements {
			var err error
			value.Elements[i], err = c.liftFlat(t.Element, slots, at, depth+1)
			if err != nil {
				return Value{}, err
			}
		}
		return value, nil
	case model.Record, model.Tuple:
		if uint64(len(t.Fields)) > c.MaxValues-c.values {
			return Value{}, exhausted()
		}
		if err := c.Transfer.charge(uint64(len(t.Fields)) * 64); err != nil {
			return Value{}, err
		}
		v := Value{Elements: make([]Value, len(t.Fields))}
		for i, f := range t.Fields {
			var err error
			v.Elements[i], err = c.liftFlat(f.Type, slots, at, depth+1)
			if err != nil {
				return Value{}, err
			}
		}
		return v, nil
	case model.Flags, model.Enum:
		bits := slots[*at].Bits
		*at++
		if t.Kind == model.Flags {
			bits &= (uint64(1) << len(t.Labels)) - 1
		} else if bits >= uint64(len(t.Labels)) {
			return Value{}, scalarTrap("invalid enum discriminant")
		}
		return Value{Bits: bits}, nil
	case model.Variant, model.Option, model.Result:
		cs := cases(t)
		index := slots[*at].Bits
		*at++
		if index >= uint64(len(cs)) {
			return Value{}, scalarTrap("invalid variant discriminant")
		}
		layout, _ := c.Layouts.Build(t)
		payload := int(layout.FlatCount) - 1
		v := Value{Case: uint32(index)}
		if child := cs[index].Type; child != nil {
			cl, _ := c.Layouts.Build(child)
			var converted [MaxFlatParams]Slot
			for i, want := range cl.Flat {
				converted[i] = slots[*at+i]
				converted[i].Type = want
				if want == wasm.I32 || want == wasm.F32 {
					converted[i].Bits = uint64(uint32(converted[i].Bits))
				}
			}
			inner := 0
			value, err := c.liftFlat(child, converted[:cl.FlatCount], &inner, depth+1)
			if err != nil {
				return Value{}, err
			}
			if err := c.Transfer.charge(64); err != nil {
				return Value{}, err
			}
			v.Elements = []Value{value}
		}
		*at += payload
		return v, nil
	}
	return Value{}, wasm.ErrUnsupported
}

func (c *Codec) LowerFlat(t *model.Type, v Value) ([]Slot, error) {
	if err := c.validate(t, v, 0); err != nil {
		return nil, err
	}
	l, _ := c.Layouts.Build(t)
	if l.FlatCount > MaxFlatParams {
		return nil, invalid("flat value requires memory")
	}
	result := make([]Slot, 0, l.FlatCount)
	return c.lowerFlat(t, v, result, 0)
}

func (c *Codec) lowerFlat(t *model.Type, v Value, out []Slot, depth uint32) ([]Slot, error) {
	if err := c.Transfer.check(); err != nil {
		return nil, err
	}
	if scalarKind(t.Kind) {
		slot, err := LowerScalar(Scalar{t.Kind, v.Bits})
		return append(out, slot), err
	}
	switch t.Kind {
	case model.String, model.List, model.Map:
		var ptr, n uint64
		var err error
		if t.Kind == model.String {
			ptr, n, err = c.Transfer.LowerString(v.Text)
		} else {
			n = uint64(len(v.Elements))
			element := t.Element
			if t.Kind == model.Map {
				layout, err := c.Layouts.Build(t)
				if err != nil {
					return nil, err
				}
				element = layout.Entry
			}
			ptr, err = c.storeList(element, v.Elements, depth)
		}
		if err != nil {
			return nil, err
		}
		flat := c.Layouts.address
		return append(out, Slot{flat, ptr}, Slot{flat, n}), nil
	case model.FixedList:
		var err error
		for _, value := range v.Elements {
			out, err = c.lowerFlat(t.Element, value, out, depth+1)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case model.Record, model.Tuple:
		var err error
		for i, f := range t.Fields {
			out, err = c.lowerFlat(f.Type, v.Elements[i], out, depth+1)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case model.Flags, model.Enum:
		return append(out, Slot{wasm.I32, v.Bits}), nil
	case model.Variant, model.Option, model.Result:
		l, _ := c.Layouts.Build(t)
		start := len(out)
		out = append(out, Slot{wasm.I32, uint64(v.Case)})
		if child := cases(t)[v.Case].Type; child != nil {
			var err error
			out, err = c.lowerFlat(child, v.Elements[0], out, depth+1)
			if err != nil {
				return nil, err
			}
		}
		for len(out) < start+len(l.Flat) {
			out = append(out, Slot{})
		}
		for i, want := range l.Flat {
			// Joining f32 with i32 reinterprets its bits; a 32-bit payload joined
			// with a 64-bit one is zero extended. No numeric float conversion.
			out[start+i].Type = want
		}
		return out, nil
	}
	return nil, wasm.ErrUnsupported
}

// FunctionLayout implements the synchronous ABI's asymmetric spill convention:
// lifted functions return an indirect result pointer; lowered host functions
// receive that pointer as the last parameter and return no Core results.
type FunctionLayout struct {
	Params, Results               []wasm.ValueType
	ParameterMemory, ResultMemory *model.Type
}

func (b *Builder) Function(params []*model.Type, result *model.Type, lifting bool) (FunctionLayout, error) {
	var out FunctionLayout
	if uint64(len(params)) > uint64(b.limits.MaxNodes-b.nodes) {
		return out, exhausted()
	}
	var fields []model.Field
	count := uint64(0)
	for _, t := range params {
		l, err := b.Build(t)
		if err != nil {
			return out, err
		}
		if count > math.MaxUint64-l.FlatCount {
			return out, exhausted()
		}
		count += l.FlatCount
		if count <= MaxFlatParams {
			out.Params = append(out.Params, l.Flat...)
		}
		fields = append(fields, model.Field{Type: t})
	}
	if count > MaxFlatParams {
		out.Params = []wasm.ValueType{b.address}
		out.ParameterMemory = &model.Type{Kind: model.Tuple, Fields: fields}
		if _, err := b.Build(out.ParameterMemory); err != nil {
			return out, err
		}
	}
	if result != nil {
		l, err := b.Build(result)
		if err != nil {
			return out, err
		}
		if l.FlatCount <= MaxFlatResults {
			out.Results = append(out.Results, l.Flat...)
		} else {
			out.ResultMemory = result
			if lifting {
				out.Results = []wasm.ValueType{b.address}
			} else {
				out.Params = append(out.Params, b.address)
			}
		}
	}
	return out, nil
}

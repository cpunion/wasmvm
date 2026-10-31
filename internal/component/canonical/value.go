package canonical

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Value is a host-owned component value. Fields and tuple/list elements retain
// type order; variants use Case and zero or one payload in Elements. Bits uses
// Scalar's representation, or the declared flags bitmap/enum discriminant.
// Lists/fixed-lists of u8 may use immutable raw bytes in Text with nil Elements.
// Resource holds opaque resource or error-context references; their validated
// type determines affine ownership versus immutable diagnostic sharing.
type Value struct {
	Bits     uint64
	Text     string
	Elements []Value
	Case     uint32
	Resource *Reference
}

// Codec belongs to one synchronous call. All type graphs/layouts are immutable;
// transfer and value budgets are aggregate across nested lists and strings.
type Codec struct {
	Transfer  *Transfer
	Layouts   *Builder
	MaxValues uint64
	values    uint64
	Resources ResourceCodec
}

// Usage reports charged value visits and copied/retained bytes for aggregate
// reservations while asynchronous values remain live.
func (c *Codec) Usage() (uint64, uint64) { return c.values, c.Transfer.used }

// Validate checks a host value before calls or transfers cause guest effects.
func (c *Codec) Validate(t *model.Type, v Value) error { return c.validate(t, v, 0) }

// Clone validates and copies the declared value shape, discarding irrelevant
// fields. Host aggregate slices never alias cached literals or instance state.
// Strings are immutable; their retained bytes still consume transfer budget.
func (c *Codec) Clone(t *model.Type, v Value) (Value, error) {
	if err := c.validate(t, v, 0); err != nil {
		return Value{}, err
	}
	return c.clone(t, v)
}

func (c *Codec) clone(t *model.Type, v Value) (Value, error) {
	if t.Kind == model.Own || t.Kind == model.Borrow || t.Kind == model.ErrorContext || t.Kind == model.Future || t.Kind == model.Stream {
		amount := uint64(128)
		if t.Kind == model.ErrorContext {
			message, _ := ErrorContextMessage(v)
			amount += uint64(len(message))
		}
		if err := c.Transfer.charge(amount); err != nil {
			return Value{}, err
		}
		if t.Kind == model.ErrorContext {
			return snapshotErrorContext(v), nil
		}
		return Value{Resource: v.Resource}, nil
	}
	if scalarKind(t.Kind) || t.Kind == model.Flags || t.Kind == model.Enum {
		return Value{Bits: v.Bits}, nil
	}
	if t.Kind == model.String {
		if err := c.Transfer.charge(uint64(len(v.Text))); err != nil {
			return Value{}, err
		}
		return Value{Text: v.Text}, nil
	}
	if byteSequence(t) {
		if v.Elements != nil {
			if err := c.Transfer.charge(2 * uint64(len(v.Elements))); err != nil {
				return Value{}, err
			}
			data, err := v.ByteString()
			return Value{Text: data}, err
		}
		if err := c.Transfer.charge(uint64(len(v.Text))); err != nil {
			return Value{}, err
		}
		return Value{Text: v.Text}, nil
	}
	if err := c.Transfer.charge(uint64(len(v.Elements)) * 64); err != nil {
		return Value{}, err
	}
	out := Value{}
	if len(v.Elements) != 0 {
		out.Elements = make([]Value, len(v.Elements))
	}
	if t.Kind == model.Variant || t.Kind == model.Option || t.Kind == model.Result {
		out.Case = v.Case
	}
	for i, element := range v.Elements {
		var child *model.Type
		switch t.Kind {
		case model.Record, model.Tuple:
			child = t.Fields[i].Type
		case model.Variant, model.Option, model.Result:
			child = cases(t)[v.Case].Type
		case model.List, model.FixedList:
			child = t.Element
		case model.Map:
			if err := c.Transfer.charge(128); err != nil {
				return Value{}, err
			}
			key, err := c.clone(t.Key, element.Elements[0])
			if err != nil {
				return Value{}, err
			}
			value, err := c.clone(t.Element, element.Elements[1])
			if err != nil {
				return Value{}, err
			}
			out.Elements[i] = Value{Elements: []Value{key, value}}
			continue
		}
		var err error
		out.Elements[i], err = c.clone(child, element)
		if err != nil {
			return Value{}, err
		}
	}
	return out, nil
}

func (c *Codec) visit(depth uint32) error {
	if err := c.Transfer.check(); err != nil {
		return err
	}
	if c.values >= c.MaxValues || depth >= c.Layouts.limits.MaxDepth {
		return exhausted()
	}
	c.values++
	return nil
}

func scalarKind(k model.Kind) bool { return k >= model.Char && k <= model.Bool }

func cases(t *model.Type) []model.Field {
	if t.Kind == model.Option {
		return []model.Field{{}, {Type: t.Element}}
	}
	return t.Cases
}

func (c *Codec) Load(t *model.Type, at uint64) (Value, error) { return c.load(t, at, 0) }

func (c *Codec) load(t *model.Type, at uint64, depth uint32) (Value, error) {
	if err := c.visit(depth); err != nil {
		return Value{}, err
	}
	l, err := c.Layouts.Build(t)
	if err != nil {
		return Value{}, err
	}
	if err := c.Transfer.rangeCheck(at, l.Size, l.Align); err != nil {
		return Value{}, err
	}
	if scalarKind(t.Kind) {
		s, err := LoadScalar(c.Transfer.Memory, at, t.Kind)
		return Value{Bits: s.Bits}, err
	}
	switch t.Kind {
	case model.Own, model.Borrow, model.ErrorContext, model.Future, model.Stream:
		r, err := c.resource()
		if err != nil {
			return Value{}, err
		}
		id, err := c.loadInt(at, 4)
		if err != nil {
			return Value{}, err
		}
		if err := c.Transfer.charge(128); err != nil {
			return Value{}, err
		}
		return r.Lift(t, uint32(id))

	case model.String:
		s, err := c.Transfer.LoadString(at)
		return Value{Text: s}, err
	case model.Record, model.Tuple:
		if uint64(len(t.Fields)) > c.MaxValues-c.values {
			return Value{}, exhausted()
		}
		if err := c.Transfer.charge(uint64(len(t.Fields)) * 64); err != nil {
			return Value{}, err
		}
		v := Value{Elements: make([]Value, len(t.Fields))}
		for i, f := range t.Fields {
			v.Elements[i], err = c.load(f.Type, at+l.Offsets[i], depth+1)
			if err != nil {
				return Value{}, err
			}
		}
		return v, nil
	case model.List, model.Map:
		ptr, n, err := c.loadPair(at)
		if err != nil {
			return Value{}, err
		}
		element := t.Element
		if t.Kind == model.Map {
			element = l.Entry
		}
		return c.loadList(element, ptr, n, depth)
	case model.FixedList:
		return c.loadList(t.Element, at, uint64(t.Length), depth)
	case model.Flags, model.Enum:
		bits, err := c.loadInt(at, l.Size)
		if err != nil {
			return Value{}, err
		}
		if t.Kind == model.Enum {
			if bits >= uint64(len(t.Labels)) {
				return Value{}, scalarTrap("invalid enum discriminant")
			}
		} else {
			bits &= (uint64(1) << len(t.Labels)) - 1
		}
		return Value{Bits: bits}, nil
	case model.Variant, model.Option, model.Result:
		cs := cases(t)
		size, _ := discriminant(len(cs))
		index, err := c.loadInt(at, size)
		if err != nil {
			return Value{}, err
		}
		if index >= uint64(len(cs)) {
			return Value{}, scalarTrap("invalid variant discriminant")
		}
		v := Value{Case: uint32(index)}
		if cs[index].Type != nil {
			if err := c.Transfer.charge(64); err != nil {
				return Value{}, err
			}
			child, err := c.load(cs[index].Type, at+l.PayloadOffset, depth+1)
			if err != nil {
				return Value{}, err
			}
			v.Elements = []Value{child}
		}
		return v, nil
	}
	return Value{}, wasm.ErrUnsupported
}

func (c *Codec) loadList(t *model.Type, at, n uint64, depth uint32) (Value, error) {
	if t.Kind == model.U8 {
		return c.loadBytes(at, n, depth)
	}
	l, err := c.Layouts.Build(t)
	if err != nil {
		return Value{}, err
	}
	if n > c.MaxValues-c.values || n > MaxStringBytes/l.Size {
		return Value{}, exhausted()
	}
	if err := c.Transfer.rangeCheck(at, n*l.Size, l.Align); err != nil {
		return Value{}, err
	}
	// Account for host value slots independently of guest element byte size.
	if err := c.Transfer.charge(n * 64); err != nil {
		return Value{}, err
	}
	v := Value{Elements: make([]Value, int(n))}
	for i := range v.Elements {
		v.Elements[i], err = c.load(t, at+uint64(i)*l.Size, depth+1)
		if err != nil {
			return Value{}, err
		}
	}
	return v, nil
}

func (c *Codec) loadInt(at, size uint64) (uint64, error) {
	var b [8]byte
	if err := c.Transfer.Memory.Read(at, b[:size]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

func (c *Codec) storeInt(at, size, value uint64) error {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], value)
	return c.Transfer.Memory.Write(at, b[:size])
}

func (c *Codec) loadPair(at uint64) (uint64, uint64, error) {
	w := c.Transfer.pointerSize()
	ptr, err := c.loadInt(at, w)
	if err != nil {
		return 0, 0, err
	}
	n, err := c.loadInt(at+w, w)
	return ptr, n, err
}

func (c *Codec) storePair(at, ptr, n uint64) error {
	w := c.Transfer.pointerSize()
	if err := c.storeInt(at, w, ptr); err != nil {
		return err
	}
	return c.storeInt(at+w, w, n)
}

// Store validates all host values before allocating or modifying guest memory.
// Realloc and memory failures may preserve earlier guest allocation effects;
// callers must retain the enclosing instance until normal cleanup completes.
func (c *Codec) Store(t *model.Type, at uint64, value Value) error {
	if err := c.validate(t, value, 0); err != nil {
		return err
	}
	return c.store(t, at, value, 0)
}

func (c *Codec) validate(t *model.Type, v Value, depth uint32) error {
	if err := c.visit(depth); err != nil {
		return err
	}
	if _, err := c.Layouts.Build(t); err != nil {
		return err
	}
	if scalarKind(t.Kind) {
		_, err := LowerScalar(Scalar{t.Kind, v.Bits})
		return err
	}
	switch t.Kind {
	case model.ErrorContext:
		_, err := ErrorContextMessage(v)
		return err
	case model.Own, model.Borrow, model.Future, model.Stream:
		r, err := c.resource()
		if err != nil {
			return err
		}
		return r.Validate(t, v)

	case model.String:
		return c.Transfer.validateString(v.Text)
	case model.List, model.FixedList:
		if byteSequence(t) && v.Elements == nil {
			if t.Kind == model.FixedList && uint64(len(v.Text)) != uint64(t.Length) {
				return invalid("host fixed-list length")
			}
			return c.byteValues(uint64(len(v.Text)), depth)
		}
		if byteSequence(t) && v.Text != "" {
			return invalid("mixed byte-list representations")
		}
		if t.Kind == model.FixedList && uint64(len(v.Elements)) != uint64(t.Length) {
			return invalid("host fixed-list length")
		}
		if uint64(len(v.Elements)) > c.MaxValues-c.values {
			return exhausted()
		}
		for _, v := range v.Elements {
			if err := c.validate(t.Element, v, depth+1); err != nil {
				return err
			}
		}
	case model.Map:
		if uint64(len(v.Elements)) > c.MaxValues-c.values {
			return exhausted()
		}
		for _, entry := range v.Elements {
			if len(entry.Elements) != 2 {
				return invalid("host map entry arity")
			}
			if err := c.visit(depth + 1); err != nil {
				return err
			}
			if err := c.validate(t.Key, entry.Elements[0], depth+2); err != nil {
				return err
			}
			if err := c.validate(t.Element, entry.Elements[1], depth+2); err != nil {
				return err
			}
		}
	case model.Record, model.Tuple:
		if len(v.Elements) != len(t.Fields) {
			return invalid("host aggregate arity")
		}
		for i, f := range t.Fields {
			if err := c.validate(f.Type, v.Elements[i], depth+1); err != nil {
				return err
			}
		}
	case model.Flags:
		if v.Bits>>len(t.Labels) != 0 {
			return invalid("host flags bitmap")
		}
	case model.Enum:
		if v.Bits >= uint64(len(t.Labels)) {
			return invalid("host enum discriminant")
		}
	case model.Variant, model.Option, model.Result:
		cs := cases(t)
		if uint64(v.Case) >= uint64(len(cs)) {
			return invalid("host variant discriminant")
		}
		child := cs[v.Case].Type
		if child == nil {
			if len(v.Elements) != 0 {
				return invalid("unexpected variant payload")
			}
		} else {
			if len(v.Elements) != 1 {
				return invalid("missing variant payload")
			}
			return c.validate(child, v.Elements[0], depth+1)
		}
	default:
		return wasm.ErrUnsupported
	}
	return nil
}

func (c *Codec) store(t *model.Type, at uint64, v Value, depth uint32) error {
	if err := c.Transfer.check(); err != nil {
		return err
	}
	l, err := c.Layouts.Build(t)
	if err != nil {
		return err
	}
	if err := c.Transfer.rangeCheck(at, l.Size, l.Align); err != nil {
		return err
	}
	if scalarKind(t.Kind) {
		return StoreScalar(c.Transfer.Memory, at, Scalar{t.Kind, v.Bits})
	}
	switch t.Kind {
	case model.Own, model.Borrow, model.ErrorContext, model.Future, model.Stream:
		r, err := c.resource()
		if err != nil {
			return err
		}
		id, err := r.Lower(t, v)
		if err != nil {
			return err
		}
		return c.storeInt(at, 4, uint64(id))

	case model.String:
		return c.Transfer.StoreString(at, v.Text)
	case model.Record, model.Tuple:
		for i, f := range t.Fields {
			if err := c.store(f.Type, at+l.Offsets[i], v.Elements[i], depth+1); err != nil {
				return err
			}
		}
	case model.List, model.Map:
		if t.Kind == model.List && byteSequence(t) && v.Elements == nil {
			ptr, err := c.lowerBytes(v.Text)
			if err != nil {
				return err
			}
			return c.storePair(at, ptr, uint64(len(v.Text)))
		}
		element := t.Element
		if t.Kind == model.Map {
			element = l.Entry
		}
		ptr, err := c.storeList(element, v.Elements, depth)
		if err != nil {
			return err
		}
		return c.storePair(at, ptr, uint64(len(v.Elements)))
	case model.FixedList:
		if byteSequence(t) && v.Elements == nil {
			return c.storeBytes(at, v.Text)
		}
		child, err := c.Layouts.Build(t.Element)
		if err != nil {
			return err
		}
		for i, value := range v.Elements {
			if err := c.store(t.Element, at+uint64(i)*child.Size, value, depth+1); err != nil {
				return err
			}
		}
	case model.Flags, model.Enum:
		return c.storeInt(at, l.Size, v.Bits)
	case model.Variant, model.Option, model.Result:
		cs := cases(t)
		size, _ := discriminant(len(cs))
		if err := c.storeInt(at, size, uint64(v.Case)); err != nil {
			return err
		}
		if child := cs[v.Case].Type; child != nil {
			return c.store(child, at+l.PayloadOffset, v.Elements[0], depth+1)
		}
	default:
		return wasm.ErrUnsupported
	}
	return nil
}

func (c *Codec) storeList(t *model.Type, values []Value, depth uint32) (uint64, error) {
	l, err := c.Layouts.Build(t)
	if err != nil {
		return 0, err
	}
	if uint64(len(values)) > MaxStringBytes/l.Size {
		return 0, exhausted()
	}
	n := uint64(len(values)) * l.Size
	if err := c.Transfer.charge(n); err != nil {
		return 0, err
	}
	if c.Transfer.Realloc == nil {
		return 0, invalid("missing realloc")
	}
	ptr, err := c.Transfer.Realloc(0, 0, l.Align, n)
	if err != nil {
		return 0, err
	}
	if err := c.Transfer.rangeCheck(ptr, n, l.Align); err != nil {
		return 0, err
	}
	for i, v := range values {
		if err := c.store(t, ptr+uint64(i)*l.Size, v, depth+1); err != nil {
			return 0, err
		}
	}
	return ptr, nil
}

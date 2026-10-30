package canonical

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/binary"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// DecodeLiteral decodes a type-directed component value definition. Its wire
// format is distinct from Canonical ABI memory. The input type must be an
// immutable, validated graph; scratch/value allocation is bounded separately
// from the input bytes. No resource, future, stream or error handles are forged.
func DecodeLiteral(t *model.Type, input []byte, limits Limits) (Value, error) {
	d := literalDecoder{reader: binary.New(input), limits: limits, admissible: map[*model.Type]bool{}}
	if uint64(len(input)) > limits.MaxBytes {
		return Value{}, exhausted()
	}
	if err := d.checkType(t, 0); err != nil {
		return Value{}, err
	}
	v, err := d.value(t, 0)
	if err == nil {
		err = d.reader.Finish()
	}
	if err != nil {
		if errors.Is(err, wasm.ErrLimit) {
			return Value{}, err
		}
		return Value{}, fmt.Errorf("%w: component value: %v", wasm.ErrInvalid, err)
	}
	return v, nil
}

type literalDecoder struct {
	reader     binary.Reader
	limits     Limits
	nodes      uint32
	bytes      uint64
	admissible map[*model.Type]bool
}

func (d *literalDecoder) charge(nodes uint32, bytes uint64) error {
	if nodes > d.limits.MaxNodes-d.nodes || bytes > d.limits.MaxBytes-d.bytes {
		return exhausted()
	}
	d.nodes += nodes
	d.bytes += bytes
	return nil
}

func (d *literalDecoder) checkType(t *model.Type, depth uint32) error {
	if t == nil {
		return invalid("missing literal type")
	}
	if d.admissible[t] {
		return nil
	}
	if depth >= d.limits.MaxDepth {
		return exhausted()
	}
	if err := d.charge(1, 0); err != nil {
		return err
	}
	switch t.Kind {
	case model.Own, model.Borrow, model.Future, model.Stream, model.ErrorContext:
		return invalid("type has no literal constructor")
	case model.Record, model.Tuple, model.Variant, model.Result:
		for _, fields := range [][]model.Field{t.Fields, t.Cases} {
			for _, f := range fields {
				if f.Type != nil {
					if err := d.checkType(f.Type, depth+1); err != nil {
						return err
					}
				}
			}
		}
	case model.List, model.FixedList, model.Option, model.Map:
		if err := d.checkType(t.Element, depth+1); err != nil {
			return err
		}
		if t.Kind == model.Map {
			if err := d.checkType(t.Key, depth+1); err != nil {
				return err
			}
		}
	default:
		if !scalarKind(t.Kind) && t.Kind != model.String && t.Kind != model.Enum && t.Kind != model.Flags {
			return invalid("unknown literal type")
		}
	}
	d.admissible[t] = true
	return nil
}

func (d *literalDecoder) elements(n uint32) ([]Value, error) {
	if n > d.limits.MaxNodes-d.nodes {
		return nil, exhausted()
	}
	if err := d.charge(0, uint64(n)*64); err != nil {
		return nil, err
	}
	return make([]Value, n), nil
}

func (d *literalDecoder) value(t *model.Type, depth uint32) (Value, error) {
	if depth >= d.limits.MaxDepth {
		return Value{}, exhausted()
	}
	if err := d.charge(1, 0); err != nil {
		return Value{}, err
	}
	var v Value
	var err error
	switch t.Kind {
	case model.Bool, model.U8, model.S8:
		var n byte
		n, err = d.reader.Byte()
		v.Bits = uint64(n)
		if t.Kind == model.S8 {
			v.Bits = uint64(int64(int8(n)))
		}
		if t.Kind == model.Bool && n > 1 {
			return v, invalid("literal boolean")
		}
	case model.U16, model.U32, model.U64:
		if t.Kind == model.U64 {
			v.Bits, err = d.reader.U64()
		} else if t.Kind == model.U16 {
			var n uint16
			n, err = d.reader.U16()
			v.Bits = uint64(n)
		} else {
			var n uint32
			n, err = d.reader.U32()
			v.Bits = uint64(n)
		}
	case model.S16, model.S32, model.S64:
		if t.Kind == model.S16 {
			var n int16
			n, err = d.reader.S16()
			v.Bits = uint64(int64(n))
		} else if t.Kind == model.S32 {
			var n int32
			n, err = d.reader.S32()
			v.Bits = uint64(int64(n))
		} else {
			var n int64
			n, err = d.reader.S64()
			v.Bits = uint64(n)
		}
	case model.F32:
		var n uint32
		n, err = d.reader.F32Bits()
		v.Bits = uint64(n)
		if n&0x7fffffff > 0x7f800000 && n != 0x7fc00000 {
			return v, invalid("noncanonical literal f32 NaN")
		}
	case model.F64:
		v.Bits, err = d.reader.F64Bits()
		if v.Bits&0x7fffffffffffffff > 0x7ff0000000000000 && v.Bits != 0x7ff8000000000000 {
			return v, invalid("noncanonical literal f64 NaN")
		}
	case model.Char:
		var bytes [4]byte
		bytes[0], err = d.reader.Byte()
		if err != nil {
			return v, err
		}
		n := 1
		for bytes[0]&(0x80>>n) != 0 && n < 4 {
			n++
		}
		if bytes[0] < utf8.RuneSelf {
			n = 1
		}
		for i := 1; i < n; i++ {
			bytes[i], err = d.reader.Byte()
			if err != nil {
				return v, err
			}
		}
		r, width := utf8.DecodeRune(bytes[:n])
		if width != n || r == utf8.RuneError && width == 1 {
			return v, invalid("literal Unicode scalar")
		}
		v.Bits = uint64(r)
	case model.String:
		var n uint32
		n, err = d.reader.U32()
		if err == nil {
			if err = d.charge(0, uint64(n)); err != nil {
				return v, err
			}
			var bytes []byte
			bytes, err = d.reader.Bytes(uint64(n))
			if err == nil && !utf8.Valid(bytes) {
				return v, invalid("literal UTF-8 string")
			}
			v.Text = string(bytes)
		}
	case model.Flags:
		var bytes []byte
		bytes, err = d.reader.Bytes(uint64(len(t.Labels)+7) / 8)
		for i, b := range bytes {
			v.Bits |= uint64(b) << (8 * i)
		}
		if len(t.Labels) < 64 {
			v.Bits &= (uint64(1) << len(t.Labels)) - 1
		}
	case model.Enum:
		var n uint32
		n, err = d.reader.U32()
		v.Bits = uint64(n)
		if err == nil && uint64(n) >= uint64(len(t.Labels)) {
			return v, invalid("literal enum discriminant")
		}
	case model.Variant, model.Option, model.Result:
		if t.Kind == model.Variant {
			v.Case, err = d.reader.U32()
		} else {
			var n byte
			n, err = d.reader.Byte()
			v.Case = uint32(n)
		}
		cs := cases(t)
		if err != nil {
			return v, err
		}
		if uint64(v.Case) >= uint64(len(cs)) {
			return v, invalid("literal variant discriminant")
		}
		if payload := cs[v.Case].Type; payload != nil {
			v.Elements, err = d.elements(1)
			if err == nil {
				v.Elements[0], err = d.value(payload, depth+1)
			}
		}
	case model.Record, model.Tuple:
		v.Elements, err = d.elements(uint32(len(t.Fields)))
		if err == nil {
			for i, field := range t.Fields {
				v.Elements[i], err = d.value(field.Type, depth+1)
				if err != nil {
					return v, err
				}
			}
		}
	case model.List, model.FixedList, model.Map:
		n := t.Length
		if t.Kind != model.FixedList {
			n, err = d.reader.U32()
		}
		if err != nil {
			return v, err
		}
		v.Elements, err = d.elements(n)
		if err != nil {
			return v, err
		}
		for i := range v.Elements {
			if t.Kind == model.Map {
				entry := &v.Elements[i]
				if err := d.charge(1, 0); err != nil {
					return v, err
				}
				entry.Elements, err = d.elements(2)
				if err == nil {
					entry.Elements[0], err = d.value(t.Key, depth+1)
				}
				if err == nil {
					entry.Elements[1], err = d.value(t.Element, depth+1)
				}
			} else {
				v.Elements[i], err = d.value(t.Element, depth+1)
			}
			if err != nil {
				return v, err
			}
		}
	}
	return v, err
}

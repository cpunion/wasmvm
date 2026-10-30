// Package canonical implements component Canonical ABI layout and conversion.
// These primitives do not establish component execution or WASI world support.
package canonical

import (
	"fmt"
	"math"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// MAX_FLAT_PARAMS and MAX_FLAT_RESULTS belong to the pinned synchronous ABI.
const MaxFlatParams, MaxFlatResults = 16, 1

type Limits struct {
	MaxNodes uint32
	MaxDepth uint32
	MaxBytes uint64
}

func DefaultLimits() Limits { return Limits{1 << 20, 128, math.MaxUint32} }

type Layout struct {
	Size, Align   uint64
	Offsets       []uint64
	PayloadOffset uint64
	// Flat is materialized only up to MaxFlatParams. FlatCount preserves the
	// full count without expanding a shared type DAG into an exponential slice.
	Flat      []wasm.ValueType
	FlatCount uint64
	// Entry is the cached tuple schema used by the map's list representation.
	Entry *model.Type
}

type Builder struct {
	limits  Limits
	address wasm.ValueType
	pointer uint64
	nodes   uint32
	active  map[*model.Type]bool
	cache   map[*model.Type]*Layout
}

// NewBuilder selects the canonical memory's address width. The input type
// graph must remain immutable for this builder's lifetime.
func NewBuilder(memory64 bool, limits Limits) *Builder {
	b := &Builder{limits: limits, address: wasm.I32, pointer: 4, active: map[*model.Type]bool{}, cache: map[*model.Type]*Layout{}}
	if memory64 {
		b.address, b.pointer = wasm.I64, 8
	}
	return b
}

func invalid(detail string) error { return fmt.Errorf("%w: canonical %s", wasm.ErrInvalid, detail) }
func exhausted() error            { return fmt.Errorf("%w: canonical layout", wasm.ErrLimit) }

// Build returns builder-owned immutable layout metadata.
func (b *Builder) Build(t *model.Type) (*Layout, error) { return b.build(t, 0) }

func (b *Builder) build(t *model.Type, depth uint32) (*Layout, error) {
	if t == nil {
		return nil, invalid("missing value type")
	}
	if b.active[t] {
		return nil, invalid("recursive value layout")
	}
	if got := b.cache[t]; got != nil {
		return got, nil
	}
	if depth >= b.limits.MaxDepth || b.nodes >= b.limits.MaxNodes {
		return nil, exhausted()
	}
	b.nodes++
	b.active[t] = true
	defer delete(b.active, t)
	out := &Layout{}
	scalar := func(size uint64, flat wasm.ValueType) {
		out.Size, out.Align = size, size
		out.Flat = []wasm.ValueType{flat}
		out.FlatCount = 1
	}
	switch t.Kind {
	case model.Bool, model.U8, model.S8:
		scalar(1, wasm.I32)
	case model.U16, model.S16:
		scalar(2, wasm.I32)
	case model.U32, model.S32, model.Char:
		scalar(4, wasm.I32)
	case model.U64, model.S64:
		scalar(8, wasm.I64)
	case model.F32:
		scalar(4, wasm.F32)
	case model.F64:
		scalar(8, wasm.F64)
	case model.ErrorContext, model.Future, model.Stream:
		scalar(4, wasm.I32)
		if t.Element != nil {
			if _, err := b.build(t.Element, depth+1); err != nil {
				return nil, err
			}
		}
	case model.FixedList:
		if t.Length == 0 {
			return nil, invalid("empty fixed list")
		}
		child, err := b.build(t.Element, depth+1)
		if err != nil {
			return nil, err
		}
		if child.Size > b.limits.MaxBytes/uint64(t.Length) || child.FlatCount > math.MaxUint64/uint64(t.Length) {
			return nil, exhausted()
		}
		out.Size, out.Align = child.Size*uint64(t.Length), child.Align
		out.FlatCount = child.FlatCount * uint64(t.Length)
		if out.FlatCount <= MaxFlatParams {
			out.Flat = make([]wasm.ValueType, 0, out.FlatCount)
			for i := uint32(0); i < t.Length; i++ {
				out.Flat = append(out.Flat, child.Flat...)
			}
		}
	case model.String, model.List, model.Map:
		out.Size, out.Align = 2*b.pointer, b.pointer
		out.Flat, out.FlatCount = []wasm.ValueType{b.address, b.address}, 2
		if t.Kind == model.List || t.Kind == model.Map {
			if _, err := b.build(t.Element, depth+1); err != nil {
				return nil, err
			}
			if t.Kind == model.Map {
				if _, err := b.build(t.Key, depth+1); err != nil {
					return nil, err
				}
				out.Entry = &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: t.Key}, {Type: t.Element}}}
				if _, err := b.build(out.Entry, depth); err != nil {
					return nil, err
				}
			}
		}
	case model.Own, model.Borrow:
		if t.ResourceID == 0 {
			return nil, invalid("missing resource identity")
		}
		scalar(4, wasm.I32)
	case model.Flags:
		n := len(t.Labels)
		if n == 0 || n > 32 {
			return nil, invalid("flag count")
		}
		size := uint64(4)
		if n <= 8 {
			size = 1
		} else if n <= 16 {
			size = 2
		}
		scalar(size, wasm.I32)
	case model.Enum:
		disc, err := discriminant(len(t.Labels))
		if err != nil {
			return nil, err
		}
		scalar(disc, wasm.I32)
	case model.Record, model.Tuple:
		if len(t.Fields) == 0 {
			return nil, invalid("empty aggregate")
		}
		if uint64(len(t.Fields)) > uint64(b.limits.MaxNodes-b.nodes) {
			return nil, exhausted()
		}
		b.nodes += uint32(len(t.Fields))
		out.Align = 1
		out.Offsets = make([]uint64, len(t.Fields))
		for i, f := range t.Fields {
			child, err := b.build(f.Type, depth+1)
			if err != nil {
				return nil, err
			}
			offset, err := b.align(out.Size, child.Align)
			if err != nil {
				return nil, err
			}
			out.Offsets[i] = offset
			out.Align = max(out.Align, child.Align)
			if child.Size > b.limits.MaxBytes-offset {
				return nil, exhausted()
			}
			out.Size = offset + child.Size
			if out.FlatCount > math.MaxUint64-child.FlatCount {
				return nil, exhausted()
			}
			out.FlatCount += child.FlatCount
			if out.FlatCount <= MaxFlatParams {
				out.Flat = append(out.Flat, child.Flat...)
			} else {
				out.Flat = nil
			}
		}
		var err error
		out.Size, err = b.align(out.Size, out.Align)
		if err != nil {
			return nil, err
		}
	case model.Variant, model.Option, model.Result:
		cases := t.Cases
		if t.Kind == model.Option {
			cases = []model.Field{{}, {Type: t.Element}}
			if t.Element == nil {
				return nil, invalid("missing option payload")
			}
		}
		if t.Kind == model.Result && len(cases) != 2 {
			return nil, invalid("result case count")
		}
		if uint64(len(cases)) > uint64(b.limits.MaxNodes-b.nodes) {
			return nil, exhausted()
		}
		b.nodes += uint32(len(cases))
		disc, err := discriminant(len(cases))
		if err != nil {
			return nil, err
		}
		out.Align = disc
		payloadAlign, payloadSize := uint64(1), uint64(0)
		var joined []wasm.ValueType
		var payloadFlat uint64
		for _, c := range cases {
			if c.Type == nil {
				continue
			}
			child, err := b.build(c.Type, depth+1)
			if err != nil {
				return nil, err
			}
			payloadAlign = max(payloadAlign, child.Align)
			payloadSize = max(payloadSize, child.Size)
			payloadFlat = max(payloadFlat, child.FlatCount)
			if payloadFlat < MaxFlatParams {
				for i, v := range child.Flat {
					if i == len(joined) {
						joined = append(joined, v)
					} else {
						joined[i] = join(joined[i], v)
					}
				}
			}
		}
		out.Align = max(out.Align, payloadAlign)
		out.PayloadOffset, err = b.align(disc, payloadAlign)
		if err != nil {
			return nil, err
		}
		if payloadSize > b.limits.MaxBytes-out.PayloadOffset {
			return nil, exhausted()
		}
		out.Size, err = b.align(out.PayloadOffset+payloadSize, out.Align)
		if err != nil {
			return nil, err
		}
		out.FlatCount = 1 + payloadFlat
		if out.FlatCount <= MaxFlatParams {
			out.Flat = append([]wasm.ValueType{wasm.I32}, joined...)
		}
	default:
		return nil, invalid("non-value or disabled type")
	}
	if out.Size > b.limits.MaxBytes {
		return nil, exhausted()
	}
	b.cache[t] = out
	return out, nil
}

func (b *Builder) align(size, alignment uint64) (uint64, error) {
	padding := (-size) & (alignment - 1)
	if size > b.limits.MaxBytes || padding > b.limits.MaxBytes-size {
		return 0, exhausted()
	}
	return size + padding, nil
}

func discriminant(cases int) (uint64, error) {
	if cases <= 0 || uint64(cases) > math.MaxUint32 {
		return 0, invalid("case count")
	}
	if cases <= 256 {
		return 1, nil
	}
	if cases <= 65536 {
		return 2, nil
	}
	return 4, nil
}

func join(a, b wasm.ValueType) wasm.ValueType {
	if a == b {
		return a
	}
	if (a == wasm.I32 || a == wasm.F32) && (b == wasm.I32 || b == wasm.F32) {
		return wasm.I32
	}
	return wasm.I64
}

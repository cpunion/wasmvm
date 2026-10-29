package canonical

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLiteralScalarsAndAggregates(t *testing.T) {
	u8 := &model.Type{Kind: model.U8}
	stringType := &model.Type{Kind: model.String}
	cases := []struct {
		typ   *model.Type
		bytes []byte
		want  Value
	}{
		{&model.Type{Kind: model.Bool}, []byte{1}, Value{Bits: 1}},
		{u8, []byte{255}, Value{Bits: 255}},
		{&model.Type{Kind: model.S8}, []byte{128}, Value{Bits: math.MaxUint64 - 127}},
		{&model.Type{Kind: model.U16}, []byte{255, 255, 3}, Value{Bits: 65535}},
		{&model.Type{Kind: model.S16}, []byte{128, 128, 126}, Value{Bits: math.MaxUint64 - 32767}},
		{&model.Type{Kind: model.U32}, []byte{255, 255, 255, 255, 15}, Value{Bits: math.MaxUint32}},
		{&model.Type{Kind: model.S32}, []byte{127}, Value{Bits: math.MaxUint64}},
		{&model.Type{Kind: model.U64}, []byte{128, 1}, Value{Bits: 128}},
		{&model.Type{Kind: model.S64}, []byte{127}, Value{Bits: math.MaxUint64}},
		{&model.Type{Kind: model.F32}, []byte{0, 0, 192, 127}, Value{Bits: 0x7fc00000}},
		{&model.Type{Kind: model.F64}, []byte{0, 0, 0, 0, 0, 0, 248, 127}, Value{Bits: 0x7ff8000000000000}},
		{&model.Type{Kind: model.Char}, []byte("界"), Value{Bits: '界'}},
		{&model.Type{Kind: model.Char}, []byte("🙂"), Value{Bits: '🙂'}},
		{&model.Type{Kind: model.Char}, []byte("a"), Value{Bits: 'a'}},
		{stringType, []byte{3, 0xe7, 0x95, 0x8c}, Value{Text: "界"}},
		{&model.Type{Kind: model.Flags, Labels: []string{"a", "b"}}, []byte{255}, Value{Bits: 3}},
		{&model.Type{Kind: model.Enum, Labels: []string{"a", "b"}}, []byte{1}, Value{Bits: 1}},
		{&model.Type{Kind: model.Record, Fields: []model.Field{{Type: u8}, {Type: stringType}}}, []byte{42, 1, 'a'}, Value{Elements: []Value{{Bits: 42}, {Text: "a"}}}},
		{&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: u8}}}, []byte{42}, Value{Elements: []Value{{Bits: 42}}}},
		{&model.Type{Kind: model.Variant, Cases: []model.Field{{}, {Type: u8}}}, []byte{1, 42}, Value{Case: 1, Elements: []Value{{Bits: 42}}}},
		{&model.Type{Kind: model.Option, Element: u8}, []byte{0}, Value{}},
		{&model.Type{Kind: model.Option, Element: u8}, []byte{1, 42}, Value{Case: 1, Elements: []Value{{Bits: 42}}}},
		{&model.Type{Kind: model.Result, Cases: []model.Field{{}, {Type: u8}}}, []byte{1, 42}, Value{Case: 1, Elements: []Value{{Bits: 42}}}},
		{&model.Type{Kind: model.FixedList, Element: u8, Length: 2}, []byte{1, 2}, Value{Elements: []Value{{Bits: 1}, {Bits: 2}}}},
		{&model.Type{Kind: model.List, Element: u8}, []byte{2, 1, 2}, Value{Elements: []Value{{Bits: 1}, {Bits: 2}}}},
		// Maps preserve order and duplicate keys, as the pinned component spec requires.
		{&model.Type{Kind: model.Map, Key: u8, Element: u8}, []byte{2, 1, 2, 1, 3}, Value{Elements: []Value{{Elements: []Value{{Bits: 1}, {Bits: 2}}}, {Elements: []Value{{Bits: 1}, {Bits: 3}}}}}},
	}
	for _, tc := range cases {
		got, err := DecodeLiteral(tc.typ, tc.bytes, DefaultLimits())
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("kind %#x: got %#v, %v; want %#v", tc.typ.Kind, got, err, tc.want)
		}
		for i := range tc.bytes {
			if _, err := DecodeLiteral(tc.typ, tc.bytes[:i], DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatalf("kind %#x prefix %d: %v", tc.typ.Kind, i, err)
			}
		}
		if _, err := DecodeLiteral(tc.typ, append(append([]byte(nil), tc.bytes...), 0), DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("trailing literal bytes accepted", tc.typ.Kind, err)
		}
	}
}

func TestInvalidLiteralsAndUninhabitedBranches(t *testing.T) {
	for _, tc := range []struct {
		kind model.Kind
		data []byte
	}{
		{model.Bool, []byte{2}},
		{model.U16, []byte{128, 128, 4}},
		{model.U16, []byte{128, 128, 128, 0}},
		{model.S16, []byte{128, 128, 2}},
		{model.S32, []byte{128, 128, 128, 128, 128, 0}},
		{model.F32, []byte{1, 0, 192, 127}},
		{model.F64, []byte{0, 0, 0, 0, 0, 0, 248, 255}},
		{model.Char, []byte{0x80}},
		{model.Char, []byte{0xed, 0xa0, 0x80}},
		{model.String, []byte{1, 0xff}},
		{model.Enum, []byte{0}},
		{model.Variant, []byte{1}},
	} {
		if _, err := DecodeLiteral(&model.Type{Kind: tc.kind}, tc.data, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(tc.kind, err)
		}
	}
	for _, kind := range []model.Kind{model.Own, model.Borrow, model.Future, model.Stream, model.ErrorContext} {
		// Even the absent branch cannot conceal a non-constructible type.
		v := &model.Type{Kind: model.Option, Element: &model.Type{Kind: kind}}
		if _, err := DecodeLiteral(v, []byte{0}, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(kind, err)
		}
	}
}

func TestLiteralAllocationAndDepthLimits(t *testing.T) {
	u8 := &model.Type{Kind: model.U8}
	list := &model.Type{Kind: model.List, Element: u8}
	for _, limits := range []Limits{{MaxBytes: 1}, {MaxNodes: 10, MaxDepth: 1, MaxBytes: 1000}, {MaxNodes: 10, MaxDepth: 10, MaxBytes: 63}} {
		if _, err := DecodeLiteral(list, []byte{1, 42}, limits); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(limits, err)
		}
	}
	if _, err := DecodeLiteral(list, []byte{255, 255, 255, 255, 15}, DefaultLimits()); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("unbounded vector allocation", err)
	}
	shared := u8
	for i := 0; i < 24; i++ {
		shared = &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: shared}, {Type: shared}}}
	}
	if _, err := DecodeLiteral(shared, nil, Limits{MaxNodes: 64, MaxDepth: 16, MaxBytes: 1024}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("type depth escaped limit", err)
	}
}

func FuzzLiteral(f *testing.F) {
	f.Add([]byte{1, 3, 'a', 'b', 'c'})
	f.Add([]byte{0})
	t := &model.Type{Kind: model.List, Element: &model.Type{Kind: model.String}}
	f.Fuzz(func(test *testing.T, bytes []byte) {
		v, err := DecodeLiteral(t, bytes, Limits{MaxNodes: 1024, MaxDepth: 16, MaxBytes: 1 << 16})
		if err == nil && len(v.Elements) > 1024 {
			test.Fatal("literal allocation escaped quota")
		}
	})
}

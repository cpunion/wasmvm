package canonical

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPinnedCollectionMemoryOracle(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/component-canonical/collection-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Type     *model.Type
		Memory64 bool
		Value    Value
		Memory   string
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 56 {
		t.Fatal("incomplete collection oracle", len(cases))
	}
	for i, tc := range cases {
		want, err := hex.DecodeString(tc.Memory)
		if err != nil {
			t.Fatal(err)
		}
		c := valueCodec(tc.Memory64)
		if err := c.Store(tc.Type, 0, tc.Value); err != nil {
			t.Fatal(i, err)
		}
		actual := make([]byte, len(want))
		if err := c.Transfer.Memory.Read(0, actual); err != nil || !reflect.DeepEqual(actual, want) {
			t.Fatal("reference bytes", i, err)
		}
		// Lift from the unchanged reference's bytes, independently of Store.
		c = valueCodec(tc.Memory64)
		if err := c.Transfer.Memory.Write(0, want); err != nil {
			t.Fatal(err)
		}
		got, err := c.Load(tc.Type, 0)
		if err != nil || !reflect.DeepEqual(expandPackedBytes(tc.Type, got), tc.Value) {
			t.Fatal("reference lift", i, got, err)
		}
	}
}

func TestCollectionMemoryAndFlatTransfers(t *testing.T) {
	mapType := &model.Type{Kind: model.Map, Key: typ(model.String), Element: typ(model.U64)}
	for _, wide := range []bool{false, true} {
		for _, encoding := range []Encoding{UTF8, UTF16, Latin1UTF16} {
			for _, tc := range []struct {
				typ   *model.Type
				value Value
			}{
				{mapType, Value{Elements: []Value{{Elements: []Value{{Text: "同"}, {Bits: 1}}}, {Elements: []Value{{Text: "同"}, {Bits: 2}}}}}},
				{mapType, Value{Elements: []Value{}}},
				{&model.Type{Kind: model.FixedList, Element: typ(model.U16), Length: 3}, Value{Elements: []Value{{Bits: 1}, {Bits: 2}, {Bits: 65535}}}},
				{&model.Type{Kind: model.FixedList, Element: typ(model.String), Length: 2}, Value{Elements: []Value{{Text: "hello"}, {Text: "界🙂"}}}},
			} {
				for _, flat := range []bool{false, true} {
					c := valueCodec(wide)
					c.Transfer.Encoding = encoding
					var got Value
					var err error
					if flat {
						slots, e := c.LowerFlat(tc.typ, tc.value)
						if e != nil {
							t.Fatal(e)
						}
						got, err = c.LiftFlat(tc.typ, slots)
					} else {
						if err := c.Store(tc.typ, 0, tc.value); err != nil {
							t.Fatal(err)
						}
						got, err = c.Load(tc.typ, 0)
					}
					if err != nil || !reflect.DeepEqual(got, tc.value) {
						t.Fatal(wide, encoding, flat, tc.typ.Kind, got, err)
					}
				}
			}
		}
	}
	// Reusing a map never creates a new synthetic entry schema per transfer.
	b := NewBuilder(false, DefaultLimits())
	l, err := b.Build(mapType)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		c := valueCodec(false)
		c.Layouts = b
		if _, err := c.LowerFlat(mapType, Value{Elements: []Value{{Elements: []Value{{Text: "key"}, {Bits: 1}}}}}); err != nil {
			t.Fatal(err)
		}
		again, _ := b.Build(mapType)
		if again.Entry != l.Entry {
			t.Fatal("map entry metadata was recreated")
		}
	}
}

func TestCollectionLimitsAndInvalidValues(t *testing.T) {
	fixed := &model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: 17}
	value := Value{Elements: make([]Value, 17)}
	c := valueCodec(false)
	if err := c.Store(fixed, 0, value); err != nil {
		t.Fatal(err)
	}
	if got, err := c.Load(fixed, 0); err != nil || !reflect.DeepEqual(expandPackedBytes(fixed, got), value) {
		t.Fatal(got, err)
	}
	if _, err := c.LowerFlat(fixed, value); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := c.Store(fixed, 0, Value{Elements: value.Elements[:16]}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	mapType := &model.Type{Kind: model.Map, Key: typ(model.U8), Element: typ(model.U32)}
	c = valueCodec(false)
	if err := c.Store(mapType, 0, Value{Elements: []Value{{}}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := c.storePair(0, 4096, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Load(mapType, 0); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := c.storePair(0, 0, 0xffffffff); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Load(mapType, 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.MaxValues = 1
	if _, err := c.LiftFlat(&model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: 2}, []Slot{{Type: wasm.I32}, {Type: wasm.I32}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

package canonical

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Preserve every existing oracle's element-wise shape while comparing the new
// host representation. Expected values and independent reference bytes stay
// unchanged, including all nested fixed-list/map cases.
func expandPackedBytes(t *model.Type, v Value) Value {
	if byteSequence(t) && v.Elements == nil {
		v.Elements = make([]Value, len(v.Text))
		for i := 0; i < len(v.Text); i++ {
			v.Elements[i].Bits = uint64(v.Text[i])
		}
		v.Text = ""
		return v
	}
	if len(v.Elements) == 0 {
		return v
	}
	v.Elements = append([]Value(nil), v.Elements...)
	for i := range v.Elements {
		switch t.Kind {
		case model.Tuple, model.Record:
			v.Elements[i] = expandPackedBytes(t.Fields[i].Type, v.Elements[i])
		case model.List, model.FixedList:
			v.Elements[i] = expandPackedBytes(t.Element, v.Elements[i])
		case model.Variant, model.Result, model.Option:
			v.Elements[i] = expandPackedBytes(cases(t)[v.Case].Type, v.Elements[i])
		case model.Map:
			entry := v.Elements[i]
			entry.Elements = []Value{expandPackedBytes(t.Key, entry.Elements[0]), expandPackedBytes(t.Element, entry.Elements[1])}
			v.Elements[i] = entry
		}
	}
	return v
}

func TestPackedByteOwnershipAndLegacyInputs(t *testing.T) {
	if unsafe.Sizeof(Value{}) != 64 {
		t.Fatal("packed storage enlarged every host value")
	}
	data := []byte{0, 255, 128, 1}
	v := ByteList(data)
	data[0] = 99
	if v.Text != "\x00\xff\x80\x01" || v.Elements != nil {
		t.Fatal("packed list aliases its mutable input", v)
	}
	legacy := Value{Elements: []Value{{Bits: 0}, {Bits: 255}, {Bits: 128}, {Bits: 1}}}
	for _, value := range []Value{v, legacy} {
		s, err := value.ByteString()
		if err != nil || s != v.Text {
			t.Fatal(s, err)
		}
	}
	for _, value := range []Value{{Text: "x", Elements: []Value{}}, {Elements: []Value{{Bits: 256}}}} {
		if _, err := value.ByteString(); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, wide := range []bool{false, true} {
		for _, fixed := range []bool{false, true} {
			typ := &model.Type{Kind: model.List, Element: typ(model.U8)}
			if fixed {
				typ.Kind, typ.Length = model.FixedList, 4
			}
			for _, source := range []Value{v, legacy} {
				for _, flat := range []bool{false, true} {
					c := valueCodec(wide)
					var got Value
					var err error
					if flat {
						slots, e := c.LowerFlat(typ, source)
						if e != nil {
							t.Fatal(e)
						}
						got, err = c.LiftFlat(typ, slots)
					} else {
						if err := c.Store(typ, 0, source); err != nil {
							t.Fatal(err)
						}
						got, err = c.Load(typ, 0)
					}
					if err != nil || !reflect.DeepEqual(got, v) {
						t.Fatal(wide, fixed, flat, got, err)
					}
					if err := c.Transfer.Memory.Write(64, []byte{7, 7, 7, 7}); err != nil {
						t.Fatal(err)
					}
					if got.Text != v.Text {
						t.Fatal("result retained guest memory")
					}
					copy, err := valueCodec(wide).Clone(typ, got)
					if err != nil || !reflect.DeepEqual(copy, v) {
						t.Fatal(copy, err)
					}
				}
			}
		}
	}
}

func TestPackedByteLimitsAndFailures(t *testing.T) {
	typ := &model.Type{Kind: model.List, Element: typ(model.U8)}
	for _, wide := range []bool{false, true} {
		for _, empty := range []bool{false, true} {
			data := "\xff\x00"
			if empty {
				data = ""
			}
			c := valueCodec(wide)
			if err := c.Store(typ, 0, Value{Text: data}); err != nil {
				t.Fatal(err)
			}
			got, err := c.Load(typ, 0)
			if err != nil || got.Text != data || got.Elements != nil {
				t.Fatal(got, err)
			}
		}
	}
	for _, mutate := range []func(*Codec){
		func(c *Codec) { c.MaxValues = 2 },
		func(c *Codec) { c.Layouts.limits.MaxDepth = 1 },
		func(c *Codec) { c.Transfer.MaxBytes = 1 },
	} {
		c := valueCodec(false)
		mutate(c)
		if _, err := c.Clone(typ, Value{Text: "xx"}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		c = valueCodec(false)
		c.storePair(0, 64, 2)
		mutate(c)
		if _, err := c.Load(typ, 0); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	for _, source := range []Value{{Text: "x", Elements: []Value{}}, {Text: "x", Elements: []Value{{Bits: 1}}}} {
		if err := valueCodec(false).Store(typ, 0, source); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	fixed := &model.Type{Kind: model.FixedList, Element: typ.Element, Length: 2}
	if err := valueCodec(false).Store(fixed, 0, Value{Text: "x"}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c := valueCodec(false)
	c.Transfer.Realloc = nil
	if _, err := c.LowerFlat(typ, Value{Text: "x"}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 4096, nil }
	if _, err := c.LowerFlat(typ, Value{Text: "x"}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = 1
	if _, err := c.LowerFlat(typ, Value{Text: "xx"}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c = valueCodec(false)
	c.Transfer.Context = ctx
	c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { cancel(); return 64, nil }
	if _, err := c.LowerFlat(typ, Value{Text: "x"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	// Large bytes use their byte size, not one 64-byte Value slot per byte.
	c = valueCodec(false)
	c.MaxValues = 2001
	c.Transfer.MaxBytes = 4000
	if err := c.Store(typ, 0, ByteList(bytes.Repeat([]byte{255}, 2000))); err != nil {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.MaxValues = 2001
	c.Transfer.MaxBytes = 4000
	c.Transfer.Memory.Write(64, bytes.Repeat([]byte{255}, 2000))
	c.storePair(0, 64, 2000)
	if got, err := c.Load(typ, 0); err != nil || got.Text != strings.Repeat("\xff", 2000) {
		t.Fatal(err)
	}
}

type byteFaultMemory struct {
	Memory
	read, write error
	cancel      context.CancelFunc
}

func (m byteFaultMemory) Read(at uint64, data []byte) error {
	if at == 64 && len(data) != 0 {
		if m.read != nil {
			return m.read
		}
		if err := m.Memory.Read(at, data); err != nil {
			return err
		}
		if m.cancel != nil {
			m.cancel()
		}
		return nil
	}
	return m.Memory.Read(at, data)
}

func (m byteFaultMemory) Write(at uint64, data []byte) error {
	if m.write != nil {
		return m.write
	}
	return m.Memory.Write(at, data)
}

func TestPackedBytesPreserveMemoryFailuresAndFlatTruncation(t *testing.T) {
	list := &model.Type{Kind: model.List, Element: typ(model.U8)}
	sentinel := errors.New("memory access failed")
	c := valueCodec(false)
	c.storePair(0, 64, 1)
	c.Transfer.Memory = byteFaultMemory{Memory: c.Transfer.Memory, read: sentinel}
	if _, err := c.Load(list, 0); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Memory = byteFaultMemory{Memory: c.Transfer.Memory, write: sentinel}
	if _, err := c.LowerFlat(list, Value{Text: "x"}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 0, sentinel }
	if _, err := c.LowerFlat(list, Value{Text: "x"}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = 1
	if _, err := c.LowerFlat(list, Value{Text: "x"}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = 0
	if _, err := c.Clone(list, Value{Text: "x"}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.storePair(0, 64, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Transfer.Context = ctx
	c.Transfer.Memory = byteFaultMemory{Memory: c.Transfer.Memory, cancel: cancel}
	if v, err := c.Load(list, 0); !errors.Is(err, context.Canceled) || v.Text != "" {
		t.Fatal(v, err)
	}
	c = valueCodec(false)
	fixed := &model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: 1}
	if v, err := c.LiftFlat(fixed, []Slot{{Type: wasm.I32, Bits: 511}}); err != nil || v.Text != "\xff" {
		t.Fatal(v, err)
	}
}

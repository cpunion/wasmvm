package slots

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSourceImagePreservesMetadataAndIndependentReaders(t *testing.T) {
	source := ir.Source{Wasm: module.Instr{Op: 0xe6, Offset: 1 << 40, A: 3, B: 7, Memory: 2, MemoryOffsetHigh: 9, Block: -64, Bits: 0xfff8000000001234, Vector: [16]byte{0xff, 3, 7}, Lane: 15, Labels: []uint32{3, 8, 0}, Types: []wasm.ValueType{wasm.I64, wasm.FuncRef}, Catches: []module.Catch{{Kind: 1, Tag: 3, Label: 2}}, On: []module.Catch{{Kind: 0, Tag: 7, Label: 9}}}, Inputs: []wasm.ValueType{wasm.V128, wasm.I32}, Outputs: []wasm.ValueType{wasm.F64}, Effects: ir.MayCallHost | ir.MayTrap}
	f := &Function{Sources: []ir.Source{source, {Wasm: module.Instr{Op: 0xb, Offset: 77}}}}
	if a, b := f.SourceIndices(0); a != 3 || b != 7 {
		t.Fatal("rich source indices", a, b)
	}
	var before ir.Source
	if got := f.ReadSource(0, &before); got != &f.Sources[0] || f.SourceCount() != 2 || f.SourceOffset(0) != source.Wasm.Offset || f.SourceOutputs(0) != 1 {
		t.Fatal("rich source view changed")
	}
	if err := f.FreezeSources(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.Sources != nil || f.SourceCount() != 2 || f.SourceOffset(0) != source.Wasm.Offset || f.SourceOutputs(0) != 1 {
		t.Fatal("source image metadata")
	}
	if a, b := f.SourceIndices(0); a != 3 || b != 7 {
		t.Fatal("packed source indices", a, b)
	}
	var first, second ir.Source
	if got := f.ReadSource(0, &first); got != &first || !reflect.DeepEqual(first, source) || first.Fuel() != 1 {
		t.Fatal(first)
	}
	f.ReadSource(1, &second)
	if second.Fuel() != 0 || second.Wasm.Offset != 77 || len(second.Wasm.Types) != 0 || !reflect.DeepEqual(first, source) {
		t.Fatal("independent execution scratch was overwritten", first, second)
	}
	source.Wasm.Labels[0] = 99
	source.Wasm.Types[0] = wasm.F32
	source.Wasm.Catches[0].Tag = 99
	source.Wasm.On[0].Tag = 99
	source.Inputs[0] = wasm.I64
	source.Outputs[0] = wasm.I32
	f.ReadSource(0, &second)
	if second.Wasm.Labels[0] != 3 || second.Wasm.Types[0] != wasm.I64 || second.Wasm.Catches[0].Tag != 3 || second.Wasm.On[0].Tag != 7 || second.Inputs[0] != wasm.V128 || second.Outputs[0] != wasm.F64 {
		t.Fatal("image retained compiler slices", second)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := f.FreezeSources(ctx); err != nil {
		t.Fatal("already published image changed", err)
	}
}

func TestSourceImageCanceledPublicationAndPointerFreeStorage(t *testing.T) {
	f := &Function{Sources: []ir.Source{{Wasm: module.Instr{Op: 0x41, Bits: 42}}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := f.FreezeSources(ctx); !errors.Is(err, context.Canceled) || f.image != nil || len(f.Sources) != 1 {
		t.Fatal("partial image published", err)
	}
	var pointerFree func(reflect.Type) bool
	pointerFree = func(typ reflect.Type) bool {
		switch typ.Kind() {
		case reflect.Array:
			return pointerFree(typ.Elem())
		case reflect.Struct:
			for i := 0; i < typ.NumField(); i++ {
				if !pointerFree(typ.Field(i).Type) {
					return false
				}
			}
			return true
		case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.String, reflect.UnsafePointer:
			return false
		}
		return true
	}
	for _, typ := range []reflect.Type{reflect.TypeOf(sourceRecord{}), reflect.TypeOf(uint32(0)), reflect.TypeOf(wasm.I32), reflect.TypeOf(module.Catch{})} {
		if !pointerFree(typ) {
			t.Fatal("runtime source array contains pointers", typ)
		}
	}
}

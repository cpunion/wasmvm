package state

import (
	"errors"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSharedHeadersAliasNativeViewsAndRetainPortablePayloads(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 8 && (unsafe.Sizeof(heapObject{}) != 32 ||
		unsafe.Offsetof(heapObject{}.Ref) != 0 || unsafe.Offsetof(heapObject{}.Data) != 8 ||
		unsafe.Offsetof(heapObject{}.Length) != 16 || unsafe.Offsetof(heapObject{}.Type) != 20 ||
		unsafe.Offsetof(heapObject{}.Serial) != 24) {
		t.Fatal("shared header ABI")
	}
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		s := New(DefaultLimits())
		if offheap {
			s = NewNative(DefaultLimits())
		}
		t.Cleanup(s.Close)
		types, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
			Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8},
				{Value: wasm.I32, Packed: 16}, {Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.V128}},
		}}}, s.limits.MaxTypes)
		if !ok {
			t.Fatal("type registry")
		}
		values := []value.Value{{Lo: 0x81}, {Lo: 0x8081}, {Lo: 0x87654321},
			{Lo: 0xfedcba9876543210}, {Lo: 0xabcdef8765432101, Hi: 0x123456789abcdef0}}
		var refs []uint64
		var headers []*heapObject
		for range 1100 {
			ref, err := s.allocateObject(types[0], uint32(len(values)))
			if err != nil {
				t.Fatal(err)
			}
			o, _ := s.heapObject(ref)
			for i, v := range values {
				s.setObject(o, uint32(i), v)
			}
			refs, headers = append(refs, ref), append(headers, o)
		}
		runtime.GC()
		for i, ref := range refs {
			o, ok := s.heapObject(ref)
			if !ok || o != headers[i] || o.Ref != ref || o.Type != types[0] || o.Serial != uint64(i+1) || s.objectBytes(o) != 95 {
				t.Fatal("header identity/accounting changed", offheap, i)
			}
			if offheap && unsafe.Pointer(o) != unsafe.Pointer(&s.nativeHeap.views[uint64(uint32(ref))*4]) {
				t.Fatal("native header duplicates its view", i)
			}
			for j, want := range values {
				if got := s.getObject(o, uint32(j)); got != want {
					t.Fatal("host collection/growth lost payload bits", offheap, i, j, got, want)
				}
			}
		}
		s.CollectReferences(nil, nil, nil)
		if s.Usage().HeapObjects != 0 || s.Usage().HeapBytes != 0 {
			t.Fatal("released header retained quota")
		}
	}
}

func TestNativeSharedHeaderRejectsChangedReservationAndEpoch(t *testing.T) {
	for _, failure := range []string{"generation", "epoch"} {
		s, types := allocationStore(t)
		if _, err := s.PrepareNativeAllocations(3); err != nil {
			t.Fatal(err)
		}
		ref := stagedAllocation(t, s, 1, 0, 0)
		if failure == "generation" {
			s.heap[uint32(ref)].Generation++
		} else {
			s.nativeHeap.views[uint64(uint32(ref))*4+3] = 1
		}
		if err := s.FinishNativeAllocations(types); !errors.Is(err, wasm.ErrInvalid) || s.heapSerial != 0 || s.Usage().HeapObjects != 0 {
			t.Fatal("invalid shared header published", failure, err, s.Usage())
		}
		if _, err := s.PrepareNativeAllocations(3); err != nil {
			t.Fatal(err)
		}
		ref = stagedAllocation(t, s, 1, 0, 0)
		if err := s.FinishNativeAllocations(types); err != nil {
			t.Fatal("failed header poisoned reuse", failure, err)
		}
		o, ok := s.heapObject(ref)
		if !ok || o.Ref != ref || o.Type != types[1] || o.Serial != 1 || s.nativeHeap.views[uint64(uint32(ref))*4+2] != uint64(types[1])<<32 {
			t.Fatal("local type/epoch did not publish in shared header", failure)
		}
	}
}

func TestSharedHeaderFieldAccessRejectsArrayBounds(t *testing.T) {
	s := New(DefaultLimits())
	defer s.Close()
	types, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
		Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}},
	}}}, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	ref, err := s.allocateObject(types[0], 1)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := s.heapObject(ref)
	for _, index := range []uint32{1, ^uint32(0)} {
		for _, write := range []bool{false, true} {
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("unchecked field access escaped payload", index, write)
					}
				}()
				if write {
					s.setObject(o, index, value.Value{Lo: 1})
				} else {
					s.getObject(o, index)
				}
			}()
		}
	}
}

package vm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostObjectsAndReferenceRoots(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		mem, err := s.NewMemory(MemoryType{Min: 1, Max: 2, HasMax: true})
		if err != nil {
			t.Fatal(err)
		}
		if err = mem.WriteMemory(65534, []byte{1, 2}); err != nil {
			t.Fatal(err)
		}
		if old, ok, err := mem.GrowMemory(1); err != nil || !ok || old != 1 {
			t.Fatal(old, ok, err)
		}
		if pages, err := mem.MemoryPages(); err != nil || pages != 2 {
			t.Fatal(pages, err)
		}
		data := make([]byte, 4)
		if err = mem.ReadMemory(65534, data); err != nil || !reflect.DeepEqual(data, []byte{1, 2, 0, 0}) {
			t.Fatal(data, err)
		}
		if _, ok, err := mem.GrowMemory(1); err != nil || ok {
			t.Fatal(ok, err)
		}
		if err = mem.ReadMemory(131072, []byte{0}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if err = mem.WriteMemory(131072, []byte{0}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if _, err = s.NewMemory(MemoryType{Min: 2, Max: 1, HasMax: true}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		if _, err = s.NewTable(TableType{Element: wasm.I32}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		if _, err = s.NewGlobal(GlobalType{Value: wasm.I64}, I32(1)); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		object, _ := s.NewExternRef("retained")
		table, err := s.NewTable(TableType{Element: wasm.ExternRef, Limits: MemoryType{Min: 1, Max: 2, HasMax: true}})
		if err != nil {
			t.Fatal(err)
		}
		global, err := s.NewGlobal(GlobalType{Value: wasm.ExternRef, Mutable: true}, object)
		if err != nil {
			t.Fatal(err)
		}
		immutable, _ := s.NewGlobal(GlobalType{Value: wasm.I32}, I32(3))
		if err = immutable.GlobalSet(I32(4)); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		if err = table.TableSet(0, object); err != nil {
			t.Fatal(err)
		}
		if old, ok, err := table.GrowTable(1, object); err != nil || !ok || old != 1 {
			t.Fatal(old, ok, err)
		}
		if _, ok, err := table.GrowTable(1, object); err != nil || ok {
			t.Fatal(ok, err)
		}
		if size, err := table.TableSize(); err != nil || size != 2 {
			t.Fatal(size, err)
		}
		object.Release()
		s.Collect()
		for _, get := range []func() (Value, error){func() (Value, error) { return table.TableGet(1) }, global.GlobalGet} {
			v, err := get()
			if err != nil {
				t.Fatal(err)
			}
			if got, err := s.ExternRef(v); err != nil || got != "retained" {
				t.Fatal(got, err)
			}
			v.Release()
		}
		if _, err = table.TableGet(2); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if err = table.TableSet(2, I32(1)); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if err = table.TableSet(0, object); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		if err = global.GlobalSet(object); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		if _, _, err = table.GrowTable(0, object); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		null, _ := NullRef(wasm.ExternRef)
		table.TableSet(0, null)
		table.TableSet(1, null)
		global.GlobalSet(null)
		s.Collect()
		u, _ := s.Usage()
		if u.ExternRefs != 0 || u.HostRoots != 0 {
			t.Fatal(u)
		}
		if _, err := table.Reference(); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		if _, err := table.FunctionType(); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		if _, err := table.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		if err := table.ReadMemory(0, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		s.Close()
		checks := []func() error{
			func() error { _, e := s.NewMemory(MemoryType{}); return e }, func() error { _, e := s.NewTable(TableType{}); return e }, func() error { _, e := s.NewGlobal(GlobalType{}, Value{}); return e },
			func() error { _, e := s.NewFunction(FunctionType{}, nil); return e }, func() error { _, e := table.Reference(); return e }, func() error { _, e := table.FunctionType(); return e },
			func() error { return mem.ReadMemory(0, nil) }, func() error { return mem.WriteMemory(0, nil) }, func() error { _, e := mem.MemoryPages(); return e }, func() error { _, _, e := mem.GrowMemory(0); return e },
			func() error { _, e := table.TableSize(); return e }, func() error { _, e := table.TableGet(0); return e }, func() error { return table.TableSet(0, null) }, func() error { _, _, e := table.GrowTable(0, null); return e },
			func() error { _, e := global.GlobalGet(); return e }, func() error { return global.GlobalSet(null) },
		}
		for i, check := range checks {
			if err := check(); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(i, err)
			}
		}
	})
}

func TestFailedStartEscapedFunctionAndStaleRoot(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		tab, _ := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Min: 1}})
		l.Define("m", "t", tab)
		d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}, {}}, Imports: []module.Import{{Module: "m", Name: "t", Kind: 1, Table: tab.external.Table.Type}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}}, {Type: 1, Body: module.Expr{{Op: 0}, {Op: 0xb}}}}, Elements: []module.Element{{Type: wasm.FuncRef, Functions: []uint32{0}, Offset: module.Expr{{Op: 0x41}, {Op: 0xb}}}}, HasStart: true, Start: 1}
		compiled := compileDefinition(t, e, d)
		info, err := compiled.Info()
		if err != nil || len(info.Imports) != 1 || !info.HasStart {
			t.Fatal(info, err)
		}
		if _, err = l.Instantiate(nil, compiled, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		u, _ := s.Usage()
		if u.Core.Instances != 1 {
			t.Fatal("lost escaped function", u)
		}
		ref, err := tab.TableGet(0)
		if err != nil {
			t.Fatal(err)
		}
		null, _ := NullRef(wasm.FuncRef)
		tab.TableSet(0, null)
		s.Collect()
		result, err := s.CallRef(nil, ref, nil, CallOptions{})
		if err != nil || bits(t, result.Values[0]) != 42 {
			t.Fatal(result, err)
		}
		ref.Release()
		s.Collect()
		u, _ = s.Usage()
		if u.Core.Instances != 0 || u.Core.Functions != 0 {
			t.Fatal(u)
		}
		l.Instantiate(nil, compiled, CallOptions{}) // Reuse the internal identity suffix.
		if _, err := s.CallRef(nil, ref, nil, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal("released root revived", err)
		}
		if _, err := s.CallRef(nil, null, nil, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		// Segment failure also reclaims a module with no escaped references.
		d = &module.Module{Memories: []module.Limits{{Min: 0}}, Data: []module.Data{{Bytes: []byte{1}, Offset: module.Expr{{Op: 0x41}, {Op: 0xb}}}}}
		before, _ := s.Usage()
		if _, err := l.Instantiate(nil, compileDefinition(t, e, d), CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		after, _ := s.Usage()
		if before != after {
			t.Fatal(before, after)
		}
	})
}

func TestLinkTypeAndZeroHandleErrors(t *testing.T) {
	e, s, l := setup(t)
	_, other, _ := setup(t)
	fn, _ := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.I32}}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
	foreign, _ := other.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
	if err := l.Define("h", "f", foreign); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if err := l.Define("h", "f", fn); err != nil {
		t.Fatal(err)
	}
	if err := l.Define("h", "f", fn); !errors.Is(err, ErrLink) {
		t.Fatal(err)
	}
	if _, err := s.Call(nil, foreign, nil, CallOptions{}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	for _, args := range [][]Value{nil, {I64(1)}, {Value{}}} {
		if _, err := fn.Call(nil, args, CallOptions{}); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
	}
	if _, err := s.NewFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	if _, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{0}}, func(*Caller, []Value) ([]Value, error) { return nil, nil }); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "f", Kind: 0}}}
	if _, err := l.Instantiate(nil, compileDefinition(t, e, d), CallOptions{}); !errors.Is(err, ErrLink) {
		t.Fatal(err)
	}
	var zero Extern
	if _, err := zero.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := zero.FunctionType(); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := zero.Reference(); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if err := zero.ReadMemory(0, nil); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	var in *Instance
	if _, err := in.Export(""); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	s.Close()
	if err := l.Define("h", "g", fn); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := l.Instantiate(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

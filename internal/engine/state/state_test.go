package state

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func build(t *testing.T, m *module.Module) *slots.Module {
	t.Helper()
	v, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	code, err := slots.Compile(ir, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return code
}
func expression(op uint32, bits uint64) module.Expr {
	return module.Expr{{Op: op, Bits: bits}, {Op: 0xb}}
}
func TestOwnershipLimitsAndGrowth(t *testing.T) {
	s := New(Limits{MaxTypes: 2, MaxFunctions: 1, MaxGlobals: 1, MaxMemories: 2, MaxTables: 2, MaxMemoryPages: 3, MaxTableElements: 4})
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}}
	host, err := s.NewHost(sig, "host")
	if err != nil {
		t.Fatal(err)
	}
	sig.Params[0] = wasm.F64
	f, ok := s.Function(host.Function)
	if !ok || f.Signature.Params[0] != wasm.I32 {
		t.Fatal("borrowed mutable host signature")
	}
	for _, id := range []uint64{0, 100} {
		if _, ok := s.Function(id); ok {
			t.Fatal("bad function handle")
		}
	}
	if _, err := s.NewHost(module.FuncType{}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if s.TypeID(module.FuncType{}) == 0 || s.TypeID(module.FuncType{}) != s.TypeID(module.FuncType{}) || s.TypeID(module.FuncType{Results: []wasm.ValueType{wasm.F64}}) != 0 {
		t.Fatal("type quota/identity")
	}
	if _, err := New(DefaultLimits()).NewHost(module.FuncType{Params: []wasm.ValueType{0}}, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := New(Limits{MaxFunctions: 1}).NewHost(module.FuncType{}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.NewGlobal(module.GlobalType{}, value.Value{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	g, err := s.NewGlobal(module.GlobalType{Value: wasm.I32}, value.Value{Lo: 7})
	if err != nil || g.Global.Value.Lo != 7 {
		t.Fatal(g, err)
	}
	if _, err := s.NewGlobal(g.Global.Type, value.Value{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.NewMemory(module.Limits{Min: 4}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.NewMemory(module.Limits{Min: 2, Max: 1, HasMax: true}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	mem, err := s.NewMemory(module.Limits{Min: 1, Max: 3, HasMax: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := New(DefaultLimits()).GrowMemory(mem.Memory, 1); ok {
		t.Fatal("foreign growth")
	}
	if old, ok := s.GrowMemory(mem.Memory, 2); !ok || old != 1 {
		t.Fatal(old, ok)
	}
	if _, ok := s.GrowMemory(mem.Memory, 1); ok {
		t.Fatal("memory quota")
	}
	if _, err := s.NewTable(module.TableType{Element: wasm.I32}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	typ := module.TableType{Element: wasm.ExternRef, Limits: module.Limits{Min: 1, Max: 3, HasMax: true}}
	tab, err := s.NewTable(typ)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 5}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if old, ok := s.GrowTable(tab.Table, 0, 4); !ok || old != 1 {
		t.Fatal(old, ok)
	}
	if old, ok := s.GrowTable(tab.Table, 2, 7); !ok || old != 1 || !reflect.DeepEqual(tab.Table.Values, []uint64{0, 7, 7}) {
		t.Fatal(old, ok, tab.Table.Values)
	}
	if _, ok := s.GrowTable(tab.Table, 1, 0); ok {
		t.Fatal("declared table max")
	}
	if _, ok := New(DefaultLimits()).GrowTable(tab.Table, 0, 0); ok {
		t.Fatal("foreign table")
	}
	if !tab.Table.Fill(0, 3, 8) || tab.Table.Fill(2, 2, 9) || !reflect.DeepEqual(tab.Table.Values, []uint64{8, 8, 8}) {
		t.Fatal(tab.Table.Values)
	}
	if !tab.Table.Init(1, 0, 2, []uint64{4, 5}) || tab.Table.Init(0, 1, 2, []uint64{4, 5}) {
		t.Fatal("table.init bounds")
	}
	if !tab.Table.Init(0, 1, 2, tab.Table.Values) || !reflect.DeepEqual(tab.Table.Values, []uint64{4, 5, 5}) {
		t.Fatal("table overlap", tab.Table.Values)
	}
	if s.Usage().MemoryPages != 3 || s.Usage().TableElements != 3 {
		t.Fatal(s.Usage())
	}
}

func TestAllocateInitializeAndRoots(t *testing.T) {
	s := New(DefaultLimits())
	host, _ := s.NewHost(module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, nil)
	table, _ := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 2, Max: 5, HasMax: true}})
	memory, _ := s.NewMemory(module.Limits{Min: 1, Max: 3, HasMax: true})
	global, _ := s.NewGlobal(module.GlobalType{Value: wasm.I32}, value.Value{Lo: 7})
	imports := map[string]Extern{"fn": host, "table": table, "memory": memory, "global": global}
	def := &module.Module{
		Types:     []module.FuncType{hostSignature(s, host)},
		Imports:   []module.Import{{Module: "m", Name: "fn", Kind: 0}, {Module: "m", Name: "table", Kind: 1, Table: table.Table.Type}, {Module: "m", Name: "memory", Kind: 2, Memory: memory.Memory.Limits()}, {Module: "m", Name: "global", Kind: 3, Global: global.Global.Type}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0xb}}}},
		Tables:    []module.TableType{{Element: wasm.ExternRef, Limits: module.Limits{Min: 1}}},
		Globals: []module.Global{
			{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x23}, {Op: 0xb}}},
			{Type: module.GlobalType{Value: wasm.FuncRef}, Init: module.Expr{{Op: 0xd2, A: 1}, {Op: 0xb}}},
			{Type: module.GlobalType{Value: wasm.ExternRef}, Init: module.Expr{{Op: 0xd0, A: uint32(wasm.ExternRef)}, {Op: 0xb}}},
			{Type: module.GlobalType{Value: wasm.V128}, Init: expression(0xfd000c, 0)},
		},
		Elements: []module.Element{
			{Type: wasm.FuncRef, Offset: expression(0x41, 0), Functions: []uint32{1}},
			{Type: wasm.FuncRef, Mode: module.Passive, Expressions: []module.Expr{{{Op: 0xd2}, {Op: 0xb}}}},
			{Type: wasm.FuncRef, Mode: module.Declarative, Functions: []uint32{0}},
		},
		Data:    []module.Data{{Offset: expression(0x41, 0), Bytes: []byte{1, 2}}, {Mode: module.Passive, Bytes: []byte{3, 4}}},
		Exports: []module.Export{{Name: "f", Kind: 0, Index: 1}, {Name: "t", Kind: 1}, {Name: "m", Kind: 2}, {Name: "g", Kind: 3, Index: 1}},
	}
	resolver := func(_, name string) (Extern, bool) { e, ok := imports[name]; return e, ok }
	compiled := build(t, def)
	in, err := s.Allocate(compiled, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if in.Globals[1].Value.Lo != 7 || in.Exports["g"].Global != in.Globals[1] {
		t.Fatal("initializer/export")
	}
	roots := 0
	in.RetainedValues(func(typ wasm.ValueType, v value.Value) {
		if !typ.IsReference() {
			t.Fatal(typ)
		}
		roots++
	})
	if roots != 8 {
		t.Fatal(roots)
	}
	if err := in.Initialize(); err != nil {
		t.Fatal(err)
	}
	if table.Table.Values[0] != in.Functions[1] || in.Data[0] != nil || in.Elements[0] != nil || in.Elements[2] != nil {
		t.Fatal("active/declarative segments")
	}
	if s.Usage().SegmentBytes != 2 || s.Usage().SegmentElements != 1 {
		t.Fatal(s.Usage())
	}
	in.DropData(1)
	in.DropData(1)
	in.DropElement(1)
	in.DropElement(1)
	if s.Usage().SegmentBytes != 0 || s.Usage().SegmentElements != 0 {
		t.Fatal(s.Usage())
	}
	before := s.Usage()
	if _, err := s.Allocate(compiled, nil); !errors.Is(err, ErrLink) || before != s.Usage() {
		t.Fatal(err, s.Usage())
	}
	foreign, _ := New(DefaultLimits()).NewHost(hostSignature(s, host), nil)
	imports["fn"] = foreign
	if _, err := s.Allocate(compiled, resolver); !errors.Is(err, ErrLink) || before != s.Usage() {
		t.Fatal(err)
	}
}
func hostSignature(s *Store, ex Extern) module.FuncType {
	f, _ := s.Function(ex.Function)
	return f.Signature
}

func TestAllocationRollback(t *testing.T) {
	def := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: expression(1, 0)}}, Memories: []module.Limits{{Min: 1}}, Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}, Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: expression(0x41, 0)}}, Elements: []module.Element{{Type: wasm.FuncRef, Mode: module.Passive, Functions: []uint32{0}}}, Data: []module.Data{{Mode: module.Passive, Bytes: []byte{1}}}}
	compiled := build(t, def)
	for _, field := range []string{"MaxTypes", "MaxFunctions", "MaxInstances", "MaxMemories", "MaxTables", "MaxMemoryPages", "MaxTableElements", "MaxGlobals", "MaxSegmentBytes", "MaxSegmentElements"} {
		t.Run(field, func(t *testing.T) {
			limits := DefaultLimits()
			reflect.ValueOf(&limits).Elem().FieldByName(field).SetUint(0)
			s := New(limits)
			if in, err := s.Allocate(compiled, nil); in != nil || !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(in, err)
			}
			if s.Usage() != (Usage{}) || len(s.functions) != 1 || len(s.instances) != 0 || len(s.registry.Keys) != 0 || len(s.tables) != 0 || len(s.memories) != 0 || len(s.globals) != 0 {
				t.Fatal("failed allocation retained state", s.Usage())
			}
		})
	}
	if _, err := New(DefaultLimits()).Allocate(nil, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	s := New(DefaultLimits())
	in, err := s.Allocate(compiled, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := in.Initialize(); err != nil {
		t.Fatal(err)
	}
}

func TestInitializationTrapPreservesImportedEffects(t *testing.T) {
	s := New(DefaultLimits())
	tab, _ := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}})
	mem, _ := s.NewMemory(module.Limits{Min: 1})
	def := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Name: "t", Kind: 1, Table: tab.Table.Type}, {Name: "m", Kind: 2, Memory: mem.Memory.Limits()}}, Functions: []module.Function{{Body: expression(1, 0)}}, Elements: []module.Element{{Type: wasm.FuncRef, Functions: []uint32{0}, Offset: expression(0x41, 0)}}, Data: []module.Data{{Bytes: []byte{42}, Offset: expression(0x41, 0)}, {Bytes: []byte{1}, Offset: expression(0x41, 65536)}}}
	in, err := s.Allocate(build(t, def), func(_, n string) (Extern, bool) {
		if n == "t" {
			return tab, true
		}
		return mem, true
	})
	if err != nil {
		t.Fatal(err)
	}
	var trap *wasm.Trap
	if err := in.Initialize(); !errors.As(err, &trap) || trap.Code != wasm.TrapMemoryBounds {
		t.Fatal(err)
	}
	b := make([]byte, 1)
	mem.Memory.Read(0, b)
	if b[0] != 42 || tab.Table.Values[0] != in.Functions[0] {
		t.Fatal("lost imported effects")
	}
	f, ok := s.Function(tab.Table.Values[0])
	if !ok || f.Instance != in {
		t.Fatal("escaped function lost instance")
	}
	second := *def
	second.Elements = append([]module.Element(nil), def.Elements...)
	second.Elements[0].Offset = expression(0x41, 1)
	in, err = s.Allocate(build(t, &second), func(_, n string) (Extern, bool) {
		if n == "t" {
			return tab, true
		}
		return mem, true
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := in.Initialize(); !errors.As(err, &trap) || trap.Code != wasm.TrapTableBounds {
		t.Fatal(err)
	}
}

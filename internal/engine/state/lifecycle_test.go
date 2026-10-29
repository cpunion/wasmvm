package state

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func ownedDefinition() *module.Module {
	return &module.Module{
		Types: []module.FuncType{{}}, Functions: []module.Function{{Body: expression(1, 0)}},
		Tables:   []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 2}}},
		Memories: []module.Limits{{Min: 1}},
		Globals:  []module.Global{{Type: module.GlobalType{Value: wasm.FuncRef, Mutable: true}, Init: module.Expr{{Op: 0xd2}, {Op: 0xb}}}},
		Elements: []module.Element{{Mode: module.Passive, Type: wasm.FuncRef, Functions: []uint32{0}}},
		Data:     []module.Data{{Mode: module.Passive, Bytes: []byte{1, 2}}},
		Exports:  []module.Export{{Name: "f", Kind: 0}, {Name: "t", Kind: 1}, {Name: "m", Kind: 2}, {Name: "g", Kind: 3}},
	}
}

func TestCollectFailedRestoresQuotasAndReleasesStorage(t *testing.T) {
	limits := DefaultLimits()
	limits.MaxInstances = 1
	limits.MaxFunctions = 1
	s := New(limits)
	compiled := build(t, ownedDefinition())
	for attempt := 0; attempt < 100; attempt++ {
		in, err := s.Allocate(compiled, nil)
		if err != nil {
			t.Fatal(attempt, err)
		}
		mem, table, global := in.Memories[0], in.Tables[0], in.Globals[0]
		if in.usageForTest() != s.Usage() {
			t.Fatal(s.Usage())
		}
		s.GrowMemory(mem, 1)
		s.GrowTable(table, 1, 0)
		s.CollectFailed(nil, nil)
		if s.Usage() != (Usage{}) || len(s.instances) != 0 || len(s.functions) != 1 || len(s.tables) != 0 || len(s.memories) != 0 || len(s.globals) != 0 {
			t.Fatal("leaked failed state", s.Usage())
		}
		if !in.disposed || in.Compiled != nil || in.Exports != nil || mem.Pages() != 0 || len(table.Values) != 0 || global.Value != (value.Value{}) {
			t.Fatal("retained payload")
		}
		if _, ok := s.GrowMemory(mem, 1); ok {
			t.Fatal("retired memory grew")
		}
		if _, ok := s.GrowTable(table, 1, 0); ok {
			t.Fatal("retired table grew")
		}
		if err := in.Initialize(); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if err := s.Publish(in); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
	}
	if err := s.Publish(nil); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	foreign, _ := New(DefaultLimits()).Allocate(compiled, nil)
	if err := s.Publish(foreign); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
}

func (in *Instance) usageForTest() Usage {
	return Usage{Instances: 1, Functions: 1, Tables: 1, Memories: 1, Globals: 1, MemoryPages: 1, TableElements: 2, SegmentElements: 1, SegmentBytes: 2}
}

func TestFailedInstanceRootsAndUnreachableCycles(t *testing.T) {
	for _, root := range []string{"function", "active", "table", "global", "published"} {
		t.Run(root, func(t *testing.T) {
			s := New(DefaultLimits())
			compiled := build(t, ownedDefinition())
			a, _ := s.Allocate(compiled, nil)
			b, _ := s.Allocate(compiled, nil)
			a.Tables[0].Values[0] = b.Functions[0]
			b.Globals[0].Value.Lo = a.Functions[0]
			hostTable, _ := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}})
			hostGlobal, _ := s.NewGlobal(module.GlobalType{Value: wasm.FuncRef}, value.Value{})
			// An externref with the same integer bits is not a function root.
			s.NewGlobal(module.GlobalType{Value: wasm.ExternRef}, value.Value{Lo: a.Functions[0]})
			var functions []uint64
			var active []*Instance
			switch root {
			case "function":
				functions = []uint64{0, 999999, a.Functions[0]}
			case "active":
				active = []*Instance{nil, a, &Instance{Owner: New(DefaultLimits())}}
			case "table":
				hostTable.Table.Values[0] = a.Functions[0]
			case "global":
				hostGlobal.Global.Value.Lo = a.Functions[0]
			case "published":
				if err := s.Publish(a); err != nil {
					t.Fatal(err)
				}
			}
			s.CollectFailed(functions, active)
			if s.Usage().Instances != 2 || a.disposed || b.disposed {
				t.Fatal("lost reachable cycle", s.Usage())
			}
			hostTable.Table.Values[0] = 0
			hostGlobal.Global.Value = value.Value{}
			s.CollectFailed(nil, nil)
			want := uint64(0)
			if root == "published" {
				want = 2
			}
			if s.Usage().Instances != want {
				t.Fatal("unreachable cycle retained", s.Usage())
			}
		})
	}
}

func TestImportedObjectsPreserveFailedOwner(t *testing.T) {
	for _, kind := range []byte{0, 1, 2, 3} {
		s := New(DefaultLimits())
		owner, _ := s.Allocate(build(t, ownedDefinition()), nil)
		names := []string{"f", "t", "m", "g"}
		ex := owner.Exports[names[kind]]
		im := module.Import{Kind: kind}
		def := &module.Module{Imports: []module.Import{im}}
		switch kind {
		case 0:
			def.Types = []module.FuncType{{}}
		case 1:
			def.Imports[0].Table = ex.Table.Type
		case 2:
			def.Imports[0].Memory = ex.Memory.Limits()
		case 3:
			def.Imports[0].Global = ex.Global.Type
		}
		user, err := s.Allocate(build(t, def), func(string, string) (Extern, bool) { return ex, true })
		if err != nil {
			t.Fatal(err)
		}
		s.Publish(user)
		s.CollectFailed(nil, nil)
		if owner.disposed || s.Usage().Instances != 2 {
			t.Fatal("lost imported owner", kind)
		}
	}
}

func TestFunctionIdentityHolesAreBounded(t *testing.T) {
	limits := DefaultLimits()
	limits.MaxFunctions = 2
	s := New(limits)
	compiled := build(t, ownedDefinition())
	a, _ := s.Allocate(compiled, nil)
	b, _ := s.Allocate(compiled, nil)
	s.CollectFailed([]uint64{b.Functions[0]}, nil)
	if !a.disposed || s.Usage().Functions != 1 {
		t.Fatal(s.Usage())
	}
	if _, err := s.NewHost(module.FuncType{}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("identity quota", err)
	}
	if _, err := s.Allocate(compiled, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("identity quota", err)
	}
	s.CollectFailed(nil, nil)
	if _, err := s.NewHost(module.FuncType{}, nil); err != nil {
		t.Fatal("suffix not reclaimed", err)
	}
}

func TestCloseReleasesAllStateAndRejectsAllocation(t *testing.T) {
	s := New(DefaultLimits())
	in, _ := s.Allocate(build(t, ownedDefinition()), nil)
	s.Publish(in)
	host, _ := s.NewHost(module.FuncType{}, new(int))
	tab, _ := s.NewTable(module.TableType{Element: wasm.ExternRef, Limits: module.Limits{Min: 1}})
	mem, _ := s.NewMemory(module.Limits{Min: 1})
	g, _ := s.NewGlobal(module.GlobalType{Value: wasm.ExternRef}, value.Value{Lo: 42})
	count := 0
	s.RetainedValues(func(typ wasm.ValueType, v value.Value) {
		if !typ.IsReference() {
			t.Fatal(typ)
		}
		count++
	})
	if count != 6 {
		t.Fatal(count)
	}
	s.Close()
	s.Close()
	s.CollectFailed(nil, nil)
	if s.Usage() != (Usage{}) || len(s.functions) != 0 || s.registry.Keys != nil || !in.disposed || len(tab.Table.Values) != 0 || mem.Memory.Pages() != 0 || g.Global.Value.Lo != 0 {
		t.Fatal("close retained resources")
	}
	if _, ok := s.Function(host.Function); ok {
		t.Fatal("closed host callback")
	}
	if s.TypeID(module.FuncType{}) != 0 {
		t.Fatal("closed type registry")
	}
	if err := s.Publish(in); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.Allocate(nil, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.NewHost(module.FuncType{}, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.NewTable(module.TableType{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.NewMemory(module.Limits{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := s.NewGlobal(module.GlobalType{}, value.Value{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	count = 0
	s.RetainedValues(func(wasm.ValueType, value.Value) { count++ })
	if count != 0 {
		t.Fatal(count)
	}
}

func TestEmptyObjectsStillConsumeQuotas(t *testing.T) {
	limits := DefaultLimits()
	limits.MaxTables = 1
	limits.MaxMemories = 1
	s := New(limits)
	if _, err := s.NewTable(module.TableType{Element: wasm.FuncRef}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.NewTable(module.TableType{Element: wasm.FuncRef}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.NewMemory(module.Limits{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.NewMemory(module.Limits{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

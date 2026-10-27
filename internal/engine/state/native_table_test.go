package state

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeTableStorageGrowthAndReclamation(t *testing.T) {
	s := NewNative(DefaultLimits())
	defer s.Close()
	if !native.Available {
		if _, err := s.NewTable(module.TableType{Element: wasm.FuncRef}); !errors.Is(err, native.ErrUnavailable) {
			t.Fatal(err)
		}
		return
	}
	for _, typ := range []wasm.ValueType{wasm.FuncRef, wasm.ExternRef} {
		ex, err := s.NewTable(module.TableType{Element: typ, Limits: module.Limits{Max: 4, HasMax: true}})
		if err != nil {
			t.Fatal(err)
		}
		table := ex.Table
		if old, ok := s.GrowTable(table, 2, 7); old != 0 || !ok {
			t.Fatal(old, ok)
		}
		if table.NativeAddress() == 0 || !table.Fill(1, 1, 9) || !table.Init(0, 0, 1, []uint64{8}) {
			t.Fatal(table)
		}
		if old, ok := s.GrowTable(table, 2, 11); old != 2 || !ok {
			t.Fatal(old, ok)
		}
		if !reflect.DeepEqual(table.Values, []uint64{8, 9, 11, 11}) {
			t.Fatal(table.Values)
		}
		if _, ok := s.GrowTable(table, 1, 0); ok {
			t.Fatal("exceeded native table limit")
		}
		if _, ok := s.GrowTable(table, 0, 0); !ok {
			t.Fatal("zero growth")
		}
		storage := table.native
		s.Close()
		if _, err := storage.Buffer(); !errors.Is(err, native.ErrClosed) || table.Values != nil || table.NativeAddress() != 0 {
			t.Fatal(err, table)
		}
		if _, ok := s.GrowTable(table, 0, 0); ok {
			t.Fatal("retired table grew")
		}
		s = NewNative(DefaultLimits())
		defer s.Close()
	}
	// Allocate must reclaim tables if a later resource exceeds its quota.
	limits := DefaultLimits()
	limits.MaxMemories = 0
	s = NewNative(limits)
	defer s.Close()
	m := build(t, &module.Module{Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}, Memories: []module.Limits{{Min: 1}}})
	if _, err := s.Allocate(m, nil); !errors.Is(err, wasm.ErrLimit) || s.Usage() != (Usage{}) || len(s.tables) != 0 {
		t.Fatal(err, s.Usage())
	}
}

func TestNativeTablesKeepFailedInstanceRoots(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	s := NewNative(DefaultLimits())
	defer s.Close()
	in, err := s.Allocate(build(t, ownedDefinition()), nil)
	if err != nil {
		t.Fatal(err)
	}
	root, err := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}})
	if err != nil {
		t.Fatal(err)
	}
	storage := in.Tables[0].native
	root.Table.Values[0] = in.Functions[0]
	s.CollectFailed(nil, nil)
	if in.disposed {
		t.Fatal("native table lost a function root")
	}
	root.Table.Values[0] = 0
	s.CollectFailed(nil, nil)
	if !in.disposed {
		t.Fatal("unreachable instance survived")
	}
	if _, err := storage.Buffer(); !errors.Is(err, native.ErrClosed) {
		t.Fatal(err)
	}
}

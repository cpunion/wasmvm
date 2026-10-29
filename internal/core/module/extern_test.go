package module

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestExternSubtype(t *testing.T) {
	r := new(TypeRegistry)
	ids, ok := r.InternModule(&Module{Types: []FuncType{{}, {Params: []wasm.ValueType{wasm.I32}}}}, 10)
	if !ok {
		t.Fatal("registry")
	}
	for _, row := range []struct {
		a, b Import
		want bool
	}{
		{Import{Kind: 0, Type: ids[0]}, Import{Kind: 0, Type: ids[0]}, true},
		{Import{Kind: 0, Type: ids[0]}, Import{Kind: 0, Type: ids[1]}, false},
		{Import{Kind: 4, Type: ids[0]}, Import{Kind: 4, Type: ids[0]}, true},
		{Import{Kind: 4, Type: ids[0]}, Import{Kind: 4, Type: ids[1]}, false},
		{Import{Kind: 0}, Import{Kind: 4}, false},
		{Import{Kind: 5}, Import{Kind: 5}, false},
		{Import{Kind: 2, Memory: Limits{Min: 2, HasMax: true, Max: 4}}, Import{Kind: 2, Memory: Limits{Min: 1, HasMax: true, Max: 5}}, true},
		{Import{Kind: 2, Memory: Limits{Min: 2}}, Import{Kind: 2, Memory: Limits{Min: 1, HasMax: true, Max: 5}}, false},
		{Import{Kind: 2, Memory: Limits{Min: 2, Index64: true}}, Import{Kind: 2, Memory: Limits{Min: 1}}, false},
		{Import{Kind: 2, Memory: Limits{Min: 2, Shared: true}}, Import{Kind: 2, Memory: Limits{Min: 1}}, false},
		{Import{Kind: 1, Table: TableType{Element: wasm.FuncRef, Limits: Limits{Min: 2}}}, Import{Kind: 1, Table: TableType{Element: wasm.FuncRef, Limits: Limits{Min: 1}}}, true},
		{Import{Kind: 1, Table: TableType{Element: wasm.ExternRef}}, Import{Kind: 1, Table: TableType{Element: wasm.FuncRef}}, false},
		{Import{Kind: 3, Global: GlobalType{Value: wasm.NoFuncRef}}, Import{Kind: 3, Global: GlobalType{Value: wasm.FuncRef}}, true},
		{Import{Kind: 3, Global: GlobalType{Value: wasm.NoFuncRef, Mutable: true}}, Import{Kind: 3, Global: GlobalType{Value: wasm.FuncRef, Mutable: true}}, false},
		{Import{Kind: 3, Global: GlobalType{Value: wasm.I32}}, Import{Kind: 3, Global: GlobalType{Value: wasm.I32, Mutable: true}}, false},
	} {
		if got := ExternSubtype(r, row.a, row.b); got != row.want {
			t.Fatal(row, got)
		}
	}
}

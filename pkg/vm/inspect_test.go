package vm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRepeatedLargeSignatureMetadataHasBoundedCopies(t *testing.T) {
	// Only two type vectors are present in the module. Thousands of functions
	// and exports must not multiply these vectors' storage in an Info snapshot.
	const count = 4096
	params := make([]wasm.ValueType, count)
	for i := range params {
		params[i] = wasm.I32
	}
	d := &module.Module{Types: []module.FuncType{{Params: params}, {Params: params}}}
	for i := 0; i < count; i++ {
		d.Functions = append(d.Functions, module.Function{Type: uint32(i % 2)})
		d.Exports = append(d.Exports, module.Export{Kind: 0, Index: uint32(i)})
	}
	info := describeModule(d)
	distinct := make(map[*FunctionType]bool)
	for _, ex := range info.Exports {
		distinct[ex.Function] = true
		if len(ex.Function.Params) != count {
			t.Fatal("lost signature")
		}
	}
	if len(distinct) != 2 {
		t.Fatalf("metadata duplicated type vectors %d times", len(distinct))
	}
	info.Exports[0].Function.Params[0] = wasm.F64
	if params[0] != wasm.I32 || info.Exports[1].Function.Params[0] != wasm.I32 {
		t.Fatal("borrowed source or merged distinct types")
	}
	if describeModule(d).Exports[0].Function.Params[0] != wasm.I32 {
		t.Fatal("snapshots alias")
	}
}

type cancelMetadata struct{ calls int }

func (c *cancelMetadata) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelMetadata) Done() <-chan struct{}       { return nil }
func (c *cancelMetadata) Value(any) any               { return nil }
func (c *cancelMetadata) Err() error {
	c.calls++
	if c.calls >= 4 {
		return context.Canceled
	}
	return nil
}
func TestInspectChecksCancellationBeforePublishingMetadata(t *testing.T) {
	if info, err := NewEngine().Inspect(&cancelMetadata{}, addBinary); !errors.Is(err, context.Canceled) || info.Exports != nil {
		t.Fatal(info, err)
	}
}

func TestStaticInspectionDoesNotCompileOrExecute(t *testing.T) {
	c := DefaultEngineConfig()
	c.Lower.MaxNodes = 0
	c.Slots.MaxNodes = 0
	e, _ := NewEngineWithConfig(c)
	if err := e.Validate(nil, addBinary); err != nil {
		t.Fatal(err)
	}
	info, err := e.Inspect(nil, addBinary)
	if err != nil || info.Exports[0].Function.Params[0] != wasm.I32 {
		t.Fatal(info, err)
	}
	info.Exports[0].Function.Params[0] = wasm.F64
	fresh, _ := e.Inspect(nil, addBinary)
	if fresh.Exports[0].Function.Params[0] != wasm.I32 {
		t.Fatal("borrowed metadata")
	}
	if _, err := e.Compile(nil, addBinary); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := e.Inspect(nil, nil); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(err)
	}
	e.Close()
	if err := e.Validate(nil, addBinary); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestCopiedObjectMetadataAndStartReceipt(t *testing.T) {
	e, _, l := setup(t)
	d := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: module.Expr{{Op: 1}, {Op: 0xb}}}}, Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}, Memories: []module.Limits{{Min: 1}}, Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "f", Kind: 0}, {Name: "t", Kind: 1}, {Name: "m", Kind: 2}, {Name: "g", Kind: 3}}, HasStart: true}
	m := compileDefinition(t, e, d)
	info, err := m.Info()
	if err != nil {
		t.Fatal(err)
	}
	info.Exports[1].Table.Limits.Min = 0
	info.Exports[2].Memory.Min = 0
	info.Exports[3].Global.Value = wasm.F64
	fresh, _ := m.Info()
	if fresh.Exports[1].Table.Limits.Min != 1 || fresh.Exports[2].Memory.Min != 1 || fresh.Exports[3].Global.Value != wasm.I32 {
		t.Fatal("borrowed object metadata")
	}
	in, err := l.Instantiate(nil, m, CallOptions{Metered: true, Fuel: 7})
	if err != nil || in.StartFuelRemaining() != 6 {
		t.Fatal(in, err)
	}
	fn := export(t, in, "f")
	fn.Call(nil, nil, CallOptions{Metered: true, Fuel: 1})
	if in.StartFuelRemaining() != 6 {
		t.Fatal("mutable receipt")
	}
}

package lower

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompleteCore2Lowering(t *testing.T) {
	root := "../../../artifacts/core2/fixtures"
	scripts, modules, functions := 0, 0, 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "script.json" {
			return nil
		}
		scripts++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var script struct {
			Commands []struct {
				Type, Filename string
				ModuleType     string `json:"module_type"`
			}
		}
		if err = json.Unmarshal(data, &script); err != nil {
			return err
		}
		for _, cmd := range script.Commands {
			if cmd.Filename == "" || cmd.Type == "assert_malformed" || cmd.Type == "assert_invalid" {
				continue
			}
			name := filepath.Join(filepath.Dir(path), cmd.Filename)
			data, err = os.ReadFile(name)
			if err != nil {
				return err
			}
			m, err := decode.Decode(data, decode.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			checked, err := validate.Check(m, validate.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			lowered, err := Compile(checked, DefaultLimits())
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}
			if lowered.Validated != checked || len(lowered.Functions) != len(m.Functions) {
				t.Fatal("module identity/function count")
			}
			modules++
			functions += len(lowered.Functions)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scripts != 148 || modules != 1716 {
		t.Fatalf("input inventory: %d scripts, %d modules", scripts, modules)
	}
	t.Logf("Lowered and verified %d functions in %d statically valid Core 2.0 modules; no execution assertions run.", functions, modules)
}

func compileTest(t *testing.T, m *module.Module) *ir.Function {
	t.Helper()
	checked, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := Compile(checked, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return lowered.Functions[0]
}
func TestLoopTupleAndDeadCode(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32, wasm.I32}}}, Functions: []module.Function{{Locals: []module.Local{{Count: 2, Type: wasm.I32}}, LocalCount: 2, Body: module.Expr{
		{Op: 0x41, Bits: 1}, {Op: 0x41, Bits: 2}, {Op: 3, Block: 1},
		{Op: 0x21, A: 1}, {Op: 0x21, A: 0}, {Op: 0x20, A: 1}, {Op: 0x20, A: 0}, {Op: 0x41}, {Op: 0xd, A: 0},
		{Op: 0xb}, {Op: 0x1a}, {Op: 0x1a}, {Op: 0xf},
		{Op: 0x41}, {Op: 0x41}, {Op: 0x6d}, {Op: 0x1a}, {Op: 0xb},
	}}}}
	f := compileTest(t, m)
	foundLoop := false
	for _, b := range f.Blocks {
		for _, op := range b.Ops {
			if f.Sources[op.Source].Wasm.Op == 0x6d {
				t.Fatal("dead division became executable")
			}
		}
		if b.Term.Kind == ir.If {
			edge := b.Term.Edges[0]
			if len(edge.Args) != 2 || len(f.Blocks[edge.Target].Params) != 2 {
				t.Fatal("lost loop tuple")
			}
			foundLoop = true
		}
	}
	if !foundLoop {
		t.Fatal("missing loop backedge")
	}
}

func TestLoweringLimits(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: module.Expr{{Op: 0xb}}}}}
	checked, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, limits := range []Limits{{0, 1000}, {1000, 0}} {
		got, err := Compile(checked, limits)
		if got != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(got, err)
		}
	}
	if got, err := Compile(nil, DefaultLimits()); got != nil || !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(got, err)
	}
	compiled, err := Compile(checked, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	b := ir.VerifyBudget{Remaining: 1000}
	if err := b.Verify(compiled.Functions[0]); err != nil {
		t.Fatal(err)
	}
	oneFunction := 1000 - b.Remaining
	if _, err := Compile(checked, Limits{1000, oneFunction}); err != nil {
		t.Fatal(err)
	}
	m.Functions = append(m.Functions, m.Functions[0])
	checked, err = validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := Compile(checked, Limits{1000, oneFunction}); got != nil || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(got, err)
	}
}

func FuzzLower(f *testing.F) {
	f.Add([]byte("\x00asm\x01\x00\x00\x00"))
	f.Add([]byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 4, 1, 0x60, 0, 0, 3, 2, 1, 0, 10, 7, 1, 5, 0, 0x41, 1, 0x1a, 0xb})
	f.Fuzz(func(t *testing.T, data []byte) {
		limits := decode.DefaultLimits()
		limits.MaxBytes = 1 << 16
		limits.MaxNodes = 4096
		limits.MaxLocals = 4096
		limits.MaxNesting = 64
		m, err := decode.Decode(data, limits)
		if err != nil {
			return
		}
		checked, err := validate.Check(m, validate.Limits{MaxLocals: 4096, MaxStack: 4096, MaxControl: 65})
		if err != nil {
			return
		}
		if _, err = Compile(checked, Limits{1 << 16, 1 << 20}); err != nil && !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	})
}

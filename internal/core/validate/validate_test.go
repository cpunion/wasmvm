package validate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompleteCore2Validation(t *testing.T) {
	root := "../../../artifacts/core2/fixtures"
	if _, err := os.Stat(root); err != nil {
		t.Fatal("prepare pinned inputs with python3 scripts/core2_corpus.py:", err)
	}
	files, malformed, invalid, valid, scripts := 0, 0, 0, 0, 0
	operations := make(map[uint32]bool)
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
		for _, command := range script.Commands {
			if command.Filename == "" || command.ModuleType == "text" {
				continue
			}
			input, err := os.ReadFile(filepath.Join(filepath.Dir(path), command.Filename))
			if err != nil {
				return err
			}
			files++
			m, err := decode.Decode(input, decode.DefaultLimits())
			name := filepath.Base(filepath.Dir(path)) + "/" + command.Filename
			if command.Type == "assert_malformed" {
				malformed++
				if !errors.Is(err, wasm.ErrMalformed) {
					t.Errorf("%s: expected malformed, got %v", name, err)
				}
				continue
			}
			if err != nil {
				t.Errorf("%s: decode: %v", name, err)
				continue
			}
			checked, err := Check(m, DefaultLimits())
			if command.Type == "assert_invalid" {
				invalid++
				if checked != nil || !errors.Is(err, wasm.ErrInvalid) {
					t.Errorf("%s: expected invalid, got %v", name, err)
				}
			} else {
				valid++
				if err != nil {
					t.Errorf("%s (%s): %v", name, command.Type, err)
				}
				for _, fn := range m.Functions {
					for _, in := range fn.Body {
						operations[in.Op] = true
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scripts != 148 || files != 4581 || malformed != 719 || invalid != 2146 || valid != 1716 {
		t.Fatalf("inventory changed: scripts=%d binaries=%d malformed=%d invalid=%d valid=%d", scripts, files, malformed, invalid, valid)
	}
	t.Logf("Checked %d modules: %d malformed, %d invalid, %d statically valid. No guest execution/linking assertions run.", files, malformed, invalid, valid)
	simd := 0
	for op := range operations {
		if op>>16 == 0xfd {
			simd++
		}
	}
	if simd != 236 {
		t.Fatalf("valid module inputs cover %d SIMD opcodes, want all 236", simd)
	}
}

func TestSIMDAverageArity(t *testing.T) {
	// i8x16.avgr_u consumes two vectors even though adjacent opcode numbers
	// describe unary conversions. This caught the initial classification error.
	for operands := 0; operands <= 3; operands++ {
		var expr module.Expr
		for i := 0; i < operands; i++ {
			expr = append(expr, module.Instr{Op: 0xfd000c})
		}
		expr = append(expr, module.Instr{Op: 0xfd007b}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
		_, err := Check(validFunction(expr), DefaultLimits())
		if (err == nil) != (operands == 2) {
			t.Fatalf("%d operands: %v", operands, err)
		}
	}
}

func TestValidationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		module *module.Module
		kind   error
	}{
		{"nil", nil, wasm.ErrInvalid},
		{"missing end", validFunction(nil), wasm.ErrInvalid},
		{"after end", validFunction(module.Expr{{Op: 0xb}, {Op: 1}}), wasm.ErrInvalid},
		{"unknown instruction", validFunction(module.Expr{{Op: 0xff}, {Op: 0xb}}), wasm.ErrInvalid},
		{"unexpected else", validFunction(module.Expr{{Op: 5}, {Op: 0xb}}), wasm.ErrInvalid},
		{"unknown SIMD", validFunction(module.Expr{{Op: 0xfd00ff + 1}, {Op: 0xb}}), wasm.ErrInvalid},
		{"unknown misc", validFunction(module.Expr{{Op: 0xfc0012}, {Op: 0xb}}), wasm.ErrInvalid},
		{"invalid block type", validFunction(module.Expr{{Op: 2, Block: -129}, {Op: 0xb}, {Op: 0xb}}), wasm.ErrInvalid},
		{"block type overflow", validFunction(module.Expr{{Op: 2, Block: 1 << 32}, {Op: 0xb}, {Op: 0xb}}), wasm.ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Check(tc.module, DefaultLimits())
			if got != nil || !errors.Is(err, tc.kind) {
				t.Fatal(got, err)
			}
		})
	}
	m := validFunction(module.Expr{{Op: 0xb}})
	m.Types[0].Params = []wasm.ValueType{wasm.I32}
	m.Functions[0].Locals = []module.Local{{Count: ^uint32(0), Type: wasm.I32}}
	if _, err := Check(m, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatalf("parameter+local index overflow: %v", err)
	}
	m.Functions[0].Locals = []module.Local{{Count: 3, Type: wasm.I32}}
	if _, err := Check(m, Limits{MaxLocals: 3, MaxControl: 1, MaxStack: 1}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatalf("parameter+local quota: %v", err)
	}
	if _, err := Check(m, Limits{MaxLocals: 4, MaxControl: 1, MaxStack: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestPolymorphicBranchesAndImplicitElse(t *testing.T) {
	// The implicit else of a typed if passes its parameters through. Even an
	// unreachable then arm cannot make a different implicit result type valid.
	m := validFunction(module.Expr{{Op: 0x41}, {Op: 0x41}, {Op: 4, Block: 1}, {Op: 0}, {Op: 0xb}, {Op: 0x1a}, {Op: 0xb}})
	m.Types = append(m.Types, module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.F32}})
	if _, err := Check(m, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	m.Types[1].Results[0] = wasm.I32
	if _, err := Check(m, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	// Unknown operands can match different branch-label types in unreachable
	// code, but an actual i32 cannot be branched to an f32 label.
	for _, reachable := range []bool{false, true} {
		expr := module.Expr{{Op: 2, Block: -1}, {Op: 2, Block: -3}}
		if reachable {
			expr = append(expr, module.Instr{Op: 0x41})
		} else {
			expr = append(expr, module.Instr{Op: 0})
		}
		expr = append(expr, module.Instr{Op: 0x41}, module.Instr{Op: 0xe, A: 0, Labels: []uint32{1}}, module.Instr{Op: 0xb}, module.Instr{Op: 0x1a}, module.Instr{Op: 0x41}, module.Instr{Op: 0xb}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
		_, err := Check(validFunction(expr), DefaultLimits())
		if (err == nil) == reachable {
			t.Fatalf("reachable=%v: %v", reachable, err)
		}
	}
}

func validFunction(body module.Expr) *module.Module {
	return &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: body}}}
}

func TestLimitsAndMetadata(t *testing.T) {
	m := validFunction(module.Expr{{Op: 0x41}, {Op: 0x1a}, {Op: 0xb}})
	checked, err := Check(m, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if checked.Definition() != m {
		t.Fatal("definition identity")
	}
	info, ok := checked.FunctionInfo(0)
	if !ok || info.MaxStack != 1 || info.MaxControl != 1 {
		t.Fatal(info, ok)
	}
	if _, ok := checked.FunctionInfo(1); ok {
		t.Fatal("invalid function index")
	}
	for _, tc := range []struct {
		name   string
		limits Limits
	}{
		{"stack", Limits{MaxLocals: 100, MaxControl: 10}},
		{"control", Limits{MaxLocals: 100, MaxStack: 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := Check(m, tc.limits); got != nil || !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(got, err)
			}
		})
	}
	m.Functions[0].Locals = []module.Local{{Count: 1, Type: wasm.I32}}
	if _, err := Check(m, Limits{MaxStack: 10, MaxControl: 10}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzValidate(f *testing.F) {
	f.Add([]byte("\x00asm\x01\x00\x00\x00"))
	f.Add([]byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 4, 1, 0x60, 0, 0, 3, 2, 1, 0, 10, 7, 1, 5, 0, 0x41, 1, 0x1a, 0xb})
	input, err := os.ReadFile("../../../pkg/vm/testdata/multi-memory.wasm")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(input)
	input, err = os.ReadFile("../../../pkg/vm/testdata/memory64.wasm")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(input)
	f.Fuzz(func(t *testing.T, input []byte) {
		limits := decode.DefaultLimits()
		limits.MaxBytes = 1 << 16
		limits.MaxNodes = 4096
		limits.MaxLocals = 4096
		limits.MaxNesting = 64
		for _, features := range []wasm.Features{0, wasm.SupportedFeatures} {
			m, err := decode.DecodeWithFeatures(input, limits, features)
			if err != nil {
				continue
			}
			checked, err := CheckWithFeatures(m, Limits{4096, 4096, 65}, features)
			if err != nil && (checked != nil || (!errors.Is(err, wasm.ErrInvalid) && !errors.Is(err, wasm.ErrLimit))) {
				t.Fatalf("invalid failure: %v", err)
			}
		}
	})
}

func BenchmarkValidate(b *testing.B) {
	m := validFunction(module.Expr{{Op: 0x41}, {Op: 0x41}, {Op: 0x6a}, {Op: 0x1a}, {Op: 0xb}})
	limits := DefaultLimits()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Check(m, limits); err != nil {
			b.Fatal(err)
		}
	}
}

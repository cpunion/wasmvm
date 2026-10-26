package numeric

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type constant struct{ Type, Value string }

// These official scripts export direct scalar numeric operations. Require
// that shape explicitly and call the production numeric primitive with their
// arguments. This is not a second interpreter or a full-module execution test.
func TestCore2ScalarNumericAssertions(t *testing.T) {
	returns, traps, invalid, textOnly := 0, 0, 0, 0
	seen := make(map[uint32]bool)
	for _, name := range []string{"i32", "i64", "f32", "f64", "f32_cmp", "f64_cmp", "float_misc", "conversions"} {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("../../../artifacts/core2/fixtures", name)
			data, err := os.ReadFile(filepath.Join(dir, "script.json"))
			if err != nil {
				t.Fatal("prepare pinned corpus:", err)
			}
			var script struct {
				Commands []struct {
					Type, Filename, Text string
					Line                 int
					ModuleType           string `json:"module_type"`
					Action               struct {
						Type, Field string
						Args        []constant
					}
					Expected []constant
				}
			}
			if err := json.Unmarshal(data, &script); err != nil {
				t.Fatal(err)
			}
			var m *module.Module
			exports := make(map[string]uint32)
			for _, cmd := range script.Commands {
				switch cmd.Type {
				case "module":
					data, err := os.ReadFile(filepath.Join(dir, cmd.Filename))
					if err != nil {
						t.Fatal(err)
					}
					m, err = decode.Decode(data, decode.DefaultLimits())
					if err != nil {
						t.Fatal(err)
					}
					if _, err := validate.Check(m, validate.DefaultLimits()); err != nil {
						t.Fatal(err)
					}
					if len(m.Imports) != 0 {
						t.Fatal("numeric script gained imports")
					}
					for _, e := range m.Exports {
						if e.Kind != 0 {
							t.Fatal("nonfunction export")
						}
						exports[e.Name] = e.Index
					}
				case "assert_invalid":
					data, err := os.ReadFile(filepath.Join(dir, cmd.Filename))
					if err != nil {
						t.Fatal(err)
					}
					bad, err := decode.Decode(data, decode.DefaultLimits())
					if err != nil {
						t.Fatal(err)
					}
					if _, err := validate.Check(bad, validate.DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
						t.Fatalf("line %d: %v", cmd.Line, err)
					}
					invalid++
				case "assert_malformed":
					if cmd.ModuleType != "text" {
						t.Fatal("unexpected binary malformed assertion")
					}
					textOnly++
				case "assert_return", "assert_trap":
					if cmd.Action.Type != "invoke" || len(cmd.Action.Args) > 2 {
						t.Fatal("unexpected numeric action")
					}
					index, ok := exports[cmd.Action.Field]
					if !ok {
						t.Fatal("missing export")
					}
					f := m.Functions[index]
					sig := m.Types[f.Type]
					if len(f.Locals) != 0 || len(sig.Results) != 1 || len(sig.Params) != len(cmd.Action.Args) || len(f.Body) != len(sig.Params)+2 {
						t.Fatalf("%s is no longer a direct numeric primitive", cmd.Action.Field)
					}
					var args [2]uint64
					for i, c := range cmd.Action.Args {
						if f.Body[i].Op != 0x20 || int(f.Body[i].A) != i {
							t.Fatal("numeric argument order changed")
						}
						args[i] = parseBits(t, c)
					}
					op := f.Body[len(sig.Params)].Op
					if !((op >= 0x45 && op <= 0xc4) || (op >= 0xfc0000 && op <= 0xfc0007)) || f.Body[len(f.Body)-1].Op != 0xb {
						t.Fatal("not a numeric primitive")
					}
					seen[op] = true
					got, trap := Eval(op, args[0], args[1])
					if cmd.Type == "assert_trap" {
						traps++
						if trap == wasm.NoTrap || trap.String() != cmd.Text {
							t.Errorf("line %d %s: trap=%v, want %s", cmd.Line, cmd.Action.Field, trap, cmd.Text)
						}
					} else {
						returns++
						if trap != wasm.NoTrap || len(cmd.Expected) != 1 || !matches(t, cmd.Expected[0], got) {
							t.Errorf("line %d %s(%x): got %#x trap=%s want=%v", cmd.Line, cmd.Action.Field, args, got, trap, cmd.Expected)
						}
					}
				default:
					t.Fatalf("unclassified command %s", cmd.Type)
				}
			}
		})
	}
	if returns != 11534 || traps != 87 || invalid != 171 || textOnly != 8 || len(seen) != 136 {
		t.Fatalf("inventory changed: return=%d trap=%d invalid=%d text=%d opcodes=%d", returns, traps, invalid, textOnly, len(seen))
	}
	t.Logf("Executed %d direct numeric assertions (%d return, %d trap), all %d scalar opcodes; %d invalid modules rejected; %d text-parser assertions classified separately.", returns+traps, returns, traps, len(seen), invalid, textOnly)
}
func parseBits(t *testing.T, c constant) uint64 {
	t.Helper()
	n, err := strconv.ParseUint(c.Value, 10, 64)
	if err != nil {
		t.Fatal(c, err)
	}
	return n
}
func matches(t *testing.T, want constant, got uint64) bool {
	t.Helper()
	if want.Value == "nan:canonical" || want.Value == "nan:arithmetic" {
		if want.Type == "f32" {
			got &= 0x7fffffff
			if want.Value == "nan:canonical" {
				return got == 0x7fc00000
			}
			return got&0x7fc00000 == 0x7fc00000
		}
		if want.Type != "f64" {
			t.Fatal("NaN of nonfloat")
		}
		got &= 0x7fffffffffffffff
		if want.Value == "nan:canonical" {
			return got == 0x7ff8000000000000
		}
		return got&0x7ff8000000000000 == 0x7ff8000000000000
	}
	return got == parseBits(t, want)
}

func TestConversionBoundaries(t *testing.T) {
	for _, width := range []int{32, 64} {
		for _, signed := range []bool{false, true} {
			upper := math.Ldexp(1, width)
			lower := float64(0)
			if signed {
				upper /= 2
				lower = -upper
			}
			for _, x := range []float64{math.Inf(-1), math.Nextafter(lower, math.Inf(-1)), lower, math.Nextafter(lower, 0), -0.5, 0, math.Nextafter(upper, 0), upper, math.Inf(1), math.NaN()} {
				_, trap := truncate(x, width, signed, false)
				valid := !math.IsNaN(x) && math.Trunc(x) >= lower && math.Trunc(x) < upper
				if (trap == wasm.NoTrap) != valid {
					t.Fatal(width, signed, x, trap)
				}
				_, trap = truncate(x, width, signed, true)
				if trap != wasm.NoTrap {
					t.Fatal("saturation trapped")
				}
			}
		}
	}
}
func TestInvalidOpcodePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid opcode silently executed")
		}
	}()
	Eval(0, 0, 0)
}

func FuzzNumeric(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(1)<<63, ^uint64(0))
	f.Add(uint64(0x7ff0000000000001), uint64(0xfff0000000000000))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		for op := uint32(0x45); op <= 0xc4; op++ {
			Eval(op, a, b)
		}
		for op := uint32(0xfc0000); op <= 0xfc0007; op++ {
			if _, trap := Eval(op, a, b); trap != wasm.NoTrap {
				t.Fatal("saturation trapped")
			}
		}
	})
}

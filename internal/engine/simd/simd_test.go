package simd

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Each listed script's first module provides pure numeric expressions. Later
// modules test integration (memory, globals and structured control), which the
// complete runtime suite must execute separately. This inventory is deliberate
// primitive scope; a new shape/opcode/count fails instead of being skipped.
var primitiveScripts = []string{
	"simd_bit_shift", "simd_bitwise", "simd_boolean", "simd_conversions",
	"simd_f32x4", "simd_f32x4_arith", "simd_f32x4_cmp", "simd_f32x4_pmin_pmax", "simd_f32x4_rounding",
	"simd_f64x2", "simd_f64x2_arith", "simd_f64x2_cmp", "simd_f64x2_pmin_pmax", "simd_f64x2_rounding",
	"simd_i16x8_arith", "simd_i16x8_arith2", "simd_i16x8_cmp", "simd_i16x8_extadd_pairwise_i8x16", "simd_i16x8_extmul_i8x16", "simd_i16x8_q15mulr_sat_s", "simd_i16x8_sat_arith",
	"simd_i32x4_arith", "simd_i32x4_arith2", "simd_i32x4_cmp", "simd_i32x4_dot_i16x8", "simd_i32x4_extadd_pairwise_i16x8", "simd_i32x4_extmul_i16x8", "simd_i32x4_trunc_sat_f32x4", "simd_i32x4_trunc_sat_f64x2",
	"simd_i64x2_arith", "simd_i64x2_arith2", "simd_i64x2_cmp", "simd_i64x2_extmul_i32x4",
	"simd_i8x16_arith", "simd_i8x16_arith2", "simd_i8x16_cmp", "simd_i8x16_sat_arith", "simd_int_to_int_extend", "simd_lane", "simd_splat",
}

type constant struct {
	Type     string
	LaneType string `json:"lane_type"`
	Value    json.RawMessage
}
type expressionOp struct {
	instruction     module.Instr
	inputs, outputs int
}

func TestOfficialPureSIMDExpressions(t *testing.T) {
	assertions := 0
	seen := make(map[uint32]bool)
	for _, name := range primitiveScripts {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("../../../artifacts/core2/fixtures/simd", name)
			raw, err := os.ReadFile(filepath.Join(dir, "script.json"))
			if err != nil {
				t.Fatal("prepare pinned corpus:", err)
			}
			var script struct {
				Commands []struct {
					Type, Filename string
					Line           int
					Action         struct {
						Type, Field, Module string
						Args                []constant
					}
					Expected []constant
				}
			}
			if err := json.Unmarshal(raw, &script); err != nil {
				t.Fatal(err)
			}
			var checked *validate.Module
			var plans [][]expressionOp
			exports := make(map[string]uint32)
			for _, cmd := range script.Commands {
				if cmd.Type == "module" {
					if checked != nil {
						break
					}
					data, err := os.ReadFile(filepath.Join(dir, cmd.Filename))
					if err != nil {
						t.Fatal(err)
					}
					m, err := decode.Decode(data, decode.DefaultLimits())
					if err != nil {
						t.Fatal(err)
					}
					checked, err = validate.Check(m, validate.DefaultLimits())
					if err != nil {
						t.Fatal(err)
					}
					if len(m.Imports)+len(m.Globals)+len(m.Memories)+len(m.Tables) != 0 {
						t.Fatal("primitive module gained runtime state")
					}
					plans = make([][]expressionOp, len(m.Functions))
					for _, e := range m.Exports {
						if e.Kind != 0 {
							t.Fatal("not a function")
						}
						exports[e.Name] = e.Index
					}
					for fi, f := range m.Functions {
						if len(f.Locals) != 0 {
							t.Fatal("primitive function gained local state")
						}
						err := checked.WalkFunction(uint32(fi), func(pc int, in module.Instr, inputs, outputs []wasm.ValueType) error {
							if in.Op == 0xb && pc == len(f.Body)-1 {
								return nil
							}
							if !(in.Op == 0x20 || in.Op >= 0x41 && in.Op <= 0xc4 || in.Op>>16 == 0xfd && pureOpcode(in.Op&0xffff)) {
								t.Fatalf("primitive expression gained opcode %#x", in.Op)
							}
							plans[fi] = append(plans[fi], expressionOp{in, len(inputs), len(outputs)})
							return nil
						})
						if err != nil {
							t.Fatal(err)
						}
					}
					continue
				}
				if cmd.Type == "assert_invalid" || cmd.Type == "assert_malformed" {
					continue
				} // Covered by the full frontend corpus test.
				if cmd.Type != "assert_return" || cmd.Action.Type != "invoke" || cmd.Action.Module != "" {
					t.Fatalf("unclassified primitive command: %s", cmd.Type)
				}
				fi, ok := exports[cmd.Action.Field]
				if !ok {
					t.Fatal("missing export")
				}
				var stack []Value
				args := make([]Value, len(cmd.Action.Args))
				for i, c := range cmd.Action.Args {
					args[i] = parseConstant(t, c)
				}
				for _, op := range plans[fi] {
					in := op.instruction
					var r Value
					switch {
					case in.Op == 0x20:
						r = args[in.A]
					case in.Op >= 0x41 && in.Op <= 0x44:
						r.Lo = in.Bits
					default:
						n := len(stack) - op.inputs
						var operands [3]Value
						copy(operands[:], stack[n:])
						stack = stack[:n]
						if in.Op>>16 == 0xfd {
							seen[in.Op&0xffff] = true
							r = Eval(&in, operands[0], operands[1], operands[2])
						} else {
							var trap wasm.TrapCode
							r.Lo, trap = numeric.Eval(in.Op, operands[0].Lo, operands[1].Lo)
							if trap != wasm.NoTrap {
								t.Fatal(trap)
							}
						}
					}
					if op.outputs != 1 {
						t.Fatal("non-expression operation")
					}
					stack = append(stack, r)
				}
				if len(stack) != len(cmd.Expected) {
					t.Fatal("expression result count")
				}
				for i, want := range cmd.Expected {
					if !matches(t, want, stack[i]) {
						t.Fatalf("line %d %s: got %#v, want %s", cmd.Line, cmd.Action.Field, stack[i], want.Value)
					}
				}
				assertions++
			}
		})
	}
	if len(primitiveScripts) != 40 || assertions != 22924 || len(seen) != 214 {
		t.Fatalf("primitive inventory changed: scripts=%d assertions=%d SIMD opcodes=%d", len(primitiveScripts), assertions, len(seen))
	}
	t.Logf("Executed %d pure expression assertions covering %d nonmemory SIMD opcodes; runtime integration is not claimed.", assertions, len(seen))
}

func laneWidth(t *testing.T, name string) uint {
	t.Helper()
	switch name {
	case "i8":
		return 8
	case "i16":
		return 16
	case "i32", "f32":
		return 32
	case "i64", "f64":
		return 64
	default:
		t.Fatal("unknown lane type", name)
		return 0
	}
}
func parseBits(t *testing.T, s string) uint64 {
	t.Helper()
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func constants(t *testing.T, c constant) (uint, []string) {
	t.Helper()
	if c.Type == "v128" {
		var xs []string
		if err := json.Unmarshal(c.Value, &xs); err != nil {
			t.Fatal(err)
		}
		w := laneWidth(t, c.LaneType)
		if len(xs) != int(128/w) {
			t.Fatal("wrong lanes")
		}
		return w, xs
	}
	var s string
	if err := json.Unmarshal(c.Value, &s); err != nil {
		t.Fatal(err)
	}
	return laneWidth(t, c.Type), []string{s}
}
func parseConstant(t *testing.T, c constant) Value {
	t.Helper()
	w, xs := constants(t, c)
	var raw [16]byte
	for i, s := range xs {
		x := parseBits(t, s)
		for j := uint(0); j < w/8; j++ {
			raw[i*int(w/8)+int(j)] = byte(x >> (8 * j))
		}
	}
	return Value{Lo: binary.LittleEndian.Uint64(raw[:8]), Hi: binary.LittleEndian.Uint64(raw[8:])}
}
func matches(t *testing.T, c constant, got Value) bool {
	t.Helper()
	w, xs := constants(t, c)
	var raw [16]byte
	binary.LittleEndian.PutUint64(raw[:8], got.Lo)
	binary.LittleEndian.PutUint64(raw[8:], got.Hi)
	for i, s := range xs {
		var x uint64
		for j := uint(0); j < w/8; j++ {
			x |= uint64(raw[i*int(w/8)+int(j)]) << (8 * j)
		}
		if s == "nan:canonical" || s == "nan:arithmetic" {
			sign, canonical := uint64(0x7fffffff), uint64(0x7fc00000)
			if w == 64 {
				sign, canonical = 0x7fffffffffffffff, 0x7ff8000000000000
			}
			x &= sign
			if s == "nan:canonical" {
				if x != canonical {
					return false
				}
			} else if x&canonical != canonical {
				return false
			}
		} else if x != parseBits(t, s) {
			return false
		}
	}
	return true
}

func TestPureOpcodeBoundary(t *testing.T) {
	count := 0
	for op := uint32(0); op <= 0x130; op++ {
		if pureOpcode(op) {
			count++
			continue
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted non-pure opcode %d", op)
				}
			}()
			Eval(&module.Instr{Op: 0xfd0000 | op}, Value{}, Value{}, Value{})
		}()
	}
	if count != 234 {
		t.Fatal(count)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("wrong prefix")
		}
	}()
	Eval(&module.Instr{Op: 12}, Value{}, Value{}, Value{})
}
func FuzzPureSIMD(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(^uint64(0), uint64(0x8000000080000000), uint64(0x7ff0000000000001), uint64(31))
	f.Fuzz(func(t *testing.T, a, b, c, d uint64) {
		for op := uint32(12); op <= 0x113; op++ {
			if pureOpcode(op) {
				Eval(&module.Instr{Op: 0xfd0000 | op}, Value{Lo: a, Hi: b}, Value{Lo: c, Hi: d}, Value{Lo: d, Hi: a})
			}
		}
	})
}

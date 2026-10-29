package codegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type scalarSpecValue struct{ Type, Value string }
type scalarSpecAction struct {
	Type, Field, Module string
	Args                []scalarSpecValue
}
type scalarSpecCommand struct {
	Type, Filename, Text string
	Line                 int
	Action               scalarSpecAction
	Expected             []scalarSpecValue
}
type scalarExecutable struct {
	compiled *CompiledFunction
	lease    *native.Lease
	context  *native.Context
}

func scalarExpected(t *testing.T, expected scalarSpecValue, got value.Value) bool {
	t.Helper()
	if got.Hi != 0 {
		return false
	}
	if expected.Value == "nan:canonical" || expected.Value == "nan:arithmetic" {
		typ := wasm.F32
		if expected.Type == "f64" {
			typ = wasm.F64
		} else if expected.Type != "f32" {
			t.Fatal("non-floating NaN assertion")
		}
		kind := nanKind(typ, got.Lo)
		return kind == 1 || expected.Value == "nan:arithmetic" && kind == 2
	}
	bits, err := strconv.ParseUint(expected.Value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return got.Lo == bits
}

// Execute complete selected scalar scripts; imports, calls, globals and SIMD
// still require the Store driver. Their suites are not silently filtered.
func TestPinnedScalarNativeExecution(t *testing.T) {
	if !native.Available {
		return
	}
	names := []string{"i32", "i64", "f32", "f64", "f32_bitwise", "f64_bitwise", "f32_cmp", "f64_cmp", "conversions", "float_literals", "store", "memory", "memory_size"}
	for _, fast := range []bool{false, true} {
		assertions, modules, frontend := 0, 0, 0
		for _, name := range names {
			directory := filepath.Join("../../../artifacts/core2/fixtures", name)
			data, err := os.ReadFile(filepath.Join(directory, "script.json"))
			if err != nil {
				t.Fatal(err)
			}
			var script struct{ Commands []scalarSpecCommand }
			if err = json.Unmarshal(data, &script); err != nil {
				t.Fatal(err)
			}
			exports := map[string]scalarExecutable{}
			var active []scalarExecutable
			var memory *native.Memory
			closeModule := func() {
				for _, ex := range active {
					ex.context.Close()
					ex.lease.Close()
				}
				active = nil
				if memory != nil {
					memory.Close()
					memory = nil
				}
			}
			for _, command := range script.Commands {
				switch command.Type {
				case "module":
					closeModule()
					exports = map[string]scalarExecutable{}
					data, err := os.ReadFile(filepath.Join(directory, command.Filename))
					if err != nil {
						t.Fatal(err)
					}
					m, err := decode.Decode(data, decode.DefaultLimits())
					if err != nil {
						t.Fatal(err)
					}
					if len(m.Imports) != 0 || m.HasStart {
						t.Fatal(name, "scalar scope includes imports/start")
					}
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
					config := Config{Unmetered: fast}
					if len(m.Memories) > 1 {
						t.Fatal("multiple memories")
					}
					if len(m.Memories) != 0 {
						limits := m.Memories[0]
						config.MemoryMinimumPages = uint32(limits.Min)
						maximum := uint64(256)
						if limits.HasMax {
							maximum = min(maximum, uint64(limits.Max))
						}
						memory, err = native.NewMemory(uint64(limits.Min)*65536, maximum*65536)
						if err != nil {
							t.Fatal(name, command.Line, err)
						}
						t.Cleanup(memory.Close)
						for _, segment := range m.Data {
							if segment.Mode != 0 || segment.Memory != 0 || len(segment.Offset) != 2 || segment.Offset[0].Op != 0x41 {
								t.Fatal("unsupported test data initialization")
							}
							if err = memory.Write(uint64(uint32(segment.Offset[0].Bits)), segment.Bytes); err != nil {
								t.Fatal(err)
							}
						}
					}
					for _, f := range code.Functions {
						compiled, lease, ctx := configuredContext(t, f, config)
						if memory != nil {
							if err = ctx.BindMemory(memory); err != nil {
								t.Fatal(err)
							}
						}
						if compiled.Helpers != 0 || compiled.Calls != 0 || compiled.EdgeCopies != 0 {
							t.Fatal(name, "scalar script requires unfinished runtime exits")
						}
						active = append(active, scalarExecutable{compiled, lease, ctx})
					}
					for _, export := range m.Exports {
						// These scripts invoke functions only; any global access in a
						// function is rejected by the helper inventory above.
						if export.Kind == 2 && memory != nil || export.Kind == 3 {
							continue
						}
						if export.Kind != 0 {
							t.Fatal(name, "non-function scalar export")
						}
						exports[export.Name] = active[export.Index]
					}
					modules++
				case "assert_return", "assert_trap":
					if command.Action.Type != "invoke" || command.Action.Module != "" {
						t.Fatal(name, command.Line, "unexpected scalar action")
					}
					ex, ok := exports[command.Action.Field]
					if !ok {
						t.Fatal(name, command.Line, "unknown export")
					}
					args := make([]value.Value, len(command.Action.Args))
					for i, arg := range command.Action.Args {
						bits, err := strconv.ParseUint(arg.Value, 10, 64)
						if err != nil {
							t.Fatal(err)
						}
						args[i].Lo = bits
					}
					out, exit, state := runFunction(t, ex.compiled, ex.lease, ex.context, args, 1000000, !fast, 4096, memory)
					if command.Type == "assert_trap" {
						traps := map[string]wasm.TrapCode{"integer divide by zero": wasm.TrapIntegerDivideByZero, "integer overflow": wasm.TrapIntegerOverflow, "invalid conversion to integer": wasm.TrapInvalidConversion}
						trap, ok := traps[command.Text]
						if !ok || exit != native.Trap || state.Payload[0] != uint64(trap) {
							t.Fatal(name, command.Line, fast, command.Text, exit, state)
						}
					} else {
						if exit != native.Return || len(out) != len(command.Expected) {
							t.Fatal(name, command.Line, fast, exit, out)
						}
						for i, expected := range command.Expected {
							if !scalarExpected(t, expected, out[i]) {
								t.Fatal(name, command.Line, fast, command.Action.Field, args, out[i], expected)
							}
						}
					}
					assertions++
				case "assert_invalid", "assert_malformed":
					// Frontend rejection is covered by the complete pinned decoder /
					// validator suite; it is not counted as native execution.
					frontend++
				default:
					t.Fatal(name, command.Line, "unaccounted scalar command", command.Type)
				}
			}
			closeModule()
		}
		if assertions != 12068 || modules != 27 || frontend != 347 {
			t.Fatal("pinned scalar execution scope changed", assertions, modules, frontend)
		}
		t.Logf("unmetered=%v: %d native assertions across %d modules / %d scripts; %d frontend-only commands excluded", fast, assertions, modules, len(names), frontend)
	}
}

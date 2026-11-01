package component

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type upstreamValue struct {
	Type  string
	Value json.RawMessage
}
type upstreamCommand struct {
	Type, Filename, Name, Module, Instance string
	Line                                   int
	Action                                 struct {
		Type, Field, Module string
		Args                []upstreamValue
	}
	Expected []upstreamValue
}

func upstreamDecodeValue(t *testing.T, typ *model.Type, raw upstreamValue) Value {
	t.Helper()
	decode := func(out any) {
		t.Helper()
		if err := json.Unmarshal(raw.Value, out); err != nil {
			t.Fatal(err)
		}
	}
	var out Value
	switch typ.Kind {
	case model.Bool:
		var b bool
		decode(&b)
		if b {
			out.Bits = 1
		}
	case model.Char:
		var s string
		decode(&s)
		r := []rune(s)
		if len(r) != 1 {
			t.Fatal("invalid character", s)
		}
		out.Bits = uint64(r[0])
	case model.String:
		decode(&out.Text)
	case model.S8, model.S16, model.S32, model.S64:
		var s string
		decode(&s)
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		out.Bits = uint64(n)
	case model.U8, model.U16, model.U32, model.U64, model.F32, model.F64:
		var s string
		decode(&s)
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		out.Bits = n
	case model.Enum, model.Flags:
		labels := []string{}
		if typ.Kind == model.Enum {
			var s string
			decode(&s)
			labels = append(labels, s)
		} else {
			decode(&labels)
		}
		for _, label := range labels {
			found := false
			for i, name := range typ.Labels {
				if name == label {
					found = true
					if typ.Kind == model.Enum {
						out.Bits = uint64(i)
					} else {
						out.Bits |= 1 << i
					}
					break
				}
			}
			if !found {
				t.Fatal("unknown label", label)
			}
		}
	case model.List, model.Tuple:
		var values []upstreamValue
		decode(&values)
		for i, value := range values {
			child := typ.Element
			if typ.Kind == model.Tuple {
				child = typ.Fields[i].Type
			}
			out.Elements = append(out.Elements, upstreamDecodeValue(t, child, value))
		}
	case model.Record:
		var fields []json.RawMessage
		decode(&fields)
		if len(fields) != len(typ.Fields) {
			t.Fatal("record field count")
		}
		for i, field := range fields {
			var pair []json.RawMessage
			if err := json.Unmarshal(field, &pair); err != nil || len(pair) != 2 {
				t.Fatal("record pair", err)
			}
			var name string
			var value upstreamValue
			json.Unmarshal(pair[0], &name)
			if name != typ.Fields[i].Name {
				t.Fatal(name)
			}
			if err := json.Unmarshal(pair[1], &value); err != nil {
				t.Fatal(err)
			}
			out.Elements = append(out.Elements, upstreamDecodeValue(t, typ.Fields[i].Type, value))
		}
	case model.Variant, model.Option, model.Result:
		var payload *upstreamValue
		switch typ.Kind {
		case model.Variant:
			var value struct {
				Case    string
				Payload *upstreamValue
			}
			decode(&value)
			found := false
			for i, c := range typ.Cases {
				if c.Name == value.Case {
					out.Case = uint32(i)
					found = true
					break
				}
			}
			if !found {
				t.Fatal("variant case", value.Case)
			}
			payload = value.Payload
		case model.Option:
			if string(raw.Value) != "null" {
				out.Case = 1
				decode(&payload)
			}
		case model.Result:
			var cases map[string]*upstreamValue
			decode(&cases)
			if len(cases) != 1 {
				t.Fatal("result case")
			}
			if value, ok := cases["Err"]; ok {
				out.Case = 1
				payload = value
			} else {
				payload = cases["Ok"]
			}
		}
		if payload != nil {
			child := typ.Element
			if typ.Kind != model.Option {
				child = typ.Cases[out.Case].Type
			}
			out.Elements = []Value{upstreamDecodeValue(t, child, *payload)}
		}
	default:
		t.Fatalf("unhandled upstream value kind %x (%s)", typ.Kind, raw.Type)
	}
	return out
}

func upstreamEqualValue(a, b Value) bool {
	if a.Bits != b.Bits || a.Text != b.Text || a.Case != b.Case || a.Resource != nil || b.Resource != nil || len(a.Elements) != len(b.Elements) {
		return false
	}
	for i := range a.Elements {
		if !upstreamEqualValue(a.Elements[i], b.Elements[i]) {
			return false
		}
	}
	return true
}

func TestUpstreamComponentRuntime(t *testing.T) {
	root := "../../artifacts/components/runtime-corpus/runtime"
	manifestData, err := os.ReadFile("../../testdata/component-runtime/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Scripts map[string]struct {
			Selected []int          `json:"selectedCommands"`
			Deferred map[string]any `json:"deferredCommands"`
			Commands int
		}
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Scripts) != 57 {
		t.Fatal("changed runtime inventory", len(manifest.Scripts))
	}
	var names []string
	for name := range manifest.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, backend := range []string{"interpreter", "native"} {
		for _, name := range names {
			row := manifest.Scripts[name]
			if len(row.Selected) == 0 {
				t.Logf("%s: %d commands require the separately scoped component-threading profile", name, len(row.Deferred))
				continue
			}
			path := filepath.Join(root, name, "script.json")
			t.Run(backend+"/"+name, func(t *testing.T) {
				config := DefaultConfig()
				config.Core.Backend = backend
				engine, err := NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer engine.Close()
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var script struct{ Commands []upstreamCommand }
				if err := json.Unmarshal(data, &script); err != nil {
					t.Fatal(err)
				}
				modules := map[string]*Compiled{}
				instances := map[string]*Instance{}
				var current *Instance
				if len(script.Commands) != row.Commands {
					t.Fatal("changed command inventory")
				}
				for _, index := range row.Selected {
					if index < 0 || index >= len(script.Commands) {
						t.Fatal("command index", index)
					}
					command := script.Commands[index]
					where := fmt.Sprintf("line %d %s", command.Line, command.Type)
					switch command.Type {
					case "module", "module_definition", "assert_invalid", "assert_uninstantiable":
						data, err := os.ReadFile(filepath.Join(filepath.Dir(path), command.Filename))
						if err != nil {
							t.Fatal(where, err)
						}
						compiled, err := engine.Compile(nil, data)
						if command.Type == "assert_invalid" {
							if !errors.Is(err, wasm.ErrInvalid) {
								t.Fatal(where, "expected invalid", err)
							}
							continue
						}
						if err != nil {
							t.Fatal(where, "compile", err)
						}
						t.Cleanup(compiled.Close)
						modules[command.Name] = compiled
						if command.Type == "module_definition" {
							continue
						}
						in, err := compiled.Instantiate(nil, Imports{}, CallOptions{})
						if command.Type == "assert_uninstantiable" {
							if err == nil {
								in.Close()
								t.Fatal(where, "expected initialization trap")
							}
							continue
						}
						if err != nil {
							t.Fatal(where, "instantiate", err)
						}
						current = in
						instances[command.Name] = in
						t.Cleanup(func() {
							in.Close()
							if err := in.WaitClosed(nil); err != nil {
								t.Error(err)
							}
						})
					case "module_instance":
						compiled := modules[command.Module]
						if compiled == nil {
							t.Fatal(where, "unknown module", command.Module)
						}
						in, err := compiled.Instantiate(nil, Imports{}, CallOptions{})
						if err != nil {
							t.Fatal(where, "instantiate", err)
						}
						current = in
						instances[command.Instance] = in
						t.Cleanup(func() {
							in.Close()
							if err := in.WaitClosed(nil); err != nil {
								t.Error(err)
							}
						})
					case "assert_return", "assert_trap", "action":
						in := current
						if command.Action.Module != "" {
							in = instances[command.Action.Module]
						}
						if in == nil || command.Action.Type != "invoke" {
							t.Fatal(where, "missing invocation")
						}
						f, err := in.Function(command.Action.Field)
						if err != nil {
							t.Fatal(where, err)
						}
						if len(command.Action.Args) != len(f.typ.Parameters) {
							t.Fatal(where, "argument count")
						}
						args := make([]Value, len(command.Action.Args))
						for i, v := range command.Action.Args {
							args[i] = upstreamDecodeValue(t, f.typ.Parameters[i].Type, v)
						}
						ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
						result, err := f.Call(ctx, args, CallOptions{Metered: true, Fuel: 10000000})
						cancel()
						if command.Type == "assert_trap" {
							if !errors.Is(err, wasm.ErrTrap) {
								t.Fatal(where, "expected semantic trap", err)
							}
							continue
						}
						if err != nil {
							t.Fatal(where, command.Action.Field, err)
						}
						if command.Type == "action" {
							continue
						}
						if len(command.Expected) == 0 {
							if result.Value != nil {
								t.Fatal(where, "unexpected result")
							}
							continue
						}
						if len(command.Expected) != 1 || result.Value == nil || f.typ.Result == nil {
							t.Fatal(where, "result count")
						}
						expected := upstreamDecodeValue(t, f.typ.Result, command.Expected[0])
						if !upstreamEqualValue(*result.Value, expected) {
							t.Fatal(where, "result", *result.Value, "expected", expected)
						}
					default:
						t.Fatal(where, "unhandled command")
					}
				}
			})
		}
	}
}

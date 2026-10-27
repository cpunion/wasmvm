package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func leb(n int) []byte {
	var b []byte
	for {
		v := byte(n & 127)
		n >>= 7
		if n != 0 {
			v |= 128
		}
		b = append(b, v)
		if n == 0 {
			return b
		}
	}
}
func section(id byte, payload []byte) []byte {
	out := append([]byte{id}, leb(len(payload))...)
	return append(out, payload...)
}
func binary(sections ...[]byte) []byte {
	out := []byte{0, 97, 115, 109, 1, 0, 0, 0}
	for _, s := range sections {
		out = append(out, s...)
	}
	return out
}
func functionModule(params, results, body []byte) []byte {
	typ := append([]byte{1, 0x60, byte(len(params))}, params...)
	typ = append(typ, byte(len(results)))
	typ = append(typ, results...)
	code := append([]byte{0}, body...)
	code = append(code, 0xb)
	codes := append([]byte{1}, leb(len(code))...)
	codes = append(codes, code...)
	return binary(section(1, typ), section(3, []byte{1, 0}), section(7, []byte{1, 3, 'r', 'u', 'n', 0, 0}), section(10, codes))
}
func writeModule(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "module.wasm")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func invoke(t *testing.T, ctx context.Context, args ...string) (int, string, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := Main(ctx, args, &out, &err)
	return code, out.String(), err.String()
}
func jsonObject(t *testing.T, s string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatal(s, err)
	}
	return out
}

func TestValidateInspectAndRun(t *testing.T) {
	path := writeModule(t, functionModule([]byte{0x7f, 0x7f}, []byte{0x7f}, []byte{0x20, 0, 0x20, 1, 0x6a}))
	code, out, err := invoke(t, nil, "validate", path)
	if code != 0 || err != "" || jsonObject(t, out)["valid"] != true {
		t.Fatal(code, out, err)
	}
	code, out, err = invoke(t, nil, "inspect", path)
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	info := jsonObject(t, out)
	if info["backend"] != "interpreter" || info["profile"] != "core2" || info["definedFunctions"] != float64(1) {
		t.Fatal(info)
	}
	exports := info["exports"].([]any)
	fn := exports[0].(map[string]any)
	if !reflect.DeepEqual(fn["params"], []any{"i32", "i32"}) || !reflect.DeepEqual(fn["results"], []any{"i32"}) {
		t.Fatal(fn)
	}
	code, out, err = invoke(t, nil, "run", "--export", "run", "--fuel", "3", path, "i32:17", "i32:25")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	result := jsonObject(t, out)
	if result["fuelRemaining"] != float64(0) || result["results"].([]any)[0].(map[string]any)["value"] != "42" {
		t.Fatal(result)
	}
	code, out, err = invoke(t, nil, "run", "--timeout", "0", "--export", "run", path, "i32:-1", "i32:0x1")
	if code != 0 || err != "" || strings.Contains(out, "fuelRemaining") || !strings.Contains(out, `"bits": "0x00000000"`) {
		t.Fatal(code, out, err)
	}
	code, _, err = invoke(t, nil, "run", "--export", "run", "--fuel", "2", path, "i32:1", "i32:2")
	if code != 1 || !strings.Contains(err, "fuel") {
		t.Fatal(code, err)
	}
	for _, args := range [][]string{{}, {"help"}, {"--help"}, {"-h"}, {"validate", "--help"}, {"bad-command"}, {"run", path}, {"validate", path, path}, {"inspect", "--bad", path}, {"run", "--export", "run", "--fuel", "-1", path}, {"validate", "--timeout", "-1s", path}, {"run", "--export", "run", path, "x"}} {
		code, _, _ := invoke(t, nil, args...)
		want := 2
		if len(args) == 1 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") || len(args) == 2 && args[1] == "--help" {
			want = 0
		}
		if code != want {
			t.Fatal(args, code, want)
		}
	}
}

func TestInspectBoundsRepeatedSignaturesBeforeJSONExpansion(t *testing.T) {
	// Both binaries are small and valid. A large signature shared by many
	// imports/exports must not expand without a presentation resource limit.
	const params = 1024
	typ := append([]byte{1, 0x60}, leb(params)...)
	typ = append(typ, bytes.Repeat([]byte{0x7f}, params)...)
	typ = append(typ, 0)
	for _, repeatedImports := range []bool{false, true} {
		count := maxInspectionTypeValues/params + 1
		imports := []byte{1, 1, 'h', 1, 'f', 0, 0}
		exports := leb(count)
		if repeatedImports {
			imports = leb(count)
		}
		for i := 0; i < count; i++ {
			if repeatedImports {
				imports = append(imports, 1, 'h', 1, 'f', 0, 0)
			} else {
				name := strconv.Itoa(i)
				exports = append(exports, leb(len(name))...)
				exports = append(exports, name...)
				exports = append(exports, 0, 0)
			}
		}
		data := binary(section(1, typ), section(2, imports))
		if !repeatedImports {
			data = append(data, section(7, exports)...)
		}
		path := writeModule(t, data)
		if code, _, err := invoke(t, nil, "validate", path); code != 0 {
			t.Fatal(code, err)
		}
		if code, out, err := invoke(t, nil, "inspect", path); code != 1 || out != "" || !strings.Contains(err, "expanded inspection signature values") {
			t.Fatal(code, len(out), err)
		}
	}
}

func TestInspectSignatureExpansionIncludesResultsAndExactLimit(t *testing.T) {
	params := make([]wasm.ValueType, maxInspectionTypeValues/2)
	for i := range params {
		params[i] = wasm.I32
	}
	signature := &vm.FunctionType{Params: params, Results: params}
	info := vm.ModuleInfo{Exports: []vm.ExportInfo{{Name: "f", ExternType: vm.ExternType{Kind: vm.FunctionKind, Function: signature}}}}
	out, err := describe(info)
	if err != nil || len(out["exports"].([]map[string]any)[0]["results"].([]string)) != len(params) {
		t.Fatal("exact expansion limit rejected", err)
	}
	signature.Results = append(signature.Results, wasm.I32)
	if out, err := describe(info); out != nil || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("result expansion exceeds budget", err)
	}
}

func TestTypedValuesRoundTrip(t *testing.T) {
	types := []byte{0x7f, 0x7e, 0x7d, 0x7c, 0x7b, 0x70, 0x6f}
	body := []byte{}
	for i := range types {
		body = append(body, 0x20, byte(i))
	}
	path := writeModule(t, functionModule(types, types, body))
	code, out, err := invoke(t, nil, "run", "--export", "run", path, "i32:-2147483648", "i64:0xffffffffffffffff", "f32bits:0x7f800001", "f64bits:0x8000000000000000", "v128:0x1234,0xabcd", "funcref:null", "externref:null")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	values := jsonObject(t, out)["results"].([]any)
	wantBits := []string{"0x80000000", "0xffffffffffffffff", "0x7f800001", "0x8000000000000000", "0x0000000000001234,0x000000000000abcd"}
	for i, want := range wantBits {
		if values[i].(map[string]any)["bits"] != want {
			t.Fatal(i, values[i])
		}
	}
	if values[5].(map[string]any)["value"] != "null" || values[6].(map[string]any)["value"] != "null" {
		t.Fatal(values)
	}
	for _, input := range []string{"i64:-1", "f32:1.25", "f64:-Inf", "f32:NaN"} {
		v, err := parseValue(input)
		if err != nil || formatValue(v).Type == "" {
			t.Fatal(input, v, err)
		}
	}
	for _, input := range []string{"bad:1", "i32:4294967296", "i32:-2147483649", "i64:18446744073709551616", "f32:no", "f64:no", "f32bits:no", "f64bits:no", "v128:1", "v128:no,1", "v128:1,no", "funcref:7", "externref:7"} {
		if _, err := parseValue(input); err == nil {
			t.Fatal(input)
		}
	}
	e := vm.NewEngine()
	s, _ := e.NewStore(vm.DefaultStoreConfig())
	defer s.Close()
	v, _ := s.NewExternRef(1)
	if formatValue(v).Value != "non-null" {
		t.Fatal(v)
	}
	v.Release()
}

func TestErrorsAndNoImplicitHostImports(t *testing.T) {
	valid := functionModule(nil, nil, nil)
	path := writeModule(t, valid)
	cases := [][]string{{"validate", "--backend", "unknown", path}, {"validate", "--profile", "core3", path}, {"inspect", filepath.Join(t.TempDir(), "absent")}, {"run", "--export", "missing", path}, {"run", "--export", "run", path, "i32:1"}}
	for _, args := range cases {
		if code, out, err := invoke(t, nil, args...); code != 1 || out != "" || err == "" {
			t.Fatal(args, code, out, err)
		}
	}
	malformed := writeModule(t, []byte{1})
	invalid := writeModule(t, functionModule(nil, nil, []byte{0x1a}))
	trap := writeModule(t, functionModule(nil, nil, []byte{0}))
	for _, command := range []string{"validate", "inspect", "run"} {
		for _, path := range []string{malformed, invalid} {
			args := []string{command}
			if command == "run" {
				args = append(args, "--export", "run")
			}
			args = append(args, path)
			if code, _, err := invoke(t, nil, args...); code != 1 || err == "" {
				t.Fatal(command, code, err)
			}
		}
	}
	if code, _, err := invoke(t, nil, "run", "--export", "run", trap); code != 1 || !strings.Contains(err, "unreachable") {
		t.Fatal(code, err)
	}
	imports := binary(section(1, []byte{1, 0x60, 0, 0}), section(2, []byte{1, 1, 'm', 1, 'f', 0, 0}), section(7, []byte{1, 3, 'r', 'u', 'n', 0, 0}))
	importPath := writeModule(t, imports)
	if code, _, err := invoke(t, nil, "validate", importPath); code != 0 {
		t.Fatal(code, err)
	}
	if code, out, err := invoke(t, nil, "inspect", importPath); code != 0 || !strings.Contains(out, `"module": "m"`) {
		t.Fatal(code, out, err)
	}
	if code, _, err := invoke(t, nil, "run", "--export", "run", importPath); code != 1 || !strings.Contains(err, "link error") {
		t.Fatal(code, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code, _, err := invoke(t, ctx, "validate", path); code != 1 || !strings.Contains(err, "canceled") {
		t.Fatal(code, err)
	}
	loop := writeModule(t, functionModule(nil, nil, []byte{3, 0x40, 0xc, 0, 0xb}))
	if code, _, err := invoke(t, nil, "run", "--timeout", "5ms", "--export", "run", loop); code != 1 || !strings.Contains(err, "deadline") {
		t.Fatal(code, err)
	}
}

func TestInspectAllObjectKindsAndStartBudget(t *testing.T) {
	objects := binary(section(4, []byte{1, 0x70, 1, 1, 2}), section(5, []byte{1, 1, 1, 2}), section(6, []byte{1, 0x7f, 1, 0x41, 0, 0xb}), section(7, []byte{3, 1, 't', 1, 0, 1, 'm', 2, 0, 1, 'g', 3, 0}))
	code, out, err := invoke(t, nil, "inspect", writeModule(t, objects))
	if code != 0 {
		t.Fatal(out, err)
	}
	for _, kind := range []string{"table", "memory", "global"} {
		if !strings.Contains(out, `"kind": "`+kind+`"`) {
			t.Fatal(out)
		}
	}
	imports := binary(section(1, []byte{1, 0x60, 0, 0}), section(2, []byte{4, 1, 'm', 1, 'f', 0, 0, 1, 'm', 1, 't', 1, 0x70, 0, 1, 1, 'm', 1, 'm', 2, 0, 1, 1, 'm', 1, 'g', 3, 0x7f, 0}), section(7, []byte{4, 1, 'f', 0, 0, 1, 't', 1, 0, 1, 'm', 2, 0, 1, 'g', 3, 0}))
	code, out, err = invoke(t, nil, "inspect", writeModule(t, imports))
	if code != 0 || strings.Contains(out, `"max"`) {
		t.Fatal(code, out, err)
	}
	// Start executes two nops; run executes one nop. One CLI budget covers both.
	start := binary(section(1, []byte{1, 0x60, 0, 0}), section(3, []byte{2, 0, 0}), section(7, []byte{1, 3, 'r', 'u', 'n', 0, 0}), section(8, []byte{1}), section(10, []byte{2, 3, 0, 1, 0xb, 4, 0, 1, 1, 0xb}))
	path := writeModule(t, start)
	for _, fuel := range []string{"0", "1", "2"} {
		if code, _, err := invoke(t, nil, "run", "--export", "run", "--fuel", fuel, path); code != 1 || !strings.Contains(err, "fuel") {
			t.Fatal(fuel, code, err)
		}
	}
	if code, out, err := invoke(t, nil, "run", "--export", "run", "--fuel", "3", path); code != 0 || !strings.Contains(out, `"fuelRemaining": 0`) {
		t.Fatal(code, out, err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestBoundedInputAndOutputFailure(t *testing.T) {
	path := writeModule(t, functionModule(nil, nil, nil))
	var stderr bytes.Buffer
	if code := Main(nil, []string{"validate", path}, failingWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "closed pipe") {
		t.Fatal(code, stderr.String())
	}
	if _, err := readModule(path, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := readModule(t.TempDir(), 100); err == nil {
		t.Fatal("read directory")
	}
}

func TestNativeCommand(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	path := writeModule(t, functionModule([]byte{0x7f, 0x7f}, []byte{0x7f}, []byte{0x20, 0, 0x20, 1, 0x6a}))
	code, out, err := invoke(t, nil, "run", "--backend", "native", "--export", "run", "--fuel", "3", path, "i32:17", "i32:25")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	result := jsonObject(t, out)
	if result["backend"] != "native" || result["fuelRemaining"] != float64(0) || !strings.Contains(out, "0x0000002a") {
		t.Fatal(result)
	}
}

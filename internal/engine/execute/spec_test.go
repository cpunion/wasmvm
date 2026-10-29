package execute

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type specValue struct {
	Type     string
	LaneType string `json:"lane_type"`
	Values   []specValue
	Value    json.RawMessage
}
type specAction struct {
	Type, Module, Field string
	Args                []specValue
}
type specCommand struct {
	Type, Name, As, Filename, Text string
	Instance, Module               string
	Line                           int
	ModuleType                     string `json:"module_type"`
	Action                         specAction
	Expected                       []specValue
	Either                         []specValue
}
type specRunner struct {
	store             *state.Store
	context           context.Context
	last              *state.Instance
	named, registered map[string]*state.Instance
	spectest          map[string]state.Extern
	definitions       map[string][]byte
	tier              int
	features          wasm.Features
	modules           map[*slots.Module]*nativeexec.Lease
	t                 *testing.T
}

func newSpec(t *testing.T, ctx context.Context, tier int) *specRunner {
	t.Helper()
	s := state.New(state.DefaultLimits())
	if tier != 0 {
		s = state.NewNative(state.DefaultLimits())
	}
	t.Cleanup(s.Close)
	r := &specRunner{tier: tier, t: t, modules: make(map[*slots.Module]*nativeexec.Lease), definitions: make(map[string][]byte), store: s, context: ctx, named: make(map[string]*state.Instance), registered: make(map[string]*state.Instance), spectest: make(map[string]state.Extern)}
	for _, fn := range []struct {
		name   string
		params []wasm.ValueType
	}{{"print", nil}, {"print_i32", []wasm.ValueType{wasm.I32}}, {"print_i64", []wasm.ValueType{wasm.I64}}, {"print_f32", []wasm.ValueType{wasm.F32}}, {"print_f64", []wasm.ValueType{wasm.F64}}, {"print_i32_f32", []wasm.ValueType{wasm.I32, wasm.F32}}, {"print_f64_f64", []wasm.ValueType{wasm.F64, wasm.F64}}} {
		ex, err := s.NewHost(module.FuncType{Params: fn.params}, fn.name)
		if err != nil {
			t.Fatal(err)
		}
		r.spectest[fn.name] = ex
	}
	for _, g := range []struct {
		name string
		typ  wasm.ValueType
		bits uint64
	}{{"global_i32", wasm.I32, 666}, {"global_i64", wasm.I64, 666}, {"global_f32", wasm.F32, uint64(math.Float32bits(666.6))}, {"global_f64", wasm.F64, math.Float64bits(666.6)}} {
		ex, err := s.NewGlobal(module.GlobalType{Value: g.typ}, value.Value{Lo: g.bits})
		if err != nil {
			t.Fatal(err)
		}
		r.spectest[g.name] = ex
	}
	ex, err := s.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 10, Max: 20, HasMax: true}})
	if err != nil {
		t.Fatal(err)
	}
	r.spectest["table"] = ex
	ex, err = s.NewMemory(module.Limits{Min: 1, Max: 2, HasMax: true})
	if err != nil {
		t.Fatal(err)
	}
	r.spectest["memory"] = ex
	return r
}
func compile(data []byte) (*slots.Module, error) {
	return compileWithFeatures(data, 0)
}
func compileWithFeatures(data []byte, features wasm.Features) (*slots.Module, error) {
	m, err := decode.DecodeWithFeatures(data, decode.DefaultLimits(), features)
	if err != nil {
		return nil, err
	}
	validated, err := validate.CheckWithFeatures(m, validate.DefaultLimits(), features)
	if err != nil {
		return nil, err
	}
	ir, err := lower.Compile(validated, lower.DefaultLimits())
	if err != nil {
		return nil, err
	}
	return slots.Compile(ir, slots.DefaultLimits())
}
func (r *specRunner) resolve(m, n string) (state.Extern, bool) {
	if m == "spectest" && n == "table64" && r.features&wasm.Memory64 != 0 {
		if ex, ok := r.spectest[n]; ok {
			return ex, true
		}
		ex, err := r.store.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 10, Max: 20, HasMax: true, Index64: true}})
		if err != nil {
			r.t.Fatal(err)
		}
		r.spectest[n] = ex
		return ex, true
	}
	if in := r.registered[m]; in != nil {
		ex, ok := in.Exports[n]
		return ex, ok
	}
	if m == "spectest" {
		ex, ok := r.spectest[n]
		return ex, ok
	}
	return state.Extern{}, false
}
func (r *specRunner) allocate(data []byte) (*state.Instance, error) {
	compiled, err := compileWithFeatures(data, r.features)
	if err != nil {
		return nil, err
	}
	in, err := r.store.Allocate(compiled, r.resolve)
	if err != nil {
		return nil, err
	}
	if r.tier != 0 {
		code, err := nativeexec.Compile(r.context, compiled, codegen.DefaultLimits())
		if err != nil {
			return nil, err
		}
		lease, err := code.Acquire()
		code.Close()
		if err != nil {
			return nil, err
		}
		r.modules[compiled] = lease
		r.t.Cleanup(lease.Close)
	}
	return in, nil
}
func (r *specRunner) instantiate(data []byte) (*state.Instance, error) {
	in, err := r.allocate(data)
	if err != nil {
		return nil, err
	}
	if err := in.Initialize(); err != nil {
		r.store.CollectFailed(nil, nil)
		return nil, err
	}
	def := in.Compiled.IR.Validated.Definition()
	if def.HasStart {
		if _, err := r.call(in.Functions[def.Start], nil); err != nil {
			r.store.CollectFailed(nil, nil)
			return nil, err
		}
	}
	return in, r.store.Publish(in)
}
func (r *specRunner) call(ref uint64, args []value.Value) ([]value.Value, error) {
	config := Config{Limits: DefaultLimits(), Metered: r.tier == 2, Fuel: ^uint64(0)}
	if r.tier != 0 {
		backend, err := nativeexec.New(r.context, r.modules, config.Limits.MaxSlots, config.Limits.MaxFrames, config.Metered)
		if err != nil {
			return nil, err
		}
		config.Backend = backend
	}
	e := New(r.store, r.context, config)
	defer e.Close()
	if err := e.Start(ref, args); err != nil {
		return nil, err
	}
	for {
		switch e.Run() {
		case Done:
			return e.Results(), nil
		case Failed:
			return nil, e.Err()
		case HostRequest:
			if err := e.ResumeHost(nil, nil); err != nil {
				return nil, err
			}
		default:
			panic("unexpected execution status")
		}
	}
}
func (r *specRunner) instance(name string) *state.Instance {
	if name != "" {
		return r.named[name]
	}
	return r.last
}
func (r *specRunner) action(t *testing.T, a specAction) ([]value.Value, error) {
	t.Helper()
	in := r.instance(a.Module)
	if in == nil {
		t.Fatal("missing action module", a.Module)
	}
	ex, ok := in.Exports[a.Field]
	if !ok {
		t.Fatal("missing export", a.Field)
	}
	switch a.Type {
	case "get":
		if ex.Kind != 3 {
			t.Fatal("get non-global")
		}
		return []value.Value{ex.Global.Value}, nil
	case "invoke":
		if ex.Kind != 0 {
			t.Fatal("invoke non-function")
		}
		args := make([]value.Value, len(a.Args))
		for i, v := range a.Args {
			args[i] = parseSpec(t, v)
		}
		return r.call(ex.Function, args)
	default:
		t.Fatal("unknown action", a.Type)
		return nil, nil
	}
}
func TestCompleteCore2Execution(t *testing.T) { completeCore2(t, 0) }
func TestCompleteCore2Native(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for tier := 1; tier <= 2; tier++ {
		t.Run(strconv.Itoa(tier), func(t *testing.T) { completeCore2(t, tier) })
	}
}
func completeCore2(t *testing.T, tier int) {
	reportPath := os.Getenv("WASMVM_CONFORMANCE_REPORT")
	if tier != 0 && reportPath != "" {
		reportPath += fmt.Sprintf(".native-%d", tier)
	}
	if reportPath != "" {
		if err := os.Remove(reportPath); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	scripts, assertions, textOnly, commands := 0, 0, 0, 0
	totals := make(map[string]int)
	err := filepath.WalkDir("../../../artifacts/core2/fixtures", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "script.json" {
			return nil
		}
		scripts++
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var script struct{ Commands []specCommand }
		if err := json.Unmarshal(raw, &script); err != nil {
			return err
		}
		name := filepath.Base(filepath.Dir(path))
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			r := newSpec(t, ctx, tier)
			for _, cmd := range script.Commands {
				commands++
				totals[cmd.Type]++
				if cmd.Type == "assert_malformed" && cmd.ModuleType == "text" {
					textOnly++
					continue
				}
				if strings.HasPrefix(cmd.Type, "assert_") {
					assertions++
				}
				r.runCommand(t, cmd, path)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal("prepare pinned corpus:", err)
	}
	if scripts != 148 || assertions != 51139 || textOnly != 1091 || commands != 54006 {
		t.Fatalf("inventory: scripts=%d assertions=%d text=%d commands=%d", scripts, assertions, textOnly, commands)
	}
	t.Logf("Executed all %d binary assertions in %d scripts; %d text-parser assertions are classified separately. Commands: %v", assertions, scripts, textOnly, totals)
	if reportPath != "" && !t.Failed() {
		pins, err := os.ReadFile("../../../testdata/core2/pins.json")
		if err != nil {
			t.Fatal(err)
		}
		backend := "wasmvm-interpreter"
		if tier != 0 {
			backend = fmt.Sprintf("wasmvm-native-tier-%d", tier)
		}
		report := struct {
			SchemaVersion                                                        int  `json:"schemaVersion"`
			Passed                                                               bool `json:"passed"`
			Profile, Backend, Compiler, RuntimeVersion, GOOS, GOARCH, PinsSHA256 string
			Scripts, BinaryAssertions, TextParserAssertions, Commands            int
			CommandCounts                                                        map[string]int
		}{1, true, "Core 2.0 binary", backend, runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH, fmt.Sprintf("%x", sha256.Sum256(pins)), scripts, assertions, textOnly, commands, totals}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(reportPath, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
func trapMatches(code wasm.TrapCode, text string) bool {
	if code.String() == text {
		return true
	}
	return code == wasm.TrapCastFailure && text == "cast" || code == wasm.TrapNullReference && (text == "null exception reference" || text == "null function reference" || text == "null i31 reference" || text == "null array reference" || text == "null structure reference") || code == wasm.TrapIndirectCallType && text == "indirect call" || code == wasm.TrapTableBounds && text == "undefined element" || code == wasm.TrapUninitializedElement && strings.HasPrefix(text, "uninitialized element")
}
func specLanes(t *testing.T, v specValue) (uint, []string) {
	t.Helper()
	name := v.Type
	if name == "v128" {
		name = v.LaneType
		var xs []string
		if err := json.Unmarshal(v.Value, &xs); err != nil {
			t.Fatal(err)
		}
		w := specWidth(t, name)
		if len(xs) != int(128/w) {
			t.Fatal("lane count")
		}
		return w, xs
	}
	var x string
	if err := json.Unmarshal(v.Value, &x); err != nil {
		t.Fatal(err)
	}
	return specWidth(t, name), []string{x}
}
func specWidth(t *testing.T, s string) uint {
	t.Helper()
	switch s {
	case "i8":
		return 8
	case "i16":
		return 16
	case "i32", "f32":
		return 32
	case "i64", "f64":
		return 64
	default:
		t.Fatal("non-numeric type", s)
		return 0
	}
}
func specUint(t *testing.T, s string) uint64 {
	t.Helper()
	if strings.HasPrefix(s, "-") {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return uint64(v)
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func parseSpec(t *testing.T, v specValue) value.Value {
	t.Helper()
	if v.Type == "externref" || v.Type == "funcref" || v.Type == "exnref" || v.Type == "anyref" {
		var s string
		if err := json.Unmarshal(v.Value, &s); err != nil {
			t.Fatal(err)
		}
		if s == "null" {
			return value.Value{}
		}
		if v.Type != "externref" && v.Type != "anyref" {
			t.Fatal("unexpected concrete funcref")
		}
		if v.Type == "anyref" {
			return value.Value{Lo: state.Internalize(specUint(t, s) + 1)}
		}
		return value.Value{Lo: specUint(t, s) + 1}
	}
	w, xs := specLanes(t, v)
	var raw [16]byte
	for i, s := range xs {
		x := specUint(t, s)
		for j := uint(0); j < w/8; j++ {
			raw[i*int(w/8)+int(j)] = byte(x >> (j * 8))
		}
	}
	return value.Value{Lo: binary.LittleEndian.Uint64(raw[:8]), Hi: binary.LittleEndian.Uint64(raw[8:])}
}
func (r *specRunner) match(t *testing.T, action specAction, want specValue, got value.Value) bool {
	if want.Type == "either" {
		for _, v := range want.Values {
			if r.match(t, action, v, got) {
				return true
			}
		}
		return false
	}
	refs := map[string]wasm.ValueType{"i31ref": wasm.I31Ref, "structref": wasm.StructRef, "arrayref": wasm.ArrayRef, "eqref": wasm.EqRef, "externref": wasm.ExternRef, "funcref": wasm.FuncRef, "exnref": wasm.ExnRef}
	if typ, ok := refs[want.Type]; ok && len(want.Value) == 0 {
		return r.store.TestReference(got.Lo, typ.NonNull())
	}
	if want.Type == "refnull" || want.Type == "nullref" || want.Type == "nullfuncref" || want.Type == "nullexternref" || want.Type == "nullexnref" {
		return got.Lo == 0 && got.Hi == 0
	}
	if want.Type == "anyref" {
		return got == parseSpec(t, want)
	}

	if want.Type == "funcref" && len(want.Value) != 0 {
		var text string
		if err := json.Unmarshal(want.Value, &text); err != nil {
			t.Fatal(err)
		}
		if text != "null" {
			index := specUint(t, text)
			in := r.instance(action.Module)
			return in != nil && index < uint64(len(in.Functions)) && got.Lo == in.Functions[index] && got.Hi == 0
		}
	}
	return matchSpec(t, want, got)
}

func matchSpec(t *testing.T, want specValue, got value.Value) bool {
	t.Helper()
	if want.Type == "externref" || want.Type == "funcref" || want.Type == "exnref" {
		if len(want.Value) == 0 {
			return got.Lo != 0
		}
		return got == parseSpec(t, want)
	}
	w, xs := specLanes(t, want)
	var raw [16]byte
	binary.LittleEndian.PutUint64(raw[:8], got.Lo)
	binary.LittleEndian.PutUint64(raw[8:], got.Hi)
	for i, s := range xs {
		var x uint64
		for j := uint(0); j < w/8; j++ {
			x |= uint64(raw[i*int(w/8)+int(j)]) << (j * 8)
		}
		if s == "nan:canonical" || s == "nan:arithmetic" {
			mask, canonical := uint64(0x7fffffff), uint64(0x7fc00000)
			if w == 64 {
				mask, canonical = 0x7fffffffffffffff, 0x7ff8000000000000
			}
			x &= mask
			if s == "nan:canonical" {
				if x != canonical {
					return false
				}
			} else if x&canonical != canonical {
				return false
			}
		} else if x != (specUint(t, s) & (^uint64(0) >> (64 - w))) {
			return false
		}
	}
	return true
}

func (r *specRunner) runCommand(t *testing.T, cmd specCommand, path string) {
	t.Helper()
	name := filepath.Base(filepath.Dir(path))
	var input []byte
	var err error
	if cmd.Filename != "" {
		input, err = os.ReadFile(filepath.Join(filepath.Dir(path), cmd.Filename))
		if err != nil {
			t.Fatal(err)
		}
	}
	fail := func(format string, args ...any) {
		t.Helper()
		t.Fatalf("%s:%d %s: %s", name, cmd.Line, cmd.Type, fmt.Sprintf(format, args...))
	}
	switch cmd.Type {
	case "module_instance":
		data, ok := r.definitions[cmd.Module]
		if !ok {
			fail("unknown module definition %s", cmd.Module)
		}
		in, err := r.instantiate(data)
		if err != nil {
			fail("%v", err)
		}
		r.last = in
		if cmd.Instance != "" {
			r.named[cmd.Instance] = in
		}
	case "module_definition":
		r.definitions[cmd.Name] = input
		// Official definition commands validate/compile without allocating an
		// instance; the declared minimum can intentionally exceed host quotas.
		compiled, err := compileWithFeatures(input, r.features)
		if err != nil {
			fail("%v", err)
		}
		if r.tier != 0 {
			code, err := nativeexec.Compile(r.context, compiled, codegen.DefaultLimits())
			if err != nil {
				fail("%v", err)
			}
			code.Close()
		}
	case "module":
		in, err := r.instantiate(input)
		if err != nil {
			fail("%v", err)
		}
		r.last = in
		if cmd.Name != "" {
			r.named[cmd.Name] = in
		}
	case "register":
		in := r.instance(cmd.Name)
		if in == nil {
			fail("unknown registration")
		}
		r.registered[cmd.As] = in
	case "action":
		if _, err := r.action(t, cmd.Action); err != nil {
			fail("%v", err)
		}
	case "assert_malformed", "assert_invalid", "assert_unlinkable", "assert_uninstantiable":
		_, err := r.instantiate(input)
		want := map[string]error{"assert_malformed": wasm.ErrMalformed, "assert_invalid": wasm.ErrInvalid, "assert_unlinkable": state.ErrLink, "assert_uninstantiable": wasm.ErrTrap}[cmd.Type]
		if !errors.Is(err, want) {
			fail("got %v, want %v (%s)", err, want, cmd.Text)
		}
	case "assert_return":
		got, err := r.action(t, cmd.Action)
		if err != nil {
			fail("%s: %v", cmd.Action.Field, err)
		}
		if len(cmd.Either) != 0 {
			if len(got) == 1 {
				for _, want := range cmd.Either {
					if r.match(t, cmd.Action, want, got[0]) {
						return
					}
				}
			}
			fail("%s results %#v match no permitted alternative", cmd.Action.Field, got)
		}
		if len(got) != len(cmd.Expected) {
			fail("results %v want %v", got, cmd.Expected)
		}
		for i, want := range cmd.Expected {
			if !r.match(t, cmd.Action, want, got[i]) {
				fail("%s result %d: got %#v want %s (%s)", cmd.Action.Field, i, got[i], want.Value, want.Type)
			}
		}
	case "assert_exception":
		_, err := r.action(t, cmd.Action)
		if !errors.Is(err, wasm.ErrException) {
			fail("got %v want exception", err)
		}
	case "assert_trap":
		_, err := r.action(t, cmd.Action)
		var trap *wasm.Trap
		if !errors.As(err, &trap) || !trapMatches(trap.Code, cmd.Text) {
			fail("%s got %v want trap %s", cmd.Action.Field, err, cmd.Text)
		}
	case "assert_exhaustion":
		_, err := r.action(t, cmd.Action)
		if !errors.Is(err, wasm.ErrLimit) || (!strings.Contains(err.Error(), "guest call stack exhausted") && !strings.Contains(err.Error(), "guest frame slots")) {
			fail("got %v want call-stack exhaustion", err)
		}
	default:
		fail("unclassified command")
	}
}

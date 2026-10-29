package cli

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func parseValue(input string) (vm.Value, error) {
	kind, raw, ok := strings.Cut(input, ":")
	if !ok {
		return vm.Value{}, fmt.Errorf("argument %q needs type:value", input)
	}
	integer := func(width int) (uint64, error) {
		if strings.HasPrefix(raw, "-") {
			n, e := strconv.ParseInt(raw, 0, width)
			return uint64(n), e
		}
		return strconv.ParseUint(raw, 0, width)
	}
	switch kind {
	case "i32":
		n, e := integer(32)
		return vm.I32(uint32(n)), e
	case "i64":
		n, e := integer(64)
		return vm.I64(n), e
	case "f32":
		n, e := strconv.ParseFloat(raw, 32)
		return vm.F32(float32(n)), e
	case "f64":
		n, e := strconv.ParseFloat(raw, 64)
		return vm.F64(n), e
	case "f32bits":
		n, e := strconv.ParseUint(raw, 0, 32)
		return vm.F32Bits(uint32(n)), e
	case "f64bits":
		n, e := strconv.ParseUint(raw, 0, 64)
		return vm.F64Bits(n), e
	case "v128":
		parts := strings.Split(raw, ",")
		if len(parts) != 2 {
			return vm.Value{}, fmt.Errorf("v128 requires low64,high64")
		}
		lo, e := strconv.ParseUint(parts[0], 0, 64)
		if e != nil {
			return vm.Value{}, e
		}
		hi, e := strconv.ParseUint(parts[1], 0, 64)
		return vm.V128(lo, hi), e
	case "funcref", "externref", "exnref":
		if raw != "null" {
			return vm.Value{}, fmt.Errorf("CLI references must be null")
		}
		t := wasm.FuncRef
		if kind == "exnref" {
			t = wasm.ExnRef
		}
		if kind == "externref" {
			t = wasm.ExternRef
		}
		return vm.NullRef(t)
	default:
		return vm.Value{}, fmt.Errorf("unknown argument type %q", kind)
	}
}

type resultValue struct {
	Type  string `json:"type"`
	Bits  string `json:"bits,omitempty"`
	Value string `json:"value,omitempty"`
}

func formatValue(v vm.Value) resultValue {
	r := resultValue{Type: v.Type().String()}
	if v.Type().IsReference() {
		r.Value = "non-null"
		if v.IsNull() {
			r.Value = "null"
		}
		return r
	}
	lo, hi, _ := v.Bits()
	switch v.Type() {
	case wasm.I32:
		r.Bits = fmt.Sprintf("0x%08x", uint32(lo))
		r.Value = strconv.FormatInt(int64(int32(lo)), 10)
	case wasm.I64:
		r.Bits = fmt.Sprintf("0x%016x", lo)
		r.Value = strconv.FormatInt(int64(lo), 10)
	case wasm.F32:
		r.Bits = fmt.Sprintf("0x%08x", uint32(lo))
		r.Value = strconv.FormatFloat(float64(math.Float32frombits(uint32(lo))), 'g', -1, 32)
	case wasm.F64:
		r.Bits = fmt.Sprintf("0x%016x", lo)
		r.Value = strconv.FormatFloat(math.Float64frombits(lo), 'g', -1, 64)
	case wasm.V128:
		r.Bits = fmt.Sprintf("0x%016x,0x%016x", lo, hi)
	}
	return r
}

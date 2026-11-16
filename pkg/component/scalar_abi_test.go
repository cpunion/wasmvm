package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarCanonicalBitsAndFailedCarrierChecks(t *testing.T) {
	codec := func() *canonical.Codec {
		return &canonical.Codec{Layouts: canonical.NewBuilder(false, canonical.DefaultLimits()), Transfer: canonical.Transfer{MaxBytes: 4096}, MaxValues: 1024}
	}
	f32, f64 := &model.Type{Kind: model.F32}, &model.Type{Kind: model.F64}
	plan := &validate.CanonicalFunction{Type: &validate.Type{Parameters: []model.Field{{Type: f32}, {Type: f64}}, Result: f64}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.F32, wasm.F64}, Results: []wasm.ValueType{wasm.F64}}}
	words := []uint64{0x80000000, 0x8000000000000000}
	args, err := newCanonicalCallPlan(plan).liftScalarArguments(codec(), words, nil)
	if err != nil || len(args) != 2 || args[0].Bits != words[0] || args[1].Bits != words[1] {
		t.Fatal(args, err)
	}
	if err := newCanonicalCallPlan(plan).lowerScalarResult(codec(), words, &args[1], nil); err != nil || words[0] != 0x8000000000000000 {
		t.Fatal(words, err)
	}
	// Canonical NaN normalization still applies at this boundary.
	words = []uint64{0x80000000, 0xfff8000000001234}
	nanArgs, err := newCanonicalCallPlan(plan).liftScalarArguments(codec(), words, nil)
	if err != nil || nanArgs[1].Bits != 0x7ff8000000000000 {
		t.Fatal(nanArgs, err)
	}
	if err := newCanonicalCallPlan(plan).lowerScalarResult(codec(), words, &nanArgs[1], nil); err != nil || words[0] != 0x7ff8000000000000 {
		t.Fatal(words, err)
	}
	if err := newCanonicalCallPlan(plan).lowerScalarResult(codec(), words, nil, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("missing numeric result accepted", err)
	}
	plan.Signature.Results[0] = wasm.I32
	if err := newCanonicalCallPlan(plan).lowerScalarResult(codec(), words, &args[1], nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("incompatible result carrier accepted", err)
	}
	oversized := &validate.CanonicalFunction{Signature: vm.FunctionType{Params: make([]wasm.ValueType, canonical.MaxFlatParams+2)}}
	if _, err := newCanonicalCallPlan(oversized).liftScalarArguments(codec(), nil, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("oversized flat signature accepted", err)
	}
	plan = &validate.CanonicalFunction{Type: &validate.Type{Parameters: []model.Field{{Type: &model.Type{Kind: model.Char}}}}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.I32}}}
	if _, err := newCanonicalCallPlan(plan).liftScalarArguments(codec(), []uint64{0xd800}, nil); err == nil {
		t.Fatal("invalid character accepted")
	}
	limited := codec()
	limited.MaxValues = 0
	if _, err := newCanonicalCallPlan(plan).liftScalarArguments(limited, []uint64{1}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	canceled := codec()
	canceled.Transfer.Context = ctx
	if _, err := newCanonicalCallPlan(plan).liftScalarArguments(canceled, []uint64{1}, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

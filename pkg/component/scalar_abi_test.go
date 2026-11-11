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
		return &canonical.Codec{Layouts: canonical.NewBuilder(false, canonical.DefaultLimits()), Transfer: &canonical.Transfer{MaxBytes: 4096}, MaxValues: 1024}
	}
	f32, f64 := &model.Type{Kind: model.F32}, &model.Type{Kind: model.F64}
	plan := &validate.CanonicalFunction{Type: &validate.Type{Parameters: []model.Field{{Type: f32}, {Type: f64}}, Result: f64}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.F32, wasm.F64}, Results: []wasm.ValueType{wasm.F64}}}
	words := []uint64{0x80000000, 0x8000000000000000}
	args, err := liftScalarArguments(codec(), plan, words)
	if err != nil || len(args) != 2 || args[0].Bits != words[0] || args[1].Bits != words[1] {
		t.Fatal(args, err)
	}
	if err := lowerScalarResult(codec(), plan, words, &args[1]); err != nil || words[0] != 0x8000000000000000 {
		t.Fatal(words, err)
	}
	// Canonical NaN normalization still applies at this boundary.
	words = []uint64{0x80000000, 0xfff8000000001234}
	nanArgs, err := liftScalarArguments(codec(), plan, words)
	if err != nil || nanArgs[1].Bits != 0x7ff8000000000000 {
		t.Fatal(nanArgs, err)
	}
	if err := lowerScalarResult(codec(), plan, words, &nanArgs[1]); err != nil || words[0] != 0x7ff8000000000000 {
		t.Fatal(words, err)
	}
	if err := lowerScalarResult(codec(), plan, words, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("missing numeric result accepted", err)
	}
	plan.Signature.Results[0] = wasm.I32
	if err := lowerScalarResult(codec(), plan, words, &args[1]); !errors.Is(err, vm.ErrType) {
		t.Fatal("incompatible result carrier accepted", err)
	}
	oversized := &validate.CanonicalFunction{Signature: vm.FunctionType{Params: make([]wasm.ValueType, canonical.MaxFlatParams+2)}}
	if _, err := liftScalarArguments(codec(), oversized, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("oversized flat signature accepted", err)
	}
	plan = &validate.CanonicalFunction{Type: &validate.Type{Parameters: []model.Field{{Type: &model.Type{Kind: model.Char}}}}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.I32}}}
	if _, err := liftScalarArguments(codec(), plan, []uint64{0xd800}); err == nil {
		t.Fatal("invalid character accepted")
	}
	limited := codec()
	limited.MaxValues = 0
	if _, err := liftScalarArguments(limited, plan, []uint64{1}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	canceled := codec()
	canceled.Transfer.Context = ctx
	if _, err := liftScalarArguments(canceled, plan, []uint64{1}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

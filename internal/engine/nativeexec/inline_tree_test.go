package nativeexec

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNestedExpansionInvocationQuotasAndRetargeting(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, indirect := range []bool{false, true} {
		d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}}}
		for index := uint32(0); index < 3; index++ {
			body := module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}}
			if index == 0 && indirect {
				body = append(body, module.Instr{Op: 0x41}, module.Instr{Op: 0x11})
			} else {
				body = append(body, module.Instr{Op: 0x10, A: index + 1})
			}
			d.Functions = append(d.Functions, module.Function{Body: append(body, module.Instr{Op: 0xb})})
		}
		d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x6a}, {Op: 0xb}}})
		if indirect {
			d.Tables = []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}
		}
		m := callSlots(t, d)
		for _, metered := range []bool{false, true} {
			s := state.NewNative(state.DefaultLimits())
			t.Cleanup(s.Close)
			in, err := s.Allocate(m, nil)
			if err != nil {
				t.Fatal(err)
			}
			e, _ := callExecution(t, s, m, in, metered, execute.Limits{MaxFrames: 128, MaxSlots: 2048})
			targets := []int{1}
			if indirect {
				targets = []int{1, 3, 2, 1}
			}
			for _, target := range targets {
				if indirect {
					in.Tables[0].Values[0] = in.Functions[target]
				}
				requiredSlots := len(m.Functions[0].Slots)
				for _, f := range m.Functions[target:] {
					requiredSlots += len(f.Slots)
				}
				requiredFrames := uint32(1 + len(m.Functions) - target)
				for _, frames := range []uint32{requiredFrames - 1, requiredFrames, 128} {
					for _, slots := range []uint32{uint32(requiredSlots) - 1, uint32(requiredSlots), 2048} {
						limits := execute.Limits{MaxFrames: frames, MaxSlots: slots}
						if err := e.SetLimits(limits); err != nil {
							t.Fatal(err)
						}
						oracle := execute.New(s, nil, execute.Config{Limits: limits, Metered: metered, Fuel: 1000})
						args := []value.Value{{Lo: 0xfffffff1}, {Lo: 37}}
						if err := oracle.Start(in.Functions[0], args); err != nil {
							t.Fatal(err)
						}
						wantStatus := oracle.Run()
						if err := e.Start(in.Functions[0], args); err != nil {
							t.Fatal(err)
						}
						gotStatus := e.Run()
						if gotStatus != wantStatus || e.Fuel() != oracle.Fuel() ||
							gotStatus == execute.Done && !reflect.DeepEqual(e.Results(), oracle.Results()) ||
							errors.Is(e.Err(), wasm.ErrLimit) != errors.Is(oracle.Err(), wasm.ErrLimit) {
							t.Fatal("nested quota or fuel differs from interpreter", indirect, metered, target, limits, gotStatus, wantStatus, e.Err(), oracle.Err(), e.Fuel(), oracle.Fuel())
						}
						wantLimit := frames < requiredFrames || slots < uint32(requiredSlots)
						if errors.Is(e.Err(), wasm.ErrLimit) != wantLimit || !wantLimit && (gotStatus != execute.Done || e.Results()[0].Lo != 22) {
							t.Fatal("original nested stack was not enforced", indirect, metered, target, limits, e.Err(), e.Results())
						}
						oracle.Close()
					}
				}
			}
		}
	}
}

package execute

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTailHostRetiresFrameButKeepsCallerAndArguments(t *testing.T) {
	input, err := os.ReadFile("testdata/tail-host.wasm")
	if err != nil {
		t.Fatal(err)
	}
	for tier := 0; tier <= 2; tier++ {
		if tier != 0 && !native.Available {
			continue
		}
		t.Run(strconv.Itoa(tier), func(t *testing.T) {
			r := newSpec(t, context.Background(), tier)
			r.features = wasm.TailCall
			host, err := r.store.NewHost(module.FuncType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}, "tail-echo")
			if err != nil {
				t.Fatal(err)
			}
			r.spectest["tail_echo"] = host
			in, err := r.allocate(input)
			if err != nil {
				t.Fatal(err)
			}
			limits := Limits{MaxFrames: 1, MaxSlots: 1}
			cfg := Config{Limits: limits, Metered: tier == 2, Fuel: 2}
			if tier != 0 {
				cfg.Backend, err = nativeexec.New(context.Background(), r.modules, 1, 1, tier == 2)
				if err != nil {
					t.Fatal(err)
				}
			}
			e := New(r.store, nil, cfg)
			defer e.Close()
			if err := e.Start(in.Exports["too-wide"].Function, nil); err != nil {
				t.Fatal(err)
			}
			if e.Run() != Failed || !errors.Is(e.Err(), wasm.ErrLimit) || e.Footprint() != (Limits{}) {
				t.Fatal(e.Err(), e.Footprint())
			}
			if err := e.Start(in.Exports["run"].Function, []value.Value{{Lo: 99}}); err != nil {
				t.Fatal(err)
			}
			if e.Run() != HostRequest || e.Footprint() != (Limits{}) {
				t.Fatal(e.Err(), e.Footprint())
			}
			if tier == 2 && e.Fuel() != 0 {
				t.Fatal("tail instruction charged incorrectly", e.Fuel())
			}
			roots := 0
			e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
				if typ == wasm.ExternRef && v.Lo == 99 {
					roots++
				}
			})
			if roots != 1 {
				t.Fatal("pending argument lost or retired locals retained", roots)
			}
			var active []*state.Instance
			e.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
			r.store.CollectFailed(nil, active)
			if len(active) != 1 || r.store.Usage().Instances != 1 {
				t.Fatal("lost unpublished host caller", len(active), r.store.Usage())
			}
			if err := e.ResumeHost([]value.Value{{Lo: 99}}, nil); err != nil || e.Run() != Done || e.Results()[0].Lo != 99 {
				t.Fatal(e.Err(), err)
			}
			active = nil
			e.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
			r.store.CollectFailed(nil, active)
			if len(active) != 0 || r.store.Usage().Instances != 0 {
				t.Fatal("completed caller retained", r.store.Usage())
			}
		})
	}
}

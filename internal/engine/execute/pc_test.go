package execute

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPCPublicationAcrossNestedHostSuspensions(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	host, err := s.NewHost(sig, "advance")
	if err != nil {
		t.Fatal(err)
	}
	m := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x10}, {Op: 0xb}}},
		{Body: module.Expr{{Op: 0x20}, {Op: 0x10, A: 1}, {Op: 0x10, A: 1}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}}},
	}}
	in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
	for _, metered := range []bool{false, true} {
		e := New(s, nil, Config{Limits: DefaultLimits(), Metered: metered, Fuel: 100})
		for _, arg := range []uint32{7, 19, 0xfffffffc} {
			if err := e.Start(in.Functions[2], []value.Value{{Lo: uint64(arg)}}); err != nil {
				t.Fatal(err)
			}
			for _, want := range []uint32{arg, arg + 101, arg + 201, arg + 302} {
				if e.Run() != HostRequest || e.Host().Args[0].Lo != uint64(want) {
					t.Fatal("lost caller continuation", metered, arg, want, e.Host(), e.Err())
				}
				for _, f := range e.frames {
					if f.pc == 0 || f.function.Code.Code[f.pc-1].Op != 0x10 {
						t.Fatal("suspended frame did not publish its next PC", f.pc)
					}
				}
				if err := e.ResumeHost([]value.Value{{Lo: uint64(want + 100)}}, nil); err != nil {
					t.Fatal(err)
				}
			}
			if e.Run() != Done || e.Results()[0].Lo != uint64(arg+403) {
				t.Fatal("nested continuation result", metered, arg, e.Err(), e.Results())
			}
		}
	}
}

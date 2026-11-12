package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type resumeCountingBackend struct {
	*Execution
	runs, resumes, parents int
}

func (b *resumeCountingBackend) RunFrames(out *state.NativeFrames) error {
	b.runs++
	if out.Resume != 0 {
		b.resumes++
	}
	err := b.Execution.RunFrames(out)
	b.parents += int(out.Parents)
	return err
}

func TestNativeHostReturnResumesWholeChainAndGrowsMetadata(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	d := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}}
	for i := 0; i < 17; i++ {
		target := uint32(i + 2)
		if i == 16 {
			target = 0
		}
		d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x20}, {Op: 0x10, A: target}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}}})
	}
	m := callSlots(t, d)
	compiled, err := Compile(context.Background(), m, codegen.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	lease, err := compiled.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	for _, metered := range []bool{false, true} {
		s := state.NewNative(state.DefaultLimits())
		defer s.Close()
		host, err := s.NewHost(sig, "double")
		if err != nil {
			t.Fatal(err)
		}
		in, err := s.Allocate(m, func(string, string) (state.Extern, bool) { return host, true })
		if err != nil {
			t.Fatal(err)
		}
		n, err := New(context.Background(), map[*slots.Module]*Lease{m: lease}, 2048, 128, metered)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = n.Resize(0, 2048); err != nil {
			t.Fatal(err)
		}
		b := &resumeCountingBackend{Execution: n}
		e := execute.New(s, nil, execute.Config{Limits: execute.Limits{MaxFrames: 128, MaxSlots: 2048}, Backend: b, Metered: metered, Fuel: 1000})
		defer e.Close()
		for round := 0; round < 2; round++ {
			before := b.runs
			if err := e.Start(in.Functions[1], []value.Value{{Lo: 10}}); err != nil || e.Run() != execute.HostRequest {
				t.Fatal(err, e.Err())
			}
			if e.Footprint().MaxFrames != 17 || e.Host().Args[0].Lo != 10 || n.callWords[0] != 16 {
				t.Fatal("missing parked chain", e.Footprint(), n.callWords[:4])
			}
			if err := e.ResumeHost([]value.Value{{Lo: 20}}, nil); err != nil || e.Run() != execute.Done || e.Results()[0].Lo != 37 {
				t.Fatal(err, e.Err(), e.Results())
			}
			if b.runs-before != 2 || b.resumes != round+1 || b.parents != 16*(round+1) || n.callWords[1] < 32 || n.resumeToken != 0 {
				t.Fatal("native return chain was not resumed", b.runs, b.resumes, b.parents, n.callWords[:4])
			}
			want := uint64(1000)
			if metered {
				want -= 4 * 17
			}
			if e.Fuel() != want {
				t.Fatal("changed precise fuel", metered, e.Fuel(), want)
			}
		}
	}
}

func TestResumeTokenChecksAndRetirement(t *testing.T) {
	fn := &state.Function{}
	other := &state.Function{}
	n := &Execution{resumeToken: 7, resumeDepth: 2, callWords: []uint64{1}, callFrames: []state.NativeFrame{{}, {Function: fn, Base: 3, PC: 9}}}
	valid := state.NativeFrames{Resume: 7, Function: fn, Base: 3, Depth: 2, PC: 10}
	if !n.resumeCalls(&valid) {
		t.Fatal("matching continuation declined")
	}
	for _, change := range []func(*state.NativeFrames){
		func(p *state.NativeFrames) { p.Resume = 0 }, func(p *state.NativeFrames) { p.Resume++ },
		func(p *state.NativeFrames) { p.Function = other }, func(p *state.NativeFrames) { p.Base++ },
		func(p *state.NativeFrames) { p.Depth++ }, func(p *state.NativeFrames) { p.PC++ },
	} {
		p := valid
		change(&p)
		if n.resumeCalls(&p) {
			t.Fatal("mismatched continuation accepted", p)
		}
	}
	n.callWords[0]++
	if n.resumeCalls(&valid) {
		t.Fatal("mismatched native depth accepted")
	}
	n.resetCalls()
	if n.resumeCalls(&valid) || n.resumeToken != 0 || n.resumeDepth != 0 || n.callFrames != nil && len(n.callFrames) != 0 {
		t.Fatal("retired continuation accepted")
	}
	_, _, err := n.Run(nil, 0, 1, 0, 0)
	if !errors.Is(err, native.ErrClosed) {
		t.Fatal(err)
	}
}

func TestParentNativeChainSurvivesForeignVeneerAndRejectsStaleChildren(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	for _, foreignNative := range []bool{false, true} {
		d := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}, Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x41}, {Op: 0x11}, {Op: 0xb}}}}}
		if foreignNative {
			d.Functions[0].Body = module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}}
		}
		foreign := callSlots(t, d)
		d = &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}}
		for _, target := range []uint32{2, 3, 0} {
			d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x20}, {Op: 0x10, A: target}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}}})
		}
		parent := callSlots(t, d)
		leases := map[*slots.Module]*Lease{}
		for _, m := range []*slots.Module{parent, foreign} {
			compiled, err := Compile(context.Background(), m, codegen.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(compiled.Close)
			lease, err := compiled.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			leases[m] = lease
		}
		for _, metered := range []bool{false, true} {
			s := state.NewNative(state.DefaultLimits())
			t.Cleanup(s.Close)
			host, err := s.NewHost(sig, "echo")
			if err != nil {
				t.Fatal(err)
			}
			other, err := s.Allocate(foreign, func(string, string) (state.Extern, bool) { return host, true })
			if err != nil {
				t.Fatal(err)
			}
			other.Tables[0].Values[0] = host.Function
			in, err := s.Allocate(parent, func(string, string) (state.Extern, bool) { return state.Extern{Owner: s, Function: other.Functions[1]}, true })
			if err != nil {
				t.Fatal(err)
			}
			n, err := New(context.Background(), leases, 2048, 128, metered)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := n.Resize(0, 2048); err != nil {
				t.Fatal(err)
			}
			b := &resumeCountingBackend{Execution: n}
			e := execute.New(s, nil, execute.Config{Limits: execute.Limits{MaxFrames: 128, MaxSlots: 2048}, Backend: b, Metered: metered, Fuel: 100})
			t.Cleanup(e.Close)
			for round := 0; round < 2; round++ {
				before := b.runs
				if err := e.Start(in.Functions[1], []value.Value{{Lo: 10}}); err != nil || e.Run() != execute.HostRequest || e.Host().Args[0].Lo != 10 {
					t.Fatal(err, e.Err())
				}
				want, fuel := uint64(23), uint64(100)
				if foreignNative {
					want++
				}
				if metered {
					fuel -= 15
					if foreignNative {
						fuel--
					}
				}
				if err := e.ResumeHost([]value.Value{{Lo: 20}}, nil); err != nil || e.Run() != execute.Done || e.Results()[0].Lo != want || e.Fuel() != fuel {
					t.Fatal(foreignNative, metered, err, e.Err(), e.Results(), e.Fuel())
				}
				if !foreignNative && (b.runs-before != 2 || b.parents != (round+1)*2) {
					t.Fatal("parent chain was lost across the foreign veneer", b.runs-before, b.parents)
				}
				if foreignNative && b.runs-before <= 2 {
					t.Fatal("stale chain bypassed the foreign native child")
				}
			}
		}
	}
}

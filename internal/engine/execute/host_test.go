package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func assertHostRetired(t *testing.T, e *Execution) {
	t.Helper()
	if e.Host() != nil || e.hostRequest.Function != nil || e.hostRequest.Caller != nil || e.hostRequest.Args != nil || len(e.hostArgs) != 0 {
		t.Fatal("retired request retained invocation roots")
	}
	for _, v := range e.hostArgs[:cap(e.hostArgs)] {
		if v != (value.Value{}) {
			t.Fatal("retired argument storage was not cleared", v)
		}
	}
}

func TestHostRequestSnapshotsTuplesAndBoundsIdleStorage(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	e := New(s, nil, Config{Limits: DefaultLimits()})
	defer e.Close()
	for _, n := range []int{0, 3, maxCachedHostArgs, 1, maxCachedHostArgs + 1, 0, 1} {
		types := make([]wasm.ValueType, n)
		args, expected := make([]value.Value, n), make([]value.Value, n)
		for i := range types {
			types[i] = []wasm.ValueType{wasm.I64, wasm.V128, wasm.ExternRef}[i%3]
			args[i] = value.Value{Lo: uint64(i + 1)}
			if types[i] == wasm.V128 {
				args[i].Hi = ^uint64(i)
			}
		}
		copy(expected, args)
		fn, err := s.NewHost(module.FuncType{Params: types, Results: types}, "echo")
		if err != nil {
			t.Fatal(err)
		}
		if err := e.Start(fn.Function, args); err != nil || e.Run() != HostRequest {
			t.Fatal(err, e.Err())
		}
		clear(args)
		request := e.Host()
		roots := 0
		e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
			if typ == wasm.ExternRef {
				roots++
				if v.Lo == 0 {
					t.Fatal("lost pending reference argument")
				}
			}
		})
		if roots != n/3 {
			t.Fatal("pending reference roots", roots, n/3)
		}
		for i, v := range request.Args {
			if v != expected[i] {
				t.Fatal("host tuple aliases inputs or lost vector bits", n, i, v)
			}
		}
		// Result handoff may alias the borrowed request tuple. Retirement must
		// follow the complete copy into result/guest-frame storage.
		if err := e.ResumeHost(request.Args, nil); err != nil || e.Run() != Done {
			t.Fatal(err, e.Err())
		}
		assertHostRetired(t, e)
		if n > maxCachedHostArgs && cap(e.hostArgs) != 0 {
			t.Fatal("oversized tuple retained in idle workspace")
		}
		e.Quiesce()
		for i, v := range e.Results() {
			if v != expected[i] {
				t.Fatal("result lost during request retirement", i, v)
			}
		}
	}
}

func TestHostRequestReentryAndFailureRetirement(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	fn, err := s.NewHost(module.FuncType{Params: []wasm.ValueType{wasm.I64}, Results: []wasm.ValueType{wasm.I64}}, "echo")
	if err != nil {
		t.Fatal(err)
	}
	outer, inner := New(s, nil, Config{Limits: DefaultLimits()}), New(s, nil, Config{Limits: DefaultLimits()})
	defer outer.Close()
	defer inner.Close()
	if err := outer.Start(fn.Function, []value.Value{{Lo: 42}}); err != nil {
		t.Fatal(err)
	}
	for _, failure := range []error{nil, wasm.ErrTrap} {
		if err := inner.Start(fn.Function, []value.Value{{Lo: 99}}); err != nil || inner.Run() != HostRequest {
			t.Fatal(err, inner.Err())
		}
		if failure == nil {
			err = inner.ResumeHost(nil, nil) // Invalid result arity.
			if err == nil {
				t.Fatal("accepted wrong host result count")
			}
		} else if err := inner.ResumeHost(nil, failure); !errors.Is(err, failure) {
			t.Fatal(err)
		}
		assertHostRetired(t, inner)
		if outer.Run() != HostRequest || outer.Host().Args[0].Lo != 42 {
			t.Fatal("nested execution overwrote suspended host arguments")
		}
	}
	outer.Abort(wasm.ErrClosed)
	assertHostRetired(t, outer)
	if err := outer.Start(fn.Function, []value.Value{{Lo: 7}}); err != nil {
		t.Fatal(err)
	}
	outer.Quiesce()
	assertHostRetired(t, outer)
	outer.Close()
	if cap(outer.hostArgs) != 0 {
		t.Fatal("closed execution retained argument storage")
	}
}

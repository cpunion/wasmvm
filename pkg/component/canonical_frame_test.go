package component

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func checkCanonicalFramesCleared(t *testing.T, in *Instance) int {
	t.Helper()
	count := 0
	for frame := in.canonicalFrames; frame != nil; frame = frame.next {
		count++
		if count > 8 {
			t.Fatal("cyclic frame cache")
		}
		if frame.caller != nil || frame.n != nil || frame.plan != nil || frame.memory.caller != nil || frame.state.codec.Resources != nil || frame.state.codec.Transfer.Context != nil || frame.state.codec.Transfer.Memory != nil || frame.state.codec.Transfer.Realloc != nil || frame.state.resources.core != nil || len(frame.args) != 0 {
			t.Fatal("released frame retains invocation state")
		}
		for _, v := range frame.args[:cap(frame.args)] {
			if v.Bits != 0 || v.Text != "" || v.Elements != nil || v.Resource != nil || v.Case != 0 {
				t.Fatal("released argument storage retains values")
			}
		}
	}
	return count
}

func TestCanonicalFramesPreserveSnapshotsAndReentry(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			var run *Function
			var callers []*Caller
			var retained [][]Value
			var expected []uint64
			nested, entered := false, false
			in := asyncInstance(t, backend, "canonical-frames", Imports{Functions: map[string]HostFunction{"step": func(c *Caller, args []Value) (*Value, error) {
				before := args[0].Bits
				callers = append(callers, c)
				alias := *c
				callers = append(callers, &alias)
				retained, expected = append(retained, args), append(expected, before)
				if !nested && !entered {
					entered, nested = true, true
					value, err := c.Call(run, []Value{{Bits: 1}, {Bits: 29}})
					nested = false
					if err != nil || value.Bits != 29*17+5 || args[0].Bits != before {
						t.Fatal("reentry overwrote parent arguments", value, args, err)
					}
				}
				return &Value{Bits: uint64(uint32(before)*17 + 5)}, nil
			}}})
			run, _ = in.Function("run")
			var first *Value
			for iteration := range 8 {
				seed := uint32(iteration) | 0x80000000
				want := seed
				for range 3 {
					want = want*17 + 5
				}
				got, err := run.Call(nil, []Value{{Bits: 3}, {Bits: uint64(seed)}}, CallOptions{})
				if err != nil || got.Value.Bits != uint64(want) {
					t.Fatal(got, err)
				}
				if first == nil {
					first = got.Value
				}
				for i, values := range retained {
					if values[0].Bits != expected[i] {
						t.Fatal("frame reuse changed a retained public argument", i, values, expected[i])
					}
					values[0].Bits = uint64(i) + 0x100000000
					expected[i] = values[0].Bits
				}
				for _, c := range callers {
					if _, err := c.Call(run, nil); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal("retained Caller became active again", err)
					}
				}
				if checkCanonicalFramesCleared(t, in) != 2 {
					t.Fatal("reentrant invocations did not retain distinct private frames")
				}
				runtime.GC()
			}
			wantFirst := uint32(0x80000000)
			for range 3 {
				wantFirst = wantFirst*17 + 5
			}
			if first.Bits != uint64(wantFirst) {
				t.Fatal("frame reuse changed an earlier public result", first)
			}
			in.Close()
			if err := in.WaitClosed(nil); err != nil || in.canonicalFrames != nil {
				t.Fatal("closed instance retained frame cache", err)
			}
		})
	}
}

func TestCanonicalFramesReturnAfterHostFailure(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"panic", "cancel", "invalid-result"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				in := asyncInstance(t, backend, "canonical-frames", Imports{Functions: map[string]HostFunction{"step": func(*Caller, []Value) (*Value, error) {
					if mode == "panic" {
						panic("frame panic")
					}
					if mode == "cancel" {
						cancel()
						return &Value{Bits: 7}, nil
					}
					return &Value{Bits: 1 << 32}, nil
				}}})
				run, _ := in.Function("run")
				func() {
					defer func() {
						p := recover()
						if mode == "panic" && p != "frame panic" || mode != "panic" && p != nil {
							t.Fatal("original host panic was not preserved", p)
						}
					}()
					_, err := run.Call(ctx, []Value{{Bits: 1}, {Bits: 7}}, CallOptions{})
					if mode == "panic" || err == nil || mode == "cancel" && !errors.Is(err, context.Canceled) {
						t.Fatal(mode, err)
					}
				}()
				if checkCanonicalFramesCleared(t, in) != 1 || in.heldValues != 0 || in.heldBytes != 0 || in.callDepth != 0 {
					t.Fatal("failed invocation leaked frame, budget or depth")
				}
			})
		}
	}
}

func TestCanonicalFramesRebindMemoryAndReallocatorAcrossLeases(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	memory, err := in.store.NewMemory(vm.MemoryType{Min: 1})
	if err != nil {
		t.Fatal(err)
	}
	n := &namespace{items: map[model.Sort][]item{model.CoreMemorySort: {{core: memory}}}, taskOps: true, task: &callTask{}}
	parent := n.task
	index := uint32(0)
	plan := &validate.CanonicalFunction{Definition: &model.Canonical{Op: 1}, Memory: &index, Realloc: &index}
	phase, allocations := 0, 0
	reallocator, err := in.store.NewScalarFunction(vm.FunctionType{Params: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(_ *vm.Caller, words []uint64) error {
		allocations++
		if !n.blocked || n.task == nil || n.task == parent {
			t.Fatal("canonical reallocator lost its isolated task/blocking state")
		}
		words[0] = uint64(phase)*65536 + 64
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	n.items[model.CoreFunctionSort] = []item{{core: reallocator}}
	var previous *canonicalFrame
	for ; phase < 2; phase++ {
		if phase != 0 {
			if _, ok, err := memory.GrowMemory(1); err != nil || !ok {
				t.Fatal(ok, err)
			}
		}
		_, err := in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
			frame := in.acquireCanonicalFrame(core, n, plan)
			defer in.releaseCanonicalFrame(frame)
			if previous != nil && previous != frame {
				t.Fatal("ordinary calls did not reuse their private frame")
			}
			previous = frame
			codec := &frame.state.codec
			if err := codec.Transfer.StoreString(0, "bound😀"); err != nil {
				return err
			}
			for range 32 {
				n.add(item{sort: model.CoreMemorySort})
			}
			runtime.GC()
			text, err := codec.Transfer.LoadString(0)
			if err != nil || text != "bound😀" || n.task != parent || n.blocked {
				t.Fatal(text, err, "reallocator changed its parent task")
			}
			return nil
		})
		if err != nil || checkCanonicalFramesCleared(t, in) != 1 {
			t.Fatal(err)
		}
	}
	if allocations != 2 {
		t.Fatal(allocations)
	}
}

func TestCanonicalFramesKeepOldBorrowScopesExpired(t *testing.T) {
	in := asyncInstance(t, "interpreter", "ownership", Imports{})
	get, _ := in.Function("get")
	makeResource, _ := in.Function("make")
	owned, err := makeResource.Call(nil, []Value{{Bits: 42}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	borrowType := get.typ.Parameters[0].Type
	plan := &validate.CanonicalFunction{Definition: &model.Canonical{Op: 1}}
	var old []*Resource
	for range 3 {
		_, err := in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
			frame := in.acquireCanonicalFrame(core, get.scope, plan)
			defer in.releaseCanonicalFrame(frame)
			codec := &frame.state.codec
			borrow := *owned.Value
			if err := codec.Validate(borrowType, borrow); err != nil {
				return err
			}
			if err := codec.AcceptResult(borrowType, &borrow); err != nil {
				return err
			}
			alias := *borrow.Resource
			old = append(old, borrow.Resource, &alias)
			for _, previous := range old[:len(old)-2] {
				if _, _, _, _, _, err := canonical.ReferenceInfo(previous); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("frame reuse revived an old borrow", err)
				}
			}
			runtime.GC()
			return finishResources(codec, true, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, previous := range old {
			if _, _, _, _, _, err := canonical.ReferenceInfo(previous); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("finished frame left a borrow active", err)
			}
		}
	}
	_, err = in.WithCaller(nil, CallOptions{}, func(c *Caller) error { return c.DropResource(owned.Value.Resource) })
	if err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFramesBoundScratchForSpilledParameters(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	memory, err := in.store.NewMemory(vm.MemoryType{Min: 1})
	if err != nil {
		t.Fatal(err)
	}
	parameters := make([]model.Field, canonical.MaxFlatParams+1)
	for i := range parameters {
		parameters[i].Type = &model.Type{Kind: model.U32}
	}
	index := uint32(0)
	plan := &validate.CanonicalFunction{Definition: &model.Canonical{Op: 1}, Type: &validate.Type{Parameters: parameters}, Memory: &index,
		ABI: canonical.FunctionLayout{ParameterMemory: &model.Type{Kind: model.Tuple, Fields: parameters}}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.I32}}}
	bound := newCanonicalCallPlan(plan)
	n := &namespace{items: map[model.Sort][]item{model.CoreMemorySort: {{core: memory}}}}
	_, err = in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
		frame := in.acquireCanonicalFrame(core, n, plan)
		defer in.releaseCanonicalFrame(frame)
		args, err := bound.liftScalarArguments(&frame.state.codec, []uint64{0}, frame.args[:0])
		if err != nil || len(args) != len(parameters) {
			t.Fatal("spilled parameters lost their declared arity", len(args), err)
		}
		frame.args = args
		return nil
	})
	if err != nil || in.canonicalFrames.args != nil {
		t.Fatal("unusually large argument scratch remained cached", err)
	}
	bad := &validate.CanonicalFunction{Type: &validate.Type{Parameters: []model.Field{{Type: &model.Type{Kind: model.String}}}}, Signature: vm.FunctionType{Params: []wasm.ValueType{wasm.I32}}}
	codec := canonical.Codec{Layouts: canonical.NewBuilder(false, canonical.DefaultLimits()), MaxValues: 100}
	if _, err := newCanonicalCallPlan(bad).liftScalarArguments(&codec, []uint64{0}, nil); !errors.Is(err, vm.ErrType) {
		t.Fatal("incomplete flattened argument representation accepted", err)
	}
}

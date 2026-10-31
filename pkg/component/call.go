package component

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Function struct {
	instance  *Instance
	typ       *validate.Type
	host      HostFunction
	asyncHost AsyncHostFunction
	scope     *namespace
	lift      func(*vm.Caller, []Value) (*Value, error)
	start     func(*vm.Caller, taskArguments, taskResolution) (*componentTask, error)
	plan      *validate.CanonicalFunction
}

type CallResult struct {
	Value         *Value
	FuelRemaining uint64
}

func (f *Function) Call(ctx context.Context, args []Value, options vm.CallOptions) (CallResult, error) {
	if f == nil || f.instance == nil {
		return CallResult{}, vm.ErrReference
	}
	var result CallResult
	run := func(c *vm.Caller) (err error) {
		complete := false
		defer f.instance.finishTasks(c, &err, &complete)
		result.Value, err = f.call(c, args)
		complete = true
		return err
	}
	var fuel uint64
	var err error
	if !f.typ.ContainsAffine() {
		fuel, err = f.instance.store.WithCaller(ctx, options, run)
	} else {
		fuel, err = f.instance.store.WithCallerCleanup(ctx, options, run, func(c *vm.Caller) error {
			var failed error
			if result.Value != nil {
				canonical.VisitReferences(f.typ.Result, result.Value, func(ref *Resource) {
					if err := f.instance.cleanupResource(c, ref); failed == nil {
						failed = err
					}
				})
			}
			return failed
		})
	}
	if err != nil {
		return CallResult{}, err
	}
	result.FuelRemaining = fuel
	return result, nil
}

// Caller expires at host callback return. Reentry uses the suspended Core
// session's cancellation, fuel and stack quotas.
type Caller struct{ *callerState }

type callerState struct {
	mu       sync.Mutex
	active   bool
	core     *vm.Caller
	instance *Instance
	function *Function
	created  []*Resource
	ctx      context.Context
}

// FunctionInfo returns this callback's immutable, instance-bound signature.
// The view may be retained; it does not extend the Caller's execution lifetime.
func (c *Caller) FunctionInfo() FunctionInfo {
	if c == nil || c.callerState == nil {
		return FunctionInfo{}
	}
	return c.function.Info()
}

func (c *Caller) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return c.core.Context()
}

func (c *Caller) enter() error {
	if c == nil || c.callerState == nil {
		return wasm.ErrClosed
	}
	if !c.mu.TryLock() {
		return vm.ErrBusy
	}
	if !c.active {
		c.mu.Unlock()
		return wasm.ErrClosed
	}
	if err := c.Context().Err(); err != nil {
		c.mu.Unlock()
		return err
	}
	return nil
}

func (c *Caller) Call(f *Function, args []Value) (*Value, error) {
	if err := c.enter(); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	return f.call(c.core, args)
}

func (f *Function) call(core *vm.Caller, args []Value) (result *Value, err error) {
	leave, err := f.enterCall(core, args)
	if err != nil {
		return nil, err
	}
	defer leave()

	if f.lift != nil {
		return f.lift(core, args)
	}
	if f.scope != nil && (f.scope.failed || f.scope.blocked) {
		return nil, wasm.ErrTrap
	}
	h, err := f.hostCall(core, args)
	if err != nil {
		return nil, err
	}
	complete := false
	defer h.finishDeferred(core, &complete, &err)
	result, err = h.invoke(core, func(c *Caller) (*Value, error) { return f.host(c, h.args) })
	if err != nil {
		return nil, err
	}
	result, err = h.accept(result)
	complete = err == nil
	return result, err
}

func (f *Function) enterCall(core *vm.Caller, args []Value) (func(), error) {
	if f == nil || f.instance == nil {
		return nil, vm.ErrReference
	}
	if f.instance.store != core.Store() {
		return nil, vm.ErrReference
	}
	if err := core.Context().Err(); err != nil {
		return nil, err
	}
	if len(args) != len(f.typ.Parameters) {
		return nil, vm.ErrType
	}
	if f.instance.callDepth > f.instance.config.Store.MaxReentry {
		return nil, wasm.ErrLimit
	}
	f.instance.callDepth++
	return func() { f.instance.callDepth-- }, nil
}

type memory struct {
	caller *vm.Caller
	extern vm.Extern
}

func (m memory) Read(at uint64, bytes []byte) error {
	return m.caller.ReadExternalMemory(m.extern, at, bytes)
}
func (m memory) Write(at uint64, bytes []byte) error {
	return m.caller.WriteExternalMemory(m.extern, at, bytes)
}

func (in *Instance) codec(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) *canonical.Codec {
	transfer := &canonical.Transfer{Context: caller.Context(), MaxBytes: in.config.MaxBytes}
	if plan != nil {
		transfer.Memory64, transfer.Encoding = plan.Memory64, plan.Encoding
		if plan.Memory != nil {
			transfer.Memory = memory{caller: caller, extern: n.items[model.CoreMemorySort][*plan.Memory].core}
		}
		if plan.Realloc != nil {
			fn := n.items[model.CoreFunctionSort][*plan.Realloc].core
			transfer.Realloc = func(old, size, alignment, next uint64) (uint64, error) {
				previous := n.task
				if n.taskOps {
					n.task = &callTask{}
				}
				defer func() { n.task = previous }()
				n.blocked = true
				defer func() { n.blocked = false }()
				kind := wasm.I32
				if plan.Memory64 {
					kind = wasm.I64
				}
				result, err := caller.Call(fn, vmValues([]canonical.Slot{{Type: kind, Bits: old}, {Type: kind, Bits: size}, {Type: kind, Bits: alignment}, {Type: kind, Bits: next}}))
				if err != nil {
					return 0, err
				}
				bits, _, err := result.Values[0].Bits()
				return bits, err
			}
		}
	}
	index := 0
	if transfer.Memory64 {
		index = 1
	}
	codec := &canonical.Codec{Transfer: transfer, Layouts: in.layouts[index], MaxValues: in.config.MaxValues}
	if n != nil && (len(n.resources) != 0 || n.handles != nil) {
		codec.Resources = &resourceCodec{in: in, n: n, core: caller}
	}
	return codec
}

func vmValues(slots []canonical.Slot) []vm.Value {
	values := make([]vm.Value, len(slots))
	for i, slot := range slots {
		switch slot.Type {
		case wasm.I32:
			values[i] = vm.I32(uint32(slot.Bits))
		case wasm.I64:
			values[i] = vm.I64(slot.Bits)
		case wasm.F32:
			values[i] = vm.F32Bits(uint32(slot.Bits))
		case wasm.F64:
			values[i] = vm.F64Bits(slot.Bits)
		}
	}
	return values
}

func flatValues(values []vm.Value) ([]canonical.Slot, error) {
	slots := make([]canonical.Slot, len(values))
	for i, v := range values {
		bits, _, err := v.Bits()
		if err != nil {
			return nil, err
		}
		slots[i] = canonical.Slot{Type: v.Type(), Bits: bits}
	}
	return slots, nil
}

func (in *Instance) canonical(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	if plan.Definition.Op >= 0x0e && plan.Definition.Op <= 0x1b || plan.Definition.Op == 0x2e || plan.Definition.Op == 0x2f {
		return in.asyncValueBuiltin(caller, n, plan)
	}
	if plan.Definition.Op == 0x24 || plan.Definition.Op == 0x25 {
		return in.backpressureBuiltin(caller, n, plan)
	}
	if plan.Definition.Op == 5 || plan.Definition.Op == 6 || plan.Definition.Op == 0x0d || plan.Definition.Op >= 0x1f && plan.Definition.Op <= 0x23 {
		return in.waitableBuiltin(caller, n, plan)
	}
	if plan.Definition.Op >= 9 && plan.Definition.Op <= 0x0c {
		return in.taskBuiltin(caller, n, plan)
	}
	if len(n.resources) != 0 || plan.Type != nil && (plan.Type.ContainsErrorContext() || plan.Type.ContainsAsyncValue()) || plan.Definition.Op >= 0x1c && plan.Definition.Op <= 0x1e || plan.Definition.Op == 1 && plan.Async {
		in.resourceTable(n)
	}
	if plan.Definition.Op >= 0x1c && plan.Definition.Op <= 0x1e {
		return in.errorContextBuiltin(caller, n, plan)
	}
	if plan.Definition.Op >= 2 && plan.Definition.Op <= 4 {
		return in.resourceBuiltin(caller, n, plan)
	}
	if plan.Definition.Op == 0 {
		fn := n.items[model.CoreFunctionSort][plan.Definition.Index].core
		f := &Function{instance: in, typ: plan.Type, scope: n, plan: plan}
		if plan.Type.Kind == model.AsyncFunction {
			f.start = func(c *vm.Caller, args taskArguments, resolve taskResolution) (*componentTask, error) {
				return in.startTask(c, n, plan, fn, args, resolve)
			}
			f.lift = in.asyncLift(f, plan)
			n.add(item{sort: model.FunctionSort, function: f})
			return nil
		}
		f.lift = func(caller *vm.Caller, args []Value) (result *Value, err error) {
			if n.failed || n.blocked || n.entered && (!plan.Async || plan.Callback != nil) {
				return nil, wasm.ErrTrap
			}
			codec := in.codec(caller, n, plan)
			for i, p := range plan.Type.Parameters {
				if err := codec.Validate(p.Type, args[i]); err != nil {
					return nil, err
				}
			}
			if codec.Resources != nil {
				if err := codec.PrepareParameters(plan.Type.Parameters, args); err != nil {
					return nil, err
				}
			}
			complete := false
			defer cleanupCanonical(n, codec, &complete, &err)
			if n.taskOps || plan.Async {
				previous := n.task
				task := &callTask{plan: plan, parameters: codec, entered: n.entered}
				n.task = task
				defer func() {
					guard := cleanupGuard{panicked: recover()}
					n.task = previous
					n.entered = task.entered
					if !complete || err != nil || guard.panicked != nil {
						n.failed = true
					}
					if task.returned != nil {
						guard.run(func() { err = finishResources(task.returned, complete, err) })
					}
					guard.finish()
				}()
			}
			n.entered = true
			defer func() { n.entered = false }()
			slots, err := lowerParameters(codec, plan, args)
			if err != nil {
				return nil, err
			}
			called, err := caller.Call(fn, vmValues(slots))
			if err != nil {
				return nil, err
			}
			if plan.Async {
				if !n.task.resolved {
					return nil, wasm.ErrTrap
				}
				complete = true
				return n.task.result, nil
			}
			if plan.Type.Result != nil {
				var value Value
				if plan.ABI.ResultMemory != nil {
					ptr, _, _ := called.Values[0].Bits()
					value, err = codec.Load(plan.Type.Result, ptr)
				} else {
					var flat []canonical.Slot
					flat, err = flatValues(called.Values)
					if err == nil {
						value, err = codec.LiftFlat(plan.Type.Result, flat)
					}
				}
				if err != nil {
					return nil, err
				}
				result = &value
			}
			if plan.PostReturn != nil {
				if err := returnBorrows(codec); err != nil {
					return nil, err
				}
				n.blocked = true
				_, err = caller.Call(n.items[model.CoreFunctionSort][*plan.PostReturn].core, called.Values)
				n.blocked = false
				if err != nil {
					return nil, err
				}
			}
			complete = true
			return result, nil
		}
		n.add(item{sort: model.FunctionSort, function: f})
		return nil
	}
	callee := n.items[model.FunctionSort][plan.Definition.Index].function
	if callee.start != nil {
		core, err := caller.NewSuspendableFunction(plan.Signature, func(c *vm.Caller, flat []vm.Value) (vm.HostResult, error) {
			if n.failed || n.blocked {
				return vm.HostResult{}, wasm.ErrTrap
			}
			return in.lowerTask(c, n, plan, flat, callee)
		})
		if err == nil {
			n.add(item{sort: model.CoreFunctionSort, core: core})
		}
		return err
	}
	core, err := caller.NewFunction(plan.Signature, func(caller *vm.Caller, flat []vm.Value) (result []vm.Value, err error) {
		if n.failed || n.blocked {
			return nil, wasm.ErrTrap
		}
		codec := in.codec(caller, n, plan)
		complete := false
		defer cleanupCanonical(n, codec, &complete, &err)
		args, err := liftArguments(codec, plan, flat)
		if err != nil {
			return nil, err
		}
		v, err := callee.call(caller, args)
		if err != nil {
			return nil, err
		}
		result, err = lowerResult(codec, plan, flat, v)
		if err != nil {
			return nil, err
		}
		complete = true
		if plan.Async {
			// Every admitted callee runs to completion in the current profile.
			// Eager resolution returns RETURNED without allocating a subtask.
			result = []vm.Value{vm.I32(2)}
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	n.add(item{sort: model.CoreFunctionSort, core: core})
	return nil
}

// CloseInstance requests cancellation of this component and all linked
// namespaces in its Store. It is nonblocking, including from a callback.
func (c *Caller) CloseInstance() error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	c.instance.Close()
	return nil
}

func lowerParameters(codec *canonical.Codec, plan *validate.CanonicalFunction, args []Value) ([]canonical.Slot, error) {
	var slots []canonical.Slot
	if typ := plan.ABI.ParameterMemory; typ != nil {
		layout, err := codec.Layouts.Build(typ)
		if err != nil {
			return nil, err
		}
		ptr, err := codec.Transfer.Realloc(0, 0, layout.Align, layout.Size)
		if err != nil {
			return nil, err
		}
		if err := codec.Store(typ, ptr, Value{Elements: args}); err != nil {
			return nil, err
		}
		slots = []canonical.Slot{{Type: plan.Signature.Params[0], Bits: ptr}}
	} else {
		for i, p := range plan.Type.Parameters {
			flat, err := codec.LowerFlat(p.Type, args[i])
			if err != nil {
				return nil, err
			}
			slots = append(slots, flat...)
		}
	}
	return slots, nil
}

func (in *Instance) asyncLift(f *Function, plan *validate.CanonicalFunction) func(*vm.Caller, []Value) (*Value, error) {
	return func(c *vm.Caller, args []Value) (*Value, error) {
		t, err := in.beginAsyncCall(c, f, plan, args)
		if err != nil {
			return nil, err
		}
		if err := in.awaitTasks(c, func() bool { return t.exited }); err != nil {
			return nil, err
		}
		return t.task.result, nil
	}
}

func (in *Instance) beginAsyncCall(c *vm.Caller, f *Function, plan *validate.CanonicalFunction, args []Value) (*componentTask, error) {
	n := f.scope
	if n != nil && (n.failed || n.blocked) {
		return nil, wasm.ErrTrap
	}
	codec := in.codec(c, n, plan)
	for i, p := range f.typ.Parameters {
		if err := codec.Validate(p.Type, args[i]); err != nil {
			return nil, err
		}
	}
	return f.start(c, func(*vm.Caller) ([]Value, error) { return args, nil }, nil)
}

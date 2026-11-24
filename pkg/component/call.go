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
	instance           *Instance
	typ                *validate.Type
	host               HostFunction
	asyncHost          AsyncHostFunction
	scope              *namespace
	lift               func(*vm.Caller, []Value) (*Value, error)
	start              func(*vm.Caller, taskArguments, taskResolution) (*componentTask, error)
	plan               *validate.CanonicalFunction
	values             *functionValues
	borrowedParameters bool
	// Accessed only while holding the instance Store lease. These descriptors
	// contain type metadata; each handoff still creates a fresh transaction.
	resultReceivers [2]*Function
}

type functionValues struct {
	parameters []canonical.ValuePlan
	result     canonical.ValuePlan
}

func newFunctionValues(typ *validate.Type) *functionValues {
	values := &functionValues{parameters: make([]canonical.ValuePlan, len(typ.Parameters)), result: canonical.NewValuePlan(typ.Result)}
	for i, parameter := range typ.Parameters {
		values.parameters[i] = canonical.NewValuePlan(parameter.Type)
	}
	return values
}

func (f *Function) valuePlans() *functionValues {
	if f.values == nil {
		f.values = newFunctionValues(f.typ)
	}
	return f.values
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

// Header copies keep the same guard while a callback's header and state share
// one allocation. Slots are consumed once and never rebound for a later call.
type callbackCallerAllocation struct {
	caller Caller
	state  callerState
}

type callbackCallerAllocator struct {
	block *[16]callbackCallerAllocation
	next  uint8
}

func (a *callbackCallerAllocator) new(core *vm.Caller, in *Instance, f *Function, prompt bool) *Caller {
	var allocation *callbackCallerAllocation
	if a.block == nil && a.next == 0 {
		a.next = 16
		allocation = &callbackCallerAllocation{}
	} else {
		if a.block == nil || a.next == uint8(len(a.block)) {
			a.block = new([16]callbackCallerAllocation)
			a.next = 0
		}
		allocation = &a.block[a.next]
		a.next++
	}
	allocation.state.active = true
	allocation.state.core = core
	allocation.state.instance = in
	allocation.state.function = f
	allocation.state.prompt = prompt
	allocation.caller.callerState = &allocation.state
	return &allocation.caller
}

func (a *callbackCallerAllocator) clear() { a.block, a.next = nil, 0 }

func newCallbackCaller(core *vm.Caller, in *Instance, f *Function, prompt bool) *Caller {
	if in != nil {
		return in.callers.new(core, in, f, prompt)
	}
	allocation := &callbackCallerAllocation{}
	allocation.state.active = true
	allocation.state.core = core
	allocation.state.instance = in
	allocation.state.function = f
	allocation.state.prompt = prompt
	allocation.caller.callerState = &allocation.state
	return &allocation.caller
}

type callerState struct {
	mu            sync.Mutex
	active        bool
	prompt        bool
	core          *vm.Caller
	instance      *Instance
	function      *Function
	inlineCreated [3]*Resource
	inlineReads   [1]*pendingHostRead
	created       []*Resource
	reads         []*pendingHostRead
	ctx           context.Context
}

func (c *callerState) addCreated(ref *Resource) {
	if c.created == nil {
		c.created = c.inlineCreated[:0]
	}
	c.created = append(c.created, ref)
}

func (c *callerState) addRead(r *pendingHostRead) {
	if c.reads == nil {
		c.reads = c.inlineReads[:0]
	}
	c.reads = append(c.reads, r)
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
	if c.prompt {
		return nil, vm.ErrType
	}
	return f.call(c.core, args)
}

func (f *Function) call(core *vm.Caller, args []Value) (result *Value, err error) {
	return f.callWithParameters(core, args, false, nil)
}

func (f *Function) callWithParameters(core *vm.Caller, args []Value, borrowed bool, frame *canonicalFrame) (result *Value, err error) {
	if err := f.enterCall(core, args); err != nil {
		return nil, err
	}
	defer f.leaveCall()

	if f.lift != nil {
		return f.lift(core, args)
	}
	if f.scope != nil && (f.scope.failed || f.scope.blocked) {
		return nil, wasm.ErrTrap
	}
	borrowResult := frame != nil
	if frame == nil {
		frame = f.instance.acquireFrame()
		defer f.instance.releaseCanonicalFrame(frame)
	}
	h, err := f.hostParameters(core, args, borrowed, frame.hostStorage())
	if err != nil {
		return nil, err
	}
	complete := false
	defer h.finishDeferred(core, &complete, &err)
	result, err = h.invoke(core, func(c *Caller) (*Value, error) { return f.host(c, h.args) })
	if err != nil {
		return nil, err
	}
	if borrowResult {
		frame.cloneBuf = frame.cloneValues[:0]
		err = h.acceptIntoScratch(result, &frame.hostValue, &frame.cloneBuf)
		if err != nil {
			return nil, err
		}
		if result != nil {
			result = &frame.hostValue
		}
	} else {
		result, err = h.accept(result)
	}
	complete = err == nil
	return result, err
}

func (f *Function) enterCall(core *vm.Caller, args []Value) error {
	if f == nil || f.instance == nil {
		return vm.ErrReference
	}
	if f.instance.store != core.Store() {
		return vm.ErrReference
	}
	if err := core.Context().Err(); err != nil {
		return err
	}
	if len(args) != len(f.typ.Parameters) {
		return vm.ErrType
	}
	if f.instance.callDepth > f.instance.config.Store.MaxReentry {
		return wasm.ErrLimit
	}
	f.instance.callDepth++
	return nil
}

func (f *Function) leaveCall() { f.instance.callDepth-- }

type memory struct {
	caller *vm.Caller
	// The namespace item is immutable after linking. Keep its capability by
	// reference instead of boxing a full Extern for every transfer session.
	extern *vm.Extern
}

func (m memory) Read(at uint64, bytes []byte) error {
	return m.caller.ReadExternalMemory(*m.extern, at, bytes)
}
func (m memory) Write(at uint64, bytes []byte) error {
	return m.caller.WriteExternalMemory(*m.extern, at, bytes)
}
func (m memory) ReadString(at, n uint64) (string, error) {
	return m.caller.ReadExternalMemoryString(*m.extern, at, n)
}
func (m memory) WriteString(at uint64, data string) error {
	return m.caller.WriteExternalMemoryString(*m.extern, at, data)
}

// Each state belongs to one call; grouping its objects does not reuse budgets
// or resource ownership across callers.
type resourceCodecState struct {
	codec     canonical.Codec
	resources resourceCodec
	memory    memory
}

func (in *Instance) initResourceCodecState(state *resourceCodecState, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, realloc bool) *canonical.Codec {
	codec := &state.codec
	if n != nil && (len(n.resources) != 0 || n.handles != nil) {
		state.resources = resourceCodec{in: in, n: n, core: caller}
		codec.Resources = &state.resources
	}
	in.configureCodecMemory(codec, &state.memory, caller, n, plan, realloc)
	return codec
}

// Synchronous transfers can use fresh caller-owned codec state from a frame.
// Endpoints retain their immutable options, never this session-bound scratch.
func (in *Instance) initBufferCodec(state *resourceCodecState, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, element *model.Type) *canonical.Codec {
	if state == nil {
		state = &resourceCodecState{}
	}
	if element != nil {
		switch element.Kind {
		case model.Bool, model.S8, model.U8, model.S16, model.U16, model.S32, model.U32,
			model.S64, model.U64, model.F32, model.F64, model.Char, model.String, model.Flags, model.Enum:
			// Only strings among these leaf elements can need canonical reallocation.
			in.configureCodecMemory(&state.codec, &state.memory, caller, n, plan, element.Kind == model.String)
			return &state.codec
		default:
			return in.initResourceCodecState(state, caller, n, plan, true)
		}
	}
	in.configureCodecMemory(&state.codec, &state.memory, caller, n, plan, false)
	return &state.codec
}

func (in *Instance) codec(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) *canonical.Codec {
	return in.initResourceCodecState(&resourceCodecState{}, caller, n, plan, true)
}

func (in *Instance) configureCodec(codec *canonical.Codec, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, realloc bool) {
	in.configureCodecMemory(codec, nil, caller, n, plan, realloc)
}

func (in *Instance) configureCodecMemory(codec *canonical.Codec, mem *memory, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, realloc bool) {
	in.configureCanonicalCodecMemory(codec, mem, caller, n, plan, realloc, nil)
}

func (in *Instance) configureCanonicalCodec(codec *canonical.Codec, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, realloc bool, frame *canonicalFrame) {
	in.configureCanonicalCodecMemory(codec, nil, caller, n, plan, realloc, frame)
}

func (in *Instance) configureCanonicalCodecMemory(codec *canonical.Codec, mem *memory, caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, realloc bool, frame *canonicalFrame) {
	transfer := &codec.Transfer
	*transfer = canonical.Transfer{Context: caller.Context(), MaxBytes: in.config.MaxBytes}
	if plan != nil {
		transfer.Memory64, transfer.Encoding = plan.Memory64, plan.Encoding
		if plan.Memory != nil {
			view := memory{caller: caller, extern: &n.items[model.CoreMemorySort][*plan.Memory].core}
			if frame != nil {
				frame.memory = view
				transfer.Memory = &frame.memory
			} else if mem != nil {
				*mem = view
				transfer.Memory = mem
			} else {
				transfer.Memory = view
			}
		}
		if realloc && plan.Realloc != nil {
			if frame == nil {
				transfer.Realloc = func(old, size, alignment, next uint64) (uint64, error) {
					return canonicalReallocate(caller, n, plan, old, size, alignment, next)
				}
			} else {
				if frame.realloc == nil {
					frame.realloc = frame.reallocate
				}
				transfer.Realloc = frame.realloc
			}
		}
	}
	codec.Layouts, codec.MaxValues = in.canonicalLayouts(n, plan), in.config.MaxValues
}

func canonicalReallocate(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, old, size, alignment, next uint64) (uint64, error) {
	fn := n.items[model.CoreFunctionSort][*plan.Realloc].core
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
	var argBuf [4]vm.Value
	slots := [4]canonical.Slot{{Type: kind, Bits: old}, {Type: kind, Bits: size}, {Type: kind, Bits: alignment}, {Type: kind, Bits: next}}
	result, err := caller.Call(fn, vmValuesInto(slots[:], argBuf[:0]))
	if err != nil {
		return 0, err
	}
	bits, _, err := result.Values[0].Bits()
	return bits, err
}

func (in *Instance) canonicalLayouts(n *namespace, plan *validate.CanonicalFunction) *canonical.Builder {
	wide := plan != nil && plan.Memory64
	index := 0
	if wide {
		index = 1
	}
	if plan != nil && plan.Definition.Op == 0 && n.wideResources {
		if n.lowerLayouts[index] == nil {
			n.lowerLayouts[index] = canonical.NewLoweringBuilder(wide, canonical.Limits{MaxNodes: in.config.Validate.MaxNodes, MaxDepth: in.config.MaxNesting, MaxBytes: in.config.MaxBytes}, n.borrowRepresentation)
		}
		return n.lowerLayouts[index]
	}
	return in.layouts[index]
}

func vmValues(slots []canonical.Slot) []vm.Value {
	if len(slots) == 0 {
		return nil
	}
	return vmValuesInto(slots, make([]vm.Value, 0, len(slots)))
}

func vmValuesInto(slots []canonical.Slot, dst []vm.Value) []vm.Value {
	for _, slot := range slots {
		switch slot.Type {
		case wasm.I32:
			dst = append(dst, vm.I32(uint32(slot.Bits)))
		case wasm.I64:
			dst = append(dst, vm.I64(slot.Bits))
		case wasm.F32:
			dst = append(dst, vm.F32Bits(uint32(slot.Bits)))
		case wasm.F64:
			dst = append(dst, vm.F64Bits(slot.Bits))
		}
	}
	return dst
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
			if n.failed || n.blocked {
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
			previous, entered := n.task, n.entered
			task := &callTask{plan: plan, parameters: codec}
			n.task, n.entered = task, true
			defer func() {
				guard := cleanupGuard{panicked: recover()}
				n.task, n.entered = previous, entered
				if !complete || err != nil || guard.panicked != nil {
					n.failed = true
				}
				if task.returned != nil {
					err = guard.finishCodec(task.returned, complete, err)
				}
				guard.finish()
			}()

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
	bound := newCanonicalCallPlan(plan)
	core, err := caller.NewBorrowedScalarFunction(plan.Signature, func(caller *vm.Caller, words []uint64) (err error) {
		if n.failed || n.blocked {
			return wasm.ErrTrap
		}
		frame := in.acquireCanonicalFrame(caller, n, plan)
		defer in.releaseCanonicalFrame(frame)
		codec := &frame.state.codec
		complete := false
		defer cleanupCanonical(n, codec, &complete, &err)
		args, err := bound.liftScalarArguments(codec, words, frame.args[:0])
		// The partial prefix is frame-owned even when a later argument fails.
		frame.args = args
		if err != nil {
			return err
		}
		v, err := callee.callWithParameters(caller, args, callee.borrowedParameters, frame)
		if err != nil {
			return err
		}
		err = bound.lowerScalarResult(codec, words, v, frame.result[:0])
		if err != nil {
			return err
		}
		complete = true
		if plan.Async {
			// Every admitted callee runs to completion in the current profile.
			// Eager resolution returns RETURNED without allocating a subtask.
			words[0] = 2
		}
		return nil
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
	frame := in.acquireCanonicalFrame(c, n, plan)
	codec := &frame.state.codec
	for i, p := range f.typ.Parameters {
		if err := codec.Validate(p.Type, args[i]); err != nil {
			in.releaseCanonicalFrame(frame)
			return nil, err
		}
	}
	in.releaseCanonicalFrame(frame)
	return f.start(c, func(*vm.Caller) ([]Value, error) { return args, nil }, nil)
}

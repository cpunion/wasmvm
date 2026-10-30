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
	instance *Instance
	typ      *validate.Type
	host     HostFunction
	lift     func(*vm.Caller, []Value) (*Value, error)
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
	fuel, err := f.instance.store.WithCaller(ctx, options, func(c *vm.Caller) error {
		var err error
		result.Value, err = f.call(c, args)
		return err
	})
	if err != nil {
		return CallResult{}, err
	}
	result.FuelRemaining = fuel
	return result, nil
}

// Caller expires at host callback return. Reentry uses the suspended Core
// session's cancellation, fuel and stack quotas.
type Caller struct {
	mu     sync.Mutex
	active bool
	core   *vm.Caller
}

func (c *Caller) Context() context.Context { return c.core.Context() }

func (c *Caller) Call(f *Function, args []Value) (*Value, error) {
	if !c.mu.TryLock() {
		return nil, vm.ErrBusy
	}
	defer c.mu.Unlock()
	if !c.active {
		return nil, wasm.ErrClosed
	}
	return f.call(c.core, args)
}

func (f *Function) call(core *vm.Caller, args []Value) (*Value, error) {
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
	defer func() { f.instance.callDepth-- }()
	if f.lift != nil {
		return f.lift(core, args)
	}
	codec := f.instance.codec(core, nil, nil)
	owned := make([]Value, len(args))
	for i, p := range f.typ.Parameters {
		v, err := codec.Clone(p.Type, args[i])
		if err != nil {
			return nil, err
		}
		owned[i] = v
	}
	caller := &Caller{core: core, active: true}
	defer func() { caller.mu.Lock(); caller.active = false; caller.mu.Unlock() }()
	result, err := f.host(caller, owned)
	if err != nil {
		return nil, err
	}
	if (result == nil) != (f.typ.Result == nil) {
		return nil, vm.ErrType
	}
	if result != nil {
		v, err := codec.Clone(f.typ.Result, *result)
		if err != nil {
			return nil, err
		}
		result = &v
	}
	return result, nil
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
	return &canonical.Codec{Transfer: transfer, Layouts: in.layouts[index], MaxValues: in.config.MaxValues}
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
	if plan.Definition.Op >= 2 && plan.Definition.Op <= 4 {
		return in.resourceBuiltin(caller, n, plan)
	}
	if plan.Definition.Op == 0 {
		fn := n.items[model.CoreFunctionSort][plan.Definition.Index].core
		f := &Function{instance: in, typ: plan.Type}
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
			complete := false
			defer func() {
				if !complete {
					n.failed = true
				}
			}()
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
			called, err := caller.Call(fn, vmValues(slots))
			if err != nil {
				return nil, err
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
	core, err := caller.NewFunction(plan.Signature, func(caller *vm.Caller, flat []vm.Value) (result []vm.Value, err error) {
		if n.failed || n.blocked {
			return nil, wasm.ErrTrap
		}
		complete := false
		defer func() {
			if !complete {
				n.failed = true
			}
		}()
		codec := in.codec(caller, n, plan)
		var args []Value
		if plan.ABI.ParameterMemory != nil {
			ptr, _, _ := flat[0].Bits()
			v, err := codec.Load(plan.ABI.ParameterMemory, ptr)
			if err != nil {
				return nil, err
			}
			args = v.Elements
		} else {
			slots, err := flatValues(flat)
			if err != nil {
				return nil, err
			}
			for _, p := range plan.Type.Parameters {
				layout, err := codec.Layouts.Build(p.Type)
				if err != nil {
					return nil, err
				}
				v, err := codec.LiftFlat(p.Type, slots[:layout.FlatCount])
				if err != nil {
					return nil, err
				}
				args = append(args, v)
				slots = slots[layout.FlatCount:]
			}
		}
		v, err := callee.call(caller, args)
		if err != nil {
			return nil, err
		}
		if v != nil {
			if plan.ABI.ResultMemory != nil {
				ptr, _, _ := flat[len(flat)-1].Bits()
				err = codec.Store(plan.Type.Result, ptr, *v)
			} else {
				var slots []canonical.Slot
				slots, err = codec.LowerFlat(plan.Type.Result, *v)
				result = vmValues(slots)
			}
			if err != nil {
				return nil, err
			}
		}
		complete = true
		return result, nil
	})
	if err != nil {
		return err
	}
	n.add(item{sort: model.CoreFunctionSort, core: core})
	return nil
}

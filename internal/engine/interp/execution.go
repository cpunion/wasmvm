// Package interp executes immutable slot code with explicit guest frames and
// host-request exits. Guest calls never recurse on the host Go stack.
package interp

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Status byte

const (
	Running Status = iota
	Done
	HostRequest
	Failed
)

type Limits struct{ MaxFrames, MaxSlots uint32 }

func DefaultLimits() Limits { return Limits{1024, 1 << 20} }

type Config struct {
	Limits  Limits
	Metered bool
	Fuel    uint64
}
type frame struct {
	function *state.Function
	pc       uint32
	base     int
	dest     []slots.Slot
}
type HostCall struct {
	Function *state.Function
	Args     []value.Value
}
type Execution struct {
	store       *state.Store
	context     context.Context
	config      Config
	frames      []frame
	values      []value.Value
	result      []value.Value
	pending     *HostCall
	pendingDest []slots.Slot
	status      Status
	err         error
	fuel, steps uint64
}

func New(store *state.Store, ctx context.Context, config Config) *Execution {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Execution{store: store, context: ctx, config: config, fuel: config.Fuel, status: Done}
}
func (e *Execution) Fuel() uint64 { return e.fuel }
func (e *Execution) Err() error   { return e.err }

// Results are borrowed until the next Start. Host calls expose copied input
// values that remain stable while the driver services the request.
func (e *Execution) Results() []value.Value { return e.result }
func (e *Execution) Host() *HostCall        { return e.pending }
func (e *Execution) Start(ref uint64, args []value.Value) error {
	if e.status == Running || e.status == HostRequest {
		return errors.New("execution already active")
	}
	fn, ok := e.store.Function(ref)
	if !ok {
		return state.ErrReference
	}
	if len(args) != len(fn.Signature.Params) {
		return errors.New("function argument count")
	}
	e.status, e.err = Running, nil
	e.result = e.result[:0]
	e.fuel = e.config.Fuel
	e.steps = 0
	if fn.Code == nil {
		e.pending = &HostCall{Function: fn, Args: append([]value.Value(nil), args...)}
		e.status = HostRequest
		return nil
	}
	if err := e.push(fn, nil); err != nil {
		e.fail(err)
		return err
	}
	copy(e.values, args)
	return nil
}
func (e *Execution) push(fn *state.Function, dest []slots.Slot) error {
	if uint64(len(e.frames)) >= uint64(e.config.Limits.MaxFrames) {
		return &wasm.Error{Kind: wasm.ErrLimit, Detail: "guest call stack exhausted"}
	}
	n := len(fn.Code.Slots)
	base := len(e.values)
	if uint64(n) > uint64(e.config.Limits.MaxSlots)-uint64(base) {
		return &wasm.Error{Kind: wasm.ErrLimit, Detail: "guest frame slots"}
	}
	needed := base + n
	if needed > cap(e.values) {
		capacity := cap(e.values) * 2
		if capacity < needed {
			capacity = needed
		}
		if uint64(capacity) > uint64(e.config.Limits.MaxSlots) {
			capacity = int(e.config.Limits.MaxSlots)
		}
		grown := make([]value.Value, needed, capacity)
		copy(grown, e.values)
		e.values = grown
	} else {
		e.values = e.values[:needed]
		clear(e.values[base:])
	}
	e.frames = append(e.frames, frame{function: fn, base: base, dest: dest})
	return nil
}
func (e *Execution) fail(err error) Status {
	e.err, e.status = err, Failed
	clear(e.frames)
	e.frames = e.frames[:0]
	e.values = e.values[:0]
	e.pending = nil
	e.pendingDest = nil
	return Failed
}

// Abort releases every active frame and pending host request. Host drivers use
// it in deferred panic cleanup before propagating the original host panic.
func (e *Execution) Abort(err error) { e.fail(err) }
func (e *Execution) poll() error     { return e.context.Err() }
func (e *Execution) ResumeHost(results []value.Value, err error) error {
	if e.status != HostRequest || e.pending == nil {
		return errors.New("no pending host request")
	}
	if err != nil {
		e.fail(err)
		return err
	}
	if len(results) != len(e.pending.Function.Signature.Results) {
		err = errors.New("host result count")
		e.fail(err)
		return err
	}
	if len(e.frames) == 0 {
		e.result = append(e.result[:0], results...)
		e.status = Done
	} else {
		base := e.frames[len(e.frames)-1].base
		for i, dst := range e.pendingDest {
			e.values[base+int(dst)] = results[i]
		}
		e.status = Running
	}
	e.pending = nil
	e.pendingDest = nil
	return nil
}
func (e *Execution) Run() Status {
	if e.status != Running {
		return e.status
	}
	if err := e.poll(); err != nil {
		return e.fail(err)
	}
	for len(e.frames) > 0 {
		fi := len(e.frames) - 1
		f := &e.frames[fi]
		fn := f.function
		code := fn.Code
		pc := f.pc
		base := f.base
		in := code.Code[pc]
		source := code.Sources[in.Source]
		im := source.Wasm
		instance := fn.Instance
		if e.config.Metered {
			cost := source.Fuel()
			if cost > e.fuel {
				return e.fail(&wasm.Error{Kind: wasm.ErrLimit, Offset: im.Offset, Detail: "execution fuel"})
			}
			e.fuel -= cost
		}
		e.steps++
		if e.steps&63 == 0 {
			if err := e.poll(); err != nil {
				return e.fail(err)
			}
		}
		f.pc++
		get := func(s slots.Slot) value.Value {
			if s == slots.NoSlot {
				return value.Value{}
			}
			return e.values[base+int(s)]
		}
		a, b, c := get(in.A), get(in.B), get(in.C)
		var result value.Value
		trap := wasm.NoTrap
		switch op := in.Op; {
		case op >= slots.ControlBase:
			kind := ir.TermKind(op - slots.ControlBase)
			if kind == ir.Trap {
				trap = wasm.TrapUnreachable
				break
			}
			if kind == ir.Return {
				outputs := code.Args[in.Extra : in.Extra+in.Count]
				if fi == 0 {
					e.result = e.result[:0]
					for _, src := range outputs {
						e.result = append(e.result, e.values[base+int(src)])
					}
					clear(e.frames)
					e.frames = e.frames[:0]
					e.values = e.values[:0]
					e.status = Done
					return Done
				}
				parentBase := e.frames[fi-1].base
				for i, src := range outputs {
					e.values[parentBase+int(f.dest[i])] = e.values[base+int(src)]
				}
				e.frames[fi] = frame{}
				e.frames = e.frames[:fi]
				e.values = e.values[:base]
				continue
			}
			index := uint32(0)
			if kind == ir.If && uint32(a.Lo) == 0 {
				index = 1
			} else if kind == ir.Table {
				index = uint32(a.Lo)
				if index >= in.Count-1 {
					index = in.Count - 1
				}
			}
			edge := code.Edges[in.Extra+index]
			if edge.PC <= pc {
				if err := e.poll(); err != nil {
					return e.fail(err)
				}
			}
			for _, move := range code.Moves[edge.Start : edge.Start+edge.Count] {
				e.values[base+int(move.Dst)] = e.values[base+int(move.Src)]
			}
			f.pc = edge.PC
			continue
		case op == 1 || op == 0x1a:
		case op == 0x1b || op == 0x1c:
			result = b
			if uint32(c.Lo) != 0 {
				result = a
			}
		case op == 0x20:
			result = e.values[base+int(im.A)]
		case op == 0x21 || op == 0x22:
			e.values[base+int(im.A)] = a
			result = a
		case op == 0x23:
			result = instance.Globals[im.A].Value
		case op == 0x24:
			instance.Globals[im.A].Value = a
		case op == 0x25:
			table := instance.Tables[im.A]
			index := uint64(uint32(a.Lo))
			if index >= uint64(len(table.Values)) {
				trap = wasm.TrapTableBounds
			} else {
				result.Lo = table.Values[index]
			}
		case op == 0x26:
			table := instance.Tables[im.A]
			index := uint64(uint32(a.Lo))
			if index >= uint64(len(table.Values)) {
				trap = wasm.TrapTableBounds
			} else {
				table.Values[index] = b.Lo
			}
		case op >= 0x28 && op <= 0x35:
			result, trap = instance.Memories[0].Load(op, uint32(a.Lo), im.B, 0, value.Value{})
		case op >= 0x36 && op <= 0x3e:
			trap = instance.Memories[0].Store(op, uint32(a.Lo), im.B, 0, b)
		case op == 0x3f:
			result.Lo = uint64(instance.Memories[0].Pages())
		case op == 0x40:
			old, ok := e.store.GrowMemory(instance.Memories[0], uint32(a.Lo))
			result.Lo = uint64(old)
			if !ok {
				result.Lo = uint64(^uint32(0))
			}
		case op >= 0x41 && op <= 0x44:
			result.Lo = im.Bits
		case op >= 0x45 && op <= 0xc4 || op >= 0xfc0000 && op <= 0xfc0007:
			result.Lo, trap = numeric.Eval(op, a.Lo, b.Lo)
		case op == 0xd0:
		case op == 0xd1:
			if a.Lo == 0 {
				result.Lo = 1
			}
		case op == 0xd2:
			result.Lo = instance.Functions[im.A]
		case op == 0x10 || op == 0x11:
			count := in.Count
			inputs := code.Args[in.Extra : in.Extra+count]
			var target uint64
			if op == 0x10 {
				target = instance.Functions[im.A]
			} else {
				count--
				index := uint64(uint32(e.values[base+int(inputs[count])].Lo))
				table := instance.Tables[im.B]
				if index >= uint64(len(table.Values)) {
					trap = wasm.TrapTableBounds
					break
				}
				target = table.Values[index]
				if target == 0 {
					trap = wasm.TrapUninitializedElement
					break
				}
			}
			callee, ok := e.store.Function(target)
			if !ok {
				return e.fail(state.ErrReference)
			}
			if op == 0x11 && callee.TypeID != instance.Types[im.A] {
				trap = wasm.TrapIndirectCallType
				break
			}
			dest := code.Args[in.Extra+in.Count : in.Extra+in.Count+uint32(len(callee.Signature.Results))]
			if err := e.poll(); err != nil {
				return e.fail(err)
			}
			if callee.Code == nil {
				args := make([]value.Value, count)
				for i := range args {
					args[i] = e.values[base+int(inputs[i])]
				}
				e.pending = &HostCall{Function: callee, Args: args}
				e.pendingDest = dest
				e.status = HostRequest
				return HostRequest
			}
			if err := e.push(callee, dest); err != nil {
				return e.fail(err)
			}
			calleeBase := e.frames[len(e.frames)-1].base
			for i := uint32(0); i < count; i++ {
				e.values[calleeBase+int(i)] = e.values[base+int(inputs[i])]
			}
			continue
		case op>>16 == 0xfd:
			sub := op & 0xffff
			switch {
			case sub <= 10 || sub >= 84 && sub <= 87 || sub == 92 || sub == 93:
				result, trap = instance.Memories[0].Load(op, uint32(a.Lo), im.B, im.Lane, b)
			case sub == 11 || sub >= 88 && sub <= 91:
				trap = instance.Memories[0].Store(op, uint32(a.Lo), im.B, im.Lane, b)
			default:
				result = simd.Eval(im, a, b, c)
			}
		case op >= 0xfc0008 && op <= 0xfc0011:
			var err error
			result, trap, err = e.bulk(instance, im, a, b, c)
			if err != nil {
				return e.fail(err)
			}
		default:
			panic("interp: unknown validated opcode")
		}
		if trap != wasm.NoTrap {
			return e.fail(&wasm.Trap{Code: trap, Function: fn.Index, Offset: im.Offset})
		}
		if in.Dst != slots.NoSlot {
			e.values[base+int(in.Dst)] = result
		}
	}
	panic("interp: running without frames")
}

// RetainedValues covers reference slots in active or suspended frames. Slot
// types remain fixed even when temporaries are reused, including edge scratch.
func (e *Execution) RetainedValues(visit func(wasm.ValueType, value.Value)) {
	for _, f := range e.frames {
		for i, t := range f.function.Code.Slots {
			if t.IsReference() {
				visit(t, e.values[f.base+i])
			}
		}
	}
	if e.pending != nil {
		for i, t := range e.pending.Function.Signature.Params {
			if t.IsReference() {
				visit(t, e.pending.Args[i])
			}
		}
	}
}

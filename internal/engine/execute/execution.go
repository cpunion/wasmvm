// Package execute runs immutable slot code with explicit guest frames and
// host-request exits. Guest calls never recurse on the host Go stack.
package execute

import (
	"context"
	"errors"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
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
	Backend Backend
}
type frame struct {
	function  *state.Function
	pc        uint32
	callPC    uint32
	suspended bool
	base      int
	dest      []slots.Slot
}
type HostCall struct {
	Function *state.Function
	Caller   *state.Instance
	Args     []value.Value
}
type Execution struct {
	store       *state.Store
	context     context.Context
	config      Config
	frames      []frame
	values      []value.Value
	result      []value.Value
	tailArgs    []value.Value
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

// SetFuel synchronizes a suspended continuation with its driver's shared
// reentry budget. The driver never changes it while Run is executing.
func (e *Execution) SetFuel(fuel uint64) { e.fuel = fuel }
func (e *Execution) Footprint() Limits   { return Limits{uint32(len(e.frames)), uint32(len(e.values))} }
func (e *Execution) Err() error          { return e.err }

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
		if e.config.Backend != nil {
			grown, err := e.config.Backend.Resize(needed, capacity)
			if err != nil {
				return err
			}
			e.values = grown
		} else {
			grown := make([]value.Value, needed, capacity)
			copy(grown, e.values)
			e.values = grown
		}
	} else {
		e.values = e.values[:needed]
	}
	clear(e.values[base:])
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
	clear(e.tailArgs)
	e.tailArgs = e.tailArgs[:0]
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
		if ex, ok := e.store.ExceptionError(err); ok {
			e.pending = nil
			e.pendingDest = nil
			if e.raise(ex) {
				e.status = Running
				return nil
			}
			return e.err
		}
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
	backend := e.config.Backend
	native := backend != nil
	metered := e.config.Metered && !native
frames:
	for len(e.frames) > 0 {
		fi := len(e.frames) - 1
		f := &e.frames[fi]
		f.suspended = false
		fn := f.function
		code := fn.Code
		base := f.base
		instance := fn.Instance
		pc := f.pc
		values := e.values[base : base+len(code.Slots)]
		for {
			if native {
				var err error
				pc, e.fuel, err = backend.Run(fn, base, len(e.frames), pc, e.fuel)
				if err != nil {
					return e.fail(err)
				}
			}
			in := code.Code[pc]
			source := &code.Sources[in.Source]
			im := &source.Wasm
			if metered {
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
			var a, b, c value.Value
			if in.A != slots.NoSlot {
				a = values[int(in.A)]
			}
			if in.B != slots.NoSlot {
				b = values[int(in.B)]
			}
			if in.C != slots.NoSlot {
				c = values[int(in.C)]
			}
			var result value.Value
			trap := wasm.NoTrap
			if in.Op>>16 == 0xfb {
				var local [5]value.Value
				operands := local[:0]
				if int(in.Count) > cap(operands) {
					operands = make([]value.Value, 0, in.Count)
				}
				for _, slot := range code.Args[in.Extra : in.Extra+in.Count] {
					operands = append(operands, values[int(slot)])
				}
				var err error
				result, err = e.store.GC(e.context, instance, im, operands, e.RetainedValues)
				if err != nil {
					var t *wasm.Trap
					if errors.As(err, &t) {
						t.Function = fn.Index
					}
					return e.fail(err)
				}
				if in.Dst != slots.NoSlot {
					values[int(in.Dst)] = result
				}
				pc++
				f.pc = pc
				continue
			}
			switch op := in.Op; op {
			case slots.ControlBase + uint32(ir.IfCast), slots.ControlBase + uint32(ir.IfNull), slots.ControlBase + uint32(ir.Jump), slots.ControlBase + uint32(ir.If), slots.ControlBase + uint32(ir.Table), slots.ControlBase + uint32(ir.Return), slots.ControlBase + uint32(ir.Trap):
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
							e.result = append(e.result, values[int(src)])
						}
						clear(e.frames)
						e.frames = e.frames[:0]
						e.values = e.values[:0]
						e.status = Done
						return Done
					}
					parentBase := e.frames[fi-1].base
					for i, src := range outputs {
						e.values[parentBase+int(f.dest[i])] = values[int(src)]
					}
					e.frames[fi] = frame{}
					e.frames = e.frames[:fi]
					e.values = e.values[:base]
					continue frames
				}
				index := uint32(0)
				if kind == ir.IfCast && !e.store.TestReference(a.Lo, module.ResolveType(im.Types[1], instance.Types)) {
					index = 1
				} else if kind == ir.IfNull && a.Lo != 0 {
					index = 1
				} else if kind == ir.If && uint32(a.Lo) == 0 {
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
					values[int(move.Dst)] = values[int(move.Src)]
				}
				pc = edge.PC
				continue

			case 8, 10:
				f.pc = pc
				var ex *state.Exception
				if op == 8 {
					var local [8]value.Value
					operands := local[:0]
					if int(in.Count) > cap(operands) {
						operands = make([]value.Value, in.Count)
					} else {
						operands = operands[:in.Count]
					}
					for i, s := range code.Args[in.Extra : in.Extra+in.Count] {
						operands[i] = values[int(s)]
					}
					var err error
					ex, err = e.store.NewException(instance.Tags[im.A], operands, e.RetainedValues)
					if err != nil {
						return e.fail(err)
					}
				} else {
					ref := values[int(code.Args[in.Extra])].Lo
					if ref == 0 {
						trap = wasm.TrapNullReference
						break
					}
					ex = &state.Exception{Owner: e.store, Ref: ref}
				}
				if e.raise(ex) {
					continue frames
				}
				return Failed
			case 1, 0x1a:
			case 0x1b, 0x1c:
				result = b
				if uint32(c.Lo) != 0 {
					result = a
				}
			case 0x20:
				result = values[int(im.A)]
			case 0x21, 0x22:
				values[int(im.A)] = a
				result = a
			case 0x23:
				result = instance.Globals[im.A].Value
			case 0x24:
				instance.Globals[im.A].Value = a
			case 0x25:
				table := instance.Tables[im.A]
				index := table.Type.Limits.Address(a.Lo)
				if index >= uint64(len(table.Values)) {
					trap = wasm.TrapTableBounds
				} else {
					result.Lo = table.Values[index]
				}
			case 0x26:
				table := instance.Tables[im.A]
				index := table.Type.Limits.Address(a.Lo)
				if index >= uint64(len(table.Values)) {
					trap = wasm.TrapTableBounds
				} else {
					table.Values[index] = b.Lo
				}
			case 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
				0x30, 0x31, 0x32, 0x33, 0x34, 0x35:
				result, trap = instance.Memories[im.Memory].Load(op, a.Lo, im.MemoryOffset(), 0, value.Value{})
			case 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
				0x3e:
				trap = instance.Memories[im.Memory].Store(op, a.Lo, im.MemoryOffset(), 0, b)
			case 0x3f:
				result.Lo = uint64(instance.Memories[im.A].Pages())
			case 0x40:
				mem := instance.Memories[im.A]
				old, ok := e.store.GrowMemory(mem, mem.Limits().Address(a.Lo))
				result.Lo = uint64(old)
				if !ok {
					result.Lo = mem.Limits().Failure()
				}
			case 0x41, 0x42, 0x43, 0x44:
				result.Lo = im.Bits
			case 0x45:
				result.Lo = numeric.I32Eqz(a.Lo)
			case 0x46:
				result.Lo = numeric.I32Eq(a.Lo, b.Lo)
			case 0x47:
				result.Lo = numeric.I32Ne(a.Lo, b.Lo)
			case 0x48:
				result.Lo = numeric.I32LtS(a.Lo, b.Lo)
			case 0x49:
				result.Lo = numeric.I32LtU(a.Lo, b.Lo)
			case 0x4a:
				result.Lo = numeric.I32GtS(a.Lo, b.Lo)
			case 0x4b:
				result.Lo = numeric.I32GtU(a.Lo, b.Lo)
			case 0x4c:
				result.Lo = numeric.I32LeS(a.Lo, b.Lo)
			case 0x4d:
				result.Lo = numeric.I32LeU(a.Lo, b.Lo)
			case 0x4e:
				result.Lo = numeric.I32GeS(a.Lo, b.Lo)
			case 0x4f:
				result.Lo = numeric.I32GeU(a.Lo, b.Lo)
			case 0x50:
				result.Lo = numeric.I64Eqz(a.Lo)
			case 0x51:
				result.Lo = numeric.I64Eq(a.Lo, b.Lo)
			case 0x52:
				result.Lo = numeric.I64Ne(a.Lo, b.Lo)
			case 0x53:
				result.Lo = numeric.I64LtS(a.Lo, b.Lo)
			case 0x54:
				result.Lo = numeric.I64LtU(a.Lo, b.Lo)
			case 0x55:
				result.Lo = numeric.I64GtS(a.Lo, b.Lo)
			case 0x56:
				result.Lo = numeric.I64GtU(a.Lo, b.Lo)
			case 0x57:
				result.Lo = numeric.I64LeS(a.Lo, b.Lo)
			case 0x58:
				result.Lo = numeric.I64LeU(a.Lo, b.Lo)
			case 0x59:
				result.Lo = numeric.I64GeS(a.Lo, b.Lo)
			case 0x5a:
				result.Lo = numeric.I64GeU(a.Lo, b.Lo)
			case 0x67:
				result.Lo = numeric.I32Clz(a.Lo)
			case 0x68:
				result.Lo = numeric.I32Ctz(a.Lo)
			case 0x69:
				result.Lo = numeric.I32Popcnt(a.Lo)
			case 0x6a:
				result.Lo = numeric.I32Add(a.Lo, b.Lo)
			case 0x6b:
				result.Lo = numeric.I32Sub(a.Lo, b.Lo)
			case 0x6c:
				result.Lo = numeric.I32Mul(a.Lo, b.Lo)
			case 0x71:
				result.Lo = numeric.I32And(a.Lo, b.Lo)
			case 0x72:
				result.Lo = numeric.I32Or(a.Lo, b.Lo)
			case 0x73:
				result.Lo = numeric.I32Xor(a.Lo, b.Lo)
			case 0x74:
				result.Lo = numeric.I32Shl(a.Lo, b.Lo)
			case 0x75:
				result.Lo = numeric.I32ShrS(a.Lo, b.Lo)
			case 0x76:
				result.Lo = numeric.I32ShrU(a.Lo, b.Lo)
			case 0x77:
				result.Lo = numeric.I32Rotl(a.Lo, b.Lo)
			case 0x78:
				result.Lo = numeric.I32Rotr(a.Lo, b.Lo)
			case 0x79:
				result.Lo = numeric.I64Clz(a.Lo)
			case 0x7a:
				result.Lo = numeric.I64Ctz(a.Lo)
			case 0x7b:
				result.Lo = numeric.I64Popcnt(a.Lo)
			case 0x7c:
				result.Lo = numeric.I64Add(a.Lo, b.Lo)
			case 0x7d:
				result.Lo = numeric.I64Sub(a.Lo, b.Lo)
			case 0x7e:
				result.Lo = numeric.I64Mul(a.Lo, b.Lo)
			case 0x83:
				result.Lo = numeric.I64And(a.Lo, b.Lo)
			case 0x84:
				result.Lo = numeric.I64Or(a.Lo, b.Lo)
			case 0x85:
				result.Lo = numeric.I64Xor(a.Lo, b.Lo)
			case 0x86:
				result.Lo = numeric.I64Shl(a.Lo, b.Lo)
			case 0x87:
				result.Lo = numeric.I64ShrS(a.Lo, b.Lo)
			case 0x88:
				result.Lo = numeric.I64ShrU(a.Lo, b.Lo)
			case 0x89:
				result.Lo = numeric.I64Rotl(a.Lo, b.Lo)
			case 0x8a:
				result.Lo = numeric.I64Rotr(a.Lo, b.Lo)
			case 0xa7:
				result.Lo = numeric.I32WrapI64(a.Lo)
			case 0xac:
				result.Lo = numeric.I64ExtendI32S(a.Lo)
			case 0xad:
				result.Lo = numeric.I64ExtendI32U(a.Lo)
			case 0xc0:
				result.Lo = numeric.I32Extend8S(a.Lo)
			case 0xc1:
				result.Lo = numeric.I32Extend16S(a.Lo)
			case 0xc2:
				result.Lo = numeric.I64Extend8S(a.Lo)
			case 0xc3:
				result.Lo = numeric.I64Extend16S(a.Lo)
			case 0xc4:
				result.Lo = numeric.I64Extend32S(a.Lo)
			case 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60, 0x61, 0x62,
				0x63, 0x64, 0x65, 0x66, 0x6d, 0x6e, 0x6f, 0x70,
				0x7f, 0x80, 0x81, 0x82, 0x8b, 0x8c, 0x8d, 0x8e,
				0x8f, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96,
				0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e,
				0x9f, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6,
				0xa8, 0xa9, 0xaa, 0xab, 0xae, 0xaf, 0xb0, 0xb1,
				0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9,
				0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf, 0xfc0000, 0xfc0001,
				0xfc0002, 0xfc0003, 0xfc0004, 0xfc0005, 0xfc0006, 0xfc0007:
				result.Lo, trap = numeric.Eval(op, a.Lo, b.Lo)
			case 0xd3:
				if a.Lo == b.Lo {
					result.Lo = 1
				}
			case 0xd0:
			case 0xd4:
				if a.Lo == 0 {
					trap = wasm.TrapNullReference
				} else {
					result = a
				}
			case 0xd1:
				if a.Lo == 0 {
					result.Lo = 1
				}
			case 0xd2:
				result.Lo = instance.Functions[im.A]
			case 0x10, 0x11, 0x12, 0x13, 0x14, 0x15:
				tail := module.IsTailCall(op)
				count := in.Count
				inputs := code.Args[in.Extra : in.Extra+count]
				var target uint64
				if op == 0x10 || op == 0x12 {
					target = instance.Functions[im.A]
				} else if op == 0x14 || op == 0x15 {
					count--
					target = values[int(inputs[count])].Lo
					if target == 0 {
						trap = wasm.TrapNullReference
						break
					}
				} else {
					count--
					table := instance.Tables[im.B]
					index := table.Type.Limits.Address(values[int(inputs[count])].Lo)
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
				if (op == 0x11 || op == 0x13 || op == 0x14 || op == 0x15) && !e.store.FunctionSubtype(callee.TypeID, instance.Types[im.A]) {
					trap = wasm.TrapIndirectCallType
					break
				}
				var dest []slots.Slot
				if tail {
					dest = f.dest
				} else {
					dest = code.Args[in.Extra+in.Count : in.Extra+in.Count+uint32(len(callee.Signature.Results))]
				}
				if err := e.poll(); err != nil {
					return e.fail(err)
				}
				if tail {
					// Preserve the tuple before retiring/reusing the current frame.
					// This supports overlapping/permuted arguments and a different
					// callee signature without charging an additional guest frame.
					if callee.Code != nil && uint64(base)+uint64(len(callee.Code.Slots)) > uint64(e.config.Limits.MaxSlots) {
						return e.fail(&wasm.Error{Kind: wasm.ErrLimit, Detail: "guest frame slots"})
					}
					if int(count) > cap(e.tailArgs) {
						e.tailArgs = make([]value.Value, count)
					} else {
						e.tailArgs = e.tailArgs[:count]
					}
					for i, src := range inputs[:count] {
						e.tailArgs[i] = values[int(src)]
					}
					e.values = e.values[:base]
					e.frames[fi] = frame{}
					e.frames = e.frames[:fi]
					if callee.Code == nil {
						e.pending = &HostCall{Function: callee, Caller: instance, Args: append([]value.Value(nil), e.tailArgs...)}
						e.pendingDest, e.status = dest, HostRequest
						clear(e.tailArgs)
						e.tailArgs = e.tailArgs[:0]
						return HostRequest
					}
					if err := e.push(callee, dest); err != nil {
						return e.fail(err)
					}
					copy(e.values[base:], e.tailArgs)
					clear(e.tailArgs)
					e.tailArgs = e.tailArgs[:0]
					continue frames
				}
				// Publish the caller continuation before a host suspension or
				// frame push. Straight-line execution keeps its PC local.
				f.pc = pc + 1
				f.callPC, f.suspended = pc, true
				if callee.Code == nil {
					args := make([]value.Value, count)
					for i := range args {
						args[i] = e.values[base+int(inputs[i])]
					}
					e.pending = &HostCall{Function: callee, Caller: instance, Args: args}
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
				continue frames
			default:
				switch {
				case op>>16 == 0xfd:
					sub := op & 0xffff
					switch {
					case sub <= 10 || sub >= 84 && sub <= 87 || sub == 92 || sub == 93:
						result, trap = instance.Memories[im.Memory].Load(op, a.Lo, im.MemoryOffset(), im.Lane, b)
					case sub == 11 || sub >= 88 && sub <= 91:
						trap = instance.Memories[im.Memory].Store(op, a.Lo, im.MemoryOffset(), im.Lane, b)
					default:
						result = simd.Eval(im, a, b, c)
					}
				case op >= 0xfc0008 && op <= 0xfc0011:
					var err error
					result, trap, err = e.bulk(instance, *im, a, b, c)
					if err != nil {
						return e.fail(err)
					}
				default:
					panic("interp: unknown validated opcode")
				}
			}
			if trap != wasm.NoTrap {
				return e.fail(&wasm.Trap{Code: trap, Function: fn.Index, Offset: im.Offset})
			}
			if in.Dst != slots.NoSlot {
				values[int(in.Dst)] = result
			}
			pc++
		}
	}
	panic("interp: running without frames")
}

// RetainedValues covers reference slots in active or suspended frames. Slot
// types remain fixed even when temporaries are reused, including edge scratch.
func (e *Execution) RetainedValues(visit func(wasm.ValueType, value.Value)) {
	if ex, ok := e.store.ExceptionError(e.err); ok {
		visit(wasm.ExnRef.NonNull(), value.Value{Lo: ex.Ref})
	}
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

// RetainedInstances preserves the owners of running and suspended function
// bodies even if their last table/global reference was removed during a call.
func (e *Execution) RetainedInstances(visit func(*state.Instance)) {
	for _, f := range e.frames {
		visit(f.function.Instance)
	}
	if e.pending != nil && e.pending.Caller != nil && (len(e.frames) == 0 || e.frames[len(e.frames)-1].function.Instance != e.pending.Caller) {
		visit(e.pending.Caller)
	}
}

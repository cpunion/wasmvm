// Package execute runs immutable slot code with explicit guest frames and
// host-request exits. Guest calls never recurse on the host Go stack.
package execute

import (
	"context"
	"errors"
	"math/bits"
	"sync/atomic"

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
	// Optional backend state belongs to this execution, never to a captured
	// Wasm continuation. A stale token always selects the ordinary entry path.
	nativeResume uint64
}
type HostCall struct {
	Function *state.Function
	Caller   *state.Instance
	Args     []value.Value
}
type Execution struct {
	store        *state.Store
	context      context.Context
	config       Config
	frames       []frame
	prompts      []prompt
	transferring *continuationBody
	values       []value.Value
	result       []value.Value
	tailArgs     []value.Value
	gcArgs       []value.Value // Synchronous scratch; cleared after each checked operation.
	hostRequest  HostCall
	hostArgs     []value.Value
	pending      *HostCall
	pendingDest  []slots.Slot
	status       Status
	err          error
	fuel, steps  uint64
	cancelFlag   *atomic.Bool
	nativeFrames state.NativeFrames
	source       ir.Source
	nativeResume uint64
}

func New(store *state.Store, ctx context.Context, config Config) *Execution {
	return &Execution{store: store, context: contextOrBackground(ctx), config: config, fuel: config.Fuel, status: Done}
}
func (e *Execution) Fuel() uint64 { return e.fuel }

// SetFuel synchronizes a suspended continuation with its driver's shared
// reentry budget. The driver never changes it while Run is executing.
func (e *Execution) SetFuel(fuel uint64) { e.fuel = fuel }
func (e *Execution) Footprint() Limits   { return Limits{uint32(len(e.frames)), uint32(len(e.values))} }
func (e *Execution) Err() error          { return e.err }

// Results are borrowed until the next Start. Host calls expose copied input
// values borrowed until ResumeHost, Abort or Quiesce. They remain stable while
// the driver services a suspended request, including nested host reentry.
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
		copy(e.prepareHost(fn, nil, len(args)), args)
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
	// Reuse frame capacity directly; the pinned LLGo scalar append otherwise
	// allocates temporary argument storage on every guest call.
	i := len(e.frames)
	if i == cap(e.frames) {
		e.frames = append(e.frames, frame{})
	} else {
		e.frames = e.frames[:i+1]
	}
	e.frames[i] = frame{function: fn, base: base, dest: dest}
	return nil
}

// Large valid tuples remain transient; idle executions retain at most 1 KiB.
const maxCachedGCArgs = 64

func (e *Execution) clearGCArgs() {
	clear(e.gcArgs)
	if cap(e.gcArgs) > maxCachedGCArgs {
		e.gcArgs = nil
	} else {
		e.gcArgs = e.gcArgs[:0]
	}
}

func (e *Execution) fail(err error) Status {
	e.err, e.status = err, Failed
	e.nativeFrames = state.NativeFrames{}
	e.nativeResume = 0
	e.source = ir.Source{}
	clear(e.frames)
	e.frames = e.frames[:0]
	clear(e.prompts)
	e.prompts = e.prompts[:0]
	e.transferring = nil
	e.values = e.values[:0]
	e.clearHost()
	e.pendingDest = nil
	clear(e.tailArgs)
	e.tailArgs = e.tailArgs[:0]
	e.clearGCArgs()
	return Failed
}

// Abort releases every active frame and pending host request. Host drivers use
// it in deferred panic cleanup before propagating the original host panic.
func (e *Execution) Abort(err error) { e.fail(err) }

// SetCancellationFlag binds an owner-controlled signal while the execution is
// idle. A false flag guarantees that the current context has no error. A true
// flag still calls Err, preserving the context's cancellation publication.
func (e *Execution) SetCancellationFlag(flag *atomic.Bool) { e.cancelFlag = flag }

func (e *Execution) poll() error {
	if e.cancelFlag != nil && !e.cancelFlag.Load() {
		return nil
	}
	return e.context.Err()
}

func (e *Execution) ResumeHost(results []value.Value, err error) error {
	if e.status != HostRequest || e.pending == nil {
		return errors.New("no pending host request")
	}

	if err != nil {
		e.nativeResume = 0
		if ex, ok := e.store.ExceptionError(err); ok {
			e.clearHost()
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
	e.clearHost()
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
	if backend == nil {
		return e.runSlots()
	}
	for {
		f := &e.frames[len(e.frames)-1]
		f.suspended = false
		if status, handled := e.runVeneer(backend); handled {
			e.nativeResume = 0
			if status != Running {
				return status
			}
			continue
		}
		pc, err := e.runBackend(backend, f.function, f.base, f.pc)
		if err != nil {
			return e.fail(err)
		}
		// Transfer validation can replace frames or relocate their values.
		// Publish the checked boundary before selecting the shared slot view.
		f = &e.frames[len(e.frames)-1]
		f.pc = pc
		e.steps++
		if e.steps&63 == 0 {
			if err := e.poll(); err != nil {
				return e.fail(err)
			}
		}
		// Select shared leaves here so a common dispatcher does not save
		// and restore another frame before each native runtime effect.
		op := f.function.Code.Code[pc].Op
		var status Status
		switch {
		case op>>16 == 0xfb:
			status = e.runGC()
		case op>>16 == 0xfd:
			status = e.runSIMD()
		case op == slots.ControlBase+uint32(ir.Return):
			status = e.returnFrame()
		case op >= 0x10 && op <= 0x15:
			status = e.runCall()
		case op == 8 || op == 10:
			status = e.runThrow()
		default:
			status = e.runRuntime()
		}
		if status != HostRequest || op != 0x10 && op != 0x11 && op != 0x14 {
			e.nativeResume = 0
		}
		if status != Running {
			return status
		}
	}
}

// runSlots keeps portable scalar/local/control execution in a frame-local loop.
// Runtime effects share their handlers with validated native exits.
func (e *Execution) runSlots() Status {
	status, steps := e.runSlotsLoop(e.steps)
	e.steps = steps
	return status
}

// Keep the portable polling counter local between yields. Each status publishes
// the exact count without a captured closure or a store on every instruction.
func (e *Execution) runSlotsLoop(steps uint64) (Status, uint64) {
	metered := e.config.Metered
	for len(e.frames) > 0 {
		fi := len(e.frames) - 1
		f := &e.frames[fi]
		f.suspended = false
		fn := f.function
		code := fn.Code
		instructions := code.Code
		base := f.base
		instance := fn.Instance
		pc := f.pc
		values := e.values[base : base+len(code.Slots)]
		for {
			// Instructions are immutable for the frame's lifetime. Read only the
			// fields needed by the selected handler instead of copying the record.
			in := &instructions[pc]
			if metered {
				source := &code.Sources[in.Source]
				cost := source.Fuel()
				if cost > e.fuel {
					return e.fail(&wasm.Error{Kind: wasm.ErrLimit, Offset: source.Wasm.Offset, Detail: "execution fuel"}), steps
				}
				e.fuel -= cost
			}
			steps++
			if steps&63 == 0 {
				if err := e.poll(); err != nil {
					return e.fail(err), steps
				}
			}
			for in.Count != 0 && (in.Op == 0x20 || in.Op == 0x41) && !metered && steps&63 != 63 {
				// Coalesced local reads already occupy their result slot.
				// Commit constants before advancing, retaining every source
				// step and stopping before the next polling boundary.
				if in.Op == 0x41 {
					values[int(in.Dst)] = value.Value{Lo: uint64(in.Extra)}
				}
				pc++
				in = &instructions[pc]
				steps++
			}
			var result value.Value
			trap := wasm.NoTrap
			// Internal control instructions have their own dispatch cases below.
			// Classify only Wasm opcodes through the extension prefixes.
			if in.Op >= 0xe0 && in.Op < slots.ControlBase {
				switch in.Op >> 16 {
				case 0xfd:
					var a, b, c value.Value
					if in.A != slots.NoSlot {
						a = values[int(in.A)]
					}
					if in.B != slots.NoSlot {
						b = values[int(in.B)]
					}
					op := in.Op
					if in.C != slots.NoSlot {
						c = values[int(in.C)]
					}
					im := &code.Sources[in.Source].Wasm
					sub := op & 0xffff
					switch {
					case sub <= 10 || sub >= 84 && sub <= 87 || sub == 92 || sub == 93:
						result, trap = instance.Memories[im.Memory].Load(op, a.Lo, im.MemoryOffset(), im.Lane, b)
					case sub == 11 || sub >= 88 && sub <= 91:
						trap = instance.Memories[im.Memory].Store(op, a.Lo, im.MemoryOffset(), im.Lane, b)
					default:
						result = simd.Eval(im, a, b, c)
					}
					goto commit
				case 0xfb:
					// Field/array access is a canonical state primitive. Keep
					// its operand routing in this loop; complex GC effects use
					// the same shared leaf as native execution.
					if n := state.GCAccessArity(in.Op); n != 0 {
						var a, b, c value.Value
						if in.A != slots.NoSlot {
							a = values[int(in.A)]
						}
						if in.B != slots.NoSlot {
							b = values[int(in.B)]
						}
						if n > 2 {
							c = values[int(code.Args[in.Extra+2])]
						}
						var err error
						result, err = e.store.GCAccess(instance, &code.Sources[in.Source].Wasm, a, b, c)
						if err != nil {
							var t *wasm.Trap
							if errors.As(err, &t) {
								t.Function = fn.Index
							}
							return e.fail(err), steps
						}
						if in.Dst != slots.NoSlot {
							values[int(in.Dst)] = result
						}
					} else {
						f.pc = pc
						if status := e.runGC(); status != Running {
							return status, steps
						}
					}
					pc++
					f.pc = pc
					goto next
				case 0xfc:
					if module.IsWideArithmetic(in.Op) {
						args := code.Args[in.Extra : in.Extra+in.Count+2]
						left, right := values[args[0]].Lo, values[args[1]].Lo
						var lo, hi uint64
						switch in.Op {
						case 0xfc0013, 0xfc0014:
							otherLo, otherHi := values[args[2]].Lo, values[args[3]].Lo
							var carry uint64
							if in.Op == 0xfc0013 {
								lo, carry = bits.Add64(left, otherLo, 0)
								hi, _ = bits.Add64(right, otherHi, carry)
							} else {
								lo, carry = bits.Sub64(left, otherLo, 0)
								hi, _ = bits.Sub64(right, otherHi, carry)
							}
						default:
							hi, lo = bits.Mul64(left, right)
							if in.Op == 0xfc0015 {
								if int64(left) < 0 {
									hi -= right
								}
								if int64(right) < 0 {
									hi -= left
								}
							}
						}
						values[args[in.Count]], values[args[in.Count+1]] = value.Value{Lo: lo}, value.Value{Lo: hi}
						pc++
						f.pc = pc
						goto next
					}
				default:
					if in.Op <= 0xe6 {
						f.pc = pc
						if status := e.runRuntime(); status != Running {
							return status, steps
						}
						goto transferred
					}
				}
			}
			// Decode scalar operands in their semantic handler. This keeps
			// optional-input guards and slot-index/value live ranges out of
			// unrelated opcodes; complete-value copies still read both words.
			// Scalar numeric opcodes form a contiguous range. Dispatch them
			// separately from sparse control, reference and host operations.
			if op := in.Op; op >= 0x45 && op <= 0xc4 {
				switch op {
				case 0x45:
					result.Lo = numeric.I32Eqz(values[in.A].Lo)
				case 0x46:
					result.Lo = numeric.I32Eq(values[in.A].Lo, values[in.B].Lo)
				case 0x47:
					result.Lo = numeric.I32Ne(values[in.A].Lo, values[in.B].Lo)
				case 0x48:
					result.Lo = numeric.I32LtS(values[in.A].Lo, values[in.B].Lo)
				case 0x49:
					result.Lo = numeric.I32LtU(values[in.A].Lo, values[in.B].Lo)
				case 0x4a:
					result.Lo = numeric.I32GtS(values[in.A].Lo, values[in.B].Lo)
				case 0x4b:
					result.Lo = numeric.I32GtU(values[in.A].Lo, values[in.B].Lo)
				case 0x4c:
					result.Lo = numeric.I32LeS(values[in.A].Lo, values[in.B].Lo)
				case 0x4d:
					result.Lo = numeric.I32LeU(values[in.A].Lo, values[in.B].Lo)
				case 0x4e:
					result.Lo = numeric.I32GeS(values[in.A].Lo, values[in.B].Lo)
				case 0x4f:
					result.Lo = numeric.I32GeU(values[in.A].Lo, values[in.B].Lo)
				case 0x50:
					result.Lo = numeric.I64Eqz(values[in.A].Lo)
				case 0x51:
					result.Lo = numeric.I64Eq(values[in.A].Lo, values[in.B].Lo)
				case 0x52:
					result.Lo = numeric.I64Ne(values[in.A].Lo, values[in.B].Lo)
				case 0x53:
					result.Lo = numeric.I64LtS(values[in.A].Lo, values[in.B].Lo)
				case 0x54:
					result.Lo = numeric.I64LtU(values[in.A].Lo, values[in.B].Lo)
				case 0x55:
					result.Lo = numeric.I64GtS(values[in.A].Lo, values[in.B].Lo)
				case 0x56:
					result.Lo = numeric.I64GtU(values[in.A].Lo, values[in.B].Lo)
				case 0x57:
					result.Lo = numeric.I64LeS(values[in.A].Lo, values[in.B].Lo)
				case 0x58:
					result.Lo = numeric.I64LeU(values[in.A].Lo, values[in.B].Lo)
				case 0x59:
					result.Lo = numeric.I64GeS(values[in.A].Lo, values[in.B].Lo)
				case 0x5a:
					result.Lo = numeric.I64GeU(values[in.A].Lo, values[in.B].Lo)
				case 0x67:
					result.Lo = numeric.I32Clz(values[in.A].Lo)
				case 0x68:
					result.Lo = numeric.I32Ctz(values[in.A].Lo)
				case 0x69, 0x7b:
					// Share one population-count intrinsic so its CPU feature
					// check stays within this handler instead of other hot cases.
					x := values[in.A].Lo
					if op == 0x69 {
						x = uint64(uint32(x))
					}
					result.Lo = numeric.I64Popcnt(x)
				case 0x6a:
					result.Lo = numeric.I32Add(values[in.A].Lo, values[in.B].Lo)
				case 0x6b:
					result.Lo = numeric.I32Sub(values[in.A].Lo, values[in.B].Lo)
				case 0x6c:
					result.Lo = numeric.I32Mul(values[in.A].Lo, values[in.B].Lo)
				case 0x71:
					result.Lo = numeric.I32And(values[in.A].Lo, values[in.B].Lo)
				case 0x72:
					result.Lo = numeric.I32Or(values[in.A].Lo, values[in.B].Lo)
				case 0x73:
					result.Lo = numeric.I32Xor(values[in.A].Lo, values[in.B].Lo)
				case 0x74:
					result.Lo = numeric.I32Shl(values[in.A].Lo, values[in.B].Lo)
				case 0x75:
					result.Lo = numeric.I32ShrS(values[in.A].Lo, values[in.B].Lo)
				case 0x76:
					result.Lo = numeric.I32ShrU(values[in.A].Lo, values[in.B].Lo)
				case 0x77:
					result.Lo = numeric.I32Rotl(values[in.A].Lo, values[in.B].Lo)
				case 0x78:
					result.Lo = numeric.I32Rotr(values[in.A].Lo, values[in.B].Lo)
				case 0x79:
					result.Lo = numeric.I64Clz(values[in.A].Lo)
				case 0x7a:
					result.Lo = numeric.I64Ctz(values[in.A].Lo)
				case 0x7c:
					result.Lo = numeric.I64Add(values[in.A].Lo, values[in.B].Lo)
				case 0x7d:
					result.Lo = numeric.I64Sub(values[in.A].Lo, values[in.B].Lo)
				case 0x7e:
					result.Lo = numeric.I64Mul(values[in.A].Lo, values[in.B].Lo)
				case 0x83:
					result.Lo = numeric.I64And(values[in.A].Lo, values[in.B].Lo)
				case 0x84:
					result.Lo = numeric.I64Or(values[in.A].Lo, values[in.B].Lo)
				case 0x85:
					result.Lo = numeric.I64Xor(values[in.A].Lo, values[in.B].Lo)
				case 0x86:
					result.Lo = numeric.I64Shl(values[in.A].Lo, values[in.B].Lo)
				case 0x87:
					result.Lo = numeric.I64ShrS(values[in.A].Lo, values[in.B].Lo)
				case 0x88:
					result.Lo = numeric.I64ShrU(values[in.A].Lo, values[in.B].Lo)
				case 0x89:
					result.Lo = numeric.I64Rotl(values[in.A].Lo, values[in.B].Lo)
				case 0x8a:
					result.Lo = numeric.I64Rotr(values[in.A].Lo, values[in.B].Lo)
				case 0xa7:
					result.Lo = numeric.I32WrapI64(values[in.A].Lo)
				case 0xac:
					result.Lo = numeric.I64ExtendI32S(values[in.A].Lo)
				case 0xad:
					result.Lo = numeric.I64ExtendI32U(values[in.A].Lo)
				case 0xc0:
					result.Lo = numeric.I32Extend8S(values[in.A].Lo)
				case 0xc1:
					result.Lo = numeric.I32Extend16S(values[in.A].Lo)
				case 0xc2:
					result.Lo = numeric.I64Extend8S(values[in.A].Lo)
				case 0xc3:
					result.Lo = numeric.I64Extend16S(values[in.A].Lo)
				case 0xc4:
					result.Lo = numeric.I64Extend32S(values[in.A].Lo)
				case 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60, 0x61, 0x62,
					0x63, 0x64, 0x65, 0x66, 0x6d, 0x6e, 0x6f, 0x70,
					0x7f, 0x80, 0x81, 0x82, 0x8b, 0x8c, 0x8d, 0x8e,
					0x8f, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96,
					0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e,
					0x9f, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6,
					0xa8, 0xa9, 0xaa, 0xab, 0xae, 0xaf, 0xb0, 0xb1,
					0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9,
					0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf:
					right := uint64(0)
					if in.B != slots.NoSlot {
						right = values[in.B].Lo
					}
					result.Lo, trap = numeric.Eval(op, values[in.A].Lo, right)
					if trap != wasm.NoTrap {
						goto commit
					}
				}
				goto write
			}
			switch op := in.Op; op {
			case slots.ControlBase + uint32(ir.Return):
				f.pc = pc
				if status := e.returnFrame(); status != Running {
					return status, steps
				}
				goto transferred
			case slots.ControlBase + uint32(ir.IfCast):
				f.pc = pc
				if status := e.runRuntime(); status != Running {
					return status, steps
				}
				goto transferred
			case slots.ControlBase + uint32(ir.IfNull), slots.ControlBase + uint32(ir.Jump), slots.ControlBase + uint32(ir.If), slots.ControlBase + uint32(ir.Table), slots.ControlBase + uint32(ir.Trap):
				kind := ir.TermKind(op - slots.ControlBase)
				if kind == ir.Trap {
					trap = wasm.TrapUnreachable
					break
				}
				index := uint32(0)
				if kind == ir.IfNull && values[in.A].Lo != 0 {
					index = 1
				} else if kind == ir.If && uint32(values[in.A].Lo) == 0 {
					index = 1
				} else if kind == ir.Table {
					index = uint32(values[in.A].Lo)
					if index >= in.Count-1 {
						index = in.Count - 1
					}
				}
				edge := code.Edges[in.Extra+index]
				if edge.PC <= pc {
					if err := e.poll(); err != nil {
						return e.fail(err), steps
					}
				}
				for _, move := range code.Moves[edge.Start : edge.Start+edge.Count] {
					values[int(move.Dst)] = values[int(move.Src)]
				}
				pc = edge.PC
				goto next

			case 8, 10:
				f.pc = pc
				if status := e.runThrow(); status != Running {
					return status, steps
				}
				goto transferred
			case 1, 0x1a:
			case 0x1b, 0x1c:
				selected := in.B
				if uint32(values[in.C].Lo) != 0 {
					selected = in.A
				}
				result = values[selected]
			case 0x20:
				result = values[int(in.Extra)]
			case 0x21, 0x22:
				result = values[int(in.A)]
				values[int(in.Extra)] = result
			case 0x23:
				result = instance.Globals[in.Extra].Value
			case 0x24:
				instance.Globals[in.Extra].Value = values[int(in.A)]
			case 0x25:
				table := instance.Tables[in.Extra]
				index := table.Type.Limits.Address(values[in.A].Lo)
				if index >= uint64(len(table.Values)) {
					trap = wasm.TrapTableBounds
				} else {
					result.Lo = table.Values[index]
				}
			case 0x26:
				table := instance.Tables[in.Extra]
				index := table.Type.Limits.Address(values[in.A].Lo)
				if index >= uint64(len(table.Values)) {
					trap = wasm.TrapTableBounds
				} else {
					table.Values[index] = values[in.B].Lo
				}
			case 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
				0x30, 0x31, 0x32, 0x33, 0x34, 0x35:
				im := &code.Sources[in.Source].Wasm
				result, trap = instance.Memories[im.Memory].Load(op, values[in.A].Lo, im.MemoryOffset(), 0, value.Value{})
			case 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
				0x3e:
				im := &code.Sources[in.Source].Wasm
				trap = instance.Memories[im.Memory].Store(op, values[in.A].Lo, im.MemoryOffset(), 0, value.Value{Lo: values[in.B].Lo})
			case 0x3f:
				result.Lo = uint64(instance.Memories[in.Extra].Pages())
			case 0x40:
				mem := instance.Memories[in.Extra]
				old, ok := e.store.GrowMemory(mem, mem.Limits().Address(values[in.A].Lo))
				result.Lo = uint64(old)
				if !ok {
					result.Lo = mem.Limits().Failure()
				}
			case 0x41:
				result.Lo = uint64(in.Extra)

			case 0x42, 0x43, 0x44:
				result.Lo = uint64(in.Extra) | uint64(in.Count)<<32
			case 0xfc0000, 0xfc0001, 0xfc0002, 0xfc0003,
				0xfc0004, 0xfc0005, 0xfc0006, 0xfc0007:
				result.Lo, trap = numeric.Eval(op, values[in.A].Lo, 0)
			case 0xd3:
				if values[in.A].Lo == values[in.B].Lo {
					result.Lo = 1
				}
			case 0xd0:
			case 0xd4:
				if values[in.A].Lo == 0 {
					trap = wasm.TrapNullReference
				} else {
					result = values[int(in.A)]
				}
			case 0xd1:
				if values[in.A].Lo == 0 {
					result.Lo = 1
				}
			case 0xd2:
				result.Lo = instance.Functions[in.Extra]
			case 0x10, 0x11, 0x12, 0x13, 0x14, 0x15:
				f.pc = pc
				if status := e.runCall(); status != Running {
					return status, steps
				}
				goto transferred
			default:
				var a, b, c uint64
				if in.A != slots.NoSlot {
					a = values[in.A].Lo
				}
				if in.B != slots.NoSlot {
					b = values[in.B].Lo
				}
				if in.C != slots.NoSlot {
					c = values[int(in.C)].Lo
				}
				switch {

				case op>>16 == 0xfe:
					if op&0xffff == 3 {
						e.store.Domain.Fence()
						break
					}
					im := &code.Sources[in.Source].Wasm
					var err error
					result, trap, err = instance.Memories[im.Memory].Atomic(e.context, e.store.Domain, op&0xffff, a, im.MemoryOffset(), b, c)
					if err != nil {
						return e.fail(err), steps
					}
				case op >= 0xfc0008 && op <= 0xfc0011:
					im := &code.Sources[in.Source].Wasm
					var err error
					result, trap, err = e.bulk(instance, *im, value.Value{Lo: a}, value.Value{Lo: b}, value.Value{Lo: c})
					if err != nil {
						return e.fail(err), steps
					}
				default:
					panic("interp: unknown validated opcode")
				}
			}
		commit:
			if trap != wasm.NoTrap {
				return e.fail(&wasm.Trap{Code: trap, Function: fn.Index, Offset: code.Sources[in.Source].Wasm.Offset}), steps
			}
		write:
			if in.Dst != slots.NoSlot {
				values[int(in.Dst)] = result
			}
			pc++
		next:
		}
	transferred:
	}
	panic("interp: running without frames")
}

// RetainedValues covers reference slots in active or suspended frames. Slot
// types remain fixed even when temporaries are reused, including edge scratch.
func (e *Execution) RetainedValues(visit func(wasm.ValueType, value.Value)) {
	if suspension, ok := e.err.(*state.Suspension); ok {
		for i, typ := range suspension.Tag.Signature.Params {
			if typ.IsReference() {
				visit(typ, suspension.Args[i])
			}
		}
	}
	if e.transferring != nil {
		e.transferring.Visit(visit, func(*state.Instance) {})
	}
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
	if suspension, ok := e.err.(*state.Suspension); ok {
		visit(suspension.Tag.OwnerInstance)
	}
	if e.transferring != nil {
		e.transferring.Visit(func(wasm.ValueType, value.Value) {}, visit)
	}
	for _, f := range e.frames {
		visit(f.function.Instance)
	}
	if e.pending != nil && e.pending.Caller != nil && (len(e.frames) == 0 || e.frames[len(e.frames)-1].function.Instance != e.pending.Caller) {
		visit(e.pending.Caller)
	}
}

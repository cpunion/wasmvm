// Package slots lowers typed SSA to immutable interpreter instructions and
// reusable typed frame slots. Guest execution is implemented separately.
package slots

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Slot uint32

const NoSlot Slot = math.MaxUint32

// ControlBase separates internal control instructions from Wasm opcodes.
const ControlBase uint32 = 1 << 24

// SharedMemoryOp is a compiler-private checked helper boundary; its original
// scalar/SIMD operation remains in Source.Wasm.
const SharedMemoryOp uint32 = 0xff0000

type Instruction struct {
	Op, Source   uint32
	Dst, A, B, C Slot
	Extra, Count uint32
}

// Calls store input slots followed by output slots in Args[Extra:], with Count
// inputs and Source.Outputs outputs, as do GC and wide arithmetic instructions.
// Tail calls store only inputs, with no local result slots. Returns store Count
// slots at Extra. Other operations have at most three inline operands and one
// inline result.
// I64/F64 constants store their low/high bits in Extra/Count. I32 uses Extra;
// its Count is a pure following-integer-operation hint (zero or one).
// Coalesced local.get uses Count as a zero/one no-copy hint. Integer comparisons
// use Count as a zero/one hint for an adjacent branch. Local, global,
// table, memory-size/grow and ref.func indices use Extra without a tuple.
// Branches address Count edges starting at Extra; A is the condition/index.
// Every instruction retains its original source and fuel boundary.
type Edge struct{ PC, Start, Count uint32 }
type Move struct{ Dst, Src Slot }

// Handler and ProtectedBlock are cold-path unwind metadata. No per-step handler
// stack is needed, and native code resumes at the same checked target PCs.
type Handler struct {
	Parent  uint32
	Catches []Catch
}
type Catch struct {
	Kind    byte
	Tag, PC uint32
	Type    uint32 // Module-local continuation type for a suspension handler.
	Dest    []Slot
}
type ProtectedBlock struct{ Start, End, Handler uint32 }
type Function struct {
	image      *sourceImage
	Veneer     *CallVeneer
	Handlers   []Handler
	Prompts    map[uint32][]Catch
	Protected  []ProtectedBlock
	Signature  module.FuncType
	Sources    []ir.Source
	Slots      []wasm.ValueType
	LocalCount uint32
	Code       []Instruction
	Args       []Slot
	Edges      []Edge
	Moves      []Move
	// ValueSlots and BlockPC retain the lowering map for diagnostics/verifiers.
	ValueSlots []Slot
	BlockPC    []uint32
}
type Module struct {
	IR        *lower.Module
	Functions []*Function
}
type Limits struct {
	MaxSlots                uint32
	MaxNodes, MaxVerifyWork uint64
}

func DefaultLimits() Limits { return Limits{1 << 20, 16 << 20, 100 << 20} }

type budget struct {
	nodes uint64
	slots uint32
}

func (b *budget) charge(n uint64) {
	if n > b.nodes {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Detail: "slot lowering nodes"})
	}
	b.nodes -= n
}
func (b *budget) addSlot(f *Function, t wasm.ValueType) Slot {
	if uint64(len(f.Slots)) >= uint64(b.slots) {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Detail: "function frame slots"})
	}
	b.charge(1)
	s := Slot(len(f.Slots))
	f.Slots = append(f.Slots, t.Storage())
	return s
}

// Compile uses a per-function frame limit and aggregate node/verification
// budgets. It returns no partial artifact. Its input must remain immutable.
func Compile(m *lower.Module, limits Limits) (result *Module, err error) {
	defer func() {
		if p := recover(); p != nil {
			if e, ok := p.(*wasm.Error); ok {
				result, err = nil, e
			} else {
				panic(p)
			}
		}
	}()
	if m == nil {
		return nil, &wasm.Error{Kind: wasm.ErrInvalid, Detail: "nil IR module"}
	}
	b := budget{limits.MaxNodes, limits.MaxSlots}
	verify := ir.VerifyBudget{Remaining: limits.MaxVerifyWork}
	b.charge(uint64(len(m.Functions)))
	result = &Module{IR: m, Functions: make([]*Function, len(m.Functions))}
	var compiler slotCompiler
	for i, f := range m.Functions {
		if err := verify.Verify(f); err != nil {
			return nil, err
		}
		result.Functions[i] = compiler.compileFunction(f, &b)
	}
	return result, nil
}

type slotCompiler struct {
	defs       []int
	pinned     []bool
	uses       []uint32
	lastUse    []uint32
	coalesced  []bool
	writeBlock []uint32
	nextWrite  []uint32
	returns    []bool
	free, all  [10][]Slot
	literals   literalPool
	moves      []Move
	planner    movePlanner
	scratch    map[wasm.ValueType]Slot
}

func storageBucket(t wasm.ValueType) int {
	switch t.Storage() {
	case wasm.I32:
		return 0
	case wasm.I64:
		return 1
	case wasm.F32:
		return 2
	case wasm.F64:
		return 3
	case wasm.V128:
		return 4
	case wasm.FuncRef:
		return 5
	case wasm.ExternRef:
		return 6
	case wasm.ExnRef:
		return 7
	case wasm.ContRef:
		return 8
	default:
		return 9
	}
}

func reuseSlice[T any](buf []T, n int) []T {
	if cap(buf) < n {
		return make([]T, n)
	}
	return buf[:n]
}

func (c *slotCompiler) compileFunction(f *ir.Function, budget *budget) *Function {
	n := len(f.Values)
	nBlocks := len(f.Blocks)
	budget.charge(uint64(n)*4 + uint64(nBlocks))
	slotCap := uint64(len(f.Locals))
	if n > 0 {
		slotCap += min(uint64(n), 8)
	}
	if uint64(budget.slots) < slotCap {
		slotCap = uint64(budget.slots)
	}
	if budget.nodes < slotCap {
		slotCap = budget.nodes
	}
	var initSlots []wasm.ValueType
	if slotCap > 0 {
		initSlots = make([]wasm.ValueType, 0, int(slotCap))
	}
	out := &Function{
		Signature:  f.Signature,
		Sources:    f.Sources,
		LocalCount: uint32(len(f.Locals)),
		Slots:      initSlots,
		ValueSlots: make([]Slot, n),
		BlockPC:    make([]uint32, nBlocks),
	}
	for _, t := range f.Locals {
		budget.addSlot(out, t)
	}
	defs := reuseSlice(c.defs, n)
	clear(defs)
	pinned := reuseSlice(c.pinned, n)
	clear(pinned)
	uses := reuseSlice(c.uses, n)
	clear(uses)
	budget.charge(uint64(n))
	lastUse := reuseSlice(c.lastUse, n)
	clear(lastUse)
	c.defs, c.pinned, c.uses, c.lastUse = defs, pinned, uses, lastUse
	var totalOps uint64
	for bi, b := range f.Blocks {
		ops := uint64(len(b.Ops)) + 1
		budget.charge(ops)
		totalOps += ops
		for _, v := range b.Params {
			defs[v] = bi
			pinned[v] = true
		}
		for _, op := range b.Ops {
			for _, v := range op.Results {
				defs[v] = bi
			}
		}
	}
	for bi, b := range f.Blocks {
		for oi, op := range b.Ops {
			for _, v := range op.Args {
				uses[v]++
				if defs[v] != bi {
					pinned[v] = true
				} else {
					lastUse[v] = uint32(oi)
				}
			}
		}
		endPos := uint32(len(b.Ops))
		for _, v := range b.Term.Args {
			uses[v]++
			if defs[v] != bi {
				pinned[v] = true
			} else {
				lastUse[v] = endPos
			}
		}
		for _, e := range b.Term.Edges {
			for _, v := range e.Args {
				uses[v]++
				if defs[v] != bi {
					pinned[v] = true
				} else {
					lastUse[v] = endPos
				}
			}
		}
	}
	coalesced := c.coalesceLocals(f, out, pinned, lastUse, budget)
	for v, t := range f.Values {
		if pinned[v] {
			out.ValueSlots[v] = budget.addSlot(out, t)
		}
	}
	if codeCap := min(totalOps, budget.nodes); codeCap > 0 {
		out.Code = make([]Instruction, 0, int(codeCap))
	}
	// Only block-local temporaries are recycled in this first allocator. Values
	// crossing blocks and block parameters keep dedicated slots; this avoids a
	// quadratic global interference graph and preserves loops conservatively.
	for i := range c.all {
		c.all[i] = c.all[i][:0]
		c.free[i] = c.free[i][:0]
	}
	free, all := &c.free, &c.all
	literals := &c.literals
	release := func(v ir.Value) {
		if !pinned[v] && !coalesced[v] && uses[v] == 0 && literals.release(out.ValueSlots[v]) {
			bucket := storageBucket(f.Values[v])
			free[bucket] = append(free[bucket], out.ValueSlots[v])
		}
	}
	for bi, b := range f.Blocks {
		literals.reset()
		for t := range free {
			if len(all[t]) != 0 || len(free[t]) != 0 {
				free[t] = append(free[t][:0], all[t]...)
			}
		}
		out.BlockPC[bi] = uint32(len(out.Code))
		for _, op := range b.Ops {
			src := f.Sources[op.Source]
			// Nontrapping integer evaluation reads every operand before writing
			// its result. Prefer reusing the left operand for two-address targets.
			code := src.Wasm.Op
			reuse := (code >= 0x45 && code <= 0x5a || code >= 0x67 && code <= 0x8a ||
				code == 0xa7 || code == 0xac || code == 0xad || code >= 0xc0 && code <= 0xc4) && src.Effects&ir.MayTrap == 0
			if reuse {
				for i := len(op.Args) - 1; i >= 0; i-- {
					v := op.Args[i]
					uses[v]--
					if uses[v] == 0 {
						release(v)
					}
				}
			}
			for _, v := range op.Results {
				if pinned[v] || coalesced[v] {
					continue
				}
				t := f.Values[v]
				key, constant := literalOf(src.Wasm, t)
				if constant {
					if slot, ok := literals.lookup(key); ok {
						out.ValueSlots[v] = slot
						continue
					}
				}
				bucket := storageBucket(t)
				if k := len(free[bucket]); k > 0 {
					out.ValueSlots[v] = free[bucket][k-1]
					free[bucket] = free[bucket][:k-1]
				} else {
					s := budget.addSlot(out, t)
					all[bucket] = append(all[bucket], s)
					out.ValueSlots[v] = s
				}
				if constant {
					literals.add(key, out.ValueSlots[v], budget)
				}
			}
			// SSA locations remain stable after release. Integer operands, local
			// copies and equal literals can alias; suspended call tuples cannot.
			in := Instruction{Op: src.Wasm.Op, Source: op.Source, Dst: NoSlot, A: NoSlot, B: NoSlot, C: NoSlot}
			switch {
			case in.Op >= 0x41 && in.Op <= 0x44:
				in.Extra, in.Count = uint32(src.Wasm.Bits), uint32(src.Wasm.Bits>>32)
				if in.Op == 0x41 {
					in.Count = 0
				}
			case in.Op >= 0x20 && in.Op <= 0x26 || in.Op == 0x3f || in.Op == 0x40 || in.Op == 0xd2:
				in.Extra = src.Wasm.A
			}
			if in.Op == 0x10 || in.Op == 0x11 || in.Op == 0x14 || in.Op>>16 == 0xfb || module.IsWideArithmetic(in.Op) || in.Op >= 0xe0 && in.Op <= 0xe6 {
				in.Extra = uint32(len(out.Args))
				if in.Op>>16 == 0xfb && len(op.Args) > 0 {
					in.A = out.ValueSlots[op.Args[0]]
				}
				if in.Op>>16 == 0xfb && len(op.Args) > 1 {
					in.B = out.ValueSlots[op.Args[1]]
				}
				in.Count = uint32(len(op.Args))
				budget.charge(uint64(len(op.Args)) + uint64(len(op.Results)))
				for _, v := range op.Args {
					out.Args = append(out.Args, out.ValueSlots[v])
				}
				if len(op.Results) > 0 {
					in.Dst = out.ValueSlots[op.Results[0]]
				}
				for _, v := range op.Results {
					out.Args = append(out.Args, out.ValueSlots[v])
				}
			} else {
				if len(op.Args) > 0 {
					in.A = out.ValueSlots[op.Args[0]]
				}
				if len(op.Args) > 1 {
					in.B = out.ValueSlots[op.Args[1]]
				}
				if len(op.Args) > 2 {
					in.C = out.ValueSlots[op.Args[2]]
				}
				if len(op.Results) > 0 {
					in.Dst = out.ValueSlots[op.Results[0]]
				}
			}
			out.Code = append(out.Code, in)
			if !reuse {
				for _, v := range op.Args {
					uses[v]--
					if uses[v] == 0 {
						release(v)
					}
				}
			}
			for _, v := range op.Results {
				release(v)
			}
		}
		// Edge arguments remain live through this block's terminator. Their slots
		// can be recycled in other blocks after the selected edge's moves complete.
		term := b.Term
		in := Instruction{Op: ControlBase + uint32(term.Kind), Source: term.Source, Dst: NoSlot, A: NoSlot, B: NoSlot, C: NoSlot}
		if len(term.Args) > 0 {
			in.A = out.ValueSlots[term.Args[0]]
		}
		if term.Kind == ir.IfCast && len(term.Args) > 1 {
			in.B = out.ValueSlots[term.Args[1]]
		}
		if term.Kind == ir.TailCall || term.Kind == ir.Throw {
			in.Op = f.Sources[term.Source].Wasm.Op
			in.A = NoSlot
		}
		if term.Kind == ir.Return || term.Kind == ir.TailCall || term.Kind == ir.Throw {
			in.Extra = uint32(len(out.Args))
			in.Count = uint32(len(term.Args))
			budget.charge(uint64(len(term.Args)))
			for _, v := range term.Args {
				out.Args = append(out.Args, out.ValueSlots[v])
			}
		}
		out.Code = append(out.Code, in)
	}
	// All target PCs and value locations now exist. Resolve each parallel tuple
	// before flattening it into a sequential, allocation-free transfer program.
	budget.charge(uint64(nBlocks))
	returns := reuseSlice(c.returns, nBlocks)
	clear(returns)
	c.returns = returns
	for bi, b := range f.Blocks {
		if len(b.Ops) != 0 || b.Term.Kind != ir.Return || len(b.Term.Args) != len(b.Params) || f.Sources[b.Term.Source].Fuel() != 0 {
			continue
		}
		budget.charge(uint64(len(b.Params)))
		returns[bi] = true
		for i, v := range b.Params {
			if b.Term.Args[i] != v {
				returns[bi] = false
				break
			}
		}
	}
	if len(c.scratch) != 0 {
		clear(c.scratch)
	}
	var scratchFn func(Slot) Slot
	for bi, b := range f.Blocks {
		at := out.BlockPC[bi] + uint32(len(b.Ops))
		if len(b.Term.Edges) == 0 {
			continue
		}
		in := &out.Code[at]
		if b.Term.Kind == ir.Jump {
			edge := b.Term.Edges[0]
			if out.BlockPC[edge.Target] > at && returns[edge.Target] {
				// Preserve the jump's source/fuel boundary and every original PC.
				// A forward pure return needs neither its tuple copy nor a second
				// dispatch. Backward edges retain their cancellation poll.
				budget.charge(uint64(len(edge.Args)))
				in.Op = ControlBase + uint32(ir.Return)
				in.Extra, in.Count = uint32(len(out.Args)), uint32(len(edge.Args))
				for _, v := range edge.Args {
					out.Args = append(out.Args, out.ValueSlots[v])
				}
				continue
			}
		}
		in.Extra = uint32(len(out.Edges))
		in.Count = uint32(len(b.Term.Edges))
		budget.charge(uint64(len(b.Term.Edges)))
		for _, e := range b.Term.Edges {
			budget.charge(uint64(len(e.Args)) * 4)
			start := uint32(len(out.Moves))
			switch len(e.Args) {
			case 0:
				out.Edges = append(out.Edges, Edge{PC: out.BlockPC[e.Target], Start: start, Count: 0})
			case 1:
				dst := out.ValueSlots[f.Blocks[e.Target].Params[0]]
				src := out.ValueSlots[e.Args[0]]
				if dst != src {
					budget.charge(1)
					out.Moves = append(out.Moves, Move{dst, src})
					out.Edges = append(out.Edges, Edge{PC: out.BlockPC[e.Target], Start: start, Count: 1})
				} else {
					out.Edges = append(out.Edges, Edge{PC: out.BlockPC[e.Target], Start: start, Count: 0})
				}
			default:
				moves := c.moves[:0]
				params := f.Blocks[e.Target].Params
				for i, v := range e.Args {
					moves = append(moves, Move{out.ValueSlots[params[i]], out.ValueSlots[v]})
				}
				c.moves = moves
				if scratchFn == nil {
					scratchFn = func(dst Slot) Slot {
						t := out.Slots[dst]
						if c.scratch == nil {
							c.scratch = make(map[wasm.ValueType]Slot)
						} else if s, ok := c.scratch[t]; ok {
							return s
						}
						s := budget.addSlot(out, t)
						c.scratch[t] = s
						return s
					}
				}
				ordered := c.planner.order(moves, scratchFn)
				budget.charge(uint64(len(ordered)))
				out.Moves = append(out.Moves, ordered...)
				out.Edges = append(out.Edges, Edge{PC: out.BlockPC[e.Target], Start: start, Count: uint32(len(ordered))})
			}
		}
	}

	for _, h := range f.Handlers {
		budget.charge(1 + uint64(len(h.Catches)))
		outH := Handler{Parent: h.Parent}
		for _, c := range h.Catches {
			catch := Catch{Kind: c.Kind, Tag: c.Tag, PC: out.BlockPC[c.Target]}
			budget.charge(uint64(len(f.Blocks[c.Target].Params)))
			for _, v := range f.Blocks[c.Target].Params {
				catch.Dest = append(catch.Dest, out.ValueSlots[v])
			}
			outH.Catches = append(outH.Catches, catch)
		}
		out.Handlers = append(out.Handlers, outH)
	}
	if len(f.Prompts) != 0 {
		out.Prompts = make(map[uint32][]Catch, len(f.Prompts))
		for source, hs := range f.Prompts {
			for _, h := range hs {
				c := Catch{Kind: h.Kind, Tag: h.Tag}
				if h.Kind == 0 {
					c.PC = out.BlockPC[h.Target]
					c.Type, _ = h.Types[len(h.Types)-1].TypeIndex()
					for _, p := range f.Blocks[h.Target].Params {
						c.Dest = append(c.Dest, out.ValueSlots[p])
					}
				}
				out.Prompts[source] = append(out.Prompts[source], c)
			}
		}
	}
	for bi, b := range f.Blocks {
		if b.Handler != 0 {
			end := uint32(len(out.Code))
			if bi+1 < len(f.Blocks) {
				end = out.BlockPC[bi+1]
			}
			budget.charge(1)
			out.Protected = append(out.Protected, ProtectedBlock{out.BlockPC[bi], end, b.Handler})
		}
	}
	cachePrefixes(out, budget)
	out.Veneer = callVeneer(out)
	if out.Veneer != nil {
		budget.charge(1)
	}
	return out
}

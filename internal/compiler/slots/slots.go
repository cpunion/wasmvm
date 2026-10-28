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

type Instruction struct {
	Op, Source   uint32
	Dst, A, B, C Slot
	Extra, Count uint32
}

// Calls store input slots followed by output slots in Args[Extra:], with Count
// inputs and Source.Outputs outputs. Returns store Count slots at Extra.
// Other operations have at most three inline operands and one inline result.
// Branches address Count edges starting at Extra; A is the condition/index.
// Every instruction retains its original source and fuel boundary.
type Edge struct{ PC, Start, Count uint32 }
type Move struct{ Dst, Src Slot }
type Function struct {
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
	f.Slots = append(f.Slots, t)
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
	for i, f := range m.Functions {
		if err := verify.Verify(f); err != nil {
			return nil, err
		}
		result.Functions[i] = compileFunction(f, &b)
	}
	return result, nil
}

func compileFunction(f *ir.Function, budget *budget) *Function {
	n := len(f.Values)
	budget.charge(uint64(n)*4 + uint64(len(f.Blocks)))
	out := &Function{Signature: f.Signature, Sources: f.Sources, LocalCount: uint32(len(f.Locals)), ValueSlots: make([]Slot, n), BlockPC: make([]uint32, len(f.Blocks))}
	for _, t := range f.Locals {
		budget.addSlot(out, t)
	}
	defs := make([]int, n)
	pinned := make([]bool, n)
	uses := make([]uint32, n)
	budget.charge(uint64(n))
	lastUse := make([]uint32, n)
	for bi, b := range f.Blocks {
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
	use := func(v ir.Value, bi, position int) {
		uses[v]++
		if defs[v] != bi {
			pinned[v] = true
		} else {
			lastUse[v] = uint32(position)
		}
	}
	for bi, b := range f.Blocks {
		budget.charge(uint64(len(b.Ops)) + 1)
		for oi, op := range b.Ops {
			for _, v := range op.Args {
				use(v, bi, oi)
			}
		}
		for _, v := range b.Term.Args {
			use(v, bi, len(b.Ops))
		}
		for _, e := range b.Term.Edges {
			for _, v := range e.Args {
				use(v, bi, len(b.Ops))
			}
		}
	}
	coalesced := coalesceLocals(f, out, pinned, lastUse, budget)
	for v, t := range f.Values {
		if pinned[v] {
			out.ValueSlots[v] = budget.addSlot(out, t)
		}
	}
	// Only block-local temporaries are recycled in this first allocator. Values
	// crossing blocks and block parameters keep dedicated slots; this avoids a
	// quadratic global interference graph and preserves loops conservatively.
	var free, all [128][]Slot
	var literals literalPool
	release := func(v ir.Value) {
		if !pinned[v] && !coalesced[v] && uses[v] == 0 && literals.release(out.ValueSlots[v]) {
			t := f.Values[v]
			free[t] = append(free[t], out.ValueSlots[v])
		}
	}
	for bi, b := range f.Blocks {
		literals = literalPool{}
		for t := range free {
			free[t] = append(free[t][:0], all[t]...)
		}
		out.BlockPC[bi] = uint32(len(out.Code))
		for _, op := range b.Ops {
			src := f.Sources[op.Source]
			for _, v := range op.Results {
				if pinned[v] || coalesced[v] {
					continue
				}
				t := f.Values[v]
				key, constant := literalOf(src.Wasm, t)
				if constant {
					if shared := literals.values[key]; shared != nil {
						shared.refs++
						out.ValueSlots[v] = shared.slot
						continue
					}
				}
				if k := len(free[t]); k > 0 {
					out.ValueSlots[v] = free[t][k-1]
					free[t] = free[t][:k-1]
				} else {
					s := budget.addSlot(out, t)
					all[t] = append(all[t], s)
					out.ValueSlots[v] = s
				}
				if constant {
					literals.add(key, out.ValueSlots[v], budget)
				}
			}
			// Map operands before releasing slots. Local copies and identical
			// literals can alias; call inputs/outputs remain distinct on suspension.
			in := Instruction{Op: src.Wasm.Op, Source: op.Source, Dst: NoSlot, A: NoSlot, B: NoSlot, C: NoSlot}
			if in.Op == 0x10 || in.Op == 0x11 {
				in.Extra = uint32(len(out.Args))
				in.Count = uint32(len(op.Args))
				budget.charge(uint64(len(op.Args)) + uint64(len(op.Results)))
				for _, v := range op.Args {
					out.Args = append(out.Args, out.ValueSlots[v])
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
			for _, v := range op.Args {
				uses[v]--
				if uses[v] == 0 {
					release(v)
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
		if term.Kind == ir.Return {
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
	budget.charge(uint64(len(f.Blocks)))
	returns := make([]bool, len(f.Blocks))
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
	scratch := make(map[wasm.ValueType]Slot)
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
			moves := make([]Move, 0, len(e.Args))
			for i, v := range e.Args {
				moves = append(moves, Move{out.ValueSlots[f.Blocks[e.Target].Params[i]], out.ValueSlots[v]})
			}
			ordered := parallelMoves(moves, func(dst Slot) Slot {
				t := out.Slots[dst]
				if s, ok := scratch[t]; ok {
					return s
				}
				s := budget.addSlot(out, t)
				scratch[t] = s
				return s
			})
			edge := Edge{PC: out.BlockPC[e.Target], Start: uint32(len(out.Moves)), Count: uint32(len(ordered))}
			budget.charge(uint64(len(ordered)))
			out.Moves = append(out.Moves, ordered...)
			out.Edges = append(out.Edges, edge)
		}
	}
	return out
}

// Package lower converts validated Core 2.0 code into typed control-flow SSA.
// All control transfers carry explicit tuples; guest loops never recurse in Go.
package lower

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct{ MaxNodes, MaxVerifyWork uint64 }

func DefaultLimits() Limits { return Limits{8 << 20, 100 << 20} }

type Module struct {
	Validated *validate.Module
	Functions []*ir.Function
	constants map[*module.Instr]uint64
}
type budget struct{ remaining uint64 }

func (b *budget) charge(n uint64) {
	if n > b.remaining {
		panic(&wasm.Error{Kind: wasm.ErrLimit, Detail: "IR nodes"})
	}
	b.remaining -= n
}

// Compile consumes no guest state. The validated definition and returned IR
// must remain immutable for all execution backends. Quotas apply across the
// entire module; no partially compiled module is returned on failure.
func Compile(m *validate.Module, limits Limits) (result *Module, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(*wasm.Error); ok {
				result, err = nil, e
			} else {
				panic(failure)
			}
		}
	}()
	if m == nil {
		return nil, &wasm.Error{Kind: wasm.ErrInvalid, Detail: "nil validated module"}
	}
	b := budget{limits.MaxNodes}
	verify := ir.VerifyBudget{Remaining: limits.MaxVerifyWork}
	definition := m.Definition()
	b.charge(uint64(len(definition.Functions)))
	result = &Module{Validated: m, Functions: make([]*ir.Function, len(definition.Functions))}
	result.foldConstants(&b)
	for index := range definition.Functions {
		f := &definition.Functions[index]
		sig := definition.Types[f.Type]
		b.charge(uint64(len(f.Body)) + uint64(len(sig.Params)) + uint64(f.LocalCount))
		out := &ir.Function{Signature: sig, Locals: append([]wasm.ValueType(nil), sig.Params...), Sources: make([]ir.Source, len(f.Body)), TypeLabels: m.TypeLabels()}
		for _, group := range f.Locals {
			for n := uint32(0); n < group.Count; n++ {
				out.Locals = append(out.Locals, group.Type)
			}
		}
		for pc, in := range f.Body {
			out.Sources[pc] = ir.Source{Wasm: in, Effects: ir.EffectsOf(in.Op)}
		}
		l := lowerer{budget: &b, definition: definition, f: out, current: 0}
		l.newBlock(nil)
		exit := l.newBlock(sig.Results)
		l.frames = append(l.frames, control{end: exit, label: exit})
		if err = m.WalkFunction(uint32(index), l.visit); err != nil {
			return nil, err
		}
		l.prune()
		if err = verify.Verify(out); err != nil {
			return nil, err
		}
		result.Functions[index] = out
	}
	return result, nil
}

type control struct {
	op                    uint32
	height                int
	end, label, otherwise ir.BlockID
	hasElse               bool
}
type lowerer struct {
	budget     *budget
	definition *module.Module
	f          *ir.Function
	current    int
	stack      []ir.Value
	frames     []control
	incoming   []int
}

func (l *lowerer) newValue(t wasm.ValueType) ir.Value {
	l.budget.charge(1)
	id := ir.Value(len(l.f.Values))
	l.f.Values = append(l.f.Values, t)
	return id
}
func (l *lowerer) newBlock(types []wasm.ValueType) ir.BlockID {
	l.budget.charge(1 + uint64(len(types)))
	id := ir.BlockID(len(l.f.Blocks))
	b := ir.Block{}
	for _, t := range types {
		b.Params = append(b.Params, l.newValue(t))
	}
	l.f.Blocks = append(l.f.Blocks, b)
	l.incoming = append(l.incoming, 0)
	return id
}
func (l *lowerer) copyValues(xs []ir.Value) []ir.Value {
	l.budget.charge(uint64(len(xs)))
	return append([]ir.Value(nil), xs...)
}
func (l *lowerer) suffix(n int) []ir.Value { return l.stack[len(l.stack)-n:] }
func (l *lowerer) edge(target ir.BlockID) ir.Edge {
	l.incoming[target]++
	return ir.Edge{Target: target, Args: l.copyValues(l.suffix(len(l.f.Blocks[target].Params)))}
}
func (l *lowerer) terminate(t ir.Terminator) {
	l.budget.charge(1 + uint64(len(t.Args)) + uint64(len(t.Edges)))
	l.f.Blocks[l.current].Term = t
	l.current = -1
}
func (l *lowerer) jump(target ir.BlockID, source uint32) {
	if l.current >= 0 {
		l.terminate(ir.Terminator{Kind: ir.Jump, Source: source, Edges: []ir.Edge{l.edge(target)}})
	}
}
func (l *lowerer) switchTo(target ir.BlockID, height int) {
	l.stack = l.stack[:height]
	if l.incoming[target] == 0 {
		l.current = -1
		return
	}
	l.current = int(target)
	l.stack = append(l.stack, l.f.Blocks[target].Params...)
}
func (l *lowerer) blockType(t int64) module.FuncType {
	if t == -64 {
		return module.FuncType{}
	}
	if t < 0 {
		return module.FuncType{Results: []wasm.ValueType{module.BlockResult(t)}}
	}
	return l.definition.Types[t]
}
func (l *lowerer) visit(pc int, in module.Instr, inputs, outputs []wasm.ValueType) error {
	source := uint32(pc)
	switch op := in.Op; {
	case op == 2 || op == 3 || op == 4:
		sig := l.blockType(in.Block)
		height := len(l.stack)
		var condition ir.Value
		if l.current >= 0 {
			if op == 4 {
				condition = l.stack[len(l.stack)-1]
				l.stack = l.stack[:len(l.stack)-1]
			}
			height = len(l.stack) - len(sig.Params)
		}
		begin, end := l.newBlock(sig.Params), l.newBlock(sig.Results)
		f := control{op: op, height: height, end: end, label: end}
		if op == 3 {
			f.label = begin
		}
		if op == 4 {
			f.otherwise = l.newBlock(sig.Params)
			if l.current >= 0 {
				l.terminate(ir.Terminator{Kind: ir.If, Source: source, Args: []ir.Value{condition}, Edges: []ir.Edge{l.edge(begin), l.edge(f.otherwise)}})
			}
		} else {
			l.jump(begin, source)
		}
		l.frames = append(l.frames, f)
		l.switchTo(begin, height)
	case op == 5:
		f := &l.frames[len(l.frames)-1]
		l.jump(f.end, source)
		f.hasElse = true
		l.switchTo(f.otherwise, f.height)
	case op == 0xb:
		f := l.frames[len(l.frames)-1]
		l.jump(f.end, source)
		if f.op == 4 && !f.hasElse {
			l.switchTo(f.otherwise, f.height)
			l.jump(f.end, source)
		}
		l.switchTo(f.end, f.height)
		l.frames = l.frames[:len(l.frames)-1]
		if len(l.frames) == 0 && l.current >= 0 {
			l.terminate(ir.Terminator{Kind: ir.Return, Source: source, Args: l.copyValues(l.stack)})
		}
	case l.current < 0:
		// Structurally unreachable operations have no run-time effect or fuel.
	case op == 0:
		l.terminate(ir.Terminator{Kind: ir.Trap, Source: source})
	case module.IsTailCall(op):
		l.budget.charge(uint64(len(inputs)))
		l.f.Sources[pc].Inputs = append([]wasm.ValueType(nil), inputs...)
		// The callee return schema is verified, but creates no local SSA results.
		l.f.Sources[pc].Outputs = append([]wasm.ValueType(nil), l.f.Signature.Results...)
		l.terminate(ir.Terminator{Kind: ir.TailCall, Source: source, Args: l.copyValues(l.suffix(len(inputs)))})
	case op == 0xd5 || op == 0xd6:
		ref := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		label := l.frames[len(l.frames)-1-int(in.A)].label
		if op == 0xd5 {
			cont := l.newBlock([]wasm.ValueType{l.f.Values[ref].NonNull()})
			l.incoming[cont]++
			l.terminate(ir.Terminator{Kind: ir.IfNull, Source: source, Args: []ir.Value{ref}, Edges: []ir.Edge{l.edge(label), {Target: cont, Args: []ir.Value{ref}}}})
			l.current = int(cont)
			l.stack = append(l.stack, l.f.Blocks[cont].Params...)
		} else {
			cont := l.newBlock(nil)
			l.incoming[cont]++
			l.stack = append(l.stack, ref)
			edge := l.edge(label)
			l.stack = l.stack[:len(l.stack)-1]
			l.terminate(ir.Terminator{Kind: ir.IfNull, Source: source, Args: []ir.Value{ref}, Edges: []ir.Edge{{Target: cont}, edge}})
			l.current = int(cont)
		}
	case op == 0xc || op == 0xf:
		label := l.frames[0].label
		if op == 0xc {
			label = l.frames[len(l.frames)-1-int(in.A)].label
		}
		l.jump(label, source)
	case op == 0xd:
		condition := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		label := l.frames[len(l.frames)-1-int(in.A)].label
		cont := l.newBlock(nil)
		l.incoming[cont]++
		l.terminate(ir.Terminator{Kind: ir.If, Source: source, Args: []ir.Value{condition}, Edges: []ir.Edge{l.edge(label), {Target: cont}}})
		l.current = int(cont)
	case op == 0xe:
		index := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		l.budget.charge(uint64(len(in.Labels)) + 1)
		edges := make([]ir.Edge, 0, len(in.Labels)+1)
		for _, idx := range in.Labels {
			edges = append(edges, l.edge(l.frames[len(l.frames)-1-int(idx)].label))
		}
		edges = append(edges, l.edge(l.frames[len(l.frames)-1-int(in.A)].label))
		l.terminate(ir.Terminator{Kind: ir.Table, Source: source, Args: []ir.Value{index}, Edges: edges})
	default:
		l.budget.charge(1 + uint64(len(inputs)) + uint64(len(outputs)))
		s := &l.f.Sources[pc]
		s.Inputs = append([]wasm.ValueType(nil), inputs...)
		s.Outputs = append([]wasm.ValueType(nil), outputs...)
		operation := ir.Operation{Source: source, Args: l.copyValues(l.suffix(len(inputs)))}
		l.stack = l.stack[:len(l.stack)-len(inputs)]
		for _, typ := range outputs {
			id := l.newValue(typ)
			operation.Results = append(operation.Results, id)
			l.stack = append(l.stack, id)
		}
		l.f.Blocks[l.current].Ops = append(l.f.Blocks[l.current].Ops, operation)
	}
	return nil
}

// Remove structurally unreachable blocks and compact value identities. Only
// reachable blocks carry instructions; unexecuted source locations remain in
// the source table for diagnostics, without becoming executable operations.
func (l *lowerer) prune() {
	f := l.f
	l.budget.charge(uint64(len(f.Blocks)) + uint64(len(f.Values)))
	seen := make([]bool, len(f.Blocks))
	pending := []ir.BlockID{0}
	seen[0] = true
	for len(pending) > 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, edge := range f.Blocks[id].Term.Edges {
			if !seen[edge.Target] {
				seen[edge.Target] = true
				pending = append(pending, edge.Target)
			}
		}
	}
	blocks := make([]ir.BlockID, len(f.Blocks))
	var compact []ir.Block
	values := make([]ir.Value, len(f.Values))
	var types []wasm.ValueType
	assign := func(id ir.Value) ir.Value {
		next := ir.Value(len(types))
		types = append(types, f.Values[id])
		values[id] = next
		return next
	}
	for i, b := range f.Blocks {
		if seen[i] {
			blocks[i] = ir.BlockID(len(compact))
			for j, id := range b.Params {
				b.Params[j] = assign(id)
			}
			for j := range b.Ops {
				for k, id := range b.Ops[j].Results {
					b.Ops[j].Results[k] = assign(id)
				}
			}
			compact = append(compact, b)
		}
	}
	remap := func(xs []ir.Value) {
		for i, id := range xs {
			xs[i] = values[id]
		}
	}
	for i := range compact {
		b := &compact[i]
		for j := range b.Ops {
			remap(b.Ops[j].Args)
		}
		remap(b.Term.Args)
		for j := range b.Term.Edges {
			edge := &b.Term.Edges[j]
			edge.Target = blocks[edge.Target]
			remap(edge.Args)
		}
	}
	f.Blocks, f.Values = compact, types
}

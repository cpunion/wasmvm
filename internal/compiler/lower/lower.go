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

	var tags []module.FuncType
	for _, im := range definition.Imports {
		if im.Kind == 4 {
			b.charge(1)
			tags = append(tags, definition.Types[im.Type])
		}
	}
	for _, t := range definition.Tags {
		b.charge(1)
		tags = append(tags, definition.Types[t])
	}
	l := lowerer{budget: &b, definition: definition, tags: tags}
	visit := l.visit
	for index := range definition.Functions {
		f := &definition.Functions[index]
		sig := definition.Types[f.Type]
		b.charge(uint64(len(f.Body)) + uint64(len(sig.Params)) + uint64(f.LocalCount))
		locals := make([]wasm.ValueType, len(sig.Params), len(sig.Params)+int(f.LocalCount))
		copy(locals, sig.Params)
		for _, group := range f.Locals {
			for n := uint32(0); n < group.Count; n++ {
				locals = append(locals, group.Type)
			}
		}
		out := &ir.Function{Signature: sig, Locals: locals, Sources: make([]ir.Source, len(f.Body)), TypeLabels: m.TypeLabels(), TypeHierarchy: m.TypeHierarchy()}
		for pc, in := range f.Body {
			out.Sources[pc] = ir.Source{Wasm: in, Effects: ir.EffectsOf(in.Op)}
		}
		l.reset(out, len(f.Body))
		l.newBlock(nil)
		exit := l.newBlock(sig.Results)
		l.frames = append(l.frames, control{end: exit, label: exit})
		if err = m.WalkFunction(uint32(index), visit); err != nil {
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
	handler    uint32
	tags       []module.FuncType
	// Per-function chunked slabs pack immutable IR slices without cross-function retention.
	typeSlab  []wasm.ValueType
	valSlab   []ir.Value
	edgeSlab  []ir.Edge
	typeChunk int
	valChunk  int
	edgeChunk int
	// Reusable prune scratch across functions in one Compile call.
	seen    []bool
	pending []ir.BlockID
	blocks  []ir.BlockID
	values  []ir.Value
	used    []bool
	mapping []uint32
}

func (l *lowerer) reset(out *ir.Function, bodyLen int) {
	l.f = out
	l.current = 0
	l.handler = 0
	l.stack = l.stack[:0]
	l.frames = l.frames[:0]
	l.incoming = l.incoming[:0]
	l.typeSlab = nil
	l.valSlab = nil
	l.edgeSlab = nil
	l.typeChunk = min(max(bodyLen*2, 16), 4096)
	l.valChunk = min(max(bodyLen*2, 16), 4096)
	l.edgeChunk = min(max(bodyLen/4+4, 8), 1024)
}

func (l *lowerer) allocTypes(n int) []wasm.ValueType {
	if n == 0 {
		return nil
	}
	if cap(l.typeSlab)-len(l.typeSlab) < n {
		l.typeSlab = make([]wasm.ValueType, 0, max(n, l.typeChunk))
	}
	start := len(l.typeSlab)
	l.typeSlab = l.typeSlab[:start+n]
	return l.typeSlab[start : start+n : start+n]
}

func (l *lowerer) copyTypes(xs []wasm.ValueType) []wasm.ValueType {
	if len(xs) == 0 {
		return nil
	}
	dst := l.allocTypes(len(xs))
	copy(dst, xs)
	return dst
}

func (l *lowerer) allocValues(n int) []ir.Value {
	if n == 0 {
		return nil
	}
	if cap(l.valSlab)-len(l.valSlab) < n {
		l.valSlab = make([]ir.Value, 0, max(n, l.valChunk))
	}
	start := len(l.valSlab)
	l.valSlab = l.valSlab[:start+n]
	return l.valSlab[start : start+n : start+n]
}

func (l *lowerer) oneValue(v ir.Value) []ir.Value {
	dst := l.allocValues(1)
	dst[0] = v
	return dst
}

func (l *lowerer) allocEdges(n int) []ir.Edge {
	if n == 0 {
		return nil
	}
	if cap(l.edgeSlab)-len(l.edgeSlab) < n {
		l.edgeSlab = make([]ir.Edge, 0, max(n, l.edgeChunk))
	}
	start := len(l.edgeSlab)
	l.edgeSlab = l.edgeSlab[:start+n]
	return l.edgeSlab[start : start+n : start+n]
}

func (l *lowerer) oneEdge(e ir.Edge) []ir.Edge {
	dst := l.allocEdges(1)
	dst[0] = e
	return dst
}

func (l *lowerer) twoEdges(a, b ir.Edge) []ir.Edge {
	dst := l.allocEdges(2)
	dst[0], dst[1] = a, b
	return dst
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
	b := ir.Block{Handler: l.handler}
	if len(types) > 0 {
		params := l.allocValues(len(types))
		for i, t := range types {
			params[i] = l.newValue(t)
		}
		b.Params = params
	}
	l.f.Blocks = append(l.f.Blocks, b)
	l.incoming = append(l.incoming, 0)
	return id
}
func (l *lowerer) copyValues(xs []ir.Value) []ir.Value {
	l.budget.charge(uint64(len(xs)))
	if len(xs) == 0 {
		return nil
	}
	dst := l.allocValues(len(xs))
	copy(dst, xs)
	return dst
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
		l.terminate(ir.Terminator{Kind: ir.Jump, Source: source, Edges: l.oneEdge(l.edge(target))})
	}
}
func (l *lowerer) switchTo(target ir.BlockID, height int) {
	l.handler = l.f.Blocks[target].Handler
	l.stack = l.stack[:height]
	if l.incoming[target] == 0 {
		l.current = -1
		return
	}
	l.current = int(target)
	l.stack = append(l.stack, l.f.Blocks[target].Params...)
}
func (l *lowerer) blockType(t int64, single *[1]wasm.ValueType) module.FuncType {
	if t == -64 {
		return module.FuncType{}
	}
	if t < 0 {
		single[0] = module.BlockResult(t)
		return module.FuncType{Results: single[:]}
	}
	return l.definition.Types[t]
}
func (l *lowerer) visit(pc int, in module.Instr, inputs, outputs []wasm.ValueType) error {
	source := uint32(pc)
	switch op := in.Op; {
	case op == 2 || op == 3 || op == 4 || op == 0x1f:
		var single [1]wasm.ValueType
		sig := l.blockType(in.Block, &single)
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
				l.terminate(ir.Terminator{Kind: ir.If, Source: source, Args: l.oneValue(condition), Edges: l.twoEdges(l.edge(begin), l.edge(f.otherwise))})
			}
		} else {
			l.jump(begin, source)
		}

		if op == 0x1f && l.incoming[begin] != 0 && len(in.Catches) != 0 {
			h := ir.Handler{Parent: l.handler}
			l.budget.charge(1 + uint64(len(in.Catches)))
			for _, c := range in.Catches {
				target := l.frames[len(l.frames)-1-int(c.Label)].label
				var types []wasm.ValueType
				if c.Kind < 2 {
					types = l.tags[c.Tag].Params
				}
				extra := 0
				if c.Kind&1 != 0 {
					extra = 1
				}
				ts := l.allocTypes(len(types) + extra)
				copy(ts, types)
				if extra != 0 {
					ts[len(types)] = wasm.ExnRef.NonNull()
				}
				l.budget.charge(uint64(len(ts)))
				h.Catches = append(h.Catches, ir.Catch{Kind: c.Kind, Tag: c.Tag, Target: target, Types: ts})
				l.incoming[target]++
			}
			l.f.Handlers = append(l.f.Handlers, h)
			l.f.Blocks[begin].Handler = uint32(len(l.f.Handlers))
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

	case op == 8 || op == 10:
		l.budget.charge(uint64(len(inputs)))
		l.f.Sources[pc].Inputs = l.copyTypes(inputs)
		l.terminate(ir.Terminator{Kind: ir.Throw, Source: source, Args: l.copyValues(l.suffix(len(inputs)))})
	case module.IsTailCall(op):
		l.budget.charge(uint64(len(inputs)))
		l.f.Sources[pc].Inputs = l.copyTypes(inputs)
		// The callee return schema is verified, but creates no local SSA results.
		l.f.Sources[pc].Outputs = l.copyTypes(l.f.Signature.Results)
		l.terminate(ir.Terminator{Kind: ir.TailCall, Source: source, Args: l.copyValues(l.suffix(len(inputs)))})
	case op == 0xfb0018 || op == 0xfb0019 || op == 0xfb0025 || op == 0xfb0026:
		var desc ir.Value
		if op >= 0xfb0025 {
			desc = l.stack[len(l.stack)-1]
			l.stack = l.stack[:len(l.stack)-1]
		}
		ref := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		label := l.frames[len(l.frames)-1-int(in.A)].label
		diff := in.Types[0]
		if in.Types[1].Nullable() {
			diff = diff.NonNull()
		}
		fall := diff
		if op == 0xfb0019 || op == 0xfb0026 {
			fall = in.Types[1]
		}
		var fallTypes [1]wasm.ValueType
		fallTypes[0] = fall
		cont := l.newBlock(fallTypes[:])
		l.incoming[cont]++
		l.stack = append(l.stack, ref)
		branch := l.edge(label)
		l.stack = l.stack[:len(l.stack)-1]
		edges := l.twoEdges(branch, ir.Edge{Target: cont, Args: l.oneValue(ref)})
		if op == 0xfb0019 || op == 0xfb0026 {
			edges[0], edges[1] = edges[1], edges[0]
		}
		var args []ir.Value
		if op >= 0xfb0025 {
			args = l.allocValues(2)
			args[0], args[1] = ref, desc
		} else {
			args = l.oneValue(ref)
		}
		l.terminate(ir.Terminator{Kind: ir.IfCast, Source: source, Args: args, Edges: edges})
		l.current = int(cont)
		l.stack = append(l.stack, l.f.Blocks[cont].Params...)

	case op == 0xd5 || op == 0xd6:
		ref := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		label := l.frames[len(l.frames)-1-int(in.A)].label
		if op == 0xd5 {
			var nonNull [1]wasm.ValueType
			nonNull[0] = l.f.Values[ref].NonNull()
			cont := l.newBlock(nonNull[:])
			l.incoming[cont]++
			l.terminate(ir.Terminator{Kind: ir.IfNull, Source: source, Args: l.oneValue(ref), Edges: l.twoEdges(l.edge(label), ir.Edge{Target: cont, Args: l.oneValue(ref)})})
			l.current = int(cont)
			l.stack = append(l.stack, l.f.Blocks[cont].Params...)
		} else {
			cont := l.newBlock(nil)
			l.incoming[cont]++
			l.stack = append(l.stack, ref)
			edge := l.edge(label)
			l.stack = l.stack[:len(l.stack)-1]
			l.terminate(ir.Terminator{Kind: ir.IfNull, Source: source, Args: l.oneValue(ref), Edges: l.twoEdges(ir.Edge{Target: cont}, edge)})
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
		l.terminate(ir.Terminator{Kind: ir.If, Source: source, Args: l.oneValue(condition), Edges: l.twoEdges(l.edge(label), ir.Edge{Target: cont})})
		l.current = int(cont)
	case op == 0xe:
		index := l.stack[len(l.stack)-1]
		l.stack = l.stack[:len(l.stack)-1]
		l.budget.charge(uint64(len(in.Labels)) + 1)
		edges := l.allocEdges(len(in.Labels) + 1)
		for i, idx := range in.Labels {
			edges[i] = l.edge(l.frames[len(l.frames)-1-int(idx)].label)
		}
		edges[len(in.Labels)] = l.edge(l.frames[len(l.frames)-1-int(in.A)].label)
		l.terminate(ir.Terminator{Kind: ir.Table, Source: source, Args: l.oneValue(index), Edges: edges})
	default:
		if op >= 0xe3 && op <= 0xe5 {
			if l.f.Prompts == nil {
				l.f.Prompts = make(map[uint32][]ir.Catch)
			}
			l.budget.charge(1 + uint64(len(in.On)))
			for _, h := range in.On {
				handler := ir.Catch{Kind: h.Kind, Tag: h.Tag}
				if h.Kind == 0 {
					handler.Target = l.frames[len(l.frames)-1-int(h.Label)].label
					l.incoming[handler.Target]++
					params := l.f.Blocks[handler.Target].Params
					handler.Types = l.allocTypes(len(params))
					for i, param := range params {
						handler.Types[i] = l.f.Values[param]
					}
					l.budget.charge(uint64(len(handler.Types)))
				}
				l.f.Prompts[source] = append(l.f.Prompts[source], handler)
			}
		}
		l.budget.charge(1 + uint64(len(inputs)) + uint64(len(outputs)))
		s := &l.f.Sources[pc]
		s.Inputs = l.copyTypes(inputs)
		s.Outputs = l.copyTypes(outputs)
		operation := ir.Operation{Source: source, Args: l.copyValues(l.suffix(len(inputs)))}
		l.stack = l.stack[:len(l.stack)-len(inputs)]
		if len(outputs) > 0 {
			res := l.allocValues(len(outputs))
			for i, typ := range outputs {
				id := l.newValue(typ)
				res[i] = id
				l.stack = append(l.stack, id)
			}
			operation.Results = res
		}
		l.f.Blocks[l.current].Ops = append(l.f.Blocks[l.current].Ops, operation)
	}
	return nil
}

func reuseSlice[T any](buf []T, n int) []T {
	if cap(buf) < n {
		return make([]T, n)
	}
	return buf[:n]
}

func (l *lowerer) markReachable(target ir.BlockID) {
	l.budget.charge(1)
	if !l.seen[target] {
		l.seen[target] = true
		l.pending = append(l.pending, target)
	}
}

// Remove structurally unreachable blocks and compact value identities. Only
// reachable blocks carry instructions; unexecuted source locations remain in
// the source table for diagnostics, without becoming executable operations.
func (l *lowerer) prune() {
	f := l.f
	nBlocks := len(f.Blocks)
	nValues := len(f.Values)
	l.budget.charge(uint64(nBlocks) + uint64(nValues))
	l.seen = reuseSlice(l.seen, nBlocks)
	clear(l.seen)
	l.pending = append(l.pending[:0], 0)
	l.seen[0] = true
	for len(l.pending) > 0 {
		id := l.pending[len(l.pending)-1]
		l.pending = l.pending[:len(l.pending)-1]
		b := f.Blocks[id]
		for _, e := range b.Term.Edges {
			l.markReachable(e.Target)
		}
		if len(f.Prompts) != 0 {
			for _, op := range b.Ops {
				for _, h := range f.Prompts[op.Source] {
					if h.Kind == 0 {
						l.markReachable(h.Target)
					}
				}
			}
		}
		for h := b.Handler; h != 0; h = f.Handlers[h-1].Parent {
			for _, c := range f.Handlers[h-1].Catches {
				l.markReachable(c.Target)
			}
		}
	}
	reachableBlocks, reachableValues := 0, 0
	for i := range nBlocks {
		if l.seen[i] {
			b := &f.Blocks[i]
			reachableBlocks++
			reachableValues += len(b.Params)
			for j := range b.Ops {
				reachableValues += len(b.Ops[j].Results)
			}
		}
	}
	l.blocks = reuseSlice(l.blocks, nBlocks)
	l.values = reuseSlice(l.values, nValues)
	blocks, values := l.blocks, l.values
	compact := make([]ir.Block, 0, reachableBlocks)
	var types []wasm.ValueType
	if reachableValues > 0 {
		types = make([]wasm.ValueType, 0, reachableValues)
	}
	for i, b := range f.Blocks {
		if l.seen[i] {
			blocks[i] = ir.BlockID(len(compact))
			for j, id := range b.Params {
				next := ir.Value(len(types))
				types = append(types, f.Values[id])
				values[id] = next
				b.Params[j] = next
			}
			for j := range b.Ops {
				res := b.Ops[j].Results
				for k, id := range res {
					next := ir.Value(len(types))
					types = append(types, f.Values[id])
					values[id] = next
					res[k] = next
				}
			}
			compact = append(compact, b)
		}
	}
	for i := range compact {
		b := &compact[i]
		for j := range b.Ops {
			args := b.Ops[j].Args
			for k, id := range args {
				args[k] = values[id]
			}
		}
		for k, id := range b.Term.Args {
			b.Term.Args[k] = values[id]
		}
		for j := range b.Term.Edges {
			edge := &b.Term.Edges[j]
			edge.Target = blocks[edge.Target]
			for k, id := range edge.Args {
				edge.Args[k] = values[id]
			}
		}
	}

	if len(f.Handlers) != 0 {
		l.used = reuseSlice(l.used, len(f.Handlers)+1)
		clear(l.used)
		for _, b := range compact {
			for h := b.Handler; h != 0 && !l.used[h]; h = f.Handlers[h-1].Parent {
				l.used[h] = true
			}
		}
		l.mapping = reuseSlice(l.mapping, len(l.used))
		clear(l.mapping)
		var handlers []ir.Handler
		for i, h := range f.Handlers {
			if !l.used[i+1] {
				continue
			}
			h.Parent = l.mapping[h.Parent]
			for j := range h.Catches {
				h.Catches[j].Target = blocks[h.Catches[j].Target]
			}
			handlers = append(handlers, h)
			l.mapping[i+1] = uint32(len(handlers))
		}
		for i := range compact {
			compact[i].Handler = l.mapping[compact[i].Handler]
		}
		f.Handlers = handlers
	}
	if f.Prompts != nil {
		prompts := make(map[uint32][]ir.Catch)
		for _, b := range compact {
			for _, op := range b.Ops {
				if hs, ok := f.Prompts[op.Source]; ok {
					for i := range hs {
						if hs[i].Kind == 0 {
							hs[i].Target = blocks[hs[i].Target]
						}
					}
					prompts[op.Source] = hs
				}
			}
		}
		f.Prompts = prompts
	}
	f.Blocks, f.Values = compact, types
}

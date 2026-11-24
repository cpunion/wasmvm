package ir

import (
	"errors"
	"fmt"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var ErrInvalid = errors.New("invalid compiler IR")

type definition struct {
	block, position int
	exists          bool
}

// Verify checks SSA definitions, operand/edge types, dominance, source schemas
// and terminators. All blocks must be reachable from block zero. Work limits
// include dominance iterations so hostile control graphs cannot cause an
// unbounded quadratic verification pass. Source schemas are compiler-owned;
// this is not a deserializer or replacement for Wasm validation.
func Verify(f *Function, maxWork uint64) (err error) {
	budget := VerifyBudget{Remaining: maxWork}
	return budget.Verify(f)
}

type verifyCursor struct{ block, next int }

// VerifyBudget bounds total verification work across a compilation, including
// failed passes. It must not be shared by concurrent compilation requests.
type VerifyBudget struct {
	Remaining uint64
	defs      []definition
	succOff   []int
	succData  []BlockID
	predCount []int
	predOff   []int
	predData  []int
	seen      []bool
	stack     []verifyCursor
	order     []int
	rank      []int
	idom      []int
}

func reuseSlice[T any](buf []T, n int) []T {
	if cap(buf) < n {
		return make([]T, n)
	}
	return buf[:n]
}

type verifier struct {
	budget *VerifyBudget
	f      *Function
}

func (v *verifier) step(n uint64) {
	if n > v.budget.Remaining {
		panic(verifyFailure{&wasm.Error{Kind: wasm.ErrLimit, Detail: "IR verification work"}})
	}
	v.budget.Remaining -= n
}

func (v *verifier) fail(message string) {
	panic(verifyFailure{fmt.Errorf("%w: %s", ErrInvalid, message)})
}

func (v *verifier) checkTypes(types []wasm.ValueType, detail string) {
	v.step(uint64(len(types)))
	for _, typ := range types {
		if !typ.Valid() {
			v.fail(detail)
		}
	}
}

func (v *verifier) matches(a, b wasm.ValueType) bool {
	if v.f.TypeHierarchy != nil {
		return v.f.TypeHierarchy.ReferenceSubtype(a, b, v.f.TypeLabels)
	}
	return wasm.ReferenceSubtype(a, b, v.f.TypeLabels)
}

func (v *verifier) valueType(id Value) wasm.ValueType {
	if uint64(id) >= uint64(len(v.f.Values)) {
		v.fail("value index")
	}
	return v.f.Values[id]
}

func (v *verifier) source(index uint32) Source {
	if uint64(index) >= uint64(len(v.f.Sources)) {
		v.fail("source index")
	}
	return v.f.Sources[index]
}

func (v *verifier) define(id Value, block, position int) {
	v.valueType(id)
	if v.budget.defs[id].exists {
		v.fail("multiple value definitions")
	}
	v.budget.defs[id] = definition{block, position, true}
}

func (v *verifier) recordSuccessor(target BlockID) {
	v.step(1)
	if uint64(target) >= uint64(len(v.f.Blocks)) {
		v.fail("successor index")
	}
	v.budget.succData = append(v.budget.succData, target)
	v.budget.predCount[target]++
}

func (v *verifier) recordSuccessors(b Block) {
	for _, e := range b.Term.Edges {
		v.recordSuccessor(e.Target)
	}
	if len(v.f.Prompts) != 0 {
		for _, op := range b.Ops {
			for _, h := range v.f.Prompts[op.Source] {
				if h.Kind == 0 {
					v.recordSuccessor(h.Target)
				}
			}
		}
	}
	for h := b.Handler; h != 0; h = v.f.Handlers[h-1].Parent {
		for _, c := range v.f.Handlers[h-1].Catches {
			v.recordSuccessor(c.Target)
		}
	}
}

func (v *verifier) intersect(a, b int) int {
	rank, idom := v.budget.rank, v.budget.idom
	for a != b {
		v.step(1)
		if rank[a] > rank[b] {
			a = idom[a]
		} else {
			b = idom[b]
		}
	}
	return a
}

func (v *verifier) use(id Value, block, position int) {
	v.step(1)
	d := v.budget.defs[id]
	if d.block == block {
		if d.position >= position {
			v.fail("value used before definition")
		}
		return
	}
	idom := v.budget.idom
	for block != d.block && block != 0 {
		v.step(1)
		block = idom[block]
	}
	if block != d.block {
		v.fail("value definition does not dominate use")
	}
}

func (budget *VerifyBudget) Verify(f *Function) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(verifyFailure); ok {
				err = e.error
			} else {
				panic(failure)
			}
		}
	}()
	v := verifier{budget: budget, f: f}
	if f == nil || len(f.Blocks) == 0 {
		v.fail("missing entry block")
	}
	nBlocks := len(f.Blocks)
	nValues := len(f.Values)
	v.step(uint64(nBlocks) + uint64(nValues) + uint64(len(f.Sources)))
	v.checkTypes(f.Signature.Params, "invalid function or local type")
	v.checkTypes(f.Signature.Results, "invalid function or local type")
	v.checkTypes(f.Locals, "invalid function or local type")
	if len(f.Locals) < len(f.Signature.Params) {
		v.fail("missing parameter locals")
	}
	for i, typ := range f.Signature.Params {
		if f.Locals[i] != typ {
			v.fail("parameter local type")
		}
	}
	for _, typ := range f.Values {
		if !typ.Valid() {
			v.fail("unknown value type")
		}
	}
	for _, s := range f.Sources {
		if s.Effects != EffectsOf(s.Wasm.Op) {
			v.fail("incorrect effects")
		}
		v.checkTypes(s.Inputs, "unknown operation type")
		v.checkTypes(s.Outputs, "unknown operation type")
	}
	budget.defs = reuseSlice(budget.defs, nValues)
	clear(budget.defs)
	budget.predCount = reuseSlice(budget.predCount, nBlocks)
	clear(budget.predCount)
	budget.succOff = reuseSlice(budget.succOff, nBlocks+1)
	budget.succData = budget.succData[:0]

	for i, h := range f.Handlers {
		v.step(1 + uint64(len(h.Catches)))
		if h.Parent > uint32(i) || len(h.Catches) == 0 {
			v.fail("handler parent")
		}
		for _, c := range h.Catches {
			if c.Kind > 3 || uint64(c.Target) >= uint64(nBlocks) {
				v.fail("catch kind/target")
			}
			ps := f.Blocks[c.Target].Params
			if len(ps) != len(c.Types) {
				v.fail("catch arity")
			}
			for j, t := range c.Types {
				v.step(1)
				if !v.matches(t, v.valueType(ps[j])) {
					v.fail("catch type")
				}
			}
			if c.Kind&1 != 0 && (len(c.Types) == 0 || c.Types[len(c.Types)-1] != wasm.ExnRef.NonNull()) {
				v.fail("catch reference")
			}
			if c.Kind == 2 && len(c.Types) != 0 || c.Kind == 3 && len(c.Types) != 1 {
				v.fail("catch-all arity")
			}
		}
	}
	for source, handlers := range f.Prompts {
		v.step(1 + uint64(len(handlers)))
		if uint64(source) >= uint64(len(f.Sources)) || f.Sources[source].Wasm.Op < 0xe3 || f.Sources[source].Wasm.Op > 0xe5 {
			v.fail("prompt source")
		}
		for _, h := range handlers {
			if h.Kind == 1 {
				continue
			}
			if h.Kind != 0 || uint64(h.Target) >= uint64(nBlocks) || len(h.Types) == 0 || len(h.Types) != len(f.Blocks[h.Target].Params) {
				v.fail("prompt target/arity")
			}
			for i, t := range h.Types {
				v.step(1)
				if !v.matches(t, v.valueType(f.Blocks[h.Target].Params[i])) {
					v.fail("prompt parameter type")
				}
			}
			if _, ok := h.Types[len(h.Types)-1].TypeIndex(); !ok || !h.Types[len(h.Types)-1].IsContinuationReference() {
				v.fail("prompt continuation type")
			}
		}
	}
	for bi, b := range f.Blocks {
		if uint64(b.Handler) > uint64(len(f.Handlers)) {
			v.fail("block handler")
		}
		budget.succOff[bi] = len(budget.succData)
		v.recordSuccessors(b)
		v.step(uint64(len(b.Params)) + uint64(len(b.Ops)) + uint64(len(b.Term.Edges)))
		for _, id := range b.Params {
			v.define(id, bi, -1)
		}
		for oi, op := range b.Ops {
			s := v.source(op.Source)
			if s.Wasm.Op != 1 && s.Wasm.Op != 0x10 && s.Wasm.Op != 0x11 && s.Wasm.Op != 0x14 && s.Wasm.Op < 0x1a {
				v.fail("control instruction in operation list")
			}
			if len(op.Args) != len(s.Inputs) || len(op.Results) != len(s.Outputs) {
				v.fail("operation arity")
			}
			v.step(uint64(len(op.Args)) + uint64(len(op.Results)))
			for i, id := range op.Args {
				if !v.matches(v.valueType(id), s.Inputs[i]) {
					v.fail("operation input type")
				}
			}
			for i, id := range op.Results {
				v.define(id, bi, oi)
				if v.valueType(id) != s.Outputs[i] {
					v.fail("operation result type")
				}
			}
		}
		t := b.Term
		s := v.source(t.Source)
		op := s.Wasm.Op
		switch t.Kind {
		case Jump:
			if len(t.Args) != 0 || len(t.Edges) != 1 {
				v.fail("jump arity")
			}
			if op != 2 && op != 3 && op != 0x1f && op != 5 && op != 0xb && op != 0xc && op != 0xf {
				v.fail("jump source")
			}
		case If, Table:
			if len(t.Args) != 1 || v.valueType(t.Args[0]) != wasm.I32 {
				v.fail("branch condition")
			}
			if t.Kind == If && (len(t.Edges) != 2 || op != 4 && op != 0xd) || t.Kind == Table && (len(t.Edges) == 0 || op != 0xe) {
				v.fail("branch targets/source")
			}
		case IfCast:
			arity := 1
			if op == 0xfb0025 || op == 0xfb0026 {
				arity = 2
			}
			if len(t.Args) != arity || !v.valueType(t.Args[0]).IsReference() || len(t.Edges) != 2 || (op != 0xfb0018 && op != 0xfb0019 && op != 0xfb0025 && op != 0xfb0026) || len(s.Wasm.Types) != 2 {
				v.fail("cast branch source/arity")
			}
			if arity == 2 && !v.valueType(t.Args[1]).IsReference() {
				v.fail("descriptor branch operand")
			}
		case IfNull:
			if len(t.Args) != 1 || !v.valueType(t.Args[0]).IsReference() || len(t.Edges) != 2 || (op != 0xd5 && op != 0xd6) {
				v.fail("null branch source/arity")
			}
		case Return:
			if op != 0xb || len(t.Edges) != 0 || len(t.Args) != len(f.Signature.Results) {
				v.fail("return arity/source")
			}
			for i, id := range t.Args {
				if !v.matches(v.valueType(id), f.Signature.Results[i]) {
					v.fail("return type")
				}
			}
		case TailCall:
			if !module.IsTailCall(op) || len(t.Edges) != 0 || len(t.Args) != len(s.Inputs) || len(s.Outputs) != len(f.Signature.Results) {
				v.fail("tail call arity/source")
			}
			v.step(uint64(len(t.Args)) + uint64(len(s.Outputs)))
			for i, typ := range s.Outputs {
				if !v.matches(typ, f.Signature.Results[i]) {
					v.fail("tail call return type")
				}
			}
			for i, id := range t.Args {
				if !v.matches(v.valueType(id), s.Inputs[i]) {
					v.fail("tail call operand type")
				}
			}

		case Throw:
			if (op != 8 && op != 10) || len(t.Edges) != 0 || len(t.Args) != len(s.Inputs) || len(s.Outputs) != 0 {
				v.fail("throw arity/source")
			}
			for i, id := range t.Args {
				v.step(1)
				if !v.matches(v.valueType(id), s.Inputs[i]) {
					v.fail("throw operand type")
				}
			}
		case Trap:
			if op != 0 || len(t.Args) != 0 || len(t.Edges) != 0 {
				v.fail("trap arity/source")
			}
		default:
			v.fail("missing terminator")
		}
		for ei, edge := range t.Edges {
			if uint64(edge.Target) >= uint64(nBlocks) {
				v.fail("block index")
			}
			params := f.Blocks[edge.Target].Params
			if len(edge.Args) != len(params) {
				v.fail("edge arity")
			}
			v.step(uint64(len(edge.Args)))
			for i, id := range edge.Args {
				actual := v.valueType(id)
				if t.Kind == IfCast && id == t.Args[0] {
					actual = s.Wasm.Types[1]
					if ei == 1 {
						actual = s.Wasm.Types[0]
						if s.Wasm.Types[1].Nullable() {
							actual = actual.NonNull()
						}
					}
				}
				if t.Kind == IfNull && ei == 1 && id == t.Args[0] {
					actual = actual.NonNull()
				}
				if !v.matches(actual, v.valueType(params[i])) {
					v.fail("edge type")
				}
			}
		}
	}
	budget.succOff[nBlocks] = len(budget.succData)
	if len(f.Blocks[0].Params) != 0 || budget.predCount[0] != 0 {
		v.fail("entry cannot have parameters or predecessors")
	}
	for _, d := range budget.defs {
		if !d.exists {
			v.fail("undefined value")
		}
	}
	budget.predOff = reuseSlice(budget.predOff, nBlocks+1)
	totalEdges := 0
	for i := range nBlocks {
		budget.predOff[i] = totalEdges
		totalEdges += budget.predCount[i]
	}
	budget.predOff[nBlocks] = totalEdges
	budget.predData = reuseSlice(budget.predData, totalEdges)
	copy(budget.predCount, budget.predOff[:nBlocks])
	succOff, succData := budget.succOff, budget.succData
	predData, predCursor := budget.predData, budget.predCount
	for bi := range nBlocks {
		for _, target := range succData[succOff[bi]:succOff[bi+1]] {
			pos := predCursor[target]
			predData[pos] = bi
			predCursor[target] = pos + 1
		}
	}
	// Iterative DFS produces reverse postorder without consuming host stack.
	seen := reuseSlice(budget.seen, nBlocks)
	budget.seen = seen
	clear(seen)
	seen[0] = true
	stack := append(budget.stack[:0], verifyCursor{0, 0})
	order := reuseSlice(budget.order, nBlocks)[:0]
	for len(stack) > 0 {
		v.step(1)
		top := &stack[len(stack)-1]
		edges := succData[succOff[top.block]:succOff[top.block+1]]
		if top.next == len(edges) {
			order = append(order, top.block)
			stack = stack[:len(stack)-1]
			continue
		}
		next := int(edges[top.next])
		top.next++
		if !seen[next] {
			seen[next] = true
			stack = append(stack, verifyCursor{block: next})
		}
	}
	budget.stack = stack
	budget.order = order
	if len(order) != nBlocks {
		v.fail("unreachable block")
	}
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	rank := reuseSlice(budget.rank, nBlocks)
	idom := reuseSlice(budget.idom, nBlocks)
	budget.rank, budget.idom = rank, idom
	for i, b := range order {
		rank[b] = i
		idom[b] = -1
	}
	idom[0] = 0
	predOff := budget.predOff
	for changed := true; changed; {
		changed = false
		for _, b := range order[1:] {
			v.step(1)
			parent := -1
			for _, p := range predData[predOff[b]:predOff[b+1]] {
				v.step(1)
				if idom[p] != -1 {
					if parent == -1 {
						parent = p
					} else {
						parent = v.intersect(parent, p)
					}
				}
			}
			if parent != idom[b] {
				idom[b] = parent
				changed = true
			}
		}
	}
	for bi, b := range f.Blocks {
		for oi, op := range b.Ops {
			for _, id := range op.Args {
				v.use(id, bi, oi)
			}
		}
		for _, id := range b.Term.Args {
			v.use(id, bi, len(b.Ops))
		}
		for _, edge := range b.Term.Edges {
			for _, id := range edge.Args {
				v.use(id, bi, len(b.Ops))
			}
		}
	}
	return nil
}

type verifyFailure struct{ error }

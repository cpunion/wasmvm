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

// VerifyBudget bounds total verification work across a compilation, including
// failed passes. It must not be shared by concurrent compilation requests.
type VerifyBudget struct{ Remaining uint64 }

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
	step := func(n uint64) {
		if n > budget.Remaining {
			panic(verifyFailure{&wasm.Error{Kind: wasm.ErrLimit, Detail: "IR verification work"}})
		}
		budget.Remaining -= n
	}
	fail := func(message string) { panic(verifyFailure{fmt.Errorf("%w: %s", ErrInvalid, message)}) }
	if f == nil || len(f.Blocks) == 0 {
		fail("missing entry block")
	}
	step(uint64(len(f.Blocks)) + uint64(len(f.Values)) + uint64(len(f.Sources)))
	for _, types := range [][]wasm.ValueType{f.Signature.Params, f.Signature.Results, f.Locals} {
		step(uint64(len(types)))
		for _, typ := range types {
			if !typ.Valid() {
				fail("invalid function or local type")
			}
		}
	}
	if len(f.Locals) < len(f.Signature.Params) {
		fail("missing parameter locals")
	}
	for i, typ := range f.Signature.Params {
		if f.Locals[i] != typ {
			fail("parameter local type")
		}
	}
	for _, typ := range f.Values {
		if !typ.Valid() {
			fail("unknown value type")
		}
	}
	for _, s := range f.Sources {
		if s.Effects != EffectsOf(s.Wasm.Op) {
			fail("incorrect effects")
		}
		for _, ts := range [][]wasm.ValueType{s.Inputs, s.Outputs} {
			step(uint64(len(ts)))
			for _, typ := range ts {
				if !typ.Valid() {
					fail("unknown operation type")
				}
			}
		}
	}
	matches := func(a, b wasm.ValueType) bool {
		if f.TypeHierarchy != nil {
			return f.TypeHierarchy.ReferenceSubtype(a, b, f.TypeLabels)
		}
		return wasm.ReferenceSubtype(a, b, f.TypeLabels)
	}
	defs := make([]definition, len(f.Values))
	pred := make([][]int, len(f.Blocks))
	valueType := func(id Value) wasm.ValueType {
		if uint64(id) >= uint64(len(f.Values)) {
			fail("value index")
		}
		return f.Values[id]
	}
	source := func(index uint32) Source {
		if uint64(index) >= uint64(len(f.Sources)) {
			fail("source index")
		}
		return f.Sources[index]
	}
	define := func(id Value, block, position int) {
		valueType(id)
		if defs[id].exists {
			fail("multiple value definitions")
		}
		defs[id] = definition{block, position, true}
	}
	for bi, b := range f.Blocks {
		step(uint64(len(b.Params)) + uint64(len(b.Ops)) + uint64(len(b.Term.Edges)))
		for _, id := range b.Params {
			define(id, bi, -1)
		}
		for oi, op := range b.Ops {
			s := source(op.Source)
			if s.Wasm.Op != 1 && s.Wasm.Op != 0x10 && s.Wasm.Op != 0x11 && s.Wasm.Op != 0x14 && s.Wasm.Op < 0x1a {
				fail("control instruction in operation list")
			}
			if len(op.Args) != len(s.Inputs) || len(op.Results) != len(s.Outputs) {
				fail("operation arity")
			}
			step(uint64(len(op.Args)) + uint64(len(op.Results)))
			for i, id := range op.Args {
				if !matches(valueType(id), s.Inputs[i]) {
					fail("operation input type")
				}
			}
			for i, id := range op.Results {
				define(id, bi, oi)
				if valueType(id) != s.Outputs[i] {
					fail("operation result type")
				}
			}
		}
		t := b.Term
		s := source(t.Source)
		op := s.Wasm.Op
		switch t.Kind {
		case Jump:
			if len(t.Args) != 0 || len(t.Edges) != 1 {
				fail("jump arity")
			}
			if op != 2 && op != 3 && op != 5 && op != 0xb && op != 0xc && op != 0xf {
				fail("jump source")
			}
		case If, Table:
			if len(t.Args) != 1 || valueType(t.Args[0]) != wasm.I32 {
				fail("branch condition")
			}
			if t.Kind == If && (len(t.Edges) != 2 || op != 4 && op != 0xd) || t.Kind == Table && (len(t.Edges) == 0 || op != 0xe) {
				fail("branch targets/source")
			}
		case IfCast:
			if len(t.Args) != 1 || !valueType(t.Args[0]).IsReference() || len(t.Edges) != 2 || (op != 0xfb0018 && op != 0xfb0019) || len(s.Wasm.Types) != 2 {
				fail("cast branch source/arity")
			}
		case IfNull:
			if len(t.Args) != 1 || !valueType(t.Args[0]).IsReference() || len(t.Edges) != 2 || (op != 0xd5 && op != 0xd6) {
				fail("null branch source/arity")
			}
		case Return:
			if op != 0xb || len(t.Edges) != 0 || len(t.Args) != len(f.Signature.Results) {
				fail("return arity/source")
			}
			for i, id := range t.Args {
				if !matches(valueType(id), f.Signature.Results[i]) {
					fail("return type")
				}
			}
		case TailCall:
			if !module.IsTailCall(op) || len(t.Edges) != 0 || len(t.Args) != len(s.Inputs) || len(s.Outputs) != len(f.Signature.Results) {
				fail("tail call arity/source")
			}
			step(uint64(len(t.Args)) + uint64(len(s.Outputs)))
			for i, typ := range s.Outputs {
				if !matches(typ, f.Signature.Results[i]) {
					fail("tail call return type")
				}
			}
			for i, id := range t.Args {
				if !matches(valueType(id), s.Inputs[i]) {
					fail("tail call operand type")
				}
			}
		case Trap:
			if op != 0 || len(t.Args) != 0 || len(t.Edges) != 0 {
				fail("trap arity/source")
			}
		default:
			fail("missing terminator")
		}
		for ei, edge := range t.Edges {
			if uint64(edge.Target) >= uint64(len(f.Blocks)) {
				fail("block index")
			}
			params := f.Blocks[edge.Target].Params
			if len(edge.Args) != len(params) {
				fail("edge arity")
			}
			step(uint64(len(edge.Args)))
			for i, id := range edge.Args {
				actual := valueType(id)
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
				if !matches(actual, valueType(params[i])) {
					fail("edge type")
				}
			}
			pred[edge.Target] = append(pred[edge.Target], bi)
		}
	}
	if len(f.Blocks[0].Params) != 0 || len(pred[0]) != 0 {
		fail("entry cannot have parameters or predecessors")
	}
	for _, d := range defs {
		if !d.exists {
			fail("undefined value")
		}
	}
	// Iterative DFS produces reverse postorder without consuming host stack.
	type cursor struct{ block, next int }
	seen := make([]bool, len(f.Blocks))
	seen[0] = true
	stack := []cursor{{0, 0}}
	var order []int
	for len(stack) > 0 {
		step(1)
		top := &stack[len(stack)-1]
		edges := f.Blocks[top.block].Term.Edges
		if top.next == len(edges) {
			order = append(order, top.block)
			stack = stack[:len(stack)-1]
			continue
		}
		next := int(edges[top.next].Target)
		top.next++
		if !seen[next] {
			seen[next] = true
			stack = append(stack, cursor{block: next})
		}
	}
	if len(order) != len(f.Blocks) {
		fail("unreachable block")
	}
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	rank := make([]int, len(order))
	idom := make([]int, len(order))
	for i, b := range order {
		rank[b] = i
		idom[b] = -1
	}
	idom[0] = 0
	intersect := func(a, b int) int {
		for a != b {
			step(1)
			if rank[a] > rank[b] {
				a = idom[a]
			} else {
				b = idom[b]
			}
		}
		return a
	}
	for changed := true; changed; {
		changed = false
		for _, b := range order[1:] {
			step(1)
			parent := -1
			for _, p := range pred[b] {
				step(1)
				if idom[p] != -1 {
					if parent == -1 {
						parent = p
					} else {
						parent = intersect(parent, p)
					}
				}
			}
			if parent != idom[b] {
				idom[b] = parent
				changed = true
			}
		}
	}
	use := func(id Value, block, position int) {
		step(1)
		d := defs[id]
		if d.block == block {
			if d.position >= position {
				fail("value used before definition")
			}
			return
		}
		for block != d.block && block != 0 {
			step(1)
			block = idom[block]
		}
		if block != d.block {
			fail("value definition does not dominate use")
		}
	}
	for bi, b := range f.Blocks {
		for oi, op := range b.Ops {
			for _, id := range op.Args {
				use(id, bi, oi)
			}
		}
		for _, id := range b.Term.Args {
			use(id, bi, len(b.Ops))
		}
		for _, edge := range b.Term.Edges {
			for _, id := range edge.Args {
				use(id, bi, len(b.Ops))
			}
		}
	}
	return nil
}

type verifyFailure struct{ error }

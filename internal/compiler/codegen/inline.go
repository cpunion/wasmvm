package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A leaf expression has no observable work except its returned scalar values.
// Limit both the input and the expansion; shared expressions are emitted once.
// Calls, traps, references, memory and control flow are deliberately ineligible.
type leafExpression struct {
	op          uint32
	bits        uint64
	argument    slots.Slot
	left, right *leafExpression
	immediate   bool
}

func leafExpressions(f *slots.Function, inputs []slots.Slot) ([]*leafExpression, bool) {
	if f == nil || len(f.Code) == 0 || len(f.Code) > 32 || len(f.Slots) > 32 ||
		len(f.Slots) < int(f.LocalCount) || len(inputs) > int(f.LocalCount) ||
		len(inputs) != len(f.Signature.Params) || len(f.Signature.Results) > 8 || checkFunction(f) != nil {
		return nil, false
	}
	values := make([]*leafExpression, len(f.Slots))
	arguments := map[slots.Slot]*leafExpression{}
	for i, typ := range f.Slots {
		if typ != wasm.I32 && typ != wasm.I64 {
			return nil, false
		}
		if i < len(inputs) {
			n := arguments[inputs[i]]
			if n == nil {
				n = &leafExpression{argument: inputs[i]}
				arguments[inputs[i]] = n
			}
			values[i] = n
		} else if i < int(f.LocalCount) {
			values[i] = &leafExpression{op: 0x41}
			if typ == wasm.I64 {
				values[i].op = 0x42
			}
		}
	}
	get := func(s slots.Slot) *leafExpression {
		if s == slots.NoSlot {
			return nil
		}
		return values[s]
	}
	for pc, in := range f.Code {
		source := f.Sources[in.Source].Wasm
		var n *leafExpression
		switch op := in.Op; {
		case op == 1 || op == 0x1a:
			continue
		case op == 0x20:
			n = get(slots.Slot(source.A))
		case op == 0x21 || op == 0x22:
			n = get(in.A)
			values[source.A] = n
		case op == 0x41 || op == 0x42:
			n = &leafExpression{op: op, bits: source.Bits}
			if op == 0x41 {
				n.bits = uint64(uint32(n.bits))
			}
		case op >= 0x45 && op <= 0x5a || op >= 0x6a && op <= 0x8a && !scalarCanTrap(op):
			n = &leafExpression{op: op, left: get(in.A), right: get(in.B)}
			if n.left == nil || n.right == nil && op != 0x45 && op != 0x50 {
				return nil, false
			}
		case op == slots.ControlBase+uint32(ir.Jump):
			// Slot lowering uses a forward edge for the function result
			// block. Its already sequenced moves preserve SSA value aliases.
			edge := f.Edges[in.Extra]
			if edge.PC != uint32(pc+1) {
				return nil, false
			}
			for _, move := range f.Moves[edge.Start : edge.Start+edge.Count] {
				if get(move.Src) == nil {
					return nil, false
				}
				values[move.Dst] = get(move.Src)
			}
			continue
		case op == slots.ControlBase+uint32(ir.Return):
			if pc != len(f.Code)-1 || int(in.Count) != len(f.Signature.Results) {
				return nil, false
			}
			result := make([]*leafExpression, in.Count)
			for i, slot := range f.Args[in.Extra : in.Extra+in.Count] {
				result[i] = get(slot)
				if result[i] == nil {
					return nil, false
				}
			}
			return result, true
		default:
			return nil, false
		}
		if n == nil {
			return nil, false
		}
		if in.Dst != slots.NoSlot {
			values[in.Dst] = n
		}
	}
	return nil, false
}

// inlineLeaf emits into a temporary buffer first: register pressure or an
// unsupported instruction leaves the original native call completely intact.
// All original continuations, call polls and caller slots remain authoritative.
func (e *functionEmitter) inlineLeaf(pc int, callee *slots.Function) bool {
	if !e.unmetered || callee == nil {
		return false
	}
	in := e.f.Code[pc]
	inputs := e.f.Args[in.Extra : in.Extra+in.Count]
	results, ok := leafExpressions(callee, inputs)
	if !ok || uint64(in.Extra)+uint64(in.Count)+uint64(len(results)) > uint64(len(e.f.Args)) {
		return false
	}
	for i, slot := range inputs {
		if e.f.Slots[slot] != callee.Signature.Params[i] {
			return false
		}
	}
	outputs := e.f.Args[in.Extra+in.Count : in.Extra+in.Count+uint32(len(results))]
	outputSlots := map[slots.Slot]bool{}
	for i, slot := range outputs {
		if outputSlots[slot] || e.f.Slots[slot] != callee.Signature.Results[i] {
			return false
		}
		outputSlots[slot] = true
	}
	a := assembler{arch: e.a.arch, bounded: true, limit: e.a.limit}
	t := functionEmitter{a: &a, registers: map[slots.Slot]byte{}}
	// These are ordinary allocated registers, never context, frame, SIMD,
	// memory-cache, operand scratch or immutable constant-pool registers.
	pool := []byte{3, 12, 13, 14, 15, 9, 10, 11}
	if e.usesMemory {
		pool = pool[:6]
	}
	if e.a.arch == "arm64" {
		pool = []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28}
	}
	protected := map[byte]bool{}
	for slot, r := range e.registers {
		if !outputSlots[slot] && !e.overwrittenBeforeExit(pc+1, slot) {
			protected[r] = true
		}
	}
	uses, regs := map[*leafExpression]int{}, map[*leafExpression]byte{}
	seen := map[*leafExpression]bool{}
	var order []*leafExpression
	var visit func(*leafExpression)
	visit = func(n *leafExpression) {
		uses[n]++
		if seen[n] {
			return
		}
		seen[n] = true
		if n.left != nil {
			visit(n.left)
			if n.right != nil {
				if n.right.op == 0x41 || n.right.op == 0x42 {
					probe := functionEmitter{a: &assembler{arch: a.arch}, registers: map[slots.Slot]byte{0: pool[0], 1: pool[1]}}
					n.immediate = probe.registerImmediate(n.op, 0, 1, n.right.bits)
				}
				if !n.immediate {
					visit(n.right)
				}
			}
		}
		order = append(order, n)
	}
	for _, result := range results {
		visit(result)
	}
	// Install all argument locations before allocation, even if their first use
	// follows another expression. Aliased arguments share an expression node.
	for _, n := range order {
		if n.op == 0 {
			r, ok := e.registers[n.argument]
			if !ok {
				return false
			}
			regs[n] = r
		}
	}
	for _, n := range order {
		if n.op == 0 {
			continue
		}
		if n.op == 0x41 || n.op == 0x42 {
			if r, ok := e.constantRegisters[n.bits]; ok {
				regs[n] = r
				continue
			}
		}
		consumed := map[*leafExpression]int{}
		if n.left != nil {
			consumed[n.left]++
		}
		if n.right != nil && !n.immediate {
			consumed[n.right]++
		}
		selected := false
		for _, r := range pool {
			if protected[r] {
				continue
			}
			busy := false
			for v, vr := range regs {
				if vr == r && uses[v] > consumed[v] {
					busy = true
				}
			}
			if busy {
				continue
			}
			t.registers = map[slots.Slot]byte{0: r}
			if n.left != nil {
				t.registers[1] = regs[n.left]
			}
			if n.right != nil && !n.immediate {
				t.registers[2] = regs[n.right]
			}
			if n.op == 0x41 || n.op == 0x42 {
				if a.arch == "arm64" {
					a.aconst(uint32(r), n.bits, n.bits > 0xffffffff)
				} else {
					a.xconst(r, n.bits)
				}
			} else if n.immediate {
				if !t.registerImmediate(n.op, 0, 1, n.right.bits) {
					continue
				}
			} else if !t.registerInteger(n.op, 0, 1, 2) {
				continue
			}
			regs[n], selected = r, true
			break
		}
		if !selected {
			return false
		}
		for v, count := range consumed {
			uses[v] -= count
		}
	}
	// Publish simultaneous results without clobbering another pending result.
	// A register cycle is left to the ordinary native call path for now.
	pending := append([]*leafExpression(nil), results...)
	writer := *e
	writer.a = &a
	for remaining := len(pending); remaining != 0; {
		progress := false
		for i, n := range pending {
			if n == nil {
				continue
			}
			dst, register := e.registers[outputs[i]]
			busy := false
			for j, other := range pending {
				if j != i && other != nil && register && regs[other] == dst && regs[n] != dst {
					busy = true
				}
			}
			if busy {
				continue
			}
			if register {
				writer.registerMove(dst, regs[n], true)
			} else {
				writer.registerMove(writer.operandRegister(0), regs[n], true)
				writer.slotStore(outputs[i], 0, false)
			}
			pending[i], progress = nil, true
			remaining--
		}
		if !progress {
			return false
		}
	}
	e.a.bytes(a.code...)
	return true
}

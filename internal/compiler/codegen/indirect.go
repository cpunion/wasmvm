package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
)

// RuntimeData is an owned uint64 array: function references in module order,
// then (element count, native base) for each table. The driver refreshes table
// bindings before every entry. Only proven matching identities enter a leaf;
// all other targets, nulls and bounds failures use the ordinary checked call.
func (e *functionEmitter) inlineIndirect(pc int, config Config) (int, uint32, bool) {
	in := e.f.Code[pc]
	source := e.f.Sources[in.Source]
	offset := uint64(config.FunctionCount)*8 + uint64(source.Wasm.B)*16
	if !e.unmetered || in.Count == 0 || config.FunctionCount == 0 || offset > math.MaxInt32-8 {
		return pc, 0, false
	}
	type candidate struct {
		index uint32
		code  []byte
		label int
	}
	var candidates []candidate
	end, needed := pc, uint32(0)
	for _, index := range config.IndirectCallees[source.Wasm.A] {
		if len(candidates) == 4 {
			break
		}
		callee := config.DirectCallees[index]
		if index >= config.FunctionCount || callee == nil || len(callee.Signature.Results) != len(source.Outputs) {
			continue
		}
		matching := true
		for i, typ := range source.Outputs {
			matching = matching && callee.Signature.Results[i] == typ
		}
		if !matching {
			continue
		}
		a := assembler{arch: e.a.arch, bounded: true, limit: e.a.limit}
		t := *e
		t.a = &a
		last, ok := t.inlineLeaf(pc, callee)
		if !ok || len(candidates) != 0 && last != end {
			continue
		}
		end, needed = last, max(needed, uint32(len(callee.Slots)))
		candidates = append(candidates, candidate{index, a.code, e.a.label()})
	}
	if len(candidates) == 0 {
		return pc, 0, false
	}
	fallback, done := e.a.label(), e.a.label()
	e.contextLoad(0, nativeabi.RuntimeDataOffset)
	e.test(true)
	e.branch(fallback, "eq")
	e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
	e.slotLoad(e.f.Args[in.Extra+in.Count-1], 0, false, false)
	e.pointerLoad(1, 2, uint32(offset))
	e.compare(true)
	e.branch(fallback, "hs")
	e.pointerLoad(1, 2, uint32(offset+8))
	if e.a.arch == "arm64" {
		e.a.word(0xf8697949) // ldr x9, [x10, x9, lsl #3]
	} else {
		e.a.bytes(0x48, 0x8b, 0x04, 0xc1) // mov rax, [rcx + rax*8]
	}
	for _, c := range candidates {
		e.pointerLoad(1, 2, c.index*8)
		e.compare(true)
		e.branch(c.label, "eq")
	}
	e.a.branch(fallback, false, 0)
	for _, c := range candidates {
		e.a.mark(c.label)
		e.a.bytes(c.code...)
		e.a.branch(done, false, 0)
	}
	e.a.mark(fallback)
	e.contextConst(nativeabi.SourceOffset, uint64(source.Wasm.Offset))
	e.leave(nativeabi.Call, uint64(pc), uint64(pc+1))
	e.a.mark(done)
	return end, needed, true
}

// Load a word from a trusted runtime allocation using operand scratch registers.
// Guest indexing is emitted separately and only after the length comparison.
func (e *functionEmitter) pointerLoad(dst, base byte, offset uint32) {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x8b, 0x80|dst<<3|base)
		e.a.word(offset)
		return
	}
	r := uint32(e.operandRegister(base))
	if offset/8 >= 4096 {
		e.a.aconst(15, uint64(offset), false)
		e.a.word(0x8b0f000f | r<<5) // add x15, base, x15
		r, offset = 15, 0
	}
	e.a.word(0xf9400000 | (offset/8)<<10 | r<<5 | uint32(e.operandRegister(dst)))
}

package decode

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (d *decoder) blockType() int64 {
	saved := d.r
	b := d.byte()
	if b == 0x40 {
		return -64
	}
	if (b == 0x63 || b == 0x64) && d.features&(wasm.FunctionReferences|wasm.GC|wasm.ExceptionHandling) != 0 {
		t := d.heapReference(b == 0x63)
		return int64(uint64(t) | 1<<63)
	}
	if wasm.ValueType(b).Valid() && d.referenceEnabled(wasm.ValueType(b)) {
		return int64(int8(b | 0x80))
	}
	d.r = saved
	t := must(d.r.S33())
	if t < 0 {
		d.fail("invalid block type")
	}
	return t
}

func (d *decoder) expr() module.Expr {
	// An explicit syntax stack keeps untrusted nesting off the host call stack.
	// 4 marks an if before else; 5 marks an if after else.
	stack := []byte{0}
	var code module.Expr
	for {
		d.charge(1)
		in := module.Instr{Offset: d.r.Offset(), Op: uint32(d.byte())}
		switch op := in.Op; {
		case op == 2 || op == 3 || op == 4 || op == 0x1f && d.features&wasm.ExceptionHandling != 0:
			in.Block = d.blockType()
			if op == 0x1f {
				in.Catches = make([]module.Catch, d.count())
				for i := range in.Catches {
					c := &in.Catches[i]
					c.Kind = d.byte()
					if c.Kind > 3 {
						d.fail("unknown catch kind")
					}
					if c.Kind < 2 {
						c.Tag = d.u32()
					}
					c.Label = d.u32()
				}
			}
			if uint64(len(stack)) > uint64(d.limits.MaxNesting) {
				d.limit("control nesting")
			}
			d.charge(1)
			stack = append(stack, byte(op))
		case op == 5:
			if stack[len(stack)-1] != 4 {
				d.failAt(in.Offset, "unexpected else")
			}
			stack[len(stack)-1] = 5
		case op == 0x0b:
			stack = stack[:len(stack)-1]
		case op == 0x0c || op == 0x0d || op == 0x10 || op >= 0x20 && op <= 0x26 || op == 0xd2:
			in.A = d.u32()
		case op == 0x08 && d.features&wasm.ExceptionHandling != 0:
			in.A = d.u32()
		case op == 0x0a && d.features&wasm.ExceptionHandling != 0:
		case op == 0x0e:
			in.Labels, in.A = d.indexes(), d.u32()
		case op == 0x11:
			in.A, in.B = d.u32(), d.u32()
		case (op == 0x12 || op == 0x13) && d.features&wasm.TailCall != 0:
			in.A = d.u32()
			if op == 0x13 {
				in.B = d.u32()
			}
		case (op == 0x14 || op == 0x15 || op == 0xd5 || op == 0xd6) && d.features&(wasm.FunctionReferences|wasm.GC) != 0:
			if op == 0x15 && d.features&wasm.TailCall == 0 {
				d.failAt(in.Offset, "tail calls are disabled")
			}
			in.A = d.u32()
		case op == 0xd4 && d.features&(wasm.FunctionReferences|wasm.GC) != 0:
		case op == 0x1c:
			in.Types = d.types()
		case op >= 0x28 && op <= 0x3e:
			d.memarg(&in)
		case op == 0x3f || op == 0x40:
			in.A = d.memoryIndex()
		case op == 0x41:
			in.Bits = uint64(uint32(must(d.r.S32())))
		case op == 0x42:
			in.Bits = uint64(must(d.r.S64()))
		case op == 0x43:
			in.Bits = uint64(must(d.r.F32Bits()))
		case op == 0x44:
			in.Bits = must(d.r.F64Bits())
		case op == 0xd0:
			t := d.heapReference(true)
			in.A, in.Bits = uint32(t), uint64(t)
		case op == 0xd3 && d.features&wasm.GC != 0:
		case op == 0xfb && d.features&wasm.GC != 0:
			d.gcInstruction(&in)
		case op == 0xfc:
			sub := d.u32()
			in.Op = op<<16 | sub
			switch {
			case sub <= 7:
			case sub == 8:
				in.A = d.u32()
				in.B = d.memoryIndex()
				d.needsDataCount = d.needsDataCount || d.inCode
			case sub == 9:
				in.A = d.u32()
				d.needsDataCount = d.needsDataCount || d.inCode
			case sub == 10:
				in.A, in.B = d.memoryIndex(), d.memoryIndex()
			case sub == 11:
				in.A = d.memoryIndex()
			case sub == 12 || sub == 14:
				in.A, in.B = d.u32(), d.u32()
			case sub >= 13 && sub <= 17:
				in.A = d.u32()
			default:
				d.failAt(in.Offset, "unknown miscellaneous opcode")
			}
		case op == 0xfd:
			sub := d.u32()
			relaxed := sub >= 0x100 && sub <= 0x113 && d.features&wasm.RelaxedSIMD != 0
			if !relaxed && (sub > 255 || simdOpcodes[sub/64]&(uint64(1)<<(sub%64)) == 0) {
				d.failAt(in.Offset, "unknown SIMD opcode")
			}
			in.Op = op<<16 | sub
			switch {
			case sub <= 11 || sub == 92 || sub == 93:
				d.memarg(&in)
			case sub == 12 || sub == 13:
				copy(in.Vector[:], d.bytes(16))
			case sub >= 21 && sub <= 34:
				in.Lane = d.byte()
			case sub >= 84 && sub <= 91:
				d.memarg(&in)
				in.Lane = d.byte()
			}
		case op == 0 || op == 1 || op == 0x0f || op == 0x1a || op == 0x1b || op == 0xd1 || op >= 0x45 && op <= 0xc4:
		default:
			d.failAt(in.Offset, "unknown opcode")
		}
		code = append(code, in)
		if len(stack) == 0 {
			return code
		}
	}
}

func (d *decoder) memoryIndex() uint32 {
	if d.features&wasm.MultiMemory == 0 {
		d.expect(0)
		return 0
	}
	return d.u32()
}

func (d *decoder) memarg(in *module.Instr) {
	at := d.r.Offset()
	align := d.u32()
	var memory uint32
	if d.features&wasm.MultiMemory != 0 && align&64 != 0 {
		align &^= 64
		memory = d.u32()
	}
	if align >= 32 {
		d.failAt(at, "malformed memory alignment flags")
	}
	in.A, in.Memory = align, memory
	if d.features&wasm.Memory64 != 0 {
		offset := must(d.r.U64())
		in.B, in.MemoryOffsetHigh = uint32(offset), uint32(offset>>32)
	} else {
		in.B = d.u32()
	}
}

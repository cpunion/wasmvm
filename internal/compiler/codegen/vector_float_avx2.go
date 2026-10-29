package codegen

func (e *functionEmitter) vectorFloatAVX2(op uint32, wide bool, d, l, r byte) {
	prefix := byte(0)
	if wide {
		prefix = 1
	}
	switch {
	case op <= 0x60:
		kind := op - 0x5b
		if kind == 3 || kind == 5 {
			l, r = r, l
		}
		e.a.xvex(1, prefix, 0xc2, d, l, r, false)
		e.a.bytes([...]byte{0, 4, 1, 1, 2, 2}[kind])
	case op == 0x8b || op == 0x8c:
		e.vectorFloatMaskAMD64(11, wide, op == 0x8c)
		code := byte(0xdb)
		if op == 0x8c {
			code = 0xef
		}
		e.a.xvex(1, 1, code, d, l, 11, false)
	case op >= 0x8d && op <= 0x90:
		code := byte(0x08)
		if wide {
			code++
		}
		e.a.xvex(3, 1, code, d, 0, l, false)
		e.a.bytes(8 | [...]byte{2, 1, 3, 0}[op-0x8d])
	case op == 0x91:
		e.a.xvex(1, prefix, 0x51, d, 0, l, false)
	case op >= 0x96:
		e.vectorMinMaxAVX2(op, wide, d, l, r)
	default:
		e.a.xvex(1, prefix, [...]byte{0x58, 0x5c, 0x59, 0x5e}[op-0x92], d, l, r, false)
	}
}

func (e *functionEmitter) vectorMinMaxAVX2(op uint32, wide bool, d, l, r byte) {
	prefix, blend := byte(0), byte(0x4a)
	if wide {
		prefix, blend = 1, 0x4b
	}
	if op >= 0x98 {
		left, right := r, l
		if op == 0x99 {
			left, right = l, r
		}
		e.a.xvex(1, prefix, 0xc2, 12, left, right, false)
		e.a.bytes(1)
		e.a.xvex(3, 1, blend, d, l, r, false)
		e.a.bytes(12 << 4)
		return
	}
	code, combine := byte(0x5d), byte(0xeb)
	if op == 0x97 {
		code, combine = 0x5f, 0xdb
	}
	e.a.xvex(1, prefix, code, 2, l, r, false)
	e.a.xvex(1, prefix, code, 12, r, l, false)
	e.a.xvex(1, 1, combine, 2, 2, 12, false)
	e.a.xvex(1, prefix, 0xc2, 12, l, r, false)
	e.a.bytes(3)
	bits := uint64(0x7fc000007fc00000)
	if wide {
		bits = 0x7ff8000000000000
	}
	constant, ok := e.vectorConstants[[2]uint64{bits, bits}]
	if !ok {
		e.vectorFloatConstant(11, wide, 0x7fc00000, 0x7ff8000000000000)
		constant = 11
	}
	e.a.xvex(3, 1, blend, d, 2, constant, false)
	e.a.bytes(12 << 4)
}

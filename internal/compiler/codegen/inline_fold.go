package codegen

import "github.com/cpunion/wasmvm/internal/core/constant"

// foldLeafExpression only rewrites total integer expressions. Constant control
// conditions can then be interpreted by the bounded leaf planner. Preserve the
// original nodes: loop locals and repeated callees may still reference them.
func foldLeafExpression(n *leafExpression) *leafExpression {
	left, right := n.left, n.right
	if left == nil || right == nil {
		return n
	}
	literal := func(v *leafExpression) bool { return v.op == 0x41 || v.op == 0x42 }
	if literal(left) && literal(right) {
		switch n.op {
		case 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e:
			return &leafExpression{op: left.op, bits: constant.Binary(n.op, left.bits, right.bits)}
		case 0x46, 0x47, 0x49, 0x4b, 0x4d, 0x4f, 0x51, 0x52, 0x54, 0x56, 0x58, 0x5a:
			a, b := left.bits, right.bits
			if n.op <= 0x4f {
				a, b = uint64(uint32(a)), uint64(uint32(b))
			}
			var result bool
			switch n.op {
			case 0x46, 0x51:
				result = a == b
			case 0x47, 0x52:
				result = a != b
			case 0x49, 0x54:
				result = a < b
			case 0x4b, 0x56:
				result = a > b
			case 0x4d, 0x58:
				result = a <= b
			case 0x4f, 0x5a:
				result = a >= b
			}
			bits := uint64(0)
			if result {
				bits = 1
			}
			return &leafExpression{op: 0x41, bits: bits}
		}
	}
	add := uint32(0x6a)
	if n.op == 0x7c || n.op == 0x7e {
		add = 0x7c
	}
	if n.op != add && n.op != add+2 {
		return n
	}
	if literal(left) && !literal(right) {
		left, right = right, left
		n = &leafExpression{op: n.op, left: left, right: right}
	}
	if !literal(right) || left.right == nil || !literal(left.right) {
		return n
	}
	if left.op == n.op {
		c := &leafExpression{op: right.op, bits: constant.Binary(n.op, left.right.bits, right.bits)}
		return &leafExpression{op: n.op, left: left.left, right: c}
	}
	if n.op == add+2 && left.op == add {
		product := foldLeafExpression(&leafExpression{op: n.op, left: left.left, right: right})
		bias := &leafExpression{op: right.op, bits: constant.Binary(n.op, left.right.bits, right.bits)}
		return &leafExpression{op: add, left: product, right: bias}
	}
	return n
}

package lower

import (
	"github.com/cpunion/wasmvm/internal/core/constant"
	"github.com/cpunion/wasmvm/internal/core/module"
)

// FoldedConstant returns a compiled literal only for an entirely static,
// validated initializer. Imported globals remain instance-dependent.
func (m *Module) FoldedConstant(expr module.Expr) (uint64, bool) {
	if len(expr) <= 2 {
		return 0, false
	}
	x, ok := m.constants[&expr[0]]
	return x, ok
}

func (m *Module) foldConstants(b *budget) {
	var local [16]uint64
	stack := local[:0]
	fold := func(expr module.Expr) {
		if len(expr) <= 2 {
			return
		}
		b.charge(uint64(len(expr)))
		stack = stack[:0]
		for i := range expr[:len(expr)-1] {
			x := &expr[i]
			switch x.Op {
			case 0x41, 0x42:
				stack = append(stack, x.Bits)
			case 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e:
				n := len(stack)
				stack[n-2] = constant.Binary(x.Op, stack[n-2], stack[n-1])
				stack = stack[:n-1]
			default:
				return
			}
		}
		if m.constants == nil {
			m.constants = make(map[*module.Instr]uint64)
		}
		m.constants[&expr[0]] = stack[0]
	}
	d := m.Validated.Definition()
	for _, g := range d.Globals {
		fold(g.Init)
	}
	for _, e := range d.Elements {
		fold(e.Offset)
		for _, expr := range e.Expressions {
			fold(expr)
		}
	}
	for _, data := range d.Data {
		fold(data.Offset)
	}
}

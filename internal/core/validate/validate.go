// Package validate checks the complete Core 2.0 static semantics. It accepts
// owned decoded modules; validation never instantiates or executes guest code.
package validate

import (
	"math"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct{ MaxLocals, MaxStack, MaxControl uint32 }

func DefaultLimits() Limits { return Limits{1 << 20, 1 << 20, 10000} }

// Module is produced only by successful validation. Its definition and all
// borrowed type slices must remain immutable after Check returns.
type Module struct {
	definition *module.Module
	functions  []FunctionInfo
	types      context
}
type FunctionInfo struct{ MaxStack, MaxControl uint32 }

// ConstantStack is the maximum validated initializer operand depth.
func (m *Module) ConstantStack() uint32              { return m.types.constantStack }
func (m *Module) TypeLabels() []uint32               { return m.types.typeLabels }
func (m *Module) TypeHierarchy() *wasm.TypeHierarchy { return &m.types.registry.Hierarchy }

func (m *Module) Definition() *module.Module { return m.definition }
func (m *Module) FunctionInfo(index uint32) (FunctionInfo, bool) {
	if uint64(index) >= uint64(len(m.functions)) {
		return FunctionInfo{}, false
	}
	return m.functions[index], true
}

type context struct {
	m               *module.Module
	limits          Limits
	features        wasm.Features
	constantStack   uint32
	funcs           []module.FuncType
	funcTypeIndices []uint32
	typeLabels      []uint32
	registry        module.TypeRegistry
	tables          []module.TableType
	memories        []module.Limits
	globals         []module.GlobalType
	importGlobals   int
	refs            map[uint32]bool
	at              uint64
	visitor         func(int, module.Instr, []wasm.ValueType, []wasm.ValueType) error
}

// Check validates a decoded module without modifying it. The caller transfers
// exclusive mutation rights to the resulting immutable module on success.
func Check(m *module.Module, limits Limits) (*Module, error) {
	return CheckWithFeatures(m, limits, 0)
}

// CheckWithFeatures enables only explicitly selected, implemented extensions.
func CheckWithFeatures(m *module.Module, limits Limits, features wasm.Features) (result *Module, err error) {
	if features & ^wasm.SupportedFeatures != 0 {
		return nil, wasm.ErrUnsupported
	}
	defer func() {
		if failure := recover(); failure != nil {
			if e, ok := failure.(*wasm.Error); ok {
				result, err = nil, e
			} else {
				panic(failure)
			}
		}
	}()
	c := context{m: m, limits: limits, features: features, refs: make(map[uint32]bool)}
	if m == nil {
		c.fail("nil module")
	}
	c.checkTypes()
	// Validation retains normalized identities and ancestry, not a second copy
	// of every decoded composite shape.
	c.registry.Types = nil
	c.registry.Keys = nil
	for _, im := range m.Imports {
		switch im.Kind {
		case 0:
			c.funcs = append(c.funcs, c.typ(im.Type))
			c.funcTypeIndices = append(c.funcTypeIndices, im.Type)
		case 1:
			c.tables = append(c.tables, im.Table)
		case 2:
			c.memories = append(c.memories, im.Memory)
		case 3:
			c.globals = append(c.globals, im.Global)
		default:
			c.fail("invalid import kind")
		}
	}
	c.importGlobals = len(c.globals)
	for _, f := range m.Functions {
		c.funcs = append(c.funcs, c.typ(f.Type))
		c.funcTypeIndices = append(c.funcTypeIndices, f.Type)
	}
	c.tables = append(c.tables, m.Tables...)
	c.memories = append(c.memories, m.Memories...)
	for _, g := range m.Globals {
		c.globals = append(c.globals, g.Type)
	}
	if len(c.memories) > 1 && c.features&wasm.MultiMemory == 0 {
		c.fail("multiple memories are not in Core 2.0")
	}
	for _, mt := range c.memories {
		max := uint64(65536)
		if mt.Index64 {
			max = 1 << 48
		}
		c.bounds(mt, max)
	}
	for _, tt := range c.tables {
		c.valueType(tt.Element, len(m.Types))
		if !tt.Element.IsReference() {
			c.fail("invalid table element type")
		}
		max := uint64(math.MaxUint32)
		if tt.Limits.Index64 {
			max = math.MaxUint64
		}
		c.bounds(tt.Limits, max)
	}
	for _, gt := range c.globals {
		c.valueType(gt.Value, len(m.Types))
	}
	names := make(map[string]bool)
	for _, e := range m.Exports {
		if names[e.Name] {
			c.fail("duplicate export name")
		}
		names[e.Name] = true
		switch e.Kind {
		case 0:
			c.function(e.Index)
			c.refs[e.Index] = true
		case 1:
			c.table(e.Index)
		case 2:
			c.memory(e.Index)
		case 3:
			c.global(e.Index)
		default:
			c.fail("invalid export kind")
		}
	}
	declare := func(expr module.Expr) {
		for _, in := range expr {
			if in.Op == 0xd2 {
				c.refs[in.A] = true
			}
		}
	}
	for _, g := range m.Globals {
		declare(g.Init)
	}
	for _, init := range m.TableInit {
		declare(init)
	}
	for _, e := range m.Elements {
		declare(e.Offset)
		for _, idx := range e.Functions {
			c.refs[idx] = true
		}
		for _, init := range e.Expressions {
			declare(init)
		}
	}
	for _, data := range m.Data {
		declare(data.Offset)
	}
	for _, g := range m.Globals {
		c.constant(g.Init, g.Type.Value)
	}
	if m.TableInit != nil && len(m.TableInit) != len(m.Tables) {
		c.fail("table initializer count")
	}
	for i, table := range m.Tables {
		if m.TableInit != nil && m.TableInit[i] != nil {
			if features&(wasm.FunctionReferences|wasm.GC) == 0 {
				c.fail("table initializers are disabled")
			}
			c.constant(m.TableInit[i], table.Element)
		} else if !table.Element.Defaultable() {
			c.fail("nondefaultable table needs an initializer")
		}
	}
	for _, e := range m.Elements {
		if !e.Type.IsReference() {
			c.fail("invalid element reference type")
		}
		c.valueType(e.Type, len(m.Types))
		if e.Mode == module.Active {
			if !c.matches(e.Type, c.table(e.Table).Element) {
				c.fail("element/table type mismatch")
			}
			c.constant(e.Offset, c.table(e.Table).Limits.AddressType())
		} else if e.Mode != module.Passive && e.Mode != module.Declarative {
			c.fail("invalid element mode")
		}
		for _, idx := range e.Functions {
			c.function(idx)
			if !c.matches(c.functionReference(idx), e.Type) {
				c.fail("function element type mismatch")
			}
		}
		for _, init := range e.Expressions {
			c.constant(init, e.Type)
		}
	}
	for _, data := range m.Data {
		if data.Mode == module.Active {
			c.memory(data.Memory)
			c.constant(data.Offset, c.memories[data.Memory].AddressType())
		} else if data.Mode != module.Passive {
			c.fail("invalid data mode")
		}
	}
	if m.HasStart {
		sig := c.function(m.Start)
		if len(sig.Params) != 0 || len(sig.Results) != 0 {
			c.fail("start function must have no parameters or results")
		}
	}
	result = &Module{definition: m, functions: make([]FunctionInfo, len(m.Functions))}
	for i := range m.Functions {
		result.functions[i] = c.body(&m.Functions[i])
	}
	result.types = c
	return result, nil
}

type visitorFailure struct{ err error }

// WalkFunction replays the validator's operand signatures without introducing
// a second opcode typing table in the compiler. Input types are in stack order;
// output types are in result order. Control instructions have no signature
// here: lowering uses their validated block/label types. Slices are borrowed
// until the next callback. Unreachable polymorphic values may have type zero;
// they must not become executable values. A visitor error stops immediately.
func (m *Module) WalkFunction(index uint32, visitor func(int, module.Instr, []wasm.ValueType, []wasm.ValueType) error) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			switch e := failure.(type) {
			case visitorFailure:
				err = e.err
			case *wasm.Error:
				err = e
			default:
				panic(failure)
			}
		}
	}()
	c := m.types
	c.index(index, len(m.functions), "defined function")
	c.visitor = visitor
	c.body(&m.definition.Functions[index])
	return nil
}

func (c *context) fail(detail string) {
	panic(&wasm.Error{Kind: wasm.ErrInvalid, Offset: c.at, Detail: detail})
}
func (c *context) limit(detail string) {
	panic(&wasm.Error{Kind: wasm.ErrLimit, Offset: c.at, Detail: detail})
}
func (c *context) bounds(l module.Limits, max uint64) {
	if l.Index64 && c.features&wasm.Memory64 == 0 {
		c.fail("64-bit indices are disabled")
	}
	if l.Min > max || l.HasMax && (l.Max > max || l.Min > l.Max) {
		c.fail("invalid limits")
	}
}
func (c *context) index(index uint32, length int, kind string) {
	if uint64(index) >= uint64(length) {
		c.fail("unknown " + kind)
	}
}
func (c *context) typ(index uint32) module.FuncType {
	c.index(index, len(c.m.Types), "type")
	if c.m.TypeDef(index).Kind != module.FunctionKind {
		c.fail("function type required")
	}
	return c.m.Types[index]
}
func (c *context) function(index uint32) module.FuncType {
	c.index(index, len(c.funcs), "function")
	return c.funcs[index]
}
func (c *context) table(index uint32) module.TableType {
	c.index(index, len(c.tables), "table")
	return c.tables[index]
}
func (c *context) memory(index uint32) { c.index(index, len(c.memories), "memory") }
func (c *context) global(index uint32) module.GlobalType {
	c.index(index, len(c.globals), "global")
	return c.globals[index]
}
func (c *context) constant(expr module.Expr, want wasm.ValueType) {
	if len(expr) < 2 || expr[len(expr)-1].Op != 0x0b {
		c.fail("constant expression must produce one value")
	}
	var local [16]wasm.ValueType
	stack := local[:0]
	for i := range expr[:len(expr)-1] {
		in := &expr[i]
		c.at = in.Offset
		var got wasm.ValueType
		switch in.Op {
		case 0x41:
			got = wasm.I32
		case 0x42:
			got = wasm.I64
		case 0x43:
			got = wasm.F32
		case 0x44:
			got = wasm.F64
		case 0xfd000c:
			got = wasm.V128
		case 0x23:
			c.index(in.A, c.importGlobals, "constant imported global")
			g := c.globals[in.A]
			if g.Mutable {
				c.fail("constant global must be immutable")
			}
			got = g.Value
		case 0xd0:
			got = in.ReferenceType()
			c.valueType(got, len(c.m.Types))
			if !got.IsReference() {
				c.fail("invalid null reference type")
			}
		case 0xd2:
			c.function(in.A)
			got = c.functionReference(in.A)
		case 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e:
			if c.features&wasm.ExtendedConst == 0 {
				c.fail("extended constant expressions are disabled")
			}
			got = wasm.I32
			if in.Op >= 0x7c {
				got = wasm.I64
			}
			if len(stack) < 2 || stack[len(stack)-1] != got || stack[len(stack)-2] != got {
				c.fail("constant operand type mismatch")
			}
			stack = stack[:len(stack)-1]
			continue
		case 0xfb0000, 0xfb0001, 0xfb0006, 0xfb0007, 0xfb0008, 0xfb001a, 0xfb001b, 0xfb001c:
			v := checker{context: c, stack: stack, control: []frame{{}}}
			v.gc(*in)
			stack = v.stack
			if uint32(len(stack)) > c.constantStack {
				c.constantStack = uint32(len(stack))
			}
			continue
		default:
			c.fail("nonconstant initializer")
		}
		if uint64(len(stack)) >= uint64(c.limits.MaxStack) {
			panic(&wasm.Error{Kind: wasm.ErrLimit, Offset: c.at, Detail: "constant operand stack"})
		}
		stack = append(stack, got)
		if uint32(len(stack)) > c.constantStack {
			c.constantStack = uint32(len(stack))
		}
	}
	if len(stack) != 1 || !c.matches(stack[0], want) {
		c.fail("initializer type mismatch")
	}
}

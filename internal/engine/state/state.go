// Package state owns Core 2.0 instance state and Store-local reference domains.
// Its mutable operations run in a serialized Store execution context; the
// embedding layer supplies concurrency, lifecycle and host-root management.
package state

import (
	"context"
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	constops "github.com/cpunion/wasmvm/internal/core/constant"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var ErrLink = errors.New("WebAssembly link error")
var ErrReference = errors.New("foreign or stale WebAssembly reference")

type Limits struct {
	MaxTypes                                                                 uint32
	MaxHeapObjects                                                           uint32
	MaxHeapBytes                                                             uint64
	MaxTags                                                                  uint32
	MaxTables, MaxMemories                                                   uint32
	MaxFunctions, MaxGlobals, MaxInstances, MaxMemoryPages, MaxTableElements uint32
	MaxSegmentBytes, MaxSegmentElements                                      uint64
}

func DefaultLimits() Limits {
	return Limits{MaxTags: 1 << 20, MaxHeapObjects: 1 << 20, MaxHeapBytes: 64 << 20, MaxTypes: 1 << 20, MaxTables: 1 << 16, MaxMemories: 1024, MaxFunctions: 1 << 20, MaxGlobals: 1 << 20, MaxInstances: 1024, MaxMemoryPages: 4096, MaxTableElements: 1 << 20, MaxSegmentBytes: 64 << 20, MaxSegmentElements: 1 << 20}
}

type Usage struct {
	SharedMemoryPages, ReservedSharedMemoryPages              uint64
	HeapObjects, HeapBytes                                    uint64
	Tags                                                      uint64
	Tables, Memories                                          uint64
	Functions, Globals, Instances, MemoryPages, TableElements uint64
	SegmentBytes, SegmentElements                             uint64
}
type Store struct {
	closed                    bool
	Domain                    *memory.Domain
	reservedShared            uint64
	nativeMemory              bool
	limits                    Limits
	usage                     Usage
	functions                 []*Function // index zero is the null function reference
	instances                 []*Instance
	tables                    []*Table
	memories                  []*Memory
	tags                      []*Tag
	globals                   []*Global
	registry                  module.TypeRegistry
	heap                      []heapSlot
	heapLayouts               map[uint32]*heapLayout
	heapFree                  []uint32
	heapSerial, heapThreshold uint64
	temporary                 []ReferenceValue
	allocating                []*Instance
	HeapSafepoint             func()
	CollectEveryAllocation    bool
}

func New(limits Limits) *Store {
	s := &Store{limits: limits, functions: []*Function{nil}, heapThreshold: 1 << 20}
	s.Domain = memory.NewDomain(1<<16, native.Available)
	s.registry.Init()
	return s
}

// NewNative selects off-heap memories and tables for native execution.
func NewNative(limits Limits) *Store { s := New(limits); s.nativeMemory = true; return s }
func (s *Store) RetainedModules(visit func(*slots.Module)) {
	for _, in := range s.instances {
		visit(in.Compiled)
	}
	for _, in := range s.allocating {
		visit(in.Compiled)
	}
}
func (s *Store) Usage() Usage {
	u := s.usage
	u.ReservedSharedMemoryPages = s.reservedShared
	for _, m := range s.memories {
		if m.Shared() {
			u.SharedMemoryPages += m.Pages()
		}
	}
	u.MemoryPages += u.SharedMemoryPages
	return u
}
func (s *Store) TypeID(t module.FuncType) uint32 {
	if s.closed {
		return 0
	}
	m := &module.Module{Types: []module.FuncType{t}}
	ids, ok := s.registry.InternModule(m, s.limits.MaxTypes)
	if !ok {
		return 0
	}
	return ids[0]
}

// Host requests are returned by the interpreter to its driver. Guest calls do
// not invoke Go callbacks or recurse through the Go stack inside dispatch.
type Function struct {
	Owner     *Store
	Signature module.FuncType
	TypeID    uint32
	Instance  *Instance
	Code      *slots.Function
	Index     uint32
	Host      any
}
type Global struct {
	OwnerInstance *Instance
	Owner         *Store
	Type          module.GlobalType
	Value         value.Value
}
type Memory struct {
	OwnerInstance *Instance
	Owner         *Store
	*memory.Memory
}
type Table struct {
	OwnerInstance *Instance
	Owner         *Store
	Type          module.TableType
	Values        []uint64
	maximum       uint64
	native        *native.Memory
}
type Extern struct {
	Owner    *Store
	Kind     byte
	Function uint64
	Table    *Table
	Memory   *Memory
	Global   *Global
	Tag      *Tag
}
type Instance struct {
	published, disposed bool
	initializerContext  context.Context
	Owner               *Store
	Compiled            *slots.Module
	Types               []uint32
	Functions           []uint64
	Tables              []*Table
	Memories            []*Memory
	Globals             []*Global
	Tags                []*Tag
	Elements            [][]uint64
	Data                [][]byte
	Exports             map[string]Extern
}

func (s *Store) Function(ref uint64) (*Function, bool) {
	if ref == 0 || ref >= uint64(len(s.functions)) || s.functions[ref] == nil {
		return nil, false
	}
	return s.functions[ref], true
}
func (s *Store) NewHost(t module.FuncType, host any, referenceType ...wasm.ValueType) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if uint64(len(s.functions)-1) >= uint64(s.limits.MaxFunctions) {
		return Extern{}, limit("functions")
	}
	t, typeID, err := s.hostSignature(t, referenceType)
	if err != nil {
		return Extern{}, err
	}
	id := uint64(len(s.functions))
	s.functions = append(s.functions, &Function{Owner: s, Signature: t, TypeID: typeID, Host: host})
	s.usage.Functions++
	return Extern{Owner: s, Kind: 0, Function: id}, nil
}
func (s *Store) NewMemory(t module.Limits) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if s.usage.Memories >= uint64(s.limits.MaxMemories) {
		return Extern{}, limit("memories")
	}
	if uint64(t.Min) > uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages-s.reservedShared {
		return Extern{}, limit("Store memory pages")
	}
	allocate := memory.New
	if s.nativeMemory {
		allocate = memory.NewNative
	}
	var m *memory.Memory
	var err error
	if t.Shared {
		m, err = memory.NewShared(t, uint32(uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages-s.reservedShared), s.Domain)
	} else {
		m, err = allocate(t, s.limits.MaxMemoryPages)
	}
	if err != nil {
		return Extern{}, err
	}
	if t.Shared {
		s.reservedShared += m.ReservationPages()
	} else {
		s.usage.MemoryPages += uint64(t.Min)
	}
	s.usage.Memories++
	owned := &Memory{Owner: s, Memory: m}
	s.memories = append(s.memories, owned)
	return Extern{Owner: s, Kind: 2, Memory: owned}, nil
}
func (s *Store) GrowMemory(m *Memory, delta uint64) (uint64, bool) {
	if s.closed || m.Owner != s {
		return 0, false
	}
	if m.Shared() {
		return m.Grow(delta)
	}
	old := m.Pages()
	if uint64(delta) > uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages-s.reservedShared {
		return old, false
	}
	_, ok := m.Grow(delta)
	if ok {
		s.usage.MemoryPages += uint64(delta)
	}
	return old, ok
}
func (s *Store) NewGlobal(t module.GlobalType, v value.Value) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if !s.validType(t.Value) {
		return Extern{}, fmt.Errorf("%w: global type", wasm.ErrInvalid)
	}
	if s.usage.Globals >= uint64(s.limits.MaxGlobals) {
		return Extern{}, limit("globals")
	}
	s.usage.Globals++
	g := &Global{Owner: s, Type: t, Value: v}
	s.globals = append(s.globals, g)
	return Extern{Owner: s, Kind: 3, Global: g}, nil
}
func (s *Store) NewTable(t module.TableType, initial ...uint64) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if t.Limits.Shared || !t.Element.IsReference() || !s.validType(t.Element) || len(initial) > 1 || len(initial) == 0 && !t.Element.Defaultable() || t.Limits.HasMax && t.Limits.Min > t.Limits.Max || !t.Limits.Index64 && (t.Limits.Min > math.MaxUint32 || t.Limits.HasMax && t.Limits.Max > math.MaxUint32) {
		return Extern{}, fmt.Errorf("%w: table type", wasm.ErrInvalid)
	}
	if s.usage.Tables >= uint64(s.limits.MaxTables) {
		return Extern{}, limit("tables")
	}
	if uint64(t.Limits.Min) > uint64(s.limits.MaxTableElements)-s.usage.TableElements {
		return Extern{}, limit("Store table elements")
	}
	maximum := uint64(s.limits.MaxTableElements)
	if t.Limits.HasMax && t.Limits.Max < maximum {
		maximum = t.Limits.Max
	}
	hostMax := uint64(math.MaxInt) / 8
	if uint64(maximum) > hostMax {
		maximum = hostMax
	}
	if t.Limits.Min > maximum {
		return Extern{}, limit("host table elements")
	}
	table := &Table{Owner: s, Type: t, maximum: maximum}
	if s.nativeMemory {
		var err error
		table.native, err = native.NewMemory(uint64(t.Limits.Min)*8, uint64(maximum)*8)
		if err != nil {
			return Extern{}, err
		}
		table.refreshNative()
	} else {
		table.Values = make([]uint64, int(t.Limits.Min))
	}
	if len(initial) != 0 && initial[0] != 0 {
		for i := range table.Values {
			table.Values[i] = initial[0]
		}
	}
	s.usage.TableElements += uint64(t.Limits.Min)
	s.usage.Tables++
	s.tables = append(s.tables, table)
	return Extern{Owner: s, Kind: 1, Table: table}, nil
}
func (s *Store) GrowTable(t *Table, delta uint64, initial uint64) (uint64, bool) {
	old := uint64(len(t.Values))
	if s.closed || t.Owner != s || delta > t.maximum-old || uint64(delta) > uint64(s.limits.MaxTableElements)-s.usage.TableElements {
		return old, false
	}
	if delta == 0 {
		return old, true
	}
	next := old + delta
	var grown []uint64
	if t.native != nil {
		if err := t.native.Grow(next * 8); err != nil {
			return old, false
		}
		t.refreshNative()
		grown = t.Values
	} else {
		grown = make([]uint64, int(next))
		copy(grown, t.Values)
	}
	for i := int(old); i < len(grown); i++ {
		grown[i] = initial
	}
	t.Values = grown
	s.usage.TableElements += uint64(delta)
	return old, true
}

// NativeAddress is borrowed only by the serialized Store executor. The table
// owner outlives every active/suspended frame. Growth invalidates the address;
// the driver refreshes its native view after every host/helper handoff.
func (t *Table) NativeAddress() uint64 {
	if t.native == nil || t.Owner == nil {
		return 0
	}
	return uint64(uintptr(unsafe.Pointer(unsafe.SliceData(t.Values))))
}

// IsNative reports whether the table owns native storage, including an empty
// table whose address is zero. Retirement clears the storage owner.
func (t *Table) IsNative() bool { return t.native != nil }

func (t *Table) refreshNative() {
	buffer, err := t.native.Buffer()
	if err != nil {
		panic(err) // Only called immediately after successful allocation/growth.
	}
	t.Values = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
}

func (t *Table) release() {
	if t.native != nil {
		t.native.Close()
		t.native = nil
	}
	t.Values = nil
}
func Range(size int, offset, count uint64) bool {
	return offset <= uint64(size) && count <= uint64(size)-offset
}
func (t *Table) Init(dst, src, count uint64, values []uint64) bool {
	if !Range(len(t.Values), dst, count) || !Range(len(values), src, count) {
		return false
	}
	copy(t.Values[int(dst):int(dst+count)], values[int(src):int(src+count)])
	return true
}
func (t *Table) Fill(dst, count, ref uint64) bool {
	if !Range(len(t.Values), dst, count) {
		return false
	}
	for i := uint64(0); i < count; i++ {
		t.Values[int(dst+i)] = ref
	}
	return true
}
func (in *Instance) DropData(index uint32) {
	in.Owner.usage.SegmentBytes -= uint64(len(in.Data[index]))
	in.Data[index] = nil
}
func (in *Instance) DropElement(index uint32) {
	in.Owner.usage.SegmentElements -= uint64(len(in.Elements[index]))
	in.Elements[index] = nil
}
func limit(detail string) error { return &wasm.Error{Kind: wasm.ErrLimit, Detail: detail} }
func sameLimits(actual, required module.Limits) bool {
	return actual.Shared == required.Shared && actual.Index64 == required.Index64 && actual.Min >= required.Min && (!required.HasMax || actual.HasMax && actual.Max <= required.Max)
}
func (s *Store) matches(ex Extern, im module.Import, in *Instance) bool {
	if ex.Owner != s || ex.Kind != im.Kind {
		return false
	}
	switch im.Kind {
	case 0:
		f, ok := s.Function(ex.Function)
		return ok && s.registry.FunctionSubtype(f.TypeID, in.Types[im.Type])
	case 1:
		if ex.Table == nil || ex.Table.Owner != s {
			return false
		}
		actual := ex.Table.Type
		actual.Limits.Min = uint64(len(ex.Table.Values))
		return actual.Element == canonicalType(im.Table.Element, in.Types) && sameLimits(actual.Limits, im.Table.Limits)
	case 2:
		if ex.Memory == nil || ex.Memory.Owner != s {
			return false
		}
		actual := ex.Memory.Limits()
		actual.Min = ex.Memory.Pages()
		return sameLimits(actual, im.Memory)
	case 4:
		return ex.Tag != nil && ex.Tag.Owner == s && ex.Tag.TypeID == in.Types[im.Type]
	case 3:
		return ex.Global != nil && ex.Global.Owner == s && ex.Global.Type.Mutable == im.Global.Mutable && (ex.Global.Type.Value == canonicalType(im.Global.Value, in.Types) || !im.Global.Mutable && s.registry.Subtype(ex.Global.Type.Value, canonicalType(im.Global.Value, in.Types)))
	}
	return false
}

func (s *Store) validType(t wasm.ValueType) bool {
	if !t.Valid() {
		return false
	}
	if id, indexed := t.TypeIndex(); indexed {
		if !t.Canonical() || id == 0 || uint64(id) >= uint64(len(s.registry.Types)) {
			return false
		}
		k := s.registry.Types[id].Definition.Kind
		return k == module.FunctionKind && t.IsFunctionReference() || k == module.StructKind && byte(t) == byte(wasm.StructRef) || k == module.ArrayKind && byte(t) == byte(wasm.ArrayRef)
	}
	return true
}

func canonicalType(t wasm.ValueType, ids []uint32) wasm.ValueType {
	if index, indexed := t.TypeIndex(); indexed && !t.Canonical() {
		return t.WithTypeIndex(ids[index], true)
	}
	return t
}

func constantValue(in *Instance, x *module.Instr) value.Value {
	switch x.Op {
	case 0x41, 0x42, 0x43, 0x44:
		return value.Value{Lo: x.Bits}
	case 0x23:
		return in.Globals[x.A].Value
	case 0xd0:
		return value.Value{}
	case 0xd2:
		return value.Value{Lo: in.Functions[x.A]}
	case 0xfd000c:
		return simd.Eval(x, value.Value{}, value.Value{}, value.Value{})
	default:
		panic("state: nonconstant validated initializer")
	}
}

// constant evaluates already validated initializers. Scratch is shared across
// all expressions in one allocation/initialization phase, never across Stores.
func constant(in *Instance, expr module.Expr, scratch []value.Value) value.Value {
	if len(expr) == 2 && expr[0].Op>>16 != 0xfb {
		return constantValue(in, &expr[0])
	}
	if in != nil {
		if x, ok := in.Compiled.IR.FoldedConstant(expr); ok {
			return value.Value{Lo: x}
		}
	}
	mark := 0
	if in != nil {
		mark = len(in.Owner.temporary)
		defer func() { clear(in.Owner.temporary[mark:]); in.Owner.temporary = in.Owner.temporary[:mark] }()
	}
	stack := scratch[:0]
	for i := range expr[:len(expr)-1] {
		x := &expr[i]
		if x.Op>>16 == 0xfb {
			op := x.Op & 0xffff
			count := 1
			resultType := wasm.I31Ref.NonNull()
			switch op {
			case 0:
				count = len(in.Owner.registry.Types[in.Types[x.A]].Definition.Fields)
				resultType = canonicalType(in.Compiled.IR.Validated.Definition().Reference(x.A, false), in.Types)
			case 1:
				count = 0
				resultType = canonicalType(in.Compiled.IR.Validated.Definition().Reference(x.A, false), in.Types)
			case 6:
				count = 2
				resultType = canonicalType(in.Compiled.IR.Validated.Definition().Reference(x.A, false), in.Types)
			case 7:
				count = 1
				resultType = canonicalType(in.Compiled.IR.Validated.Definition().Reference(x.A, false), in.Types)
			case 8:
				count = int(x.B)
				resultType = canonicalType(in.Compiled.IR.Validated.Definition().Reference(x.A, false), in.Types)
			case 26:
				resultType = wasm.AnyRef
			case 27:
				resultType = wasm.ExternRef
			}
			n := len(stack) - count
			v, err := in.Owner.GC(in.initializerContext, in, x, stack[n:])
			if err != nil {
				panic(constantFailure{err})
			}
			stack = append(stack[:n], v)
			in.Owner.temporary = append(in.Owner.temporary, ReferenceValue{resultType, v})
			continue
		}
		switch x.Op {
		case 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e:
			n := len(stack)
			stack[n-2] = value.Value{Lo: constops.Binary(x.Op, stack[n-2].Lo, stack[n-1].Lo)}
			stack = stack[:n-1]
		default:
			v := constantValue(in, x)
			stack = append(stack, v)
			if in != nil {
				t := wasm.ValueType(0)
				if x.Op == 0xd2 {
					t = wasm.FuncRef
				} else if x.Op == 0x23 {
					t = in.Globals[x.A].Type.Value
				}
				if t.IsReference() {
					in.Owner.temporary = append(in.Owner.temporary, ReferenceValue{t, v})
				}
			}
		}
	}
	return stack[0]
}

func constantScratch(compiled *slots.Module, local []value.Value) []value.Value {
	n := compiled.IR.Validated.ConstantStack()
	if n <= uint32(cap(local)) {
		return local[:n]
	}
	return make([]value.Value, n)
}

func (s *Store) FunctionSubtype(actual, expected uint32) bool {
	return s.registry.FunctionSubtype(actual, expected)
}
func (s *Store) ReferenceSubtype(actual, expected wasm.ValueType) bool {
	return s.registry.Subtype(actual, expected)
}

type constantFailure struct{ err error }

func (s *Store) hostSignature(t module.FuncType, referenceType []wasm.ValueType) (module.FuncType, uint32, error) {
	// Copy signatures before publication, so a host cannot mutate call typing.
	t = module.FuncType{Params: append([]wasm.ValueType(nil), t.Params...), Results: append([]wasm.ValueType(nil), t.Results...)}
	for _, ts := range [][]wasm.ValueType{t.Params, t.Results} {
		for _, typ := range ts {
			if !s.validType(typ) {
				return module.FuncType{}, 0, fmt.Errorf("%w: host value type", wasm.ErrInvalid)
			}
		}
	}
	if len(referenceType) > 1 {
		return module.FuncType{}, 0, fmt.Errorf("%w: host function type", wasm.ErrInvalid)
	}
	var typeID uint32
	if len(referenceType) == 1 {
		rt := referenceType[0]
		id, ok := rt.TypeIndex()
		if !ok || !rt.IsFunctionReference() || !s.validType(rt) || module.SignatureKey(t, nil) != module.SignatureKey(s.registry.Types[id].Function, nil) {
			return module.FuncType{}, 0, fmt.Errorf("%w: host function type", wasm.ErrInvalid)
		}
		typeID = id
	} else {
		typeID = s.TypeID(t)
	}
	if typeID == 0 {
		return module.FuncType{}, 0, limit("function types")
	}
	return t, typeID, nil
}

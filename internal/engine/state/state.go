// Package state owns Core 2.0 instance state and Store-local reference domains.
// Its mutable operations run in a serialized Store execution context; the
// embedding layer supplies concurrency, lifecycle and host-root management.
package state

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
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
	MaxTables, MaxMemories                                                   uint32
	MaxFunctions, MaxGlobals, MaxInstances, MaxMemoryPages, MaxTableElements uint32
	MaxSegmentBytes, MaxSegmentElements                                      uint64
}

func DefaultLimits() Limits {
	return Limits{MaxTypes: 1 << 20, MaxTables: 1 << 16, MaxMemories: 1024, MaxFunctions: 1 << 20, MaxGlobals: 1 << 20, MaxInstances: 1024, MaxMemoryPages: 4096, MaxTableElements: 1 << 20, MaxSegmentBytes: 64 << 20, MaxSegmentElements: 1 << 20}
}

type Usage struct {
	Tables, Memories                                          uint64
	Functions, Globals, Instances, MemoryPages, TableElements uint64
	SegmentBytes, SegmentElements                             uint64
}
type Store struct {
	closed       bool
	nativeMemory bool
	limits       Limits
	usage        Usage
	functions    []*Function // index zero is the null function reference
	instances    []*Instance
	tables       []*Table
	memories     []*Memory
	globals      []*Global
	types        map[string]uint32
}

func New(limits Limits) *Store {
	return &Store{limits: limits, functions: []*Function{nil}, types: make(map[string]uint32)}
}

// NewNative selects off-heap memories and tables for native execution.
func NewNative(limits Limits) *Store { s := New(limits); s.nativeMemory = true; return s }
func (s *Store) RetainedModules(visit func(*slots.Module)) {
	for _, in := range s.instances {
		visit(in.Compiled)
	}
}
func (s *Store) Usage() Usage { return s.usage }
func (s *Store) TypeID(t module.FuncType) uint32 {
	if s.closed {
		return 0
	}
	key := make([]byte, 0, len(t.Params)+1+len(t.Results))
	for _, v := range t.Params {
		key = append(key, byte(v))
	}
	key = append(key, 0)
	for _, v := range t.Results {
		key = append(key, byte(v))
	}
	if id, ok := s.types[string(key)]; ok {
		return id
	}
	if uint64(len(s.types)) >= uint64(s.limits.MaxTypes) {
		return 0
	}
	id := uint32(len(s.types) + 1)
	s.types[string(key)] = id
	return id
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
	maximum       uint32
	native        *native.Memory
}
type Extern struct {
	Owner    *Store
	Kind     byte
	Function uint64
	Table    *Table
	Memory   *Memory
	Global   *Global
}
type Instance struct {
	published, disposed bool
	Owner               *Store
	Compiled            *slots.Module
	Types               []uint32
	Functions           []uint64
	Tables              []*Table
	Memories            []*Memory
	Globals             []*Global
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
func (s *Store) NewHost(t module.FuncType, host any) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if uint64(len(s.functions)-1) >= uint64(s.limits.MaxFunctions) {
		return Extern{}, limit("functions")
	}
	// Copy signatures before publication, so a host cannot mutate call typing.
	t = module.FuncType{Params: append([]wasm.ValueType(nil), t.Params...), Results: append([]wasm.ValueType(nil), t.Results...)}
	for _, ts := range [][]wasm.ValueType{t.Params, t.Results} {
		for _, typ := range ts {
			if !typ.Valid() {
				return Extern{}, fmt.Errorf("%w: host value type", wasm.ErrInvalid)
			}
		}
	}
	typeID := s.TypeID(t)
	if typeID == 0 {
		return Extern{}, limit("function types")
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
	if uint64(t.Min) > uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages {
		return Extern{}, limit("Store memory pages")
	}
	allocate := memory.New
	if s.nativeMemory {
		allocate = memory.NewNative
	}
	m, err := allocate(t, s.limits.MaxMemoryPages)
	if err != nil {
		return Extern{}, err
	}
	s.usage.MemoryPages += uint64(t.Min)
	s.usage.Memories++
	owned := &Memory{Owner: s, Memory: m}
	s.memories = append(s.memories, owned)
	return Extern{Owner: s, Kind: 2, Memory: owned}, nil
}
func (s *Store) GrowMemory(m *Memory, delta uint32) (uint32, bool) {
	if s.closed || m.Owner != s {
		return 0, false
	}
	old := m.Pages()
	if uint64(delta) > uint64(s.limits.MaxMemoryPages)-s.usage.MemoryPages {
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
	if !t.Value.Valid() {
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
func (s *Store) NewTable(t module.TableType) (Extern, error) {
	if s.closed {
		return Extern{}, wasm.ErrClosed
	}
	if !t.Element.IsReference() || t.Limits.HasMax && t.Limits.Min > t.Limits.Max {
		return Extern{}, fmt.Errorf("%w: table type", wasm.ErrInvalid)
	}
	if s.usage.Tables >= uint64(s.limits.MaxTables) {
		return Extern{}, limit("tables")
	}
	if uint64(t.Limits.Min) > uint64(s.limits.MaxTableElements)-s.usage.TableElements {
		return Extern{}, limit("Store table elements")
	}
	maximum := s.limits.MaxTableElements
	if t.Limits.HasMax && t.Limits.Max < maximum {
		maximum = t.Limits.Max
	}
	hostMax := uint64(math.MaxInt) / 8
	if uint64(maximum) > hostMax {
		maximum = uint32(hostMax)
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
	s.usage.TableElements += uint64(t.Limits.Min)
	s.usage.Tables++
	s.tables = append(s.tables, table)
	return Extern{Owner: s, Kind: 1, Table: table}, nil
}
func (s *Store) GrowTable(t *Table, delta uint32, initial uint64) (uint32, bool) {
	old := uint32(len(t.Values))
	next := uint64(old) + uint64(delta)
	if s.closed || t.Owner != s || next > uint64(t.maximum) || uint64(delta) > uint64(s.limits.MaxTableElements)-s.usage.TableElements {
		return old, false
	}
	if delta == 0 {
		return old, true
	}
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
	return actual.Min >= required.Min && (!required.HasMax || actual.HasMax && actual.Max <= required.Max)
}
func (s *Store) matches(ex Extern, im module.Import, definition *module.Module) bool {
	if ex.Owner != s || ex.Kind != im.Kind {
		return false
	}
	switch im.Kind {
	case 0:
		f, ok := s.Function(ex.Function)
		if !ok {
			return false
		}
		t := definition.Types[im.Type]
		if len(t.Params) != len(f.Signature.Params) || len(t.Results) != len(f.Signature.Results) {
			return false
		}
		for i, v := range t.Params {
			if v != f.Signature.Params[i] {
				return false
			}
		}
		for i, v := range t.Results {
			if v != f.Signature.Results[i] {
				return false
			}
		}
		return true
	case 1:
		if ex.Table == nil || ex.Table.Owner != s {
			return false
		}
		actual := ex.Table.Type
		actual.Limits.Min = uint32(len(ex.Table.Values))
		return actual.Element == im.Table.Element && sameLimits(actual.Limits, im.Table.Limits)
	case 2:
		if ex.Memory == nil || ex.Memory.Owner != s {
			return false
		}
		actual := ex.Memory.Limits()
		actual.Min = ex.Memory.Pages()
		return sameLimits(actual, im.Memory)
	case 3:
		return ex.Global != nil && ex.Global.Owner == s && ex.Global.Type == im.Global
	}
	return false
}
func constant(in *Instance, expr module.Expr) value.Value {
	x := expr[0]
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

package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
)

// ValueKind identifies a component value shape. Zero denotes no type/payload.
type ValueKind = model.Kind

const (
	KindBool         = model.Bool
	KindS8           = model.S8
	KindU8           = model.U8
	KindS16          = model.S16
	KindU16          = model.U16
	KindS32          = model.S32
	KindU32          = model.U32
	KindS64          = model.S64
	KindU64          = model.U64
	KindF32          = model.F32
	KindF64          = model.F64
	KindChar         = model.Char
	KindString       = model.String
	KindList         = model.List
	KindFixedList    = model.FixedList
	KindMap          = model.Map
	KindTuple        = model.Tuple
	KindRecord       = model.Record
	KindVariant      = model.Variant
	KindOption       = model.Option
	KindResult       = model.Result
	KindEnum         = model.Enum
	KindFlags        = model.Flags
	KindOwn          = model.Own
	KindBorrow       = model.Borrow
	KindFuture       = model.Future
	KindStream       = model.Stream
	KindErrorContext = model.ErrorContext
)

// FunctionCheck validates a host provider's signature before any guest starts.
// Nil checks preserve the raw callback API's guest-declared schema contract.
type FunctionCheck func(FunctionInfo) error

type pendingFunctionCheck struct {
	name  string
	typ   *validate.Type
	check FunctionCheck
}

// FunctionInfo and ValueType expose immutable checked graphs. Views may be
// retained after binding; they never expose mutable model or registry slices.
type FunctionInfo struct {
	typ       *validate.Type
	resources map[uint64]*resourceType
}

type ValueType struct {
	typ       *model.Type
	resources map[uint64]*resourceType
}

func (f FunctionInfo) ParameterCount() int {
	if f.typ == nil {
		return 0
	}
	return len(f.typ.Parameters)
}
func (f FunctionInfo) Parameter(i int) (string, ValueType) {
	if i < 0 || i >= f.ParameterCount() {
		return "", ValueType{}
	}
	p := f.typ.Parameters[i]
	return p.Name, ValueType{p.Type, f.resources}
}
func (f FunctionInfo) Result() ValueType {
	if f.typ == nil {
		return ValueType{}
	}
	return ValueType{f.typ.Result, f.resources}
}
func (t ValueType) Valid() bool { return t.typ != nil }
func (t ValueType) Kind() ValueKind {
	if t.typ == nil {
		return 0
	}
	return t.typ.Kind
}
func (t ValueType) Element() ValueType {
	if t.typ == nil {
		return ValueType{}
	}
	return ValueType{t.typ.Element, t.resources}
}
func (t ValueType) Key() ValueType {
	if t.typ == nil {
		return ValueType{}
	}
	return ValueType{t.typ.Key, t.resources}
}
func (t ValueType) Length() uint32 {
	if t.typ == nil {
		return 0
	}
	return t.typ.Length
}
func (t ValueType) FieldCount() int {
	if t.typ == nil {
		return 0
	}
	return len(t.typ.Fields)
}
func (t ValueType) Field(i int) (string, ValueType) {
	if i < 0 || i >= t.FieldCount() {
		return "", ValueType{}
	}
	f := t.typ.Fields[i]
	return f.Name, ValueType{f.Type, t.resources}
}
func (t ValueType) CaseCount() int {
	if t.typ == nil {
		return 0
	}
	return len(t.typ.Cases)
}
func (t ValueType) Case(i int) (string, ValueType) {
	if i < 0 || i >= t.CaseCount() {
		return "", ValueType{}
	}
	f := t.typ.Cases[i]
	return f.Name, ValueType{f.Type, t.resources}
}
func (t ValueType) LabelCount() int {
	if t.typ == nil {
		return 0
	}
	return len(t.typ.Labels)
}
func (t ValueType) Label(i int) string {
	if i < 0 || i >= t.LabelCount() {
		return ""
	}
	return t.typ.Labels[i]
}

// IsResource compares nominal host declarations, not representation integers.
func (t ValueType) IsResource(host *HostResourceType) bool {
	if host == nil || host.definition == nil || (t.Kind() != KindOwn && t.Kind() != KindBorrow) {
		return false
	}
	rt := t.resources[t.typ.ResourceID]
	return rt != nil && rt.host == host.definition
}

package model

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Binary is an owned, unvalidated component definition. Definitions preserve
// binary order: component sections may repeat and their order is meaningful.
type Binary struct {
	Definitions []Definition
	CoreTypes   *module.Module
}

type Definition struct {
	Kind       byte
	Offset     uint64
	Bytes      []byte
	Component  *Binary
	CoreType   *CoreTypeDef
	CoreModule *module.Module
	Type       *TypeDef
	Instance   *InstanceExpr
	Alias      *Alias
	Canonical  *Canonical
	Import     *Import
	Export     *Export
	Start      *Start
	Value      *ValueDefinition
}

type Sort uint16

const (
	CoreFunctionSort Sort = 0x100
	CoreTableSort    Sort = 0x101
	CoreMemorySort   Sort = 0x102
	CoreGlobalSort   Sort = 0x103
	CoreTagSort      Sort = 0x104
	CoreTypeSort     Sort = 0x110
	CoreModuleSort   Sort = 0x111
	CoreInstanceSort Sort = 0x112
	FunctionSort     Sort = 1
	ValueSort        Sort = 2
	TypeSort         Sort = 3
	ComponentSort    Sort = 4
	InstanceSort     Sort = 5
)

type SortIndex struct {
	Sort  Sort
	Index uint32
}
type Attribute struct {
	Kind  byte
	Value string
}
type Name struct {
	Text       string
	Attributes []Attribute
}
type Export struct {
	Name Name
	Item SortIndex
	Type *ExternType
}
type Import struct {
	Name Name
	Type ExternType
}
type ExternType struct {
	Sort  Sort
	Index uint32
	Bound byte
	Value *TypeRef
}
type Argument struct {
	Name string
	Item SortIndex
}
type InstanceExpr struct {
	Instantiate bool
	Index       uint32
	Arguments   []Argument
	Exports     []Export
}
type Alias struct {
	Sort                   Sort
	Kind                   byte
	Instance, Count, Index uint32
	Name                   string
}
type Start struct {
	Function  uint32
	Arguments []uint32
	Results   uint32
}
type Canonical struct {
	Op        byte
	Index     uint32
	Extra     uint32
	Flag      byte
	CoreValue wasm.ValueType
	Results   *TypeRef
	Options   []CanonicalOption
}
type CanonicalOption struct {
	Kind  byte
	Index uint32
}
type ValueDefinition struct {
	Type  TypeRef
	Bytes []byte
}

type TypeRef struct {
	Indexed   bool
	Index     uint32
	Primitive Kind
}
type NamedType struct {
	Name string
	Type TypeRef
}
type TypeCase struct {
	Name string
	Type *TypeRef
}
type TypeDef struct {
	Kind           Kind
	Element        *TypeRef
	Length         uint32
	Key            *TypeRef
	Fields         []NamedType
	Tuple          []TypeRef
	Cases          []TypeCase
	Labels         []string
	Parameters     []NamedType
	Result         *TypeRef
	Declarations   []TypeDeclaration
	CoreTypes      *module.Module
	Representation wasm.ValueType
	Destructor     *uint32
	Resource       uint32
}
type TypeDeclaration struct {
	Kind     byte
	CoreType *CoreTypeDef
	Type     *TypeDef
	Alias    *Alias
	Import   *Import
	Export   *Import
}
type CoreTypeDef struct {
	Module       bool
	Start, Count uint32
	Declarations []CoreDeclaration
	Types        *module.Module
}
type CoreDeclaration struct {
	Kind         byte
	Name, Module string
	Type         module.Import
	Alias        *Alias
	Start, Count uint32
}

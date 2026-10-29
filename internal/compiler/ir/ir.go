// Package ir represents typed Core control flow independently of execution
// backends. Values are SSA identities; mutable locals remain explicit operations.
package ir

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Value uint32
type BlockID uint32

// Source is an immutable operation schema derived from validation, preserving
// the original immediate, types, byte offset and effect classification. An
// optimizer must construct and verify a new schema when changing an operation.
type Source struct {
	Wasm            module.Instr
	Inputs, Outputs []wasm.ValueType
	Effects         Effects
}

// Fuel charges one unit per executed Wasm instruction. Else/end delimiters and
// compiler-generated edge transfers have no additional charge. A loop header
// is charged when entered from outside; a branch back to its body charges the
// branch and subsequent body operations. This is fuel model version 1.
func (s *Source) Fuel() uint64 {
	if s.Wasm.Op == 5 || s.Wasm.Op == 0xb {
		return 0
	}
	return 1
}

type Operation struct {
	Source        uint32
	Args, Results []Value
}
type Edge struct {
	Target BlockID
	Args   []Value
}
type TermKind byte

const (
	NoTerm TermKind = iota
	Jump
	If
	Table
	Return
	Trap
)

type Terminator struct {
	Kind   TermKind
	Source uint32
	Args   []Value
	Edges  []Edge
}
type Block struct {
	Params []Value
	Ops    []Operation
	Term   Terminator
}
type Function struct {
	Signature module.FuncType
	Locals    []wasm.ValueType
	Values    []wasm.ValueType
	Sources   []Source
	Blocks    []Block
}

type Effects uint32

const (
	ReadLocal Effects = 1 << iota
	WriteLocal
	ReadGlobal
	WriteGlobal
	ReadMemory
	WriteMemory
	ReadTable
	WriteTable
	ReadSegment
	WriteSegment
	MayTrap
	MayCallHost
	MayAllocate
	MaySuspend
)

func EffectsOf(op uint32) Effects {
	switch {
	case op == 0:
		return MayTrap
	case op == 0x10 || op == 0x11:
		return ReadGlobal | WriteGlobal | ReadMemory | WriteMemory | ReadTable | WriteTable | ReadSegment | WriteSegment | MayTrap | MayCallHost | MayAllocate | MaySuspend
	case op == 0x20:
		return ReadLocal
	case op == 0x21 || op == 0x22:
		return WriteLocal
	case op == 0x23:
		return ReadGlobal
	case op == 0x24:
		return WriteGlobal
	case op == 0x25:
		return ReadTable | MayTrap
	case op == 0x26:
		return WriteTable | MayTrap
	case op >= 0x28 && op <= 0x35:
		return ReadMemory | MayTrap
	case op >= 0x36 && op <= 0x3e:
		return WriteMemory | MayTrap
	case op == 0x3f:
		return ReadMemory
	case op == 0x40:
		return ReadMemory | WriteMemory | MayAllocate
	case op >= 0x6d && op <= 0x70 || op >= 0x7f && op <= 0x82 || op >= 0xa8 && op <= 0xab || op >= 0xae && op <= 0xb1:
		return MayTrap
	case op == 0xfc0008:
		return ReadSegment | WriteMemory | MayTrap
	case op == 0xfc0009 || op == 0xfc000d:
		return WriteSegment
	case op == 0xfc000a:
		return ReadMemory | WriteMemory | MayTrap
	case op == 0xfc000b:
		return WriteMemory | MayTrap
	case op == 0xfc000c:
		return ReadSegment | WriteTable | MayTrap
	case op == 0xfc000e:
		return ReadTable | WriteTable | MayTrap
	case op == 0xfc000f:
		return ReadTable | WriteTable | MayAllocate
	case op == 0xfc0010:
		return ReadTable
	case op == 0xfc0011:
		return WriteTable | MayTrap
	case op>>16 == 0xfd:
		sub := op & 0xffff
		if sub == 11 || sub >= 88 && sub <= 91 {
			return WriteMemory | MayTrap
		}
		if sub <= 10 || sub >= 84 && sub <= 87 || sub == 92 || sub == 93 {
			return ReadMemory | MayTrap
		}
	}
	return 0
}

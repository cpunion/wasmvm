// Package nativeexec connects owned generated code to the common frame machine.
// Calls, host requests and explicit runtime helpers never interpret a function.
package nativeexec

import (
	"context"
	"errors"
	"math"
	"runtime"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Module struct {
	code    *native.Code
	entries map[*slots.Function]functionEntries
}
type Lease struct {
	code    *native.Lease
	entries map[*slots.Function]functionEntries
}

type functionEntries struct {
	ordinary, metered, expanded uint32
	inlineSlots                 uint32
	hasExpansion                bool
	directRuntimeData           bool
	usesRuntimeData             bool
	globals                     []uint32
}

// Compile prepares both policy tiers within one aggregate code-byte budget.
// Publication is atomic: no executable allocation survives a failed compile.
func Compile(ctx context.Context, m *slots.Module, limits codegen.Limits) (*Module, error) {
	if !native.Available {
		return nil, wasm.ErrUnsupported
	}
	out := &Module{entries: make(map[*slots.Function]functionEntries, len(m.Functions))}
	var code []byte
	var entries []uint32
	definition := m.IR.Validated.Definition()
	var minimum uint32
	var importedFunctions uint32
	var tables, memories uint32
	var globals []module.GlobalType
	for _, im := range definition.Imports {
		if im.Kind == 0 {
			importedFunctions++
		}
		if im.Kind == 2 {
			memories++
			if memories == 1 {
				minimum = im.Memory.Min
			}
		}
		if im.Kind == 1 {
			tables++
		}
		if im.Kind == 3 {
			globals = append(globals, im.Global)
		}
	}
	tables += uint32(len(definition.Tables))
	memories += uint32(len(definition.Memories))
	for _, g := range definition.Globals {
		globals = append(globals, g.Type)
	}
	if len(definition.Memories) != 0 && memories == uint32(len(definition.Memories)) {
		minimum = definition.Memories[0].Min
	}
	callees := make(map[uint32]*slots.Function, len(m.Functions))
	for i, f := range m.Functions {
		callees[importedFunctions+uint32(i)] = f
	}
	// Index structurally identical types once. Bound code expansion per type,
	// independently of map iteration order and table contents at instantiation.
	keys := make([]string, len(definition.Types))
	for i, typ := range definition.Types {
		key := make([]byte, 0, len(typ.Params)+len(typ.Results)+1)
		for _, p := range typ.Params {
			key = append(key, byte(p))
		}
		key = append(key, 0)
		for _, r := range typ.Results {
			key = append(key, byte(r))
		}
		keys[i] = string(key)
	}
	bySignature := map[string][]uint32{}
	for i, f := range m.Functions {
		key := keys[definition.Functions[i].Type]
		if len(bySignature[key]) < 4 && codegen.CanExpandLeaf(f) {
			bySignature[key] = append(bySignature[key], importedFunctions+uint32(i))
		}
	}
	indirect := map[uint32][]uint32{}
	for i, key := range keys {
		indirect[uint32(i)] = bySignature[key]
	}
	wantExpansion := map[*slots.Function]bool{}
	features := codegen.X86Features(native.CPUFeatures())
	// Publish all required ordinary tiers before spending spare code budget
	// on optional expansion, including when later functions are large.
	for tier := 0; tier < 3; tier++ {
		for _, f := range m.Functions {
			indexes := out.entries[f]
			if tier == 2 && !wantExpansion[f] {
				continue
			}
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			bounded := limits
			bounded.MaxCodeBytes = min(limits.MaxCodeBytes, uint64(math.MaxUint32)) - uint64(len(code))
			config := codegen.Config{Unmetered: tier != 1, MemoryMinimumPages: minimum, X86Features: features,
				FunctionCount: importedFunctions + uint32(len(m.Functions)), TableCount: tables, MemoryCount: memories, Globals: globals}
			if tier == 2 {
				config.DirectCallees = callees
				config.IndirectCallees = indirect
			}
			body, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, f, bounded, config)
			if tier == 2 && errors.Is(err, wasm.ErrLimit) {
				// The optional expansion cannot consume the ordinary tiers'
				// code budget or make an otherwise valid module unavailable.
				continue
			}
			if err != nil {
				return nil, err
			}
			if tier == 2 && body.InlinedCalls == 0 {
				continue
			}
			index := uint32(len(entries))
			switch tier {
			case 0:
				indexes.ordinary = index
				indexes.globals = body.GlobalIndices
				indexes.directRuntimeData = body.UsesRuntimeData
				wantExpansion[f] = body.Calls != 0
			case 1:
				indexes.metered = index
			case 2:
				indexes.expanded, indexes.inlineSlots, indexes.hasExpansion = index, body.InlineSlots, true
				indexes.usesRuntimeData = body.UsesRuntimeData
			}
			entries = append(entries, uint32(len(code)))
			code = append(code, body.Entry.Code...)
			out.entries[f] = indexes
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(entries) != 0 {
		var err error
		out.code, err = native.Publish(code, entries, limits.MaxCodeBytes)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (m *Module) Acquire() (*Lease, error) {
	l := &Lease{entries: m.entries}
	if m.code != nil {
		var err error
		l.code, err = m.code.Acquire()
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}
func (m *Module) Close() {
	if m.code != nil {
		m.code.Close()
	}
}
func (l *Lease) Close() {
	if l.code != nil {
		l.code.Close()
	}
}

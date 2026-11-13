// Package nativeexec connects owned generated code to the common frame machine.
// Calls, host requests and explicit runtime helpers never interpret a function.
package nativeexec

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Module struct {
	code    *native.Code
	entries []functionEntries
	first   uint32
	linked  bool
}
type Lease struct {
	code    *native.Lease
	entries []functionEntries
	first   uint32
	linked  bool
}

type functionEntries struct {
	function                    *slots.Function
	ordinary, metered, expanded uint32
	inlineSlots, inlineFrames   uint32
	hasExpansion                bool
	directRuntimeData           bool
	usesRuntimeData             bool
	globals                     []uint32
	allocationBatch             uint32
	nativeMemory                bool
}

// Compile prepares both policy tiers within one aggregate code-byte budget.
// Publication is atomic: no executable allocation survives a failed compile.
func Compile(ctx context.Context, m *slots.Module, limits codegen.Limits) (*Module, error) {
	if !native.Available {
		return nil, wasm.ErrUnsupported
	}
	compiled, err := compileModule(ctx, m, limits, true)
	if errors.Is(err, wasm.ErrLimit) {
		// Linking is an optimization. A module that fits the complete ordinary
		// tiers must still compile when its linked metadata exceeds the budget.
		return compileModule(ctx, m, limits, false)
	}
	return compiled, err
}

func compileModule(ctx context.Context, m *slots.Module, limits codegen.Limits, allowLink bool) (*Module, error) {
	out := &Module{}
	var code []byte
	var entries []uint32
	definition := m.IR.Validated.Definition()
	var minimum uint32
	var importedFunctions uint32
	var tables, memories uint32
	var shared []bool
	var pageSizes []uint64
	var globals []module.GlobalType
	for _, im := range definition.Imports {
		if im.Kind == 0 {
			importedFunctions++
		}
		if im.Kind == 2 {
			shared = append(shared, im.Memory.Shared)
			pageSizes = append(pageSizes, im.Memory.PageBytes())
			memories++
			if memories == 1 && !im.Memory.Index64 {
				minimum = uint32(im.Memory.Min)
			}
		}
		if im.Kind == 1 {
			tables++
		}
		if im.Kind == 3 {
			globals = append(globals, im.Global)
		}
	}
	for _, m := range definition.Memories {
		shared = append(shared, m.Shared)
		pageSizes = append(pageSizes, m.PageBytes())
	}
	tables += uint32(len(definition.Tables))
	memories += uint32(len(definition.Memories))
	for _, g := range definition.Globals {
		globals = append(globals, g.Type)
	}
	if len(definition.Memories) != 0 && memories == uint32(len(definition.Memories)) && !definition.Memories[0].Index64 {
		minimum = uint32(definition.Memories[0].Min)
	}
	out.first = importedFunctions
	out.entries = make([]functionEntries, len(m.Functions))
	callees := make(map[uint32]*slots.Function, len(m.Functions))
	for i, f := range m.Functions {
		callees[importedFunctions+uint32(i)] = f
	}
	var linkedCallees map[uint32]*slots.Function
	var linkedGlobals []uint32
	if allowLink {
		seen := map[uint32]bool{}
		for _, f := range m.Functions {
			for _, in := range f.Code {
				if in.Op == 0x11 || in.Op == 0x14 || in.Op == 0x10 && callees[f.Sources[in.Source].Wasm.A] != nil {
					linkedCallees = callees
				}
				if in.Op == 0x23 || in.Op == 0x24 {
					seen[f.Sources[in.Source].Wasm.A] = true
				}
			}
		}
		for index := range seen {
			linkedGlobals = append(linkedGlobals, index)
		}
		slices.Sort(linkedGlobals)
	}
	out.linked = linkedCallees != nil
	relocations := [2][]codegen.CallRelocation{}
	dispatchSites := [2][]uint32{}
	var dispatchBytes uint64
	targets := [2]map[uint32]uint32{{}, {}}
	usesMemory := false
	// Index structurally identical types once. Bound code expansion per type,
	// independently of map iteration order and table contents at instantiation.
	keys := m.IR.Validated.TypeLabels()
	bySignature := map[uint32][]uint32{}
	leafCandidates := map[uint32]bool{}
	for i, f := range m.Functions {
		key := keys[definition.Functions[i].Type]
		if codegen.CanExpandLeaf(f, callees) {
			leafCandidates[importedFunctions+uint32(i)] = true
			if len(bySignature[key]) < 4 {
				bySignature[key] = append(bySignature[key], importedFunctions+uint32(i))
			}
		}
	}
	indirect := map[uint32][]uint32{}
	for i, key := range keys {
		indirect[uint32(i)] = bySignature[key]
	}
	linkedSignatures := make([]uint32, len(m.Functions))
	for i, f := range definition.Functions {
		linkedSignatures[i] = keys[f.Type]
	}
	wantExpansion := map[*slots.Function]bool{}
	features := codegen.X86Features(native.CPUFeatures())
	armFeatures := codegen.ARM64Features(native.ARM64CPUFeatures())
	// Publish all required ordinary tiers before spending spare code budget
	// on optional expansion, including when later functions are large.
	for tier := 0; tier < 3; tier++ {
		bounded := limits
		bounded.MaxCodeBytes = min(limits.MaxCodeBytes, uint64(math.MaxUint32)) - uint64(len(code)) - dispatchBytes
		config := codegen.Config{Unmetered: tier != 1, NativeAllocations: tier != 1, NativeThreads: true, MemoryMinimumPages: minimum, X86Features: features, ARM64Features: armFeatures,
			MemoryPageSizes: pageSizes, HeapTypes: definition.TypeDefs,
			FunctionCount: importedFunctions + uint32(len(m.Functions)), ScalarImports: importedFunctions, TableCount: tables, MemoryCount: memories, SharedMemories: shared, Globals: globals}
		config.LinkedCallees, config.LinkedGlobals = linkedCallees, linkedGlobals
		config.SharedDispatch = true
		if linkedCallees != nil {
			config.LinkedSignatures, config.LinkedTypeLabels = linkedSignatures, keys
		}
		if tier == 2 {
			config.DirectCallees = callees
			config.IndirectCallees = indirect
		}
		compiler := newFunctionCompiler(ctx, m.Functions, bounded, config, tier == 2, wantExpansion)
		defer compiler.close()
		for functionIndex, f := range m.Functions {
			row := functionIndex
			indexes := out.entries[row]
			indexes.function = f
			if tier == 2 && !wantExpansion[f] {
				continue
			}
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			remaining := min(limits.MaxCodeBytes, uint64(math.MaxUint32)) - uint64(len(code)) - dispatchBytes
			result := compiler.next(remaining)
			if result.panicked {
				panic(result.panicValue)
			}
			body, err := result.body, result.err
			policy := 0
			if tier == 1 {
				policy = 1
			}
			var newDispatchBytes uint64
			if err == nil {
				if len(body.DispatchRelocations) != 0 && len(dispatchSites[policy]) == 0 {
					newDispatchBytes = uint64(len(linkedSignatures)) * 32
				}
				if uint64(len(body.Entry.Code))+newDispatchBytes > remaining {
					err = wasm.ErrLimit
				}
			}
			if tier == 2 && errors.Is(err, wasm.ErrLimit) {
				// The optional expansion cannot consume the ordinary tiers'
				// code budget or make an otherwise valid module unavailable.
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("native function %d tier %d (%d instructions, %d slots, %d/%d code bytes): %w", importedFunctions+uint32(functionIndex), tier, len(f.Code), len(f.Slots), len(code), limits.MaxCodeBytes, err)
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
				indexes.allocationBatch = body.NativeAllocationBatch
				if body.Calls != 0 {
					for _, instruction := range f.Code {
						source := f.Sources[instruction.Source].Wasm
						if instruction.Op == 0x10 && leafCandidates[source.A] || instruction.Op == 0x11 && len(indirect[source.A]) != 0 {
							wantExpansion[f] = true
							break
						}
					}
				}
			case 1:
				indexes.metered = index
			case 2:
				indexes.expanded, indexes.inlineSlots, indexes.hasExpansion = index, body.InlineSlots, true
				indexes.inlineFrames = body.InlineFrames
				indexes.usesRuntimeData = body.UsesRuntimeData
			}
			indexes.nativeMemory = indexes.nativeMemory || body.UsesMemory
			usesMemory = usesMemory || body.UsesMemory
			dispatchBytes += newDispatchBytes
			for _, site := range body.DispatchRelocations {
				dispatchSites[policy] = append(dispatchSites[policy], uint32(len(code))+site)
			}
			if tier != 2 {
				targets[policy][importedFunctions+uint32(functionIndex)] = uint32(len(code)) + body.Entry.BodyOffset
			}
			for _, relocation := range body.CallRelocations {
				relocation.Offset += uint32(len(code))
				relocations[policy] = append(relocations[policy], relocation)
			}
			entries = append(entries, uint32(len(code)))
			code = append(code, body.Entry.Code...)
			out.entries[row] = indexes
		}
		compiler.close()
	}
	for policy, sites := range dispatchSites {
		if len(sites) == 0 {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		table, calls, err := codegen.LinkedDispatchTable(runtime.GOARCH, codegen.Config{
			FunctionCount: importedFunctions + uint32(len(m.Functions)), LinkedCallees: linkedCallees, LinkedSignatures: linkedSignatures,
		})
		if err != nil {
			return nil, err
		}
		offset := uint32(len(code))
		code = append(code, table...)
		if err := codegen.LinkDispatch(runtime.GOARCH, code, sites, offset); err != nil {
			return nil, err
		}
		for _, call := range calls {
			call.Offset += offset
			relocations[policy] = append(relocations[policy], call)
		}
	}
	for policy := range relocations {
		if err := codegen.LinkCalls(runtime.GOARCH, code, relocations[policy], targets[policy]); err != nil {
			return nil, err
		}
	}
	if out.linked && usesMemory {
		for f, entry := range out.entries {
			entry.nativeMemory = true
			out.entries[f] = entry
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(entries) != 0 {
		integerOnly, err := integerModule(ctx, m)
		if err != nil {
			return nil, err
		}
		publish := native.Publish
		if integerOnly {
			publish = native.PublishIntegerCode
		}
		out.code, err = publish(code, entries, limits.MaxCodeBytes)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (m *Module) Acquire() (*Lease, error) {
	l := &Lease{entries: m.entries, linked: m.linked, first: m.first}
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

func (l *Lease) function(fn *state.Function) (*functionEntries, bool) {
	index := uint64(fn.Index) - uint64(l.first)
	if index >= uint64(len(l.entries)) {
		return nil, false
	}
	entry := &l.entries[index]
	return entry, fn.Code != nil && entry.function == fn.Code
}

package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct {
	MaxCodeBytes              uint64
	MaxInstructions, MaxSlots uint32
}

func DefaultLimits() Limits { return Limits{64 << 20, 1 << 20, 1 << 20} }

// CompiledFunction is immutable. A continuation is an original slot-code PC,
// never an interior machine address. ReferenceSlots conservatively describes
// every reference-bearing slot in the current frame (root-map ID zero).
type CompiledFunction struct {
	Entry                      nativeabi.Entry
	Function                   *slots.Function
	Offsets                    []uint32
	ReferenceSlots             []uint32
	Helpers, Calls, EdgeCopies uint32
	FusedIntegers              uint32
	ForwardedPredicates        uint32
	UsesMemory                 bool
	BoundsElided               uint32
	MemoryGrows                uint32
	ForwardedScalars           uint32
	InlinedCalls, InlineSlots  uint32
	UsesRuntimeData            bool
}

// CompileFunction consumes trusted, immutable output of slot lowering. It emits
// complete control flow, integer operations, local/constant/select operations
// and explicit runtime exits for the remaining Core 2.0 semantics. A driver
// must service RuntimeHelper/Call; the compiler never inserts an interpreter.
// Payload[3] enables semantic fuel; Budget bounds native instructions per slice.
// RuntimeHelper/Call publish the instruction PC in Payload[0] and the next PC
// in Continuation. Return publishes its PC for multi-result marshalling.
// Move publishes a large edge-copy plan in Payload[0] and its destination PC
// in Continuation; the driver copies it in bounded, cancelable chunks.
func CompileFunction(arch string, f *slots.Function, limits Limits) (*CompiledFunction, error) {
	return compileFunction(arch, f, limits, Config{})
}

// CompileUnmeteredFunction emits bounded interrupt polls without periodic host
// round trips or dynamic fuel checks. The driver must supply canonical scalar
// values (Hi == 0), as required by value.Value, and Payload[3] == 0. Budget is
// unused. All scalar slots are canonical on exit; vectors retain all 128 bits.
// The shared runtime selects this tier for calls without semantic fuel.
func CompileUnmeteredFunction(arch string, f *slots.Function, limits Limits) (*CompiledFunction, error) {
	return compileFunction(arch, f, limits, Config{Unmetered: true})
}

// Config records validated module assumptions. Entry checks the minimum memory
// length before any optimized access; every resume repeats this binding check.
type Config struct {
	Unmetered          bool
	MemoryMinimumPages uint32
	// DirectCallees contains only immutable definitions in the same module.
	// An expanded entry requires one additional logical frame and InlineSlots
	// remaining slots. The runtime must select an ordinary entry otherwise.
	DirectCallees map[uint32]*slots.Function
	// IndirectCallees indexes bounded leaf candidates by the call's type index.
	// Every candidate must have that validated structural signature. Identity
	// is checked against the current instance's function reference at runtime.
	IndirectCallees map[uint32][]uint32
	FunctionCount   uint32
}

func CompileFunctionWithConfig(arch string, f *slots.Function, limits Limits, config Config) (*CompiledFunction, error) {
	return compileFunction(arch, f, limits, config)
}

func compileFunction(arch string, f *slots.Function, limits Limits, config Config) (out *CompiledFunction, err error) {
	unmetered := config.Unmetered
	if config.MemoryMinimumPages > 65536 {
		return nil, wasm.ErrInvalid
	}
	defer func() {
		if p := recover(); p != nil {
			if e, ok := p.(*wasm.Error); ok {
				out = nil
				err = e
			} else {
				panic(p)
			}
		}
	}()
	if f == nil || len(f.Code) == 0 || len(f.Slots) < int(f.LocalCount) {
		return nil, wasm.ErrInvalid
	}
	if uint64(len(f.Code)) > uint64(limits.MaxInstructions) || uint64(len(f.Slots)) > uint64(limits.MaxSlots) || uint64(len(f.Slots)) > math.MaxInt32/16 {
		return nil, wasm.ErrLimit
	}
	// Each continuation/table edge needs at least one word of code or metadata.
	// Bound label arrays before allocating them, even for enormous br_tables.
	if uint64(len(f.Code)) > limits.MaxCodeBytes/4 || uint64(len(f.Edges)) > limits.MaxCodeBytes/4 {
		return nil, wasm.ErrLimit
	}
	if err := checkFunction(f); err != nil {
		return nil, err
	}
	empty, err := nativeabi.Wrap(arch, nil, limits.MaxCodeBytes)
	if err != nil {
		return nil, err
	}
	a := assembler{arch: arch, labels: map[int]int{}, bounded: true, limit: limits.MaxCodeBytes - uint64(len(empty.Code)), longBranches: true, compactBranches: arch == "arm64"}
	e := functionEmitter{a: &a, f: f, exit: a.label(), bad: a.label(), zero: a.label(), overflow: a.label(), unmetered: unmetered, memoryMinimum: uint64(config.MemoryMinimumPages) * 65536}
	pcs := make([]int, len(f.Code))
	for i := range pcs {
		pcs[i] = a.label()
	}
	for _, in := range f.Code {
		if in.Op >= 0x28 && in.Op <= 0x40 {
			e.usesMemory = true
			break
		}
	}
	e.header(uint64(len(f.Slots))*16, len(pcs))
	if e.usesMemory {
		e.prepareMemory()
	}
	if unmetered {
		e.prepareRegisters()
		e.prepareConstants(config)
		e.prepareRuntimeRegisters(config)
		e.contextLoad(0, nativeabi.PayloadOffset+24)
		e.test(true)
		e.branch(e.bad, "ne")
		e.contextLoad(0, nativeabi.ContinuationOffset)
	}
	a.dispatch(pcs)
	out = &CompiledFunction{Function: f, Offsets: make([]uint32, len(pcs)), UsesMemory: e.usesMemory}
	for i, t := range f.Slots {
		if t.IsReference() {
			out.ReferenceSlots = append(out.ReferenceSlots, uint32(i))
		}
	}
	regions := e.memoryRegions()
	var independentEntries []int
	combinedThrough := -1
	for pc, in := range f.Code {
		if pc <= combinedThrough {
			continue
		}
		// Only fallthrough from the preceding constant proves an immediate.
		// Keep the original PC as a generic cold entry for branches/resumption.
		if unmetered && pc > 0 && pc%256 != 0 {
			previous := f.Code[pc-1]
			if (previous.Op == 0x41 || previous.Op == 0x42) && previous.Dst == in.B && in.B != slots.NoSlot &&
				e.registerImmediate(in.Op, in.Dst, in.A, f.Sources[previous.Source].Wasm.Bits) {
				independentEntries = append(independentEntries, pc)
				continue
			}
		}
		if e.compareBranchFallthrough(pc, pcs) {
			out.ForwardedPredicates++
			independentEntries = append(independentEntries, pc)
			continue
		}
		a.mark(pcs[pc])
		src := f.Sources[in.Source]
		isControl := in.Op >= slots.ControlBase
		if unmetered {
			if pc%256 == 0 || in.Op == 0x10 || in.Op == 0x11 || isControl {
				e.interrupt(uint32(pc))
			}
		} else {
			e.boundary(uint32(pc), src, in.Op == 0x10 || in.Op == 0x11 || isControl)
		}
		if region, ok := regions[pc]; ok {
			e.emitMemoryRegion(pc, region, out)
			for entry := pc + 1; entry <= region.end; entry++ {
				independentEntries = append(independentEntries, entry)
			}
			combinedThrough = region.end
			continue
		}
		if end, ok := e.combineMaskShift(pc); ok {
			out.FusedIntegers++
			for entry := pc + 1; entry <= end; entry++ {
				independentEntries = append(independentEntries, entry)
			}
			combinedThrough = end
			continue
		}
		if e.logicalShift(pc, in) {
			out.FusedIntegers++
			independentEntries = append(independentEntries, pc+1, pc+2)
			combinedThrough = pc + 2
			continue
		}
		if end, ok := e.combineScalarCopies(pc); ok {
			out.ForwardedScalars++
			for entry := pc + 1; entry <= end; entry++ {
				independentEntries = append(independentEntries, entry)
			}
			combinedThrough = end
			continue
		}

		if end, ok := e.combineMultiplyAdd(pc); ok {
			out.FusedIntegers++
			for entry := pc + 1; entry <= end; entry++ {
				independentEntries = append(independentEntries, entry)
			}
			combinedThrough = end
			continue
		}
		if end, ok := e.combineIntegerConstant(pc); ok {
			out.FusedIntegers++
			for entry := pc + 1; entry <= end; entry++ {
				independentEntries = append(independentEntries, entry)
			}
			combinedThrough = end
			continue
		}
		if e.forwardScalarResult(pc) {
			out.ForwardedScalars++
			independentEntries = append(independentEntries, pc+1)
			combinedThrough = pc + 1
			continue
		}
		switch op := in.Op; {
		case isControl:
			kind := ir.TermKind(op - slots.ControlBase)
			switch kind {
			case ir.Return:
				e.leave(nativeabi.Return, uint64(pc), uint64(pc))
			case ir.Trap:
				e.trap(wasm.TrapUnreachable, uint64(src.Wasm.Offset))
			case ir.Jump:
				e.edge(in.Extra, pcs, pc+1)
			case ir.If:
				e.conditionalEdges(in, pcs, pc+1)
			case ir.Table:
				targets := make([]int, in.Count)
				for i := range targets {
					targets[i] = a.label()
				}
				e.slotLoad(in.A, 0, false, false)
				e.constant(1, uint64(in.Count-1))
				e.compare(false)
				valid := a.label()
				e.branch(valid, "lo")
				e.constant(0, uint64(in.Count-1))
				a.mark(valid)
				a.dispatch(targets)
				for i, l := range targets {
					a.mark(l)
					e.edge(in.Extra+uint32(i), pcs, -1)
				}
			}
		case plainOpcode(op):
			e.plain(in, false)
		case op == 0x40:
			out.MemoryGrows++
			e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			e.leave(nativeabi.GrowMemory, uint64(pc), uint64(pc+1))
		case op == 0x10 || op == 0x11:
			out.Calls++
			var end int
			var expanded bool
			if op == 0x10 {
				end, expanded = e.inlineLeaf(pc, config.DirectCallees[src.Wasm.A])
				if expanded {
					out.InlineSlots = max(out.InlineSlots, uint32(len(config.DirectCallees[src.Wasm.A].Slots)))
				}
			} else {
				var needed uint32
				end, needed, expanded = e.inlineIndirect(pc, config)
				if expanded {
					out.InlineSlots = max(out.InlineSlots, needed)
					out.UsesRuntimeData = true
				}
			}
			if expanded {
				out.InlinedCalls++
				if end != pc {
					independentEntries = append(independentEntries, end)
					combinedThrough = end
				}
				continue
			}
			e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			e.leave(nativeabi.Call, uint64(pc), uint64(pc+1))
		case helperOpcode(op):
			out.Helpers++
			e.contextConst(nativeabi.SourceOffset, uint64(src.Wasm.Offset))
			e.leave(nativeabi.RuntimeHelper, uint64(pc), uint64(pc+1))
		default:
			return nil, wasm.ErrUnsupported
		}
	}
	// A validated function never falls through its final terminator.
	for _, pc := range independentEntries {
		a.mark(pcs[pc])
		if f.Code[pc].Op == slots.ControlBase+uint32(ir.If) {
			e.interrupt(uint32(pc))
			e.conditionalEdges(f.Code[pc], pcs, -1)
		} else {
			e.plain(f.Code[pc], false)
			a.branch(pcs[pc+1], false, 0)
		}
	}
	e.emitPollExits()
	for _, trap := range e.trapExits {
		a.mark(trap.label)
		e.trap(trap.code, trap.source)
	}
	a.mark(e.bad)
	e.rawStatus(nativeabi.BadState)
	a.mark(e.zero)
	e.trapCode(wasm.TrapIntegerDivideByZero)
	a.mark(e.overflow)
	e.trapCode(wasm.TrapIntegerOverflow)
	a.mark(e.exit)
	out.EdgeCopies = e.edgeCopies
	out.BoundsElided = e.boundsElided
	body, err := a.finish()
	if err != nil {
		return nil, err
	}
	out.Entry, err = nativeabi.Wrap(arch, body, limits.MaxCodeBytes)
	if err != nil {
		return nil, err
	}
	for i, l := range pcs {
		out.Offsets[i] = out.Entry.BodyOffset + uint32(a.labels[l])
	}
	return out, nil
}

func isInteger(op uint32) bool {
	return op >= 0x45 && op <= 0x5a || op >= 0x67 && op <= 0x8a || op == 0xa7 || op == 0xac || op == 0xad || op >= 0xc0 && op <= 0xc4
}
func integerWide(op uint32) bool {
	return op >= 0x50 && op <= 0x5a || op >= 0x79 && op <= 0x8a || op == 0xac || op == 0xad || op >= 0xc2
}
func helperOpcode(op uint32) bool {
	return op >= 0x23 && op <= 0x26 || op >= 0x28 && op <= 0x40 || op >= 0x5b && op <= 0x66 || op >= 0x8b && op <= 0xbf || op == 0xd2 || op >= 0xfc0000 && op <= 0xfc0011 || op>>16 == 0xfd
}

func checkFunction(f *slots.Function) error {
	valid := func(s slots.Slot) bool { return s != slots.NoSlot && uint64(s) < uint64(len(f.Slots)) }
	for _, s := range f.Args {
		if !valid(s) {
			return wasm.ErrInvalid
		}
	}
	for _, m := range f.Moves {
		if !valid(m.Dst) || !valid(m.Src) {
			return wasm.ErrInvalid
		}
	}
	for _, e := range f.Edges {
		if uint64(e.PC) >= uint64(len(f.Code)) || uint64(e.Start)+uint64(e.Count) > uint64(len(f.Moves)) {
			return wasm.ErrInvalid
		}
	}
	for _, in := range f.Code {
		if uint64(in.Source) >= uint64(len(f.Sources)) {
			return wasm.ErrInvalid
		}
		for _, s := range []slots.Slot{in.A, in.B, in.C, in.Dst} {
			if s != slots.NoSlot && !valid(s) {
				return wasm.ErrInvalid
			}
		}
		if in.Op >= 0x20 && in.Op <= 0x22 && f.Sources[in.Source].Wasm.A >= f.LocalCount {
			return wasm.ErrInvalid
		}
		if in.Op >= slots.ControlBase {
			if in.Op > slots.ControlBase+uint32(ir.Trap) {
				return wasm.ErrInvalid
			}
			kind := ir.TermKind(in.Op - slots.ControlBase)
			switch kind {
			case ir.Return:
				if uint64(in.Extra)+uint64(in.Count) > uint64(len(f.Args)) {
					return wasm.ErrInvalid
				}
			case ir.Trap:
			case ir.Jump, ir.If, ir.Table:
				if in.Count == 0 || kind == ir.Jump && in.Count != 1 || kind == ir.If && in.Count != 2 || uint64(in.Extra)+uint64(in.Count) > uint64(len(f.Edges)) {
					return wasm.ErrInvalid
				}
			default:
				return wasm.ErrInvalid
			}
		}
		if in.Op == 0x10 || in.Op == 0x11 {
			if uint64(in.Extra)+uint64(in.Count)+uint64(len(f.Sources[in.Source].Outputs)) > uint64(len(f.Args)) {
				return wasm.ErrInvalid
			}
		}
	}
	return nil
}

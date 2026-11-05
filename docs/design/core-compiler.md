# Core Compiler and Execution Design

Date: **2026-10-26**. Status: **implementation contract; see the validation record for implemented increments**. This is the core mechanism supplement to the [architecture](architecture.md), [native ABI](native-abi.md), and [performance contract](performance-contract.md).

## Complete profiles and common pipeline

The first acceptance profile is the complete **Core 2.0 binary execution and embedding contract**, using `WebAssembly/spec` commit `fffc6e12fa454e475455a7b58d3b5dc343980c10` (`wg-2.0`), including `test/core` and `test/core/simd`. It includes all scalar instructions and conversions, structured control flow, calls and indirect calls, globals, memories, multiple tables, reference types, multi-value/block parameters, sign extension, nontrapping conversions, bulk memory/table operations, passive/declarative segments, and fixed SIMD. Text parsing uses pinned development tooling; a native WAT parser and the JavaScript API are not VM requirements.

Later profiles add the full pinned Core 3.0 binary scope, including recursive types/WasmGC, typed function references, tail calls, final EH, extended constants, Memory64/**Table64**, multi-memory, and relaxed SIMD. Core Threads remains a separate profile switch. Every enabled feature must be implemented in the selected backend before compilation succeeds. A declaration, decoded opcode, or interpreter-only implementation does not establish native support.

```mermaid
flowchart LR
    Bytes[Module bytes and limits] --> Decode[Bounded decoding]
    Decode --> Validate[Type and control validation]
    Validate --> Module[Immutable validated module]
    Module --> IR[Typed control-flow IR]
    IR --> Slots[Slot lowering and fusion]
    Slots --> Interp[Portable interpreter]
    IR --> Opt[SSA optimization]
    Opt --> Select[Target lowering and selection]
    Select --> RA[Register allocation and spills]
    RA --> Emit[Machine code and metadata]
    Emit --> Native[Native execution]
    Interp --> Runtime[State, memory, roots, budgets, exits]
    Native --> Runtime
```

`Compile` validates and prepares the requested backend. Interpreter compilation means decoding/validation and bytecode lowering; native compilation means generating **host CPU machine code from guest Wasm**. Building the VM with Go or LLGo is a separate compilation step. No external LLVM optimizer or Rust/C++ Wasm compiler is required for the planned native path.

## Frontend and validated module

Bound module bytes, section/vector sizes, function bodies, locals, nesting, and all intermediate allocations before allocating. Check LEB widths, integer overflow, UTF-8 names, section ordering/uniqueness, index spaces, function/code agreement, and instruction/immediate boundaries. Retain source offsets for diagnostics. Malformed encoding, invalid typing, disabled/unsupported feature, and resource exhaustion have distinct errors. Declared maxima never override host quotas.

The validator implements operand and control stacks, stack polymorphism in unreachable code, block inputs/results, branch arity, mutability, import/export types, constant expressions, segment rules, and feature dependencies. Parsing does not grant permission to execute. Only validator-produced immutable modules can enter lowering or instantiation. Reference types and signatures are canonical identifiers, with Store ownership checked separately at runtime.

Instantiation resolves every import, checks Engine/Store/type domains and limits, allocates state, validates initializer/data/element ranges, and executes the start function with the ordinary execution budget. Failure releases allocations and references created by the attempt. Side effects already performed on imported memory, tables, or by host functions follow the specification and are **not** transactionally rolled back. Preserve declaration and initialization order; do not run a start function during compilation or benchmark setup while claiming it is excluded.

## Typed control-flow IR

Use basic blocks with typed parameters, explicit branch edges and argument lists, and values identified by dense indexes. Lower the validated operand stack into SSA value identities; retain mutable locals through explicit local operations initially, then promote eligible locals through a verified pass. Block parameters express merges; branches and parallel copies must handle cyclic moves. A verifier checks dominance, types, block arity, terminators, source locations, and definitions after each transforming pass in tests/debug builds.

Each operation records effects: reads/writes to a particular state domain, potential trap, host call, allocation, suspension, and original Wasm fuel charge. Memory operations identify the memory and address width; table operations identify the table and index width. Calls distinguish direct, indirect, host, and later tail calls. Keep numeric values and reference handles typed throughout lowering; never erase references into anonymous scalar spill slots.

Do not fold a trapping expression into a compile-time module rejection when the specification traps only upon execution. Dead-code removal, common-subexpression elimination, constant folding, bounds-check elimination, and inlining must preserve observable effects, trap order, permitted floating-point outcomes, and resource policy. No unsafe fast-math. Preserve source maps from generated instructions to guest operations for traps, traces and metering.

Resource policy is part of compilation configuration and the cache key. With deterministic fuel enabled, charge original Wasm instructions at their semantic boundaries, before their effects. Fused instructions retain sub-operation costs and exits. Coarse prepaid blocks are allowed only where identical exhaustion/side-effect behavior is proved; otherwise retain finer checks or use a separately named budget mode. Optimizing away a loop must not erase its configured fuel exhaustion behavior. Cancellation polling is a separate bounded-latency contract.

### Implemented IR contract

`internal/compiler/lower` lowers every statically valid Core 2.0 function to
`internal/compiler/ir`. Instruction schemas come from the validator's typed
walk; there is no duplicated opcode typing table. Sources retain byte offsets,
immediates, effects and input/output types. Mutable locals remain explicit.
Control edges carry ordered tuples, and unreachable blocks/values are removed
without executing or folding trapping operations.

The verifier checks single definitions, dominance, use order, function/local
types, operand and edge types, terminators, reachability and effect metadata.
It uses iterative traversal and bounded dominance work. Node and verification
budgets apply across the whole module; zero is a literal zero budget. Source
schemas are internal trusted compiler data, not a serialized input format.

Fuel model version 1 charges one unit before each executed Wasm instruction,
except `else` and `end` delimiters. Compiler-generated transfers have no extra
charge. A loop header is charged on entry from outside; a backedge charges its
branch and the subsequent body instructions. These costs are retained in IR;
runtime metering and cancellation are subsequent implementation work.

## Interpreter lowering and hot loop

Lower SSA values and locals into reusable typed slots with precomputed frame sizes. Resolve branches, function indexes, local offsets, immediates, memory offsets and stack effects at compile time. Keep immutable instruction arrays compact and contiguous. A separate read-only pool holds uncommon wide immediates/branch tables. Dispatch uses a Go `switch`; ordinary Go has no portable computed-goto contract. Do not implement dispatch through per-op closures, reflection, or `interface{}` values.

The execution hot loop uses a program counter, explicit frame base and slot arrays. Operand reads are guarded direct slot loads; backend/metering selection is fixed for each Run invocation. Immutable source/instruction metadata is borrowed so scalar dispatch does not copy wide SIMD immediates. Scalar slots hold exact bit patterns; reference slots hold checked handles with type/root metadata. Frame allocation, locals, outgoing arguments, results, and spill areas use pre-sized storage, growing only at explicit checked boundaries. Guest calls push explicit frames rather than recursively calling Go. Return and `br_table` copy overlapping tuples correctly. Preserve suspended frame state across host calls and resume.

The active PC and bounded frame-value view now stay local to that loop. Publish
the caller's next PC before guest push or host suspension, and refresh both
views at every frame transition or Run entry. Arguments copied after push read
the preserved global arena, since growth can invalidate the former view.
Nontrapping integer instructions call shared small numeric primitives directly,
avoiding the generic evaluator's second opcode dispatch. Typed-bit instantiation
exposes the same canonical implementations to LLGo's package optimizer; it is
not a separate semantic evaluator. Trapping and FP operations keep the generic
evaluator. Source/fuel/trap identities and all polls remain unchanged.

Ordinary GC access consumes the predecoded scalar operands and shares a
nonallocating typed heap helper. Allocating and bulk/cast operations use normal
methods rather than constructing unrelated local callbacks. A temporary root
collector is installed only when no embedding Store visitor exists. Slow
operands borrow cleared per-execution scratch, retaining at most 64 values
(1 KiB); larger validated tuples remain transient. Constructor roots are
restored on errors and collector panics. Fuel, source PCs and poll boundaries
are unchanged. See the [GC interpreter comparison](../research/performance-gc-interpreter-2026-11-02.md).

The LLGo release profile uses `-O3 -lto=full` for engine/CLI builds, the native
probe and semantic CI. Source comparisons apply identical flags to base/head.
Default `-Os` diagnostics retain observed regressions; an inlined helper or a
release-profile improvement does not establish a default-build performance win.

Start with a simple verified lowering. Measure register/slot dispatch against a predecoded stack-dispatch prototype using identical complete semantics. The default design is slot bytecode; change it only with measured evidence covering both compilers, dispatch cost, bytecode size, compile latency, and representative workloads. Do not maintain two permanent production interpreters merely to preserve the experiment.

Optimize demonstrated bottlenecks with direct local/constant forms, copy elimination, block layout, and bounded superinstructions. Fusion must preserve fuel, trap locations, memory effects and safepoints. Pre-resolve internal direct calls; indirect calls retain null, bounds, signature, ownership, and instance checks. Poll at calls, backedges and a maximum straight-line work interval; chunk bulk operations. An unmetered mode still observes cancellation and resource limits. There is no unchecked performance mode for untrusted guests.

Keep allocation out of warmed numeric execution after frames/results are reserved. API allocations, growth and host callbacks are measured separately. Private linear memory can use inlined checked access after profiling; shared memory must retain the common synchronization contract. Check address-plus-offset overflow before bounds and host-index conversion. Preserve the specified failure result for expected grow failures.

### Implemented slot lowering

`internal/compiler/slots` emits flat 32-byte instructions, inline ordinary
operands/results, and separate pools for call/return arguments, edges and moves.
Every emitted instruction retains its Wasm source/fuel identity. Blocks have
resolved instruction offsets. Function locals occupy the frame prefix; scalar,
vector and reference slots keep distinct types for future root scanning.

The initial allocator recycles block-local temporaries by type and last use.
Cross-block values and block parameters have dedicated slots. This conservative
policy bounds compilation without constructing a global interference graph;
whole-function allocation is a later measured optimization. Results do not
clobber input slots at an instruction or suspended-call boundary. Parallel
edge copies use a linear expected-time resolver and typed scratch for cycles.
Per-function frame slots and aggregate lowering/verification work are bounded.

An unconditional forward edge to an empty return block can return its incoming
tuple directly. This keeps every original instruction PC and the predecessor's
source/fuel charge while removing the intermediate tuple copy and dispatch.
Backward edges retain their cancellation poll. Conditional, effectful and
reordered/subset return blocks keep their original transfer plan. Native leaf
analysis stops at the reached return, preserving call expansion when a former
return block remains independently addressable afterwards.

Symbolic tests independently simulate slot reads, writes and edge transfers for
every valid corpus function. This establishes lowering preservation, not
instruction execution, runtime rooting or interpreter performance. The complete
Core 2.0 execution corpus and current performance limits are recorded separately
in the [validation record](../validation.md).

## Native compilation

The complete native pipeline is: typed IR verification, optimization, legalization to target operations, instruction selection, register allocation, spill insertion, frame layout, branch resolution/relocations, code emission, and publication with trap/safepoint/source maps. Implement amd64 and arm64 emitters in Go. The target allocator requires linear scan with separate integer/vector classes, fixed-register constraints, call clobbers, live-range splitting, parallel-move resolution, and verified spill reloads. Target tests check actual execution as well as encoding/disassembly.

The conservative allocator assigns stable homes by bounded CFG loop-use
priorities and also selects the immutable metadata inventory. Merged backward
edge intervals provide capped frequency hints, including nested/overlapping
spans. They do not prove reachability or justify removing a check. Tuple/move
and expanded-leaf literal scans have separate work budgets and conservative
priority fallbacks. Candidate ordering is deterministic by score, class and key.
AArch64 integer/reference homes compete with immutable literal/runtime words
before the inventory fills: homes prefer x19-x28, while all classes can use
x1-x8. FP/vector homes retain their separate shared pool. Immutable words reload
on every entry, including missing-view zeros for checked indirect fallback. Leaf
expansion excludes them through the same target inventory. Every mutable
allocation reloads at entry and spills on all exits; reserved ABI registers and
the separate FP/vector bank remain excluded. Scalar `select`
now uses CMOV/CSEL and directly names allocated integer/reference registers,
falling back to scratch loads for spilled or FP bit patterns. It tests only low
i32 condition bits and preserves source/destination aliases. The precise tier
selects both input words; unmetered scalar storage follows its existing
canonical-input contract. V128 retains its complete vector transfer path.

Guarded loop regions can select a different integer/reference location plan
using the same bounded slot-use priorities. Common homes keep their registers;
every displaced word reaches the canonical frame before any replacement loads.
Private backedges retain the local plan, while ordinary loop exits restore the
enclosing plan. Deferred poll, trap and call exits capture the plan at their
originating instruction, preserving complete snapshots and original PCs.
Checked cold/interior entries retain an explicit mapping. Entry/exit transfer
costs gate loop-region selection; FP/vector homes keep their incoming mapping
within this local transform. See the [loop-plan evidence](../research/performance-loop-register-plans-2026-11-05.md).

The subsequent global slot-residency pass scans conservative first/last access
intervals over basic blocks, with separate integer and shared FP/vector banks.
Expiration and furthest-end eviction split residency; a spilled slot can regain
a register later. Every displaced word reaches its canonical frame before any
replacement loads. Branch copies retain the existing typed move plans, while
mapping transitions belong to their actual source edge. One shared cold-entry
transition per distinct plan preserves every original continuation without
duplicating the full spill/reload sequence at every PC. Poll, call and trap
exits retain all value classes and original source identities. Mutable homes
load for the chosen cold entry; immutable metadata remains function-wide.

Tuple traversal and materialized map cells have independent work bounds;
exhaustion retains the conservative allocator. The pass preserves dead frame
words instead of treating its residency intervals as semantic liveness. It
does not itself construct SSA value-version live ranges or split allocation
inside a basic block. See the
[global-plan validation and static comparison](../research/performance-global-register-plans-2026-11-05.md).

A subsequent bounded loop pass versions each write, includes dead definition
points and observes every final frame word. Copy-connected slots of the same
storage type can share a physical home when all overlapping versions have
identical bits. A distinct initial entry plan spills displaced words before
loading shared homes. The body retains complete integer/reference, scalar FP
and V128 snapshots at its poll and ordinary exits. Pairwise interference work
has a body-proportional bound; exhaustion keeps the incoming plan. Shared-home
writes split instruction selection, while read-only sequences retain fusion.
Vector aliases expire on every write to their physical source register.

The private-loop emitter handles arithmetic and already-proven memory regions
alongside the existing affine whole-loop proof. The memory proof is optional;
helpers, traps, interior entries, edge copies and poll-window crossings remain
ineligible. Both conditional polarities, copied boolean conditions and
unconditional loops use the same bounded poll and cold-entry contracts. This
is local copy coalescing; whole-function SSA allocation, more general live-range
splitting/spills and performance acceptance remain required. See the
[loop-value validation and actual machine-code comparison](../research/performance-loop-value-coalescing-2026-11-05.md).

Use the platform's accepted baseline CPU feature set, with explicit detection and recorded features for optional instructions. Specialized artifacts cannot run on a weaker CPU. Floating-point/vector operations require correct Wasm NaNs, signed zero, rounding, conversions and lane behavior. Division and conversion instructions often need explicit prechecks; host instruction traps and language overflow behavior are not automatically Wasm semantics.

Initially emit explicit memory/table/stack bounds checks and explicit trap exits. Signal-handler recovery is not the default safety mechanism. Eliminate checks only with validated range/effect proofs covering growth, calls, 64-bit overflow and aliases. Imported or mutable function/table targets cannot be treated as immutable without valid guards. Mark the entire profile supported only when each operation has verified native lowering or a documented runtime helper with the same semantics; helper use is reported in performance evidence. Native execution must never interpret an unsupported function silently.

The unmetered straight-line pass groups at least three affine accesses anchored in unchanged
i32 locals within its existing <=256-instruction, positive-minimum-memory
regions. Guards read the first access's existing address operand and validate
the entire span against memory32 and the current binding. Failed guards enter
the original checked instruction sequence at that access, preserving earlier
effects and the actual first trap. Interior cold PCs keep all checks. The precise
tier and unmatched addresses retain their existing lowering.

After proving no address addition wrapped, the emitter folds biases into memory
offsets. Bounded backwards liveness removes only dead i32 constant/add temporary
writes; guards and possible traps restore full physical-frame requirements.
Remapped accesses and removed instructions bound older fusion selectors. Fixed
register weights reflect the emitted reads/writes, including guard operands.
For arm64 spans up to 4096 bytes, let `u = uint32(address - firstBias)` and
`span = firstBias + after`. One unsigned interval check rejects both subtraction
underflow and an end past memory32. A widened end check against the bound memory
still admits an exact 4 GiB endpoint. Larger spans and amd64 keep the general
lowering; guard failure preserves the original operand and checked continuation.

Fusion stops at remapped accesses and deleted operations. Physical-state scans
may cross accesses already proved safe, but stop at every later guard, possible
trap, poll, CFG edge, helper and region end. These separate bounds allow dead
literal removal without granting a fusion selector an original, stale address.
AArch64 shift-plus-ADD/SUB/AND/OR/XOR selection uses the same observation scan;
it retains live constants, rejects observed intermediate values and preserves
aliases and all original cold PCs. Rotate forms apply only to logical operands.
Local versions invalidate stale facts; CFG/helper/growth/poll boundaries reset
analysis. All memory operations still execute in order.

A bounded loop analysis also recognizes one-block, literal-bounded i32 equality
exits with independent invariant bases and a nonzero unsigned stride. Calls,
growth, possible arithmetic traps, internal CFG entries, mutated bases and
nonlinear addresses retain ordinary checked code. The whole candidate,
including its branch, stays within one 256-instruction interval; it does not
require three accesses sharing a root.

At the original header, require `i < limit` and `(limit - i) % stride == 0`.
For each affine group, compute its initial i32 origin with wrapping arithmetic,
then check the widened origin plus its complete future displacement and
nonwrapping access extent against memory32 and the current binding. A folded
address additionally proves its canonical local anchor and whole folded offset.
Guard work is bounded and polled. Failure enters the original checked header
without effects or an early trap. Success enters the shared body with private
backedges and no repeated span checks. All interior dispatcher PCs keep their
checked originals; backedge interruption and physical-frame requirements stay
intact, and driver returns lose the proof. The precise fuel tier is unchanged.
See the [affine-loop comparison](../research/performance-affine-loop-spans-2026-11-05.md).
Global live-range allocation and broader loop/range optimization remain required.

Use a private guest ABI with explicit VM frames: guest calls push checked logical frames and branch to native callee labels; returns branch to saved continuations. The initial backend keeps one bounded C entry frame, without hardware-stack switching or a native C/Go frame per guest call. Host interaction, collection, cancellation and traps leave through the explicit status protocol in the [native ABI](native-abi.md). Whole-function inlining and more advanced optimization follow correctness and profiling; both disabled and enabled optimizations must pass differential tests. A persistent AOT/cache file format is later work; generating native code in-process is required in M2.

The retained backend-contract experiment requires `Resize`, `RunFrames`,
`Interrupt` and `Close` together. Its shared driver validates explicit frame
packets through one noninlined native bridge; it has no optional frame-capability
assertion, capability cache or legacy single-frame fallback. Portable execution
keeps a nil backend. Frame transfers, canonical state, roots, fuel, cancellation
and owned storage retain the same runtime contracts. The low-level single-frame
convenience entry is separate from this driver protocol. The experiment removes
repeated LLGo method-table allocations without adding execution-state fields,
but overall performance and final-head integration gates remain open. See the
[source, semantic and performance evidence](../research/performance-native-backend-contract-2026-11-05.md).

The subsequent [shared-slot dispatcher experiment](../research/performance-slot-dispatch-driver-2026-11-05.md)
moves native transfer into the outer driver. Both backends retain one instruction
handler; a native exit services exactly one already-charged boundary and publishes
its continuation before reentry. Portable execution keeps its frame-local loop.
This improves measured interpreter workloads but remains held for native/portable
regressions and final acceptance. Native effect boundaries still enter the large
slot-dispatch stack frame; a separate runtime-effect handler shared by both
drivers motivates the next experiment, alongside global allocation and the other gates.

The [runtime-effects experiment](../research/performance-runtime-effects-2026-11-05.md)
removes the portable boundary mode. Native exits select shared call/return,
exception, stack, GC and SIMD effects, with complete next-PC/frame publication.
The portable scalar/local/CFG loop keeps its direct execution, sharing effect
handlers and canonical state primitives. Ordinary scalar native exits fail
instead of falling back to that loop. Smaller static frames still introduce
extra transfers and do not establish throughput gains: GC, SIMD, atomic and
public-call regressions keep this experiment held. Direct native leaf routing,
portable primitive routing and adequately controlled measurements remain work.

The [direct-routing follow-up](../research/performance-direct-runtime-routing-2026-11-05.md)
selects native GC/SIMD/call/throw/return leaves in the outer driver, avoiding the
general dispatcher prologue on those exits. Portable SIMD, GC access and atomic
operand routing stays in its hot loop, calling the existing canonical state,
memory and SIMD primitives. Complex effects still share their leaves and full
PC/frame/fuel/root/cancellation contracts; scalar native fallback remains invalid.
Local dual-compiler semantics pass. The hosted comparison and all final-head
platform, coverage, review and controlled performance gates remain separate.

## Direct globals

Both emitters lower validated numeric and reference global.get/set in precise and unmetered
entries. An execution-owned native view provides per-index cell pointers;
imported aliases share cells. Scalar precise copies preserve both words, while
unmetered scalar values retain the canonical zero high word. V128 accesses use
allocated vector operands and exact 128-bit loads/stores, including spilled
values. Allocated scalar integer/FP registers load/store directly, avoiding
integer/FP bit transfers. Unmetered scalar stores retain canonical zero words
without rewriting them. Spare arm64 caller registers cache immutable pointers, sharing the
existing allocator with table metadata. Mutable cell contents are always live.

The driver snapshots only referenced globals before Run and publishes
writes after every native exit, before host/helper handling, root tracing or errors.
Reference cells carry Store-local handles. Reentry refreshes cells after nested
guest/host writes. Missing views fail before any guest
instruction and original cold continuations retain the same contract. See the
[scope](../../benchmarks/native/globals-plan.md) and [native ABI](native-abi.md).
This removes helper transitions; broader global live ranges remain open.

## Direct GC field access

Both emitters lower struct.get/get_s/get_u/set, array.get/get_s/get_u/set and
array.len against the same packed field layout as the portable runtime. Each
access checks null, slot bounds and the complete generation-tagged identity
before loading a payload address; arrays also check the unsigned i32 index.
Validated subtype prefixes retain their physical layout. Packed signed loads
produce canonical i32 bits, and vectors use the existing vector allocator.
Precise fuel, trap offsets, cold continuations, spills and root maps retain
their ordinary contracts. Allocation, casts and bulk GC operations continue to
use typed runtime exits.

Native Stores own a compacting payload arena and a shared numeric object-header
view. Collection and arena/view growth run only at serialized safepoints and
update shared payload pointers before reentry. Guest handles never change when data
moves. No generated store needs a Go or incremental guest write barrier:
payloads contain numeric handles, and the collector traces the same bytes after
native execution stops. Hidden descriptor edges, exception tags and continuation
metadata keep their existing typed root visitors. Portable Stores keep Go-backed
payloads without native relocation metadata in their object size class.

Payload capacity is bounded by MaxHeapBytes, header-view virtual capacity by
(MaxHeapObjects+1)*32 bytes. Payload growth transiently retains at most two
copies; stable header growth enables pages in place without moving records.
Logical object/descriptor/continuation charges remain unchanged. Compaction
also bounds retired allocation metadata after failed instances or handle reuse,
independently of arena high-water capacity. Store close frees both native
allocations after execution drains. The [shared header contract](shared-heap-header.md)
specifies typed portable roots, native backing and complete batch publication.
See the [GC comparison](../research/performance-gc-2026-11-02.md)
for measured gains and remaining deficits.

The [heap-traffic follow-up](../research/performance-gc-followup-2026-11-02.md)
adds entry-local checked-address caches on arm64 and bounded store/read
forwarding on both emitters. Cache keys reset at every entry and remain separate
from guest register homes. Forwarding preserves all original cold PCs and
checks, packed/vector bits, poll intervals and precise-tier behavior.

Adjacent `ref.i31` and `i31.get_s/u` operations also forward their integer
payload in the unmetered tier. Preserve a boxed slot that later exits can
observe, while extracting from the original integer to avoid extending its
dependency chain through tagging. Cached operands use direct ARM64 AND/SBFX
or AMD64 unsigned immediate AND; spilled/signed-AMD64 operands keep generic
lowering. Original interior PCs retain independently checked entries, and
combinations cannot cross a poll boundary. The [i31 comparison](../research/performance-i31-payload-2026-11-03.md)
keeps all source stages, ordinary C, competitors and adverse controls; one
kernel's gain does not close portfolio acceptance.

## Direct tables and function references

Both emitters lower table.get/table.set and ref.func in precise and unmetered
entries through the shared runtime view. Table length checks precede scaled
addressing and preserve exact source/trap order. Reads and writes use the same
native table cells as the Store's reference tracer, including imported aliases.
Function IDs come from the current instance's immutable handle array. No Go
pointer or machine function address is encoded as a guest reference. Spare
arm64 metadata registers share the existing allocation pool; all element values
remain live reads. Allocated reference registers access table cells directly;
arm64 uses an extended length comparison and UXTW addressing to consume i32
indices without changing their raw upper words. Precise and spilled paths share
the element encoder with guarded indirect calls. Reference slots retain the
conservative spill/root protocol.

Ordinary entries request their own runtime view independently of optional leaf
expansion. Host reentry, table growth and bulk helpers rebind on the next entry;
empty, Go-backed and retired storage retain distinct boundary outcomes. Bulk
operations and growth continue through explicit owned helpers. See the
[scope](../../benchmarks/native/tables-plan.md) and [native ABI](native-abi.md).
This removes the basic table/reference helper transitions, without accepting
general call, register-allocation or whole-profile performance.

## Implemented native struct allocation

Ordinary native entries may initialize fixed-layout struct objects in batches
of at most 64 reserved handles. All capacity/quota guards precede writes;
field initialization precedes the output and cursor commit. The existing driver
publishes every completed batch into the shared typed collector on every exit,
including trap/cancel/error paths, before helpers or host reentry. No payload
relocation, reclamation or live-handle reuse occurs inside an entry. Precise
fuel, forced GC, descriptors and unsupported shapes retain complete helpers.
The private view contains only numeric IDs and owned native addresses; the
stable entry record ABI is unchanged. See the
[allocation evidence](../research/performance-native-allocation-2026-11-02.md).

## Implemented private memory primitives

`internal/engine/memory` enforces Core 2.0 page limits, independent host quotas,
widened effective-address arithmetic and host-index bounds. It implements all
scalar and fixed-SIMD loads/stores, growth, host copies and bulk copy/fill/init.
Every rejected range is detected before mutation. Overlapping copies preserve
snapshot semantics. Lane loads preserve other lanes; zero loads clear the
remaining lanes. Shared-memory synchronization is outside this Core 2.0 package.

Buffers have exact requested capacity. Growth preserves existing bytes and
zeroes the new pages; quota failure leaves memory unchanged. The old and new
buffers may coexist transiently during copying and until host collection, so
the page quota bounds logical memory rather than process RSS. Store-wide memory
accounting and scheduling of chunked bulk operations remain runtime work.

Interpreter dispatch resolves original source metadata only when required for
semantics, fuel or traps. Existing instruction words cache scalar constant bits
and single indices; instruction size remains 32 bytes. Guest frames write
directly into retained capacity, avoiding LLGo append temporaries on ordinary
calls. Context binding caches Done once; a nonblocking poll calls Err only after
closure, while nil-channel contexts retain Err polling. Rebind/quiescence update
both references. The [three-way gap audit](../research/performance-interpreter-dispatch-2026-11-02.md)
retains competitor, host-compiler and ordinary native C comparisons, including
regressions and the rejected unchecked-frame experiment.

Interpreter loads keep lane updates in value form, preventing a vector helper
from forcing scalar results onto the host heap. Fixed opcode widths are read-only
metadata, and scalar dispatch uses a smaller frame than the vector helper. The
[paired comparison](../research/performance-interpreter-memory-2026-11-02.md)
retains allocation-only and frame-outlining results separately.

Ordinary native entries lower private single-memory `memory.fill`/`memory.copy`
directly on both targets, including i64 addresses/counts. Complete checked ranges
precede all writes. arm64 uses paired vector accesses with exact tails; amd64
uses REP strings and restores direction/context/frame state on every exit.
Overlapping copies retain direction, and at-most-4096-byte chunks poll the
existing cancellation word before writing. Precise, shared and multi-memory
operations retain their runtime helpers. The existing state/root/lease contracts
and code-size limits apply without another native ABI or execution framework.

## Implemented native shared memory and atomics

Shared memories reserve a stable, host-quota-bounded native extent. Their
accessible length lives in a separate atomic native cell; every emitted bounds
observation reloads it, including during a running entry. Both targets directly
emit all scalar atomic widths, RMW/CAS and fences. arm64 selects LSE only after
CPU probing and otherwise retains cancellation-polled LL/SC loops. Wait mismatch,
zero timeout and empty notification emit checked paths; blocking and queued
notification use the existing bounded domain helper. Precise fuel charges the
original instruction once across retries and helper resumption.

Host/interpreter shared accesses use opaque ISA instructions to preserve Wasm's
mixed-size overlapping access semantics. Portable backings retain domain locks.
Pending waiters publish a native reservation before their SC value comparison;
a notifier observing a reservation takes the ordering lock. Native allocation
ownership, logical growth, queue cleanup and per-Store heaps remain separate.
See the [implementation and comparison](../research/performance-native-threads-2026-11-02.md).

## Implemented Core 2.0 execution

`internal/engine/state` owns typed imports, canonical signature identities,
Store-local function references, globals, tables, memories and segment state.
Allocation resolves every import before side effects and rolls back its new
state on failure. Initialization follows active elements, declarative drops,
active data and a separately driven start call. Earlier imported-state effects
survive later traps; escaped functions retain their originating instance.

`internal/engine/execute` executes flat slot code with explicit bounded guest
frames. Host calls return requests to a driver, which may run a nested
continuation and then resume the suspended caller. Frame and pending-host
reference slots have root visitors. All opcodes use the owned numeric, SIMD,
memory and table semantics. No competitor runtime is invoked.

Fuel model v1 charges before effects. Cancellation is checked at entry, calls,
backedges, at most every 64 straight-line instructions, and between 4-KiB memory
or 512-element table chunks. Bounds traps precheck entire bulk ranges;
cancellation may leave partial progress. The execution object can be reused;
a warmed scalar call has zero allocations in the Go allocation-counter test.

These internal mechanisms pass the complete binary corpus on both supported
platforms under Go and LLGo. The [public driver](../embedding.md) now supplies
single-executor admission, callback reentry, shared fuel/frame quotas, opaque
host roots, safepoint collection, cancellation and deferred close. Active and
suspended frames preserve instance ownership. Imported effects and escaped
functions survive failed instantiation; unreachable failed cycles are reclaimed.
Canonical types retain a bounded Store-lifetime cache, while function identity
holes and empty memory/table objects have explicit quotas. Public call costs
are separate from the internal dispatch allocation test and remain to be
measured by the own-engine benchmark integration.

## Runtime, lifecycle and observability

Both backends use the same Engine/Store/Instance ownership model, memory/table/heap domains, host signature checks and typed outcomes. Compiled artifacts are immutable and reference counted or otherwise retained until all instances and active/suspended executions release them. Closing an Engine prevents new work and cannot unmap code used by a continuation. Store `Close` is an idempotent, nonblocking close request: prevent new work, signal cancellation and resolve waits once, then reclaim after active callbacks/executions leave. It is safe from a host callback and never joins itself. A separate `WaitClosed(ctx)` is for an external caller to await reclamation with a deadline; calling it from an active callback of that Store is outside its contract. Host-boundary deferred cleanup releases reentry tokens, temporary roots and execution/code leases before propagating a host panic, allowing the caller to recover without leaving a busy Store. An arbitrary blocking user callback remains outside any promised preemption bound.

Define explicit exits for return, host request, trap, budget exhaustion, cancellation, allocation/growth, and suspension; later add exception and async state without using Go panic as guest control flow. Root maps cover active and suspended frames, locals, temporaries, host handles and exception payloads. Every collection/reentry boundary spills live guest references before allocation or entering the host. Keep Go objects rooted in Go-managed tables; machine registers and native stacks contain handles rather than sole copies of host pointers.

Debugging must retain module/function/byte-offset and guest stack information without depending on native PC readability after code is freed. Benchmark phase timers wrap actual API or internal phase intervals; phases that overlap through lazy compilation/tiering are reported as such. The first backend is eagerly compiled with tiering disabled, so first-call and warmed-call costs have clear meaning.

## Tests and completion evidence

| Layer | Required checks |
| --- | --- |
| Decode/validate | Complete pinned profile tests including malformed/invalid/unlinkable inputs; truncation/LEB/allocation fuzzing; no panics for hostile bytes |
| IR/optimization | Verifiers; branch/tuple/loop cases; differential optimized/unoptimized execution; fuel, trap and side-effect order; legal NaN outcome sets |
| Interpreter | Every opcode and trap; explicit frame/stack limits; host reentry, growth and cancellation; full Core 2.0 corpus |
| Native | Same corpus on both CPUs/host compilers; spills, long branches, large frames, register clobbers, SIMD and indirect calls; ABI/lifecycle stress |
| Runtime | Import identity, aliasing, partial instantiation failures, start traps, repeated close, close from a callback, host panic/recover cleanup, suspend/resume, memory quotas and stale/foreign handles |
| Performance | Own base/head and competitors; isolated subsystem workloads plus representative applications; raw results and compiler/platform metadata |

Run semantic tests with Go and LLGo, and retain Go race/fuzz diagnostics separately. Target at least **90% statement coverage per owned executable core package**, with uncovered paths explained; assembly/machine-code correctness requires target execution tests rather than that percentage. Coverage and passing differential samples cannot substitute for a complete conformance manifest. Missing tests or unsupported operations block profile acceptance. [The roadmap](../roadmap.md) closes a milestone only after all its correctness, compiler, lifecycle, and performance gates pass.

## Individually enabled constant expressions

`EngineConfig.Features` is an explicit additive `wasm.Features` mask over the
complete default `core2` profile. `wasm.ExtendedConst` enables the six i32/i64
add/sub/mul instructions in constant expressions. Unknown feature bits fail
with `ErrUnsupported`; enabling this flag does not select full Core 3.0.
`Inspect` and compiled metadata report the enabled mask; the CLI accepts
`--extended-const` for validate, inspect and run.

Validation uses a typed operand stack, permits only imported immutable globals,
requires exactly one result of the destination type and applies `MaxStack` to
initializer depth as well as function bodies. Decode node limits bound total
initializer work. Validation records the maximum initializer depth. Allocation
and active-segment initialization each reuse one bounded scratch stack, with
small expressions on local storage. Evaluation uses unsigned wrapping integer
arithmetic, preserves subtraction order and resolves imports per instance.
Both backends share this path, including segment trap ordering and cleanup;
initializers never execute a guest function or consume function fuel. Entirely static
integer expressions fold once during bounded lowering into an immutable
expression-identity-to-bits map, leaving the validated definition untouched.
Instance-dependent expressions use the same wrapping arithmetic primitives.
Private globals share one instance-lifetime cell allocation and pre-sized
pointer views; host globals retain their independent lifetime and quotas.

The full archived feature proposal corpus is pinned in
[testdata/extended-const](../../testdata/extended-const/README.md), alongside
explicit default-profile rejection, resource, wrapping, cross-instance and
public API tests. Core 3.0 profile acceptance remains separate.

## Proper tail calls

`wasm.TailCall` enables the two instructions in the original Tail Call proposal.
The feature bit reaches decoding and validation independently: default Core 2.0
still rejects their binary encodings. Tail calls require exactly matching return
types, permit different parameter tuples, and make following code unreachable.
The function-references proposal's `return_call_ref` remains a separate feature.

Lowering emits a terminal `ir.TailCall`, verified against its typed operand
and callee return schemas, with no intermediate result slots. Slot code retains the original Wasm
opcode and source location, and stores the complete input tuple. Both native
policy tiers publish a checked call boundary; no whole-function interpretation
or competitor fallback is involved.

The common frame machine resolves and checks the target before retiring the
caller. A reusable pointer-free tuple buffer protects overlapping arguments
before frame storage can grow or be reused. Replacement checks the callee's
actual slot requirement at the caller's former base, rather than requiring
space for both frames. The callee inherits the caller's original destination.
An ordinary call continues to consume a new frame.

A host tail transfer retires the guest frame before suspension. `HostCall.Caller`
keeps the originating instance available to host memory access and collection,
including across module boundaries. Pending arguments remain typed roots;
the removed frame's locals do not. Host reentry sees the released frame quota.
Each tail instruction consumes one fuel unit, and all transfers poll cancellation.
Native direct tail-entry optimizations remain local follow-up work.

## Multiple memories

`wasm.MultiMemory` enables multiple imports/definitions and indexed scalar,
fixed-SIMD, size/grow, active data and bulk operations. Memarg bit 6 selects an
explicit memory index before the offset; alignment remains separate. Default
Core 2.0 retains its original grammar and single-memory validation rule.

Typed IR preserves the memory identity in every memarg, and validation checks
the selected memory independently. Native code selects entry-stable length/base
pairs in the existing ABI view; the interpreter uses the same module indices.
Bulk copy validates both complete ranges before writing, copies between
independent memories and preserves memmove ordering when imported indices alias
the same object, including across scheduling chunks. Store page/memory quotas,
failed allocation rollback and host callback capabilities remain unchanged.

## Individually enabled relaxed SIMD

`wasm.RelaxedSIMD` enables the complete 0x100–0x113 instruction range. Reserved
opcodes and all prototype encodings remain rejected. The existing vector
validator records exact unary, binary and ternary types, including unreachable
stack handling; IR, slot lowering, quotas, fuel and cancellation stay shared.

The selected projection uses fixed-SIMD swizzle, saturating conversions,
bitselect, strict min/max and Q15, separately rounded multiplication/addition,
and signed byte products. Both native targets use that same policy. Existing
strict legalizations are reused; multiply/add and byte dots generate packed
native operations. Four-byte dots widen before pair sums, including
`-128 * -128`, then accumulate modulo i32. Two-byte dots wrap at i16 only after
the pair sum. Baseline amd64 needs only SSE2; arm64 needs baseline NEON.
Target-specific fused or relaxed instructions remain local optimizations.

The full pinned seven-script feature corpus contains 85 binary assertions.
The runner matches each WABT `either` as one complete permitted result vector,
never as a Cartesian product of independently accepted lane values. Additional
tests cover every opcode, default-off and reserved grammar, exact arity and
operand types, all native destinations/spills/cold continuations, code-size
limits, all four feature flags together, single-frame tails, fuel, cancellation
and workspace reuse. Go and LLGo run the same semantic matrix.

## Individually enabled Memory64 and Table64

`wasm.Memory64` enables 64-bit limits and address/index types for memories and
tables. A module limit stores `uint64` bounds and `Index64`; the validator
selects i32/i64 for scalar/SIMD addresses, size/growth, indirect calls, active
segments and bulk operations. `memory.init`/`table.init` keep i32 segment
source/count operands. Cross-memory/table copies take destination and source
address types separately, with the narrower count type.

The decoded instruction retains the offset's low word in `B` and its upper
word in `MemoryOffsetHigh`, filling padding on both native targets without
increasing their instruction size. Existing i32 affine proofs and code remain
available. I64 accesses currently retain complete checks: carry from the
address plus offset and from the access end traps before adding the native
base. Scalar and all fixed-SIMD accesses are generated on both native targets.
Owned embedding tests additionally execute all 22 SIMD memory forms with i64
addresses, lane preservation, high indices, traps and complete memory effects.
Table get/set and guarded indirect leaf calls retain full index bits before
length comparison; checked frame dispatch uses the same width contract.

Growth checks the remaining quota/maximum before addition. A declared 64-bit
maximum never grants an unbounded allocation: host limits and addressable
storage remain independent bounds. Failure returns all one bits in the
object's address type. Import matching includes that type, and every host
reentry refreshes existing trusted descriptors after growth.

The feature corpus contains all nine dedicated archived memory64 scripts, all
three memory bulk scripts and all eight table-operation scripts, plus `call_indirect.wast`. All 13,069
binary assertions in those 21 complete scripts run on every execution tier;
70 malformed-text assertions are classified separately. The sole upstream
`module definition` command compiles without instantiation, preserving the
intended validation-only check for a minimum beyond host quotas. Producer
normalization and four data-count encoder corrections retain before/after
hashes. Generic Core 3.0 data/table declaration cases requiring unimplemented
GC/typed-reference constant contexts remain in that later capability scope;
the complete pinned Core 2.0 corpus remains independently required.

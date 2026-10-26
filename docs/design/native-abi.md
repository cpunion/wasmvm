# Native execution ABI and compiler gates

Status: **boundary implementation in progress**, not a complete native backend or performance acceptance claim. This document defines the initial boundary for the project's own compiler. Required M2 targets are **linux/amd64 and darwin/arm64, each with both Go and LLGo hosts**. Other OS/architecture combinations remain unsupported until their execution, lifecycle, and ABI suites pass.

## Compiler and host boundary

The decoder, validator, typed IR, optimizer, register allocator, machine-code encoders, relocation logic, and interpreter are written in Go. They must build and test with `CGO_ENABLED=0`; the compiler can emit and inspect code without enabling execution. There is no LLVM, Cranelift, wazero, or other engine dependency in the implementation path.

Initial native execution uses a small C-compatible platform adapter. The standard Go build calls it through cgo; the LLGo build calls the same native C ABI using an explicit C declaration. The adapter owns executable-memory operations and the indirect entry call. It does not decode Wasm, execute an interpreter, optimize IR, or implement guest semantics. This distinction preserves a pure-Go interpreter and compiler implementation while making the initial native execution dependency explicit.

| Build configuration | Required behavior |
| --- | --- |
| Go, `CGO_ENABLED=0` | Interpreter and all compiler logic work; requesting native execution returns a specific unavailable-backend error |
| Go, `CGO_ENABLED=1`, accepted target | Own emitted code executes through the platform adapter |
| LLGo, accepted target | Same interpreter/compiler source and native protocol execute through its C ABI adapter |
| Unaccepted target | No inferred native support from a successful cross-build |

An explicitly requested native backend must not silently run the interpreter. A future automatic selection policy must report the selected backend separately from the requested policy.

The proposed entry has one code address and one pointer to a fixed-layout execution record:

```c
uint32_t wasmvm_enter(const void *entry, struct wasmvm_exec_ctx *ctx);
```

Only the platform adapter converts an approved executable address into a callable native function pointer. Generated entry code accepts `ctx` according to that platform's C ABI and returns an exit status. It never receives a Go closure, interface, slice header, or Go-private ABI frame. Define every record field with fixed-width integer or native-pointer storage; derive Go offsets from the fixed record and verify size/alignment/offsets against C `static_assert` and execution probes on both target/compiler pairs.

Go does not directly call C function-pointer values; its documented bridge mechanism is a C wrapper. Its internal function ABI is explicitly unstable and differs from the platform C ABI. These are reasons to isolate the entry adapter, rather than reinterpret a Go `func` value as executable memory. Sources: [cgo function-pointer bridge](https://pkg.go.dev/cmd/cgo#hdr-Go_references_to_C), [Go internal ABI](https://go.dev/src/cmd/compile/abi-internal).

LLGo already has explicit native C calling-convention support and C-bound runtime declarations. Its Go closure representation also has compiler-owned environment variants; none belongs in this execution ABI. Local source inspection used LLGo `38bfd05ee102e36710a2df5bca47834a622b667c`, including `ssa/stdcall.go`, `runtime/internal/ffi/call_llgo_nest.go`, and `runtime/internal/clite/bdwgc/bdwgc.go`. These establish available mechanisms, not proof that the new adapter is correct. See the pinned [LLGo C-call lowering](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/ssa/stdcall.go).

## Guest frames, registers, and resumption

The first implementation does **not switch the hardware stack pointer to a Go-allocated guest stack**. The adapter and native entry use a fixed-size C stack frame. Guest parameters, locals, spills, saved continuation labels, and call frames live in a separate, explicitly bounded off-heap VM frame arena.

Guest calls push logical frames and branch to compiled callee entries. Guest returns restore the caller's frame and branch to its saved continuation. Guest recursion therefore consumes the VM arena instead of unbounded Go or C stack space. Tail calls replace the current logical frame. Internal calls use a documented project ABI; only entry/exit crosses the platform C ABI.

The context contains frame-arena base/limit/current offsets, module and function identifiers, continuation identifier, value/result slots, memory descriptors, exit payload, current guest instruction location, root-map identifier, and remaining execution budget. Persistent frame links use offsets, not pointers into a Go slice. Growing the arena occurs only after a host exit; resumption reloads its base. Bounds and total stack-byte/depth quotas are checked before committing a frame push.

Each ISA defines a register-role table for context, current guest frame, memory base, scratch registers, allocatable integers, and allocatable vectors. Callee-saved registers modified by generated code are saved once at native entry and restored on every exit. The initial arm64 allocator excludes x18 and preserves the platform frame/link conventions. SIMD values spanning a yield are spilled in full, independent of the platform's narrower preservation rules. The amd64 adapter follows the System V contract; Windows is a separate future ABI. Sources: [AAPCS64](https://github.com/ARM-software/abi-aa/blob/main/aapcs64/aapcs64.rst), [x86-64 psABI](https://gitlab.com/x86-psABIs/x86-64-ABI).

Hardware-call fast paths, hardware guest-stack switching, or direct generated-code-to-Go callbacks are later design changes. They require new root, unwind, stack-limit, cancellation, and dual-compiler evidence; they are not implied by an instruction encoder passing arithmetic tests.

## Exit protocol and safepoints

Every slow operation uses an explicit exit record. Initial statuses are `Return`, `Trap`, `HostCall`, `GrowMemory`, `GrowStack`, `Poll`, `FuelExhausted`, and `GuestGC`. `FuelExhausted` maps to the typed budget outcome rather than a generic guest trap; a poll that observes cancellation returns the separate canceled outcome through the driver. Additional feature-specific statuses need a defined owner and resume contract.

Before returning a resumable status, native code spills live state into its logical frame, publishes a continuation and exact root-map identifier, and fills the exit payload. It then restores the original C frame/register state and returns normally. The Go driver validates the status and ownership, performs the operation, writes typed results, and calls the C-ABI entry dispatcher again. That dispatcher establishes the normal prologue/register state, reloads the frame/context, validates the retained continuation identifier, and branches to its private-ABI label. An interior continuation label is never cast to the C entry function type. Reload memory bases/lengths, frame-arena bases and mutable dispatch metadata after every exit, including host reentry: an aliased imported memory can grow inside a callback. Retain cached addresses only where an explicit stable-address/pinning contract proves them valid. A continuation belongs to one Store/execution and retained code generation; it cannot be resumed concurrently or after its Execution/Store has closed. Releasing a CompiledModule handle does not invalidate code still retained by an instance or execution. No native return address or continuation is supplied directly by guest memory.

Host functions execute after native code has returned. Reentrant host calls create a distinct execution context while retaining suspended parent frames. Host-programming panics retain the documented Go behavior and are not silently converted into guest traps; neither a Go panic nor `longjmp` unwinds through generated frames. Deferred host-boundary cleanup releases the interrupted execution/reentry token, temporary roots and code leases before propagating a host panic; a caller that recovers must not find the Store permanently busy. Guest exception handling walks explicit guest frames and must preserve its own tag/value semantics.

Polling is mandatory even when the caller has no deadline. Emit polls on loop backedges and call boundaries, and split long straight-line regions so an execution-budget check cannot be optimized away. Respect the performance contract's maximum of 1,024 original guest instruction points between polls and 64 KiB between bulk-operation polls. Backend scheduling budgets may tighten these limits, never exceed them. Start with a conservative configurable maximum region budget within those caps; determine its accepted value by measured worst-case exit latency, rather than asserting that a source-op count is a time guarantee. Variable-work instructions such as bulk memory operations must be chunked with progress continuations when their maximum work exceeds the cancellation contract. Atomic waits always exit to the host wait implementation.

There are two distinct accounting contracts:

- **Scheduling/interrupt policy:** the sliced baseline decrements a driver-refilled instruction budget and returns at exhaustion. The unmetered compiler polls an atomic interrupt word at control transfers and every 256 slot instructions in straight-line code, without mandatory Go round trips. These are distinct policies, recorded in comparisons; neither bypasses cancellation. A native C entry releases the Go scheduler resource, and both host compilers must demonstrate GC/timer/goroutine progress with `GOMAXPROCS=1`. Native worker admission, requested collection/yield, and controlled latency gates remain requirements for public runtime integration.
- **Semantic fuel:** when enabled, charges defined guest operations consistently across interpreter and compiler. Optimizations cannot remove its charges or reorder a fuel trap past visible effects.

The host checks cancellation, close, scheduler work, and requested guest collection after exits. Accepted native execution must also pass the independent measured cancellation/exit latency ceilings; deterministic work caps alone are not elapsed-time evidence. It yields to the host scheduler when servicing polling exits. Host Go GC uses its normal cgo transition; LLGo retains its normal registered native thread and stack. Neither host collector is required to understand generated guest frames. Long native slices and suspended host calls are still tested while another host goroutine repeatedly forces collection.

Cancellation and lifecycle control words have an explicit cross-language atomic protocol. Use aligned fixed-width, lock-free atomics in the platform adapter and corresponding acquire/release machine instructions in emitted code; assert the C layout and lock-free property on each accepted target. Other Go goroutines request cancellation through that adapter, never by racing a plain Go field against a generated load. Ordinary frame/context payload fields have one owner: native code while running, the driver after an acknowledged exit. Inspection, collection, or mutation waits for that handoff. Closing retains the control record until request writers, executing code, and suspended continuations have all released it. Go's race detector is an additional host-code check, not instrumentation of JIT memory accesses.

The initial adapter creates no foreign threads and calls no Go code from generated execution. LLGo entry originates on a runtime-managed, collector-registered thread; guest workers use bounded host-runtime workers. A future embedding callback from an arbitrary native thread must attach both LLGo runtime state and collector state before entering Go, with matching detach ownership. Registering a stack only with BDWGC is not sufficient runtime attachment, and an unsupported foreign-thread entry must be rejected rather than assumed safe.

## Roots and ownership

Native execution records, guest frame storage, and native-visible linear memories are off-heap allocations owned by the Store/execution. Guest-heap objects may remain in the typed host registry behind handles and be accessed through checked runtime exits; any later directly native-accessible payload must use the same pointer-free ownership rules. They contain raw scalar bits, guest reference handles, and pointers to VM-owned native allocations. They contain **no raw Go heap pointers**. Generated stores must never rely on a missing Go write barrier.

Host references use generation-checked handles. A strongly reachable, typed Go registry owns the corresponding Go values; a numeric handle alone is not a GC root. Active and suspended executions retain their Store, code, registry entries, and memory owners until release. Explicit `KeepAlive` at adapter exits supplements this ownership discipline; it does not make hidden native pointers visible to a collector.

At allocating, yielding, host-call, trapping, and exceptional boundaries, all live guest references must be represented in exact root maps. The initial baseline allocator spills conservatively; the optimizing allocator may reduce spills only with checked liveness maps. Roots include operands, locals, globals, tables, pending host arguments/results, exception payloads, and suspended/reentrant frames. Register-resident references cannot be the only copy when an execution returns to the host. Guest collection uses these maps and guest-object metadata, independently of Go or LLGo host collection. A Store has one active guest executor by contract, with suspended reentrant parents included in the root set. Collection prevents re-entry and host root mutation until tracing/reclamation completes. Future shared-heap execution would require acknowledgements from every worker; it is not enabled by Core Threads, whose Stores have separate heaps.

Avoid retaining Go memory across native calls. If a later zero-copy API needs it, it must follow cgo pinning rules, document its lifetime, and pass `GOEXPERIMENT=cgocheck2` tests. A slice header or interface cannot be hidden in an off-heap context. LLGo's conservative collector is not justification for breaking the standard Go contract. Source: [cgo pointer and handle rules](https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers).

## Code publication and platform profiles

Code moves through `AllocatedWritable -> Relocated -> SealedExecutable -> Published -> Retired -> Freed`. New execution leases can be acquired only for `Published` code. `Retired` code may still execute or resume under an existing retained instance/execution lease, but accepts no new unretained entries. `Freed` code is never executable; transition to it requires all leases and active entries to be released. Validate every relocation, range, alignment, and branch target before publication. Store immutable trap/source/root-map metadata with the code generation. Synchronize publication so readers cannot observe incomplete bytes or metadata.

On Linux, allocate writable non-executable memory, finish emission/relocation, perform the architecture's required instruction-cache maintenance, and change protection to executable non-writable. Do not leave ordinary RWX mappings. On darwin/arm64, use a platform JIT arena and the supported write-protection mechanism; respect the embedding application's Hardened Runtime and entitlements. Apple's callback-allowlist mode cannot be mixed with `pthread_jit_write_protect_np`. Keep an entire write transaction inside one adapter call on one OS thread, invalidate the instruction cache before execution, and publish only after write access has ended. Missing permission returns an explicit native-unavailable error. Source: [Apple JIT porting guide](https://developer.apple.com/documentation/apple-silicon/porting-just-in-time-compilers-to-apple-silicon).

Published code is immutable. Tier replacement publishes a new code generation and updates dispatch metadata; it does not patch bytes currently being executed. Active calls, suspended continuations, function references, and host reentrancy retain the old generation. Closing a module retires code; unmapping waits until those references are quiescent. Do not rely on a finalizer to prevent use-after-unmap. M2 initially uses in-memory code caching only; persisted native artifacts require a separate trust and compatibility design.

CPU profiles are explicit inputs to compilation and part of every cache key, together with target OS/architecture, engine/IR/native-ABI versions, guest feature profile, module digest, and semantic options. Bring up scalar amd64 and arm64 baseline instructions first; M2 acceptance also requires the complete fixed-SIMD profile through native lowering or explicitly measured helpers. Probe optional extensions through the platform adapter, record the selected profile, and never infer support from the build machine. SIMD expansion and optional instructions require both semantic tests and feature-masked execution tests. Preserve caller FP control state and establish the Wasm-required floating-point behavior; ambient host rounding/flush modes must not change guest semantics.

## Traps, optimization, and diagnostics

Emit explicit guest bounds, alignment, divide-by-zero, integer-overflow, conversion, indirect-call type, null-reference, and stack-limit checks where the Wasm semantics require them. Use unsigned widened address arithmetic and overflow checks before forming a host address. Initial correctness does not depend on recovering SIGSEGV, installing process-global signal handlers, or interpreting a host panic as a guest bounds trap.

Every trap site maps to its module/function/instruction offset and preserves the specified prior side effects. Optimizations may not hoist a trapping operation past a condition, host call, memory effect, or observable fuel/trace event without proving the change valid. Exception tables, source locations, stack maps, and continuation maps are compiler outputs verified alongside machine code. Instrumented execution preserves its documented event sequence; a faster uninstrumented mode must be a separate configuration recorded in benchmark results.

The compiler pipeline is typed validation output -> control-flow IR -> verified lowering -> machine IR -> register allocation -> encoding/relocation -> immutable executable. Begin with explicit loads/spills and local allocation. Add SSA optimizations, global allocation, inlining, loop transformations, and check elimination only after each pass has differential and adversarial tests. Do not make LLVM part of the host dependency graph to obtain this tier.

## Early unmetered compiler tier

`CompileUnmeteredFunction` shares decoding, typed slot IR, instruction lowering,
relocations and the C boundary with the sliced baseline. It rejects metered
entry records; deterministic fuel continues through the existing precise tier.
The static policy removes per-instruction fuel-flag tests and periodic host
round trips, while retaining mandatory acquire-load interrupt checks.

A small fixed allocator assigns the most referenced scalar slots to eight
amd64 or ten arm64 registers, excluding context, frame and constrained integer
scratch registers. Assignments remain identical across CFG edges. Entries load
the retained frame; every exit spills the cached slots, including reference
handles. Remaining scalars and all vectors use the frame. Scalar `Hi == 0` is the
existing `value.Value` contract; vector copies retain all 128 bits. The driver
must preserve that contract on initial entry, helper results and resumption.
Common arithmetic names allocated registers directly, while constrained
instructions use the existing checked lowering. There is no workload-specific
pattern or competitor fallback.

This is an early architecture improvement, not a complete native runtime.
Native memory, guest calls, runtime exits, precise metering optimization, global
liveness/allocation and the full-profile performance gates remain unfinished.
Memory operations and guest calls must receive their own same-workload
competitor comparisons before their development PRs can merge.

## M2 acceptance sequence

These gates describe milestone completeness, not permission to postpone performance. Measure architecture and optimize alongside N1–N3. Each feature-development PR must reach competitor parity or better for its affected workloads before merging; full N5 acceptance still requires the entire portfolio.

| Gate | Concrete deliverable | Required evidence |
| --- | --- | --- |
| N0: Boundary | C adapter, record layout, executable lifecycle, scalar return, status/resume | Go + LLGo execution on linux/amd64 and darwin/arm64; callee-saved/FP-state canaries, repeated enter/exit, forced host GC, protection and cleanup tests |
| N1: Baseline lowering | Typed IR, both scalar encoders, branches, scalar numeric/memory/control operations and explicit VM calls | Encoder/disassembler checks, official declared-profile cases, instruction edge cases, interpreter/native differential traces and results |
| N2: Runtime exits | Imports, nested/reentrant host calls, growth, polls, cancellation, quotas, root publication and resumable operations | Close/cancel/GC during active and suspended execution, deep recursion, growth rollback, no stale handles, bounded wait/teardown; both host compilers |
| N3: Profile completion | Every feature assigned to the milestone on both ISA backends | No unsupported instruction replaced by hidden interpreter execution; full declared conformance matrix, cross-feature suites, trapping/exception/reference lifetime evidence |
| N4: Optimizing tier | Measured IR simplification, register allocation, inlining and loop/check optimization | Per-pass differential tests, root-map verifier, unchanged semantic fuel/trace contracts, code-size and compilation-budget measurements |
| N5: Performance acceptance | Phase and steady-call corpus including calls, branches, loads/stores, numeric/SIMD where declared, and realistic guests | Pinned same-fixture competitor comparisons, independent process rounds, separate startup/compile/first/warm measurements, host-boundary control, confidence/noise limits, Go/LLGo parity |

A fast control or arithmetic probe is not a complete compiler. A compiler-wide skip, a successful cross-build, or an interpreter fallback does not satisfy a native gate. Architecture-specific opcode coverage and negative/edge cases are tracked separately from host harness statement coverage.

The C boundary adds transition and marshalling cost. It is paid per native execution slice, not per guest instruction. Long slices amortize it but must still meet responsiveness and GC/lifecycle contracts. Short calls may remain slower than an engine using a private Go assembly bridge. Measure this with the no-work call control; do not subtract its median as an exact correction or claim an unmeasured speed target.

A no-cgo native adapter is a later optimization, enabled only after the same matrix proves scheduler/GC interaction, register preservation, stack scanning, traps, cancellation, and lifecycle safety. It must improve the intended workloads without bypassing the semantics or making native support depend on a released LLGo ABI assumption that has not been tested.

## Existing evidence and limits

The existing local probe used wazero v1.12.0, Go 1.27.3, LLGo compiler `75e966c4fbcf8077497a41093262ad8677711fec`, and LLGo runtime contribution `bc7c991d8fe91459c8afc1eb7d0e374124d95374` on darwin/arm64. Go interpreter and native runs returned 42; LLGo interpreter returned 42; LLGo native compiled/instantiated but its first `CallWithStack` returned a recovered nil-pointer fault. The guest had no imports, memory, start function, GC, or threads. This is recorded in [LLGo issue #2921](https://github.com/xgo-dev/llgo/issues/2921); local raw metadata/logs are under `artifacts/llgo-wazevo-native/`.

Wazevo's [pinned arm64 entry code](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/internal/engine/wazevo/backend/isa/arm64/abi_entry_arm64.s) uses Go assembly argument slots and guest-stack transitions. The observed fault does **not** isolate its root cause and does not prove that a C ABI adapter fails. It establishes that linking the runtime hooks and accepting interpreter tests did not establish native compatibility. The new boundary must earn its own N0 evidence before any native support or speed claim.

## Implemented boundary

`internal/compiler/nativeabi` emits fixed C entry/exit frames for both target
ISAs in pure Go. `internal/platform/native` publishes immutable code, retains
active and suspended leases, owns pointer-free native records/frame storage,
and enters through the same C ABI under Go and LLGo. Cancellation uses a
lock-free release/acquire control word; the adapter normalizes and restores
the host floating-point environment. Every field offset is checked against C.

Linux mappings transition RW to RX. Darwin uses a shared, bounded 64-MiB
MAP_JIT arena with thread write protection and instruction-cache invalidation;
released ranges coalesce for reuse. The process reserves the arena until exit.
Callback-allowlist JIT mode is outside this initial platform adapter; hardened
embedders must supply the applicable JIT entitlements.

Native ABI probes execute arithmetic, frame stores, status/resumption, register
clobbers and cancellation under both host compilers. They also exercise GC
during native work, concurrent publication, retirement during an active entry,
and strict cgo pointer checks. The function compiler now emits control flow, integer operations, typed frame
maps and checked continuation dispatch. Calls, runtime helpers and large tuple
copies publish explicit exits; connecting them to the runtime driver remains
subsequent work. Public native
backend selection remains unsupported until the complete Core 2.0 backend is
implemented and validated.

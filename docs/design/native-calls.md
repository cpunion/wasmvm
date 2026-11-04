# Native guest calls

Status: shared frame transfer and invocation budgets are implemented at engine source `4862cc8`; native call/return branch emission is unfinished. No new performance result is claimed. The required final scope remains the complete backend and full performance acceptance.

## Shared runtime boundary

`state.NativeFrames` is a borrowed Go-side packet. `execute.FrameBackend` optionally executes it; existing backends keep their old boundary. A nonempty packet replaces the existing top frame with the complete root/child chain. `execute.adoptNativeFrames` validates all parents, PCs, direct targets, bases, tuple bounds and current invocation quotas before mutating the shared stack. The already charged final boundary is serviced once, after locals are refreshed; restarting the backend at that PC would double-charge fuel. Root destinations survive adoption. All references are spilled before any host, GC, exception or continuation operation. Packet references are cleared on success/error and Abort. A returned error preserves the current owned arena view for reuse even after relocation.

The real native backend uses the packet now, without emitting hidden calls yet. It passes current invocation quotas into the existing execution/expansion path rather than relying on the cached workspace maximum. The eight complete affected packages pass Go/LLGo, with 96 actual conformance reports and all scoped coverage above 90%. Model tests exercise hidden frame transfers; they are not proof of generated native calls.

## One hardware entry, bounded guest frames

Generate ordinary same-instance direct calls and returns as machine branches, not recursive C calls. Existing per-function C ABI wrappers have the same fixed prologue/epilogue for an ISA. Branch into the callee at `Entry.BodyOffset`, bypassing its prologue. Any callee's shared epilogue can return through the one original hardware frame. Preserve caller callee-saved registers and the adapter FP/W^X contract. The C adapter remains only an entry/platform boundary and implements no guest call semantics.

Use `Payload[1]` as an optional trusted native call-view address; production code currently does not use it. Keep existing raw probes and codegen configurations unchanged when linking is disabled. Document this field designation; do not confuse Go-side packet pointers with off-heap ABI data. The call view holds top/capacity/current frame and slot limits plus bounded parent records. Parent records contain caller function index, value-base byte offset, call PC, compiler-owned resume address and immutable destination-plan address. No Go pointer enters generated execution.

Allocate the parent metadata proportionally (for example initial 16 records, growing at idle boundaries up to the invocation's declared frame maximum). Retain bounded growth/fallback and account metadata bytes. Exhaustion leaves an ordinary Call boundary before committing an extra frame; materialize existing hidden parents before the common driver grows the arena or reports a limit. Current remaining limits can shrink after suspended parents redistribute budgets, independently of reserved capacity.

## Call emission

Before changing live state, check metadata capacity, current depth, logical slot quota and the actual native arena capacity against the callee frame. Spill caller registers once and preserve complete 128-bit values. Clear callee storage and copy arguments into its slots. Use bounded loops for large tuples/frames with cancellation checks; a canceled partial initialization may fall back to the already charged Call before publishing the child. Do not spill clobbered cached registers again on that fallback: use a raw typed exit because the caller was already spilled.

Publish the parent record, new used/depth/function fields and continuation 0 only after preparation succeeds, then branch to the callee body header. Preserve the current source offset. Every original cold continuation remains valid. Metered calls use the metered callee; unmetered calls use its ordinary tier. Tail/imported/indirect/cross-instance calls retain existing checked typed exits until separately optimized; their complete semantics remain supported.

Use cross-function PC-relative branch relocations recorded by compiler labels, whose final offsets are read after assembler branch compaction. Link all targets after module layout. Validate each target as the selected policy tier's known function body; enforce ISA alignment/range and total code budget. Avoid exposing arbitrary executable addresses or pinning an extra code lease for each call.

At the call site, place a nearby compiler-owned resume stub and immutable result-destination index table. Save their PC-relative addresses in the parent record. The stub branches back to the caller's body header with the restored original continuation, reloading all register/constant/runtime state. Existing assembler ADR uses x13 and can address nearby metadata; x86 LEA uses a relative displacement. Large function headers use ordinary long branches, not a potentially out-of-range ADR.

## Return emission

If there are no hidden parents, leave through the existing Return boundary. Otherwise spill the callee once, read its selected return source tuple and the saved parent destination plan, and copy complete values into the disjoint parent frame. For large tuples, use a bounded loop and cancellation checks with immutable source/destination index tables rather than unbounded code expansion. Restore used/depth/function and the caller continuation, decrement metadata top, then branch to the saved resume stub. Do not overwrite caller state or discard live roots before all required copies are safe. Avoid in-place tuple permutations through the callee's argument slots.

## Stable runtime views and observable exits

Linked functions need one stable runtime descriptor layout. Current per-function global binding counts change memory/heap offsets. Compile linked tiers against a sorted module-wide union of required globals and bind that union once per entry; retain imported-global alias identity. Every callee uses the same function/table/global/memory/heap offsets. Bind the first memory if any reachable linked body can use it. Refresh views at each true runtime reentry after growth or host effects; no Go/world callback occurs while hidden frames exist.

Audit native allocation batches: a callee must not assume that its caller reserved a compatible batch. Preserve checked reservation limits and the existing slow allocation boundary when no usable batch exists; never disable WasmGC capabilities or remove their safety checks. Shared-memory atomics/growth and table mutation must retain their current synchronization and ownership contracts.

`nativeexec.run` currently assumes the original function/entry on Poll, Trap and helper validation. When Context.Function changes, resolve the active function against the same instance and module; use the matching policy entry on Poll resume. Trap attribution and pending-helper opcode checks use the active function. After any observable boundary, convert all remaining native parent records plus the current frame into `NativeFrames.Frames`, expose the complete values slice, then clear native top for the next shared-machine invocation. Fuel/cancellation failures must preserve exact consumed fuel, cleanup and instance reuse rules.

Before the common machine services any helper/host/GC/exception/stack-switching operation, adopt the full chain. This keeps the existing handler search, root tracing, stack traces, pending host destinations and continuation capture unchanged. Crossing a runtime boundary can deliberately end a linked region; a later return then uses the ordinary shared frame. No second interpreter or runtime driver is introduced.

## Required qualification

Implement both amd64 and arm64 emission. Exercise actual generated direct recursion/mutual calls and nested helpers, multi-value/reference/SIMD tuples, frame/slot/metadata growth and overflow, memory/global updates, GC, host reentry, exceptions, Stack Switching, traps, exact fuel, cancellation and reuse/close under both Go and LLGo. Keep default/raw codegen tests valid. Run the full affected package suites and exact coverage, then complete same-cohort source comparisons over the original portfolio plus compiled C/Rust modules and pinned competitors/native C. Preserve all adverse rows and compilation/startup/first/warm/cleanup phases. Do not multiply improvements across independent cohorts or rerun to select favorable timing.

The local frame-contract branch has no PR while the single external contribution remains open. #2983 belongs to another contributor. Do not duplicate it or change the production LLGo profile as part of this work.

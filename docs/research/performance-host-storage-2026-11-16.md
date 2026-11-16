# Synchronous host storage and native operand follow-ups

The owner requested completing the ordered implementation work before the
unified competitor event. This record describes structural changes and semantic
evidence. It contains no new timing claim and does not close performance or
production-release acceptance. The earlier [mechanism investigation](performance-p3-mechanisms-2026-11-16.md)
and [canonical invocation study](performance-canonical-invocations-2026-11-16.md)
retain their separate observations and adverse results.

## Implemented mechanisms

PR [#188](https://github.com/cpunion/wasmvm/pull/188) adds explicit Core borrowed
scalar carriers, including immediate suspendable results. Ordinary callbacks
retain their owned stack contract. Pending callbacks keep owned values and
fresh callback capabilities. Completion polling combines subscription and
consumption under one lock; atomic readiness is only a scheduling hint. Source
reads share immutable type/value plans while retaining independent transactions.
Proven amd64 vector accesses reuse the checked scalar address encoder.

PR [#189](https://github.com/cpunion/wasmvm/pull/189) groups fresh affine reference
allocations and endpoint pairs. Published guards are never recycled or rebound.
It also adds opt-in synchronous borrowed component parameters. P3 HTTP types
consume normalized canonical values without another aggregate snapshot. Hosts
copy aggregate data they retain into provider objects; immutable strings and
transferred owned resources retain their ordinary contracts. Public calls,
async parameters and returned results remain owned snapshots. Borrowed retention
uses the same logical validation, value and transfer charges as cloning.

The next implementation reuses a cleared synchronous transaction shell attached
to each private canonical frame. A shell never backs a pending task and no public
Caller points into it. Reentry takes another frame. Failure cleanup completes
before the function, codec, resource, context and budget state clears. A cached
frame retains one lazy shell allocation; it does not retain old invocation data.

String/list/map pointer-length descriptors now use one checked 8/16-byte copy,
sharing the existing string descriptor implementation. The memory capability
remains held for that copy. It exposes no backing slice and preserves byte-wise
shared synchronization rather than promising atomic descriptors. Invalid short
ranges cannot partially overwrite a descriptor.

Proven nonwrapping i32 scales also feed amd64 SIMD loads through the existing SIB
encoder. Bounds proofs, narrow lane reads, SSE/AVX feature policy and original
cold entries stay in place. Stores, unproven scales, memory64 and precise fuel
retain their existing paths.

ARM constant-shift consumers can execute before an observable intermediate is
restored, provided the consumer does not overwrite the original source. This
keeps the original number of shifts and preserves every frame word while
removing that shift from the consumer's result dependency. The existing immediate
encoder emits the restored intermediate. A compile-only diagnostic selects one
additional `i32.shr_u`/`i32.xor` group in each pinned Clang/Rust SHA module.
This selection count does not establish a time or throughput improvement.

A further native entry increment uses Store workspace exclusivity through
`NewOwned`/`BeginOwned` and `RunOwnedState`. It removes the two Context mutex
acquisitions from each owned native entry, sharing the same state checks and
code/memory pins with the general entry. Only cancellation may overlap; the
owner drains delivery before rebinding or closing. Other Context clients keep
their concurrent-close path. Native code still enters through the supported C
transition. A test holds two owned entries active through Go collection, then
retires their code lease and cancels/drains both entries before storage release.
This addresses repeated host-side synchronization without adopting the rejected
assembly bridge below; full state/frame materialization still exists.

The following SDK pin advances LLGo from `5860fb6f` to main `1b44c645` (five
commits: horizontal SIMD subtraction and PCLN/LTO flag handling). The existing
`-O3 -lto=full` profile stays fixed. Old-SDK phase and complete-P3 controls were
built at committed engine source `4d6fc1a4` before changing the SDK checkout.
The next unified event builds both source base/head with the new SDK and treats
the fixed-engine old/new compiler comparison as a separate baseline event.
Neither SDK adoption nor source optimization is a performance result by itself.

PR #190's longest job was macOS LLGo VM at 606 seconds. Its trace records a
472-second build, 1,116 seconds of overlapping link work and about 30 seconds of
test execution; the codegen test link takes 209 seconds. Moving that package
to the shorter component shard retains the same 32 root packages, one auxiliary
module, nine shards and full LTO profile. Its effect on the maximum job duration
must be assessed from the next CI run, not inferred by adding parallel spans.

## Validation and rejected entry bridge

The tests exercise retained public arguments/results, copied expired Caller
headers, nested reentry, forced collection, resource moves and borrow expiration,
pending completion/cancellation, panic cleanup, logical budgets, memory32/64,
full descriptor bounds and all original native PCs. Vector tests compare every
frame word and memory byte against checked execution. Linux amd64 execution uses
local Docker; macOS arm64 uses the local host. Final PR CI retains both compilers
and platforms, including the pinned LLGo `-O3 -lto=full` profile and per-package
coverage gates. Local Docker execution is semantic evidence, not timing evidence.

The owned-context follow-up's local Docker full native package run aborts in
Rosetta when the existing stable-memory test requests an unreservable extent
near 2^63 bytes. Replaying the unchanged test against the pre-change native
sources reproduces the same abort. Targeted entry/GC/lease tests and complete
nativeexec/VM packages pass there; local Go/LLGo arm64 runs pass. This Docker
failure is retained as an environment limitation, not counted as a full native
package pass. Native amd64 CI remains required without skipping that test.

The earlier direct Go assembly entry experiment remains rejected. One native
entry could run while another goroutine requested collection; Go could not
asynchronously preempt its unregistered native PC. The concurrent lease/GC test
deadlocked while the cgo baseline passed. Interrupt checks alone do not guarantee
a return to the Go scheduler in the unmetered tier. The patch was withdrawn;
rejected patches, logs and the failed experiment remain under
`artifacts/native-entry-bridge-20261116/`.

A safe broader bridge still needs runtime cooperation or a complete bound on
every reachable path, including linked calls, returns and helper loops. The
transaction and operand changes above preserve the supported cgo/native entry
and all GC/root/lease contracts. They do not complete that larger ABI change.

Final comparison must cover complete P3 requests and Core/native portfolios,
both host compilers, independent competitors and native oracles, retaining
startup/preparation separately from steady execution. Reduced allocation counts,
static selections or passing semantic tests cannot establish competitive parity.

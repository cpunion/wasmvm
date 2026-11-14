# Checked internal native call entries

Ordinary host entries keep their complete frame, continuation and runtime-view
checks. Compiler-owned native calls use an interior entry only after checking
metadata capacity, logical depth/slots and physical arena capacity, initializing
the child frame and copying complete parameters. The caller restores the frame,
root map and compiler-known PC. The shared same-instance runtime view cannot
move during native execution. Saved returns use immutable compiler metadata.

Memory preparation, metered boundaries, cancellation and root publication remain
shared; foreign/host reentry still uses the ordinary checked entry. No second
executor, public ABI change or narrower value-copy contract is introduced.
Nonlinked bodies add no emitted instructions. The
[ABI design](../design/native-abi.md#checked-internal-call-entries) records the
preconditions. Tests execute private linked targets, ordinary host/poll resumes,
misaligned/insufficient frames, invalid PCs and missing runtime views, alongside
the existing recursion, mixed tuples, quotas, GC roots, reentry and cleanup suites.

## Exact-source comparison

Selected source c3cd0c4b versus base433d995d, M4 Max/macOS arm64, Go1.27.3 and
LLGoac129220 with `-O3 -lto=full`, LLVM22.1.8 and normal GC. Both sources use
symmetric four-file Git overlays in one existing worktree; no full snapshots.
The pure Go reports assert the actual source/compiler revisions. Pinned independent
references remain wazero1.12.0, Wasmtime48.0.1 and Wasmer7.3.0. Their original
adapters/policies and the unchecked native C arithmetic control remain explicit.

The general-call kernel uses five million source iterations, each calling a
looped callee for two affine steps. Twenty fixed blocks contain240 fresh
processes: Go base/head, three references and C, each with an identical-binary
alias. Seven warm calls follow first call and two warmups; every checksum is
checked. The paired Go head/base ratio is **0.96572 [0.95772,0.97387]**. All
kernel controls resolve2.5%. Head remains3.295x wazero,3.501x Wasmtime,3.472x
Wasmer and3.465x unchecked C. This is a source improvement, not parity.

The complete P3 Rust HTTP guest has separate16B/64KiB events, each with20 fixed
blocks,200 fresh processes and20 individually interleaved512-request batches.
Go and LLGo base/head plus frozen Wasmtime each have a binary alias. Four Stores,
sixteen clients, snapshot input, buffered output and full body/status/header/
trailer/CRC checks remain unchanged. All400 selected-candidate processes and
4,096,000 measured responses pass. First call, two warmups, final flush,
preparation and cleanup remain separate. No owned build/test/profile/edit/prune
runs during a timed event. Intervals are individual paired-bootstrap diagnostics.

| Go head/base |16B|64KiB|
|---|---:|---:|
| CPU ratio,95% interval |1.00180 [0.99613,1.00814]|0.99460 [0.98973,1.00007]|
| Instructions |0.98708 [0.98574,0.98841]|0.98619 [0.98492,0.98733]|
| Throughput |0.99317 [0.97831,1.00650]|0.99507 [0.97465,1.01490]|

Request instruction work falls1.29%/1.38%; CPU and throughput improvement are
not established. The16B baseline CPU control fails2.5%, while both64KiB own CPU
controls pass. Managed allocations remain approximately696/756 per request.
Descriptive Go/Wasmtime CPU points are2.057x/2.304x, with unresolved reference
CPU controls; these do not close competitive acceptance. Managed counters omit
Rust/C allocations and cannot rank total cross-language allocation work.

LLGo CPU head/base is0.99559 [0.97850,1.01328] and0.99788 [0.98704,1.00845],
with unresolved source CPU controls. It remains2.455x/2.205x Go in these events.
No LLGo source CPU gain or established regression is claimed. Go's16B compile
point is+4.36% with an interval crossing1; all compile controls remain unresolved.
Two LLGo host build walls of78.37/87.42 seconds are one unreplicated pair, kept
separate from Wasm compilation. Every adverse cold/warm point remains in data.

## Cost localization and rejected work

A separate160-process counterfactual replaces general guest calls with an
explicitly inlined equivalent loop, reducing the time point72.3% toward C.
Clearing only parameters/locals instead of all temporary slots reduces the point
2.4%. Some controls fail; these are localization evidence, never production
candidates or accepted optimizations. The closed native kernel allocates neither
Go objects nor a Go invocation per guest iteration: allocator/GC tuning alone
cannot explain its remaining native gap.

Small return-copy unrolling was rejected. Its240-process event worsens time2.71%
with an unresolved candidate control. A separate40-process equal-work audit
reproduces CPU+2.86% and cycles+2.92%, despite instructions−2.42% and CPI+5.47%;
both own controls resolve2.5%. This is reduced work with worse execution cost,
not evidence of an LLGo defect. Exact dependency/layout/branch causes remain open.

The initial entry candidate fe1dc69b improves the kernel5.43%, but P3 CPU points
of+0.61%/+0.31% cross1. Its independent source event is retained. A direct
40-process v1/v2 audit resolves the apparent candidate discrepancy: v2/v1 CPU
is0.99816 [0.99367,1.00201], cycles0.99884 [0.99503,1.00243], instructions0.95934
and CPI1.04117; all own controls resolve2.5%. More removed instructions establish
neither an additional CPU win nor a v2 CPU regression. The selected design reuses
the already-validated view and avoids a dead continuation reload; full request
and portfolio acceptance remain open. No candidate events are pooled or extended.

The next substantial target is parameter/result frame-memory traffic and native
context/metadata updates. Register arguments/results and cached native state need
a scoped design preserving full128-bit values, spill/root boundaries, quotas,
precise fuel, cancellation and host/foreign resumption. Instruction count alone
is an insufficient optimization gate.

## Validation and CI

All32 Go test packages pass with all31 executable packages above90% statement
coverage. Pinned LLGo executes codegen/nativeexec semantics; Go lifecycle race
checks and vet pass. Final Go/LLGo × Linux amd64/macOS arm64 CI is required.

The preceding CI change reduced Linux threads from684 to617 seconds, still17
seconds above the10-minute goal. Its root service tests cost119 seconds and P3
401 seconds. Move root service tests beside legacy/native-reference validation,
and move the small compat package from component to threads. All32 root packages
remain covered exactly once across the same8 nonempty shards, with both extra
modules and production flags retained. Actual final-head durations remain required.

[All identities, controls, cold phases and adverse points](data/performance-internal-call-entries-2026-11-14.json),
[all2,048 projected process records](evidence/performance-internal-call-entries-2026-11-14.jsonl.gz)
and [protocols, source overlays, fixtures and validation logs](data/performance-internal-call-entries-2026-11-14-protocol.tar.gz)
retain the independent events. Only unused per-request latency arrays are
projected to counts/extrema/exact hashes; full originals stay local. This change
does not close competitive, full-portfolio or release acceptance.

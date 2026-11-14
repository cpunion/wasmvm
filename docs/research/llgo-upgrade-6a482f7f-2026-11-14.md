# LLGo 6a482f7f baseline and collection-work regression

Upgrade the active LLGo pin from ac129220 to 6a482f7f, keeping Go 1.27.3,
LLVM 22.1.8, `-O3 -lto=full` and default production GC unchanged. The single reused
SDK worktree is clean at the new revision; old compiler/runtime executables were
frozen before advancing it. No full source copies or upstream mutations occur.

[PR #3102](https://github.com/xgo-dev/llgo/pull/3102) fixes
[issue #3101](https://github.com/xgo-dev/llgo/issues/3101): methods on private types
could be emitted without complete placement analysis and then reused from package
archives with temporary arrays on the heap. The change analyzes all emitted
methods and invalidates old placement/archive facts.
[PR #3103](https://github.com/xgo-dev/llgo/pull/3103) corrects CallersFrames PCs;
[PR #3096](https://github.com/xgo-dev/llgo/pull/3096) enforces native test host
deadlines. Neither is assumed to improve ordinary steady execution.

## Correctness verification

The same preallocated-capacity probe checks 100,000 AppendUint32 calls with the
production profile. Old/new managed allocation events are 100,000 → 0; the direct
capacity/reslice/PutUint32 and empty-observer controls remain 0. The upstream
isolated-cache two-client regression passes: first populate encoding/binary's
archive without AppendUint32, then reuse it for the checked append loop. New
LLGo also executes wasmvm codegen/nativeexec/component suites. All active CI,
benchmark, packaging and documented pins move together. Final dual-platform,
dual-compiler CI remains required before merging.

## Independent compiler events

The runtime source is fixed at wasmvm 4d64f951, whose tree equals integrated
393bf75c and whose production runtime equals c3cd0c4b. Old and new LLGo use the
same fixture bytes, arguments and release flags. Go 1.27.3 is an unchanged
anchor; the P3 event also retains frozen independent Wasmtime 48.0.1. Historical
source optimizations are not pooled into this compiler event.

Each complete P3 payload event uses 20 fixed blocks/160 fresh processes: old/new
LLGo, Go and Wasmtime, each with an identical-binary alias. Twenty 512-request
warm batches are individually interleaved. First call, two warmups, final flush,
all setup/compilation phases and cleanup remain separate. All measured responses
pass body/status/header/trailer/CRC checks. No owned build/test/profile/source
edit/pruning overlaps timed work. Intervals are individual diagnostics.

| New/old LLGo | 16 B | 64 KiB |
|---|---:|---:|
| CPU ratio, 95% interval |1.07782 [1.05621,1.09812]|1.11784 [1.08409,1.15298]|
| Instructions |1.07289 [1.05417,1.09041]|1.10016 [1.07020,1.13254]|
| Throughput |0.95544 [0.93840,0.97043]|0.93684 [0.91551,0.95773]|
| Wasm compile-wall |0.89774 [0.84773,0.94569]|0.93075 [0.89576,0.96777]|

Several source CPU/instruction controls fail 2.5%, and every compile control
remains unresolved. Preserve the adverse steady points and favorable cold
points; neither an accepted exact regression magnitude nor a cold speedup claim
follows solely from those intervals. This upgrade fixes the demonstrated cache
allocation defect; it does not claim universal performance improvement.

Five independent 20-block/120-process core events retain 7 warm calls after 2
warmups and all checksums. New/old time ratios are:

| Case | Ratio, 95% interval |
|---|---:|
| Integer interpreter |1.00234 [0.99516,1.01000]|
| Integer native |0.99586 [0.98886,1.00170]|
| Memory interpreter |1.00818 [0.99724,1.01935]|
| Narrow unaligned native |0.99129 [0.98087,1.00160]|
| General native calls |1.00420 [0.99628,1.01309]|

All intervals cross 1. Memory-interpreter head and narrow-native baseline
controls remain unresolved. These finite cases do not accept the full portfolio
or resolve the earlier independent Linux narrow-layout question.

## Collection and initialization-state audit

Managed allocations/request remain approximately 1165/1430 with equal allocated
bytes, while collections per 10,240 requests change as follows:

| Payload | Old median/range | New median/range |
|---|---:|---:|
|16 B|24 / 19–35|39 / 27–40|
|64 KiB|30.5 / 28–41|60 / 33–111|

Initial live and physical heap footprints also differ. The 16 B live-heap starting
median falls 32.36 MB → 20.44 MB, and 64 KiB falls 74.54 MB → 55.59 MB. This is consistent
with less transient compilation allocation changing later collection scheduling;
heap sizing versus retained conservative roots is not yet isolated.

Separate bounded no-GC diagnostics use 10 fixed blocks/40 processes per payload,
five 512-request measured batches and a final flush. `GC_DONT_GC=1` applies only to
these diagnostic children; all measured collections must be zero. New/old
instruction ratios are 1.00035 [0.99926,1.00137] and 0.99933 [0.99832,1.00057].
CPU ratios are 1.00662 [0.99242,1.02439] and 1.00499 [0.99187,1.01716]. The 16 B
own CPU controls pass 2.5%; 64 KiB old CPU and throughput controls remain unresolved.
The near-equal instruction work implicates collection in the normal-event gap,
but does not prove every cycle is GC work or close its exact root cause.

Do not compensate by disabling production GC, adding artificial allocations or
silently changing process-wide policy. Keep the lower-allocation compiler fix,
retain the observed tradeoff and investigate adaptive collector sizing/live-root
behavior separately. No new code-generation defect is inferred from throughput
alone. Go-first wasmvm optimization and all performance/release gates stay open.

[Identities, raw phases, controls and adverse points](data/llgo-upgrade-6a482f7f-2026-11-14.json),
[all 1,000 projected process records](evidence/llgo-upgrade-6a482f7f-2026-11-14.jsonl.gz)
and [protocols and verification logs](data/llgo-upgrade-6a482f7f-2026-11-14-protocol.tar.gz)
retain all events. Only unused per-request latency vectors are projected to
counts/extrema/exact hashes; full originals remain local. No cohort is extended
or combined with an earlier one.

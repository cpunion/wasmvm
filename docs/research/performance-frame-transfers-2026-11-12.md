# Direct physical register/frame transfers

Native compilation uses **3.5% less CPU** in the paired diagnostic comparison
with source `df70fc9` (95% bootstrap interval: 2.2–4.8% less). Compilation wall
and warm HTTP throughput do not resolve an improvement. **Go competitive
performance acceptance remains open.**

## Implementation

Allocated integer registers and F64 registers now load/store their canonical
frame words directly. This removes the intermediate integer scratch-register
move from spills and register-plan transitions on arm64 and amd64. F32 reloads
use an exact-width load; F32 spills retain the existing zero-extension path.

Every displaced value is still stored. The change preserves raw low words,
untouched high words, cold continuations, NaN payloads, cancellation points,
fuel accounting, frame quotas and the native ABI. It does not discard values
based on semantic liveness. Large offsets use the existing bounded scratch
address calculation; vector transfers retain their existing lowering.

## Compilation comparison

M4 Max, macOS 27.0.1 arm64, Go 1.27.3, Clang 22.1.8, GOMAXPROCS=4 and
Wasmtime 48.0.1 Cranelift. Twenty randomized fresh-process blocks include
independent A/A aliases. All 120 processes execute two complete checked P3
HTTP responses. Own builds, tests and profilers do not overlap measurement.

| Configuration | Compile wall ms | A/A wall ms | CPU ms | A/A CPU ms | Instructions billions |
|---|---:|---:|---:|---:|---:|
| Go base |72.45|72.64|133.43|133.61|1.725|
| Go candidate |71.19|70.90|127.94|128.26|1.655|
| Wasmtime |57.67|57.14|unavailable|unavailable|unavailable|

Public Compile includes decoding, validation, lowering, emission and
publication. Protocol markers bound CPU/instruction counters to that phase.
Process startup, input loading, instantiation, execution and cleanup are
separate. Wasmtime exposes its Rust compile wall timer; its isolated compiler
CPU is not inferred from other measurements.

The paired candidate/base CPU ratio is 0.965 [0.952, 0.978]. CPU A/A intervals
fit within ±2.5%. The compile-wall ratio is 0.988 [0.969, 1.007]; wall A/A
controls remain unresolved. The candidate/reference compile-wall ratio is
1.229 [1.183, 1.268]. These are unadjusted diagnostic intervals, not a closed
all-platform or portfolio gate.

## Warm execution remains open

A separate fixed four-block, 24-process comparison runs 1,966,080 checked warm
requests, plus preparation/first/warmup/flush batches. Four independent Stores,
16 clients, 256 requests per client and 16-byte bodies use the same complete
HTTP response, body, trailer and CRC checks for both runtimes.

Candidate instructions fall approximately 1.5%. Candidate CPU is 90.38/89.85
microseconds per request versus base 90.82/90.86 and Wasmtime 43.14/42.76.
Candidate throughput is 40,520/41,463 requests/s versus base 40,939/40,873 and
Wasmtime 51,922/52,957. Timing controls and external Chrome/StorageManagement
load prevent a throughput-win claim. All aliases and adverse records remain.
These observations do not replace the preceding
[quiet full HTTP deficit](performance-native-runtime-image-2026-11-12.md).

## Validation

The complete affected Go codegen/nativeexec/execute/VM/component/P2/P3 suites
pass. Actual-source standalone codegen statement coverage is 95.2%; this is
package coverage, not specification coverage. Go codegen/nativeexec/VM race
checks pass. Pinned LLGo `3d92a58 -O3 -lto=full` executes the complete macOS
codegen/nativeexec/VM suites.

New raw-frame regressions emit both ISAs and execute the host ISA, including
extended integer/FP registers, F32 zero-extension, preserved F32/F64 NaN payloads,
untouched high words, source immutability and offsets around 16/32/64-KiB
boundaries. Linux amd64 Docker executes the complete codegen/nativeexec suites
with Go and pinned LLGo, including these native boundary regressions. This is
scoped semantic evidence; emulation is never used for performance claims. The
final-head hosted matrix must still qualify all packages on both platforms.

[Identities, counters and controls](data/performance-frame-transfers-2026-11-12.json)
and [bounded process/batch records](evidence/performance-frame-transfers-2026-11-12.jsonl.gz)
retain the diagnostic evidence. Large request-latency arrays retain count,
extrema and hash summaries; originals remain local.

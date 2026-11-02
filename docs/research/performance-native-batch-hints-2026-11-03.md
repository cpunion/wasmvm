# Native allocation batch hints: 2026-11-03

Status: **validated local candidate; performance acceptance remains open**.
Exact source `8462dfa` follows the combined dispatcher/allocation source
`fda5e8f`. Baseline binaries are reused only after source and binary hashes are
verified; they retain their original revision. No rejected all-I32 prefix source
is present. This experiment isolates reservation sizing from earlier changes.

## Change and semantics

Reuse the existing bounded register-frequency vector as a profitability hint.
A native constructor inside a backward-edge interval, or more than 64 static
native constructors, requests the existing 256-handle batch. Small acyclic
sequences request 64, matching the earlier reservation size. Unknown flow keeps
64; metered/nonallocating functions request none. This is not a loop correctness
proof or workload-specific detector. Existing capacity, payload and quota guards
continue to select the checked fallback, and every exit publishes or rolls back.
No extra CFG analysis or generated instruction family is added.

The 86 native bodies across both architectures and policy tiers are identical;
foreign bodies are compile-only. Tests cover constructors before/inside/after
loops, the 64/65 straight-line boundary, metered execution, invalid requested
sizes, partial publication/reuse and quotas. Existing 255-constructor trap and
cancellation tests still require the full native batch. All 31 Go packages and
30 package coverage gates pass. Affected codegen/state/nativeexec/VM execute with
pinned LLGo; scoped race, no-cgo and vet pass. Final contribution-head CI is a
separate integration requirement.

## Matching-work measurements

Shared M4 Max/macOS arm64, Go 1.27.3 and pinned LLGo `8e45ed5`, `-O3 -lto=full`.
Seven unchanged kernels, five fresh process blocks, two warmups/seven calls,
serial sample interleaving and exact baseline A/A aliases for both compilers.
GC uses 10000 iterations, calls/memory 1000000, empty control one. Named native
references retain their pinned versions; Wasmer/wazero GC rejections are explicit.
Ordinary unchecked C uses the same scalar work/counts and remains a separate
policy reference. All preparation, first/warm calls, readiness and cleanup stay
separate. 275 VM observations, ten C observations, 57 summaries, 55 alternate
checks and eight unsupported configurations complete.

Warm ns/source iteration, except empty control ns/call:

| Workload | Go native base → head | LLGo native base → head | Wasmtime | Wasmer | wazero native | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 3666.000000 → 4750.000000 | 4917.000000 → 5042.000000 | 1000.000000 | 3000.000000 | 375.000000 | — |
| core/function-calls | 0.781458 → 0.787916 | 0.776916 → 0.776292 | 0.899000 | 0.770000 | 1.275417 | 0.768000 |
| core/memory-load-store | 0.434042 → 0.431750 | 0.434167 → 0.427458 | 0.467000 | 0.520000 | 0.789500 | 0.273000 |
| gc/i31-roundtrip | 3.016700 → 3.037500 | 3.275000 → 3.120800 | 1.900000 | — | — | — |
| gc/struct-access | 1.916700 → 1.816700 | 2.237500 → 1.987500 | 1.200000 | — | — | — |
| gc/struct-allocation | 43.945900 → 45.858300 | 52.508300 → 53.595800 | 105.700000 | — | — | — |
| gc/struct-indirect-access | 5.283300 → 5.354100 | 5.279200 → 5.200000 | 2.800000 | — | — | — |

Short struct-access medians improve Go 5.2% and LLGo 11.2%. Paired-block LLGo
ratios range 0.830–1.120, so this is descriptive evidence, not a controlled win.
The allocation loop regresses Go 4.4% (A/A +0.8%) and LLGo 2.1% (A/A +2.0%);
its unchanged native body and batch size do not waive those adverse samples.
Go indirect access regresses 1.3% (A/A -2.1%). Empty-call changes remain noisy:
Go +29.6% with A/A +28.4%, LLGo +2.5% with A/A +20.3%. No repeat-until-pass,
control subtraction, corpus inflation or cross-campaign pooling is used.

An independent exact-archive diagnostic confirms the allocation loop still
publishes 39 batches/9962 native objects with 38 fallbacks on the third call,
for both sources/compilers. Its global keeps the 10000-node chain after explicit
collection. Third-call approximate host allocation remains 808 bytes for Go;
LLGo is 41728 → 41856 bytes. Cold and warm results are retained. These test
executables supply counters only, never production throughput.

## Three remaining gaps

* Competitors: native allocation is 53.596 versus Wasmtime 105.700 ns/iteration,
  but struct access is 1.988 versus 1.200 and indirect access 5.200 versus 2.800.
  Neither the allocation win nor short-sequence improvement closes GC parity.
* Go/LLGo: allocation is 53.596 versus 45.858 (1.169x). Empty adapter readiness
  is about 93.1 versus 3.9 ms, separately from warmed calls 5.04 versus 4.75 µs.
  These boundaries require independent compiler/runtime investigation.
* Ordinary native: memory is 0.427458 versus C 0.273000 (1.566x), despite leading
  Wasmtime 0.467000 in this campaign. Calls are much closer to C. Bounds, overflow,
  cold-entry and bounded cancellation obligations remain in the VM.

Existing LLGo #2932, #2957, #2958, #2965 and #2966 are open at this inspection.
[#2959](https://github.com/xgo-dev/llgo/issues/2959) is closed by merged
[#2963](https://github.com/xgo-dev/llgo/pull/2963); pinned results remain valid,
while current upstream was not rebuilt or timed. [#2971](https://github.com/xgo-dev/llgo/pull/2971)
remains open. No throughput-only gap becomes a new defect or duplicate report,
and this candidate does not change the compiler baseline.

The earlier complete combined interpreter/native portfolio remains distinct.
SIMD/RMW dispatch, x86 memory, latency/allocation and expanded WASI world
conformance still require work; neither this scoped candidate nor semantic
coverage completes those milestones.

[Lossless measurements, A/A blocks, source/binary identities, diagnostics and tests](evidence/performance-native-batch-hints-2026-11-03.json.gz)

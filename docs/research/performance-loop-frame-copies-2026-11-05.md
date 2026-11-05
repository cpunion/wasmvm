# Private-loop frame reconstruction: 2026-11-05

**Local Go/LLGo semantics pass. x86 Clang SHA emits fewer hot frame accesses;
elapsed competitive acceptance remains open.**
Engine `993cbb3a59f9b19b58523ce8f239c2d6b37ec3dd`; baseline document head
`2b37289`, engine `3355a8d5293198474c8cd97379317093f9e983ae`.
The [preceding source events](performance-cfg-value-allocation-2026-11-05.md)
retain their separate failed and corrected measurements.

## Frame mechanism

A bounded private-loop pass reuses the existing write-version graph. A copy-only
nonresident integer word can avoid its hot frame write when its entry value is
unread and an independent materialized word carries exactly its current version
at every subsequent point. Read aliases follow those versions as the original
source changes. Chosen anchors stay materialized, so reconstruction has no cycles.
Shared homes stay excluded as anchors: a new read could otherwise extend a
version beyond the lifetime proven by coalescing. Variable residency plans retain
their current behavior.

Selectors stop at deferred writes and alias changes. Observation scans account
for the forwarded reads and cannot cross a changing mapping. Ordinary exits and
the generated cold poll exit reconstruct every deferred raw word through the
existing move emitter. Taken private backedges omit reconstruction only because
the next iteration overwrites each target before reading its entry word. Header
polls reject this transform. Every original cold PC and the precise tier retain
their original instructions. Recipes currently cover i32/i64 copies; general
value rematerialization remains open.

The ordinary full-CFG analysis also skips reconstructing write versions when both
physical banks can hold every slot and there are no handler/prompt entries.
Conservative slot allocation stays available. This avoids paying for a global
pressure analysis on uncongested functions; it does not remove guest operations.

## Semantic evidence

The complete codegen/nativeexec/public VM suites pass with Go 1.27.3 and LLGo
`3571c25` (`-O3 -lto=full`) on macOS arm64. Vet passes. The three suites
cross-cover **codegen 97.65% (8286/8485)** and **nativeexec 92.54% (459/496)**.
These are exact package measurements, not specification or competitor coverage.

Independent instruction models compare every original cold PC and all frame
words, ordinary exits, the first generated private backedge poll, last-valid
memory accesses and a subsequent out-of-bounds trap. Cases include i32/i64,
local.get/set/tee, changing anchors, repeated target definitions, arithmetic
wraparound, nonzero upper raw words and source bits above i32. The precise tier
supplies an additional reference for its canonical-word cases. Initial-read and
header-poll cases reject deferral. An anchor gap and zero work budget also reject
it. Both targets compile; the local arm64 host executes. Earlier development
compile/test-construction failures remain recorded separately.

## Actual code

All 66 baseline-Go/head-Go/head-LLGo arm64 publication witnesses pass the
independent result oracles. All **22 baseline/candidate images and entry-offset
lists match**, and candidate Go/LLGo bytes agree. This architecture's fixtures
therefore establish no new guest-code throughput benefit from frame recipes.
No redundant sustained arm64 timing campaign was collected.

Production x86 module compilation uses the unchanged feature mask 7 and stops
before native publication on the arm64 host. The Clang compression loop selects
two deferred words; the Rust compression loop selects none.

| x86 compression loop | Hot instructions | Frame references | Polls | Complete function bytes |
| --- | ---: | ---: | ---: | ---: |
| Clang | 92 → 88 | 24 → 20 | 1 → 1 | 10956 → 10984 |
| Rust | 85 → 85 | 18 → 18 | 1 → 1 | 10866 → 10866 |

Both schedule loops remain at 47 instructions, zero frame references and one
poll. Cold materialization grows the Clang function by 28 bytes despite its
shorter hot body. Static counts do not establish elapsed improvement.

## Go preparation allocation events

`testing.AllocsPerRun(5)` performs one warmup and five fresh Engine/Compile/Close
operations per unchanged module. Fixture reads and serialization are outside
counting; no Store or guest call participates. The diagnostic compiles for the
local macOS arm64 target. These are Go heap event counts,
not LLGo counters, executable/peak memory or elapsed latency.

All 22 cases record fewer events, by **119–730 per compile**:

| Workload | Baseline events | Candidate events |
| --- | ---: | ---: |
| core/integer-loop | 1019 | 623 |
| memory/narrow-unaligned | 1605 | 875 |
| SHA-256 / Clang | 7511 | 6910 |
| SHA-256 / Rust | 7480 | 6838 |

[The preceding CI](https://github.com/cpunion/wasmvm/actions/runs/42036970753)
completes for `2b37289`/engine `3355a8d`: all 21 required jobs pass, with one
optional skip. The longest job is 565 s (9m25s), excluding queue time. These
results apply to the baseline; current final-head validation remains required.
Contribution review also remains required.
[Linux comparison 42051975324](https://github.com/cpunion/wasmvm/actions/runs/42051975324)
now completes for exact source `993cbb3` against `3355a8d`: all six required
build/measurement jobs pass, 22 cases, 362 alternate checks and 12 expected
reference-GC rejections. Seven cases resolve all four diagnostic controls. The
shared build takes 405s; the longest measurement group takes 495s, excluding
queue time. Complete separate lifecycle phases and all competitor/native-C rows
remain retained. No timing is pooled with another source or target.

Clang SHA records Go/LLGo warm elapsed ratios of 0.941/0.957 against baseline;
Rust records 0.994/1.000 with unresolved controls. The Clang frame-traffic change
therefore accompanies a 4.3% lower LLGo median in this advisory event; it does
not complete controlled acceptance. Smaller elapsed ratios are faster.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.221× | 1.358× | 7.696× | 4.084× | 2.288× | 3.400× | 2.049× | unavailable | inconclusive |
| core/integer-loop | 1.011× | 0.809× | 1.022× | 1.021× | 1.017× | 0.344× | 0.547× | 1.023× | inconclusive |
| core/function-calls | 1.001× | 0.816× | 0.511× | 0.499× | 0.498× | 0.368× | 0.846× | 0.829× | inconclusive |
| core/memory-load-store | 1.001× | 0.930× | 0.714× | 1.172× | 1.170× | 0.516× | 0.714× | 1.613× | inconclusive |
| memory/narrow-unaligned | 0.995× | 0.967× | 0.862× | 1.315× | 1.312× | 0.499× | 0.718× | 2.081× | resolved, advisory |
| memory/dependent-address | 0.987× | 0.893× | 1.086× | 1.442× | 1.441× | 0.562× | 0.358× | 1.685× | resolved, advisory |
| memory/float64-unaligned | 1.008× | 0.946× | 0.714× | 1.184× | 1.177× | 0.700× | 0.415× | 1.199× | resolved, advisory |
| runtime/simd-memory-full | 0.994× | 0.938× | 0.891× | 1.105× | 1.069× | 0.427× | 0.979× | 0.924× | resolved, advisory |
| runtime/simd-memory-widen | 1.002× | 1.023× | 0.844× | 1.906× | 1.899× | 0.440× | 1.006× | 0.492× | resolved, advisory |
| runtime/simd-memory-lanes | 0.995× | 1.129× | 0.981× | 1.183× | 1.192× | 0.483× | 1.176× | 1.376× | resolved, advisory |
| gc/i31-roundtrip | 0.987× | 0.908× | unsupported | 0.922× | unsupported | unsupported | 0.442× | unavailable | inconclusive |
| gc/struct-access | 1.043× | 0.845× | unsupported | 2.113× | unsupported | unsupported | 0.363× | unavailable | inconclusive |
| gc/struct-allocation | 1.300× | 1.393× | unsupported | 0.306× | unsupported | unsupported | 1.389× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.986× | 0.892× | unsupported | 2.849× | unsupported | unsupported | 0.764× | unavailable | inconclusive |
| threads/private-memory | 1.083× | 1.118× | 0.930× | 2.763× | 2.156× | 0.862× | 0.646× | unavailable | inconclusive |
| threads/shared-memory | 1.090× | 0.673× | 1.048× | 2.952× | 2.315× | 0.871× | 0.648× | unavailable | inconclusive |
| threads/atomic-rmw | 1.010× | 0.718× | 1.311× | 1.285× | 1.174× | 0.859× | 0.638× | unavailable | inconclusive |
| threads/notify-empty | 1.179× | 0.935× | 0.080× | 0.359× | 0.262× | 0.794× | 0.516× | unavailable | inconclusive |
| threads/wait-mismatch | 1.022× | 0.708× | 0.056× | 0.238× | 0.575× | 0.739× | 0.512× | unavailable | inconclusive |
| threads/wait-timeout | 0.863× | 0.688× | 0.006× | 0.050× | 0.024× | 0.097× | 0.001× | unavailable | inconclusive |
| applications/sha256-clang | 0.997× | 0.914× | 0.772× | 1.473× | 1.430× | 0.556× | 0.765× | 1.772× | resolved, advisory |
| applications/sha256-rust | 0.981× | 0.927× | 0.770× | 1.540× | 1.456× | 0.575× | 0.773× | 1.844× | inconclusive |

The source's [CI run 42057704970](https://github.com/cpunion/wasmvm/actions/runs/42057704970)
is **cancelled**, despite passing semantic evidence. Twenty required jobs pass;
VM shard tests pass in 394s, but finalizing a 26,623-byte artifact retries HTML
responses for 378s and exceeds the 12-minute job limit. The preserved archive
contains 240 passing, matching conformance reports across four profiles and all
30 package coverage gates (minimum 90.77%). This is not a green merge gate.
The [subsequent increment](performance-loop-compare-immediates-2026-11-06.md)
bounds required artifact finalization to two minutes and retains the failure.
New final-head CI and contribution review remain required.

Pinned local baseline/reference binaries are reused only after exact engine,
harness, toolchain, flag and executable checks. No new LLGo defect is inferred.

Integration follows the one-active-contribution policy while LLGo #2991 remains
open. LLGo #2983 stays with its existing contributor. Application-effective
general rematerialization, remaining preparation costs, broader WASI-world
qualification and controlled/release acceptance remain open. This increment does
not complete the requested performance goal or a milestone.

The [data index](data/loop-frame-copies-2026-11-05.json) retains exact source,
qualification, all allocation rows, publication identities and raw input hashes.
Local evidence: `artifacts/loop-frame-copies-20261105/evidence.tar.gz`, 1854290 bytes, 224 files; SHA-256 `e7da5ce8408823e1c1dfd70dedd48954ea5568a1bb31ee6513fed816dc0a868a`.

Supplemental Linux archive: `artifacts/loop-frame-copies-20261105/hosted-evidence-current.tar.gz`, 2095193 bytes; SHA-256
`841890717284818cec46c8f90f2d476ca71df69817e61be5bb0517e074a644f4`.

Cancelled-CI archive: `artifacts/loop-compare-immediates-20261106/ci-eab-evidence.tar.gz`, 6115008 bytes; SHA-256
`4f6df9a3797258f6ce4ce1815e017de6078e351256875418502013da0d807ee6`.

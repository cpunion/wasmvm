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
is dispatched once for the exact `993cbb3` source against `3355a8d`, with one shared
build and all five original-count groups. It retains the Go/LLGo, competitor and
native-C rows and separate lifecycle phases. It targets x86, where the actual
code changed; its queued observation is not a performance result. Pinned local
baseline/reference binaries are reused only after exact engine, harness,
toolchain, flag and executable checks. No new LLGo defect is inferred.

Integration follows the one-active-contribution policy while LLGo #2991 remains
open. LLGo #2983 stays with its existing contributor. Application-effective
general rematerialization, remaining preparation costs, broader WASI-world
qualification and controlled/release acceptance remain open. This increment does
not complete the requested performance goal or a milestone.

The [data index](data/loop-frame-copies-2026-11-05.json) retains exact source,
qualification, all allocation rows, publication identities and raw input hashes.
Local evidence: `artifacts/loop-frame-copies-20261105/evidence.tar.gz`, 1854290 bytes, 224 files; SHA-256 `e7da5ce8408823e1c1dfd70dedd48954ea5568a1bb31ee6513fed816dc0a868a`.

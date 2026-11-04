# Compact GC metadata and default allocation: 2026-11-04

The candidate reduces ordinary object records from **104 to 64 bytes** and
separates zero-operand constructors from temporary-reference cleanup. Against
merged PR115, the local long-chain native kernel improves **14.0% under LLGo**;
empty-object interpreter allocation improves **36.4%**. Long-chain allocation
still takes about **4.6x Wasmtime**. These shared-host observations do not close
controlled performance acceptance.

## Runtime contract

Ordinary structs and arrays keep a 64-byte common record. Exception tags,
custom descriptors and continuations use optional 32-byte cells. A 32-bit
optional-cell index fits the common record's existing padding; zero means
absent. This preserves the original per-object allocation epoch and rollback,
without a transaction journal or Store layout change. Optional cells are cleared
before handle reuse, including their host references. Tracing retains exception
tag precedence and continuation/descriptor roots.

Logical pages still contain at most 64 objects. Their typed backing allocations
grow through 1, 2, 4 and at most 8 logical pages, without moving live cells.
The last page/slab stops at the handle quota, including the null slot. Optional
cells materialize only for the corresponding slab. Small Stores start with one
page; large Stores allocate up to 512 common cells at once. Native views,
generations, serial exhaustion, payload layout, fuel/cancellation and byte/object
quotas are unchanged. No benchmark-specific guest pattern selects this path.

Both compilers independently report common records 104→64 bytes, optional cells
32 bytes, page headers 32→56 bytes and handle slots 16 bytes. Store remains
608 bytes under Go / 616 under LLGo. These are record sizes, not peak RSS or total
heap savings. The exact four-helper layout measurements from indexed source
`9a659e4afd29203e9f6cada81b765694f648e6ff` are reused only after verifying that
all type-declaring source bytes and compiler identities/profiles are identical;
only `gc_helpers.go` and its test changed in the measured follow-up.

Default constructors have no operand references to borrow. Their zero-operand
path now bypasses the rooted helper's deferred cleanup; real operands retain
the original panic-safe root extent. LLGo disassembly shows the old ordinary
path entering 48-byte exception-frame and 32-byte deferred-call allocations.
The new wrapper checks the operand count first and has no direct `AllocU` call.
This does not imply allocation-free object construction. The existing LLGo
[defer issue #2932](https://github.com/xgo-dev/llgo/issues/2932) remains separate
from a compiler fix. Issue #2983 is handled independently; no duplicate upstream
contribution was made.

## Fixed source, semantics and phases

Measured engine source is `e9b1c5114e472801bc653a3a37842178e5c2cba7`, against
`bd0c30c89829a4e1007655ce32be76483013714d`. The baseline engine inventory is
byte-identical to merged main `cf006c9169cf5aae69e227227f9056a1c07febb6`.
Go 1.27.3, LLGo `983465e779527ff496857a9a1915c1b57131091f`, LLVM 22.1.8 and
LLGo `-O3 -lto=full` remain pinned. No production `GOEXPERIMENT=simd` is enabled.

Both compilers execute five affected packages and 48 pinned reports each,
including all five mandatory extensions at three tiers and forced Stack
Switching collection. Tests cover stable cell addresses, truncated quotas,
optional slab isolation/clearing, generation reuse, nested rollback after GC,
native batch serial exhaustion, resolver failure/panic recovery, and default
constructor quota/collector failures while preserving outer roots. Final PR-head
platform CI, coverage and review remain required.

The Apple M4 Max/macOS arm64 host is shared. Every scope retains preparation,
first call, warm execution and cleanup independently; no startup, empty call,
timer or A/A score is subtracted. Kernel rates include the guest loop and
required collection. Counts, retention shapes and independent cohorts are never
pooled. A/A aliases use the exact baseline binaries.

## Longer affected scopes before reference-clock correction

Eight primary cases retain 300 observations / 60 alternate checks / 75 oracle
checks. Separate empty-object and typed-exception scopes each retain 70 / 14 /
75. Each uses five fresh interleaved process blocks, two warmups and seven warm
calls. The original reference clock profile is retained for these observations;
the separate clock correction below is not applied retroactively.

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | -6.25% / +0.30% | -14.00% / -1.04% |
| gc/struct-allocation / interpreter | -1.68% / +0.82% | -1.55% / +3.87% |
| gc/struct-access / native | +0.10% / -0.18% | +0.23% / +0.09% |
| gc/struct-indirect-access / native | +0.50% / +0.35% | +0.15% / -0.39% |
| gc/struct-access / interpreter | -0.31% / +0.36% | -1.88% / -1.32% |
| threads/wait-mismatch / native | +0.21% / +3.77% | +2.55% / +2.86% |
| gc/i31-roundtrip / native | +0.26% / -0.16% | -0.70% / -0.64% |
| core/memory-load-store / native | -1.89% / -0.54% | +1.43% / +0.77% |
| gc/empty-struct-allocation / native | -7.46% / +0.20% | -16.60% / +0.20% |
| gc/empty-struct-allocation / interpreter | -5.66% / -1.88% | -36.43% / -0.82% |
| exception/throw-catch / native | +2.18% / -1.64% | +1.30% / +9.09% |
| exception/throw-catch / interpreter | +0.61% / +4.95% | -0.71% / +0.75% |

The 300k-object linked kernel changes 47.504→40.854 ns/object under LLGo,
versus Go 42.666→39.998 and Wasmtime 8.920. LLGo remains 2.1% above Go and
4.58x Wasmtime. LLGo interpreter linked allocation is 131.220 versus WasmEdge
104.340; its −1.55% self change accompanies +3.87% A/A and does not establish a
resolved improvement.

The distinct 300k empty-object kernel retains only the newest escaped identities.
Native LLGo changes 19.247→16.052 ns/iteration, versus Wasmtime 20.583;
interpreter LLGo changes 107.269→68.193, versus WasmEdge 90.620. Native Go is
16.412 and interpreter Go 81.779. This is not the linked-chain retention shape.

Retain LLGo native memory +1.43% (A/A +0.77%), wait-mismatch +2.55%
(A/A +2.86%), and native exceptions +1.30% (A/A +9.09%). Go native exceptions
+2.18% (A/A −1.64%) is also retained. No observed regression is relabeled a pass,
and the noisy exception cohort is not replaced by an earlier result.

## Reference-clock baseline event and failed portfolio

The original twenty-case run failed its strict disjoint-interval check during
WasmEdge's integer-loop block, after two complete cases. It retains 103 VM
observations, its C control, 45 alternate checks and the partial block. The old
script validated before persisting the complete block, so the offending raw
WasmEdge record was lost; its identity is inferred from the final sample order
and saved prefix. Do not claim the missing record was recovered or validated.
The parent observed load averages 296.49 / 81.54 / 35.66 before the run and
257.82 / 117.73 / 52.82 after it. These are load averages, not CPU percentages.

Darwin's `clock_gettime(3)` documents that `CLOCK_MONOTONIC` includes system
sleep and that `CLOCK_UPTIME_RAW` matches `mach_absolute_time` while excluding
sleep. Existing C references used MONOTONIC or MONOTONIC_RAW; the Python observer,
Go and pinned LLGo use the uptime domain. This confirms a clock-domain mismatch,
not that sleep during the lost observation has been proven. The C references
and ordinary C controls now select `CLOCK_UPTIME_RAW` on Darwin, preserving Linux
CLOCK_MONOTONIC. A compiled test checks all four complete C sources against the
observer's before/after timestamps; all 82 benchmark-tooling tests pass.

Rebuilt reference binaries form an explicit separate baseline event. Engine/SDK
pins, owned binaries and compiler flags stay unchanged; header/source/compiler
and binary hashes are retained. The failed portfolio remains independent. The
new scope saves each complete raw block before validation and keeps the strict
wall-time check. No tolerance, correction or subtraction hides the failure.

## Separate clock-aligned original portfolio

The corrected reference-clock scope verifies **864 VM/reference observations,
27 C observations, 288 alternate checks, 12 unsupported-reference outcomes and
297 summaries**. Three fresh interleaved blocks retain the original counts:
one public-call control, 1M ordinary iterations and 10k GC/notify/wait iterations,
two warmups and seven warm calls. All preparation and first-call phases remain
in the bundle. Ordinary C lacks Wasm bounds, cancellation, fuel and guest GC.
The host's load averages fall from 93.54 / 52.85 / 45.80 to
29.91 / 33.94 / 40.86 during the scope. This is not a controlled acceptance run.

Native rates, ns/source iteration (call-control is ns/public call):

| Workload | Go | LLGo | Wasmtime | Wasmer | wazero compiled | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.9956 | 1.2527 | 2.0871 | 2.2333 | 2.7118 | — |
| core/call-control | 13209 | 8625 | 3500 | 7292 | 1500 | — |
| core/integer-loop | 3.3896 | 3.3486 | 3.4452 | 3.3625 | 3.7093 | 3.3381 |
| core/function-calls | 1.7896 | 1.7035 | 2.0014 | 1.6805 | 3.23 | 1.6794 |
| core/memory-load-store | 0.9774 | 0.9691 | 1.0532 | 1.1368 | 2.1763 | 0.5917 |
| memory/narrow-unaligned | 1.279 | 1.1549 | 1.2823 | 1.2129 | 2.033 | 0.7482 |
| memory/dependent-address | 3.0874 | 2.7497 | 3.0846 | 3.486 | 4.6562 | 2.3359 |
| memory/float64-unaligned | 0.9991 | 0.9882 | 0.9681 | 0.9785 | 1.2155 | 0.905 |
| runtime/simd-memory-full | 4.7738 | 4.6079 | 5.3668 | 5.116 | 5.5239 | 4.9284 |
| runtime/simd-memory-widen | 1.7501 | 1.7337 | 1.9214 | 1.7391 | 1.9838 | 1.9192 |
| runtime/simd-memory-lanes | 6.8946 | 6.7891 | 7.916 | 7.9467 | 8.2394 | 6.0966 |
| gc/i31-roundtrip | 2.5041 | 2.4375 | 2.7083 | unsupported | unsupported | — |
| gc/struct-access | 2.9666 | 3.0125 | 1.8125 | unsupported | unsupported | — |
| gc/struct-allocation | 44.925 | 51.4 | 147.0625 | unsupported | unsupported | — |
| gc/struct-indirect-access | 7.6583 | 7.0084 | 3.9709 | unsupported | unsupported | — |
| threads/shared-memory | 1.0938 | 1.0782 | 1.4023 | 1.278 | 1.5998 | — |
| threads/atomic-rmw | 3.4073 | 3.9132 | 3.8856 | 6.52 | 3.7528 | — |
| threads/notify-empty | 4.325 | 3.2125 | 15.0583 | 10.9875 | 47.8875 | — |
| threads/wait-mismatch | 2.6584 | 2.2375 | 12.4791 | 3.8083 | 34.5666 | — |
| threads/wait-timeout | 3.2916 | 2.4083 | 64.4334 | 252.3334 | 355.1458 | — |

Interpreter rates retain the same original counts and units:

| Workload | Go | LLGo | WasmEdge | wazero interpreter | LLGo change / A/A |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 152.0794 | 130.9659 | 116.8337 | 153.7847 | +32.24% / -7.50% |
| core/call-control | 13542 | 8250 | 8542 | 3166 | -13.91% / +10.44% |
| core/integer-loop | 92.234 | 72.9395 | 155.1393 | 190.8757 | -8.33% / -8.11% |
| core/function-calls | 150.793 | 185.9586 | 205.9493 | 222.592 | +15.69% / -11.37% |
| core/memory-load-store | 190.2619 | 137.3667 | 204.0257 | 332.3268 | -0.63% / +3.00% |
| memory/narrow-unaligned | 206.0042 | 134.4379 | 204.4696 | 339.6815 | -16.96% / +3.58% |
| memory/dependent-address | 91.1776 | 61.359 | 87.185 | 139.1469 | -6.53% / +1.97% |
| memory/float64-unaligned | 114.0039 | 89.9133 | 115.57 | 186.903 | -0.08% / -0.76% |
| runtime/simd-memory-full | 127.0605 | 105.2103 | 115.8316 | 203.2848 | +4.73% / -1.50% |
| runtime/simd-memory-widen | 125.1056 | 103.9135 | 110.9052 | 221.36 | -6.22% / +2.81% |
| runtime/simd-memory-lanes | 151.1891 | 124.8828 | 124.1223 | 252.7592 | +0.90% / -0.35% |
| gc/i31-roundtrip | 64.5291 | 49.7125 | 60.6291 | unsupported | +1.48% / +1.35% |
| gc/struct-access | 71.7542 | 57.4916 | 75.0209 | unsupported | +0.21% / -0.49% |
| gc/struct-allocation | 146.7292 | 165.2958 | 142.8708 | unsupported | -9.11% / -0.08% |
| gc/struct-indirect-access | 129.0583 | 103.65 | 131.2583 | unsupported | -2.26% / -0.44% |
| threads/shared-memory | 111.6471 | 52.2447 | 52.594 | 84.6574 | +0.87% / -0.03% |
| threads/atomic-rmw | 109.4645 | 88.0028 | 88.6885 | 104.3562 | +3.62% / +1.73% |
| threads/notify-empty | 88.7875 | 70.3792 | 87.7666 | 93.6375 | -3.24% / -1.58% |
| threads/wait-mismatch | 92.9917 | 63.6167 | 59.4084 | 96.5791 | +2.75% / +1.24% |
| threads/wait-timeout | 98.0791 | 59.0083 | 143.95 | 423.1583 | -2.08% / -1.65% |

Retain LLGo interpreter private-memory **+32.24%** (A/A −7.50%), function calls
**+15.69%** (A/A −11.37%), full SIMD memory **+4.73%** (A/A −1.50%), atomic RMW
**+3.62%** (A/A +1.73%) and wait-mismatch **+2.75%** (A/A +1.24%). Retain native
LLGo short GC access **+19.11%**, atomic RMW **+13.77%**, private memory **+4.22%**
and wait-mismatch **+3.67%**. The separate long GC access scope's +0.23% with
+0.09% A/A does not overwrite the original short outcome. Go regressions,
including native private-memory +37.12%, interpreter private-memory +2.19%,
dependent-address +27.57% and public-call +105.71%, remain in the raw record.
These results prevent a portfolio-parity or regression-free claim.

Disassembly of the baseline, indexed-metadata and default-constructor LLGo
runners retains 2,736 instructions in `Execution.Run`, with the same instruction
and branch positions. Observed differences involve PC-relative data/dispatch
table addresses. This does not exclude cache/layout effects or identify the
performance regression's cause. The independent longer-count three-source control below preserves these
outcomes and investigates the duration/source effects.

## Independent longer-count three-source control

This follow-up is a distinct duration/ablation experiment, not a replacement for
the original twenty-case cohort. It uses 5M private-memory and function-call
iterations, the same fixed main/indexed/default binaries, five fresh interleaved
blocks, two warmups and seven warm calls. Exact main A/A aliases and the corrected
reference-clock profile are retained. It verifies 100 observations, 20 alternate
checks, 75 oracle sanity checks and 20 summaries. Own process medians span
264–519 ms, rather than interpreting one public call as a throughput result.

Rates are ns/source iteration; changes compare default source with main:

| Kernel / compiler | Main | Indexed | Default | Change / A/A |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory / go | 87.485 | 84.368 | 82.341 | -5.88% / +3.62% |
| threads/private-memory / llgo | 67.923 | 67.989 | 68.995 | +1.58% / -2.66% |
| core/function-calls / go | 79.802 | 80.401 | 80.134 | +0.42% / +0.27% |
| core/function-calls / llgo | 72.667 | 73.124 | 72.975 | +0.42% / -0.46% |

The original large interpreter regressions are not reproduced at these counts;
this does not identify their cause or erase them. Address/layout effects and
shared-host scheduling remain unresolved. The source adds no instructions to
these non-GC interpreter paths. Final platform semantics and an independent
Linux PR comparison remain required before integration; controlled portfolio
acceptance stays open.

## Earlier variants and remaining work

All sources compare the same verified merged-PR115 engine inventory and pinned
compiler profile. The earlier prototype plan's `baselinePR` label was stale;
its `baselineMain`, actual binary and full source hashes identify `cf006c9`.
The stale label is retained in the raw evidence rather than rewritten.

| Source | Representation | LLGo native linked allocation | LLGo interpreter empty allocation | Decision |
| --- | --- | ---: | ---: | --- |
| `9acf20b` | 64-byte common records, optional pointer, transaction journal, individual pages | +1.94% | +6.57% | held |
| `57ee512` | same records/journal with bounded coallocated slabs | −6.17% | +3.33% | held |
| `9a659e4` | indexed optional cells, original epochs/rollback and bounded slabs | −14.41% | +3.64% | held |
| `e9b1c51` | indexed cells plus zero-operand cleanup split | −14.00% | −36.43% | full-portfolio regressions under investigation |

These are independent five-block cohorts, not stages of one averaged score.
No held source received a redundant original twenty-case run: the prepared
scopes have no timing results. An independent BDWGC size probe retains allocator
rounding observations; it does not measure physical collector block consumption
or prove an LLGo bug. No duplicate allocator issue was opened.

Long-lived allocation metadata traffic, collector strategy, indirect GC access,
interpreter regressions, host boundaries, ordinary-C ratios, controlled portfolio
acceptance and expanded WASI acceptance remain open. The implementation scope
and performance acceptance are distinct. No milestone is completed by these
measurements.

## Durable evidence

The [evidence packet](evidence/performance-gc-compact-2026-11-04.json.xz) retains
all four source patches, plans, build/compiler/binary/source hashes, full affected
Go/LLGo execution reports, record-size proofs, disassembly, all raw phases and
adverse outcomes, clock-baseline provenance and the failed portfolio limitation.
It contains **781 files**, **857,400 compressed bytes**, SHA-256
`1bfe1552262c11e425c620aa0e0b2b74e1fa69cca2b8233f02ba3aacb6355b8b`. Each JSON entry declares UTF-8 or base64 content and the decoded original
SHA-256; every original byte/hash round-tripped successfully before publication.
Use Python `lzma.decompress` and `json.loads` to inspect it. No SDK or executable
binary is bundled; pinned build/source provenance is retained.

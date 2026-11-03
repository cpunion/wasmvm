# Native heap sweep and page summaries: 2026-11-04

The candidate reduces long linked-allocation time by **8.5% under LLGo** and
**5.9% under Go** against the merged PR114 engine. Native field access stays
near its baseline. Empty allocation retains a smaller regression but remains
faster than Wasmtime in its separate native study. Long linked allocation still
takes **5.32x Wasmtime**; this increment does not close performance acceptance.

Local affected execution, original twenty-case comparison and separate duration
controls are verified. Final PR-head CI and review remain required. The
[PR114 final audit](pr114-final-ci-2026-11-04.md) establishes the merged baseline.

## Runtime change

Sweep nonempty native objects while compacting their payloads, using the same
generation-checked handles. Relocation records shrink from **24 to 16 bytes**
and contain no host-object pointer. Stationary payloads need no copy or address
publication. A shared live-cell release retains quotas, cleared metadata,
generation retirement, native views and the existing free-handle chain.

Empty structs, zero-length arrays, zero-argument exceptions and continuations
have no payload to relocate. Their existing 64-cell metadata pages gain a
conservative `hasEmpty` summary. Allocation sets the flag; collection scans
only flagged pages and resets each summary to its survivors. A cleared layout
already identifies released cells. Rollback retains no per-object tombstone or
separate bitmap allocation. Native batches coalesce repeated page writes.
A mixed page can require both scans; this is not a claim that every heap has
only one traversal. Portable collection retains its full slot sweep.

Actual Go/LLGo reflection reports 104-byte objects and 16-byte slots unchanged,
page headers 24→32 bytes, and the native-heap header unchanged at 152 bytes.
Store sizes remain 608 Go / 616 LLGo. The grouped-publication follow-up changes
only a function body; identical type files and compiler identities justify reuse
of that exact layout proof. These are record sizes, not measured peak RSS.

The byte/object quotas, serial overflow, guest layout, root graph, safepoints,
mandatory cancellation and fuel contracts are unchanged. Both compilers execute
five affected packages and 48 pinned reports each, including all five mandatory
extensions at three tiers and forced Stack Switching collection. New tests
cover mixed/empty batches, unpublished reservations, stale generations,
rollback, arena growth, repeated collection and post-sweep thresholds.

## Fixed source and measurement

Measured source `4a30329c641c85c0b9400e6ed63268931be08475` compares with
`18d7b09592b6b93e0b1d2e24e2321a9a4cc7714d`, whose engine inventory equals
merged main `cb1e2459b88fbac32dcb77eb762fb3c5143ef12e`. Go 1.27.3 and LLGo
`983465e779527ff496857a9a1915c1b57131091f` retain their pinned profiles;
LLGo uses `-O3 -lto=full`, without production `GOEXPERIMENT=simd`.
LLVM and reference/SDK pins are unchanged. This Apple M4 Max/macOS arm64 host
is shared; the observations are descriptive, not controlled acceptance.

Every scope preserves preparation, first call, warm call and cleanup separately.
Kernel rates include the guest loop and necessary collection; they are not bare
allocator timings. No clock, startup, empty-call or A/A median is subtracted.
Exact baseline A/A aliases, alternate inputs and independent integer oracles are
retained. Different counts, seeds and retention shapes are never pooled.

The six-case primary study verifies 220 observations, 44 alternate checks,
50 oracle checks and 44 summaries: five fresh process blocks, two warmups and
seven measured calls. Linked allocation uses 300,000 objects with a 12-byte
packed payload and 64 accounting bytes each. Old/new live chains fit the
unchanged 64 MiB budget. Values below are ns/object for complete guest kernels.

| Backend | Go base → head | LLGo base → head | Reference |
| --- | ---: | ---: | ---: |
| native | 45.873 → 43.170 (-5.9%) | 50.857 → 46.533 (-8.5%) | Wasmtime 8.743 |
| interpreter | 109.349 → 112.201 (+2.6%) | 136.639 → 136.249 (-0.3%) | WasmEdge 103.783 |

Native allocation A/A is +0.64% Go / −0.02% LLGo. LLGo native remains 7.8%
above Go. Retain Go interpreter allocation +2.61% (A/A +1.20%) and direct field
access +1.90% (A/A −0.73%). LLGo interpreter allocation −0.29% versus A/A +0.40%
is not evidence of a resolved throughput improvement.

The separate zero-payload workload verifies 70 observations, 14 alternate and
75 oracle checks, with the same five-block protocol and 300,000 allocations.
Each fresh empty identity is compared with the preceding escaped object, and
only the latest global/local identities survive. This is a different retention
shape from the linked chain. LLGo native changes 18.892→19.215 ns/iteration
(**+1.71%, A/A +0.32%**), versus Wasmtime 20.680. Go changes 17.479→17.711
(+1.32%, A/A +0.58%). Retain Go interpreter +2.43% (A/A +0.18%); LLGo interpreter
−0.31% versus A/A −0.79% is inconclusive. The fixture is project-authored and
independently parsed/validated by pinned wasm-tools 1.262.0.

## Original portfolio and ordinary native C

The unchanged original twenty-case scope verifies 864 VM/reference observations,
27 C observations, 288 alternate checks, twelve explicit unsupported GC outcomes
and 297 regenerated summaries. Three fresh interleaved blocks, two warmups and
seven measured calls retain the original counts: one call-control invocation,
1M ordinary iterations and 10k GC/notify/wait iterations. C is an explicitly
unchecked reference without Wasm bounds, fuel, cancellation or guest-GC contracts.
A dash means no measured equivalent; GC rejections remain `unsupported`.

Native rates, ns/source iteration (call-control is ns/public call):

| Workload | Go | LLGo | Wasmtime | Wasmer | wazero compiled | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.3364 | 0.3185 | 1.020 | 0.8120 | 1.129 | — |
| core/call-control | 3334.000 | 3125.000 | 1000.000 | 3000.000 | 416.000 | — |
| core/integer-loop | 1.573 | 1.537 | 1.553 | 1.539 | 1.695 | 1.539 |
| core/function-calls | 0.7963 | 0.7840 | 0.8920 | 0.7700 | 1.415 | 0.7720 |
| core/memory-load-store | 0.4486 | 0.4348 | 0.4780 | 0.5280 | 0.8709 | 0.2678 |
| memory/narrow-unaligned | 0.7055 | 0.6847 | 0.6910 | 0.7180 | 1.039 | 0.4581 |
| memory/dependent-address | 1.850 | 1.837 | 2.067 | 2.095 | 3.140 | 1.410 |
| memory/float64-unaligned | 0.7389 | 0.7228 | 0.7080 | 0.7150 | 0.8546 | 0.6644 |
| runtime/simd-memory-full | 3.381 | 3.361 | 3.607 | 3.427 | 3.726 | 3.479 |
| runtime/simd-memory-widen | 1.055 | 1.040 | 1.028 | 1.041 | 1.143 | 1.037 |
| runtime/simd-memory-lanes | 5.074 | 5.051 | 5.777 | 5.818 | 6.049 | 4.407 |
| gc/i31-roundtrip | 1.821 | 1.821 | 2.100 | unsupported | unsupported | — |
| gc/struct-access | 2.025 | 1.938 | 1.400 | unsupported | unsupported | — |
| gc/struct-allocation | 40.900 | 42.967 | 105.100 | unsupported | unsupported | — |
| gc/struct-indirect-access | 5.342 | 5.192 | 3.000 | unsupported | unsupported | — |
| threads/shared-memory | 0.8148 | 0.7880 | 1.017 | 0.8140 | 1.125 | — |
| threads/atomic-rmw | 1.822 | 1.810 | 1.812 | 3.854 | 1.944 | — |
| threads/notify-empty | 1.446 | 1.417 | 6.900 | 5.100 | 20.712 | — |
| threads/wait-mismatch | 1.925 | 1.625 | 9.000 | 3.100 | 24.887 | — |
| threads/wait-timeout | 2.067 | 1.617 | 43.100 | 168.200 | 240.200 | — |

Interpreter rates use the same original counts and units:

| Workload | Go | LLGo | WasmEdge | wazero interpreter | LLGo/base change |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 48.787 | 42.183 | 38.401 | 62.130 | +1.02% |
| core/call-control | 2500.000 | 2834.000 | 2000.000 | 792.000 | +9.67% |
| core/integer-loop | 41.130 | 32.923 | 48.639 | 81.528 | +0.85% |
| core/function-calls | 63.044 | 56.331 | 62.078 | 79.456 | -0.27% |
| core/memory-load-store | 81.403 | 69.679 | 90.702 | 145.493 | -0.51% |
| memory/narrow-unaligned | 107.189 | 92.322 | 113.268 | 182.552 | -0.40% |
| memory/dependent-address | 53.771 | 45.218 | 53.327 | 95.035 | +0.51% |
| memory/float64-unaligned | 86.050 | 71.970 | 83.937 | 138.409 | -0.53% |
| runtime/simd-memory-full | 91.559 | 74.347 | 77.258 | 149.032 | +0.67% |
| runtime/simd-memory-widen | 75.247 | 64.005 | 60.577 | 142.444 | -0.89% |
| runtime/simd-memory-lanes | 113.607 | 92.584 | 90.148 | 186.953 | -1.02% |
| gc/i31-roundtrip | 49.867 | 37.554 | 45.000 | unsupported | +2.35% |
| gc/struct-access | 54.396 | 44.112 | 54.300 | unsupported | -0.87% |
| gc/struct-allocation | 111.850 | 140.900 | 105.600 | unsupported | +3.58% |
| gc/struct-indirect-access | 98.446 | 83.625 | 96.000 | unsupported | +0.80% |
| threads/shared-memory | 82.225 | 40.852 | 38.353 | 61.449 | -0.11% |
| threads/atomic-rmw | 61.240 | 40.007 | 31.698 | 43.731 | -2.35% |
| threads/notify-empty | 39.862 | 35.550 | 40.200 | 42.062 | -0.29% |
| threads/wait-mismatch | 69.529 | 48.871 | 43.800 | 69.162 | -1.39% |
| threads/wait-timeout | 67.088 | 45.100 | 96.200 | 289.371 | +2.34% |

Integer/call/SIMD point estimates approach the unchecked C kernels; ordinary
memory load/store still takes about 1.62x C under LLGo. This does not establish
portfolio parity. All base/head values, all phases and both host compilers remain
in the bundle, including Go regressions and tiny public-boundary controls.

Short linked allocation reports 42.967 ns under LLGo versus Wasmtime 105.100,
whereas the separately planned 300k-object study reports 46.533 versus 8.743.
Keep both. Do not use the short case to overturn the sustained deficit: the
scale dependence has not been isolated. Collection-count, retention-policy and
representation diagnostics remain necessary; startup subtraction cannot explain
these warm intervals.

Retain original LLGo interpreter call-control +9.67% (A/A −20.98%), short linked
allocation +3.58% (A/A +1.91%), i31 +2.35% (A/A +0.80%), zero-timeout +2.34%
(A/A +0.97%), and native zero-timeout +4.30%. They are not relabeled passes.

The separate three-case longer-count control verifies 115 observations,
23 alternate checks, 75 oracle checks and 23 summaries, with five fresh blocks.
It uses 100M native/5M interpreter i31 iterations and 5M native timeout iterations,
with exact native/interpreter A/A:

| Control | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/i31-roundtrip native | +0.20% / +0.18% | +0.03% / -0.03% |
| gc/i31-roundtrip interpreter | +0.34% / -0.28% | +0.61% / +0.89% |
| threads/wait-timeout native | -3.27% / +1.72% | -3.89% / -3.33% |

These controls remain separate observations, not replacements or adjusted scores.
Native timeout's −3.89% LLGo change accompanies A/A −3.33%; it does not prove a
source win. The original public-boundary noise and all short outcomes remain.

## Earlier variants and evidence

All screened sources remain recorded: full per-object empty records regressed
LLGo empty allocation 15.52%; a bitmap narrowed this to 5.57%; once-per-word
bitmap sweeping left 4.92%; page summaries left 2.87%; grouped publication leaves
1.71%. Each is a source change, not an unchanged timing retry. Earlier six-case
allocation improvements and all adverse results remain available and are not
pooled into the selected source. A read-only recursive tool lookup overlapped
later cases in the first variant; its incident is retained. Its attempted full
portfolio command failed on a broken local evidence symlink before any result
file or timing; that source was held and was not run again. The selected source's
original twenty-case scope completed once.

The [evidence bundle](evidence/performance-native-gc-sweep-2026-11-04.json.gz)
retains raw scope results, phases, commands, compiler/binary/source hashes,
A/A aliases, checks, held-source decisions and the prior final CI audit.
It contains 772 files (5,471,063 compressed bytes), SHA-256
`a1fad365a334a2c5fdef885be98092351d2a94dbce87883743709b2d83b15290`. Every decoded original byte/hash was verified.

LLGo's named-result fix PR2963 is already production pin `983465e`. New compiler
changes require a separate baseline event. LLGo issue2983 is being handled
independently; no duplicate upstream contribution was made here. Native GC,
interpreter gaps, public boundaries, ordinary-C ratios, controlled portfolio
acceptance and expanded WASI acceptance remain open.

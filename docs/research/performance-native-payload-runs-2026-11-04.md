# Native GC payload runs and held metadata layouts: 2026-11-04

Merged PR #116 remains the source baseline. Two further metadata reductions were
rejected after their fixed-source measurements. The new candidate instead
coalesces adjacent live payloads into overlapping-safe relocation runs.
Controlled performance acceptance remains open.

## Contracts and diagnostic profiles

The source baseline uses 64-byte common object records, 16-byte generation/free
slots and a 32-byte native view. CPU sampling uses the exact production-profile
LLGo runner, the original 300k-node linked fixture and quota, 1000 calls and a
three-second macOS sample. These are diagnostics, not timing acceptance or
retries of earlier measurements. Main-thread sample counts are not normalized
wall times.

| Source | Main-thread samples | Collector | Batch finish | Batch preparation |
| --- | ---: | ---: | ---: | ---: |
| Merged #116 engine alias e9b1c51 | 2373 | 1279 | 762 | 95 |
| Direct cells 92cfac3 | 2374 | 1327 | 633 | 211 |
| Numeric slots cc87358 | 2368 | 1427 | 626 | see raw profile |

The direct-cell profile adds a `heapCell` call during reservation. Numeric slots
retain dense generation/free-chain scans, but every live lookup still resolves
a cell through its page. Neither source reduces linked-allocation time. The
numeric profile attributes 663 samples to payload compaction, including 134
inside `memmove`; these motivate relocating adjacent survivors as one run.
The samples do not establish a compiler defect. LLGo #2983 is independently
owned; no duplicate contribution or new compiler issue is opened here.

The pinned Wasmtime 48.0.1 SDK enables DRC and null collectors but not copying.
The harness uses Auto, and the [exact selection implementation](https://github.com/bytecodealliance/wasmtime/blob/7bac2c2775808aaec5d4aa5627a5e447b51102cf/crates/wasmtime/src/config.rs)
prefers copying if compiled, then DRC, then null. Thus Auto selects DRC in this
SDK; this is an inference from source and the matching generated header, not a
runtime getter. Its documentation says cycles remain until Store drop. The
owned tracing collector continues to reclaim unreachable cycles. Collector
policy helps explain workloads but does not waive the measured throughput gap.

## Held source-changing metadata experiments

Both experiments retain the unchanged 64-byte common records, payloads, quota,
native 32-byte ABI, cycle collection and compiler pins. Direct cells move
generation/free links into common records and eliminate slots. Numeric slots
keep generation/free links in 8-byte pointer-free slots and use the high
generation bit as publication state. Both preserve low 31-bit handle generations.

Each experiment independently retains eight primary cases (300 observations),
sparse heap, empty allocation and typed exceptions (70 each): 510 observations,
102 alternate-input checks and 75 oracle checks per scope. Each case uses five
fresh interleaved process blocks, two warmups and seven warm calls. A/A uses
the exact baseline binaries. The original outcomes are neither pooled nor
replaced by later cohorts. No full twenty-case portfolio is run for a held source.

Rates below are source-head changes and exact-binary A/A changes; negative means
lower time. Shared-host observations are descriptive.

### Direct layout

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | +8.24% / +1.67% | +18.79% / +0.38% |
| gc/struct-allocation / interpreter | +1.96% / +0.93% | +1.41% / -0.33% |
| gc/struct-access / native | -0.00% / +0.22% | -0.22% / -0.09% |
| gc/struct-indirect-access / native | -0.17% / +1.26% | -1.93% / -0.64% |
| gc/struct-access / interpreter | +2.44% / -0.24% | +0.06% / +1.46% |
| threads/wait-mismatch / native | -2.41% / -1.11% | -7.99% / -0.76% |
| gc/i31-roundtrip / native | +0.10% / +0.49% | -0.06% / +1.44% |
| core/memory-load-store / native | -0.32% / +0.18% | -0.17% / -0.24% |
| gc/sparse-empty-allocation / native | +7.36% / +0.27% | +14.11% / -0.77% |
| gc/sparse-empty-allocation / interpreter | +18.18% / +0.28% | +23.66% / -0.01% |
| gc/empty-struct-allocation / native | +6.09% / +0.00% | +15.62% / -0.40% |
| gc/empty-struct-allocation / interpreter | -0.51% / -1.23% | -0.47% / +0.03% |
| exception/throw-catch / native | +0.35% / -0.19% | -3.18% / -1.19% |
| exception/throw-catch / interpreter | -1.60% / +0.05% | -0.80% / -0.12% |

### Numeric layout

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | +5.60% / +0.30% | +18.83% / +3.45% |
| gc/struct-allocation / interpreter | +0.29% / -3.79% | +6.24% / +3.86% |
| gc/struct-access / native | -0.03% / -0.12% | -0.24% / -0.32% |
| gc/struct-indirect-access / native | +0.22% / -0.25% | +0.05% / +0.10% |
| gc/struct-access / interpreter | -0.32% / -0.19% | +0.10% / -0.31% |
| threads/wait-mismatch / native | -0.13% / -6.24% | -0.58% / +1.42% |
| gc/i31-roundtrip / native | -0.22% / -0.22% | -0.18% / -0.13% |
| core/memory-load-store / native | +0.06% / -0.31% | +0.05% / +0.20% |
| gc/sparse-empty-allocation / native | +0.41% / -0.61% | +0.40% / +0.51% |
| gc/sparse-empty-allocation / interpreter | +1.96% / +0.33% | -1.26% / -0.47% |
| gc/empty-struct-allocation / native | +0.48% / -0.52% | +0.12% / +1.79% |
| gc/empty-struct-allocation / interpreter | +3.52% / +0.34% | +0.29% / -1.31% |
| exception/throw-catch / native | -0.05% / -0.94% | +0.39% / -0.91% |
| exception/throw-catch / interpreter | -1.67% / -0.35% | +0.78% / +1.05% |

The sparse fixture primes a 300k-node chain during the first call only, then
runs the same empty-identity/LCG kernel. The discarded chain leaves a large
handle high-water mark while warm live occupancy stays small. Its WAT, Wasm,
pinned wasm-tools producer, hashes and validation are retained. Priming stays
in firstCall; it is never subtracted from warm execution or hidden. Sparse
portable collection exposes the direct-cell regression that the numeric layout
avoids, but numeric linked-native allocation still regresses 18.83% under LLGo.

Both sources execute five affected packages and 48 exact pinned reports per
compiler, including all mandatory extensions at three tiers and forced Stack
Switching collection. Independent Go/LLGo layout probes verify direct slots
16→0 bytes and Store 608/616→592/600; numeric slots 16→8 and unchanged Store.
These are record sizes, not measured RSS or total-memory claims.

## Coalesced payload relocation

Source 4768cc9 retains the original pointer slots. Collection and capacity
maintenance still compact live payloads in allocation order and republish every
moved view before reentry. The change merges only physically adjacent survivor
ranges. A gap flushes the prior run; the final run flushes at completion. Moves
use Go copy overlap semantics. Because destination offsets never exceed source
offsets and runs are processed in order, flushing cannot overwrite a later
source. No guest pattern or type selects an alternate runtime.

Individual headers may point at their destination before that run is copied.
This happens only inside the serialized safepoint, which exposes no guest
payloads or callbacks until relocation completes. Logical bytes, free-list
generations, rollback epochs, empty-object tracking and native ABI are unchanged.

Tests exercise variable-sized overlapping runs, isolated survivors, dead
prefixes/suffixes, zero-byte objects, recycled nonsequential handles, maintenance
without tracing, repeated tracing, complete byte preservation, cleared marks
and native-view/inventory coherence. The original relocation tests remain.

## Fixed-source affected measurements

Source `4768cc9f40f555772fd3503ba07914fe7a9a4317` is compared with the
byte-identical merged PR #116 engine alias `e9b1c5114e472801bc653a3a37842178e5c2cba7`.
Go 1.27.3, LLGo `983465e779527ff496857a9a1915c1b57131091f`, LLVM 22.1.8
and LLGo `-O3 -lto=full` remain pinned. No production SIMD experiment is enabled.
All references reuse the separate corrected Darwin CLOCK_UPTIME_RAW epoch;
there is no toolchain/SDK upgrade or timing subtraction.

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | -4.50% / -0.21% | -5.91% / -0.94% |
| gc/struct-allocation / interpreter | +0.27% / -0.17% | -0.55% / -0.48% |
| gc/struct-access / native | -0.16% / -0.22% | -0.12% / -0.33% |
| gc/struct-indirect-access / native | +0.53% / +0.44% | +0.02% / -0.59% |
| gc/struct-access / interpreter | -0.24% / -0.11% | -0.47% / +0.08% |
| threads/wait-mismatch / native | -1.48% / -0.31% | +0.77% / -0.86% |
| gc/i31-roundtrip / native | +0.09% / +0.11% | +0.21% / +0.02% |
| core/memory-load-store / native | -0.24% / -0.30% | +0.09% / -0.01% |
| gc/sparse-empty-allocation / native | +0.09% / +0.03% | +0.44% / -0.26% |
| gc/sparse-empty-allocation / interpreter | +1.32% / +0.60% | +0.40% / -0.54% |
| gc/empty-struct-allocation / native | +0.13% / +0.08% | -0.24% / +0.36% |
| gc/empty-struct-allocation / interpreter | +0.63% / -0.87% | +1.52% / +0.37% |
| exception/throw-catch / native | +0.47% / -0.04% | +1.02% / +1.76% |
| exception/throw-catch / interpreter | +0.16% / -0.11% | +0.43% / +1.21% |

The long linked-native case changes Go 39.619→37.835 and LLGo 40.340→37.955
ns/object, versus Wasmtime 8.894. LLGo is 0.3% above Go and still 4.27x Wasmtime.
Linked interpretation is Go 108.313 / LLGo 126.066 versus WasmEdge 104.172.
Retain the adverse observations: LLGo wait-mismatch+0.77%, i31+0.21%,
native memory+0.09%, sparse native+0.44% and interpreter+0.40%, empty
interpretation+1.52%, native exceptions+1.02% and interpreter exceptions+0.43%.
Their A/A values remain in the table; none is relabeled a pass or replaced.

All 14affected cases retain 510observations/102alternate-input checks, plus
75 oracle checks per scope. Both compilers execute the five affected packages
and 48 exact pinned reports each. Setup failures are also retained: the initial
fresh worktree lacked corpus links (the Go semantic run failed), a subsequent
preflight rejected the untracked artifacts symlink before execution, and the
exception scope initially lacked its fixture copy (zero timing observations).
The corrected fixture paths reuse identical pinned bytes. No successful timing
case or completed semantic run was repeated unchanged.

## Independent original twenty-case portfolio

The original-count scope verifies 864 VM/reference observations, 27 ordinary-C
observations, 288 alternate checks, 12 unsupported-reference results and 297 summaries.
Three fresh interleaved blocks retain two warmups and seven calls. Counts remain
one public call, 1M ordinary-loop iterations and 10k GC/notify/wait iterations.
Every raw complete block is saved before validation; all preparation, firstCall,
warm execution and cleanup remain separate in the bundle. The host is shared;
these descriptive observations do not establish controlled performance acceptance.

Native rates, ns/source iteration (call-control is ns/public call):

| Workload | Go | LLGo | Wasmtime | Wasmer | wazero compiled | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.3415 | 0.3201 | 0.9866 | 0.8225 | 1.1492 | — |
| core/call-control | 2416 | 3500 | 917 | 2125 | 375 | — |
| core/integer-loop | 1.5764 | 1.5602 | 1.55 | 1.5615 | 1.7161 | 1.5445 |
| core/function-calls | 0.801 | 0.783 | 0.9172 | 0.7821 | 1.4536 | 0.7715 |
| core/memory-load-store | 0.4587 | 0.4355 | 0.4794 | 0.5395 | 0.8768 | 0.27 |
| memory/narrow-unaligned | 0.7068 | 0.6888 | 0.6955 | 0.7281 | 1.0745 | 0.4573 |
| memory/dependent-address | 1.8705 | 1.8425 | 2.0818 | 2.1195 | 3.1394 | 1.4168 |
| memory/float64-unaligned | 0.7368 | 0.7288 | 0.7087 | 0.7137 | 0.8776 | 0.6689 |
| runtime/simd-memory-full | 3.3937 | 3.379 | 3.6622 | 3.4776 | 3.7472 | 3.5865 |
| runtime/simd-memory-widen | 1.0599 | 1.0406 | 1.044 | 1.041 | 1.139 | 1.0393 |
| runtime/simd-memory-lanes | 5.0543 | 4.9091 | 5.818 | 5.8138 | 6.0772 | 4.4269 |
| gc/i31-roundtrip | 1.9584 | 1.8334 | 2.1 | — | — | — |
| gc/struct-access | 2.1834 | 2.1542 | 1.45 | — | — | — |
| gc/struct-allocation | 34.7042 | 36.1792 | 106.425 | — | — | — |
| gc/struct-indirect-access | 5.5833 | 5.2625 | 2.975 | — | — | — |
| threads/shared-memory | 0.8117 | 0.7767 | 0.9825 | 1.0198 | 1.1347 | — |
| threads/atomic-rmw | 1.818 | 1.8009 | 1.794 | 3.7486 | 1.9421 | — |
| threads/notify-empty | 1.4958 | 1.4916 | 6.9291 | 4.9125 | 21.3667 | — |
| threads/wait-mismatch | 2.025 | 1.5916 | 9.0083 | 2.9375 | 25.2666 | — |
| threads/wait-timeout | 2.1416 | 1.8708 | 43.4541 | 171.1125 | 245.0833 | — |

Interpreter rates retain those original counts:

| Workload | Go | LLGo | wazero interpreter | WasmEdge interpreter |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 48.1433 | 38.3108 | 62.4278 | 38.358 |
| core/call-control | 2500 | 3292 | 917 | 2375 |
| core/integer-loop | 38.6083 | 29.1373 | 81.8066 | 48.599 |
| core/function-calls | 58.4263 | 53.771 | 79.7172 | 62.1903 |
| core/memory-load-store | 82.0874 | 62.4678 | 144.2586 | 91.4646 |
| memory/narrow-unaligned | 107.1664 | 81.4219 | 185.1068 | 113.1012 |
| memory/dependent-address | 50.089 | 40.34 | 94.7752 | 53.5057 |
| memory/float64-unaligned | 83.3017 | 64.9747 | 138.6616 | 84.3065 |
| runtime/simd-memory-full | 89.3181 | 71.1681 | 151.7611 | 77.4589 |
| runtime/simd-memory-widen | 72.3986 | 62.7802 | 143.4225 | 60.8449 |
| runtime/simd-memory-lanes | 109.7468 | 91.5962 | 187.3233 | 90.6208 |
| gc/i31-roundtrip | 46.775 | 37.2458 | — | 44.8667 |
| gc/struct-access | 52.1792 | 42.7583 | — | 55.2792 |
| gc/struct-allocation | 106.4833 | 125.5417 | — | 107.7958 |
| gc/struct-indirect-access | 98.7583 | 78.7625 | — | 96.2583 |
| threads/shared-memory | 81.8386 | 37.5583 | 62.3764 | 38.4551 |
| threads/atomic-rmw | 62.147 | 37.9968 | 43.5362 | 31.8848 |
| threads/notify-empty | 41.8166 | 32.05 | 42.3542 | 40.4959 |
| threads/wait-mismatch | 70.3333 | 45.025 | 67.3792 | 43.5375 |
| threads/wait-timeout | 68.5875 | 41.4375 | 293.3041 | 97.975 |

The native C controls omit Wasm bounds, fuel, cancellation and guest GC; their
gap is separately classified. Unsupported GC references remain explicit. The
short 10k allocation portfolio differs from the long 300k retained-chain scope;
neither result substitutes for the other.

Full-portfolio source changes remain independent:

| Workload | Go native | Go interpreter / A/A | LLGo native | LLGo interpreter / A/A |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | +1.70% | +1.19% / +0.47% | -0.19% | +0.48% / -0.59% |
| core/call-control | -24.69% | +39.59% / -2.29% | +13.49% | +43.63% / +7.29% |
| core/integer-loop | -0.37% | -0.58% / +0.11% | +0.07% | -0.61% / +0.34% |
| core/function-calls | -0.10% | +1.21% / +0.64% | -0.11% | +0.70% / -1.29% |
| core/memory-load-store | +1.35% | -0.08% / -0.23% | +0.40% | +1.02% / -0.38% |
| memory/narrow-unaligned | +0.36% | +1.11% / -1.00% | -1.21% | +0.49% / +0.48% |
| memory/dependent-address | +0.69% | -0.18% / -0.10% | -0.19% | -0.02% / +0.19% |
| memory/float64-unaligned | -1.18% | +0.17% / -0.34% | -0.25% | -1.27% / -0.21% |
| runtime/simd-memory-full | +0.03% | -0.24% / -0.12% | +0.40% | -0.57% / +0.05% |
| runtime/simd-memory-widen | +0.15% | -0.81% / -0.43% | +0.10% | -2.76% / +0.03% |
| runtime/simd-memory-lanes | +1.00% | -2.52% / -1.33% | -1.02% | +0.21% / -0.11% |
| gc/i31-roundtrip | +13.80% | +1.70% / +1.50% | -0.90% | +4.43% / +0.33% |
| gc/struct-access | +7.16% | -0.83% / +4.05% | +7.71% | +0.86% / -0.98% |
| gc/struct-allocation | +2.01% | -0.94% / -2.10% | +3.21% | +2.46% / -0.99% |
| gc/struct-indirect-access | +6.43% | +3.30% / +0.55% | +0.00% | -0.32% / -3.41% |
| threads/shared-memory | +0.26% | +0.18% / -0.03% | -2.17% | -1.57% / +1.80% |
| threads/atomic-rmw | +0.52% | +3.49% / -0.18% | +0.56% | +0.83% / -0.34% |
| threads/notify-empty | -0.28% | +9.09% / +3.62% | +1.99% | -1.03% / -2.28% |
| threads/wait-mismatch | +9.96% | +5.22% / +0.14% | -3.54% | +0.36% / -2.33% |
| threads/wait-timeout | -19.56% | +2.55% / +0.01% | -2.82% | +3.01% / +0.46% |

The single-public-call controls regress LLGo native +13.49% and interpreter
+43.63% (interpreter A/A +7.29%); Go native changes −24.69% while interpreter
regresses +39.59%. These unresolved outcomes are not dismissed as noise. Short
GC native access is +7.71% under LLGo and short allocation +3.21%, despite
long-count access −0.12% and linked allocation −5.91%. Counts and cohorts differ;
the longer cases do not erase the original-count observations.

No entire-portfolio or target-platform performance acceptance is claimed.
Final-head Go/LLGo execution on both supported platforms, all 30 coverage gates,
shard-completeness and review are merge requirements. Their run links and
artifact audit are recorded in the contribution PR; local execution alone does
not replace them.

## Evidence

The [compressed per-file JSON bundle](evidence/performance-native-payload-runs-2026-11-04.json.xz)
contains raw timings, phases, sample orders, checksums, source patches, source
inventories, compiler/SDK hashes, profiles, successful execution reports and
failed setup attempts. UTF8/base64 payloads each have a decoded SHA256 hash.
The packaging helper verifies every decoded byte against the original file.
CI artifacts remain linked from the contribution PR. Earlier PR #116 timing and
clock-baseline evidence stays in its independent bundle.

Bundle: 658 files, 757680 bytes; SHA256
`dbea457769ff4de1d853e82c8757928321ea4872a6fa4ccf57103b9a1d4dbc00`.

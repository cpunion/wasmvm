# Compiled component schema comparisons

**Implemented local candidate; performance acceptance remains open.** Reusing
immutable schema traversal reduces instructions and allocations. Marginal
production LLGo throughput improves by a 6.56% point estimate; experimental
LLGo improves by only 0.92%. Neither reaches competitive parity. Full hosted
final-head validation and review are pending.

## Mechanism and checked semantics

Implementation `98ca48cb83f97dc0a7784d1c0b54aa5a0c5455f7` follows the call-veneer
candidate `b02cbe87a108e3283c4f38450fcf3d3cc989b1f6` (implementation `8a25766`).
The [comparison-plan design](../design/value-comparison-plans.md) uses the
existing equality visitor to compile structural results and the ordered nominal
resource predicates for each schema pair and exact node budget. Matching always
resolves current resource bindings. No runtime identity result is cached.

This preserves predicate order, repeated obligations before visited edges,
short-circuit mismatch and budget-error precedence. Limits produce fresh errors.
Each actor retains at most 64 programs and 4096 nominal operations; capacity
exhaustion uses the ordinary comparator. Shutdown releases type roots after
active calls and source cleanup. Both backends share the mechanism.

A separate previous-candidate counter copy inventories 1,000 checked requests:
`asyncEnd.matches` and `Instance.end` perform 47 comparisons/request in 15
call-site/pair groups, representing 12 distinct ordered type pairs. Some graphs
have 52 or 54 nodes. Calibration-only groups are excluded. The task-return check
is outside that inventory; it also uses the new mechanism. Instrumented timing
is excluded. Full Canonical ABI conversion remains separate checked work.

## Complete request comparison

M4 Max/macOS arm64, Go 1.27.3, production LLGo `3d92a58`, experimental compiler
`e2cca2c`/runtime `f2e4cb426`, independent Wasmtime 48.0.1. All LLGo builds use
`-O3 -lto=full`, identical within each source pair. The experimental toolchain
includes multiple changes and is neither a production upgrade nor an isolated
measurement of LLGo #3057.

The same 16-byte Rust P3 HTTP guest uses buffered responses, four workers,
16 clients and 256 requests/client. Status, headers, full body, trailers and
CRC32 remain checked. Three fixed rounds (seed `2026111217`) each run main,
previous candidate, head and independent main A/A processes for each compiler,
plus Wasmtime. Each process runs first execution, two warmups and 30 measured
4096-request batches. All **39 processes** and **4,792,320 measured responses**
pass. The retained production main binaries have exactly equivalent
implementation `0aaab1ad` to main `4ce75302`; executable identities are recorded.

Darwin process counters cover 29 inter-batch intervals/process and 4,632,576
checked requests, summing all threads with 125/3 ns per Mach tick. Startup and
cleanup are outside this window; readiness/bookkeeping remains included.
Throughput is each process's median batch rate, then the median of three
process observations. Instructions are hardware retired instructions.

| Configuration | req/s | CPU µs/request | Instructions/request |
|---|---:|---:|---:|
| Main / Go | 26,584 | 147.93 | 1,891,944 |
| Previous / Go | 30,074 | 132.03 | 1,688,863 |
| Head / Go | 34,089 | 115.91 | 1,347,879 |
| Main / production LLGo | 6,643 | 674.12 | 2,636,350 |
| Previous / production LLGo | 7,535 | 607.48 | 2,377,912 |
| Head / production LLGo | 8,029 | 577.71 | 2,120,778 |
| Main / experimental LLGo | 12,448 | 371.18 | 2,565,016 |
| Previous / experimental LLGo | 13,657 | 341.29 | 2,283,550 |
| Head / experimental LLGo | 13,783 | 334.71 | 2,052,209 |
| Wasmtime | 75,385 | 24.54 | 207,921 |

Versus the previous candidate, throughput points improve **13.35% / 6.56% /
0.92%** and instructions fall **20.19% / 10.81% / 10.13%** for Go/production
LLGo/experimental LLGo. CPU points fall 12.21% / 4.90% / 1.93%.
Only **6/9 timing A/A controls** resolve within 2.5%; all **9/9 instruction
controls** resolve. The experimental timing change is unresolved. Frequency and
neighbor load are uncontrolled; no sustained acceptance or controlled ranking
is established by these point estimates.

The combined branch improves throughput points **28.23% / 20.88% / 10.72%**
versus main and reduces instructions **28.76% / 19.56% / 19.99%**. These are
direct comparisons within this cohort, not compounded historical percentages.
Wasmtime remains **2.21× / 9.39× / 5.47×** faster in throughput; the experimental
LLGo instruction gap is **9.87×**. Its same-cohort previous throughput gap was
5.52×; the preceding report's 5.31× belongs to another cohort. Wasmtime's ready
`Full<Bytes>` body/generated bindings differ from the owned public
`http.Request`/`io.Reader` adapter and bounded I/O workers. No new ordinary
native-host or other-runtime comparison is claimed by this event.

## Allocations, preparation and footprint

Four separate allocation-observer processes check 20,000 requests each after
warmup. Go uses public `MemStats.Mallocs`; production LLGo sums exact native
size-class allocation counters. Every empty observer control reports zero.
Events/request fall **1059.008 -> 1011.004** for Go and **2998.035 -> 2889.049**
for LLGo. Observer timings are excluded. These are allocation events, not bytes
or collection CPU, and do not constitute competitive parity.

Separate phase medians use milliseconds. `firstCall` is the first 4096-request
batch; instantiation sums workers. Observed lifetime-peak footprint at batch
markers includes preparation and is not a final post-exit peak.

| Head configuration | Validate/compile | Instantiate | First batch | Cleanup | Observed peak MiB |
|---|---:|---:|---:|---:|---:|
| Go | 165.16 | 0.53 | 124.09 | 0.07 | 113.3 |
| Production LLGo | 243.46 | 0.90 | 546.09 | 0.59 | 150.9 |
| Experimental LLGo | 233.26 | 0.82 | 350.32 | 0.60 | 162.3 |
| Wasmtime | 52.32 | 0.18 | 58.61 | 0.23 | 21.8 |

## Residual profile and next work

One separate Time Profiler process runs the exact head experimental binary,
checks 122,880 measured responses and exits successfully. All 76,065 exported
samples are `Running`, with waiting-thread recording off. The kernel window
contains 75,259 samples and first execution/two warmups/30 measured batches,
135,168 checked requests. Normalized sample weight is **556.78 CPU µs/request**,
materially above **334.71** in the unprofiled counter cohort. Instrumentation,
sampling and execution-condition divergence prevent absolute category deltas;
profiled timing is excluded from all source throughput comparisons.

Allocation/free/accounting occupies 23.2% of samples, collector paths 10.8%,
component calls/tasks 16.1%, provider 11.2%, Core execution 10.5% and canonical
values 7.1%. Anonymous/JIT/unresolved leaves retain 11.9%. Recursive
`equality.value` traversal is absent from the warm hot frames. Cache lookup and
nominal checks still execute and may be inlined into component frames:
`Instance.matchValue` alone has 4.82 normalized sample µs/request attributed to
its nearest owned frame. The near-zero standalone type-comparison category
therefore does not mean all validation work has disappeared.

Continue with general checked Canonical ABI conversion/staging and temporary
ownership, then isolate channel/scheduling costs with representative probes.
Allocation and collector cost remain substantial, but avoiding repeated schema
traversal demonstrates a separate source of instruction work. The current
profile does not prove a new LLGo bug; existing runtime issues are not duplicated.

## Validation and integration

All 32 Go internal/pkg test groups pass. All 31 executable-package coverage
gates pass, minimum **90.86%**; validate/component reach **92.12% / 92.46%**.
Whole-module vet and no-cgo build pass; component race tests pass. Production
LLGo executes all three affected packages on macOS arm64. Local Docker Linux
amd64 executes six Go/LLGo package groups, including actual P3 Rust guest
execution on both backends. The eight LLGo shards still cover all test packages.

Differential tests cover cycles/shared graphs, exact budgets, ordered/stateful
nominal predicates, resource rebinding/missing bindings, cache/nominal limits,
fallback and shutdown root retirement. An initial limit-error classifier
assumed a concrete error type instead of the existing wrapped sentinel; the
final implementation uses `errors.Is`. The failed attempt and the diagnostic
probe's calibration contamination remain recorded; neither qualifies a result.

The candidate stays on the contribution branch while
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057) occupies the single active
PR slot. Final hosted CI/review, cross-target controlled performance and release
acceptance remain open. [Data and provenance](data/performance-value-comparison-plans-2026-11-11.json)
retain all process arrays, A/A controls and qualifications.
[Compressed evidence](evidence/performance-value-comparison-plans-2026-11-11.jsonl.gz)
retains counters, snapshots, phase timings and checked batch summaries. Full
per-request arrays remain in the local raw event with its recorded checksum.

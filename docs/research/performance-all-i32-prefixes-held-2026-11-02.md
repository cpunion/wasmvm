# Held general I32 prefix experiment: 2026-11-02

Status: **held for performance; not part of the selected implementation**.
Base `fda5e8f` is the combined coalesced-read/native-batch candidate; tested head
`87a3853` removes the following-integer-consumer requirement from I32 prefix
hints. It keeps nonempty destinations, original PCs, all instruction words,
constant commit order, metered execution and polling. The dispatcher loop is
unchanged; the compiler simply permits it to consume more I32 constants.
No workload identifiers, byte padding or safety shortcuts are used.

All 31 Go packages and 30 coverage gates pass, as do pinned LLGo slots/codegen/
execute/VM, targeted race, no-cgo and vet. New tests check metadata before
memory/call/global/atomic/GC/control instructions, no-slot/final-PC rejection,
130 adjacent constants passed to a real guest callee, bit width, precise fuel
and source-64 polling. Existing trap/visible-side-effect and wrap tests pass.
All 86 native guest bodies match, across both architectures and tiers; foreign
code is compile-only. Passing semantics does not establish performance.

## Full unchanged-work comparison

Shared M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` / `-O3 -lto=full`.
All twenty original kernels, three fresh serially interleaved process blocks,
two warmups/seven calls, both compilers/backends/base/head, exact baseline A/A
aliases for both backends, all five pinned reference modes and six applicable
unchecked C controls. Counts, quotas, bytes and oracles match the combined-base
campaign: twelve large loops one million; four GC and three notify/wait cases
ten thousand; empty call one. The complete result retains 984 VM observations,
18 C observations, 334 summaries, 328 alternate checks, twelve explicit GC
reference rejections and every separate measurement phase.

Warm ns/source iteration (empty control is ns/call):

| Workload | Go interpreter base → head | LLGo interpreter base → head | LLGo baseline A/A | WasmEdge |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 50.933708 → 46.306250 | 48.211208 → 43.861083 | 48.257625 | 38.405000 |
| core/call-control | 3083.000000 → 2917.000000 | 3208.000000 → 3084.000000 | 3292.000000 | 3000.000000 |
| core/integer-loop | 44.198083 → 44.065667 | 42.127584 → 42.240167 | 42.265458 | 48.537000 |
| core/function-calls | 67.399500 → 67.658417 | 72.248875 → 72.171208 | 72.245000 | 61.926000 |
| core/memory-load-store | 85.152250 → 85.723875 | 97.290208 → 80.801458 | 97.006542 | 90.605000 |
| memory/narrow-unaligned | 111.127375 → 110.611375 | 104.171250 → 120.384083 | 103.180208 | 112.876000 |
| memory/dependent-address | 56.499166 → 56.587917 | 54.444042 → 54.549666 | 54.474500 | 53.328000 |
| memory/float64-unaligned | 89.368791 → 89.720833 | 83.094916 → 100.500958 | 82.652292 | 84.294000 |
| runtime/simd-memory-full | 106.945167 → 105.981958 | 100.026458 → 115.818250 | 99.797500 | 77.582000 |
| runtime/simd-memory-widen | 87.136084 → 89.951583 | 90.079208 → 91.260000 | 89.798209 | 61.046000 |
| runtime/simd-memory-lanes | 131.260875 → 130.525500 | 127.582625 → 159.788625 | 127.928958 | 90.631000 |
| gc/i31-roundtrip | 53.554200 → 55.450000 | 49.041700 → 52.779200 | 50.879200 | 45.200000 |
| gc/struct-access | 58.462500 → 57.529200 | 55.754200 → 58.316700 | 56.004200 | 55.400000 |
| gc/struct-allocation | 122.991700 → 118.591700 | 160.420800 → 156.425000 | 158.887500 | 111.400000 |
| gc/struct-indirect-access | 100.045800 → 103.383400 | 97.550000 → 96.408400 | 94.691600 | 95.900000 |
| threads/shared-memory | 87.003250 → 82.639417 | 48.369208 → 44.134584 | 48.420584 | 38.368000 |
| threads/atomic-rmw | 65.226000 → 60.473250 | 63.597750 → 59.194083 | 63.312875 | 31.713000 |
| threads/notify-empty | 45.258300 → 39.975000 | 54.479200 → 48.754200 | 53.537500 | 41.500000 |
| threads/wait-mismatch | 74.995800 → 70.608300 | 71.579100 → 68.337500 | 74.012500 | 44.900000 |
| threads/wait-timeout | 72.683300 → 71.279100 | 68.658300 → 66.625000 | 69.062500 | 97.800000 |

LLGo memory improves 16.9%, RMW 6.9%, private/shared loops about 9%, and notify
10.5%. These useful cases cannot erase substantial regressions: SIMD lanes
**25.2%**, float memory **20.9%**, full SIMD **15.8%**, narrow memory **15.6%**.
Their corresponding LLGo baseline A/A values remain close to baseline.
Float paired head/base ratios are 1.2042/1.2747/1.2171 versus A/A
0.9914/0.9947/1.0170. Lane paired ratios are 1.2422/1.2584/1.2524 versus A/A
1.0008/1.0055/0.9906. The large adverse observations repeat in every block.

Go improves thirteen of twenty interpreter cases and LLGo eleven; Go widening
and i31/indirect GC regress about 3.2–3.5%. Native adverse observations remain
in the complete evidence, including Go timeout +25.2%, Go/LLGo struct access
+17.5%/+14.1%, LLGo i31 +9.3% and notify +9.2%. Native GC is Go
46.8625→45.4250 (A/A 48.0084), LLGo 55.3000→56.4500 (A/A 57.1334), Wasmtime
107.4000. A/A noise and guest-body identity do not prove native regressions
harmless or accept any narrow gate.

Keep the earlier combined source as the selected candidate. This experiment
shows that broader semantically safe fusion does not guarantee a faster whole
interpreter under the pinned compiler. The precise metadata/branch/code-layout
contribution has not been isolated; no LLGo issue is filed from throughput alone.
No favorable subset is substituted for the twenty-case result, and no fixture-
specific eligibility rule or alignment workaround is introduced. Original
compiler, ordinary-native, startup and controlled-acceptance gaps stay open.

[Complete failed-performance evidence, identities, all phases and semantic logs](evidence/performance-all-i32-prefixes-held-2026-11-02.json.gz)

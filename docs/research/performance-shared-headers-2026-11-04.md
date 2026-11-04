# Shared GC headers: 2026-11-04

The implementation unifies ordinary object metadata and native views in one
32-byte header on the supported 64-bit targets. The source baseline is merged
PR #117. Two intermediate sources remain held with their complete observations.
Performance acceptance remains open.

## Representation and lifetime

A header contains the complete handle, native/portable payload pointer, 32-bit
length and canonical type, and full 64-bit allocation serial. The 16-byte pointer
slot table remains. Its live-only spare word contains mark/optional flags; free
slots retain their complete 32-bit next index. Optional typed cells hold tags,
descriptors, continuations and their exceptional byte charges. No generation,
epoch or logical quota bits are borrowed. Logical metadata accounting stays 64
bytes per object plus the original descriptor/continuation charges.

Portable headers are typed Go allocations and retain payload pointers across
host GC. Native headers alias the owned native view itself. They contain only
owned native payload pointers; optional host pointers remain in typed Go pages.
The 32-byte native header replaces the former 64-byte common record plus 32-byte
view. This is a representation calculation, not a measured RSS saving. Optional
cells grow 32→40 bytes; page headers stay 56 bytes and slots 16. Store size changes
608→624 bytes in Go / 616→632 in LLGo because the bounded layout cache is a slice.

Native word two packs low 32-bit length and high 32-bit type. Constructors stage
instance-local types and a zero epoch; Finish validates the complete batch,
payload range, handle generations and quotas before publishing canonical types
and epochs in the same headers. Both emitters load 32-bit lengths for array.len
and bounds. Payload compaction/growth updates one shared pointer instead of
republishing a second record. Native ABI entry records remain unchanged; the
private view and generated code belong to the same runtime build.

Private native Memory.Grow previously moved allocations. The initial development
test exposed dangling header pointers, so the view now uses a bounded anonymous
PROT_NONE reservation with read/write pages enabled only as its logical extent
grows. Stable storage shares the existing Memory owner/binding/active-entry
lifetime API. Close unmaps the reservation after uses drain. Shared guest memory
behavior is unchanged. Metadata slabs are quota-truncated; growth never clears
pending constructors or moves existing headers.

Reference tracing reads exact eight-byte fields from admitted layouts and
initialized lengths. It does not reconstruct a variable byte slice per marked
object. Field get/set retains array bounds and typed layout access. Allocation
keeps metadata materialization errors on the cold path before payload/handle
commit; the common path does not return a 24-byte pointer/error tuple. Sweeps
reuse known slots, mark flags and optional cells when releasing objects. These
paths apply to every admitted type; no guest benchmark pattern selects them.

## Frozen sources and semantic checks

Baseline engine 4768cc9 is byte-identical to merged main 26ebd6c. The fixed
production profile is Go 1.27.3, LLGo 983465e779527ff496857a9a1915c1b57131091f,
LLVM 22.1.8 and LLGo -O3 -lto=full. No production SIMD experiment is enabled.
Reference SDKs, clocks and fixtures remain in the corrected-clock epoch from
PR #116. No compiler/profile upgrade is mixed into these source comparisons.

Each frozen source executes eight affected packages under both compilers and
48 exact pinned reports each, including five mandatory extensions at three tiers
and forced Stack Switching collection. Extra Go cgocheck2 and no-cgo checks
remain separately identified. These are local checks, not the final platform
matrix. The direct package coverage profile excludes cross-package execution
coverage; it does not replace the final 30 package coverage gates.

Tests cover exact header offsets/aliasing, host-GC payload liveness, unaligned
1/2/4/8/16-byte fields, stable cells across pages/relocation, invalid array indexes,
staged local/canonical types and epochs, changed reservations, quota/serial
exhaustion, nested rollback, optional cells, cycles and all typed roots. Stable
memory tests cross page boundaries, preserve/zero bytes, reject invalid growth,
exclude active private access and drain retired bindings before unmapping.

The initial development build failure after map→slice conversion and the moving
Memory.Grow pointer failure are retained in logs. Intermediate development source
snapshots were not captured; frozen measured sources, patches and inventories
are complete. No successful timing is repeated unchanged.

## Held source-changing ablations

### First shared header: bb19cec

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | -5.37% / +1.35% | +11.72% / +0.48% |
| gc/struct-allocation / interpreter | +1.38% / +0.02% | +9.92% / +2.84% |
| gc/struct-access / native | -0.08% / -0.16% | +0.03% / -0.16% |
| gc/struct-indirect-access / native | +0.16% / +0.42% | -0.14% / +0.21% |
| gc/struct-access / interpreter | +7.81% / +0.03% | +6.63% / -1.79% |
| threads/wait-mismatch / native | -3.81% / -1.84% | -0.23% / +3.03% |
| gc/i31-roundtrip / native | -0.04% / +0.02% | +0.07% / +0.16% |
| core/memory-load-store / native | +0.34% / +0.06% | -0.11% / -0.11% |

### Exact-width tracing: 59f74f6

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | -3.36% / +2.92% | -17.05% / -0.81% |
| gc/struct-allocation / interpreter | -1.69% / -0.37% | +2.01% / -1.04% |
| gc/struct-access / native | -0.06% / -0.11% | -0.19% / -0.03% |
| gc/struct-indirect-access / native | +0.42% / +0.31% | +0.80% / -0.06% |
| gc/struct-access / interpreter | -1.70% / -0.38% | +3.47% / -0.01% |
| threads/wait-mismatch / native | +0.14% / -0.50% | -3.55% / -0.28% |
| gc/i31-roundtrip / native | +0.19% / -0.13% | +0.08% / -0.16% |
| core/memory-load-store / native | -0.17% / -0.64% | +0.04% / +0.35% |
| gc/sparse-empty-allocation / native | -16.37% / -0.15% | -16.50% / -0.33% |
| gc/sparse-empty-allocation / interpreter | +1.03% / -0.10% | +4.49% / -0.19% |
| gc/empty-struct-allocation / native | -15.40% / +0.12% | -16.50% / +0.71% |
| gc/empty-struct-allocation / interpreter | +2.77% / -0.04% | +4.97% / +0.46% |
| exception/throw-catch / native | -0.39% / +0.25% | +1.67% / +0.23% |
| exception/throw-catch / interpreter | -1.92% / -0.29% | -2.61% / +0.35% |

The first source retains eight cases / 300 observations / 60 alternate checks and
75 oracle checks. Its native LLGo linked-allocation regression 11.72% (A/A +0.48%)
keeps it held. The trace follow-up retains 14 cases / 510 observations / 102 alternate
checks and75 oracle checks per scope. Its native linked gain 17.05% accompanies
LLGo interpreter empty +4.97% and sparse +4.49%, so a known-slot follow-up was
implemented before promotion. No full twenty-case portfolio is run for either held source.

CPU diagnostics use the exact PR #117 and bb19cec native LLGo binaries, 1000 calls
of the 300k linked fixture and three-second macOS sampling. Head main-thread
samples 2237 include 1464 under collection, 417 under batch finish, 172 in objectData
and 53 in release accounting. Sample counts are not normalized wall-time scores.
Disassembly shows the variable slice helper returning via an indirect aggregate;
AllocU sites are cold unsafe.Slice error paths, not hot allocation evidence.
No new LLGo bug is claimed without an independent compiler reproduction. Issue
2983 remains independently owned; no duplicate contribution was made.

## Final source and measurement scopes

Final measured engine: fe149a44aa86d175d5e35c6f15345960d2b7a6a1. All process
startup, read, setup, validation/compilation, instantiation, lookup, first-call,
warm execution and cleanup phases remain separate. Kernels include their guest
loops and required GC. A/A aliases use exact baseline binaries. No startup,
clock-pair or empty-call score is subtracted, and cohorts/counts are not pooled.

### Affected long-count scopes

Each case uses five fresh interleaved process blocks, two warmups and seven
warm calls. Eight primary cases retain 300 observations; sparse/empty/exception
scopes retain 70 each: 510 total observations and 102 alternate checks, with 75
oracle checks per scope. Smaller original-count cases below remain separate.

| Workload / backend | Go change / A/A | LLGo change / A/A |
| --- | ---: | ---: |
| gc/struct-allocation / native | -9.13% / +0.25% | -19.47% / -1.23% |
| gc/struct-allocation / interpreter | -5.52% / -1.75% | -2.31% / -0.51% |
| gc/struct-access / native | -0.11% / -0.03% | -0.01% / -0.02% |
| gc/struct-indirect-access / native | +0.14% / +0.16% | -0.61% / -0.44% |
| gc/struct-access / interpreter | -2.87% / -0.60% | +3.79% / +0.21% |
| threads/wait-mismatch / native | +0.21% / +0.42% | -3.79% / +0.70% |
| gc/i31-roundtrip / native | -0.01% / -0.02% | -0.46% / -0.14% |
| core/memory-load-store / native | +0.37% / +0.49% | -0.41% / -0.65% |
| gc/sparse-empty-allocation / native | -22.86% / -0.48% | -22.77% / +0.42% |
| gc/sparse-empty-allocation / interpreter | -4.05% / -0.43% | +5.02% / +0.60% |
| gc/empty-struct-allocation / native | -23.67% / -0.43% | -23.16% / -0.12% |
| gc/empty-struct-allocation / interpreter | -2.44% / +0.86% | +4.21% / +0.78% |
| exception/throw-catch / native | -0.66% / -0.55% | +2.78% / -0.13% |
| exception/throw-catch / interpreter | -2.93% / -1.30% | -1.66% / +2.60% |

Long linked-native allocation changes Go 38.211→34.723 and LLGo
38.714→31.177 ns/object, versus Wasmtime 9.012. LLGo remains
3.46x Wasmtime. These rates include LCG work, object construction and required GC.
Retain LLGo interpreter access +3.79%, sparse +5.02%, empty +4.21% and native
exceptions +2.78%. A/A values remain explicit; no observation is relabeled a pass.

Sparse priming builds a 300k-node chain only during firstCall, then discards
its head before the empty-identity kernel. Warm occupancy is small while handle
high-water capacity stays large. Priming remains in firstCall and is never
subtracted. Fixture bytes and producer/validation hashes match the PR #117 scope.

### Independent original twenty-case portfolio

This scope verifies 864 VM/reference observations, 27 ordinary-native C records,
288 alternate checks, 12 unsupported-reference results and 297 summaries. Three
fresh interleaved process blocks retain the original one-call control, 1M ordinary
iterations and 10k GC/notify/wait iterations, with two warmups and seven calls.
Raw complete blocks are saved before validation. This shared-host comparison
does not accept the controlled performance gate.

Native rates, ns/source iteration (call-control is ns/public call):

| Workload | Go | LLGo | Wasmtime | Wasmer | wazero compiled | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.3279 | 0.3232 | 0.9242 | 0.9533 | 1.1427 | — |
| core/call-control | 3333 | 3458 | 875 | 2000 | 333 | — |
| core/integer-loop | 1.5562 | 1.5329 | 1.5203 | 1.434 | 1.6506 | 1.4264 |
| core/function-calls | 0.7973 | 0.7788 | 0.9028 | 0.7761 | 1.4249 | 0.7596 |
| core/memory-load-store | 0.4507 | 0.4303 | 0.472 | 0.5282 | 0.8818 | 0.2687 |
| memory/narrow-unaligned | 0.7019 | 0.6921 | 0.6935 | 0.7176 | 1.0501 | 0.4541 |
| memory/dependent-address | 1.8486 | 1.8308 | 2.0674 | 2.0983 | 3.1504 | 1.4126 |
| memory/float64-unaligned | 0.7475 | 0.7261 | 0.7049 | 0.7151 | 0.8698 | 0.6611 |
| runtime/simd-memory-full | 3.3755 | 3.3573 | 3.6219 | 3.4658 | 3.7892 | 3.373 |
| runtime/simd-memory-widen | 1.0663 | 1.0427 | 1.0465 | 1.0392 | 1.1542 | 1.029 |
| runtime/simd-memory-lanes | 5.0223 | 4.9461 | 5.7789 | 5.7779 | 6.0513 | 4.4218 |
| gc/i31-roundtrip | 1.8625 | 1.95 | 2.1541 | — | — | — |
| gc/struct-access | 2.1875 | 2.025 | 1.4666 | — | — | — |
| gc/struct-allocation | 31.5416 | 26.3709 | 106.0542 | — | — | — |
| gc/struct-indirect-access | 5.4666 | 5.3458 | 3.2042 | — | — | — |
| threads/shared-memory | 0.7995 | 0.79 | 1.0204 | 0.9811 | 1.1591 | — |
| threads/atomic-rmw | 1.8119 | 1.8075 | 1.8088 | 3.8285 | 1.9571 | — |
| threads/notify-empty | 1.4209 | 1.4625 | 7.0833 | 5.0458 | 20.825 | — |
| threads/wait-mismatch | 1.75 | 1.7042 | 9.1625 | 2.8917 | 25.0208 | — |
| threads/wait-timeout | 2.5208 | 1.9417 | 43.4333 | 171.0708 | 243.675 | — |

Interpreter rates retain those same original counts:

| Workload | Go | LLGo | wazero interpreter | WasmEdge interpreter |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 47.8027 | 38.9983 | 60.9097 | 38.3442 |
| core/call-control | 2250 | 2292 | 958 | 2166 |
| core/integer-loop | 38.674 | 30.2621 | 81.4489 | 48.5939 |
| core/function-calls | 57.324 | 54.57 | 82.5012 | 62.0346 |
| core/memory-load-store | 81.1536 | 65.037 | 145.3808 | 90.4984 |
| memory/narrow-unaligned | 105.1379 | 88.1624 | 183.0757 | 112.8146 |
| memory/dependent-address | 50.548 | 41.7477 | 94.1935 | 53.3777 |
| memory/float64-unaligned | 84.6887 | 68.3794 | 139.034 | 84.1265 |
| runtime/simd-memory-full | 89.8766 | 73.8488 | 151.0455 | 77.3271 |
| runtime/simd-memory-widen | 72.1297 | 64.4003 | 143.7568 | 60.6078 |
| runtime/simd-memory-lanes | 109.484 | 92.2916 | 187.8642 | 90.5179 |
| gc/i31-roundtrip | 48.7208 | 37.575 | — | 44.8292 |
| gc/struct-access | 51.8417 | 44.4125 | — | 55.6875 |
| gc/struct-allocation | 103.1791 | 123.8333 | — | 105.3542 |
| gc/struct-indirect-access | 98.5916 | 78.6708 | — | 96.3667 |
| threads/shared-memory | 84.2432 | 38.3764 | 61.5963 | 38.5384 |
| threads/atomic-rmw | 60.9077 | 39.0603 | 43.4472 | 31.8629 |
| threads/notify-empty | 41.6 | 34.2458 | 42.5542 | 39.7958 |
| threads/wait-mismatch | 70.9458 | 46.7625 | 68.2417 | 44.5833 |
| threads/wait-timeout | 68.4083 | 42.3292 | 288.8375 | 97 |

The unchecked C controls omit Wasm bounds, fuel, cancellation and guest GC.
They are separately classified. Unsupported GC references remain explicit; the
pinned Wasmtime uses inferred DRC selection and different cycle policy, as
documented in the [PR #117 study](performance-native-payload-runs-2026-11-04.md).
Those policies explain context and do not waive measured gaps.

Full-portfolio source changes remain independent:

| Workload | Go native | Go interpreter / A/A | LLGo native | LLGo interpreter / A/A |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | -1.39% | +2.28% / +2.00% | +2.95% | +2.33% / +1.58% |
| core/call-control | +77.76% | -10.00% / +3.32% | +10.66% | -8.32% / -11.68% |
| core/integer-loop | +1.99% | +1.17% / +0.10% | +4.26% | +4.10% / +0.49% |
| core/function-calls | -0.13% | -2.32% / +0.50% | -0.54% | +2.67% / -0.51% |
| core/memory-load-store | -1.13% | -1.18% / -0.52% | -0.91% | +4.32% / +0.07% |
| memory/narrow-unaligned | -0.40% | -0.78% / -0.86% | +0.37% | +8.44% / -0.15% |
| memory/dependent-address | -0.16% | +0.98% / +0.18% | -0.09% | +4.30% / -0.29% |
| memory/float64-unaligned | +1.59% | +2.85% / +0.65% | +0.20% | +4.87% / -0.98% |
| runtime/simd-memory-full | -0.93% | -0.77% / -0.63% | +0.02% | +2.62% / -0.15% |
| runtime/simd-memory-widen | +0.36% | +0.07% / -0.65% | -0.37% | +3.17% / +0.75% |
| runtime/simd-memory-lanes | -0.64% | -0.80% / -0.95% | -0.87% | +0.30% / -1.29% |
| gc/i31-roundtrip | -0.45% | +3.81% / +0.92% | +2.86% | +5.24% / +2.98% |
| gc/struct-access | +3.14% | -4.53% / -2.15% | -1.62% | +5.20% / +0.00% |
| gc/struct-allocation | -8.45% | -0.77% / +4.10% | -23.69% | +0.63% / +0.92% |
| gc/struct-indirect-access | -2.38% | -1.52% / -1.54% | +6.92% | +2.81% / +2.62% |
| threads/shared-memory | -0.54% | +2.79% / +0.15% | -0.39% | +1.87% / -0.49% |
| threads/atomic-rmw | -0.38% | -2.17% / -0.89% | +0.88% | +2.43% / -0.21% |
| threads/notify-empty | +5.25% | +0.11% / +0.91% | +5.09% | +4.00% / +0.77% |
| threads/wait-mismatch | +6.60% | +0.47% / -0.67% | +5.69% | +4.87% / +1.24% |
| threads/wait-timeout | +10.00% | +0.11% / -0.91% | +8.63% | +7.88% / +6.17% |

Counts and retention shapes differ from the long scopes; neither cohort
substitutes for the other. Every regression and its A/A control remains above.
There is no whole-portfolio or target-platform performance acceptance claim.

Unresolved original-count regressions include LLGo interpreter narrow memory
+8.44% (A/A −0.15%), native public call +10.66%, and native indirect GC access
+6.92%. Go native public call is +77.76%. These remain observed regressions;
long-count gains do not erase them or establish a whole-portfolio win.

Final-head Go/LLGo execution on both supported platforms, all 30 per-package
coverage gates, shard completeness and review remain merge requirements. CI
run links and the source/artifact audit belong in the contribution PR. Local
eight-package execution does not substitute for that full target matrix.

## Evidence

The [compressed per-file JSON bundle](evidence/performance-shared-headers-2026-11-04.json.xz)
retains the three frozen sources/patches, all timing/phase/sample-order records,
compiler and SDK identities, semantic reports, pointer/portable checks, probes,
CPU diagnostics and failed working-tree checks. Each UTF8/base64 entry has a
decoded SHA256 hash; the packaging helper round-trips every byte. CI artifacts
are linked separately from the contribution PR. Earlier PR #117 and clock-baseline
bundles remain independent.

Bundle: 598 files, 728612 bytes; SHA256
`496a1ebd0be6528775ec6bb55b70ad0343ac66c36a1f8e00aea02e366a30a5b9`.

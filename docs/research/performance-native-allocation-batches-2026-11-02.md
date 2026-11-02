# Bounded native allocation batches: 2026-11-02

Status: **measured follow-up candidate, awaiting integration and final CI**.
PR #101 remains the sole active contribution; these changes stay on a follow-up
branch while its final checks run. Controlled cross-platform performance and
the complete WASI acceptance boundary remain open.

The private numeric reservation view now holds at most 256 handles instead of
64, still capped by actual object/byte/serial quotas. All generated constructor
guards, original polls, exact exits, roots, generations and metered behavior stay
unchanged. Publication and rollback run after every native exit. Header size
grows 576→2112 bytes; active validation scratch arrays grow 768→3072 bytes.
Both remain fixed bounds. No workload names, special lifetimes or second
executor are introduced.

The [idle-cleanup change](performance-native-idle-cleanup-2026-11-02.md) is held
as a standalone throughput claim; this later combined comparison includes it.
Baseline `25298a6` is the eight-copy PR #101 engine (production code equals
`dc580d6`). Middle `c4fef4f` only separates idle cleanup. Head `d11bd32` also
increases the batch. Measured production binaries keep those actual revisions;
later changes add public boundary tests/docs only. Exact source/compiler/binary
identities are retained. All **86 guest bodies remain byte-identical** across
both architectures and tiers; foreign bodies are compile-only.

## Complete scoped screening

Go 1.27.3; pinned LLGo `8e45ed5` / `-O3 -lto=full`; macOS arm64 / Apple M4 Max.
Three fresh serially interleaved process blocks, two warmups/seven calls, both
host compilers, exact baseline A/A aliases and applicable pinned competitors.
The linked allocation chain keeps 10,000 iterations and unchanged quotas;
nonallocating GC controls use one million, calls five million, memory 50 million,
empty call one. Ordinary Clang-22 O3 C applies to the two scalar controls, with
its missing VM contracts explicitly retained. There is no invented C GC.

The 64→256 screening completes 165 VM/6 C observations, 55 alternate checks,
57 summary rows and 8 explicit GC-unsupported outcomes. Warm ns/iteration
(empty control is ns/call):

| Workload | Go base → head | LLGo base → head | Wasmtime | C |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 8583.000000 → 11541.000000 | 13500.000000 → 11250.000000 | 2000.000000 | — |
| core/function-calls | 0.918492 → 0.916758 | 0.914367 → 0.917150 | 1.062600 | 0.914000 |
| core/memory-load-store | 0.702729 → 0.700006 | 0.672112 → 0.694047 | 0.788040 | 0.445400 |
| gc/i31-roundtrip | 4.324083 → 4.254167 | 4.222375 → 4.360333 | 3.091000 | — |
| gc/struct-access | 2.252583 → 2.247167 | 2.280375 → 2.391208 | 1.754000 | — |
| gc/struct-allocation | 113.195900 → 90.570800 | 139.912500 → 136.245800 | 180.500000 | — |
| gc/struct-indirect-access | 7.717459 → 7.887792 | 7.684667 → 7.731875 | 4.354000 | — |

These three-block data are inconclusive: Go allocation improves about 20% but
its A/A also improves about 14%; LLGo improves 2.6% with a 1.3% A/A change.
Adverse LLGo memory/i31/struct-access results are +3.3%/+3.3%/+4.9%; Go
indirect access +2.2%. All phases and unfavorable controls remain retained.
They are not pooled with the separate comparison below.

## One predeclared twenty-block combination check

A single additional campaign tests only the affected original 10,000-object
chain, with 20 fresh process blocks, seven interleaved calls and two warmups.
There is no optional extension or repeat-until-pass. All three source roles,
original baseline A/A aliases and Wasmtime run in every block: 180 observations,
9 summaries and 9 alternate checks. Compiler/profile, guest bytes, quotas and
oracles remain identical. Preparation and warm work stay separate:

| Configuration | Warm ns/iteration | Readiness, ms | Compile/validate, µs | First call, ms | Cleanup, µs |
| --- | ---: | ---: | ---: | ---: | ---: |
| base-go-native | 47.500000 | 4.244 | 202.500 | 0.862 | 134.542 |
| base-llgo-native | 63.831250 | 93.502 | 248.542 | 1.413 | 161.333 |
| head-go-native | 46.393750 | 4.338 | 200.855 | 0.870 | 136.271 |
| head-llgo-native | 56.883300 | 93.129 | 229.771 | 1.031 | 160.750 |
| wasmtime-cranelift | 107.000000 | 4.605 | 1066.500 | 0.138 | 119.500 |
| middle-go-native | 49.491650 | 4.398 | 198.521 | 0.891 | 135.938 |
| middle-llgo-native | 62.562450 | 92.527 | 218.917 | 1.074 | 160.292 |
| aa-go-native | 48.535450 | 4.204 | 202.520 | 0.922 | 135.958 |
| aa-llgo-native | 63.970850 | 93.055 | 229.812 | 1.313 | 158.500 |

The combined LLGo median improves **63.831→56.883 ns/iteration (10.9%)**; all
20 paired head/base blocks improve (ratios 0.8252–0.9667). Go improves
47.500→46.394 (2.3%), with paired ratios 0.8670–1.1583. Head LLGo remains
**1.226x Go**; it takes about 53.2% of this configured Wasmtime's warm time.
This is one allocation workload, not general compiler or competitor parity.

Post-hoc diagnostic paired-block bootstrap intervals (20,000 resamples, seed
2026110251, Bonferroni across six comparisons) are retained, not used as a new
predeclared acceptance gate. LLGo head/base geometric ratio is 0.8833 with
interval 0.8605–0.9084; Go 0.9742 with 0.9368–1.0148. A/A intervals remain
wide: LLGo 0.9797–1.0683, Go 0.9831–1.0479. This fails narrow calibration and
does not establish controlled acceptance. LLGo readiness remains about 93 ms,
far above its 10 ms target. No startup win or allocation-free VM claim is made.

## Independent mechanism counters

An isolated diagnostic test executes the unchanged guest under both compilers,
covering cold and two explicitly collected warm calls. On the third call:

| Engine source | Published native batches | Native objects | Checked fallback objects | Go host allocation | LLGo host allocation |
| --- | ---: | ---: | ---: | ---: | ---: |
| Original 64 | 154 | 9847 | 153 | 808 B | 162032 B |
| Idle cleanup, 64 | 154 | 9847 | 153 | 808 B | 153712 B |
| Idle cleanup, 256 | 39 | 9962 | 38 | 808 B | 38032 B |

These exact batch/object counters corroborate reduced host publication and
fallback frequency. TotalAlloc is approximate host-collector accounting; test
executables are not production timing evidence. The initial diagnostic wrongly
expected collection to drop the guest's global sink. The unchanged fixture
intentionally retains its latest 10,000-node chain; the corrected assertion
requires that chain to survive and the replaced chain to be reclaimed. Both
diagnostic sources/failure and all final results are retained. No VM fix or
LLGo issue is inferred from that harness error. Residual LLGo allocation remains
an investigation alongside the independently reported compiler/runtime issues.

## Semantic validation

All 31 Go packages and 30 coverage gates pass, pinned LLGo state/nativeexec/VM
execute successfully, and affected race/no-cgo/vet pass. Tests cover full/partial
batches, quota-limited reservations at 1/63/64/65/255/256/257, publication,
reuse, invalid views, root chains, traps and cancellation with 255 constructors.
Public guest tests additionally cross 255/256/257 under both backends, small and
large quotas, and forced collection; Go and pinned LLGo pass. These additional
tests change no measured production source. Final contribution-head
Linux/macOS Go/LLGo CI and controlled acceptance remain required.

[All scoped/combined samples, counters, source identities, phases and failures](evidence/performance-native-allocation-batches-2026-11-02.json.gz)

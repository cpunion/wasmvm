# Combined dispatcher and native allocation candidate: 2026-11-02

Status: **unmerged full integration candidate; native regressions remain open**.
Baseline `25298a6` is the current PR #101 engine, equal in production code to
`dc580d6`. Head `fda5e8f` combines the restored `d8a3569` coalesced-read prefix
implementation with the independently measured idle-cleanup/256-handle batch
combination. No rejected guard variant or fixture-specific optimization enters
this source. Earlier component measurements are not relabelled as this result.

The same prefix loop preserves every source PC, exact metered stepping, constant
commit order and 64-step polling/wrap. Native reservation/publication preserves
quotas, byte budgets, references and all exits. The 86 captured guest bodies
match byte-for-byte on both architectures and tiers; foreign code is compile-only.
Host dispatch/cleanup performance can still change and is measured below.

All 31 Go packages and 30 coverage gates pass; pinned LLGo slots/codegen/execute/
state/nativeexec/VM execute. Affected race and targeted prefix race, no-cgo and
vet pass. This is local validation, not final contribution-head Linux/macOS CI.
The sole active PR remains #101.

## Complete matching-work comparison

Go 1.27.3 and LLGo `8e45ed5` with `-O3 -lto=full`, shared M4 Max/macOS arm64.
All 20 original kernels use unchanged guest bytes/counts: twelve large loops
one million, four GC and three notify/wait cases ten thousand, empty call one.
Three fresh serially interleaved process blocks, two warmups/seven calls, five
pinned reference modes, both own compilers/backends/base/head and exact baseline
A/A aliases for **both** backends. Six applicable unchecked ordinary C controls
remain distinct from VM policy. No C GC implementation is substituted.

All 984 VM observations, 18 C observations, 334 summaries, 328 alternate checks
and 12 explicit unsupported GC configurations complete. Every process readiness,
preparation, first-call, warm-call, cleanup and protocol sample remains in the
compressed evidence. No cross-campaign or cross-host pooling is used.

Warm ns/source iteration; empty control is ns/call:

| Workload | Go interpreter base → head | LLGo interpreter base → head | WasmEdge | LLGo native base → head | Wasmtime native | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 77.179417 → 75.521958 | 70.862333 → 66.689792 | 56.470000 | 0.437625 → 0.454750 | 1.129000 | — |
| core/call-control | 10333.000000 → 10875.000000 | 8750.000000 → 9417.000000 | 8000.000000 | 12958.000000 → 15209.000000 | 4000.000000 | — |
| core/integer-loop | 115.436459 → 81.365792 | 79.180417 → 67.056666 | 77.268000 | 2.345000 → 2.198875 | 2.333000 | 2.154000 |
| core/function-calls | 137.355667 → 113.134917 | 162.469583 → 118.284416 | 98.567000 | 1.196083 → 1.211750 | 1.413000 | 1.149000 |
| core/memory-load-store | 138.478958 → 104.392083 | 144.313417 → 116.411416 | 109.865000 | 0.530250 → 0.521667 | 0.581000 | 0.360000 |
| memory/narrow-unaligned | 161.604084 → 130.272083 | 143.404750 → 118.633792 | 133.204000 | 0.810584 → 0.804459 | 0.790000 | 0.517000 |
| memory/dependent-address | 74.782375 → 65.224250 | 70.137334 → 62.215041 | 60.677000 | 2.084208 → 2.082083 | 2.333000 | 1.592000 |
| memory/float64-unaligned | 119.289667 → 98.028667 | 105.228500 → 90.011750 | 90.920000 | 0.785667 → 0.836708 | 0.760000 | 0.708000 |
| runtime/simd-memory-full | 135.134625 → 112.692125 | 125.079500 → 103.288375 | 81.977000 | 3.617083 → 3.486709 | 3.875000 | — |
| runtime/simd-memory-widen | 101.545667 → 88.163125 | 101.045750 → 91.089916 | 61.307000 | 1.046500 → 1.044959 | 1.040000 | — |
| runtime/simd-memory-lanes | 162.191250 → 130.294833 | 151.740459 → 127.285250 | 89.958000 | 5.073083 → 5.058000 | 5.839000 | — |
| gc/i31-roundtrip | 62.029200 → 53.445900 | 54.904100 → 50.487500 | 45.400000 | 3.220800 → 3.658300 | 2.100000 | — |
| gc/struct-access | 75.791700 → 62.070900 | 67.925000 → 56.987500 | 55.400000 | 2.466700 → 2.875000 | 1.500000 | — |
| gc/struct-allocation | 133.200000 → 123.129100 | 168.700000 → 154.287500 | 107.100000 | 62.604200 → 54.745800 | 106.000000 | — |
| gc/struct-indirect-access | 120.666600 → 102.016700 | 108.512500 → 96.600000 | 97.100000 | 5.791700 → 6.095800 | 3.100000 | — |
| threads/shared-memory | 89.364875 → 85.860583 | 51.797125 → 48.521709 | 38.124000 | 0.842667 → 0.861458 | 1.024000 | — |
| threads/atomic-rmw | 68.413375 → 64.369666 | 66.806667 → 63.640084 | 31.753000 | 1.821625 → 1.827750 | 1.817000 | — |
| threads/notify-empty | 49.537500 → 46.108400 | 55.937500 → 52.379100 | 40.900000 | 1.879100 → 2.195900 | 7.200000 | — |
| threads/wait-mismatch | 82.187500 → 76.283400 | 76.262500 → 72.854200 | 43.700000 | 2.133300 → 2.291700 | 9.400000 | — |
| threads/wait-timeout | 82.350000 → 73.683400 | 75.870800 → 68.554200 | 97.200000 | 2.258300 → 2.350000 | 43.200000 | — |

Both host compilers improve all nineteen nonempty interpreter cases in this
shared-host comparison. LLGo integer/calls/memory improve 15.3%/27.2%/19.3%.
Integer leads this WasmEdge configuration, while memory remains 6.0% behind,
calls 20.0%, SIMD widening 48.6% and RMW about 100.4%. Compiler/runtime and
C deficits are distinct. Native LLGo GC improves 62.604→54.746 (12.6%), still
about 1.16x Go in this campaign. These are useful combined observations, not
controlled or aggregate parity.

## Adverse observations and limits

Every adverse case stays visible. Native Go/LLGo GC struct access regress
15.7%/16.6%, i31 about 13.6% each, and LLGo notify 16.9%. Their corresponding
A/A changes are -6.6%/+13.8%, +7.6%/+14.9%, and +4.4%. This does not establish
which changes are intrinsic or harmless. Native Go private memory regresses
21.3% while its A/A regresses 29.5%; Go native calls +9.1%, scalar memory +4.1%.
Empty-call controls regress Go/LLGo interpreter 5.2%/7.6%, native 28.5%/17.4%,
with large per-block variation. They do not satisfy the latency target.

Reservation of up to 256 handles for a function that only constructs once is
a possible fixed-cost contributor to short GC access cases. It is a source
inspection hypothesis, not a measured causal attribution. The complete original
workload and its adverse result are not replaced by a larger count. Future
batch selection should reuse compiler flow analysis and retain the same safety
contracts. Body identity alone cannot dismiss these host-boundary differences.

LLGo native scalar memory is 0.521667
versus C 0.360000 (about 1.45x), while
Wasmtime is 0.581000. Readiness and first-call costs remain separate; high
LLGo startup latency is not closed by interpreter throughput. Existing LLGo
issues stay independent of these ratios; no new compiler defect is filed from
timing alone. A/A is too noisy for the declared controlled acceptance gate.

A separate next candidate generalizes I32 constant prefixes beyond the earlier
following-integer-consumer hint. It keeps this exact source as its baseline so
its metadata effect is isolated. This report is not rewritten around that
candidate. Integration, remaining native regressions, controlled full-portfolio
performance and expanded WASI acceptance stay open.

[All phases, source/binary identities, A/A controls and validation](evidence/performance-core-dispatch-allocation-2026-11-02.json.gz)

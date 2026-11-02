# Packed vector loads: 2026-11-02

Two scoped comparisons retain the complete outcomes of an unsuccessful inline
candidate and its outlined follow-up. Both replace per-lane splat construction
with word replication. A widening load reads its already checked eight bytes
once and spreads input lanes into the two result words, then extends signs in
the unused upper halves. This applies to private memory and portable shared
memory under its existing domain lock. Concurrent native shared widening loads
retain their original per-lane accesses. Splat still performs its original one
checked scalar read, then replicates only the resulting value.

No range, offset-overflow, memory64 normalization, shared-access or trap check
is removed. Neither instruction dispatch/fuel nor native compiler lowering
changes. The helper uses portable integer operations, with no architecture-
specific assembly or second execution path for whole guest functions.

## Candidate decision and exact sources

Base `4d91cce174d3341fad34e1111cfa19a2ff0722b6` is the previously measured
integer-pair engine, with corrected I32 result bits. The first candidate
`aaef9c1` improves widening but regresses LLGo scalar memory 16.3%, full SIMD
memory 12.1% and lane memory 19.7%; it is held. Its native widening observation
also regresses 7.5%. All observations remain retained.

The follow-up `2221a919a46ce23527a5e446e4c8f069546bf997` adds only a noinline
boundary to the widening helper, keeping its extra operations out of the
common loader. No instruction or byte padding is added to tune placement.
The affected source implementations and complete linked-symbol/dispatch
assembly diagnostics are retained. Equal interpreter function sizes do not
establish equal timing; relocation/global-layout changes are not a new LLGo
compiler defect without an independent reproduction.

In this second campaign LLGo widening improves 120.236→100.346 ns/iteration
(16.5%); Go improves 120.094→100.495 (16.3%). LLGo scalar memory is approximately
flat, 100.331→99.629, while full/lane SIMD are 119.492→117.529 and
152.166→150.680. Native LLGo observations are 0.1–0.8% slower, retained below.
This supports advancing the outlined candidate, not a controlled parity claim.
Widening still takes 1.65x WasmEdge's 60.690; full/lane take 1.52x/1.67x.
It leads wazero interpreter on these cases, while both named references remain
separate rows. No synthetic fastest-per-case engine is constructed.

Each campaign uses Apple M4 Max/macOS arm64, Go 1.27.3 and pinned LLGo
`8e45ed5` `-O3 -lto=full`, the same three SIMD-memory fixtures plus a scalar
memory control, one million source iterations, three fresh serially interleaved
process blocks, two warmups and seven measured calls. Independent original and
alternate-input checks, exact baseline Go/LLGo interpreter A/A aliases and all
five pinned reference configurations pass. The scalar memory control also
interleaves the ordinary unchecked C reference; no C implementation is invented
for the SIMD cases. Each campaign has 61 complete summaries. Preparation,
instantiation, lookup, first/warm calls, readiness and cleanup remain separate.
Shared-host load is retained, without affinity, isolation or timing subtraction.

## All outcomes

Units: ns/source iteration. Columns retain base/head separately.

### Inline candidate (held)

| Workload | Configuration | Base | Head / reference |
| --- | --- | ---: | ---: |
| core/memory-load-store | go interpreter | 112.350 | 112.327 |
| core/memory-load-store | go native | 0.489 | 0.496 |
| core/memory-load-store | llgo interpreter | 100.596 | 117.010 |
| core/memory-load-store | llgo native | 0.476 | 0.473 |
| core/memory-load-store | aa-go-interpreter | — | 114.084 |
| core/memory-load-store | aa-llgo-interpreter | — | 100.377 |
| core/memory-load-store | wasmtime-cranelift | — | 0.479 |
| core/memory-load-store | wasmer-cranelift | — | 0.536 |
| core/memory-load-store | wazero-compiler | — | 0.875 |
| core/memory-load-store | wazero-interpreter | — | 142.926 |
| core/memory-load-store | wasmedge-interpreter | — | 91.003 |
| core/memory-load-store | native-c | — | 0.269 |
| runtime/simd-memory-full | go interpreter | 129.926 | 128.694 |
| runtime/simd-memory-full | go native | 3.379 | 3.389 |
| runtime/simd-memory-full | llgo interpreter | 119.307 | 133.725 |
| runtime/simd-memory-full | llgo native | 3.364 | 3.365 |
| runtime/simd-memory-full | aa-go-interpreter | — | 130.524 |
| runtime/simd-memory-full | aa-llgo-interpreter | — | 118.537 |
| runtime/simd-memory-full | wasmtime-cranelift | — | 3.605 |
| runtime/simd-memory-full | wasmer-cranelift | — | 3.421 |
| runtime/simd-memory-full | wazero-compiler | — | 4.010 |
| runtime/simd-memory-full | wazero-interpreter | — | 150.338 |
| runtime/simd-memory-full | wasmedge-interpreter | — | 77.129 |
| runtime/simd-memory-lanes | go interpreter | 179.930 | 176.739 |
| runtime/simd-memory-lanes | go native | 5.378 | 5.477 |
| runtime/simd-memory-lanes | llgo interpreter | 165.864 | 198.617 |
| runtime/simd-memory-lanes | llgo native | 5.465 | 5.449 |
| runtime/simd-memory-lanes | aa-go-interpreter | — | 177.159 |
| runtime/simd-memory-lanes | aa-llgo-interpreter | — | 166.490 |
| runtime/simd-memory-lanes | wasmtime-cranelift | — | 6.286 |
| runtime/simd-memory-lanes | wasmer-cranelift | — | 6.283 |
| runtime/simd-memory-lanes | wazero-compiler | — | 6.530 |
| runtime/simd-memory-lanes | wazero-interpreter | — | 199.371 |
| runtime/simd-memory-lanes | wasmedge-interpreter | — | 97.766 |
| runtime/simd-memory-widen | go interpreter | 132.564 | 109.226 |
| runtime/simd-memory-widen | go native | 1.184 | 1.164 |
| runtime/simd-memory-widen | llgo interpreter | 134.246 | 112.368 |
| runtime/simd-memory-widen | llgo native | 1.141 | 1.227 |
| runtime/simd-memory-widen | aa-go-interpreter | — | 133.123 |
| runtime/simd-memory-widen | aa-llgo-interpreter | — | 135.837 |
| runtime/simd-memory-widen | wasmtime-cranelift | — | 1.160 |
| runtime/simd-memory-widen | wasmer-cranelift | — | 1.163 |
| runtime/simd-memory-widen | wazero-compiler | — | 1.319 |
| runtime/simd-memory-widen | wazero-interpreter | — | 156.784 |
| runtime/simd-memory-widen | wasmedge-interpreter | — | 68.865 |

### Outlined follow-up (pending)

| Workload | Configuration | Base | Head / reference |
| --- | --- | ---: | ---: |
| core/memory-load-store | go interpreter | 112.193 | 111.866 |
| core/memory-load-store | go native | 0.489 | 0.483 |
| core/memory-load-store | llgo interpreter | 100.331 | 99.629 |
| core/memory-load-store | llgo native | 0.465 | 0.468 |
| core/memory-load-store | aa-go-interpreter | — | 113.021 |
| core/memory-load-store | aa-llgo-interpreter | — | 100.067 |
| core/memory-load-store | wasmtime-cranelift | — | 0.473 |
| core/memory-load-store | wasmer-cranelift | — | 0.528 |
| core/memory-load-store | wazero-compiler | — | 0.878 |
| core/memory-load-store | wazero-interpreter | — | 142.824 |
| core/memory-load-store | wasmedge-interpreter | — | 90.560 |
| core/memory-load-store | native-c | — | 0.264 |
| runtime/simd-memory-full | go interpreter | 129.774 | 128.541 |
| runtime/simd-memory-full | go native | 3.382 | 3.377 |
| runtime/simd-memory-full | llgo interpreter | 119.492 | 117.529 |
| runtime/simd-memory-full | llgo native | 3.361 | 3.365 |
| runtime/simd-memory-full | aa-go-interpreter | — | 130.562 |
| runtime/simd-memory-full | aa-llgo-interpreter | — | 120.292 |
| runtime/simd-memory-full | wasmtime-cranelift | — | 3.608 |
| runtime/simd-memory-full | wasmer-cranelift | — | 3.444 |
| runtime/simd-memory-full | wazero-compiler | — | 3.814 |
| runtime/simd-memory-full | wazero-interpreter | — | 148.876 |
| runtime/simd-memory-full | wasmedge-interpreter | — | 77.342 |
| runtime/simd-memory-lanes | go interpreter | 163.187 | 161.017 |
| runtime/simd-memory-lanes | go native | 5.064 | 5.024 |
| runtime/simd-memory-lanes | llgo interpreter | 152.166 | 150.680 |
| runtime/simd-memory-lanes | llgo native | 5.010 | 5.048 |
| runtime/simd-memory-lanes | aa-go-interpreter | — | 162.517 |
| runtime/simd-memory-lanes | aa-llgo-interpreter | — | 153.523 |
| runtime/simd-memory-lanes | wasmtime-cranelift | — | 5.818 |
| runtime/simd-memory-lanes | wasmer-cranelift | — | 5.845 |
| runtime/simd-memory-lanes | wazero-compiler | — | 6.075 |
| runtime/simd-memory-lanes | wazero-interpreter | — | 187.047 |
| runtime/simd-memory-lanes | wasmedge-interpreter | — | 90.308 |
| runtime/simd-memory-widen | go interpreter | 120.094 | 100.495 |
| runtime/simd-memory-widen | go native | 1.059 | 1.057 |
| runtime/simd-memory-widen | llgo interpreter | 120.236 | 100.346 |
| runtime/simd-memory-widen | llgo native | 1.040 | 1.046 |
| runtime/simd-memory-widen | aa-go-interpreter | — | 120.720 |
| runtime/simd-memory-widen | aa-llgo-interpreter | — | 120.239 |
| runtime/simd-memory-widen | wasmtime-cranelift | — | 1.031 |
| runtime/simd-memory-widen | wasmer-cranelift | — | 1.039 |
| runtime/simd-memory-widen | wazero-compiler | — | 1.155 |
| runtime/simd-memory-widen | wazero-interpreter | — | 143.237 |
| runtime/simd-memory-widen | wasmedge-interpreter | — | 60.690 |

## Correctness and integration limits

The first source passes all 31 Go semantic packages and all 30 coverage gates
(memory 92.52%), pinned LLGo memory/execute/VM suites, scoped race, no-cgo and
vet. After adding the outlining directive, the affected Go/LLGo packages and
quality checks pass again. New byte-oracle tests exhaust all 256 byte values at
each input position and mixed-sign/nonzero neighboring lanes across every
widening and splat operation. Existing range-end, unaligned, offset-overflow,
shared-memory, memory64 and complete guest SIMD tests remain in place.

Final-head dual-platform semantic CI and the contribution's declared benchmark
job remain required before integration. The controlled portfolio and the native
C memory gap remain open. Later evidence-only merges/commits do not change the
measured `2221a91` runtime implementation. The preceding inline outcome is never
replaced with the favorable outlined run or pooled into its measurements.

[Lossless evidence](evidence/performance-packed-vector-loads-2026-11-02.json.gz)
contains both full campaigns, all phases/calls/orders/oracles/A/A controls,
source/binary/SDK identities, candidate code, validation and diagnostic programs,
assembly and every result, including the unsuccessful candidate.

Inline host load: [13.21533203125, 10.15380859375, 9.07763671875]→[11.4677734375, 10.4541015625, 9.34423828125].

Outlined host load: [12.7822265625, 10.64111328125, 9.61572265625]→[8.80029296875, 9.87939453125, 9.4541015625].

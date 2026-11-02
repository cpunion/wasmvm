# Held coalesced-local-copy experiment: 2026-11-02

The six-line runtime fast path is held: most scalar LLGo gains are only 1–2%,
with several results overlapping same-binary A/A variation. GC allocation
regresses 160.713→165.475 ns/iteration; its baseline alias is 166.150, showing
why that difference alone is not proof of an optimization defect. Raw unfavorable
samples remain retained. Private memory is essentially flat.

Static lowering confirms self copies: integer mixing has seven self-copy
local.get instructions among thirty total, scalar memory fifteen among fifty-six,
and private/RMW loops three each. Static counts are not dynamic-frequency or
speedup predictions. Skipping only the copy still pays instruction dispatch,
operand selection and source polling. The observed small gain motivates testing
reduced dispatch work rather than adding unsafe memory access.

Base `e76f172a371482b721812367005a5d31309c8d9a`; candidate
`f319f02b323857ca564c58f242ad2c40865664e4`. The conditional skips data movement
when a local read already occupies its result slot. Original PC, fuel charge and
64-step poll remain. A focused test checks fuel exhaustion at 63 and cancellation
at 64; existing tests preserve vectors, references, suspension and frame reuse.
All 31 Go packages, 30 coverage gates, affected pinned LLGo suites, scoped race,
no-cgo and vet pass. This does not establish performance acceptance.

Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler/interpreter and
WasmEdge 0.17.1 interpreter. Twenty unchanged cases, three fresh serially
interleaved blocks, two warmups and seven warm calls, independent and alternate
checks, exact baseline Go/LLGo interpreter A/A aliases and six applicable C
controls. Twelve loops use one million iterations; four GC and three notify/wait
cases retain ten thousand; the empty control performs one call. Twelve GC
reference rejections are explicit unsupported outcomes. All phases, checksums,
source/binary/toolchain/SDK identities and regressions are retained. This shared
host is not calibrated, and no startup subtraction or parity verdict is used.

| Workload | Go base | Head | LLGo base | Head | LLGo baseline alias |
| --- | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 2875.000 | 3000.000 | 3083.000 | 3292.000 | 2958.000 |
| core/function-calls | 86.577 | 83.884 | 89.978 | 89.336 | 91.154 |
| core/integer-loop | 64.614 | 63.462 | 59.719 | 58.699 | 59.880 |
| core/memory-load-store | 125.193 | 122.317 | 112.538 | 110.252 | 112.874 |
| gc/i31-roundtrip | 67.392 | 65.733 | 57.862 | 55.583 | 58.171 |
| gc/struct-access | 78.292 | 79.608 | 69.683 | 68.987 | 69.871 |
| gc/struct-allocation | 139.067 | 137.654 | 160.713 | 165.475 | 166.150 |
| gc/struct-indirect-access | 127.542 | 128.921 | 119.242 | 114.508 | 117.938 |
| memory/dependent-address | 74.149 | 71.673 | 68.511 | 67.792 | 67.789 |
| memory/float64-unaligned | 122.342 | 118.052 | 110.351 | 107.944 | 112.681 |
| memory/narrow-unaligned | 157.625 | 154.991 | 140.323 | 138.491 | 141.007 |
| runtime/simd-memory-full | 133.330 | 133.794 | 120.345 | 119.164 | 120.727 |
| runtime/simd-memory-lanes | 164.399 | 162.501 | 152.379 | 148.889 | 151.936 |
| runtime/simd-memory-widen | 124.843 | 124.434 | 122.413 | 121.717 | 122.942 |
| threads/atomic-rmw | 69.910 | 69.760 | 66.151 | 65.311 | 65.604 |
| threads/notify-empty | 51.892 | 52.742 | 55.367 | 55.029 | 55.275 |
| threads/private-memory | 58.991 | 58.378 | 55.303 | 55.289 | 55.132 |
| threads/shared-memory | 92.793 | 93.294 | 54.597 | 54.357 | 54.380 |
| threads/wait-mismatch | 81.808 | 83.525 | 75.829 | 76.350 | 75.800 |
| threads/wait-timeout | 83.050 | 85.058 | 74.796 | 75.213 | 74.267 |

Units are ns/source iteration except call-control (ns/call). The complete
Go/native/reference tables and every raw phase are in the
[lossless evidence](evidence/performance-interpreter-local-aliases-2026-11-02.json.gz),
with the exact patch and diagnostic source. The initial diagnostic selected the
extended decoder but the base validator and was correctly rejected; the retained
successful diagnostic selects the same explicit feature profile in both. No
compiler bug is inferred from that configuration mistake.

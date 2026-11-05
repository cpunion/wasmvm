# Default shared memory: complete hosted comparison, 2026-11-05

**Hosted diagnostic; controlled performance acceptance remains open.**
[Run 41976051215](https://github.com/cpunion/wasmvm/actions/runs/41976051215) measures engine
`fb6e872b4758eb5c4e52c334e23c9e12036a58bb`, published at
`9f19606e7e765287c5c11c0558317bdd2a1c7ca2`, against
`747514e40729ea54d482a331dece171f4ba56769`. The
[implementation record](performance-shared-default-memory-2026-11-05.md)
describes the stable-base and monotonic-length proof. All six jobs succeed;
the audit verifies 22 workloads, 362 alternate correctness configurations and
12 explicit reference GC rejections. Nine workloads resolve all four A/A
controls. The longest job is the 6m47s shared build; measurement groups take
29s (GC), 1m44s (Threads), 3m7s (Core), 3m26s (applications) and 4m8s
(memory/SIMD). Source identities, compiler flags, ISA policy and sampling protocol remain
pinned throughout this event.

The shared-memory native elapsed ratio is **0.681x head/base under LLGo**
and **0.761x under Go**. The LLGo result remains **1.267x wazero, 2.853x
Wasmtime and 2.053x Wasmer**. Every Threads control is inconclusive; these
diagnostic medians do not establish a controlled improvement or parity.
Private-memory and atomic native LLGo head/base ratios are 0.911x and 0.991x.

All ratios below are **elapsed time**, lower is better. Values belong to their
own workload/group/host; ratios from different source events are never
multiplied. Native C is the separate unchecked baseline. Compilation, setup,
instantiation, first execution, cleanup, process readiness and warmed API
execution retain distinct raw intervals; no startup or timer interval is
subtracted from a call.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.873× | 1.036× | 6.206× | 3.645× | 1.934× | 2.885× | 1.628× | unavailable | inconclusive |
| core/integer-loop | 0.995× | 0.820× | 1.141× | 1.140× | 1.139× | 0.348× | 0.444× | 1.146× | resolved, advisory |
| core/function-calls | 0.997× | 0.890× | 0.375× | 0.512× | 0.499× | 0.482× | 0.900× | 0.822× | resolved, advisory |
| core/memory-load-store | 1.001× | 0.947× | 0.675× | 1.203× | 1.207× | 0.546× | 0.561× | 1.521× | resolved, advisory |
| applications/sha256-clang | 0.986× | 0.884× | 1.035× | 1.817× | 1.812× | 0.541× | 0.783× | 2.274× | resolved, advisory |
| applications/sha256-rust | 0.987× | 0.913× | 0.976× | 1.782× | 1.748× | 0.557× | 0.798× | 2.235× | resolved, advisory |
| gc/i31-roundtrip | 1.264× | 0.943× | unsupported | 1.044× | unsupported | unsupported | 0.832× | unavailable | inconclusive |
| gc/struct-access | 1.112× | 0.931× | unsupported | 1.683× | unsupported | unsupported | 0.747× | unavailable | inconclusive |
| gc/struct-allocation | 1.261× | 1.880× | unsupported | 0.320× | unsupported | unsupported | 1.428× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.992× | 0.918× | unsupported | 1.432× | unsupported | unsupported | 0.772× | unavailable | inconclusive |
| memory/narrow-unaligned | 1.000× | 0.962× | 0.814× | 1.360× | 1.356× | 0.767× | 0.819× | 2.000× | inconclusive |
| memory/dependent-address | 1.002× | 0.855× | 1.006× | 1.291× | 1.288× | 0.780× | 0.797× | 1.508× | resolved, advisory |
| memory/float64-unaligned | 1.001× | 0.926× | 0.443× | 1.325× | 1.300× | 0.828× | 0.839× | 1.317× | resolved, advisory |
| runtime/simd-memory-full | 0.998× | 0.982× | 0.893× | 1.035× | 1.032× | 0.483× | 0.933× | 0.853× | resolved, advisory |
| runtime/simd-memory-widen | 1.003× | 1.079× | 0.909× | 1.010× | 1.007× | 0.409× | 1.034× | 0.493× | resolved, advisory |
| runtime/simd-memory-lanes | 0.997× | 1.264× | 1.037× | 1.309× | 1.314× | 0.904× | 1.266× | 1.941× | inconclusive |
| threads/private-memory | 0.946× | 1.113× | 1.258× | 2.455× | 1.669× | 0.946× | 0.549× | unavailable | inconclusive |
| threads/shared-memory | 1.057× | 0.612× | 1.267× | 2.853× | 2.053× | 0.925× | 0.544× | unavailable | inconclusive |
| threads/atomic-rmw | 0.969× | 0.709× | 1.505× | 1.351× | 1.199× | 0.930× | 0.568× | unavailable | inconclusive |
| threads/notify-empty | 0.968× | 0.989× | 0.077× | 0.332× | 0.225× | 0.820× | 0.469× | unavailable | inconclusive |
| threads/wait-mismatch | 0.941× | 0.701× | 0.060× | 0.246× | 0.555× | 0.730× | 0.477× | unavailable | inconclusive |
| threads/wait-timeout | 0.938× | 0.697× | 0.006× | 0.051× | 0.025× | 0.094× | 0.001× | unavailable | inconclusive |

The major deficits remain the tiny public API call, SHA, GC access and shared
memory. LLGo/Go differences remain visible, including native GC allocation
(1.261x) and interpreted GC allocation (1.880x). These observations do not
independently identify a new LLGo defect; existing reported reproductions stay
separate, and LLGo #2983 remains with its contributor.

Native guest-call loops beat the listed references in this event, while the
native integer result is 1.146x unchecked C. The earlier event on a different
CPU gave another integer ratio; neither replaces the other or proves global
near-native performance. All adverse rows and controls remain retained.

The [raw evidence index](data/shared-default-memory-hosted-2026-11-05.json)
contains pinned source/compiler/reference/fixture identities and every phase
summary. References remain wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and
WasmEdge 0.17.1. These results do not measure the subsequent constrained-register
lending or complete-write instruction-selection changes. Final-head CI,
review, controlled competitor/native and broader application/WASI-world
acceptance remain open.

Local archive: `artifacts/loop-value-coalescing-20261105/shared-default/hosted-evidence.tar.gz`, 2081066 bytes, SHA-256
`d52be26cdc593f10424e87769dd253e4bd0d54b8400f9e2d376b10c4bcd8546e`.

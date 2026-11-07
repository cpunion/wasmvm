# CFG residency: complete hosted comparison, 2026-11-05

**Hosted diagnostic; controlled performance acceptance remains open.**
[Run 41973062370](https://github.com/cpunion/wasmvm/actions/runs/41973062370) measures engine `2be7d9f17b54b7ed5582faabd741f0f5fa9c01c7`,
published at `747514e40729ea54d482a331dece171f4ba56769`, against
`a47055833582068b88e9aeb51d922ef84662db4f`. All six jobs succeed. The audit
validates all 22 workloads, 362 alternate correctness configurations and 12
explicit reference GC rejections; it independently recalculates all controls.
The longest job is 8m42s; build is 5m10s, Core 1m46s, GC 37s, Threads 1m29s
and applications 3m27s. Six workloads resolve all four A/A controls.

The [actual CFG/cost correction and shared selectors](performance-cfg-register-residency-2026-11-05.md)
recover native memory from the earlier severe regression in this new matched
source event: LLGo head/base is 0.996x and Go is 1.003x. Integer is 0.986x
under LLGo; Clang/Rust SHA are 0.964x/1.008x. Earlier failed events remain
retained. Ratios from different hosts are never multiplied or pooled.

All ratios below are **elapsed time**, lower is better. Native means guest
Wasm compiled to machine code by the named runtime; Go/LLGo identify the host
compiler used to build wasmvm. C is the separately built unchecked native
baseline. Workloads are evaluated within their own group/host.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.643× | 3.019× | 8.183× | 3.720× | 2.544× | 4.455× | 2.440× | unavailable | inconclusive |
| core/integer-loop | 0.995× | 0.918× | 1.009× | 1.007× | 1.000× | 0.646× | 0.704× | 1.012× | inconclusive |
| core/function-calls | 1.003× | 0.868× | 0.435× | 0.435× | 0.441× | 0.653× | 0.850× | 0.824× | inconclusive |
| core/memory-load-store | 0.994× | 0.831× | 0.718× | 1.138× | 1.134× | 0.746× | 0.744× | 1.544× | resolved, advisory |
| applications/sha256-clang | 0.975× | 0.923× | 1.025× | 1.948× | 1.895× | 0.557× | 0.789× | 2.343× | inconclusive |
| applications/sha256-rust | 0.972× | 0.927× | 0.982× | 1.959× | 1.860× | 0.563× | 0.765× | 2.333× | resolved, advisory |
| gc/i31-roundtrip | 1.086× | 0.839× | unsupported | 1.010× | unsupported | unsupported | 0.729× | unavailable | inconclusive |
| gc/struct-access | 0.888× | 0.867× | unsupported | 1.644× | unsupported | unsupported | 0.423× | unavailable | inconclusive |
| gc/struct-allocation | 1.321× | 1.645× | unsupported | 0.304× | unsupported | unsupported | 1.352× | unavailable | inconclusive |
| gc/struct-indirect-access | 1.016× | 0.859× | unsupported | 1.588× | unsupported | unsupported | 0.469× | unavailable | inconclusive |
| memory/narrow-unaligned | 0.996× | 0.979× | 0.900× | 1.381× | 1.382× | 0.500× | 0.707× | 2.190× | resolved, advisory |
| memory/dependent-address | 0.988× | 0.899× | 1.017× | 1.377× | 1.368× | 0.646× | 0.358× | 1.607× | resolved, advisory |
| memory/float64-unaligned | 1.002× | 0.952× | 0.706× | 1.181× | 1.167× | 0.767× | 0.417× | 1.193× | inconclusive |
| runtime/simd-memory-full | 0.996× | 0.944× | 0.892× | 1.107× | 1.068× | 0.482× | 0.722× | 0.923× | inconclusive |
| runtime/simd-memory-widen | 1.003× | 1.029× | 0.764× | 1.726× | 1.737× | 0.391× | 0.994× | 0.444× | resolved, advisory |
| runtime/simd-memory-lanes | 0.996× | 1.138× | 0.953× | 1.152× | 1.159× | 0.472× | 1.188× | 1.340× | resolved, advisory |
| threads/private-memory | 1.002× | 1.087× | 1.282× | 2.483× | 1.741× | 0.645× | 0.534× | unavailable | inconclusive |
| threads/shared-memory | 1.112× | 0.604× | 1.756× | 4.076× | 2.854× | 0.640× | 0.528× | unavailable | inconclusive |
| threads/atomic-rmw | 0.871× | 0.698× | 1.395× | 1.218× | 1.097× | 0.651× | 0.578× | unavailable | inconclusive |
| threads/notify-empty | 0.945× | 0.997× | 0.082× | 0.331× | 0.235× | 0.540× | 0.471× | unavailable | inconclusive |
| threads/wait-mismatch | 0.947× | 0.665× | 0.068× | 0.257× | 0.633× | 0.654× | 0.472× | unavailable | inconclusive |
| threads/wait-timeout | 0.946× | 0.661× | 0.006× | 0.053× | 0.026× | 0.094× | 0.001× | unavailable | inconclusive |

Long integer work is close to native C, and generated guest calls beat these
references in this event. Significant deficits remain: the tiny public API
call, SHA, GC access and shared memory. A fast allocation or wait/notify row
does not close those gaps. The interpreter is usually faster than the pinned
interpreter references here, with visible allocation/lane/API exceptions.
Only six rows resolve the diagnostic controls, and hosted hardware is not
isolated, so none establishes controlled portfolio acceptance.

## Compilation, instantiation and first execution

Each phase below is a separately timed interval, in microseconds. These are
medians over the five processes for the Clang SHA guest. The complete raw
records also retain file load, export lookup, cleanup, process readiness,
warmups, all warm samples and the timer pairs. First-call time includes the
guest work; it is not pure entry overhead. No timer or phase is subtracted.

| Runtime | Engine setup (us) | Wasm compile/validate (us) | Instantiate (us) | First call (us) |
| --- | ---: | ---: | ---: | ---: |
| wasmvm / Go native | 15.45 | 4168.25 | 79.04 | 3377.13 |
| wasmvm / LLGo native | 9.30 | 7527.66 | 37.12 | 3311.79 |
| wazero compiler | 141.57 | 838.77 | 52.58 | 3419.96 |
| Wasmtime Cranelift | 197.71 | 2902.47 | 85.58 | 2773.24 |
| Wasmer Cranelift | 176.44 | 2486.59 | 86.47 | 3004.96 |

LLGo compile/validate remains about 1.81x Go for this guest; native warm
execution is 0.975x Go. Actual published arm64 code equality is recorded in
the separate static experiment, rather than inferred for this x86 host.
Both host compilation costs and generated-code throughput still require work.

The [raw evidence index](data/cfg-register-residency-hosted-2026-11-05.json)
contains all source/toolchain/fixture identities, phase summaries and archive
hashes. Pinned references: wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and
WasmEdge 0.17.1. No flags, ISA policy, sampling protocol or prior cohort changed.

Local archive: `artifacts/loop-value-coalescing-20261105/cfg-costs/hosted-evidence.tar.gz`, 2074512 bytes, SHA-256
`f191aeec800c1fb9ca73601c174c20aa558c5c50c4bbeb8914d65636388129a7`.

The [subsequent shared-memory implementation](performance-shared-default-memory-2026-11-05.md)
is a different source event and is not measured by these ratios. Final-head CI,
review, controlled performance and broader application/WASI-world acceptance
remain open. LLGo #2983 stays with its existing contributor.

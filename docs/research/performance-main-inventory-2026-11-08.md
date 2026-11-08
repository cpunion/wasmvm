# Complete integrated-main comparison: 2026-11-08

Measured `c735526` versus main `43afd27`; the production Go sources are identical.
The CLI schema repair does not change the engine. The measured head merged as
`eac84b0` after all 23 required CI jobs, 240 conformance reports and 30 coverage
gates passed. [PR #129](https://github.com/cpunion/wasmvm/pull/129) records
[the full table](https://github.com/cpunion/wasmvm/pull/129#issuecomment-6717792413).

The [Linux event](https://github.com/cpunion/wasmvm/actions/runs/42449932354)
retains the complete 22-case original portfolio and a separate four-case
sustained GC cohort. Five groups have independent hosted machines; they are
never pooled into a single-host or controlled result. Go 1.27.3, LLGo `3571c25`
(`-O3 -lto=full`), LLVM 22.1.8, references, fixtures, counts and policies remain
fixed. Five process blocks contain seven interleaved warm calls after two
warmups. Allocation uses process means including every pause; all other
primary statistics retain process medians. Setup/compile/instantiate/first/
warm/cleanup and process intervals stay separate without subtraction.

Original integrity checks verify 362 alternate configurations and 12 explicitly
unsupported GC configurations; sustained verifies 56 and 12. Four source/
compiler identities agree across every group and sustained reuses the same
binaries. Original controls resolve **7/22**, sustained **2/4**. Unresolved
and adverse results remain; no favorable retry, profile/count/reference sweep
or synthetic fastest-engine row is used. Native C is unchecked and unavailable
for GC/Threads/control cases; unsupported is never timed as zero.

| Group | CPU model |
| --- | --- |
| core | INTEL(R) XEON(R) PLATINUM 8573C |
| memory-simd | AMD EPYC 7763 64-Core Processor |
| gc | AMD EPYC 7763 64-Core Processor |
| threads | AMD EPYC 9V45 96-Core Processor |
| applications | INTEL(R) XEON(R) PLATINUM 8573C |

Ratios compare complete elapsed API intervals, **lower is better**.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control / median-v1 | 1.847× | 2.030× | 8.287× | 4.933× | 2.444× | 2.714× | 1.596× | unavailable | inconclusive |
| core/integer-loop / median-v1 | 0.999× | 0.772× | 0.998× | 1.006× | 1.002× | 0.696× | 0.929× | 1.018× | resolved, advisory |
| core/function-calls / median-v1 | 0.995× | 0.792× | 0.434× | 0.758× | 0.836× | 0.753× | 1.101× | 0.819× | resolved, advisory |
| core/memory-load-store / median-v1 | 0.997× | 0.827× | 0.588× | 1.073× | 1.059× | 0.827× | 0.994× | 1.437× | resolved, advisory |
| memory/narrow-unaligned / median-v1 | 0.985× | 0.933× | 0.876× | 1.308× | 1.304× | 0.491× | 0.674× | 2.072× | resolved, advisory |
| memory/dependent-address / median-v1 | 0.988× | 0.801× | 0.966× | 1.304× | 1.301× | 0.574× | 0.350× | 1.521× | resolved, advisory |
| memory/float64-unaligned / median-v1 | 1.004× | 0.873× | 0.681× | 1.129× | 1.123× | 0.713× | 0.410× | 1.145× | inconclusive |
| runtime/simd-memory-full / median-v1 | 0.997× | 0.875× | 0.858× | 1.065× | 1.029× | 0.493× | 0.758× | 0.890× | resolved, advisory |
| runtime/simd-memory-widen / median-v1 | 1.000× | 0.878× | 0.678× | 1.534× | 1.530× | 0.448× | 0.942× | 0.395× | resolved, advisory |
| runtime/simd-memory-lanes / median-v1 | 0.994× | 1.061× | 0.935× | 1.132× | 1.139× | 0.474× | 1.152× | 1.315× | inconclusive |
| gc/i31-roundtrip / median-v1 | 0.997× | 0.800× | unsupported | 0.901× | unsupported | unsupported | 0.559× | unavailable | inconclusive |
| gc/struct-access / median-v1 | 0.967× | 0.766× | unsupported | 1.753× | unsupported | unsupported | 0.369× | unavailable | inconclusive |
| gc/struct-allocation / mean-including-pauses-v1 | 1.322× | 1.283× | unsupported | 0.498× | unsupported | unsupported | 1.373× | unavailable | inconclusive |
| gc/struct-indirect-access / median-v1 | 1.007× | 0.789× | unsupported | 1.038× | unsupported | unsupported | 0.554× | unavailable | inconclusive |
| threads/private-memory / median-v1 | 1.900× | 0.999× | 1.239× | 2.632× | 2.004× | 1.073× | 1.061× | unavailable | inconclusive |
| threads/shared-memory / median-v1 | 1.391× | 0.613× | 0.974× | 2.042× | 1.664× | 1.104× | 1.075× | unavailable | inconclusive |
| threads/atomic-rmw / median-v1 | 1.016× | 0.595× | 1.128× | 1.117× | 1.085× | 1.016× | 1.047× | unavailable | inconclusive |
| threads/notify-empty / median-v1 | 1.314× | 0.777× | 0.049× | 0.124× | 0.111× | 0.857× | 0.803× | unavailable | inconclusive |
| threads/wait-mismatch / median-v1 | 1.204× | 0.601× | 0.039× | 0.098× | 0.342× | 0.624× | 0.902× | unavailable | inconclusive |
| threads/wait-timeout / median-v1 | 0.833× | 0.612× | 0.004× | 0.027× | 0.009× | 0.088× | 0.001× | unavailable | inconclusive |
| applications/sha256-clang / median-v1 | 0.989× | 0.822× | 0.727× | 1.185× | 1.162× | 0.763× | 0.972× | 1.531× | inconclusive |
| applications/sha256-rust / median-v1 | 1.000× | 0.826× | 0.703× | 1.101× | 1.092× | 0.767× | 0.997× | 1.507× | inconclusive |

## Separate sustained GC

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| gc/i31-roundtrip / median-v1 | 0.997× | 0.788× | unsupported | 0.721× | unsupported | unsupported | 0.723× | unavailable | inconclusive |
| gc/struct-access / median-v1 | 0.995× | 0.738× | unsupported | 1.258× | unsupported | unsupported | 0.696× | unavailable | resolved, advisory |
| gc/struct-allocation / mean-including-pauses-v1 | 1.075× | 1.192× | unsupported | 0.684× | unsupported | unsupported | 1.433× | unavailable | inconclusive |
| gc/struct-indirect-access / median-v1 | 0.992× | 0.775× | unsupported | 0.937× | unsupported | unsupported | 0.737× | unavailable | resolved, advisory |

## Remaining gaps

The Core group observes LLGo native integer/function/private-memory loops at
2.037/1.036/1.000 ns per source iteration. Wasmtime is 2.025/1.366/0.932.
Optimized guest calls may be inlined; these figures are not machine-call latency.
The public-call control remains 6.911 microseconds under LLGo, versus Go 3.742,
wazero 0.834, Wasmtime 1.401 and Wasmer 2.828. Its controls are unresolved.
Invocation ownership/context/marshalling costs remain inside that interval;
it must not be reported as guest instruction latency.

Narrow/dependent memory and widening SIMD remain behind Wasmtime/Wasmer;
private/shared Threads memory, direct GC access, allocation interpretation and
SHA/native-C differences also remain visible. Individual ratios are tied to
their group/host and cannot establish broad parity. Existing LLGo issues
[#2932](https://github.com/xgo-dev/llgo/issues/2932),
[#2974](https://github.com/xgo-dev/llgo/issues/2974) and
[#3012](https://github.com/xgo-dev/llgo/issues/3012) do not explain every gap.
Unmerged compiler work is not a substitute for the pinned production profile.

The [complete machine record](data/performance-main-inventory-2026-11-08.json)
retains every original/sustained observation, phase, host identity and audit.
Immutable archive: `artifacts/closeout-20261108/final-evidence.tar.gz`,
10476130 bytes, 323 files; SHA-256
`fc8e74339cf63110a8aad1b1e048815a62f9c2582490b16ed56fdcbb7a4a18c8`.
Controlled performance, required first-call/latency/allocation/memory evidence
and production release remain open.

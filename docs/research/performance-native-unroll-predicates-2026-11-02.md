# Unrolled comparison liveness: 2026-11-02

The candidate removes only proven-dead writes of intermediate comparison results
inside an already proven, nontrapping, bounded unrolled memory group. Operand
reads precede writes, including aliased local operations. If the following body
can read a comparison before overwriting it, its known scalar result remains
materialized. The final comparison is always emitted normally. Original PCs,
cold paths, unsigned bounds/wrap guards, source fuel and cancellation points
remain present. No new bytecode, native ABI or host execution framework is added.

The shared scalar read/write classification is used by ordinary fallthrough
liveness and the unroll proof. The production change is one bounded scan and a
plan flag; it does not recognize fixture names, constants or expected results.

Base `2221a919a46ce23527a5e446e4c8f069546bf997`; measured candidate
`93190e993918ab50df5fc81adfd509c8a4fcf900`. Later evidence-only commits do not change that implementation.
Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift and wazero 1.12.0 compiler.

## Scope and observations

A compile-only capture compares 86 native body pairs across twenty fixtures,
amd64/arm64, both policy tiers and optional expansion. Exactly four bodies
change: unmetered scalar memory and float64 memory on each architecture, all
24 bytes smaller. Metered code and every other captured function are byte-
identical. Cross-amd64 uses its baseline feature mask; ARM64 retains the actual
host mask. Foreign code is never published or executed. This does not establish
performance on another architecture or CPU feature mask.

The scoped native study measures those two scalar memory kernels and four
scalar controls with unchanged guest code. Core memory uses 50,000,000 source
iterations, the other five 5,000,000. All named configurations, including C,
receive matching counts and checksums. Three fresh serially interleaved process
blocks, two warmups and seven measured calls retain exact baseline Go/LLGo
native A/A aliases. Independent original and alternate-input oracles pass.
All preparation, first/warm, readiness and cleanup phases remain separate.
Interpreter references are not timed in this compiler-only scope; the complete
final-head PR benchmark and controlled portfolio remain required.

The table uses medians of process medians, not pooled calls. Scalar memory
improves Go 0.450596→0.443957 and LLGo 0.450242→0.442667 ns/iteration.
Every paired memory block improves: Go ratios 0.974990–0.987679, LLGo
0.982586–0.988415. Baseline A/A ratios span 1.000216–1.006816 and
0.996080–1.004540, respectively. These do not resolve the declared controlled
±0.25% calibration requirement. The observation supports further evaluation,
not a controlled parity verdict. Float64 is inconclusive: Go regresses 0.86%,
LLGo improves 0.72%, with its A/A faster by about 1% as well.

C memory remains 0.265180 versus LLGo 0.442667 (1.67x); Wasmtime is
0.468460, Wasmer 0.516500 and wazero compiler 0.777575. Native LLGo calls
regress about 1.0% and dependent-address 1.5%; their baseline A/A controls
also move, and these regressions remain retained rather than dismissed from
unchanged guest code. The separate C assembly exposes induction updates and
independent accumulators absent from our expanded loop. That motivates further
address/reduction work; the small cleanup does not explain or close the native
memory gap. No LLGo issue is inferred from these throughput/layout differences.

| Workload | Configuration | ns/source iteration |
| --- | --- | ---: |
| core/function-calls | aa-go-native | 0.770100 |
| core/function-calls | aa-llgo-native | 0.763608 |
| core/function-calls | base-go-native | 0.766942 |
| core/function-calls | base-llgo-native | 0.753592 |
| core/function-calls | head-go-native | 0.774267 |
| core/function-calls | head-llgo-native | 0.761192 |
| core/function-calls | native-c | 0.729400 |
| core/function-calls | wasmer-cranelift | 0.741800 |
| core/function-calls | wasmtime-cranelift | 0.884800 |
| core/function-calls | wazero-compiler | 1.303208 |
| core/integer-loop | aa-go-native | 1.460367 |
| core/integer-loop | aa-llgo-native | 1.434508 |
| core/integer-loop | base-go-native | 1.452858 |
| core/integer-loop | base-llgo-native | 1.447725 |
| core/integer-loop | head-go-native | 1.448333 |
| core/integer-loop | head-llgo-native | 1.453750 |
| core/integer-loop | native-c | 1.410400 |
| core/integer-loop | wasmer-cranelift | 1.440400 |
| core/integer-loop | wasmtime-cranelift | 1.410600 |
| core/integer-loop | wazero-compiler | 1.469075 |
| core/memory-load-store | aa-go-native | 0.451694 |
| core/memory-load-store | aa-llgo-native | 0.449921 |
| core/memory-load-store | base-go-native | 0.450596 |
| core/memory-load-store | base-llgo-native | 0.450242 |
| core/memory-load-store | head-go-native | 0.443957 |
| core/memory-load-store | head-llgo-native | 0.442667 |
| core/memory-load-store | native-c | 0.265180 |
| core/memory-load-store | wasmer-cranelift | 0.516500 |
| core/memory-load-store | wasmtime-cranelift | 0.468460 |
| core/memory-load-store | wazero-compiler | 0.777575 |
| memory/dependent-address | aa-go-native | 1.736867 |
| memory/dependent-address | aa-llgo-native | 1.698008 |
| memory/dependent-address | base-go-native | 1.743375 |
| memory/dependent-address | base-llgo-native | 1.714617 |
| memory/dependent-address | head-go-native | 1.745367 |
| memory/dependent-address | head-llgo-native | 1.739958 |
| memory/dependent-address | native-c | 1.290400 |
| memory/dependent-address | wasmer-cranelift | 1.910600 |
| memory/dependent-address | wasmtime-cranelift | 1.894400 |
| memory/dependent-address | wazero-compiler | 2.811808 |
| memory/float64-unaligned | aa-go-native | 0.705967 |
| memory/float64-unaligned | aa-llgo-native | 0.698292 |
| memory/float64-unaligned | base-go-native | 0.708025 |
| memory/float64-unaligned | base-llgo-native | 0.705292 |
| memory/float64-unaligned | head-go-native | 0.714092 |
| memory/float64-unaligned | head-llgo-native | 0.700208 |
| memory/float64-unaligned | native-c | 0.655600 |
| memory/float64-unaligned | wasmer-cranelift | 0.699400 |
| memory/float64-unaligned | wasmtime-cranelift | 0.699800 |
| memory/float64-unaligned | wazero-compiler | 0.797267 |
| memory/narrow-unaligned | aa-go-native | 0.678650 |
| memory/narrow-unaligned | aa-llgo-native | 0.677075 |
| memory/narrow-unaligned | base-go-native | 0.680125 |
| memory/narrow-unaligned | base-llgo-native | 0.678433 |
| memory/narrow-unaligned | head-go-native | 0.678575 |
| memory/narrow-unaligned | head-llgo-native | 0.678575 |
| memory/narrow-unaligned | native-c | 0.454800 |
| memory/narrow-unaligned | wasmer-cranelift | 0.708800 |
| memory/narrow-unaligned | wasmtime-cranelift | 0.686800 |
| memory/narrow-unaligned | wazero-compiler | 0.949633 |

## Ordinary C admission and validation

The isolated C diagnostic expands only its CLI count admission from one million
to 100 million, with a separate conservative five-million cap for float64.
Its timed kernel/protocol source is unchanged, and the complete emitted C
kernel assembly matches the prior binary's assembly. This is a declared control
input-admission change, not a silent replacement of an earlier baseline.
The native C algorithm has fewer Wasm contracts; missing VM phases are not
reported as measured zeroes. Clang 22.1.8 `-O3 -Wall -Wextra -Werror`, source,
binary and assembly identities remain retained.

The float cap guarantees an in-range result even at seed byte 255. An independent
integer cycle-count calculation predicts 2,038,251,984 at five million iterations
and seed 0xffffffff; the C executable matches it. An oversized float request is
rejected before execution, avoiding the undefined C conversion/guest truncation
trap that a uniform raised cap would permit. Measured floating results use the
same valid seed/count across all configurations.

All 31 Go semantic packages, all 30 coverage gates, pinned LLGo codegen/execute/VM,
codegen race, no-cgo and vet pass at the measured candidate. New tests compare
all original PCs, scalar slots, visible memory and exit state against ordinary
code, including live/dead comparisons, both backedge polarities, groups/tails,
unsigned wrap and read-before-write aliases. Existing fuel/cancellation and
code-budget regressions remain in the passing suites. Final-head dual-platform
checks and semantic artifact audit are still required before integration.

[Lossless evidence](evidence/performance-native-unroll-predicates-2026-11-02.json.gz)
retains every native/C record, phase, sample order, original/alternate oracle,
A/A control, source/binary/SDK identity, compiler/capture/measurement program,
C admission/assembly proof and all validation outcomes. The initial architecture
study is retained separately inside the same evidence; its capture is diagnostic,
not timed execution.

Shared-host load: [6.86669921875, 8.81884765625, 9.587890625]→[5.83740234375, 8.35986328125, 9.388671875]. No affinity or host isolation is claimed.

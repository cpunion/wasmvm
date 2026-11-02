# Held native reduction split: 2026-11-02

Do not adopt this candidate. Splitting the bounded memory group's I32 sum into
two register accumulators makes the target slower under both Go and pinned LLGo.
Its semantic proof succeeds; its profitability hypothesis does not.

Base `93190e993918ab50df5fc81adfd509c8a4fcf900`; experimental candidate
`a37d9cf69b12c98361384ed3689e73f27c3e371a`. The production contribution
branch keeps the base implementation. The prototype remains a local branch and
its source is retained in the evidence, not introduced through this report.

The candidate proves that an I32 sum is read only by its single update, accepts
validated self-copies, rejects other observations/writes, and requires a temporary
result to die before group end. It uses a spare ARM caller register in alternating
copies and merges modulo-32-bit results before any exit/poll. Other architectures
and exhausted inventories retain ordinary lowering. No scalar/float reassociation
is inferred from relative competitor speed alone.

A compile-only comparison of 86 body pairs changes exactly the unmetered ARM64
scalar-memory function, 3,772→3,780 bytes. All amd64, metered and other captured
bodies are byte-identical. Foreign code is never published/executed. CPU masks
are the declared baseline amd64/actual ARM64 diagnostic profiles, not every
possible feature configuration. The added zero initialization and group merge
are real extra instructions; the experiment does not isolate their individual
costs or prove the entire C lead is caused by one dependency chain.

Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`;
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift and wazero 1.12.0 compiler. The same
six-case native scope as the preceding cleanup runs three fresh serially
interleaved process blocks, two warmups, seven calls, exact native baseline A/A
aliases, 50 million core-memory iterations and five million for the other cases.
Independent original/alternate-input checks pass. Every engine and applicable
C control uses matching inputs/counts. The C executable/source are reused from
the [preceding scoped campaign](performance-native-unroll-predicates-2026-11-02.md),
including its declared count admission, safe float cap and unchanged kernel
assembly proof. All preparation/first/warm/cleanup/readiness phases remain
separate. This native-only study does not replace final PR/portfolio evidence.

Target process-median observations, ns/source iteration:

| Host | Base | Candidate | Baseline A/A |
| --- | ---: | ---: | ---: |
| Go | 0.443783 | 0.455365 | 0.449548 |
| LLGo | 0.447676 | 0.457256 | 0.446228 |

Every paired target block regresses: Go head/base 1.014129–1.034377,
LLGo 1.014403–1.027729. A/A spans 0.997523–1.013253 and
0.993993–1.001309 respectively. The candidate/base medians regress 2.61% and
2.14%. C is 0.267280, Wasmtime 0.468280, Wasmer 0.518760 and wazero compiler
0.786854. Being faster than some competitors is insufficient to accept the
owned-baseline regression. Unrelated small control improvements are not used
to average away this target failure. No LLGo issue is inferred from owned
machine-code selection.

## All observations

| Workload | Configuration | ns/source iteration |
| --- | --- | ---: |
| core/function-calls | aa-go-native | 0.776967 |
| core/function-calls | aa-llgo-native | 0.764658 |
| core/function-calls | base-go-native | 0.774100 |
| core/function-calls | base-llgo-native | 0.767792 |
| core/function-calls | head-go-native | 0.768358 |
| core/function-calls | head-llgo-native | 0.762142 |
| core/function-calls | native-c | 0.735600 |
| core/function-calls | wasmer-cranelift | 0.760400 |
| core/function-calls | wasmtime-cranelift | 0.884800 |
| core/function-calls | wazero-compiler | 1.313367 |
| core/integer-loop | aa-go-native | 1.557183 |
| core/integer-loop | aa-llgo-native | 1.547075 |
| core/integer-loop | base-go-native | 1.551550 |
| core/integer-loop | base-llgo-native | 1.556358 |
| core/integer-loop | head-go-native | 1.540225 |
| core/integer-loop | head-llgo-native | 1.545375 |
| core/integer-loop | native-c | 1.541600 |
| core/integer-loop | wasmer-cranelift | 1.538200 |
| core/integer-loop | wasmtime-cranelift | 1.545400 |
| core/integer-loop | wazero-compiler | 1.587425 |
| core/memory-load-store | aa-go-native | 0.449548 |
| core/memory-load-store | aa-llgo-native | 0.446228 |
| core/memory-load-store | base-go-native | 0.443783 |
| core/memory-load-store | base-llgo-native | 0.447676 |
| core/memory-load-store | head-go-native | 0.455365 |
| core/memory-load-store | head-llgo-native | 0.457256 |
| core/memory-load-store | native-c | 0.267280 |
| core/memory-load-store | wasmer-cranelift | 0.518760 |
| core/memory-load-store | wasmtime-cranelift | 0.468280 |
| core/memory-load-store | wazero-compiler | 0.786854 |
| memory/dependent-address | aa-go-native | 1.830675 |
| memory/dependent-address | aa-llgo-native | 1.819067 |
| memory/dependent-address | base-go-native | 1.816775 |
| memory/dependent-address | base-llgo-native | 1.804792 |
| memory/dependent-address | head-go-native | 1.826017 |
| memory/dependent-address | head-llgo-native | 1.817733 |
| memory/dependent-address | native-c | 1.405800 |
| memory/dependent-address | wasmer-cranelift | 2.089400 |
| memory/dependent-address | wasmtime-cranelift | 2.058200 |
| memory/dependent-address | wazero-compiler | 3.010217 |
| memory/float64-unaligned | aa-go-native | 0.714542 |
| memory/float64-unaligned | aa-llgo-native | 0.708200 |
| memory/float64-unaligned | base-go-native | 0.714750 |
| memory/float64-unaligned | base-llgo-native | 0.710583 |
| memory/float64-unaligned | head-go-native | 0.713075 |
| memory/float64-unaligned | head-llgo-native | 0.715875 |
| memory/float64-unaligned | native-c | 0.660400 |
| memory/float64-unaligned | wasmer-cranelift | 0.704400 |
| memory/float64-unaligned | wasmtime-cranelift | 0.700000 |
| memory/float64-unaligned | wazero-compiler | 0.807600 |
| memory/narrow-unaligned | aa-go-native | 0.682533 |
| memory/narrow-unaligned | aa-llgo-native | 0.674608 |
| memory/narrow-unaligned | base-go-native | 0.677925 |
| memory/narrow-unaligned | base-llgo-native | 0.670842 |
| memory/narrow-unaligned | head-go-native | 0.677742 |
| memory/narrow-unaligned | head-llgo-native | 0.673500 |
| memory/narrow-unaligned | native-c | 0.451600 |
| memory/narrow-unaligned | wasmer-cranelift | 0.706000 |
| memory/narrow-unaligned | wasmtime-cranelift | 0.685200 |
| memory/narrow-unaligned | wazero-compiler | 0.976350 |

All 31 Go semantic packages and 30 coverage gates pass; pinned LLGo
codegen/execute/VM, codegen race, no-cgo and vet pass as well. Focused tests
compare every original PC, scalar slot, memory byte and exit state, covering
wrap, groups/tails, local versus temporary sums and observed/escaped values.
Selection rejects doubling, subtraction, missing homes and exhausted spare
registers. Existing cancellation/fuel/budget tests remain passing. Correctness
does not substitute for the failed throughput objective.

The next hypothesis should address other differences, including repeated group
admission/control and address induction, rather than keeping this additional
planner because it looks like the C optimizer. Neither C's complete mechanism
nor broad native performance acceptance is established by this experiment.

[Lossless evidence](evidence/performance-native-reduction-held-2026-11-02.json.gz)
retains the held implementation/tests, 60 summaries, all calls/phases/orders,
A/A/oracle/source/binary/SDK identities, code comparisons, validation and every
build/capture/measurement program.

Shared-host load: [7.611328125, 8.91259765625, 8.248046875]→[8.396484375, 8.91650390625, 8.2724609375]. No affinity or host isolation is claimed.

# Bounded native memory-loop expansion: experimental evaluation

Candidate `52fae050fab26fb7a08720601d80bd306d4ddef7` is **not integrated**.
The source implements a general bounded loop transformation, but the current
shared-host data do not establish a repeatable improvement or competitive
acceptance. Preserve the candidate and its complete evidence for controlled
evaluation; do not infer parity from selected wins.

## Mechanism and contracts

The existing private-memory range proof admits unsigned I32 unit-stride loops
with invariant bounds. The native emitter groups four iterations (two for
larger bodies), with at most 256 original slots between cancellation polls.
A dynamic remaining-iteration guard sends short tails and wraparound cursors
through the original body. Every original PC retains its ordinary continuation
entry; intermediate comparison results and ordered memory effects are retained.
Exact-fuel execution uses its original tier. Shared memory, helper calls,
possible traps, guarded ranges and oversized bodies remain excluded.

The first candidate expanded serial invariant-address load/store chains and
regressed the private increment loop from 0.319 to 0.944 ns/source iteration.
The final candidate conservatively requires independently generated varying
addresses and excludes pointer chasing. This filter is a cost heuristic, not
a new memory-safety proof. Store-forwarding effects are a hypothesis for the
initial regression, without hardware-counter attribution.

## Paired results and remaining gaps

Base `ffd6300e596735d2c51181897ae92499fcde347b`, Apple M4 Max/macOS arm64,
Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`. Twelve unchanged kernels
run at 1,000,000 source iterations with independent and alternate-input
oracles, three fresh interleaved process blocks, two warmups and seven calls.
Wasmtime 48.0.1/Wasmer 7.3.0 use Cranelift; wazero 1.12.0 uses its compiler.
The separate C reference uses Clang -O3 without fast-math and has fewer runtime
contracts. Startup, preparation, first call, warm execution and cleanup remain
separate; no timing subtraction is used. Units below are ns/source iteration.

| Workload | Go base | Head | LLGo base | Head | Native C | Wasmtime | Wasmer | wazero |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core/integer-loop | 2.137 | 2.143 | 2.115 | 2.126 | 2.007 | 2.097 | 2.113 | 2.385 |
| core/function-calls | 1.030 | 1.006 | 0.997 | 0.990 | 0.987 | 1.149 | 0.979 | 1.633 |
| core/memory-load-store | 0.726 | 0.706 | 0.769 | 0.697 | 0.400 | 0.645 | 0.819 | 1.410 |
| memory/narrow-unaligned | 1.266 | 1.154 | 1.252 | 1.175 | 0.763 | 1.303 | 1.329 | 1.719 |
| memory/dependent-address | 2.767 | 2.778 | 2.975 | 2.871 | 2.139 | 3.087 | 3.335 | 5.020 |
| memory/float64-unaligned | 1.111 | 1.116 | 1.118 | 1.109 | 1.001 | 1.075 | 1.076 | 1.314 |
| runtime/simd-memory-full | 4.255 | 4.279 | 4.245 | 4.269 | — | 4.916 | 4.362 | 4.779 |
| runtime/simd-memory-widen | 1.438 | 1.439 | 1.428 | 1.425 | — | 1.412 | 1.444 | 1.454 |
| runtime/simd-memory-lanes | 6.923 | 6.684 | 6.841 | 6.674 | — | 7.919 | 7.923 | 8.035 |
| threads/private-memory | 0.410 | 0.425 | 0.411 | 0.453 | — | 1.300 | 1.301 | 1.306 |
| threads/shared-memory | 1.254 | 1.257 | 1.254 | 1.248 | — | 1.174 | 1.327 | 1.611 |
| threads/atomic-rmw | 2.488 | 2.473 | 2.468 | 2.476 | — | 2.463 | 5.260 | 2.505 |

The scalar memory row improves in this sample but remains behind native C
and Wasmtime. The private-memory control regresses despite unchanged generated
guest code. SIMD full-width memory and atomic RMW also regress slightly.
The evidence retains every row and both interpreter control kernels; none is
removed from the portfolio or reclassified as accepted performance.

## Longer controls and uncertainty

A predeclared follow-up uses the same binaries, 25,000,000 private increments
and 3,000,000 scalar memory iterations, six fresh interleaved process blocks
and seven calls. Independent A/A aliases execute the exact baseline binaries:

| Workload / compiler | Base | Head | Same-binary baseline alias |
| --- | ---: | ---: | ---: |
| threads/private-memory / go | 0.682 | 0.674 | 0.688 |
| threads/private-memory / llgo | 0.735 | 0.684 | 0.693 |
| core/memory-load-store / go | 1.307 | 1.127 | 1.280 |
| core/memory-load-store / llgo | 1.329 | 1.224 | 1.222 |

Shared-host load was roughly 24–36 runnable tasks on 16 cores in the first
campaign. Longer controls also vary materially; even identical-binary aliases
differ by approximately 6–8% for LLGo. These controls expose uncertainty and do
not calibrate a statistical acceptance threshold. Private-memory and dependent
address guest code is byte-identical between base/head for both amd64 and arm64,
with hashes retained. A final-block private-memory warm call increased from
roughly 0.4 ms to 2.3 ms; the raw sample remains included.

## Validation and retained evidence

All 31 Go semantic packages, 30 cross-package coverage gates, pinned LLGo
codegen/execute/VM suites, codegen race and no-cgo checks, and vet pass. Tests
cover original continuation entries and frames, short tails, unusual cursors,
unsigned wrap, both branch polarities, constant bounds, code budgets and proof
exclusions. The candidate still requires controlled cross-platform performance
evaluation before integration.

[Lossless compressed evidence](evidence/performance-native-loop-unroll-2026-11-02.json.gz)
retains both complete campaigns, rejected first candidate, all phase samples,
build/source/toolchain identities, measurement drivers, longer A/A controls,
code-identity hashes and validation logs. Decode with Python's `gzip` and `json`.
The original pinned LLGo source worktree disappeared externally after timing;
subsequent checks use an exact git-archive recovery of the same revision.
The compiler binary and timed binaries retain their original hashes.


## Further four-case controls after load fell

The same candidate/baseline binaries were retained for a new declared follow-up:
25,000,000 private increments, 3,000,000 scalar/narrow/dependent memory
iterations, six fresh process blocks, seven calls, and two same-baseline binary
aliases. All checksums and phase records pass. Load was roughly 10 at the
start and 15 at the end; this remains a shared host. Units are ns/source
iteration:

| Workload | Go base | Head | Go A/A | LLGo base | Head | LLGo A/A | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.311 | 0.311 | 0.311 | 0.309 | 0.311 | 0.310 | 0.530 | 0.416 |
| core/memory-load-store | 0.517 | 0.476 | 0.522 | 0.512 | 0.481 | 0.512 | 0.476 | 0.534 |
| memory/narrow-unaligned | 0.824 | 0.842 | 0.813 | 0.827 | 0.875 | 0.840 | 0.841 | 0.825 |
| memory/dependent-address | 1.968 | 1.979 | 1.971 | 1.938 | 1.983 | 1.956 | 2.218 | 2.247 |

Scalar memory improves approximately 8% for Go / 6% for LLGo and approaches
Wasmtime in this sample. Narrow memory instead regresses approximately 2% / 6%;
its LLGo A/A alias differs by about 1.6%. Dependent-address code remains
byte-identical, but its measurements also differ by about 2%. This does not
establish a profitable portfolio transformation, and the candidate remains
held. A broader profitability model and controlled cross-platform evidence
are still required.

The first follow-up attempt mistakenly included the unchanged C reference at
3,000,000 iterations. Its explicit 1,000,000-iteration limit exited before the
readiness handshake. That invalid setup/log is retained; no timings from the
failed attempt are accepted. The corrected longer controls omit C rather than
change its baseline or mix loop counts. The earlier equal-count native-C
comparisons remain separate. The evidence contains every corrected phase and
same-binary sample, plus the failed attempt's source and diagnostic.

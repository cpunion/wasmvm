# Bounded native word-loop expansion: 2026-11-02

The measured candidate reduces LLGo native scalar memory from **0.526 to
0.473 ns/source iteration**, approximately 10%, versus Wasmtime **0.480** and
ordinary C **0.269** on this local ARM64 run. Go improves **0.540→0.493**.
This closes the measured Wasmtime gap but still leaves about 1.76x C cost.
Controlled cross-platform performance acceptance remains open.

## Conservative selection and execution contracts

Base `ffd6300e596735d2c51181897ae92499fcde347b`; measured candidate
`06196a9bba2e6619d400be96ac14b8b90053fd26`. The later `027cfa8` changes only a
test deadline. Word accesses in proven private-memory regions can expand by two
or four copies when independent varying addresses and an unsigned unit-stride
recurrence prove safe execution. No observation, potentially trapping scalar
operation or unproven access may enter an expanded group. Each group executes
at most 256 original slot instructions; original dispatcher PCs retain ordinary
checked entries. Short tails and wrap boundaries use the original path. Fuel,
limits and cancellation contracts remain intact. This is a general native
selection rule shared by amd64 and arm64.

Earlier broader expansion regressed serial constant-address and narrow-memory
loops. The final rule excludes pointer chasing, invariant addresses and subword
accesses. Exact generated-code comparisons confirm that the narrow, dependent
and private-memory controls remain byte-identical to base on **both architectures**
(six checked pairs). Their timing differences are not attributed to new machine
code. The [earlier experiments](performance-native-loop-unroll-2026-11-02.md)
retain their failures and regressions.

## Complete native comparison

Apple M4 Max/macOS arm64; Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1 and Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler.
Three fresh serially interleaved blocks, two warmups and seven warm calls;
1,000,000 source iterations. All results and alternate inputs pass independent
oracles. Applicable C controls use identical counts and checked outputs.
Units: ns/source iteration. All phases are retained separately.

| Workload | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core/function-calls | 0.787 | 0.800 | 0.782 | 0.784 | 0.908 | 0.789 | 0.774 |
| core/integer-loop | 1.576 | 1.581 | 1.563 | 1.566 | 1.554 | 1.549 | 1.545 |
| core/memory-load-store | 0.540 | 0.493 | 0.526 | 0.473 | 0.480 | 0.532 | 0.269 |
| memory/dependent-address | 1.873 | 1.860 | 1.843 | 1.839 | 2.063 | 2.095 | 1.414 |
| memory/float64-unaligned | 0.724 | 0.727 | 0.725 | 0.718 | 0.711 | 0.715 | 0.661 |
| memory/narrow-unaligned | 0.695 | 0.694 | 0.686 | 0.689 | 0.693 | 0.715 | 0.453 |
| runtime/simd-memory-full | 3.382 | 3.373 | 3.372 | 3.376 | 3.611 | 3.459 | — |
| runtime/simd-memory-lanes | 5.098 | 5.018 | 5.001 | 5.043 | 5.804 | 5.812 | — |
| runtime/simd-memory-widen | 1.047 | 1.053 | 1.043 | 1.046 | 1.035 | 1.036 | — |
| threads/atomic-rmw | 1.845 | 1.849 | 1.819 | 1.829 | 1.819 | 3.883 | — |
| threads/private-memory | 0.328 | 0.326 | 0.323 | 0.331 | 1.022 | 0.900 | — |
| threads/shared-memory | 0.848 | 0.842 | 0.850 | 0.839 | 1.031 | 1.030 | — |

The full-memory and lane SIMD results include small regressions; private-memory
LLGo changes 0.323→0.331 despite identical guest code. No unfavorable sample is
removed. Two interpreter controls remain unchanged within observed noise. There
are no independent A/A aliases in this final twelve-kernel run; exact machine-code
identity is a control for three cases, not a general noise calibration. This is
a shared-host observation, not a portfolio or cross-platform acceptance result.

All 31 Go semantic packages and 30 cross-package coverage gates, affected pinned
LLGo codegen/execute/VM suites, codegen race, no-cgo and vet pass. Tests exercise
all original entry PCs, overlap ordering, unsigned wrap, tails, fuel, limits and
cancellation. An initial run lacked its artifact link. Another encountered the
old 100 ms thread-fixture compilation deadline; eight Go repetitions and LLGo
pass after compiling the fixture before the execution deadline. The complete
final Go matrix passes; these earlier failures are retained.

[Lossless evidence](evidence/performance-native-word-loops-2026-11-02.json.gz)
contains all samples/phases, sources and executable/SDK hashes, six code-identity
proofs, the exact patch, measurement programs and validation logs. Native C has
fewer runtime contracts; missing VM setup is not reported as invented zero cost.

# Long native SIMD and ordinary C comparison: 2026-11-03

This separately declared campaign follows the [held vector-constant experiment](performance-vector-constants-held-2026-11-03.md).
It reuses its exact original Go/LLGo base/head binaries, without recompiling or
repeating the short interpreter campaign. The original 25.1% native widening
regression remains retained. The new observations do not replace it or establish
controlled all-platform acceptance. No held vector-constant code enters main.

## Method and independent C controls

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7` is the implementation integrated
by PR111. Held head is `ac51429fa3ac43282d48c6bd6fc494ee549e113a`. Fixed Go 1.27.3,
LLGo `983465e` with `-O3 -lto=full`, Apple M4 Max/macOS arm64. No compiler or
production-profile upgrade is part of this measurement.

The plan fixes five fresh serially interleaved process blocks, two warmups/seven
warm calls, seed 305419896, alternate seed 4294967295 with 257 iterations, and
exact baseline Go/LLGo native A/A aliases. Widening uses 100 million iterations,
scalar memory 200 million, full-vector memory 30 million and lane memory 20
million. All VM warm intervals are 82–167 ms, compared with roughly 1–2 ms in
the original short controls. Readiness, preparation, first call, warm execution,
cleanup, original outcomes and sample orders remain separate. No subtraction
or fastest-per-case baseline is used. Shared-host controls are descriptive,
without controlled affinity or a calibrated whole-portfolio parity verdict.

The ordinary C reference now adds the three SIMD memory algorithms. Standard C
arrays and alias-safe memcpy allow Clang to optimize/vectorize freely; generated
ARM64 assembly contains signed widening, vector xor/add and vector loads/stores.
It resets fixture memory within each call. These unchecked algorithms lack VM
bounds, fuel, cancellation and root/ownership contracts. They are not a C API
adapter or a complete VM equivalent. Little-endian representation matches both
supported benchmark hosts. Existing scalar algorithms are unchanged; the C
source/binary identity changes explicitly and is retained.

234 independent Python-oracle cases check all nine C algorithms, including SIMD
overlap, signed bytes and overflow. A separate ASan/UBSan C entry is checked on
96 short Python-oracle cases before computing four long-count checksums; all
ten configurations verify those checksums on every timed invocation. The
complete grid contains 180 VM/reference observations, twenty C observations,
forty alternate checks and forty independently regenerated summaries.

## Main implementation and reference outcomes

Warm ns/source iteration, lower is better. Values below all use the **base**
implementation; held candidate results are separate below. Reference versions
are wazero 1.12.0 compiler, Wasmtime 48.0.1 Cranelift and Wasmer 7.3.0 Cranelift.

| Kernel | Go | LLGo | wazero | Wasmtime | Wasmer | Ordinary C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SIMD widening | 1.024362 | 1.024143 | 1.029357 | 1.023630 | 1.025370 | 1.024030 |
| Scalar memory | 0.418333 | 0.419737 | 0.791882 | 0.470930 | 0.522650 | 0.277056 |
| SIMD full-vector memory | 3.325372 | 3.328314 | 3.711097 | 4.029033 | 3.438667 | 3.610850 |
| SIMD lane memory | 5.016121 | 4.983410 | 5.781967 | 5.823550 | 5.788250 | 4.508400 |

Widening is descriptively close to C and the native references. Full-vector
memory takes less time than these references/C. Lane memory still takes about
1.105x C, and scalar memory about 1.515x C. Native SIMD parity in these kernels
does not establish fast SIMD interpretation or Component/WASI boundaries.

## Held candidate and adverse outcomes

| Kernel | Go base → held | LLGo base → held | LLGo held/base |
| --- | ---: | ---: | ---: |
| SIMD widening | 1.024362 → 1.025058 | 1.024143 → 1.023775 | −0.04% |
| Scalar memory | 0.418333 → 0.419743 | 0.419737 → 0.419227 | −0.12% |
| SIMD full-vector memory | 3.325372 → 3.328957 | 3.328314 → 3.329667 | +0.04% |
| SIMD lane memory | 5.016121 → 4.982719 | 4.983410 → 5.016210 | +0.66% |

The earlier 25.1% widening difference does not recur here; this does not
retroactively invalidate that observation. Retain the new 0.66% LLGo lane
regression, with block ratios 1.0087/1.0106/1.0066/1.0046/1.0081. Its baseline
A/A median ratio is 1.00085. Go scalar memory also increases 0.34%, with a 1.00387
A/A ratio in the same campaign's median summary comparison for that kernel.
No performance acceptance or adoption of the held candidate follows from
favorable widening data alone.

Initial script assumptions about reference build metadata, alternate-report
hashes and C readiness caused three setup failures before sustained interleaved
timing. Their exact scripts, logs and partial validation records are preserved;
they are measurement setup errors, not runtime/compiler failures. The final
complete grid and independent verifier pass without dropping an outcome.

[Plans, source/tests, sanitizers, commands, original failures, exact source/binary
identities, assembly, all phases and independent verification](evidence/performance-native-simd-long-2026-11-03.json.gz).

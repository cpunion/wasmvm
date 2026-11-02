# Ordinary C memory mechanism: 2026-11-02

This one-kernel experiment narrows the ordinary-native gap without adding a
production optimization. Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5`
`-O3 -lto=full`, Clang 22.1.8 and Wasmtime 48.0.1 Cranelift remain separate
configurations. Engine source is `93190e993918ab50df5fc81adfd509c8a4fcf900`;
archived source/binary identities are checked before reuse. This is not the
later x86 immediate-cost candidate's performance.

The existing scalar memory algorithm receives 50 million iterations in every
configuration. Its checked seed/count oracle and alternate all-one seed/17-count
control agree. C admission retains the previously declared count limit, and the
timed source stays identical across all three flag variants. Memory reset and
VM polling/checks remain inside their respective calls; no timings are subtracted.

Three fresh process blocks serially interleave seven warm calls after two
warmups. Every call checks its checksum. Ordinary C O3 and Go native each have
an exact-binary A/A alias. Units are ns/source iteration, medians of process
medians, not pooled calls:

| Configuration | ns/iteration |
| --- | ---: |
| wasmvm Go native | 0.447541 |
| wasmvm LLGo native | 0.447864 |
| Wasmtime Cranelift | 0.468780 |
| Go native A/A | 0.446921 |
| C O3 | 0.266760 |
| C O3 A/A | 0.269740 |
| C O3, loop/SLP vectorization disabled | 0.268920 |
| C O3, vectorization and loop unrolling disabled | 0.365600 |

Disabling vectorization leaves the complete emitted C `run` assembly text
identical, SHA-256
`7d87c89cb8135d67797f4bef73f0afead4cb9314e8c68cbf2de87f17365d1c8b`.
Its observed 0.8% time difference is smaller than the C A/A displacement.
The hot memory loop contains scalar loads/stores. SIMD auto-vectorization does
not explain the optimized-C lead for this kernel.

Disabling unrolling changes that assembly and raises the C observation 37.1%.
The optimized C loop expands four iterations with induction-based address
updates and four partial sums. wasmvm also expands four, but repeats source
counter/address operations and uses one running sum. Its 0.448 versus C 0.267
is still about 1.68x. Even the C unrolling-disabled control is faster, so counting
unrolled copies alone cannot close the gap. These are code/protocol observations
and an optimization direction, not causal attribution of each nanosecond.
A previous two-accumulator wasmvm prototype regressed and remains held.

C A/A ratios range 1.0089–1.0159 and Go A/A 0.9885–1.0014. They do not meet
controlled ±0.25% calibration. Shared-host load and all 24 records/eight alternate
checks stay retained; no affinity or isolated-host performance claim is made.
Go and LLGo are nearly equal on this body, which does not explain their separate
GC/interpreter differences or justify another LLGo issue.

The first local diagnostic wrongly applied the VM readiness parser to standalone
C output. It failed before timed blocks; its source/log/partial record remain.
The corrected C oracle reads its declared JSON protocol, and all three measured
blocks pass. This is a measurement-script correction, not an engine defect.

[Lossless evidence](evidence/performance-native-c-memory-mechanism-2026-11-02.json.gz)
retains all code, flags, assemblies, phases, samples, order, oracles, A/A aliases,
source/binary identities and the unsuccessful initial invocation.

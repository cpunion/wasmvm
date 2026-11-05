# Register homes for nonresident loop values: 2026-11-05

**Full four-profile Go/LLGo CI passes; elapsed-time acceptance remains open.**
Engine source: `00bb57898295bb2994e2362cae8cdf3c85b431af`;
baseline: `67ce0eb52e5b5f7be40e5d0838f4139bc4bc3043` (engine `2f8e44f`).
The [preceding event](performance-loop-scratch-registers-hosted-2026-11-05.md)
is retained and does not measure this change.

The bounded loop interference graph previously included only resident slots.
Equal uncached copies could not join a register group, even when their complete
lifetime and final frame word matched a resident value. The graph now includes
every referenced slot, in deterministic order. A group must contain an existing
home; memory-only groups acquire no new registers.

When the sole live-in is uncached, the entry plan loads that original frame word
into the resident member's home. Overlapping versions and every final frame
word still constrain interference. Displaced values reach the canonical frame
before reloads. Full plans include every new alias for poll/return spills;
ordinary cold PCs retain their checked mapping. Selectors still split at
shared-home writes. All three banks use one analysis, with bounded work and
conservative rejection. ABI, physical inventory, ISA, compiler flags, feature
scope, fuel, cancellation and complete-frame contracts remain unchanged.

## Execution evidence

The complete codegen package passes Go 1.27.3 and LLGo `3571c25`
(`-O3 -lto=full`) on macOS arm64. Go codegen coverage is **96.43% (7541/7820)**;
this is one-package coverage, not the preceding ten-package cross-coverage.
Exact codegen `go vet` also passes.

New cases cover uncached live-in selection for integers, F32/F64, V128 and
references, and reject groups with no resident home. Generated execution models
include 120 extra equal final words, exceeding either target's register bank.
They compare every original cold PC, complete frame words, valid/OOB memory and
forced private-loop polls with the independent precise tier. Both targets are
compiled; the host architecture executes.

An additional real-guest check executes all unchanged 92 P2 and 60 P3 upstream
consumers on interpreter/native under each host compiler: **608
compiler/backend/guest checks** on macOS arm64. The manifest audit verifies every
expected name. Each compiler reports 388 passing test nodes including parents,
interactions and negative controls; those nodes are not 388 distinct guests.
There are no test skips. Declared manifest exclusions remain explicit and
unchanged. This does not establish broader upstream-world or performance
acceptance.

## Static target output

The x86 diagnostic uses the production module compile/link path, exact producer
bytes and feature mask 7. It dumps output before native publication and does not
execute x86 code on the arm64 host. All spans and raw metadata remain retained.

| Compression loop | Instructions | Frame references | Polls | Complete function bytes |
| --- | ---: | ---: | ---: | ---: |
| Clang | 92 → 92 | 24 → 24 | 1 → 1 | 10956 → 10956 |
| Rust | 88 → 85 | 20 → 18 | 1 → 1 | 10843 → 10866 |

Both schedule loops stay at 47 instructions, zero frame references and one poll.
The Rust function grows by 23 bytes despite its shorter hot loop; entry/exit
materialization is retained. Core integer/memory modules stay at 2188/4284 bytes.
These counts do not establish elapsed improvement or competitor parity.

[Full CI 41992896773](https://github.com/cpunion/wasmvm/actions/runs/41992896773)
passes all 21 required jobs, with one intentional optional skip. The independent
artifact audit verifies 240 conformance reports across all four compiler/platform
combinations, 31 test packages in seven LLGo shards per platform and 30 package
coverage gates. Minimum coverage is **90.88%**. The longest job takes **504 s
(8m24s)**; queue time is separate. All Stack Switching reports force collection.
These results apply to `00bb578`, not the later invariant-literal source.

[Comparison 41992901763](https://github.com/cpunion/wasmvm/actions/runs/41992901763)
now completes successfully on Linux amd64: one shared build and all five groups,
22 workloads, 362 alternative-correctness checks and 12 explicit unsupported GC
reference configurations. The independent audit recomputes every A/A interval,
checks the complete phase records and five-process/seven-call inventory, and
matches both compiler builds against exact archived engine hashes. Reference
versions and release flags are pinned. The longest measurement job is **493 s
(8m13s)**; the shared build is 330 s. Neither run was resubmitted.

**10/22 workloads resolve all four diagnostic A/A controls.** The other 12
remain inconclusive; hosted conditions do not establish controlled milestone
acceptance even for the resolved rows. The following are complete warm-call
ratios for the `00bb578` source; lower is better. They do not measure later
invariant-literal, typed-preparation or CFG value-allocation changes.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.161× | 1.324× | 5.336× | 2.472× | 1.574× | 2.454× | 1.433× | unavailable | inconclusive |
| core/integer-loop | 0.996× | 0.836× | 1.066× | 1.063× | 1.060× | 0.807× | 0.956× | 1.067× | resolved, advisory |
| core/function-calls | 1.000× | 0.844× | 0.708× | 0.943× | 1.009× | 0.778× | 1.089× | 0.822× | resolved, advisory |
| core/memory-load-store | 1.009× | 0.840× | 0.644× | 1.066× | 1.040× | 0.915× | 1.002× | 1.419× | resolved, advisory |
| memory/narrow-unaligned | 0.995× | 0.979× | 0.861× | 1.297× | 1.295× | 0.547× | 0.782× | 2.053× | resolved, advisory |
| memory/dependent-address | 0.992× | 0.876× | 1.017× | 1.350× | 1.346× | 0.505× | 0.351× | 1.576× | resolved, advisory |
| memory/float64-unaligned | 0.998× | 0.947× | 0.688× | 1.172× | 1.166× | 0.609× | 0.413× | 1.190× | inconclusive |
| runtime/simd-memory-full | 0.995× | 0.926× | 0.891× | 1.107× | 1.069× | 0.512× | 0.901× | 0.925× | resolved, advisory |
| runtime/simd-memory-widen | 0.998× | 1.026× | 0.759× | 1.716× | 1.715× | 0.487× | 0.949× | 0.441× | resolved, advisory |
| runtime/simd-memory-lanes | 0.996× | 1.125× | 0.955× | 1.151× | 1.159× | 0.482× | 1.130× | 1.341× | resolved, advisory |
| gc/i31-roundtrip | 0.997× | 0.895× | unsupported | 0.949× | unsupported | unsupported | 0.398× | unavailable | inconclusive |
| gc/struct-access | 0.896× | 0.841× | unsupported | 1.997× | unsupported | unsupported | 0.364× | unavailable | inconclusive |
| gc/struct-allocation | 1.242× | 1.420× | unsupported | 0.309× | unsupported | unsupported | 1.406× | unavailable | inconclusive |
| gc/struct-indirect-access | 1.006× | 0.871× | unsupported | 2.770× | unsupported | unsupported | 0.747× | unavailable | inconclusive |
| threads/private-memory | 1.135× | 1.082× | 1.515× | 2.692× | 1.765× | 0.656× | 0.616× | unavailable | inconclusive |
| threads/shared-memory | 1.032× | 0.651× | 1.387× | 2.526× | 1.682× | 0.672× | 0.635× | unavailable | inconclusive |
| threads/atomic-rmw | 1.060× | 0.687× | 1.651× | 1.578× | 1.222× | 0.549× | 0.623× | unavailable | inconclusive |
| threads/notify-empty | 1.021× | 0.869× | 0.091× | 0.393× | 0.292× | 0.367× | 0.487× | unavailable | inconclusive |
| threads/wait-mismatch | 1.043× | 0.676× | 0.073× | 0.294× | 0.623× | 0.407× | 0.491× | unavailable | inconclusive |
| threads/wait-timeout | 0.938× | 0.660× | 0.007× | 0.054× | 0.027× | 0.095× | 0.001× | unavailable | inconclusive |
| applications/sha256-clang | 0.992× | 0.938× | 0.816× | 1.551× | 1.509× | 0.549× | 0.861× | 1.862× | resolved, advisory |
| applications/sha256-rust | 0.986× | 0.942× | 0.770× | 1.541× | 1.447× | 0.552× | 0.813× | 1.835× | resolved, advisory |

SHA-256 is observed at 0.77–0.82× wazero JIT, 1.54–1.55× Wasmtime and
1.84–1.86× native C. LLGo/Go native is 0.986–0.992× for these two workloads.
The native compiler still trails the stronger application compilers despite
smaller hot-loop instruction counts. Cheap wait/notify probes are not general
thread scheduler or real-contention throughput results. Tiny API-call and GC
rows that fail calibration cannot establish superiority or regression.

Contribution review and final-head CI, controlled elapsed acceptance,
rematerialization/frame reconstruction and broader WASI-world qualification
remain open. Integration stays on this branch under the one-active-contribution
policy while LLGo #2991 remains open. LLGo #2983 stays with its existing contributor.

The [evidence index](data/loop-nonresident-values-2026-11-05.json) records exact
sources, commands, compiler hashes, coverage, manifest audits and raw file hashes.
Local archive: `artifacts/loop-nonresident-values-20261105/evidence.tar.gz`,
440324 bytes, SHA-256
`0d5e55345c6c478d20061c48827922415e641136230ad1f669c67ad7d5dacfab`.

The subsequent CI bundle is separate:
`artifacts/loop-nonresident-values-20261105/ci-evidence.tar.gz`, 5810541 bytes,
267 files, SHA-256
`a07c8a20fb6fe1415163f42b98e099b49737da624c7ec146df1197f2f1c993b3`.
The initial archive remains immutable.

The hosted closeout is a separate immutable bundle: `artifacts/loop-nonresident-values-20261105/hosted-evidence.tar.gz`, 2085776 bytes, 15 files; SHA-256 `5edbe676fed3ef3446cac4b16e6fda1e0dc3dc58830a148d73849b256569edd9`.

# Register homes for nonresident loop values: 2026-11-05

**Local dual-compiler execution passes; full CI and elapsed-time acceptance
remain open.** Engine source: `00bb57898295bb2994e2362cae8cdf3c85b431af`;
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
and [comparison 41992901763](https://github.com/cpunion/wasmvm/actions/runs/41992901763)
are incomplete at the recorded observation, with jobs queued or running. Neither
workflow has been resubmitted. The comparison retains all five groups, both
compilers, competitor/native rows, separate phases and adverse observations.
Final-head CI, review, elapsed-time/controlled acceptance, general value
allocation/splitting/rematerialization and broader WASI-world qualification remain
open. Integration stays on this branch under the one-active-contribution policy
while LLGo #2991 remains open. LLGo #2983 stays with its existing contributor.

The [evidence index](data/loop-nonresident-values-2026-11-05.json) records exact
sources, commands, compiler hashes, coverage, manifest audits and raw file hashes.
Local archive: `artifacts/loop-nonresident-values-20261105/evidence.tar.gz`,
440324 bytes, SHA-256
`0d5e55345c6c478d20061c48827922415e641136230ad1f669c67ad7d5dacfab`.

# Mixed scalar forwarding: 2026-11-06

**Local qualification passes; final x86 execution and elapsed acceptance remain
required.** Engine `96ea7fb` uses the last good frame-operand engine `5099a76`
as its measured baseline. The raw-FP allocation experiment `0e32719` is
**rejected and reverted** in `4a744df`; its complete adverse evidence remains
separate below.

## Shared selection

Copy/result forwarding previously accepted fully resident integer operations.
It now also accepts the existing mixed GPR/frame forms for i32/i64
add/subtract/multiply/AND/OR/XOR. Eligibility and emission share one check and
retain the existing binary opcode inventory. Both copy and result forwarding
use the common scalar selector. There is no new register bank, ISA requirement,
execution mode or benchmark-specific path.

Original cold PCs, independently observable temporary/local words, deferred
reconstruction boundaries, exact fuel and polling keep their existing
contracts. Unsupported operations, the precise tier and cached typed FP/vector
inputs reject the mixed form before emitting code.

## Qualification and actual code

Complete codegen/nativeexec/public-VM suites pass under Go 1.27.3 and LLGo
`3571c25` with `-O3 -lto=full`; vet passes. Exact Go package coverage is
**codegen 97.69% (8460/8660)** and **nativeexec 92.54% (459/496)**.
This does not measure specification coverage.

New tests select both forwarding forms for all twelve arithmetic opcodes. A
pressured ordinary function checks **105 cold entries × four input patterns ×
twelve opcodes**, every Lo/Hi word, continuation, payload, fuel and budget.
Existing private-loop tests retain forced polls, original PCs and late traps.
Both target functions compile; the ARM64 counterpart executes locally. Newly
selected x86 forwarding still needs actual Linux execution under both compilers.
The existing integer encoders retain their independent LLVM byte witnesses.

All 44 fresh ARM64 publication witnesses pass their checksum oracles. Across
22 fixtures, baseline/head Go/LLGo images and entry lists are identical. No
additional local warm campaign is collected for unchanged guest code.

Production x86 compilation/linking retains feature mask 7 and stops before
publication on this host. Hot spans must contain a known private poll.

| x86 function | Instructions | Frame references | Compression bytes | Wrapper bytes |
| --- | ---: | ---: | ---: | ---: |
| Clang SHA | 70 → 69 | 20 → 20 | 10864 → 10859 | 15868 → 15892 |
| Rust SHA | 67 → 65 | 18 → 18 | 10746 → 10738 | 16178 → 16202 |

Both schedule loops retain 43 instructions, zero frame references and one poll;
compression polling also remains one per iteration. Wrapper growth and complete
preparation/first/warm phases stay visible. Static reductions establish no
elapsed gain. [Linux event 42093658819](https://github.com/cpunion/wasmvm/actions/runs/42093658819)
compares this exact source with `5099a76` over all 22 original-count cases,
Go/LLGo, competitors and applicable native C. Its result remains required.

## Rejected raw-FP strategy

The completed [raw-source Linux event 42088602463](https://github.com/cpunion/wasmvm/actions/runs/42088602463)
passes six build/measurement jobs, all 22 cases, 362 alternate checksum checks
and 12 expected reference-GC rejections. Nine cases resolve all four advisory
controls; both SHA cases do. Ratios compare elapsed time; lower is faster.

| Rejected Linux workload | Go head/base | LLGo head/base | LLGo / Wasmtime | / native C |
| --- | ---: | ---: | ---: | ---: |
| Clang SHA | 1.234 | 1.217 | 1.630 | 1.957 |
| Rust SHA | 1.236 | 1.218 | 1.604 | 1.911 |

That strategy removes all 20/18 compression frame references but regresses
elapsed execution substantially under both host compilers. Local ARM64 Clang
also observes LLGo head/base 1.015. It is withdrawn from production; the new
selector uses the established GPR/frame strategy. The regression is sufficient
to reject the allocation policy; isolating its microarchitectural cause needs
separate evidence. All raw observations, phases, controls, failed tests and
semantic witnesses remain in the [rejected experiment record](performance-raw-integer-registers-2026-11-06.md).
There is no favorable rerun, ISA/flag/alignment sweep or cross-source pooling.

[CI 42091050950](https://github.com/cpunion/wasmvm/actions/runs/42091050950)
belongs to the rejected raw source and remains a separate validation event.
Final corrected-head CI follows its terminal outcome. Final target semantics,
review/integration, broader world/release conformance, preparation costs and
controlled competitive/native acceptance remain open. LLGo #2983 stays with its
existing contributor; external LLGo #2991 remains the sole active contribution
PR and is outside this repository's merge authority.

The [machine-readable record](data/mixed-scalar-forwarding-2026-11-06.json) pins
all source/compiler identities, static measurements and archive checksums.

Immutable local archive `artifacts/loop-mixed-forwarding-20261106/evidence.tar.gz`: 1776913 bytes, 246 files;
SHA-256 `fff33ac681e8a72a2d65ed8d1eefc0d378ae6839b33fe200382758c57eec2c84`.

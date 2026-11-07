# Mixed scalar forwarding: 2026-11-06

**Final four-profile semantic CI passes; elapsed results are advisory.** Engine `96ea7fb` uses the last good frame-operand engine `5099a76`
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
Both target functions compile and execute on their actual targets under Go and
LLGo in completed CI 42096193934.
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
Go/LLGo, competitors and applicable native C. Its completed result is advisory: 22 cases, 362 alternate checksum checks,
12 expected reference-GC rejections and eight cases resolving all four A/A
controls. Both SHA cases resolve. Lower ratios mean less elapsed time.

| Corrected Linux workload | Go head/base | LLGo head/base | LLGo / Go | / wazero JIT | / Wasmtime | / native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Clang SHA | 0.983 | 0.978 | 0.998 | 0.696 | 1.318 | 1.574 |
| Rust SHA | 1.011 | 1.011 | 0.997 | 0.671 | 1.334 | 1.585 |

The Rust loss remains recorded. This selector does not establish controlled
performance acceptance or close the Wasmtime/native gap. All preparation, first,
warm, shutdown and process intervals remain separate in the completed archive.
The shared build takes 329 seconds; the longest measurement group takes 496
seconds. No timing phases or source events are pooled.

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
[Corrected-head CI 42096193934](https://github.com/cpunion/wasmvm/actions/runs/42096193934)
passes all 21 required jobs. Artifact verification checks 240 pinned execution
reports across all four profiles, all 31 LLGo packages once per platform, and
30 coverage gates (minimum 90.75%). The longest job is 646 seconds, so the
5–10 minute target is not fully met. Integration/review, broader world/release
conformance, preparation costs and controlled competitive/native acceptance
remain open. LLGo #2983 stays with its existing contributor; upstream has merged
LLGo #2991. The compiler pin remains unchanged.

The [machine-readable record](data/mixed-scalar-forwarding-2026-11-06.json) pins
all source/compiler identities, static measurements and archive checksums.

Immutable local archive `artifacts/loop-mixed-forwarding-20261106/evidence.tar.gz`: 1776913 bytes, 246 files;
SHA-256 `fff33ac681e8a72a2d65ed8d1eefc0d378ae6839b33fe200382758c57eec2c84`.

Completed Linux and CI archive: `artifacts/loop-mixed-forwarding-20261106/completed-evidence.tar.gz`, 8026921 bytes, 284 files; SHA-256 `d2ffad7b81ae80c63a1ea5e796cb981cf46fc7de61913c29d4869f15c02ddf3e`. The earlier local archive is unchanged.

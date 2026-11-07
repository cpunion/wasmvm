# Private-loop comparison immediates: 2026-11-06

**Local Go/LLGo semantics pass; controlled performance acceptance remains open.**
Engine `4054f1b96a707fa6f5334221574b8e393f8effc0`; baseline engine
`15567dfe8068f5e279c61c67f9bd06d37ac0176e`, documentation head `faff6d6`.
The [preceding predicate snapshot event](performance-loop-predicates-2026-11-06.md)
retains its separate 18-case results and preparation gap.

## Shared mechanism

A proved private loop may remove the integer literal immediately before its
final comparison and compare its other operand directly with an immediate.
Both the comparison result and literal entry versions must be unread. Shared
literal homes, existing hoists, copy-reconstruction anchors, aliased comparison
inputs, header polls and unencodable constants retain the existing path. The
same native comparison encoder handles ordinary eqz and the immediate form,
including x86 signed imm8/imm32 and ARM64 optional imm12 shift by 12.

The observable-frame snapshot restores the literal before copies and the
original comparison on cancellation. On ordinary exit, a literal/result alias
needs only the final boolean; separate slots retain the literal too. Every
original cold PC and precise-fuel instruction remains. The transform introduces
no unchecked execution or alternative runtime.

## Rejected interaction and correction

Initial source `b92b249` passed complete local semantic tests but emitted one
extra x86 instruction in each SHA loop; ARM64 hot counts were unchanged. Removing
the late literal caused observation scans to skip its logical definition.
Earlier immediate/arithmetic/local fusion then preserved two unnecessary moves.
This is a compiler dataflow interaction, not evidence that an immediate is faster.
All initial output, source, tests and disassembly remain retained. No warm timing
was collected for this rejected source.

The correction retains logical uses and definitions of snapshot-recoverable
instructions in observation scans. Read-before-write ordering, copy aliases,
mapping boundaries, guards, traps and polls retain their existing bounds.
Physical selection still stops at omitted instructions. Earlier temporary words
may disappear only when the snapshot proves they cannot be observed.

## Semantic and actual-output evidence

Complete codegen/nativeexec/public VM suites pass with Go 1.27.3 and LLGo
`3571c25`, using `-O3 -lto=full`; vet passes. The suites cross-cover **codegen
97.67% (8396/8596)** and **nativeexec 92.54% (459/496)**. These are exact Go
package statement measurements, not coverage of a specification or competitor.

Independent full-word models now cover every binary integer comparison, both
backedge directions, literal/result aliases and separate operands, register
pressure, initial reads, every original cold PC, first generated backedge poll
and resumption from its published branch snapshot. Immediate boundary cases
include sign extension, i32 normalization, shifted ARM64 limits and fallback.
Negative tests retain shared homes, copy anchors, hoists and mismatched widths;
an unencodable immediate emits no partial code. Existing precise-reference
tests also compare all words, memory, late traps and shared-memory execution.
An initial fixture insertion accidentally moved its backedge past the intended
prefix; the corrected fixture keeps that instruction inside the body. Its
failure log is retained.

All 44 fresh candidate publication witnesses pass checksum oracles, and all 22
candidate Go/LLGo images and entry lists match. The 22 baseline witnesses are
reused after fixture/source/image verification. Ten actual ARM64 images differ;
the four GC images remain unchanged.

| SHA loop | ARM64 instructions | x86 instructions | x86 frame references | Polls |
| --- | ---: | ---: | ---: | ---: |
| Clang schedule | 28 → 27 | 44 → 43 | 0 → 0 | 1 → 1 |
| Clang compression | 46 → 45 | 85 → 84 | 20 → 20 | 1 → 1 |
| Rust schedule | 28 → 27 | 44 → 43 | 0 → 0 | 1 → 1 |
| Rust compression | 40 → 39 | 82 → 81 | 18 → 18 | 1 → 1 |

Complete x86 compression functions grow by seven bytes each: Clang
11000 → 11007, Rust 10882 → 10889. Complete ARM64 SHA module sizes remain
101332 and 101500 bytes; cold restoration offsets the shorter hot code. Other
changed ARM64 modules shrink by four bytes. ARM64 Clang compression retains one
frame reference, Rust zero. x86 uses unchanged feature mask 7 and production
module compilation/linking, stopping before native publication on this ARM64
host. Static output establishes no x86 execution or elapsed gain.

## Measurement contract

Before timing, select every group containing a changed actual ARM64 image:
applications, core, memory/SIMD and threads, 18 cases. Retain GC's existing
evidence because all four images match. Five fresh process blocks, two warmups
and seven interleaved samples retain original counts/seeds, both host compilers,
own interpreters, all references and applicable native-C cases. Baseline and
reference binaries are reused only after engine, harness, toolchain/profile,
executable and library verification. Candidate executables are fresh.

Each case retains four identical-binary A/A controls. The predeclared marginal
95% ±2.5% diagnostic calibration does not replace controlled acceptance. Local
macOS has no CPU affinity or controlled physical frequency/neighbors. Elapsed
ratios are lower-is-faster; failed calibration, adverse observations and separate
lifecycle phases remain visible. No phase subtraction, cross-source/target
pooling, favorable retry or flag/ISA/alignment sweep is used.

## Local elapsed results

All four groups complete: **18 fixtures, 306 alternate checks and 1,585 measured
process records**. **11/18 cases resolve all four A/A controls.** These are
median warm elapsed ratios; lower is faster. Diagnostic calibration does not
establish controlled acceptance or statistical significance for small changes.

| Workload | LLGo head/base | LLGo / Go | / wazero | / Wasmtime | / Native C | All four A/A resolve |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.026 | 1.191 | 7.353 | 3.857 | — | No |
| core/integer-loop | 1.005 | 0.988 | 0.919 | 1.015 | 1.009 | Yes |
| core/function-calls | 0.989 | 0.973 | 0.548 | 0.862 | 1.003 | Yes |
| core/memory-load-store | 1.005 | 0.975 | 0.506 | 0.918 | 1.616 | Yes |
| threads/private-memory | 1.082 | 0.879 | 1.276 | 1.199 | — | No |
| threads/shared-memory | 1.080 | 0.824 | 1.261 | 1.220 | — | No |
| threads/atomic-rmw | 0.939 | 0.889 | 1.036 | 1.125 | — | No |
| threads/notify-empty | 0.924 | 0.820 | 0.062 | 0.187 | — | No |
| threads/wait-mismatch | 1.026 | 0.742 | 0.059 | 0.163 | — | No |
| threads/wait-timeout | 0.992 | 0.559 | 0.006 | 0.036 | — | No |
| memory/narrow-unaligned | 0.991 | 0.988 | 0.624 | 0.952 | 1.422 | Yes |
| memory/dependent-address | 1.008 | 0.988 | 0.586 | 0.888 | 1.306 | Yes |
| memory/float64-unaligned | 0.990 | 0.975 | 0.845 | 1.025 | 1.104 | Yes |
| runtime/simd-memory-full | 0.988 | 0.986 | 0.891 | 0.915 | 0.946 | Yes |
| runtime/simd-memory-widen | 1.005 | 0.986 | 0.919 | 1.004 | 1.008 | Yes |
| runtime/simd-memory-lanes | 1.004 | 1.002 | 0.827 | 0.865 | 1.129 | Yes |
| applications/sha256-clang | 0.985 | 0.977 | 0.669 | 1.043 | 1.198 | Yes |
| applications/sha256-rust | 1.012 | 0.989 | 0.649 | 1.134 | 1.292 | Yes |

Clang SHA records a 1.5% lower LLGo median; Rust records a 1.2% higher median,
despite fewer hot instructions. Both retain their calibration records. Dependent
integer memory also records a 0.8% higher median. Both plain thread-memory cases
record about 8% higher medians with unresolved controls; unchanged atomic/helper
cases move too. These adverse and unresolved values remain open. No source is
retested to obtain a favorable result.

Preparation medians are reported separately from warm/first calls:

| Producer | Go baseline → head | LLGo baseline → head |
| --- | ---: | ---: | ---: |
| Clang | 2.308 → 2.248 ms | 4.583 → 4.395 ms |
| Rust | 2.322 → 2.268 ms | 4.682 → 4.760 ms |

The current LLGo/Go preparation gap remains about 1.96–2.10×. These medians have
no preparation-specific A/A gate; warm calibration does not calibrate this phase
or establish an upstream compiler bug. Complete runtime phases, both own
interpreters, Wasmer, WasmEdge interpreter and every process/order remain in
the raw record. No elapsed phase is subtracted.

## Preceding Linux event and CI outcome

The preceding source’s [Linux run 42051975324](https://github.com/cpunion/wasmvm/actions/runs/42051975324)
completes: 22 cases, 362 alternate checks, 12 explicit reference-GC capability
rejections and all six build/measurement jobs pass. Seven cases resolve all
four controls. Source `993cbb3` against `3355a8d` records a 4.3% lower Clang SHA
LLGo median; Rust is effectively unchanged with unresolved controls. These
are separate source/target results, not pooled with this ARM64 event. The
[frame-copy report](performance-loop-frame-copies-2026-11-05.md) retains the full
competitor matrix.

The [preceding CI](https://github.com/cpunion/wasmvm/actions/runs/42057704970)
retains 240 matching, passing conformance reports across four compiler/platform
profiles, 31 packages in seven LLGo shards per platform, all pinned manifests
and compiler identities, Stack Switching forced collection, and 30 coverage
gates with minimum 90.77%. **The overall workflow is cancelled, not green.**
Twenty required jobs pass. VM shard execution passes in 394s, then artifact
finalization receives HTML instead of JSON on three two-minute attempts. The
26,623-byte upload eventually finalizes after 378s; the job exceeds its 12-minute
limit and finishes cancelled after 839s. The longest successful job is 566s.

The required shard upload now has a two-minute step limit. Evidence remains
mandatory and semantic scope/release flags are unchanged; an upload timeout
still fails validation. Five shard-tooling tests and the 31-package matrix
validation pass. New final-head CI remains required; the cancelled older run is
not presented as a successful merge gate or rerun solely to replace its outcome.

The subsequent [Linux event 42070246689](https://github.com/cpunion/wasmvm/actions/runs/42070246689)
now completes for exact source `4054f1b` against `993cbb3`, covering the cumulative
predicate and literal changes: all six build/measurement jobs pass, 22 cases,
362 alternate checks and 12 expected GC capability rejections. Seven cases
resolve all four controls. SHA LLGo warm head/base medians are 0.956/0.941
(Clang/Rust), with unresolved controls for both; these advisory results remain
separate from the local ARM64 event. The full matrix follows. Lower is faster.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 2.056× | 2.916× | 11.489× | 4.646× | 3.867× | 5.041× | 3.481× | unavailable | inconclusive |
| core/integer-loop | 0.997× | 0.846× | 1.010× | 1.007× | 1.005× | 0.568× | 0.658× | 1.011× | inconclusive |
| core/function-calls | 0.995× | 0.885× | 0.435× | 0.440× | 0.438× | 0.649× | 0.905× | 0.822× | inconclusive |
| core/memory-load-store | 0.991× | 0.887× | 0.664× | 1.050× | 1.052× | 0.508× | 0.779× | 1.421× | resolved, advisory |
| memory/narrow-unaligned | 0.997× | 0.945× | 0.849× | 1.275× | 1.272× | 0.485× | 0.679× | 2.019× | resolved, advisory |
| memory/dependent-address | 0.987× | 0.841× | 1.006× | 1.332× | 1.332× | 0.416× | 0.353× | 1.561× | resolved, advisory |
| memory/float64-unaligned | 0.985× | 0.926× | 0.688× | 1.169× | 1.161× | 0.524× | 0.422× | 1.177× | resolved, advisory |
| runtime/simd-memory-full | 0.996× | 0.906× | 0.873× | 1.084× | 1.046× | 0.507× | 0.982× | 0.905× | resolved, advisory |
| runtime/simd-memory-widen | 1.007× | 1.024× | 0.679× | 1.539× | 1.557× | 0.497× | 0.577× | 0.398× | resolved, advisory |
| runtime/simd-memory-lanes | 0.996× | 1.050× | 0.947× | 1.144× | 1.150× | 0.450× | 1.122× | 1.331× | resolved, advisory |
| gc/i31-roundtrip | 1.015× | 0.891× | unsupported | 0.923× | unsupported | unsupported | 0.797× | unavailable | inconclusive |
| gc/struct-access | 1.097× | 0.899× | unsupported | 1.584× | unsupported | unsupported | 0.732× | unavailable | inconclusive |
| gc/struct-allocation | 1.289× | 1.825× | unsupported | 0.324× | unsupported | unsupported | 1.395× | unavailable | inconclusive |
| gc/struct-indirect-access | 1.064× | 0.945× | unsupported | 1.489× | unsupported | unsupported | 0.762× | unavailable | inconclusive |
| threads/private-memory | 1.019× | 1.072× | 0.831× | 2.489× | 2.141× | 0.510× | 0.534× | unavailable | inconclusive |
| threads/shared-memory | 1.072× | 0.613× | 0.807× | 2.743× | 2.058× | 0.538× | 0.565× | unavailable | inconclusive |
| threads/atomic-rmw | 1.269× | 0.686× | 1.746× | 1.583× | 1.392× | 0.861× | 0.558× | unavailable | inconclusive |
| threads/notify-empty | 1.025× | 0.950× | 0.077× | 0.307× | 0.223× | 0.781× | 0.464× | unavailable | inconclusive |
| threads/wait-mismatch | 0.964× | 0.670× | 0.052× | 0.209× | 0.545× | 0.440× | 0.459× | unavailable | inconclusive |
| threads/wait-timeout | 0.864× | 0.653× | 0.005× | 0.046× | 0.023× | 0.095× | 0.001× | unavailable | inconclusive |
| applications/sha256-clang | 0.980× | 0.874× | 0.747× | 1.416× | 1.376× | 0.496× | 0.782× | 1.702× | inconclusive |
| applications/sha256-rust | 0.988× | 0.881× | 0.724× | 1.447× | 1.355× | 0.493× | 0.818× | 1.723× | inconclusive |

[Final CI 42073348453](https://github.com/cpunion/wasmvm/actions/runs/42073348453)
also completes for `ab377e7`/engine `4054f1b`: all 21 required jobs pass, with one
intentional optional skip. The audit verifies 240 matching reports across four
profiles, pinned manifests/compiler identities, forced Stack Switching
collection, 31 packages in seven LLGo shards per platform and 30 package coverage
gates (minimum 90.84%). The longest job is 616s (10m16s), so the desired ten-minute
budget is still slightly exceeded. Its required evidence upload takes two
seconds; compiler setup and semantic execution take 117s and 461s. The cancelled
earlier CI remains retained. This completed CI validates the immediate-selection
source, not the subsequent frame-operand source.

Contribution review/integration, actual new-source x86 execution, general
rematerialization, expanded world/release conformance and controlled performance
acceptance remain open. The [frame-operand follow-up](performance-frame-operands-2026-11-06.md)
records the next source without pooling evidence.

## Evidence

The compact [machine-readable record](data/loop-compare-immediates-2026-11-06.json)
contains source/profile identities and all local case summaries. The immutable
archive retains all initial rejected and corrected code, tests, diagnostic
patches, images/disassembly, raw process records, full lifecycle phases,
alternative checks and sample orders.

Archive `artifacts/loop-compare-immediates-20261106/evidence.tar.gz`: 6362658 bytes, 417 files; SHA-256
`c948e755241e4585c686b3dc7a3314f40ad274be88f7780881addc3c8ab20470`.

Preceding Linux archive `artifacts/loop-frame-copies-20261105/hosted-evidence-current.tar.gz`: 2095193 bytes; SHA-256
`841890717284818cec46c8f90f2d476ca71df69817e61be5bb0517e074a644f4`.

Preceding cancelled-CI archive `artifacts/loop-compare-immediates-20261106/ci-eab-evidence.tar.gz`: 6115008 bytes; SHA-256
`4f6df9a3797258f6ce4ce1815e017de6078e351256875418502013da0d807ee6`.

Completed Linux archive: `artifacts/loop-compare-immediates-20261106/hosted-evidence-current.tar.gz`; SHA-256
`16ff5e70fe005a9ca32e619a5d4e4f690056b174bcd46494fb77a18960b408a6`.

Final CI archive: `artifacts/loop-frame-operands-20261106/ci-ab-evidence.tar.gz`; SHA-256
`b172110acb478326e1cf32f475c11f5824503fee9ef2e453d639d822e779750b`.

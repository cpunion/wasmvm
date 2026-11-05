# Integer arithmetic with frame operands: 2026-11-06

**Local Go/LLGo suites and reference encodings pass. Actual new x86 execution
and elapsed acceptance remain required.** Engine
`5099a76ddad5910dd2cfff195ca8a4bca9be9cf0`; baseline engine
`4054f1b96a707fa6f5334221574b8e393f8effc0`, documentation head `ab377e7`.
The [preceding comparison-immediate event](performance-loop-compare-immediates-2026-11-06.md)
retains its separate ARM64 and completed Linux measurements.

## Shared instruction selection

The former fallback moves both integer inputs through RAX/RCX whenever a
destination or input has no ordinary GPR home. The new x86 path uses a direct
VM-frame operand for add/subtract/multiply/AND/OR/XOR, for i32 and i64. Register
and frame forms share the existing binary opcode inventory. Fully resident
operations keep their existing encoding. A missing destination uses RAX and the
existing slot store. Commutative forms may exchange inputs; subtraction uses a
separate accumulator when the destination aliases its right input.

Read aliases resolve before selection. A cached FP/vector input cannot be read
through stale frame backing and retains the existing path. Unsupported
instructions and the precise tier emit nothing before falling back. All inputs
are consumed before a destination write; i32 results clear upper Lo bits and
each slot retains its own raw Hi word. These are execution-owned, validated VM
frame reads. Guest linear-memory checks/proofs, original cold PCs, roots,
polling, exact fuel and resource limits retain their existing contracts. No new
register bank, runtime path or reference engine is introduced.

## Semantic and encoding evidence

Complete codegen/nativeexec/public VM suites pass on macOS ARM64 under Go
1.27.3 and LLGo `3571c25` (`-O3 -lto=full`); vet passes. They cross-cover
**codegen 97.69% (8449/8649)** and **nativeexec 92.54% (459/496)**. These are
exact Go package statement measurements, not specification coverage or evidence
that x86 code executed locally.

The new test constructs **684 complete x86 kernels** covering all twelve binary
opcode forms, source/destination residency combinations, destination aliases
and remapped copy inputs. Each numeric execution tests 64 operand pairs and
every raw Lo/Hi word, with unchanged fuel/budget and a known continuation.
On this ARM64 host the kernels are emitted only; Linux amd64 CI must execute
them under both host compilers before integration. Negative cases cover
metering, the other architecture, unsupported operations, absent inputs and
cached FP/vector sources, including no partial emission on fallback.

A separate diagnostic compares **588 individual instruction encodings** with
LLVM 22.1.8's assembler: MOV and the six arithmetic/logical forms, both widths,
low/high registers and multiple full frame displacements. All bytes match.
Go/LLGo also emit identical bytes for all 684 complete kernels. The diagnostic
adds one test file to an archived source snapshot; its full source hashes,
compiler identities, flags and outputs remain recorded. Byte equality and
assembler witnesses are separate from actual CPU execution.

## Actual generated code

All 44 fresh ARM64 publication witnesses pass their checksum oracles. Candidate
Go/LLGo images and entry lists agree across all 22 fixtures. Baseline witnesses
are reused only after source, fixture and image verification. **All 22 ARM64
baseline/candidate images and entry lists match**, so no new sustained ARM64
warm campaign is collected.

Production x86 compilation/linking uses unchanged feature mask 7 and stops
before publication on this ARM64 host. The compression functions select 35
mixed integer operations each, including their cold bodies.

| x86 SHA loop | Hot instructions | Frame references | Polls | Complete function bytes |
| --- | ---: | ---: | ---: | ---: |
| Clang compression | 84 → 70 | 20 → 20 | 1 → 1 | 11007 → 10864 |
| Rust compression | 81 → 67 | 18 → 18 | 1 → 1 | 10889 → 10746 |

Both schedule loops remain at 43 instructions, zero frame references and one
poll. The change removes fourteen hot instructions from each compression
body and 143 bytes from each complete function. **It does not remove the
remaining 20/18 frame accesses.** Static instruction reductions establish no
elapsed gain; register pressure and raw-frame recovery remain further work.

## Source-pinned target comparison

The new [Linux event 42078773487](https://github.com/cpunion/wasmvm/actions/runs/42078773487)
compares exact source `5099a76` against `4054f1b`, with one shared build and
all 22 original-count cases. It begins only after the preceding Linux event
completes. Both host compilers, own interpreters, all references, applicable
native-C rows and separate lifecycle phases retain the existing protocol:
five fresh process blocks, two warmups and seven interleaved calls, with four
identical-binary A/A controls per case. Hosted diagnostic calibration does not
replace controlled acceptance. No phase subtraction, cross-source/target
pooling, favorable retry or flag/ISA/alignment sweep is used. At this record's
observation the shared build passes and measurement groups are queued.

## Preceding validation and remaining gates

[CI 42073348453](https://github.com/cpunion/wasmvm/actions/runs/42073348453)
completes for `ab377e7`/engine `4054f1b`: **all 21 required jobs pass**, with one
intentional optional skip. The audit verifies 240 matching conformance reports
across four compiler/platform profiles, all pinned manifests/compiler identities,
Stack Switching forced collection, 31 test packages in seven LLGo shards per
platform and 30 package coverage gates (minimum **90.84%**). This validates the
baseline, not the new frame-operand source.

The longest job is **616s (10m16s)**, slightly over the desired ten minutes. Its
pinned compiler setup takes 117s and semantic execution 461s; required evidence
upload takes two seconds. The earlier finalization failure remains recorded;
this successful run retains the two-minute required evidence-step limit.

The preceding [Linux comparison 42070246689](https://github.com/cpunion/wasmvm/actions/runs/42070246689)
also completes: six required build/measurement jobs, 22 cases, 362 alternate
checks, 12 expected reference-GC rejections, and seven cases with all four
controls resolved. SHA LLGo elapsed head/base medians are 0.956 (Clang) and
0.941 (Rust), with unresolved controls for both. Their ratios to Wasmtime remain
1.416/1.447 and to native C 1.702/1.723. These are separate-source advisory
observations and demonstrate remaining gaps, not controlled gains.

New final-head CI, actual x86 kernel execution, contribution review/integration,
general allocation/rematerialization, preparation costs, expanded world/release
conformance and controlled competitive/native acceptance remain open.

## Evidence

The [machine-readable record](data/frame-operands-2026-11-06.json) retains exact
source/profile identities, coverage, publication comparisons, assembler proof
and separate prior-event outcomes. The immutable archive includes all source
changes, qualification output, complete kernel/encoding dumps from both
compilers, reference assembly, actual ARM64 images and x86 function disassembly.

Current archive `artifacts/loop-frame-operands-20261106/evidence.tar.gz`: 1994348 bytes, 239 files; SHA-256
`15bf355687f25a62599ef7a140f6579c4657fda9d2b23ea1a0dbba3e9d769cfc`.

Baseline full-CI archive `artifacts/loop-frame-operands-20261106/ci-ab-evidence.tar.gz`: 5914474 bytes, 268 files; SHA-256
`b172110acb478326e1cf32f475c11f5824503fee9ef2e453d639d822e779750b`.

Preceding Linux archive `artifacts/loop-compare-immediates-20261106/hosted-evidence-current.tar.gz`: 2096744 bytes, 16 files; SHA-256
`16ff5e70fe005a9ca32e619a5d4e4f690056b174bcd46494fb77a18960b408a6`.

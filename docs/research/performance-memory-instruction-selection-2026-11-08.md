# Native memory instruction selection

Engine candidate `398a6ad` builds on integrated main `04a0d6e`. It adds no
feature flags or execution-policy exceptions. Both host compilers select from
the same existing hardware-feature inventory.

## Widening vector loads

All six signed/unsigned `v128.load8x8`, `v128.load16x4` and `v128.load32x2`
forms still read **exactly eight bytes** and overwrite the full 128-bit result.

| Existing target profile | Selection |
| --- | --- |
| SSE2 baseline | Eight-byte MOVQ, zero/sign-mask construction and low-half unpack |
| SSE4.1 | One PMOVSXBW/WD/DQ or PMOVZXBW/WD/DQ memory instruction |
| AVX2 | One VEX.128 VPMOVSX/ZX memory instruction, using the existing OS-state guard |
| ARM64 | Existing NEON implementation |

The previous emitter used the SSE2 sequence even when SSE4.1/AVX2 was enabled.
The new forms retain the original address calculation, bounds/trap sites,
metering and cancellation points. They do not read a 16-byte source. Legacy
and VEX encodings for all six operations and low/high destinations were checked
against the pinned LLVM 22.1.8 assembler before execution qualification.

## Scalar byte extraction

A contained unsigned byte load at stored-byte offset one can use AH/CH/DH/BH.
Other or spilled sources first copy into scratch RAX, exposing AH without a
separate shift. Signed i32 results use the corresponding sign extension.
Full signed i64 extension retains the old path because REX.W disables high-byte
register names. Checked stores, copies and independent cold entries remain;
shared/multi-memory and metered paths retain their original selection.

## Scope and evidence

Exact-source amd64 compile/link diagnostics under the previously recorded
feature mask 7 cover all 22 portfolio workloads. Only two linked modules change:

| Workload | Base bytes | Candidate bytes |
| --- | ---: | ---: |
| Narrow unaligned integers | 15,409 | 15,405 |
| SIMD widening memory | 9,258 | 9,213 |

The other 20 module images are byte-identical. These are compiler diagnostics
on ARM64, **not amd64 guest execution or timing evidence**. The
[machine index](data/memory-instruction-selection-2026-11-08.json) retains every
image hash and scope. Runtime comparisons keep all six memory/SIMD workloads.

Independent preceding Linux events provide the following observations, in
ns per source iteration. They use different source/host events and are not pooled.

| Engine source | LLGo native base → head | Go native base → head | Wasmtime | Control scope |
| --- | ---: | ---: | ---: | --- |
| `e7660b0`, integrated by PR #132 | 2.245 → 1.996 | 2.245 → 1.997 | 1.710 | Affected case resolved; 4/6 workloads |
| `6cf7e1f`, byte extraction alone | 2.003 → 1.724 | 2.011 → 1.727 | 1.714 | Affected case unresolved; 4/6 workloads |

The latter approaches Wasmtime but does not establish parity. Its Go-interpreter
row also observes 192.024 → 208.942 ns despite unchanged interpreter sources;
all adverse and unresolved rows remain in the raw data. The separate M4 event
keeps identical published base/head Go/LLGo guest images and resolves 1/6 workload
controls. No ARM64 improvement follows from that noisy unchanged-code event.

## Qualification contract

The [raw archive](data/memory-instruction-selection-2026-11-08.json.xz) retains
all three preceding six-case events: each has 102 alternative-input checks,
510 ordinary/A/A process observations and 30 unchecked native-C observations.
Source, binary and fixture hashes, separate API phases, seven interleaved warm
calls after two warmups, and five independent process blocks remain explicit.
No timer subtraction, sample selection, ISA/profile/count sweep or source/host
pooling is used. Five-process calibration is advisory and does not close the
at-least-20-pair controlled performance contract.

Go 1.27.3, LLGo `38de1e218b8146ba73f9c6f2e60e3c31f29d96b6`, LLVM 22.1.8
and LLGo `-O3 -lto=full` stay fixed. The local Go/LLGo codegen/native/VM suites
pass; local execution is ARM64, with amd64 emission checked separately. Codegen
statement coverage is 96.7%. The widening tests additionally execute baseline,
SSE4.1 and AVX2 profiles on capable amd64 hosts, covering all six forms, register
pressure, unaligned/end-of-memory addresses, traps, cold entries and exact fuel.
Exact code-size quotas are checked for both allocated and spilled destinations.

Integration requires final-head Go/LLGo Linux amd64/macOS arm64 semantics,
all 23 required CI jobs, 240 execution reports and all 30 coverage gates. Final
combined-source measurements are recorded separately in the contribution PR.
Competitor/native-C deficits, controlled M2 acceptance and release gates remain
open.

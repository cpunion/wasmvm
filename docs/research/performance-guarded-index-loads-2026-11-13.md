# Guarded indexed loads: one dependent-memory kernel

Go-native execution improves **25.6%** against its source baseline in this fixed
native Linux event. Its paired time ratios are **0.8648 vs Wasmtime**,
**0.8616 vs Wasmer**, and **1.0107 vs unchecked native C**. All six A/A controls
resolve within the diagnostic ±2.5% margin. This is one hosted workload,
not full-portfolio parity or controlled milestone acceptance.

The LLGo native backend improves **25.1%**, while its interpreter regresses
**13.6%**. That adverse result remains open under the Go-first integration
decision; it is not averaged away with native gains.

## Implementation and safety

Base `b546ea76663c2ee763ce303a7835e1f83096e59d`; measured production head
`95943842febc2d3571547c72f2d62bc79f232f88`.

The existing bounded i32 range transfer records nonwrapping scales 2/4/8
immediately consumed by numeric loads. amd64 can fold those scales into SIB
addresses only on proven normal paths. A low-bit-mask loop additionally checks
the whole incoming 64-bit Lo word once and verifies that a complete transfer
through its bounded body preserves the range at its backedge. The successful
path uses the original index directly; unusual raw inputs retain checked cold
execution. The mask/product and factor must occupy distinct temporary slots,
and every omitted logical word must be overwritten before any observation.

On the dependent-load fixture, a move, mask and shift leave the recurring
index dependency chain. A four-instruction guard executes once before the
fast header; the backedge skips it. Cancellation polling remains inside the
loop. No guest-specific name or checksum selects the transformation.

Fuel-enabled code, arbitrary cold PCs, whole frames, traps, cancellation,
address overflow and memory policy retain their original contracts. The
transformation does not apply to arm64 or memory64. Diagnostic dumps confirm
all five arm64 fixtures and four other amd64 memory fixtures are byte-identical.
Separate actual nativeexec image checks confirm unchanged Clang/Rust SHA-256
images under the Docker CPU profile. Those checks establish that profile's
code identity, not universal host performance. No Docker timing is promoted.

## Fixed native Linux comparison

[Run 43448812030](https://github.com/cpunion/wasmvm/actions/runs/43448812030)
executes on an AMD EPYC 7763 GitHub runner. One inherited CPU affinity,
GOMAXPROCS=1, 20 predeclared fresh-process blocks, two warmups and seven
randomized interleaved calls per runtime per block. Each call executes
1,000,000 source iterations, including deterministic memory reset. All results
and 19 separate alternate-input probes pass. All **400 process observations**
remain: 380 runtime rows and 20 native C rows. Each Go-native and LLGo role has
an identical-binary alias where declared in the saved protocol.

Go 1.27.3, LLGo `56fa9ae -O3 -lto=full`, Clang 22.1.8; wazero 1.12.0,
Wasmtime 48.0.1 Cranelift, Wasmer 7.3.0 Cranelift and WasmEdge 0.17.1
interpreter. Compiler/source/binary/fixture/SDK identities remain in the raw
record. wasmvm keeps checked memory and bounded cancellation, with optional
fuel disabled. Reference cancellation defaults differ; native C omits Wasm
isolation and safety policy. Physical frequency and neighbors are uncontrolled.

Times below are medians of the 20 process medians, in ns/source iteration.
A/A aliases and all other phases remain in the machine record.

| Configuration | Native/compiled | Interpreter |
| --- | ---: | ---: |
| wasmvm Go base | 2.57655 | 107.72050 |
| wasmvm Go head | **1.90836** | 107.81550 |
| wasmvm LLGo base | 2.54338 | 86.09287 |
| wasmvm LLGo head | **1.90428** | 97.66205 |
| Wasmtime | 2.20449 | N/A |
| Wasmer | 2.21106 | N/A |
| wazero | 2.95752 | 179.50726 |
| WasmEdge | N/A | 244.22755 |
| Native C, unchecked | 1.88788 | N/A |

Individual paired-block geometric time ratios use 20,000 seeded bootstrap
resamples. They are marginal diagnostic intervals, not simultaneous portfolio
acceptance. Ratio points can differ from ratios of the table's medians.

| Comparison | Ratio | Individual 95% interval |
| --- | ---: | --- |
| Go native head/base | 0.74401 | [0.74071, 0.74748] |
| LLGo native head/base | 0.74891 | [0.74702, 0.75074] |
| Go native / Wasmtime | 0.86483 | [0.86319, 0.86631] |
| Go native / Wasmer | 0.86165 | [0.85921, 0.86362] |
| Go native / wazero compiler | 0.64391 | [0.63964, 0.64813] |
| Go native / native C | 1.01073 | [1.00780, 1.01295] |
| LLGo interpreter head/base | **1.13594** | **[1.13246, 1.14003]** |

Preparation, first execution and warm work stay separate. Go native base/head
compile-validation medians are 0.761/0.764 ms, instantiation 0.063/0.066 ms,
and first-call 2.618/1.999 ms. These small preparation observations do not
establish a compilation-speed improvement. Interpreter native compilation is
not applicable, not a zero-duration native compiler result. No timer,
startup, control or phase medians are subtracted from execution.

## Retained neutral event and LLGo follow-up

The earlier scaling-only head `9ef24ba` completed its own fixed 20-block,
400-observation event on EPYC 9V45:
[run 43441525795](https://github.com/cpunion/wasmvm/actions/runs/43441525795).
All six controls resolve. Go-native head/base is 1.00263
[0.99892, 1.00622]; no speedup is claimed. LLGo interpreter regresses 10.1%.
It is retained independently, never pooled with the guarded-loop event.
The intervening build was cancelled after the temporary-independence proof
audit changed the source, before a measurement cohort completed.

The final LLGo interpreter's source dispatcher is unchanged. The exact ELF
binaries show `runSlots` remains 7,175 bytes with the same instruction
and local-control sequence, apart from relocated data references. Its entry
moves from offset 32 to 48 modulo 64. This supports investigating executable
layout; it does **not** prove an alignment cause or an LLGo correctness defect.
A controlled layout ablation is still needed before an upstream bug claim.
The compiler pin and production flags were not changed to conceal this result.
Go interpreter's paired ratio is 1.00175 [1.00011, 1.00348], without dedicated
Go-interpreter A/A controls in this event.

## Validation and remaining work

Full Go semantics and all 31 executable package coverage gates pass at ≥90%.
Affected compiler/native/VM race tests and pinned production LLGo execution
pass. Linux amd64 Docker semantics and both target code-generation paths pass.
Tests include every cold PC, raw upper Lo bits, numeric load widths,
multiplication/modulo shifts, independent temporary slots, non-inductive
backedges, remapped/aliased registers, complete-frame exits, generated loop
polls and entry cancellation. Final-head hosted Go/LLGo Linux amd64/macOS arm64
checks and resolved reviews remain mandatory before merging.

This change closes neither private/shared-memory, SIMD, SHA-256 nor complete
P3 performance gaps. The prior complete P3 comparison still executes about
3.8–4.0× Wasmtime's instructions. Continue profiling actual host/guest work
instead of treating allocation counts or GC as the whole gap.

- [Machine summary](data/performance-guarded-index-loads-2026-11-13.json) includes every control, separate phase medians and image identities.
- [Final raw event](data/performance-guarded-index-loads-2026-11-13-guarded-loop.json.gz) and [neutral raw event](data/performance-guarded-index-loads-2026-11-13-scaling-only.json.gz) preserve every observation, warm sample, checksum, order, phase and build identity losslessly.
- [Reproduction and semantic evidence](data/performance-guarded-index-loads-2026-11-13-protocol.tar.gz) preserves the analysis, plans, image overlays/disassembly, LLGo dispatcher comparison and validation logs. No executables or checkout snapshots are committed.

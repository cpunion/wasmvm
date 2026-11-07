# Direct scalar memory operands: 2026-11-07

**Local qualification and actual Linux benchmark execution pass; final-head
semantic CI remains required.** Engine `d64119b` compares against `96ea7fb`, with the same pinned
Go 1.27.3, LLGo `3571c25` release profile and reference versions.

## Selection and safety

The existing bounded memory-region analysis now tracks whether each locally
executed i32 definition is zero extended, and whether it is an exact constant.
Proven scalar accesses can use the x86 stable memory base plus a constant or
canonical GPR index and displacement. All 23 scalar load/store opcodes share
the existing encoders. No register bank, execution framework or ISA is added.

Incoming frame words have unknown upper bits. Facts reset at every region entry;
unknown or elided writes kill them. Local copies, select results and remapped
addresses retain their actual original-PC facts. Precise mode, memory64,
vectors, unproven accesses, unknown upper bits and large displacements retain
the established address calculation. Shared minimum-size guards, cold frames,
raw Lo/Hi words, original PCs, fuel, cancellation and trap contracts are unchanged.

## Qualification

Complete codegen/nativeexec/public-VM tests pass under Go and release LLGo;
vet passes. Exact package coverage is **codegen 97.71% (8529/8729)** and
**nativeexec 92.54% (459/496)**, rather than specification coverage.

Tests exercise all scalar widths/signs and displacement boundaries, plus
constant/copy/arithmetic/select facts, invalidated facts, remapped addresses,
new entries, memory64 and precise-mode fallback. Complete-frame tests perform
**5,184 executions per local compiler** across nine load/store pairs, dynamic
and constant addresses, private/shared memory, four offsets, nine cold PCs and
four raw input patterns. They verify every Lo/Hi word, whole-memory effects,
return/trap state, original continuation, fuel and budget. Both targets compile;
the ARM64 counterpart executes locally. Actual x86 paths require Linux CI.

Independent LLVM 22.1.8 assembly matches **35,328 individual instructions**:
23 opcodes × 16 data registers × 16 index/no-index choices × six displacements.
Go and LLGo emit identical bytes, including REX low-byte/high-register forms,
SIB aliases, signed loads and FP widths. This is encoding evidence; it does not
execute x86 instructions on ARM64.

Two fixed development failures remain in the archive: applying memory
classification to a non-memory opcode, and a complete-frame oracle's return-PC
assertion. Neither is counted as a successful qualification result.

## Published code and measured scope

All 44 fresh ARM64 checksum witnesses pass. Across 22 workloads, exact-baseline
and candidate Go/LLGo published images and entry lists are identical. The 22
baseline publication witnesses are reused only after source, compiler, profile,
executable, library and fixture hashes match. No additional ARM64 warm campaign
is collected for identical guest code.

All 22 x86 modules compile and link with the established feature mask 7. Fourteen
modules change. Static loop counts use only backward spans containing the
compiler's actual `PrivatePollOffsets`, excluding embedded dispatch-table data.

| x86 private loop | Instructions, base → head | Frame references |
| --- | ---: | ---: |
| Narrow unaligned memory | 26 → 22 | 2 → 2 |
| Dependent address | 14 → 13 | 0 → 0 |
| Private/shared memory | 10 → 8 | 0 → 0 |
| Clang/Rust SHA schedule | 43 → 35 | 0 → 0 |
| Clang SHA compression | 69 → 68 | 20 → 20 |
| Rust SHA compression | 65 → 64 | 18 → 18 |

Clang/Rust compression functions shrink from 10859/10738 to 10827/10706 bytes.
Static reductions establish no elapsed gain. The
[source-pinned Linux event 42366175293](https://github.com/cpunion/wasmvm/actions/runs/42366175293)
retains all 22 original-count workloads, both compilers/backends, four A/A
controls, alternate checksum checks, competitors and applicable native C.
Preparation, publication, first call, warm calls, shutdown and process intervals
remain separate. The event completes all six required jobs: 22 cases, 362 alternate checksum
checks and 12 expected reference-GC rejections. Seven cases resolve all four A/A
controls; both SHA cases do. Full actual-target final-head semantic CI remains
required; controlled milestone acceptance stays open.

## Completed Linux observations

Ratios compare full warm-call intervals; lower is faster. These hosted results
are advisory. The Rust/Clang SHA changes observe approximately 5.8%/5.7% less
LLGo elapsed time against the exact `96ea7fb` baseline.

| Workload | LLGo head/base | / Wasmtime | / native C | All four A/A |
| --- | ---: | ---: | ---: | --- |
| Clang SHA | 0.943 | 1.183 | 1.549 | Resolved |
| Rust SHA | 0.942 | 1.099 | 1.514 | Resolved |
| Dependent address | 0.976 | 1.304 | 1.522 | Resolved |
| Float64 unaligned | 0.954 | 1.131 | 1.149 | Resolved |
| Narrow unaligned | 1.064 | 1.315 | 2.077 | Inconclusive |
| Core memory load/store | 1.042 | 1.225 | 1.686 | Inconclusive |

The adverse narrow/core observations and tiny-call-control loss remain in the
complete record. Unresolved controls do not become passes, and there is no
favorable retry. Go/LLGo, interpreter/native, all competitors and applicable C
rows retain independent preparation, first, warm, shutdown and process phases.
Cross-event ratios are not pooled. Build takes 383 seconds; the longest group
takes 493 seconds. The next main-baseline event measures the cumulative branch
separately rather than extrapolating from these incremental measurements.

The preceding mixed-forwarding event and the rejected raw-FP allocation event
remain independent in the [completed study](performance-mixed-scalar-forwarding-2026-11-06.md).
The latter regresses Linux SHA by about 22% and stays reverted. LLGo #2983
remains with its existing contributor; the compiler pin is unchanged after
upstream merged LLGo #2991.

The [machine-readable record](data/direct-memory-operands-2026-11-07.json)
pins all identities and complete static summaries. Immutable local archive:
`artifacts/memory-operands-20261106/local-evidence.tar.gz`, 3203770 bytes,
473 files; SHA-256
`e8a05c67d607a21153a7b37aec03dde3e8dae7de9032aa8edbc1fbc7c6b3d732`.

Completed hosted archive: `artifacts/memory-operands-20261106/hosted-evidence.tar.gz`, 2103812 bytes, 18 files; SHA-256 `e322557bc34a8ca183d5f7ac472648e1904468485c7d326b58fdac51ef3084bd`.

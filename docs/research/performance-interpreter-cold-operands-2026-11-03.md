# Scalar operand lifetime after extension dispatch: 2026-11-03

Status: source-changing correction, local validation and complete diagnostic scope pass; new exact-head cross-platform CI is required. Performance acceptance stays open.

Both earlier PR112 heads pass full semantic CI but retain material Linux scalar
regressions. The [second CI hold](pr112-control-ci-2026-11-03.md) records
LLGo integer/calls/memory +13.8%/+5.5%/+13.2% on `5b767da`. No unchanged timing
rerun replaces either attempt.

## Assembly evidence and change

Exact base/head engine sources are exported and built for linux/amd64 with Go
1.27.3, LLVM 22.1.8 and LLGo `983465e`, using the same `-O3 -lto=full` profile.
The compiler source mount is read-only. The container executes under emulation
on an ARM64 host; this scope builds and disassembles only, without measuring
or claiming native throughput. Source inventories, compiler/binary hashes,
commands, raw and source-annotated assembly are retained.

Initial prefix dispatch moves the common scalar `a.Lo` to stack storage in the
linked amd64 Run function; baseline keeps it in a register. Total frame sizes
are 0x718/0x6d8 respectively, so a smaller frame alone does not establish a
cheaper loop. The new source loads SIMD/GC inputs within their handlers and
loads scalar inputs after extension classification. Its linked scalar load
keeps `a.Lo` in a register again. This motivates the change without attributing
an exact number of cycles or all observed regressions to one spill.

Opcode classification, checked memory, shared result/trap writeback, fuel/poll
frequency, guest roots and GC PC publication remain intact. Wide arithmetic
and Stack Switching no longer prepare unused scalar inputs. No opcode, backend,
public API or supported proposal is removed.

## Complete local interpreter comparison

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7` is PR111's implementation; new
source `452df712b53bfdc66e8e7ca80fd42d10e8a4cf4a`. Fixed Go 1.27.3/LLGo `983465e`,
full LTO, Apple M4 Max/macOS arm64. Three fresh serially interleaved process
blocks, two warmups/seven samples and the unchanged twenty-kernel counts.
The complete grid retains 864 VM/reference and 27 ordinary-C observations,
288 alternate checks, twelve explicit unsupported GC outcomes and 297
independently regenerated summaries. Baseline Go/LLGo interpreter A/A aliases,
all phases and original outcomes remain separate. No subtraction or pooling
with earlier campaigns. Warm ns/source iteration, except empty ns/call:

| Kernel | Go base → head | LLGo base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 49.920 → 49.169 | 43.415 → 40.867 | 38.268 | 61.257 |
| core/call-control | 2625.000 → 2709.000 | 3167.000 → 3042.000 | 3000.000 | 1042.000 |
| core/integer-loop | 41.442 → 41.294 | 35.511 → 34.279 | 48.552 | 81.683 |
| core/function-calls | 63.490 → 60.944 | 59.456 → 56.477 | 62.123 | 79.874 |
| core/memory-load-store | 84.244 → 83.572 | 74.700 → 72.132 | 90.758 | 145.359 |
| memory/narrow-unaligned | 109.597 → 109.020 | 98.226 → 93.881 | 113.402 | 185.196 |
| memory/dependent-address | 55.172 → 54.138 | 48.424 → 46.580 | 53.516 | 98.158 |
| memory/float64-unaligned | 87.541 → 87.102 | 78.516 → 76.998 | 84.158 | 138.395 |
| runtime/simd-memory-full | 107.027 → 95.964 | 95.195 → 75.739 | 77.785 | 149.616 |
| runtime/simd-memory-widen | 88.786 → 78.521 | 85.383 → 70.394 | 61.223 | 144.777 |
| runtime/simd-memory-lanes | 130.315 → 116.290 | 125.046 → 97.968 | 90.206 | 185.355 |
| gc/i31-roundtrip | 54.304 → 53.846 | 42.812 → 42.712 | 45.300 | — |
| gc/struct-access | 57.333 → 54.796 | 50.612 → 48.171 | 56.500 | — |
| gc/struct-allocation | 123.037 → 114.300 | 146.004 → 141.717 | 107.400 | — |
| gc/struct-indirect-access | 99.596 → 102.513 | 90.508 → 88.817 | 96.600 | — |
| threads/shared-memory | 83.476 → 83.509 | 43.434 → 40.576 | 38.667 | 62.865 |
| threads/atomic-rmw | 63.734 → 62.519 | 43.817 → 40.426 | 32.024 | 44.016 |
| threads/notify-empty | 42.300 → 42.404 | 39.271 → 37.375 | 40.700 | 42.833 |
| threads/wait-mismatch | 71.929 → 70.500 | 53.496 → 49.658 | 45.000 | 70.500 |
| threads/wait-timeout | 71.421 → 70.617 | 49.479 → 46.083 | 97.800 | 287.742 |

All nineteen sustained LLGo interpreter cases improve, 0.2–21.7%; SIMD improves
17.6–21.7%. Full-vector memory leads this WasmEdge; widening and lanes remain
about 1.150x/1.086x its time. Go SIMD improves 10.3–11.6%, while indirect GC
increases 2.9%. These shared-host observations do not establish controlled
acceptance or resolution of Linux regressions.

## Native controls and validation

The complete native/reference/C table remains in the lossless evidence. Retain
LLGo native empty +21.3%, i31 +8.9%, direct GC +2.7%, widening +1.6% and
wait-mismatch +35.5%. Go native empty +16.5%, i31/direct GC +7.9%/+8.0% and
wait-mismatch/timeout +12.2%/+10.6% also remain. Earlier source identities'
long/batch controls do not waive these outcomes. No native control is rerun
to substitute a favorable observation. Ordinary-C memory, public-boundary,
startup/cancellation and expanded WASI-world acceptance remain open.

Both Go and pinned full-LTO LLGo pass the complete affected SIMD/execute/VM
suites, with 48 passing reports each. All five mandatory proposals execute
three tiers, including forced collection for Stack Switching. The unchanged
C/harness implementation previously passes 81 tests and both full CI heads.
New final-head 31-package Go/LLGo × Linux amd64/macOS arm64 CI, thirty coverage
gates, performance and review checks remain required before merge.

[Complete sources, source-annotated amd64 assembly, original outcomes, semantic
reports and independent verification](evidence/performance-interpreter-cold-operands-2026-11-03.json.gz).

# Integer-image entry effects

The final Go source comparison reduces complete P3 request CPU by about
2.1–2.3% and increases throughput by 2.2–2.7%. This is a bounded entry
optimization. Full VM competitiveness and release acceptance remain open.

## Compiler proof and runtime contract

The compiler classifies the complete immutable image before publication,
including unused bodies and cold continuations. Floating/vector slots or
signatures, FP instructions, SIMD and unknown operations retain the complete
floating-environment transition. The integer whitelist covers the current
integer, scalar-memory, bulk, atomic and reference/control lowerings. These
selectors must not observe or change floating control/status; future FP-based
integer legalizations must update this contract. SIMD source stays conservative
even though some integer selectors use non-saturating bitwise SIMD internally.

The property belongs to immutable published code, with no public VM option.
Every linked target belongs to the same image. Imports and foreign targets
leave native execution before their own artifact is entered. The ordinary
publication path and arbitrary machine-code probes keep complete normalization.

The C adapter shares admission and cancellation checks between both cases.
Verified integer code leaves the ambient FP environment untouched. Other code
still establishes Wasm FP behavior and restores the complete host state.
Frame/slot quotas, roots, continuation validation, metering and cleanup remain
unchanged. Proof scans poll cancellation at function boundaries and every
256 slots/instructions before executable publication.

Tests reject mixed images and unknown cold instructions; cancel during each
proof scan; and check every returned status plus invalid admission/cancellation
under non-default host FP state. Actual generated integer arithmetic/comparison
selectors execute under that state for baseline and detected CPU profiles,
including ARM popcount's internal SIMD. The diagnostic canary that reads FP
state is deliberately outside the compiler's proof.

## Two separate measured events

Both use the frozen complete Rust HTTP P3 component, snapshot requests and
buffered responses, four Stores, sixteen clients and 512 checked requests per
batch. Each fresh process excludes preparation, first call and two warmups,
then measures twenty batches and a separately checked final flush. Individual
batches interleave across processes; other children remain stopped. CPU and
instruction snapshots bracket complete batches, including a final background
tail; API batch timers remain separate. GC is enabled. No owned build, test,
profiler, pruning or source edit overlaps either event.

The initial `0fed4e4` event has 20 paired blocks, four roles and identical
aliases: 320 fresh processes across 16 B and 64 KiB payloads. It compares Go
base/head, pinned LLGo head and the actual Wasmtime 48 P3 provider. Go source
CPU ratios are 0.98137/0.97923; its CPU/throughput A/A controls resolve ±2.5%.
LLGo controls remain unresolved at both sizes, as does Wasmtime's 16 B CPU
control. Keep those observations; they do not establish reference CPU ranks.

After adding bounded compilation-proof cancellation, the final `a8084b3`
event independently remeasures both Go sources: 20 blocks/80 processes per
payload, 160 total. It is not an extension or pooling of the earlier event.
Subsequent changes add tests/documentation only; production sources match the
measured revision. Go 1.27.3 and Clang 22.1.8 stay fixed throughout.

| Final-source comparison | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU head/base, individual 95% interval | 0.97869 [0.97170, 0.98581] | 0.97739 [0.97231, 0.98251] |
| Throughput head/base, individual 95% interval | 1.02193 [1.01565, 1.02823] | 1.02725 [1.02106, 1.03351] |
| Retired-instruction head/base | 1.00108 [0.99958, 1.00303] | 1.00006 [0.99815, 1.00199] |
| CPU medians, base → head, microseconds/request | 105.37 → 103.74 | 143.78 → 141.25 |
| Managed allocations/request, base and head | 705 | 789 |

All final Go CPU, instruction and throughput A/A intervals resolve ±2.5%.
Process blocks are the statistical unit; intervals use a seeded 20,000-sample
paired log-ratio bootstrap and are individual diagnostics, not simultaneous
full-portfolio acceptance. Unchanged instruction totals and omitted FP-state
synchronization are consistent with the CPU improvement; no pure instruction
cost is manufactured by subtraction.

The initial comparative event remains a separate description of the gap:

| Payload | Go prototype CPU, main/alias | LLGo head CPU, main/alias | Wasmtime CPU, main/alias |
| --- | ---: | ---: | ---: |
| 16 B | 108.22 / 107.72 µs | 301.18 / 306.61 µs | 37.48 / 38.38 µs |
| 64 KiB | 140.54 / 141.96 µs | 339.88 / 378.65 µs | 48.85 / 49.99 µs |

These do not substitute older reference medians into the final source pairing.
Go still executes roughly four times the reference's instructions. LLGo
reports 1,160/1,425 actual managed allocation events versus Go's 705/789;
the collectors count different managed activity. Wasmtime's allocation counter
covers its Go shim only, not Rust/native allocations. None of these counts
alone identifies an LLGo compiler defect.

## Validation and evidence

All 32 Go test packages pass. Actual-source integration coverage covers all
31 executable packages at ≥90%: native 95.10%, nativeexec 93.25%, codegen
98.01%, VM 93.52%. Relevant Go race checks pass. Pinned LLGo `56fa9ae`,
`-O3 -lto=full`, executes native/nativeexec/VM/component/P2/P3 tests; final
proof-cancellation and generated-selector tests pass as well. Local Docker
validates x86 semantics only. The full real Linux/macOS dual-compiler PR
matrix, CLI, coverage and benchmark checks remain required before integration.

[Identities, protocols, controls and evidence](data/performance-integer-entry-effects-2026-11-13.json)
retain all 480 process observations. Committed projections keep every batch
timer, CPU/instruction snapshot, allocation/GC counter, phase and checksum.
The two unused per-request latency arrays are replaced by counts, distributions
and exact hashes; full originals remain in local artifacts. No process/sample
is removed, and all reported claims can be recomputed from committed fields.
This reduces the evidence from about 30 MiB to 2.8 MiB.

The remaining work targets general guest calls, component ownership and
scheduling costs. This improvement does not close those larger gaps.

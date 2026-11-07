# Loop register plans: 2026-11-05

**Retained candidate; controlled acceptance remains open.** Loop-local integer homes reduce C/Rust SHA native durations against the preceding affine-loop source by **13.28% / 17.56% under LLGo**, **14.37% / 12.04% under Go**. Same-binary LLGo head native A/A differences are **+5.32% / +5.00%**, so this noisy shared-host cohort does not establish competitor parity.

Measured engine `d472a76ff9a8aebfdacd6ba74eeccc962d404e59`; preceding N `0cde7c5578386d280eebce6b6cd7faaaa45f2d09`; production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Nine engine/test files change against N; 28 differ against production, including tests. Go 1.27.3 and LLGo 3571 (`-O3 -lto=full`), guests, reference binaries, policies, clocks and alignment remain fixed.

## Mechanism and validation

Function-wide homes forced independent loops to share the same register inventory. The compiler now records explicit location plans for loop regions and deferred poll/trap/call exits. Bounded region priorities reuse the existing slot-use model, retain common homes and assign free registers deterministically. Estimated cache savings must exceed entry/exit transfers. No scan or allocation proportional to all unused function slots is repeated for each loop.

Before changing homes, spill every displaced word before loading any replacement. Ordinary loop exits restore the enclosing function map; private backedges keep the local map. Deferred exits use the map at their originating instruction, retaining complete canonical scalar words, all vector bits, reference handles and original PCs. FP/vector and immutable constant/runtime homes keep their function assignment. Guards, checked interior entries, all interruption polls, fuel tiers and the shared runtime remain intact. This location infrastructure advances allocation; **full global live-range splitting/linear scan remains unfinished**.

Ten complete affected packages pass both pinned host compilers on macOS arm64, with **108 actual conformance reports** and coverage **90.71%–100%**. Actual generated-code tests exercise register pressure, complete scalar/vector tuples, aliases, wrapping, every loop interior entry, short-loop placement and deferred poll/trap/call snapshots before/inside/after transitions. Actual emulated Linux amd64 tests also pass under diagnostic Go 1.27.0; this does not replace pinned target CI. Race, cgocheck2 and no-cgo checks pass. Eight application configurations pass 64 checksum calls and 48 expected traps. Full final-head four-profile qualification/review remain required.

Both compiled SHA core functions select two local maps. Diagnostic ordinary function sizes change C 11060→11160 bytes and Rust 10952→11024 bytes. These per-function diagnostic emissions are not published whole-module dumps.

## Fixed comparison

One production/N/O cohort: three fresh process blocks, two warmups/seven individually interleaved samples, same-binary LLGo production/head controls. Original20 independent verification: 1224 VM/reference+27 C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes, 417 summaries. Separate application verification: 126 VM/reference+6 C observations, 42 alternate checks, 44 summaries. All outcomes retained; no timing retry or pooling with earlier cohorts.

Warm ns/message block at 5000 blocks/call includes generation, chaining and padding; the independent checksum is first 32 digest bits. These are two compiled Core compute modules for one algorithm, not a complete application/WASI-world portfolio.

| Guest | LLGo production → N → O | Go O | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 780.467 → 251.133 → 217.783 | 226.550 | 215.833 | 207.183 | 314.008 | 179.192 |
| sha256-rust | 740.208 → 265.292 → 218.708 | 226.742 | 194.433 | 208.492 | 361.517 | 177.625 |

LLGo native point estimates are 30.64%/39.50% below wazero, 0.90%/12.48% above Wasmtime, 5.12%/4.90% above Wasmer and 21.54%/23.13% above unchecked C. Production native A/A differences are −0.75%/−1.57%; head controls are about +5%. Policies differ and C omits VM obligations. Both controls and absolute reference rates vary: do not infer a controlled margin or multiply improvements from earlier cohorts.

| Guest | Go interpreter vs production / N | LLGo interpreter vs production / N | LLGo interpreter A/A production / head |
| --- | ---: | ---: | ---: |
| sha256-clang | -3.05% / -1.74% | +2.36% / -0.22% | -1.17% / +0.46% |
| sha256-rust | +0.62% / -0.84% | +8.92% / +0.76% | +1.56% / -4.64% |

Startup and preparation remain separate: LLGo readiness is 99–104 ms against Go 4.45–5.14 ms; LLGo native validate/compile is 3.33–3.95 ms against Go 1.97–2.02 ms. Load, setup, instantiate, lookup, first/warmup/warm and cleanup records remain unmodified. Phase medians are never summed or subtracted.

## Complete original20 outcomes

Positive changes mean slower. All Go/LLGo backends, reference rates, phases and A/A controls remain machine-readable.

| Workload | Go native vs N | LLGo native vs N | LLGo native vs production | LLGo interpreter vs production |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | -0.88% | -0.01% | +0.59% | +0.91% |
| core/call-control | +9.50% | -5.77% | +6.53% | -20.76% |
| core/integer-loop | -0.45% | -0.61% | +0.41% | +2.50% |
| core/function-calls | +0.24% | -0.08% | +0.40% | -2.52% |
| core/memory-load-store | +0.05% | -0.12% | +1.24% | +4.73% |
| memory/narrow-unaligned | +0.11% | +1.55% | +1.78% | +2.43% |
| memory/dependent-address | -0.13% | -0.36% | +0.00% | +4.08% |
| memory/float64-unaligned | -2.16% | -0.24% | +0.49% | +2.86% |
| runtime/simd-memory-full | -0.65% | -0.24% | -0.08% | -0.14% |
| runtime/simd-memory-widen | -0.29% | +0.11% | +0.06% | +0.55% |
| runtime/simd-memory-lanes | +3.24% | -2.43% | -1.12% | +5.48% |
| gc/i31-roundtrip | +2.68% | -12.35% | -2.99% | +1.30% |
| gc/struct-access | -7.46% | -9.77% | -0.90% | +3.33% |
| gc/struct-allocation | -1.41% | -3.67% | -5.30% | +1.53% |
| gc/struct-indirect-access | -0.67% | -0.34% | +3.48% | +0.46% |
| threads/shared-memory | +0.69% | -0.45% | -0.39% | +0.63% |
| threads/atomic-rmw | +1.20% | -0.93% | -0.08% | -1.62% |
| threads/notify-empty | -7.48% | -11.41% | +0.24% | -1.92% |
| threads/wait-mismatch | +1.30% | -7.43% | +4.81% | +0.71% |
| threads/wait-timeout | -4.03% | -10.68% | -6.45% | +3.00% |

Retain Go native i31+14.03%, wait-timeout+25.88% against production and every other adverse row, even where unrelated controls are noisy. LLGo interpreter memory-load/store+4.73%, dependent-address+4.08%, SIMD lanes+5.48% and Rust SHA+8.92% remain unresolved. A native application improvement does not erase interpreter gaps, initialization/preparation costs or GC/threads acceptance.

## Integration and next work

Keep this candidate while preserving N and held experiments unchanged. Complete semantic CI, review and controlled acceptance; extend location maps toward full live-range allocation/rematerialization and a broader application/WASI-world portfolio. The sole external PR2991 remains open and the earlier parallel-PR question is unanswered, so no new contribution PR is opened. LLGo #2983 belongs to another contributor. Production main is unchanged.

The N baseline [CI run](https://github.com/cpunion/wasmvm/actions/runs/41886225379) passed all 19 required jobs, with one intentional optional-reference skip. Two jobs failed to acquire hosted macOS runners before succeeding on attempt 3. The earlier async job reached its 12-minute limit; its package subsequently passed. Cancellations remain distinct from engine failures, and successful jobs were not rerun.

Its macOS HTTP job took 11m56s: P1/P2 links consumed 72.39s/174.98s and tests 87.64s/77.55s; entry took 10m34s, including threads link/test 70.02s/81.78s. The candidate moves P1 and legacy threads into one shared-dependency shard, leaving P2 and CLI in separate existing jobs. Only the paired macOS legacy shard serializes package work. Seven shards still run all 31 test packages exactly once per LLGo platform, with identical release flags and corpora. This trades two platform-job setups for shorter critical paths; final-head CI must establish the actual improvement. The candidate still needs its own full matrix.

The [residual native profile](performance-loop-residual-profile-2026-11-05.md) confirms fewer instructions and canonical-frame accesses inside the guarded loops while retaining their interruption checks.

[Lossless evidence](evidence/performance-loop-register-plans-2026-11-05.json.xz) includes exact source/compilers/reference identities, actual qualification/coverage and ownership checks, amd64 execution, original20/application observations and all controls, phase records, diagnostic selection counts and independent verifiers. Source aliases and fixed seeds were checked before measurement; no timing changed.

# Bounded affine loop spans: 2026-11-05

**Retained candidate; acceptance remains open.** Against the preceding outlined-boundary source, native C/Rust SHA durations fall **6.38% / 5.50% under LLGo** and **6.85% / 6.04% under Go**. The new proof handles independent invariant bases and non-unit strides within the existing bounded compiler pass; runtime, ISA, source guests and release profiles are unchanged.

Measured engine `0cde7c5578386d280eebce6b6cd7faaaa45f2d09`, preceding L `22f1a1dd635c102b0c63b27c57d1148a85da4743`, production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Six engine/test files change against L; 23 engine files differ against production, including tests. The held frame-scope experiment is not included.

## Mechanism and qualification

For an equality-exit loop, require an entry cursor below the limit and a positive stride dividing the remaining distance. Otherwise take the original checked loop. Derive each address as an invariant local plus the entry cursor and an i32 bias. Prove all future access intervals with widened endpoints, retaining memarg offsets, 32-bit wrapping and exact trap order. Fold an address only when its canonical anchor plus the whole folded offset is also proved. Mutation, unknown/nonlinear addresses, trapping operations and unsupported control retain the checked path. Existing backedge polls and checked interior PCs survive; preheader guard work is bounded and polled. Metered tiers retain their original fuel semantics.

All ten complete affected packages pass pinned Go 1.27.3 and LLGo 3571 (`-O3 -lto=full`) on macOS arm64, with **108 actual conformance reports** and coverage **90.71%–100%**. New actual native tests cover multiple/aliased bases, unit/power-of-two/non-power-of-two strides, high input bits, wrapping and negative offsets, side effects before traps and every interior PC; guard-only tests cover memory32 endpoints without multi-GiB mappings. Both ISAs compile, and actual emulated Linux amd64 guard tests pass with diagnostic Go 1.27.0; that older toolchain is not pinned production acceptance. Race, cgocheck2 and no-cgo checks pass. Eight compiled-application configurations pass 64 checksum calls and 48 expected traps.

The retained L hosted run exposed an OfficialGoGuest tier2 compile deadline at 90 s. The exact local L fixture passes all three tiers in 19.27 s test execution. Hosted trace includes 348.23 s CLI test linking, 115.86 s thread-test linking and 194.45 s thread execution with overlap. N reuses the same eagerly compiled native artifact for ordinary/metered official Go guest calls, with independent process/store/instance state and unchanged 90 s compile budget and all output assertions. Both host compilers execute every tier. On the small macOS runners, entry/http shards now use `-p=1` to bound overlapping package builds, full-LTO links and test processes; all 31 packages, six shards, release flags and deadlines stay unchanged. Full hosted final-head qualification and review are still required.

## Fixed comparison

One production/L/N cohort uses unchanged pinned compilers/references, three fresh process blocks, two warmups/seven individually interleaved samples and same-binary LLGo production/head A/A controls. Original20 verification: 1224 VM/reference observations, 27 C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes, 417 summaries. Separate application verification: 126 VM/reference observations, 6 C observations, 42 alternate checks, 44 summaries. No timing retries or pooling with earlier cohorts. Shared-host results are descriptive.

Warm ns/message block for the complete C/Rust SHA modules at 5000 blocks/call, including message generation, chaining and padding. The independently checked result is first 32 digest bits. Two frontends for one Core compute algorithm do not constitute a complete application/WASI-world portfolio.

| Guest | LLGo production → L → N | Go N | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 709.667 → 250.900 → 234.892 | 238.450 | 185.733 | 192.475 | 292.717 | 160.617 |
| sha256-rust | 663.850 → 250.750 → 236.958 | 240.817 | 184.017 | 186.008 | 319.608 | 159.717 |

LLGo native point estimates are 19.75%/25.86% below wazero, 26.47%/28.77% above Wasmtime and 46.24%/48.36% above unchecked C. LLGo/Go duration ratios are 0.9851/0.9840. Head native A/A changes are −0.57%/+0.90% (production +0.84%/+3.30%). Policies differ; C omits VM obligations. These observations do not close controlled targets.

| Guest | Go interpreter vs production / L | LLGo interpreter vs production / L | LLGo interpreter A/A production / head |
| --- | ---: | ---: | ---: |
| sha256-clang | +1.13% / -0.99% | +4.19% / +0.75% | +0.34% / -0.08% |
| sha256-rust | +2.20% / +0.54% | +4.12% / +0.12% | -0.14% / -0.29% |

Preparation and lifecycle phases stay separate. LLGo readiness is 92–93 ms against Go 3.9–4.3 ms; LLGo native validation/compilation is 3.64–3.66 ms against Go 1.89–1.95 ms. C-guest LLGo preparation rises from 3.323 to 3.660 ms against L; Rust is 3.647→3.642 ms. Ordinary diagnostic core-function code grows 10188→11060 bytes (C),10132→10952 (Rust), with two selected loop guards each; these are per-function diagnostic emissions, not published whole-module dumps. All load/setup/instantiate/lookup/first/warmup/cleanup records remain in evidence; medians are neither summed nor subtracted.

## Complete original20 outcomes

Positive duration changes are slower. Each comparison below comes directly from this cohort; all Go/LLGo interpreter/native values, references and A/A controls remain machine-readable.

| Workload | Go native vs L | LLGo native vs L | LLGo native vs production | LLGo interpreter vs production |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | +0.10% | +0.26% | +0.59% | +0.33% |
| core/call-control | +23.28% | +25.05% | +1.07% | -15.64% |
| core/integer-loop | +0.07% | +0.36% | +0.25% | +2.94% |
| core/function-calls | -0.04% | +1.40% | +0.87% | -1.37% |
| core/memory-load-store | -0.21% | +0.68% | +0.83% | +6.19% |
| memory/narrow-unaligned | -0.38% | -0.12% | -0.09% | +6.74% |
| memory/dependent-address | -0.60% | -0.23% | +0.29% | +3.30% |
| memory/float64-unaligned | +0.62% | +0.40% | +1.55% | +3.86% |
| runtime/simd-memory-full | +0.22% | -0.41% | +0.75% | +0.04% |
| runtime/simd-memory-widen | +0.13% | +0.82% | +1.36% | -0.58% |
| runtime/simd-memory-lanes | +0.01% | +2.55% | +0.34% | +1.16% |
| gc/i31-roundtrip | -5.46% | -14.32% | -0.65% | -1.12% |
| gc/struct-access | -6.07% | -6.27% | +11.01% | +3.72% |
| gc/struct-allocation | +6.87% | -2.47% | -7.49% | +1.76% |
| gc/struct-indirect-access | -14.76% | -0.31% | -0.93% | +3.01% |
| threads/shared-memory | -1.45% | +0.39% | +0.77% | +2.45% |
| threads/atomic-rmw | -0.46% | +0.64% | +0.78% | -1.22% |
| threads/notify-empty | +13.37% | +6.01% | +17.22% | -1.84% |
| threads/wait-mismatch | -11.65% | -1.78% | +9.12% | +0.06% |
| threads/wait-timeout | -2.35% | -7.86% | +0.00% | +0.30% |

Retain native LLGo struct-access+11.01%, notify+17.22%, wait-mismatch+9.12% against production and all other adverse outcomes. Some short controls and GC allocation rows have large A/A variation; they cannot establish parity. LLGo interpreter memory-load/store+6.19%, narrow+6.74% and Go private-memory+4.94%, notify+8.63% also remain unresolved. A native application gain does not erase these rows.

## Remaining work

Keep N as the candidate, preserving L and the held M experiment. Inspect residual generated-code slot/register traffic before choosing further native optimization. Complete the hosted four-profile semantic matrix, review, contribution workflow and controlled portfolio/application/WASI-world acceptance. Production main remains unchanged. External PR2991 remains open and no parallel PR permission was received, so no new PR is opened. LLGo #2983 is handled elsewhere; no duplicate fix/issue is included.

[Lossless evidence](evidence/performance-affine-loop-spans-2026-11-05.json.xz) retains exact source/compiler/reference identities, qualification/coverage and ownership checks, both complete cohorts, phases, all adverse rows and independent verifiers. A copied secondary full-plan sourceRevision label retained the held M ID; the correct primary source, actual builds and every head observation identify N. The original full plan is preserved with an explicit source-label erratum; application aliases were corrected before its first timing. No timing changed. Development input-setup and plan-generation mistakes are retained separately from successful engine checks.

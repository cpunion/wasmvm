# Native memory departure: 2026-11-04

**Selected for final CI; not merged.** The combined admission/departure change reduces LLGo native warm execution of the compiled C/Rust SHA-256 modules by **29.11% / 28.63%** against production in this new same-cohort comparison. Departure alone contributes **19.47% / 18.86%** against the admission source. Go and LLGo application samples are close after this change; controlled parity and overall performance acceptance remain open.

## Change and validation

Measured engine `9a8d59743a2dd115d18b4998258483a62567449a`, intermediate admission `6187370b0a0b674b1c13db6a0bd349499b84b17f`, baseline engine `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine files identical to production main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). One file changes 17 inserted/3 removed lines against production. Admission snapshots address/size under the mutex, then unlocks explicitly. Departure decrements the active count under that mutex and unlocks directly unless retirement can reclaim storage. A cold `releaseAndUnlock` helper retains deferred unlocking around reclamation. Active/bound storage lifetime, private/shared access, error precedence, cancellation, quotas and ABI stay the same.

Go 1.27.3 and LLGo `3571c2523e010e6842c4dd8a5213554fdc72d2a3`, LLVM 22.1.8, LLGo release `-O3 -lto=full`, macOS arm64. No toolchain/profile, guest, alignment or reference-policy change. Production, admission and combined sources use the same host compilers in both backends. LLGo production/head have identical-binary A/A aliases for each backend.

Seven complete affected packages pass under both compilers, with **96 actual Core/proposal reports**. The compiled application probe additionally passes **64 checksum calls and 48 expected unreachable traps**, including successful reuse after traps and completed store close. Race, cgocheck2 and no-cgo checks pass; no-cgo verifies explicit native unavailability. These checks do not replace final Linux amd64/macOS arm64 × Go/LLGo 31-package qualification and its 30 coverage gates.

Exact scoped Go statement coverage:

| Package | Coverage |
| --- | ---: |
| internal/compiler/codegen | 97.44% |
| internal/compiler/slots | 99.68% |
| internal/engine/execute | 96.35% |
| internal/engine/nativeexec | 92.92% |
| internal/platform/native | 95.03% |
| pkg/component | 91.23% |
| pkg/vm | 93.00% |

Static LLGo `Memory.end` changes from **132 to 68 instructions**, removing two direct `AllocU` call sites from the ordinary departure body. The new cold helper has 115 instructions and retains two allocator call sites. Admission retains the earlier 143→187 instruction change and two→zero direct allocator sites. This describes static emitted sites, not measured allocation events, total code-size improvement or a whole-engine zero-allocation claim. LLGo #2983 remains with its existing contributor; this work does not change SIMD local-array lowering.

## Original20 source comparison

Three fresh process blocks, two warmups/seven interleaved samples, randomized individual calls. The complete original20 cohort verifies **1224 VM/reference observations, 27 native C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes and 417 summary rows**. Counts stay at 1M ordinary iterations, original 10k GC/notify/wait counts and one call-control. Every phase, adverse outcome, source/fixture/SDK/binary identity and A/A control remains in evidence. No previous cohort is replaced or pooled; old unsupported outcomes require exact unchanged binary/fixture identity.

Changes are warm-duration percentages; positive is slower. Production/combined and admission/combined ratios come directly from this one cohort. A/A is production/head LLGo.

| Workload | LLGo native vs production | vs admission | Native A/A | Go native vs production | LLGo interpreter vs production | Interpreter A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | -1.11% | -1.43% | -0.60% / +1.30% | -2.12% | -1.52% | -0.32% / -1.64% |
| core/call-control | +1.19% | -5.61% | -1.24% / -19.06% | -27.18% | -19.06% | +0.00% / +17.65% |
| core/integer-loop | -0.34% | -0.00% | -0.45% / -0.18% | +0.29% | -0.13% | -0.02% / +0.11% |
| core/function-calls | +0.23% | -0.11% | +0.49% / -0.03% | -0.30% | -0.11% | +0.59% / +0.37% |
| core/memory-load-store | -0.56% | -1.13% | -0.19% / +0.87% | -1.17% | -0.51% | -0.99% / +0.59% |
| memory/narrow-unaligned | +2.24% | +0.63% | +0.46% / -0.60% | -0.67% | +0.42% | +0.60% / -0.06% |
| memory/dependent-address | -0.22% | +0.00% | +0.31% / -0.09% | -0.04% | +0.56% | +0.54% / +0.39% |
| memory/float64-unaligned | +0.93% | +0.26% | +1.20% / -0.33% | +2.58% | -0.30% | -1.22% / +0.23% |
| runtime/simd-memory-full | +0.47% | -0.29% | +0.93% / -0.61% | +0.39% | +0.52% | +0.31% / +0.55% |
| runtime/simd-memory-widen | -0.81% | -0.28% | -1.04% / -0.68% | +0.36% | -1.60% | +0.20% / +0.40% |
| runtime/simd-memory-lanes | +0.77% | +0.62% | -0.06% / -0.73% | -0.13% | +1.53% | -0.00% / -1.90% |
| gc/i31-roundtrip | +0.95% | -11.13% | +0.72% / +1.18% | -7.32% | +1.82% | +1.33% / -0.16% |
| gc/struct-access | +4.09% | -4.78% | +7.50% / -0.43% | -17.42% | -0.34% | -1.25% / -1.03% |
| gc/struct-allocation | +6.73% | -5.06% | +14.34% / +0.68% | -2.01% | -3.94% | -4.77% / -1.59% |
| gc/struct-indirect-access | -0.47% | -3.85% | +0.78% / -0.94% | -5.17% | +0.28% | -0.02% / +0.01% |
| threads/shared-memory | -0.24% | -0.52% | -0.12% / +0.58% | +1.39% | -1.14% | +0.81% / +1.07% |
| threads/atomic-rmw | +0.59% | +0.08% | +0.34% / -0.36% | -0.53% | -3.24% | +0.55% / +0.15% |
| threads/notify-empty | -5.94% | -20.75% | -3.56% / +14.51% | -14.25% | -2.37% | -0.71% / -2.21% |
| threads/wait-mismatch | +1.10% | -20.22% | +5.24% / -1.36% | -11.45% | -2.71% | +1.01% / +1.35% |
| threads/wait-timeout | -1.43% | -5.91% | +9.29% / -1.45% | -14.06% | -2.24% | -1.57% / +2.44% |

Retain native narrow unaligned **+2.24%** against production (**+0.63%** against admission), Go native float64 **+2.58%**, and LLGo native GC allocation **+6.73%** (production A/A **+14.34%**). Tiny call/notify/wait and GC controls have substantial variation. Three shared-host blocks do not establish narrow controlled parity, and no timing retry or sample extension was used to replace adverse rows.

## Compiled application comparison

Independent C and Rust authored SHA-256 implementations include message generation, chaining and padding; the checked result is the first 32 digest bits. Each call resets state and processes 5000 blocks. Both guest frontends use LLVM; this is two compiled Core compute modules, not a complete application or WASI-world portfolio. Three process blocks/two warmups/seven samples independently verify **126 VM/reference observations, 6 native C observations, 42 alternate checks and 44 summaries**.

Warm ns/message block; A/A percentages are production/head LLGo native.

| Guest | LLGo production → admission → combined | Combined vs production | Departure increment | Native A/A | Go combined | LLGo/Go |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | 685.792 → 603.642 → 486.133 | -29.11% | -19.47% | -0.71% / +0.35% | 484.233 | 1.004x |
| applications/sha256-rust | 681.758 → 599.625 → 486.550 | -28.63% | -18.86% | +0.45% / +1.63% | 487.850 | 0.997x |

The remaining competitor, host-compiler and native-C gaps use this same cohort:

| Guest | LLGo native | Wasmtime | Wasmer | wazero compiler | Native C | LLGo / Wasmtime | LLGo / C | LLGo interpreter | WasmEdge interpreter | wazero interpreter |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | 486.133 | 191.333 | 186.425 | 290.708 | 161.233 | 2.541x | 3.015x | 17405.292 | 25939.417 | 39822.792 |
| applications/sha256-rust | 486.550 | 184.667 | 188.275 | 319.058 | 159.983 | 2.635x | 3.041x | 17343.108 | 24247.500 | 40073.225 |

LLGo/Go warm native ratios are approximately 1.004x and 0.997x here. This removes the earlier large host-compiler difference on these inputs, not all Go/LLGo differences. Native execution still takes about 2.5–2.6x Wasmtime, 1.5–1.7x wazero compiler and 3x unchecked C. Native C omits VM guards, fuel, cancellation and lifecycle; reference policies stay separately identified. Interpreter rows are shown separately from compiled runtimes.

## Startup and preparation stay separate

Median milliseconds for selected native configurations. First call executes all 5000 blocks; warm is the complete warm call, not per-block latency. Compilation includes the adapter's validate/compile phase and is not subtracted from a process total. All other phase/reference rows remain in evidence.

| Guest | Runtime | Process ready | Validate/compile | Instantiate | First call | Warm call | Cleanup |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | wasmtime-cranelift | 4.8655 | 1.2763 | 0.0644 | 0.8996 | 0.9567 | 0.1775 |
| applications/sha256-clang | wazero-compiler | 4.6604 | 0.2913 | 0.0159 | 1.3180 | 1.4535 | 0.0118 |
| applications/sha256-clang | wasmer-cranelift | 6.3765 | 1.7189 | 0.0330 | 0.8810 | 0.9321 | 0.0679 |
| applications/sha256-clang | head-go-native | 4.1629 | 1.6742 | 0.0245 | 2.5099 | 2.4212 | 0.0160 |
| applications/sha256-clang | base-llgo-native | 95.5833 | 3.5044 | 0.0230 | 3.5389 | 3.4290 | 0.0056 |
| applications/sha256-clang | head-llgo-native | 93.6377 | 3.4691 | 0.0220 | 2.6883 | 2.4307 | 0.0087 |
| applications/sha256-rust | wasmtime-cranelift | 4.7174 | 1.2609 | 0.0578 | 0.9084 | 0.9233 | 0.1324 |
| applications/sha256-rust | wazero-compiler | 4.3774 | 0.3369 | 0.0182 | 1.4648 | 1.5953 | 0.0107 |
| applications/sha256-rust | wasmer-cranelift | 6.3831 | 1.6257 | 0.0348 | 0.9225 | 0.9414 | 0.0659 |
| applications/sha256-rust | head-go-native | 4.3254 | 1.7986 | 0.0192 | 2.3715 | 2.4392 | 0.0186 |
| applications/sha256-rust | base-llgo-native | 89.7680 | 3.0322 | 0.0237 | 3.4653 | 3.4088 | 0.0115 |
| applications/sha256-rust | head-llgo-native | 90.2861 | 3.4762 | 0.0227 | 2.3981 | 2.4327 | 0.0156 |

LLGo readiness remains approximately 90ms versus Go 4ms in these samples. This is a separate initialization gap, not guest execution cost or an isolated new compiler defect. Rust LLGo validate/compile is **+14.64%** against production (A/A production/head **+9.16% / -4.86%**), and cleanup **+36.36%**. C-frontend cleanup is **+55.97%**, with large A/A variation. Keep these observations open for controlled latency investigation; warm improvements do not erase them. The shared host was busy (all before/after load values are retained), so these measurements are descriptive rather than controlled acceptance.

## Disposition

Keep this compact combined change for final target CI and review. Production main is unchanged. The sole existing external contribution [LLGo #2991](https://github.com/xgo-dev/llgo/pull/2991) remains open with 112 successful/3 skipped checks at the latest read. No external upstream merge is authorized. The owner question about a second contribution PR remains pending, so no parallel PR was opened.

Next architectural work should reduce ordinary guest Call/Return transitions through the host while retaining explicit bounded frames, precise root maps, quotas, cancellation, metering and complete runtime semantics. The two compiled inputs have 5001 logical calls; the earlier diagnostic found the same guest machine bodies under both host compilers. These observations motivate investigation, not a claimed independent call-cost measurement or predicted speedup. Do not replace the shared frame machine or move guest recursion onto the hardware stack.

[Lossless evidence](evidence/performance-native-memory-departure-2026-11-04.json.xz) contains the complete engine patch, exact source/toolchain/SDK identities, all scoped semantic/coverage/pointer checks, disassembly, both complete new cohorts, source/fixture identities, independent verifiers and compiled application inputs. Executables, SDK binaries and archived engine build trees are excluded. Full production CI, controlled compiler/platform performance and expanded application/WASI-world acceptance remain unfinished.

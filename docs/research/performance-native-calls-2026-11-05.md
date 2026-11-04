# Native guest calls: 2026-11-05

**Native throughput improves substantially; interpreter regressions remain open. Not merge-ready.** Bounded direct native guest calls reduce LLGo warm execution of the compiled C/Rust SHA-256 modules by **60.76% / 64.95%** against production in one new same-cohort comparison. Their durations are **0.851x / 0.723x wazero compiler**, **1.363x / 1.349x Wasmtime** and **1.562x / 1.469x unchecked native C**. The Go interpreter worsens **12.31% / 10.76%**, and LLGo interpreter **8.22% / 5.56%**; these outcomes prevent an overall performance acceptance claim.

## Implementation and scope

Measured engine `2e0c605b2570696b8eed9fe796fe53ca3529ac12`, intermediate memory optimization `9a8d59743a2dd115d18b4998258483a62567449a`, production engine `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-file identical to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). The complete source event changes 17 engine files, including tests and the preceding shared frame-transfer foundation. No compiler/profile, guest, alignment or reference-policy change is included.

Native calls branch into callee body headers and return through saved compiler-owned continuations while retaining one fixed C hardware frame. Bounded parent metadata carries function IDs, slot bases, call PCs and immutable result-move plans. Full values are spilled/copied before observable boundaries. The existing common frame machine still owns host, GC, exception and continuation effects. Both policy tiers preserve exact fuel and current invocation quotas. Shared module runtime views keep global aliases and heap/memory offsets coherent across callees. The C adapter implements no guest call semantics.

Frame initialization uses bounded arm64 vector blocks or x86 `rep stosq`, checking cancellation at most every 4 KiB. Large tuples copy full 128-bit values with cancellation checks. Self-recursion updates the current frame base even when the function index stays unchanged, including large Move exits. Metadata starts with at most 16 parents and grows at idle boundaries; native bytes are 32+64×reserved records, bounded by workspace limits. Code-budget exhaustion falls back to complete ordinary compilation. See the [architecture](../design/native-calls.md).

The earlier scalar-initialization source `60bf043` passed semantics before the vector implementation. No timing was taken for that intermediate. Its qualification and the initial rejected test fixture (missing Jump Count=1, corrected before execution) remain in evidence. There was no timing retry or selection across initialization variants.

## Exact qualification

Go 1.27.3, LLGo `3571c2523e010e6842c4dd8a5213554fdc72d2a3`, LLVM 22.1.8, LLGo `-O3 -lto=full`, macOS arm64. All nine complete affected packages pass under both host compilers, producing **96 actual Core/proposal reports**. Generated-code tests execute large mixed tuples, nested host/root publication, parent-only roots under forced guest GC, 40-level self-recursion, metadata growth, depth exhaustion, cancellation and reuse. The compiled application probe passes **64 checksum calls / 48 expected unreachable traps**, including reuse and close. Race, cgocheck2 and no-cgo checks pass; unavailable-native checks are not native execution evidence.

| Package | Scoped Go coverage |
| --- | ---: |
| internal/compiler/codegen | 97.38% |
| internal/compiler/nativeabi | 100.00% |
| internal/compiler/slots | 99.68% |
| internal/engine/execute | 96.63% |
| internal/engine/nativeexec | 92.61% |
| internal/engine/state | 95.21% |
| internal/platform/native | 95.03% |
| pkg/component | 91.20% |
| pkg/vm | 93.00% |

Both ISAs are emitted and their relocations checked locally. A supplementary fresh Linux amd64 container executes the generated-code tests under **Go 1.27.0 and emulation** (pinned image digest and logs retained). This confirms execution beyond an emitter-only check but does not replace final pinned Go 1.27.3/LLGo Linux CI or hardware performance acceptance. The container completed successfully and was removed. Required final qualification remains Go/LLGo × Linux amd64/macOS arm64, 31 packages/profile and 30 coverage gates.

## Original20 cohort

One new fixed three-source cohort: production, memory optimization, native calls. Three fresh process blocks, two warmups/seven randomized interleaved samples; LLGo production/head have identical-binary A/A for each backend. **1224 VM/reference observations, 27 native-C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes and 417 summaries** independently verify. Original counts, adverse rows, host loads, source/binary/fixture/SDK identities and every phase remain. Older cohorts are not pooled or substituted. Positive duration changes are slower.

| Workload | LLGo native vs production | Native calls vs memory | Native A/A | Go native vs production | LLGo interpreter vs production | Interpreter A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | +4.48% | +0.08% | +0.94% / -2.76% | +0.36% | +3.30% | -1.17% / -0.99% |
| core/call-control | +16.36% | +42.39% | -31.73% / -28.92% | -3.12% | -22.87% | +7.23% / +79.68% |
| core/integer-loop | +1.09% | -3.08% | +0.44% / -0.65% | -0.44% | +1.48% | +2.53% / -0.81% |
| core/function-calls | -0.02% | -0.73% | +0.65% / +1.15% | +0.62% | +0.22% | -0.01% / +0.02% |
| core/memory-load-store | -2.00% | +0.00% | -1.90% / +0.17% | +0.59% | +3.31% | -0.05% / +0.52% |
| memory/narrow-unaligned | +0.72% | -0.38% | +0.83% / -0.39% | +0.36% | +3.74% | -0.98% / -0.27% |
| memory/dependent-address | +1.04% | +0.66% | +1.00% / +0.02% | +0.00% | +1.19% | -1.01% / +0.16% |
| memory/float64-unaligned | -1.12% | -0.43% | +0.47% / +0.69% | -0.11% | +5.69% | -0.38% / -0.59% |
| runtime/simd-memory-full | +0.56% | +0.11% | +0.12% / +0.09% | -0.19% | +1.09% | -0.19% / +0.11% |
| runtime/simd-memory-widen | -1.16% | +0.08% | -0.78% / +0.16% | -1.29% | +1.82% | -0.31% / +0.18% |
| runtime/simd-memory-lanes | -0.14% | +1.35% | -2.04% / +0.25% | +1.06% | +2.95% | +0.59% / +0.06% |
| gc/i31-roundtrip | +6.24% | -1.12% | +8.86% / +2.46% | +8.46% | +2.48% | -1.94% / -0.31% |
| gc/struct-access | +9.88% | +11.58% | -5.32% / -1.90% | +7.02% | +6.53% | +1.22% / -0.52% |
| gc/struct-allocation | -1.50% | -6.27% | +5.24% / +0.89% | -1.49% | -1.72% | -4.87% / -1.23% |
| gc/struct-indirect-access | -1.28% | -1.91% | -6.32% / +4.68% | +5.55% | +0.68% | -0.52% / -2.81% |
| threads/shared-memory | +0.01% | +0.78% | -1.21% / -1.00% | +2.40% | +6.44% | -1.68% / -1.22% |
| threads/atomic-rmw | +0.71% | +0.30% | +0.47% / -0.12% | +0.33% | -0.30% | -0.54% / -0.37% |
| threads/notify-empty | +15.08% | +4.29% | +3.54% / -7.97% | -17.87% | +0.96% | +1.35% / +3.76% |
| threads/wait-mismatch | -4.93% | -2.41% | -7.27% / +13.33% | -7.15% | +2.39% | +1.49% / -0.35% |
| threads/wait-timeout | +16.29% | +4.04% | +3.26% / +0.43% | -4.63% | -2.24% | -4.70% / -1.52% |

Native GC struct access is **+9.88%** against production (**+11.58%** against memory); native short notify **+15.08%** and wait-timeout **+16.29%** remain unresolved. Single-call/short-count and GC controls have material variation; they are retained rather than silently waived. Sustained ordinary native cases are mostly close to their controls. Three busy shared-host process blocks are descriptive and do not establish controlled parity.

## Compiled application cohort

Independent C and Rust authored SHA-256 modules process 5000 message blocks per call, including message generation, chaining and padding; the independent oracle checks the first 32 digest bits. Both frontends use LLVM. This is two compiled Core compute modules, not a complete application/WASI-world portfolio. **126 VM/reference observations, 6 C observations, 42 alternate checks and 44 summaries** verify separately from original20.

Warm ns/message block; all three sources occur in this one cohort.

| Guest | LLGo production → memory → native calls | vs production | vs memory | Go native calls | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | 744.400 → 569.925 → 292.067 | -60.76% | -48.75% | 268.317 | 214.242 | 224.700 | 343.225 | 186.933 |
| applications/sha256-rust | 781.267 → 567.742 → 273.833 | -64.95% | -51.77% | 294.917 | 203.025 | 217.108 | 378.600 | 186.367 |

| Guest | Native A/A production/head | Go interpreter production → head | Change | LLGo interpreter production → head | Change | Interpreter A/A | WasmEdge interpreter | wazero interpreter |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | +2.19% / -4.97% | 24659.717 → 27694.708 | +12.31% | 19181.925 → 20759.167 | +8.22% | +2.00% / +3.98% | 28942.408 | 45412.658 |
| applications/sha256-rust | -3.23% / +0.45% | 24284.275 → 26896.733 | +10.76% | 18795.825 → 19841.208 | +5.56% | -2.76% / +1.69% | 26270.942 | 44417.433 |

The new native mechanism approximately halves duration again against the memory-only source. It is faster than wazero compiler on both modules but still slower than Wasmtime/Wasmer/native C. LLGo/Go differs by module and its controls; a universal host-compiler win is not claimed. Interpreter regressions exceed their application A/A differences and require investigation. Interpreter output still beats the named interpreter references; that does not erase the own-baseline regression.

## Lifecycle measurements stay separate

Median milliseconds. First/warm call executes all 5000 blocks; warm is the complete call. Compile includes the adapter's validation/compilation phase. No startup/clock/median subtraction manufactures pure instruction cost; C omits VM safety/cancellation/fuel/lifecycle policies. All policies and remaining phases stay in raw evidence.

| Guest | Runtime | Process ready | Validate/compile | First call | Warm call | Cleanup |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | wasmtime-cranelift | 5.2309 | 1.2901 | 1.0269 | 1.0712 | 0.1438 |
| applications/sha256-clang | wazero-compiler | 4.6747 | 0.3002 | 1.3453 | 1.7161 | 0.0157 |
| applications/sha256-clang | wasmer-cranelift | 6.7161 | 1.7014 | 1.1160 | 1.1235 | 0.0687 |
| applications/sha256-clang | head-go-native | 4.7077 | 2.1729 | 1.4445 | 1.3416 | 0.0087 |
| applications/sha256-clang | base-llgo-native | 101.7652 | 3.5730 | 3.9442 | 3.7220 | 0.0088 |
| applications/sha256-clang | head-llgo-native | 98.5100 | 3.6919 | 1.2313 | 1.4603 | 0.0100 |
| applications/sha256-clang | memory-llgo-native | 107.0295 | 3.5727 | 2.7375 | 2.8496 | 0.0073 |
| applications/sha256-rust | wasmtime-cranelift | 5.1410 | 1.3123 | 1.0993 | 1.0151 | 0.1824 |
| applications/sha256-rust | wazero-compiler | 4.9888 | 0.3943 | 1.7041 | 1.8930 | 0.0099 |
| applications/sha256-rust | wasmer-cranelift | 7.1406 | 1.7298 | 1.1147 | 1.0855 | 0.0676 |
| applications/sha256-rust | head-go-native | 5.3480 | 2.0783 | 1.4675 | 1.4746 | 0.0112 |
| applications/sha256-rust | base-llgo-native | 109.1195 | 3.7238 | 3.6503 | 3.9063 | 0.0130 |
| applications/sha256-rust | head-llgo-native | 112.3384 | 3.9391 | 1.2646 | 1.3692 | 0.0115 |
| applications/sha256-rust | memory-llgo-native | 103.4780 | 3.7575 | 2.3745 | 2.8387 | 0.0130 |

LLGo readiness remains about 100ms versus Go 5ms; it is a separate unresolved initialization gap. Shared-host load and each preparation/cleanup change remain visible. Native warm improvements do not establish preparation or readiness acceptance.

## Follow-up and disposition

Retain the native mechanism and investigate the interpreter regression before integration. The shared dispatcher gained the native frame packet branch; its effect on generated host code is a candidate explanation, not an established cause. Static linked LLGo `Execution.Run` grows 2729→2929 instructions and 32→34 direct allocation sites; Go changes 4176→3166 instructions. These are static sites/body sizes, not dynamic allocation counts or proof of the throughput cause. The next architecture step should isolate backend frame transfer from portable dispatch while preserving identical runtime semantics; do not tune padding or choose favorable timings.

No benchmark/compiler failure occurred in the new qualified cohorts. One copied verifier metadata field initially said one engine file; it was corrected from the exact source manifests to 17 while raw timings and the original verifier log remain preserved. No new upstream issue is inferred from timing or static site counts, and LLGo #2983 remains with its existing contributor. Production main is unchanged. The sole existing external PR2991 remains open; the earlier parallel-PR question has no answer, so no additional PR was opened.

[Lossless evidence](evidence/performance-native-calls-2026-11-05.json.xz) includes the exact patch, qualified source/toolchains, both complete cohorts, verifiers, semantic/coverage/pointer reports, generated-code tests, static dispatch disassembly, compiled application inputs, source/provenance corrections and the supplementary emulated-ISA diagnosis. Executables, SDKs and archived build trees are excluded. Final pinned target CI/review, controlled performance and expanded application/WASI-world acceptance remain open.

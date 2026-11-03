# LLGo SIMD intrinsic local-buffer escape: 2026-11-03

Independent minimal reproduction reported as [LLGo #2983](https://github.com/xgo-dev/llgo/issues/2983).
The compiler emits actual NEON add.2d, but the three nonescaping local array
buffers around array load/store intrinsics each retain an AllocU(16) call.
Go allocates zero; caller-provided buffers allocate zero under both compilers.

Pinned Go 1.27.3, LLGo `983465e` with `-O3 -lto=full`, LLVM 22.1.8/macOS arm64.
This independent diagnostic enables `GOEXPERIMENT=simd`; the project release
profile remains unchanged. The compiler's array declarations carry noescape,
and the minimal function has scalar parameters and an unnamed scalar result,
separating it from the previously reported aggregate-parameter/named-result cases.

| 100,000 calls | Round 0 bytes | Round 1 | Round 2 |
| --- | ---: | ---: | ---: |
| Go local arrays | 0 | 0 | 0 |
| LLGo local arrays | 9,596,928 | 9,601,952 | 9,596,640 |
| Go supplied buffers | 0 | 0 | 0 |
| LLGo supplied buffers | 0 | 0 | 0 |

Every round checks checksum 5001650000. Reported LLGo TotalAlloc is approximate;
three allocator calls in final linked assembly independently corroborate the
roughly 96 bytes/call. No Mallocs or throughput claim is inferred. The wider
four-width add/sub probe separately passes 524,288 reference comparisons under
both compilers and reproduces the approximate allocation. Initial use of amd64
bitcast method names on arm64 and an accidentally included saved source file
are retained setup failures, not LLGo compatibility findings.

Current upstream `10ebf99` has no cl/ssa implementation changes since the tested
pin, but was not rebuilt. General heap-to-stack #2244 is related; aggregate
parameter #2974 and named-result #2959 are distinct reproductions. No compiler
pin or flags are changed to hide the allocation. Native host SIMD is a potential
next step only after its compiler/profile and complete VM acceptance are measured.

[Minimal/wider programs, commands, original outcomes, final assemblies, toolchain and submitted issue](evidence/performance-llgo-simd-array-escape-2026-11-03.json.gz).

# LLGo native allocation: redundant clearing

Reported as [LLGo #2966](https://github.com/xgo-dev/llgo/issues/2966).
Pinned LLGo `8e45ed5` with `-O3 -lto=full` emits an actual BDWGC `GC_malloc`
call followed by `bzero` for a dynamic `make([]byte, n)`. BDWGC 8.2.12's
non-atomic allocation contract already guarantees cleared storage. Current
upstream `08e759300a2d523d5a37de0b3b31a78b20bd2843` retains the same AllocZ
implementation; it was inspected, not rebuilt. This is an independently
identified redundant runtime operation, not an inference from guest throughput.

A minimal Go program checks zero bytes before modifying the result. Final
Mach-O disassembly and indirect-symbol bindings identify both calls. The native
BDWGC implementation is distinct from no-GC, bare-metal, Wasm and aligned
fallback allocators, whose clearing contracts must be checked separately.

An independent C mechanism diagnostic uses the same BDWGC library, Clang
22.1.8 `-O3`, Apple M4 Max/macOS arm64, matching noinline allocator wrappers
and null checks, zero-byte observations and identical writes. Three fresh
rotated process blocks each use two warmup and seven measured batches.
Collection is disabled within a batch and forced between batches.

| Request | Allocations/batch | GC_malloc | GC_malloc plus clear |
| --- | ---: | ---: | ---: |
| 64 B | 1,000,000 | 7.238 ns/allocation | 8.704 ns/allocation |
| 4096 B | 10,000 | 299.0 ns/allocation | 341.5 ns/allocation |

The initial C diagnostic had an unmatched null check in the clearing wrapper.
It remains retained, but these values and the updated issue use the corrected
matching-check source. These shared-host descriptive observations quantify only
the allocator mechanism. They do not measure an LLGo compiler patch, establish
a wasmvm speedup, or explain the full LLGo/Go GC-allocation difference. The
production pinned toolchain remains unchanged. Any future compiler upgrade
requires its own source-identity and baseline event.

[Lossless evidence](evidence/llgo-native-redundant-clear-2026-11-02.json.gz)
includes both diagnostic sources/runs, all raw batches, toolchain/library
identities, final disassembly/bindings, the minimal Go reproduction, the upstream
source inspected and the issue body.

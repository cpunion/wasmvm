# Bulk memory fill: 2026-11-02

Bounded bulk primitives reduce fill overhead in both execution backends. The
remaining native runtime-helper transitions are expensive; **competitive
performance acceptance remains open**.

## Method

Base `380a4ec08f6f29283801702aac7474031f5b2ec3` is the validated head of merged
PR #89. Measured head `c5da50069533f43a21afa8a8b73bbdfdb3af8e7b` contains the
implementation; the subsequent baseline merge and documentation do not change
engine sources. Both revisions are built independently with Go 1.27.3 and
LLGo `8e45ed5` (`-O3 -lto=full`) on Apple M4 Max, macOS arm64.

The existing bulk-overlap module fills 8 KiB and performs an overlapping 8 KiB
copy per source iteration. Its binary, 1,000-iteration configuration and
independent checksum oracle are unchanged. The private-memory module is an
unaffected control. Each pair has three interleaved fresh process blocks,
two warmups and seven measured calls per process, plus alternate-seed/count
checks. Statistics are medians of process medians. Every call validates its
checksum. Startup, loading, preparation, instantiation, lookup, first call,
warm calls and cleanup remain separately timed; no timing is subtracted.

References are Wasmtime 48.0.1 Cranelift, wazero 1.12.0 (compiler/interpreter),
Wasmer 7.3.0 Cranelift and WasmEdge 0.17.1 interpreter. Executable and SDK library
hashes match the pinned survey inventory. These are descriptive workstation
measurements, without controlled equivalence or regression acceptance.

## Results

Warm bulk workload, **ns/source iteration; lower is better**:

| Compiler/backend | Base | Head |
| --- | ---: | ---: |
| Go interpreter | 2,411.250 | 334.333 |
| Go native | 2,529.750 | 486.666 |
| LLGo interpreter | 2,638.334 | 431.167 |
| LLGo native | 3,261.250 | 1,207.625 |

| Reference | Warm ns/source iteration |
| --- | ---: |
| Wasmtime Cranelift | 155.000 |
| Wasmer Cranelift | 159.000 |
| wazero compiler | 185.250 |
| WasmEdge interpreter | 230.000 |
| wazero interpreter | 265.000 |

LLGo native remains 7.79x Wasmtime; its native helper transitions must be
addressed separately. Private-memory control observations and every preparation
phase are retained in the [evidence](evidence/performance-bulk-2026-11-02.json).
An earlier copy-only experiment also remains visible: LLGo interpreter reached
675 ns/iteration and native 1,200; native performance did not materially change
when replacing that fill primitive with the LLVM-friendly loop.

## Mechanism and validation

The complete destination range is checked before any write. Zero fills use
`clear`. For nonzero fills, Go expands an initialized slice prefix using bulk
copies; LLGo iterates over the already bounded slice, enabling LLVM's bulk-fill
optimization. Pure primitive measurements retain the original bounded loop,
slice loop and copying algorithms at 32, 256, 4,096 and 8,192 bytes. These
diagnostics explain compiler-specific choices; they do not replace the actual
Wasm workload comparison.

Both compilers share range validation, shared-memory locking and execution
semantics. Runtime cancellation still polls between at most 4,096-byte chunks,
and overlapping copies retain their direction and whole-operation bounds
checks. No native ABI or compiler-specific execution framework is added.

Go and LLGo tests check every byte and canary for private, native and shared
memory, zero/nonzero values, tails, chunk boundaries, empty ranges and rejected
overflow/out-of-bounds ranges without partial writes. Affected Core 2/3,
Threads and Custom Page Sizes inventories pass under both compilers; memory
race tests pass. Final target CI remains required before merging.

The evidence retains all raw timings, orders, checksums, phase records,
source/compiler/binary/SDK hashes, the earlier experiment and primitive probe.
Local scripts and complete source inventories are under
`artifacts/bulk-fill-20261102`, with hashes in the committed evidence.

# Incremental native frame transfers

Native execution now publishes only the changed suffix of an authenticated
same-instance call chain after a host return. A private header word tracks the
lowest changed frame; pops lower it and later pushes already belong to that
suffix. The existing common frame machine validates the transferred suffix and
keeps its untouched prefix. No public API or second executor is introduced.

Depth/slot quotas, physical metadata bounds, continuation authentication,
reference roots, foreign reentry and retirement remain enforced. The native
backend retains the full chain for authentication; host-array growth preserves
the prefix. A new 17/18-frame differential test covers repeated host calls,
pushes/pops, both growth paths, forced host GC, reference roots, metered fuel,
quota failure and reuse. Successful transfers contain 24 frame descriptions
rather than the former 88. Malformed dirty indices and overflowing metadata
are rejected. The [ABI design](../design/native-abi.md#incremental-native-frame-publication)
describes ownership and lifetime.

## Complete request comparison

M4 Max/macOS arm64, Go1.27.3, LLGo ac129220 with `-O3 -lto=full`, LLVM22.1.8,
normal GC. Baseline 1ebf33e5 versus production-code candidate 4927def4 uses the
same Rust P3 HTTP guest, four Stores, sixteen clients, immutable snapshot input
and buffered output. The independent Wasmtime48.0.1 adapter remains frozen.
LLGo base/head are freshly linked with symmetric five-file Git overlays in the
same existing worktree; no full source snapshots are created.

Each payload has a separate, fixed 20-block/200-process event: Go base/head,
LLGo base/head and Wasmtime, each with an identical-binary alias. Twenty
512-request warm batches are individually interleaved; first call, two warmups,
final flush, preparation and cleanup remain separate. All 400 final processes
and 4,096,000 measured responses pass body/status/header/trailer/CRC checks.
No owned build/test/profile/source edit or pruning overlaps either event.
Intervals are individual paired-bootstrap diagnostics, not portfolio acceptance.

| Go head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU ratio, 95% interval | 0.98746 [0.98026, 0.99423] | 0.98926 [0.98385, 0.99400] |
| Instruction ratio | 0.97211 [0.97023, 0.97367] | 0.96934 [0.96835, 0.97039] |
| Throughput ratio | 1.00909 [0.99962, 1.01859] | 1.00677 [0.99701, 1.01574] |

Both Go CPU/instruction controls resolve 2.5% in both events. CPU and executed
instruction work decrease; throughput improvement is not established. Managed
allocation counts remain approximately 696/756 per request, and median GC
collections remain 35/76 per 10,240 requests. This supports reduced frame work
rather than an allocation reduction.

Go CPU remains 2.098x/2.185x Wasmtime; instruction counts remain 3.790x/3.623x.
These are complete-service comparisons with different adapters/providers and
recorded policy differences. Managed counters exclude Rust/C allocations and
cannot establish a cross-language total-allocation ranking.

LLGo CPU head/base points are 1.00837 [0.98734, 1.03086] and
0.99468 [0.97995, 1.00893]. Both baseline CPU A/A controls fail the 2.5% check;
there is no accepted LLGo CPU improvement or established regression. Median
collections agree at 24/36, but process ranges differ. The 16 B adverse CPU
point, lower throughput points and all cold-phase adverse points remain in the
data. LLGo compile-wall ratios of 1.0083/1.0152 also cross 1 and have unresolved
controls. No favorable retry or cohort extension was added. Its CPU remains
2.731x/2.359x Go; causal compiler/GC work remains open.

## Independent call kernels

Two additional fixed 20-block events use public VM APIs and the same three
pinned compiler references: wazero1.12.0, Wasmtime48.0.1 and Wasmer7.3.0. The
indirect-leaf kernel uses 20 million source iterations; a looped callee forces
the general call path for five million iterations. Each process has seven warm
calls after two warmups, with all results checked. The generic comparison also
reuses the ordinary LLVM C affine recurrence for twice as many scalar steps,
normalized per guest outer iteration; C omits VM contracts. Preparation, first
call and process readiness remain separate. The owned diagnostic WAT, binaries'
hashes, independent affine oracle and exact commands are retained.

| Warm ns/source iteration | Go base | Go head | wazero | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Indirect leaf | 0.9010 | 0.8443 | 2.1373 | 1.1483 | 0.8278 | — |
| General call, two affine steps | 8.4256 | 7.8604 | 2.3292 | 2.2044 | 2.2006 | 2.1948 |

Paired head/base ratios are 0.93900 [0.92958, 0.94932] and
0.94435 [0.93636, 0.95265], with both own controls resolving 2.5%. These kernel
wins may include generated-code layout effects; their cause is not isolated by
this event. The leaf kernel is still 1.019x Wasmer, and its wazero control is
unresolved. General calls remain 3.431x wazero, 3.580x Wasmtime, 3.606x Wasmer
and 3.610x unchecked C; all generic-kernel controls resolve 2.5%.

The general call gap is the next substantial compiler/runtime target. Isolate
parameter/result frame-memory traffic, frame initialization and call checks
before choosing register argument passing or broader inlining. This transfer
change does not claim to solve that gap.

## Validation and evidence

All 32 Go test packages pass, with all 31 executable packages above 90% statement
coverage (nativeexec 93.14%, codegen 97.93%). Pinned LLGo executes the native
frame/resumption suite. Go vet and nativeexec/execute/vm race tests pass. The
unchanged eight-shard inventory covers all 32 packages exactly once; the final
Go/LLGo × Linux amd64/macOS arm64 CI head must pass before merging.

[Identities, all controls, cold phases and adverse points](data/performance-native-frame-deltas-2026-11-14.json),
[all 968 projected screen/final process records](evidence/performance-native-frame-deltas-2026-11-14.jsonl.gz)
and [protocol, diagnostic fixture, plans and validation logs](data/performance-native-frame-deltas-2026-11-14-protocol.tar.gz)
retain the independent events. Only unused per-request latency vectors are
projected to counts/extrema/exact hashes; full originals remain local. No
historical events are pooled. Competitive, full-portfolio and release
acceptance remain open.

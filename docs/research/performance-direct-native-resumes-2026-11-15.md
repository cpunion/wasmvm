# Direct continuations after internal native returns

Unmetered compiler-owned returns restore the ordinary caller environment and
branch to their exact original PC, avoiding arbitrary host-continuation
dispatch. Frame/root state, memory, scalar/vector constants, runtime bindings,
heap-cache keys and the continuation's register plan are restored. Complete
parameter/result copies, quotas, cancellation and root publication remain.
Metered, host and foreign entries retain their existing dispatcher. See the
[ABI contract](../design/native-abi.md#checked-internal-call-entries).

The implementation reuses literal/binding selection plans when emitting private
stubs; it does not repeat whole-function analysis per call. Stubs are bounded
by call sites and charged to the existing code budget. Emit late indirect-call
exits before resolving their saved return addresses.

## Source and protocol

Measured base 5e25ac7d has the same tree as integrated a2aa727a. Measured head
e9383987 has the same tree as rebased af30e410. Five-file symmetric Git overlays
build eight Go/LLGo phase and P3 executables in the existing worktree. Go 1.27.3,
LLGo 6a482f7f, LLVM 22.1.8, `-O3 -lto=full` and default GC are fixed. M4 Max,
macOS arm64; independent references remain wazero 1.12.0, Wasmtime 48.0.1 and
Wasmer 7.3.0. No owned build/test/profile/edit/pruning overlaps timed cohorts.

The general-call event has 20 fixed blocks/320 fresh processes: Go and LLGo
base/head, three references and unchecked native C, each with a binary alias.
Each process retains first call, two warmups, seven interleaved five-million-
iteration calls and all checksums. C performs 2N equivalent affine steps,
normalized per guest iteration, with no VM limits/cancellation/frame contract.

Go head/base warm time is **0.96834 [0.94758, 0.98903]**; most controls remain
unresolved at 2.5%. Head/reference points are **3.233x wazero, 3.384x Wasmtime,
3.398x Wasmer and 3.415x C**. LLGo head/base is 1.00798 [0.98768, 1.03123],
also with unresolved controls. These points do not establish competitive parity.

A separate fixed 10-block/40-process audit measures CPU, instructions and cycles
over 20 interleaved five-million-iteration calls plus a separate flush. Both
own controls resolve 2.5%. Go head/base results are:

| Metric | Ratio and individual 95% interval |
| --- | ---: |
| CPU | 0.98109 [0.97752, 0.98480] |
| Cycles | 0.98020 [0.97709, 0.98338] |
| Instructions | 0.95764 [0.95758, 0.95771] |
| Cycles/instruction | 1.02355 [1.02033, 1.02687] |

The final CPU reduction is **1.89%**, not the initial prototype's 5.13% from its
separate audit. Different source/layout and host cohorts remain distinct;
neither event is substituted or pooled. Removing instructions does not promise
proportional CPU savings, and exact dependency/layout attribution stays open.

## Complete P3 requests

Each payload uses 20 fixed blocks/200 fresh processes: Go/LLGo base/head and
frozen Wasmtime, each with a binary alias. Twenty 512-request batches are
individually interleaved with four Stores/sixteen clients, snapshot input and
buffered output. Body/status/header/trailer/CRC checks cover every response.
Preparation, first call, two warmups, final flush and cleanup remain separate.

| Go head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU | 1.00051 [0.99538, 1.00536] | 0.99669 [0.99069, 1.00240] |
| Instructions | 0.99589 [0.99482, 0.99699] | 0.99406 [0.99229, 0.99576] |
| Throughput | 0.99948 [0.97843, 1.02461] | 1.00782 [0.99036, 1.02784] |

Both Go CPU controls resolve; CPU improvement is not established. Throughput
controls remain unresolved. Go/Wasmtime CPU points remain 2.231x/2.393x, with
unresolved reference CPU controls. Managed allocations remain approximately
696/756 per request; these counters omit Rust/C allocations.

LLGo CPU ratios are 0.97937 [0.96423, 0.99463] and
1.02963 [0.98179, 1.07911]. The 16 B own CPU controls resolve; the adverse 64 KiB
point and its controls remain unresolved. Allocations/request stay 1165/1430.
The 64 KiB initial heap median changes 56.04 MB to 35.56 MB, with collection
medians 60 to 62 per 10,240 requests and wide ranges. This matches the collection
sensitivity tracked in [LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106).

One separate bounded no-GC 64 KiB diagnostic has 10 blocks/40 processes and five
512-request measured batches plus a flush, requiring zero measured collections.
LLGo instructions become **0.99768 [0.99636, 0.99897]**, with both instruction
controls resolving. CPU is 0.99839 [0.98345, 1.01179], with unresolved controls.
Collection work contributes to the adverse normal point; exact heap sizing
versus conservative roots remains open. Production GC is unchanged.

The 64 KiB Go/LLGo compile-wall points increase 5.26%/7.45%; all cold controls
remain unresolved. Private stubs add emitted code while sharing selection work.
Exact cold scaling/publication effects remain follow-up work; no cold win or
accepted exact regression magnitude is claimed.

## Rejected experiments and validation

Scratch-vector transfer of small complete results was rejected: an independent
40-process audit gives CPU +11.89%, instructions -8.00%, cycles +12.00% and
cycles/instruction +21.74%, with both own controls resolving. A separate
60-process scalar-pair intervention is worse still: CPU +18.48%, instructions
-7.06% and cycles +18.49%; all own controls resolve. Neither is production.
The scalar result rules out attributing the whole loss to vector traffic;
exact dependencies/layout remain unresolved. Historical screen templates have
stale scope prose; actual commands, hashes, sourceRevision records and overlays
identify the variants. The archive retains these original records and errata.

Expanded testing caught the first resume prototype using callee-owned cached
bindings during memory preparation. A focused shared-memory case crashes at
address 7654321, the callee's constant, before correction. Prepare memory from
the trusted view before reloading caller caches, matching the ordinary entry.
The corrected focused direct/indirect/reference cases preserve mixed scalar/
128-bit results and caller constants, in both policies and repeated calls.
The complete official wasi-threads suite also passes. Failed logs remain.

All 32 Go root test packages pass; all 31 measured executable packages exceed
90% statement coverage. New pinned LLGo executes codegen/nativeexec suites,
including the focused regression; affected vet checks pass. Full final-head
Go/LLGo x Linux amd64/macOS arm64 CI remains required before merge.

[Identities, all comparisons and adverse points](data/performance-direct-native-resumes-2026-11-15.json),
[all 1,084 projected process records](evidence/performance-direct-native-resumes-2026-11-15.jsonl.gz)
and [protocols, source overlays, fixtures and validation logs](data/performance-direct-native-resumes-2026-11-15-protocol.tar.gz)
retain the independent events. Only unused request-latency vectors become
counts/extrema/exact hashes; full originals remain local. Full-portfolio,
competitive and production-release acceptance remain open.

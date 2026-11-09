# Private loop word forwarding

Baseline: `566797bdb46ad153e2fc87f0faa7ca0c8d5a2068` (PR #136).
Controlled performance and production-release acceptance remain open.

## Problem and mechanism

The previous Linux100000-iteration Threads cohort observes private/shared
native API intervals at1.401/1.402 times Wasmtime. Invocation allocation savings
do not eliminate these sustained memory costs. This contribution targets one
avoidable dependency: reloading the private word written by the preceding
iteration. It does not recognize an increment or any named workload.

The existing bounded native-loop planner accepts at most256 nontrapping IR
instructions with proven memory accesses and unchanged backedge polling. An
additional linear scan recognizes exactly one full-width integer load followed
by one matching store at the same proven constant memory32 address. Only the
private default memory qualifies. Shared memories, memory views, remapped spans,
helpers, trapping bodies, interior/header polls and deferred frame-copy plans
retain their existing paths. Other memory effects reject this proof.

A register absent from every active loop mapping, immutable constant, runtime
metadata and heap cache holds the word's bits. It comes from the existing target
inventory and already-proved auxiliary scratch inventory; no bank or ABI grows.
Register pressure declines the optimization. An initial load from the checked
minimum-memory span fills this register before the first iteration. Moving that
read across its nontrapping, memory-effect-free prefix cannot expose a new trap
or mutation. Each original store still executes; its low32/all64 stored bits
refresh the cached word. Each later load reads those bits into its original
canonical slot, regardless of intervening reuse of that slot.

The original per-backedge cancellation check and full-frame publication remain.
No state survives a native entry, helper, callback or cold continuation. A
borrowed amd64 length register is restored on the ordinary exit. Original cold
PCs retain ordinary checked loads/stores and precise-tier fuel is unchanged.
Large/unencodable initial displacements use the existing path.

## Qualification and predeclared comparison

Both target assemblers compile32/64-bit, aligned/unaligned and displacement-edge
cases. Executable tests compare every original continuation with the ordinary
checked tier, including complete frame words, untouched sentinels, every memory
byte, wraparound and a forced real backedge poll. Rejection cases cover shared
and alternate memories, missing proofs, changed addresses/widths/order, extra
reads, unknown indices and current/future register ownership. Full affected Go
and pinned LLGo tests, race/vet, final four-profile semantic CI, all coverage
gates and extracted CLI package verification remain required.

Before timing, finish every local build/test/probe and freeze the final source.
Keep Go1.27.3, LLGo38de1e218 `-O3 -lto=full`, LLVM22.1.8 and named references.
Retain the existing five fresh interleaved blocks, two warmups/seven measured
calls, alternate checksums and every A/A interval. Linux measures original22,
separate sustainedGC4 and the unchanged six-case100000-iteration Threads v2
cohort. Local arm64 measures Core4, originalThreads6 and sustainedThreads6.
Verify all source/fixture/compiler/reference identities. Retain all phases,
adverse/unresolved controls, unsupported probes and allocation process means
including pauses. Do not sweep source/profile/ISA/counts, retime unchanged
sources, subtract phases or pool hosts/cohorts. This single experiment does not
close controlled performance or authorize release publication.

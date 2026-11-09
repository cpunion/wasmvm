# Deferred loop induction

Baseline: `6923e034b42f89e93315121eed5941315a9fd659` (PR #137).
Controlled performance and release acceptance remain open.

## Mechanism

The existing bounded native loop can defer a counter used only by its +1 update
and unsigned i32 comparison against a stable dynamic bound. The counter's
existing, exclusive register holds bound-counter. A guard requires counter<bound
before replacing that register; failed guards execute the ordinary checked path,
including do-while, zero-bound and wraparound behavior.

The loop body retains every original memory operation. Its backedge still checks
the interrupt word before each branch. Only after that check does one decrement
set the branch flags. This combines induction update and comparison without
reducing cancellation frequency. A cold poll restores bound-countdown+1 before
the existing predicate materialization and complete-frame spill. An ordinary
exit already decremented and restores bound-countdown instead. Both paths consume
the temporary state before leaving the native entry; cold continuations retain
the original instructions and precise-tier fuel is unchanged.

No native ABI or register inventory grows. Unknown counter uses, nonmatching
updates, aliases, moving homes, deferred copy recipes, affine memory guards,
header polls and borrowed amd64 length homes decline this proof. Private-word
forwarding remains separate; shared loads/stores are never forwarded by it.
Compiler work remains bounded by the existing <=256-instruction loop region
and fixed register inventories.

## Qualification and declared comparison

Tests compare every original continuation and forced real backedge polls with
the ordinary checked tier, including full frames, untouched sentinels, all memory
bytes, dynamic bounds0/1/max, below/equal/above-bound counters and wraparound.
Both target encodings enforce code quotas. Additional public embedding tests
execute private/shared modules under both backends, cancel long-running loops
and call again after cancellation. Go affected suites/race/vet and pinned LLGo
release tests precede final source timing. Full four-profile semantic CI,
coverage, stable consumer and extracted CLI package checks remain mandatory.

This PR uses the existing scoped Threads comparison: six original10000-iteration
cases and the distinct, unchanged six-case100000-iteration sustained-v2 cohort,
on hosted Linux amd64 and local macOS arm64. A fixed-profile compile-only audit
checks all22 portfolio modules on both targets before local timing; changes
outside the declared Threads scope require investigation before measuring.
The previous complete portfolio event remains recorded separately; no new timing
or performance conclusion is attributed to unchanged images.

Freeze final source, finish all local builds/tests/probes, and keep Go1.27.3,
LLGo38de1e218 `-O3 -lto=full`, LLVM22.1.8, references and producer hashes fixed.
Retain five fresh randomized/interleaved blocks,2warmups/7calls, independent
alternative checksums and every A/A interval. Preparation, first/warm API and
cleanup remain separate. Preserve adverse/unresolved controls and unsupported
outcomes. Do not sweep source/profile/ISA/counts, retime unchanged sources,
subtract phases or pool hosts/cohorts. Scoped diagnostics do not close controlled
M1–M6 performance acceptance or authorize versioned publication.

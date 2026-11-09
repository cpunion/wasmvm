# ARM64 heap caches in the shared caller budget

Baseline: `6923e034b42f89e93315121eed5941315a9fd659` (PR #137).
Controlled performance and production-release acceptance remain open.

## Diagnosis and rejected revision

The first PR #139 revision, `b12c9cf9d91c9162293b1a744f6d334d8504683c`,
preferred an invariant array identity over changing struct identities in the
last free cache pair. Its complete local original/sustained GC comparisons
observed indirect-access LLGo native head/base ratios of **1.3556/1.3489**;
not all A/A controls resolved. All raw observations, failures and controls
remain archived. It is withdrawn, not a new baseline or an accepted optimization.
Its pending remote validations were stopped when that implementation was
withdrawn; no hosted performance or full semantic pass is claimed for it.

A deterministic identity trace of the existing sustained fixture explains the
opportunity-cost problem: the original struct cache still hits 1499997 of
2000000 accesses (about75%). Changing identities do not imply zero useful reuse.
The compile-only ARM64 loop window grows400 to488bytes when that cache is lost.
Neither a predicted hit rate nor static instruction counts establish performance.

## Revised mechanism

Select checked cache pairs in the existing shared register planner, alongside
scalar homes, constants and metadata. ARM64 considers repeated-access classes
at a conservative potential saving of four instructions per access, charged
as two instructions to each of the pair's two words. This is a generic static
profitability estimate, not a runtime hit-rate claim. It can retain both array
and struct caches instead of reserving caller words for cheaper literals first.
There is no fixed-root bonus or workload-name selection.

Keep the same eight ARM64 caller registers and maximum two pairs. Each pair
must fit completely; no partial reservation is made. The existing slot scan,
leaf scratch and immutable metadata exclude selected pairs. Full-handle hit
comparisons, null/index/generation/array checks on misses, bounded polling,
collection/reentry and cold continuations retain their production contracts.
Reset every key on native entry in deterministic register order. The amd64
scoring, one-pair limit and physical inventory remain unchanged.

## Qualification and frozen measurement plan

Exercise competing literals, complete-pair budgets, disjoint mutable/immutable
mappings and deterministic key resets. Retain all native heap tests for packed
and vector values, every cold PC, changed/stale/null references, allocation,
collection, bounds and traps. Affected full Go/race/vet and pinned LLGo release
execution precede timing; final required four-profile semantic CI, coverage,
world-consumer and extracted CLI verification remain necessary for merge.

Before timing, compile all22 portfolio images for both fixed target profiles.
Investigate any non-GC image change and broaden the affected scope before
measurement if necessary. Use GC original4 and separate sustained4 (300000
allocation,1000000 other iterations), all named references and explicit
unsupported probes, independently on Linux amd64 and macOS arm64.

Freeze the revised source and finish every local build/test/probe before its
single timing event. Keep Go1.27.3, LLGo38de1e218 `-O3 -lto=full`, LLVM22.1.8,
fixture/reference hashes, five interleaved blocks,2warmups/7calls, timers,
checksums and statistics unchanged. Fresh base/head Go/LLGo binaries are
required. Do not reuse own-engine binaries from the rejected revision, retime
unchanged sources, sweep counts/profiles/ISA or pool cohorts/hosts. Preserve
all phases, raw samples/orders, allocation means including every pause and all
adverse/unresolved A/A intervals. These advisory runs do not close M1–M6
controlled acceptance or authorize a production-performance claim.

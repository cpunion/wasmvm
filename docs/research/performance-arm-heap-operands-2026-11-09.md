# ARM64 checked heap operands in allocated homes

Follow-up to the shared caller-budget implementation in merged PR #139,
baseline `281f2cb1a5e63356c237e21ff2125157473e937f`. Its final source passed
all required CI and the Linux comparison; controlled performance acceptance
remains open. This follow-up requires its own final-head qualification and
scoped comparison before integration.

## Mechanism

A cached struct hit formerly copied an allocated complete reference to x9
before comparing it, and scalar fields moved through x9 between an existing
allocated home and checked payload memory. ARM64 can name these homes directly.

Compare a mapped struct reference directly for the same null/full-handle hit
condition; only a miss copies it into the ordinary checked resolver operand.
Encode scalar field/array loads and stores with the mapped integer/reference
register where available. Read deferred source aliases through frameRead and
write the actual destination home. Packed signed loads keep the same W-register
sign extension/zero upper word. Float/vector or unallocated homes, precise fuel
and amd64 retain their existing constrained-operand paths. Allocation and
private-loop loads explicitly keep their original scratch operand.

No cache bank, validity check, array bound, poll, continuation, frame ownership
or ABI changes. Every cold-PC body still starts with reset keys and full miss
validation. Existing every-width/every-continuation tests cover both layouts,
packed signs, vector/floating bit patterns, raw upper bits, changed/stale/null
handles, relocation after reentry, pressure and forced poll reconstruction.
Full affected Go/race/vet and pinned LLGo release execution remain mandatory.
The first local build exposed two remaining helper callers after the signature
change; both were corrected before qualification or any measurement. The failed
build log remains diagnostic evidence, never an execution result.

## Measurement boundary

Preserve one active PR. Freeze this source after rebasing on the merged
dependency, then finish all local builds/tests/probes before one predeclared event.
First audit all22 fixed-profile module/function/tier images for both targets.
Scope measurement to all changed families before timing; retain original and
sustained GC cohorts independently. Reuse pinned profiles, fixtures, references,
counts, process/call ordering, means including allocation pauses and every
adverse/unresolved A/A interval. No unchanged-source retiming, source/ISA/profile
or count sweep, phase subtraction, host/cohort pooling or controlled-acceptance
claim is permitted. Performance and production release remain open.

## Immutable global values in the existing planned words

The prepared follow-up now also retains selected immutable scalar/reference
values, rather than their cell pointers, in the same ARM64 metadata registers.
At every native entry the driver has bound valid global cells; load each
selected value once before cold-PC dispatch. Native global.get then copies its
full value directly to an integer/float home or the ordinary spill path. Mutable
globals, V128, unallocated metadata, precise fuel and amd64 retain live cell
loads. No register, native ABI field or runtime feature flag is added.

Import matching requires actual and expected global mutability to agree
(internal/core/module/extern.go); the public API denies immutable GlobalSet.
Cells cannot change during one serialized native entry. Reentry reloads the
binding and value; complete GC handles remain canonical, global roots remain
traced and address-cache validation/reset is unchanged. Existing metadata and
slot selection still exclude each other's physical registers.

Added tests exercise eligibility, every cold entry, actual cell refresh between
entries, aliased global bindings, signed/NaN/reference/vector bit patterns,
integer/float/spilled homes, pressure and deterministic target images. Retain
mutable-global, host-reentry, forced-GC and invalid-import tests. The complete
Core3 corpus and pinned Go/LLGo execution remain required after this addition.

A compile-only diagnostic of the same Wasmtime48.0.1 SDK and four GC binaries
finds the immutable global root resident outside the indirect-access loop;
that image uses32-bit heap-relative reference loads and4-byte array reference
strides. It also contains null, bounds/overflow and GC barrier paths. The
recorded Wasmtime policy lacks the owned backedge control-word poll. These
static observations identify architecture differences, not measured cycle
contributions, a new benchmark, or permission to remove checks. The original
identity trace has75% struct-cache hits and an alternating store-hit pattern;
do not assume unpredictable branches or add a two-way cache on that premise.
Preserve raw SDK/image hashes, disassembly and source traces separately.

## CI package balance

The final PR #139 Linux core shard took608 seconds, with477 seconds in its
compile/link/test step. Individual overlapping LTO links included phase.test
at169.936 seconds; its actual tests took0.169 seconds. Move only
internal/bench/phase from core to component, whose observed total was319
seconds. Both shards then contain nine packages. Phase uses committed fixtures
and does not require an extra corpus input. Keep all31 packages exactly once
per platform, eight shards, the pinned release profile, every semantic report
and every coverage gate. This is a scheduling change; a shorter final CI job
must be observed rather than assumed from overlapping link durations.

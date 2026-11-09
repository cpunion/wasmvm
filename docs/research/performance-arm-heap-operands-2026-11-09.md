# ARM64 checked heap operands in allocated homes

Prepared follow-up to shared caller-budget revision
`1178f72be66bc45009f696a3b07762a2cfbb1921` (PR #139).
That PR remains subject to final required CI and hosted comparison; this branch
is not a second active contribution PR or an accepted performance result.

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

Do not time this follow-up while PR #139 or local preparation is active.
Preserve one active PR. Once its dependency is resolved, rebase and freeze this
source, then finish all builds/tests/probes before one predeclared event.
First audit all22 fixed-profile module/function/tier images for both targets.
Scope measurement to all changed families before timing; retain original and
sustained GC cohorts independently. Reuse pinned profiles, fixtures, references,
counts, process/call ordering, means including allocation pauses and every
adverse/unresolved A/A interval. No unchanged-source retiming, source/ISA/profile
or count sweep, phase subtraction, host/cohort pooling or controlled-acceptance
claim is permitted. Performance and production release remain open.

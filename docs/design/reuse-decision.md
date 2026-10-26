# Historical Decision: Wazero-First Evaluation

**Superseded by [Project-Owned Core](own-core.md) on 2026-10-26.** The text below preserves the M0 evaluation decision and its original milestone numbering. Its reuse-first instructions and next-work recommendations are historical; the current [roadmap](../roadmap.md) governs implementation.

Date: **2026-10-26**. Status: **M0 evaluation decision, subject to the contribution PR's required checks and review before merge**. Evidence and upstream pins are recorded in the [wazero evaluation](../research/wazero-evaluation.md). This selects the next implementation route; it does not accept M1–M5 or claim complete runtime support.

## Decision

Prefer reuse over a new VM core. Use the pinned wazero interpreter as the initial implementation foundation, preserve its existing Core/P1 implementation, and address demonstrated missing capabilities as separate work items. Fix missing Go runtime contracts in LLGo instead of maintaining a wazero workaround for those compiler defects. Do not start an independent decoder/validator/interpreter merely because the current release does not satisfy every long-term goal.

An existing implementation that passes the complete required feature, compiler, lifecycle, and platform matrix would eliminate the need for another core. `wasmvm` would then become an integration and verification project, with only the adapters or policies that its consumers need. Avoid inventing a parallel public API when the existing API already meets those needs.

Current evidence does **not** support “use upstream wazero unchanged and all stages are complete.” It also does **not** support “an entirely new VM is necessary.” The immediate decision is a reuse-first implementation route with explicit evidence gates.

## Why this direction

| Evidence | Consequence |
| --- | --- |
| Wazero provides a substantial Core 1.0/2.0 interpreter and P1 host implementation | Rewriting those foundations would duplicate proven work |
| Stable v1.12.0 includes experimental threads, final EH, typed refs, tail calls, and extended constants | Several planned M1/M2 capabilities can potentially be reused after validation |
| Released LLGo fails to link; an LLGo runtime fix permits the unchanged wazero evaluator to execute | Contribute the compatibility fix to LLGo and pin its tested revision; no wazero backend patch is required for this path |
| WasmGC, Memory64, multi-memory, relaxed SIMD, and native component/P2/P3 hosting remain outside the inspected upstream support | Later milestones still need implementation or a separately validated reusable source |
| A blocked atomic wait delays host cancellation until its own timeout | Lifecycle correctness is an adoption criterion, not an optional optimization |
| Wago's Go smoke tests pass, but its LLGo run fails during ARM64 assembly translation; Wazy remains a source-only follow-up | Choose the interpreter path with a demonstrated compatibility fix while preserving alternatives for later feature work |

## Architecture consequences

The independent pure Go interpreter remains a fallback design rather than the next code-writing task. Existing public package declarations are placeholders, not a requirement to wrap every wazero type. M0 determines whether `pkg/vm` adds useful policy/lifecycle functionality or should stay minimal.

Keep evaluation dependencies and compatibility experiments separate from claims about a production core. The diagnostic backend-isolation patch is retained only to reproduce the investigation with LLGo v1.0.8. Required CI executes unchanged wazero using the LLGo memory-hook contribution, not a silently modified engine. Do not use a runtime configuration choice as a substitute for checking the actual compiler package graph and linked runtime symbols. The compiler contribution is unreleased and its upstream review status remains a documented dependency.

Adopting wazero changes implementation provenance, not the acceptance requirements. The shared Go/LLGo tests, supported-feature manifests, explicit WASI versions, resource/lifecycle behavior, and guest toolchain verification still apply. Any design choice inherited from an engine that differs from the current architecture must be documented and tested before that milestone is accepted.

## Alternatives

| Route | Assessment |
| --- | --- |
| Use upstream wazero unchanged | Insufficient for the complete target; the unmodified released LLGo v1.0.8 also fails to link it |
| Reuse wazero, fix compiler contracts in LLGo, and contribute required engine/host features | Selected initial route; retains existing strengths and exposes real remaining costs |
| Import the open GC PR as a completed capability | Rejected as an acceptance shortcut; it is unmerged, interpreter-only, and needs GC/lifetime review |
| Reuse Wago or a wazero-derived implementation | Alternative for later feature decisions; Wago's compiler boundary is measured, Wazy's claims remain unvalidated |
| Wrap Wasmtime/WAMR as the production engine | Does not meet the intended Go implementation constraint; useful as a reference engine |
| Build an independent VM now | Deferred until evidence shows reuse would require an unmaintainable redesign or cannot meet the essential contracts |

## Required implementation work remains staged

| Milestone | Effect of reuse |
| --- | --- |
| M0 | Complete compiler/feature experiments, patch scope, candidate comparison, guest inventory, and license review before closing the gate |
| M1 | Reuse decode/validate/execute/P1 where available; implement only necessary embedding/CLI policy and prove limits, errors, and cancellation |
| M2a | Validate shared-memory semantics; implement or reuse legacy wasi-threads and complete process/blocked-wait lifecycle behavior |
| M2b | Validate existing typed-ref/EH/tail-call paths; choose and fully validate a WasmGC implementation rather than treating the open PR as done |
| M3 | Close Core 3.0 gaps and run the pinned LLGo guest matrix; isolate individual feature results from combined-workload acceptance |
| M4/M5 | Add or reuse versioned component/WASI layers; component parsing or an async ABI alone is insufficient |
| M6 | Optimize or adopt experimental proposals only after the required correctness gates |

Each stage must pass its entire declared acceptance scope. Reuse may reduce the amount of new code, but it does not reduce testing obligations or convert expected unsupported probes into completed features.

## Contribution and coverage requirements

Changes to `cpunion/wasmvm` go through a contribution branch and PR, with required checks passing before merge. For upstream projects outside the permitted owner boundary, use an allowed fork and contribution PR; never push directly to upstream or merge its PR. Preserve source licenses and notices, and record the origin and regeneration commands for test fixtures.

Coverage reports describe the code they measure. Measure owned evaluator/integration code, test error and unsupported-feature paths, and report compiler execution separately. VM conformance requires the appropriate pinned suites and actual feature results; a high statement-coverage percentage for the evaluator does not establish engine completeness.

## M0 deliverable record and next work

| M0 work package | Delivered evidence |
| --- | --- |
| Candidate pins and licenses | Wazero stable/main/GC-PR snapshots, separate pinned module files and checksums, Wago beta.13 comparison, Apache-2.0 provenance |
| Specification, suite, and interface inputs | [Immutable Core/proposal/P1/WASI-test revisions and license notices](../research/spec-inputs.md), plus downloaded and SHA-256-verified WASI 0.2.12/0.3.2 WIT release assets; full upstream suite execution remains assigned to later milestones |
| Dual-compiler experiment | Shared evaluator, original v1.0.8 failure, scoped symbol inventory, LLGo issue #2918 and contribution PR #2919, unchanged-engine execution with the fix |
| Feature corpus | 13 owned WAT/Wasm pairs, source/binary hashes, exact rejection assertions, negative cases, concurrent shared memory, cancellation observation, owned-code coverage |
| LLGo guest inventory | Committed source/module and hashes, producer workaround, actual imports/features, successful Wasmer reference run |
| Reuse decision and upkeep | This decision, explicit feature/lifecycle gaps, contribution policy, and no imported engine implementation source |
| Additional requested performance work | Same-input comparison across four runtime projects, startup/small/sustained workloads, raw samples and contention caveats |

All outcomes are bounded experiments, including recorded candidate failures. M0 does not require porting every candidate or passing a complete engine conformance suite; those obligations belong to the selected implementation's later milestones. The PR merge gate requires the final shared evaluator to pass under Go and the pinned fixed LLGo compiler, plus the coverage and integrity checks.

The next milestone is M1. Start with the existing embedding API and smallest useful policy/CLI integration, then prove every M1 limit, error, cancellation, and P1 requirement. In particular, do not silently replace deterministic fuel with a wall-clock deadline or call the entire P1 API accepted based on one output fixture. M2–M5 retain their full acceptance scope.

Revisit the route if GC/reference ownership cannot be made reliable, required resource/cancellation semantics cannot be enforced, or measured maintenance costs favor another existing implementation. A new core becomes justified only after a concrete comparison demonstrates that need.

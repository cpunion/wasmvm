# Native batch commit and rollback: 2026-11-08

Source `9242933` compares with main `431da80`. Go 1.27.3 and LLGo `3571c25`
retain the pinned `-O3 -lto=full` release profile; references and CPU policies are
unchanged. **No reliable whole-workload speedup is established.** This increment
removes mutable closure captures while preserving transaction cleanup.

## Protocol

Batch publication previously deferred a closure over the Store and a mutable
committed count. It now registers `defer s.releaseNativeReservations(0)`. After
successful validation/publication, the existing release method explicitly
returns unused reservations and clears `pendingCount`. The deferred rollback
then observes an idle batch and does nothing. Errors and panics still invoke
the fixed rollback while the batch is pending.

No Store field, native word, execution framework or quota is added. Full batch
validation still precedes object publication. Handles, epochs, payloads, roots,
unused reservations, cancellation/trap and precise-mode contracts stay intact.
The active-entry unwind guard remains in the native driver.

The new fault test truncates a private metadata page after staging, catches the
publication panic, restores the page and verifies complete reservation cleanup,
preexisting object identity/accounting, next-batch reuse and traced reference
chains. Existing invalid-prefix, quota, empty-object, generation/epoch, forced
collection and pre-entry cancellation/panic cases remain.

## Linked code and allocation evidence

Exact base/head Go and LLGo runners were built independently. The final linked
LLGo publication method contains **nine → six static direct allocator call
sites**, including its cold error blocks. Inspection removes the three early
receiver/counter/capture allocations; the two defer-management allocations
remain covered by existing LLGo #2932. Go has no direct `runtime.newobject`
call in either version. Static sites are not runtime allocation-event counts.

An independent untimed counter probe repeats complete prepare/stage/publish/
collect cycles with the same fixture helper. Its source, fixed counts and
approximate LLGo TotalAlloc deltas are retained separately. Fixture-helper
allocation is included, so its absolute totals cannot describe production
allocation cost or the entire Go/LLGo gap. It does not use Mallocs as a counter.

All 22 x86 modules remain byte-identical. All 22 ARM64 images and entry lists
match base/head and Go/LLGo; 44 fresh candidate checksum publications pass.
Baseline publications are reused only after complete engine/harness, compiler,
profile, executable and fixture identity checks. Host publication changes even
though guest code does not, so this is not a reason to skip native warm tests.

## Measurements and limitations

The [Linux event](https://github.com/cpunion/wasmvm/actions/runs/42427030933)
retains all original22 workloads and the separate fixed sustained GC cohort.
Allocation throughput uses the median of process means, including every
observed warm call/pause within each process; original medians and both A/A
reductions remain explicit. First/preparation/cleanup/process intervals are
separate, with no subtraction or cross-host/source pooling.

| Linux allocation / process means | Go head/base | LLGo head/base | LLGo / Wasmtime | All A/A |
| --- | ---: | ---: | ---: | --- |
| Original, 10,000 | 0.985 | 1.007 | 0.538 | Inconclusive |
| Sustained, 300,000 | 1.006 | 0.996 | 0.798 | Inconclusive |

A separate local ARM64 GC-only campaign uses the same five-block, seven-call
protocol and independent exact-source executables. Its original-count
allocation ratios are Go **0.988**, LLGo **0.971**, LLGo/Wasmtime **0.376**;
controls are inconclusive. The local sustained cohort overlaps an added Go
coverage run for **13.79 seconds**. It is preserved as a contaminated diagnostic,
excluded from performance conclusions, and not favorably rerun. Original local
and independently hosted Linux measurements have separate records.

Complete affected State/nativeexec/public-VM tests pass under Go and release
LLGo; vet passes. The initial three-package local coverage subset covers State
87.03%. One additional Go execute-corpus run brings combined State coverage to
**95.21% (1490/1565)** and nativeexec to **92.54% (459/496)**. The initial coverage
expectation failure is retained. Final four-profile CI, all 31 packages per
LLGo target, all package coverage gates and resolved review remain mandatory.

The final PR body records completed CI and Linux observations. Controlled
full-portfolio, steady-state GC and broader world/release acceptance stay open.
The [machine record](data/native-batch-commit-2026-11-08.json) retains local
observations, overlap timestamps, source/compiler identities and linked-code
limits. Local archive: `artifacts/native-batch-commit-20261108/local-evidence.tar.gz`,
3498913 bytes, 480 files; SHA-256
`d4f5a88476c105f55869f12e6a78b3814e1e39c0c32978ed2706ff2685fce398`.

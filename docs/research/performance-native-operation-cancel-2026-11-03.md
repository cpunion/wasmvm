# Operation-owned native interruption: 2026-11-03

Selected production `4bb81a964981058dd36026c7a927845ee05d2888`, baseline
`1e166519d42c42788d20e900036e6d188a640abd` (PR #108). Exact contribution-head
Linux/macOS CI and controlled performance acceptance remain required. Local
diagnostics below retain the benefits and adverse cases separately.

## Ownership and bounded teardown

Public operations deliver native interruption through their existing frame
registry. Workspace borrowing/rebinding and registration share the Store lock;
detachment, native suspension and the cache decision share its return transaction.
A background parent needs no cancellation watcher. A cancelable parent has one
operation-level watcher covering native calls, nested reentry and continuations.
Store.Close and panic cancellation deliver interruption directly under the lock.
No callback of a retired operation can reach a later operation's cached workspace.
Interrupted native storage is discarded; its control word is never reset.
Standalone native executions retain their automatic watcher, with explicit
NewOwned/BeginOwned entry points for callers that assume delivery and draining.
Guest code, C ABI, polling, quotas, roots and host-buffer checks remain intact.

Caller contexts remain distinct standard cancel contexts with deadlines, values
and causes. The immutable LLGo cancellation view is retained. This removes
per-workspace AfterFunc registration from public calls without pooling contexts.
Native interruption adds no second frame registry or execution framework.

The investigation also reproduces an existing bounded-exit bug on exact main:
WithCaller waits for a concurrently used capability before canceling it when
its callback returns or panics. The injected native/panic test fails after two
seconds; explicit Store.Close rescues the blocked baseline so no test goroutine
or guest loop leaks. The selected source cancels before waiting and publishes
final fuel after capability operations drain. Normal return and panic pass with
both backends. A separate host-panic test drains a concurrent native reentry.
Tests also exercise close/parent cancellation of nested native frames, retained
context metadata/expiry, delayed old watchers, reused workspaces, GC and sticky
interrupt retirement. Panic cancellation remains locked; only final cleanup
after capabilities drain can skip an empty interruption dispatch.

All 31 Go packages, thirty coverage gates, scoped race/vet and core no-cgo pass.
Pinned LLGo executes the affected VM/component/P1/P2/P3 packages. Earlier full
native/platform/execute checks are also retained; final dual-platform CI is still
required. The standalone regression test is reformatted with gofmt between the
baseline injection and committed test; original/canonical hashes and the exact
format transformation are retained.

## Three separate measurements

Apple M4 Max/macOS arm64, Go 1.27.3, LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, `-O3 -lto=full`.
All builds/tests/allocation probes finish before each timing campaign. Every
campaign runs once: five fresh serially interleaved process blocks, two warmups,
seven samples, original fixture hashes and independent original/alternate oracles.
Base binaries and named reference binaries are reused only with verified exact
revision/profile/hash identities. No cross-campaign pooling or timing subtraction.

The first source `7600a74` improves background LLGo native public calls 31.4%
and sessions 40.2%, but regresses live-parent public calls 4.4% (A/A -0.2%) and
background interpreter calls 3.8%. Source `1cd9dc7` folds registration/removal
into existing workspace transactions: background improves 32.5%, sessions 42.8%,
background interpreter changes -0.2%, but live native is still +2.5% (A/A +0.8%).
Neither earlier campaign is replaced. A third pre-session source has passing
quality checks but no timing campaign; the final selected source includes the
independently reproduced session fix and avoids empty final interruption work.

Each complete campaign has **485 observations, 97 summaries/alternate checks**:
both compilers/backends/base/head, public/session/dispatch, raw entry components,
extra live-parent public API modes, native A/A aliases with both parent policies,
named default/enabled-background/live wazero, Wasmtime, Wasmer and ordinary C.
Empty call-control uses one guest iteration; integer-loop uses 64. Timers include
10000 calls and per-call result checks. Raw firstCallNs is the first batch, not
an unbatched first invocation. Preparation/readiness/first/warm/cleanup stay
separate. Raw native entry is an uncancelled Context.RunState component with
grouped preparation and unmeasured cleanup (raw zero retained).

Final warm ns/call:

| Boundary | Go base → head | Change | LLGo base → head | Change |
| --- | ---: | ---: | ---: | ---: |
| Empty native, background | 588.742 → 395.575 | -32.8% | 1682.479 → 1164.438 | -30.8% |
| Empty native, live parent | 636.112 → 638.896 | +0.4% | 2010.421 → 2039.513 | +1.4% |
| Empty native session | 377.408 → 244.613 | -35.2% | 1044.054 → 591.833 | -43.3% |
| Empty native dispatch | 144.692 → 136.800 | -5.5% | 259.458 → 271.054 | +4.5% |
| Empty interpreter, background | 326.892 → 321.008 | -1.8% | 953.263 → 935.304 | -1.9% |
| Empty interpreter, live parent | 363.337 → 365.454 | +0.6% | 1155.746 → 1246.754 | +7.9% |
| Integer native, background | 585.167 → 408.333 | -30.2% | 1679.013 → 1156.612 | -31.1% |
| Integer native, live parent | 618.038 → 624.938 | +1.1% | 1869.683 → 1898.504 | +1.5% |
| Integer native session | 405.154 → 297.017 | -26.7% | 1046.283 → 642.725 | -38.6% |
| Integer native dispatch | 206.637 → 201.679 | -2.4% | 336.200 → 334.150 | -0.6% |
| Integer interpreter, background | 3122.796 → 3102.683 | -0.6% | 3635.042 → 3631.375 | -0.1% |
| Integer interpreter, live parent | 3169.912 → 3156.471 | -0.4% | 3881.492 → 3889.037 | +0.2% |

Empty native A/A moves Go +1.2%/+2.3% (background/live), LLGo +3.5%/-2.5%;
integer A/A Go +0.2%/-0.7%, LLGo -1.6%/-0.1%. Shared-host load is
16.59/17.65/14.44 before and 14.57/17.59/14.86 after. These controls do not
calibrate controlled acceptance. The live-interpreter empty regression and
short dispatch movements remain adverse observations with unresolved causes.
Do not waive them because their guest compiler body is unchanged.

A separate unchanged args-sizes public-host control has ninety VM/five C
observations and eighteen VM/one C alternate checks. It uses the ordinary
HostFunction with two checked writes and errno, omitting P1 Process admission;
it is not a complete P1 implementation. Final ns/callback:

| Mode | Base → head | Change | A/A change |
| --- | ---: | ---: | ---: |
| Go interpreter | 219.533 → 221.400 | +0.9% | -1.5% |
| Go native | 236.492 → 230.971 | -2.3% | +1.4% |
| LLGo interpreter | 410.596 → 394.913 | -3.8% | -3.1% |
| LLGo native | 627.350 → 624.638 | -0.4% | -1.0% |

Cancellation registration is per outer call, not per callback in this loop;
these controls do not establish a native host-callback throughput win.

## Allocations and remaining gaps

Separate allocation diagnostics retain all three rounds/zero-call controls.
Third-round native public reported TotalAlloc per 10000 calls is Go
8080112 → 4080000 (about 400 bytes and five allocation events saved per call),
LLGo 31678256 → 20962448 (about 1072 fewer reported bytes). Sessions are Go
2080664 → 640408, LLGo 16794832 → 8168400. LLGo BDWGC bytes are approximate;
its zero Mallocs is unsupported instrumentation, not allocation-event evidence.
The whole operation protocol changes together, so no individual registration,
defer, lock or private return is assigned the entire saving. Known upstream
[defer](https://github.com/xgo-dev/llgo/issues/2932) and
[startup](https://github.com/xgo-dev/llgo/issues/2973) reports remain applicable;
no new compiler defect is independently established by this VM comparison.

The final full LLGo background native API is still 2.94x Go. The same empty
campaign reports wazero compiler default/background 14.904 ns, enabled/live
383.738, Wasmtime 66.546, Wasmer 153.363 and C volatile identity 0.867.
Lifecycle/cancellation settings differ; C has no VM policies and is a lower
bound. Component measurements cannot replace full API scores. Sustained
machine-code/native-C, startup, live-parent/cancellation/allocation, complete
portfolio and expanded WASI-world acceptance remain open.

Setup evidence retains an initial undefined test-helper build failure, missing
LLGo revision environment in the host builder and a dirty-document guard that
stopped a later build before timing. Each was corrected and logged; compiler
profiles and guest semantics were not changed to obtain a passing measurement.

[All three complete campaigns, phases, plans, oracles, compiler/binary/source
identities, quality checks, baseline deadlock and intermediate quality evidence](evidence/performance-native-operation-cancel-2026-11-03.json.gz).

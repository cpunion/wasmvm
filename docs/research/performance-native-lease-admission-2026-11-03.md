# Native lease admission and call-boundary gaps: 2026-11-03

Selected production source: `17f65d93c9e81b6d16935f7062399da9b80fa4d5`.
Base: `344bbc35f9818f89dcb2779c6819619fa06b1da7` (merged PR #107).
Exact contribution-head Linux/macOS CI and controlled portfolio acceptance
remain required. Local results below are shared-host diagnostics.

## Change and lifetime contract

A platform Lease now combines its closed bit and admitted-entry count in one
atomic word. Admission loads and CAS-increments; departure atomically decrements.
Close prevents later admission and releases the code pin after the last admitted
entry returns. This replaces the per-entry Lease/Code locks and Code's separate
active counter. Final pin release still takes the Code lock. There is no change
to generated guest code, C ABI, scheduler boundary, host-buffer checking,
Context/memory locking, cancellation or execution quotas.

The new tests execute two native cancellation loops through one lease, close
the owner and lease, reject a third entry, and confirm that the mapping survives
the first return and forced GC until the second return. Concurrent admission
versus close, invalid-index rollback, overflow rejection, zero/idempotent close
and subsequent successful native execution are also checked. Existing register,
FP, GC and cancellation probes remain. Local checks pass: all 31 Go packages,
30 coverage gates, affected race/vet, core no-cgo and six affected LLGo packages
(platform, nativeexec, VM, component, P1 and P3). An initial private-test signature
build failure was fixed by adapting the existing FP/register probes to Lease.end.

## Measurement

Apple M4 Max/macOS arm64; Go 1.27.3; LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, `-O3 -lto=full`.
No compiler/profile change. Builds, semantic tests and allocation diagnostics
finish before timing. Each study runs once with five fresh interleaved process
blocks, two warmups and seven samples, original fixture bytes/checksums and an
alternate seed. Exact base A/A aliases expose temporal variation. All controls,
process readiness and preparation/first/warm/cleanup phases remain in evidence;
no phase medians are subtracted or cross-campaign results pooled.

The short-call harness times batches of 10000 calls, including per-call result
checks. Its public API, retained session and internal dispatch are different
boundaries. Empty call-control uses one guest iteration; integer-loop uses 64.
There are 385 observations, 77 summaries and 77 alternate checks. The standalone
native-entry probe executes a raw ABI add through Context.RunState; it is an
uncancelled component, not a full VM. Its preparation is grouped under compile/
validate and cleanup is unmeasured (raw zero retained).

Warm ns/call; percentage is head/base minus one:

| Boundary | Go base → head | Change | LLGo base → head | Change |
| --- | ---: | ---: | ---: | ---: |
| Empty public interpreter | 261.12 → 258.70 | -0.9% | 840.60 → 841.78 | +0.1% |
| Empty public native | 491.02 → 479.02 | -2.4% | 1588.96 → 1502.83 | -5.4% |
| Empty native session | 319.38 → 312.63 | -2.1% | 1004.38 → 911.20 | -9.3% |
| Empty native dispatch | 125.78 → 122.65 | -2.5% | 330.23 → 225.53 | -31.7% |
| Raw native entry component | 91.38 → 81.23 | -11.1% | 189.13 → 104.26 | -44.9% |
| Integer public interpreter | 3103.55 → 3135.61 | +1.0% | 3605.70 → 3573.44 | -0.9% |
| Integer public native | 569.87 → 565.38 | -0.8% | 1695.92 → 1620.82 | -4.4% |
| Integer native dispatch | 214.82 → 199.38 | -7.2% | 421.77 → 329.77 | -21.8% |

Empty native API A/A moves Go -0.3%, LLGo -0.6%; integer A/A moves Go +2.7%,
LLGo +1.5%. The other interpreter/session/dispatch controls remain in evidence.
These observations support selecting the compact lease change for full CI;
they do not establish an all-workload or all-platform win.

A separate unchanged args-sizes fixture executes 10000 callbacks per call.
Our diagnostic provider uses the ordinary public HostFunction, two checked
memory writes and errno, omitting P1 Process admission. This is a public-host
regression control, not a complete P1 implementation or equal reference policy.
Eighteen VM configurations and ordinary C produce 90 VM/five C observations,
with eighteen VM/one C alternate checks:

| Public-host mode | Base → head ns/callback | Change | A/A change |
| --- | ---: | ---: | ---: |
| Go interpreter | 214.804 → 211.400 | -1.6% | +0.7% |
| Go native | 240.683 → 233.475 | -3.0% | -1.1% |
| LLGo interpreter | 390.167 → 393.121 | +0.8% | -0.2% |
| LLGo native | 674.429 → 605.862 | -10.2% | +2.0% |

## Remaining gaps and mechanism

The same short-call campaign reports these empty-call medians:

| Mode | ns/call |
| --- | ---: |
| wasmvm Go native public API | 479.02 |
| wasmvm LLGo native public API | 1502.83 |
| wazero compiler, default/background | 13.71 |
| wazero compiler, termination enabled/live parent | 358.32 |
| Wasmtime Cranelift | 56.83 |
| Wasmer Cranelift | 135.59 |
| Ordinary C volatile function-pointer identity | 0.91 |

LLGo's full native API is still about 3.1 times the Go result. Competitor
lifecycle/cancellation policies differ; ordinary C has no VM lifecycle,
validation, guest memory checks or cancellation. Ratios to C are a lower-bound
comparison, not an equal-workload target. Internal/component times exclude
outer contracts and cannot replace public API scores. Public-host default
reference ns/callback are wazero compiler 24.29, Wasmtime 54.80, Wasmer 42.60,
and ordinary C 0.90; enabled wazero compiler is 39.44. The large host gap remains.

The pinned wazero v1.12.0 source explains part of the short-call difference:
default call paths skip ensureTermination; enabled paths call
CloseModuleOnCanceledOrTimeout and start a cancellation watcher goroutine per
call, including background parents. Enabled compilation also instruments
termination checks. See the pinned [compiler call engine](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/wazevo/call_engine.go),
[interpreter](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/interpreter/interpreter.go)
and [module lifecycle](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/wasm/module_instance.go).
Our Store cancellation and retained Caller.Context contracts remain enabled;
this change removes no such policy. Faster default reference modes still
remain visible and motivate a cheaper implementation of our full contracts.

Separate allocation samples use three rounds of 10000 calls, with GC before
each round. Third-round LLGo native public reported TotalAlloc drops
33914096 → 31682176 bytes; session drops 19042880 → 16800480, about 224 fewer
reported bytes per call. Go remains approximately 8080000 public/2080664 session
bytes. LLGo's BDWGC TotalAlloc is approximate; its reported zero Mallocs is
unsupported instrumentation, not proof of zero allocation events. Removing
locks/deferred scopes and changing the private begin return together do not
isolate each constituent saving. This is consistent with existing
[LLGo #2932](https://github.com/xgo-dev/llgo/issues/2932), not a new independently
established compiler defect. The independent startup cause and upstream update
are in the [startup marker report](performance-llgo-startup-markers-2026-11-03.md)
and [LLGo #2973 update](https://github.com/xgo-dev/llgo/issues/2973#issuecomment-6593936533).

Next acceptance work must retain public contracts while reducing their
allocation/synchronization cost, measure complete sustained kernels separately
from call/startup overhead, and check Linux as well as macOS. Portfolio,
zero-allocation, startup and controlled cancellation gates remain open.

## Rejected experiment and initial diagnostic

The earlier boundary-only study on unchanged main has 255 observations,
51 summaries/checks. It exposes public/session/dispatch/entry and default/
enabled-background/live reference modes. Its host conditions differ markedly
from later campaigns; its absolute times are not pooled or subtracted.

The independent held source `4f14bcd334657f7ba66b8af412ceb93a598a5a19`
adds numeric marshalling and a bounded 64-value argument cache with nested-call,
reference/GC, mixed scalar/vector, result-ownership and error cleanup tests.
It is excluded from the selected lease change. Despite about 32 Go bytes and
one Go allocation saved per call (LLGo about 48 reported bytes), full LLGo
empty native API regresses 0.4%, integer improves 2.9% while A/A improves 3.7%,
and public-host interpreter regresses 3.0% (A/A +0.8%). There is no stable
material full-API win to justify the extra cache state. Its complete 385-call
and 90-VM/five-C-host observations and adverse controls remain retained.

Held-candidate quality also retains an initial full-upstream-world race timeout
in P3 tcp_streams. Isolating that exact case with the unchanged 30-second limit
passes baseline 14.290 s/head 14.400 s; this does not erase the original timeout.
The standard CI race scope passes. An initial expanded no-cgo run selected
unsupported native component/P1 variants; the exact P1 case also returns
Unsupported on baseline. The standard core no-cgo scope passes. Full official
world coverage remains in ordinary Go/LLGo semantic tests. No broad skip was
added. All attempts, command logs and separate baseline checks are retained.

[Sources, compiler/binary hashes, plans, all raw phases, independent summary and
matrix verification, allocations, quality logs, held experiment and initial
boundary diagnostic](evidence/performance-native-lease-admission-2026-11-03.json.gz).

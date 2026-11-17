# Native call bounds and rejected scalar return experiment

The retained change folds physical and logical call limits once at host entry.
It reduces the closed general-call kernel's time by **6.3% under Go and 6.8%
under LLGo** in its fixed own-only regression event. These results do not
update the [last complete competitor comparison](performance-combined-optimizations-2026-11-16.md)
or close the native-call performance gate.

## Mechanism and contracts

The checked private entry computes the minimum of the logical slot limit and
the physical arena size, and the minimum of the logical depth limit and the
depth reachable with the remaining private records. Depth minus record top and
both allocations remain invariant during one native run. Growth or a new host
entry recomputes the limits. The emitter can therefore remove the duplicate
physical comparisons while retaining the effective limit checks.

The call header and 24-byte records are unchanged. Existing header size,
record capacity and top are authenticated before access or growth; arithmetic
overflow is checked. Standalone and metered bodies keep independent logical
and physical checks. Canonical full-width values, cancellation polls, root
publication, cold PCs and admission checks retain their existing contracts.
The ARM64 sample removes six instructions/24 bytes from its caller, shrinking
1,512 to 1,488 bytes; its 1,280-byte callee is unchanged.

## Fixed events

Both events use Go 1.27.3, LLVM 22.1.8 and LLGo **9e45f103**, with
`-O3 -lto=full` and default GC, on the same macOS ARM64 M4 Max host. Each
cohort has six fixed fresh-process blocks and symmetric same-binary A/A
participants. General calls retain five million source iterations, two warmups
and seven measured calls. Complete P3 requests retain four workers, sixteen
clients, twenty 512-request windows and complete response checks. Setup,
startup, first call and cleanup are separate. No owned build, test, profile,
edit or analysis overlaps timing; external desktop load remains uncontrolled.

Ratios use paired block geometric means and marginal bootstrap 95% intervals.
Both relevant A/A intervals must fit [0.975,1.025] to resolve. No event is
extended or pooled after inspecting results. Managed allocation counters do
not account for foreign allocations, and no competitor acceptance is claimed.

The bounds-only event measures source base **266ac846** against **e9c2918f**.
Their complete source trees equal merged base **9f33573f** and rebased bounds
commit **21c00387**, respectively.

| Bounds-only head/base | Ratio [95% interval] | Controls |
| --- | --- | --- |
| General calls, Go | **0.9375 [0.9326,0.9427]** | Resolve |
| General calls, LLGo | **0.9317 [0.9274,0.9362]** | Resolve |
| Complete P3 CPU, Go 16 B | 0.9979 [0.9904,1.0053] | Resolve |
| Complete P3 CPU, Go 64 KiB | **0.9875 [0.9754,0.9969]** | Resolve |
| Complete P3 CPU, LLGo 16 B | 0.9972 [0.9882,1.0051] | Resolve |
| Complete P3 CPU, LLGo 64 KiB | 1.0117 [1.0015,1.0231] | Unresolved |

All request throughput controls are unresolved. The adverse LLGo 64 KiB CPU
point is retained; it is neither an accepted regression nor a passing result.

A separate three-source event adds complete scalar-pair return carriers to the
bounds change. It moves the indexed destination copy to a caller continuation
with a compile-known slot, retaining both words and cancellation ordering.
The sample removes another 24 bytes from each function, but the incremental
general-call ratios are **1.0058 [0.9944,1.0217] under Go** (controls resolve)
and **1.0111 [0.9977,1.0246] under LLGo** (controls unresolved). There is no
demonstrated incremental speedup. The combined Go ratio is 0.9439
[0.9376,0.9500] with resolving controls; the LLGo ratio is 0.9453
[0.9362,0.9545] with unresolved controls. Request contrasts and all controls
remain in the evidence. This experiment is **not integrated**; it remains on
`experiment/native-result-pairs-20261117` at **d65da11c**. Reduced instruction
count alone does not establish throughput improvement or isolate its cause.

## Validation and SDK provenance

The retained bounds source passes the full Go suite and **31 executable package
coverage gates above 90%**, four affected packages under LLGo, race/vet, no-cgo
and amd64 Docker semantic tests. Direct tests cover arena and record growth,
depth/top invariance, tighter logical quotas, saturated arithmetic, overflowing
slot bytes, malformed capacities/top and a short header. Existing linked-call
suites cover complete values, cancellation, roots, dynamic targets and cold PCs.
Docker execution is semantic evidence, not native amd64 performance evidence.

CI and build defaults subsequently move to LLGo **2e3e02f4**, which merges
#3127's Windows cgo metadata and reference-test/CI fixes. The timing events above
remain explicitly attributed to 9e45f103; they are not relabeled as measurements
from the newer compiler. Final target CI and the unified competitor comparison
remain separate requirements.

Local immutable evidence is under `artifacts/native-call-bounds-20261117` and
`artifacts/native-result-pairs-20261117`: source overlays and tree qualification,
binary/toolchain hashes, frozen campaigns, compact raw process records,
startup/setup phases, A/A controls, request counters, instruction audits and
validation logs. The [committed summary](data/native-call-bounds-2026-11-17.json)
retains both complete analyses and their evidence hashes.

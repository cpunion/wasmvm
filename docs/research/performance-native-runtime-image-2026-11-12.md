# Native return chains and component transaction costs

Go candidate `57ffbf5` improves complete P3 HTTP execution over the
`d469ce2` runtime (`9831c30` documentation checkpoint). **Competitive parity
remains open.** The same-cohort snapshot throughput gaps to Wasmtime are
**1.48× / 1.73×** for 16 B / 64 KiB responses.

## Retained changes

Native compilation discards compiler-only IR and private decoded function
bodies after publication. Runtime source metadata uses pointer-free records;
the interpreter retains its direct rich-source path. In a separate four-Store
heap diagnostic, live bytes fall from 32,658,224 to 16,389,688, scannable bytes
from 29,220,544 to 9,338,016, and live objects from 61,942 to 8,903.

The native backend preserves its validated C-owned return chain across ordinary
host returns. A frame-local token also preserves a parent chain across Go-only
foreign call veneers; actual foreign native entries invalidate stale tokens.
Exception searches and continuation capture/restore discard resume state.
Frame packets, logical quotas, metered execution and exact result handoff remain
checked. Diagnostics observe **122 -> 106 -> 58 native entries per request**,
with 50 host callbacks and 28 ordinary guest boundaries unchanged. Artifact
lookup uses bounded indexed rows and code identity checks; scalar source
immediates avoid rebuilding rich source records.

Private host-result storage reuses the execution's pending argument allocation;
public callback stacks remain independent and retainable. Component Caller
headers and their lifetime guards share one fresh allocation. Ownership
transactions keep their first reference identity inline and use a map for
additional identities. Immutable layout-cache hits avoid recursive builder
entry. Resource bindings are immutable, with an atomic affine bit preserving
single-winner moves and copied-header invalidation. Synchronous reads wait on
operation, callback and VM lease cancellations independently, avoiding a child
context and asynchronous cancellation bridge.

Overlapping continuations also exposed a real cleanup bug: an idle workspace
could overwrite another cached workspace without closing it. The surplus
workspace now closes normally. No finalizer supplies this ownership guarantee.

## Matched complete-request comparison

M4 Max/macOS arm64, Go 1.27.3, Clang 22.1.8, Wasmtime 48.0.1. Both sources
use the same Rust P3 HTTP component, four Stores, 16 clients and checked complete
responses. Mutable Reader and explicitly immutable snapshot inputs remain
separate. Fuel is off; ordinary bounds, cancellation and logical limits remain.
Compilation, input preparation, first batch and cleanup are outside warm timing.

Three randomized rounds, seed `2026111493`, include an identical-binary A/A
alias for every configuration. Each process executes a first batch, two warmups
and 30 measured batches of 4,096 requests. All **72 processes and 8,847,360
measured responses** pass status, headers, full body, trailers and CRC checks.
Whole-process CPU/instruction windows cover 8,552,448 checked requests. No owned
local builds, tests, profiling or cache pruning overlap timing.

| Payload | Configuration | req/s | A/A req/s | CPU µs/request | Instructions/request | Allocations/request |
|---|---|---:|---:|---:|---:|---:|
| 16 B | Base Reader | 35,975 | 36,808 | 110.23 | 1,285,355 | 882 |
| 16 B | Head Reader | 43,887 | 44,129 | 89.74 | 1,059,090 | 795 |
| 16 B | Base snapshot | 42,475 | 42,608 | 93.11 | 1,109,067 | 789 |
| 16 B | Head snapshot | 51,170 | 51,347 | 76.59 | 912,042 | 705 |
| 16 B | Wasmtime | 75,733 | 75,654 | 24.36 | 208,033 | 1* |
| 64 KiB | Base Reader | 25,398 | 25,515 | 154.81 | 1,857,002 | 978 |
| 64 KiB | Head Reader | 31,002 | 31,109 | 124.07 | 1,481,861 | 879 |
| 64 KiB | Base snapshot | 31,940 | 31,567 | 123.62 | 1,542,389 | 885 |
| 64 KiB | Head snapshot | 38,338 | 38,584 | 101.14 | 1,235,941 | 789 |
| 64 KiB | Wasmtime | 66,238 | 65,891 | 32.37 | 297,926 | 1* |

\* Go reference counters exclude Rust/native allocations. They do not establish
the reference runtime's total allocation count. All CPU measurements include
every process thread. Points are process medians; A/A aliases are not pooled
selectively.

Head/base snapshot throughput improves **20.47% / 20.03%**, with CPU lower by
**17.74% / 18.18%**. Reader throughput improves **21.99% / 22.06%**. Instructions
fall 17.60–20.20% across these configurations. All **36/36 timing and instruction
A/A controls** stay within ±2.5%; **34/36 CPU controls** do, with a 4.10% worst
CPU deviation. These local observations support this source comparison, not
cross-platform release acceptance. Snapshot CPU gaps remain **3.14× / 3.12×**;
Reader throughput gaps remain **1.73× / 2.14×**. Other engines' earlier Core
measurements are separate events and acquire no new P3 ranking here.

The data retains every lifecycle phase and observed process footprint, including
preparation. Input preparation covers the four owned workers and requires
explicit immutability; ordinary Reader input retains its copy/worker contract.

## Rejected experiments and remaining work

Independent screens reject new numeric reallocation APIs, small paired memory
guards, universal codec pools, generic ready-value probes, broader host service
loops, coallocated scalar invocation buckets, inline borrower storage and
ready-only read watchers as unproven whole-request improvements. In particular,
saving 38/50 scalar callback allocations did not reliably reduce CPU time.
The forward-edge polling prototype remains outside production: its small
instruction reduction did not establish a CPU gain, and exact-site cancellation
tests require adaptation to its proposed bound. The current polling policy
remains unchanged.

A separate dense guest-handle-table prototype passes three Go semantic groups
and observes 52,435 / 39,095 req/s in the final cohort. It retains host ID and
generation rules, but has no final LLGo/cold-lifecycle qualification and is
**not included** in candidate `57ffbf5` or its performance claims.

Warm profiling of intermediate `9e354fa` attributes about 17 sampled CPU
µs/request to native execution, with allocation and GC still material. Exact
allocation sampling identifies a fresh 64-KiB response buffer as roughly half
of allocated bytes in the large case. Next work must reduce actual canonical,
resource and buffer costs; allocation counts alone cannot close the gap.
Retain ordinary owned-response measurements alongside any explicit caller-owned
buffer API. No unchecked mode or competitor fallback is introduced.

## Validation and evidence

All 32 Go package groups pass; all 31 executable packages exceed the 90%
statement gate, minimum **90.85%**. Affected canonical/component/P3 tests pass
Go and fixed-profile LLGo on macOS arm64 and Linux amd64 Docker. The preceding
native bridge also passes slots/execute/nativeexec/vm/component/P3 under both
compilers/platforms. Linux emulation supplies semantic evidence only. Full
PR CI must still execute all packages on all four compiler/platform combinations.

Full execute/vm/component/P3 race tests pass on `532244b`; atomic reference
canonical/component/P3 race tests and final focused move/invalidation/read
cancellation races pass. Affected vet passes; the whole-module no-cgo build
passes on `57ffbf5`. Existing corpus/shard coverage and the pinned LLGo
`3d92a58 -O3 -lto=full` profile remain unchanged.

[Data, phase arrays, controls and source identities](data/performance-native-runtime-image-2026-11-12.json)
link to [all process and batch records](evidence/performance-native-runtime-image-2026-11-12.jsonl.gz).
Individual request latency arrays use summaries and raw-array hashes in the
published bundle; unmodified arrays remain in the local raw artifact. No
per-request latency acceptance is claimed. Distinct prototype/source events
remain labeled and cannot be compounded into one improvement.

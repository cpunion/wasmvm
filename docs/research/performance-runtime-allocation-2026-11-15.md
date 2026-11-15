# Runtime allocation and latest LLGo main

The largest coarse LLGo P3 profile group is allocator/GC work, followed by owned
host canonical and lifetime setup. This is a profiling direction, not an exact
exclusive-time decomposition. Final changes reduce fresh capability allocations
and immutable signature construction. **No resolved full-request CPU or
throughput gain, competitive parity or release acceptance is claimed.**

## Sources and method

Base engine **ec6396bf** / compiler-pin commit **2fb98dda**; final implementation
**1385093b**. The reused Go baseline binary identifies **55f3e93a**, whose Go and
module files match ec6396bf. The compiler-upgrade event separately compares LLGo
**6a482f7f** with latest main **0539ea6e** at fixed engine source. Final source
comparisons use 0539ea6e on both sides, LLVM 22.1.8, **`-O3 -lto=full`**, Go 1.27.3
and default production GC. No experimental LLGo runtime patch ships.

Host: Apple M4 Max/macOS arm64, with other user build/link/test and desktop
activity. Fixed cohorts are not extended or pooled. No owned build, test,
profile, analysis, edit or pruning overlaps timed cohorts. Intervals are
individual block-bootstrap 95% intervals; A/A controls resolve only when their
whole interval is inside 0.975–1.025. Full response bodies, status, headers,
trailers and CRCs are checked. Startup, validation/compilation, instantiation,
first call, two warmups and steady requests remain separate.

Each final normal-GC payload has ten blocks: 60 processes for Go base/head and
Wasmtime aliases, then an independent 80-process LLGo base/head, Go head and
Wasmtime event. Four Stores/sixteen clients execute twenty individually
interleaved 512-request batches plus a separate flush. Separate bounded no-GC
LLGo diagnostics have eight blocks/32 processes and five batches plus flush,
asserting zero measured collections. Production GC remains enabled.

## What latest LLGo main changes

At fixed engine source, allocations per 16 B request fall **1165 → 1014**
(**−13.0%**), and 64 KiB **1430 → 1195** (**−16.4%**), with resolving controls.
The interval-capture fix in LLGo #3100 is a plausible contributor; this event
upgrades the whole main range and does not isolate that one commit.
Normal-GC CPU points are 0.94560/0.89456 of the previous compiler, but own CPU
controls fail: these are not established 5.4%/10.5% speedups.

The independent general native-call compiler event has 140 processes, including
wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and an unchecked LLVM C oracle.
New/old LLGo warm time is **1.00634 [0.99738, 1.01595]**, with resolving controls;
no warm speedup is established. Reference time points remain **2.29–2.41x**.
This compiler-upgrade event precedes the final Caller allocation change; it is
not a fresh general-call measurement of the final source. C omits VM contracts.

## Distribution evidence and rejected experiments

Six separate profiles cover 131,072 checked warm requests each, with setup,
first call and warmups excluded. LLGo CPU Profiler has 44,899 samples; the coarse
allocator/GC group accounts for about **48.4% of cycle weight**, owned code 19.7%,
other runtime work 10.8%, probable generated guest code 8.4%. These weights are
not exact exclusive CPU fractions or directly comparable across runtimes.
The Go CPU Profiler returned only one sample and is rejected for distribution
claims. Time Profiler includes substantial kernel waits even under its Running
label; its percentages cannot be treated as exclusive CPU. Cumulative stacks
overlap. Full summaries, request checks and raw XML hashes remain in evidence.

At fixed main, a bounded no-GC counterfactual gives CPU **0.87820
[0.85872, 0.89840]**, with unresolved base CPU control. It neither measures an
exact GC percentage nor removes the substantial non-collection runtime gap.
An untimed Go allocation-site diagnostic finds about 45 resource reference,
39 Core Caller and 23 component Caller allocations per request. These are
allocation counts, not CPU shares. Ownership and expired aliases must survive.

Rejected experiments are retained separately:

- Outlining cold `getg` initialization reduces bounded no-GC instructions
  **2.3–2.5%**, with resolving instruction controls, but not CPU. Normal and no-GC
  throughput adverse points and unresolved controls remain. SDK main is restored;
  the unmerged runtime experiment is not shipped or proposed as a proven fix.
- Avoiding codec construction on transfer refresh leaves this async request's Go
  allocations unchanged. That path is not the measured hotspot.
- Combining a host call and its codec initially reduces Go allocations 696→678
  and CPU points about 0.7–0.9%, but regresses LLGo. Independent source ablation
  attributes the adverse direction to coallocation, not immutable metadata.
  At 16 B its bounded no-GC CPU point is **1.06945 [1.05482, 1.08360]**; own CPU
  controls remain unresolved. The original combined 16 B no-GC event gives
  **1.04580 [1.03705, 1.05630]**, with both own CPU controls resolving. Rejected.

The allocator routing explains why fewer objects can cost more. Actual LLGo
assembly allocates separate **152 B + 256 B** objects, or a combined **408 B**
object. Installed [BDWGC 8.2.12](https://github.com/bdwgc/bdwgc/blob/v8.2.12/thread_local_alloc.c) uses 16-byte granules and 25 tiny free lists;
`GC_malloc_kind` branches to `GC_malloc_kind_global` at index ≥25, including its
interior-pointer allowance. The combined object crosses that threshold, while
the separate objects stay on the thread-local route. Headers, library hash and
both generated/allocator disassemblies are retained. This establishes a routing
change; exact allocator-versus-code-layout shares of the CPU regression are not
isolated. It does not establish a new LLGo compiler defect.

## Final implementation and complete requests

Core Caller guards are allocated individually for the first capability, then
in fresh groups of four within that execution lease. A batch is **256 B** on
this host, retaining the tiny-object route. Every slot is consumed once; copied
headers retain their independent guard through GC. No capability is reset or
rebound. Finished drivers release their last block; retained capabilities keep
only their blocks alive. Small batches bound retention and unused capacity.

Result handoffs reuse two immutable signature descriptors under the existing
exclusive Store lease, distinguishing absent/present results. Scheduled work
shares its private immutable empty signature. Arguments, ownership transactions,
Callers, budgets and cancellation remain fresh. The rejected host-call/codec
coallocation is absent.

| Managed allocations/request | 16 B base → final | 64 KiB base → final |
| --- | ---: | ---: |
| Go | 696 → **663** (−4.7%) | 756 → **714** (−5.6%) |
| LLGo main | 1014 → **981** (−3.3%) | 1195 → **1153** (−3.5%) |

All allocation-count controls resolve. Go bytes fall about 54,673→54,233 and
124,321→123,881; LLGo about 73,953→72,959 and 214,625→213,489. Foreign Rust/C,
JIT and linear-memory allocations are outside these host allocation counters.

| Final/base CPU [95% interval] | 16 B | 64 KiB |
| --- | ---: | ---: |
| Go normal GC | 0.99404 [0.98483, 1.00437] | 0.99588 [0.98856, 1.00327] |
| LLGo normal GC | 0.98754 [0.97274, 1.00394] | 0.98622 [0.94371, 1.02406] |
| LLGo bounded no GC | 0.99378 [0.97296, 1.01373] | 0.98646 [0.97493, 1.00333] |

Both Go CPU controls resolve; LLGo normal controls fail. No-GC head 16 B and both
64 KiB CPU controls fail. No CPU interval establishes improvement. Go
instructions fall about 0.3%, and LLGo no-GC about 0.6%, with resolving own
instruction controls. Throughput gains are unproven; every control and phase
remains in the machine-readable report. No-GC is diagnostic, not acceptance.

Final P3 Go/Wasmtime CPU points are **2.34x/2.10x**; LLGo/Go **2.75x/2.49x**,
LLGo/Wasmtime **6.14x/5.45x**. Reference CPU controls fail in these events, so the
points describe the remaining broad gap, not precise certified ratios. Wazero
and Wasmer do not supply equivalent P3 worlds in this comparison. General native
calls, host canonical dispatch/resource lifetimes, scheduler/runtime work,
interpreter/other kernels and cross-platform acceptance remain open priorities.

## Validation and evidence

All 32 Go root test packages pass; all 31 measured executable packages meet 90%
statement coverage. Affected VM/component/P3 race tests and both extra Go modules
pass. LLGo executes VM/component/P3 tests under the final production profile.
New cases retain copied Callers across twenty callbacks and two leases, trigger
GC and check expired capabilities/contexts. Repeated result handoffs check fresh
ownership, request contexts and failed-consumer cleanup. Final-head dual-platform
CI is required before merge.

The [report](data/performance-runtime-allocation-2026-11-15.json) links
**1380 checked process records in 22 separate cohorts**, all rejected experiments,
partial source overlays/patches, toolchain identities, allocator/code evidence,
load snapshots, raw lifecycle phases, GC and allocation counters. Only unused
per-request timing vectors become count/extrema/canonical-JSON hashes; originals
and full traces remain local. Published profile summaries exclude environment-
bearing trace metadata. Protocol errata include the failed no-GC environment
setup, wrong-runner startup and corrected test/profiler expectations. No missing
rows are reconstructed and no failed event is relabeled as passing.

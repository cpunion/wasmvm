# Free-handle bookkeeping and remaining GC costs: 2026-11-04

Status: local affected semantics, original twenty-case comparison and separately
planned duration/public-call controls verified. Final-head CI and review are
required before integration. Controlled performance and expanded WASI acceptance
remain open.

## Representation and contracts

The [merged PR113 baseline](pr113-final-ci-2026-11-04.md) uses typed trace layouts.
Its current-source profile still spends substantial time tracing, publishing
native allocation batches, compacting payloads and maintaining free handles.
Five-second samples are diagnosis only; their timings do not enter comparisons.

Replace the separate free-index slice with a numeric link in each free handle
slot. The link uses the existing four-byte alignment space: both pinned compilers
report **16-byte slots** before and after. The Store shrinks by 16 bytes within
each compiler (Go 624→608; LLGo 632→616); their absolute layouts differ.
There is no free-index backing array to grow or copy. The actual full-LTO LLGo
release-object body changes from 75 to 64 instructions and loses its common-path
`appendValue[uint32]` call. Cold panic paths remain; this is not a claim that the
entire runtime allocates nothing.

LIFO reuse, generation checks/retirement, uncommitted native reservations, all
root kinds, quotas, safepoints and the numeric native view keep their contracts.
Reservation failures return all unused handles without exposing a valid prefix.
Both portable and native stores use the same chain. No rejected compaction or
object-padding change is included.

## Fixed-source comparisons

Engine source `18d7b09592b6b93e0b1d2e24e2321a9a4cc7714d` is compared with
`354abd3e49d27be4a1fed5cb6004e39ae5b580b7`, whose engine inventory equals merged
main `6d791852c86655a21ce29a22bd0a15c07bb74333`. Go 1.27.3 and LLGo
`983465e779527ff496857a9a1915c1b57131091f` use the unchanged production profile
(`-O3 -lto=full` for LLGo). Reference versions/SDKs and fixture bytes are unchanged.
These macOS arm64 measurements use an Apple M4 Max on a shared host; they are not
controlled Linux/macOS portfolio acceptance.

The primary six-case duration study independently verifies 220 observations,
44 alternate checks, 50 affine-oracle checks and 44 summaries. Five process blocks,
two warmups, seven measured calls and exact native/interpreter baseline A/A.
Allocation uses the same 300k-object live chain: 12-byte packed payload plus 64
accounting bytes/object, below the unchanged 64 MiB quota even with old/new chains.

Allocation, **ns/object**:

| Backend | Go base → head | LLGo base → head | Reference |
| --- | ---: | ---: | ---: |
| Native | 50.340 → 48.367 (−3.9%) | 60.760 → 54.354 (−10.5%) | Wasmtime 9.560 |
| Interpreter | 119.792 → 114.997 (−4.0%) | 146.473 → 145.634 (−0.6%) | WasmEdge 110.900 |

Native A/A changes −2.09% Go / −0.59% LLGo. Go's smaller gain and LLGo interpreter
−0.6% versus A/A +1.14% need caution. LLGo native remains **5.69x Wasmtime** and
12.4% above Go in this campaign. Cross-campaign absolute values are not pooled.
Retain Go native direct-GC+1.99% (A/A +1.11%) and all other primary controls.

The original twenty-case scope independently verifies 864 VM/reference and 27
ordinary-C observations,288 alternate checks, twelve explicit unsupported GC
outcomes and297 regenerated summaries. All five named reference configurations
and nine C algorithms remain. Native short allocation changes −2.2% Go / −11.1%
LLGo; interpreter allocation changes +0.9% Go / −6.7% LLGo. The one-minute host
load changes 8.51→7.25. GC is not assigned an unsupported ordinary-C equivalent.

Retain every original adverse row, including Go empty interpreter+47.7%, LLGo
empty native+9.0%, Go/LLGo short i31 native+8.5%/+4.4%, LLGo interpreter scalar
memory+1.67% and dependent memory+2.27%, and LLGo native notify/timeout
+16.6%/+11.0%. Tiny-loop and single-call percentages are not precise instruction
costs. No startup, clock, empty-call or other median is subtracted.

## Separate adverse controls

A separately declared six-case duration study verifies 240 observations,
48 alternate checks, 75 oracle checks and 48 summaries, with exact baseline A/A.
It uses 100M native / 5M interpreter i31 iterations and 5M iterations for native
notify/timeout and interpreter scalar/dependent memory. All original shorter
observations remain intact; these scopes are not pooled or substituted.

| Control | Go delta / A/A | LLGo delta / A/A |
| --- | ---: | ---: |
| gc/i31-roundtrip / native | +0.07% / +0.00% | +0.22% / +0.01% |
| gc/i31-roundtrip / interpreter | +1.25% / +0.54% | +0.75% / -0.75% |
| threads/notify-empty / native | -0.13% / +0.27% | +0.21% / -0.18% |
| threads/wait-timeout / native | +2.60% / +1.19% | +3.30% / +0.12% |
| core/memory-load-store / interpreter | +0.53% / +0.46% | +1.52% / +0.41% |
| memory/dependent-address / interpreter | -1.07% / +0.40% | +1.57% / +0.02% |

The longer study still observes LLGo scalar/dependent interpretation about 1.5%
slower and native timeout+3.3%; these remain adverse outcomes. This candidate's
GC gain does not close controlled or whole-portfolio acceptance.

A separate 10000-public-call study verifies 180 observations,36 alternate checks
and 36 summaries for empty and 64-step guests, both backends/compilers, exact A/A,
Wasmtime/Wasmer and wazero default/enabled cancellation policies. LLGo empty
native changes 836.733→860.221ns (+2.81%, A/A +2.55%); interpreter changes
646.467→655.167 (+1.35%, A/A +1.16%). Retain Go 64-step interpretation+3.02%
(A/A−1.42%). Empty compiled references measure Wasmtime 59.046, Wasmer 134.562,
wazero default 13.558/enabled 220.713 ns. Public-boundary/compiler gaps stay open.

## Validation and CI dependency recovery

Both exact compilers pass the state/SIMD/execute/VM suites and 48 expected
execution reports each, all five required extensions/three tiers including
forced Stack Switching GC. Both additionally execute the native allocation,
trap and cancellation suite. New tests cover failed allocation order, payload
clearing, stale generations, identity retirement, invalid mixed-size native
prefixes, unused reservations, repeated collection and close.

The initial footprint verifier incorrectly required the entire Store layout to
match across compilers. All four helper executions succeeded; corrected checks
verify each compiler's 16-byte slot and 16-byte Store reduction. The original
verifier/log and correction remain. No timing measurement had started at failure.
No new LLGo bug is inferred; issue 2983 remains with its other contributor.

PR113's LLVM key download failed DNS resolution before tests in multiple jobs.
A small Linux-only installer now matches the pinned setup action's dependency
set, adds bounded curl/APT retries and verifies LLVM 22.1.8. All five setup
consumers use the correct checkout path. macOS dependency setup, compiler pins,
release flags, 31-package semantic matrix and benchmark scope remain unchanged.
The shell passes syntax validation; both workflows and all five setup paths are
parsed/checked, and the five shard-planner tests pass. Actual final-head hosted
installation, full semantics/coverage and review remain required before merge.

[Complete observations, all adverse controls, compiler/source inventories,
semantic reports, actual assembly/size probes, current-source profile,
precursor experiments and PR113 final CI](evidence/performance-gc-free-chain-2026-11-04.json.gz).

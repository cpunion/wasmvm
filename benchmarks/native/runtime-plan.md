# Complete Core 2.0 runtime development scope

This branch connects generated functions to the public VM and shared runtime.
It does not close M2's performance gate or implement Core 3.0, components or WASI.

Compare exact baseline `4286585` and candidate interpreter API execution, candidate
native API and reusable dispatch under Go/LLGo, and pinned wazero 1.12.0,
Wasmtime 48.0.1 and Wasmer 7.3.0. The baseline public native backend is unavailable;
its interpreter rows remain explicitly identified. Preparation, first call and
warmed execution are separate. Dispatch additionally splits frontend phases and
native emission/publication; public API timings include frame/context allocation.

The declared portfolio is the existing six embedded workloads plus four owned
fixtures: indirect calls, i32x4 SIMD, overlapping 8-KiB bulk memory operations,
and globals/tables. Each uses runtime seed/count parameters and an independent
Python oracle. State reset is inside each call. The bulk case performs fill plus
overlapping copy; its unit is one source iteration, not a single byte operation.
Generate with `python3 benchmarks/native/runtime_fixtures.py` and pinned
wasm-tools 1.258.0. Manifest digests identify the exact shared bytes.

Every family must pass calibrated own-baseline and competitor comparison under
both compilers on both required targets before merge. Current local diagnostics
are advisory. Explicit runtime helpers establish semantics but not speed; calls,
SIMD and mutable state need their own measurements. Existing PR #26 remains the
only open contribution while later capabilities stay on development branches.

## Local integration diagnostic

Three randomized process blocks on an M4 Max, seven warm calls and two warmups
per process, with 60 additional alternate-input checksum checks. Medians below
are ns/source iteration; preparation and first calls are in the raw record.
These are uncalibrated diagnostics, not a performance pass.

| Family | Native Go | Native LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Sustained integer | 1.610 | 1.673 | 1.766 | 1.669 | 1.974 |
| Sustained scalar memory | 0.562 | 0.634 | 1.158 | 0.594 | 0.654 |
| Direct calls | 362.054 | 1383.400 | 1.917 | 1.273 | 1.193 |
| Indirect calls | 300.239 | 1262.550 | 2.431 | 1.267 | 0.933 |
| SIMD i32x4 | 627.129 | 2708.012 | 2.263 | 2.100 | 2.200 |
| Bulk overlap | 3270.459 | 4936.666 | 202.625 | 190.000 | 191.000 |
| Globals/tables | 708.508 | 2724.475 | 34.112 | 25.200 | 5.900 |

The call/SIMD/state helper architecture is substantially slower than competitors,
including in reusable dispatch. Moving setup outside the timer does not remove
this gap. Guest-to-guest entry and SIMD/state code generation remain required
optimization work. The empty-call control is resolution-limited in two C
reference runners and is not ranked. All ten workloads, own-baseline/interpreter
rows, dispatch rows, samples, checksums, scripts and binary hashes are preserved
under `artifacts/native-runtime/`; the checked integration source tree is
`29f9dd4bfff6ee4fd555698f3785204a65343106`. Earlier PR #26 measurements cover its
narrow integer compiler scope separately.

The next call optimization expands bounded, straight-line, nontrapping integer
leaf functions in the caller's generated code. The existing direct-call workload
is the primary affected case; indirect calls, integer/memory loops and all
alternate-input checks remain regression rows. Use `f0b1ded` as the exact owned
baseline. Keep preparation, first call and warm execution separate.

Expansion uses disposable registers, preserves observable caller slots and
never creates a hardware guest stack. It applies only to the unmetered tier.
An execution selects the expanded entry only when the original logical callee
frame would fit both depth and slot quotas; otherwise it selects the ordinary
native entry. The precise-fuel tier, recursive/indirect/imported calls and
effectful callees retain their complete existing execution paths. This step
does not accept the remaining guest-call architecture or the full native profile.

After result forwarding, `2365102`'s three-block local direct-call diagnostic
measured Go/LLGo 0.784/0.769 ns per iteration, versus wazero 1.321, Wasmtime 0.967
and Wasmer 0.860. The identical-round native baseline was 260.909/1075.161.
Indirect calls remain slow (LLGo 1087.408 versus Wasmer 0.800); integer/memory
rows and all 75 alternate-input checks are retained. The earlier `27fff72`
diagnostic was slower than Cranelift and remains in `artifacts/native-inline`.
These are diagnostic observations, not calibrated or full-profile acceptance.

The planned local direct-call check uses the existing 20/60 independent process
blocks, seven warm samples, two warmups, identical-binary controls, simultaneous
intervals and unchanged ±0.5%/±0.25% margins. All candidates and controls use
25 million iterations of the same direct-call fixture; the independent Python
oracle calculates its checksum before measurement. Three descriptive baseline
processes retain 150,000 iterations. `warm-call-interleave-v1` now also applies
to the owned public API runner; only warm execution is gated by this check.
Preparation/first call, all earlier observations and both target/compiler
requirements remain separate. Passing this case cannot accept indirect calls,
helpers or the complete M2 portfolio.

The completed 60-block local run at `982b379` remains **blocked**. Warm medians
were Go/LLGo 1.194/1.181 ns per iteration, wazero 1.973, Wasmtime 1.390 and
Wasmer 1.183. LLGo ratios were 0.603 [0.586, 0.620] against wazero, 0.851
[0.821, 0.882] against Wasmtime, and 1.002 [0.974, 1.035] against Wasmer.
The LLGo A/A interval was [0.968, 1.019]; neither host compiler calibrated.
Thus the short diagnostic's apparent Wasmer win is not an accepted result.
The oracle checksum was `249740414`; all observations, planned looks, binary
hashes, source hashes and dispatch orders are in
`artifacts/native-inline/gate-results.json`. No margins or failed records changed.

Go/LLGo full semantic suites pass on arm64, with the complete native Core 2.0
corpus and additional all-slot/cold-entry, alias, logical-quota, exact-fuel,
cancellation and aggregate-code-budget tests. Go amd64 execution under
Docker/Rosetta passes the compiler, native driver, complete execution corpus and
public API tests; this is semantic evidence, not an amd64 timing result.
Codegen coverage is 97.32%, native execution 95.59%, public API 96.92%, and the
updated phase adapter 95.2%; every executable owned package remains above 90%.
The branch preserves one open PR (#26) and does not claim M2 completion.

The next optimization uses exact baseline `4d04b6b`: guarded expansion of
bounded integer leaves at indirect call sites. The primary affected workload is
`indirect-calls`; direct calls and sustained integer/memory loops are regression
rows. Keep all existing alternate-input oracles, pinned competitors, separate
phase timings and acceptance margins. Store-owned native tables contain only
reference handles. Generated code checks the live table index and target identity
before expansion; the ordinary call path retains all other targets and traps.
Native views are refreshed after host calls and helpers, including table growth.
Precise fuel and logical frame/slot limits keep their existing contracts.

The initial guarded implementation (`627f944`) passed the full Go/LLGo arm64
semantic suites and Go amd64 execution, including the complete Core 2.0 corpus.
Coverage was codegen 97.37%, native execution 94.71%, state 97.78%, public API
96.92%; all executable owned packages passed 90%. New tests cover target/type
changes, null/bounds traps, canonical type aliases, table growth, imported tables,
host/cross-instance reentry, reference roots, rollback, exact fuel and logical
quotas. Native storage also passed `cgocheck2`, vet and the no-cgo compiler build.

Its three-block local diagnostic measured indirect-call Go/LLGo at 1.022/0.908
ns per iteration, baseline 241.263/1068.711, wazero 2.093, Wasmtime 1.056 and
Wasmer 0.756. The candidate remained slower than Wasmer; this is not acceptance.
All four families and 95 alternate-input checks are retained in
`artifacts/native-indirect/guarded-results.json`. The primary diagnostic used
one million iterations for every engine, including the baseline. The next
change caches stable metadata during each native entry using spare arm64
constant registers; it does not cache mutable elements or relax any checks.

Caching alone (`05bcc50`) did not improve the three-block diagnostic: LLGo was
1.107 ns/iteration versus guarded 1.062 and Wasmer 0.843 in the same run. Those
samples remain in `cached-results.json`. The next implementation (`3519b7a`)
moves checked-call exits out of the successful loop, compares cached identities
directly, and fuses single-use integer products into arm64 MADD while preserving
shared products and aliases. Its diagnostic measured Go/LLGo 0.776/0.754 versus
cached 1.018/1.071, guarded 1.036/0.904, wazero 2.059, Wasmtime 1.072 and Wasmer
0.762. Direct-call/integer/memory rows and all 95 alternate-input checks remain
in `compact-results.json`. These short samples still do not accept the feature.

The planned local indirect-call check uses 25 million iterations, 20/60
independent process blocks, seven warm calls, two warmups and per-call random
interleaving. Use the same A/A controls, simultaneous intervals, ±0.5% parity
and ±0.25% calibration margins as the earlier checks. The independent Python
oracle calculates the expected result before measurement. Three descriptive
`4d04b6b` baseline processes retain the original 15,000 iterations; their cost
is reported separately. Both planned looks and every raw observation are retained.
This checks one indirect-call case on this host, not the whole native profile.

For architectural comparison, the pinned Wasmer 7.3.0 CLI's Cranelift output
for these Wasm bytes keeps the loop in registers and uses a checked hardware
indirect call (`blr`); the arithmetic callee uses `madd`, and each callee returns
without a host-runtime handoff. The independent Wasmtime 48.0.1 CLI IR also has
native indirect calls and type/table checks. These CLI diagnostics are separate
from the measured C API runners; files and commands are retained under
`artifacts/native-indirect/{wasmer-code,wasmtime-code}`. They explain why host
round trips were the initial architectural bottleneck without asserting that
the two independently built tools have byte-identical generated code.

The final local indirect-call check at `8937a5e` (engine `3519b7a`) retained
60 process blocks, both planned looks and 426 observations. Independent oracle
checksum: `643576559`. Warm medians are ns/source iteration; compile/validate
and first-call medians are milliseconds, measured separately:

| Runtime | Warm ns/iteration | Compile/validate ms | First call ms |
| --- | ---: | ---: | ---: |
| wasmvm Go | 0.773 | 0.196 | 19.292 |
| wasmvm LLGo | 0.772 | 0.579 | 19.239 |
| wazero | 1.983 | 0.139 | 49.844 |
| Wasmtime | 1.069 | 1.177 | 26.707 |
| Wasmer | 0.775 | 1.256 | 19.281 |

LLGo passed this host's competitor comparisons and A/A calibration: ratios
were 0.3891 [0.3852, 0.3935] versus wazero, 0.7218 [0.7200, 0.7233] versus
Wasmtime, and 0.9950 [0.9900, 0.9981] versus Wasmer. Its A/A interval was
[0.99851, 1.00244], inside the unchanged ±0.25% band. Go passed the competitor
comparisons but its A/A interval [0.99660, 1.00121] remained inconclusive.
**The combined check remains blocked**, and does not authorize a merge or
accept other indirect targets, the complete native profile or another target.
Baseline Go/LLGo warm medians were 254.386/1059.186 at 15,000 iterations;
those three baseline processes are descriptive, not simultaneous inference.
The raw record, hashes, commands, phases and sample orders are retained in
`artifacts/native-indirect/gate-results.json`; earlier failed diagnostics remain.

Final semantic regression passed under Go and LLGo on arm64 and Go on amd64
under Docker/Rosetta, including the complete native Core 2.0 corpus and the new
full/partial/exhausted metadata-cache and shared-product/alias tests. Final
coverage: codegen 97.38% (2715/2788), native execution 94.71% (179/189), state
97.78% (440/450), public API 96.92% (723/746). All executable owned packages
remain above 90%. PR #26 remains the sole open PR, with its independent ARM64
calibration gate still blocked; these later changes remain on the development
branch. General native guest calls, SIMD/state/bulk performance and the remaining
Core 3.0/Threads/component/WASI deliverables are still open.

The next vector-code increment uses exact baseline `c7fbc5c`. The affected
portfolio includes the existing `simd-i32x4` workload, with direct/indirect calls
and sustained integer/memory loops as regressions. Compare the public native
API under Go and LLGo with the same pinned competitors, independent checksums
and separate preparation/first/warm phases. Introduce full-width register
ownership and native integer lane/bitwise operations on baseline SSE2 and NEON;
unsupported packed operations keep explicit typed helpers. Preserve every cold
continuation, fuel boundary and full 128-bit value across spills and host exits.
This increment does not accept the complete SIMD performance profile or M2.

The first vector implementation (`686b6ac`) lowers 122 additional basic SIMD
opcodes, plus the existing vector literal, on both targets. It covers splats,
lane extraction/replacement, bitwise/select/reductions, integer comparisons,
abs/neg, shifts, wrapping add/sub/mul, i8/i16 saturating arithmetic and integer
min/max. Baseline SSE2 uses owned instruction sequences where a packed operation
is absent; NEON uses native operations. Remaining packed operations retain
typed runtime helpers. No optional CPU feature is assumed.

Three-block diagnostics measured LLGo SIMD at 2.029 ns/iteration versus the
baseline 2355.079, wazero 1.775, Wasmtime 1.800 and Wasmer 1.800: the initial
implementation remained slower. Forwarding (`e02330f`) reduced this to 2.000
but did not reach parity. The affine version (`f08d7df`) measured Go/LLGo
1.363/1.483 versus wazero 1.796, Wasmtime/Wasmer 1.800. Every run retains all
five families and 115 alternate-input checks in `artifacts/native-simd/`.
These 10,000-iteration SIMD measurements include public-call overhead and are
diagnostic only; earlier failures remain in the same directory.

The planned sustained SIMD check uses 25 million iterations, the independent
Python lane oracle, 20/60 process blocks, seven warm calls and two warmups.
Keep per-call randomized interleaving, simultaneous intervals, ±0.5% parity
and ±0.25% A/A calibration. Three descriptive baseline processes retain the
original 10,000 iterations. Preparation and first-call phases remain separate;
this one case cannot accept the full SIMD portfolio or another target.

The sustained run completed both planned looks (60 process blocks) for engine
`f08d7df`, with test/documentation head `0028450`. Final warm rates and preparation
medians are below; candidates use 25 million iterations per call. Baselines use
10,000 iterations and are descriptive only.

| Runtime | Warm ns/iteration | Compile/validate ms | First call ms |
| --- | ---: | ---: | ---: |
| Own Go | 1.286 | 0.158 | 32.109 |
| Own LLGo | 1.283 | 0.235 | 32.050 |
| wazero | 1.807 | 0.108 | 45.049 |
| Wasmtime | 1.793 | 0.702 | 44.878 |
| Wasmer | 1.793 | 1.108 | 44.745 |
| Own baseline Go | 528.979 | 0.150 | 5.285 |
| Own baseline LLGo | 2306.154 | 0.244 | 23.904 |

LLGo competitor ratios (simultaneous intervals) are 0.7059 [0.6809, 0.7166]
against wazero, 0.7149 [0.7097, 0.7191] against Wasmtime and 0.7187
[0.7135, 0.7269] against Wasmer. All competitor comparisons are faster.
However, Go A/A [0.98370, 1.00541] and LLGo A/A [0.99052, 1.00882] both fail
the unchanged ±0.25% calibration band: **the combined gate is blocked**.
Host load averages before/after were 12.78/11.98 (one minute). No rerun replaces
these observations. Raw commands, source/binary hashes, phases, sample orders,
both looks and checksums remain in local `artifacts/native-simd/gate-results.json`.
Native amd64 performance and the complete SIMD portfolio are not accepted.

## SIMD memory increment

Baseline: `4407b0c`. Lower all 22 fixed-SIMD memory operations through the shared
widened address check and existing memory owner, with explicit full-width spills
and independent checked cold entries. Extend the bounded local range analysis;
do not infer loop invariants or assume aligned guest addresses.

Before implementation, declare three additional public native API workloads:
unaligned overlapping 128-bit load/store, signed widening plus splat loads, and
lane load/store plus zero-extending loads. Each uses deterministic initialization,
an independent byte/lane Python oracle and the pinned Wasm producer. Keep own
Go/LLGo base/head and wazero/Wasmtime/Wasmer rows, alternate inputs and scalar,
call and existing SIMD regression rows. Retain startup, preparation, first call
and warm execution separately. Short diagnostics do not accept the profile;
sustained checks retain the existing 20/60-block protocol, seven samples, two
warmups, ±0.5% parity and ±0.25% A/A calibration, with no discarded failures.
Each sustained candidate call executes 25 million iterations. Adjust simultaneous
intervals across all three families (24 comparisons and two planned looks).
The baseline retains 10,000 iterations in three descriptive processes per family.

The first implementation (`4b59731`) removed all 22 memory helper exits. Its
three-block LLGo diagnostics were 4.083 ns/iteration for full vectors, 1.408 for
widen/splat and 5.733 for lane operations; pinned competitors measured
4.000–4.200, 1.163–1.200 and 6.400–6.617 respectively. Widening remained slower.
The register-address version (`246b906`) measured 3.733, 1.229 and 5.112 versus
3.600–4.200, 1.017–1.100 and 5.600–5.987. Both runs and the five regression
families remain in local `artifacts/native-simd-memory/`. These short calls
include public-call overhead and do not resolve sustained equivalence.


The vector-alias version (`9cc2f0f`) completed the declared sustained run. The
full-vector family stopped at 20 blocks on a confirmed slower comparison; both
remaining families completed 60 blocks. Median warm ns/iteration:

| Family | Own Go | Own LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full vector overlap | 5.095 | 5.117 | 5.205 | 5.392 | 4.680 |
| Widen and splat | 1.219 | 1.220 | 1.231 | 1.217 | 1.213 |
| Lanes and zero loads | 5.484 | 5.465 | 6.238 | 6.198 | 6.121 |

LLGo/full versus Wasmer is **1.0836 [1.0458, 1.1373], slower**. Lane comparisons
are faster against all three references (paired ratios 0.8714/0.8769/0.8870).
Widening is inconclusive against Wasmtime and Wasmer. All A/A comparisons remain
inconclusive, so the combined gate is **blocked**. Raw phases, source/binary
hashes, alternate-input checks and both looks are retained in local
`artifacts/native-simd-memory/aliases-gate-results.json` and `aliases-results.json`.
The descriptive baseline remains in the same record; no baseline helper timing
is used as a competitor acceptance criterion.

Disassembly of the unchanged pinned Wasmer library's serialized full-vector
function shows base-plus-immediate addressing before the UXTW-indexed access.
The own emitter instead forms a dependent base-plus-index-plus-immediate chain
for unaligned offsets. The next change selects the former generic addressing
form for proven accesses, sharing it between scalar and vector instructions.
Checked cold entries and memory ownership remain unchanged. Its diagnostics
retain all eight families, the preceding engine and the original baseline;
acceptance margins and sustained sampling remain unchanged.


The base-offset implementation (`aa89f0d`) has passing Go/LLGo semantics on both
required targets ([CI](https://github.com/cpunion/wasmvm/actions/runs/40473998419)).
Its three-block diagnostic reduces full-vector LLGo from 3.938 to 3.401
ns/iteration, versus 3.879/4.207/3.791 for the references in that same run.
All eight families and 175 alternate-input checks are retained in
`artifacts/native-simd-memory/baseoffset-results.json`. The subsequent sustained
run is recorded separately in `baseoffset-gate-results.json`; improvements do
not overwrite the earlier slower gate.

The sustained base-offset run completed 60 blocks for each family:

| Family | Own Go | Own LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full vector overlap | 3.440 | 3.434 | 3.867 | 4.100 | 3.596 |
| Widen and splat | 1.059 | 1.057 | 1.074 | 1.057 | 1.057 |
| Lanes and zero loads | 5.204 | 5.177 | 6.051 | 6.045 | 5.998 |

Warm ns/iteration, with preparation and first calls retained separately. LLGo
full-vector ratios are 0.8889 [0.8817, 0.8954], 0.8413 [0.8341, 0.8492] and
0.9559 [0.9500, 0.9617] versus wazero/Wasmtime/Wasmer: the previously slower
path is now faster against every reference. Lane ratios are 0.8573/0.8584/0.8631,
also faster. Widening remains inconclusive versus Wasmtime and Wasmer; their
upper bounds are 1.0067 and 1.0062. LLGo A/A intervals are [0.9940, 1.0021],
[0.9917, 1.0047] and [0.9952, 1.0058], outside ±0.25%. The combined gate remains
**blocked**; no result is discarded and native amd64 performance remains open.

An additional experiment (`38340db`) forwarded the equality established by a
preceding `local.tee` within the same bounded region. It passed local Go/LLGo
semantic checks, including all-slot comparisons at original PCs and a negative
poll-boundary fault injection. However, its three-block diagnostic did not
materially improve the targeted LLGo widening path (previous 1.099, candidate
1.101 ns/iteration), and lane timing was 5.130 versus 5.142. The experiment was
reverted, preserving the compact proven addressing implementation. All eight
families and 175 alternate-input checks remain in local `tee-results.json`;
no additional sustained run or redundant target CI was launched for this
unretained change. The final production source equals `aa89f0d`, whose semantic
matrix and sustained results above remain the applicable evidence.

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

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

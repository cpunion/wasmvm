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

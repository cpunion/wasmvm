# Scalar floating-point development scope

Status: local implementation work; not accepted and not a second contribution PR.

Implement native f32/f64 comparisons, arithmetic, rounding, min/max, sign operations and bit reinterpretation on both required ISAs. Use the existing function compiler, continuation protocol, frame slots and FP environment; do not add another execution framework.

Before submission, compare the exact own baseline, Go/LLGo candidates and pinned wazero, Wasmtime and Wasmer native engines using the same parameterized Wasm binaries. Declare three equally visible families: dependent f32/f64 multiply/add/divide chains, min/max/comparison with signed zeros, and rounding/square-root chains. Preserve the integer regression series. Separate preparation, first call and warm execution. Require parity or better for every affected family on Linux amd64 and macOS arm64 with calibrated A/A controls; no float feature is accepted on integer-only evidence. Numerical inputs must remain finite, include alternate seeds/counts and use independent expected results; NaN/edge behavior belongs to the semantic suite.

The initial six owned fixtures use one million iterations, the shared `run(i32 seed, i32 iterations) -> i32` protocol and seed `305419896`. f32 rounds after every operation; f64 checksums fold both output words. `float_fixtures.py` computes expectations independently in Python, and wasm-tools 1.258.0 produces the binaries. Alternate seed/count checks precede timing. Use the existing native and reference phase runners; no separate execution harness or public-backend claim is introduced.

The implementation emits all 44 operations in this scope on both ISAs. amd64 uses only baseline SSE2, including explicit rounding sequences; arm64 uses scalar FP instructions. FP bit reinterpretation and sign operations preserve NaN payloads. SSE min/max handle unordered inputs and signed zeros explicitly. The common compiler continues to enforce fuel boundaries, canonical scalar slots and cancellation. Native memory/calls and public runtime integration remain open.

Local implementation evidence at `c8b9a66`: all 44 instructions execute in Go/LLGo arm64 tests and isolated amd64 Go tests. Register caching and copy forwarding reduce the initial 3–6× deficits substantially. The six-case M4 Max diagnostics are close on f32 arithmetic and both rounding cases, but min/max remains about 16–20% slower and f64 arithmetic has a smaller deficit. They do not authorize a PR merge. Raw stage comparisons, 90 alternate-parameter checks and reproduction metadata are retained locally in `artifacts/native-float/`. Exact codegen statement coverage is 94.5%; the public native API and full-profile execution remain incomplete.

## Next scalar conversion scope

Implement all 26 remaining scalar integer/float numeric conversions, including
promotion/demotion, checked truncation and the eight saturating truncations.
Extend the affected portfolio with signed/unsigned i32/f32 and i64/f64 conversion
chains, declared before implementation; retain the six arithmetic cases and the
integer regression series. Preserve exact trap categories/locations, negative
fractions truncating to unsigned zero, saturation, NaN policy, fuel and every
existing continuation guarantee. No conversion feature is accepted on arithmetic
benchmarks alone.

The local implementation now lowers all 26 conversions on both ISAs, bringing
native scalar numeric coverage to 136 operations. Checked conversions keep
explicit bounds and cold trap exits; arm64 saturating conversions use FCVTZS/ZU,
with direct allocated registers for nontrapping casts and reinterpretation.
All 11,970 execution assertions in the ten pinned scalar numeric scripts pass
in both native tiers under Go/LLGo arm64 and isolated Go amd64 execution.
The 263 invalid/malformed frontend commands are explicitly excluded from that
native count. Exact codegen statement coverage is 95.4%; full native Core 2.0
execution and the final pinned two-target matrix remain incomplete.

The four declared conversion chains are still approximately 8–9% slower than
Cranelift in local three-block diagnostics, despite improving from 14.6 to about
10.5 ns/source iteration for the signed f32 case. Each iteration contains several
casts and integer operations; this is not latency per individual conversion.
Raw initial/final records and restored machine-code identity checks are retained
in `artifacts/native-conversions/`. An IEEE-bit range-check experiment had no
stable benefit and was removed. No failed performance gate is waived.

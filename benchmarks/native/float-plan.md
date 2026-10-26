# Scalar floating-point development scope

Status: local implementation work; not accepted and not a second contribution PR.

Implement native f32/f64 comparisons, arithmetic, rounding, min/max, sign operations and bit reinterpretation on both required ISAs. Keep integer/float conversions as explicit runtime exits until their complete trapping and saturating semantics are implemented. Use the existing function compiler, continuation protocol, frame slots and FP environment; do not add another execution framework.

Before submission, compare the exact own baseline, Go/LLGo candidates and pinned wazero, Wasmtime and Wasmer native engines using the same parameterized Wasm binaries. Declare three equally visible families: dependent f32/f64 multiply/add/divide chains, min/max/comparison with signed zeros, and rounding/square-root chains. Preserve the integer regression series. Separate preparation, first call and warm execution. Require parity or better for every affected family on Linux amd64 and macOS arm64 with calibrated A/A controls; no float feature is accepted on integer-only evidence. Numerical inputs must remain finite, include alternate seeds/counts and use independent expected results; NaN/edge behavior belongs to the semantic suite.

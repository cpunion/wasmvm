# Shared AArch64 integer register capacity

Baseline `883990e2ed719c72259747a9ff9439ba0ac6446e`, engine `3a9f291`.
Use AArch64 caller-saved x1-x8 for mutable integer/reference slots when the
callee-saved cache is full. Assign immutable constants and runtime metadata
only to remaining caller-saved registers. Share one target register inventory
with leaf expansion, excluding immutable registers there. Keep context, frame,
operand/vector scratch, memory bases, platform/link registers and the native
ABI unchanged. Every mutable allocation spills on exits and reloads on entry.

Preserve exact scalar/reference bits, vector state, original PCs, fuel,
cancellation, growth, host reentry and code ownership. Test overlapping register
classes, partial/exhausted caller capacity, inline arguments/outputs/pressure,
indirect target mutation, large frames and all cold continuations. Compile both
targets and run complete Go/LLGo release semantics, >=90% measured-package
coverage and the final Linux amd64/macOS arm64 matrix.

Compare exact source base/head and native API/dispatch with identical compiler
pins/flags and the existing twelve families plus pinned native references.
Record preparation, first/warm calls and code/frame traffic separately. Keep
all regressions; unchanged x86 code and interpreter timing are not an inferred
speedup. The current competitor/calibration gates and single active PR remain.

# LLGo main snapshot upgrade: 2026-11-14

The new production pin is
[ac1292206d672c2aeb546780a1d585ad98edbc3d](https://github.com/xgo-dev/llgo/commit/ac1292206d672c2aeb546780a1d585ad98edbc3d),
the main snapshot selected for this upgrade. The previous pin is 56fa9aedc957.
Go 1.27.3, LLVM/Clang 22.1.8, normal GC and `-O3 -lto=full` remain unchanged.
Later main commits require a separate, explicit baseline update.

## Relevant changes and qualification

The snapshot includes scalar interface-wrapper tail calls, native defer-callee
capture, named function assertion identity, native finalizers/callbacks,
HostLayout/libffi padding, horizontal SIMD addition and native external-linker
options. Memory32 compiler-target callback/assertion fixes are distinct from
wasmvm's guest linear-memory32 support; the host benchmark remains linux/amd64
or darwin/arm64.

The new compiler builds and executes the complete native-platform, interpreter,
nativeexec, VM, component and WASI P3 test packages on local macOS arm64. The
contribution's complete Go/LLGo/platform matrix remains required before merging.
The CLI packaging checks and benchmark provenance checks use the same new pin.
No compiler failure is replaced by an expected rejection or a reference engine.

## Compiler comparison protocol

This is a host compiler/runtime baseline event, separate from a VM source change.
Compare the exact frozen 688b40b6 engine and its common instrumentation with both
LLGo pins, changing only compiler identity in the generated provenance file.
That engine source closure matches the retained implementation merged by
[PR #171](https://github.com/cpunion/wasmvm/pull/171). Check source and executable
hashes, compiler revision/dependencies, LLVM version and release flags before
executing either role.

Reuse the frozen Go and independent reference executables. Report source/input
identities for every reused binary. Keep old/new LLGo interpreter and native
execution distinct, together with unchanged Go controls, named competitors and
unchecked native C. Measure integer dispatch, dependent addresses and nonzero
offset mixed-width memory operations with 20 fixed blocks per cohort. Include
identical-binary aliases for both LLGo roles and all selected references.

Preparation, instantiation, first execution and warmed calls retain separate
timers. Validate the original and alternate checksums. Each cohort retains all
observations, failed controls and adverse outcomes; do not pool distinct hosts
or discard a regression because another workload improves. Results and final
disposition are recorded in the contribution's benchmark comment and body.

The [prior regression investigation](performance-operand-regressions-2026-11-14.md)
and LLGo issues [#3094](https://github.com/xgo-dev/llgo/issues/3094),
[#3097](https://github.com/xgo-dev/llgo/issues/3097) and
[#3098](https://github.com/xgo-dev/llgo/issues/3098) remain open. Updating the pin
does not establish recovery or whole-portfolio performance acceptance.

## CI work distribution

The preceding macOS legacy job took 675 seconds: 116 seconds installing LLGo,
176 seconds in its semantic shard and 347 seconds in the P3 measurement-package
validation. Compilation/linking remains separate from test execution in the
saved LLGo traces. The legacy shard now receives the small service adapter and
native service-reference validation; the Threads shard receives the complete
P3 measurement validation. All 32 repository test packages still run exactly
once per LLGo platform, and both additional measurement modules remain covered.
No tests, platforms, profile flags or timeouts are removed to shorten CI.

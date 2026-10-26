# Interpreter-only diagnostic patch

`interpreter-only-v1.12.0.patch` applies only to wazero **v1.12.0**. It is an experiment used to separate host-compiler portability from guest feature coverage. It is not a production fork or a claim of complete LLGo compatibility.

The patch makes three narrowly scoped changes:

1. Remove the unconditional native `wazevo` engine import from `runtime.go`. Explicit compiler-engine requests panic with a diagnostic instead of silently selecting the interpreter. Evaluation callers must select `NewRuntimeConfigInterpreter` explicitly.
2. Replace native-backend CPU detection on amd64 with zero native feature flags.
3. Do the same on arm64, removing the dependency on `x/sys/cpu` runtime hooks from the interpreter path.

On the initial darwin/arm64 probe, unmodified wazero failed LLGo linking on `runtime.memmove` and `runtime.getAuxv`, before any guest ran. The diagnostic isolation allowed the same core arithmetic/trap, host callback, and P1 output probes to execute successfully. The broader evaluator and CI provide separate evidence for their own fixtures; this initial result is not a conformance claim.

`scripts/run.sh <go|llgo> <upstream|interpreter-only> [go test args...]` copies the evaluator and, for patched mode, the pinned dependency into a temporary directory. It never changes the repository module files or module cache. It preserves the compiler's exit status and writes logs, dependency metadata, toolchain information, and Go coverage under `artifacts/wazero/<compiler>-<mode>/`. Set `WASMVM_EVAL_ARTIFACTS` to override that output directory. Test arguments are forwarded as separate arguments; use an absolute path for an additional output file because execution takes place in the temporary copy.

`scripts/record-upstream-llgo.sh` checks only the known, pre-execution linker limitation. It rejects arbitrary failures and unexpected upstream success. Its successful exit means the limitation was reproduced; it does not mean guest tests passed. The independently patched LLGo evaluator must pass normally.

A production adaptation would need proper backend/build-tag boundaries, automatic-backend selection behavior, platform coverage, runtime/GC/ABI auditing, and the relevant upstream tests. Those requirements cannot be replaced by this diagnostic patch.

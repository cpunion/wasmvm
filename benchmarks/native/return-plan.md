# Forwarded return scope

Baseline: `bc92e03854605d87b5f8fae71b0409af6541d11d`. Fold an unconditional
forward edge into a pure return block in the shared slot compiler. Return the
edge's original operands directly, avoiding an intermediate tuple copy and an
extra interpreter dispatch. Keep instruction PCs and source/fuel boundaries
stable; conditional edges, backward edges and effectful blocks are unchanged.
Native leaf expansion must still recognize the resulting early return.

Verify result ordering and reference/vector values, exact fuel before observable
effects and return, host suspension, nested calls and frame-arena relocation.
Run the complete Core 2.0 interpreter/native corpus with both host compilers,
retain at least 90% package coverage, and run the final supported-target matrix.

Compare the exact base/head public API and separate dispatch diagnostics under
Go and LLGo, including the source-call, control, integer, memory, and SIMD
regression families against pinned relevant competitors. Use the existing
phase protocol, three process blocks, seven warm calls, two warmups, alternate
seeds/counts and independent checksums. Report preparation separately and retain
regressions. Local diagnostics do not accept M2 or resolve PR #26's calibration
gate; no new contribution PR or merge is authorized by a diagnostic result.

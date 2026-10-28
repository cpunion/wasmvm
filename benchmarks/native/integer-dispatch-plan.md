# Direct integer semantics scope

Source baseline: `1d4b6d9936ad0727b3b252d5a9ed2bf8b8d4535d`. Replace the
interpreter's second opcode dispatch for nontrapping integer operations with
calls to small, shared semantic functions. The generic numeric evaluator and
the primary instruction switch use those same functions; do not duplicate
arithmetic semantics or add another interpreter. Include both widths of
comparisons, counts, arithmetic, bit operations, shifts/rotations, integer
width conversions and sign extensions. Trapping operations and floating-point
semantics retain their existing evaluator.

Adopt exact LLGo `8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, whose upstream
PR #2936 passes its complete CI but remains unmerged. Record this as a compiler
baseline event. Compare VM source base/head using the same new compiler and
retain the former compiler's measurements separately; never attribute the
bitcast compiler improvement to the interpreter dispatch change.

Keep PCs, source/fuel accounting, cancellation, frame limits and all backend
boundaries. Run the complete Core 2.0 corpus under both compilers/backends and
the final amd64/arm64 matrix, with at least 90% coverage per executable package.
Existing numeric assertion and boundary tests remain the semantic oracle.

Use the existing phase protocol, exact archived base/head sources, pinned
competitors, alternate inputs and three interleaved process blocks with seven
warm calls and two warmups. Include integer, direct/indirect calls, memory,
SIMD and floating-point regression families, keeping API/dispatch and all
preparation/first-call phases separate. Inspect emitted dispatch for helper
inlining and retain code-size changes, failures and regressions. Diagnostics
do not waive the full performance gates or PR #26's ARM calibration failure.

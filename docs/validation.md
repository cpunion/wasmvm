# Verification Record

Date: 2026-10-26. The initial sections record the design scaffold's host compiler checks. The M0 section records subsequent candidate execution. No production VM API or full Wasm/WASI conformance is claimed.

## Local results

Working directory: `/Users/lijie/source/goplus/wasmvm`. Platform: darwin/arm64.

| Check | Result |
| --- | --- |
| `./scripts/check.sh go` | Passed; Go 1.27.3, `CGO_ENABLED=0` |
| `./scripts/check.sh llgo` | Passed; Homebrew LLGo v1.0.8 |
| `go test -race -count=1 -timeout=10m ./...` | Passed |
| `go vet ./...` | Passed |
| `actionlint .github/workflows/ci.yml` | Passed |
| `gofmt` | Formatted, with no remaining differences |

Both compilers ran the same five tests in `internal/compat/compat_test.go`: numeric/NaN/negative-zero bit patterns and little-endian encoding; typed reference cycles across host GC; goroutine synchronization with atomics/mutexes; condition-variable wakeup; and context cancellation propagation. The public packages contain only `doc.go`; `[no test files]` is not counted as feature acceptance.

LLGo resolved to the `/opt/homebrew/bin/llgo` wrapper, which executed `/opt/homebrew/Cellar/llgo/1.0.8/libexec/bin/llgo`. `LLGO_ROOT=/opt/homebrew/Cellar/llgo/1.0.8/libexec`; GOROOT was the Homebrew Go 1.27.3 installation. The devel binary/source combination observed earlier in the llgo checkout was different from this project test environment; see [llgo-ci.md](design/llgo-ci.md).

## Remote CI

**Executed and passed:** [GitHub Actions run 40240658954](https://github.com/cpunion/wasmvm/actions/runs/40240658954), testing commit `4968459404a8fc67d6bf6eb27502aeb2a1dd56ea`.

| Job | Result | Duration |
| --- | --- | --- |
| `go / linux-amd64` | Success; shared tests, gofmt/vet, and race checks all passed | 38 seconds |
| `llgo / linux-amd64` | Success; pinned revision installation/check and shared tests passed | 2 minutes 37 seconds |

Both jobs ran the same `test -count=1 -timeout=10m ./...`, without `continue-on-error` or compiler-specific test skips. The current probes have therefore been executed with both Go and LLGo on local darwin/arm64 and CI linux/amd64.

The initial follow-up documentation commit recorded these results and links using `[skip ci]`; it changed no code, tests, or workflow and did not reinstall the toolchains. The results above refer to the listed tested commit. Subsequent documentation translation and development planning do not add VM capabilities or constitute new runtime verification.

## Initial scaffold limitations

- No production VM instruction, WASI ABI, WasmGC, guest-thread, or Component Model implementation has been written in `pkg`.
- No full official Wasm/WASI conformance suite has been run. Subsequent bounded candidate probes and performance measurements are described below.
- The initial scaffold did not validate an LLGo guest; M0 now includes a separate reference-engine run.
- `go 1.24.0` is a conservative source language declaration; the minimum version has not been tested with the Go 1.24 toolchain.
- Local success does not establish compatibility with every OS/architecture, and host race checks do not prove the Wasm memory model.

## Initial documentation translation checks

The English translation and expanded development plan were checked across all nine Markdown files. The checks found no remaining Chinese characters, missing relative file links, unmatched fenced code blocks, or whitespace errors. All original external source URLs were preserved. A consistency review confirmed the M0–M6 sequence, M2a/M2b parallelism, and the distinction between existing scaffolding and planned VM capabilities.

This update changes documentation only. Runtime tests were not rerun because source code, tests, and the workflow are unchanged; the execution evidence above remains tied to its original commit. The documentation commit uses `[skip ci]` to avoid repeating toolchain installation for text changes.

## M0 candidate evaluation

- The standalone wazero evaluator executes seven capability fixtures and records six exact unsupported cases. Shared-memory concurrency/wait/notify/growth, cancellation observations, and invalid-input/result paths are tested separately.
- The local Go race run passes with **100% statement coverage of the owned evaluator**. This is not engine, compiler, benchmark-runner, or specification coverage.
- Released LLGo v1.0.8 fails before execution on missing runtime symbols. The runtime fix in [LLGo PR #2919](https://github.com/xgo-dev/llgo/pull/2919), commit `bc7c991d8fe91459c8afc1eb7d0e374124d95374`, runs the unchanged evaluator. The temporary engine-isolation patch is historical diagnostic evidence only.
- [Candidate results](../experiments/wazero/results/darwin-arm64.json), the [Wago probe](../experiments/wago/README.md), and the [generated LLGo guest/reference run](../experiments/llgo-guest/README.md) retain their exact scopes and limitations.
- The [benchmark report](research/performance.md) records actual executions across four runtime projects, including raw samples, fixture hashes, backend choices, and external host contention. Python harness tests validate statistics, fixture integrity, and output/exit assertions; timing is not a CI pass/fail performance threshold.
- The expanded CI requires both compilers to execute unchanged wazero, enforces an owned-evaluator coverage floor of 90%, and uploads raw evaluation logs. M0 acceptance requires successful final-PR checks and review before merge.

The [M0 CI run at `bbbb375`](https://github.com/cpunion/wasmvm/actions/runs/40251182967) passed both Linux amd64 jobs. Downloaded artifacts confirm the shared fixture matrix under fixed LLGo, the recorded atomic-wait cancellation limitation, and **100.0% owned-evaluator coverage** under Go. Final acceptance is tracked by [PR #1](https://github.com/cpunion/wasmvm/pull/1), including its final-head checks and review findings.

The original next-work recommendation followed the [M0 reuse decision](design/reuse-decision.md). It is superseded by the [own-core decision](design/own-core.md) and revised [roadmap](roadmap.md); the experiments do not accept any production implementation milestone.

## Own-core design revision

The implementation route now follows [Project-Owned Core](design/own-core.md). The revision defines complete Core 2.0 interpreter/native stages, Core 3.0 plus Threads on both backends, the native ABI, package boundaries and provisional performance gates. It changes documentation and development instructions only; no production engine, native bridge, conformance runner or performance enforcement is implemented by this revision.

Local checks on darwin/arm64 passed: `./scripts/check.sh go` (Go 1.27.3), `./scripts/check.sh llgo` (Homebrew LLGo v1.0.8), `go vet ./...`, and `go test -race -count=1 ./...`. These execute the existing host compatibility probes. Changed Markdown files passed relative-link, fenced-block, English-text and whitespace checks. The Core 2.0 `wg-2.0` tag and directory-specific license were checked at `fffc6e12fa454e475455a7b58d3b5dc343980c10`; no corpus execution is implied.

Final contribution-PR checks are required before merge. Existing benchmark CI remains a candidate-engine comparison and cannot establish the new own-core budgets. The source, tests and workflows are unchanged, so this revision adds no coverage percentage or new runtime acceptance claim.

## M1 binary-reader increment

The first owned executable core package, `internal/binary`, provides bounded
borrowed reads and nested subreaders, absolute error offsets, u32/u64/s32/s33/s64
LEB128, exact little-endian floating-point bits, and byte-limited UTF-8 names.
Lengths are checked before conversion to a host index or string allocation.
Module decoding, validation, aggregate allocation accounting, execution and
the complete pinned Core 2.0 corpus remain unfinished; M1 stays open.

Local darwin/arm64 validation uses Go 1.27.3 and the LLGo compiler implementation
from `d32073c65e65b14b51a51e017e6d2023f1da60ce` with the runtime at
`8b1389c627c77a02e56af03f08d9cdc19edffe77` from
[LLGo PR #2919](https://github.com/xgo-dev/llgo/pull/2919). The intervening LLGo
commits change runtime hooks and tests, not the compiler implementation. Both
compilers execute the same reader tests, including fuzz seeds. Boundary tests
cover nested limits, maximum u64 length claims, every possible terminal LEB
byte, signed extrema, legal padding, malformed encodings, invalid Unicode,
embedded NUL names, NaN payloads and signed zero.

The owned binary reader has **100% Go statement coverage (74/74 statements)**.
Coverage is measured for executable packages under `internal` and `pkg`, not
ignored local diagnostic artifacts or foreign engines. The new coverage
checker enforces 90% independently per package and rejects empty, corrupt or
missing required package data. Five checker tests include an aggregate-99%
case whose uncovered package must still fail. A ten-second Go fuzz run with
two workers completed 1,916,181 executions without a failure. This is a bounded
local observation, not exhaustive validation.

CI now declares Go and LLGo on Linux amd64 and macOS arm64, pinned to the LLGo
revision above, with separate coverage artifacts and Go fuzz runs. The actual
final-head jobs must pass before this increment merges; configuring the matrix
alone is not acceptance. Existing benchmark jobs still measure reference
engines, and no own-engine performance gate is closed by this reader increment.

The reader increment merged in [PR #5](https://github.com/cpunion/wasmvm/pull/5)
after all six checks passed for exact head
`44eacb2de51033ffc929a0ab0ca454319c2b1698`. The
[Go/LLGo run](https://github.com/cpunion/wasmvm/actions/runs/40266256500) passed
all four Linux amd64/macOS arm64 jobs; logs confirm Go 1.27.3 darwin/arm64 and
100% reader coverage on the macOS runner. The
[benchmark run and PR report](https://github.com/cpunion/wasmvm/actions/runs/40266256532)
also passed, with their existing reference-engine scope. The merge commit is
`7c68127a024743552480d7a6efcf033d2aa81f43`.

## M1 complete corpus input preparation

The [Core 2.0 inventory](../testdata/core2/README.md) includes all 148 WAST
scripts under the pinned `test/core`, recursively including SIMD. WABT 1.0.41
converts them with validation disabled so invalid binaries remain available
for negative VM tests. The manifest records every source hash, script identity
and command classification; a second record checksums all 5,822 generated
files, including both upstream notices. Production VM code is not copied.

The local pinned darwin/arm64 producer generated and verified 54,006 commands:
51,139 binary assertions, 1,091 malformed-text assertions, and 1,776 module,
registration and action commands. These counts describe inputs only. The
binary VM does not implement a WAT parser, and no text-parser or binary guest
assertion is counted as executed by this preparation step.

Eleven corpus-tool tests cover source selection including SIMD, archive path and
link handling, download size/digest/redirect failures, cached corruption,
script classifications, missing/extra/modified files, and changed pins/counts/notices.
They run alongside the five per-package coverage-gate tests. CI regenerates
and checks the same fixture bytes on both supported platforms and retains the
producer executable hash and input/output identities. Final-head CI remains
required before merging this increment; M1's complete execution gates stay open.

## M1 complete Core 2.0 binary decoding

The owned frontend implements all Core 2.0 sections, segment variants, scalar
and reference immediates, and all 236 fixed SIMD opcodes. It preserves raw float
bits and absolute source offsets. An explicit syntax stack bounds control
nesting without recursive guest parsing. The decoder snapshots input and bounds
bytes, aggregate decoded nodes, names, nesting and expanded local counts before
allocation/use; decoded modules are unvalidated and cannot execute.

The complete 148-script corpus checks 4,581 binary modules: 719 malformed
binaries are rejected, and the remaining 3,862 decode successfully, including
semantic-invalid and linking/trapping inputs that later stages must reject or
execute appropriately. No semantic-invalid, linking, trap, or return assertion
is counted as passed by decoding. Two pinned WABT data-count corrections preserve
the original semantic-invalid cases; see the [input record](../testdata/core2/README.md).

Local Go and LLGo tests pass. Go statement coverage is 99.6% for
`internal/core/decode` and 100% for `pkg/wasm`; `internal/core/module` contains only
type declarations. The decoder's defensive rethrow of unexpected implementation
panics is not intentionally triggered; a future such panic must fail fuzz/tests,
not be disguised as malformed guest input. A ten-second Go fuzz run with two
workers executed 682,250 inputs without failure. These are bounded observations,
not a semantic conformance or security-audit claim.

Seventeen Python tool tests cover corpus preparation/corrections and coverage
gates. CI now requires the decoder/public-type coverage packages and fuzzes
module decoding in addition to the binary primitives. The final head must still
pass Go/LLGo on Linux amd64 and macOS arm64 before merge. M1 remains open for
semantic validation, IR/lowering, complete interpreter/runtime lifecycle, public
embedding API, CLI and actual own-engine benchmarks.

## M1 complete Core 2.0 static validation

The decoder increment merged in [PR #7](https://github.com/cpunion/wasmvm/pull/7)
after all six checks passed for `f5e1833ec1fd0b8198333f6138450e6e0bccdc6e`,
including Go/LLGo on both supported platforms and the reference-engine
benchmark/report workflow. The merge commit is
`8451a83` (full identity is available in the PR).

`internal/core/validate` checks index spaces, exports, limits, constant
initializers, declared function references, block parameters/results, operand
typing, unreachable-stack polymorphism, every scalar/reference/SIMD operation,
lane/alignment bounds and bulk memory/table rules. Validation creates no runtime
state and executes no guest code. Decoded module ownership transfers to the
validated-module record; subsequent compiler/runtime consumers must treat all
module/type slices as immutable. Locals (including parameters), operand-stack
height and control depth have explicit independent quotas.

The full pinned input inventory contains 719 malformed binaries, 2,146
semantic-invalid binaries and 1,716 statically valid binaries. All classifications
pass locally. Valid function bodies include all 236 fixed SIMD opcodes; an
initial unary/binary classification error for `i8x16.avgr_u` was caught by the
corpus and now has an explicit arity regression test. Branch-table unknown
operands, implicit else types, exact quota boundaries and parameter/local index
overflow have additional regressions. Linking, instantiation, traps and return
assertions are not counted as executed by static validation.

Go statement coverage for `internal/core/validate` is 96.9%. Remaining paths
mostly defend against manually corrupted internal module records that the
decoder cannot produce (invalid value/reference types and segment kinds); the
unexpected-panic rethrow remains deliberately visible to fuzz/testing. Local Go race/vet checks and the shared Go/LLGo suite pass. A 15-second Go
fuzz run with two workers completed 1,604,465 inputs without failure; this is a
bounded observation, not a proof. Final-head cross-platform CI remains required.
M1 is still open for IR, interpreter, runtime/lifecycle, API, CLI and own-engine
performance measurements.


## M1 typed control-flow IR

The validator merged in [PR #8](https://github.com/cpunion/wasmvm/pull/8) after
all six checks passed for `742a93f6aff980c5e15016d426713a0e4fe34a20`.
The [Go/LLGo matrix](https://github.com/cpunion/wasmvm/actions/runs/40273077218)
passed on Linux amd64 and macOS arm64; the
[reference benchmark/report run](https://github.com/cpunion/wasmvm/actions/runs/40273077222)
also passed.

Typed lowering produces and verifies 6,236 functions in all 1,716 statically
valid corpus modules. Tests exercise multi-value loop edges, dead trapping
instructions, shared compilation budgets and invalid IR: wrong types/arity,
missing definitions/terminators, use before definition, non-dominating uses,
unreachable blocks and corrupted effect metadata. Visitor tests check operand
ordering, multi-result calls, early errors and unreachable polymorphic types.
The shared validator walk supplies opcode signatures to the compiler.

Local Go and LLGo tests, Go race/vet checks, and Python tooling tests pass.
Go statement coverage is 99.5% for `internal/compiler/ir` and 98.9% for
`internal/compiler/lower`; the intentional unexpected-panic rethrows remain
uncovered. A 15-second Go lowering fuzz run with two workers completed 548,666
inputs without failure. These checks validate lowering structure, not executed
Wasm results. CI requires both packages at 90% coverage and adds lowering fuzz
checks. Final-head cross-platform CI is still required before merge. M1 remains
open for slot lowering, complete execution, lifecycle, public API/CLI and
own-engine performance measurements.


## M1 typed slots and branch transfers

`internal/compiler/slots` lowers all 6,236 functions in the 1,716 statically
valid Core 2.0 corpus modules. A symbolic checker reads every allocated operand
and simulates every selected edge's sequential moves against simultaneous SSA
assignment. This includes local reuse, calls/results, references, vector values,
branches, loops and return tuples without substituting a reference engine.

The corpus's 28,202 SSA values use 18,454 temporary slots, including edge scratch
and excluding declared locals. This is a storage-layout observation, not a
runtime speed or memory-usage benchmark. Ten thousand deterministic randomized
parallel-copy graphs cover overlap, fan-out, identity moves and cycles. Explicit
typed cyclic-loop tests verify vector/reference scratch and scratch reuse.

Local Go/LLGo, Go race/vet and the per-package coverage gate pass. The slot
package has 99.4% statement coverage; its defensive unexpected-panic rethrow is
not deliberately exercised. A 15-second Go fuzz campaign with two workers
completed 730,957 inputs without failure. CI requires this package's 90% floor
and adds slot-lowering fuzzing. Final-head cross-platform CI is required before
merge. Guest execution assertions, runtime lifecycle, public API/CLI and
own-engine measurements remain open, so M1 is not complete.


## M1 complete scalar numeric primitives

The IR and slot increments merged in [PR #9](https://github.com/cpunion/wasmvm/pull/9)
and [PR #10](https://github.com/cpunion/wasmvm/pull/10) after all six checks passed
for exact heads `be277a95afe4d5df9fd1deb38822791e12813cff` and
`7e8c30600329f6fac1354ff97210e1b2b5d4ead9`, respectively.

`internal/engine/numeric` implements every Core 2.0 scalar numeric instruction:
128 scalar opcodes plus eight saturating conversions. Inputs/results retain
exact bit representations; divide-by-zero, integer overflow and invalid
conversion return explicit trap codes. Conversion ranges are checked after
truncation and before host conversion. Arithmetic float results normalize NaNs;
bit-preserving `abs`, `neg`, `copysign` and reinterpret operations do not.

Eight pinned scripts (`i32`, `i64`, `f32`, `f64`, `f32_cmp`, `f64_cmp`,
`float_misc`, `conversions`) provide 11,534 expected-return and 87 trap
assertions. The primitive test adapter requires each invoked export to consist
of ordered parameter reads, exactly one numeric operation and `end`; changed
shapes, commands, opcode inventories or counts fail. All 136 operations are
exercised. Another 171 invalid modules are rejected by the frontend; eight
malformed-text assertions remain explicitly outside binary execution scope.
This is direct numeric primitive testing, not execution of complete modules or
acceptance of the full Core 2.0 profile.

The tests found two host-math differences that the Wasm layer now handles:
Go `math.Min/Max` prioritize extremal infinity over NaN, while Wasm requires
NaN propagation; pure-Go and intrinsic math implementations can differ in NaN
payload and signaling behavior. The Wasm implementation returns permitted
canonical quiet NaNs for arithmetic on both compilers, retaining exact NaN
payloads where bit operations require them. See the
[official numeric rules](https://webassembly.github.io/spec/core/exec/numerics.html).
These differences do not establish a violation of Go's NaN API contract.

Local shared Go/LLGo tests and Go race/vet pass. Numeric primitives and public
trap codes have 100% statement coverage. A 15-second two-worker numeric fuzz run
completed 1,199,022 operand pairs across every scalar opcode without host panics.
CI requires numeric package coverage and fuzz checks. Final-head CI remains
required; complete SIMD, memory/table state, control/call execution, lifecycle,
API/CLI and own-engine measurements remain open within M1.

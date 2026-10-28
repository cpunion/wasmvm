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


## M1 nonmemory fixed-SIMD primitives

Scalar numeric primitives merged in [PR #11](https://github.com/cpunion/wasmvm/pull/11)
after all six checks passed for exact head
`0ab925fe2cc0a552915faed0a6a660886c762a4b`, including both compilers on both
supported platforms and the existing reference benchmark/report job.

`internal/engine/simd` implements all 214 nonmemory fixed-SIMD opcodes, including
lane operations, shuffle/swizzle, comparisons, integer wrap/saturation, Q15
rounding, dot products, widening/narrowing, pairwise addition, extended
multiplication, float arithmetic/rounding/pseudo-min/max and conversions.
`internal/engine/value` holds exact 128-bit storage; it does not implement
Store reference ownership. The other 22 SIMD instructions access memory and
remain part of the subsequent checked memory implementation.

The primitive adapter explicitly selects the first pure-expression module in
40 named official scripts and executes all its 22,924 return assertions. It
requires state-free modules, pure expressions, fixed counts and all 214
nonmemory opcodes; an unexpected function shape or command fails. Later modules
exercise globals, structured control and memory and remain required for the
complete execution suite. They are not counted as executed by this adapter.
Malformed/invalid classification is already checked by the full frontend test.

Shared Go/LLGo tests pass locally. The SIMD package has 97.3% statement coverage;
uncovered branches defensively reject opcode groups already excluded by the
entry inventory guard. A 15-second two-worker fuzz campaign completed 455,375
operand tuples across all 214 operations without failure. Go vet/race and
per-package coverage checks pass. CI adds the SIMD coverage floor and fuzz
campaign; final-head CI is required before merge. M1 remains open for checked
memory, full interpreter/runtime lifecycle, API/CLI and own-engine benchmarks.

## M1 private memory primitives

Nonmemory SIMD merged in [PR #12](https://github.com/cpunion/wasmvm/pull/12)
after all six checks passed for exact head
`189e7c21c646def087aceceb9e2b15717f3b56f0`.

`internal/engine/memory` implements all 14 scalar loads, nine scalar stores,
17 SIMD loads and five SIMD stores, plus page growth, copied host access and
bulk fill/copy/init. Tests independently construct byte-level expected results
for every access width, signed extension, vector widening/splat/zero, every
lane and unaligned/end-of-memory address. Range checks widen address+offset,
reject overflow-sized inputs, and precede mutation. Bulk tests check both
copy overlap directions, exact endpoints and zero-length operations.

Go and LLGo tests pass locally, as do Go race/vet and tooling checks. Memory
coverage is 99.3%; the sole uncovered statement clamps maximum pages on a
32-bit host, outside the current Linux amd64/macOS arm64 target matrix. A
15-second two-worker fuzz campaign executed 1,606,911 range/operand tuples
without failure. Growth tests preserve old bytes, check new zero pages and
exercise declaration/host quotas without allocating an entire 4-GiB memory.
Logical-page quotas do not promise a hard process-RSS ceiling during copying
or deferred host collection.

These are memory primitive tests. Full-module instantiation/segment semantics,
Store-wide accounting, chunked cancellation, tables/references, explicit call
frames, API/CLI and own-engine benchmarks remain required before M1 acceptance.
CI adds memory coverage and fuzz checks; final-head results are required before
this increment merges.


## M1 complete Core 2.0 interpreter integration

Private memory merged in [PR #13](https://github.com/cpunion/wasmvm/pull/13)
after all six checks passed for exact head
`9863399d75b6facd2927ec915640d422902f8923`.

The owned interpreter and runtime state now pass all **51,139 binary assertions
in all 148 pinned scripts** locally under Go and LLGo. Counts are unchanged:
45,734 return assertions, 2,408 traps, 15 stack-exhaustion assertions, 2,146
invalid binaries, 719 malformed binaries, 83 unlinkable modules and 34
uninstantiable modules. The runner also performs all 1,599 ordinary module,
155 action and 22 registration commands. Exactly 1,091 malformed-text parser
assertions remain outside the binary VM's scope; none of the binary assertions
is skipped. Compiler lowering and actual execution are distinct recorded steps.

The first integrated run exposed two adapter errors: spectest float globals
must contain 666.6 rather than 666, and logical stack exhaustion includes both
frame-count and frame-slot limits. Correcting the adapter produced the complete
passing inventory. The large-frame recursion script remains enabled and hits
the slot quota without using recursive Go calls.

Additional tests cover all allocation quotas/rollback, import ownership,
canonical type limits, table overlap/growth, passive/declarative drops, and
initialization traps that preserve earlier imported-memory/table effects and
escaped function identity. Interpreter tests cover fuel before side effects,
bounded polling and partial bulk cancellation, root visitors, host suspension,
resume errors, nested continuations, explicit abort/panic cleanup and reuse.
A warmed scalar call performs zero allocations in the Go counter test; this
is not a throughput benchmark or a performance-gate pass.

Local Go/LLGo tests and Go vet/race pass. Statement coverage is **98.6%** for
`internal/engine/execute`, **96.0%** for `internal/engine/state`, and **100%** for
public value/error/trap contracts. Remaining paths defend against corrupted
internal instructions/references and 32-bit host capacity constraints. A
15-second two-worker execution fuzz run completed 1,049,762 inputs without
failure, with bounded frontend, instance, frame, slot and execution budgets.

CI now writes an actual own-core conformance artifact containing the compiler,
platform, pins checksum, backend, command classifications and passing assertion
counts. The new core package coverage gates and execution fuzz run are required.
Final-head remote checks still must pass before merge. M1 remains open for
public lifecycle/ownership, host roots and close/reclamation (including failed
instances without escaped references), API/CLI and own-engine benchmarks. The
existing benchmark PR report still measures reference engines at this point.


## M1 failed-instance reclamation and internal close

The complete interpreter merged in [PR #14](https://github.com/cpunion/wasmvm/pull/14)
after all six checks passed for `223d46a6ffbfe29eba3220f418a4034757908865`.
All four [CI execution artifacts](https://github.com/cpunion/wasmvm/actions/runs/40279556265)
report the same 51,139 passing binary assertions, 148 scripts and pinned-input
checksum under Go/LLGo on Linux amd64/macOS arm64.

The internal Store now traces successful instances, host tables/globals,
explicit function roots and active frame owners. Failed instances are reclaimed
only when unreachable, including unreachable cycles. Imported memory/table
effects survive failure; escaped functions and imported objects preserve their
origin instance. Tests cover removal of each root kind, 100 repeated failures
under one-instance/one-function quotas, bounded identity holes, suspended calls,
rollback of object registries, empty-object quotas and idempotent final close.
Reclamation clears backing storage and restores live resource accounting.
Interned type signatures retain a bounded Store-lifetime cache.

Local shared Go/LLGo tests and Go vet/race pass with the complete corpus still
enabled. Coverage is 97.6% for state, 98.7% for the interpreter and 99.3% for
memory; defensive invalid-internal-record and 32-bit host paths remain uncovered.
Every executable core package exceeds 90%. Final-head CI is required before
merge. Internal final close assumes no execution lease remains; public
nonblocking close, lease synchronization, host reference handles, API/CLI and
own-engine benchmarks are still required M1 work.


## M1 public embedding API and host references

Failed-instance reclamation merged in [PR #15](https://github.com/cpunion/wasmvm/pull/15)
after all six checks passed for `579c37a1ce30597765f92d342f70df5a3c1b2737`.

`pkg/vm` now exposes Engine, CompiledModule, Store, Linker, Instance, Extern and
Execution, typed numeric/reference Values, explicit host functions and copied
memory access. The [embedding guide](embedding.md) specifies ownership, root
release/collection, shared reentry budgets, nonblocking close and callback
limits. Two executable external-package examples compile/call a module and
link/call a guest import without a competitor runtime.

Tests cover concurrent compilation, immutable metadata, closed compiled-module
leases, cross-Store and released references, root/object quotas, caller-scoped
result handoff, suspended-local roots during nested collection, escaped
functions after a failed start, repeated Execution use, combined fuel/frame/
slot limits, concurrent Caller rejection, host errors/panics and deferred close
while a callback holds a lease. Start panic cleanup restores instance/memory
usage. Public calls preserve context and host errors separately from traps.

Local Go/LLGo shared tests and the complete Core 2.0 corpus pass, as do Go
vet/race checks. Public VM statement coverage is 97.4%; remaining paths include
cancellation between bounded compilation phases and defensive stale internal
records. Every executable core package remains above 90%, and CI now requires
`pkg/vm` explicitly. Final-head remote checks are required before merge.
M1 still requires the core CLI, own-engine phase benchmarks and final acceptance.
The public call driver has not passed a controlled throughput/allocation gate.


## M1 core CLI and static inspection

The embedding API merged in [PR #16](https://github.com/cpunion/wasmvm/pull/16)
after all six checks passed for `c40acf552203110f89bddbeec6661675ae6ac266`,
including both compilers on Linux amd64 and macOS arm64.

`cmd/wasmvm` now implements `validate`, `inspect` and typed core `run` through
`pkg/vm`. Validation/inspection do not lower, link or execute modules. Copied
metadata includes function signatures and memory/table/global types. Run has
explicit backend/profile reporting, checked scalar/vector/null-reference
arguments, exact result bits, context cancellation and one instruction budget
covering both start and the selected export. No implicit host/WASI imports are
installed. See the [CLI guide](cli.md).

Local Go/LLGo suites, the complete binary corpus and Go vet/race checks pass.
Both actual CLI binaries execute validate/inspect/run against the pinned
call-control fixture; unsupported native selection fails explicitly. CLI
coverage is 99.5% and public VM coverage is 96.8%; the uncovered CLI branch is
the defensive Store-creation error after constructing an unclosed local Engine.
Every executable core package remains above 90%. CI builds and executes the
CLI with both compilers on both platforms and adds its coverage requirement.
Final-head CI is required before merge. Own-engine phase benchmarks and final
M1 acceptance remain open; command wall time is not an engine throughput result.


## M1 owned-engine measurement integration

The CLI's four Go/LLGo platform jobs and measurement job passed for
`b23f512e2629d01b81d323fcc60e9221cff23282`. GitHub initially refused to allocate
a runner for its PR-comment job because of an account billing limit. At the
owner's request the repository was made public; rerunning the failed job then
passed. [PR #17](https://github.com/cpunion/wasmvm/pull/17) merged only after all
six checks were successful.

The phase harness now builds the actual base/head engine sources with both
host compilers. Public API results are primary; separate internal dispatch
rows retain cancellation/bounds/trap checks while exposing reusable interpreter
execution. Internal preparation records decode, validate, lower and slot
allocation beneath an outer interval. Every first/warmup/measured call checks
its checksum. Separately built production CLI binaries retain complete process
latency measurements with checked typed JSON, rather than inferring that latency
from the embedding executable. Owned fixtures retain their independent oracle and alternate-seed
vectors. Engine commits, source digests, common adapter hashes, compiler and
binary identities are validated before a complete report is accepted.

The local measurement package covers both boundaries, all oracle vectors,
mutable-state/checksum failures, traps, forbidden imports/start, bad signatures,
invalid binaries and ready/ack protocol errors. Its statement coverage is 95.3%;
remaining branches defend against impossible current internal states or
allocation errors under fixed valid benchmark configuration. Python tests reject
missing process/build matrices, wrong engine/compiler identities, contaminated
source archives, bad phase accounting and unavailable-baseline substitution.

Final actual base/head measurements, cross-platform CI and PR publication are
still required before this increment and M1 can be accepted. Hosted timings
are advisory; they cannot pass M2's controlled optimization/native gates.


The first complete owned-engine hosted run for `982bc81` passed 270 phase
processes and 72 production CLI processes on an EPYC 9V45 runner
([raw artifact](https://github.com/cpunion/wasmvm/actions/runs/40289815252)).
The subsequent documentation-only head `e30d995` hit the 60-second process
limit on an EPYC 7763 runner after two completed observations
([failure artifact](https://github.com/cpunion/wasmvm/actions/runs/40293132561)).
The workflow now allows 180 seconds per measured process and 90 minutes per
job. Workload sizes, rounds, warmups, measured calls and checksum checks are
unchanged. Raw reports record the deadline and the active failed phase case.
This extends the initial-baseline execution budget; it does not pass an M2
performance gate or erase the failed run. Final-head checks remain required.


At the owner's request to keep CI efficient, PR feedback now excludes only the
two duplicate sustained integer/memory cases; all workload families and all
selected runtime/compiler configurations remain. Per-case loop counts and
sampling/checksum rules are unchanged. Explicit full mode retains sustained
cases and the legacy CLI reference series. The completed full run above remains
historical evidence. Default PR reports state their narrower scope; no M2 hard
performance budget is relaxed. Go/LLGo full semantic execution remains on both
platforms, while pure-Go race/coverage/fuzz/tooling checks run once on Linux.
Merge pushes no longer repeat already-reviewed workflows. Active contributions
are processed sequentially, with subsequent work retained on branches until
the preceding PR passes and merges.


## M1 metadata and final-close resource audit

Metadata snapshots now copy each referenced function signature once. Repeated
imports/functions/exports cannot multiply a large type vector's storage, and
snapshots remain detached from engine code and other snapshots. CLI inspection
preflights its repeated JSON signature values against a 1,048,576-value budget
before allocating expanded strings; validation still accepts these valid
modules. Inspection checks cancellation again before publishing metadata.
Final Store close also drops reference-map backing storage after invalidating
roots and draining the active operation.

Regression tests exercise thousands of shared signatures, snapshot isolation,
import/export JSON amplification, the exact expansion boundary and result-type
accounting, cancellation and closed-root cleanup. Full local Go/LLGo suites,
the complete Core 2.0 binary corpus, Go vet/race and package coverage checks
pass. CLI coverage is 99.5% and public VM coverage is 96.8%; all executable core
packages exceed 90%. Final-head CI remains required before merge. M1 remains
open for representative owned-engine performance evidence and final acceptance.

## Native execution boundary

The owned C adapter and pure-Go amd64/arm64 entry-frame emitter implement code
publication, off-heap frame storage, fixed ABI records, typed exits, atomic
cancellation, FP environment preservation, and code leases. Local darwin/arm64
Go 1.27.3 and pinned LLGo executions pass arithmetic/resumption, register
clobber, concurrent close/publication and forced host-GC tests. Go race and
`GOEXPERIMENT=cgocheck2` checks pass. The existing complete Core 2.0 binary
corpus remains green under both compilers. Native entry probes are separate
from Wasm conformance: a full native Wasm backend is not implemented yet.

Go coverage is 100% for `internal/compiler/nativeabi` and above 95% for
`internal/platform/native`; native allocation/protection failures are not
forced by exhausting the host. C and emitted machine instructions are checked
by execution probes, not included in Go statement coverage.

CI executes the boundary on both target architectures and host compilers.
Linux Go semantic execution also collects coverage in the same run. Historical
M0 candidate evaluations run on manual dispatch; PR benchmarks retain owned
base/head and competitors. Extra benchmark expansion branches remain deferred
so functional implementation proceeds without additional peripheral gates.

## Native integer instruction selection

`internal/compiler/codegen` emits all 66 Core 2.0 integer operations for amd64
and arm64: comparisons, arithmetic, bit counts, shifts/rotates, wrapping and
sign extension. Divide-by-zero and signed division overflow exit with Wasm trap
codes before a hardware fault; signed remainder of the minimum value by -1
returns zero. amd64 uses baseline BSR/BSF and SWAR population count without
requiring BMI or POPCNT. Branch relocations check alignment and signed reach.

The same 505 edge/random input pairs execute for every instruction under both
host compilers (33,330 calls per target/compiler), compared with the validated
owned scalar semantics. Local darwin/arm64 Go and LLGo pass; Go code-generator
coverage is 99.6%. ABI fields now have one pure-Go definition shared by the
emitter and adapter and are still checked against the C layout. These entries
validate instruction selection; whole-function control flow, memory, calls,
metering and native profile conformance remain subsequent compiler work.

## Native function control flow

The native function compiler now emits locals, constants, select, integer
operations, branches/loops, relative jump tables, multi-value copies, fuel and
bounded polling. Runtime-helper, call and large tuple-copy exits retain typed
frame slots and checked continuation IDs; their runtime driver is subsequent
work. Native selection in the public API remains unsupported at this increment.

Both local host compilers compile all 6,236 functions in 1,716 valid pinned
Core 2.0 modules. On arm64 this produces 8,779,052 code bytes and 7,126 explicit
helper sites. This is compilation coverage, not native conformance execution.
Actual function tests cover loops, br_table, tuple cycles, vectors/references,
fuel/trap agreement with the interpreter, helper resumption, frame bounds,
large offsets, long branches and cancellation during generated loops. Codegen
coverage is 96.2%; Go race/cgocheck2 and shared Go/LLGo tests pass locally.

## Complete Core 2.0 native runtime integration

This increment supersedes the earlier public-backend limitation. Both the
interpreter and native backend execute the pinned Core 2.0 corpus through the
same explicit frame machine. Each native tier (unmetered and precise fuel)
passes 51,139 binary assertions across 148 scripts; 1,091 text-parser assertions
remain separately classified. Native execution compiles and publishes every
function, including functions reached through imports, start and indirect calls.
Fixed SIMD, globals/tables, bulk operations and growth use explicit runtime
boundaries. Numeric, scalar-memory and control operations execute generated
machine code. These helper paths do not establish competitive performance.

Shared embedding tests force both backends for host reentry, common fuel/frame
quotas, reference roots, cancellation, panic cleanup, failed instantiation,
escaped references, imported memory and code lifetime after owner close. Native
frame growth, large edge copies, artifact mismatch and retired-buffer access
have dedicated tests. One off-heap memory owner is shared by generated code and
the runtime; entry/exit does not copy the complete guest memory.

Local Go and pinned LLGo execute the complete native corpus on darwin/arm64.
Go also passes it on linux/amd64 under Docker/Rosetta; this is semantic evidence,
not native amd64 timing or an LLGo amd64 CI result. Go race, strict cgo pointer
checks, vet and a cgo-disabled build pass. CI retains the full semantic corpus
for all four compiler/target jobs, while race and strict-pointer jobs focus on
ownership and concurrency without repeating the entire pure semantic corpus.

Performance diagnostics cover the existing six embedded workloads plus indirect
calls, SIMD, overlapping bulk memory and globals/tables. Public API and reusable
dispatch remain separate; startup, compilation, instantiation and first-call
times are excluded from warmed execution and reported independently. Baseline
`4286585` has no public native backend, so its interpreter rows are identified
as such. The comparison uses pinned wazero 1.12.0, Wasmtime 48.0.1 and Wasmer
7.3.0, shared fixture bytes and independent checksums. Local uncalibrated timing
cannot accept M2 or authorize merging. Core 3.0, Threads, components and WASI
1/2/3 are still required and are not implemented by this increment.

Go statement coverage for this integration is 98.41% for `engine/execute`,
94.69% for `engine/nativeexec`, 96.09% for `platform/native`, 96.92% for `pkg/vm`,
and 97.88% for `compiler/codegen`; every measured executable package exceeds
90%. Native allocator/protection failure and asynchronous cancellation timing
leave defensive paths uncovered. Generated instructions and C code are verified
by execution, not included in those Go coverage percentages.

## Native vector register ownership and integer SIMD

Engine `f08d7df` lowers 122 additional basic SIMD opcodes on baseline SSE2 and
AArch64 NEON, sharing vector/scalar FP register allocation. Integer affine
fusion on arm64 preserves modulo-lane arithmetic, observable intermediate values
and every original cold continuation. Remaining SIMD operations retain explicit
owned helpers; this does not claim complete direct SIMD emission.

Tests compare every selected opcode with the owned SIMD evaluator across lane
immediates, destination aliases, register pressure, full-width spills, large
frames and cold entries. Fuel counts guest instructions rather than emitted
machine instructions; cancellation preserves all 128 bits. Fused and unfused
execution agree on every frame slot at every entry PC. The complete Core 2.0
corpus passes both native tiers with Go/LLGo on arm64 and Go under amd64
Docker/Rosetta. Measured Go statement coverage for `compiler/codegen` is 92.52%
on arm64 locally and 91.53% on amd64 in CI. Other exact package measurements are
94.71% for `engine/nativeexec`, 97.78% for `engine/state` and 96.92% for `pkg/vm`;
every measured executable package remains above 90%.

The [SIMD measurement record](../benchmarks/native/runtime-plan.md) retains the
two slower implementations and the final 60-block run. LLGo's final warm median
is 1.283 ns/iteration versus 1.793–1.807 for the three pinned competitors. Both
host compilers' A/A intervals remain inconclusive, so performance acceptance is
blocked despite the competitor advantage. This does not complete M2.

CI run [40462722506](https://github.com/cpunion/wasmvm/actions/runs/40462722506)
exposed an outdated CLI smoke assertion after the LLGo/Linux package suite
passed: it still expected native selection to fail. The smoke script now takes
an explicit build capability and verifies validate/inspect/run on both available
backends, or unsupported errors for all three commands in a cgo-disabled Go
build. It cannot accept either result by probing and adjusting its expectation.

Final code head `ab19548` passes all four Go/LLGo × Linux amd64/macOS arm64
[CI jobs](https://github.com/cpunion/wasmvm/actions/runs/40463558710), including
the CLI correction, complete interpreter/native corpus and shared lifecycle
tests. The amd64 Go job also passes coverage, race, strict-pointer, fuzz and
no-cgo checks. This supersedes the earlier missing LLGo amd64 semantic evidence;
it does not supply native amd64 performance evidence or close the blocked gate.

## Native fixed-SIMD memory operations

Engine `4b59731` adds all 22 SIMD memory opcodes on baseline SSE2 and NEON;
`246b906` selects widened ARM64 register-offset addresses. Engine `9cc2f0f`
forwards equal vector operands within bounded regions while retaining physical
copies for observable frames. Every opcode, lane and access width is compared
against the owned memory evaluator, including unaligned addresses, immediate
offsets 0/1/16/8193/65535/0xffffffff, large dynamic addresses, register pressure,
precise fuel and independently resumed interior PCs. Additional tests cover
missing bindings, cancellation, code limits, growth/reentry and all intermediate
slots at traps. Removing the new vector-memory trap barrier makes its regression
test fail; the diagnostic failure is retained.

The complete Go/LLGo × Linux amd64/macOS arm64 matrix passes at `9cc2f0f` in
[CI 40471490866](https://github.com/cpunion/wasmvm/actions/runs/40471490866).
Local full-suite Go coverage measures `compiler/codegen` at **93.12% (3233/3472
statements)**; every owned executable package remains above 90%. Native runtime,
state and public VM coverage remain 94.71%, 97.78% and 96.92%, respectively.
These percentages describe owned Go statements, not native instruction or
specification coverage. Full Core 2.0 interpreter and both native-tier assertion
counts remain unchanged.

Three independently checked workloads cover overlapping full-vector accesses,
widening/splat loads and lane/zero-extending accesses. Their benchmark scope and
failed diagnostics are recorded in [runtime-plan.md](../benchmarks/native/runtime-plan.md).
All timings exclude startup/compilation/instantiation, which remain separate raw
phase fields. Initial formal measurements confirm a full-vector access regression
against Wasmer; this feature is not accepted for merge based on semantic CI.

Engine `aa89f0d` separates small ARM64 memory immediates from the dynamic index,
matching the addressing dependency observed in the unchanged Wasmer library's
serialized machine code. Scalar and vector tests include offsets 4095/4096
across the encoding boundary. Go full-suite and LLGo targeted tests pass;
codegen coverage is **93.13% (3239/3478 statements)**. All four jobs pass in
[CI 40473998419](https://github.com/cpunion/wasmvm/actions/runs/40473998419).
Downloaded execution artifacts confirm 51,139 binary assertions across 148
scripts for each of the 12 compiler/target/backend-tier combinations, including
both native tiers; this is actual execution evidence, not just corpus inventory.
The sustained full-vector comparison now favors both host compilers against all
three references, but A/A calibration remains inconclusive. See the retained
before/after measurements and outstanding gates in the runtime benchmark plan.

## Native SIMD floating-point operations

Engine `c1c63d5` lowers all 52 Core 2.0 SIMD floating-point instructions directly
on both targets, including SSE2 rounding and signed/unsigned saturation without
optional ISA assumptions. Directed and randomized mixed-lane tests cover NaN
classes, signed zeros, subnormals, rounding halves, saturation boundaries,
aliases, spills, code budgets, exact fuel and every independent continuation.
Only specification-permitted NaN results use class comparisons; all other bits
and unaffected slots must match exactly.

All four [CI jobs](https://github.com/cpunion/wasmvm/actions/runs/40481474907)
pass. Downloaded execution artifacts independently confirm 51,139 binary
assertions across 148 scripts for all 12 compiler/target/backend-tier
combinations. Local full-suite coverage measures `compiler/codegen` at **93.88%
(3511/3740 statements)**; each measured owned executable package remains at least
90%. These are statement measurements, not instruction coverage claims.

Eleven packed floating-point workloads use independent Python lane oracles.
All 275 baseline/reference checks and 110 initial candidate checks pass across
five seed/iteration cases. The initial short diagnostic used sequential baseline
calls and interleaved candidate/reference calls; it is unsuitable for quantifying
small baseline regressions. Hashing 44 function/target/fuel-tier outputs confirms
that all eight earlier regression families retain identical generated code at
`c1c63d5`. A second diagnostic uses the same interleaving protocol for both
baseline and candidate and 100,000 iterations for the floating-point workloads.
Conversion results are promising; arithmetic, min/max, rounding and casts still
have slower comparisons. Raw phases and failed measurements are retained. See
[the scoped record](../benchmarks/native/vector-float-plan.md); no M2 performance
or merge gate is closed by these diagnostics.

The subsequent constant/selection optimizations through `0084b38` pass the
complete local Go and LLGo package suites and amd64 codegen/runtime/VM checks.
Updated full-suite coverage is **93.96% (3530/3757 statements)** for codegen;
all other measured executable packages remain at least 90%. Go vet and the
cgo-disabled build pass. The paired final diagnostic improves min/max by about
10–14%, but still has 1–3% slower arithmetic/min/max/rounding/cast comparisons
against the fastest reference. This is retained as failed performance evidence,
not a parity claim.

Final optimized code passes all four jobs at `a11d35a` in
[CI 40484968078](https://github.com/cpunion/wasmvm/actions/runs/40484968078).
Downloaded execution reports again verify all 12 compiler/target/backend-tier
combinations with 51,139 assertions each. A separate one-million-iteration
diagnostic reduces the apparent gaps: min/max and casts are within the nominal
parity band, while arithmetic/rounding still trail the fastest reference by
about 0.8–1.5%. No calibrated acceptance is inferred. The scope record separates
public-call overhead and an explicitly invalid-for-production no-poll experiment;
production cancellation remains intact.

## Complete fixed-SIMD lowering and detected x86 profiles

Engine `2d859c9` adds the final 39 integer/permutation opcodes; `49a5719`
adds explicit SSSE3/SSE4.1 inputs and platform CPUID detection. Both baseline
SSE2 and NEON cover all **236** Core 2.0 fixed-SIMD operations with no SIMD
runtime-helper exit. The [four semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40507703659)
pass; downloaded artifacts verify the complete 51,139 binary assertions / 148
scripts for all 12 compiler/target/backend-tier combinations. The first packed
shift/reduction engine `64e5e69` also passes the
[complete matrix](https://github.com/cpunion/wasmvm/actions/runs/40509424784).

The inventory test checks every reserved hole and forces baseline, SSSE3,
SSE4.1 and combined emission, with exact/exceeded code limits. Execution tests
use only host-advertised extensions and cover mixed saturation/Q15/dot boundaries,
all index classes, input/result aliases, spills, cold entries, fuel and masked
shift counts. The amd64 local execution probe reports `x86Features=0x3`;
optional instructions are actually exercised. Go/LLGo full package suites and
amd64 codegen/runtime/VM tests pass through `0ccf8f4`. Latest local codegen
coverage is **97.25% (4101/4217 statements)**; every measured executable package
is at least 90%. Go vet, cgo-disabled compilation and the measurement tooling
checks pass. A local container build once read source during an edit and failed
with truncated-file syntax errors; the committed, frozen snapshot passes.

NEON selection now names allocated operands directly. Compiler-created masks
share the immutable literal pool and its loop/cost priority. ARM64 uses otherwise
unused caller-saved v24-v31; a 20-literal test consumes the full pool and resumes
every PC, comparing every slot. Packed shifts replace lane-by-lane work on both
ISAs. Initial ARM64 i32/i64 reduction regressions are retained; subsequent
weighted masks, scalar i64 sign extraction and packed unsigned minima address
them. SSE4.1 min/max selection retains canonical NaN and signed-zero behavior;
pseudo-min/max keep selected raw bits exactly.

The existing native comparison workflow now has an explicit `native-simd`
scope, with both targets and both own compilers. It measures 33 integer,
floating-point, shift and reduction families without repeating semantic or CLI
suites. The [first complete two-target diagnostic](https://github.com/cpunion/wasmvm/actions/runs/40509770203)
retains **924 alternate-input oracle checks per target**, three interleaved
process blocks, seven calls and every phase/CPU mask. It exposes real slower
comparisons and does not pass a merge gate. Earlier setup failures (a short
baseline SHA and the macOS/darwin manifest naming mismatch) are retained and
fixed. No M1/M2 or PR #26 acceptance gate is closed by diagnostic success.

Final optimized engine `0ccf8f4` passes all four jobs at `eaee3f7` in
[CI 40512028373](https://github.com/cpunion/wasmvm/actions/runs/40512028373).
All 12 downloaded execution reports retain the full assertion/script counts.
The [final amd64/arm64 diagnostic](https://github.com/cpunion/wasmvm/actions/runs/40512069149)
completes all 33 families and 924 alternate-input checks per target. Paired
LLGo improvements include roughly 21–23% for amd64 popcount/shuffle, 26–32%
for amd64 min/max and 15–31% for ARM64 i16/i32/i64 reductions. The selected
cases still include slower competitor comparisons; all results and hashes are
recorded in the integer SIMD scope. M1/M2 and PR #26 gates remain open.


## OS-validated AVX2 lowering

Engine `f09d952` adds an explicit AVX2 compiler bit and VEX.128 three-operand
SIMD lowering. Synthetic CPU/OS probes cover missing XSAVE/OSXSAVE/AVX, leaf 7
AVX2 and both XCR0 state bits. Unknown/wrong-target profiles remain invalid;
all eight known masks have complete SIMD emission checks. Execution selects
only host-advertised profiles, with alias/spill/cold-entry, NaN/signed-zero,
Q15 overflow and invalid swizzle-index cases. Local amd64 reports mask 3;
the actual Linux CI benchmark reports mask 7 and therefore executes AVX2.

All four [semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40522056164)
pass. All 12 downloaded reports independently retain 51,139 assertions and
148 scripts. Local full Go/LLGo suites, amd64 fallback execution, vet and
cgo-disabled compilation pass. Codegen statement coverage is **97.19%
(4292/4416)**; every measured executable package exceeds 90%.

The [two-target diagnostic](https://github.com/cpunion/wasmvm/actions/runs/40522063062)
compares `c980109` with `f09d952` over 33 families and 924 alternate-input
oracle checks per target. LLGo amd64 shuffle improves from 2.503 to 2.150
ns/source iteration and average from 1.911 to 1.590. Narrowing (2.174) and
Q15 (2.247) beat their fastest measured references (2.186 and 2.381), but
swizzle regresses from 2.437 to 2.803 against a 1.941 reference. Widening
and other families also retain slower comparisons. ARM64 engine code is
unchanged, and its process variation must not be presented as an AVX2 gain.
These short diagnostics do not pass performance acceptance or PR #26.

Raw result SHA-256: amd64
`d6bf96b9bf6b1a144bc1c90078e6484b4c724c2d0a1d5551c4ffa1c39a8cb8ac`;
arm64 `848a29b1afc7c34b8e530f1c29687ef2e8645b5bbae598d14858749f0673f296`.
The next scoped comparison retains the eight existing scalar/call/memory
regression families and adds one-invocation call-control, for 42 cases.

## Reusable call workspace validation (development)

Engine `a918237` retains one bounded idle outer-call workspace per Store.
Targeted tests cover policy/fuel reset, function/slot changes, zeroed locals,
traps and fuel exhaustion, canceled loops followed by fresh calls, host panic
cleanup, nested reentry isolation, result encoding, released host roots and
Store close. Native lifecycle tests observe the real asynchronous cancellation
writer and prove that its sticky context cannot be reused; retired linear
memory cannot remain bound to a suspended workspace. Shared driver tests
reject rebinding active guest/host continuations and preserve borrowed results
through quiescence.

Local complete Go/LLGo suites, focused race checks, vet, cgo-disabled build and
68 measurement-tool tests pass. The final [four-job semantic CI](https://github.com/cpunion/wasmvm/actions/runs/40525608274)
passes. All 12 downloaded execution artifacts retain the complete 51,139
assertions / 148 scripts. Linux CI coverage measures execute **98.47%
(321/326)**, nativeexec **94.09% (207/220)** and public VM **96.42% (754/782)**;
every measured executable package exceeds 90%. Strict native-pointer checks
and the unavailable-native profile pass locally too.

The [paired comparison](https://github.com/cpunion/wasmvm/actions/runs/40525614537)
completes 42 families and **1,176 alternate-input checks per target**, using the
exact AVX2 baseline `f09d952`. One-invocation control results are ns/call:

| Target/compiler | Baseline | Workspace reuse | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| amd64 Go | 5,217 | 2,775 | 260 | 571 | 862 |
| amd64 LLGo | 12,719 | 9,354 | 260 | 571 | 862 |
| arm64 Go | 5,292 | 3,000 | 583 | 1,000 | 2,000 |
| arm64 LLGo | 6,958 | 7,333 | 583 | 1,000 | 2,000 |

LLGo single-call overhead is still far above the references, and the ARM64
candidate is slower in this short diagnostic. Reuse alone does not solve the
boundary. Sustained function/indirect-call loops beat several references, but
SIMD/memory families retain slower results. These are full public-call phases,
not pure native kernel measurements; no calibrated parity claim follows.
Raw result SHA-256: amd64
`85b747efa30caf756f6f11fdb4073f6a9542cb0c2ca5d6868c2cd620b89c88fc`;
arm64 `79fc97e391860c82ea9e74b322625ff2b8e85c04d7a35e8c811c5acb0455647a`.
M1/M2 and PR #26 remain open.


## Fallthrough polling and boundary attribution

Engine `20d0a41` removes only no-copy next-PC jump polls and replaces amd64's
load/test with one aligned full-width memory comparison. Cancellation tests
inject the flag after the C adapter check and verify forward/backward/copied/
conditional/table edges, periodic PC 256, long fallthrough chains, exact PCs and
live values under both policy tiers. Full local Go/LLGo and amd64 suites pass;
every measured executable package exceeds 90% coverage.

All four [semantic CI jobs](https://github.com/cpunion/wasmvm/actions/runs/40527305218)
pass; all 12 downloaded reports retain 51,139 assertions / 148 scripts. The
[paired two-target comparison](https://github.com/cpunion/wasmvm/actions/runs/40527313197)
completes 42 families and 1,176 alternate-input checks per target. LLGo amd64
function calls improve from 1.050 to 0.822 ns/source iteration and indirect
calls from 1.115 to 0.988; fastest references measure 1.647 and 1.750. Other
cases regress: i32 shifts measure 2.589 to 2.736 on amd64 and 3.698 to 3.870
on arm64. These short diagnostics retain every result and do not pass the
portfolio or calibrated performance gates.

Raw SHA-256: amd64
`1cb074e3d7d1464a213caa6c6d1e62cf34fcff794b3977eb5a97a5c193ef362e`;
arm64 `a5cc759a4e75b6bde7aaa1e6f90ca5879a7549e5d118e150515c596f4ccdf295`.

A separate repeated-call diagnostic on an Apple M4 Max, Go 1.27.3, pinned LLGo
`8b1389c`, engine `20d0a41` and GOMAXPROCS=1 isolates boundary components:

| Operation | Go ns/op | LLGo ns/op |
| --- | ---: | ---: |
| Mutex lock/unlock | 2.4 | 7.7 |
| Create/cancel context | 61.7 | 319.3 |
| Create/cancel context plus native cancellation registration | 321.4 | 1,273.3 |
| Native state write / read | 7.1 / 6.8 | 58.9 / 57.9 |
| Native entry alone | 83.1 | 239.7 |
| Shared native driver Start/Run | 126.1 | 568.8 |
| Public native call, including result inspection | 587.4 | 2,544.5 |
| Public host call, including result inspection | 544.0 | 2,346.3 |

The standard-library/cancellation/host path is a substantial residual cost;
these measurements must not be called pure Wasm instruction throughput. A
standalone non-loop deferred-unlock reproduction measures Go 4.29 ns versus
LLGo 51.71 ns (~96 reported bytes), while explicit unlock measures 2.73/7.97
ns and zero reported bytes. LLGo disassembly contains two AllocU calls (48 and
24 bytes); hosted ReadMemStats does not populate Mallocs, so its zero allocs/op
is not allocation evidence. [LLGo issue #2932](https://github.com/xgo-dev/llgo/issues/2932)
retains the complete standalone reproduction and relates it to the existing
escape-analysis proposal. Local raw JSON, binaries, source and disassembly
hashes are retained in `artifacts/native-avx2/call-profile-provenance.json`.
This identifies a compiler cost without weakening wasmvm cleanup or replacing
its shared runtime with compiler-specific execution paths.


## Rejected amd64 comparison replay

Candidate `1d5a3d5` preserves the original poll locations and replays a pure
comparison on unchanged registers, avoiding a boolean-to-branch dependency.
Aliased/spilled inputs retain the old path. Expanded tests cover all predicates,
widths, copied edges, source/result aliases, cold branch entries and real
cancellation at the original branch PC, including PC 256. Full local Go/LLGo,
amd64 execution, vet and cgo-disabled compilation pass; local codegen coverage
is **97.21% (4314/4438)** and every measured executable package exceeds 90%.
All four [candidate semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40530016726)
pass, with all 12 downloaded 51,139-assertion / 148-script reports verified.

The [paired diagnostic](https://github.com/cpunion/wasmvm/actions/runs/40530024451)
completes 42 families and 1,176 alternate-input checks per target. LLGo amd64
average regresses 1.637 -> 1.796 ns/source iteration and integer-sustained
2.134 -> 2.234; indirect calls improve 2.272 -> 2.179. Results do not establish
a stable portfolio win. **The replay implementation is withdrawn**; production
compiler sources return byte-for-byte to `20d0a41`, retaining the expanded tests.
No additional full performance rerun is needed to remeasure identical production
sources. Updated tests still receive the complete host/compiler semantic matrix.

Raw candidate SHA-256: amd64
`589fa5231cfa5a7f3c26541498bad889cf1a385585e464dda1a769a3dfec21a9`;
arm64 `ceb85fffe66597891e8b143bdd39f6e1da8510d48b452aa70c1b8ab4cdc26413`.
The pinned [Cranelift 48.0.1 lowering](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/cranelift/codegen/src/isa/x64/lower.isle#L2456)
selects VPOPCNTB when AVX512VL/BITALG are available, with a nibble-table fallback.
This is a concrete additional ISA path to investigate when explaining competitor
popcount differences. Actual instruction use on each measured host still needs
disassembly; a timing ratio alone does not identify its selected ISA.
PR #26 remains the only open contribution PR and stays blocked; no M1/M2,
Core 3.0/Threads, component or WASI acceptance is inferred from these diagnostics.


Retained head `4b09555` passes the final
[four compiler/target jobs](https://github.com/cpunion/wasmvm/actions/runs/40531604986).
All 12 downloaded reports verify the complete 51,139 assertions / 148 scripts.
Linux CI coverage is codegen **94.23% (4168/4423)**, execute **98.47%
(321/326)**, nativeexec **94.09% (207/220)** and public VM **96.42% (754/782)**;
every measured executable package exceeds 90%. Production sources are identical
to `20d0a41`; the retained tests additionally execute aliased comparisons,
original cancellation continuations and periodic PC 256 on both targets.


## Short-vector AVX-512 and exact SIMD immediates (development)

Engine `0bc1ba3` adds separate AVX512F+VL / AVX512BITALG+VL capabilities and
complete CPU/OS-state gating. Synthetic tests remove every required CPU/state
bit independently. LLVM 22 golden encodings cover EVEX.128 low/high registers;
complete SIMD emission and exact/exceeded code budgets cover all 32 known masks.
Host execution tests include only actually advertised capabilities, preserving
baseline execution on older hosts. Local Go/LLGo complete suites, amd64 fallback
execution, vet and cgo-disabled compilation pass. Local codegen coverage is
**97.21% (4321/4445)**; every measured executable package exceeds 90%.

The [initial two-target comparison](https://github.com/cpunion/wasmvm/actions/runs/40596344538)
completes 42 families and 1,176 alternate-input checks per target, comparing
`36b1d46` with `0bc1ba3`. The native amd64 host advertises **mask 7 for both
compilers**, so neither new AVX-512 path executes in this measurement. LLGo
popcount measures 2.332 -> 2.350 ns/source iteration; i64 shifts 3.973 -> 3.879.
These unchanged-profile measurements are not evidence for AVX-512 speed or
correctness. That initial run leaves native validation open; the later capable-host run
below supplies execution evidence. Neither diagnostic establishes calibrated
feature acceptance.

Engine `64932ea` introduces exact SIMD immediate forms. Its
[four semantic CI jobs](https://github.com/cpunion/wasmvm/actions/runs/40598533536)
pass. A batch of **5,253 native ABI probes** checks every 8-bit immediate form
against independent integer lane expansion and mathematical FP oracles, plus
unsupported NaN, negative-zero, subnormal and unequal-half patterns. Probes
rotate destination registers and publish one bounded native image. Existing
compiler tests retain pool exhaustion, raw bits, aliases, spills, all cold
entries and exact code limits.

The following shared-mask and shuffle-copy reductions through `c52c58f` pass
complete local Go/LLGo suites and amd64 tests, vet and cgo-disabled compilation.
Latest local codegen coverage is **97.17% (4389/4517)**; every executable package
remains above 90%. The [final four semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40609496290)
pass through `7f25ed9`, with all 12 complete corpus reports verified.
The [paired 42-family comparison](https://github.com/cpunion/wasmvm/actions/runs/40609500312)
against `0bc1ba3` passes 1,176 alternate-input checks per target. Results are
mixed: LLGo ARM unsigned f32 conversion improves 5.007 -> 4.938 ns/source
iteration, while shuffle changes 2.170 -> 2.183 versus wazero's 2.092.
AMD shuffle changes 2.096 -> 2.140 versus Wasmtime's 1.639. These are short
diagnostics, not portfolio acceptance. Reports retain CPU descriptions and
selected feature masks so ISA availability can be separated from selection.

The [later native run](https://github.com/cpunion/wasmvm/actions/runs/40611342681)
compares `7f25ed9` with `ccf7dfa` and supplies a capable **AMD EPYC 9V45** host:
both own host compilers select **mask 31** and actually execute EVEX popcount
and signed i64 shift. LLGo popcount is 1.526 ns/source iteration versus wazero
2.860, Wasmtime 1.455 and Wasmer 2.814; i64 shift is 2.672 versus 7.511,
2.520 and 3.472. Both own revisions already contain EVEX selection, so this
is execution/competitor evidence, not an isolated head/base ISA speedup.
The [same revision's semantic matrix](https://github.com/cpunion/wasmvm/actions/runs/40610849392)
passes all four jobs and 12 complete reports. Native shuffle still trails the
fastest reference; optional-ISA support does not close the performance gate.
M1/M2, PR #26 and subsequent GC/Threads/component/WASI stages remain open.


## Shared interpreter operand and metadata access (development)

**Local provenance correction:** the early local rows below used a bootstrap
`llgo (devel)` CLI. Its build info identifies `bc7c991...` with local changes,
while its runtime source tree was pinned to `8b1389c`. Those rows remain
bootstrap diagnostics, not verified fixed-compiler evidence. The linked CI
builds independently report the pinned `8b1389c627c7` version and are unaffected.
The final paired run below rebuilds both revisions with a clean pinned CLI;
`3daa886` now rejects a missing/wrong executable version even when the revision
environment variable is set, and records the CLI's Go build information.

The [declared scope](../benchmarks/native/operand-plan.md) removes per-instruction
getter closures and captured receiver/base allocations. The first method-based
candidate `ccf7dfa` still emits three LLGo calls; `f6e77f7` reads guarded slots
inline and snapshots backend/metering selection once per Run. Its
[four semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40612478202)
pass; all 12 complete reports are verified. Disassembly retains the earlier
allocating/calling variants rather than attributing their results to later code.

Engine `43bb8a7` borrows immutable Source/Wasm metadata instead of copying it
for scalar dispatch. A read-only fuel receiver avoids another aggregate copy
when metering. Existing SIMD/bulk helper value boundaries, immediate bits, fuel,
trap offsets, cancellation and frame ownership remain unchanged. Full local Go
and affected LLGo tests pass; every measured executable package exceeds 90%,
including execution **98.48% (323/328)** and IR **99.52% (208/209)**.
The [final engine matrix](https://github.com/cpunion/wasmvm/actions/runs/40613158896)
passes all four compiler/target jobs; all 12 reports independently retain 148
scripts and 51,139 binary assertions. The [paired native regressions](https://github.com/cpunion/wasmvm/actions/runs/40613243428)
complete 42 families and 1,176 alternate-input checks per target. This amd64
host selects mask 7; the earlier mask-31 result remains separate. LLGo native
call-control improves 23144 -> 19046 ns/call on amd64 and 7709 -> 6542 on arm64,
but remains slower than wazero's 882 / 541. Sustained and SIMD results remain
mixed; the source optimization does not change generated native loop code.


Local macOS arm64 diagnostics compare exact `7f25ed9` / `43bb8a7`, using recorded
Go/LLGo and pinned wazero interpreter builds: three interleaved process blocks, seven
calls/two warmups, 100,000 source iterations per loop call. The six families
pass 216 alternate-input/runtime checks. Startup, preparation, instantiation,
first call and public API versus reused internal dispatch remain separate.
Public API warm medians, ns/source iteration:

| Family | Go base -> head | LLGo base -> head | wazero interpreter (Go) |
| --- | ---: | ---: | ---: |
| Integer | 227.255 -> 161.807 | 1071.383 -> 179.261 | 132.827 |
| Function calls | 311.167 -> 203.893 | 1410.549 -> 355.843 | 143.813 |
| Memory | 435.147 -> 288.059 | 2180.318 -> 374.500 | 250.843 |
| Shuffle | 325.438 -> 270.686 | 1862.438 -> 866.340 | 181.599 |
| Popcount | 314.368 -> 268.457 | 1517.051 -> 977.724 | 214.815 |

The single-invocation public LLGo control is 6792 -> 5083 ns/call versus wazero
1208. The scalar gains do not establish parity: calls and SIMD remain slower.
These local measurements have no all-target A/A calibration. LLGo's default
build is size-oriented `-Os`, without LTO; a separate identical-flags base/head
`-O3 -lto=full` experiment is retained separately and does not overwrite these results.
Raw default results: SHA-256
`d52a152903ac23cbefb2bc3fd1c2f6c645d9a81dfc2f63174b0bdc89130707f0`,
retained with build/source/binary provenance under `artifacts/operand-source`.
Earlier method and inline-only diagnostics remain under `artifacts/operand`
and `artifacts/operand-inline`; all are diagnostics, not merge authorization.


The separate release-options experiment interleaves both default and optimized
LLGo base/head configurations with Go and wazero in the same six-family run
(312 alternate-input checks). The complete local LLGo `-O3 -lto=full` package
suite passes. Identical source `43bb8a7`, public API warm ns/source iteration:

| Family | LLGo default | LLGo -O3/full LTO | wazero interpreter (Go) |
| --- | ---: | ---: | ---: |
| Integer | 170.102 | 137.136 | 113.196 |
| Function calls | 279.239 | 209.645 | 123.236 |
| Memory | 248.451 | 202.625 | 164.993 |
| Shuffle | 562.570 | 536.358 | 123.445 |
| Popcount | 625.957 | 568.359 | 133.795 |

Single-call LLGo control is 5041 -> 4375 ns/call versus wazero 1209. The
optimized host build is not a VM-source gain and does not establish parity.
Its full target matrix/release adoption remain open; default CI flags are
unchanged. Raw results SHA-256:
`e70ef6e4e6c31463228e8d9614f1a3730b462236943c8ff452adc1275831c30d`
under `artifacts/llgo-release`, with both flag sets and exact archived sources.
Even the optimized SIMD evaluator still allocates a 112-byte instruction copy
and a 16-byte addressed result on its entry path; scalar host options cannot
remove this source-level ownership problem. The next scoped change borrows
immutable SIMD instructions and returns lane updates by value.


## SIMD value ownership and packed integer arithmetic (development)

Engine `1e296a8` borrows immutable pure-SIMD instructions and returns lane
updates by value. This removes the normal-entry 112-byte instruction and
16-byte result allocation requests emitted by the bootstrap LLGo CLI; local
reported allocation traffic drops from roughly 160 B/op to zero for integer
operations. These are compiler requests/runtime counters, not a general
allocation-free claim. Floating-point operations still allocate in standard
math bit-conversion helpers. A standard-library-only reproducer confirms
~32 B for a round trip and ~48 B for a scalar add with a clean pinned CLI,
including full LTO; see [LLGo #2935](https://github.com/xgo-dev/llgo/issues/2935).

Engine `3bd89e3` implements packed modular add/subtract for all four integer
lane widths, isolating carries/borrows with lane high-bit masks. Byte popcount
counts bit pairs/nibbles/bytes in two words. The existing evaluator and native
emission remain shared; no workload-specific execution path is introduced.
Tests exhaust 65,536 byte pairs and check wider boundary/random values against
independent scalar oracles. The [four final semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40617055082)
pass; all 12 reports verify 148 scripts / 51,139 binary assertions. SIMD package
coverage is **97.76% (262/268)**; every measured executable package exceeds 90%.
The clean pinned CLI also passes local optimized SIMD/execution/VM tests.

Final local comparison: exact `1e296a8` -> `3daa886` (production engine
`3bd89e3`), Apple M4 Max. Both revisions were rebuilt with the clean CLI from
`8b1389c627c77a02e56af03f08d9cdc19edffe77`, with `vcs.modified=false` verified.
Go 1.27.3, default LLGo and `-O3 -lto=full` LLGo configurations stay separate.
Three interleaved process blocks, seven calls/two warmups, 100,000 source
iterations; six families pass 312 alternate-input/runtime checks. Startup,
preparation, instantiation, first call and reused dispatch are separate raw
fields. Public API warm medians, ns/source iteration:

| Family | Go base -> head | LLGo default base -> head | LLGo -O3/full LTO base -> head | wazero interpreter (Go) |
| --- | ---: | ---: | ---: | ---: |
| Integer | 111.974 -> 104.863 | 123.978 -> 121.927 | 96.893 -> 94.625 | 87.728 |
| Function calls | 134.974 -> 131.425 | 231.472 -> 230.347 | 152.685 -> 152.242 | 96.700 |
| Memory | 201.463 -> 192.273 | 242.352 -> 245.255 | 180.465 -> 179.195 | 158.957 |
| Shuffle | 159.355 -> 136.951 | 206.082 -> 187.865 | 151.568 -> 135.518 | 113.770 |
| Popcount | 163.200 -> 116.743 | 230.100 -> 146.793 | 159.444 -> 109.879 | 134.422 |

Only the final optimized-LLGo popcount row is ~18% lower than wazero's elapsed
time here (Go is ~13% lower). Default LLGo popcount and other families still
trail; neither a configuration win nor a single family accepts the interpreter.
Single public-call LLGo control is 3083 -> 2917 ns/call with full LTO versus
wazero 792. Packed pure evaluation leaves native emission unchanged and supplies no new
native timing evidence. Earlier native results and PR #26's A/A failure remain
separate.

Isolated Eval component medians confirm the algorithmic effect under the same
clean CLI: optimized LLGo i32x4.add **15.459 -> 6.355 ns/op**, byte popcount
**43.064 -> 4.665 ns/op**. These are primitive timings with operand/checksum
updates, not whole-runtime comparisons. Popcount reports zero allocation bytes;
an add block retains 0.036 B/op of process-level counter traffic, so counters
alone do not prove zero engine events. f32x4.add still reports ~192 B/op,
consistent with the separately reproduced compiler-helper cost. M1/M2 and
Core 3.0/Threads/components/WASI 1/2/3 remain open.

Final raw results SHA-256: `391a54424bfd1402cc440d11e4bac83ef95b0c0dae853347cbf032cf40ba2d89`.
Component results SHA-256: `e5325303b39a22de991c9153919ab45748c4006b0d388221dda5f233e1399c7c`.
Retained under `artifacts/simd-packed-pinned` with exact source/binary/CLI hashes.
These short local diagnostics lack all-target A/A calibration and do not authorize merging.


## Proposed LLGo caller-side math bitcasts

[LLGo PR #2936](https://github.com/xgo-dev/llgo/pull/2936) addresses
[issue #2935](https://github.com/xgo-dev/llgo/issues/2935) by selecting the
existing four LLVM bitcast intrinsics on every target. 386 retains its earlier
x87/signaling-NaN protection. Direct calls no longer allocate temporary scalar
storage; ordinary function values retain their existing supported path. No VM
math implementation or compiler-specific VM fallback is added.

Exact compiler base `748cf18118e4d2e9cd4f8e953490e50945572ef6` and proposed
head `8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` were built from clean checkouts,
with full revisions and `vcs.modified=false` retained. Caller-side IR checks
cover windows/386, linux/amd64, darwin/arm64 and wasip1/wasm. Raw-bit tests include
signed zero, subnormals, infinities and quiet/signaling NaNs; function-value and
reflection tests pass. The complete local wasmvm package suite also passes
under the fixed compiler. Upstream CI/review and production-pin adoption remain
separate requirements; the wasmvm production pin is still `8b1389c`.

Local Apple M4 Max standalone default-compiler medians (100,000 calls, three
process blocks): f32/f64 bit round trips change roughly 26 ns / 32 reported B
per call to 1 ns / zero reported B. Scalar additions change roughly 37 ns /
48 B to 1 ns / zero B. Full LTO removes the same allocator calls; one f64 block
retains 0.026 B/call of process-counter traffic. Disassembly shows register
moves/addition without allocator or math-helper calls. Hosted Mallocs zero is
not used as evidence.

The same exact wasmvm source `6d19583` is then rebuilt with both compilers,
keeping default and `-O3 -lto=full` configurations separate. Ten interpreter
families pass 440 alternate-input/runtime checks. Three interleaved process
blocks, seven calls/two warmups and 100,000 source iterations retain public API,
reused dispatch and all preparation/first-call phases separately. Selected
public API warm medians, ns/source iteration:

| Family | LLGo default base -> proposed | LLGo -O3/full LTO base -> proposed | wazero interpreter (Go) |
| --- | ---: | ---: | ---: |
| f32 arithmetic | 662.571 -> 200.403 | 424.703 -> 131.972 | 104.114 |
| f64 arithmetic | 394.211 -> 180.997 | 266.575 -> 119.538 | 109.734 |
| Unsigned f32 conversion | 826.282 -> 270.272 | 533.722 -> 180.191 | 167.623 |
| Mixed casts | 617.868 -> 213.508 | 394.951 -> 149.140 | 135.463 |

These gains remove a compiler-side cause of floating-point overhead but still
trail the reference. Integer/call/memory/SIMD/control regressions remain visible
in the raw report. No all-target or calibrated performance acceptance is claimed.
Runtime raw SHA-256:
`460eed4194ceb8fc7a3fdda5a171eed0e8e4b733bb045c436c181dc85f9ccdb8`.
Raw reports, source/binary hashes and clean compiler build information remain
under `artifacts/llgo-bitcasts`. M1/M2 and later profiles remain open.


## Per-frame metadata and opcode dispatch

The scoped follow-up caches immutable function/code/instance/base metadata for
one active frame and refreshes it after frame transitions. The common driver
still reloads the preserved value arena after push; a moving-arena regression
explicitly invalidates old views during four nested frame/slot expansions.
Every original PC, fuel charge, cancellation poll and trap boundary remains.

The first metadata-only candidate `1e0b917` has mixed local LLGo results and is
not an accepted standalone performance improvement. Candidate `884d302` replaces
the scalar boolean-condition ladder with a value switch. Validated scalar and
control opcodes can use dense/range dispatch; SIMD and bulk retain their
existing handling. Disassembly confirms an indirect jump-table branch. No
bytecode layout, opcode semantics, backend or instruction fusion is added.
These changes are integrated through production engine `a87954f`.

Local exact `1e0b917` -> `884d302`, with the same proposed LLGo compiler
`8e45ed5e` on both sides, keeps the compiler change separate from VM-source
changes. Go/LLGo default and optimized execution/VM tests pass locally; every
measured executable package exceeds 90%. The four-job Go/LLGo × amd64/arm64 production matrix uses the existing
`8b1389c` pin. [Final CI](https://github.com/cpunion/wasmvm/actions/runs/40625855069)
passes all four jobs; all 12 reports independently retain 148 scripts and
51,139 binary assertions. Every measured executable package exceeds 90%.
Ten families pass 520 alternate-input/runtime checks, with three interleaved
blocks, seven calls/two warmups and 100,000 source iterations. Public API warm
medians, ns/source iteration:

| Family | Go base -> head | LLGo default base -> head | LLGo -O3/full LTO base -> head | wazero interpreter (Go) |
| --- | ---: | ---: | ---: | ---: |
| Integer | 94.376 -> 68.459 | 119.142 -> 95.966 | 86.332 -> 75.823 | 83.977 |
| Calls | 124.394 -> 97.290 | 232.915 -> 190.785 | 175.259 -> 142.285 | 89.389 |
| Memory | 175.568 -> 124.823 | 225.419 -> 185.632 | 161.274 -> 147.697 | 146.613 |
| Shuffle | 126.486 -> 104.427 | 173.359 -> 156.327 | 122.733 -> 112.055 | 104.705 |
| Popcount | 104.884 -> 80.472 | 139.710 -> 123.518 | 102.957 -> 90.880 | 125.236 |
| f32 arithmetic | 142.925 -> 116.083 | 191.196 -> 172.333 | 130.433 -> 121.518 | 105.084 |
| f64 arithmetic | 127.171 -> 100.799 | 176.282 -> 150.619 | 117.696 -> 106.623 | 110.037 |
| Unsigned f32 conversion | 220.499 -> 186.807 | 273.004 -> 242.811 | 181.162 -> 165.236 | 169.175 |
| Mixed casts | 171.884 -> 138.850 | 206.954 -> 179.051 | 143.195 -> 136.102 | 136.690 |

A few point estimates match or beat wazero; calls, f32 and other configurations
still trail. The tiny public-call control also regresses in this local run
(optimized LLGo 3084 -> 3917 ns/call, wazero 1041); keep this observation even
though it is below reliable percentage-gating resolution. Startup, preparation,
first call and reused dispatch remain separate. No all-target A/A or complete
interpreter/native acceptance is claimed. No later stage is marked complete.

Final source-comparison raw SHA-256:
`1812eaea07cef340af8fbe241823f2418292d7fdc4f12f926df9a1a92257adf9`.
Raw reports/source/binary hashes remain under the isolated frame worktree's
`artifacts/frame-dispatch` and `artifacts/dispatch-switch`. The compiler proposal
is still unmerged; these rows do not change the production toolchain pin.


The [final paired PR-scope CI](https://github.com/cpunion/wasmvm/actions/runs/40625856666)
also compares exact `6d19583` / `a87954f` with the **production default** Go/LLGo
build flags and compiler `8b1389c`, separately from the proposed-toolchain local
rows. Linux public interpreter ns/source iteration:

| Family | Go base -> head | LLGo default base -> head | wazero interpreter |
| --- | ---: | ---: | ---: |
| Integer | 190.68 -> 132.13 | 224.14 -> 182.62 | 175.12 |
| Calls | 248.00 -> 226.15 | 441.07 -> 430.60 | 181.27 |
| Memory | 344.30 -> 246.05 | 445.21 -> 394.84 | 323.05 |

Linux native integer timing passes its narrow unchanged competitor/A/A gate
(Go 1.870, LLGo 1.868 ns/source iteration; wazero 1.908, Wasmtime/Wasmer 1.868).
The macOS native job reaches 60 blocks but remains inconclusive: LLGo A/A
interval 0.986–1.022 fails the unchanged ±0.25% calibration band, and the
Wasmtime/Wasmer comparisons are also inconclusive. Its raw failure is retained
rather than using the Linux pass as all-target acceptance. Interpreter throughput and public-call
controls also retain deficits. These results leave PR #26 and M1/M2 open.

## Shared return forwarding

The [scoped change](../benchmarks/native/return-plan.md) compares exact base
`bc92e03854605d87b5f8fae71b0409af6541d11d` with engine
`b91095284c4af0a06d418637774ed90bbe6fcc0b`. Slot lowering changes an unconditional
forward edge to a pure, ordered return into a direct return of the original
operands. Original PCs and the predecessor's source/fuel charge remain stable.
Backward, conditional, effectful and reordered/subset return blocks retain
their transfer plans. Native leaf expansion still recognizes the reached exit.

The complete local Go and production-pinned LLGo package suites pass. New tests
cover reverse-ordered mixed scalar/vector/reference results, explicit and
implicit return fuel boundaries, ineligible control/tuple shapes, and native
leaf recognition. Existing host suspension, arena relocation, resource limits
and lifecycle tests remain enabled. All measured executable packages exceed
90% coverage: slots 99.50%, codegen 97.17%, execute 98.50%, nativeexec 94.09%
and public VM 96.42%. [Final production CI](https://github.com/cpunion/wasmvm/actions/runs/40628184046)
passes all four Go/LLGo × Linux amd64/macOS arm64 jobs; all 12 execution reports
retain 148 scripts and 51,139 binary assertions.

Local phase diagnostics rebuild both exact source revisions using the same
Go 1.27.3 and production LLGo `8b1389c` compilers. Default and `-O3 -lto=full`
LLGo rows are separate; the proposed bitcast compiler is not used. Seven
families and 22 runtime/backend/boundary configurations pass 616 alternate-input
checks. Three interleaved process blocks retain seven calls, two warmups and
100,000 source iterations, with preparation and first call separate. Public
interpreter warm medians in ns/source iteration:

| Family | Go base -> head | LLGo default base -> head | LLGo -O3/full LTO base -> head | wazero interpreter (Go) |
| --- | ---: | ---: | ---: | ---: |
| Direct source calls | 90.141 -> 80.296 | 190.915 -> 162.900 | 135.088 -> 128.990 | 80.628 |
| Indirect calls | 95.120 -> 85.415 | 177.342 -> 170.141 | 139.556 -> 135.005 | 90.725 |
| Integer | 67.202 -> 66.168 | 87.620 -> 87.517 | 74.379 -> 76.372 | 82.193 |
| Memory | 124.914 -> 123.686 | 180.500 -> 181.988 | 145.061 -> 147.198 | 146.674 |
| Shuffle | 101.985 -> 103.625 | 149.090 -> 151.347 | 111.527 -> 111.072 | 103.353 |

Native direct-source-call medians, also ns/source iteration, compare all three
pinned native references separately:

| Build | Base -> head | wazero native | Wasmtime Cranelift | Wasmer Cranelift |
| --- | ---: | ---: | ---: | ---: |
| Go | 1.067 -> 1.059 | 1.580 | 0.950 | 0.840 |
| LLGo default | 0.888 -> 0.891 | 1.580 | 0.950 | 0.840 |
| LLGo -O3/full LTO | 0.881 -> 0.920 | 1.580 | 0.950 | 0.840 |

Go interpreter call point estimates approach wazero, while LLGo interpreter
and native calls still have deficits. Tiny public-call controls remain above
the M2 budget: Go interpreter 2459 -> 2250 ns/call and LLGo default 2875 ->
2834 ns/call; LLGo native regresses 6417 -> 6875 ns/call in this diagnostic.
Preparation also needs investigation: optimized LLGo native source-call
preparation medians rise 247,042 -> 622,958 ns, with base process values
225,708–571,417 and head values 249,167–623,208. Do not attribute this spread
or unchanged-loop timing differences to return forwarding without controlled
evidence. SIMD floating-point results, other regressions, internal dispatch,
startup, instantiation and all first-call samples remain in the raw report.

The call fixture retains 33 slot instructions and 14 frame slots, with edge
moves reduced 2 -> 1 and argument-pool entries 5 -> 6. Static baseline-target
function emission retains one expanded leaf call. Expanded bodies grow
3172 -> 3263 bytes on amd64 and 2880 -> 2948 on arm64 because the former
return continuation remains independently addressable; these are static body
sizes, not peak-memory measurements or complete publication sizes.

Raw SHA-256:
`a5652a5da78273c6050769586c75561cba13cb7707696e61341fc51a21cdbdb2`.
Build/source/binary provenance, shape counts and raw observations remain under
`artifacts/return-forward`. An initial harness invocation rejected its own
one-sample correctness mode against the shared seven-sample validator; its
failed report is retained separately before the protocol-consistent run.
No new sustained benchmark is run for an unchanged native integer kernel.
These local diagnostics do not close the missing all-target performance,
call/preparation deficits or PR #26's ARM A/A calibration gate. M1/M2 and
Core 3.0/Threads/Component Model/WASI 1/2/3 remain unfinished.

## Integer dispatch, active state and LLGo release profile

The source baseline is `6e061ac2263dc80dbe99441f392ad762ff99f87d`.
Candidate `af96ddc` extracts 58 nontrapping integer primitives shared by the
generic evaluator and main dispatcher. Candidate `9769078` instantiates those
same typed-bit bodies in consuming packages; default LLGo's 58 integer helper
calls disappear from the main-loop disassembly, but its timings still regress.
Neither candidate is accepted as an independent performance win.

Engine `a7f2b1e128873a63139d2e02191a16468171810b` keeps its active PC and
bounded frame-value view local, publishing the next PC at guest/host call
boundaries. Arguments copied after push still read the relocated global arena.
New repeated nested-host suspension tests retain continuation/result order in
both fuel modes; moving-arena, cancellation, trap and reentry tests remain.

Two configuration changes are recorded separately from VM-source improvements.
The LLGo pin advances `8b1389c` -> `8e45ed5e` after
[upstream PR #2936](https://github.com/xgo-dev/llgo/pull/2936) passes its CI; the
contribution remains unmerged. Its caller-side math bitcasts remove the earlier
FP allocations. The supported LLGo release build explicitly uses
`-O3 -lto=full`, consistently in the common phase/CLI builder, native probe and
semantic CI. Both source revisions in each comparison use identical compiler
revisions and flags. Old default `-Os` rows are retained, not relabeled.

[Release-profile CI](https://github.com/cpunion/wasmvm/actions/runs/40636529944)
passes all four Go/LLGo × Linux amd64/macOS arm64 jobs through engine `a7f2b1e`;
all 12 downloaded reports retain 148 scripts and 51,139 binary assertions.
The earlier default-flag matrix also passes through `9769078`
([CI](https://github.com/cpunion/wasmvm/actions/runs/40634334604)).

Twelve local families and 22 runtime/backend/boundary configurations pass 1056
alternate-input checks. Three interleaved process blocks retain seven calls and
two warmups. Public interpreter medians, ns/source iteration:

| Family | Go base -> head | LLGo release base -> head | wazero interpreter |
| --- | ---: | ---: | ---: |
| Integer | 64.915 -> 61.488 | 73.636 -> 64.158 | 81.925 |
| Direct source calls | 79.368 -> 77.109 | 128.330 -> 117.747 | 79.792 |
| Indirect calls | 85.522 -> 80.021 | 131.971 -> 122.247 | 90.265 |
| Memory | 123.038 -> 118.917 | 144.348 -> 129.967 | 146.837 |
| Signed f64 conversion | 160.580 -> 153.627 | 151.183 -> 128.740 | 162.623 |
| Unsigned f64 conversion | 177.412 -> 164.926 | 176.510 -> 143.082 | 191.050 |

The same run retains the default-flag regressions: LLGo source calls
160.054 -> 182.055, memory 180.742 -> 186.922 and shuffle 151.855 -> 153.918
ns/iteration. Go SIMD and several release FP/SIMD/control rows also regress.
Removing helper calls does not prove a throughput gain, and switching profiles
does not accept a default-build deficit. Release calls and f32 SIMD still trail
wazero. Preparation/first-call and complete phase records remain separate.

Application inputs reuse exact portfolio commit
[`2b20a74`](https://github.com/cpunion/wasmvm/tree/2b20a742fdfd35e8e03ea9d2360458794ef9e039/benchmarks/embedded/portfolio):
the Clang 22.1.8 image filter and Rust 1.98.0 packet CRC. Source/WAT/binary hashes
and independent checksums are verified. Counts retain 40 image passes and 64
packets per call. Interpreter medians in microseconds per pass/packet:

| Application | Go base -> head | LLGo release base -> head | wazero interpreter |
| --- | ---: | ---: | ---: |
| C image filter | 261.311 -> 248.251 | 388.256 -> 354.282 | 350.913 |
| Rust packet CRC | 92.868 -> 83.594 | 92.305 -> 78.306 | 127.669 |

The image filter still trails wazero under LLGo. These local application inputs
do not replace a complete controlled portfolio or all-target performance gate.
Reports retain the mixed first candidates under `artifacts/integer-dispatch`
and `artifacts/integer-consumer`, plus the final active-state comparison under
`artifacts/active-frame`. Their raw SHA-256 values are respectively
`68cad181172ed0431be0348325ef2e98be1ed5a70af71900e186baeed720ed2a`,
`3cb5db09b428f0aa69dd275af1602802bc5ed27f6674f44bf6516279b84f0b74`,
and `1c7bb456282980604682eea1b7742b84138183c48a326024c31083bd4ecd37e9`.

## Native scalar select

Native `select` previously emitted branches. Candidate `5a96815` emits
CMOV/CSEL; final engine `f1513a828a1264e99bca25620894c4ce1e23cc29` directly
uses allocated integer/reference registers with spill/FP scratch fallback.
Condition tests use low i32 bits, including when a destination aliases the
condition. Scalar NaN payloads and reference bits are preserved. The precise
tier selects both input words, including its existing noncanonical continuation
state; unmetered scalars retain the canonical-input contract. V128 keeps its
complete prior implementation. Tests cover both targets' encodings, both native
tiers, operand/destination aliases, high condition bits and register pressure.

Full local Go and LLGo release execution suites pass. Every measured executable
package exceeds 90% coverage: codegen 97.15%, execute 98.72%, numeric 100%,
nativeexec 94.09% and VM 96.42%. The final
[CI](https://github.com/cpunion/wasmvm/actions/runs/40639955242) passes all four
Go/LLGo release × Linux amd64/macOS arm64 jobs. All 12 downloaded execution
reports retain 148 scripts and 51,139 binary assertions each; Linux codegen
coverage is 94.30%, with every executable package above 90%.

The exact source comparison retains base `a7f2b1e`, intermediate `5a96815` and
head `f1513a8`, all with the same pinned compilers/release flags. Twelve families
and 15 native API/dispatch/reference configurations pass 720 alternate-input
checks. Public native CRC medians, microseconds per 256-byte packet:

| Build | Branches | Scratch CMOV/CSEL | Register CMOV/CSEL | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Go | 7.704 | 3.477 | 2.870 | 1.700 | 1.703 | 1.766 |
| LLGo release | 7.589 | 3.443 | 2.874 | 1.700 | 1.703 | 1.766 |

This general opcode improvement is about 2.6x faster on that local application,
but remains slower than every reference. Image-filter native LLGo is 3.132
microseconds/pass versus wazero 2.982, Wasmtime 1.150 and Wasmer 1.200.
Fixed slot-frequency allocation, temporary-value traffic and broader compiler
optimization remain required work. Other native regressions remain visible:
LLGo memory 0.825 -> 0.854 ns/source iteration, for example. Tiny call controls
remain above budget. No phase or failed observation is subtracted or discarded.

The CRC still has 138 slot instructions and 21 frame slots. Static unmetered
function bodies shrink 3924 -> 3860 bytes on arm64 and 4336 -> 4258 on amd64;
the implementation adds no runtime helper or CRC-specific execution path.

Native CRC preparation process medians are Go 363,041 -> 349,166 ns and LLGo
997,208 -> 902,334 ns; all individual preparation, first-call and cleanup samples
remain in the report. The native dispatch validator now requires its existing
measured `native` preparation component, instead of rejecting it as an unknown
field; tests retain strict interpreter/native schemas and aggregate accounting.
The initial harness rejection and intermediate measurements remain under
`artifacts/native-select`, raw SHA-256
`a1eeea5ccb77c35780b99bb50088c94471ad22cdf935599b3fdb8e9e74fec960`.
Final raw/source/binary provenance remains under `artifacts/select-register`,
raw SHA-256 `7f912a5ddec85f4b15be286708a1b7e8a02e17d4acbd60612ae32136a3b24120`.
These diagnostics do not accept native/application parity, resolve PR #26's
ARM A/A failure, or complete M1/M2. Core 3.0/GC/Threads/components/WASI 1/2/3
remain required subsequent work.

Manual semantic workflows now make unchanged wazero/wago compatibility
evaluations an explicit `evaluate_references` option, defaulting to false.
The four own-engine/compiler/target jobs, corpus, coverage, race, pointer and
CLI checks remain required; this removes repeated M0 work from core iterations.

## Shared local and literal slots

Engine `4f531dd620a3c4920fb70d59e78286df7e2e02bd` coalesces block-local
local.get/tee results only until the next local write. Identical typed numeric
and vector literals share storage only while their SSA definitions overlap;
last-alias release recycles it. Cross-block values remain dedicated. Linear
scans and block-local pools retain bounded compilation. Native register weights
ignore self-copies, include argument/return tuples, and avoid duplicate tee
destinations. No Wasm opcode, source/fuel boundary or original PC is removed.

The independent symbolic checker models local versions and exact literal bits.
Tests cover old values across writes, tee, blocks, types/NaN payloads/signed
zero/vector halves, alias release, bounded frames, quotas, suspended reference
roots, zero-initialized reuse and arbitrary native cold-entry slot contents.
Complete local Go and LLGo release suites pass: each of six execution reports
retains 148 scripts and 51,139 binary assertions. Every executable package is
above 90% coverage; locally slots is 99.60%, codegen 96.81%, execute 98.72%.

The first local-only candidate `7bfd24b` remains an unaccepted intermediate:
its initial native CRC process medians regress under both host compilers.
Its [CI](https://github.com/cpunion/wasmvm/actions/runs/40646194361) retains 12
passing execution reports, but three jobs fail uploading artifacts from the
disabled reference evaluation. Commit `ff0e525fec3fdaea20b848651d8a7587ab7a9982`
fixes that condition; its engine sources equal `4f531dd` exactly. The final
[CI](https://github.com/cpunion/wasmvm/actions/runs/40649187641) passes all four
Go/LLGo release x Linux amd64/macOS arm64 jobs. All 12 downloaded execution
reports retain 148 scripts and 51,139 binary assertions each. Linux codegen
coverage is 93.86%, slots 99.60%, with every executable package above 90%.

Final measurements compare exact base `f1513a8`, intermediate `7bfd24b` and
head `ff0e525`, with identical pinned Go/LLGo release compilers and reference
binaries. Twelve families and 22 configurations pass 1056 alternate-input
checks; intermediate rows cover native API execution. Three interleaved process
blocks retain seven calls after two warmups, not a controlled acceptance run.
The same pinned C/Rust portfolio commit `2b20a74`, producers, hashes, counts and
independent checksum oracles are reused. Public native medians, microseconds
per image pass or 256-byte CRC packet:

| Application / host | Base | Intermediate | Head | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| C image / Go | 3.366 | 3.360 | 3.375 | 3.429 | 1.200 | 1.225 |
| C image / LLGo | 3.097 | 2.957 | 2.927 | 3.429 | 1.200 | 1.225 |
| Rust CRC / Go | 2.987 | 3.236 | 2.039 | 1.818 | 1.531 | 1.641 |
| Rust CRC / LLGo | 2.645 | 2.873 | 1.730 | 1.818 | 1.531 | 1.641 |

LLGo native CRC is about 35% faster than its base and ahead of wazero in this
run, but still slower than Wasmtime/Wasmer. Image filtering still has a large
deficit against those engines. CRC instructions remain 138 while frame slots
fall 21 -> 15; unmetered code shrinks 4258 -> 3803 bytes on amd64 and 3860 ->
3568 on arm64. Image slots fall 17 -> 16, code 6575 -> 6281 / 7260 -> 6972.

The report retains regressions instead of averaging them away. Go native
indirect calls measure 1.077 -> 1.323 ns/source iteration; LLGo native memory
0.752 -> 0.783 and f32 SIMD 3.882 -> 3.944. LLGo interpreter image filtering
measures 364.717 -> 376.356 microseconds/pass versus wazero 352.943, and direct
calls 143.485 -> 146.909 ns/iteration versus wazero 91.869. Tiny call controls
remain unresolved. Go interpreter image/CRC improve 257.457 -> 247.429 and
89.928 -> 85.182 microseconds, while LLGo CRC measures 80.042 -> 80.826.

Preparation and first call remain separate. Native CRC preparation process
medians are Go 384208 -> 381584 ns and LLGo 912875 -> 850125 ns; first calls
163875 -> 112875 / 159625 -> 105875 ns. Image preparation is Go 442125 ->
535375 ns, an observed regression, and LLGo 1310500 -> 1125500 ns. No startup,
preparation or API boundary median is subtracted from warmed calls.

The initial report remains under `artifacts/local-slots`, raw SHA-256
`87f9bf014665017b95b51fad276515177cb8c667b10e2f9340a9f1f597184d99`.
Final raw/source/binary provenance and code/frame shapes remain under
`artifacts/literal-slots`, raw SHA-256
`2f80425ac39e1af9e8b3e7ab3dd91eab1424a3f4bf29ed7f98d839cfa8c78df4`.
These mixed local diagnostics do not accept parity, resolve PR #26's ARM A/A
failure, or close M1/M2. Core 3.0/GC/Threads/components/WASI 1/2/3 remain open.

## Last-use integer operand storage

Engine `342676e65ca3ead426150a41e7890d3819a23004` releases consumed operands
before allocating results for all 58 nontrapping scalar integer primitives,
including wrap/extend/sign-extension. Locals, cross-block values and live
literal aliases stay protected; calls retain distinct suspended input/output
storage. Left-first reuse suits two-address targets. x86 directly selects
right-aliased commutative operations and captures shift counts before replacing
an aliased destination; subtraction keeps its scratch fallback.
The shared layout changes no instruction, fuel boundary or continuation ABI.

Tests cover all 58 operations, typed results, repeated SSA arguments, still-live
literal aliases, both operand aliases, arbitrary cold entries and the existing
full corpus/lifecycle matrix. Complete local Go/LLGo release suites pass.
[Final CI](https://github.com/cpunion/wasmvm/actions/runs/40656577132) passes all
four jobs; all 12 downloaded execution reports retain 148 scripts and 51,139
binary assertions. Linux coverage is slots 99.61%, codegen 93.94%, execute
98.72%, nativeexec 94.09%; every measured executable package remains above 90%.

The exact measured base is `ff0e525` (engine `4f531dd`), head `342676e`, with
unchanged compiler/reference pins and release flags. The same twelve families
and 20 API/dispatch/reference configurations pass 960 alternate-input checks.
Three interleaved process blocks, seven calls and two warmups remain local
diagnostics. Public LLGo native medians:

| Family | Unit | Base | Head | wazero | Wasmtime | Wasmer |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Integer | ns/iteration | 1.644 | 1.660 | 1.827 | 1.580 | 1.620 |
| Direct calls | ns/iteration | 0.915 | 0.880 | 1.597 | 0.930 | 0.880 |
| Memory | ns/iteration | 0.710 | 0.675 | 1.078 | 0.540 | 0.660 |
| f32 SIMD | ns/iteration | 3.747 | 3.651 | 4.234 | 3.640 | 3.640 |
| C image filter | microseconds/pass | 2.877 | 2.896 | 3.393 | 1.150 | 1.200 |
| Rust CRC | microseconds/packet | 1.732 | 1.682 | 1.807 | 1.578 | 1.672 |

Matching rounded call medians and small differences are not a controlled parity
decision. Image filtering still trails Wasmtime/Wasmer substantially; integer
and several other rows still regress. LLGo unsigned-f64 conversion measures
9.246 -> 9.581 ns/iteration, Go interpreter shuffle 104.865 -> 109.970.
LLGo interpreter image filtering measures 342.657 -> 339.222 microseconds/pass
versus wazero 354.118; calls remain slower, 120.548 -> 119.932 ns/iteration
versus 80.216. Interpreter CRC measures 77.919 -> 78.490 microseconds/packet.

Preparation remains separate: CRC Go 342917 -> 323041 ns, LLGo 924083 ->
894875 ns; image Go 434292 -> 429625 ns and LLGo 1108083 -> 1252750 ns, an
observed regression. CRC first calls measure 102667 -> 102750 / 101083 ->
92250 ns. No phase is subtracted from warmed execution.

CRC frames fall 15 -> 14 slots, but static unmetered code grows 3803 -> 3813
bytes on amd64 and 3568 -> 3572 on arm64. Smaller frames alone do not establish
faster code. A separate probe uses the image module's actual two-page minimum:
amd64 code falls 6295 -> 6230 bytes; arm64 remains 6984. It removes none of 13
static memory access checks and retains three global helper exits. Guarded
address groups with a checked cold path merit investigation; this own-code
finding does not establish why another engine is faster.

Source/binary provenance, phase records, configured code hashes and shapes
remain under `artifacts/operand-register`, raw SHA-256
`f308bbf71cc6287bd0c017d5e20292b8d4605c5671fb1116b3507d6ba33845a5`.
The preliminary slot-only builds/shapes remain under `artifacts/operand-slots`;
their x86 alias fallback was refined before the final warm comparison. The
existing compiler, competitor and PR #26 calibration gates remain open.

## Guarded memory regions and address folding

Engine `3a9f291089dbee4ec07b03e6bdb839190232a565` implements the
[guarded-memory scope](../benchmarks/native/guarded-memory-plan.md): versioned
affine address groups, runtime span guards, original checked cold fallbacks,
address folding, bounded dead temporary removal and emitted-operation register
weights. Zero-minimum modules, unmatched patterns and the precise tier keep
their existing checks. Original instructions, fuel, PCs and frame contracts
remain unchanged. Memory operations are neither removed nor reordered.

Tests exercise every scalar/SIMD memory opcode and lane, exact span ends,
unaligned accesses, large offsets, i32 wrap, stale local versions, CFG/helper
boundaries, first/interior cold PCs, canonical frame contents, earlier effects
before a later trap, arithmetic-trap order, growth, bindings, limits and bounded
cancellation. A constant-vector-store regression first reproduces an old fusion
using a pre-remapping address; access boundaries fix it. The failing log remains
under `artifacts/guarded-memory/vector-regression.log`.

Complete local Go and LLGo release suites pass. Each of six execution reports
retains 148 scripts and 51,139 binary assertions. Local codegen coverage is
97.48%, execute 98.72%, nativeexec 94.09%; every measured executable package
exceeds 90%. The final
[CI](https://github.com/cpunion/wasmvm/actions/runs/40672341762) passes all four
Go/LLGo release x Linux amd64/macOS arm64 jobs. All 12 downloaded execution
reports retain 148 scripts and 51,139 binary assertions each. Linux codegen
coverage is 94.39%; every measured executable package remains above 90%.

Exact native comparisons retain base `342676e`, guard-only `5073a99`, folded
`0e0318d` and final `3a9f291`. Compiler/reference pins and release flags are
unchanged. The final run covers twelve families, thirteen configurations and
624 alternate-input checks; intermediate rows cover native API execution.
Three interleaved process blocks retain seven calls after two warmups. Unchanged
interpreter timing is reused; its complete semantic matrix still runs.
Final public native medians:

| Application / host | Base | Head | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| C image / Go, microseconds/pass | 2.782 | 2.097 | 2.782 | 1.075 | 1.125 |
| C image / LLGo, microseconds/pass | 2.763 | 2.148 | 2.782 | 1.075 | 1.125 |
| Rust CRC / Go, microseconds/packet | 1.623 | 1.648 | 1.507 | 1.531 | 1.516 |
| Rust CRC / LLGo, microseconds/packet | 1.601 | 1.619 | 1.507 | 1.531 | 1.516 |

Image filtering improves about 25% under Go and 22% under LLGo, but still trails
Wasmtime/Wasmer. Thirteen explicit access checks become two span guards on the
normal image path; cold checks remain. Fourteen temporary constant/add writes
disappear. Complete function bodies, including entry/exit and cold paths, grow
6230 -> 7389 bytes on amd64 and 6984 -> 7616 on arm64. CRC machine-code hashes
remain identical across these engine revisions; its observed warm differences
are not a code-generation improvement.

Mixed regressions remain visible: LLGo native shuffle 1.611 -> 1.653,
f64 SIMD 4.082 -> 4.226 and direct calls 0.825 -> 0.848 ns/source iteration.
LLGo memory measures 0.626 -> 0.606 versus Wasmtime 0.500; small or rounded
differences do not establish controlled parity. Native image preparation is
Go 459250 -> 469166 ns, LLGo 1127834 -> 1051208 ns; first calls 110709 ->
79250 / 113083 -> 82667 ns. CRC preparation is 340750 -> 342458 /
825583 -> 830166 ns. No startup/preparation/boundary median is subtracted.

A separate disassembly probe uses the same Wasmtime 48.0.1 C library and
Cranelift/speed configuration as the reference adapter. Its image guest
function occupies 508 bytes; the inner span at 0x110..0x1a0 has 37 instructions,
nine byte loads, no per-access bounds comparisons and no stack spills. It is
not a normalized size comparison with wasmvm's ABI/cold-path body. The adapter
does not enable fuel/epoch interruption. The dump is evidence for further
loop-range, register-allocation and policy analysis, not a measurement assigning
the remaining speed difference to one cause. Retain assembly and library,
input/object hashes under `artifacts/guarded-weights`.

Raw reports remain under `artifacts/guarded-memory` (guard-only),
`artifacts/guarded-fold` and `artifacts/guarded-weights`, SHA-256 respectively:
`e64393eaacbbdf779b5632ee50d857e762147d11561ead048abc696173ed102e`,
`a7790bbc5290eaf7e03a98f0d65f53724503ed3474a83751bee533939f67e12b`,
`ae8025b6aae2ef0051a71bf339ed4498f122ee3107a36ae4a71826bb6206a4f1`.
The later refinements do not consistently improve every row; no intermediate
failure/regression is discarded. These diagnostics do not accept parity,
resolve PR #26's ARM A/A failure, or close M1/M2 or later milestones.

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

## Shared AArch64 caller register capacity

Engine `9ef5a125700f546e5f441d49939f5915224fad95` implements the
[caller-register scope](../benchmarks/native/caller-registers-plan.md).
Mutable integer/reference slots use x19-x28 then x1-x8. Immutable constants
and runtime metadata share only the remaining caller registers; leaf expansion
uses the common inventory without clobbering immutable words. ABI reservations,
separate vector allocation, canonical exit spills and entry reloads remain.
This expands the fixed allocator; it does not implement global live ranges.

Tests cover exhausted and partial capacity, direct/indirect expansion, live
metadata mutation, literal/spill state, 24 live integer/reference slots at large
frame offsets and every original continuation. The pressure fallback fixture
now exceeds 18 registers. Full local Go/LLGo release suites and vet pass. All six
execution reports retain 148 scripts and 51,139 binary assertions each; local
codegen coverage is 97.44%, execute 98.72%, nativeexec 94.09%, with every measured
executable package above 90%. Final
[CI](https://github.com/cpunion/wasmvm/actions/runs/40678438843) passes all four
Go/LLGo release x Linux amd64/macOS arm64 jobs. All 12 downloaded execution
reports retain those counts. Linux codegen coverage is 94.49%; every measured
executable package remains above 90%.

Exact source base `3a9f291` and head `9ef5a12` use unchanged compiler/reference
pins and release flags. Twelve native API/dispatch families across eleven
configurations pass 528 alternate-input checks. Three interleaved process
blocks retain seven calls after two warmups. These are local diagnostics;
startup, compilation/validation, instantiation, first call, warm calls and
cleanup remain separate. Public native medians:

| Workload / compiler | Unit | Base | Head | wazero | Wasmtime | Wasmer |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| C image / Go | microseconds/pass | 2.094 | 1.611 | 2.958 | 1.075 | 1.175 |
| C image / LLGo | microseconds/pass | 2.138 | 1.643 | 2.958 | 1.075 | 1.175 |
| Rust CRC / Go | microseconds/packet | 1.585 | 1.607 | 1.484 | 1.500 | 1.547 |
| Rust CRC / LLGo | microseconds/packet | 1.651 | 1.630 | 1.484 | 1.500 | 1.547 |
| f32 SIMD / LLGo | ns/iteration | 3.729 | 3.782 | 3.750 | 3.650 | 3.710 |
| Memory / LLGo | ns/iteration | 0.635 | 0.619 | 0.859 | 0.520 | 0.600 |

Image filtering improves about 23% on both compilers but still trails
Wasmtime/Wasmer. The image inner backedge span falls from 63 instructions with
two frame accesses (0x7d8..0x8d0) to 56 with none (0x874..0x950). The same nine
byte loads remain; the live accumulator now stays in x1 and its rotation uses
an immediate. Existing span checks and interrupt polling remain. The pinned
Cranelift span has 37 instructions. This is code-generation evidence, not a
controlled attribution of every timing difference or a safety-policy match.

Larger exit/cold code remains a cost: complete arm64 image bodies grow
7616 -> 9524 bytes, CRC 3572 -> 4096. All twelve amd64 guest-code hashes and the
ten non-application arm64 hashes remain identical; their observed warm changes
must not be attributed to different guest instructions. CRC Go regresses,
while its small LLGo gain does not establish competitor parity. Unchanged
interpreter timing is not repeated; complete interpreter semantics still run.

Compilation/validation preparation remains separate: image Go 499625 -> 499084 ns,
LLGo 1018375 -> 1153541 ns, an observed regression. Image first calls measure
83250 -> 61292 / 82916 -> 62208 ns. CRC preparation is Go 338292 -> 346583 ns,
LLGo 1116542 -> 817791 ns; first calls 105042 -> 102084 / 100334 -> 100959 ns.
No phase median is subtracted from warm execution. Source/binary provenance,
all phase records, code hashes and disassembly remain under
`artifacts/caller-registers`, raw SHA-256
`7ab4bfc8d6ccb48e34c78a11bdad130fed4f0969acbb806e65b969faf299632b`.
The competitor, all-target calibration and PR #26 gates remain open. This
branch is outside that PR's unchanged head; M1/M2 and all later stages remain
unaccepted.

A separate Wasmtime 48.0.1 policy-cost probe retains the original comparator and
rebuilds paired epoch-disabled/enabled C adapters with the same SDK and C flags.
Epoch-enabled code sets a one-tick deadline without incrementing the engine
during timing; a separate pre-expired-deadline probe traps with `interrupt`.
Thirty-two alternate-input checks pass. This proves the configuration is active,
not asynchronous cancellation latency or policy equivalence. Its image inner
span grows 37 -> 42 instructions, adding three epoch-check instructions and two
stack accesses. It retains nine byte loads. Codegen and check placement change
with the policy; the original unchecked-policy assembly is not the complete
interruptible comparator.

Timing remains inconclusive: the fresh three-block run observes LLGo image
2.194 microseconds/pass, original Wasmtime 1.450, rebuilt epoch-off 1.550 and
epoch-on 1.450. CRC measures 2.411, 2.297, 2.234 and 2.172 microseconds/packet.
The recorded host load is approximately 20/37/28, and the unchanged original
runners shift substantially from the earlier run. Do not infer zero/negative
interruption cost, substitute these values for the original comparison, or
accept parity. Retain SDK/source/binary hashes, configuration, phase records,
expired-deadline output and disassembly under `artifacts/caller-registers/epoch`,
raw SHA-256 `a6c9ae6b0e112db31c362603a4af8db42c2d9a0d80e5bdc543ec50c063c0a624`.

## Compact span guards and shifted operands

Engine `102623b24e9d12cae96e5c934f5abdb1afd01a82` implements the
[compact-guard scope](../benchmarks/native/compact-guards-plan.md). AArch64
normalizes spans up to 4096 bytes around their root, combining unsigned lower
and upper checks before testing a widened end against bound memory length.
The original mutable operand and checked fallback remain intact, including
wrapping addresses and an exact 4 GiB endpoint. Larger spans and amd64 retain
the general guard lowering.

Physical-state scans now cross already-proven accesses independently of fusion
boundaries. Every later guard, possible trap, poll, CFG edge, helper and region
end remains an observation boundary. This permits dead immediate removal and
AArch64 shifted ADD/SUB/AND/OR/XOR operands; rotated forms remain limited to
logical operations. Live intermediates/constants, aliases and original cold PCs
keep their full state. Precise fuel and all existing safety policies remain.

Tests execute cached/spilled guard operands, large raw high bits, normalized
underflow/overflow, immediate encoding limits, exact memory32 endpoints,
interior/failed cold continuations and later-guard frame observations. Scalar
fusion tests cover both widths, masked counts, aliases, live constants and
rejected trap/poll/region/target cases. Full local Go and LLGo release suites
and vet pass. All six execution reports retain 148 scripts and 51,139 binary
assertions each. Local codegen coverage is 97.49%, execute 98.72%, nativeexec
94.09%; every measured executable package exceeds 90%. Linux codegen coverage
is 94.44%, with every measured executable package above 90%. Final
[four-job CI](https://github.com/cpunion/wasmvm/actions/runs/40686721435) passes
Go/LLGo release x Linux amd64/macOS arm64. All 12 downloaded execution reports
retain those script and binary-assertion counts.

Exact base `9ef5a12` and head `102623b` retain compiler/reference pins and release
flags. Twelve native API/dispatch families across eleven configurations pass
528 alternate-input checks. Three interleaved blocks, seven calls and two
warmups remain local diagnostics under host load approximately 18/19/24.
No CPU-heavy task from this work runs during timing; unrelated host work is not
controlled. All preparation/first/warm/cleanup phases remain separate.
Public native medians:

| Workload / compiler | Unit | Base | Head | wazero | Wasmtime | Wasmer |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| C image / Go | microseconds/pass | 1.671 | 1.439 | 2.904 | 1.100 | 1.125 |
| C image / LLGo | microseconds/pass | 1.658 | 1.482 | 2.904 | 1.100 | 1.125 |
| Rust CRC / LLGo | microseconds/packet | 1.697 | 1.629 | 1.619 | 1.484 | 1.562 |
| f64 SIMD / LLGo | ns/iteration | 4.413 | 4.554 | 4.418 | 4.400 | 4.380 |
| Memory / LLGo | ns/iteration | 0.642 | 0.666 | 0.927 | 0.560 | 0.640 |
| Signed f64 conversion / LLGo | ns/iteration | 10.117 | 10.280 | 21.027 | 10.130 | 10.210 |

Image filtering improves about 14% under Go and 11% under LLGo locally, while
remaining slower than Wasmtime/Wasmer. Its configured arm64 inner span falls
56 -> 48 instructions (0x874..0x950 -> 0x860..0x91c), retaining nine byte loads
and no frame accesses. Its hot span guard falls ten -> six instructions;
observation-aware literals and shifted operands account for the other four
removed instructions. The precise tier and cold generic entries remain.
Complete image bodies shrink 7389 -> 7371 bytes on amd64, 9524 -> 9472 on arm64.

Configured static code hashes change in image, f64 SIMD and both scalar-convert
families on arm64, and only image on amd64. The f64 SIMD body grows 2688 -> 2692
bytes: ORR with a shifted operand reduces the post-loop sequence, while its
extra original-PC continuation grows cold code. Its loop is not an inferred
speedup. CRC hashes remain identical; its warm difference is not different guest
instructions. Regressions above remain visible and no row proves controlled
competitor parity. These shape probes configure direct leaf candidates;
indirect expansion and host/compiler boundary costs remain measured by the
actual runtime runners rather than inferred from those static probes.

Image compile/validate preparation is Go 551584 -> 512792 ns, LLGo 1035458 ->
1120458 ns, an observed LLGo regression. First calls are 64459 -> 54125 /
64584 -> 56417 ns. CRC preparation is Go 375250 -> 345708 ns, LLGo 822250 ->
813958 ns; first calls 109167 -> 101458 / 101458 -> 105959 ns. No phase median
is subtracted. Source/binary provenance, full phase records, hashes and assembly
remain under `artifacts/compact-guards`, raw SHA-256
`c953045c95729be64a9b1247bd6043fdceb3049e195050e9de067ce5063f9e81`.
Existing compiler, competitor and PR #26 calibration gates remain open. The
branch remains outside that PR's unchanged head; neither M1/M2 nor the later
Core 3.0/Threads/components/WASI stages are accepted by this increment.

## Remaining-span guards for bounded loops

Engine `4304f73fb52e2e946e6f2b8877d90e4b87d7ee4c` implements the
[bounded-loop scope](../benchmarks/native/loop-guards-plan.md). One-block i32 +1
loops with a literal exit bound can share one proof for the remaining accesses.
The root must advance with the induction variable, its base stay invariant and
the update follow every access. Internal entries, edge moves, helpers, growth,
possible arithmetic traps, multiple guard groups, nonlinear recurrences and
long prefixes retain ordinary checked code. Analysis stays within 256 operations.

At the original header, the canonical prefix precedes a widened remaining-span
check. Failed checks continue at the exact first checked access, preserving
prefix effects, wrapping behavior and the actual first trap. Successful entries
share a hot body and a private prefix/backedge without repeated span checks.
Original dispatch PCs and cold conditions never inherit that proof. Backedges
retain polling, and driver resumption or memory rebinding always rechecks.
Warm-only temporary deletion still preserves every physical exit slot.

Tests compare all frame words, complete native states and memory effects against
ordinary checked execution from every original PC. Cases include first-access
headers, both branch polarities, commuted comparisons, non-adjacent exits,
future-iteration out-of-bounds effects, i32 wrapping, memory rebinding,
cancellation, exact code limits and rejected recurrences/control. Separate guard
probes execute cached/spilled operand combinations, near-uint32-limit counters,
large biases and exact 4 GiB endpoints without allocating multi-GiB memory.
A test-only follow-up `877c6d8` adds 97 frontend-validated scalar/SIMD width/lane
loop cases, comparing successful private backedges, future-iteration bounds
failures, wrapped addresses, complete frames and memory effects with ordinary
checked execution. Production sources remain identical to `4304f73`.
The first overflow-rejection fixture described a legal 4 GiB span; its failed
assertion is retained in `artifacts/loop-guards/loop-go.log`, and the rejection
case now exceeds memory32. This was a fixture correction, not a hidden engine
failure. Both valid and invalid endpoints remain tested.

Full local Go/LLGo release semantics and vet pass. All six execution reports retain
148 scripts and 51,139 binary assertions each. Codegen coverage is 97.51%, execute
98.72%, nativeexec 94.09%; every measured executable package exceeds 90%. Final
Linux codegen coverage is 94.60%, with all 19 measured packages above 90%. The
[engine CI](https://github.com/cpunion/wasmvm/actions/runs/40698445626) passes all
four combinations. The final
[typed-loop CI](https://github.com/cpunion/wasmvm/actions/runs/40705521756) passes
Go/LLGo release x Linux amd64/macOS arm64. All 12 downloaded reports retain the
script and binary-assertion counts. It reruns semantics for the added tests,
without repeating benchmark timing.

Exact source base `102623b` and head `4304f73` retain compiler/reference pins and
release flags. Twelve native API/dispatch families, eleven configurations and
528 alternate-input checks pass. Three interleaved process blocks, seven calls
and two warmups remain local diagnostics; host load is approximately 7/8/8.
No CPU-heavy task from this work runs during timing. Every preparation, first,
warm and cleanup phase remains separate. Public native medians:

| Workload / compiler | Unit | Base | Head | wazero | Wasmtime | Wasmer |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| C image / Go | microseconds/pass | 1.405 | 1.249 | 2.817 | 1.100 | 1.100 |
| C image / LLGo | microseconds/pass | 1.443 | 1.330 | 2.817 | 1.100 | 1.100 |
| Rust CRC / Go | microseconds/packet | 1.615 | 1.639 | 1.483 | 1.516 | 1.531 |
| Rust CRC / LLGo | microseconds/packet | 1.744 | 1.667 | 1.483 | 1.516 | 1.531 |
| Indirect calls / LLGo | ns/iteration | 0.843 | 0.926 | 2.063 | 1.100 | 0.830 |
| Memory / LLGo | ns/iteration | 0.597 | 0.609 | 0.858 | 0.500 | 0.580 |
| Integer / LLGo | ns/iteration | 1.639 | 1.654 | 1.535 | 1.580 | 1.590 |

Image warm time improves about 11% under Go and 8% under LLGo, but still trails
Wasmtime/Wasmer. Its arm64 hot backedge span falls 48 -> 41 instructions
(0x860..0x91c -> 0x8ac..0x94c), retaining nine byte loads, no frame accesses and
interrupt polling. One loop-range check replaces the repeated inner span check;
Warm-only elisions increase 14 -> 16 logical temporary operations. The complete
bodies grow 7371 -> 7562 bytes on amd64 and 9472 -> 9520 on arm64. Code size and
first-entry work remain costs, not hidden by warm throughput.

All eleven other configured static code hashes remain identical on both targets;
their timing differences are not different guest instructions. This includes CRC,
whose small LLGo gain and Go regression do not establish parity. Shape probes
configure direct leaf candidates; actual runtime runners measure indirect
expansion and host/compiler boundaries. Interpreter timing is unchanged and not
repeated; its full semantic matrix still runs. Small or rounded medians, and the
tiny-call control, are not controlled acceptance.

Image compilation/validation preparation is Go 478333 -> 503541 ns (regression),
LLGo 1004541 -> 958333 ns. First calls are 54250 -> 48167 / 56041 -> 51459 ns.
CRC preparation is Go 333667 -> 331250 ns, LLGo 898208 -> 801291 ns; first calls
100959 -> 101375 / 102625 -> 103458 ns. No phase median is subtracted. Exact
source/binary metadata, full phase records, interval probes, code hashes and
disassembly remain under `artifacts/loop-guards`, raw SHA-256
`75da03c922b2419b58474ad432ac3e49437ab6b62d5e915a36e1287c27faf825`.
Existing competitor and PR #26 calibration gates remain open. This branch stays
outside that PR's unchanged head. M1/M2 acceptance, global allocation, broader
loop analysis and subsequent Core 3.0/GC/Threads/components/WASI stages remain open.

## Direct numeric global cells

Scope is [direct numeric globals](../benchmarks/native/globals-plan.md), from
baseline `ccbcbd5868de49651402737ccef0e4e9cc8b13e5` (production engine
`4304f73fb52e2e946e6f2b8877d90e4b87d7ee4c`). Implementation
`9ad273d7b7e76026ee4fa3c997de958c78fe394b` adds typed numeric lowering;
`a26c57cebe8bc67734304e86c781d9b7fa020f65` selects allocated integer/FP registers
directly, removes canonical zero-word stores and avoids hash-table construction
for small binding sets. Both amd64/arm64 emitters and precise/unmetered entries
support i32/i64/f32/f64/v128. References retain the helper/root protocol.

Execution-owned native cells deduplicate imported aliases by Global identity.
The driver snapshots only referenced numeric globals and flushes after every
native exit, before host/helper handling, traps, fuel or cancellation. Pointer
caches never cache mutable values. Suspend/close clear Go binding roots, and a
function without globals cannot flush a prior function's stale cells. No Go
address crosses the ABI. Unmetered scalar cells follow the canonical zero-word
contract; precise entries retain exact two-word copies.

Tests cover all numeric types and NaN/vector bits, every original continuation,
missing views, reference/invalid-assumption fallbacks, large displacements, code
limits, integer/FP/vector register pressure and spills, small/large alias sets,
rebinding/growth, root cleanup, host observation, same/cross-instance reentry,
collection, and writes before trap/fuel/cancellation. Full local Go/LLGo release
semantics and vet pass at the final engine. Each of six reports retains 148
scripts and 51,139 binary assertions. Local coverage: codegen 97.66%, execute
98.72%, nativeexec 95.24%, public VM 96.42%; all 19 measured executable packages
exceed 90%.

The [final CI](https://github.com/cpunion/wasmvm/actions/runs/40718959222)
passes Go/LLGo release x Linux amd64/macOS arm64 at `a26c57c`. All twelve
downloaded execution reports pass with the same script/assertion counts. Linux
coverage is codegen 95.40%, execute 98.72%, nativeexec 95.24%, with every measured
package above 90%. The earlier `9ad273d` run passed both Go jobs, then its
remaining LLGo jobs were canceled in favor of the optimized source; it is not
claimed as a completed four-job matrix. Documentation-only follow-up changes
leave the tested production source unchanged.

Performance retains Go 1.27.3 and LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` with `-O3 -lto=full`. The exact
`4304f73` baseline binaries and source archives are hash-verified before reuse.
References remain wazero 1.12.0, Wasmtime 48.0.1 / Cranelift speed and Wasmer
7.3.0 / Cranelift, with the existing policy distinctions unchanged. Both first
and optimized candidates pass 792 alternate-input oracle checks across eighteen
families and eleven API/dispatch/reference configurations. Five owned numeric
global loops use wasm-tools 1.258.0 and independent integer oracles, with source
and binary hashes. Existing twelve families and mixed globals/tables remain.

Three interleaved process blocks, seven calls and two warmups are local ARM64
diagnostics only. Initial host load is about 27/18/14, optimized load about
10/12/12. No CPU-heavy task from this work runs during timing. Startup,
compilation, instantiation, first call, warm execution and cleanup stay separate.
Primary public API medians (numeric/mixed globals: 10,000 iterations):

| Workload / compiler | Unit | Base | Head | wazero | Wasmtime | Wasmer |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Global i32 / Go | ns/iteration | 230.383 | 2.704 | 1.550 | 1.600 | 2.100 |
| Global i32 / LLGo | ns/iteration | 744.796 | 1.633 | 1.550 | 1.600 | 2.100 |
| Global i64 / LLGo | ns/iteration | 742.587 | 1.533 | 1.704 | 1.500 | 1.900 |
| Global f32 / LLGo | ns/iteration | 757.654 | 1.975 | 1.663 | 3.000 | 3.400 |
| Global f64 / LLGo | ns/iteration | 747.433 | 2.000 | 1.575 | 3.000 | 3.300 |
| Global v128 / LLGo | ns/iteration | 740.250 | 1.700 | 1.779 | 2.600 | 3.200 |
| Globals/tables / LLGo | ns/iteration | 1865.475 | 1140.475 | 34.592 | 21.700 | 6.100 |
| C image / Go | microseconds/pass | 1.313 | 1.372 | 3.011 | 1.150 | 1.175 |
| C image / LLGo | microseconds/pass | 1.278 | 1.299 | 3.011 | 1.150 | 1.175 |
| Rust CRC / LLGo | microseconds/packet | 1.714 | 1.818 | 1.682 | 1.563 | 1.672 |

The first candidate's f32/f64 LLGo API medians were 5.067/5.133 ns/iteration;
its image API regressed 1.380 -> 1.439 microseconds/pass. That entire slower
comparison remains retained. Cross-run differences are not calibrated speedup
estimates. Optimized LLGo regressions also remain: integer 1.623 -> 1.649,
function calls 0.848 -> 0.861, memory 0.600 -> 0.630, unsigned conversion
10.400 -> 10.602 ns/iteration, plus the image/CRC rows above. CRC generated
code remains byte-identical; this does not dismiss its measured regression.

Primary LLGo dispatch-only global medians are 0.758--1.200 ns/iteration versus
1.533--2.000 through the public API. An additional component-scaling diagnostic
retains both 10,000 and 100,000 global iterations and 40/400 image passes; it
compares the same head and references, without replacing the fixed-count base/
head results. At 100,000 iterations, LLGo API/dispatch f32 is 0.763/0.732,
f64 0.786/0.744 and v128 0.622/0.548 ns/iteration, while wazero is
0.672/0.673/0.525. Absolute results shift between runs. Fixed call costs matter,
but isolated scaling still retains real competitor gaps; it does not accept a
backend or prove a causal allocation-cost attribution.

Image ARM64 code shrinks 9,520 -> 9,092 bytes with all three global helper exits
removed; the bounded loop guard remains. Pure f32/f64 now use direct FP
loads/stores without loop FP/GPR bit transfers. Mixed globals/tables still has
three table/reference helper exits and a large competitor gap. Direct checked
[table access](../benchmarks/native/tables-plan.md) is the next affected runtime/compiler scope. Global live-range
allocation, API/runtime costs and all required controlled target comparisons
remain open. PR #26 remains the single active draft at its unchanged head;
this branch work does not mark M1/M2 or Core 3.0/Threads/components/WASI complete.

Full local records are `artifacts/globals` and `artifacts/globals-optimized`:
source/build identities, reports, coverage, full phase observations, sample
orders, generated-code hashes and disassembly. Initial result SHA-256:
`487e7f72c5214f881d537a8560161fc70b01a1cc554d605d29a034d50a65124f`;
optimized result:
`e98b1b9305e13e8fc2a750886b184f3f63821eabcad47d237bf3c5fd17ffc754`;
scaling result:
`bc78190d45bd97416c445c16e1e51797e6de189242a7e45be850788a87385cf3`.
All-target parity, calibration, API/mixed-state costs and broader portfolio
acceptance remain blocked. No sustained timing suite is repeated for unchanged
kernels and these diagnostics do not pass the merge gates.

## Direct checked tables and function references

Scope is [table/reference lowering](../benchmarks/native/tables-plan.md), with
baseline engine `a26c57cebe8bc67734304e86c781d9b7fa020f65`. Engine
`dd79a743f9897104353514f3a32c6d1b236a0a5e` lowers table.get/table.set/ref.func
on amd64/arm64 in precise and unmetered entries. Engine
`c921b5de933aea8d5a6724ec3eff106c2525f0bd` directly accesses allocated reference
registers and shares the element encoder with guarded indirect calls. ARM64
uses extended comparisons and UXTW addressing to truncate i32 indices without
changing raw home values. Precise and spilled paths retain checked semantics.

Both tiers bind the execution-owned function/table metadata independently of
optional call expansion. Function references load full instance-specific handles,
including imported functions. Table writes use Store-owned native cells directly;
imported aliases and root visitors see the same storage. Bounds precede address
formation, with exact source/trap ordering. Empty native tables trap safely;
Go-backed, nil and retired tables are rejected before native access, including
empty Go tables. Missing views fail before any guest instruction. Growth and bulk
helpers still exit and rebind. Every exit publishes the existing reference map.

Tests cover both reference types, null/full-width handles, every original PC,
large displacements, exact/overflowing indices, aliased tables, missing views,
invalid assumptions, code limits, index/value/result spills, imported/defined
function identities, host observation, same/cross-instance reentry, growth,
collection and cleanup, and ordered writes before trap/fuel/cancellation. A host
callback clears a table, releases its host root and collects while a native frame
retains the reference, then grows/replaces the table and updates a numeric global
before resumption. Both returned reference identities and the new global value
are checked. Initial binding-test failures assumed the return occupied the last
physical PC and that a retired instance reached table binding; these fixture
assumptions were corrected, retaining the logs rather than changing engine errors.

Full local Go/LLGo release suites and vet pass. Each local report retains 148
scripts and 51,139 binary assertions. Local coverage: codegen 97.70%, execute
98.72%, nativeexec 94.93%, public VM 96.42%; every measured package exceeds 90%.
The [final CI](https://github.com/cpunion/wasmvm/actions/runs/40766482589) passes
all four Go/LLGo release x Linux amd64/macOS arm64 combinations at `c921b5d`.
All twelve downloaded reports pass with the same script/assertion counts. Linux
codegen coverage is 95.50%, nativeexec 94.93%; all 19 executable packages exceed
90%. Only one final CI matrix is dispatched after local performance corrections.
Documentation-only follow-up commits leave the tested production source unchanged.

Performance retains Go 1.27.3, LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` with `-O3 -lto=full`, wazero
1.12.0, Wasmtime 48.0.1 / Cranelift speed and Wasmer 7.3.0 / Cranelift. The exact
baseline binaries/source archives and reference binaries are hash-verified.
Both candidates pass 924 alternate-input oracle checks across 21 families and
11 API/dispatch/reference configurations. Three owned table/reference fixtures
use wasm-tools 1.258.0, source/binary hashes and scalar oracles. Weighted indirect
calls verify reference identity after each loop, inside guest timing. All eighteen
previous families, phase separation and compiler/policy distinctions remain.

Three interleaved process blocks, seven calls and two warmups remain local ARM64
diagnostics. Host load is approximately 11/12/9 before the first run and 22/22/15
before the register run. No CPU-heavy task from this work runs during timing.
Preparation, instantiation, first call, warm calls and cleanup stay separate.
Final primary public API medians (ns/source iteration):

| Workload / compiler | Base | Head | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Globals/tables / Go | 353.383 | 3.279 | 29.387 | 21.500 | 5.900 |
| Globals/tables / LLGo | 1120.171 | 1.887 | 29.387 | 21.500 | 5.900 |
| Table roundtrip / Go | 274.460 | 1.681 | 14.949 | 10.740 | 5.210 |
| Table roundtrip / LLGo | 883.293 | 1.552 | 14.949 | 10.740 | 5.210 |
| Table copy / Go | 350.480 | 1.752 | 1.543 | 1.150 | 5.640 |
| Table copy / LLGo | 1295.402 | 1.650 | 1.543 | 1.150 | 5.640 |
| ref.func loop / Go | 348.658 | 1.467 | 30.613 | 20.600 | 5.250 |
| ref.func loop / LLGo | 1289.155 | 1.314 | 30.613 | 20.600 | 5.250 |

The first candidate retained a table-copy gap (LLGo 1.725 versus Wasmtime 1.230)
and other regressions; its full result remains. The register candidate still
trails in table copy. Other retained LLGo changes include integer 1.625 -> 1.663,
f32 globals 1.283 -> 1.308 ns/iteration and call control 4.750 -> 5.417
microseconds/call. Image API changes 1.204 -> 1.249 microseconds/pass and still
trails Wasmtime 1.075. Go call control, memory, scalar conversion and image
regressions also remain in the full report. All 102 generated-code hashes across
17 unaffected families x two targets x three tiers match baseline exactly; this
does not dismiss their observed timing changes or accept the runtime boundary.

Compilation and first-call costs are explicit. Mixed globals/tables LLGo compile
changes 0.193 -> 0.424 ms (regression), while first call falls 10.923 ms -> 15.875
microseconds. Table-copy LLGo compile is 2.253 -> 2.213 ms, first call 128.556 ms
-> 157.041 microseconds. Image LLGo compile is 1.047 -> 0.992 ms, first call
50.375 -> 50.959 microseconds. None of these phases is subtracted from warm calls.

Static diagnostics configure the same indirect leaf candidates as the runtime.
The mixed ARM64 expanded body shrinks 2,040 -> 2,000 bytes, removing its three
remaining basic table/reference helper exits. All three new fixtures remove 19
static helper sites, including their post-loop identity checks. The copy hot span
shrinks 55 -> 49 instructions but still has two frame reads and twelve runtime-
view/metadata loads. The current frequency allocator assigns cold tail homes
before metadata, exhausting the spare cache. The pinned Wasmtime serialized-code
inspection shows a 38-instruction loop span and reuses the stored reference for
the subsequent get, omitting that element reload. Its default configuration also
lacks our bounded interruption poll. These are code observations, not a claim
that any one difference explains the entire timing gap. The dump helper is
excluded from all timing comparisons; original comparator policies stay fixed.

Records in `artifacts/tables` and `artifacts/tables-registers` retain source/build
identities, all phase observations, orders, coverage, reports, exact code hashes
and both engines' disassembly. Initial result SHA-256:
`f3f8509b20480bbb2dcff0cbb238bbb7da2ba8c7e10b08f55817c03344ea2dc3`;
register result:
`dca50817531a86e8092f0c1546065e6a460306fbcdc869c1275bd107ac6a31c6`.
The [next allocation scope](../benchmarks/native/register-allocation-plan.md)
integrates loop priorities and runtime metadata with scalar homes. Store/load
forwarding, global live ranges, general calls and API costs remain distinct work.
Table-copy parity, all-target calibration and broader portfolio acceptance remain
open. PR #26 remains the single active draft at its unchanged head; no merge or
M1/M2/Core 3.0/Threads/components/WASI completion is inferred from these results.

## Loop-weighted shared register priorities

Scope is [shared register allocation](../benchmarks/native/register-allocation-plan.md),
from engine `c921b5de933aea8d5a6724ec3eff106c2525f0bd` to
`139ccf18454b4a010b8922d04ccf53040631a0cd`. Mutable scalar homes and immutable
integer literal/runtime-word candidates compete in one deterministic ranking.
ARM homes prefer x19-x28 and share x1-x8 with immutable words; other targets keep
their existing inventories. FP/vector homes retain their separate shared bank.
Cold homes no longer consume the entire inventory before metadata is considered.

Merged backward-CFG intervals supply priority hints of 8 per nesting level,
capped at 512. Shared latches merge; nested and overlapping/irreducible intervals
remain hints, not correctness proofs. Explicit edge, tuple/move and leaf-literal
work budgets lose only optional priorities when exhausted. Tuple inputs/results,
edge moves, local-copy elimination and emitted memory-region plans participate.
Literal scores measure materialization savings. No target-specific workload
identity, new feature assumption, removed bounds check or reduced poll policy
is introduced. Every cold entry reloads its selected homes/immutable words;
exit spills, reference maps and raw scalar/vector contracts remain unchanged.

Tests cover forward/self/nested/shared/overlapping backedges, saturation,
pathological shared lists, tuple/move/leaf budgets, class/register exclusion,
cold tails, deterministic bytes and code limits. A pressure fixture compares
all physical slots, table effects and terminal state against precise execution
from every original PC, including full-width handles and wrapped/invalid i32
indices. Existing reference collection/reentry, cache-pressure, spill, numerical,
SIMD, cancellation, helper/host rebinding and complete Core 2.0 suites stay active.
Full local Go/LLGo release semantics and vet pass, with six reports retaining
148 scripts and 51,139 binary assertions each. Local codegen coverage is 97.76%,
nativeexec 94.93%; all 19 measured executable packages exceed 90%.

The [final CI](https://github.com/cpunion/wasmvm/actions/runs/40774851885)
passes all four Go/LLGo release x Linux amd64/macOS arm64 combinations at
`139ccf1`. All twelve downloaded execution reports pass with the same script
and assertion counts. Linux codegen coverage is 95.59%, nativeexec 94.93%; all
19 executable packages exceed 90%. Only one final CI matrix is dispatched for
this allocator source. Documentation-only follow-up commits leave its production
source unchanged.

Performance retains Go 1.27.3 and LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` with `-O3 -lto=full`, plus wazero
1.12.0, Wasmtime 48.0.1 / Cranelift speed and Wasmer 7.3.0 / Cranelift. The
exact `c921b5d` baseline binaries and source archives are hash-verified before
reuse; candidate builds use the exact committed source and common head harness.
All 924 alternate-input oracle checks pass across 21 families and eleven own
API/dispatch/reference configurations. Three interleaved process blocks, seven
calls and two warmups remain local ARM64 diagnostics. Host load changes about
9/11/11 -> 8/10/11, with no CPU-heavy task from this work during timing. Startup,
compilation, instantiation, first calls, warmed calls and cleanup remain separate.

Primary public API medians, ns/source iteration unless otherwise labeled:

| Workload / compiler | Base | Head | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Table copy / Go | 1.649 | 1.166 | 1.331 | 1.180 | 5.690 |
| Table copy / LLGo | 1.595 | 1.123 | 1.331 | 1.180 | 5.690 |
| Table roundtrip / LLGo | 1.630 | 1.366 | 16.250 | 11.190 | 5.330 |
| Globals/tables / LLGo | 1.500 | 1.408 | 30.817 | 22.500 | 5.800 |
| ref.func loop / LLGo | 1.306 | 1.177 | 28.735 | 22.370 | 5.590 |
| Global i32 / LLGo | 1.012 | 0.842 | 0.554 | 0.900 | 0.900 |
| Global i64 / LLGo | 0.867 | 0.783 | 0.650 | 1.100 | 0.800 |
| Image / LLGo, microseconds/pass | 1.210 | 1.290 | 2.995 | 1.175 | 1.225 |
| CRC / LLGo, microseconds/packet | 1.697 | 1.663 | 1.568 | 1.609 | 1.531 |

Table-copy LLGo is about 30% faster than its own baseline in this diagnostic;
its median also falls below the three references. This does not pass controlled
parity/calibration on all required compiler/target profiles. Retained LLGo
regressions include shuffle 1.755 -> 1.874, integer 1.638 -> 1.655, unsigned
conversion 10.261 -> 10.338, f32 globals 1.363 -> 1.392 and f64 globals
1.267 -> 1.283 ns/iteration, plus the image row above. Go function-call,
indirect-call, call-control, image and v128-global regressions remain in the full
report. Candidate workload bytes are unchanged; register choices legitimately
change generated code in multiple families, so those timing changes are not
dismissed as byte-identical kernels.

Table-copy LLGo compilation changes 2.299 -> 1.981 ms and first call
162.958 -> 104.083 microseconds. Mixed globals/tables LLGo compilation changes
0.539 -> 0.379 ms, first call 14.917 -> 15.250 microseconds (regression).
Image LLGo compilation changes 1.092 -> 0.983 ms, first call
49.667 -> 49.625 microseconds. These phases are not hidden in warm throughput.

The same indirect-leaf configuration used by the runtime is retained in static
probes. The ARM64 copy hot span falls 49 -> 41 instructions, with its runtime-
view/metadata loads reduced from twelve to four. Two frame parameter reads and
the stored-reference reload remain. The complete expanded body shrinks
25,220 -> 20,996 bytes; the image body grows 9,092 -> 9,104 bytes on ARM64 and
7,326 -> 7,367 on baseline amd64. Scalar homes, metadata and constants still
have fixed lifetimes for an entire entry; this is not global live-range
splitting or table store/load forwarding.

Records in `artifacts/register-allocation` retain compiler/source identities,
coverage, all six local and twelve CI reports, all phase observations, sample
orders, code hashes and disassembly. Result SHA-256:
`65d270f91230774771af8d6e882f1c74aa088c7c51d5683dc11b2cde61df613c`.
All-target controlled performance, remaining portfolio/API/general-call costs,
global allocation and M1/M2 acceptance remain open. PR #26 is still the single
active draft at its unchanged head. Core 3.0/GC/Threads/components/WASI 1/2/3
are not completed or implied by this allocator change.

## Core Threads implementation

The feature adds all 67 atomic instructions and shared ordinary/SIMD/bulk/host
memory access on the interpreter and both native targets. Its pinned complete
applicable corpus contributes 17 scripts, 946 actually executed commands and
579 binary assertions per tier; six text-parser assertions remain separately
classified. Nested thread/wait commands execute concurrently with isolated
Stores and explicit memory capabilities, and join before cleanup.

Local macOS arm64 validation uses Go 1.27.3 and LLGo revision
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` with `-O3 -lto=full`.
Full Go and LLGo package tests, Go race tests including the entire Threads
corpus, vet, cgo-free build/adapter checks and `cgocheck2` ownership tests pass.
All 21 executable packages exceed 90% statement coverage; the lowest is
Core validation at 93.21%, with memory at 97.33% and public VM at 93.28%.
The exact-head Linux/macOS × Go/LLGo CI reports remain the integration gate.

Authored tests cover concurrent cross-Store RMW, shared growth, mixed-width
atomic alignment/bounds, Memory64, ordinary/SIMD/bulk/host access, wait32/64 queue
identity, notify counts, mismatch/timeout, cancellation, Store/capability close,
waiter limits and per-Store maximum reservations. Native shared-memory accesses
use individually checked synchronized helpers; no whole-function fallback or
raw shared-memory descriptor is used. Helper overhead, contention/scaling and
representative feature performance remain explicit follow-ups. Phase-separated
exact-base/head and pinned-competitor measurements are retained with this PR's
artifact evidence; they do not close M3 performance acceptance.

## WASI Preview 1 host implementation

The Linux/macOS adapter supplies all 46 typed imports, checked Memory32 ABI,
rights, stable directory handles, explicit sockets and independently owned P1
process resources. It executes all 72 complete pinned official programs in the
interpreter, unmetered native and precise-fuel native tiers. Reports count guest
programs, not unknown internal assertion totals. A separate Go 1.27.3 program
combines goroutines, GC, panic/recover, stdio, filesystem I/O, randomness and
timed waits on all three tiers under both host compilers.

Local Go/LLGo full package tests, Go race checks, cgo-free builds, vet and
cgocheck2 pass on macOS arm64. Every executable package exceeds 90% statement
coverage; the new host capability package measures 91.35% and P1 90.59% in the
retained full Go profile. The exact-head CI matrix remains the integration gate.
Source revisions, original producer hashes and custom-section stripping are
recorded in the P1 input manifest; original instructions/data are unchanged.

Authored cases cover symlink/rename races, no-follow and dot-dot escapes,
read-only hard-link rights, output-pointer faults before effects, descriptor and
poll-worker quotas, partial I/O, all socket operations, nonblocking/sync flags,
Store versus Process shutdown and file-stream cancellation/descriptor lifetime.
File.Fd's conversion back to blocking mode and negative Unix I/O counts exposed
by blocked-pipe tests were corrected; SyscallConn preserves controlled access.

The real Go guest exceeds the default 64 MiB aggregate native code budget. Its
rejection is retained with function/tier/budget diagnostics. The test explicitly
selects 256 MiB; the default remains unchanged. Darwin's bounded MAP_JIT virtual
reservation grows to 1 GiB to permit that explicitly budgeted image. This does
not establish compact-code or compilation-performance acceptance. Legacy
wasi-threads, threaded LLGo guests, components/WASI 0.2/0.3 and remaining
performance gates are still open.
## Component frontend and lexical type validation

The frontend/Canonical ABI foundation merged in
[PR #40](https://github.com/cpunion/wasmvm/pull/40). Its final head `0183b15`
passed the four Go/LLGo Linux-amd64/macOS-arm64 jobs in
[run 40919202259](https://github.com/cpunion/wasmvm/actions/runs/40919202259).
All 180 existing Core/proposal/P1/thread execution reports were downloaded and
checked, in addition to the new 132-command binary grammar and 1,024-case
reference-layout tests. The 16 invalid grammar cases still only passed syntax
decoding; this did not establish semantic validation or component execution.

The next increment adds bounded value/function/resource and lexical interface
type checking, Core module type groups/descriptors through the shared Core
validator, name/attribute constraints and independent instance-import resource
binders. Local full Go package tests and targeted LLGo Core/component/engine
tests pass, including full Core 2.0/3.0 regression execution. Race, cgo-free and
Linux cross builds pass. The declaration-prefix fuzzer completed 1,400,969
executions in 10 seconds after resource substitution was added, without a
failure. This count is not a full component conformance count.

Whole-binary linking, canonical declaration checking, interface subtyping,
component-instantiation import substitution, exported resource visibility and
component execution remain open. The [type design](design/component-validation.md)
records those boundaries; M5/M6 are not complete.

## Core composition and canonical definition validation

The lexical type subsystem merged in [PR #41](https://github.com/cpunion/wasmvm/pull/41)
at `ddf15ee`, after [all four Go/LLGo Linux/macOS jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40926004925).
All 180 execution reports were downloaded and verified. Its component validator
coverage was 95.83%; Core validator coverage was 93.73%.

The next ordered-checker increment adds Core module instances and shared import
matching, typed aliases, synchronous canonical signatures/options and boundary
nameability. Three complete pinned semantic scripts pass 81 commands: five valid
definitions and 76 invalid rejections. These are validation results, not guest
execution counts. Local full Go and LLGo Core/engine regression pass; subsequent
nameability changes were rechecked under Go/race and LLGo. Interface instance
construction/subtyping, value/start definitions and async remain unsupported.
The larger external-visibility script remains diagnostic, with unsupported
component-instance/ascription paths preserved as failures; it is not counted as
passing conformance. M5/M6 completion and competitor performance acceptance stay
open.


## Component instances and binder-aware interface matching

The ordered checker merged in [PR #42](https://github.com/cpunion/wasmvm/pull/42)
at `777c748`, after [all four Go/LLGo Linux/macOS jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40933962369).
All 180 Core/proposal/P1/thread execution reports were downloaded and verified.

The next increment adds component instantiation and synthesized instances,
interface subtyping/ascription, argument/resource substitution, fresh exported
resources and outer component aliases. Public name tokens stay separate from
nominal identity. Imported resources retain identity across child round trips;
representation authority stays with their defining component. All traversals,
including substitution chains, consume the compilation work budget.

The pinned offline inventory now contains 13 complete generated semantic
scripts and 461 binary commands: 100 accepted definitions and 361 invalid
rejections. Five producer text-parser assertions are recorded separately.
wasm-tools 1.258.0 cannot generate the official version-suffix attribute script;
its source and explicit producer limitation are retained. No pass is claimed.
Stream/future constructors, waitable wait/poll and indirect thread creation have
static plans only, without async execution.

Local full Go tests, Go race checks and LLGo component tests pass. The changed
validator package has 92.6% statement coverage. Bounded whole-component fuzzing
completed 816,987 executions in 10 seconds without a failure. The coverage
percentage measures this package, not Component Model conformance. Component
value/start validation, remaining async plans, actual execution and WASI 0.2/0.3
worlds remain open; M5/M6 and full performance acceptance are not complete.

Scoped local measurements on darwin/arm64 use Go 1.27.3 and LLGo `8e45ed5`
with `-O3 -lto=full`, five 300 ms samples per row. Values below are medians in
microseconds per validation. Base is `c086823`, with the identical benchmark
harness overlaid; head is this implementation. Decode and validation contain no
guest execution. Builds happen before timing, and test-process startup is
outside the Go benchmark timer.

| Fixture and phase | Go base → head | LLGo base → head |
| --- | --- | --- |
| Canonical math: validation after decode | 3.244 → 3.382 (+4.3%) | 6.406 → 6.527 (+1.9%) |
| Canonical math: owned decode + validation | 4.801 → 4.921 (+2.5%) | 8.917 → 9.120 (+2.3%) |
| Child composition: validation after decode | Unsupported → 5.422 | Unsupported → 10.171 |
| Child composition: owned decode + validation | Unsupported → 8.015 | Unsupported → 14.592 |

The small existing-fixture overhead remains a local optimization follow-up.
These samples are descriptive, not calibrated performance acceptance. Test
binary build wall times were Go base/head 0.427/0.312 s and LLGo 19.610/17.862 s;
cache state differs, so those are build diagnostics, not throughput comparisons.
LLGo allocation counters are unavailable and are not interpreted as zero.
Competitor warm component-validation comparisons remain pending; CLI compile
or process-start measurements are not substituted for that phase. Raw outputs
and the base overlay are retained under
`artifacts/components/instances-benchmark/`.

```sh
go test -run '^$' -bench 'Benchmark(CanonicalModule|ComponentComposition)Validation' -benchtime=300ms -count=5 ./internal/component/validate
llgo test -O3 -lto=full -run '^$' -bench 'Benchmark(CanonicalModule|ComponentComposition)Validation' -benchtime=300ms -count=5 ./internal/component/validate
```


## Literal values, linear consumption and start definitions

Component instances and binder-aware interface subtyping merged in
[PR #43](https://github.com/cpunion/wasmvm/pull/43) at `a472af5`, after
[all four Go/LLGo Linux/macOS jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40963539368).
All 180 execution reports were downloaded and verified.

The next increment decodes type-directed component literal values into the
existing canonical value representation, checks value imports/exports/equality
bounds and validates component start arguments/results. Unknown values use
abstract symbols; exact constants retain typed contents unless an explicit
ascription hides them. Resource-bearing or async-handle literal types are
rejected transitively, including unselected branches. Value obligations follow
moves and nested instance aliases and reject reuse or unconsumed values.
Nothing in these plans invokes a guest or host function.

The original 13 generated semantic scripts remain green. The pinned
wasm-tools 1.258.0 producer rejects value-definition text, so the new value/start
binary is authored from the pinned grammar and explicitly is not an independent
oracle. Authored tests cover all literal constructors, truncation, Unicode,
scalar width/NaN encoding, type/arity mismatch, equality-bound arguments,
ascription, nested instance moves and shared compilation/allocation limits.
Both host compilers run the same tests. Local full Go coverage passes every
existing per-package gate: binary 100.00%, canonical 94.50%, component decoder
94.26% and component validator 92.43%. Component/binary race checks pass.
Literal and whole-component fuzz runs complete without failure (199,178 and
837,726 executions respectively); these are robustness counts, not conformance.

Scoped measurements use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full` on local
arm64. Five 300 ms samples alternate base/head order with no concurrent local
builds or tests. The earlier contended samples are retained separately. Base
is `a472af5` with the identical harness overlaid, and values below are medians
in microseconds per operation; no guest execution is timed.

| Fixture and phase | Go base → head | LLGo base → head |
| --- | --- | --- |
| Canonical math: validation after decode | 3.357 → 3.385 (+0.8%) | 6.503 → 6.539 (+0.6%) |
| Canonical math: owned decode + validation | 5.081 → 4.989 | 9.089 → 9.044 |
| Child composition: validation after decode | 5.571 → 5.800 (+4.1%) | 10.084 → 10.347 (+2.6%) |
| Child composition: owned decode + validation | 8.108 → 8.330 (+2.7%) | 14.572 → 14.852 (+1.9%) |
| Value/start: validation after decode | Unsupported → 2.197 | Unsupported → 3.711 |
| Value/start: owned decode + validation | Unsupported → 3.024 | Unsupported → 5.050 |

The small composition overhead from new value obligations remains a tracked
local optimization. Test-binary build diagnostics are separate: Go base/head
0.455/0.709 s, LLGo 19.527/22.121 s, with differing caches. Process startup is
outside the benchmark timer; LLGo allocation counters are unavailable.
Competitor warm value/start validation remains unmeasured because the pinned
reference producer lacks this feature. Existing-phase competitor comparison
and calibrated performance acceptance remain open. Raw logs, source overlay
and repeat script are retained in `artifacts/components/values-benchmark/`.

Component async validation, composition execution, canonical calls/ownership
and versioned WASI 0.2/0.3 worlds remain required. This static increment does
not complete M5 or M6.


## Synchronous component execution

Literal values and start plans merged in [PR #44](https://github.com/cpunion/wasmvm/pull/44)
at `7a5f5a8` after [all four Go/LLGo jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40966221064).
All 180 Core/proposal/P1/thread execution reports were downloaded and verified.

The next increment executes ordered Core/component composition and synchronous
canonical calls on the existing interpreter/native compiler. Real execution
covers host callbacks, three string encodings, memory64, spilled parameters and
results, post-return, starts, values, outer closures and narrowed instance views.
Limits cover pure component expansion and host-only reentry as well as Core
execution. Errors, traps, cancellation and callback lifetime are tested.
[Runtime scope](design/component-runtime.md) lists unsupported capabilities;
resources, remaining async and WASI 0.2/0.3 worlds are not complete.

Local full Go tests and the exact per-package coverage gate pass. Coverage is
90.34% for `pkg/component`, 93.64% for `pkg/vm`, 94.30% for canonical primitives
and 92.45% for component validation. Go race checks and LLGo release-profile
component/Core API tests pass, including both own backends. Bounded execution
fuzzing completed 304,374 executions without failure. These are package tests
and robustness counts, not complete independent Component Model conformance.

The component numeric composition also returns `9` under Wasmtime 48.0.1.
Separate embedded measurements use its checksum-pinned C API SDK (SHA256
`9e3c636ed487a41026ff76388c5fa6f3a48ea0968408d033ed4b5e8082c20d69`).
Five local arm64 samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`.
The same two-u32 `add` component and checked changing arguments isolate tiny
call latency; this does not measure sustained instruction throughput.

| Phase (median microseconds) | Go interpreter | Go native | LLGo interpreter | LLGo native | Wasmtime C API |
| --- | ---: | ---: | ---: | ---: | ---: |
| Owned compile, excluding Engine setup | 13.033 | 25.239 | 59.277 | 172.641 | 986.000 |
| Store setup + instantiate + close | 3.167 | 3.484 | 12.795 | 15.193 | 28.000 (close excluded) |
| Public warm call | 0.436 | 0.751 | 2.643 | 4.610 | 0.346 |
| Shared Core session diagnostic | 0.334 | 0.562 | 2.499 | 4.155 | — |

Own instantiation includes cleanup; the reference's setup/instantiate phases
exclude deletion and export lookup. First call, lookup, Engine setup and file
load are retained separately in raw logs. Process startup and binary build are
outside the warm timers. The source baseline has no component execution API,
so it has no successful warm row. Allocation counters under LLGo remain
unavailable. Fuel is disabled in these descriptive samples; cancellation and
logical safety policies still differ, and calibrated acceptance remains open.

LLGo tiny-call latency is about 7.6×/13.3× Wasmtime for interpreter/native and
has not reached the target. A three-instruction call emphasizes API/adapter,
allocation and Core handoff costs; these results do not establish slower
compiled instruction throughput. Sharing the Core session removes only part
of the overhead. Keep this gap visible while execution coverage advances under
the owner capability-first decision; do not claim a performance pass.

The initial reference harness crashed with an uninitialized output value.
Explicitly initializing its output struct makes the probe succeed; the failed
source and debugger output are retained, not counted as timing results.
Raw logs, SDK provenance and phase summaries are under
`artifacts/components/runtime-benchmark/`; the independent C probe and Go/LLGo
commands are documented in [the phase probes](../benchmarks/components/README.md).


## Map and fixed-list ABI transfers

Synchronous component execution merged in [PR #45](https://github.com/cpunion/wasmvm/pull/45)
at `3091da1` after [all four semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40981892654).
All 180 existing execution reports were downloaded and verified.

The next collection increment executes fixed-list memory/flat conversions and
map transfers through the existing list/tuple representation. The entry schema
is cached per map layout; repeated calls cannot grow metadata by recreating it.
Maps preserve duplicate keys and order. Flat-slot limits, guest ranges, host
allocation budgets, aggregate values and nesting remain enforced.

Real component fixtures cover map and fixed-list callbacks through both own
Core backends, including direct parameters and spilled parameters/results.
wasm-tools 1.258.0 independently validates these generated binary fixtures.
The unchanged pinned Python Canonical ABI reference additionally supplies 56
independent memory byte cases for maps/fixed lists, with both pointer widths;
Store output and Load input are checked separately. These cases are distinct
from the existing 1,024 layout cases.
Local full Go tests, LLGo component tests and race checks pass. Statement
coverage is 93.89% for canonical primitives, 90.20% for `pkg/component` and
92.45% for component validation; all existing package gates remain green.

The `BenchmarkCollectionConversion` probe excludes layout construction, process
startup, compilation and instantiation. It measures lower+lift conversion with
one reused layout and reset scratch memory. Five 300 ms local arm64 samples use
Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`. The exact source base `3091da1`
returns unsupported for both conversions, preserved as a failed capability row.
Median lower+lift times are Go map/fixed-list **869.3/509.0 ns**, LLGo **1,584.0/948.5 ns**. No successful base throughput or pure-conversion competitor parity is claimed;
Wasmtime's complete function-call timing is not substituted for this phase.
The synchronous runtime's tiny-call competitor gap remains open.

Resource own/borrow transfer, generative host resource bindings, component/Core
module host binding and component async/WASI 0.2/0.3 worlds remain required.


## Component resource instruction execution

Map and fixed-list transfers merged in [PR #46](https://github.com/cpunion/wasmvm/pull/46)
at `1f6db34` after [all four semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40984842528).
All 180 downloaded execution reports were verified against that exact head.

The resource instruction increment executes resource.new/rep/drop using nominal
runtime declarations and independent component handle domains. Imported type
round trips retain identity; repeated instantiation freshens it. Representations
preserve i32/i64 width. Destructors execute through the active Core Caller and
consume ownership before guest code; traps cannot resurrect handles. Post-return
leave guards, a shared Store handle quota, failed instantiation and sealed
shutdown domains retain their bounded lifetime contracts.

A bounded 10-second execution fuzz run completes 496,113 cases without a panic.
Local full Go, pinned LLGo release component tests and Go race checks pass with
both own backends. Coverage is canonical **94.04%**, component runtime **90.17%**,
validation **92.45%**; every required package remains above 90%. An initial full
Go run under concurrent compiler load hit the pre-existing wasi-threads 100 ms
setup deadline. Its failing log is retained; the subsequent full run passes.
wasm-tools 1.258.0 independently validates the six new fixtures. Wasmtime
48.0.1 executes the 32-bit churn/type-round-trip probes. Its cm64-enabled compile
panics on I64/I32 assertions for the 64-bit fixture; that failed reference row
is retained rather than counted as a semantic or performance pass.

Three local 200 ms samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full` on
Apple M4 Max arm64. A warm call executes 128 guest resource lifetimes, including
the loop, new/rep/drop and the Core destructor; results are checked per batch.
Compile, setup, lookup and first call stay outside these timers. The independent
Wasmtime C API probe executes the same binary and batch. Medians in ns/lifetime:

| Probe | Go interpreter | Go native | LLGo interpreter | LLGo native | Wasmtime C API |
| --- | ---: | ---: | ---: | ---: | ---: |
| Complete guest resource lifetime | 578.8 | 1,399.0 | 1,448.0 | 3,836.0 | 29.5 |

These measurements expose a substantial helper/dispatch gap; LLGo remains about
49x/130x behind the reference on this workload. They do not establish controlled
acceptance: fuel is disabled, own cancellation/quotas remain enabled and the
reference uses its default Store policy. Source base `1f6db34` rejects resource
instructions, so it has no successful component throughput row.

A separate pure table probe removes Core dispatch, guest looping and destructors
and measures New + Rep + Drop. The exact base is extracted with git archive and
receives only the unchanged benchmark file; no implementation patch is applied.
Sequential source-base/head medians are Go **46.65 -> 62.41 ns** (+33.8%), LLGo
**181.2 -> 143.3 ns** (-20.9%). This isolates the shared quota/64-bit table change
from complete guest execution. The first contended table samples are retained
as diagnostics. The competitor exposes no equivalent public table primitive;
its complete guest timing is not substituted for pure-table performance.
LLGo allocation counters remain unavailable. Raw logs, failed reference output,
phase summary and toolchain pins are in `artifacts/components/resources-*`.

Own/borrow codec transactions, opaque host capabilities, host resource/module/
component bindings, component async and WASI 0.2/0.3 worlds remain required. This
instruction increment does not complete M5/M6 or meet the performance target.


## Opaque ownership calls and host resource bindings

Resource instructions merged in [PR #47](https://github.com/cpunion/wasmvm/pull/47)
at `bedd83b` after [all four semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40988542787).
All 180 downloaded Core/proposal/P1/thread reports were verified at that head.

Own/borrow call values now execute through the existing flat/memory codec on
both own backends. Opaque capabilities share move state across Go aliases;
borrowed callbacks expire, guest borrow escape traps, nominal host bindings
respect equality and Store authority, and aggregate/spilled transfers retain
those contracts. Host failure, panic, cancellation and unpublished input/result
cleanup attempt host destruction and release quota. Guest destruction runs only
while the defining component can safely be entered. Code close and reference
collection retain a live resource's destructor. Same-component synchronous
reentry is rejected; cross-component forwarding/local destructors remain valid,
and pure host recursion has a shared depth limit. Copied callback capabilities
share expiry/serialization state too.

Full local Go, LLGo release component/Core API tests, race and vet checks pass.
Coverage is canonical **93.04%**, component runtime **90.26%**, component validator
**92.46%**, Core embedding API **93.78%**; all package gates pass. Bounded execution
fuzzing completes **948,816** cases in ten seconds without a panic. The original
1,024 independent layout oracle cases and 56 collection-memory cases remain
present; the new opaque-reference tests use a separate filename.

wasm-tools 1.258.0 independently validates the new fixtures. The pinned Wasmtime
48.0.1 C API probe executes make/get/pass/consume, checks a guest destructor and
explicit drop, then rejects a moved capability. Its negative case runs last:
that canonical trap locks down the reference instance, while wasmvm rejects bad
host capabilities before guest effects and preserves usability. An initial
reference harness placed that rejection earlier and then failed with `cannot
enter component instance`; this was harness ordering, not a successful timing
sample. Local cm64 borrow conversion is rejected before instantiation; cm64 own
representation preservation passes both own backends. Standalone resource
values/starts remain unsupported.

Three sequential 200 ms samples per own compiler/backend use Go 1.27.3 and
LLGo `8e45ed5 -O3 -lto=full` on Apple M4 Max arm64. The source base is exact
`bedd83b`, extracted with git archive and only the unchanged benchmark/fixture
added. Both compilers reject its own/borrow feature, retained as a failed
capability row. The successful existing resource-instruction baseline measures
128 guest lifetimes per public call; the new ownership probe measures four
public calls per lifetime (make/get/pass/consume). These are distinct scopes.

| Probe (ns/lifetime, medians) | Go interpreter | Go native | LLGo interpreter | LLGo native | Wasmtime C API |
| --- | ---: | ---: | ---: | ---: | ---: |
| Guest new/rep/drop: source base -> head | 566.1 -> 572.9 | 1,363 -> 1,397 | 1,435 -> 1,434 | 3,501 -> 3,473 | 26.68 |
| Four public ownership calls | 2,923 | 5,297 | 8,656 | 14,453 | 1,201.63 |

Startup, compile, lookup and first lifetime stay outside the warm timers; every
batch/result and final destructor checksum are checked. Fuel is disabled, own
cancellation/logical quotas remain enabled, and the reference uses default Store
policy. Samples are descriptive and uncalibrated; LLGo allocation counters remain
unavailable. LLGo ownership calls remain about **7.2x/12.0x** slower than the
reference, with no performance acceptance claim. The guest-loop/helper gap also
remains open. Raw source-base outcomes, final phase logs, reference outputs and
summaries are under `artifacts/components/ownership-*`.

The performance investigation uses Wasmtime v48.0.1's primary
[component trampoline implementation](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/cranelift/src/compiler/component.rs):
new/rep use register arguments to libcalls, and drop calls compiled destructors
from the generated trampoline. wasmvm currently uses generic host callbacks and
numeric value marshalling. A separate Go allocation profile attributes about
4% of allocations to resource entries, with execution frames and marshalling
dominating. This supports optimizing the shared call boundary; it is not a
quantitative LLGo cost breakdown and does not justify workload-specific paths.

Remaining M5/M6 work includes standalone resource values/starts, local cm64
borrow ABI, host component/Core-module binding, component async and the declared
WASI 0.2.12/0.3.2 worlds. Passing these call fixtures does not complete them.


## Checked WASI environment interface

Opaque own/borrow calls merged in [PR #48](https://github.com/cpunion/wasmvm/pull/48)
at `7f89e1a` after [all four semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/40996755338).
All 180 downloaded execution reports were verified at that exact head.

The new host contract hook exposes immutable FunctionInfo/ValueType views and
checks every import after nominal resources bind, before guest initialization.
A valid component that substitutes own for a provider's borrow parameter is
rejected. Failure/cancellation before an actual Core start publishes no instance
and executes no callback. Retained views preserve nominal/aggregate information
without exposing mutable model graphs or copying each function's type graph.
Raw callbacks remain available under the original guest-declared schema contract.

`wasi:cli/environment@0.2.12` now implements all three declared operations:
get-environment, get-arguments and initial-cwd. Explicit UTF-8 configuration is
copied, limited and never inherited implicitly. Repeated/mutated host results do
not change it; duplicate environment names, empty arguments/values and Unicode
survive Canonical ABI memory/list/tuple/option transfers. The pinned WIT/archive
SHA256 and CLI license are retained in `testdata/wasi-preview2`. Independently
validated wasm-tools 1.258.0 fixtures run through both own backends.

Full local Go, pinned LLGo release, race and vet checks pass. Every executable
package coverage gate passes: component runtime **90.70%**, preview2 environment
**98.65%**, canonical **93.04%** and component validation **92.46%**. The reference
C API probe uses Wasmtime 48.0.1's actual WASIp2 provider and explicit argv/env,
checking all three complete return values outside its warm timer. No complete
upstream WASI suite/world or second guest-language toolchain is claimed.

Three sequential 200 ms own samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`
on Apple M4 Max arm64. The actual environment workload lowers three argument
strings into guest memory, counts them and resets allocation in post-return.
Compilation, binding, export lookup and first call are excluded. Wasmtime uses
the same binary/configuration and 100,000 warm calls per sample. Source base
`2d195db` has no P2 provider, so it has no successful interface throughput row.
The existing component public-call probe is a separate base/head regression check.

| Median nanoseconds | Go interpreter | Go native | LLGo interpreter | LLGo native | Wasmtime C API |
| --- | ---: | ---: | ---: | ---: | ---: |
| Actual P2 argument-count call | 2,394 | 5,284 | 6,816 | 14,709 | 438.8 |
| Existing public-call base -> head | 450.8 -> 450.3 | 752.0 -> 769.3 | 1,411 -> 1,433 | 2,419 -> 2,288 | — |

Fuel is disabled; own cancellation/logical limits remain enabled while the
reference uses its default Store policy. These are descriptive local samples,
not calibrated acceptance. LLGo is about 15.5x/33.5x behind the reference on this
workload; the target remains open. LLGo allocation counters remain unavailable.
Raw phase logs and summaries are under `artifacts/components/environment-*`.

Other CLI/HTTP interfaces, complete declared worlds, remaining component
resource values/starts/host module bindings and async/WASI 0.3 remain required.
This complete environment interface does not complete M5/M6.


## WASI exit and linked-instance shutdown

The complete `wasi:cli/exit@0.2.12` interface implements both `exit(result)` and
`exit-with-code(u8)`. Both own backends check status 0/1/42/255, typed host errors,
nonreturning Core execution, invalid schema/input, cancellation and closure of
held exports in both linked child instances. An exit in a Core start returns a
typed code with no published instance. `Caller.CloseInstance` shares the existing
callback lifetime/serialization guard; expired/copied/canceled callers cannot
acquire extra shutdown authority. Host callback outcomes survive shutdown's
cancellation. Final Store reclamation occurs before the public exit call returns.

Full Go, affected-package LLGo release, race and vet checks pass. The full
cross-package coverage profile passes every package gate: component **90.74%**,
preview2 **95.88%**. New start-exit cases additionally pass both compilers.
wasm-tools 1.258.0 independently validates the project-authored fixtures; the
pinned WIT/license remain unchanged. Wasmtime 48.0.1's actual WASIp2 C API provider
executes the same exit binary and confirms all four statuses.

Three sequential local samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full` on
Apple M4 Max arm64. Every exit consumes a fresh instance. Own samples use batches
of 256 prebound instances, excluding preparation/compile/lookup and checking all
exit results outside the timer. Wasmtime uses 10,000 exits per sample, with those
phases and final verification/destruction reported separately. Source base
`025e230` has no exit provider: both compilers reject the unchanged exit benchmark
for missing APIs, retained as unsupported rather than a throughput result.

| Median nanoseconds | Go interpreter | Go native | LLGo interpreter | LLGo native | Wasmtime C API |
| --- | ---: | ---: | ---: | ---: | ---: |
| Exit call (own includes Store closure) | 1,652 | 2,851 | 3,142 | 33,096 | 427.5 |
| Reference verification + explicit Store destruction | — | — | — | — | 457.5 |
| Existing public-call base -> head | 451.7 -> 448.2 | 775.5 -> 769.4 | 1444 -> 1575 | 2366 -> 2362 | — |

This is a one-shot exit lifetime, not recurring warm throughput. Fuel is disabled;
own cancellation/logical quotas remain enabled while the reference uses default
Store policy. The timed policies differ: own shutdown cancels and reclaims its
Store as the call unwinds, while the C API host explicitly destroys the reference
Store later. Keep both reference rows visible. The LLGo native gap remains open
and is not attributed to a specific cleanup/call stage without a profile. Local
samples are uncalibrated, and LLGo allocation counters are unavailable. No speed
parity or performance acceptance is claimed. Raw phases, source-base outcomes and
summaries are under `artifacts/components/exit-*`. Other interfaces/worlds,
component mechanisms and async/WASI 0.3 remain required; M5/M6 stay open.


## Packed byte-sequence transport

The checked environment interface merged in [PR #49](https://github.com/cpunion/wasmvm/pull/49)
at `025e230` after [the four final semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41002891297).
All 180 downloaded execution reports passed at that exact head.

The shared Canonical ABI codec now lifts list/fixed-list<u8> into immutable raw
bytes, reusing Value.Text with nil Elements. This keeps Value at 64 bytes on the
accepted platforms, removes per-byte host object expansion and supports NUL,
invalid UTF-8 and all byte values independently of string encoding. ByteList
owns mutable inputs; ByteString accepts both representations. Legacy inputs
remain checked, while byte-sequence results now use the packed representation.
Mixed representations are invalid. Logical element/nesting/range budgets remain
enforced; transfer budgets charge actual scratch/output/retained byte counts.

Go and LLGo release execute pure and host-callback round trips under both own
backends, cm32/cm64, empty and large sequences, allocator reuse, metering,
invalid shapes and retained-result isolation. Cancellation around bulk copies,
realloc/memory errors and transfer/element budgets pass. All existing 1,024 layout
and 56 collection reference cases remain present with unchanged expected values
and independent reference bytes. A test-only shape expansion compares packed
results at byte-sequence nodes. Full Go, affected LLGo release, race and vet pass;
all executable package coverage gates pass (canonical **93.24%**, component
**90.75%**, P2 **95.88%**). wasm-tools 1.258.0 validates and reproduces all four new
fixtures exactly. This is shared IO/random infrastructure; those interfaces and
complete WASI worlds remain required.

The 16 KiB echo probe uses the same cm32 guest, content and owned output in all
engines. Three sequential 200 ms own samples use Go 1.27.3 and LLGo
`8e45ed5 -O3 -lto=full` on Apple M4 Max arm64. Payload construction, file/engine/compiler,
Store/import binding/instantiation, export lookup and first call are excluded
from warm timers and separately reported. Every first payload byte and every
call's length/endpoints are checked. Source base `1fc138c` receives only the
unchanged benchmark/fixture: legacy rows execute, packed rows fail their payload
checksum under both compilers and are retained as unsupported.

Wasmtime 48.0.1 uses both the pinned prebuilt dynamic C API and a new typed Rust
reference (`&[u8]` input, owned Vec<u8> output), built with Rust 1.98.0, Cargo
release/locked dependencies and explicitly declared features. The default host
nightly could not compile the dependency; this failed build and the independent
reference profile are retained. No own compiler baseline was upgraded. Reference
samples use 1,000 dynamic C calls and 100,000 typed calls; owned-result deletion
is included. C API output slots are initialized/reset because v48.0.1's Rust
replacement drops the old C value. An initial invalid-slot crash is retained as
a harness failure and excluded from successful timings.


| Median ns / 16 KiB call | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Legacy source base | 2,066,890 | 2,076,478 | 7,990,498 | 7,908,796 |
| Legacy input at head | 1,588,158 | 1,453,173 | 6,142,223 | 7,267,104 |
| Packed head | 6,079 | 6,996 | 13,003 | 20,346 |

Wasmtime dynamic C API: **353,960 ns**; typed Rust: **469.9 ns**.
LLGo packed transfer improves the original legacy workload by roughly
615x/389x, but still trails the typed reference by roughly
28x/43x. Surpassing the dynamic C API does not meet the performance target.

The API effect is supported by Wasmtime's pinned primary implementation:
[dynamic C calls](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/c-api/src/component/func.rs)
convert the argument/result graph through dynamic Val values; its
[typed primitive-list implementation](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasmtime/src/runtime/component/func/typed.rs)
uses a checked contiguous memory window and integral copy loops eligible for
LLVM memcpy, with owned typed list lifting. This explains an architectural cost
difference; it is not a profile assigning all measured time to a single function.
Own typed call plans and allocator/call-boundary costs remain optimization work.

Startup/first phases are reported separately in the raw own metrics and both
reference JSON logs. Fuel is disabled; own cancellation/logical quotas remain
enabled and reference Stores use defaults. Results are descriptive local samples,
not calibrated performance acceptance; LLGo allocation counters are unavailable.
The initial overlapping head LLGo phase sample is retained and excluded in favor
of its isolated rerun. Raw phase logs, dependency build failures, fixture hashes,
source-base unsupported rows and final summaries are under
`artifacts/components/packed-*`. This shared transport does not complete WASI
IO/random, complete worlds, remaining component mechanisms or async/WASI 0.3.


A composition check before merge exposed a byte-value identity regression:
packed and element-wise data normalized differently when bound to an equal
value or literal. The failing source-head regression is retained. Owned clones
now normalize both forms to immutable packed bytes before equality checks,
including nested aggregates. New value import/constant-bound binaries check all
representation combinations and reject unequal content. Their authored value
grammar is not supported by wasm-tools 1.258.0; its validation diagnostics are
retained separately from the four independently validated Core ABI fixtures.
Final local full Go, affected LLGo/race and updated byte benchmarks pass; earlier
head timings remain as historical raw logs. Required CI must run again at the
corrected head before merging.


## Complete WASI random interfaces

The exit interface merged in [PR #50](https://github.com/cpunion/wasmvm/pull/50)
at `f3eca72` after [all four final semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41004817111).
All 180 downloaded execution reports passed at that exact head. Packed-byte
normalization merged in [PR #51](https://github.com/cpunion/wasmvm/pull/51) at
`a661ec7` after [all four corrected-head jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41010470056),
with all 180 execution reports verified. The random implementation is rebased
on that exact production tree and its affected Go/LLGo/race checks pass.

All five operations across `wasi:random/random`, `insecure` and `insecure-seed` at
0.2.12 now execute through the shared component runtime. NewRandom acquires
independent AES-256 keys and hash-map seed before binding, then guest callbacks
generate without entropy IO. Secure/insecure streams retain independent state,
u64 consumes the matching byte stream, and seed results are owned/stable. Public
configuration controls request bounds; deterministic secure seeds are not an
API option. Empty/default-maximum/oversized requests, exact import contracts,
entropy initialization failure and packed transfer limits pass.

Both own backends execute the same fixtures with Go and LLGo release, including
cancellation/Store close while waiting for a stream lease, cancellation during
generation, concurrent separate Stores sharing a provider, retained results and
independent OpenSSL AES-256-CTR vectors. Full Go, affected LLGo, race and vet pass;
every executable package coverage gate passes, with P2 **95.29%** (162/170).
wasm-tools 1.258.0 validates/reproduces all five new components. The unchanged
release random WIT, archive/file hashes and WASI license are retained. Wasmtime
48.0.1's actual WASIp2 provider executes the same binary and verifies all five
operations before timing. These targeted checks do not replace an upstream
conformance suite, complete CLI/HTTP world or two guest-language toolchains.

Three sequential 200 ms own samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`
on Apple M4 Max arm64. Each warm operation generates 16 KiB of secure bytes,
lowers the list to guest memory, reads its length in the guest and resets the
allocator in post-return. All file/engine/compiler, entropy initialization,
Store/binding/instantiation, lookup and first-call phases are reported outside
the warm timer. Reference context/linker/provider setup and first call are
separate; its warm scalar result avoids dynamic per-byte C API result conversion.
Source base `eb68b71` has no random provider: both compilers reject the unchanged
benchmark for missing APIs, retained as unsupported. The existing environment
operation supplies an independent common-path regression baseline.

| Median ns / call | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Secure 16 KiB generation/count | 7,119 | 8,483 | 65,337 | 69,723 | 39,741.4 |
| Existing environment base -> head | 2,461 -> 2,454 | 5,384 -> 5,377 | 6,751 -> 6,824 | 14,425 -> 14,433 | — |

| Separate 16 KiB diagnostic (ns) | Go default | Go purego | LLGo release |
| --- | ---: | ---: | ---: |
| Cipher only, preallocated buffer | 1,596 | 48,453 | 57,817 |
| Provider generation/allocation/lease | 2,499 | 49,743 | 59,261 |

These diagnostic scopes are not an additive time profile. The large LLGo cipher
cost already exists outside the VM. LLGo's pinned
[default build tags](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/internal/build/build.go#L1577)
select purego; the Go software-path control supports investigating hardware AES
support. A separate stdlib-only reproduction passes its OpenSSL vector and
measures 1,589 / 49,051 / 55,368 ns under Go default / Go purego / LLGo release.
This performance gap is reported in [llgo #2948](https://github.com/xgo-dev/llgo/issues/2948).

The reference's pinned [random context](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/random.rs)
uses StdRng for secure random and SmallRng for insecure data; its
[P2 byte operation](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/p2/host/random.rs)
collects random values through an iterator. Own calls generate bounded chunks
through Go's AES-256-CTR implementation. Algorithms and limits differ; observed
LLGo times remain about **1.64x/1.75x** behind the reference. Core call/bulk-copy
costs and LLGo crypto acceleration remain open, with no parity/acceptance claim.

Fuel is disabled, own cancellation/logical quotas remain enabled, and reference
Stores use default policy. Samples are local and uncalibrated; LLGo allocation
counters are unavailable. Raw startup/warm and diagnostic logs, source-base
failures, fixture/WIT hashes and summaries are under `artifacts/components/random-*`
and `api-free-aes-*`. IO, clocks, filesystem, sockets, complete worlds, remaining
component mechanisms and async/WASI 0.3 remain required; M5/M6 stay open.


## Host Core module imports

The random interfaces merged in [PR #52](https://github.com/cpunion/wasmvm/pull/52)
at `d25fb3d` after [all four final semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41014302827).
All 180 downloaded execution reports passed at that exact head. This change is
rebased on the merged production tree; only preceding validation notes changed.

`Engine.CompileCore` and root/nested `Imports.CoreModules` execute host-compiled
modules through the selected own interpreter/native compiler. Matching shares
the existing Core type relation and canonical registry, covering functions,
tables, memory32/64 limits, globals, tags and GC subtype identity. Every binding
is checked before guest starts. Tests cover missing/foreign/closed owners,
incompatible descriptors/import requirements, trapping starts, bounded matching,
cancellation, owner Close during binding/start, concurrent Stores, forced GC and
retained execution after Engine/component/host-owner Close. Failed unpublished
instances release their module leases without closing the host's owner.

Full Go, affected LLGo release, race and vet pass locally. Every executable
package clears 90% statement coverage: Core module **98.58%** (209/212), component
validator **92.41%** (1827/1977), component **91.04%** (976/1072), VM **93.82%**
(1017/1084). The 23 project-authored host fixtures reproduce and independently
validate with wasm-tools 1.258.0. Named GC heap-type references inside a component
module type trigger that producer's emission panic; the final fixture uses
numeric indices and validates. This producer limitation is distinct from the
runtime result.

Wasmtime 48.0.1 C API executes the ordinary and rich memory64 modules and rejects
all nine incompatible rich descriptor variants. Component compilation of the
same valid GC module-type fixture aborts at `types_builder.rs:889` with
`heap types are not supported yet`. This failed reference capability is retained;
both own backends execute the GC case successfully. It contributes no successful
reference timing row.

Three sequential 200 ms own samples and three 1,000,000-call reference samples
on Apple M4 Max arm64 use Go 1.27.3, LLGo `8e45ed5 -O3 -lto=full`, the pinned
Wasmtime 48.0.1 C SDK, and Rust 1.98.0/locked wasmtime 48.0.1. Each warm call
adds changing u32 arguments through the same component and external Core binary,
checking every result. Source base `6c8a5b4` has no host Core-module API: both
compilers reject the unchanged new probe; preserve this unsupported result.
The existing public-call workload supplies an independent base/head regression
comparison:

| Median ns / call | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Host Core module head | 458.6 | 789.2 | 1,475 | 2,392 |
| Existing public call base -> head | 461.9 -> 475.3 | 796.0 -> 785.5 | 1,462 -> 1,444 | 2,338 -> 2,315 |

Same host-module workload: Wasmtime dynamic C API **270.4 ns**, typed Rust
**96.6 ns**. The stronger typed reference remains roughly **15x/25x** faster
than LLGo. Reusing an existing execution contract introduces no separate module
call dispatch path; the warm gap remains in the shared call boundary, codecs,
allocation and safety checks. This is an architectural investigation direction,
not a measured attribution of the whole gap to any one operation. Typed call
plans and lower allocation remain required optimization work.

File load, Engine, Core/component compilation, Store/linker/binding/instantiation,
lookup and first call have separate raw phase metrics. Own Store/binding/setup
is one reported phase; references report Store/linker and instantiation
separately. Process startup and runner compilation are excluded. Fuel is off;
own cancellation/logical limits remain enabled, while reference Stores use
default policy. Reference linkers clone Core modules before deleting original
handles; own bound leases survive owner Close, but new bindings are then
rejected. These embedding policies differ. Samples are local and uncalibrated;
LLGo allocation counters remain unavailable. Raw base/head probes, references,
fixture hashes and coverage summaries are under `artifacts/components/host-module-*`.
No performance acceptance or complete M5/M6/world support is claimed.


## Complete WASI wall clock interface

Host Core module imports merged in [PR #53](https://github.com/cpunion/wasmvm/pull/53)
at `43db0bc` after [all four final-head jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41020858224).
All 180 downloaded execution reports passed. This provider is rebased on that
production tree; only preceding validation notes and package comments changed.

Both `wasi:clocks/wall-clock@0.2.12` operations execute through the same component
and selected own Core backend. Exact record field/type/order and zero-parameter
contracts reject all six independently valid mismatch fixtures at binding.
Go/LLGo tests cover zero epoch, nanosecond limits, full u64 seconds, retained
record ownership, configured source errors, invalid resolution, incomplete
configuration and cancellation/Store close during host work. Default reads
expose microsecond ticks with matching declared resolution; this is not a claim
about native OS precision or monotonicity. Callback providers may expose their
own valid clock contracts. The pinned release clocks WIT, hash and license are
retained; all seven authored binaries reproduce and validate with wasm-tools
1.258.0. Wasmtime 48.0.1's actual WASIp2 provider executes the same positive binary
and validates both operations before timing.

Affected Go/LLGo release, race and vet pass; P2 statement coverage is **95.79%**
(205/214), with every unchanged executable package's prior gate retained. These
checks are targeted interface tests, not complete clocks/CLI/HTTP worlds or a
replacement for upstream conformance and two guest-language toolchains.

Three sequential 200 ms own samples on Apple M4 Max arm64 use Go 1.27.3 and LLGo
`8e45ed5 -O3 -lto=full`. Three 100,000-call actual WASIp2 C samples use the pinned
Wasmtime 48.0.1 SDK. Each warmed operation reads datetime, lowers it to guest
memory and returns its seconds field. File/Engine/compiler, provider setup,
Store/binding/instantiation, lookup and first phases are outside warm timing.
Reference context/linker/provider and instantiate phases remain separate.
Source base `54c4e13` lacks the new clock API: both compilers reject the unchanged
probe, retained as unsupported. Common-path environment data is separate:

| Median ns / call | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Wall-clock now/guest seconds | 870.5 | 1,350 | 2,773 | 4,163 |
| Existing environment base -> head | 2,484 -> 2,512 | 5,393 -> 5,402 | 6,807 -> 6,916 | 14,348 -> 14,253 |

Actual Wasmtime WASIp2: **323.0 ns**. LLGo remains roughly **8.6x/12.9x** behind
this reference. Separate clock/provider diagnostic medians are Go **30.1/56.1 ns**
and LLGo **54.8/95.5 ns**. This isolates a cheap clock source from the larger
component-call workload; the independent scopes are not an additive profile
and do not assign the remaining gap to one function. The common canonical/call
boundary remains optimization work.

Fuel is disabled, own cancellation/logical quotas remain enabled and reference
Stores use default policies. Clock precision differs as documented; the warm
result is scalar, avoiding per-byte/per-field dynamic result expansion. These
local samples are uncalibrated; LLGo allocation counters remain unavailable.
Raw phase/diagnostic logs, reference JSON, source-base API failures, fixture/WIT
hashes and coverage are under `artifacts/components/wall-clock-*`. IO/poll,
monotonic subscriptions, streams, complete worlds and remaining component/async
mechanisms remain required. No M5/M6 or performance acceptance is claimed.


## Complete poll and monotonic-clock interfaces

Wall-clock merged in [PR #54](https://github.com/cpunion/wasmvm/pull/54) at
`1ddce99` after [all four final semantic jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41023242783).
All 180 downloaded execution reports passed. This increment is rebased on the
merged production tree; preceding validation notes and package comments differ.

All three `wasi:io/poll@0.2.12` operations and all four stable monotonic-clock
operations now execute through the shared component runtime with one checked
nominal pollable identity. Tests cover coherent readiness notifications,
duplicate indices, empty/list/domain quotas, missing/foreign/closed capabilities,
all eight independently valid contract mismatches, cancellation and Store/provider
close during block/poll, timer races, concurrent close, failed publication and
provisional/detached cleanup. No waiter goroutine is created per polled source.
Provider Close prevents new publication and joins in-flight creation/cleanup
and already dispatched internal timer callbacks. The explicit provider owns
abandoned entries until closed; Store close invalidates handles and cancels calls
without synthesizing guest destructors.

Absolute and relative subscriptions preserve full unsigned input ranges. Tests
check past/zero deadlines, real delayed wakeup without shortened duration, huge
pending instants/durations and deterministic resource drops. Signed timer limits
are chunked; relative u64 deadline overflow remains pending rather than wrapping.
Clock origin and microsecond ticks are provider policy, separately documented.
Go/LLGo release, both own backends, race and vet pass locally. P2 statement
coverage is **94.75%** (451/476), and every executable package gate passes using
unchanged-package profiles plus the updated affected profile. All ten authored
fixtures reproduce and independently validate with wasm-tools 1.258.0; unchanged
release IO WIT, hashes and the existing WASI license are retained.

The same component executes all seven operations with Wasmtime 48.0.1's actual
WASIp2 provider. An initial probe incorrectly assumed a past deadline's first
nonblocking ready must return true. The pinned reference
[explicitly yields once](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/p2/host/clocks.rs)
for scheduler fairness. The final shared probe blocks before asserting completed
readiness, then checks exact duplicate indices, huge deadlines and drops. The
original reference failure remains separately recorded. Targeted authored tests
are not complete upstream conformance, guest-toolchain or world acceptance.

Three sequential 200 ms own samples use Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`
on Apple M4 Max arm64. Each warm call executes 128 complete zero-duration
subscription/create, block and resource.drop lifetimes in guest code. Results
and final own domain entry counts are checked. Three 10,000-call reference
samples use the pinned Wasmtime 48.0.1 C SDK, same binary and batch size. Scalar
C API call overhead is amortized across 128 lifetimes. The source base `145a3bc`
lacks Poller/MonotonicClock: both compilers reject the unchanged new probe,
retained as unsupported. The unchanged wall-clock workload is a separate
common-path baseline:

| Median ns | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Complete subscription lifetime | 2,397.8 | 3,260.1 | 15,736.7 | 19,608.4 |
| Existing wall-clock call base -> head | 863.4 -> 850.5 | 1,337 -> 1,332 | 2,798 -> 2,834 | 4,253 -> 4,267 |

Actual Wasmtime WASIp2: **359.2 ns / lifetime**, or 45,983.9 ns / 128-lifetime
call. LLGo remains roughly **44x/55x** behind. Go reports about **69 allocations
and 3.5 kB per lifetime** in this complete path, a significant allocation budget
for shared call/resource conversion work. These measurements do not attribute
the entire gap to one function. Shared typed call plans, ownership conversion
and allocation reduction remain architectural performance investigation/work;
this capability increment claims no performance parity or acceptance.

File/Engine/compiler, provider/Store binding/instantiation, lookup and first
calls are outside warm timing and reported separately. Fuel is disabled, own
cancellation/logical quotas remain enabled, and reference Stores use defaults.
Reference immediate subscriptions yield to Tokio; own subscriptions are directly
ready. Preserve that policy difference even though the reference remains faster.
Samples are local and uncalibrated; LLGo allocation counters are unavailable.
Raw original/final samples, failed reference probe, pinned reference source,
fixture hashes and coverage summaries are under `artifacts/components/poll-*`.
IO error/streams/stdio, remaining component resource/host/async mechanisms,
complete CLI/HTTP worlds and WASI 0.3 remain required; M5/M6 stay open.


## Complete WASI IO, stdio and terminal interfaces

Poll/monotonic-clock merged in [PR #55](https://github.com/cpunion/wasmvm/pull/55)
at `f1187c8` after [all four final jobs passed](https://github.com/cpunion/wasmvm/actions/runs/41026397758).
All 180 downloaded execution reports passed. This increment implements every
operation in error/streams, stdin/stdout/stderr and the five terminal interfaces
at 0.2.12: 22 callable operations plus their resource declarations/implicit drops.
It shares the existing poll and Core execution contracts.

The complete generated WIT/component exercises all 15 stream operations,
stdio getters, exact data/counts, EOF and drops. A failure probe creates a real
last-operation-failed resource, reads its debug string, drops it and verifies
subsequent closed state. Own and actual Wasmtime 48.0.1 WASIp2 execute these
same binaries. Virtual terminals return none in the common fixture; own tests
also enable all three explicit terminal capabilities. Additional own tests cover
ready/blocked partial IO, queue/backpressure, flush state/permits, Store/provider
cancellation, parent/child ownership, expired/foreign capabilities, source quotas,
failed publication, source panic, cleanup errors and bounded UTF-8 snapshots.

Eight independently valid WIT mutations reject wrong provider contracts before
guest effects. All ten authored components reproduce and validate with
wasm-tools 1.258.0 using the verified unchanged WASI archive. The selected world,
generated Core WAT/component and checksum manifest are retained, with
`scripts/build_wasi_io_fixtures.py` as the pinned reproducer. An initial error
probe used the void-result payload offset for a u64-result assertion; both
engines trapped. The corrected probe uses the independently checked alignment;
the original failures remain as diagnostics. Full affected Go/LLGo release,
both own backends, race and vet pass locally. P2 statement coverage is **90.98%**
(1180/1297), and every executable package's coverage gate passes.

The reference rejects over-permit writes, blocking payloads above 4,096 bytes
and dropping parents with live child pollables. Its stdio backend accepts an
initial write without check-write; own code enforces the WIT precondition. The
pinned [generic IO host](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-io/src/impls.rs)
delegates check/write to the concrete stream backend. Preserve this observed
policy difference and the original harness-only endpoint-check failure separately;
it does not explain the common blocking-write benchmark by itself.

Three sequential 200 ms own samples on Apple M4 Max arm64 use Go 1.27.3 and LLGo
`8e45ed5 -O3 -lto=full`. Each warmed guest call performs 128 complete blocking
4 KiB write-and-flush operations through stdout, with explicit getter/drop and
checked sink byte totals/endpoints. Three 1,000-call reference samples use the
pinned actual Wasmtime 48.0.1 WASIp2 C provider and identical checks/batch. Scalar
C API result overhead is amortized. Source base `62e9d5d` lacks the new IO APIs:
both compilers reject the unchanged benchmark, retained as unsupported. The
unchanged wall-clock call is a separate common-path regression baseline:

| Median ns | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Complete 4 KiB write/flush | 2,895.3 | 3,482.2 | 6,388.3 | 8,434.1 |
| Existing wall-clock call base -> head | 854 -> 860.9 | 1,318 -> 1,335 | 2,777 -> 2,773 | 4,131 -> 4,221 |

Actual Wasmtime WASIp2: **7,987.3 ns / write**, 1,022,376 ns / 128-write call,
**512.8 MB/s**. Own LLGo interpreter/native achieve **641.2 / 485.6 MB/s** in
this virtual checked-sink workload: interpreter is about 20% faster; native
remains about 6% slower. Local data alone is not calibrated performance acceptance
or proof of general filesystem/network throughput.

| Independent 4 KiB diagnostic (ns) | Go | LLGo release |
| --- | ---: | ---: |
| Checked sink only | 5.76 | 49.78 |
| Blocking backend | 385.8 | 1,069 |
| Queued backend | 926.1 | 24,960 |

These are separate scopes, not an additive time profile. The sink counts logical
bytes and validates endpoints; its byte rate is not a physical memory-bandwidth
measurement. Blocking operations can use the prompt source directly when no
older output is queued, avoiding mandatory worker handoff; nonblocking writes
retain a bounded queued worker. The LLGo scheduling/allocation cost visible in
the queued diagnostic makes that distinction material. Shared canonical/native
call-boundary work and queued LLGo costs remain optimization targets.

File/Engine/compiler, provider setup, Store/binding/instantiation, lookup and
first-call phases are outside warm timing and reported separately. Fuel is off,
own cancellation/quotas remain enabled, and reference Stores use defaults with
different buffering/scheduling internals. No real OS IO occurs in these sink
samples. LLGo allocation counters remain unavailable. Raw phases, diagnostics,
source-base failures, reference probes, fixture hashes and coverage are under
`artifacts/components/io-*`. Filesystem/preopens, sockets/HTTP, complete worlds,
remaining component mechanisms and native async/WASI 0.3 remain required;
M5/M6 and general performance acceptance stay open.

## Complete WASI filesystem interfaces (2026-10-30)

The filesystem/preopens provider implements all 30 pinned WASI 0.2.12 callable
operations on Linux/Darwin. The authored guest exercises file/stream IO,
flags/types/stat/timestamps/hashes, links, directory iteration and explicit drops
on both own backends; Wasmtime 48.0.1's actual WASIp2 provider executes the same
binary. Eleven components independently reproduce/validate with wasm-tools
1.258.0: the guest, an exact-domain embedding API view, and nine valid signature
mutations that must fail provider binding. A host-induced revocation probe checks
filesystem error provenance through guest stream/error operations.

Go 1.27.3 and LLGo `8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`
(`-O3 -lto=full`) pass the affected package tests. Full local Go tests, targeted
race checks and vet pass. Preview 2 statement coverage is **91.9%**; every
executable-package coverage gate passes when the current P2 profile is combined
with PR #56's exact unchanged-source CI coverage for other packages. The final four-job Go/LLGo × Linux/macOS matrix passed at source head
`f3e54bd16881cf2fa21301a9b64de0b13e8da30e`; all 180 execution reports were
verified before [PR #57](https://github.com/cpunion/wasmvm/pull/57) merged
([CI](https://github.com/cpunion/wasmvm/actions/runs/41041101413)). Tests include readonly and
cross-directory grants, absolute/parent/symlink escapes, bounded partial reads
and writes, u64 overflow, nanosecond times, enumeration quotas/encoding, expired
and stale resources, failed publication/rollback, independent streams, atomic
append versus positional IO, pipe readiness/backpressure, error downcasts,
provider revocation and concurrent shutdown. APFS rejects non-UTF-8 file names
before enumeration; that observed OS error is recorded, while Linux exercises
the returned-name encoding rejection. An initial enum/flags result-offset guest
assertion and initial benchmark wrong-cwd failure are retained as diagnostics.

Local serial descriptive samples on Apple M4 Max (three 200 ms own samples and
three 1,000-call reference samples) use the same component and local filesystem.
Each warmed call opens one preopen/file, performs **128 stats**, asserts the
16-byte size in the guest and drops both handles. The per-stat rate includes
amortized open/drop. Guest file creation and the reference's semantic exercise
are outside timed warm calls. Own fuel is off; cancellation, resource and
transfer bounds remain on. Wasmtime uses SDK default WASIp2 policies/scheduling;
C scalar-result conversion is amortized across 128 operations.

| Warm scope | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Complete stat, ns/operation | 3,896 | 4,607 | 9,658 | 11,834 | 5,583 |
| Complete call, ns/128 stats | 498,701 | 589,744 | 1,236,235 | 1,514,727 | 714,602 |

LLGo remains about **1.73× / 2.12× slower** on this full workload. Pure host
fstat is **278 / 368 ns** (Go/LLGo), while stat-value construction is
**176 / 309 ns**. These independent diagnostics are not additive; the full guest
boundary/aggregate transfer dominates the observed gap, rather than host fstat
alone. Go allocates roughly 533/534 KiB and 7,957/7,978 objects per batch across
interpreter/native; LLGo allocation counters are unavailable. Avoid claiming a
throughput win from the raw syscall or C embedding comparison.

Separate median preparation costs (ns) are retained in
`artifacts/components/filesystem-performance-summary.json`:

| Phase | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Component compilation | 954,500 / 4,310,791 | 1,675,750 / 5,990,958 | 5,960,000 |
| Provider/Store setup | 66,666 / 74,292 | 90,917 / 100,875 | 124,000 |
| Bind/instantiate | 340,084 / 302,292 | 306,375 / 1,500,916 | 124,000 |
| First call | 483,000 / 647,458 | 1,000,583 / 1,368,000 | 785,000 |
| Teardown | 12,708 / 20,125 | 103,500 / 85,250 | 200,000 |

Load, engine creation and lookup are also recorded separately. Source base
`cca230cd8417befc6f3a8e0c31c3b8cf015333be` lacks the FileSystem provider APIs;
both compiler failures are preserved as API-baseline absence, without claiming
that its component engine cannot compile the binary. The unchanged wall-clock
common workload changes Go **835.7 -> 823.4 / 1,276 -> 1,270 ns**, LLGo
**2,678 -> 2,728 / 4,049 -> 4,132 ns**. These are descriptive, uncalibrated
samples; revision regression and competitor performance acceptance remain open.
Aggregate canonical call/transfer allocation and LLGo overhead remain explicit
optimization work under the current integration policy. Sockets/HTTP, complete
worlds/toolchain acceptance, remaining component mechanisms and WASI 0.3 remain
required.

## Network capabilities and IP name lookup (2026-10-30)

All five operations in the pinned network, instance-network and IP name lookup
interfaces execute through checked nominal resources. Go 1.27.3, LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` (`-O3 -lto=full`) and race checks
pass. Preview 2 statement coverage is **92.1%**. Every executable package retains
its 90% gate when the current P2 profile is combined with PR #57's exact
unchanged-source Linux CI profile for other packages. The final Go/LLGo × Linux/macOS matrix passed at source head
`bf7d37c4a5c7f4c57cabaeed090157f8254ee753`; all 180 execution reports were
verified before [PR #58](https://github.com/cpunion/wasmvm/pull/58) merged
([CI](https://github.com/cpunion/wasmvm/actions/runs/41044534178)). Tests cover both own backends,
IPv4/IPv6/mapped literals, IDNA, actual localhost lookup, pending/ready/EOF,
resolver error classes/panics, result/name/stream quotas, permissions, timeout,
caller cancellation, failed publication/rollback, typed error downcasts,
expired Callers, child pollables, provider revocation and concurrent shutdown.
The IDNA dependency is pinned to `x/net v0.61.0` / `x/text v0.43.0`; no engine is
used as the implementation. Literal lookups allocate no DNS timer or worker;
named queries use a private `PreferGo` resolver or an authorized callback.

Nine components independently reproduce/validate with wasm-tools 1.258.0.
**Wasmtime 48.0.1's C WASIp2 linker cannot bind the full consumer:**
`network-error-code` has no matching function implementation. The original
failure is retained in `network-reference-initial.log`. The standard WIT
contains the operation; the benchmark does not install a stub or replacement
provider. A separately retained narrow consumer omits only the unused function
requirement. Both engines execute that exact common binary, including IPv4 and
IPv6 guest assertions. Six valid contract mutations reject binding on the own
provider; the complete own API and error provenance are tested separately.

Local serial samples on Apple M4 Max (three 200 ms own samples and three
1,000-call reference samples) measure **128 complete literal lookup lifetimes**
per guest call: resolve, subscribe/block/check-ready, next address, next EOF,
drop pollable and drop stream, with one network handle per batch. This is not
DNS wire/server latency or DNS throughput. No network packet is sent by this
workload. Fuel is off; own quotas/cancellation remain enabled. Both providers
explicitly grant lookup; their default scheduling/safety policies differ. C
scalar result conversion is amortized across 128 lifetimes.

| Warm scope | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Complete lifetime, ns | 8,153 | 10,457 | 24,525 | 32,137 | 990 |
| Complete call, ns/128 lifetimes | 1,043,544 | 1,338,516 | 3,139,139 | 4,113,542 | 126,721 |

The LLGo gap is **24.8× / 32.5×** on this workload. Pure parse is
**10.75 / 14.67 ns** (Go/LLGo), IP-value construction **47.05 / 72.46 ns**.
These independent diagnostic scopes are not additive; the large full-lifetime
gap points to canonical/resource call overhead and requires further profiling
and optimization. Go allocates about 1.625 MB and 31,022/31,031 objects per batch;
LLGo allocation counters are unavailable. A separate sampled Go allocation
profile attributes 17.18% of objects to Instance.codec, 14.88% to NewReference
and 10.16% to NewBorrowScope (42.22% together); Function.call adds 9.59%.
This identifies shared allocation mechanisms to investigate, not LLGo allocation
counts. The short Darwin CPU profile is dominated by runtime wait/scheduling
frames and is not used to attribute LLGo costs. Neither the raw parse nor this
narrow consumer establishes full networking performance acceptance.

Separate median preparation costs (ns) are retained in
`artifacts/components/network-performance-summary.json`:

| Phase | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Component compilation | 344,708 / 1,028,250 | 551,041 / 1,676,541 | 1,983,000 |
| Provider/Store setup | 12,500 / 14,334 | 15,708 / 15,042 | 84,000 |
| Bind/instantiate | 50,250 / 62,125 | 91,834 / 102,958 | 90,000 |
| First call | 924,125 / 1,534,500 | 2,592,291 / 3,427,125 | 151,000 |
| Teardown | 10,791 / 17,792 | 83,250 / 90,875 | 148,000 |

Load, engine creation and lookup are separate metrics as well. The initial
DefaultResolver run is retained separately from the final private Go resolver
policy. Source base `f3e54bd16881cf2fa21301a9b64de0b13e8da30e` lacks NewNetwork;
both expected compiler failures are retained as API-baseline absence. Existing
filesystem benchmark binaries with identical production source supply the
common source-base control. Wall-clock warm medians change Go
**858.5 -> 860.4 / 1,322 -> 1,330 ns**, LLGo
**2,729 -> 2,811 / 4,098 -> 4,050 ns**. Samples are descriptive and uncalibrated;
revision regression and competitor performance acceptance remain open.
TCP/UDP/HTTP, complete worlds/toolchains, remaining component mechanisms and
WASI 0.3 remain required. The current integration policy permits this complete
interface capability to advance while the measured performance deficit remains
explicit work.

## Complete WASI TCP interfaces (2026-10-30)

The Linux/Darwin provider implements all 28 TCP methods and socket creation
from pinned WASI 0.2.12. The complete guest executes IPv4/IPv6, options,
bind/listen/connect/accept, bidirectional IO, shutdown and explicit resource
drops on interpreter/native. Eleven components reproduce and independently
validate with wasm-tools 1.258.0; eight valid signature mutations reject binding.
Nested public socket-address names exposed a component validator defect:
instance freshening now preserves captured visible compound names while still
freshening local resource binders and rejecting private-name escape.

Go 1.27.3 full package tests/coverage, targeted race tests, vet and affected LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f` (`-O3 -lto=full`) tests pass.
Every executable package passes 90% statement coverage; Preview 2 is **91.7%**,
component validation **92.8%** with cross-package coverage, and public component
**91.0%**. Tests include default-denied/explicit grants, callback panic/cancel,
pending-operation conflicts, full IPv6 addresses, port/family/option rejection,
partial IO/backpressure, connection refusal, EOF/error downcasts, stream/socket
and pollable/socket lifetime, quotas/failed-publication rollback, inherited
listener options, owner revocation, concurrent close and retained cleanup
errors. An external-FD-close fault injection initially failed the ordinary
success-only test cleanup; the helper now asserts the expected cleanup error.

Wasmtime 48.0.1 cannot link the full consumer's `network-error-code`. The common
consumer omits only that unused requirement. The C API additionally returns
access-denied on TCP creation: address inheritance does not enable TCP, and the
C API has no allow-TCP setter. These failures remain in `tcp-c-api-final.log`
and `tcp-common-c-api-final.log`. The performance reference is the actual Rust
WASIp2 provider with `inherit_network().allow_tcp(true)`, not a replacement.
Pinned Rust 1.98.0 builds successfully; an initial root-default-nightly compiler
failure in a Wasmtime dependency remains retained without cache patches. See
[the C API policy](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/c-api/src/wasi.rs)
and [WASI context defaults](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/ctx.rs).

Local serial descriptive medians on Apple M4 Max use three samples, 100 own and
1,000 reference batches per sample. Both run the same component's IPv4/IPv6
semantic exercise, open a persistent IPv4 loopback session and warm ten batches
before timing. Each guest batch makes **128 roundtrips**, 4 KiB in each
direction, checking returned lengths and pattern endpoints. Fuel is off;
quotas/cancellation remain enabled. Own grants cover loopback only; reference
inherits network authority, with different default scheduling/safety policies.
TCP_NODELAY retains the OS default in both providers.

| Warm scope | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Roundtrip, ns | 35,263 | 49,521 | 50,879 | 60,872 | 29,763 |
| Batch of 128, ns | 4,513,642 | 6,338,672 | 6,512,492 | 7,791,602 | 3,809,643 |

LLGo is **1.71× / 2.05× slower** than this reference; performance acceptance
remains open. Go allocates approximately 7.55/7.62 MB and 21,275/22,818 objects
per batch; LLGo's allocation counters are not valid evidence. Aggregate canonical
transfer/resource overhead remains optimization work; these full TCP figures
are not a pure syscall/kernel measurement and do not isolate the cost of each
internal stage. No performance win is claimed.

Startup is excluded from warm timing. Separate median phases (ns):

| Phase | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Component compilation | 892,375 / 3,860,708 | 1,586,459 / 7,454,583 | 11,252,042 |
| Provider/Store setup | 24,666 / 29,041 | 33,250 / 33,041 | 80,958 |
| Bind/instantiate | 419,709 / 186,583 | 290,334 / 319,875 | 96,958 |
| Connection session setup | 127,125 / 164,750 | 217,125 / 272,667 | 112,667 |
| First 128 roundtrips | 4,862,959 / 6,641,917 | 6,945,709 / 7,965,792 | 4,212,250 |
| Session shutdown/drop | 33,875 / 41,916 | 204,416 / 222,709 | 39,500 |
| Owner teardown | 16,834 / 28,167 | 84,084 / 82,667 | 169,041 |

Load, engine creation and lookup phases, compiler/binary/fixture identities and
raw samples are retained under `artifacts/components/tcp-*`, including
`tcp-performance-summary.json`. Source base `409b859` lacks NewTCP under both
compilers; preserve that API absence separately from engine execution support.
The unchanged wall-clock workload changes Go **856.6 -> 851.2 / 1,325 -> 1,324
ns**, LLGo **2,658 -> 2,652 / 4,030 -> 4,021 ns**. These uncontrolled samples do
not establish a regression guarantee. UDP/HTTP, complete worlds/toolchain
conformance, remaining component mechanisms and native async/WASI 0.3 remain
required; M5/M6 stay open. The final-head Go/LLGo × Linux/macOS matrix passed at
`ec3fa07bcb5f35baa2470eababd691d016316a75`; all 180 execution reports and
coverage gates were verified before [PR #59](https://github.com/cpunion/wasmvm/pull/59)
merged ([CI](https://github.com/cpunion/wasmvm/actions/runs/41066359930)).

## Complete WASI UDP interfaces (2026-10-30)

All 19 Linux/Darwin UDP/create and datagram-stream operations execute on both
own backends. The shared guest checks connected/unconnected IPv4/IPv6,
zero-length packets, options, readiness, explicit peers and resource drops;
the actual Wasmtime 48.0.1 Rust WASIp2 provider executes the identical common
binary with UDP explicitly enabled. Eleven components reproduce and validate
with wasm-tools 1.258.0; eight valid type mutations reject binding. The full
consumer's missing `network-error-code` requirement remains separate from the
narrow common consumer; no reference provider is substituted.

Go 1.27.3 and pinned LLGo release tests, race checks and vet pass locally.
Preview 2 statement coverage is **91.5%**. Tests cover separate bind/send/receive
grants, denied packet filtering, callback panic/cancellation, zero/oversized
packets, partial batch success, send-permit misuse, peer mismatch, stream
reconfiguration, quotas and partial-publication rollback, orphaned pollables,
blocked-poll cancellation, parent revocation, concurrent close and retained
physical-close failures. Shared transport helpers preserve complete IPv6
flow/scope values and FD leases; TCP protocol state remains independent.

Local serial medians on Apple M4 Max use three samples (100 own / 1,000
reference batches each), ten warmup batches and the same persistent connected
IPv4 guest. Each batch makes 128 roundtrips, **4 KiB per direction**, including
check-send/send and receive/readiness. Socket/stream setup, first batch and
teardown are excluded from warm timing. Fuel is off; own quotas/cancellation
remain enabled with loopback grants. Reference network inheritance, policy and
completion-based scheduling differ.

| Warm scope | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Roundtrip, ns | 53,371 | 57,061 | 89,737 | 103,764 | 42,166 |
| Batch of 128, ns | 6,831,475 | 7,303,845 | 11,486,282 | 13,281,848 | 5,397,239 |

LLGo is **2.13× / 2.46× slower**; performance acceptance remains open. Go
allocates 9.45/9.09 MB and 71,019/64,692 objects per batch. LLGo allocation
counters are unavailable. These full guest/provider rates include canonical
aggregate transfer, permissions, resources and kernel IO; they do not isolate
raw UDP syscall efficiency. General call/transfer overhead remains explicit
optimization work.

| Separate phase, ns | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Component compilation | 729,417 / 2,965,625 | 1,210,958 / 4,238,500 | 7,365,125 |
| Provider/Store setup | 23,375 / 23,208 | 28,583 / 29,708 | 79,917 |
| Bind/instantiate | 357,166 / 119,542 | 197,041 / 204,625 | 83,084 |
| Endpoint/stream setup | 87,583 / 103,917 | 153,000 / 211,334 | 55,333 |
| First batch | 6,754,583 / 7,077,042 | 12,387,084 / 15,648,250 | 4,409,500 |
| Session drop | 12,959 / 21,292 | 143,833 / 145,042 | 28,792 |
| Owner teardown | 24,667 / 28,000 | 80,792 / 76,042 | 163,833 |

Source base TCP head `ec3fa07` lacks NewUDP under both compilers; that API
absence is retained independently of engine support. Unchanged wall-clock
medians change Go **886.9 -> 885.0 / 1,371 -> 1,350 ns**, LLGo **2,796 -> 2,974 /
4,251 -> 4,349 ns**. These uncontrolled local samples are descriptive; they do
not establish performance acceptance or a regression guarantee. Raw phases,
compiler/binary identities, producer hashes and failures are retained in
`artifacts/components/udp-*`. The final-head Go/LLGo × Linux/macOS matrix passed at
`0c6fa04cd16798f7312d24ee4929e6321bb1def2`; all 180 execution reports and
coverage gates were verified before [PR #60](https://github.com/cpunion/wasmvm/pull/60)
merged ([CI](https://github.com/cpunion/wasmvm/actions/runs/41070858048)).
HTTP, complete worlds/toolchain conformance, remaining component mechanisms
and native async/WASI 0.3 remain required; M5/M6 stay open.

## HTTP resource and fixture preparation (2026-10-30)

The unchanged HTTP WIT is now retained from the pinned WASI archive. Its stable
surface comprises 51 types operations plus two handlers. The separately gated
informational-outbound-responses extension is not enabled by these fixtures.
No HTTP interface or complete world is added to implementedInterfaces.

Private ordered field storage preserves case/bytes, case-insensitive lookup,
immutable views, detached snapshots/clones and atomic bounded mutations. A
bounded circular body buffer implements partial writes, backpressure,
cancellable blocking transport reads, nonblocking ports, flush readiness,
EOF/errors and discard-to-trailers behavior without per-port workers.
Go/LLGo release tests, targeted Go race tests and vet pass for this preparation;
Preview 2 coverage is **91.9%**. These helpers are not yet an HTTP provider.

Ten components reproduce and independently validate with wasm-tools 1.258.0:
a field probe importing the full stable types interface, its exact-domain API
view and eight valid signature mutations. Both own backends compile all ten;
this establishes type/canonical compilation, not own HTTP guest execution.
The probe exercises all nine field operations and invalid syntax on the actual
Wasmtime 48.0.1 Rust HTTPp2 provider. Its default send implementation remains
enabled and unmodified, although the field probe performs no network requests.
Pinned Rust 1.98.0 builds the reference with the retained Cargo.lock.

An initial probe read the unit-result header-error payload at +4 rather than
its byte-aligned +1 location; the resulting reference assertion failure remains
in http-fields-reference-initial.log. The corrected binary passes. An initial
concurrent buffer test ignored an already-ready snapshot before waiting; the
retained timeout led to respecting the level check and testing final flush
readiness. Raw evidence is under artifacts/components/http-*.

All HTTP resource bindings, client/proxy execution, transport/TLS/timeouts,
permission/ownership integration, full performance comparisons and world/guest
toolchain conformance remain required before HTTP acceptance. Canonical async
and WASI 0.3 remain separate open work.


## Stable WASI HTTP client and proxy (2026-10-31)

This supersedes the preparation-only scope above. `HTTP` implements all 51
stable types operations, outgoing handling and the standard guest incoming
handler on Linux/Darwin. `HTTPHandler` creates independent request Stores and
streams responses concurrently with guest execution. URL and resolved-connect
grants default to deny; private transports, bounded resources/headers/body
queues, timeout options, ownership rollback and joined shutdown share existing
IO/network/canonical contracts. No reference engine is used for own execution.

The full authored proxy-world consumer checks HTTP/HTTPS requests, 4096-byte
bodies, request/reply trailers and guest incoming handling on interpreter/native
and Go/LLGo release. An independent strict server validates every request byte.
API tests cover immutable/cloned fields, all metadata/options, single-use
body/future states, parent/child lifetime traps, foreign resources, quotas,
failed publication, cancellation, TLS trust rejection, response truncation and
typed IO downcasts. A 257-byte body queue exercises real incoming streaming and
discard-to-trailers backpressure. Empty/redirect responses do not follow
redirects. All 39 error-code payloads roundtrip through response-outparam.set;
this is payload/ABI coverage, not 39 generated physical network failures.
Eight valid contract mutations reject binding before guest effects.

Local final Go/LLGo package binaries, Go race tests, vet and no-cgo build pass.
Preview 2 statement coverage is **90.27% (4806/5324)**. The required four-job
final-head CI and full execution-report/coverage audit remain merge conditions.
The actual unmodified Wasmtime 48.0.1 Rust HTTPp2 provider executes the same
HTTP client and incoming handler; the latter independently checks status,
complete response bytes and trailers. Private-root HTTPS is tested on own code,
without claiming a default-hook reference TLS comparison. Complete-world and
two guest-language-toolchain acceptance remain open, as do M5/M6 and WASI 0.3.

Three serial samples on Apple M4 Max / Darwin arm64 use Go 1.27.3, pinned LLGo
`8e45ed5 -O3 -lto=full` and Rust 1.98.0. Each call checks 128 lifetimes after ten
warmup batches. Fields and body use 1000 measured batches/sample; full requests
use 20. Fuel is off; quotas/cancellation and loopback grants remain enabled.
The body workload produces/finishes 4096 bytes without a network consumer;
it is not a bidirectional streaming throughput claim. Preparation and first
batch are outside warmed measurements. Medians below are ns per lifetime:

| Scope | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime HTTPp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Fields operations/drop | 22,396 | 32,386 | 53,733 | 81,412 | 2,328 |
| 4 KiB body production/finish | 12,366 | 15,484 | 30,174 | 39,966 | 1,266 |
| 4 KiB POST/response/trailers | 137,349 | 192,985 | 354,735 | 440,687 | 129,309 |

LLGo remains about **23.1×/35.0× slower for fields**, **23.8×/31.6× for
body production**, and **2.74×/3.41× for full requests**. Performance acceptance
is unmet. Native code is slower than the interpreter for these host-call-heavy
workloads; Go fields batches allocate 67,985/74,138 objects for the two backends.
Shared canonical/host-call and resource-transfer overhead is explicit follow-up
work. LLGo reports zero allocation counts; that counter is unavailable and does
not mean zero allocation. These are local descriptive samples, not controlled
regression or platform acceptance.

The full requests use one common external server for all processes. It records
12 connections/33,024 checked requests for each own compiler (both backends,
including benchmark calibration/warmups), versus 11,904 connections/11,904
requests for the reference. All final runs report zero invalid messages. The
actual reference opens a connection per request; own code pools connections.
These rates include complete policy/connection costs and do not prove pure
engine parity. Reference sample rates span 120,972–161,625 ns/request.

Separate complete-request phases, in ns (first means a batch of 128 and includes
its initial connection; connection establishment is not isolated by subtraction):

| Phase | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Load | 353,958 / 353,958 | 99,667 / 99,667 | 40,209 |
| Engine | 1,000 / 916 | 1,167 / 1,167 | 85,666 |
| Compile | 1,486,917 / 4,988,334 | 2,082,000 / 9,085,291 | 14,693,042 |
| Provider/Store setup | 38,625 / 40,791 | 49,250 / 52,375 | 123,292 |
| Bind/instantiate | 642,958 / 509,459 | 434,834 / 473,917 | 103,916 |
| Lookup | 292 / 250 | 375 / 291 | 875 |
| First batch | 18,657,333 / 24,005,542 | 49,509,292 / 55,969,916 | 16,942,083 |
| Owner teardown | 24,167 / 29,791 | 111,417 / 112,958 | 179,917 |

Source base `a37898a` lacks NewHTTP under both compilers; that absence is retained
as an API-baseline failure. The unchanged wall-clock medians, interpreter/native,
change Go 843.6/1315 -> 837.6/1302 ns, LLGo 2754/4226 -> 2706/3989 ns. No regression
guarantee is inferred. Raw logs, phases, source/compiler/binary hashes and
server counters remain in `artifacts/components/http-*`; reproduction commands
are in [the benchmark guide](../benchmarks/components/README.md).

Failures remain visible: the earlier valid guest without an explicit request
Trailer declaration loses its trailer under the default reference sender and
fails strict validation. Its binary/logs are retained separately. A sustained
100-batch reference attempt fails around 15,977 server connections; local
TIME_WAIT/ephemeral-port pressure is suspected, not established as a provider
bug. Final shorter samples use a fresh shared endpoint without changing guest
assertions or sender behavior. Sustained request acceptance remains open.

A separate standard-library-only cancellation reproducer reports Go client
cancellation at 10–28 µs and LLGo at 1.00–1.01 s (eight samples). See
[llgo #2953](https://github.com/xgo-dev/llgo/issues/2953). Tests allow three seconds
for eventual completion; protocol timeout values and typed error assertions
remain unchanged. The pinned compiler is not silently patched or upgraded.


The first final-head macOS LLGo CI attempt exposed a test that polled a truncated
response only 100 times before the pinned netpoll fallback tick. The typed IO
assertion now uses blocking-read with a three-second call deadline, preserving
both the stream failure and its HTTP error-code downcast. The original failed
job is retained in pr61-ci-failed.log; no production or benchmark code changed.


## Stable world assembly and independent CLI guests (2026-10-31)

`NewWorld` now composes all stable 0.2.12 CLI/HTTP host providers with shared
nominal IO, error, pollable and network domains. Configuration supplies explicit
args, stdio, preopens and grants. Constructor rollback joins both primary and
cleanup failures; repeated/concurrent close joins every resource owner.
`RunCommand` validates its standard unit-result entry before invoking it and
preserves success/failure/exit codes, traps and cancellation. `Function.Info`
exposes immutable export signatures and nominal resources. Incoming HTTP uses
that view to reject wrong signatures before resource publication or effects.

Pinned Clang 22.1.8 / wasi-libc 33 and Rust 1.98.0 independently compile the C
and Rust guests with wit-bindgen 0.54.0 and WASI 0.2.12 WIT. Rust calls the C
arithmetic component after wasm-tools 1.258.0 composition; the commands exercise
arguments, randomness, polling, preopened directory stat and verified stdout.
Both guests execute on both own backends and Go/LLGo release, and on actual
Wasmtime 48.0.1 WASIp2. The full stable CLI import fixture separately checks
binding after language DCE narrows unused imports. Seven valid CLI controls and
one incoming-handler result mutation exercise failure/exit/entry-signature
behavior and rejection before their callback effect. Unified HTTP client and
incoming calls also execute on both backends.

The source/dependency lock, WIT, source/binary checksums and runtime notices are
retained under testdata/wasi-preview2/guests and the fixture manifest. Producer
reproduction validates every binary; Go also checks manifest input/output hashes.
Local package tests, race/vet and no-cgo build pass. Component statement coverage
is **91.09% (982/1078)**; Preview 2 is **90.53% (4904/5417)**. Final-head four-job
CI, execution reports and coverage are required before merge. These results do
not close complete upstream world conformance, synchronous resource value/start
and local cm64-borrow gaps, M5 performance gates, or WASI 0.3/native async.

Three serial 1000-call samples after ten warmups use one read-only preopen,
identical args and bounded captured output. Every run verifies the exact output
bytes. Fuel is off; quota/cancellation checks remain. Table entries are warm
ns/command and include the command's real WASI, canonical and resource costs:

| Command | Go interpreter | Go native | LLGo interpreter | LLGo native | Actual Wasmtime WASIp2 |
| --- | ---: | ---: | ---: | ---: | ---: |
| C | 31,069 | 40,293 | 75,338 | 129,046 | 7,530 |
| Rust -> C composition | 40,280 | 50,784 | 90,160 | 151,488 | 7,171 |

LLGo is **9.81×/17.13× slower** for C and **12.57×/21.12×** for the composed
command. Performance remains below target. Go's native composed command
allocates 445 objects/call; LLGo's zero allocation count is unavailable data,
not zero allocation. Native bridging is slower than interpreter execution for
these host-call-heavy guests. Shared transfer/dispatch and allocation costs
remain explicit optimization work. Random algorithms and default provider
resource/scheduling policies differ; these local samples establish no controlled
regression guarantee. No pure arithmetic or equal random-provider claim is made.

Separate phases for the composed command, in ns:

| Phase | Go interpreter/native | LLGo interpreter/native | Wasmtime |
| --- | ---: | ---: | ---: |
| Load | 47,083 / 47,083 | 56,834 / 56,834 | 30,750 |
| Engine | 791 / 750 | 1,459 / 1,958 | 76,250 |
| Compile | 6,607,083 / 36,163,875 | 14,784,083 / 69,123,084 | 19,269,917 |
| Provider/Store setup | 76,000 / 82,083 | 80,417 / 101,750 | 127,375 |
| Bind/instantiate | 248,708 / 262,833 | 333,458 / 368,959 | 84,041 |
| Lookup | 291 / 333 | 333 / 292 | 1,083 |
| First call | 313,208 / 172,292 | 219,792 / 312,166 | 272,709 |
| Owner teardown | 26,625 / 29,792 | 140,666 / 142,459 | 107,500 |

Source base `b1f249e` lacks NewWorld under both compilers. Unchanged wall-clock
base/head medians are Go 824.8 -> 830.7 / 1256 -> 1275 ns and LLGo 2642 -> 2771 /
3946 -> 4209 ns. Raw logs, all phases, binary/source/compiler identities and
producer failures remain in artifacts/components/world-*.

The actual reference links/runs both language guests. Its complete CLI import
consumer instead fails with no matching wasi:sockets/network@0.2.12 linker
implementation; that diagnostic is retained without substituting a provider.
An initial harness build used obsolete DirPerms/FilePerms names; it now uses
48.0.1's actual FsPerms API. Initial Clang lacked a default Wasm compiler-rt
archive; the fixed producer explicitly links its required libc without default
runtime libraries, and component validation confirms resolved imports. Rust's
explicit generate_all enables the pinned external WIT package bindings. These
producer/harness failures remain distinct from own runtime execution results.


Final cleanup review reproduced a valid io.ReadCloser receiving two Close calls
on body timeout: the timer and pump/shutdown paths raced, and the second close
became a spurious provider cleanup error. A shared once-guarded reader now owns
that input across all three paths. The regression fails before the fix and
passes on timeout and explicit shutdown with Go/LLGo and race checks. Source
cooperation for interruptible Read/Close remains the HTTP embedding contract.
The earlier preflight at 4e457ca was canceled; the corrected final head requires
its own complete four-job matrix. Final command samples above were rerun.

The affected full HTTP workload was also rerun serially: three 20-batch samples
of 128 checked requests, 4096 bytes per direction and trailers on a shared
external server. Table entries are ns/request, source base 4e457ca versus the
reader-ownership fix; default connection-policy differences remain unchanged.

| Compiler/backend | Before | After | Actual Wasmtime HTTPp2 |
| --- | ---: | ---: | ---: |
| go/interpreter | 136,594 | 136,967 | 120,550 |
| go/native | 176,194 | 176,650 | 120,550 |
| llgo/interpreter | 343,274 | 346,142 | 120,550 |
| llgo/native | 445,260 | 428,861 | 120,550 |

All common-server message/trailer checks pass; counters remain in
world-http-performance-summary.json. Own connections are pooled, while the
reference opens a connection per request. This is complete-request policy cost,
not pure VM throughput. Local variance and the additional ownership guard do not
establish performance parity or a controlled regression guarantee. Earlier raw
command results are preserved with the before-reader suffix.

## Synchronous resource values and start ownership (2026-10-31)

The value extension now executes resource-bearing standalone imports/exports and
component starts through the existing canonical resource codec and Core Caller.
Factories run after static bindings/provider checks, can mint bound nominal
resources and retain aggregate initialization budgets. Aliases, narrowed views,
child components and nested host interfaces share affine value occurrences.
Resource exports move exactly once, including empty containers; resource-free
exports retain detached repeatable reads. Borrow imports stay pinned through
initialization and expire before publication. Transitive borrow value exports
are rejected according to the pinned Binary.md rule; exporting a borrow type is
still allowed. Equality-bound borrows compare live table/type/handle identity,
including separate tokens for the same lender.

Twelve project-authored binary fixtures retain their generator and hashes:
nine successful execution inputs, two valid inputs with deliberate later-start
failure, and one statically invalid borrow export. The pinned producer cannot
encode value/start text, so these are authored grammar/semantic tests, not an
independent conformance oracle. Successful guest creation, ownership transfer,
explicit guest destructors and failure cleanup execute on both own backends.
Tests also cover stale/foreign capabilities, duplicate own aggregates, nested
imports, factory failure/panic/cancellation, expired Callers, constraints,
resource/byte/value limits, concurrent extraction and failed initialization
cleanup with active host and guest destructors. Go and LLGo release-profile
component suites pass; Go race/vet pass. Local statement coverage is
`pkg/component` **91.79%** (1,085/1,182), `internal/component/validate` **91.81%**
(1,828/1,991).

Serial local phase probes use Apple M4 Max arm64, Go 1.27.3 and pinned LLGo
`8e45ed5` with `-O3 -lto=full`; medians of three 10,000-instance samples.
The fixture creates a real guest resource during start and explicitly drops it
through its guest destructor. Source base `eb6551d` rejects it at Compile.
Values below are microseconds; process startup/runner build are excluded:

| Phase | Go interpreter / native | LLGo interpreter / native |
| --- | ---: | ---: |
| File load | 30.125 / 34.667 | 56.958 / 50.458 |
| Engine creation | 1.541 / 1.542 | 3.416 / 3.000 |
| Compile | 114.459 / 219.250 | 157.708 / 703.250 |
| First Store/bind/instantiate/start | 28.041 / 30.416 | 66.500 / 78.375 |
| First value extraction | 1.500 / 1.208 | 2.250 / 1.875 |
| Repeated Store/bind/instantiate/start | 6.921 / 7.781 | 22.341 / 24.258 |
| Repeated value extraction | 0.393 / 0.391 | 1.601 / 1.664 |
| Repeated explicit drop/destructor | 0.320 / 0.721 | 0.989 / 2.562 |
| Repeated close | 0.096 / 0.279 | 0.228 / 0.562 |

The public Instantiate API combines Store creation, composition, binding and
start; preparation is not pure guest execution. Affine extraction requires a
new instance. Raw logs separately retain first drop/close. Actual unmodified
Wasmtime 48.0.1 rejects the fixture with `support for component model values is
not enabled`; no direct ratio for the experimental extension is available.

The existing common `ownership.wasm` workload separately executes four public
calls per complete make/get/pass/consume lifetime. Medians of three samples,
10,000 own lifetimes and 100,000 reference lifetimes; nanoseconds/lifetime:

| Compiler | Base interpreter / native | Head interpreter / native |
| --- | ---: | ---: |
| Go | 3,041 / 5,326 | 3,010 / 5,149 |
| LLGo release | 8,939 / 14,925 | 8,961 / 14,954 |
| Actual Wasmtime 48.0.1 component C API | 1,244.79 | 1,244.79 |

This common workload shows no material local LLGo regression (about +0.2%),
while the reference remains substantially faster. Warm timers exclude load,
engine/compile, instance preparation, lookup, first lifetime and close. Own fuel
is off with logical resource/cancellation checks; the reference uses its default
Store policy. This is descriptive local evidence, not controlled performance
acceptance. Raw per-phase samples, reference rejection, source-base outcomes,
runner identity and common-workload logs are retained in
`artifacts/components/resource-values-*`.

Local cm64 borrows, abstract host/component bindings, complete upstream world
conformance, compiler/interpreter performance acceptance and native async/WASI
0.3 remain required. This increment does not complete M5/M6 or the project.

## Equality-bound non-resource type imports (2026-10-31)

Function, instance and component type imports now bind as checked equality
metadata at root and nested host interfaces, matching existing defined-value
bindings. They are already resolved static targets, so they do not introduce a
host-selected runtime identity. Two broad runtime rejections are removed; no
new type registry or alternate execution framework is added. Resource imports
retain explicit nominal declarations and equality checks; actual host component
imports remain rejected before effects.

Three WAT/Wasm fixtures retain source/binary hashes and pass unmodified
wasm-tools 1.258.0 parse/validation. Both own backends execute a child component's
Core addition through imported function/instance/component type aliases and
checked ascriptions. Nested host type exports work; a resource-capturing
function alias preserves nominal host callback checks and borrow/drop behavior.
Go/LLGo release suites, race and vet pass. Local `pkg/component` statement
coverage is **91.86%** (1,084/1,180).

Actual Wasmtime 48.0.1 rejects the complete fixture with `unsupported type
export`; separate minimal inputs containing only an equality-bound function,
instance or component type import produce the same rejection. This is a backend
support limit, distinct from pinned wasm-tools validation. Source base
`ccfa657` rejects the new inputs before execution too. No direct speed ratio is
claimed for an unsupported capability.

Local Apple M4 Max arm64 medians of three 1,000-operation samples; Go 1.27.3,
LLGo `8e45ed5 -O3 -lto=full`; microseconds:

| New type-import fixture phase | Go interpreter / native | LLGo interpreter / native |
| --- | ---: | ---: |
| Compile and compiled-owner Close | 20.911 / 32.286 | 39.322 / 64.946 |
| Store/bind/instantiate/Close | 5.886 / 5.905 | 10.562 / 11.154 |
| Public warm add | 0.437 / 0.695 | 1.234 / 1.932 |
| Shared-session warm add (diagnostic) | 0.355 / 0.554 | 0.819 / 1.386 |

Load/engine creation and runner build/startup are outside these timers. Each
first call is separately recorded outside warm timers in the raw benchmark
logs. Combined preparation includes explicit Close and is not pure binding or
compile throughput. The shared-session diagnostic amortizes one existing Core
session; it is not public API latency.

The unchanged composition control separately compares source base/head with
actual Wasmtime's component C API. Medians of three samples, 10,000 own calls
and 1,000,000 reference calls; nanoseconds/public warm call:

| Compiler | Base interpreter / native | Head interpreter / native |
| --- | ---: | ---: |
| Go | 461.6 / 777.9 | 452.8 / 796.9 |
| LLGo release | 1,468 / 2,239 | 1,474 / 2,207 |
| Actual Wasmtime 48.0.1 | 244.2 | 244.2 |

LLGo changes approximately +0.4% / -1.4% in this local control; parity remains
unmet. Own fuel is off with cancellation/quotas retained; reference Store policy
is its default. These are uncalibrated observations, not controlled performance
acceptance. Raw first/preparation/warm/control samples, source-base outcomes and
isolated reference rejections are retained in `artifacts/components/type-imports-*`.
Host component bindings, local cm64 borrows, complete upstream world conformance,
performance acceptance and native async/WASI 0.3 remain open; M5/M6 are incomplete.

## Host component imports and provider boundaries (2026-10-31)

`Imports.Components` now binds independently compiled roots in the same Engine,
including nested host interfaces and component arguments. The existing executor
uses each closure's correct embedded Core map. One bounded schema registry
reuses ordinary subtype/binder relations, remaps source resource/value symbols
and imports canonical recursive Core/GC groups without mutating compiled pools.
Compilation leases preserve in-flight bindings and execution after owner Close;
new bindings reject closed/foreign/zero owners. Root compilation uses the same
lease rule. Failure, cancellation, panic, instance limits and concurrent close
release leases and unpublished ownership.

A failing regression exposed imported host functions sharing the consumer's
canonical failure guard. A healthy imported component's destructor could not
call its provider after the consumer trapped. Provider namespaces now share
resolved nominal bindings while retaining independent execution guards. The
regression fails on both own backends before the change and passes afterwards;
a direct re-exported host alias remains callable after its consumer traps.
Canonical guest entry/leave/realloc/post-return/reentry guards remain intact.

Sixteen fixtures retain source/binary/generator hashes. Fourteen are ordinary
pinned wasm-tools 1.258.0 parse/validation outputs. Two append authored value/start
sections to producer fragments and additionally pass independent binary
validation with `wasm-tools validate --features cm-values`. These are not an
independent runtime oracle. Tests execute repeated generative resources,
narrowed views, nested/delegated component arguments, Core-module parameters
whose recursive/GC identities differ between pools, table/global/tag descriptors,
real struct allocation/access, standalone owned values and explicit/failed-start
destructors. They also cover equality/literal constraints, schema budgets,
owner close during binding, cancellation, panic, rollback, concurrent Stores,
forced collection and code lifetime.

Full local Go package tests and the relevant LLGo release suites pass. Race/vet
and pinned corpus checks pass. Full local darwin/arm64 coverage is component
**91.90%** (1,135/1,235), validator **92.28%** (1,995/2,162), Core module metadata
**98.06%** (253/258); every executable package exceeds 90%.

Actual unmodified Wasmtime 48.0.1 rejects the consumer before instantiation with
`root-level component imports are not supported`. Its pinned inliner contains
that explicit boundary. Source base `297e30e` also rejects host component imports.
No direct ratio is assigned to an unsupported operation. Local phase and
common-workload measurements follow below; raw results remain in
`artifacts/components/host-components-*`.

Apple M4 Max arm64 phase medians, three serial 10,000-operation samples;
Go 1.27.3, LLGo `8e45ed5 -O3 -lto=full`; microseconds:

| Imported-component phase | Go interpreter / native | LLGo interpreter / native |
| --- | ---: | ---: |
| Load consumer and guest | 69.709 / 67.791 | 92.667 / 84.625 |
| Engine creation | 1.667 / 1.709 | 3.417 / 3.084 |
| Guest compilation | 96.292 / 166.417 | 96.458 / 213.917 |
| Consumer compilation | 38.000 / 76.917 | 59.917 / 487.000 |
| First Store/bind/instantiate | 59.000 / 47.459 | 360.166 / 65.042 |
| Repeated Store/bind/instantiate | 10.856 / 11.591 | 34.373 / 34.861 |
| First public add | 19.166 / 27.250 | 12.750 / 33.875 |
| Public warm add | 0.451 / 0.863 | 1.872 / 2.955 |
| Repeated instance Close | 0.180 / 0.231 | 0.568 / 0.724 |

Lookup and final Close remain separate in raw JSON. Binding includes schema
matching and composition; it is not pure guest execution. Build/process startup
is outside these timers. Initial 10,000-call controls showed large variance and
are preserved as diagnostics. Longer alternating base/head controls use three
100,000-call add samples and three 1,000-batch HTTP fields samples (128 complete
lifetimes/batch). Nanoseconds per call/lifetime:

| Common workload | Go base -> head interpreter / native | LLGo base -> head interpreter / native | Actual Wasmtime 48.0.1 |
| --- | ---: | ---: | ---: |
| Public composition add | 470.4 -> 480.9 / 837.7 -> 801.2 | 1,436 -> 1,456 / 2,320 -> 2,328 | 248.7 |
| Actual HTTP fields lifetime | 24,367 -> 24,876 / 34,467 -> 35,229 | 59,890 -> 56,038 / 105,291 -> 106,202 | 2,629.3 |

The add reference uses the actual component C API; HTTP uses the unmodified
Wasmtime WASIp2 HTTP provider with its typed Rust API. Reference samples use
1,000,000 add calls and 1,000 HTTP batches. Warm timers exclude preparation,
lookup, first call and Close. Own fuel is off with logical cancellation/resource
checks; the reference uses its default Store policy. Long local LLGo controls
change about +1.4%/+0.3% for add and -6.4%/+0.9% for fields. These are noisy local
observations, not controlled acceptance. The reference remains substantially
faster; compiler/interpreter and Canonical ABI performance targets stay open.
Local cm64 borrows, complete upstream world conformance and native async/WASI 0.3
remain required. This increment does not complete M5/M6 or the project.

## Upstream WASI standard-library interoperability (2026-10-31)

Fourteen unmodified Wasmtime 48.0.1 test programs, pinned at
`7bac2c2775808aaec5d4aa5627a5e447b51102cf`, compile with Rust 1.98.0's
wasm32-wasip2 target (edition 2024, `-O -C panic=abort`). The producer strips
custom sections and independently validates every component with wasm-tools
1.258.0. Source/license, toolchain archive and output hashes are retained in
`upstream-manifest.json`; rebuilding does not alter any upstream guest body.

The standard libraries import WASI 0.2.9 and export CLI run 0.2.0. Source base
`a8aef41` fails at the first 0.2.9 host instance. Stable interface lookup now
connects these guests to the existing checked 0.2.12 World: exact names win,
otherwise the greatest stable compatible version wins. Prerelease/0.0.x
fallback is excluded; numeric ordering and build-metadata ties are deterministic.
Import scans spend the aggregate binding budget. Existing contracts, nominal
resources, failure cleanup and plain-name behavior remain enforced.

The actual append program also found descriptor mutation permissions being
confused with the open-flags directory constraint. Ordinary writable files now
open correctly, while explicit directory constraints/read-only grants remain
checked. A separate sustained probe found the project-authored C control guest
not dropping descriptors inside a freed list. It now explicitly drops them;
three repeated C/Rust commands pass with a two-descriptor provider quota, where
the original C fixture fails. The original 10,000-call quota failure is preserved.
No provider limit was raised. Corrected identical guest bytes are supplied to
both source revisions and the reference.

All fourteen upstream programs execute with expected outcomes on both own
backends under Go/LLGo release, and on the actual unmodified Wasmtime WASIp2
provider. This includes empty/Unicode args, absent argv0, environment, stdio,
clocks/sleep, normal/nonzero exit, panic, file reads, seek-preserving append and
sync. Nonzero exit and panic remain explicit outcomes. The independent typed
reference accepts empty args as well as the CLI-tested programs.

Full Go tests, final relevant Go/LLGo suites, race/vet and corpus/provenance checks
pass. Local full-production coverage: component **92.00%** (1,162/1,263), validator
**92.33%** (2,009/2,176), P2 **90.51%** (4,905/5,419); all executable packages
remain >=90%. These are package statement metrics, not specification coverage.

Apple M4 Max arm64; LLGo `8e45ed5 -O3 -lto=full`, Go 1.27.3. Three serial samples
per first-call workload, fuel off with own logical quotas/cancellation retained;
reference default Store policy. Commands are one-shot and may exit or trap.
Build/process startup is outside the timers. Microseconds:

### Rust standard-library arguments

| Runtime | Load | Engine | Compile | Provider / provider+Store (ref) | Store/bind/instantiate / instantiate (ref) | Lookup + signature | First call | Close |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 63.042 | 0.750 | 14,465.959 | 100.834 | 352.584 | 1.333 | 90.833 | 38.917 |
| go-native | 69.166 | 0.792 | 75,546.541 | 115.834 | 293.875 | 1.292 | 99.208 | 36.334 |
| llgo-interpreter | 132.666 | 1.166 | 31,894.041 | 163.708 | 613.708 | 1.666 | 133.250 | 332.250 |
| llgo-native | 81.958 | 0.833 | 146,145.125 | 183.375 | 487.500 | 1.708 | 177.875 | 328.500 |
| wasmtime | 47.000 | 90.334 | 40,493.500 | 143.041 | 90.750 | 0.875 | 36.375 | 128.000 |

### Rust standard-library file append

| Runtime | Load | Engine | Compile | Provider / provider+Store (ref) | Store/bind/instantiate / instantiate (ref) | Lookup + signature | First call | Close |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 65.958 | 2.958 | 17,710.625 | 109.959 | 387.667 | 1.292 | 353.292 | 28.583 |
| go-native | 61.208 | 1.584 | 92,235.542 | 124.000 | 354.416 | 1.541 | 415.750 | 33.625 |
| llgo-interpreter | 83.041 | 1.541 | 35,476.542 | 161.334 | 3,104.250 | 2.750 | 775.083 | 274.875 |
| llgo-native | 72.459 | 0.791 | 162,914.875 | 165.916 | 560.666 | 1.500 | 653.333 | 299.625 |
| wasmtime | 42.750 | 85.542 | 47,971.083 | 151.916 | 86.459 | 0.917 | 598.500 | 151.167 |

Own binding includes Store creation; the reference provider setup includes its
Store and linker. First invocation excludes lookup and preparation; it includes
real standard-library/provider work, not only guest arithmetic. No warmed
throughput ratio is inferred from these first-call numbers.

Common corrected C/Rust command controls use three alternating base/head samples,
10,000 warmed calls each. Preparation, lookup, first call and teardown are outside
warm timers. Final test-binary relink medians, microseconds per command:

| Command | Go base -> head interpreter / native | LLGo base -> head interpreter / native | Actual Wasmtime 48.0.1 |
| --- | ---: | ---: | ---: |
| C | 31.492 -> 31.017 / 41.484 -> 40.965 | 119.917 -> 98.951 / 187.499 -> 164.761 | 24.108 |
| Rust-to-C | 39.808 -> 40.072 / 50.673 -> 51.275 | 156.932 -> 124.715 / 294.352 -> 202.511 | 19.225 |

Earlier diagnostic medians are retained separately: LLGo C 72.255/128.292 ->
71.414/128.574 us, Rust-to-C 87.507/148.635 -> 88.034/146.161 us; reference
7.905/7.203 us. Final relinking and active host background indexing accompanied
large run-to-run variance, including the independent reference. There is no
causal optimization claim or controlled acceptance from these local timings.
The reference remains substantially faster; performance gates remain open.

Reproduce own one-shot phases with `benchmarks/components/upstream.go` and the
actual typed reference with `benchmarks/components/rust/src/bin/upstream.rs`.
`BenchmarkWASICommand` and the existing `command` reference reproduce common
controls with the same corrected fixture bytes. Raw final samples, phase JSON,
earlier diagnostic medians and explicit failures remain under
`artifacts/components/upstream-*`.

This is a bounded upstream CLI subset. Full upstream/world conformance, local
cm64 borrow conversion, performance acceptance and native async/WASI 0.3 still
remain required. M5/M6 and the project are incomplete.

## Error-context execution (2026-10-31)

The owned compiler/runtime executes all three pinned diagnostic builtins and
copyable error-context values through existing flat/memory aggregate codecs,
synchronous host callbacks, imported components and standalone value bindings.
Both own backends execute Unicode encodings and optional memory64, typed handle
rejection, byte/live-handle quotas, post-return may-leave rejection, cancellation,
host panic and rollback. A mutable host reference carrier cannot mutate the
retained diagnostic object. Shared table/budget tests include concurrent
namespaces and unique-message accounting. No external engine fallback is used.

Nine pinned fixtures retain source/output/generator/helper SHA-256. Seven pass
independent wasm-tools 1.258.0 validation. The producer emits an i32 builtin for
cm64 despite the pinned memory64 signature, and its value-import syntax differs
from the project's pinned extension; both exact rejections are preserved. Those
two fixtures execute on the own runtime but are not an independent oracle.
Actual unmodified Wasmtime 48.0.1 executes guest new/drop with message capture
and the forbidden post-return case traps. The cross-component sample traps
with cannot-enter and the cm64 sample is rejected; neither is reported passing.
Source base `af798f4` rejects the new canonical operations before execution.

Full Go tests and relevant Go/LLGo release suites, race and vet pass. Local full
production coverage: canonical **94.24%** (1,341/1,423), component validator
**92.43%** (2,040/2,207), component runtime **92.20%** (1,206/1,308). Every
executable package remains >=90%. Package statement coverage is not specification
coverage. These results advance the first M6 primitive, not task/world completion.

Apple M4 Max arm64, Go 1.27.3 and LLGo `8e45ed5 -O3 -lto=full`; three serial
samples, 3,000 warmed calls with 128 complete context lifetimes per call. Both
implementations capture the same UTF-8 input, with no discard substitution.
Own logical quotas/cancellation remain enabled and fuel is off; reference Store
policy is default. Microseconds except the last column (nanoseconds/lifetime):

| Runtime | Load | Engine | Compile | Store + instantiate | Lookup | First call | Close | Warm context |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 41.417 | 1.584 | 239.625 | 41.084 | 0.375 | 74.167 | 3.125 | 426.908 |
| go-native | 44.000 | 1.417 | 668.875 | 42.000 | 0.416 | 109.167 | 16.542 | 644.180 |
| llgo-interpreter | 50.750 | 2.959 | 580.292 | 88.084 | 0.417 | 119.375 | 3.125 | 1435.063 |
| llgo-native | 55.667 | 2.958 | 1,622.583 | 90.125 | 0.458 | 233.917 | 14.833 | 2376.130 |
| wasmtime-48.0.1 | 42.208 | 91.583 | 2,049.666 | 54.917 | 2.833 | 16.166 | 63.667 | 66.134 |

First call is one batch. Build/process startup, lookup, preparation and Close are
outside warm execution. Guest instructions, public batch entry and the two
canonical callbacks remain in each complete lifecycle; no pure-dispatch or
whole-process efficiency ratio is inferred. The optional Rust reference enables
the pinned concurrent-runtime feature required by error-context, independently
of earlier P2 reference profiles.

The common unchanged own/borrow control runs 100,000 four-call lifetimes per
sample, three alternating source base/head samples. Startup, lookup and first
lifetime are outside warm timers. Nanoseconds per complete lifetime:

| Host | Interpreter base -> head | Native base -> head | Actual Wasmtime C API |
| --- | ---: | ---: | ---: |
| go | 2,882 -> 2,935 | 5,239 -> 5,214 | 1,194.52 |
| llgo | 8,628 -> 8,824 | 14,649 -> 14,721 | 1,194.52 |

Local LLGo common controls change about +2.3%/+0.5%. The reference is much
faster for the new lifecycle as well as the common path. These uncalibrated
observations are not parity or optimization acceptance; compiler/interpreter
and canonical callback performance remain explicit open work. Raw samples,
producer/reference rejections and support probes are retained under
`artifacts/components/error-context-*`. Future/stream/task scheduling, complete
WASI 0.3.2 worlds, local cm64 borrows and earlier conformance/performance gates
remain required. M5/M6 and the overall project are incomplete.


## Run-to-completion async lifts (2026-10-31)

Implementation `d2427c8`, source base `88e1c4e`: async function types, synchronous
lift/lower of those types, stackful task.return and stackless callback EXIT/YIELD
execute on both own backends. The same Store session, value codecs and handle
budgets remain shared. TLS is isolated across initialization, realloc, nested
tasks and calls; borrow validation precedes task.return and post-return effects.
Explicit async host bindings and synchronous WASI 0.2 export checks preserve
versioned interface contracts. Async lower/subtasks, suspended waitables,
cooperative task cancellation, futures/streams and WASI 0.3.2 worlds remain open.

Fourteen authored fixtures independently validate with wasm-tools 1.258.0 against
the grammar pinned at Component Model `d095e2f`. This includes cm32/cm64, all
three string encodings, direct/17-slot indirect transfers, memory aliases,
instantiated resource aliases, ownership, borrow release and failure cleanup.
Guest realloc preserves old bytes during reference UTF-16 transcoding. Full Go
and affected pinned LLGo release tests pass, with race/vet and cgo-free builds.
Coverage: canonical 94.26%, validator 92.69%, component 92.62%; all 29 executable
packages remain at least 90%. Coverage is not complete async conformance.

Actual Wasmtime 48.0.1 matches 30 independently invoked positive/trapping outcomes.
Two precise reference boundaries are retained: its compiler rejects i64 context
with `unsupported context.get type: I64`; its cm64 high-pointer task.return reads
`diagnostic-λ` from offset zero for pointer 4294967296 instead of trapping. Both
own backends reject that out-of-bounds pointer. These two outcomes are not
counted as passed reference conformance. Complete upstream async/world coverage
remains required.

Local Apple M4 Max arm64 diagnostic: Go 1.27.3; LLGo `8e45ed5` rebuilt from its
clean pinned checkout with `-O3 -lto=full`; LLVM 22.1.8; Rust 1.98.0 and actual
Wasmtime 48.0.1. An unrelated moving development LLGo checkout was detected and
excluded before final verification. Source base/head use the same fixed binary
and flags. Three serial samples alternate runtime/revision order. Each new
probe performs 3,000 calls after first-call and ten warmups; argument 128 yields
129 callback invocations per callback task. First-call and all preparation
phases are outside warmed timing. Process/runtime initialization before main is
excluded; Wasmtime's separately built Tokio executor is timed explicitly.

Warm **ns per complete public call**:

| Runtime | Callback (129 turns) | Stackful | Synchronous Core ABI |
| --- | ---: | ---: | ---: |
| go-interpreter | 46,790 | 912 | 542 |
| go-native | 117,940 | 2,109 | 1,294 |
| llgo-interpreter | 175,644 | 3,684 | 2,384 |
| llgo-native | 437,724 | 7,279 | 4,497 |
| wasmtime-48.0.1 | 40,736 | 1,080 | 716 |

The amortized callback figures include task entry/return, the guest loop,
context get/set, VM/host boundaries and callback dispatch. They are not pure
opcode or scheduler costs. LLGo callback interpreter/native figures are 1,361.6 / 3,393.2 ns per turn, versus 315.8 ns for the reference.

Callback probe preparation/first/close phases, **ns**:

| Runtime | Load | Executor | Engine | Compile | Store + instantiate | Lookup | First | Close |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 34,542 | 0 | 1,750 | 220,708 | 30,292 | 209 | 58,583 | 2,500 |
| go-native | 35,375 | 0 | 1,375 | 601,000 | 34,583 | 334 | 139,750 | 4,750 |
| llgo-interpreter | 62,667 | 0 | 2,625 | 704,583 | 96,083 | 375 | 107,625 | 2,667 |
| llgo-native | 56,125 | 0 | 3,083 | 1,602,750 | 73,833 | 417 | 293,416 | 9,375 |
| wasmtime-48.0.1 | 23,958 | 13,542 | 84,833 | 2,091,916 | 22,292 | 3,083 | 78,791 | 65,792 |

A zero executor field means no separately constructed host executor. The owned
probe keeps logical quotas and context cancellation enabled, with fuel off;
Wasmtime uses its ordinary finite-fixture defaults with concurrency enabled.
This is a local diagnostic, not matched-policy/all-target acceptance. The source
base rejects task.return before guest execution; it has no new-feature timing.

The common four-call owned resource lifetime uses 100,000 complete lifetimes per
sample, with startup/lookup/first excluded. Base -> head medians, **ns/lifetime**:

| Compiler | Interpreter | Native |
| --- | ---: | ---: |
| go | 2,870 -> 2,882 | 5,165 -> 5,213 |
| llgo | 8,630 -> 8,725 | 14,576 -> 14,491 |

Actual pinned Wasmtime C API reference: **1,207.32 ns/lifetime**. An initial
LLGo control regression of 15–17% was removed by retaining the ordinary
synchronous deferred exit and saving nested state inside the existing task.
Final local control differences are small; no confidence-bound acceptance is
claimed. LLGo's absolute competitor deficits and M1–M6 performance gates remain
open. New task mechanisms do not close native async/WASI 0.3 functionality.

Reproduction: [async probe instructions](../benchmarks/components/README.md).
Raw `async-lift-phase-*`, `async-lift-common-*`, reference outcomes, binary hashes,
toolchain metadata and the initial regression diagnostic remain under
`artifacts/components/async-lift-*`.


## Eager async canonical lower (2026-10-31)

Implementation `43211d2`, source base `7491e45`: async lower uses four flat
argument slots and indirect results, shares the ordinary call bridge and returns
RETURNED (2) after admitted run-to-completion callees finish. It covers sibling
component calls, synchronous and stackful lift ABIs, mixed numeric slots,
spilled arguments, strings and nominal own/borrow transfers. Pending subtasks,
waitable producers, cooperative cancellation, backpressure, futures/streams and
WASI 0.3.2 worlds remain required; M6 is open.

The shared lower transaction now adopts returned ownership before validating
result memory. A direct source-base reproduction of the synchronous ABI traps
but drops zero unpublished results; the candidate traps and drops exactly one.
Host effects precede a result-pointer failure; malformed spilled arguments fail
before host effects. Tests also cover panic, cancellation, bad result shapes,
expired loans and complete shared-quota reclamation on both own backends.

Twelve authored fixtures pin Component Model `d095e2f` and wasm-tools 1.258.0.
Nine independently validate; three precise cm64 async-lower producer rejections
are retained because the producer expects i32 result pointers instead of i64.
The owned cm64 cases follow the pinned optional memory64 rule and are not
independent conformance evidence. Actual Wasmtime 48.0.1 matches 20 component
and string invocations and independently rejects those three cm64 signatures.
Full Go, affected fixed-LLGo release tests, race/vet and cgo-free builds pass.
Coverage: canonical 94.28%, validator 92.74%, component 92.72%; all 29 executable
packages remain >=90%. This is not full async/world conformance acceptance.

Local Apple M4 Max arm64: Go 1.27.3; LLGo `8e45ed5 -O3 -lto=full` from the clean
pinned checkout; LLVM 22.1.8; Rust 1.98.0; actual Wasmtime 48.0.1. Source base and
head use identical toolchains and flags. Three serial samples alternate runtime
and revision order. Each new probe runs 3,000 public calls, each containing 128
checked cross-component async lowers, after first call and ten warmups. The
source base rejects async lower before guest execution, so it has no new-feature
timing. Process/runtime initialization before main remains excluded.

Warm **ns per complete 128-lower call**, and explicitly amortized **ns/lower**:

| Runtime | Warm call | Amortized lower |
| --- | ---: | ---: |
| go-interpreter | 109,356 | 854.35 |
| go-native | 186,177 | 1,454.51 |
| llgo-interpreter | 348,113 | 2,719.64 |
| llgo-native | 613,483 | 4,792.84 |
| wasmtime-48.0.1 | 139,401 | 1,089.07 |

Amortized timing includes the guest loop/checks, canonical ABI transfers, the
callee, task.return and VM/host transitions. It is not a pure opcode cost.
Preparation, first batch and close remain separate, **ns**:

| Runtime | Load | Executor | Engine | Compile | Store + instantiate | Lookup | First | Close |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 42,750 | 0 | 1,750 | 217,542 | 33,875 | 250 | 144,500 | 3,208 |
| go-native | 44,584 | 0 | 1,625 | 346,334 | 59,250 | 417 | 232,292 | 6,959 |
| llgo-interpreter | 47,375 | 0 | 3,083 | 513,250 | 78,083 | 375 | 233,875 | 3,209 |
| llgo-native | 61,084 | 0 | 2,750 | 826,292 | 73,167 | 416 | 776,250 | 11,583 |
| wasmtime-48.0.1 | 30,334 | 14,667 | 87,791 | 1,959,958 | 39,459 | 2,542 | 208,708 | 17,632,750 |

The separate executor field measures reference Tokio construction; Go runtime
initialization is outside this helper. Fuel is off; owned quotas and context
cancellation remain enabled, while the reference uses its finite-fixture
defaults. These local rows are not matched-policy/all-target acceptance. The
reference close cost is retained rather than folded into steady-state timing.

The ordinary four-call resource lifetime is a common base/head control: 100,000
complete lifetimes/sample, with preparation/lookup/first excluded. Medians,
**ns/lifetime**, base -> head:

| Compiler | Interpreter | Native |
| --- | ---: | ---: |
| go | 3,093 -> 2,933 | 5,869 -> 5,486 |
| llgo | 9,023 -> 8,723 | 15,163 -> 14,646 |

Actual Wasmtime C API reference: **1,191.84 ns/lifetime**. Local control medians
do not establish a controlled improvement. LLGo's absolute competitor deficits
and earlier milestone performance gates remain open. The probe reuses the async
lift helper with its new `churn` entry; no separate benchmark runtime is added.

Raw `async-lower-phase-*`, `async-lower-common-*`, toolchain/binary hashes,
source-base capability/rollback reproductions and reference outcomes remain in
`artifacts/components/async-lower-*`. Reproduce with the
[component probe instructions](../benchmarks/components/README.md).

## Pending callback subtasks (2026-10-31)

Source base `1de979f84f6e7009814ef19debdb593791eea301`; runtime implementation
`3d8156fff070459cf93a4a81ccaead25b62d13f6`; measurement/helper commit
`c2f6a5b8340c915591aeb93c3b0f623930c16761`. Later evidence-only edits do not
change the runtime or measured helpers.

Callback-lifted functions now suspend after YIELD/WAIT, return real async-lower
subtasks and resume through the existing Core session. Waitable membership and
ready queues share the tagged resource handle domain. Cooperative cancellation,
resolve delivery, explicit drop, borrow lifetimes, task/handle limits and failed
session cleanup execute on both owned backends. No goroutine or duplicate VM
provides the continuation. `Config.MaxTasks` defaults to 1,024.

Eleven fixtures independently validate with wasm-tools 1.258.0 against Component
Model `d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`. They cover three pending
sibling calls, set migration, callback WAIT, immediate/deferred/returning
cancellation, three string encodings, resource handoff and rollback, and memory64
event payloads. Negative paths include wrong handle kinds, early/duplicate drop,
joined cancellation, invalid event/result pointers, synchronous suspension,
panic, host errors, cancellation and close. A task-entry contract test verifies
that cancellation under implicit backpressure never reads argument memory.

Full local Go execution and affected pinned LLGo execution pass, along with race,
vet and `CGO_ENABLED=0` builds. All 29 executable packages meet 90% statement
coverage; final-head CI is audited separately in the PR. These are coverage and
execution results for the declared profile, not full async conformance.

Actual Wasmtime 48.0.1 records 47 guest outcomes: 45 satisfy the common contracts;
two expose a memory64 disagreement. Both owned backends reject an event pointer
of `0x1_0000_0000` into a one-page memory. The reference returns NONE for both
`poll(4294967296)` and a guest-internal i64 constant, while rejecting ordinary
out-of-range/unaligned pointers. The constant reproduction avoids CLI argument
conversion. These disagreements remain failures of reference equivalence.
Async cancellation may return BLOCKED before resolving instead of resolving
eagerly; the probes retain the allowed scheduling alternatives.

### Local phase and throughput evidence

Apple M4 Max arm64; Go 1.27.3; clean LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f -O3 -lto=full`; Rust 1.98.0;
Wasmtime 48.0.1. Three serial samples alternate runtime and base/head order.
Each sample checks 3,000 calls of 128 complete pending lifetimes, excluding the
first call and ten warmups. Each lifetime includes the guest loop/checks, one
callback, canonical conversion, handle/set creation/join/wait/drop, result
publication and VM/host transitions. It is not a pure opcode or isolated lower
cost. The helper explicitly records 128 logical guest callbacks per outer call.

| Runtime | ns/128-lifetime call | Amortized ns/complete lifetime |
| --- | ---: | ---: |
| go-interpreter | 309,263.89 | 2,416.12 |
| go-native | 609,991.49 | 4,765.56 |
| llgo-interpreter | 1,206,751.97 | 9,427.75 |
| llgo-native | 2,233,168.76 | 17,446.63 |
| wasmtime-48.0.1 | 244,634.76 | 1,911.21 |

All other phases are separate (**ns**, medians); process startup is excluded.

| Runtime | Load | Executor | Engine | Compile | Store+instantiate | Lookup | First | Close |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-interpreter | 37,209 | 0 | 2,334 | 399,084 | 70,459 | 417 | 360,291 | 3,084 |
| go-native | 33,667 | 0 | 2,333 | 1,397,416 | 70,542 | 333 | 1,025,500 | 7,958 |
| llgo-interpreter | 59,375 | 0 | 3,333 | 871,084 | 106,333 | 458 | 1,299,584 | 2,833 |
| llgo-native | 56,209 | 0 | 3,000 | 3,480,625 | 117,167 | 417 | 2,202,084 | 7,791 |
| wasmtime-48.0.1 | 28,000 | 14,042 | 107,792 | 4,044,583 | 38,416 | 3,292 | 320,500 | 108,125 |

Existing callback control: 1,000 calls/sample, each performing 129 callbacks
without pending subtask/set management (**ns/call**, interpreter/native).

| Compiler | Source base | Head | Observed change |
| --- | ---: | ---: | ---: |
| go | 46,075.50 / 115,619.54 | 48,396.38 / 116,776.54 | +5.04% / +1.00% |
| llgo | 176,525.71 / 440,016.08 | 197,512.75 / 469,558.12 | +11.89% / +6.71% |

The actual reference control is 39,908.29 ns/call.

Common resource control: 100,000 four-call lifetimes/sample (**ns/lifetime**,
interpreter/native).

| Compiler | Source base | Head |
| --- | ---: | ---: |
| go | 2,837 / 5,081 | 2,892 / 5,041 |
| llgo | 8,531 / 14,506 | 9,009 / 14,796 |

The source base rejects the required async builtin profile before guest
execution; no unsupported base throughput is invented. Reference flags enable
async, stackful async, more async builtins, threading grammar and concurrency;
the workload does not execute component threads. Owned logical quotas and
cancellation stay on with fuel off; reference defaults differ. Uncontrolled
local medians do not establish confidence intervals, matched-policy or all-target
acceptance. LLGo competitor deficits and observed callback/resource regressions
remain open.

Raw local evidence is retained in `artifacts/components/pending-{toolchain,performance}.json`,
`pending-phase-*.json`, `pending-control-*.json`, `pending-common-*.log`,
`pending-reference-check.json`, `pending-base-unsupported.log`, and the local
Go/LLGo/coverage/race logs. Compiler, fixture, helper, lockfile and binary hashes
are retained. The PR contains the concise comparison and final CI audit.

Arbitrary stackful suspension, async host completion, explicit backpressure,
futures/streams and WASI 0.3.2 worlds remain required. This increment does not
complete M6 or its performance gate.

## Suspended component tasks and host completion — 2026-10-31

The owned Core continuation API parks interpreter/native frames at suspendable
host boundaries. Shared frame/slot quotas, fuel, roots, fresh Callers, abandonment
and session-close cleanup have execution tests on both backends. Component
callback/stackful/synchronous-lifted async calls use that same API, including
synchronous lower/wait/cancel and explicit 16-bit backpressure. Five new fixtures
are generated by `build_async_completion_fixtures.py`, independently parsed and
validated by wasm-tools 1.258.0, and hashed with their owned source and generator
in `async-completion.json`.

Local Go race tests and pinned LLGo `8e45ed5` with `-O3 -lto=full` execute the
suspension and external-host completion contracts on both backends. Tests cover
fresh/expired Callers, externally signaled wakeups, eager/pending completion,
cooperative cancellation, duplicate/reused completions, wrong result shape,
errors/panics, owned and borrowed resource lifetime, aggregate pending-value and
byte quotas, quota restoration, context cancellation and instance shutdown.
Cancellation may resolve eagerly or return BLOCKED before an event; both legal
outcomes are retained. Counter underflow/overflow and cancel-before-entry with
invalid spilled argument/result pointers are actual guest executions.

Actual Wasmtime 48.0.1 matches eight common semantic outcomes: both stackful
churn contracts, sync/async cancellation, backpressure entry/cancellation, and
counter underflow/overflow. These use the pinned async/stackful/more-builtins,
threading-grammar and component-memory64 flags. CLI reference execution cannot
supply the custom asynchronous host imports, so those host embedding tests are
owned-engine evidence only. Raw outcomes are in the ignored local artifact
`artifacts/components/completion-reference-check.json`.

A first local full proposal-corpus race invocation exceeded its two-minute
harness timeout; the normal complete execute suite passes, and the final race
invocation uses CI's explicit exclusion of complete upstream corpora. This is
not an unsupported-feature skip. Semantic CI retains its complete four-way
Go/LLGo/platform matrix. Performance campaigns are deferred by the owner decision;
benchmark dispatch and the `run-benchmarks` PR label retain explicit opt-in
measurement. Futures/streams, WASI 0.3 worlds, complete world conformance and
performance acceptance remain open.

## Future/stream guest protocols and affine transport — 2026-10-31

Sixteen hashed owned fixtures cover readable/writable directions, cancellation,
peer closure, future completion precedence, forwarding, unit/packed-byte streams,
zero-length readiness, memory64, cross-component strings, synchronous suspension,
exported endpoint moves/drops and generative nominal resource identity. Go race
and pinned LLGo release-profile tests execute the same guests on both own
backends. Opposite-direction handle-table move tests also check lock ordering,
authority and quota preservation. Successful imported ownership effects survive
consumer failure, and failed transfers restore handle availability.

Twelve fixtures are independently parsed/validated by wasm-tools 1.258.0. Two
forward binaries patch only a signature-equivalent placeholder canonical entry,
using the pinned component binary grammar, because the producer lacks forward.
Two full-width memory64 copy fixtures parse but are rejected by its incorrectly
i32 canonical copy signatures. `future-streams.json` records every input hash,
the generator and both actual producer rejections. These four cases are owned
execution evidence, not independent producer passes.

Actual Wasmtime 48.0.1 matches twelve common executions (scalar/byte/unit
copies, maximum unit count and zero-length readiness); it rejects both forward
and memory64 fixture groups. Raw outcomes and failures are retained in
`artifacts/components/future-stream-reference-check.json`. Custom host import
round trips, generative binding and string composition execute in the owned
engines but are not claimed as reference-provider conformance. Host I/O endpoint
production/consumption, WASI 0.3 worlds and complete M6 acceptance remain open.
No performance comparison or optimization gate is claimed during capability
completion.


## Host async I/O substrate — 2026-11-01

Eight owned fixtures generated by `build_host_source_fixtures.py` independently
parse and validate with wasm-tools 1.258.0. `host-sources.json` hashes their WAT,
binaries and generator against the pinned component specification. Both owned
backends execute typed u32/string/owned-resource futures, u32/u8/unit streams,
and guest producers whose task.return precedes a synchronous pending write.

Tests cover lazy starter ordering, external completion, fresh/expired Callers,
normal/early/cooperative cancellation, source and scheduled-work errors/panics,
wrong payload/counts, stream consumption before producer exit, owned-payload
handoff, rejected work ownership and aggregate quota restoration. Instance close
also closes never-started sources; teardown continues through close panics and
WaitClosed observes completion. Source-close failure after an owned copy cleans
the uncommitted destination. Successful imported effects remain committed.

Local Go race and pinned LLGo release-profile tests exercise these contracts.
The host embedding fixtures require custom providers and are not claimed as
Wasmtime provider conformance. No new performance result is claimed; WASI 0.3.2
interfaces/worlds, remaining reentry/cm64 contracts and full acceptance stay open.


## WASI 0.3.2 base, clocks and native stdio — 2026-11-01

Twenty base/clock/command and eleven stream/terminal fixtures independently parse
and validate with wasm-tools 1.258.0. The generators and source/binary hashes are
retained in preview3/testdata/base.json and io.json; base provenance additionally
pins the WIT archive and reused owned input sources. Exact P3 names and changed
contracts execute through both owned backends under Go race and pinned LLGo
release-profile tests.

Coverage includes immutable environment, random zero/short/scalar/seed operations,
signed system times, async relative/absolute waits, overflow/cancellation/provider
close, cross-Store quotas, stdio read/drop/EOF/error and output partial progress,
backpressure/final flush/error/future-drop. Publication limits release operation
reservations; pending instance/provider shutdown terminates workers. Nominal
terminal resources, malformed signatures and async command success/failure/exit/
trap/rejection execute as distinct outcomes. The new executable P3 package remains
subject to the same 90% per-package gate. No full CLI/HTTP world, independent guest
toolchain, reference-provider or performance acceptance is claimed by this subset.

Actual Wasmtime 48.0.1 CLI with its native WASI 0.3 providers matches seven
common executions: command success/failure, stdin EOF, stdout/stderr bytes,
monotonic wait and signed system-clock record. The retained local artifact
preview3-reference-check.json records the binary/fixture hashes, feature flags
and raw results. This is independent provider evidence for those cases, not
complete world conformance or a timing comparison.

## Synchronous LIFO reentry and destructor tasks — 2026-11-01

The owned runtime executes ordinary sync reentry while an outer async callback
holds exclusive ownership and explicit backpressure. Context slots start fresh
and outer task/entered/exclusive/storage state is restored after nested calls.
Guest destructors use fresh synchronous tasks for both local drops and host drops
while the defining guest is active. Borrowing, reentry/fuel limits, post-return
leave guards and the inability of sync calls to block remain checked.

Three owned fixtures independently parse/validate with wasm-tools 1.258.0; their
source, binary and generator hashes are pinned in sync-reentry.json. Go race and
pinned LLGo tests execute both owned backends, including fully guest-composed
indirect call cycles, recursion quotas, nested traps, denied waits, cancellation
and original-panic preservation. Wasmtime 48.0.1 rejects both composed variants
with cannot-enter-component traps and process exit 134; raw commands, fixture
hashes and diagnostics are retained in sync-reentry-reference-check.json. These
reference failures are recorded separately from own execution and do not close
full conformance or performance acceptance.

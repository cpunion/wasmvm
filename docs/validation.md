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
`internal/engine/interp`, **96.0%** for `internal/engine/state`, and **100%** for
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


## M1 representative Core 2.0 portfolio

The owned phase harness adds eight [versioned workloads](../benchmarks/embedded/portfolio/README.md)
to the original six cases: float/conversion, irregular multi-value control,
indirect calls/references/globals, SIMD, bulk/unaligned memory, deep calls, a
Clang-generated image filter and a Rust-generated packet CRC. Exact producer
versions, flags, source/binary hashes and independent checksums are committed.
Initialization/reset remains visible inside each call. Existing baseline work
sizes and repetition counts are unchanged.

The same fourteen inputs run through all fifteen configurations, producing
630 phase processes and 168 actual CLI processes. PR comments retain primary
base/head and competitor rows; complete phase data remains in raw JSON.
Small validation vectors pass locally through both own boundaries under Go
and LLGo and both wazero backends. Python checks validate provenance, sizes,
seed/count dependence and an independent zlib CRC cross-check. Full CI C API
validation and final-head measurements are required before merge. M1 remains
open for host-callback/preparation-shape baselines and final acceptance; no
M2 controlled performance or native-backend gate is claimed.

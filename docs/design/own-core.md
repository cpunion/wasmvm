# Decision: Project-Owned Core

Date: **2026-10-26**. Status: **selected implementation direction; implementation and performance gates remain open**. This decision supersedes the [M0 reuse-first decision](reuse-decision.md) following the owner's explicit choice to own the compiler and VM. M0 measurements and upstream compatibility findings remain valid evidence.

## Decision and scope

Write the decoder, validator, intermediate representations, interpreter, optimizer, native machine-code backends, and runtime state in Go in this repository. Keep one public embedding API with explicit interpreter/native selection. Wazero, Wasmtime, Wasmer, and WasmEdge remain independent reference engines and benchmark competitors, rather than production execution dependencies. No competitor is a hidden fallback for unimplemented instructions.

The immediate priority is a complete, bounded Core execution system with a fast portable interpreter and a required native compiler. Core 2.0 is the first complete binary profile; Core 3.0 and separate Core Threads follow on **both** backends. WASI P1, legacy wasi-threads, and LLGo guest integration follow the complete core; Component Model/P2/P3 remain required later stages. [The roadmap](../roadmap.md) defines acceptance and preserves the mapping from earlier milestone numbers.

Ownership provides control over internal representations, resource policy, native ABI, feature sequencing, and fixes. It also transfers validator, compiler, memory-safety, conformance, and platform maintenance responsibility to this project. Lower future maintenance cost is a hypothesis to assess at each gate, not a consequence established by code size.

## What the evaluation establishes

The inspected wazero v1.12.0 source contains about **61,502 nonblank, noncomment production Go lines**, including about 32,003 in the native compiler, 10,358 in the interpreter, 8,854 in core decoding/validation/runtime, and 5,528 in WASI/system support. Its `_test.go` files add about 70,723 code lines, excluding separate test helpers and external corpora. These are source categories, not an estimate for reproducing all required features; assembly, generated fixtures, and documentation are outside those counts. The source pin is recorded in the [evaluation](../research/wazero-evaluation.md).

The measured wazero native backend is competitive on the existing small numeric/memory workloads. This decision is therefore about implementation ownership and the complete target contracts, not a claim that the reference engine is slow. Existing performance measurements do not predict a new engine's throughput. Neither that baseline nor the unmerged GC patch establishes the effort required for complete Core 3.0, both execution backends, WASI, or LLGo parity.

The LLGo memory-hook fixes remain compiler compatibility work. The [native-entry failure](https://github.com/xgo-dev/llgo/issues/2921) is a separate unresolved issue; an interpreter pass does not establish native ABI compatibility. Keep the reproductions and references even though the production core changes direction.

## Engineering decisions

| Area | Decision | Required evidence |
| --- | --- | --- |
| Frontend | One bounded decoder and typed validator, independent of execution backends | Complete pinned positive/negative profile corpus, fuzzing and resource-limit tests |
| Intermediate representations | Typed control-flow IR with explicit effects and source locations; distinct interpreter and native lowerings | IR verifiers, trap/fuel/side-effect preservation, differential tests |
| Interpreter | Predecoded register/slot bytecode, explicit guest frames, specialization justified by measurements | Full profile conformance and the interpreter performance gate |
| Native compiler | Go-written optimization, instruction selection, register allocation and amd64/arm64 emission | Same profile corpus, ABI/lifetime tests, compile and execution budgets |
| Native entry | Small C ABI bridge initially; no dependency on LLVM or a foreign Wasm engine | Actual Go and LLGo execution, GC/stack/cancellation stress and call-boundary measurement |
| Portability | Interpreter and compiler logic build without cgo; standard-Go native execution initially requires cgo for the bridge | Explicit build/backend matrix; no claim that native execution is cgo-free |
| Runtime | Common ownership, linear memory, guest handles, traps, safepoints, budgets and continuations | Cross-backend lifecycle and feature-interaction tests |
| Performance | Measure each subsystem, exact own base/head builds, and pinned competitors | [Performance contract](performance-contract.md); no inference from startup-inclusive CLI timings |

The [compiler design](core-compiler.md) and [native ABI](native-abi.md) specify the mechanisms. Small platform shims are allowed; the execution engine and code generator remain Go code. An eventual cgo-free native bridge is a separately measured optimization, not a reason to delay a correct native backend or silently omit LLGo testing.

## Implementation and acceptance discipline

Implement reviewable contributions in order, through PRs with final-head checks passing before merge. Contributions within a milestone may land incrementally; **the milestone stays incomplete until every deliverable and gate passes**. Never rename an incomplete opcode subset as a completed Core profile. Explicit native mode fails with a typed unsupported-backend/profile error when unavailable; it does not switch to the interpreter.

Keep feature/backend/platform support machine-readable when execution is implemented. Do not expose internal bytecode, SSA nodes, machine registers, platform pointers, or foreign-runtime objects through `pkg`. Add platform and API surface only when backed by tests. Imported tests, tools, and any deliberately reused source retain their licenses and provenance; independence is not a reason to copy upstream code without attribution.

This change records the implementation decision and acceptance design. It adds no production VM capability. The current benchmark job still measures candidate engines; wiring actual wasmvm base/head execution is an M1 deliverable and a prerequisite for reporting own-core performance.

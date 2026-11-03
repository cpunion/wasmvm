# PR112 cold-operand CI: semantic success, continued performance hold

Head `7a2ffd0381a1d4c511ab6f660ac3651f1b3b3c4f` passes all nineteen required
jobs in [semantic CI](https://github.com/cpunion/wasmvm/actions/runs/41638103341).
Longest required job is 507 seconds (Linux LLGo core), with queue delay separate.
Nineteen hash-verified archives prove all 240 execution reports, sixty per
compiler/platform, all 31 packages, thirty coverage gates and all five required
proposal snapshots on three tiers. Stack Switching forces collection. Twelve
LLGo traces and independent P3 execution logs verify the complete package inventory.

The [benchmark](https://github.com/cpunion/wasmvm/actions/runs/41638103348)
and automatic PR report complete. Independently checked source inventories,
compiler/profile identities, all checksums, 228 observations and 76 regenerated
summaries across nineteen configurations pass. Both base and head contain 1475
engine files, with exact production Go 1.27.3/LLGo `983465e` and full-LTO flags.

**Integration remains held.** Restoring scalar operand register storage alone
does not resolve Linux scalar regressions. Warm ns/source iteration except
empty ns/call:

| Kernel/backend | Go base → head | LLGo base → head | Reference |
| --- | ---: | ---: | --- |
| Empty/interpreter | 534 → 532 | 1933 → 1935 | WasmEdge 269; wazero 142 |
| Integer/interpreter | 88.716 → 97.531 | 67.251 → 75.372 | WasmEdge 69.411; wazero 91.274 |
| Guest calls/interpreter | 129.746 → 137.811 | 124.655 → 132.652 | WasmEdge 91.594; wazero 123.107 |
| Memory/interpreter | 171.974 → 189.803 | 144.900 → 161.307 | WasmEdge 136.721; wazero 162.497 |
| Empty/native | 962 → 954 | 1873 → 1899 | Wasmtime 259; Wasmer 278 |
| Integer/native | 2.384564 → 2.387524 | 2.393084 → 2.417272 | Wasmtime 2.049712; Wasmer 2.022964 |
| Guest calls/native | 1.344560 → 1.413140 | 1.351700 → 1.348073 | Wasmtime 1.337513; Wasmer 1.003920 |
| Memory/native | 1.570627 → 1.471833 | 1.324920 → 1.357833 | Wasmtime 0.971553; Wasmer 0.938027 |

Retain LLGo interpreter integer/calls/memory **+12.1%/+6.4%/+11.3%**, Go
interpreter **+9.9%/+6.2%/+10.4%**, Go native calls +5.1% and LLGo native
memory +2.5%. Every LLGo head process is slower than every baseline process
in all three sustained scalar cases. All raw phases, first calls and controls
remain separate. No discarded observation, reference substitution or unchanged
timing rerun replaces these results.

The source-changing immutable-instruction follow-up borrows records from the
frame-local code slice. Actual linked amd64 assembly removes the eager eight-field
copy; this is build/disassembly evidence under emulation, not throughput. Its own
complete local matrix and final-head target CI remain required. No LLGo #2983
implementation is included; another contributor owns that issue.

Controlled portfolio, public-boundary and ordinary-C acceptance remain open.

[Original CI/benchmark outcomes, source identities, coverage, traces, execution logs
and verification](evidence/pr112-cold-operands-ci-2026-11-03.json.gz).

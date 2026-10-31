# Stable WASI 0.2 world assembly

`NewWorld(WorldConfig)` composes the existing versioned providers for stable
WASI 0.2.12 CLI command and HTTP proxy consumers on Linux/Darwin. It creates one
IO owner and reuses its nominal streams, errors and pollables throughout the
filesystem, network, TCP, UDP, HTTP and clock interfaces. Configuration remains
explicit: args/environment, preopens, stdio and network callbacks grant only
what the embedder supplies. Defaults use EOF/discarded IO, no preopens and denied
network access. The provider does not inherit process state.

World assembly shares existing host operations and contract checkers. `Imports`
returns fresh maps, and the final HTTP imports provide the single exported
monotonic clock. `IO` and `HTTP` expose the same owners used by those imports.
`HTTPHandler` factories may create a World per request and supply its HTTP owner
alongside a fresh component instance.

IO sources transfer when the inner NewIO constructor succeeds. Later constructor
failure closes/joins accepted providers and joins cleanup errors to the primary
failure. Close requests shutdown in reverse dependency order; WaitClosed joins
all owners and retains the first cleanup failure. Close bound instances
separately: immutable configuration imports and guest execution belong to the
Store lifetime, while resource owners have explicit close/wait lifetimes.

`RunCommand` checks the `wasi:cli/run@0.2.12#run` signature before invoking it.
A unit success/failure result maps to code 0/1. A nonreturning WASI exit preserves
its code, including zero; traps and cancellation remain errors. Instantiation
and Store/provider ownership stay explicit. `Function.Info` exposes immutable
export signatures, including the defining namespace's nominal resource types,
without entering the Store. HTTP incoming handling uses the same view to reject
incorrect own/request/outparam or result signatures before publication/effects.

## Validation boundaries

C and Rust commands compile independently with pinned compilers and wit-bindgen.
The Rust command calls the C component's arithmetic export after composition,
then exercises configured arguments, random bytes, polling, filesystem stat and
stdout through real WASI interfaces. Both execute on interpreter/native and
Go/LLGo release. Actual Wasmtime WASIp2 independently executes the same bytes.
The [guest sources](../../testdata/wasi-preview2/guests/README.md), producer,
source and binary checksums, dependency lock and notices are retained.

Compiler DCE narrows unused language imports. A separate fixture retains the
entire stable CLI command import surface, binds it with the same World and
returns success. Individual provider suites supply operation execution coverage;
the full-import fixture alone proves binding, not those operations. Seven valid
CLI entry controls exercise success, result failure, trap, exit and incorrect
signatures. A valid incoming-handler result mutation must reject before its
observable callback. HTTP client/incoming tests also use the unified imports.

These results advance CLI/world integration and independent guest toolchains.
Complete upstream world conformance, local cm64 borrows and M5 performance
gates remain open. Native
async, futures/streams and WASI 0.3 require separate M6 implementation.

## Performance boundaries

`BenchmarkWASICommand` runs the compiled C and Rust-to-C commands with lookup and
preparation outside warmed calls. Both own backends and the actual Rust reference
capture/verify bounded output, configure identical args and one read-only preopen,
and use real provider operations. Fuel is off; quotas/cancellation remain enabled.
The benchmark does not isolate arithmetic efficiency from WASI/canonical/resource
costs. It records load, engine, compilation, provider/Store setup, binding,
lookup, first invocation and owner teardown separately. Reference and own default
random algorithms, resource and scheduling policies remain different.

Build the reference with `cargo +1.98.0 build --locked --release --manifest-path
benchmarks/components/rust/Cargo.toml --bin command`. Invoke `command COMPONENT
PREOPEN COUNT c|rust|full`; `full` preserves the complete-consumer reference link
failure separately from successful language guests. Set
`WASMVM_COMMAND_BENCH_PREOPEN` to the same directory for own benchmarks and run
fixed 1000-call samples serially. World API absence on the source base is recorded
separately from unchanged wall-clock base/head samples. Performance acceptance
requires its own evidence and is not inferred from successful command execution.

## Stable version compatibility and upstream standard libraries

Host instance imports and public Function/Value paths use exact names first.
If an exact versioned interface is absent, the shared name resolver consults
stable SemVer tracks: nonzero major versions share that major; 0.x versions
with nonzero minor share that minor. Prereleases and 0.0.x have no alternate
track. The greatest numeric version wins; equal versions differing only in build
metadata use lexical name order for deterministic selection. Version syntax
shares the existing validator parser. Plain names and exact provider contracts
retain their usual behavior, including failure when an exact provider has the
wrong shape. No operation/resource signature or nominal identity check is bypassed.

Import fallback scans charge the existing aggregate instantiation budget before
parsing host keys. Export maps are already immutable and bounded by validation;
repeated lookups do not consume a Store-lifetime metadata quota. Exact lookup is
a small shared wrapper, with version parsing/scanning in the fallback helper.

Fourteen unmodified Wasmtime 48.0.1 test programs now compile directly with
Rust 1.98.0's wasm32-wasip2 target. Their standard libraries import WASI 0.2.9 and
export CLI run 0.2.0; the World remains a checked WASI 0.2.12 provider. Before
this change, every available guest failed to link at the first 0.2.9 interface.
Tests now execute arguments/environment, stdio, clocks/sleep, exit/panic, file
read/seek/append and sync on both own backends and host compilers. Expected exit
failure and guest traps remain distinct outcomes. Sources, license, toolchain
archive pins, stripped component hashes and the rebuilding script are retained
in [the upstream fixture inventory](../../testdata/wasi-preview2/upstream-wasmtime).

The append guest also exposed a provider bug: requesting the mutation capability
incorrectly forced O_DIRECTORY even when open-flags did not require a directory.
Only open-flags now enforces that object-kind constraint. Ordinary files clear
the directory-only capability from returned flags. Parent mutation grants,
explicit directory constraints and read-only checks remain enforced.

This bounded upstream CLI subset supplements operation suites and two-language
composition. It does not establish full upstream world conformance, cm64 local
borrow execution, performance acceptance, native async or WASI 0.3 support.

A sustained common-command probe additionally found an ownership error in the
project-authored C guest: wit-bindgen's list free helper frees strings/storage,
but does not drop contained owned descriptors. The guest now explicitly drops
every returned preopen descriptor. Repeated C/Rust commands pass with a two-
descriptor provider quota; the old C fixture fails that same regression. The
regenerated C and Rust-to-C components are supplied unchanged to source base,
candidate and reference during comparison. The original 10,000-call quota
failure remains diagnostic evidence; provider limits were not raised to hide it.

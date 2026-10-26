# Core command line

Build with `go build -o wasmvm ./cmd/wasmvm` or, with the pinned toolchain,
`llgo build -o wasmvm ./cmd/wasmvm`. CI builds and executes both binaries on
Linux amd64 and macOS arm64. The interpreter build also supports `CGO_ENABLED=0`.

```sh
./wasmvm validate module.wasm
./wasmvm inspect module.wasm
./wasmvm run --export add --fuel 1000 module.wasm i32:17 i32:25
```

Options precede the filename. `validate` performs binary decoding and static
validation only: it does not lower code, link imports, instantiate or execute
start. `inspect` performs the same frontend checks and prints copied metadata,
including import/export signatures, memory/table limits and global types.
Inspection does not require imports to be available. `run` compiles, instantiates
and calls the explicitly named export using the public embedding API.

Inspection limits the total number of parameter/result types repeated across
all import/export JSON signatures to 1,048,576. Exceeding this presentation
budget reports a resource-limit error before expansion or output; the binary
can still pass `validate`. This bounds amplification when many functions share
one large binary signature. All commands retain the frontend's input/node limits.

All commands accept `--backend interpreter`, `--profile core2`, and
`--timeout 30s`. `--timeout 0` disables the default deadline. Other backends or
profiles fail explicitly. Core run installs no host imports or WASI capabilities;
modules requiring imports fail linking. WASI command execution is an M4 item.
Ctrl-C requests cancellation. The VM checks deadlines at documented bounded
phase/execution boundaries; this is not a hard preemption guarantee for OS file
reads, host allocations or arbitrary embedding callbacks.

`--fuel N` on run covers the start function and selected export together, using
the instantiation receipt to transfer the remaining budget. `--fuel 0` enables
metering with zero fuel; omitting the flag disables instruction metering.
Segment initialization is not instruction fuel. Trap, quota or cancellation
errors may leave initialization/execution effects in the temporary Store, which
is always closed before command exit.

Arguments use explicit `type:value` syntax:

| Type | Examples | Meaning |
| --- | --- | --- |
| i32 / i64 | `i32:-1`, `i64:0xffffffffffffffff` | Signed decimal or exact unsigned bits; range checked |
| f32 / f64 | `f32:1.25`, `f64:-Inf`, `f32:NaN` | Floating-point value |
| f32bits / f64bits | `f32bits:0x7f800001` | Exact bits, including NaN payload and negative zero |
| v128 | `v128:0x1234,0xabcd` | Low 64 bits followed by high 64 bits |
| funcref / externref | `funcref:null` | Only null references can be supplied from command text |

Successful output is JSON. Numeric results contain the Wasm type, exact bits
and a readable value; vectors contain low/high bits. Reference results report
null/non-null without exposing a reusable Store identity. Every report names
the backend and profile. Metered run also reports `fuelRemaining`.

Errors go to stderr. Exit code 0 means success, 1 means a VM or I/O failure,
and 2 means invalid command syntax. `--help` describes the accepted options.

For a checked example already in the repository:

```sh
./wasmvm run --export run --fuel 20 \
  benchmarks/embedded/fixtures/call-control.wasm i32:305419896 i32:1
```

This returns i32 bits `0x12345678`. These CLI checks establish correct execution;
process startup and JSON/file I/O costs are separate from engine benchmarks.

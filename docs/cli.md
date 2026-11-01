# Command line

Build `cmd/wasmvm` with Go or the pinned LLGo release profile:

```sh
go build -o wasmvm ./cmd/wasmvm
llgo build -O3 -lto=full -o wasmvm ./cmd/wasmvm
```

Options precede the module path. `wasmvm --help` lists all flags. Core commands:

```sh
wasmvm validate module.wasm
wasmvm inspect --profile core3 module.wasm
wasmvm run --backend native --export add module.wasm i32:20 i32:22
wasmvm run --export work --fuel 100000 --timeout 2s module.wasm
```

Core results are JSON, with exact numeric/vector values and optional fuel
remaining. Core run installs no host imports. `--profile core3` enables the
complete final Core 3.0 binary profile. `--threads` independently enables shared
memory and atomics; it does not supply a thread-creation ABI. Standalone feature
flags retain their archived proposal behavior. Native execution requires a
supported platform adapter; no backend fallback is performed.

`--compact-imports`, `--wide-arithmetic` and `--custom-page-sizes` independently enable the pinned
[core extensions](design/core-extensions.md). Both default to disabled, even
with `--profile core3`, and apply to validate, inspect and run.

## WASI Preview 1

On Linux/macOS, `--wasi` installs the P1 Process adapter. The default export is
`_start`; positional values after the module are plain guest arguments.
Guest stdout/stderr are written directly, and its exit status is returned.
Arguments/environment/directories are explicit:

```sh
wasmvm run --wasi hello.wasm argument
wasmvm run --wasi --env MODE=test --dir /work=/tmp/work module.wasm
wasmvm run --wasi --read-only-dir /data=/srv/data --stdin input.txt module.wasm
wasmvm run --wasi --backend native --profile core3 \
  --native-code-bytes 268435456 --dir /work=/tmp/work module.wasm
```

No host environment or directories are inherited. Stdin defaults to the CLI's
stdin; `--stdin` selects a file. `--env`, `--dir` and `--read-only-dir` may be
repeated. Standard streams use cancellable nonblocking adapters, restoring
borrowed file-status flags on cleanup. Fuel includes initialization and the
invocation; the timeout also covers compilation and host waits.

The default native code quota remains 64 MiB per module. Large Go guests can
require an explicit larger quota as shown above. Missing native permissions or
exhausted resources are reported, without switching to interpretation.

[Declared P1 host behavior](design/wasi-preview1.md) distinguishes optional
unsupported clocks/signals from missing imports.

## Legacy WASI threads

`--wasi-threads` installs P1 plus `wasi.thread-spawn` and enables Core Threads.
Each worker receives a separate Store/instance; shared memories and process
resources are retained until the group has joined. Use the Core 3.0 profile
for LLGo's final EH:

```sh
wasmvm run --wasi-threads --profile core3 --backend native \
  --native-code-bytes 268435456 --max-threads 256 \
  --dir /work=/tmp/work guest-llgo.wasm
```

The command entry is `_start`. `--max-threads` includes the main thread and
pending creation, default 256. With this adapter, fuel is a separate budget
for each thread including initialization; the timeout covers the entire
process. Exit/trap/cancellation stops and joins all workers. See
[the lifecycle contract](design/wasi-threads.md). Components/WASI 0.2/0.3
remain separate capabilities.

## WASI component commands

`--wasi-preview 2` runs the standard synchronous WASI 0.2 command entry;
`--wasi-preview 3` runs the native-async WASI 0.3 entry. The providers use pinned
WASI 0.2.12/0.3.2 interfaces and checked stable version compatibility. Both owned
backends execute the same components. The default embedded Core profile is
Core 3.0; an explicit `--profile` still selects its declared feature scope.

```sh
wasmvm run --wasi-preview 2 --dir /data=/srv/data command-p2.wasm argument
wasmvm run --wasi-preview 3 --backend native --fuel 10000000 \
  --read-only-dir /data=/srv/data --stdin input.txt command-p3.wasm argument
```

The module basename is argv0. Arguments, environment, writable/read-only preopens
and stdin use the existing CLI options. Environment/directories/network are not
inherited; network operations are denied by default. Stdout/stderr stream directly
with bounded backpressure and cancellation. Borrowed stdio flags are restored
after the instance, providers and readiness workers close/join.

The command uses its standard `wasi:cli/run` export; `--export` and
`--wasi-threads` do not apply to component commands. Success/failure results map
to 0/1, explicit exit preserves the guest code, and traps/cancellation report an
error. Fuel includes component initialization and invocation in one total
budget; timeout also includes compilation and host I/O. `--wasi-preview 1`
selects the existing P1 mode; `--wasi` remains its shorthand.

For explicit HTTP/socket authorities, custom component exports or service/
middleware hosting, use the [embedding API](embedding.md) and versioned World
configuration. The command adds no second provider or execution framework.

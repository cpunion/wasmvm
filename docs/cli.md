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
unsupported clocks/signals from missing imports. Core Threads and P1 do not
provide legacy `wasi.thread-spawn`; that adapter and components/WASI 0.2/0.3
remain subsequent capabilities.

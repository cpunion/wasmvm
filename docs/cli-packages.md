# CLI candidate packages

The existing Go/LLGo CI produces extracted-and-executed CLI candidate archives
on Linux amd64 (`ubuntu-24.04`) and macOS arm64 (`macos-15`). These are
unversioned CI artifacts. Controlled performance acceptance and the first
versioned release remain open.

## Download and run

Choose a successful [Go and LLGo run](https://github.com/cpunion/wasmvm/actions/workflows/ci.yml).
Each platform has `cli-go-<platform>` and `cli-llgo-<platform>` artifacts. Go
artifacts contain both a native-enabled CLI and a cgo-free portable-interpreter
CLI. LLGo artifacts use the pinned `-O3 -lto=full` native profile. The compiler
used to build the CLI is independent of the selected guest execution backend.

```sh
gh run download RUN_ID --repo cpunion/wasmvm --name cli-go-linux-amd64 --dir cli
cd cli
sha256sum -c SHA256SUMS
# On macOS: shasum -a 256 -c SHA256SUMS
# Select one archive; do not extract multiple variants into the same directory.
mkdir unpacked
tar -xzf wasmvm-COMMIT-linux-amd64-go-native.tar.gz -C unpacked
unpacked/bin/wasmvm --help
```

Archive names include the checked-out source commit, target, host compiler and
native/portable mode. In PR runs that checkout may be GitHub's temporary merge
commit; `sourceTree` identifies the tested contents. For durable main-commit
artifacts, dispatch the same CI workflow on `main` after merging. No additional
benchmark or duplicate semantic matrix is required for packaging.

The archive includes `manifest.json`, `loader.txt`, this project's MIT license
and dependency/toolchain notices. Read its runtime-library requirements before
running. Native Go packages require the recorded system C libraries. LLGo
packages also depend on the actual external libraries listed in the manifest
(such as BDWGC/libffi); those libraries are not bundled. The build compiler
and LLVM are not installation requirements unless listed as dynamic libraries.
The macOS loader record includes the deployment target; Linux records the
maximum required GLIBC symbol version. Build-host OS information alone is not
a minimum-OS compatibility guarantee.

## Contents and verification

Each archive has external `.manifest.json` and `.verification.json` records;
`SHA256SUMS` covers all three files. Metadata retains the exact source commit
and tree, binary/compiler hashes, Go/LLGo versions, release flags, native mode,
module versions/checksums, copied license hashes and dynamic-loader information.
Canonical tar/gzip metadata does not promise bit-for-bit compiler rebuilds.

Packaging extracts the archive into a temporary directory, verifies its binary,
license files and executable permission, then runs **that extracted binary**.
Core validate/inspect/run, P1 stdin/stdout, P2/P3 commands, P3 exit/failure and
trap diagnostics exercise every available backend. Native archives execute
18 CLI processes; portable archives execute 12, including explicit native
rejection. These smoke checks supplement the full semantic matrix; they do not
replace it or certify every guest program. No conformance inputs, competitor
engines or SDKs are redistributed inside the CLI archive.

Artifacts expire after 30 days. A versioned release will require the remaining
[release gates](release-readiness.md), an exact-source main/tag build and an
accurate release body; generating a candidate does not close those gates.

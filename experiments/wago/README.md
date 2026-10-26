# Wago comparison probe

This standalone module pins Wago `v0.1.0-beta.13` and preserves the project-authored arithmetic, unreachable-trap, and host-callback probes used during M0. It is a candidate comparison, not a production dependency or a full Wago conformance suite.

```sh
cd experiments/wago
go test -v -count=1 -timeout=2m ./...
llgo test -v -count=1 -timeout=2m ./...
```

Observed on 2026-10-26, darwin/arm64, Go 1.27.3 and Homebrew LLGo v1.0.8:

- Go: arithmetic, the unreachable trap, and the host callback passed.
- LLGo: assembly translation failed before test execution in `engine_int_arm64.s`, `enterNativeIntRaw`, with an unsupported `ADR` instruction. This is a compiler/backend compatibility finding, not proof that Wago cannot be adapted.
- Wago GC, threading, WASI plugins, and complete conformance were not executed in this comparison. Those capabilities must not be inferred from these two tests.

The WAT inputs were authored for this evaluation and compiled with WABT `wat2wasm` 1.0.41. Wago is obtained through its pinned module dependency and retains its upstream Apache-2.0 license. No Wago implementation source is vendored here.

The [reuse decision](../../docs/design/reuse-decision.md) records how this comparison affects the next work. Current CI validates the selected wazero evaluation path; this alternative's LLGo failure remains an explicitly reported local observation.

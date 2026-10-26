# Native integer architecture gate

This probe runs the existing `integer-mix.wasm` directly through the project
compiler and C entry adapter. It is not the public VM API: native memory,
guest-call dispatch and runtime helpers remain incomplete. No unsupported
operation silently falls back. The one kernel measures the scalar register and
interrupt-polling changes; it cannot accept the complete native backend.

Build with `go build -o /tmp/native-probe ./benchmarks/native` or the pinned LLGo
compiler. Send `run\n` on stdin and pass
`--wasm benchmarks/embedded/fixtures/integer-mix.wasm`. `--tier unmetered` keeps
bounded interrupt checks without periodic host round trips. `--tier sliced
--budget 64` selects the original precise sliced compiler with semantic fuel
disabled. The probe records frontend, native emission/publication, context setup,
first call, seven warm calls after two extra warmups, code size and host poll
exits. Checksums, initial frame setup and printing are outside the call timer.

The existing PR phase job reuses its isolated exact-revision source trees and
pinned wazero 1.12.0, Wasmtime 48.0.1 and Wasmer 7.3.0 compiler runners. It adds
20 interleaved process blocks for this kernel and publishes raw records and the
feature comparison in the same PR comment. Both host compilers must pass every
competitor comparison. A paired process bootstrap uses simultaneous 95% bounds;
slower or inconclusive results fail the feature gate. Shared-host results remain
provisional and do not replace controlled full-portfolio acceptance. macOS arm64
requires separate evidence before merging; the existing benchmark job is Linux
amd64 only. Future feature PRs must declare and compare their affected workload
families, not reuse this kernel to excuse an unrelated deficit.

For the bootstrap baseline predating `CompileUnmeteredFunction`, only this
probe's compiler-function reference is adapted to `CompileFunction`; its recorded
tier is `sliced`. Archived engine sources are checked unchanged and both probe
and executable hashes are retained. Later baselines use their own unmetered
compiler. Own baseline ratios remain visible; only competitor ratios decide
this feature gate, so an unchanged baseline does not fail merely due to A/A
noise. Preparation and first-call costs remain separately visible.

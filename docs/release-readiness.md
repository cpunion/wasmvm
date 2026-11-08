# Release contract and remaining acceptance

Review date: **2026-11-08**. This record separates delivered capabilities from
quantitative performance and production-release acceptance. It does not certify
all possible programs or constitute an independent security audit.

## Delivered functional scope

The declared scope includes the owned Core 2.0/3.0 interpreter and amd64/arm64
compiler, GC, Core Threads, P1 and legacy threads, Component Model/Canonical ABI,
stable WASI 0.2.12/0.3.2 worlds and the five separately gated non-JS extensions.
Draft TLS, config/keyvalue, component threading and other unselected proposals
remain explicitly outside that scope. Phase 3 snapshots remain experimental.

[PR #128](https://github.com/cpunion/wasmvm/pull/128) merged after 23 required
jobs passed. Its exact-head artifact audit verifies 240 execution reports,
31 LLGo packages exactly once per platform and all 30 coverage gates. The
merged source tree equals the tested tree. Linux amd64 and macOS arm64 are the
validated native targets; the portable interpreter also builds without cgo.

The stable upstream world inventory contains **92 P2 and 60 P3 consumers**.
An additional raw-log audit finds all **608 LLGo consumer pass events** across
both backends and platforms, with no missing or duplicate case. Go full-package
execution also passed; no Go per-test JSON is invented. Provider operation,
ownership, cancellation and cross-component tests supplement those consumers.
The original duplex executor diagnostics and separately named repaired variants
remain explicit in the [P2 inventory](../testdata/wasi-preview2/upstream/README.md).
Passing a declared inventory is not exhaustive WASI certification.

## Boundary review

The scoped source/test review checks these existing contracts:

| Boundary | Mechanism and exercised failures |
| --- | --- |
| Core/component input | Owned byte snapshots; checked readers, aggregate node/name/nesting budgets; malformed and over-limit inputs |
| Validation/lowering | Typed ownership of validated modules, bounded control/type work, verified IR/slot transfers |
| Native code and memory | Code/frame quotas, checked branch conditions, W^X publication, retained leases, full GC identities and guarded cold continuations |
| Guest ownership | Store serialization, typed roots, explicit release, failed-instantiation cleanup, cancellation and close while suspended |
| Filesystem/network | Explicit grants, `os.Root` and retained directory handles, relative-path checks, rights reduction, bounded operations and denied ambient authority |
| P2/P3 providers | Shared nominal resource identities, bounded workers/streams, cancellation, partial I/O, completion publication and joined shutdown |

Seven existing fuzz targets additionally ran for a predeclared 30 seconds each
with two workers: Core decode/validate/execution, component decode/validate,
Canonical literals and public component execution. All passed. Execution counts
are mutation attempts, not distinct assertions, coverage percentages or proof
that no defect exists. The [machine record](research/data/closeout-2026-11-08.json)
retains commands, logs, counts and the source identity.

## Complete benchmark correctness

[Issue #67](https://github.com/cpunion/wasmvm/issues/67) identified a validator
that rejected the CLI's existing `features: 0` output. It now requires exact
integer zero, permits only historical baseline omission, and still rejects
other policies, wrong JSON types, extra fields and corrupt checksums. All 91
measurement-tool unit tests pass. The original failed run's 48 CLI observations
and 16 summary cells replay without changing its raw data. Two fresh exact-source
Go/LLGo CLI binaries also pass 48 real processes across four workload families.
Logical base/head roles share those binaries because production Go sources are
unchanged; their retained times are correctness evidence, not an optimization.
Generic PR timing remains opt-in; explicit scoped/full campaigns remain separate
from the unchanged full semantic matrix.

## Public versioning policy

Until the first tag, consumers must pin a commit. All seven public `pkg/...`
packages belong to one module and will share its version. Internal packages and
measurement adapters are not supported external APIs.

For a future `v0.MINOR.PATCH` release, patch updates preserve exported Go source
compatibility, documented ownership/lifecycle rules and CLI result
representations. Correctness/security fixes may tighten rejection of invalid
inputs; such changes require release notes. Additive CLI fields are documented.
A minor update may change experimental APIs, with migration notes. Proposal
snapshot changes remain explicit feature/version events. No v1 stability
commitment or production release is implied by this policy.

## Candidate packaging

The existing semantic CI also builds verified CLI candidate packages for both
platforms/host compilers, plus portable Go interpreter variants. Archives retain
source/tree, compiler/profile, binary/module/license hashes, actual loader
requirements and extracted-binary Core/P1/P2/P3 smoke records. See the
[package contract](cli-packages.md). They remain unversioned CI artifacts;
controlled performance acceptance and production publication are still open.

## Gates still requiring evidence

- Full controlled performance acceptance: complete pinned portfolio, production
  fuel/cancellation policy, preparation/first/warm/cleanup, allocations and peak
  owned-memory evidence. Preserve all named competitors and native C where
  applicable. Hosted advisory results do not close this gate.
- Reproducible Linux amd64 and macOS arm64 acceptance measurements. GitHub-hosted
  `ubuntu-24.04` and `macos-15` already execute these targets; self-hosted runners
  are not a prerequisite. The runner API inventories self-hosted machines only.
  Record CPU/microcode, image, affinity, power/frequency and background-load
  policy, including controls the host does not expose. Keep independent samples
  and A/A calibration; unresolved noise is not performance acceptance.
- Final packaging and version publication after the applicable gates pass:
  exact source/toolchain/profile, dependency notices, actual platform/loader
  requirements, checksums and an accurate release body. MIT is selected and
  third-party notices remain in their inventories.

Do not mark M2 or production release complete while these gates remain unmet.
Optional M7 work is not a hidden requirement to implement every future proposal.

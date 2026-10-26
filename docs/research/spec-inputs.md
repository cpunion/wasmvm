# Initial Specification and Conformance Input Pins

Recorded on **2026-10-26** for the M0 acceptance gate. These are immutable
baseline inputs for later milestones, separate from the engine and toolchain
pins in the [wazero evaluation](wazero-evaluation.md). Repository revisions and
license notices were inspected; both WIT release archives were downloaded and
their SHA-256 digests verified. **No complete upstream conformance suite has
been executed by this project.** The owned M0 probes are not substitutes for
these suites.

## Source revisions

Use the full commit below, rather than a moving branch or a release tag alone.
The milestone column identifies intended adoption, not completed support.

| Input | Repository and immutable revision | Scope and intended milestone | Upstream license notice |
| --- | --- | --- | --- |
| Core 2.0 specification and official tests | [WebAssembly/spec wg-2.0](https://github.com/WebAssembly/spec/tree/fffc6e12fa454e475455a7b58d3b5dc343980c10), `fffc6e12fa454e475455a7b58d3b5dc343980c10` | M1/M2 complete Core 2.0 binary scope; include `test/core` and `test/core/simd`, not an ad hoc subset of Core 3.0 tests | [Directory-specific LICENSE](https://github.com/WebAssembly/spec/blob/fffc6e12fa454e475455a7b58d3b5dc343980c10/LICENSE); preserve Apache-2.0 notices for interpreter/test inputs |
| Core specification and reference interpreter | [WebAssembly/spec](https://github.com/WebAssembly/spec/tree/aa4dcf44b8de3f3f3270ca067422172518ecc341), `aa4dcf44b8de3f3f3270ca067422172518ecc341` | Core 3.0 baseline; M3 full-profile target | [Directory-specific LICENSE](https://github.com/WebAssembly/spec/blob/aa4dcf44b8de3f3f3270ca067422172518ecc341/LICENSE): W3C Software and Document Notice and License for `document/` and `specification/`; Apache-2.0 for `spectec/`, `interpreter/`, and `test/`; CC0 for `proposals/`; CC-BY-4.0 for `papers/` |
| Generated Core test suite | [WebAssembly/testsuite](https://github.com/WebAssembly/testsuite/tree/44f1503f98409fa3737437bf9449a37a5def4210), `44f1503f98409fa3737437bf9449a37a5def4210` | Additional M3 corpus source; record generation provenance and coverage before adoption | [Apache-2.0](https://github.com/WebAssembly/testsuite/blob/44f1503f98409fa3737437bf9449a37a5def4210/LICENSE) |
| Core Threads proposal and tests | [WebAssembly/threads](https://github.com/WebAssembly/threads/tree/cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0), `cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0` | Shared linear memory, atomics, wait/notify; M3; `test/core/` is an input candidate, not an executed suite | [Directory-specific LICENSE](https://github.com/WebAssembly/threads/blob/cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0/LICENSE): W3C notice for `document/`; Apache-2.0 for `interpreter/` and `test/`; CC-BY-4.0 for `papers/` |
| Component Model and Canonical ABI | [WebAssembly/component-model](https://github.com/WebAssembly/component-model/tree/d095e2f60ddae1f9a1e9b1e2fb547ec1114da677), `d095e2f60ddae1f9a1e9b1e2fb547ec1114da677` | `design/mvp/` binary, WIT, Canonical ABI, and concurrency definitions; M5/M6 | [Apache-2.0 unless a subdirectory states otherwise](https://github.com/WebAssembly/component-model/blob/d095e2f60ddae1f9a1e9b1e2fb547ec1114da677/LICENSE) |
| WASI Preview 1 definitions | [WebAssembly/WASI, wasi-0.1 snapshot](https://github.com/WebAssembly/WASI/tree/fae981bae14809d91f9bc2d63852d461f331d161), `fae981bae14809d91f9bc2d63852d461f331d161` | `preview1/` WITX/API baseline for M4; do not substitute a current P2/P3 source tree | [W3C Community Contributor License Agreement notice](https://github.com/WebAssembly/WASI/blob/fae981bae14809d91f9bc2d63852d461f331d161/LICENSE.md) |
| WASI 0.2.12 release metadata | [WebAssembly/WASI v0.2.12](https://github.com/WebAssembly/WASI/tree/281ba75fafcd50961ef55f9e52747afcc9b71ede), `281ba75fafcd50961ef55f9e52747afcc9b71ede` | M5 version target; use the separately checksummed WIT archive below | [W3C Community Contributor License Agreement notice](https://github.com/WebAssembly/WASI/blob/281ba75fafcd50961ef55f9e52747afcc9b71ede/LICENSE.md), with package notices under `proposals/` |
| WASI 0.3.2 release metadata | [WebAssembly/WASI v0.3.2](https://github.com/WebAssembly/WASI/tree/bdadbbb8cadd287cc5014c4b5a0bf76efc60291c), `bdadbbb8cadd287cc5014c4b5a0bf76efc60291c` | M6 version target; use the separately checksummed WIT archive below | [W3C Community Contributor License Agreement notice](https://github.com/WebAssembly/WASI/blob/bdadbbb8cadd287cc5014c4b5a0bf76efc60291c/LICENSE.md), with package notices under `proposals/` |
| WASI test sources and runner | [WebAssembly/wasi-testsuite](https://github.com/WebAssembly/wasi-testsuite/tree/58b6206a46013fb53d23990e5a650584a5361043), `58b6206a46013fb53d23990e5a650584a5361043` | Source baseline for P1 and later version-specific coverage; runner/producer integration and supported subsets remain to be validated | [Apache-2.0](https://github.com/WebAssembly/wasi-testsuite/blob/58b6206a46013fb53d23990e5a650584a5361043/LICENSE) |
| Legacy wasi-threads | [WebAssembly/wasi-threads](https://github.com/WebAssembly/wasi-threads/tree/f8ea03cc0ffbde143415e945497de112f6f9fd53), `f8ea03cc0ffbde143415e945497de112f6f9fd53` | M4 ABI and upstream test-reference snapshot | No root license file or repository-wide license metadata was found at this revision. Reference-only in M0; establish the applicable file notices before copying tests or source |
| Proposal status registry | [WebAssembly/proposals](https://github.com/WebAssembly/proposals/tree/df6a2aca1cd9d779eb4f2f3e19dae8277f336015), `df6a2aca1cd9d779eb4f2f3e19dae8277f336015` | Phase/status provenance for the [standards survey](standards.md); not an executable conformance input | No root license file or repository-wide license metadata was found; linked as reference metadata only |

The Core 2.0 pin was added with the own-core design; the `wg-2.0` tag was resolved to the immutable commit above. Corpus conversion and execution remain M1 work.

The Core specification and generated suite have separate pins; this record
does not assert that one was generated from the other. Before executing an
adopted subset, record its generation provenance, feature configuration,
reference-engine version, and any intentional exclusions.

## Released WIT artifacts

The WASI release tags are not a sufficient pin for the packaged WIT files.
The following release assets are the initial M5/M6 interface inputs. Their
downloaded digests match GitHub's release-asset `digest` metadata.

| Release asset | Bytes | Verified SHA-256 | Inspected contents |
| --- | ---: | --- | --- |
| [wasi-wit-0.2.12.tar.gz](https://github.com/WebAssembly/WASI/releases/download/v0.2.12/wasi-wit-0.2.12.tar.gz) | 32,250 | `2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992` | 21 WIT files including dependency copies; `wasi:cli`, `clocks`, `filesystem`, `http`, `io`, `random`, and `sockets` package versions are `0.2.12` |
| [wasi-wit-0.3.2.tar.gz](https://github.com/WebAssembly/WASI/releases/download/v0.3.2/wasi-wit-0.3.2.tar.gz) | 25,493 | `2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311` | 15 WIT files including dependency copies; `wasi:cli`, `clocks`, `filesystem`, `http`, `random`, and `sockets` package versions are `0.3.2` |

These archives contain WIT files only, without embedded license files. Retain
the pinned WASI and applicable package notices when importing them; the
upstream [CLI notice](https://github.com/WebAssembly/WASI/blob/281ba75fafcd50961ef55f9e52747afcc9b71ede/proposals/cli/LICENSE.md)
and [filesystem notice](https://github.com/WebAssembly/WASI/blob/bdadbbb8cadd287cc5014c4b5a0bf76efc60291c/proposals/filesystem/LICENSE.md)
also identify the W3C Community Contributor License Agreement. This inventory
records upstream notices and does not relicense their contents. No upstream
archive, test corpus, or implementation source is vendored by this document.

Reproduce the artifact check in a temporary directory:

```sh
gh release download v0.2.12 --repo WebAssembly/WASI --pattern wasi-wit-0.2.12.tar.gz
gh release download v0.3.2 --repo WebAssembly/WASI --pattern wasi-wit-0.3.2.tar.gz
shasum -a 256 -c <<'CHECKSUMS'
2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992  wasi-wit-0.2.12.tar.gz
2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311  wasi-wit-0.3.2.tar.gz
CHECKSUMS
```

## Adoption and update rule

M0 accepts the input inventory, provenance, and verified artifact digests.
Later milestones must add executable adapters and results for their complete
declared scope. A pin, a parsed WIT package, or an expected unsupported result
is not a conformance pass. Update these revisions and digests through a PR,
review semantic and license changes, and rerun every affected accepted subset
before claiming compatibility with the new baseline.

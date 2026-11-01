# Explicit core extensions

These switches extend Core 3.0; selecting `core3` does not enable them.
Both extensions below run through the owned decoder, validator, typed IR,
slot code and interpreter/native backends, without a reference-engine fallback.

| Extension | API feature | CLI option | Specification snapshot |
| --- | --- | --- | --- |
| Compact Import Section | `wasm.CompactImports` | `--compact-imports` | [`0624de8`](https://github.com/WebAssembly/compact-import-section/tree/0624de829e2663da799e708444dec80337a3fe54) |
| Wide Arithmetic | `wasm.WideArithmetic` | `--wide-arithmetic` | [`024d4ce`](https://github.com/WebAssembly/wide-arithmetic/tree/024d4ce1b28cee7fd006d9c5175e3b7ece52b429) |

Compact imports decode both `0x7f` groups with individual external types and
`0x7e` groups with one shared external type. The abstract module keeps its
existing flat import list in source order. Normal linking, type identity,
index spaces and ownership therefore require no parallel implementation.
Empty groups add no imports. Expanded entries consume the decoder node budget,
and embedded component modules use the same parser. Empty marker names,
group lengths and every external-type feature gate are checked before use.

Wide Arithmetic adds `i64.add128`, `i64.sub128`, `i64.mul_wide_s` and
`i64.mul_wide_u` at `0xfc` subopcodes 19–22. Pairs are low word followed by
high word. Addition/subtraction wrap modulo 2^128; multiplication returns the
full signed/unsigned 128-bit product. Neither result is a `v128` value.
The validator checks exact i64 signatures and normal operand limits.
Slot code uses the existing bounded input/output tuples, preserving both
results through control flow, calls and native resumption. AMD64 emits
ADD/ADC, SUB/SBB and MUL/IMUL; ARM64 emits ADDS/ADC, SUBS/SBC and MUL with
SMULH/UMULH. These instructions do not exit to an interpreter helper.
Tuple-aware validation rejects malformed internal plans before emission;
all inputs are consumed before aliased result slots are written.

The complete affected upstream scripts are committed in
`testdata/compact-import-section` and `testdata/wide-arithmetic`. Their
manifests pin source archive sizes/hashes, original WAST, test licenses,
the producer version/binary hash and every generated input. Regenerate with
`python3 scripts/core_extensions_corpus.py --generate --producer /path/to/wasm-tools`
after removing the old generated directories. Verification needs no producer
or network: `python3 scripts/core_extensions_corpus.py`.

The semantic matrix executes all scripts under Go and LLGo, on Linux amd64
and macOS arm64, in interpreter and both native tiers. The import corpus has
2 scripts, 70 commands, 32 binary assertions and 3 separately classified
text-parser assertions. The arithmetic corpus has 1 script, 109 commands
and 107 binary assertions. Additional tests compare 512 randomized operations
per tier against arbitrary-precision arithmetic, exercise carry/borrow across
large native frame offsets, overlapping input/output slots, fuel exhaustion,
disabled features, wrong operand types and resource quotas.

The owner also requires Custom Page Sizes, the non-JavaScript parts of Custom
Descriptors and Stack Switching. Those remain pending until each pinned
snapshot passes both backends and the Go/LLGo matrix. Phase 3 snapshots remain
explicit experimental gates rather than claims of final standardization.
Performance acceptance remains open; this extension work does not erase the
November 1 performance audit or claim new parity results.

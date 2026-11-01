# Explicit core extensions

These switches extend Core 3.0; selecting `core3` does not enable them.
The extensions below run through the owned decoder, validator, typed IR,
slot code and interpreter/native backends, without a reference-engine fallback.

| Extension | API feature | CLI option | Specification snapshot |
| --- | --- | --- | --- |
| Compact Import Section | `wasm.CompactImports` | `--compact-imports` | [`0624de8`](https://github.com/WebAssembly/compact-import-section/tree/0624de829e2663da799e708444dec80337a3fe54) |
| Wide Arithmetic | `wasm.WideArithmetic` | `--wide-arithmetic` | [`024d4ce`](https://github.com/WebAssembly/wide-arithmetic/tree/024d4ce1b28cee7fd006d9c5175e3b7ece52b429) |
| Custom Page Sizes | `wasm.CustomPageSizes` | `--custom-page-sizes` | [`61c873a`](https://github.com/WebAssembly/custom-page-sizes/tree/61c873a1860acad2772539164f284cc5110f1b1b) |
| Custom Descriptors (non-JS, experimental Phase 3) | `wasm.CustomDescriptors` with GC | `--custom-descriptors` | [`a791af2`](https://github.com/WebAssembly/custom-descriptors/tree/a791af285f3876fb0bc6df89c218ff912347cf46) |
| Stack Switching (experimental Phase 3) | `wasm.StackSwitching` with GC, typed references and EH | `--stack-switching` | [`d9fbe34`](https://github.com/WebAssembly/stack-switching/tree/d9fbe345495d15af623b981440d6aaacd5ad015c) |

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

Custom Page Sizes permits exactly 1-byte and 64-KiB pages, independently for
each memory. `MemoryType.PageSize == 0` preserves the 64-KiB default; an explicit
nonzero value requires the feature gate. Binary exponent values below 64 are
decoded without overflow; validation then rejects unsupported sizes.
Memory imports require equal effective page sizes; explicit 65536 matches the
default. Guest `memory.size` and `memory.grow` count that memory's pages, while
addresses and bulk-memory lengths remain byte counts. The 1-byte i32 memory
limit is 2^32-1 pages; the i64 limit is 2^64-1. The default limits remain 2^16
and 2^48 pages respectively. Counts are checked before byte multiplication.

Host `MaxMemoryPages` and memory usage counters continue to count 64-KiB quota
units, rounding up each memory's byte length. Shared maxima reserve the same
units independently in every importing Store. Private growth consumes only the
additional quota units; growth within an already charged unit remains possible
when the Store is at its quota. Failed instantiation, Store close and shared
lease retirement reclaim the corresponding rounded units. Guest pages reported
by `Extern.MemoryPages` retain their memory-specific sizes. Native byte views,
bounds guards and static `memory.size` lowering use each indexed memory's size;
the compiler never derives byte bounds using an unconditional 64-KiB multiplier.

All four affected Custom Page Sizes scripts are pinned in
`testdata/custom-page-sizes`: 80 commands, 53 binary assertions and 4 separately
classified malformed-text assertions. The same generator verifies/regenerates
them (`--proposal custom-page-sizes` selects a single corpus). An owned
`pkg/vm/testdata/custom-pages.wat` fixture, produced by the same pinned wasm-tools
version, additionally exercises mixed private/shared, i32/i64 memories, byte
boundary traps, zeroing, atomics, cross-Store leases, quotas and cleanup.

Custom Descriptors implements the pinned Core definitions and exact reference
types, including exact function imports (`0x20`), descriptor allocation,
descriptor retrieval, descriptor-equality casts and both branch variants.
The CLI enables its GC prerequisite; embedding callers select GC or `core3`
alongside `wasm.CustomDescriptors`. Its Phase 3 status remains experimental.
JS prototypes and JavaScript embedding APIs are outside this implementation.

Descriptor/describes clauses are reciprocal members of one recursive group,
with matching finality and a checked subtype square. Described types precede
their descriptor types, preventing cycles in descriptor chains. Canonical
recursive keys include both clauses and exactness. Exact types accept the
same concrete type, excluding proper subtypes; ordinary types retain declared
subtyping. Defined functions have exact `ref.func` types, while ordinary
imports remain inexact unless explicitly imported exactly. Linking checks
concrete canonical identity for exact imports, and component module interfaces
preserve the distinction.

Guest objects keep their descriptor as a typed-handle edge. The collector
traces that hidden edge, including descriptor/object cycles and meta-descriptor
chains, without retaining Go pointers in guest frames. The edge consumes eight
additional charged heap bytes. Constructor operands stay rooted during forced
collection; null descriptors trap before allocation. Normal casts compare
concrete types, while descriptor-equality casts compare descriptor identity.
Both backends use the existing shared checked GC and control-flow operations.

All 11 affected Core scripts are pinned in `testdata/custom-descriptors`:
684 commands, 597 binary assertions and 16 separately classified malformed
text assertions. The owned `pkg/vm/testdata/custom-descriptors.wat` fixture,
produced by the same pinned wasm-tools, also tests forced collection during
construction and exception/host boundaries, host payload reachability,
descriptor cycles, exact hidden-edge accounting and object/byte quota failures.

Stack Switching implements typed one-shot continuations, `cont.new`,
`cont.bind`, `suspend`, `resume`, `resume_throw`, `resume_throw_ref` and `switch`.
Continuation types have their own reference hierarchy (`cont` / `nocont`),
canonical identities and checked function signatures. Control tags may return
values; exception handling still rejects result-bearing exception tags.
Resume handlers use existing typed control-flow edges and bounded slot tuples.

Prompts delimit frames in the existing guest execution arena. Capturing a
continuation copies its guest frames, values and nested prompts into a charged
guest heap object. It does not capture a Go/C stack, retain machine addresses or
native-code leases, or start a goroutine. Both backends restore the same logical
frames and slot PCs. Binding and resumption consume one-shot handles; reuse
traps. Invalid null-exception injection leaves the continuation unconsumed.
Suspended references, bound arguments and owning instances participate in
tracing; frame/slot and heap object/byte quotas apply across capture/restore.
Host callbacks remain ordinary explicit requests and cannot capture host stacks.

An unhandled control tag reaches the Go boundary as `vm.Suspension`, matching
`wasm.ErrSuspension`, separately from traps and exceptions. Its private payload
root retains tag identity and arguments. `Payload` returns owned argument
values; callers release those values and the suspension. This signal is
terminal at the host boundary, rather than an implicit host continuation API.

All four affected Stack Switching scripts are pinned in `testdata/stack-switching`:
162 commands and 112 binary assertions, including `assert_suspension`, with no
text-parser assertions. The runner forces collection at every guest allocation.
The owned `pkg/vm/testdata/stack-switching.wat` fixture also exercises saved
multi-frame roots, bound guest/host functions, exact-reference interactions,
null injection, consumed handles, signal ownership, cancellation and quota
failure/cleanup. All run locally under Go and LLGo on both backends; final
four-target integration validation remains required for PR #87.

Phase 3 snapshots remain explicit experimental gates rather than claims of
final standardization.
Performance acceptance remains open; this extension work does not erase the
November 1 performance audit or claim new parity results.

# Complete Core 3.0 binary corpus

Source: [WebAssembly/spec](https://github.com/WebAssembly/spec/tree/aa4dcf44b8de3f3f3270ca067422172518ecc341),
revision `aa4dcf44b8de3f3f3270ca067422172518ecc341`. All **259** `.wast` scripts
under `test/core`, including every subdirectory, are retained completely.
Each execution tier processes **65,344 commands**, executes **61,395 binary
assertions**, and separately classifies **1,239 text-parser assertions**.
No source command or assertion is removed.

`inputs.tar.gz` contains the complete original sources, test license and
produced inputs. `pins.json` fixes the upstream archive and producer revision;
`manifest.json` fixes all 7,744 extracted file hashes, per-script command counts,
producer correction and text-to-binary conversions. CI verifies/extracts the
committed 1.3 MiB bundle without downloading a producer:

```sh
python3 scripts/core3_corpus.py
```

Reproduce using wasm-tools **1.258.0**, revision
`5c6d31c78f8bd503f558441ef9c732950a141d1a`:

```sh
python3 scripts/core3_corpus.py --generate --producer /path/to/wasm-tools \
  --output artifacts/core3-reproduce
```

Generation checks the pinned source archive's size/digest. The producer binary
checksum is retained in `producer.json` in the selected output directory.
The local macOS arm64 producer checksum is
`f6d76114ae7a25d972e64e51598b1788572fd4d1d0b3f7f24da8a01f77864d9f`.
Two independent generations produce byte-identical bundles.

One deliberate multiple-supertype negative cannot be parsed by this producer.
The adapter temporarily removes its second parent in an owned source copy,
then restores both parent indices in the generated binary and increases the
type-section length. Input/output binary hashes are checked before publication;
the original `assert_invalid`, source line and unmodified original source are
retained. No successful validation is substituted for that negative assertion.

Ten valid quoted-text modules are compiled to binary and actually instantiated
and executed. Both original text and produced bytes are retained, with explicit
conversion records. Of the separately classified text-parser assertions, 1,233
are `assert_malformed` and six are quoted `assert_invalid` forms whose numeric
operands exceed their binary representation. Text parsing is a producer concern;
it is not claimed as a VM execution result. The six remaining text negatives
are not mislabeled as binary assertions.

Statement coverage, native execution and source/fixture verification are
separate evidence. Core Threads, WASI and Component Model tests are outside
this core specification inventory.

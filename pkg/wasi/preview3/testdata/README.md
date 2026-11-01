# WASI 0.3.2 provider fixtures

`build_preview3_base_fixtures.py` builds 20 base/clock/command fixtures;
`build_preview3_io_fixtures.py` builds 11 native-stream/terminal fixtures. Both use
wasm-tools 1.258.0 parse/validate with all features. The JSON manifests hash WAT,
binaries and generators. Base provenance also hashes reused owned P2 fixture
sources and pins the WASI 0.3.2 WIT archive. Fixtures are project-authored provider
integration tests, not a complete upstream conformance suite or complete world.

`build_preview3_filesystem_schema.py` derives exact function/value contracts from
the pinned WIT archive after verifying its files. The filesystem fixture generator
builds four filesystem fixtures plus a stdio cancellation/progress regression
adapted from an owned component fixture. filesystem-schema.json and filesystem.json
pin the schemas, generators, source input and all WAT/binary outputs.

filesystem-guest contains a real async Core consumer of preopens/open/stat/drop;
filesystem-api exposes the full import contract for embedding integration tests.
Malformed flags/time variants test binding rejection. The cancellation guest
checks canonical progress reporting when an OS read wins a cancellation race.
Wasmtime 48.0.1 executes the two open/stat guest cases; its rejection of imported
function reexports in the API fixture is separate reference evidence.

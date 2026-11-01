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

The socket schema/fixture generators verify the same pinned WIT archive and
produce four independently validated fixtures. sockets.json and
sockets-schema.json pin generators, schemas, WAT and binaries. sockets-guest
contains a real Core async UDP producer; sockets-api exposes all 41 operations
for provider tests. The two negative fixtures alter the error variant and async
connect signature. Explicit-authority Wasmtime datagram execution and the denied
guest assertion are recorded separately from API-only embedding tests.

HTTP fixtures use `scripts/build_preview3_http_fixtures.py` and the pinned
`http-schema.json`/`http.json` provenance. The API fixture covers all 37 operations;
malformed async/header-error contracts are rejected during binding. The outgoing
Core guest sets an explicit URI and uses native trailer/transmission futures.
The server guest returns its response before producing bounded byte batches,
then observes trailer/transmission completion. Both real guests also execute in
Wasmtime 48.0.1. `http-handler-api` exercises an explicitly supplied upstream
handler. These six fixtures are integration evidence, not full independent
language-toolchain or upstream world conformance.

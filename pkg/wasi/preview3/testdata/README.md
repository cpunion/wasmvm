# WASI 0.3.2 provider fixtures

`build_preview3_base_fixtures.py` builds 20 base/clock/command fixtures;
`build_preview3_io_fixtures.py` builds 11 native-stream/terminal fixtures. Both use
wasm-tools 1.258.0 parse/validate with all features. The JSON manifests hash WAT,
binaries and generators. Base provenance also hashes reused owned P2 fixture
sources and pins the WASI 0.3.2 WIT archive. Fixtures are project-authored provider
integration tests, not a complete upstream conformance suite or complete world.

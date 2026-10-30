# Canonical ABI reference inputs

The unchanged `definitions.py` comes from the Component Model specification
revision `d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`,
`design/mvp/canonical-abi/definitions.py`. Its Apache-2.0 notice is retained.
`generate.py` creates 512 deterministic nested value types, evaluates their
memory layout and flattened Core types with that independent reference, and
records 1,024 cases across both memory address widths. This is foundation-level
ABI validation, not complete component conformance or WASI interface coverage.

```sh
python3 testdata/component-canonical/generate.py
go test ./internal/component/canonical
```

The generator seed, exact inventory and all input/output hashes are recorded
in `manifest.json`. The production implementation never imports or executes
the Python reference. Cases cover primitives, strings/lists, records/tuples,
variants, flags/enums, option/results and mixed nested payloads. Separate Go
tests cover malformed inputs, resource/lifecycle behavior and quota boundaries.

The collection memory oracle adds 56 independent byte-level store/load cases
for maps (including ordered duplicate keys) and fixed lists, with memory32/64.
`generate_collections.py` invokes the same unchanged pinned reference and records
the type-directed host values and exact guest bytes. Its source/output hashes
are in the manifest; these cases are separate from the 1,024 layout cases.

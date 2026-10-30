# WASI 0.2.12 interface inputs

`cli.wit` is copied unchanged from the verified WASI release archive recorded in
`manifest.json`. The pinned WASI repository revision is
`281ba75fafcd50961ef55f9e52747afcc9b71ede`; its CLI license notice is preserved
as `LICENSE.WASI.md`. File/archive SHA256 values remain separate provenance.

Only `wasi:cli/environment@0.2.12` is implemented in this increment, covering
get-environment, get-arguments and initial-cwd. No complete CLI or HTTP world
is claimed. Runtime fixtures are project-authored WAT converted and validated
with wasm-tools 1.258.0; they are not an upstream conformance suite.

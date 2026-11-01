# CLI execution fixtures

`wasi-stdio.wat`/`.wasm` are owned P1 stdio/cancellation fixtures.
`components.json` hashes three owned command fixtures generated with
`scripts/build_component_cli_fixtures.py` and wasm-tools 1.258.0 parse/validate.
The generator and reused owned P3 stdin source are pinned separately.

The P2/P3 initialization commands each spend three instructions before their
two-instruction run function, checking a single total fuel budget. The P3 stdin
command adds a standard command entry to the existing real stream/future
consumer; pipe cancellation exercises pending readiness and joined cleanup.
CLI tests also execute the independently produced C/Rust world fixtures directly
from their existing pinned provider inventories.

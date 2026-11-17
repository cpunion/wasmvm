Describe the problem, resulting behavior, and validation.

Include a fresh complete comparison for this PR. Keep the body concise and link
the complete tables, pinned versions, commands, binary/fixture hashes, raw
evidence, own baseline deltas, and measurement controls.

| Workload / phase / unit | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | --- | --- | --- | --- | --- | --- |

Give interpreter, native execution, and P3 separate tables. Report startup,
preparation, first call, and warm execution separately. Cover every declared
portfolio case in the linked report; the body may show the key rows. Label
unsupported profiles, unmeasured cases, incomparable boundaries, and unresolved
controls explicitly. Do not substitute historical results for current PR data or
claim an overall win from a selected subset.

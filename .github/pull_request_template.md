Describe the problem, resulting behavior, and validation.

Include a fresh complete comparison for this PR. Keep the body concise and link
the complete tables, pinned versions, commands, binary/fixture hashes, raw
evidence, own baseline deltas, and measurement controls.

| Workload / phase / unit | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | --- | --- | --- | --- | --- | --- |

Show absolute observations followed by signed speed differences, for example
`100 ns (+25.00%)`, and name the reference runtime per table/row. Positive means
faster. For time `t`, speed Δ = `(reference / candidate - 1) * 100%`; for
throughput `q`, speed Δ = `(candidate / reference - 1) * 100%`. Own source/SDK
comparisons also show quantity Δ = `(candidate / baseline - 1) * 100%`, keeping
time reduction distinct from speedup. Transform both confidence bounds with the
same formula, reversing their order for reciprocal speed ratios. Retain
unresolved controls and missing/unsupported/failing cells. Use descriptive
median differences in the observation tables and paired block ratios/CIs for
qualified claims. Allocation/instruction differences are quantity deltas, not
execution-speed claims. Keep phases with incomparable API boundaries separate.

Give interpreter, native execution, and P3 separate tables. Report startup,
preparation, first call, and warm execution separately. Cover every declared
portfolio case in the linked report; the body may show the key rows. Label
unsupported profiles, unmeasured cases, incomparable boundaries, and unresolved
controls explicitly. Do not substitute historical results for current PR data or
claim an overall win from a selected subset.

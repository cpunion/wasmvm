# Relaxed SIMD proposal corpus

All seven scripts in `test/core/relaxed-simd` come from
[WebAssembly/relaxed-simd](https://github.com/WebAssembly/relaxed-simd/tree/c3f9359af2cd607cc46b0a3274f90ea52543a2f2),
revision `c3f9359af2cd607cc46b0a3274f90ea52543a2f2`.

The manifest pins all 23 source/license/fixture hashes and script counts.
WABT `wast2json` 1.0.41, revision
`03a00a1334e6121fb0cce4fccbd6bb109b68acaa`, produced each fixture directory from
its same-named `.wast` source with:

```sh
wast2json SOURCE.wast --enable-relaxed-simd --no-check -o SOURCE/script.json
```

Each interpreter/native tier executes 93 commands and 85 binary assertions
across all seven scripts, with no text-only assertions. Coverage includes all
20 instructions, permitted alternate vectors, rounding, NaNs, signed zeros,
conversion boundaries, out-of-range swizzles, masks, overflow and consistency.
The runner matches complete `either` vectors. Owned tests additionally cover
feature composition, native registers/spills, fuel, cancellation and limits.
`LICENSE.tests` retains the upstream Apache-2.0 notice. CI requires no producer
download. This suite does not establish complete Core 3.0 support.

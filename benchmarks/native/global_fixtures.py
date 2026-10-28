#!/usr/bin/env python3
"""Owned numeric-global loops with independent integer checksum oracles."""

import hashlib
import json
from pathlib import Path
import subprocess

KINDS = ("i32", "i64", "f32", "f64", "v128")


def checksum(kind, seed, count):
    initial = seed & 1023
    if kind == "v128":
        lanes = [(initial + count * step) & 0xffffffff for step in (1, 3, 5, 7)]
        return lanes[0] ^ lanes[1] ^ lanes[2] ^ lanes[3]
    if kind not in KINDS:
        raise ValueError(kind)
    # f32 inputs and the declared 10000-iteration run remain exact integers.
    return (initial + 3 * count) & 0xffffffff


def source(kind):
    init = "local.get 0 i32.const 1023 i32.and"
    if kind == "v128":
        zero = "v128.const i32x4 0 0 0 0"
        init += " i32x4.splat"
        step = "v128.const i32x4 1 3 5 7 i32x4.add"
        result = "global.get $g i32x4.extract_lane 0 " + " ".join(
            f"global.get $g i32x4.extract_lane {i} i32.xor" for i in (1, 2, 3))
    else:
        zero = f"{kind}.const 0"
        if kind != "i32":
            init += " " + {"i64": "i64.extend_i32_u", "f32": "f32.convert_i32_u", "f64": "f64.convert_i32_u"}[kind]
        step = f"{kind}.const 3 {kind}.add"
        result = "global.get $g"
        if kind != "i32":
            result += " " + {"i64": "i32.wrap_i64", "f32": "i32.trunc_f32_u", "f64": "i32.trunc_f64_u"}[kind]
    return f"""(module
  (global $g (mut {kind}) ({zero}))
  (func (export "run") (param i32 i32) (result i32) (local $i i32)
    {init} global.set $g
    (loop $loop
      global.get $g {step} global.set $g
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    {result}))
"""


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise ValueError(version)
    root = Path(__file__).with_name("global-fixtures")
    root.mkdir(exist_ok=True)
    rows = []
    for kind in KINDS:
        wat, wasm = root / (kind + ".wat"), root / (kind + ".wasm")
        wat.write_text(source(kind))
        subprocess.run(["wasm-tools", "parse", str(wat), "-o", str(wasm)], check=True)
        subprocess.run(["wasm-tools", "validate", str(wasm)], check=True)
        seed, count = 0x12345678, 10000
        rows.append(dict(id=kind, seed=seed, iterations=count, expectedChecksum=checksum(kind, seed, count),
                         unit="ns/source iteration", wat=wat.name, wasm=wasm.name,
                         watSha256=hashlib.sha256(wat.read_bytes()).hexdigest(),
                         wasmSha256=hashlib.sha256(wasm.read_bytes()).hexdigest()))
    (root / "manifest.json").write_text(json.dumps(dict(producer=version,
        origin="authored in wasmvm; no third-party benchmark code", workloads=rows), indent=2) + "\n")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Generate owned full-runtime workloads and independent scalar oracles."""

import hashlib
import json
from pathlib import Path
import subprocess

MASK = 0xffffffff


def checksum(kind, seed, count):
    x = seed
    if kind == "simd-i32x4":
        lanes = [seed] * 4
        for i in range(count):
            lanes = [(((v + a) * b) & MASK) ^ i
                     for v, a, b in zip(lanes, (1, 3, 5, 7), (3, 5, 7, 11))]
        return lanes[0] ^ lanes[1] ^ lanes[2] ^ lanes[3]
    for i in range(count):
        if kind == "indirect-calls":
            x = (((x << 13) | (x >> 19)) & MASK) ^ 0x9e3779b9 if i & 1 else (x * 1664525 + 1013904223) & MASK
        elif kind == "bulk-overlap":
            x = (x + ((x + i) & 255)) & MASK
        elif kind == "globals-tables":
            x = (x + i + 1) & MASK
        else:
            raise ValueError(kind)
    return x


def source(kind):
    setup, locals_, init, body, result = "", "(local $x i32)", "local.get 0 local.set $x", "", "local.get $x"
    if kind == "indirect-calls":
        setup = """(type $step (func (param i32) (result i32)))
          (func $a (type $step) local.get 0 i32.const 1664525 i32.mul i32.const 1013904223 i32.add)
          (func $b (type $step) local.get 0 i32.const 13 i32.rotl i32.const 0x9e3779b9 i32.xor)
          (table 2 funcref) (elem (i32.const 0) $a $b)"""
        body = "local.get $x local.get $i i32.const 1 i32.and call_indirect (type $step) local.set $x"
    elif kind == "simd-i32x4":
        locals_ = "(local $v v128)"
        init = "local.get 0 i32x4.splat local.set $v"
        body = """local.get $v v128.const i32x4 1 3 5 7 i32x4.add
          v128.const i32x4 3 5 7 11 i32x4.mul
          local.get $i i32x4.splat v128.xor local.set $v"""
        result = "local.get $v i32x4.extract_lane 0 " + " ".join(f"local.get $v i32x4.extract_lane {i} i32.xor" for i in (1, 2, 3))
    elif kind == "bulk-overlap":
        setup = "(memory 1 1)"
        body = """i32.const 256 local.get $x local.get $i i32.add i32.const 8192 memory.fill
          i32.const 272 i32.const 256 i32.const 8192 memory.copy
          local.get $x i32.const 287 i32.load8_u i32.add local.set $x"""
    elif kind == "globals-tables":
        setup = """(global $g (mut i32) (i32.const 0)) (func $target)
          (table 2 funcref) (elem declare func $target)"""
        init = "local.get 0 global.set $g"
        body = """local.get $i i32.const 1 i32.and ref.func $target table.set 0
          global.get $g local.get $i i32.add
          local.get $i i32.const 1 i32.and table.get 0 ref.is_null i32.eqz
          i32.add global.set $g"""
        result = "global.get $g"
    else:
        raise ValueError(kind)
    return f"""(module{(" " + setup) if setup else ""}
      (func (export "run") (param i32 i32) (result i32) (local $i i32) {locals_}
        {init}
        (loop $loop {body}
          local.get $i i32.const 1 i32.add local.tee $i
          local.get 1 i32.lt_u br_if $loop)
        {result}))
"""


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise ValueError(version)
    root = Path(__file__).with_name("runtime-fixtures")
    root.mkdir(exist_ok=True)
    rows = []
    for kind, iterations in (("indirect-calls", 15000), ("simd-i32x4", 10000), ("bulk-overlap", 1000), ("globals-tables", 10000)):
        wat, wasm = root / f"{kind}.wat", root / f"{kind}.wasm"
        wat.write_text(source(kind))
        subprocess.run(["wasm-tools", "parse", str(wat), "-o", str(wasm)], check=True)
        subprocess.run(["wasm-tools", "validate", str(wasm)], check=True)
        seed = 0x12345678
        rows.append({"id": kind, "seed": seed, "iterations": iterations,
                     "expectedChecksum": checksum(kind, seed, iterations),
                     "unit": "ns/source iteration", "wat": wat.name, "wasm": wasm.name,
                     "watSha256": hashlib.sha256(wat.read_bytes()).hexdigest(),
                     "wasmSha256": hashlib.sha256(wasm.read_bytes()).hexdigest()})
    (root / "manifest.json").write_text(json.dumps({"producer": version,
        "origin": "authored in wasmvm; no third-party benchmark code", "workloads": rows}, indent=2) + "\n")


if __name__ == "__main__":
    main()

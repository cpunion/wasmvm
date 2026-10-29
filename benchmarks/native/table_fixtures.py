#!/usr/bin/env python3
"""Owned table/reference loops and independent handle-identity oracles."""

import hashlib
import json
from pathlib import Path
import subprocess

KINDS = ("table-roundtrip", "table-copy-handles", "ref-func")
MASK = 0xffffffff


def checksum(kind, seed, count):
    if kind not in KINDS:
        raise ValueError(kind)
    table, x = [0] * 16, seed
    for i in range(count):
        if kind == "table-copy-handles":
            table[i & 15] = 7 if ((i + seed) & 1) == 0 else 11
            null = 0
        else:
            table[(i + seed) & 15] = 7 if kind == "ref-func" or i & 1 == 0 else 0
            null = int(table[(i + seed) & 15] == 0)
        x = (x * 1664525 + null) & MASK
    return (x + sum((i + 1) * v for i, v in enumerate(table))) & MASK


def source(kind):
    setup = "(table $t 16 funcref) (elem declare func $a $b)"
    target_table = "$t"
    if kind == "table-copy-handles":
        setup = "(table $source 16 funcref) (table $t 16 funcref) (elem (table $source) (i32.const 0) func " + " ".join(["$a", "$b"] * 8) + ")"
        body = """local.get $i i32.const 15 i32.and
      local.get $i local.get 0 i32.add i32.const 15 i32.and table.get $source
      table.set $t
      local.get $x i32.const 1664525 i32.mul
      local.get $i i32.const 15 i32.and table.get $t ref.is_null i32.add local.set $x"""
    else:
        ref = "ref.func $a" if kind == "ref-func" else "local.get $i i32.const 1 i32.and if (result funcref) ref.null func else ref.func $a end"
        body = f"""local.get $i local.get 0 i32.add i32.const 15 i32.and
      {ref} table.set $t
      local.get $x i32.const 1664525 i32.mul
      local.get $i local.get 0 i32.add i32.const 15 i32.and table.get $t ref.is_null i32.add local.set $x"""
    # Weighted indirect calls verify the produced references' identities, rather
    # than checking only null/non-null. These calls remain inside guest timing.
    tail = "\n".join(f"""local.get $x i32.const {i + 1}
    i32.const {i} table.get {target_table} ref.is_null
    if (result i32) i32.const 0 else i32.const {i} call_indirect {target_table} (type $target) end
    i32.mul i32.add local.set $x""" for i in range(16))
    return f"""(module
  (type $target (func (result i32)))
  (func $a (type $target) i32.const 7)
  (func $b (type $target) i32.const 11)
  {setup}
  (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
    local.get 0 local.set $x
    (loop $loop
      {body}
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    {tail}
    local.get $x))
"""


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise ValueError(version)
    root = Path(__file__).with_name("table-fixtures")
    root.mkdir(exist_ok=True)
    rows = []
    for kind in KINDS:
        wat, wasm = root / (kind + ".wat"), root / (kind + ".wasm")
        wat.write_text(source(kind))
        subprocess.run(["wasm-tools", "parse", str(wat), "-o", str(wasm)], check=True)
        subprocess.run(["wasm-tools", "validate", str(wasm)], check=True)
        seed, count = 0x12345678, 100000
        rows.append(dict(id=kind, seed=seed, iterations=count, expectedChecksum=checksum(kind, seed, count),
                         unit="ns/source iteration", wat=wat.name, wasm=wasm.name,
                         watSha256=hashlib.sha256(wat.read_bytes()).hexdigest(),
                         wasmSha256=hashlib.sha256(wasm.read_bytes()).hexdigest()))
    (root / "manifest.json").write_text(json.dumps(dict(producer=version,
        origin="authored in wasmvm; no third-party benchmark code", workloads=rows), indent=2) + "\n")


if __name__ == "__main__":
    main()

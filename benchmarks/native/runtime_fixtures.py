#!/usr/bin/env python3
"""Generate owned full-runtime workloads and independent scalar oracles."""

import hashlib
import json
from pathlib import Path
import subprocess
import struct

MASK = 0xffffffff


def checksum(kind, seed, count):
    x = seed
    if kind.startswith("simd-memory-"):
        return memory_checksum(kind, seed, count)
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


def memory_checksum(kind, seed, count):
    memory = bytearray((i * 193 + 127) & 255 for i in range(256))
    lanes = [seed] * 4
    for i in range(count):
        if kind == "simd-memory-full":
            address = (i & 127) + 1
            struct.pack_into("<4I", memory, address, *lanes)
            memory[address + 16] = (seed + i) & 255
            lanes = list(struct.unpack_from("<4I", memory, address + 1))
            lanes = [(v + a) & MASK for v, a in zip(lanes, (1, 3, 5, 7))]
        elif kind == "simd-memory-widen":
            packed = struct.pack("<4I", *lanes)
            old = struct.unpack("<8H", packed)
            wide = struct.unpack_from("<8b", memory, i & 127)
            scalar = struct.unpack_from("<H", memory, (i * 17) & 127)[0]
            updated = [((v ^ (w & 65535)) + scalar) & 65535 for v, w in zip(old, wide)]
            lanes = list(struct.unpack("<4I", struct.pack("<8H", *updated)))
        elif kind == "simd-memory-lanes":
            address = (i & 127) + 1
            struct.pack_into("<I", memory, address, lanes[3])
            memory[address + 4] = (seed + i) & 255
            lanes[1] = struct.unpack_from("<I", memory, address + 1)[0]
            lanes[0] ^= struct.unpack_from("<I", memory, address)[0]
            lanes = [(v + a) & MASK for v, a in zip(lanes, (1, 3, 5, 7))]
        else:
            raise ValueError(kind)
    return lanes[0] ^ lanes[1] ^ lanes[2] ^ lanes[3]


def source(kind):
    setup, locals_, init, body, result = "", "(local $x i32)", "local.get 0 local.set $x", "", "local.get $x"
    if kind == "indirect-calls":
        setup = """(type $step (func (param i32) (result i32)))
          (func $a (type $step) local.get 0 i32.const 1664525 i32.mul i32.const 1013904223 i32.add)
          (func $b (type $step) local.get 0 i32.const 13 i32.rotl i32.const 0x9e3779b9 i32.xor)
          (table 2 funcref) (elem (i32.const 0) $a $b)"""
        body = "local.get $x local.get $i i32.const 1 i32.and call_indirect (type $step) local.set $x"
    elif kind == "simd-i32x4" or kind.startswith("simd-memory-"):
        locals_ = "(local $v v128)"
        init = "local.get 0 i32x4.splat local.set $v"
        body = """local.get $v v128.const i32x4 1 3 5 7 i32x4.add
          v128.const i32x4 3 5 7 11 i32x4.mul
          local.get $i i32x4.splat v128.xor local.set $v"""
        result = "local.get $v i32x4.extract_lane 0 " + " ".join(f"local.get $v i32x4.extract_lane {i} i32.xor" for i in (1, 2, 3))
        if kind.startswith("simd-memory-"):
            data = "".join(f"\\{(i * 193 + 127) & 255:02x}" for i in range(256))
            setup = f'(memory 1 1) (data (i32.const 0) "{data}")'
            locals_ += " (local $address i32)"
            if kind == "simd-memory-full":
                body = """local.get $i i32.const 127 i32.and i32.const 1 i32.add local.set $address
                  local.get $address local.get $v v128.store align=1
                  local.get $address local.get 0 local.get $i i32.add i32.store8 offset=16
                  local.get $address v128.load offset=1 align=1
                  v128.const i32x4 1 3 5 7 i32x4.add local.set $v"""
            elif kind == "simd-memory-widen":
                body = """local.get $v local.get $i i32.const 127 i32.and v128.load8x8_s align=1 v128.xor
                  local.get $i i32.const 17 i32.mul i32.const 127 i32.and v128.load16_splat align=1
                  i16x8.add local.set $v"""
            elif kind == "simd-memory-lanes":
                body = """local.get $i i32.const 127 i32.and i32.const 1 i32.add local.set $address
                  local.get $address local.get $v v128.store32_lane align=1 3
                  local.get $address local.get 0 local.get $i i32.add i32.store8 offset=4
                  local.get $address local.get $v v128.load32_lane offset=1 align=1 1
                  local.get $address v128.load32_zero align=1 v128.xor
                  v128.const i32x4 1 3 5 7 i32x4.add local.set $v"""
            else:
                raise ValueError(kind)
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
    for kind, iterations in (("indirect-calls", 15000), ("simd-i32x4", 10000), ("bulk-overlap", 1000), ("globals-tables", 10000),
                             ("simd-memory-full", 10000), ("simd-memory-widen", 10000), ("simd-memory-lanes", 10000)):
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

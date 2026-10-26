#!/usr/bin/env python3
"""Generate the project's own deterministic Core/WASI P1 benchmark fixtures."""

import hashlib
import json
from pathlib import Path
import subprocess


ROOT = Path(__file__).resolve().parent
MASK = (1 << 32) - 1


def rotate_left(value, count):
    return ((value << count) | (value >> (32 - count))) & MASK


def definitions():
    yield "startup", 0, 0, "", "    i32.const 0\n", "startup/P1 output control with no workload loop"

    integer_count = 250_000
    value = 0x12345678
    for index in range(integer_count):
        value = ((rotate_left(value ^ index, 5) * 1664525) + 1013904223) & MASK
    integer_body = f"""
    (local $i i32) (local $x i32)
    i32.const 305419896 local.set $x
    loop $again
      local.get $x local.get $i i32.xor i32.const 5 i32.rotl
      i32.const 1664525 i32.mul i32.const 1013904223 i32.add local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      i32.const {integer_count} i32.lt_u br_if $again
    end
    local.get $x
"""
    yield "integer-loop", integer_count, value, "", integer_body, "i32 rotate/xor/multiply/add loop"

    sustained_integer_count = integer_count * 20
    value = 0x12345678
    for index in range(sustained_integer_count):
        value = ((rotate_left(value ^ index, 5) * 1664525) + 1013904223) & MASK
    sustained_integer_body = integer_body.replace(
        f"i32.const {integer_count} i32.lt_u", f"i32.const {sustained_integer_count} i32.lt_u"
    )
    yield "integer-sustained", sustained_integer_count, value, "", sustained_integer_body, "20x integer-loop work; same operations and state evolution"

    call_count = 150_000
    value = 7
    for index in range(call_count):
        value = (rotate_left((value + index) & MASK, 7) ^ 0x9E3779B9) & MASK
    helper = """
  (func $mix (param $x i32) (param $i i32) (result i32)
    local.get $x local.get $i i32.add i32.const 7 i32.rotl
    i32.const 2654435769 i32.xor)
"""
    call_body = f"""
    (local $i i32) (local $x i32)
    i32.const 7 local.set $x
    loop $again
      local.get $x local.get $i call $mix local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      i32.const {call_count} i32.lt_u br_if $again
    end
    local.get $x
"""
    yield "function-calls", call_count, value, helper, call_body, "direct-call source workload; native backends may inline"

    memory_count = 150_000
    memory = [0] * 8192
    checksum = 0
    for index in range(memory_count):
        slot = (index * 17) & 8191
        value = ((memory[slot] ^ index) + 0x9E3779B9) & MASK
        memory[slot] = value
        checksum = (checksum + value) & MASK
    memory_body = f"""
    (local $i i32) (local $address i32) (local $x i32) (local $sum i32)
    loop $again
      local.get $i i32.const 17 i32.mul i32.const 8191 i32.and
      i32.const 4 i32.mul i32.const 4096 i32.add local.set $address
      local.get $address i32.load local.get $i i32.xor
      i32.const 2654435769 i32.add local.set $x
      local.get $address local.get $x i32.store
      local.get $sum local.get $x i32.add local.set $sum
      local.get $i i32.const 1 i32.add local.tee $i
      i32.const {memory_count} i32.lt_u br_if $again
    end
    local.get $sum
"""
    yield "memory-load-store", memory_count, checksum, "", memory_body, "strided read/modify/write over 32 KiB of linear memory"

    sustained_memory_count = memory_count * 20
    memory = [0] * 8192
    checksum = 0
    for index in range(sustained_memory_count):
        slot = (index * 17) & 8191
        value = ((memory[slot] ^ index) + 0x9E3779B9) & MASK
        memory[slot] = value
        checksum = (checksum + value) & MASK
    sustained_memory_body = memory_body.replace(
        f"i32.const {memory_count} i32.lt_u", f"i32.const {sustained_memory_count} i32.lt_u"
    )
    yield "memory-sustained", sustained_memory_count, checksum, "", sustained_memory_body, "20x memory-load-store work over the same 32 KiB region"


def module(identifier, expected, helper, body):
    output = f"{identifier}:{expected}\n"
    wat_output = output.replace("\n", "\\0a")
    wat = f""";; Authored for wasmvm. Regenerate with: python3 benchmarks/generate.py
;; The computed checksum is checked before any success output is written.
(module
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory (export "memory") 1)
  (data (i32.const 16) "{wat_output}")
{helper}
  (func $work (result i32)
{body})
  (func (export "_start")
    call $work i32.const {expected} i32.ne
    if unreachable end
    i32.const 0 i32.const 16 i32.store
    i32.const 4 i32.const {len(output)} i32.store
    i32.const 1 i32.const 0 i32.const 1 i32.const 8 call $fd_write
    if unreachable end))
"""
    return wat, output


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    entries = []
    for identifier, count, expected, helper, body, description in definitions():
        wat, output = module(identifier, expected, helper, body)
        wat_path = ROOT / "fixtures" / f"{identifier}.wat"
        wasm_path = wat_path.with_suffix(".wasm")
        wat_path.parent.mkdir(parents=True, exist_ok=True)
        wat_path.write_text(wat)
        subprocess.run(["wasm-tools", "parse", str(wat_path), "-o", str(wasm_path)], check=True)
        subprocess.run(["wasm-tools", "validate", str(wasm_path)], check=True)
        entries.append({
            "id": identifier, "iterations": count, "description": description,
            "expectedChecksum": expected, "expectedStdout": output,
            "wat": f"fixtures/{identifier}.wat", "wasm": f"fixtures/{identifier}.wasm",
            "watSha256": hashlib.sha256(wat_path.read_bytes()).hexdigest(),
            "wasmSha256": hashlib.sha256(wasm_path.read_bytes()).hexdigest(),
        })
    manifest = {
        "schemaVersion": 1, "origin": "authored in wasmvm; no third-party benchmark code",
        "producer": version, "generationCommand": "python3 benchmarks/generate.py",
        "profile": "Core MVP + WASI Preview 1 fd_write; one 64 KiB memory",
        "workloads": entries,
    }
    (ROOT / "fixtures" / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Generated and validated {len(entries)} fixtures with {version}")


if __name__ == "__main__":
    main()

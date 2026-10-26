#!/usr/bin/env python3
"""Generate owned, parameterized Core MVP fixtures for embedded phase metrics."""

import hashlib
import json
from pathlib import Path
import subprocess


ROOT = Path(__file__).resolve().parent
MASK = (1 << 32) - 1
MAX_ITERATIONS = (1 << 31) - 1
MEMORY_WORDS = 8192
SEED = 0x12345678
KINDS = ("call-control", "integer-mix", "function-calls", "memory-load-store")
CASES = (
    ("call-control", "call-control", 1),
    ("integer-loop", "integer-mix", 250_000),
    ("integer-sustained", "integer-mix", 5_000_000),
    ("function-calls", "function-calls", 150_000),
    ("memory-load-store", "memory-load-store", 150_000),
    ("memory-sustained", "memory-load-store", 3_000_000),
)


def rotate_left(value, count):
    return ((value << count) | (value >> (32 - count))) & MASK


def checksum(kind, seed, iterations):
    """Independent Python arithmetic oracle; never execute the generated Wasm."""
    if kind not in KINDS:
        raise ValueError("unknown workload kind")
    if type(seed) is not int or not 0 <= seed <= MASK:
        raise ValueError("seed must be an unsigned 32-bit integer")
    if type(iterations) is not int or not 1 <= iterations <= MAX_ITERATIONS:
        raise ValueError("iterations must be a positive signed 32-bit integer")
    if kind == "call-control":
        if iterations != 1:
            raise ValueError("call-control has exactly one host invocation, not a guest loop")
        return seed
    if kind == "integer-mix":
        value = seed
        for index in range(iterations):
            value = (rotate_left(value ^ index, 5) * 1664525 + 1013904223) & MASK
        return value
    if kind == "function-calls":
        value = seed
        for index in range(iterations):
            value = rotate_left((value + index) & MASK, 7) ^ 0x9E3779B9
        return value
    # Reset is part of every invocation, including repeated calls on one instance.
    memory = [0] * MEMORY_WORDS
    result = seed
    for index in range(iterations):
        slot = (index * 17) % MEMORY_WORDS
        value = ((memory[slot] ^ index ^ seed) + 0x9E3779B9) & MASK
        memory[slot] = value
        result = (result + value) & MASK
    return result


def module(kind):
    if kind not in KINDS:
        raise ValueError("unknown workload kind")
    declarations = ""
    locals_ = ""
    guard = "local.get $iterations i32.const 0 i32.le_s if unreachable end"
    if kind == "call-control":
        body = """
    local.get $iterations i32.const 1 i32.ne if unreachable end
    local.get $seed
"""
    elif kind == "integer-mix":
        locals_ = "(local $i i32) (local $x i32)"
        body = """
    local.get $seed local.set $x
    loop $again
      local.get $x local.get $i i32.xor i32.const 5 i32.rotl
      i32.const 1664525 i32.mul i32.const 1013904223 i32.add local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x
"""
    elif kind == "function-calls":
        declarations = """
  ;; A source-level direct call per iteration; native backends may inline it.
  (func $mix (param $x i32) (param $i i32) (result i32)
    local.get $x local.get $i i32.add i32.const 7 i32.rotl
    i32.const 2654435769 i32.xor)
"""
        locals_ = "(local $i i32) (local $x i32)"
        body = """
    local.get $seed local.set $x
    loop $again
      local.get $x local.get $i call $mix local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x
"""
    else:
        declarations = "  (memory 1 1)\n"
        locals_ = "(local $clear i32) (local $i i32) (local $address i32) (local $x i32) (local $sum i32)"
        body = """
    ;; Exactly 8192 i32 stores reset the 32 KiB working region on every call.
    ;; This cost is included in run; it is not charged to instantiation alone.
    loop $reset
      local.get $clear i32.const 0 i32.store
      local.get $clear i32.const 4 i32.add local.tee $clear
      i32.const 32768 i32.lt_u br_if $reset
    end
    local.get $seed local.set $sum
    loop $again
      local.get $i i32.const 17 i32.mul i32.const 8191 i32.and
      i32.const 4 i32.mul local.set $address
      local.get $address i32.load local.get $i i32.xor local.get $seed i32.xor
      i32.const 2654435769 i32.add local.set $x
      local.get $address local.get $x i32.store
      local.get $sum local.get $x i32.add local.set $sum
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $sum
"""
    source = f""";; Authored for wasmvm. Regenerate: python3 benchmarks/embedded/generate.py
;; Runtime parameters determine the work and checksum; no WASI/import/start code.
(module
{declarations}
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    {locals_}
    {guard}
{body}))
"""
    return "\n".join(line.rstrip() for line in source.splitlines()) + "\n"


def validation_cases(kind):
    pairs = ((0, 1), (1, 1), (MASK, 1), (0xDEADBEEF, 1)) if kind == "call-control" else (
        (0, 1), (1, 2), (MASK, 17), (0xDEADBEEF, 257), (7, MEMORY_WORDS + 1),
    )
    return [{"seed": seed, "iterations": count, "expectedChecksum": checksum(kind, seed, count)}
            for seed, count in pairs]


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    fixtures = ROOT / "fixtures"
    fixtures.mkdir(parents=True, exist_ok=True)
    artifacts = {}
    for kind in KINDS:
        wat = fixtures / f"{kind}.wat"
        wasm = wat.with_suffix(".wasm")
        wat.write_text(module(kind))
        subprocess.run(["wasm-tools", "parse", str(wat), "-o", str(wasm)], check=True)
        subprocess.run(["wasm-tools", "validate", str(wasm), "--features=mvp"], check=True)
        artifacts[kind] = {
            "wat": f"fixtures/{wat.name}", "wasm": f"fixtures/{wasm.name}",
            "watSha256": hashlib.sha256(wat.read_bytes()).hexdigest(),
            "wasmSha256": hashlib.sha256(wasm.read_bytes()).hexdigest(),
        }
    descriptions = {
        "call-control": "one exported host call returning its runtime seed; no guest work loop",
        "integer-mix": "i32 rotate/xor/multiply/add loop with runtime seed and loop count",
        "function-calls": "one direct helper call in the source per iteration; native backends may inline",
        "memory-load-store": "strided read/modify/write over 32 KiB; 8192 reset stores per invocation are included",
    }
    entries = []
    for identifier, kind, count in CASES:
        entries.append({
            "id": identifier, "kind": kind, "description": descriptions[kind],
            "seed": SEED, "iterations": count, "expectedChecksum": checksum(kind, SEED, count),
            "unit": "ns/call" if kind == "call-control" else "ns/iteration",
            "memoryResetBytes": MEMORY_WORDS * 4 if kind == "memory-load-store" else 0,
            "memoryResetIncluded": kind == "memory-load-store",
            "validationCases": validation_cases(kind), **artifacts[kind],
        })
    manifest = {
        "schemaVersion": 1,
        "origin": "authored in wasmvm; no third-party benchmark code",
        "producer": version,
        "generationCommand": "python3 benchmarks/embedded/generate.py",
        "profile": "Core MVP; no imports, WASI, start function, or data segments; only run is exported",
        "api": {"export": "run", "parameters": ["i32", "i32"], "result": "i32",
                "seed": "unsigned 32-bit bit pattern", "iterationsMin": 1,
                "iterationsMax": MAX_ITERATIONS, "controlIterations": 1,
                "checksum": "compare the result as an unsigned 32-bit bit pattern"},
        "measurement": {
            "iteration": "one source workload loop iteration, not one Wasm or native instruction",
            "control": "one host invocation, without a guest loop; do not divide by a fictitious loop count",
            "memory": "reset cost is included in every timed call and amortized over its workload iteration count",
            "hostBoundary": "timed calls include the embedding API boundary; the call control is reported separately",
        },
        "workloads": entries,
    }
    (fixtures / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Generated {len(artifacts)} MVP modules and {len(entries)} parameterized cases with {version}")


if __name__ == "__main__":
    main()

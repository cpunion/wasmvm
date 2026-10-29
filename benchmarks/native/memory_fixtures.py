"""Owned scalar memory workloads; byte-oriented independent Python oracle."""
import hashlib
import json
from pathlib import Path
import struct
import subprocess

ROOT = Path(__file__).resolve().parent
SEED, ITERATIONS = 305419896, 1_000_000
FAMILIES = ("narrow-unaligned", "dependent-address", "float64-unaligned", "growth-first-call")


def checksum(family, seed, count):
    if family == "growth-first-call":
        return seed ^ count
    memory = bytearray(65536)
    total = seed
    if family == "dependent-address":
        for i in range(4096):
            struct.pack_into("<I", memory, i*4, (i*109+seed) & 4095)
        address = seed & 4095
        for _ in range(count):
            address = struct.unpack_from("<I", memory, address*4)[0] & 4095
            total = (total+address) & 0xffffffff
    elif family == "narrow-unaligned":
        for i in range(count):
            address = ((i*13) & 2047)*4+1
            old = struct.unpack_from("<h", memory, address)[0] & 0xffffffff
            new = ((old ^ seed ^ i)+31337) & 65535
            struct.pack_into("<H", memory, address, new)
            total = (total + (new ^ memory[address+1])) & 0xffffffff
    else:
        total = 0.0
        for i in range(2048):
            struct.pack_into("<d", memory, i*8+1, float(seed & 255))
        for i in range(count):
            address = ((i*17) & 2047)*8+1
            number = struct.unpack_from("<d", memory, address)[0]+0.125
            struct.pack_into("<d", memory, address, number)
            total += number
        total = int(total)
    return total & 0xffffffff


def wat(family):
    if family == "growth-first-call":
        return '''(module (memory 1 65)
          (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
            i32.const 64 memory.grow i32.const -1 i32.ne
            if
              i32.const 65536 i32.load if unreachable end
            end
            memory.size i32.const 65 i32.ne if unreachable end
            i32.const 65536 local.get $seed local.get $iterations i32.xor i32.store
            i32.const 65536 i32.load))\n'''
    if family == "narrow-unaligned":
        reset = "local.get $j i32.const 4 i32.mul i32.const 0 i32.store"
        limit = 2049
        setup = "local.get $seed local.set $sum"
        body = '''local.get $i i32.const 13 i32.mul i32.const 2047 i32.and
          i32.const 4 i32.mul i32.const 1 i32.add local.set $address
          local.get $address i32.load16_s align=1 local.get $seed i32.xor local.get $i i32.xor
          i32.const 31337 i32.add i32.const 65535 i32.and local.set $x
          local.get $address local.get $x i32.store16 align=1
          local.get $sum local.get $x local.get $address i32.load8_u offset=1 i32.xor i32.add local.set $sum'''
        result = "local.get $sum"
    elif family == "dependent-address":
        reset = "local.get $j i32.const 4 i32.mul local.get $j i32.const 109 i32.mul local.get $seed i32.add i32.const 4095 i32.and i32.store"
        limit = 4096
        setup = "local.get $seed local.set $sum local.get $seed i32.const 4095 i32.and local.set $address"
        body = '''local.get $address i32.const 4095 i32.and i32.const 4 i32.mul i32.load
          i32.const 4095 i32.and local.tee $address local.get $sum i32.add local.set $sum'''
        result = "local.get $sum"
    else:
        reset = "local.get $j i32.const 8 i32.mul i32.const 1 i32.add local.get $seed i32.const 255 i32.and f64.convert_i32_u f64.store align=1"
        limit, setup = 2048, ""
        body = '''local.get $i i32.const 17 i32.mul i32.const 2047 i32.and
          i32.const 8 i32.mul i32.const 1 i32.add local.set $address
          local.get $address f64.load align=1 f64.const 0.125 f64.add local.set $number
          local.get $address local.get $number f64.store align=1
          local.get $total local.get $number f64.add local.set $total'''
        result = "local.get $total i32.trunc_f64_u"
    return f'''(module (memory 1 1)
      (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
        (local $i i32) (local $j i32) (local $address i32) (local $x i32) (local $sum i32)
        (local $number f64) (local $total f64)
        local.get $iterations i32.const 0 i32.le_s if unreachable end
        loop $reset
          {reset}
          local.get $j i32.const 1 i32.add local.tee $j i32.const {limit} i32.lt_u br_if $reset
        end
        {setup}
        loop $again
          {body}
          local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
        end
        {result}))\n'''


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise SystemExit("use pinned wasm-tools 1.258.0")
    directory = ROOT / "memory-fixtures"
    directory.mkdir(exist_ok=True)
    cases = []
    for family in FAMILIES:
        source, binary = directory / (family+".wat"), directory / (family+".wasm")
        source.write_text("\n".join(line.rstrip() for line in wat(family).splitlines())+"\n")
        subprocess.run(["wasm-tools", "parse", str(source), "-o", str(binary)], check=True)
        cases.append({"id": family, "wasm": binary.name, "seed": SEED, "iterations": ITERATIONS,
                      "expectedChecksum": checksum(family, SEED, ITERATIONS),
                      "measurement": "first call grows 64 pages; warm calls measure failed growth" if family == "growth-first-call" else "warm ns/iteration; deterministic reset included",
                      "wasmSha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                      "sourceSha256": hashlib.sha256(source.read_bytes()).hexdigest()})
    (directory / "manifest.json").write_text(json.dumps({"producer": version, "workloads": cases}, indent=2)+"\n")


if __name__ == "__main__":
    main()

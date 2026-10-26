"""Owned parameterized scalar FP fixtures; independent Python arithmetic oracle."""
import hashlib
import json
import math
from pathlib import Path
import struct
import subprocess

ROOT = Path(__file__).resolve().parent
SEED, ITERATIONS = 305419896, 1_000_000
FAMILIES = ("arithmetic", "minmax", "rounding")


def initial(seed, wide):
    bits = 0x3f800000 | (seed & 0x807fffff)
    if wide:
        bits = 0x3ff0000000000000 | ((seed & 0x7fffff) << 29) | ((seed & 0x80000000) << 32)
    return struct.unpack("<d" if wide else "<f", struct.pack("<Q" if wide else "<I", bits))[0]


def checksum(family, wide, seed, iterations):
    def fp(x):
        return x if wide else struct.unpack("<f", struct.pack("<f", x))[0]
    x = initial(seed, wide)
    for _ in range(iterations):
        if family == "arithmetic":
            x = fp(fp(fp(x * 1.0001220703125) + 0.125) / 1.125)
        elif family == "minmax":
            x = max(min(fp(x + 0.125), 2.0), -2.0)
            if x > 1.5:
                x = -x
        else:
            y = fp(fp(fp(math.ceil(x) + math.floor(x)) + math.trunc(x)) + round(x))
            x = fp(fp(fp(math.sqrt(fp(fp(x * x) + 0.5))) / fp(y + 8.0)) + 0.25)
    bits = struct.unpack("<Q" if wide else "<I", struct.pack("<d" if wide else "<f", x))[0]
    return (bits ^ (bits >> 32)) & 0xffffffff


def wat(family, wide):
    t = "f64" if wide else "f32"
    if wide:
        init = "local.get $seed i32.const 0x7fffff i32.and i64.extend_i32_u i64.const 29 i64.shl local.get $seed i32.const 0x80000000 i32.and i64.extend_i32_u i64.const 32 i64.shl i64.or i64.const 0x3ff0000000000000 i64.or f64.reinterpret_i64"
    else:
        init = "local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or f32.reinterpret_i32"
    if family == "arithmetic":
        body = f"local.get $x {t}.const 1.0001220703125 {t}.mul {t}.const 0.125 {t}.add {t}.const 1.125 {t}.div local.set $x"
    elif family == "minmax":
        body = f"""local.get $x {t}.const 0.125 {t}.add {t}.const 2 {t}.min {t}.const -2 {t}.max local.set $x
        local.get $x {t}.const 1.5 {t}.gt if local.get $x {t}.neg local.set $x end"""
    else:
        body = f"""local.get $x {t}.ceil local.get $x {t}.floor {t}.add
        local.get $x {t}.trunc {t}.add local.get $x {t}.nearest {t}.add local.set $y
        local.get $x local.get $x {t}.mul {t}.const 0.5 {t}.add {t}.sqrt
        local.get $y {t}.const 8 {t}.add {t}.div {t}.const 0.25 {t}.add local.set $x"""
    result = "local.get $x i32.reinterpret_f32"
    if wide:
        result = "local.get $x i64.reinterpret_f64 local.get $x i64.reinterpret_f64 i64.const 32 i64.shr_u i64.xor i32.wrap_i64"
    return f"""(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x {t}) (local $y {t})
      local.get $iterations i32.const 0 i32.le_s if unreachable end
      {init} local.set $x
      loop $again
        {body}
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end
      {result}))\n"""


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise SystemExit("use pinned wasm-tools 1.258.0")
    directory = ROOT / "float-fixtures"
    directory.mkdir(exist_ok=True)
    cases = []
    for family in FAMILIES:
        for wide in (False, True):
            name = family + ("-f64" if wide else "-f32")
            source, binary = directory / (name+".wat"), directory / (name+".wasm")
            source.write_text(wat(family, wide))
            subprocess.run(["wasm-tools", "parse", str(source), "-o", str(binary)], check=True)
            cases.append({"id": name, "family": family, "wide": wide, "wasm": binary.name,
                          "seed": SEED, "iterations": ITERATIONS, "expectedChecksum": checksum(family, wide, SEED, ITERATIONS),
                          "wasmSha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                          "sourceSha256": hashlib.sha256(source.read_bytes()).hexdigest()})
    (directory / "manifest.json").write_text(json.dumps({"producer": version, "workloads": cases}, indent=2)+"\n")


if __name__ == "__main__":
    main()

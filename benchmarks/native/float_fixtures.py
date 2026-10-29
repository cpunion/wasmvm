"""Owned parameterized scalar FP fixtures; independent Python arithmetic oracle."""
import hashlib
import json
import math
from pathlib import Path
import struct
import subprocess

ROOT = Path(__file__).resolve().parent
SEED, ITERATIONS = 305419896, 1_000_000
FAMILIES = ("arithmetic", "minmax", "rounding", "convert-signed", "convert-unsigned")


def initial(seed, wide):
    bits = 0x3f800000 | (seed & 0x807fffff)
    if wide:
        bits = 0x3ff0000000000000 | ((seed & 0x7fffff) << 29) | ((seed & 0x80000000) << 32)
    return struct.unpack("<d" if wide else "<f", struct.pack("<Q" if wide else "<I", bits))[0]


def checksum(family, wide, seed, iterations):
    def fp(x):
        return x if wide else struct.unpack("<f", struct.pack("<f", x))[0]
    if family.startswith("convert-"):
        return conversion_checksum(wide, family.endswith("unsigned"), seed, iterations)
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


def conversion_checksum(wide, unsigned, seed, iterations):
    width = 64 if wide else 32
    mask, top = (1 << width)-1, 1 << (width-1)
    x = seed * ((1 << 32)+1) if wide else seed
    multiplier = 6364136223846793005 if wide else 1664525
    for index in range(iterations):
        x = (x * multiplier + 1013904223 + index) & mask
        if unsigned:
            x = (x & ((top >> 1)-1)) | ((index & 1) * top)
        else:
            x = (x & (top-1)) - (top >> 1)
        number = float(x)
        if not wide:
            number = struct.unpack("<f", struct.pack("<f", number))[0]
        checked = int(number) & mask
        raw = (checked ^ seed ^ index) & mask
        number = struct.unpack("<d" if wide else "<f", struct.pack("<Q" if wide else "<I", raw))[0]
        if math.isnan(number):
            saturated = 0
        else:
            minimum, maximum = (0, mask) if unsigned else (-top, top-1)
            saturated = minimum if number <= minimum else maximum if number >= maximum else int(number)
        x = checked ^ (saturated & mask)
    return (x ^ (x >> 32)) & 0xffffffff


def conversion_wat(wide, unsigned):
    integer, fp = ("i64", "f64") if wide else ("i32", "f32")
    sign = "u" if unsigned else "s"
    width = 64 if wide else 32
    top = 1 << (width-1)
    seed, index = "local.get $seed", "local.get $i"
    if wide:
        seed += " i64.extend_i32_u"
        index += " i64.extend_i32_u"
    init = seed
    if wide:
        init += " i64.const 4294967297 i64.mul"
    multiplier = 6364136223846793005 if wide else 1664525
    body = f"local.get $x {integer}.const {multiplier} {integer}.mul {integer}.const 1013904223 {integer}.add {index} {integer}.add"
    if unsigned:
        body += f" {integer}.const {(top >> 1)-1} {integer}.and {index} {integer}.const 1 {integer}.and {integer}.const {width-1} {integer}.shl {integer}.or"
    else:
        body += f" {integer}.const {top-1} {integer}.and {integer}.const {top >> 1} {integer}.sub"
    body += f" {fp}.convert_{integer}_{sign} {integer}.trunc_{fp}_{sign} local.set $x"
    body += f" local.get $x local.get $x {seed} {integer}.xor {index} {integer}.xor {fp}.reinterpret_{integer} {integer}.trunc_sat_{fp}_{sign} {integer}.xor local.set $x"
    result = "local.get $x"
    if wide:
        result += " local.get $x i64.const 32 i64.shr_u i64.xor i32.wrap_i64"
    return init, body, result


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
    elif family == "rounding":
        body = f"""local.get $x {t}.ceil local.get $x {t}.floor {t}.add
        local.get $x {t}.trunc {t}.add local.get $x {t}.nearest {t}.add local.set $y
        local.get $x local.get $x {t}.mul {t}.const 0.5 {t}.add {t}.sqrt
        local.get $y {t}.const 8 {t}.add {t}.div {t}.const 0.25 {t}.add local.set $x"""
    result = "local.get $x i32.reinterpret_f32"
    if wide:
        result = "local.get $x i64.reinterpret_f64 local.get $x i64.reinterpret_f64 i64.const 32 i64.shr_u i64.xor i32.wrap_i64"
    x_type = t
    if family.startswith("convert-"):
        init, body, result = conversion_wat(wide, family.endswith("unsigned"))
        x_type = "i64" if wide else "i32"
    return f"""(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x {x_type}) (local $y {t})
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

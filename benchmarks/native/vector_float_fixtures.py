#!/usr/bin/env python3
"""Owned packed-FP workloads with independent scalar, per-lane Python oracles."""
import hashlib
import json
import math
from pathlib import Path
import struct
import subprocess

from float_fixtures import initial

ROOT = Path(__file__).with_name("vector-float-fixtures")
FAMILIES = ("arithmetic", "minmax", "rounding", "convert-signed", "convert-unsigned")
MASK = 0xffffffff


def checksum(family, wide, seed, count):
    def fp(x):
        return x if wide else struct.unpack("<f", struct.pack("<f", x))[0]
    n = 2 if wide else 4
    if family.startswith("convert-"):
        x = [(seed ^ (0x9e3779b9 * i)) & MASK for i in range(4)]
        unsigned = family.endswith("unsigned")
        for index in range(count):
            x = [(v * (1664525 + 2*i) + 1013904223 + i + index) & MASK for i, v in enumerate(x)]
            out = []
            for i, v in enumerate(x[:n]):
                signed = v if unsigned or v < 0x80000000 else v - (1 << 32)
                number = fp(fp(float(signed)) + (0.75 + i*0.125))
                lower, upper = (0, MASK) if unsigned else (-(1 << 31), (1 << 31)-1)
                out.append(min(max(int(number), lower), upper) & MASK)
            x = out + [0]*(4-n)
        return x[0] ^ x[1] ^ x[2] ^ x[3]
    if family == "casts":
        x = [fp(initial(seed, False) + i*0.25) for i in range(4)]
        for _ in range(count):
            x = [fp(fp(v * (0.875-i*0.125) + (0.25+i*0.25)) + (0.125+i*0.125))
                 for i, v in enumerate(x[:2])] + [0.5, 0.75]
    else:
        x = [fp(initial(seed, wide) + i*0.25) for i in range(n)]
        for _ in range(count):
            out = []
            for i, v in enumerate(x):
                if family == "arithmetic":
                    z = fp(fp(fp(v * (1.0001220703125 + i*0.0001220703125)) + (0.125+i*0.125)) / (1.125+i*0.125))
                elif family == "minmax":
                    z = max(min(fp(v + (0.125+i*0.125)), 2.0+i), -2.0-i)
                    z = max(min(z, 1.75+i*0.25), -1.75-i*0.25)
                    if z > 1.5+i*0.25:
                        z = -z
                elif family == "rounding":
                    y = fp(fp(fp(math.ceil(v) + math.floor(v)) + math.trunc(v)) + round(v))
                    z = fp(fp(fp(math.sqrt(fp(fp(v*v) + 0.5))) / fp(y + 8.0+i)) + 0.25+i*0.125)
                    z = abs(z)
                else:
                    raise ValueError(family)
                out.append(z)
            x = out
    raw = struct.pack("<2d" if wide else "<4f", *x)
    lanes = struct.unpack("<4I", raw)
    return lanes[0] ^ lanes[1] ^ lanes[2] ^ lanes[3]


def wat(family, wide):
    t, n = ("f64x2", 2) if wide else ("f32x4", 4)
    def vector(values, typ=None):
        return "v128.const " + (typ or t) + " " + " ".join(map(str, values))
    def const(value):
        return vector([value(i) if callable(value) else value for i in range(n)])
    init = "local.get $seed i32.const 0x807fffff i32.and i32.const 0x3f800000 i32.or i32x4.splat"
    if wide:
        init = "local.get $seed i32.const 0x7fffff i32.and i64.extend_i32_u i64.const 29 i64.shl local.get $seed i32.const 0x80000000 i32.and i64.extend_i32_u i64.const 32 i64.shl i64.or i64.const 0x3ff0000000000000 i64.or i64x2.splat"
    init += f" {const(lambda i: i*0.25)} {t}.add"
    if family == "arithmetic":
        body = f"local.get $x {const(lambda i: 1.0001220703125+i*0.0001220703125)} {t}.mul {const(lambda i: .125+i*.125)} {t}.add {const(lambda i: 1.125+i*.125)} {t}.div local.set $x"
    elif family == "minmax":
        body = f"""local.get $x {const(lambda i: .125+i*.125)} {t}.add
          {const(lambda i: 2+i)} {t}.min {const(lambda i: -2-i)} {t}.max
          {const(lambda i: 1.75+i*.25)} {t}.pmin {const(lambda i: -1.75-i*.25)} {t}.pmax local.set $x
          local.get $x {t}.neg local.get $x local.get $x {const(lambda i: 1.5+i*.25)} {t}.gt v128.bitselect local.set $x"""
    elif family == "rounding":
        body = f"""local.get $x {t}.ceil local.get $x {t}.floor {t}.add
          local.get $x {t}.trunc {t}.add local.get $x {t}.nearest {t}.add local.set $y
          local.get $x local.get $x {t}.mul {const(.5)} {t}.add {t}.sqrt
          local.get $y {const(lambda i: 8+i)} {t}.add {t}.div {const(lambda i: .25+i*.125)} {t}.add {t}.abs local.set $x"""
    elif family.startswith("convert-"):
        sign = "u" if family.endswith("unsigned") else "s"
        init = "local.get $seed i32x4.splat " + vector([(0x9e3779b9*i) & MASK for i in range(4)], "i32x4") + " v128.xor"
        convert = f"f64x2.convert_low_i32x4_{sign}" if wide else f"f32x4.convert_i32x4_{sign}"
        truncate = f"i32x4.trunc_sat_f64x2_{sign}_zero" if wide else f"i32x4.trunc_sat_f32x4_{sign}"
        body = f"""local.get $x {vector([1664525+2*i for i in range(4)], 'i32x4')} i32x4.mul
          {vector([1013904223+i for i in range(4)], 'i32x4')} i32x4.add
          local.get $i i32x4.splat i32x4.add {convert}
          {const(lambda i: .75+i*.125)} {t}.add {truncate} local.set $x"""
    elif family == "casts" and not wide:
        body = """local.get $x f64x2.promote_low_f32x4
          v128.const f64x2 0.875 0.75 f64x2.mul v128.const f64x2 0.25 0.5 f64x2.add
          f32x4.demote_f64x2_zero v128.const f32x4 0.125 0.25 0.5 0.75 f32x4.add local.set $x"""
    else:
        raise ValueError(family)
    result = "local.get $x i32x4.extract_lane 0 " + " ".join(f"local.get $x i32x4.extract_lane {i} i32.xor" for i in (1, 2, 3))
    return f"""(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128) (local $y v128)
      local.get $iterations i32.eqz if unreachable end
      {init} local.set $x
      loop $again
        {body}
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end {result}))\n"""


def main():
    version = subprocess.check_output(["wasm-tools", "--version"], text=True).strip()
    if version != "wasm-tools 1.258.0":
        raise ValueError(version)
    ROOT.mkdir(exist_ok=True)
    rows = []
    for family, wide in [(f, w) for f in FAMILIES for w in (False, True)] + [("casts", False)]:
        name = family + ("-f64x2" if wide else "-f32x4")
        source, binary = ROOT / (name + ".wat"), ROOT / (name + ".wasm")
        source.write_text(wat(family, wide))
        subprocess.run(["wasm-tools", "parse", str(source), "-o", str(binary)], check=True)
        subprocess.run(["wasm-tools", "validate", str(binary)], check=True)
        seed, count = 0x12345678, 10000
        rows.append({"id": name, "family": family, "wide": wide, "seed": seed, "iterations": count,
                     "expectedChecksum": checksum(family, wide, seed, count), "unit": "ns/source iteration",
                     "wat": source.name, "wasm": binary.name,
                     "watSha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                     "wasmSha256": hashlib.sha256(binary.read_bytes()).hexdigest()})
    (ROOT / "manifest.json").write_text(json.dumps({"producer": version,
        "origin": "authored in wasmvm; no third-party benchmark code", "workloads": rows}, indent=2) + "\n")


if __name__ == "__main__":
    main()

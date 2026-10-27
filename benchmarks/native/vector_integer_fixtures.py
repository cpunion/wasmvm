#!/usr/bin/env python3
"""Complete packed-integer families with independent scalar lane oracles."""
import hashlib
import json
from pathlib import Path
import struct
import subprocess

ROOT = Path(__file__).with_name('vector-integer-fixtures')
MASK = 0xffffffff


def pack(values, bits):
    return b''.join((v & ((1 << bits)-1)).to_bytes(bits//8, 'little') for v in values)


def lanes(raw, bits, signed=False):
    return [int.from_bytes(raw[i:i+bits//8], 'little', signed=signed)
            for i in range(0, 16, bits//8)]


def saturate(v, bits, unsigned=False):
    low, high = (0, (1 << bits)-1) if unsigned else (-(1 << (bits-1)), (1 << (bits-1))-1)
    return max(low, min(high, v))


def binary(op, raw, rhs):
    if op == 'shuffle':
        return bytes((raw+rhs)[i] for i in SHUFFLE)
    if op == 'swizzle':
        indices = [(a ^ b) & 31 for a, b in zip(raw, rhs)]
        return bytes(raw[i] if i < 16 else 0 for i in indices)
    if op == 'i8x16.popcnt':
        return bytes(v.bit_count() for v in raw)
    dest, kind = op.split('.', 1)
    bits = int(dest.split('x')[0][1:])
    if kind.startswith('extend_'):
        unsigned = kind.endswith('_u')
        source = lanes(raw, bits//2, not unsigned)
        n = 128//bits
        start = n if '_high_' in kind else 0
        return pack(source[start:start+n], bits)
    if kind.startswith('extmul_'):
        unsigned = kind.endswith('_u')
        a, b = lanes(raw, bits//2, not unsigned), lanes(rhs, bits//2, not unsigned)
        n = 128//bits
        start = n if '_high_' in kind else 0
        return pack([a[i]*b[i] for i in range(start, start+n)], bits)
    if kind.startswith('extadd_pairwise_'):
        a = lanes(raw, bits//2, kind.endswith('_s'))
        return pack([a[i]+a[i+1] for i in range(0, len(a), 2)], bits)
    if kind.startswith('narrow_'):
        unsigned = kind.endswith('_u')
        return pack([saturate(v, bits, unsigned) for v in lanes(raw, bits*2, True)+lanes(rhs, bits*2, True)], bits)
    if kind == 'avgr_u':
        return pack([(a+b+1)//2 for a, b in zip(lanes(raw, bits), lanes(rhs, bits))], bits)
    if kind == 'q15mulr_sat_s':
        return pack([saturate((a*b+16384)//32768, 16) for a, b in zip(lanes(raw, 16, True), lanes(rhs, 16, True))], 16)
    if kind == 'dot_i16x8_s':
        a, b = lanes(raw, 16, True), lanes(rhs, 16, True)
        return pack([a[i]*b[i]+a[i+1]*b[i+1] for i in range(0, 8, 2)], 32)
    raise ValueError(op)


SHUFFLE = (31, 0, 17, 2, 19, 4, 21, 6, 23, 8, 25, 10, 27, 12, 29, 14)
RHS = bytes((i*73+131) & 255 for i in range(16))
SWIZZLE = bytes((15, 0, 1, 14, 16, 255, 8, 7, 31, 128, 2, 13, 4, 11, 6, 9))
FAMILIES = {
    'shuffle': [('shuffle', RHS)],
    'swizzle': [('swizzle', SWIZZLE)],
    'popcount': [('i8x16.popcnt', None)],
    'narrow': [(f'i{bits}x{128//bits}.narrow_i{bits*2}x{64//bits}_{sign}', RHS)
               for bits in (8, 16) for sign in ('s', 'u')],
    'average': [('i8x16.avgr_u', RHS), ('i16x8.avgr_u', RHS)],
    'q15': [('i16x8.q15mulr_sat_s', pack([-32768, 32767, -1, 1, -16384, 16384, 32767, -32768], 16))],
    'dot': [('i32x4.dot_i16x8_s', RHS)],
    'pairwise': [(f'i{bits}x{128//bits}.extadd_pairwise_i{bits//2}x{256//bits}_{sign}', None)
                 for bits in (16, 32) for sign in ('s', 'u')],
}
for category in ('extend', 'extmul'):
    for bits in (16, 32, 64):
        FAMILIES[f'{category}-i{bits}'] = [(f'i{bits}x{128//bits}.{category}_{half}_i{bits//2}x{256//bits}_{sign}', RHS if category == 'extmul' else None)
                                          for half in ('low', 'high') for sign in ('s', 'u')]


def checksum(family, seed, count):
    x = pack([seed]*4, 32)
    for i in range(count):
        # Mix every iteration before the measured capability to avoid a fixed
        # point hiding input data dependencies or unused output lanes.
        x = pack([((v + a + i) & MASK) for v, a in zip(lanes(x, 32), (1, 3, 5, 7))], 32)
        for op, rhs in FAMILIES[family]:
            x = binary(op, x, rhs)
    a, b, c, d = lanes(x, 32)
    return a ^ b ^ c ^ d


def vector(raw):
    return 'v128.const i8x16 ' + ' '.join(map(str, raw))


def wat(family):
    ops = []
    for op, rhs in FAMILIES[family]:
        if op == 'swizzle':
            ops.append('local.tee $x local.get $x')
        if rhs is not None:
            ops.append(vector(rhs))
        if op == 'shuffle':
            ops.append('i8x16.shuffle ' + ' '.join(map(str, SHUFFLE)))
        elif op == 'swizzle':
            ops.append('v128.xor ' + vector(bytes([31]*16)) + ' v128.and i8x16.swizzle')
        else:
            ops.append(op)
    result = 'local.get $x i32x4.extract_lane 0 ' + ' '.join(f'local.get $x i32x4.extract_lane {i} i32.xor' for i in (1, 2, 3))
    return f'''(module (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
      (local $i i32) (local $x v128)
      local.get $iterations i32.eqz if unreachable end
      local.get $seed i32x4.splat local.set $x
      loop $again
        local.get $x v128.const i32x4 1 3 5 7 i32x4.add
        local.get $i i32x4.splat i32x4.add
        {' '.join(ops)} local.set $x
        local.get $i i32.const 1 i32.add local.tee $i local.get $iterations i32.lt_u br_if $again
      end {result}))\n'''


def main():
    version = subprocess.check_output(['wasm-tools', '--version'], text=True).strip()
    if version != 'wasm-tools 1.258.0':
        raise ValueError(version)
    ROOT.mkdir(exist_ok=True)
    rows = []
    for family in FAMILIES:
        source, binary_path = ROOT/(family+'.wat'), ROOT/(family+'.wasm')
        source.write_text(wat(family))
        subprocess.run(['wasm-tools', 'parse', str(source), '-o', str(binary_path)], check=True)
        subprocess.run(['wasm-tools', 'validate', str(binary_path)], check=True)
        seed, count = 0x12345678, 10000
        rows.append({'id': family, 'seed': seed, 'iterations': count,
                     'expectedChecksum': checksum(family, seed, count), 'unit': 'ns/source iteration',
                     'wat': source.name, 'wasm': binary_path.name,
                     'watSha256': hashlib.sha256(source.read_bytes()).hexdigest(),
                     'wasmSha256': hashlib.sha256(binary_path.read_bytes()).hexdigest()})
    (ROOT/'manifest.json').write_text(json.dumps({'producer': version,
        'origin': 'authored in wasmvm; no third-party benchmark code', 'workloads': rows}, indent=2)+'\n')


if __name__ == '__main__':
    main()

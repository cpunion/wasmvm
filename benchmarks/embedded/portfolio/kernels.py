"""Authored Core 2.0 diagnostic kernels and independent arithmetic oracles."""
import math
import struct

MASK = 0xffffffff
CASES = (
    ('float-convert', 25000, 'numeric'),
    ('control-multivalue', 25000, 'control'),
    ('indirect-reference', 25000, 'calls-references'),
    ('simd-shuffle', 25000, 'simd'),
    ('bulk-unaligned', 1000, 'memory'),
    ('deep-calls', 2000, 'calls-references'),
    ('image-filter-c', 40, 'applications'),
    ('packet-crc-rust', 64, 'applications'),
)


def rot(x, n):
    return ((x << n) | (x >> (32 - n))) & MASK


def f32(x):
    return struct.unpack('<f', struct.pack('<f', x))[0]


def checksum(kind, seed, count):
    if kind not in {x[0] for x in CASES} or type(seed) is not int or not 0 <= seed <= MASK or type(count) is not int or not 1 <= count < 2**31:
        raise ValueError('invalid portfolio input')
    x = seed
    if kind == 'float-convert':
        for i in range(count):
            f = struct.unpack('<f', struct.pack('<I', (x & 0x7fffff) | 0x3f800000))[0]
            f = f32(f32(f * 1.5) + 0.25)
            bits = struct.unpack('<I', struct.pack('<f', f))[0]
            x = bits ^ int(math.sqrt((i & 255) + 1) * 4096) ^ i
    elif kind == 'control-multivalue':
        for i in range(count):
            k = x & 3
            v = (x + i) & MASK if k == 0 else rot(x, 5) if k == 1 else ((x ^ i) + 1013904223) & MASK if k == 2 else (x * 1664525) & MASK
            t = (v + i) & MASK
            x = t ^ rot(t, 3)
    elif kind == 'indirect-reference':
        for i in range(count):
            x = (x + i) & MASK if x & 1 == 0 else rot(x ^ i, 7)
    elif kind == 'simd-shuffle':
        v = [seed] * 4
        for i in range(count):
            v = [((a + i) & MASK) ^ b for a, b in zip(v, (1, 0x12345678, 0x9e3779b9, MASK))]
            v = [v[3], v[0], v[1], v[2]]
            v = [rot(a, 3) for a in v]
        x = ((v[0] + rot(v[1], 7)) & MASK) ^ v[2] ^ rot(v[3], 13)
    elif kind == 'bulk-unaligned':
        memory = bytearray(512)
        for i in range(count):
            memory[:] = bytes(512)
            memory[1:258] = bytes([(seed + i) & 255]) * 257
            memory[3:260] = memory[1:258]
            v = int.from_bytes(memory[253:261], 'little')
            x = rot(x, 3) ^ (v & MASK) ^ (v >> 32) ^ i
    elif kind == 'deep-calls':
        for i in range(count):
            # Thirty-two recursive edges (33 helper frames), with arithmetic on return.
            v = (x + i + sum(range(1, 33))) & MASK
            for depth in range(1, 33):
                v = rot(v, 1) ^ depth
            x = v
    elif kind == 'image-filter-c':
        image = [0] * 1024
        for i in range(1024):
            x = (x * 1664525 + 1013904223) & MASK
            image[i] = x >> 24
        x = seed
        for iteration in range(count):
            for y in range(1, 31):
                for col in range(1, 31):
                    p = y * 32 + col
                    v = (image[p-33] + 2*image[p-32] + image[p-31] + 2*image[p-1] + 4*image[p] + 2*image[p+1] + image[p+31] + 2*image[p+32] + image[p+33]) >> 4
                    x = (rot(x, 5) ^ (v + iteration + p)) & MASK
    else:
        for i in range(count):
            crc = MASK
            state = (seed + i) & MASK
            for _ in range(256):
                state = (state * 1664525 + 1013904223) & MASK
                crc ^= state >> 24
                for _ in range(8):
                    crc = (crc >> 1) ^ (0xedb88320 if crc & 1 else 0)
            x = rot(x, 3) ^ (crc ^ MASK)
    return x & MASK


DECLARATIONS = {
    'indirect-reference': '''
  (type $binary (func (param i32 i32) (result i32)))
  (table 2 2 funcref)
  (elem (i32.const 0) $add $mix)
  (global $state (mut i32) (i32.const 0))
  (func $add (type $binary) local.get 0 local.get 1 i32.add)
  (func $mix (type $binary) local.get 0 local.get 1 i32.xor i32.const 7 i32.rotl)
''',
    'simd-shuffle': '(memory 1 1)',
    'bulk-unaligned': '(memory 1 1)',
    'deep-calls': '''
  (func $recurse (param $x i32) (param $n i32) (result i32)
    local.get $n i32.eqz
    if (result i32) local.get $x
    else
      local.get $x local.get $n i32.add
      local.get $n i32.const 1 i32.sub call $recurse
      i32.const 1 i32.rotl local.get $n i32.xor
    end)
''',
}
BODIES = {
    'float-convert': '''
      local.get $x i32.const 8388607 i32.and i32.const 1065353216 i32.or
      f32.reinterpret_i32 f32.const 1.5 f32.mul f32.const 0.25 f32.add i32.reinterpret_f32
      local.get $i i32.const 255 i32.and i32.const 1 i32.add
      f64.convert_i32_u f64.sqrt f64.const 4096 f64.mul i32.trunc_sat_f64_u
      i32.xor local.get $i i32.xor local.set $x
''',
    'control-multivalue': '''
      block $done (result i32)
        block $three
          block $two
            block $one
              block $zero
                local.get $x i32.const 3 i32.and br_table $zero $one $two $three
              end
              local.get $x local.get $i i32.add br $done
            end
            local.get $x i32.const 5 i32.rotl br $done
          end
          local.get $x local.get $i i32.xor i32.const 1013904223 i32.add br $done
        end
        local.get $x i32.const 1664525 i32.mul
      end
      local.get $i
      block (param i32 i32) (result i32 i32)
        i32.add local.tee $tmp local.get $tmp i32.const 3 i32.rotl
      end
      i32.xor local.set $x
''',
    'indirect-reference': '''
      global.get $state i32.const 1 i32.and local.tee $tmp
      table.get 0 ref.is_null if unreachable end
      local.get $tmp local.get $tmp table.get 0 table.set 0
      global.get $state local.get $i local.get $tmp call_indirect (type $binary)
      global.set $state
''',
    'simd-shuffle': '''
      i32.const 1
      local.get $v local.get $i i32x4.splat i32x4.add
      v128.const i32x4 1 305419896 2654435769 -1 v128.xor
      local.tee $v local.get $v
      i8x16.shuffle 12 13 14 15 0 1 2 3 4 5 6 7 8 9 10 11
      local.tee $v i32.const 3 i32x4.shl
      local.get $v i32.const 29 i32x4.shr_u v128.or v128.store align=1
      i32.const 1 v128.load align=1 local.set $v
''',
    'bulk-unaligned': '''
      i32.const 0 i32.const 0 i32.const 512 memory.fill
      i32.const 1 local.get $seed local.get $i i32.add i32.const 257 memory.fill
      i32.const 3 i32.const 1 i32.const 257 memory.copy
      i32.const 253 i64.load align=1 local.tee $wide i32.wrap_i64
      local.get $wide i64.const 32 i64.shr_u i32.wrap_i64 i32.xor
      local.get $x i32.const 3 i32.rotl i32.xor local.get $i i32.xor local.set $x
''',
    'deep-calls': 'local.get $x local.get $i i32.add i32.const 32 call $recurse local.set $x',
}


def module(kind):
    if kind not in BODIES:
        raise ValueError('compiled application has a separate producer')
    before = 'local.get $seed local.set $x'
    after = 'local.get $x'
    if kind == 'indirect-reference':
        before, after = 'local.get $seed global.set $state', 'global.get $state'
    if kind == 'simd-shuffle':
        before = 'local.get $seed i32x4.splat local.set $v'
        after = '''local.get $v i32x4.extract_lane 0 local.get $v i32x4.extract_lane 1 i32.const 7 i32.rotl i32.add
        local.get $v i32x4.extract_lane 2 i32.xor local.get $v i32x4.extract_lane 3 i32.const 13 i32.rotl i32.xor'''
    return f'''(module
  {DECLARATIONS.get(kind, '')}
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    {before}
    loop $again
      {BODIES[kind]}
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    {after}))
'''

#!/usr/bin/env python3
"""Generate owned host-callback and preparation-shape diagnostic fixtures."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent
SEED = 0x12345678


def checksum(seed, count):
    for i in range(count):
        x = (seed + i) & 0xffffffff
        seed = ((x << 7) | (x >> 25)) & 0xffffffff
        seed ^= 0x9e3779b9
    return seed


def main():
    assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
    host = '''(module
  (import "host" "mix" (func $mix (param i32 i32) (result i32)))
  (func (export "run") (param $seed i32) (param $count i32) (result i32)
    (local $i i32)
    local.get $count i32.const 0 i32.le_s if unreachable end
    loop $again
      local.get $seed local.get $i call $mix local.set $seed
      local.get $i i32.const 1 i32.add local.tee $i local.get $count i32.lt_u br_if $again
    end
    local.get $seed))
'''
    cases = [
        ('small-module', '(module (func))', 'valid', 1, 0),
        ('many-functions', '(module ' + '(func) '*512 + ')', 'valid', 512, 0),
        ('large-functions', '(module ' + ('(func (param i32) (result i32) local.get 0 ' + 'i32.const 1 i32.add '*4096 + ')')*2 + ')', 'valid', 2, 0),
        ('deep-control', '(module (func '+ 'block '*256 + 'nop ' + 'end '*256 + '))', 'valid', 1, 0),
        ('many-locals', '(module (func (local ' + 'i32 '*16384 + ')))', 'valid', 1, 0),
        ('many-imports', '(module ' + ' '.join(f'(import "host" "f{i}" (func (param i32) (result i32)))' for i in range(128)) + ')', 'valid', 0, 128),
        ('tables-globals', '(module ' + '(table 64 64 funcref) '*8 + '(global (mut i32) (i32.const 7)) '*128 + ')', 'valid', 0, 0),
        ('data-64k', '(module (memory 2 2) (data (i32.const 0) "' + '\\ab'*65536 + '"))', 'valid', 0, 0),
        ('invalid-stack', '(module (func (result i32) i64.const 0))', 'invalid', 0, 0),
        ('excessive-locals', '(module (func (local '+ 'i32 '*1048577 + ')))', 'limit', 0, 0),
    ]
    entries = []
    for name, source, expected, functions, imports in cases + [('host-callback', host, 'valid', 1, 1)]:
        # The large-local source is generated deterministically but need not be
        # committed as four megabytes of repeated text. Retain its hash instead.
        wat = ROOT / (name + '.wat')
        wasm = ROOT / (name + '.wasm')
        wat.write_text(source+'\n')
        subprocess.run(['wasm-tools','parse',str(wat),'-o',str(wasm)],check=True)
        source_hash = hashlib.sha256(wat.read_bytes()).hexdigest()
        if name == 'excessive-locals': wat.unlink()
        entries.append({'id':name, 'wasm':'subsystems/'+wasm.name, 'wasmSha256':hashlib.sha256(wasm.read_bytes()).hexdigest(),
                        'sourceSha256':source_hash, 'inputBytes':wasm.stat().st_size,
                        'outcome':expected, 'definedFunctions':functions, 'imports':imports})
    bad = ROOT/'malformed.wasm'
    bad.write_bytes(b'\x00asm\x01\x00\x00\x00\x01\x80')
    entries.insert(-1, {'id':'malformed', 'wasm':'subsystems/malformed.wasm', 'wasmSha256':hashlib.sha256(bad.read_bytes()).hexdigest(), 'sourceSha256':None, 'inputBytes':10, 'outcome':'malformed','definedFunctions':0,'imports':0})
    host_info = entries.pop()
    hosts = []
    for name, count in [('host-control',1), ('host-loop',1000)]:
        hosts.append(dict(host_info, id=name, seed=SEED, iterations=count, expectedChecksum=checksum(SEED,count),
                          unit='ns/callback iteration', validationCases=[{'seed':s,'iterations':n,'expectedChecksum':checksum(s,n)} for s,n in ((0,1),(1,2),(0xffffffff,17))]))
    manifest = {'schemaVersion':1,'version':'subsystems-v1','producer':'wasm-tools 1.258.0','origin':'authored in wasmvm',
                'generatorSha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'host':hosts,'preparation':entries}
    (ROOT/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')


if __name__ == '__main__': main()

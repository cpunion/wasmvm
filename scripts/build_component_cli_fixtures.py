#!/usr/bin/env python3
"""Build command lifecycle fixtures with the pinned component producer."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT/'internal/cli/testdata'
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
fixtures = {}
for preview, version in [(2, '0.2.12'), (3, '0.3.2')]:
    async_type = 'async' if preview == 3 else ''
    fixtures[f'component-init-p{preview}'] = f'''(component
 (core module $m
  (func $start nop nop nop) (start $start)
  (func (export "run") (result i32) nop i32.const 0))
 (core instance $m (instantiate $m))
 (func $run {async_type} (result (result)) (canon lift (core func $m "run")))
 (instance $run (export "run" (func $run)))
 (export "wasi:cli/run@{version}" (instance $run)))'''

source = ROOT/'pkg/wasi/preview3/testdata/stdin.wat'
fixtures['component-stdin-p3'] = source.read_text().rstrip()[:-1]+'''
 (core module $command
  (import "" "read" (func $read (param i32) (result i32)))
  (func (export "run") (result i32) i32.const 0 call $read drop i32.const 0))
 (core instance $command (instantiate $command (with "" (instance (export "read" (func $m "run"))))))
 (func $command async (result (result)) (canon lift (core func $command "run")))
 (instance $command (export "run" (func $command)))
 (export "wasi:cli/run@0.3.2" (instance $command)))'''

sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
manifest = {'producer': 'wasm-tools 1.258.0', 'generatorSHA256': sha(Path(__file__)),
            'sourcesSHA256': {str(source.relative_to(ROOT)): sha(source)}, 'fixtures': {}}
for name, text in fixtures.items():
    path = OUT/(name+'.wat')
    path.write_text(text.rstrip()+'\n')
    subprocess.run(['wasm-tools', 'parse', path, '-o', path.with_suffix('.wasm')], check=True)
    subprocess.run(['wasm-tools', 'validate', '--features', 'all', path.with_suffix('.wasm')], check=True)
    manifest['fixtures'][name] = {ext: sha(OUT/(name+'.'+ext)) for ext in ('wat', 'wasm')}
(OUT/'components.json').write_text(json.dumps(manifest, indent=2)+'\n')
print('Built', len(fixtures), 'validated component command fixtures')

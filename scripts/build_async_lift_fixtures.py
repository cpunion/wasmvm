#!/usr/bin/env python3
"""Build pinned run-to-completion async lift and task context fixtures."""
import hashlib
import json
import re
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
base = (OUT / 'async-lift.wat').read_text()
wide = base.replace('context.get i32', 'context.get i64').replace('context.set i32', 'context.set i64')
wide = wide.replace('(func $get (result i32))', '(func $get (result i64))').replace('(func $set (param i32))', '(func $set (param i64))')
wide = wide.replace('call $get', 'call $get i32.wrap_i64').replace('call $set', 'i64.extend_i32_u call $set')
wide = wide.replace('(core instance $env', '(core func $get1 (canon context.get i64 1))\n (core func $set1 (canon context.set i64 1))\n (core instance $env', 1)
wide = wide.replace('(export "get" (func $get))', '(export "get1" (func $get1)) (export "set1" (func $set1)) (export "get" (func $get))', 1)
wide = wide.replace('(func $check', '(import "env" "get1" (func $get1 (result i64)))\n (import "env" "set1" (func $set1 (param i64)))\n (func (export "wide-context") (param i64) (result i64) call $get i64.eqz if else unreachable end call $get1 i64.eqz if else unreachable end local.get 0 call $set local.get 0 i64.const -1 i64.xor call $set1 call $get1 local.get 0 i64.const -1 i64.xor i64.ne if unreachable end call $get)\n (func $check', 1)
wide = wide.rstrip()[:-1] + '\n (func (export "wide-context") (param "bits" u64) (result u64) (canon lift (core func $code "wide-context"))))\n'
fixtures = {'async-lift': base, 'async-context64': wide, 'async-host': (OUT / 'async-host.wat').read_text()}
reentry = base.replace('(core func $return', '(import "hook" (func $hook))\n (core func $hook (canon lower (func $hook)))\n (core func $return', 1)
reentry = reentry.replace('(core instance $env ', '(core instance $env (export "hook" (func $hook)) ', 1)
reentry = reentry.replace('(core module $code', '(core module $code (import "env" "hook" (func $hook))', 1)
reentry = reentry.replace('call $yield drop call $get', 'call $hook call $yield drop call $get')
fixtures['async-reentry'] = reentry
fixtures['async-init'] = base.replace('(func $check', '(func $initialize call $get if unreachable end i32.const 123 call $set call $yield drop)\n (start $initialize)\n (func $check', 1)
fixtures['async-tls'] = (OUT / 'async-tls.wat').read_text()
fixtures['async-cli'] = '''(component
 (core module $code (func (export "run") (result i32) i32.const 0))
 (core instance $code (instantiate $code))
 (func $run async (result (result)) (canon lift (core func $code "run")))
 (instance $cli (export "run" (func $run)))
 (export "wasi:cli/run@0.2.12" (instance $cli)))
'''
fixtures['async-http'] = '''(component
 (import "wasi:http/types@0.2.12" (instance $types
  (export "incoming-request" (type (sub resource)))
  (export "response-outparam" (type (sub resource)))))
 (alias export $types "incoming-request" (type $request))
 (alias export $types "response-outparam" (type $response))
 (core module $code (func (export "handle") (param i32 i32)))
 (core instance $code (instantiate $code))
 (func $handle async (param "request" (own $request)) (param "response-out" (own $response)) (canon lift (core func $code "handle")))
 (instance $handler (export "handle" (func $handle)))
 (export "wasi:http/incoming-handler@0.2.12" (instance $handler)))
'''
fixtures['async-alias'] = '''(component
 (import "a" (type $a (sub resource)))
 (import "b" (type $b (sub resource)))
 (import "create" (func $create (result (own $b))))
 (core func $create (canon lower (func $create)))
 (core func $return (canon task.return (result (own $b))))
 (core instance $env (export "create" (func $create)) (export "return" (func $return)))
 (core module $code
  (import "env" "create" (func $create (result i32)))
  (import "env" "return" (func $return (param i32)))
  (func (export "make") call $create call $return))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "make") async (result (own $a)) (canon lift (core func $code "make") async)))
'''
base = (OUT / 'async-values.wat').read_text()
fixtures['async-values'] = base
for encoding in ['utf16', 'latin1+utf16']:
    # The explicit bad-encoding declaration stays opposite to the lift.
    lines = []
    for line in base.splitlines():
        if '(canon ' in line:
            if 'string-encoding=' in line:
                if encoding == 'utf16':
                    line = line.replace('string-encoding=utf16', 'string-encoding=utf8')
            else:
                line = re.sub(r'\(memory \$(mem|alias|other)\)', lambda m: m[0] + ' string-encoding=' + encoding, line)
        lines.append(line)
    text = '\n'.join(lines) + '\n'
    fixtures['async-values-' + ('utf16' if encoding == 'utf16' else 'compact')] = text
text = base.replace('(memory (export "memory") 1)', '(memory (export "memory") i64 1)').replace('(memory (export "other") 1)', '(memory (export "other") i64 1)').replace('(memory 1)', '(memory i64 1)')
text = text.replace('(data (i32.const 0)', '(data (i64.const 0)').replace('(global $bump (mut i32) (i32.const 1024))', '(global $bump (mut i64) (i64.const 1024))')
text = text.replace('(param $old i32) (param $old-size i32) (param $alignment i32) (param $size i32) (result i32) (local $p i32)', '(param $old i64) (param $old-size i64) (param $alignment i64) (param $size i64) (result i64) (local $p i64)')
text = text.replace('local.get $size i32.add i32.const 7 i32.add i32.const -8 i32.and', 'local.get $size i64.add i64.const 7 i64.add i64.const -8 i64.and')
text = text.replace('local.get $old-size local.get $size i32.lt_u', 'local.get $old-size local.get $size i64.lt_u')
for name in ['ret', 'ret-other', 'ret-encoding']:
    text = text.replace(f'(func ${name} (param i32 i32))', f'(func ${name} (param i64 i64))')
text = text.replace('(func $ret-wide (param i32))', '(func $ret-wide (param i64))').replace('(export "echo") (param i32 i32)', '(export "echo") (param i64 i64)').replace('(export "spill") (param i32)', '(export "spill") (param i64)')
text = text.replace('local.get $n i32.const 4 i32.mul i32.const 64 i32.add', 'local.get $n i64.extend_i32_u i64.const 4 i64.mul i64.const 64 i64.add')
text = text.replace('i32.const 64 call $ret-wide', 'i64.const 64 call $ret-wide').replace('i32.const 65535 i32.const 13', 'i64.const 4294967296 i64.const 13').replace('i32.const 0 i32.const 13', 'i64.const 0 i64.const 13')
fixtures['async-values64'] = text
# Reuse the ordinary resource fixture and convert only its creation ABI.
text = (OUT / 'ownership.wat').read_text()
text = text.replace('(core module $guest', '(core func $return (canon task.return (result (own 1))))\n  (core module $guest')
text = text.replace('(import "env" "new"', '(import "env" "return" (func $return (param i32)))\n    (import "env" "new"')
text = text.replace('(func (export "make") (param i32) (result i32) local.get 0 call $new)', '(func (export "make") (param i32) local.get 0 call $new call $return)\n    (func (export "bad") (param i32) local.get 0 call $new call $return unreachable)')
text = text.replace('(core instance $env ', '(core instance $env (export "return" (func $return)) ')
text = text.replace('(export "make") (param "rep" u32)', '(export "make") async (param "rep" u32)').replace('(core func $make)))', '(core func $make) async))')
text = text[:-2] + '\n (func (export "bad") async (param "rep" u32) (result (own 1)) (canon lift (core func $guest "bad") async)))\n'
fixtures['async-resources'] = text
pins = {}
for name, text in fixtures.items():
    path = OUT / (name + '.wat')
    path.write_text(text)
    output = path.with_suffix('.wasm')
    subprocess.run(['wasm-tools', 'parse', str(path), '-o', str(output)], check=True)
    subprocess.run(['wasm-tools', 'validate', '--features', 'all', str(output)], check=True)
    pins[name] = {extension: hashlib.sha256((OUT / (name + '.' + extension)).read_bytes()).hexdigest() for extension in ['wat', 'wasm']}
(OUT / 'async-lift.json').write_text(json.dumps({'componentModelCommit': 'd095e2f60ddae1f9a1e9b1e2fb547ec1114da677', 'producer': 'wasm-tools 1.258.0', 'generatorSHA256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'resourceSourceSHA256': hashlib.sha256((OUT / 'ownership.wat').read_bytes()).hexdigest(), 'sha256': pins}, indent=2) + '\n')
print('Built', len(pins), 'independently validated async lift fixtures')

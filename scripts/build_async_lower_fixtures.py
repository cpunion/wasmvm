#!/usr/bin/env python3
"""Build pinned eager async-lower ABI, composition and ownership fixtures."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
base = (OUT/'async-lower.wat').read_text()
fixtures = {'async-lower': base}
fixtures['async-lower64'] = base.replace('(memory (export "memory") 1)', '(memory (export "memory") i64 1)').replace('(memory 1)', '(memory i64 1)').replace('(func $add (param i32 i32)', '(func $add (param i32 i64)').replace('i32.const 32', 'i64.const 32')
sync = base.replace('(export "add") (param i32) local.get 0 i32.const 7 i32.add call $return', '(export "add") (param i32) (result i32) local.get 0 i32.const 7 i32.add').replace('(core func $code "add") async)', '(core func $code "add"))')
fixtures['async-lower-sync-callee'] = sync

values = (OUT/'async-values.wat').read_text()
memory = values[values.index(' (core module $memory'):values.index(' (core func $ret ')]

arity = '''(component
 (import "a4" (func $a4 async (param "a" u32) (param "b" u32) (param "c" u32) (param "d" u32) (result u32)))
 (import "a5" (func $a5 async (param "a" u32) (param "b" u32) (param "c" u32) (param "d" u32) (param "e" u32) (result u32)))
 (import "mixed" (func $mixed async (param "a" u32) (param "b" u64) (param "c" float32) (param "d" float64) (result u64)))
 (import "unit" (func $unit async))
 MEMORY
 (core func $a4 (canon lower (func $a4) (memory $mem) async))
 (core func $a5 (canon lower (func $a5) (memory $mem) async))
 (core func $mixed (canon lower (func $mixed) (memory $mem) async))
 (core func $unit (canon lower (func $unit) async))
 (core instance $env (export "memory" (memory $mem)) (export "a4" (func $a4)) (export "a5" (func $a5)) (export "mixed" (func $mixed)) (export "unit" (func $unit)))
 (core module $code
  (import "env" "memory" (memory 1))
  (import "env" "a4" (func $a4 (param i32 i32 i32 i32 i32) (result i32)))
  (import "env" "a5" (func $a5 (param i32 i32) (result i32)))
  (import "env" "mixed" (func $mixed (param i32 i64 f32 f64 i32) (result i32)))
  (import "env" "unit" (func $unit (result i32)))
  (func (export "a4") (param i32 i32 i32 i32) (result i32)
   local.get 0 local.get 1 local.get 2 local.get 3 i32.const 32 call $a4 i32.const 2 i32.ne if unreachable end i32.const 32 i32.load)
  (func (export "a5") (param i32 i32 i32 i32 i32) (result i32)
   i32.const 64 local.get 0 i32.store i32.const 68 local.get 1 i32.store i32.const 72 local.get 2 i32.store i32.const 76 local.get 3 i32.store i32.const 80 local.get 4 i32.store
   i32.const 64 i32.const 32 call $a5 i32.const 2 i32.ne if unreachable end i32.const 32 i32.load)
  (func (export "mixed") (param i32 i64 f32 f64) (result i64)
   local.get 0 local.get 1 local.get 2 local.get 3 i32.const 32 call $mixed i32.const 2 i32.ne if unreachable end i32.const 32 i64.load)
  (func (export "unit") (result i32) call $unit)
  (func (export "bad-result") (param i32) (result i32) i32.const 1 i32.const 2 i32.const 3 i32.const 4 local.get 0 call $a4)
  (func (export "bad-args") (param i32) (result i32) local.get 0 i32.const 32 call $a5))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "a4") (param "a" u32) (param "b" u32) (param "c" u32) (param "d" u32) (result u32) (canon lift (core func $code "a4")))
 (func (export "a5") (param "a" u32) (param "b" u32) (param "c" u32) (param "d" u32) (param "e" u32) (result u32) (canon lift (core func $code "a5")))
 (func (export "mixed") (param "a" u32) (param "b" u64) (param "c" float32) (param "d" float64) (result u64) (canon lift (core func $code "mixed")))
 (func (export "unit") (result u32) (canon lift (core func $code "unit")))
 (func (export "bad-result") (param "ptr" u32) (result u32) (canon lift (core func $code "bad-result")))
 (func (export "bad-args") (param "ptr" u32) (result u32) (canon lift (core func $code "bad-args"))))
'''.replace('MEMORY', memory)
fixtures['async-lower-arity'] = arity

strings = '''(component
 (component $server
  MEMORY
  (core func $ret (canon task.return (result string) (memory $mem)))
  (core instance $env (export "ret" (func $ret)))
  (core module $code (import "env" "ret" (func $ret (param i32 i32))) (func (export "echo") (param i32 i32) local.get 0 local.get 1 call $ret))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") async (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) (realloc $realloc) async)))
 (instance $server (instantiate $server))
 (alias export $server "echo" (func $echo))
 (component $client
  (import "echo" (func $echo async (param "text" string) (result string)))
  MEMORY
  (core func $echo (canon lower (func $echo) (memory $mem) (realloc $realloc) async))
  (core instance $env (export "echo" (func $echo)))
  (core module $code
   (import "env" "echo" (func $echo (param i32 i32 i32) (result i32)))
   (func (export "echo") (param i32 i32) (result i32) local.get 0 local.get 1 i32.const 32 call $echo i32.const 2 i32.ne if unreachable end i32.const 32))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) (realloc $realloc))))
 (instance $client (instantiate $client (with "echo" (func $echo))))
 (alias export $client "echo" (func $call))
 (export "echo" (func $call)))
'''.replace('MEMORY', memory)
fixtures['async-lower-strings'] = strings
for suffix, encoding in [('utf16', 'utf16'), ('compact', 'latin1+utf16')]:
    fixtures['async-lower-strings-'+suffix] = '\n'.join(line.replace('(memory $mem)', '(memory $mem) string-encoding='+encoding) if '(canon ' in line else line for line in strings.splitlines())+'\n'

resources = '''(component
 (import "r" (type $r (sub resource)))
 (type $output (tuple (own $r) u32))
 (export "output" (type $output))
 (import "make" (func $make async (param "arg" (borrow $r)) (result 2)))
 MEMORY
 (core func $make (canon lower (func $make) (memory $mem) async))
 (core func $drop (canon resource.drop $r))
 (core instance $env (export "make" (func $make)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "make" (func $make (param i32 i32) (result i32)))
  (import "env" "drop" (func $drop (param i32)))
  (func (export "call") (param i32 i32) (result i32)
   local.get 0 local.get 1 call $make i32.const 2 i32.ne if unreachable end
   local.get 0 call $drop local.get 1))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "call") (param "arg" (borrow $r)) (param "ptr" u32) (result 2) (canon lift (core func $code "call") (memory $mem)))
 (export "direct" (func $make)))
'''.replace('MEMORY', memory)
fixtures['async-lower-resources'] = resources
fixtures['async-lower-resource-sync'] = resources.replace('(memory $mem) async)', '(memory $mem))').replace('(func $make (param i32 i32) (result i32))', '(func $make (param i32 i32))').replace('call $make i32.const 2 i32.ne if unreachable end', 'call $make')

def memory64(text):
    text = text.replace('(memory (export "memory") 1)', '(memory (export "memory") i64 1)').replace('(memory (export "other") 1)', '(memory (export "other") i64 1)').replace('(memory 1)', '(memory i64 1)')
    text = text.replace('(data (i32.const 0)', '(data (i64.const 0)').replace('(global $bump (mut i32) (i32.const 1024))', '(global $bump (mut i64) (i64.const 1024))')
    text = text.replace('(param $old i32) (param $old-size i32) (param $alignment i32) (param $size i32) (result i32) (local $p i32)', '(param $old i64) (param $old-size i64) (param $alignment i64) (param $size i64) (result i64) (local $p i64)')
    text = text.replace('local.get $size i32.add i32.const 7 i32.add i32.const -8 i32.and', 'local.get $size i64.add i64.const 7 i64.add i64.const -8 i64.and').replace('local.get $old-size local.get $size i32.lt_u', 'local.get $old-size local.get $size i64.lt_u')
    return text

text = memory64(strings).replace('(func $ret (param i32 i32))', '(func $ret (param i64 i64))').replace('(export "echo") (param i32 i32)', '(export "echo") (param i64 i64)').replace('(func $echo (param i32 i32 i32)', '(func $echo (param i64 i64 i64)').replace('(result i32) local.get 0 local.get 1 i32.const 32', '(result i64) local.get 0 local.get 1 i64.const 32').replace('end i32.const 32', 'end i64.const 32')
fixtures['async-lower-strings64'] = text
text = memory64(resources).replace('(func $make (param i32 i32)', '(func $make (param i32 i64)').replace('(export "call") (param i32 i32) (result i32)', '(export "call") (param i32 i64) (result i64)').replace('(param "ptr" u32)', '(param "ptr" u64)')
fixtures['async-lower-resources64'] = text
fixtures['async-lower-resource-sync64'] = text.replace('(memory $mem) async)', '(memory $mem))').replace('(func $make (param i32 i64) (result i32))', '(func $make (param i32 i64))').replace('call $make i32.const 2 i32.ne if unreachable end', 'call $make')

pins = {}
for name, text in fixtures.items():
    path=OUT/(name+'.wat');path.write_text(text)
    subprocess.run(['wasm-tools','parse',str(path),'-o',str(path.with_suffix('.wasm'))],check=True)
    validation=subprocess.run(['wasm-tools','validate','--features','all',str(path.with_suffix('.wasm'))],capture_output=True,text=True)
    if name in ['async-lower64','async-lower-strings64','async-lower-resources64']:
        assert validation.returncode!=0 and 'type mismatch for export' in validation.stderr and 'expected: (func (param ' in validation.stderr and 'found:    (func (param ' in validation.stderr,validation.stderr
        assert ('i32 i64) (result i32))' in validation.stderr and 'i32 i32) (result i32))' in validation.stderr) or ('i64 i64 i64) (result i32))' in validation.stderr and 'i64 i64 i32) (result i32))' in validation.stderr),validation.stderr
    else:
        assert validation.returncode==0,validation.stderr
    pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
    pins[name]['producerValidated']=validation.returncode==0
    if validation.returncode:pins[name]['producerRejection']=validation.stderr.strip()
(OUT/'async-lower.json').write_text(json.dumps({'componentModelCommit':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'memorySourceSHA256':hashlib.sha256((OUT/'async-values.wat').read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',sum(x['producerValidated'] for x in pins.values()),'independently validated fixtures;',sum(not x['producerValidated'] for x in pins.values()),'explicit cm64 producer boundaries')

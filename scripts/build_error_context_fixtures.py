#!/usr/bin/env python3
"""Build pinned error-context ABI, encoding, composition and guard fixtures."""
import hashlib,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
base=(OUT/'error-context.wat').read_text()
fixtures={'error-context':base}
for encoding in ['utf16','latin1+utf16']:
 s='\n'.join(line.replace('(memory $mem)',f'(memory $mem) string-encoding={encoding}') if '(canon ' in line else line for line in base.split('\n'))
 if encoding=='utf16':s=s.replace('"context-\\ce\\bb"','"c\\00o\\00n\\00t\\00e\\00x\\00t\\00-\\00\\bb\\03"').replace('i32.const 10 call $new','i32.const 9 call $new')
 fixtures['error-context-'+('utf16' if encoding=='utf16' else 'compact')]=s
s=base.replace('(memory (export "memory") 1)','(memory (export "memory") i64 1)').replace('(memory 1)','(memory i64 1)')
s=s.replace('(data (i32.const 0)', '(data (i64.const 0)')
s=s.replace('(global $bump (mut i32) (i32.const 1024))','(global $bump (mut i64) (i64.const 1024))')
s=s.replace('(param i32 i32 i32) (param $size i32) (result i32)','(param i64 i64 i64) (param $size i64) (result i64)').replace('(local $ptr i32)','(local $ptr i64)')
s=s.replace('local.get $size i32.add i32.const 7 i32.add i32.const -8 i32.and','local.get $size i64.add i64.const 7 i64.add i64.const -8 i64.and')
s=s.replace('(func $new (param i32 i32)', '(func $new (param i64 i64)').replace('(func $message (param i32 i32))','(func $message (param i32 i64))')
s=s.replace('(export "make") (param i32 i32)','(export "make") (param i64 i64)')
s=s.replace('(export "inspect") (param i32) (result i32)','(export "inspect") (param i32) (result i64)')
s=s.replace('(export "record") (param i32 i32) (result i32)','(export "record") (param i32 i32) (result i64)').replace('(export "post-record") (param i32)','(export "post-record") (param i64)')
for n in [32,64,68]:s=s.replace('i32.const '+str(n),'i64.const '+str(n))
s=s.replace('i32.const 0 i32.const 10 call $new','i64.const 0 i64.const 10 call $new').replace('(param "ptr" u32)','(param "ptr" u64)').replace('(param "len" u32)','(param "len" u64)')
fixtures['error-context64']=s
fixtures['error-context-post-return']=base.replace('(realloc $realloc) )','(realloc $realloc) (post-return $post))',1)
# A real independently compiled component argument crosses a namespace boundary.
fixtures['error-context-components']='''(component
 (type $library (component
  (export "make" (func (param "message" string) (result error-context)))
  (export "echo" (func (param "context" error-context) (result error-context)))
  (export "inspect" (func (param "context" error-context) (result string)))
  (export "release" (func))))
 (import "code" (component $code (type $library)))
 (instance $api (instantiate $code))
 (alias export $api "make" (func $make))
 (alias export $api "echo" (func $echo))
 (alias export $api "inspect" (func $inspect))
 (alias export $api "release" (func $release))
 (export "make" (func $make)) (export "echo" (func $echo))
 (export "inspect" (func $inspect)) (export "release" (func $release)))'''
fixtures['error-context-host']='''(component
 (import "host" (func $host (param "context" error-context) (result error-context)))
 (core func $host (canon lower (func $host)))
 (core func $drop (canon error-context.drop))
 (core instance $env (export "host" (func $host)) (export "drop" (func $drop)))
 (core module $guest
  (import "env" "host" (func $host (param i32) (result i32)))
  (import "env" "drop" (func $drop (param i32)))
  (global $pending (mut i32) (i32.const 0))
  (func $release (export "release") global.get $pending if global.get $pending call $drop i32.const 0 global.set $pending end)
  (func (export "call") (param i32) (result i32) (local $result i32)
   call $release local.get 0 call $host local.set $result local.get 0 call $drop
   local.get $result global.set $pending local.get $result))
 (core instance $guest (instantiate $guest (with "env" (instance $env))))
 (alias core export $guest "call" (core func $call))
 (alias core export $guest "release" (core func $release))
 (func (export "call") (param "context" error-context) (result error-context) (canon lift (core func $call)))
 (func (export "release") (canon lift (core func $release))))'''
# Cross-namespace lower/lift and explicit guest drops, without host component imports.
fixtures['error-context-cross']='(component\n (component $library\n'+base[len('(component'):].rstrip()+'\n'+r'''
 (instance $api (instantiate $library))
 (alias export $api "make" (func $make))
 (alias export $api "echo" (func $echo))
 (alias export $api "inspect" (func $inspect))
 (core module $memory
  (memory (export "memory") 1)
  (data (i32.const 0) "context-\ce\bb")
  (global $next (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32) (param $n i32) (result i32) (local $p i32)
   global.get $next local.tee $p local.get $n i32.add global.set $next local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "realloc" (core func $realloc))
 (core func $make (canon lower (func $make) (memory $mem)))
 (core func $echo (canon lower (func $echo)))
 (core func $inspect (canon lower (func $inspect) (memory $mem) (realloc $realloc)))
 (core func $drop (canon error-context.drop))
 (core instance $env (export "make" (func $make)) (export "echo" (func $echo))
  (export "inspect" (func $inspect)) (export "drop" (func $drop)) (export "memory" (memory $mem)))
 (core module $guest
  (import "env" "make" (func $make (param i32 i32) (result i32)))
  (import "env" "echo" (func $echo (param i32) (result i32)))
  (import "env" "inspect" (func $inspect (param i32 i32)))
  (import "env" "drop" (func $drop (param i32)))
  (import "env" "memory" (memory 1))
  (func (export "run") (param $n i32) (result i32) (local $i i32) (local $a i32) (local $b i32)
   block $end loop $next
    local.get $i local.get $n i32.ge_u br_if $end
    i32.const 0 i32.const 10 call $make local.set $a
    local.get $a call $echo local.set $b local.get $a call $drop
    local.get $b i32.const 32 call $inspect local.get $b call $drop
    i32.const 36 i32.load i32.const 10 i32.ne if unreachable end
    local.get $i i32.const 1 i32.add local.set $i br $next
   end end local.get $i))
 (core instance $guest (instantiate $guest (with "env" (instance $env))))
 (alias core export $guest "run" (core func $run))
 (func (export "run") (param "n" u32) (result u32) (canon lift (core func $run))))
'''
pins={}
for name,text in fixtures.items():
 wat=OUT/(name+'.wat');wat.write_text(text if text.endswith('\n') else text+'\n')
 wasm=OUT/(name+'.wasm')
 subprocess.run(['wasm-tools','parse',str(wat),'-o',str(wasm)],check=True)
 validation=subprocess.run(['wasm-tools','validate','--features','cm-error-context,cm64',str(wasm)],capture_output=True,text=True)
 if name=='error-context64':
  assert validation.returncode!=0 and 'expected: (func (param i64 i64)' in validation.stderr and 'found:    (func (param i32 i32)' in validation.stderr,validation.stderr
 else: assert validation.returncode==0,validation.stderr
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
 pins[name]['producerValidated']=validation.returncode==0
 if validation.returncode: pins[name]['producerRejection']=validation.stderr.strip()
# The values proposal has no producer text syntax at this pin.
import build_resource_value_fixtures as values
binary=values.component(values.val_import('first',b'\x64'),values.definitions(10,values.imp('same',b'\x02\0\0')),values.definitions(11,values.exp('first',2,0),values.exp('second',2,1)))
path=OUT/'error-context-values.wasm';path.write_bytes(binary)
validation=subprocess.run(['wasm-tools','validate','--features','cm-values,cm-error-context',str(path)],capture_output=True,text=True)
pins['error-context-values']={'wasm':hashlib.sha256(binary).hexdigest(),'producerValidated':validation.returncode==0,'textGenerated':False}
if validation.returncode:pins['error-context-values']['producerRejection']=validation.stderr.strip()

(OUT/'error-context.json').write_text(json.dumps({'componentModelCommit':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'helperSHA256':hashlib.sha256((ROOT/'scripts/build_resource_value_fixtures.py').read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',sum(x['producerValidated'] for x in pins.values()),'independently validated fixtures;',sum(not x['producerValidated'] for x in pins.values()),'explicit producer boundaries')

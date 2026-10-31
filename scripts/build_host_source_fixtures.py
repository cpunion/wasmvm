#!/usr/bin/env python3
"""Generate owned lazy host-source consumers with the pinned component producer."""
import hashlib
import json
from pathlib import Path
import subprocess
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
fixtures={}
for kind,element in [('future','u32'),('stream','u32'),('stream','u8'),('future','string'),('stream','')]:
 future=kind=='future';unit=element=='';text=element=='string'
 one='' if future else 'i32.const 1 '
 ptr='-1' if unit else '64'
 completed=0 if future else 17
 opts='' if unit else '(memory $mem)'+(' (realloc $realloc)' if text else '')
 name='host-source-'+kind+'-'+(element or 'unit')
 fixtures[name]=f'''(component
 (type $t ({kind}{(' '+element) if element else ''})) (export "endpoint-type" (type $t))
 (import "make" (func $make (result $t))) (import "started" (func $started (result u32)))
 (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
 (core func $make (canon lower (func $make))) (core func $started (canon lower (func $started)))
 (core func $read (canon {kind}.read $t {opts})) (core func $read-async (canon {kind}.read $t {opts} async))
 (core func $cancel (canon {kind}.cancel-read $t)) (core func $drop (canon {kind}.drop-readable $t))
 (core func $yield (canon thread.yield))
 (core instance $env (export "memory" (memory $mem)) (export "make" (func $make)) (export "read" (func $read)) (export "read-async" (func $read-async))
  (export "cancel" (func $cancel)) (export "drop" (func $drop)) (export "started" (func $started)) (export "yield" (func $yield)))
 (core module $code
  (import "env" "memory" (memory 1))
  (import "env" "make" (func $make (result i32)))
  (import "env" "read" (func $read (param i32 i32{'' if future else ' i32'}) (result i32)))
  (import "env" "read-async" (func $read-async (param i32 i32{'' if future else ' i32'}) (result i32)))
  (import "env" "cancel" (func $cancel (param i32) (result i32))) (import "env" "drop" (func $drop (param i32)))
  (import "env" "yield" (func $yield (result i32))) (import "env" "started" (func $started (result i32)))
  (func (export "run") (param $mode i32) (result i32) (local $r i32)
   call $make local.set $r
   local.get $mode if
    local.get $r i32.const {ptr} {one}call $read-async i32.const -1 i32.ne if unreachable end
    local.get $mode i32.const 3 i32.ge_u if
     loop $wait call $started i32.eqz if call $yield drop br $wait end end
     local.get $mode i32.const 4 i32.eq if unreachable end
    end
    local.get $r call $cancel i32.const 2 i32.ne if unreachable end
   else
    local.get $r i32.const {ptr} {one}call $read i32.const {completed} i32.ne if unreachable end
   end
   local.get $r call $drop
   {'i32.const '+str(completed) if unit else 'i32.const 64'+('' if text else ' i32.load')}))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (param "mode" u32) (result {'string' if text else 'u32'}) (canon lift (core func $code "run"){(' (memory $mem)') if text else ''})))'''
for kind in ['future','stream']:
 one='' if kind=='future' else 'i32.const 1 '
 done=0 if kind=='future' else 17
 fixtures['host-result-'+kind]=f'''(component
 (type $t ({kind} u32)) (export "endpoint-type" (type $t))
 (core module $memory (memory (export "memory") 1)) (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon {kind}.new $t))
 (core func $return (canon task.return (result $t) (memory $mem)))
 (core func $write (canon {kind}.write $t (memory $mem)))
 (core func $drop (canon {kind}.drop-writable $t))
 (core instance $env (export "memory" (memory $mem)) (export "new" (func $new)) (export "return" (func $return))
  (export "write" (func $write)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "memory" (memory 1)) (import "env" "new" (func $new (result i64)))
  (import "env" "return" (func $return (param i32)))
  (import "env" "write" (func $write (param i32 i32{'' if kind=='future' else ' i32'}) (result i32)))
  (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (local $pair i64) (local $w i32) (local $status i32)
   call $new local.tee $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $pair i32.wrap_i64 call $return
   i32.const 32 i32.const 42 i32.store
   local.get $w i32.const 32 {one}call $write
   local.tee $status i32.const {done} i32.ne if local.get $status i32.const 1 i32.ne if unreachable end end local.get $w call $drop))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result $t) (canon lift (core func $code "run") (memory $mem) async)))'''

owned=fixtures['host-source-future-u32'].replace('(component\n','(component\n (import "r" (type $r (sub resource)))\n',1).replace('(future u32)','(future (own $r))',1)
owned=owned.replace(' (core func $cancel ', ' (core func $drop-resource (canon resource.drop $r))\n (core func $cancel ',1)
owned=owned.replace('(core instance $env ', '(core instance $env (export "drop-resource" (func $drop-resource)) ',1)
owned=owned.replace(' (core module $code\n', ' (core module $code\n (import "env" "drop-resource" (func $drop-resource (param i32)))\n',1)
owned=owned.replace('i32.const 64 i32.load))','i32.const 64 i32.load call $drop-resource i32.const 42))',1)
fixtures['host-source-future-owned']=owned

pins={}
for name,text in fixtures.items():
 path=OUT/(name+'.wat');path.write_text(text+'\n')
 subprocess.run(['wasm-tools','parse',str(path),'-o',str(path.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(path.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'host-sources.json').write_text(json.dumps({'componentModelCommit':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'producer-validated host-source consumers')

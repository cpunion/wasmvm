#!/usr/bin/env python3
"""Build producer-validated native WASI 0.3.2 stdio consumers/producers."""
import hashlib,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'pkg/wasi/preview3/testdata'
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
fixtures={}
error='(type $e (enum "io" "illegal-byte-sequence" "pipe")) (export "error-code" (type $err (eq $e)))'
fixtures['stdin']='''(component
 (import "wasi:cli/stdin@0.3.2" (instance $stdin
 '''+error+'''
 (type $s (stream u8)) (type $f (future (result (error $err)))) (type $r (tuple $s $f))
 (export "read-via-stream" (func (result $r)))))
 (alias export $stdin "error-code" (type $err))
 (alias export $stdin "read-via-stream" (func $input))
 (type $s (stream u8)) (type $f (future (result (error $err))))
 (core module $mem (memory (export "memory") 1)) (core instance $mem (instantiate $mem))
 (alias core export $mem "memory" (core memory $memory))
 (core func $input (canon lower (func $input) (memory $memory)))
 (core func $read (canon stream.read $s (memory $memory))) (core func $drop-stream (canon stream.drop-readable $s))
 (core func $result (canon future.read $f (memory $memory))) (core func $drop-future (canon future.drop-readable $f))
 (core instance $env (export "memory" (memory $memory)) (export "input" (func $input)) (export "read" (func $read))
 (export "drop-stream" (func $drop-stream)) (export "result" (func $result)) (export "drop-future" (func $drop-future)))
 (core module $m
  (import "env" "memory" (memory 1)) (import "env" "input" (func $input (param i32)))
  (import "env" "read" (func $read (param i32 i32 i32) (result i32)))
  (import "env" "drop-stream" (func $drop-stream (param i32)))
  (import "env" "result" (func $result (param i32 i32) (result i32)))
  (import "env" "drop-future" (func $drop-future (param i32)))
  (func (export "run") (param $drop i32) (result i32) (local $s i32) (local $f i32) (local $sum i32) (local $status i32)
   i32.const 0 call $input i32.const 0 i32.load local.set $s i32.const 4 i32.load local.set $f
   local.get $drop i32.eqz if
    loop $next
     local.get $s i32.const 32 i32.const 1 call $read local.tee $status i32.const 4 i32.shr_u
     if local.get $sum i32.const 32 i32.load8_u i32.add local.set $sum end
     local.get $status i32.const 15 i32.and i32.eqz br_if $next
    end
   end
   local.get $s call $drop-stream
   local.get $f i32.const 64 call $result i32.eqz if else unreachable end
   local.get $f call $drop-future
   i32.const 64 i32.load8_u i32.const 24 i32.shl
   i32.const 65 i32.load8_u i32.const 16 i32.shl i32.or local.get $sum i32.or))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") async (param "drop" u32) (result u32) (canon lift (core func $m "run"))))'''
for output in ['stdout','stderr']:
 fixtures[output]='''(component
 (import "wasi:cli/'''+output+'''@0.3.2" (instance $output
 '''+error+'''
 (type $s (stream u8)) (type $f (future (result (error $err))))
 (export "write-via-stream" (func (param "data" $s) (result $f)))))
 (alias export $output "error-code" (type $err)) (alias export $output "write-via-stream" (func $output))
 (type $s (stream u8)) (type $f (future (result (error $err))))
 (core module $mem (memory (export "memory") 1) (data (i32.const 32) "abcdef")) (core instance $mem (instantiate $mem))
 (alias core export $mem "memory" (core memory $memory))
 (core func $output (canon lower (func $output))) (core func $new (canon stream.new $s))
 (core func $write (canon stream.write $s (memory $memory))) (core func $drop-stream (canon stream.drop-writable $s))
 (core func $result (canon future.read $f (memory $memory))) (core func $drop-future (canon future.drop-readable $f))
 (core instance $env (export "memory" (memory $memory)) (export "output" (func $output)) (export "new" (func $new)) (export "write" (func $write))
 (export "drop-stream" (func $drop-stream)) (export "result" (func $result)) (export "drop-future" (func $drop-future)))
 (core module $m
  (import "env" "memory" (memory 1)) (import "env" "output" (func $output (param i32) (result i32)))
  (import "env" "new" (func $new (result i64))) (import "env" "write" (func $write (param i32 i32 i32) (result i32)))
  (import "env" "drop-stream" (func $drop-stream (param i32)))
  (import "env" "result" (func $result (param i32 i32) (result i32))) (import "env" "drop-future" (func $drop-future (param i32)))
  (func (export "run") (param $mode i32) (result i32) (local $pair i64) (local $w i32) (local $f i32) (local $i i32) (local $status i32)
   call $new local.tee $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $pair i32.wrap_i64 call $output local.set $f
   local.get $mode i32.const 1 i32.eq if local.get $f call $drop-future end
   local.get $mode i32.const 2 i32.ne if
    block $closed loop $next
     local.get $w i32.const 32 local.get $i i32.add i32.const 1 call $write local.set $status
     local.get $status i32.const 15 i32.and br_if $closed
     local.get $status i32.const 4 i32.shr_u i32.const 1 i32.ne if unreachable end
     local.get $i i32.const 1 i32.add local.tee $i i32.const 6 i32.lt_u br_if $next
    end end
   end
   local.get $w call $drop-stream
   local.get $mode i32.const 1 i32.eq if local.get $status return end
   local.get $f i32.const 64 call $result i32.eqz if else unreachable end
   local.get $f call $drop-future
   i32.const 64 i32.load8_u i32.const 8 i32.shl i32.const 65 i32.load8_u i32.or))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") async (param "mode" u32) (result u32) (canon lift (core func $m "run"))))'''
fixtures['stdout-wrong-result']=fixtures['stdout'].replace('(enum "io" "illegal-byte-sequence" "pipe")','(enum "io" "pipe" "illegal-byte-sequence")')
for i,terminal in enumerate(['stdin','stdout','stderr']):
 resource='terminal-input' if i==0 else 'terminal-output'
 fixtures['terminal-'+terminal]=f'''(component
 (import "wasi:cli/{resource}@0.3.2" (instance $resource (export "{resource}" (type (sub resource)))))
 (alias export $resource "{resource}" (type $r))
 (import "wasi:cli/terminal-{terminal}@0.3.2" (instance $terminal
  (alias outer 1 $r (type $outer)) (export "{resource}" (type $r (eq $outer)))
  (type $option (option (own $r))) (export "get-terminal-{terminal}" (func (result $option)))))
 (alias export $terminal "get-terminal-{terminal}" (func $get))
 (core module $mem (memory (export "memory") 1)) (core instance $mem (instantiate $mem))
 (alias core export $mem "memory" (core memory $memory))
 (core func $get (canon lower (func $get) (memory $memory))) (core func $drop (canon resource.drop $r))
 (core instance $env (export "get" (func $get)) (export "drop" (func $drop)) (export "memory" (memory $memory)))
 (core module $m (import "env" "memory" (memory 1)) (import "env" "get" (func $get (param i32))) (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (result i32) (local $present i32)
   i32.const 0 call $get i32.const 0 i32.load8_u local.tee $present
   if i32.const 4 i32.load call $drop end local.get $present))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") (result u32) (canon lift (core func $m "run"))))'''
fixtures['stdin-wrong-stream']=fixtures['stdin'].replace('(type $s (stream u8))','(type $s (stream u16))',1)
fixtures['stdin-zero-read']=fixtures['stdin'].replace('local.get $drop i32.eqz if','local.get $s i32.const -1 i32.const 0 call $read i32.const 15 i32.and i32.eqz if else unreachable end local.get $drop i32.eqz if',1)
fixtures['stdout-wrong-future']=fixtures['stdout'].replace('(type $f (future (result (error $err))))','(type $f (future u32))',1)
fixtures['stdout-wrong-direction']=fixtures['stdout'].replace('(type $f (future (result (error $err))))','(type $f (stream u32))',1)
pins={}
for name,s in fixtures.items():
 p=OUT/(name+'.wat');p.write_text(s.rstrip()+'\n')
 subprocess.run(['wasm-tools','parse',str(p),'-o',str(p.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(p.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'io.json').write_text(json.dumps({'wasiVersion':'0.3.2','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated WASI 0.3.2 IO fixtures')

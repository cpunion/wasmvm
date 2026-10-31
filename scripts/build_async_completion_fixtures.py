#!/usr/bin/env python3
"""Build stackful suspension and backpressure fixtures from pinned owned inputs."""
import hashlib,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];OUT=ROOT/'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
base=(OUT/'pending-subtasks.wat').read_text()
stackful=base.replace('(core func $cancel (canon task.cancel))','(core func $cancel (canon task.cancel))\n (core func $yield (canon thread.yield))',1)
stackful=stackful.replace('(export "cancel" (func $cancel)))','(export "cancel" (func $cancel)) (export "yield" (func $yield)))',1)
stackful=stackful.replace('(import "env" "cancel" (func $cancel))','(import "env" "cancel" (func $cancel))\n  (import "env" "yield" (func $yield (result i32)))',1)
stackful=stackful.replace('(func (export "start") (param i32) (result i32) local.get 0 call $set i32.const 1)', '(func (export "start") (param i32) local.get 0 call $set call $yield if call $cancel return end call $get i32.const 7 i32.add call $ret)',1)
stackful=stackful.replace('async (callback $callback)', 'async',1)
fixtures={'async-stackful':stackful}
# Sync lower of an async callee must itself park when the callee is unresolved.
sync=stackful[:stackful.index('(component $client')]+'(component $client\n (import "add" (func $add async (param "n" u32) (result u32)))\n (core func $add (canon lower (func $add)))\n (core instance $env (export "add" (func $add)))\n (core module $code (import "env" "add" (func $add (param i32) (result i32)))\n  (func (export "churn") (param $n i32) (result i32) (local $i i32)\n   block $done loop $loop local.get $i local.get $n i32.ge_u br_if $done\n    local.get $i call $add local.get $i i32.const 7 i32.add i32.ne if unreachable end\n    local.get $i i32.const 1 i32.add local.set $i br $loop end end local.get $i))\n (core instance $code (instantiate $code (with "env" (instance $env))))\n (func (export "churn") async (param "n" u32) (result u32) (canon lift (core func $code "churn"))))\n (instance $client (instantiate $client (with "add" (func $add))))\n (alias export $client "churn" (func $churn)) (export "churn" (func $churn)))'
fixtures['async-stackful-sync-lower']=sync

# A five-slot parameter tuple spills at async lower. Cancellation while
# backpressured must not touch deliberately invalid argument/result pointers.
pressure=stackful
pressure=pressure.replace('(core func $ret (canon task.return (result u32)))','''(core func $ret (canon task.return (result u32)))
 (core func $inc (canon backpressure.inc)) (core func $dec (canon backpressure.dec))''',1)
pressure=pressure.replace('(export "yield" (func $yield)))','(export "yield" (func $yield)) (export "inc" (func $inc)) (export "dec" (func $dec)))',1)
pressure=pressure.replace('(import "env" "yield" (func $yield (result i32)))','''(import "env" "yield" (func $yield (result i32)))
 (import "env" "inc" (func $inc)) (import "env" "dec" (func $dec))
 (func (export "pressure") (param i32) local.get 0 if call $inc else call $dec end)''',1)
pressure=pressure.replace('(func (export "start") (param i32) ', '(func (export "start") (param i32 i32 i32 i32 i32) ',1)
pressure=pressure.replace('(param "n" u32) (result u32)', '(param "n" u32) (param "a" u32) (param "b" u32) (param "c" u32) (param "d" u32) (result u32)',2)
pressure=pressure.replace('(func (export "add") async', '(func (export "pressure") (param "on" u32) (canon lift (core func $code "pressure")))\n (func (export "add") async',1)
pressure=pressure.replace('(alias export $server "add" (func $add))','(alias export $server "add" (func $add))\n (alias export $server "pressure" (func $pressure))',1)
pressure=pressure.replace(' (component $client',' (component $client\n (import "pressure" (func $pressure (param "on" u32))))',1).replace('(param "on" u32))))','(param "on" u32)))',1)
pressure=pressure.replace('  (core func $add (canon lower', '  (core func $pressure (canon lower (func $pressure)))\n  (core func $add (canon lower',1)
pressure=pressure.replace('(core instance $env (export "memory" (memory $mem)) (export "add" (func $add))','(core instance $env (export "pressure" (func $pressure)) (export "memory" (memory $mem)) (export "add" (func $add))',1)
pressure=pressure.replace('(import "env" "memory" (memory 1))','(import "env" "pressure" (func $pressure (param i32)))\n (import "env" "memory" (memory 1))',1)
marker=' (func $start (param i32)'
pos=pressure.index(marker,pressure.index('(component $client'))
code=''' (func (export "pressure") (param $cancel i32) (result i32) (local $sub i32) (local $set i32) (local $state i32)
  i32.const 1 call $pressure
  i32.const 256 i32.const 42 i32.store i32.const 260 i32.const 1 i32.store
  i32.const 264 i32.const 2 i32.store i32.const 268 i32.const 3 i32.store i32.const 272 i32.const 4 i32.store
  local.get $cancel if (result i32) i32.const 65532 else i32.const 256 end
  local.get $cancel if (result i32) i32.const 3 else i32.const 32 end
  call $add local.tee $sub i32.const 15 i32.and if unreachable end
  local.get $sub i32.const 4 i32.shr_u local.set $sub
  local.get $cancel if
   local.get $sub call $cancel-async local.set $state
   local.get $sub call $subdrop i32.const 0 call $pressure local.get $state return
  end
  i32.const 0 call $pressure
  call $new local.set $set local.get $sub local.get $set call $join
  loop $wait
   local.get $set i32.const 64 call $wait i32.const 1 i32.ne if unreachable end
   i32.const 68 i32.load local.tee $state i32.const 1 i32.eq if i32.const 256 i32.const 999 i32.store br $wait end
  end
  local.get $state i32.const 2 i32.ne if unreachable end
  local.get $sub call $subdrop local.get $set call $drop i32.const 32 i32.load)
'''
pressure=pressure[:pos]+code+pressure[pos:]
pressure=pressure.replace('(func (export "pressure") (param i32) local.get 0 if call $inc else call $dec end)', '(func (export "pressure") (param i32) (local $i i32) local.get 0 i32.const 2 i32.eq if loop $overflow call $inc local.get $i i32.const 1 i32.add local.tee $i i32.const 65536 i32.lt_u br_if $overflow end else local.get 0 if call $inc else call $dec end end)',1)
pressure=pressure.rstrip()[:-1]+' (export "control" (func $pressure)))'
pressure=pressure.replace('  (func (export "churn") async', '  (func (export "pressure") async (param "cancel" u32) (result u32) (canon lift (core func $code "pressure")))\n  (func (export "churn") async',1)
pressure=pressure.replace('(with "add" (func $add))))','(with "add" (func $add)) (with "pressure" (func $pressure))))',1)
pressure=pressure.replace(' (alias export $client "multi"', ' (alias export $client "pressure" (func $pressure-call)) (export "pressure" (func $pressure-call))\n (alias export $client "multi"',1)
fixtures['async-backpressure']=pressure
# Imported completions may become ready on another host thread, but all canonical
# work and fresh-Caller resource operations remain inside the Store session.
host='(component\n (import "add" (func $add async (param "n" u32) (result u32)))\n (import "notify" (func $notify))\n'+base[base.index(' (component $client'):]
host=host.replace('  (import "add"', '  (import "notify" (func $notify))\n  (import "add"',1)
host=host.replace('  (core func $add (canon lower', '  (core func $notify (canon lower (func $notify)))\n  (core func $add (canon lower',1)
host=host.replace('(core instance $env (export "memory"', '(core instance $env (export "notify" (func $notify)) (export "memory"',1)
host=host.replace('(import "env" "memory"', '(import "env" "notify" (func $notify))\n (import "env" "memory"',1)
host=host.replace('  local.get $packed i32.const 4', '  call $notify\n  local.get $packed i32.const 4',1)
host=host.replace('(with "add" (func $add))))','(with "add" (func $add)) (with "notify" (func $notify))))',1)
host=host.rstrip()[:-1]+' (export "direct" (func $add)))'
fixtures['async-host-completion']=host

fixtures['async-host-resources']='(component\n (import "r" (type $r (sub resource)))\n (import "pass" (func $pass async (param "arg" (own $r)) (result (own $r))))\n (import "borrow" (func $borrow async (param "arg" (borrow $r)) (result u32)))\n (import "make" (func $make async (result (own $r))))\n (type $blob (list u8)) (export "blob-type" (type $blob))\n (import "blob" (func $blob async (param "arg" $blob) (result $blob)))\n (export "pass" (func $pass)) (export "borrow" (func $borrow))\n (export "make" (func $make)) (export "blob" (func $blob)))'

pins={}
for name,text in fixtures.items():
 path=OUT/(name+'.wat');path.write_text(text.rstrip()+'\n')
 subprocess.run(['wasm-tools','parse',str(path),'-o',str(path.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(path.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'async-completion.json').write_text(json.dumps({'componentModelCommit':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sourceSHA256':hashlib.sha256((OUT/'pending-subtasks.wat').read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated continuation/backpressure fixtures')

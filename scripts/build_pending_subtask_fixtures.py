#!/usr/bin/env python3
"""Build real callback suspensions, waitable events and cancellation fixtures."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'pkg/component/testdata'
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'

SERVER = '''(component $server
 (core func $ret (canon task.return (result u32)))
 (core func $get (canon context.get i32 0))
 (core func $set (canon context.set i32 0))
 (core func $cancel (canon task.cancel))
 (core instance $env (export "ret" (func $ret)) (export "get" (func $get)) (export "set" (func $set)) (export "cancel" (func $cancel)))
 (core module $code
  (import "env" "ret" (func $ret (param i32)))
  (import "env" "get" (func $get (result i32)))
  (import "env" "set" (func $set (param i32)))
  (import "env" "cancel" (func $cancel))
  (func (export "start") (param i32) (result i32) local.get 0 call $set i32.const 1)
  (func (export "callback") (param i32 i32 i32) (result i32)
   local.get 0 i32.const 6 i32.eq if call $cancel i32.const 0 return end
   local.get 0 if unreachable end
   call $get i32.const 7 i32.add call $ret i32.const 0))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (alias core export $code "callback" (core func $callback))
 (func (export "add") async (param "n" u32) (result u32) (canon lift (core func $code "start") async (callback $callback))))'''

WAIT = '''(core func $new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $poll (canon waitable-set.poll (memory $mem)))
 (core func $drop (canon waitable-set.drop))
 (core func $join (canon waitable.join))
 (core func $subdrop (canon subtask.drop))
 (core func $cancel (canon subtask.cancel))
 (core func $cancel-async (canon subtask.cancel async))
 (core instance $env (export "memory" (memory $mem)) (export "add" (func $add)) (export "new" (func $new)) (export "wait" (func $wait)) (export "poll" (func $poll)) (export "drop" (func $drop)) (export "join" (func $join)) (export "subdrop" (func $subdrop)) (export "cancel" (func $cancel)) (export "cancel-async" (func $cancel-async)))'''
IMPORTS = '''(import "env" "memory" (memory 1))
 (import "env" "add" (func $add (param i32 i32) (result i32)))
 (import "env" "new" (func $new (result i32)))
 (import "env" "wait" (func $wait (param i32 i32) (result i32)))
 (import "env" "poll" (func $poll (param i32 i32) (result i32)))
 (import "env" "drop" (func $drop (param i32)))
 (import "env" "join" (func $join (param i32 i32)))
 (import "env" "subdrop" (func $subdrop (param i32)))
 (import "env" "cancel" (func $cancel (param i32) (result i32)))
 (import "env" "cancel-async" (func $cancel-async (param i32) (result i32)))'''

CODE = '''(core module $code
 IMPORTS
 (func $start (param i32) (result i32) (local $packed i32)
  local.get 0 i32.const 32 call $add local.tee $packed i32.const 15 i32.and i32.const 1 i32.ne if unreachable end
  local.get $packed i32.const 4 i32.shr_u)
 (func $finish (param $sub i32) (param $set i32) (result i32)
  local.get $sub call $subdrop local.get $set call $drop i32.const 32 i32.load)
 (func (export "churn") (param $n i32) (result i32) (local $i i32) (local $sub i32) (local $set i32)
  block $done loop $loop local.get $i local.get $n i32.ge_u br_if $done
   local.get $i call $start local.set $sub call $new local.set $set
   local.get $sub local.get $set call $join
   local.get $set i32.const 64 call $poll if unreachable end
   local.get $set i32.const 64 call $wait i32.const 1 i32.ne if unreachable end
   i32.const 64 i32.load local.get $sub i32.ne if unreachable end
   i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
   local.get $sub local.get $set call $finish local.get $i i32.const 7 i32.add i32.ne if unreachable end
   local.get $i i32.const 1 i32.add local.set $i br $loop end end local.get $i)
 (func (export "cancel") (param $async i32) (result i32) (local $sub i32) (local $state i32)
  i32.const 1 call $start local.set $sub
  local.get $async if (result i32) local.get $sub call $cancel-async else local.get $sub call $cancel end local.set $state
  local.get $state i32.const -1 i32.eq if call $new local.set $async local.get $sub local.get $async call $join local.get $async i32.const 64 call $wait drop local.get $sub call $subdrop local.get $async call $drop local.get $state return end local.get $sub call $subdrop local.get $state)
 (func (export "bad") (param $mode i32) (result i32) (local $sub i32) (local $set i32)
  i32.const 1 call $start local.set $sub call $new local.set $set
  local.get $mode i32.eqz if local.get $sub call $subdrop end
  local.get $mode i32.const 1 i32.eq if local.get $sub local.get $set call $join local.get $sub call $cancel drop end
  local.get $mode i32.const 2 i32.eq if local.get $sub local.get $set call $join local.get $set call $drop end
  local.get $mode i32.const 3 i32.eq if local.get $set call $subdrop end
  local.get $mode i32.const 4 i32.eq if local.get $sub call $drop end
  local.get $mode i32.const 5 i32.eq if local.get $sub call $cancel drop local.get $sub call $cancel drop end
  local.get $mode i32.const 6 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 3 call $poll drop end
  local.get $mode i32.const 7 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 65532 call $wait drop end
  local.get $mode i32.const 8 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 0 local.get $set call $join end
  local.get $mode i32.const 9 i32.eq if i32.const 2 call $start drop end
  i32.const 0))'''.replace('IMPORTS', IMPORTS)
# mode 8 passes a set handle as a waitable, with no unrelated stack operands.
CODE = CODE.replace('local.get $set i32.const 0 local.get $set call $join', 'local.get $set i32.const 0 call $join')

text = '''(component
 SERVER
 (instance $server (instantiate $server))
 (alias export $server "add" (func $add))
 (component $client
  (import "add" (func $add async (param "n" u32) (result u32)))
  (core module $memory (memory (export "memory") 1))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem))
  (core func $add (canon lower (func $add) (memory $mem) async))
  WAIT
  CODE
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "churn") (param "n" u32) (result u32) (canon lift (core func $code "churn")))
  (func (export "cancel") (param "async" u32) (result u32) (canon lift (core func $code "cancel")))
  (func (export "bad") (param "mode" u32) (result u32) (canon lift (core func $code "bad"))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "churn" (func $churn)) (export "churn" (func $churn))
 (alias export $client "cancel" (func $cancel)) (export "cancel" (func $cancel))
 (alias export $client "bad" (func $bad)) (export "bad" (func $bad)))'''.replace('SERVER', SERVER).replace('WAIT', WAIT).replace('CODE', CODE)
multi = """ (func (export "multi") (param $n i32) (result i32) (local $a i32) (local $b i32) (local $c i32) (local $s i32) (local $other i32)
  local.get $n i32.const 32 call $add i32.const 4 i32.shr_u local.set $a
  local.get $n i32.const 1 i32.add i32.const 40 call $add i32.const 4 i32.shr_u local.set $b
  local.get $n i32.const 2 i32.add i32.const 48 call $add i32.const 4 i32.shr_u local.set $c
  call $new local.set $s call $new local.set $other
  local.get $a local.get $s call $join local.get $b local.get $s call $join local.get $c local.get $s call $join
  local.get $b local.get $other call $join local.get $b local.get $s call $join local.get $other call $drop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $poll if unreachable end local.get $s call $drop
  i32.const 32 i32.load i32.const 40 i32.load i32.add i32.const 48 i32.load i32.add)
"""
text=text.replace('  i32.const 0))','  i32.const 0)'+multi+')')
text=text.replace('  (func (export "churn") (param "n"', '  (func (export "multi") async (param "n" u32) (result u32) (canon lift (core func $code "multi")))\n  (func (export "churn") (param "n"')
text=text.replace(' (alias export $client "churn"', ' (alias export $client "multi" (func $multi)) (export "multi" (func $multi))\n (alias export $client "churn"')
text = text.replace('(func (export "churn") (param "', '(func (export "churn") async (param "').replace('(func (export "cancel") (param "', '(func (export "cancel") async (param "').replace('(func (export "bad") (param "', '(func (export "bad") async (param "')
fixtures = {'pending-subtasks': text}
fixtures['pending-sync-trap'] = text.replace('(func (export "churn") async', '(func (export "churn")')
for name, replacement in [('pending-cancel-return', 'call $get i32.const 7 i32.add call $ret'), ('pending-cancel-deferred', 'i32.const 1 return')]:
 fixtures[name] = text.replace('if call $cancel i32.const 0 return end', 'if '+replacement+' i32.const 0 return end')

CALLBACK = """(component
 SERVER
 (instance $server (instantiate $server)) (alias export $server "add" (func $add))
 (component $client
  (import "add" (func $add async (param "n" u32) (result u32)))
  (core module $memory (memory (export "memory") 1)) (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem))
  (core func $add (canon lower (func $add) (memory $mem) async))
  (core func $ret (canon task.return (result u32)))
  (core func $get0 (canon context.get i32 0)) (core func $get1 (canon context.get i32 1))
  (core func $set0 (canon context.set i32 0)) (core func $set1 (canon context.set i32 1))
  WAIT
  (core instance $tls (export "ret" (func $ret)) (export "get0" (func $get0)) (export "get1" (func $get1)) (export "set0" (func $set0)) (export "set1" (func $set1)))
  (core module $code
   IMPORTS
   (import "tls" "ret" (func $ret (param i32)))
   (import "tls" "get0" (func $get0 (result i32))) (import "tls" "get1" (func $get1 (result i32)))
   (import "tls" "set0" (func $set0 (param i32))) (import "tls" "set1" (func $set1 (param i32)))
   (func (export "start") (param i32) (result i32)
    call $new call $set1
    local.get 0 i32.const 32 call $add i32.const 4 i32.shr_u call $set0
    call $get0 call $get1 call $join call $get1 i32.const 4 i32.shl i32.const 2 i32.or)
   (func (export "callback") (param i32 i32 i32) (result i32)
    local.get 0 i32.const 1 i32.ne if unreachable end
    local.get 1 call $get0 i32.ne if unreachable end
    local.get 2 i32.const 2 i32.ne if unreachable end
    call $get0 call $subdrop call $get1 call $drop i32.const 32 i32.load call $ret i32.const 0))
  (core instance $code (instantiate $code (with "env" (instance $env)) (with "tls" (instance $tls))))
  (alias core export $code "callback" (core func $callback))
  (func (export "run") async (param "n" u32) (result u32) (canon lift (core func $code "start") async (callback $callback))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "run" (func $run)) (export "run" (func $run)))""".replace('SERVER',SERVER).replace('WAIT',WAIT).replace('IMPORTS',IMPORTS)
fixtures['pending-callback'] = CALLBACK
fixtures['pending-deadlock'] = CALLBACK.replace('i32.const 4 i32.shr_u call $set0', 'i32.const 4 i32.shr_u call $set0').replace('call $get0 call $get1 call $join call $get1', 'call $get1')

# Strings cross a suspended sibling call before result-memory allocation.
strings = (OUT/'async-lower-strings.wat').read_text()
strings = strings.replace('  (core instance $env (export "ret" (func $ret)))', '  (core func $get0 (canon context.get i32 0))\n  (core func $get1 (canon context.get i32 1))\n  (core func $set0 (canon context.set i32 0))\n  (core func $set1 (canon context.set i32 1))\n  (core instance $env (export "ret" (func $ret)) (export "get0" (func $get0)) (export "get1" (func $get1)) (export "set0" (func $set0)) (export "set1" (func $set1)))')
strings = strings.replace('(core module $code (import "env" "ret" (func $ret (param i32 i32))) (func (export "echo") (param i32 i32) local.get 0 local.get 1 call $ret))', '(core module $code (import "env" "ret" (func $ret (param i32 i32)))\n   (import "env" "get0" (func $get0 (result i32))) (import "env" "get1" (func $get1 (result i32)))\n   (import "env" "set0" (func $set0 (param i32))) (import "env" "set1" (func $set1 (param i32)))\n   (func (export "echo") (param i32 i32) (result i32) local.get 0 call $set0 local.get 1 call $set1 i32.const 1)\n   (func (export "callback") (param i32 i32 i32) (result i32) call $get0 call $get1 call $ret i32.const 0))')
strings = strings.replace('  (func (export "echo") async', '  (alias core export $code "callback" (core func $callback))\n  (func (export "echo") async').replace('(realloc $realloc) async)))', '(realloc $realloc) async (callback $callback))))')
strings = strings.replace('  (core instance $env (export "echo" (func $echo)))', WAIT.replace('(export "add" (func $add))', '(export "echo" (func $echo))'))
strings = strings.replace('(import "env" "echo" (func $echo (param i32 i32 i32) (result i32)))', IMPORTS.replace('(import "env" "add" (func $add (param i32 i32) (result i32)))','(import "env" "echo" (func $echo (param i32 i32 i32) (result i32)))'))
strings = strings.replace('(param i32 i32) (result i32) local.get 0 local.get 1 i32.const 32 call $echo i32.const 2 i32.ne if unreachable end i32.const 32', '(param i32 i32) (result i32) (local $sub i32) (local $set i32)\n    local.get 0 local.get 1 i32.const 32 call $echo i32.const 4 i32.shr_u local.set $sub\n    call $new local.set $set local.get $sub local.get $set call $join\n    local.get $set i32.const 64 call $wait i32.const 1 i32.ne if unreachable end\n    i32.const 68 i32.load i32.const 2 i32.ne if unreachable end\n    local.get $sub call $subdrop local.get $set call $drop i32.const 32')
strings = strings.replace('(func (export "echo") (param "', '(func (export "echo") async (param "')
fixtures['pending-strings'] = strings.rstrip()
for suffix, encoding in [('utf16','utf16'),('compact','latin1+utf16')]:
 fixtures['pending-strings-'+suffix] = '\n'.join(line.replace('(memory $mem)', '(memory $mem) string-encoding='+encoding) if '(canon ' in line and 'waitable-set.' not in line else line for line in strings.splitlines())

values = (OUT/'async-values.wat').read_text()
memory = values[values.index(' (core module $memory'):values.index(' (core func $ret ')]
resource = '''(component
 (import "r" (type $r (sub resource)))
 (type $out (tuple (own $r) u32)) (export "out" (type $out))
 (import "make" (func $make async (param "arg" (borrow $r)) (result 2)))
 (component $server
  (import "r" (type $r (sub resource)))
  (type $out (tuple (own $r) u32)) (export "out" (type $out))
  (import "make" (func $make async (param "arg" (borrow $r)) (result 2)))
  MEMORY
  (core func $ret (canon task.return (result $out) (memory $mem)))
  (core func $get (canon context.get i32 0)) (core func $set (canon context.set i32 0))
  (core func $drop (canon resource.drop $r))
  (core func $make (canon lower (func $make) (memory $mem)))
  (core instance $env (export "memory" (memory $mem)) (export "ret" (func $ret)) (export "get" (func $get)) (export "set" (func $set)) (export "drop" (func $drop)) (export "make" (func $make)))
  (core module $code
   (import "env" "memory" (memory 1)) (import "env" "ret" (func $ret (param i32 i32)))
   (import "env" "get" (func $get (result i32))) (import "env" "set" (func $set (param i32)))
   (import "env" "drop" (func $drop (param i32))) (import "env" "make" (func $make (param i32 i32)))
   (func (export "start") (param i32) (result i32) local.get 0 call $set i32.const 1)
   (func (export "callback") (param i32 i32 i32) (result i32)
    call $get i32.const 64 call $make call $get call $drop
    i32.const 64 i32.load i32.const 68 i32.load call $ret i32.const 0))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (alias core export $code "callback" (core func $callback))
  (func (export "add") async (param "arg" (borrow $r)) (result 2) (canon lift (core func $code "start") (memory $mem) async (callback $callback))))
 (instance $server (instantiate $server (with "r" (type $r)) (with "make" (func $make))))
 (alias export $server "add" (func $add))
 (component $client
  (import "r" (type $r (sub resource)))
  (type $out (tuple (own $r) u32)) (export "out" (type $out))
  (import "add" (func $add async (param "arg" (borrow $r)) (result 2)))
  MEMORY
  (core func $add (canon lower (func $add) (memory $mem) async))
  (core func $borrow-drop (canon resource.drop $r))
  WAIT
  (core instance $resource (export "drop" (func $borrow-drop)))
  (core module $code
   IMPORTS
   (import "resource" "drop" (func $borrow-drop (param i32)))
   (func (export "call") (param i32 i32) (result i32) (local $sub i32) (local $set i32)
    local.get 0 local.get 1 call $add i32.const 4 i32.shr_u local.set $sub
    call $new local.set $set local.get $sub local.get $set call $join
    local.get $set i32.const 96 call $wait i32.const 1 i32.ne if unreachable end
    i32.const 100 i32.load i32.const 2 i32.ne if unreachable end
    local.get $sub call $subdrop local.get $set call $drop local.get 0 call $borrow-drop local.get 1))
  (core instance $code (instantiate $code (with "env" (instance $env)) (with "resource" (instance $resource))))
  (func (export "call") (param "arg" (borrow $r)) (param "ptr" u32) (result 2) (canon lift (core func $code "call") (memory $mem))))
 (instance $client (instantiate $client (with "r" (type $r)) (with "add" (func $add))))
 (alias export $client "call" (func $call)) (export "call" (func $call)))'''.replace('MEMORY',memory).replace('WAIT',WAIT).replace('IMPORTS',IMPORTS)
resource = resource.replace('(func (export "call") (param "', '(func (export "call") async (param "')
fixtures['pending-resources'] = resource

wide = '''(component
 (component $server
  (core func $ret (canon task.return))
  (core instance $env (export "ret" (func $ret)))
  (core module $code (import "env" "ret" (func $ret))
   (func (export "start") (param i32) (result i32) i32.const 1)
   (func (export "callback") (param i32 i32 i32) (result i32) call $ret i32.const 0))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (alias core export $code "callback" (core func $callback))
  (func (export "add") async (param "n" u32) (canon lift (core func $code "start") async (callback $callback))))
 (instance $server (instantiate $server)) (alias export $server "add" (func $add))
 (component $client
  (import "add" (func $add async (param "n" u32)))
  (core module $memory (memory (export "memory") i64 1)) (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem))
  (core func $add (canon lower (func $add) async))
  WAIT
  (core module $code
   IMPORTS
   (func (export "churn") (param $n i32) (result i32) (local $i i32) (local $sub i32) (local $set i32)
    block $done loop $loop local.get $i local.get $n i32.ge_u br_if $done
     local.get $i call $add i32.const 4 i32.shr_u local.set $sub call $new local.set $set
     local.get $sub local.get $set call $join
     local.get $set i64.const 64 call $wait i32.const 1 i32.ne if unreachable end
     i64.const 68 i32.load i32.const 2 i32.ne if unreachable end
     local.get $sub call $subdrop local.get $set call $drop
     local.get $i i32.const 1 i32.add local.set $i br $loop end end local.get $i)
   (func (export "poll") (param $ptr i64) (result i32) (local $set i32) (local $event i32)
    call $new local.tee $set local.get $ptr call $poll local.set $event local.get $set call $drop local.get $event))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "churn") async (param "n" u32) (result u32) (canon lift (core func $code "churn")))
  (func (export "poll") (param "ptr" u64) (result u32) (canon lift (core func $code "poll"))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "churn" (func $churn)) (export "churn" (func $churn))
 (alias export $client "poll" (func $poll)) (export "poll" (func $poll)))'''.replace('WAIT',WAIT).replace('IMPORTS',IMPORTS.replace('(memory 1)','(memory i64 1)').replace('(func $add (param i32 i32)', '(func $add (param i32)').replace('(func $wait (param i32 i32)', '(func $wait (param i32 i64)').replace('(func $poll (param i32 i32)', '(func $poll (param i32 i64)'))
wide = wide.replace('   (func (export "poll") (param $ptr i64)', '   (func (export "high") (result i32) (local $set i32) call $new local.tee $set i64.const 4294967296 call $poll)\n   (func (export "poll") (param $ptr i64)')
wide = wide.replace('  (func (export "poll") (param "ptr" u64)', '  (func (export "high") (result u32) (canon lift (core func $code "high")))\n  (func (export "poll") (param "ptr" u64)')
wide = wide.replace(' (alias export $client "poll"', ' (alias export $client "high" (func $high)) (export "high" (func $high))\n (alias export $client "poll"')
fixtures['pending-memory64'] = wide

pins = {}
for name,text in fixtures.items():
 path=OUT/(name+'.wat');path.write_text(text+'\n')
 subprocess.run(['wasm-tools','parse',str(path),'-o',str(path.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(path.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'pending-subtasks.json').write_text(json.dumps({'componentModelCommit':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sources':{name:hashlib.sha256((OUT/name).read_bytes()).hexdigest() for name in ['async-values.wat','async-lower-strings.wat']},'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'independently validated pending-subtask fixtures')

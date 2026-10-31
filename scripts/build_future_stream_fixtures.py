#!/usr/bin/env python3
"""Generate owned future/stream protocol guests using the pinned producer."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "pkg/component/testdata"
assert subprocess.check_output(["wasm-tools", "--version"], text=True).strip() == "wasm-tools 1.258.0"
fixtures = {}
for kind in ["stream", "future"]:
    future = kind == "future"
    length = "" if future else " i32"
    one = "" if future else "i32.const 1 "
    completed = 0 if future else 16
    fixtures["async-" + kind + "-protocol"] = f'''(component
 (type $t ({kind} u32))
 (core module $memory (memory (export "memory") 1))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon {kind}.new $t))
 (core func $read (canon {kind}.read $t (memory $mem) async))
 (core func $write (canon {kind}.write $t (memory $mem) async))
 (core func $cancel (canon {kind}.cancel-read $t async))
 (core func $drop-r (canon {kind}.drop-readable $t))
 (core func $drop-w (canon {kind}.drop-writable $t))
 (core func $join (canon waitable.join))
 (core func $set-new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $set-drop (canon waitable-set.drop))
 (core instance $env (export "memory" (memory $mem))
  (export "new" (func $new)) (export "read" (func $read)) (export "write" (func $write))
  (export "cancel" (func $cancel)) (export "drop-r" (func $drop-r)) (export "drop-w" (func $drop-w))
  (export "join" (func $join)) (export "set-new" (func $set-new)) (export "wait" (func $wait)) (export "set-drop" (func $set-drop)))
 (core module $code
  (import "env" "memory" (memory 1))
  (import "env" "new" (func $new (result i64)))
  (import "env" "read" (func $read (param i32 i32{length}) (result i32)))
  (import "env" "write" (func $write (param i32 i32{length}) (result i32)))
  (import "env" "cancel" (func $cancel (param i32) (result i32)))
  (import "env" "drop-r" (func $drop-r (param i32))) (import "env" "drop-w" (func $drop-w (param i32)))
  (import "env" "join" (func $join (param i32 i32)))
  (import "env" "set-new" (func $set-new (result i32)))
  (import "env" "wait" (func $wait (param i32 i32) (result i32)))
  (import "env" "set-drop" (func $set-drop (param i32)))
  (func (export "run") (param $n i32) (result i32) (local $r i32) (local $w i32) (local $pair i64) (local $set i32)
   call $new local.tee $pair i32.wrap_i64 local.set $r
   local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $r i32.const 64 {one}call $read i32.const -1 i32.ne if unreachable end
   local.get $r call $cancel i32.const 2 i32.ne if unreachable end
   local.get $r i32.const 64 {one}call $read i32.const -1 i32.ne if unreachable end
   call $set-new local.set $set local.get $r local.get $set call $join
   i32.const 32 local.get $n i32.store
   local.get $w i32.const 32 {one}call $write i32.const {completed} i32.ne if unreachable end
   local.get $set i32.const 96 call $wait i32.const {4 if future else 2} i32.ne if unreachable end
   i32.const 100 i32.load i32.const {completed} i32.ne if unreachable end
   local.get $w call $drop-w local.get $r call $drop-r local.get $set call $set-drop
   i32.const 64 i32.load))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (param "n" u32) (result u32) (canon lift (core func $code "run"))))'''
# Transfer affine readable ends through an immediate host round trip, then
# into a distinct component whose string buffer uses its own memory/allocator.
for kind in ["stream", "future"]:
    future = kind == "future"
    one = "" if future else "i32.const 1 "
    payload = 0 if future else 17
    done = 0 if future else 16
    fixtures["async-"+kind+"-strings"] = f'''(component
 (type $t ({kind} string)) (export "endpoint-type" (type $t))
 (import "pass" (func $pass (param "end" $t) (result $t)))
 (component $consumer
  (type $t ({kind} string)) (export "endpoint-type" (type $t))
  (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
   (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
    global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
  (core func $read (canon {kind}.read $t (memory $mem) (realloc $realloc) async))
  (core func $drop (canon {kind}.drop-readable $t))
  (core instance $env (export "read" (func $read)) (export "drop" (func $drop)))
  (core module $code
   (import "env" "read" (func $read (param i32 i32{'' if future else ' i32'}) (result i32)))
   (import "env" "drop" (func $drop (param i32)))
   (func (export "consume") (param i32) (result i32)
    local.get 0 i32.const 64 {one}call $read i32.const {done} i32.ne if unreachable end
    local.get 0 call $drop i32.const 64))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "consume") async (param "end" $t) (result string) (canon lift (core func $code "consume") (memory $mem))))
 (instance $consumer (instantiate $consumer)) (alias export $consumer "consume" (func $consume))
 (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
  (data (i32.const 256) "caf\\c3\\a9\\e9\\9b\\aa")
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
 (core func $pass (canon lower (func $pass)))
 (core func $consume (canon lower (func $consume) (memory $mem) (realloc $realloc) async))
 (core func $new (canon {kind}.new $t)) (core func $write (canon {kind}.write $t (memory $mem) async))
 (core func $drop (canon {kind}.drop-writable $t))
 (core func $join (canon waitable.join)) (core func $set-new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem))) (core func $set-drop (canon waitable-set.drop))
 (core instance $env (export "memory" (memory $mem)) (export "pass" (func $pass)) (export "consume" (func $consume))
  (export "new" (func $new)) (export "write" (func $write)) (export "drop" (func $drop))
  (export "join" (func $join)) (export "set-new" (func $set-new)) (export "wait" (func $wait)) (export "set-drop" (func $set-drop)))
 (core module $code
  (import "env" "memory" (memory 1)) (import "env" "pass" (func $pass (param i32) (result i32)))
  (import "env" "consume" (func $consume (param i32 i32) (result i32)))
  (import "env" "new" (func $new (result i64)))
  (import "env" "write" (func $write (param i32 i32{'' if future else ' i32'}) (result i32)))
  (import "env" "drop" (func $drop (param i32))) (import "env" "join" (func $join (param i32 i32)))
  (import "env" "set-new" (func $set-new (result i32))) (import "env" "wait" (func $wait (param i32 i32) (result i32)))
  (import "env" "set-drop" (func $set-drop (param i32)))
  (func (export "run") (result i32) (local $r i32) (local $w i32) (local $pair i64) (local $set i32)
   call $new local.tee $pair i32.wrap_i64 local.set $r local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   i32.const 32 i32.const 256 i32.store i32.const 36 i32.const 8 i32.store
   local.get $w i32.const 32 {one}call $write i32.const -1 i32.ne if unreachable end
   local.get $r call $pass i32.const 96 call $consume i32.const 2 i32.ne if unreachable end
   call $set-new local.set $set local.get $w local.get $set call $join
   local.get $set i32.const 128 call $wait i32.const {5 if future else 3} i32.ne if unreachable end
   i32.const 132 i32.load i32.const {payload} i32.ne if unreachable end
   local.get $w call $drop local.get $set call $set-drop i32.const 96))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result string) (canon lift (core func $code "run") (memory $mem))))'''

# Exercise forwarding, peer closure, illegal state transitions and zero-length
# readiness independently from ordinary copying and cross-component transport.
for kind in ["stream", "future"]:
    future = kind == "future"
    text = fixtures["async-" + kind + "-protocol"]
    text = text.replace(' (core func $drop-r ', f' (core func $forward (canon {kind}.forward $t))\n (core func $cancel-w (canon {kind}.cancel-write $t async))\n (core func $drop-r ', 1)
    text = text.replace('(core instance $env ', '(core instance $env (export "forward" (func $forward)) (export "cancel-w" (func $cancel-w)) ', 1)
    text = text.replace('  (import "env" "memory"', '  (import "env" "forward" (func $forward (param i32 i32)))\n  (import "env" "cancel-w" (func $cancel-w (param i32) (result i32)))\n  (import "env" "memory"', 1)
    one = "" if future else "i32.const 1 "
    done = 0 if future else 16
    init = 'call $new local.tee $pair i32.wrap_i64 local.set $r local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w'
    edges = f'''(func (export "edges") (param $mode i32) (result i32)
     (local $r i32) (local $w i32) (local $r2 i32) (local $w2 i32) (local $pair i64) (local $set i32)
     {init}
     local.get $mode i32.const 1 i32.eq if local.get $w i32.const 64 {one}call $read drop end
     local.get $mode i32.const 2 i32.eq if local.get $r i32.const 2 {one}call $read drop end
     local.get $mode i32.const 3 i32.eq if local.get $r i32.const 65536 {one}call $read drop end
     local.get $mode i32.const 4 i32.eq if local.get $r i32.const 64 {one}call $read drop local.get $r call $drop-r end
     local.get $mode i32.const 5 i32.eq if local.get $r i32.const 64 {one}call $read drop local.get $r call $cancel drop local.get $r call $cancel drop end
     local.get $mode i32.const 6 i32.eq if local.get $w call $cancel-w drop end
     local.get $mode i32.const 8 i32.eq if
      local.get $r call $drop-r local.get $w i32.const 32 {one}call $write i32.const 1 i32.ne if unreachable end
      local.get $w call $drop-w i32.const 1 return end
     local.get $mode i32.const 9 i32.eq if local.get $r local.get $w call $forward i32.const 42 return end
     local.get $mode i32.const 11 i32.eq if local.get $r local.get $w call $forward local.get $r call $drop-r end
     local.get $mode i32.const 12 i32.eq if local.get $w i32.const 32 {one}call $write drop local.get $w call $drop-w end
     call $new local.tee $pair i32.wrap_i64 local.set $r2 local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w2
     i32.const 32 i32.const 42 i32.store
     local.get $w i32.const 32 {one}call $write i32.const -1 i32.ne if unreachable end
     local.get $r2 i32.const 64 {one}call $read i32.const -1 i32.ne if unreachable end
     local.get $r local.get $w2 call $forward
     call $set-new local.set $set local.get $w local.get $set call $join local.get $r2 local.get $set call $join
     local.get $set i32.const 96 call $wait drop i32.const 100 i32.load i32.const {done} i32.ne if unreachable end
     local.get $set i32.const 96 call $wait drop i32.const 100 i32.load i32.const {done} i32.ne if unreachable end
     local.get $w call $drop-w local.get $r2 call $drop-r local.get $set call $set-drop i32.const 64 i32.load)
'''
    text = text.replace('  (func (export "run")', edges+'  (func (export "run")',1)
    text = text.rstrip()[:-1]+' (func (export "edges") async (param "mode" u32) (result u32) (canon lift (core func $code "edges"))))'
    fixtures["async-"+kind+"-edges"] = text

unit = fixtures["async-stream-protocol"].replace('(stream u32)', '(stream)')
unit = unit.replace('(canon stream.read $t (memory $mem) async)', '(canon stream.read $t async)').replace('(canon stream.write $t (memory $mem) async)', '(canon stream.write $t async)')
# Unit streams count events without reading either pointer or allocating a Value
# per event, even at the maximum 28-bit length.
unit = unit.replace('i32.const 64 i32.const 1 call $read', 'i32.const -1 i32.const 268435455 call $read')
unit = unit.replace('i32.const 32 i32.const 1 call $write', 'i32.const 3 i32.const 268435455 call $write')
unit = unit.replace('i32.const 16 i32.ne', 'i32.const -16 i32.ne')
unit = unit.replace('i32.const 64 i32.load))', 'local.get $n))',1)
fixtures["async-stream-unit"]=unit

# Memory64 keeps full-width pointers, lengths and stream copy results.
for kind in ["stream", "future"]:
    future = kind == "future"
    wide = fixtures["async-"+kind+"-protocol"].replace('(memory (export "memory") 1)', '(memory (export "memory") i64 1)').replace('(memory 1)', '(memory i64 1)')
    old = '(param i32 i32'+('' if future else ' i32')+') (result i32)'
    new = '(param i32 i64'+('' if future else ' i64')+') (result '+('i32' if future else 'i64')+')'
    wide = wide.replace('(func $read '+old+')', '(func $read '+new+')').replace('(func $write '+old+')', '(func $write '+new+')')
    wide = wide.replace('(func $wait (param i32 i32)', '(func $wait (param i32 i64)')
    wide = wide.replace('i32.const 64 '+('' if future else 'i32.const 1 ')+'call $read', 'i64.const 64 '+('' if future else 'i64.const 1 ')+'call $read')
    wide = wide.replace('i32.const 32 '+('' if future else 'i32.const 1 ')+'call $write', 'i64.const 32 '+('' if future else 'i64.const 1 ')+'call $write')
    if not future:
        for name, want in [('read', -1), ('write', 16)]:
            wide = wide.replace(f'call ${name} i32.const {want} i32.ne', f'call ${name} i64.const {want} i64.ne')
    wide = wide.replace('i32.const 32 local.get $n i32.store', 'i64.const 32 local.get $n i32.store')
    wide = wide.replace('i32.const 96 call $wait', 'i64.const 96 call $wait').replace('i32.const 100 i32.load', 'i64.const 100 i32.load').replace('i32.const 64 i32.load', 'i64.const 64 i32.load')
    fixtures["async-"+kind+"-memory64"]=wide

zero = fixtures["async-stream-protocol"]
code = '''(func (export "zero") (param $mode i32) (result i32) (local $pair i64) (local $r i32) (local $w i32) (local $set i32) (local $copied i32)
 call $new local.tee $pair i32.wrap_i64 local.set $r local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
 local.get $mode i32.const 1 i32.eq if
  local.get $w i32.const -1 i32.const 0 call $write i32.const -1 i32.ne if unreachable end
  local.get $r i32.const 3 i32.const 0 call $read i32.const -1 i32.ne if unreachable end
  call $set-new local.set $set local.get $w local.get $set call $join
  local.get $set i32.const 96 call $wait drop i32.const 100 i32.load if unreachable end
  local.get $w i32.const 0 call $join local.get $set call $set-drop
 else
  local.get $r local.get $mode i32.const 3 i32.eq if (result i32) i32.const 64 else i32.const -1 end local.get $mode i32.const 3 i32.eq if (result i32) i32.const 1 else i32.const 0 end call $read i32.const -1 i32.ne if unreachable end
  local.get $w i32.const 32 local.get $mode i32.const 2 i32.eq call $write local.set $copied
  local.get $mode i32.const 2 i32.eq if
   local.get $copied i32.const -1 i32.ne if unreachable end
   call $set-new local.set $set local.get $r local.get $set call $join
   local.get $set i32.const 96 call $wait drop i32.const 100 i32.load if unreachable end
   local.get $r i32.const 0 call $join local.get $set call $set-drop
   local.get $w call $cancel-w i32.const 2 i32.ne if unreachable end
   local.get $w call $drop-w local.get $r call $drop-r i32.const 0 return
  else local.get $copied if unreachable end end
 end
 local.get $r call $cancel i32.const 2 i32.ne if unreachable end
 local.get $w call $drop-w local.get $r call $drop-r i32.const 0)
'''
zero = zero.replace(' (core func $drop-r ', ' (core func $cancel-w (canon stream.cancel-write $t async))\n (core func $drop-r ',1)
zero = zero.replace('(core instance $env ', '(core instance $env (export "cancel-w" (func $cancel-w)) ',1)
zero = zero.replace('  (import "env" "memory"', '  (import "env" "cancel-w" (func $cancel-w (param i32) (result i32)))\n  (import "env" "memory"',1)
zero = zero.replace('  (func (export "run")',code+'  (func (export "run")',1)
zero = zero.rstrip()[:-1]+' (func (export "zero") async (param "mode" u32) (result u32) (canon lift (core func $code "zero"))))'
fixtures["async-stream-zero"]=zero

for kind in ["stream", "future"]:
    one = "" if kind == "future" else "i32.const 1 "
    fixtures["async-"+kind+"-exports"] = f'''(component
 (type $t ({kind} u32)) (export "endpoint-type" (type $t))
 (import "keep" (func $keep (param "end" $t))) (import "drop" (func $drop (param "end" $t)))
 (core module $memory (memory (export "memory") 1)) (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon {kind}.new $t)) (core func $write (canon {kind}.write $t (memory $mem) async))
 (core func $drop-w (canon {kind}.drop-writable $t))
 (core func $keep (canon lower (func $keep))) (core func $drop (canon lower (func $drop)))
 (core instance $env (export "new" (func $new)) (export "write" (func $write)) (export "drop-w" (func $drop-w))
  (export "keep" (func $keep)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "new" (func $new (result i64))) (import "env" "write" (func $write (param i32 i32{'' if kind=='future' else ' i32'}) (result i32)))
  (import "env" "drop-w" (func $drop-w (param i32))) (import "env" "keep" (func $keep (param i32))) (import "env" "drop" (func $drop (param i32)))
  (global $w (mut i32) (i32.const 0))
  (func $make (export "make") (result i32) (local i64)
   call $new local.tee 0 i64.const 32 i64.shr_u i32.wrap_i64 global.set $w local.get 0 i32.wrap_i64)
  (func (export "write") (result i32) (local i32)
   global.get $w i32.const 32 {one}call $write local.set 0 global.get $w call $drop-w local.get 0)
  (func (export "drop") (param i32) local.get 0 call $drop)
  (func (export "leak") call $make call $keep unreachable))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "make") (result $t) (canon lift (core func $code "make")))
 (func (export "write") (result u32) (canon lift (core func $code "write")))
 (func (export "drop") (param "end" $t) (canon lift (core func $code "drop")))
 (func (export "leak") (canon lift (core func $code "leak"))))'''

fixtures["async-future-generative"] = '''(component
 (component $child
  (type $r (resource (rep i32))) (export "r" (type $r))
  (type $t (future (own 1))) (export "endpoint-type" (type $t))
  (core func $new (canon future.new $t)) (core func $drop (canon future.drop-readable $t))
  (core instance $env (export "new" (func $new)) (export "drop" (func $drop)))
  (core module $code
   (import "env" "new" (func $new (result i64))) (import "env" "drop" (func $drop (param i32)))
   (func (export "make") (result i32) call $new i32.wrap_i64)
   (func (export "take") (param i32) local.get 0 call $drop))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "make") (result $t) (canon lift (core func $code "make")))
  (func (export "take") (param "end" $t) (canon lift (core func $code "take"))))
 (instance $one (instantiate $child)) (instance $two (instantiate $child))
 (export "one" (instance $one)) (export "two" (instance $two)))'''

fixtures["async-stream-bytes"] = fixtures["async-stream-protocol"].replace("(stream u32)", "(stream u8)")

for kind in ["stream", "future"]:
    future = kind == "future"
    one = "" if future else "i32.const 1 "
    done = 0 if future else 17
    text = fixtures["async-"+kind+"-strings"]
    # The callee parks before reading. The caller's synchronous write must park
    # too, then resume after the callee supplies its own buffer.
    text = text.replace(f'(canon {kind}.read $t (memory $mem) (realloc $realloc) async)', f'(canon {kind}.read $t (memory $mem) (realloc $realloc))',1)
    text = text.replace(f'(canon {kind}.write $t (memory $mem) async)', f'(canon {kind}.write $t (memory $mem))',1)
    text = text.replace('  (core func $read ', '  (core func $yield (canon thread.yield))\n  (core func $read ',1)
    text = text.replace('  (core instance $env ', '  (core instance $env (export "yield" (func $yield)) ',1)
    text = text.replace('   (import "env" "read"', '   (import "env" "yield" (func $yield (result i32)))\n   (import "env" "read"',1)
    text = text.replace('    local.get 0 i32.const 64 ', '    call $yield drop local.get 0 i32.const 64 ',1)
    text = text.replace(' (core func $join ', ' (core func $subdrop (canon subtask.drop))\n (core func $join ',1)
    text = text.replace(' (core instance $env (export "memory"', ' (core instance $env (export "subdrop" (func $subdrop)) (export "memory"',1)
    text = text.replace('  (import "env" "memory"', '  (import "env" "subdrop" (func $subdrop (param i32)))\n  (import "env" "memory"',1)
    start = text.index('  (func (export "run")')
    end = text.index('\n (core instance $code',start)
    text = text[:start]+f'''  (func (export "run") (result i32) (local $r i32) (local $w i32) (local $pair i64) (local $set i32) (local $sub i32)
     call $new local.tee $pair i32.wrap_i64 local.set $r local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
     i32.const 32 i32.const 256 i32.store i32.const 36 i32.const 8 i32.store
     local.get $r call $pass i32.const 96 call $consume local.tee $sub i32.const 15 i32.and i32.const 1 i32.ne if unreachable end
     local.get $sub i32.const 4 i32.shr_u local.set $sub
     local.get $w i32.const 32 {one}call $write i32.const {done} i32.ne if unreachable end
     call $set-new local.set $set local.get $sub local.get $set call $join
     local.get $set i32.const 128 call $wait i32.const 1 i32.ne if unreachable end
     i32.const 132 i32.load i32.const 2 i32.ne if unreachable end
     local.get $sub call $subdrop local.get $w call $drop local.get $set call $set-drop i32.const 96))'''+text[end:]
    fixtures["async-"+kind+"-sync-copy"]=text

def uleb(value):
    out = bytearray()
    while True:
        byte = value & 127
        value >>= 7
        out.append(byte | (128 if value else 0))
        if not value:
            return bytes(out)

def read_uleb(data, at):
    value = shift = 0
    while True:
        byte = data[at]
        at += 1
        value |= (byte & 127) << shift
        if byte < 128:
            return value, at
        shift += 7

def forward_patch(data, future):
    # The pinned producer lacks forward syntax/opcodes. Its validated placeholder
    # waitable.join has the identical Core signature; only this canonical entry
    # is authored from the pinned binary grammar, not claimed producer validation.
    out = data[:8]
    at = 8
    patched = False
    while at < len(data):
        section = data[at]
        length, body = read_uleb(data, at+1)
        payload = data[body:body+length]
        if not patched and section == 8:
            count, cursor = read_uleb(payload, 0)
            if count >= 5:
                for _ in range(4):
                    op = payload[cursor]
                    cursor += 1
                    _, cursor = read_uleb(payload, cursor)
                    if op in (0x0f, 0x10, 0x16, 0x17):
                        options, cursor = read_uleb(payload, cursor)
                        for _ in range(options):
                            option = payload[cursor]
                            cursor += 1
                            if option in (3, 4, 5, 7):
                                _, cursor = read_uleb(payload, cursor)
                    elif op in (0x11, 0x18):
                        cursor += 1
                    else:
                        assert op in (0x0e, 0x15), op
                assert payload[cursor] == 0x23
                payload = payload[:cursor]+bytes([0x2f if future else 0x2e, 0])+payload[cursor+1:]
                patched = True
        out += bytes([section])+uleb(len(payload))+payload
        at = body+length
    assert patched
    return out

pins = {}
producer_rejections = {}
for name, text in fixtures.items():
    wat = OUT / (name + ".wat")
    wat.write_text(text + "\n")
    authored = name.endswith("-edges")
    if authored:
        placeholder = OUT / (name+"-producer.wat")
        placeholder.write_text(text.replace("(canon stream.forward $t)", "(canon waitable.join)").replace("(canon future.forward $t)", "(canon waitable.join)")+"\n")
        subprocess.run(["wasm-tools", "parse", str(placeholder), "-o", str(wat.with_suffix(".wasm"))], check=True)
        placeholder.unlink()
        subprocess.run(["wasm-tools", "validate", "--features", "all", str(wat.with_suffix(".wasm"))], check=True)
        wat.with_suffix(".wasm").write_bytes(forward_patch(wat.with_suffix(".wasm").read_bytes(), "future" in name))
    else:
        subprocess.run(["wasm-tools", "parse", str(wat), "-o", str(wat.with_suffix(".wasm"))], check=True)
        result = subprocess.run(["wasm-tools", "validate", "--features", "all", str(wat.with_suffix(".wasm"))], capture_output=True, text=True)
        if name.endswith("-memory64"):
            assert result.returncode != 0 and 'type mismatch for export `read`' in result.stderr, result.stderr
            producer_rejections[name] = result.stderr.strip()
        else:
            assert result.returncode == 0, result.stderr
    pins[name] = {ext: hashlib.sha256((OUT / (name + "." + ext)).read_bytes()).hexdigest() for ext in ["wat", "wasm"]}
(OUT / "future-streams.json").write_text(json.dumps({"componentModelCommit": "d095e2f60ddae1f9a1e9b1e2fb547ec1114da677", "producer": "wasm-tools 1.258.0", "authoredForward": [name for name in fixtures if name.endswith("-edges")], "producerRejections": producer_rejections, "generatorSHA256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), "sha256": pins}, indent=2) + "\n")
print("Built", len(pins)-4, "producer-validated guests, 2 authored forwards and 2 recorded memory64 producer rejections")

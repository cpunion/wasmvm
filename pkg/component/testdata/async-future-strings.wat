(component
 (type $t (future string)) (export "endpoint-type" (type $t))
 (import "pass" (func $pass (param "end" $t) (result $t)))
 (component $consumer
  (type $t (future string)) (export "endpoint-type" (type $t))
  (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
   (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
    global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
  (core func $read (canon future.read $t (memory $mem) (realloc $realloc) async))
  (core func $drop (canon future.drop-readable $t))
  (core instance $env (export "read" (func $read)) (export "drop" (func $drop)))
  (core module $code
   (import "env" "read" (func $read (param i32 i32) (result i32)))
   (import "env" "drop" (func $drop (param i32)))
   (func (export "consume") (param i32) (result i32)
    local.get 0 i32.const 64 call $read i32.const 0 i32.ne if unreachable end
    local.get 0 call $drop i32.const 64))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "consume") async (param "end" $t) (result string) (canon lift (core func $code "consume") (memory $mem))))
 (instance $consumer (instantiate $consumer)) (alias export $consumer "consume" (func $consume))
 (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
  (data (i32.const 256) "caf\c3\a9\e9\9b\aa")
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
 (core func $pass (canon lower (func $pass)))
 (core func $consume (canon lower (func $consume) (memory $mem) (realloc $realloc) async))
 (core func $new (canon future.new $t)) (core func $write (canon future.write $t (memory $mem) async))
 (core func $drop (canon future.drop-writable $t))
 (core func $join (canon waitable.join)) (core func $set-new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem))) (core func $set-drop (canon waitable-set.drop))
 (core instance $env (export "memory" (memory $mem)) (export "pass" (func $pass)) (export "consume" (func $consume))
  (export "new" (func $new)) (export "write" (func $write)) (export "drop" (func $drop))
  (export "join" (func $join)) (export "set-new" (func $set-new)) (export "wait" (func $wait)) (export "set-drop" (func $set-drop)))
 (core module $code
  (import "env" "memory" (memory 1)) (import "env" "pass" (func $pass (param i32) (result i32)))
  (import "env" "consume" (func $consume (param i32 i32) (result i32)))
  (import "env" "new" (func $new (result i64)))
  (import "env" "write" (func $write (param i32 i32) (result i32)))
  (import "env" "drop" (func $drop (param i32))) (import "env" "join" (func $join (param i32 i32)))
  (import "env" "set-new" (func $set-new (result i32))) (import "env" "wait" (func $wait (param i32 i32) (result i32)))
  (import "env" "set-drop" (func $set-drop (param i32)))
  (func (export "run") (result i32) (local $r i32) (local $w i32) (local $pair i64) (local $set i32)
   call $new local.tee $pair i32.wrap_i64 local.set $r local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   i32.const 32 i32.const 256 i32.store i32.const 36 i32.const 8 i32.store
   local.get $w i32.const 32 call $write i32.const -1 i32.ne if unreachable end
   local.get $r call $pass i32.const 96 call $consume i32.const 2 i32.ne if unreachable end
   call $set-new local.set $set local.get $w local.get $set call $join
   local.get $set i32.const 128 call $wait i32.const 5 i32.ne if unreachable end
   i32.const 132 i32.load i32.const 0 i32.ne if unreachable end
   local.get $w call $drop local.get $set call $set-drop i32.const 96))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result string) (canon lift (core func $code "run") (memory $mem))))

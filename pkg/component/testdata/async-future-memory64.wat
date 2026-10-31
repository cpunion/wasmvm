(component
 (type $t (future u32))
 (core module $memory (memory (export "memory") i64 1))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon future.new $t))
 (core func $read (canon future.read $t (memory $mem) async))
 (core func $write (canon future.write $t (memory $mem) async))
 (core func $cancel (canon future.cancel-read $t async))
 (core func $drop-r (canon future.drop-readable $t))
 (core func $drop-w (canon future.drop-writable $t))
 (core func $join (canon waitable.join))
 (core func $set-new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $set-drop (canon waitable-set.drop))
 (core instance $env (export "memory" (memory $mem))
  (export "new" (func $new)) (export "read" (func $read)) (export "write" (func $write))
  (export "cancel" (func $cancel)) (export "drop-r" (func $drop-r)) (export "drop-w" (func $drop-w))
  (export "join" (func $join)) (export "set-new" (func $set-new)) (export "wait" (func $wait)) (export "set-drop" (func $set-drop)))
 (core module $code
  (import "env" "memory" (memory i64 1))
  (import "env" "new" (func $new (result i64)))
  (import "env" "read" (func $read (param i32 i64) (result i32)))
  (import "env" "write" (func $write (param i32 i64) (result i32)))
  (import "env" "cancel" (func $cancel (param i32) (result i32)))
  (import "env" "drop-r" (func $drop-r (param i32))) (import "env" "drop-w" (func $drop-w (param i32)))
  (import "env" "join" (func $join (param i32 i32)))
  (import "env" "set-new" (func $set-new (result i32)))
  (import "env" "wait" (func $wait (param i32 i64) (result i32)))
  (import "env" "set-drop" (func $set-drop (param i32)))
  (func (export "run") (param $n i32) (result i32) (local $r i32) (local $w i32) (local $pair i64) (local $set i32)
   call $new local.tee $pair i32.wrap_i64 local.set $r
   local.get $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $r i64.const 64 call $read i32.const -1 i32.ne if unreachable end
   local.get $r call $cancel i32.const 2 i32.ne if unreachable end
   local.get $r i64.const 64 call $read i32.const -1 i32.ne if unreachable end
   call $set-new local.set $set local.get $r local.get $set call $join
   i64.const 32 local.get $n i32.store
   local.get $w i64.const 32 call $write i32.const 0 i32.ne if unreachable end
   local.get $set i64.const 96 call $wait i32.const 4 i32.ne if unreachable end
   i64.const 100 i32.load i32.const 0 i32.ne if unreachable end
   local.get $w call $drop-w local.get $r call $drop-r local.get $set call $set-drop
   i64.const 64 i32.load))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (param "n" u32) (result u32) (canon lift (core func $code "run"))))

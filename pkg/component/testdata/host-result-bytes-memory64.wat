(component
 (type $t (stream u8)) (export "endpoint-type" (type $t))
 (core module $memory (memory (export "memory") i64 1))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon stream.new $t))
 (core func $return (canon task.return (result $t) (memory $mem)))
 (core func $write (canon stream.write $t (memory $mem)))
 (core func $drop (canon stream.drop-writable $t))
 (core instance $env (export "memory" (memory $mem)) (export "new" (func $new))
  (export "return" (func $return)) (export "write" (func $write)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "memory" (memory i64 1))
  (import "env" "new" (func $new (result i64)))
  (import "env" "return" (func $return (param i32)))
  (import "env" "write" (func $write (param i32 i64 i64) (result i64)))
  (import "env" "drop" (func $drop (param i32)))
  (data (i64.const 32) "\80\ff")
  (func (export "run") (local $pair i64) (local $w i32) (local $status i64)
   call $new local.tee $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $pair i32.wrap_i64 call $return
   local.get $w i64.const 32 i64.const 1 call $write local.tee $status i64.const 15 i64.and i64.const 1 i64.eq
   if local.get $w call $drop return end
   local.get $status i64.const 16 i64.ne if unreachable end
   local.get $w i64.const 33 i64.const 1 call $write local.tee $status i64.const 16 i64.ne
   if local.get $status i64.const 15 i64.and i64.const 1 i64.ne if unreachable end end
   local.get $w call $drop))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result $t) (canon lift (core func $code "run") (memory $mem) async)))

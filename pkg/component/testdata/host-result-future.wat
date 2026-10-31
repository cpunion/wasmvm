(component
 (type $t (future u32)) (export "endpoint-type" (type $t))
 (core module $memory (memory (export "memory") 1)) (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $new (canon future.new $t))
 (core func $return (canon task.return (result $t) (memory $mem)))
 (core func $write (canon future.write $t (memory $mem)))
 (core func $drop (canon future.drop-writable $t))
 (core instance $env (export "memory" (memory $mem)) (export "new" (func $new)) (export "return" (func $return))
  (export "write" (func $write)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "memory" (memory 1)) (import "env" "new" (func $new (result i64)))
  (import "env" "return" (func $return (param i32)))
  (import "env" "write" (func $write (param i32 i32) (result i32)))
  (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (local $pair i64) (local $w i32) (local $status i32)
   call $new local.tee $pair i64.const 32 i64.shr_u i32.wrap_i64 local.set $w
   local.get $pair i32.wrap_i64 call $return
   i32.const 32 i32.const 42 i32.store
   local.get $w i32.const 32 call $write
   local.tee $status i32.const 0 i32.ne if local.get $status i32.const 1 i32.ne if unreachable end end local.get $w call $drop))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result $t) (canon lift (core func $code "run") (memory $mem) async)))

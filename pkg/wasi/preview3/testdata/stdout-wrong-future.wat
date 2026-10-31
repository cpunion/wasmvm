(component
 (import "wasi:cli/stdout@0.3.2" (instance $output
 (type $e (enum "io" "illegal-byte-sequence" "pipe")) (export "error-code" (type $err (eq $e)))
 (type $s (stream u8)) (type $f (future u32))
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
 (func (export "run") async (param "mode" u32) (result u32) (canon lift (core func $m "run"))))

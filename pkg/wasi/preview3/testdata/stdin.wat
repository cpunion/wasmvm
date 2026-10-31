(component
 (import "wasi:cli/stdin@0.3.2" (instance $stdin
 (type $e (enum "io" "illegal-byte-sequence" "pipe")) (export "error-code" (type $err (eq $e)))
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
 (func (export "run") async (param "drop" u32) (result u32) (canon lift (core func $m "run"))))

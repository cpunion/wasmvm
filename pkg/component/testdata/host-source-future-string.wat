(component
 (type $t (future string)) (export "endpoint-type" (type $t))
 (import "make" (func $make (result $t))) (import "started" (func $started (result u32)))
 (core module $memory (memory (export "memory") 1) (global $heap (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
 (core func $make (canon lower (func $make))) (core func $started (canon lower (func $started)))
 (core func $read (canon future.read $t (memory $mem) (realloc $realloc))) (core func $read-async (canon future.read $t (memory $mem) (realloc $realloc) async))
 (core func $cancel (canon future.cancel-read $t)) (core func $drop (canon future.drop-readable $t))
 (core func $yield (canon thread.yield))
 (core instance $env (export "memory" (memory $mem)) (export "make" (func $make)) (export "read" (func $read)) (export "read-async" (func $read-async))
  (export "cancel" (func $cancel)) (export "drop" (func $drop)) (export "started" (func $started)) (export "yield" (func $yield)))
 (core module $code
  (import "env" "memory" (memory 1))
  (import "env" "make" (func $make (result i32)))
  (import "env" "read" (func $read (param i32 i32) (result i32)))
  (import "env" "read-async" (func $read-async (param i32 i32) (result i32)))
  (import "env" "cancel" (func $cancel (param i32) (result i32))) (import "env" "drop" (func $drop (param i32)))
  (import "env" "yield" (func $yield (result i32))) (import "env" "started" (func $started (result i32)))
  (func (export "run") (param $mode i32) (result i32) (local $r i32)
   call $make local.set $r
   local.get $mode if
    local.get $r i32.const 64 call $read-async i32.const -1 i32.ne if unreachable end
    local.get $mode i32.const 3 i32.ge_u if
     loop $wait call $started i32.eqz if call $yield drop br $wait end end
     local.get $mode i32.const 4 i32.eq if unreachable end
    end
    local.get $r call $cancel i32.const 2 i32.ne if unreachable end
   else
    local.get $r i32.const 64 call $read i32.const 0 i32.ne if unreachable end
   end
   local.get $r call $drop
   i32.const 64))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (param "mode" u32) (result string) (canon lift (core func $code "run") (memory $mem))))

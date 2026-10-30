(component
 (type $bytes (list u8))
 (type $echo-type (func (param "bytes" $bytes) (result $bytes)))
 (import "echo" (func $host (type $echo-type)))
 (core module $m
  (memory (export "memory") i64 2)
  (global $next (mut i64) (i64.const 256))
  (func (export "reset") (param i64) i64.const 256 global.set $next)
  (func (export "realloc") (param i64 i64 i64 i64) (result i64)
   (local $p i64)
   global.get $next local.get 2 i64.const 1 i64.sub i64.add
   i64.const 0 local.get 2 i64.sub i64.and local.tee $p
   local.get 3 i64.add global.set $next local.get $p)
  (func (export "echo") (param i64 i64) (result i64)
   i64.const 0 local.get 0 i64.store
   i64.const 8 local.get 1 i64.store
   i64.const 0))
 (core instance $m (instantiate $m))
 (alias core export $m "memory" (core memory $memory))
 (alias core export $m "realloc" (core func $realloc))
 (alias core export $m "reset" (core func $reset))
 (alias core export $m "echo" (core func $echo))
 (core func $lower (canon lower (func $host) (memory $memory) (realloc $realloc)))
 (core module $bounce
  (import "env" "echo" (func $echo (param i64 i64 i64)))
  (func (export "bounce") (param i64 i64) (result i64)
   local.get 0 local.get 1 i64.const 0 call $echo i64.const 0))
 (core instance $env (export "echo" (func $lower)))
 (core instance $bounce (instantiate $bounce (with "env" (instance $env))))
 (alias core export $bounce "bounce" (core func $bounce))
 (func (export "echo") (param "bytes" $bytes) (result $bytes)
  (canon lift (core func $bounce) (memory $memory) (realloc $realloc) (post-return $reset))))

(component
 (type $bytes (list u8))
 (core module $m
  (memory (export "memory") 2)
  (global $next (mut i32) (i32.const 256))
  (func (export "reset") (param i32) i32.const 256 global.set $next)
  (func (export "realloc") (param i32 i32 i32 i32) (result i32)
   (local $p i32)
   global.get $next local.get 2 i32.const 1 i32.sub i32.add
   i32.const 0 local.get 2 i32.sub i32.and local.tee $p
   local.get 3 i32.add global.set $next local.get $p)
  (func (export "echo") (param i32 i32) (result i32)
   i32.const 0 local.get 0 i32.store
   i32.const 4 local.get 1 i32.store
   i32.const 0))
 (core instance $m (instantiate $m))
 (alias core export $m "memory" (core memory $memory))
 (alias core export $m "realloc" (core func $realloc))
 (alias core export $m "reset" (core func $reset))
 (alias core export $m "echo" (core func $echo))
 (func (export "echo") (param "bytes" $bytes) (result $bytes)
  (canon lift (core func $echo) (memory $memory) (realloc $realloc) (post-return $reset))))

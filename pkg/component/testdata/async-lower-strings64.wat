(component
 (component $server
   (core module $memory
  (memory (export "memory") i64 1)
  (memory (export "other") i64 1)
  (data (i64.const 0) "diagnostic-\ce\bb")
  (global $bump (mut i64) (i64.const 1024))
  (func (export "realloc") (param $old i64) (param $old-size i64) (param $alignment i64) (param $size i64) (result i64) (local $p i64)
   global.get $bump local.tee $p local.get $size i64.add i64.const 7 i64.add i64.const -8 i64.and global.set $bump
   local.get $p local.get $old local.get $old-size local.get $size local.get $old-size local.get $size i64.lt_u select memory.copy
   local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "memory" (core memory $alias))
 (alias core export $memory "other" (core memory $other))
 (alias core export $memory "realloc" (core func $realloc))

  (core func $ret (canon task.return (result string) (memory $mem)))
  (core instance $env (export "ret" (func $ret)))
  (core module $code (import "env" "ret" (func $ret (param i64 i64))) (func (export "echo") (param i64 i64) local.get 0 local.get 1 call $ret))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") async (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) (realloc $realloc) async)))
 (instance $server (instantiate $server))
 (alias export $server "echo" (func $echo))
 (component $client
  (import "echo" (func $echo async (param "text" string) (result string)))
   (core module $memory
  (memory (export "memory") i64 1)
  (memory (export "other") i64 1)
  (data (i64.const 0) "diagnostic-\ce\bb")
  (global $bump (mut i64) (i64.const 1024))
  (func (export "realloc") (param $old i64) (param $old-size i64) (param $alignment i64) (param $size i64) (result i64) (local $p i64)
   global.get $bump local.tee $p local.get $size i64.add i64.const 7 i64.add i64.const -8 i64.and global.set $bump
   local.get $p local.get $old local.get $old-size local.get $size local.get $old-size local.get $size i64.lt_u select memory.copy
   local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "memory" (core memory $alias))
 (alias core export $memory "other" (core memory $other))
 (alias core export $memory "realloc" (core func $realloc))

  (core func $echo (canon lower (func $echo) (memory $mem) (realloc $realloc) async))
  (core instance $env (export "echo" (func $echo)))
  (core module $code
   (import "env" "echo" (func $echo (param i64 i64 i64) (result i32)))
   (func (export "echo") (param i64 i64) (result i64) local.get 0 local.get 1 i64.const 32 call $echo i32.const 2 i32.ne if unreachable end i64.const 32))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) (realloc $realloc))))
 (instance $client (instantiate $client (with "echo" (func $echo))))
 (alias export $client "echo" (func $call))
 (export "echo" (func $call)))

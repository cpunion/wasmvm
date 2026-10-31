(component
 (component $server
   (core module $memory
  (memory (export "memory") 1)
  (memory (export "other") 1)
  (data (i32.const 0) "diagnostic-\ce\bb")
  (global $bump (mut i32) (i32.const 1024))
  (func (export "realloc") (param $old i32) (param $old-size i32) (param $alignment i32) (param $size i32) (result i32) (local $p i32)
   global.get $bump local.tee $p local.get $size i32.add i32.const 7 i32.add i32.const -8 i32.and global.set $bump
   local.get $p local.get $old local.get $old-size local.get $size local.get $old-size local.get $size i32.lt_u select memory.copy
   local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "memory" (core memory $alias))
 (alias core export $memory "other" (core memory $other))
 (alias core export $memory "realloc" (core func $realloc))

  (core func $ret (canon task.return (result string) (memory $mem) string-encoding=latin1+utf16))
  (core instance $env (export "ret" (func $ret)))
  (core module $code (import "env" "ret" (func $ret (param i32 i32))) (func (export "echo") (param i32 i32) local.get 0 local.get 1 call $ret))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") async (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) string-encoding=latin1+utf16 (realloc $realloc) async)))
 (instance $server (instantiate $server))
 (alias export $server "echo" (func $echo))
 (component $client
  (import "echo" (func $echo async (param "text" string) (result string)))
   (core module $memory
  (memory (export "memory") 1)
  (memory (export "other") 1)
  (data (i32.const 0) "diagnostic-\ce\bb")
  (global $bump (mut i32) (i32.const 1024))
  (func (export "realloc") (param $old i32) (param $old-size i32) (param $alignment i32) (param $size i32) (result i32) (local $p i32)
   global.get $bump local.tee $p local.get $size i32.add i32.const 7 i32.add i32.const -8 i32.and global.set $bump
   local.get $p local.get $old local.get $old-size local.get $size local.get $old-size local.get $size i32.lt_u select memory.copy
   local.get $p))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "memory" (core memory $alias))
 (alias core export $memory "other" (core memory $other))
 (alias core export $memory "realloc" (core func $realloc))

  (core func $echo (canon lower (func $echo) (memory $mem) string-encoding=latin1+utf16 (realloc $realloc) async))
  (core instance $env (export "echo" (func $echo)))
  (core module $code
   (import "env" "echo" (func $echo (param i32 i32 i32) (result i32)))
   (func (export "echo") (param i32 i32) (result i32) local.get 0 local.get 1 i32.const 32 call $echo i32.const 2 i32.ne if unreachable end i32.const 32))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) string-encoding=latin1+utf16 (realloc $realloc))))
 (instance $client (instantiate $client (with "echo" (func $echo))))
 (alias export $client "echo" (func $call))
 (export "echo" (func $call)))

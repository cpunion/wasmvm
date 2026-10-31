(component
 (import "r" (type $r (sub resource)))
 (type $output (tuple (own $r) u32))
 (export "output" (type $output))
 (import "make" (func $make async (param "arg" (borrow $r)) (result 2)))
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

 (core func $make (canon lower (func $make) (memory $mem)))
 (core func $drop (canon resource.drop $r))
 (core instance $env (export "make" (func $make)) (export "drop" (func $drop)))
 (core module $code
  (import "env" "make" (func $make (param i32 i32)))
  (import "env" "drop" (func $drop (param i32)))
  (func (export "call") (param i32 i32) (result i32)
   local.get 0 local.get 1 call $make
   local.get 0 call $drop local.get 1))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "call") (param "arg" (borrow $r)) (param "ptr" u32) (result 2) (canon lift (core func $code "call") (memory $mem)))
 (export "direct" (func $make)))

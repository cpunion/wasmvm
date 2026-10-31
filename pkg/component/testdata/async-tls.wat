(component
 (import "text" (func $text (result string)))
 (core func $get (canon context.get i32 0))
 (core func $set (canon context.set i32 0))
 (core instance $ctx (export "get" (func $get)) (export "set" (func $set)))
 (core module $memory
  (import "ctx" "get" (func $get (result i32)))
  (import "ctx" "set" (func $set (param i32)))
  (memory (export "memory") 1)
  (global $bump (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32) (param $size i32) (result i32) (local $p i32)
   call $get if unreachable end i32.const 99 call $set
   global.get $bump local.tee $p local.get $size i32.add global.set $bump local.get $p))
 (core instance $memory (instantiate $memory (with "ctx" (instance $ctx))))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "realloc" (core func $realloc))
 (core func $text (canon lower (func $text) (memory $mem) (realloc $realloc)))
 (core func $ret (canon task.return (result string) (memory $mem)))
 (core instance $env (export "get" (func $get)) (export "set" (func $set))
  (export "text" (func $text)) (export "ret" (func $ret)) (export "memory" (memory $mem)))
 (core module $code
  (import "env" "get" (func $get (result i32)))
  (import "env" "set" (func $set (param i32)))
  (import "env" "text" (func $text (param i32)))
  (import "env" "ret" (func $ret (param i32 i32)))
  (import "env" "memory" (memory 1))
  (func (export "run")
   call $get if unreachable end i32.const 7 call $set
   i32.const 32 call $text
   call $get i32.const 7 i32.ne if unreachable end
   i32.const 32 i32.load i32.const 36 i32.load call $ret))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "run") async (result string) (canon lift (core func $code "run") (memory $mem) async)))

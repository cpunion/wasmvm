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
  (core func $get0 (canon context.get i32 0))
  (core func $get1 (canon context.get i32 1))
  (core func $set0 (canon context.set i32 0))
  (core func $set1 (canon context.set i32 1))
  (core instance $env (export "ret" (func $ret)) (export "get0" (func $get0)) (export "get1" (func $get1)) (export "set0" (func $set0)) (export "set1" (func $set1)))
  (core module $code (import "env" "ret" (func $ret (param i32 i32)))
   (import "env" "get0" (func $get0 (result i32))) (import "env" "get1" (func $get1 (result i32)))
   (import "env" "set0" (func $set0 (param i32))) (import "env" "set1" (func $set1 (param i32)))
   (func (export "echo") (param i32 i32) (result i32) local.get 0 call $set0 local.get 1 call $set1 i32.const 1)
   (func (export "callback") (param i32 i32 i32) (result i32) call $get0 call $get1 call $ret i32.const 0))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (alias core export $code "callback" (core func $callback))
  (func (export "echo") async (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) string-encoding=latin1+utf16 (realloc $realloc) async (callback $callback))))
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
(core func $new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $poll (canon waitable-set.poll (memory $mem)))
 (core func $drop (canon waitable-set.drop))
 (core func $join (canon waitable.join))
 (core func $subdrop (canon subtask.drop))
 (core func $cancel (canon subtask.cancel))
 (core func $cancel-async (canon subtask.cancel async))
 (core instance $env (export "memory" (memory $mem)) (export "echo" (func $echo)) (export "new" (func $new)) (export "wait" (func $wait)) (export "poll" (func $poll)) (export "drop" (func $drop)) (export "join" (func $join)) (export "subdrop" (func $subdrop)) (export "cancel" (func $cancel)) (export "cancel-async" (func $cancel-async)))
  (core module $code
   (import "env" "memory" (memory 1))
 (import "env" "echo" (func $echo (param i32 i32 i32) (result i32)))
 (import "env" "new" (func $new (result i32)))
 (import "env" "wait" (func $wait (param i32 i32) (result i32)))
 (import "env" "poll" (func $poll (param i32 i32) (result i32)))
 (import "env" "drop" (func $drop (param i32)))
 (import "env" "join" (func $join (param i32 i32)))
 (import "env" "subdrop" (func $subdrop (param i32)))
 (import "env" "cancel" (func $cancel (param i32) (result i32)))
 (import "env" "cancel-async" (func $cancel-async (param i32) (result i32)))
   (func (export "echo") (param i32 i32) (result i32) (local $sub i32) (local $set i32)
    local.get 0 local.get 1 i32.const 32 call $echo i32.const 4 i32.shr_u local.set $sub
    call $new local.set $set local.get $sub local.get $set call $join
    local.get $set i32.const 64 call $wait i32.const 1 i32.ne if unreachable end
    i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
    local.get $sub call $subdrop local.get $set call $drop i32.const 32))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "echo") async (param "text" string) (result string) (canon lift (core func $code "echo") (memory $mem) string-encoding=latin1+utf16 (realloc $realloc))))
 (instance $client (instantiate $client (with "echo" (func $echo))))
 (alias export $client "echo" (func $call))
 (export "echo" (func $call)))

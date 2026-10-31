(component
 (import "r" (type $r (sub resource)))
 (type $out (tuple (own $r) u32)) (export "out" (type $out))
 (import "make" (func $make async (param "arg" (borrow $r)) (result 2)))
 (component $server
  (import "r" (type $r (sub resource)))
  (type $out (tuple (own $r) u32)) (export "out" (type $out))
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

  (core func $ret (canon task.return (result $out) (memory $mem)))
  (core func $get (canon context.get i32 0)) (core func $set (canon context.set i32 0))
  (core func $drop (canon resource.drop $r))
  (core func $make (canon lower (func $make) (memory $mem)))
  (core instance $env (export "memory" (memory $mem)) (export "ret" (func $ret)) (export "get" (func $get)) (export "set" (func $set)) (export "drop" (func $drop)) (export "make" (func $make)))
  (core module $code
   (import "env" "memory" (memory 1)) (import "env" "ret" (func $ret (param i32 i32)))
   (import "env" "get" (func $get (result i32))) (import "env" "set" (func $set (param i32)))
   (import "env" "drop" (func $drop (param i32))) (import "env" "make" (func $make (param i32 i32)))
   (func (export "start") (param i32) (result i32) local.get 0 call $set i32.const 1)
   (func (export "callback") (param i32 i32 i32) (result i32)
    call $get i32.const 64 call $make call $get call $drop
    i32.const 64 i32.load i32.const 68 i32.load call $ret i32.const 0))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (alias core export $code "callback" (core func $callback))
  (func (export "add") async (param "arg" (borrow $r)) (result 2) (canon lift (core func $code "start") (memory $mem) async (callback $callback))))
 (instance $server (instantiate $server (with "r" (type $r)) (with "make" (func $make))))
 (alias export $server "add" (func $add))
 (component $client
  (import "r" (type $r (sub resource)))
  (type $out (tuple (own $r) u32)) (export "out" (type $out))
  (import "add" (func $add async (param "arg" (borrow $r)) (result 2)))
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

  (core func $add (canon lower (func $add) (memory $mem) async))
  (core func $borrow-drop (canon resource.drop $r))
  (core func $new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $poll (canon waitable-set.poll (memory $mem)))
 (core func $drop (canon waitable-set.drop))
 (core func $join (canon waitable.join))
 (core func $subdrop (canon subtask.drop))
 (core func $cancel (canon subtask.cancel))
 (core func $cancel-async (canon subtask.cancel async))
 (core instance $env (export "memory" (memory $mem)) (export "add" (func $add)) (export "new" (func $new)) (export "wait" (func $wait)) (export "poll" (func $poll)) (export "drop" (func $drop)) (export "join" (func $join)) (export "subdrop" (func $subdrop)) (export "cancel" (func $cancel)) (export "cancel-async" (func $cancel-async)))
  (core instance $resource (export "drop" (func $borrow-drop)))
  (core module $code
   (import "env" "memory" (memory 1))
 (import "env" "add" (func $add (param i32 i32) (result i32)))
 (import "env" "new" (func $new (result i32)))
 (import "env" "wait" (func $wait (param i32 i32) (result i32)))
 (import "env" "poll" (func $poll (param i32 i32) (result i32)))
 (import "env" "drop" (func $drop (param i32)))
 (import "env" "join" (func $join (param i32 i32)))
 (import "env" "subdrop" (func $subdrop (param i32)))
 (import "env" "cancel" (func $cancel (param i32) (result i32)))
 (import "env" "cancel-async" (func $cancel-async (param i32) (result i32)))
   (import "resource" "drop" (func $borrow-drop (param i32)))
   (func (export "call") (param i32 i32) (result i32) (local $sub i32) (local $set i32)
    local.get 0 local.get 1 call $add i32.const 4 i32.shr_u local.set $sub
    call $new local.set $set local.get $sub local.get $set call $join
    local.get $set i32.const 96 call $wait i32.const 1 i32.ne if unreachable end
    i32.const 100 i32.load i32.const 2 i32.ne if unreachable end
    local.get $sub call $subdrop local.get $set call $drop local.get 0 call $borrow-drop local.get 1))
  (core instance $code (instantiate $code (with "env" (instance $env)) (with "resource" (instance $resource))))
  (func (export "call") async (param "arg" (borrow $r)) (param "ptr" u32) (result 2) (canon lift (core func $code "call") (memory $mem))))
 (instance $client (instantiate $client (with "r" (type $r)) (with "add" (func $add))))
 (alias export $client "call" (func $call)) (export "call" (func $call)))

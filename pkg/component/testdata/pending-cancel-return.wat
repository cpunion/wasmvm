(component
 (component $server
 (core func $ret (canon task.return (result u32)))
 (core func $get (canon context.get i32 0))
 (core func $set (canon context.set i32 0))
 (core func $cancel (canon task.cancel))
 (core instance $env (export "ret" (func $ret)) (export "get" (func $get)) (export "set" (func $set)) (export "cancel" (func $cancel)))
 (core module $code
  (import "env" "ret" (func $ret (param i32)))
  (import "env" "get" (func $get (result i32)))
  (import "env" "set" (func $set (param i32)))
  (import "env" "cancel" (func $cancel))
  (func (export "start") (param i32) (result i32) local.get 0 call $set i32.const 1)
  (func (export "callback") (param i32 i32 i32) (result i32)
   local.get 0 i32.const 6 i32.eq if call $get i32.const 7 i32.add call $ret i32.const 0 return end
   local.get 0 if unreachable end
   call $get i32.const 7 i32.add call $ret i32.const 0))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (alias core export $code "callback" (core func $callback))
 (func (export "add") async (param "n" u32) (result u32) (canon lift (core func $code "start") async (callback $callback))))
 (instance $server (instantiate $server))
 (alias export $server "add" (func $add))
 (component $client
  (import "add" (func $add async (param "n" u32) (result u32)))
  (core module $memory (memory (export "memory") 1))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem))
  (core func $add (canon lower (func $add) (memory $mem) async))
  (core func $new (canon waitable-set.new))
 (core func $wait (canon waitable-set.wait (memory $mem)))
 (core func $poll (canon waitable-set.poll (memory $mem)))
 (core func $drop (canon waitable-set.drop))
 (core func $join (canon waitable.join))
 (core func $subdrop (canon subtask.drop))
 (core func $cancel (canon subtask.cancel))
 (core func $cancel-async (canon subtask.cancel async))
 (core instance $env (export "memory" (memory $mem)) (export "add" (func $add)) (export "new" (func $new)) (export "wait" (func $wait)) (export "poll" (func $poll)) (export "drop" (func $drop)) (export "join" (func $join)) (export "subdrop" (func $subdrop)) (export "cancel" (func $cancel)) (export "cancel-async" (func $cancel-async)))
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
 (func $start (param i32) (result i32) (local $packed i32)
  local.get 0 i32.const 32 call $add local.tee $packed i32.const 15 i32.and i32.const 1 i32.ne if unreachable end
  local.get $packed i32.const 4 i32.shr_u)
 (func $finish (param $sub i32) (param $set i32) (result i32)
  local.get $sub call $subdrop local.get $set call $drop i32.const 32 i32.load)
 (func (export "churn") (param $n i32) (result i32) (local $i i32) (local $sub i32) (local $set i32)
  block $done loop $loop local.get $i local.get $n i32.ge_u br_if $done
   local.get $i call $start local.set $sub call $new local.set $set
   local.get $sub local.get $set call $join
   local.get $set i32.const 64 call $poll if unreachable end
   local.get $set i32.const 64 call $wait i32.const 1 i32.ne if unreachable end
   i32.const 64 i32.load local.get $sub i32.ne if unreachable end
   i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
   local.get $sub local.get $set call $finish local.get $i i32.const 7 i32.add i32.ne if unreachable end
   local.get $i i32.const 1 i32.add local.set $i br $loop end end local.get $i)
 (func (export "cancel") (param $async i32) (result i32) (local $sub i32) (local $state i32)
  i32.const 1 call $start local.set $sub
  local.get $async if (result i32) local.get $sub call $cancel-async else local.get $sub call $cancel end local.set $state
  local.get $state i32.const -1 i32.eq if call $new local.set $async local.get $sub local.get $async call $join local.get $async i32.const 64 call $wait drop local.get $sub call $subdrop local.get $async call $drop local.get $state return end local.get $sub call $subdrop local.get $state)
 (func (export "bad") (param $mode i32) (result i32) (local $sub i32) (local $set i32)
  i32.const 1 call $start local.set $sub call $new local.set $set
  local.get $mode i32.eqz if local.get $sub call $subdrop end
  local.get $mode i32.const 1 i32.eq if local.get $sub local.get $set call $join local.get $sub call $cancel drop end
  local.get $mode i32.const 2 i32.eq if local.get $sub local.get $set call $join local.get $set call $drop end
  local.get $mode i32.const 3 i32.eq if local.get $set call $subdrop end
  local.get $mode i32.const 4 i32.eq if local.get $sub call $drop end
  local.get $mode i32.const 5 i32.eq if local.get $sub call $cancel drop local.get $sub call $cancel drop end
  local.get $mode i32.const 6 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 3 call $poll drop end
  local.get $mode i32.const 7 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 65532 call $wait drop end
  local.get $mode i32.const 8 i32.eq if local.get $sub local.get $set call $join local.get $set i32.const 0 call $join end
  local.get $mode i32.const 9 i32.eq if i32.const 2 call $start drop end
  i32.const 0) (func (export "multi") (param $n i32) (result i32) (local $a i32) (local $b i32) (local $c i32) (local $s i32) (local $other i32)
  local.get $n i32.const 32 call $add i32.const 4 i32.shr_u local.set $a
  local.get $n i32.const 1 i32.add i32.const 40 call $add i32.const 4 i32.shr_u local.set $b
  local.get $n i32.const 2 i32.add i32.const 48 call $add i32.const 4 i32.shr_u local.set $c
  call $new local.set $s call $new local.set $other
  local.get $a local.get $s call $join local.get $b local.get $s call $join local.get $c local.get $s call $join
  local.get $b local.get $other call $join local.get $b local.get $s call $join local.get $other call $drop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $wait i32.const 1 i32.ne if unreachable end i32.const 64 i32.load call $subdrop
  local.get $s i32.const 64 call $poll if unreachable end local.get $s call $drop
  i32.const 32 i32.load i32.const 40 i32.load i32.add i32.const 48 i32.load i32.add)
)
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "multi") async (param "n" u32) (result u32) (canon lift (core func $code "multi")))
  (func (export "churn") async (param "n" u32) (result u32) (canon lift (core func $code "churn")))
  (func (export "cancel") async (param "async" u32) (result u32) (canon lift (core func $code "cancel")))
  (func (export "bad") async (param "mode" u32) (result u32) (canon lift (core func $code "bad"))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "multi" (func $multi)) (export "multi" (func $multi))
 (alias export $client "churn" (func $churn)) (export "churn" (func $churn))
 (alias export $client "cancel" (func $cancel)) (export "cancel" (func $cancel))
 (alias export $client "bad" (func $bad)) (export "bad" (func $bad)))

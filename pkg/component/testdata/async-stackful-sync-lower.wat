(component
 (component $server
 (core func $ret (canon task.return (result u32)))
 (core func $get (canon context.get i32 0))
 (core func $set (canon context.set i32 0))
 (core func $cancel (canon task.cancel))
 (core func $yield (canon thread.yield))
 (core instance $env (export "ret" (func $ret)) (export "get" (func $get)) (export "set" (func $set)) (export "cancel" (func $cancel)) (export "yield" (func $yield)))
 (core module $code
  (import "env" "ret" (func $ret (param i32)))
  (import "env" "get" (func $get (result i32)))
  (import "env" "set" (func $set (param i32)))
  (import "env" "cancel" (func $cancel))
  (import "env" "yield" (func $yield (result i32)))
  (func (export "start") (param i32) local.get 0 call $set call $yield if call $cancel return end call $get i32.const 7 i32.add call $ret)
  (func (export "callback") (param i32 i32 i32) (result i32)
   local.get 0 i32.const 6 i32.eq if call $cancel i32.const 0 return end
   local.get 0 if unreachable end
   call $get i32.const 7 i32.add call $ret i32.const 0))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (alias core export $code "callback" (core func $callback))
 (func (export "add") async (param "n" u32) (result u32) (canon lift (core func $code "start") async)))
 (instance $server (instantiate $server))
 (alias export $server "add" (func $add))
 (component $client
 (import "add" (func $add async (param "n" u32) (result u32)))
 (core func $add (canon lower (func $add)))
 (core instance $env (export "add" (func $add)))
 (core module $code (import "env" "add" (func $add (param i32) (result i32)))
  (func (export "churn") (param $n i32) (result i32) (local $i i32)
   block $done loop $loop local.get $i local.get $n i32.ge_u br_if $done
    local.get $i call $add local.get $i i32.const 7 i32.add i32.ne if unreachable end
    local.get $i i32.const 1 i32.add local.set $i br $loop end end local.get $i))
 (core instance $code (instantiate $code (with "env" (instance $env))))
 (func (export "churn") async (param "n" u32) (result u32) (canon lift (core func $code "churn"))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "churn" (func $churn)) (export "churn" (func $churn)))

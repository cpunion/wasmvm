(component
 (component $server
  (core func $return (canon task.return (result u32)))
  (core instance $env (export "return" (func $return)))
  (core module $code
   (import "env" "return" (func $return (param i32)))
   (func (export "add") (param i32) local.get 0 i32.const 7 i32.add call $return))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "add") async (param "value" u32) (result u32) (canon lift (core func $code "add") async)))
 (instance $server (instantiate $server))
 (alias export $server "add" (func $add))
 (component $client
  (import "add" (func $add async (param "value" u32) (result u32)))
  (core module $memory (memory (export "memory") i64 1))
  (core instance $memory (instantiate $memory))
  (alias core export $memory "memory" (core memory $mem))
  (core func $add (canon lower (func $add) (memory $mem) async))
  (core instance $env (export "add" (func $add)) (export "memory" (memory $mem)))
  (core module $code
   (import "env" "memory" (memory i64 1))
   (import "env" "add" (func $add (param i32 i64) (result i32)))
   (func (export "churn") (param $n i32) (result i32) (local $i i32)
    block $end loop $next
     local.get $i local.get $n i32.ge_u br_if $end
     local.get $i i64.const 32 call $add i32.const 2 i32.ne if unreachable end
     i64.const 32 i32.load local.get $i i32.const 7 i32.add i32.ne if unreachable end
     local.get $i i32.const 1 i32.add local.set $i br $next
    end end local.get $i))
  (core instance $code (instantiate $code (with "env" (instance $env))))
  (func (export "churn") (param "count" u32) (result u32) (canon lift (core func $code "churn"))))
 (instance $client (instantiate $client (with "add" (func $add))))
 (alias export $client "churn" (func $churn))
 (export "churn" (func $churn)))

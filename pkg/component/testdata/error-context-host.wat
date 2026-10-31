(component
 (import "host" (func $host (param "context" error-context) (result error-context)))
 (core func $host (canon lower (func $host)))
 (core func $drop (canon error-context.drop))
 (core instance $env (export "host" (func $host)) (export "drop" (func $drop)))
 (core module $guest
  (import "env" "host" (func $host (param i32) (result i32)))
  (import "env" "drop" (func $drop (param i32)))
  (global $pending (mut i32) (i32.const 0))
  (func $release (export "release") global.get $pending if global.get $pending call $drop i32.const 0 global.set $pending end)
  (func (export "call") (param i32) (result i32) (local $result i32)
   call $release local.get 0 call $host local.set $result local.get 0 call $drop
   local.get $result global.set $pending local.get $result))
 (core instance $guest (instantiate $guest (with "env" (instance $env))))
 (alias core export $guest "call" (core func $call))
 (alias core export $guest "release" (core func $release))
 (func (export "call") (param "context" error-context) (result error-context) (canon lift (core func $call)))
 (func (export "release") (canon lift (core func $release))))

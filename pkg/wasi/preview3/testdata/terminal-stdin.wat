(component
 (import "wasi:cli/terminal-input@0.3.2" (instance $resource (export "terminal-input" (type (sub resource)))))
 (alias export $resource "terminal-input" (type $r))
 (import "wasi:cli/terminal-stdin@0.3.2" (instance $terminal
  (alias outer 1 $r (type $outer)) (export "terminal-input" (type $r (eq $outer)))
  (type $option (option (own $r))) (export "get-terminal-stdin" (func (result $option)))))
 (alias export $terminal "get-terminal-stdin" (func $get))
 (core module $mem (memory (export "memory") 1)) (core instance $mem (instantiate $mem))
 (alias core export $mem "memory" (core memory $memory))
 (core func $get (canon lower (func $get) (memory $memory))) (core func $drop (canon resource.drop $r))
 (core instance $env (export "get" (func $get)) (export "drop" (func $drop)) (export "memory" (memory $memory)))
 (core module $m (import "env" "memory" (memory 1)) (import "env" "get" (func $get (param i32))) (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (result i32) (local $present i32)
   i32.const 0 call $get i32.const 0 i32.load8_u local.tee $present
   if i32.const 4 i32.load call $drop end local.get $present))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") (result u32) (canon lift (core func $m "run"))))

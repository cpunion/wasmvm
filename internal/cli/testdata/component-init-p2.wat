(component
 (core module $m
  (func $start nop nop nop) (start $start)
  (func (export "run") (result i32) nop i32.const 0))
 (core instance $m (instantiate $m))
 (func $run  (result (result)) (canon lift (core func $m "run")))
 (instance $run (export "run" (func $run)))
 (export "wasi:cli/run@0.2.12" (instance $run)))

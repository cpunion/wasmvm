(component
 (core module $code (func (export "run") (result i32) i32.const 0))
 (core instance $code (instantiate $code))
 (func $run async (result (result)) (canon lift (core func $code "run")))
 (instance $cli (export "run" (func $run)))
 (export "wasi:cli/run@0.2.12" (instance $cli)))

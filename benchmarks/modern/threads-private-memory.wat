(module (memory 1 1 )
 (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
 (local $i i32) (local $sum i32)
 i32.const 0 local.get $seed i32.store
 loop $again
 i32.const 0 i32.const 0 i32.load i32.const 1 i32.add i32.store
 local.get $i i32.const 1 i32.add local.tee $i
 local.get $iterations i32.lt_u br_if $again end
 i32.const 0 i32.load))

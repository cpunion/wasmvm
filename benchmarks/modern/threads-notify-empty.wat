(module (memory 1 1 shared)
 (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
 (local $i i32) (local $sum i32)
 i32.const 0 local.get $seed i32.store
 loop $again
 i32.const 0 i32.const 1 memory.atomic.notify drop
 local.get $i i32.const 1 i32.add local.tee $i
 local.get $iterations i32.lt_u br_if $again end
 local.get $seed local.get $sum i32.add))

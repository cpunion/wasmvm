(module
(func (export "run") (param $x i32) (param $n i32) (result i32) (local $i i32)
(loop $again (local.set $x (i31.get_u (ref.i31 (i32.add (i32.mul (local.get $x) (i32.const 1664525)) (i32.const 1013904223))))) (local.set $i (i32.add (local.get $i) (i32.const 1))) (br_if $again (i32.lt_u (local.get $i) (local.get $n))))  (local.get $x)))

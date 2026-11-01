(module
(type $t (struct (field (mut i32))))(func (export "run") (param $x i32) (param $n i32) (result i32) (local $i i32) (local $r (ref $t)) (local.set $r (struct.new $t (local.get $x)))
(loop $again (struct.set $t 0 (local.get $r) (i32.add (i32.mul (local.get $x) (i32.const 1664525)) (i32.const 1013904223))) (local.set $x (struct.get $t 0 (local.get $r))) (local.set $i (i32.add (local.get $i) (i32.const 1))) (br_if $again (i32.lt_u (local.get $i) (local.get $n))))  (local.get $x)))

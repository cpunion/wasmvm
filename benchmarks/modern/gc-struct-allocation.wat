(module
(type $t (struct (field i32) (field (ref null $t)))) (global $sink (mut (ref null $t)) (ref.null $t))(func (export "run") (param $x i32) (param $n i32) (result i32) (local $i i32) (local $r (ref null $t))
(loop $again (local.set $r (struct.new $t (i32.add (i32.mul (local.get $x) (i32.const 1664525)) (i32.const 1013904223)) (local.get $r))) (local.set $x (struct.get $t 0 (local.get $r))) (local.set $i (i32.add (local.get $i) (i32.const 1))) (br_if $again (i32.lt_u (local.get $i) (local.get $n)))) (global.set $sink (local.get $r)) (local.get $x)))

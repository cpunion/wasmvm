(module
(type $t (struct (field (mut i32))))
(type $a (array (ref $t)))
(global $objects (ref $a) (array.new_fixed $a 2 (struct.new $t (i32.const 0)) (struct.new $t (i32.const 1))))
(func (export "run") (param $x i32) (param $n i32) (result i32) (local $i i32)
(struct.set $t 0 (array.get $a (global.get $objects) (i32.const 0)) (local.get $x))
(struct.set $t 0 (array.get $a (global.get $objects) (i32.const 1)) (i32.xor (local.get $x) (i32.const 0xdeadbeef)))
(loop $again
(struct.set $t 0 (array.get $a (global.get $objects) (i32.and (local.get $x) (i32.const 1))) (i32.add (i32.mul (local.get $x) (i32.const 1664525)) (i32.const 1013904223)))
(local.set $x (i32.xor (struct.get $t 0 (array.get $a (global.get $objects) (i32.and (i32.shr_u (local.get $x) (i32.const 1)) (i32.const 1)))) (local.get $i)))
(local.set $i (i32.add (local.get $i) (i32.const 1)))
(br_if $again (i32.lt_u (local.get $i) (local.get $n)))) (local.get $x)))

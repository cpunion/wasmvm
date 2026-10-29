(module
 (rec (type $node (struct
  (field (mut (ref null $node))) (field (mut i32)) (field funcref) (field externref))))
 (import "h" "escape" (func $escape (param (ref $node))))
 (func $answer (result i32) (i32.const 77))
 (elem declare func $answer)
 (func $start
  (call $escape (struct.new $node (ref.null $node) (i32.const 77) (ref.func $answer) (ref.null extern)))
  (unreachable))
 (start $start)
)

(component
 (type $poll (instance
  (export "pollable" (type $p (sub resource)))
  (type $ready (func (param "self" (borrow $p)) (result bool)))
  (export "[method]pollable.ready" (func (type $ready)))
  (type $block (func (param "self" (borrow $p)) (result bool)))
  (export "[method]pollable.block" (func (type $block)))
  (type $input (list (borrow $p)))
  (type $output (list u32))
  (type $poll (func (param "in" $input) (result $output)))
  (export "poll" (func (type $poll)))))
 (import "wasi:io/poll@0.2.12" (instance (type $poll))))

(module
 (type (func (param i64)))
 (type $different-index (func (param i32) (result i32)))
 (type $apply (func (param i32 (ref null $different-index)) (result i32)))
 (import "p" "apply" (func $apply (type $apply)))
 (func (export "run") (type $apply) (return_call $apply (local.get 0) (local.get 1)))
)

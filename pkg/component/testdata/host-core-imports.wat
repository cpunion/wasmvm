(module
 (import "env" "seed" (func $seed (result i32)))
 (func (export "add") (param i32 i32) (result i32) local.get 0 local.get 1 i32.add call $seed i32.add))

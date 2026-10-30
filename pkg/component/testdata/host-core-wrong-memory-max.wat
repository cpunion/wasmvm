(module
 (func (export "apply") (param i32) (result i32) local.get 0 i32.const 1 i32.add)
 (memory (export "memory") i64 2 5)
 (table (export "table") 2 3 funcref)
 (global (export "global") i32 (i32.const 9))
 (tag (export "tag") (param i32)))

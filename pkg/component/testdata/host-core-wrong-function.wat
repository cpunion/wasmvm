(module
 (func (export "apply") (param i64) (result i64) local.get 0 i64.const 1 i64.add)
 (memory (export "memory") i64 2 3)
 (table (export "table") 2 3 funcref)
 (global (export "global") i32 (i32.const 9))
 (tag (export "tag") (param i32)))

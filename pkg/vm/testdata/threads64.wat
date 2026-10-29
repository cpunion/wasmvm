(module
 (memory $shared (export "memory") i64 1 2 shared)
 (func (export "run") (param i64) (result i64)
  (i64.atomic.store $shared (local.get 0) (i64.const 99))
  (i64.atomic.rmw32.add_u $shared (local.get 0) (i64.const 5)))
 (func (export "load") (param i64) (result i64) (i64.load $shared (local.get 0)))
 (func (export "notify") (param i64 i32) (result i32) (memory.atomic.notify $shared (local.get 0) (local.get 1)))
)

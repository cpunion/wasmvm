(module
 (memory 1)
 (func (export "load") (result i32) (i32.atomic.load (i32.const 0)))
 (func (export "wait") (result i32) (memory.atomic.wait32 (i32.const 0) (i32.const 0) (i64.const 0)))
 (func (export "notify") (result i32) (memory.atomic.notify (i32.const 0) (i32.const -1)))
)

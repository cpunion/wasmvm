(module (import "env" "memory" (memory 1 2 shared))(import "wasi" "thread-spawn" (func (param i32) (result i32)))(func (export "_start")))

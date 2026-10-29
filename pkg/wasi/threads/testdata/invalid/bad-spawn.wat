(module (import "env" "memory" (memory 1 2 shared))(import "wasi" "thread-spawn" (func))(func (export "_start"))(func (export "wasi_thread_start") (param i32 i32)))

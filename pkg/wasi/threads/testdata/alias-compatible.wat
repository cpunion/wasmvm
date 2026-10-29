(module
 (import "env" "memory" (memory 1 3 shared))
 (import "env" "memory" (memory 2 4 shared))
 (func (export "_start")
  (if (i32.ne (memory.size 0) (i32.const 2)) (then unreachable))
  (if (i32.ne (memory.grow 1 (i32.const 1)) (i32.const 2)) (then unreachable))
  (if (i32.ne (memory.size 0) (i32.const 3)) (then unreachable))
  (if (i32.ne (memory.grow 0 (i32.const 1)) (i32.const -1)) (then unreachable))))

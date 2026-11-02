;; Build with pinned wasm-tools 1.262.0: wasm-tools parse shared-concurrent.wat -o shared-concurrent.wasm
(module
  (import "env" "memory" (memory $shared 1 3 shared))
  (func (export "signal") (param $at i32) (param $value i32)
    (i32.atomic.store (local.get $at) (local.get $value)))
  (func (export "observe-growth") (result i32)
    (i32.atomic.store (i32.const 0) (i32.const 1))
    (loop $wait
      (br_if $wait (i32.eqz (i32.atomic.load (i32.const 4)))))
    (i32.add (memory.size) (i32.load (i32.const 65536))))
  (func (export "count") (param $count i32) (result i32) (local $i i32)
    (loop $next
      (drop (i32.atomic.rmw.add (i32.const 8) (i32.const 1)))
      (local.set $i (i32.add (local.get $i) (i32.const 1)))
      (br_if $next (i32.lt_u (local.get $i) (local.get $count))))
    atomic.fence
    (i32.atomic.load (i32.const 8)))
  (func (export "wait-zero") (result i32)
    (memory.atomic.wait32 (i32.const 16) (i32.const 0) (i64.const -1)))
  (func (export "wake") (result i32)
    (i32.atomic.store (i32.const 16) (i32.const 1))
    (memory.atomic.notify (i32.const 16) (i32.const 1))))

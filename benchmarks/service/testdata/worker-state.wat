;; Project-authored fixture. Producer: wasm-tools 1.258.0.
;; Mutable memory and globals must persist inside each independent worker.
(module
  (memory 1)
  (global $counter (mut i32) (i32.const 0))
  (func (export "run") (param i32 i32) (result i32)
    (local $old i32)
    global.get $counter
    i32.const 0 i32.load
    i32.add local.set $old
    global.get $counter i32.const 1 i32.add global.set $counter
    i32.const 0
    i32.const 0 i32.load i32.const 1 i32.add
    i32.store
    local.get $old))

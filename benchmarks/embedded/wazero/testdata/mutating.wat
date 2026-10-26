;; Project-authored unit fixture. Producer: wasm-tools 1.258.0.
(module
  (global $counter (mut i32) (i32.const 0))
  (func (export "run") (param i32 i32) (result i32)
    global.get $counter i32.const 1 i32.add global.set $counter
    global.get $counter))

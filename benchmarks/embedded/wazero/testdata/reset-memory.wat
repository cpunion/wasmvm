;; Project-authored unit fixture. Producer: wasm-tools 1.258.0.
(module
  (memory 1)
  (func (export "run") (param $seed i32) (param $count i32) (result i32)
    i32.const 0 local.get $seed i32.store
    i32.const 0 i32.load))

;; Project-authored unit fixture. Producer: wasm-tools 1.258.0.
(module
  (memory 1)
  (data (i32.const 65536) "x")
  (func (export "run") (param i32 i32) (result i32)
    local.get 0))

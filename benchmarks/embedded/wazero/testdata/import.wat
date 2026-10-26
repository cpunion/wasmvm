;; Project-authored unit fixture. Producer: wasm-tools 1.258.0.
(module
  (import "host" "x" (func))
  (func (export "run") (param i32 i32) (result i32)
    local.get 0))

;; Authored for wasmvm. Regenerate: python3 benchmarks/embedded/generate.py
;; Runtime parameters determine the work and checksum; no WASI/import/start code.
(module

  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)

    local.get $iterations i32.const 0 i32.le_s if unreachable end

    local.get $iterations i32.const 1 i32.ne if unreachable end
    local.get $seed
))

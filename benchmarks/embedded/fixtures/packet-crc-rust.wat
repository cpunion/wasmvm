(module $packet-crc-rust.wasm
  (type (;0;) (func (param i32 i32) (result i32)))
  (memory (;0;) 2 2)
  (global $__stack_pointer (;0;) (mut i32) i32.const 65536)
  (export "memory" (memory 0))
  (export "run" (func $run))
  (func $run (;0;) (type 0) (param i32 i32) (result i32)
    (local i32 i32 i32 i32 i32 i32)
    block ;; label = @1
      block ;; label = @2
        local.get 1
        br_if 0 (;@2;)
        local.get 0
        local.set 2
        br 1 (;@1;)
      end
      local.get 0
      local.set 2
      i32.const 0
      local.set 3
      loop ;; label = @2
        local.get 3
        local.get 0
        i32.add
        local.set 4
        i32.const 256
        local.set 5
        i32.const -1
        local.set 6
        loop ;; label = @3
          i32.const -306674912
          i32.const 0
          i32.const -306674912
          i32.const 0
          i32.const -306674912
          i32.const 0
          i32.const -306674912
          i32.const 0
          local.get 4
          i32.const 1664525
          i32.mul
          i32.const 1013904223
          i32.add
          local.tee 4
          i32.const 24
          i32.shr_u
          local.get 6
          i32.xor
          local.tee 6
          i32.const 1
          i32.and
          select
          local.get 6
          i32.const 1
          i32.shr_u
          local.tee 6
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 7
          i32.const -306674912
          i32.const 0
          local.get 6
          i32.const 1
          i32.and
          select
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 6
          i32.const 1
          i32.and
          select
          i32.const -306674912
          i32.const 0
          local.get 7
          i32.const 1
          i32.and
          select
          local.get 6
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 6
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 7
          i32.const 1
          i32.and
          select
          i32.const -306674912
          i32.const 0
          local.get 6
          i32.const 1
          i32.and
          select
          local.get 7
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 6
          i32.xor
          i32.const 1
          i32.shr_u
          local.tee 7
          i32.const 1
          i32.and
          select
          i32.const -306674912
          i32.const 0
          local.get 6
          i32.const 1
          i32.and
          select
          local.get 7
          i32.xor
          i32.const 1
          i32.shr_u
          i32.xor
          local.set 6
          local.get 5
          i32.const -1
          i32.add
          local.tee 5
          br_if 0 (;@3;)
        end
        local.get 2
        i32.const 3
        i32.rotl
        local.get 6
        i32.xor
        i32.const -1
        i32.xor
        local.set 2
        local.get 3
        i32.const 1
        i32.add
        local.tee 3
        local.get 1
        i32.ne
        br_if 0 (;@2;)
      end
    end
    local.get 2
  )
  (@producers
    (processed-by "rustc" "1.98.0 (88d9e12ae 2026-08-18)")
  )
)

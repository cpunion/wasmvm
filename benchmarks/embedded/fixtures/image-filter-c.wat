(module $image-filter-c.wasm
  (type (;0;) (func (param i32 i32) (result i32)))
  (memory (;0;) 2 2)
  (global $__stack_pointer (;0;) (mut i32) i32.const 65536)
  (export "memory" (memory 0))
  (export "run" (func $run))
  (func $run (;0;) (type 0) (param i32 i32) (result i32)
    (local i32 i32 i32 i32 i32 i32 i32 i32)
    global.get $__stack_pointer
    i32.const 1024
    i32.sub
    local.tee 2
    global.set $__stack_pointer
    i32.const 0
    local.set 3
    local.get 0
    local.set 4
    loop ;; label = @1
      local.get 2
      local.get 3
      i32.add
      local.tee 5
      i32.const 3
      i32.add
      local.get 4
      i32.const 158984081
      i32.mul
      i32.const -1426500812
      i32.add
      local.tee 6
      i32.const 24
      i32.shr_u
      i32.store8
      local.get 5
      i32.const 2
      i32.add
      local.get 4
      i32.const -1354167659
      i32.mul
      i32.const -775096599
      i32.add
      i32.const 24
      i32.shr_u
      i32.store8
      local.get 5
      i32.const 1
      i32.add
      local.get 4
      i32.const 389569705
      i32.mul
      i32.const 1196435762
      i32.add
      i32.const 24
      i32.shr_u
      i32.store8
      local.get 5
      local.get 4
      i32.const 1664525
      i32.mul
      i32.const 1013904223
      i32.add
      i32.const 24
      i32.shr_u
      i32.store8
      local.get 6
      local.set 4
      local.get 3
      i32.const 4
      i32.add
      local.tee 3
      i32.const 1024
      i32.ne
      br_if 0 (;@1;)
    end
    block ;; label = @1
      local.get 1
      i32.eqz
      br_if 0 (;@1;)
      i32.const 0
      local.set 7
      i32.const 33
      local.set 8
      loop ;; label = @2
        local.get 2
        local.set 3
        local.get 8
        local.set 6
        i32.const 1
        local.set 9
        loop ;; label = @3
          i32.const 0
          local.set 5
          loop ;; label = @4
            local.get 6
            local.get 5
            i32.add
            local.get 3
            local.get 5
            i32.add
            local.tee 4
            i32.const 33
            i32.add
            i32.load8_u
            i32.const 2
            i32.shl
            local.get 4
            i32.const 32
            i32.add
            i32.load8_u
            local.get 4
            i32.const 1
            i32.add
            i32.load8_u
            i32.add
            local.get 4
            i32.const 34
            i32.add
            i32.load8_u
            i32.add
            local.get 4
            i32.const 65
            i32.add
            i32.load8_u
            i32.add
            i32.const 1
            i32.shl
            i32.add
            local.get 4
            i32.const 2
            i32.add
            i32.load8_u
            i32.add
            local.get 4
            i32.load8_u
            i32.add
            local.get 4
            i32.const 64
            i32.add
            i32.load8_u
            i32.add
            local.get 4
            i32.const 66
            i32.add
            i32.load8_u
            i32.add
            i32.const 4
            i32.shr_u
            i32.add
            local.get 0
            i32.const 5
            i32.rotl
            i32.xor
            local.set 0
            local.get 5
            i32.const 1
            i32.add
            local.tee 5
            i32.const 30
            i32.ne
            br_if 0 (;@4;)
          end
          local.get 3
          i32.const 32
          i32.add
          local.set 3
          local.get 6
          i32.const 32
          i32.add
          local.set 6
          local.get 9
          i32.const 1
          i32.add
          local.tee 9
          i32.const 31
          i32.ne
          br_if 0 (;@3;)
        end
        local.get 8
        i32.const 1
        i32.add
        local.set 8
        local.get 7
        i32.const 1
        i32.add
        local.tee 7
        local.get 1
        i32.ne
        br_if 0 (;@2;)
      end
    end
    local.get 2
    i32.const 1024
    i32.add
    global.set $__stack_pointer
    local.get 0
  )
  (@producers
    (processed-by "Homebrew clang" "22.1.8")
  )
  (@custom "target_features" (after code) "\08+\0bbulk-memory+\0fbulk-memory-opt+\16call-indirect-overlong+\0amultivalue+\0fmutable-globals+\13nontrapping-fptoint+\0freference-types+\08sign-ext")
)

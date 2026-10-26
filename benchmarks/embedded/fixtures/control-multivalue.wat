(module
  
  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed local.set $x
    loop $again
      
      block $done (result i32)
        block $three
          block $two
            block $one
              block $zero
                local.get $x i32.const 3 i32.and br_table $zero $one $two $three
              end
              local.get $x local.get $i i32.add br $done
            end
            local.get $x i32.const 5 i32.rotl br $done
          end
          local.get $x local.get $i i32.xor i32.const 1013904223 i32.add br $done
        end
        local.get $x i32.const 1664525 i32.mul
      end
      local.get $i
      block (param i32 i32) (result i32 i32)
        i32.add local.tee $tmp local.get $tmp i32.const 3 i32.rotl
      end
      i32.xor local.set $x

      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x))

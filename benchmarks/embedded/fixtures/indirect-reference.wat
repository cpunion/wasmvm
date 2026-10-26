(module
  
  (type $binary (func (param i32 i32) (result i32)))
  (table 2 2 funcref)
  (elem (i32.const 0) $add $mix)
  (global $state (mut i32) (i32.const 0))
  (func $add (type $binary) local.get 0 local.get 1 i32.add)
  (func $mix (type $binary) local.get 0 local.get 1 i32.xor i32.const 7 i32.rotl)

  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed global.set $state
    loop $again
      
      global.get $state i32.const 1 i32.and local.tee $tmp
      table.get 0 ref.is_null if unreachable end
      local.get $tmp local.get $tmp table.get 0 table.set 0
      global.get $state local.get $i local.get $tmp call_indirect (type $binary)
      global.set $state

      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    global.get $state))

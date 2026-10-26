(module
  
  (func $recurse (param $x i32) (param $n i32) (result i32)
    local.get $n i32.eqz
    if (result i32) local.get $x
    else
      local.get $x local.get $n i32.add
      local.get $n i32.const 1 i32.sub call $recurse
      i32.const 1 i32.rotl local.get $n i32.xor
    end)

  (func (export "run") (param $seed i32) (param $iterations i32) (result i32)
    (local $i i32) (local $x i32) (local $tmp i32) (local $v v128) (local $wide i64)
    local.get $iterations i32.const 0 i32.le_s if unreachable end
    local.get $seed local.set $x
    loop $again
      local.get $x local.get $i i32.add i32.const 32 call $recurse local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      local.get $iterations i32.lt_u br_if $again
    end
    local.get $x))

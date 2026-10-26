(module
  (import "host" "mix" (func $mix (param i32 i32) (result i32)))
  (func (export "run") (param $seed i32) (param $count i32) (result i32)
    (local $i i32)
    local.get $count i32.const 0 i32.le_s if unreachable end
    loop $again
      local.get $seed local.get $i call $mix local.set $seed
      local.get $i i32.const 1 i32.add local.tee $i local.get $count i32.lt_u br_if $again
    end
    local.get $seed))


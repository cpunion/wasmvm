(module
  (type $target (func (result i32)))
  (func $a (type $target) i32.const 7)
  (func $b (type $target) i32.const 11)
  (table $t 16 funcref) (elem declare func $a $b)
  (func (export "run") (param i32 i32) (result i32) (local $i i32) (local $x i32)
    local.get 0 local.set $x
    (loop $loop
      local.get $i local.get 0 i32.add i32.const 15 i32.and
      ref.func $a table.set $t
      local.get $x i32.const 1664525 i32.mul
      local.get $i local.get 0 i32.add i32.const 15 i32.and table.get $t ref.is_null i32.add local.set $x
      local.get $i i32.const 1 i32.add local.tee $i
      local.get 1 i32.lt_u br_if $loop)
    local.get $x i32.const 1
    i32.const 0 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 0 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 2
    i32.const 1 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 1 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 3
    i32.const 2 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 2 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 4
    i32.const 3 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 3 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 5
    i32.const 4 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 4 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 6
    i32.const 5 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 5 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 7
    i32.const 6 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 6 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 8
    i32.const 7 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 7 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 9
    i32.const 8 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 8 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 10
    i32.const 9 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 9 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 11
    i32.const 10 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 10 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 12
    i32.const 11 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 11 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 13
    i32.const 12 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 12 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 14
    i32.const 13 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 13 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 15
    i32.const 14 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 14 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
local.get $x i32.const 16
    i32.const 15 table.get $t ref.is_null
    if (result i32) i32.const 0 else i32.const 15 call_indirect $t (type $target) end
    i32.mul i32.add local.set $x
    local.get $x))

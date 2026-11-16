(component
(import "step" (func $step (param "x" u32)(result u32)))
(core func $lower (canon lower (func $step)))
(core module $m
(import "env" "step" (func $step (param i32)(result i32)))
(func (export "run") (param $n i32) (param $seed i32) (result i32)
(local $i i32) (local $x i32)
local.get $seed local.set $x
block $done loop $again
 local.get $i local.get $n i32.ge_u br_if $done
 local.get $x call $step local.set $x
 local.get $i i32.const 1 i32.add local.set $i
 br $again
end end
local.get $x)
)
(core instance $env (export "step" (func $lower)))
(core instance $i (instantiate $m (with "env" (instance $env))))
(func (export "run") (param "n" u32)(param "seed" u32)(result u32)(canon lift (core func $i "run")))
)

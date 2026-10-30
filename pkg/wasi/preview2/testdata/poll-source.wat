(component
 (type $poll (instance
  (export "pollable" (type $p (sub resource)))
  (type $ready (func (param "self" (borrow $p)) (result bool)))
  (export "[method]pollable.ready" (func (type $ready)))
  (type $block (func (param "self" (borrow $p))))
  (export "[method]pollable.block" (func (type $block)))
  (type $input (list (borrow $p)))
  (type $output (list u32))
  (type $poll (func (param "in" $input) (result $output)))
  (export "poll" (func (type $poll)))))
 (import "wasi:io/poll@0.2.12" (instance $poll (type $poll)))
 (alias export $poll "pollable" (type $pollable))
 (import "create" (func $create (result (own $pollable))))
 (alias export $poll "[method]pollable.ready" (func $ready))
 (alias export $poll "[method]pollable.block" (func $block))
 (alias export $poll "poll" (func $poll))
 (type $clock (instance
  (alias outer 1 $pollable (type $outer-p))
  (export "pollable" (type $p (eq $outer-p)))
  (type $read (func (result u64)))
  (export "now" (func (type $read)))
  (export "resolution" (func (type $read)))
  (type $subscribe (func (param "when" u64) (result (own $p))))
  (export "subscribe-instant" (func (type $subscribe)))
  (export "subscribe-duration" (func (type $subscribe)))))
 (import "wasi:clocks/monotonic-clock@0.2.12" (instance $clock (type $clock)))
 (alias export $clock "now" (func $now))
 (alias export $clock "resolution" (func $resolution))
 (alias export $clock "subscribe-instant" (func $instant))
 (alias export $clock "subscribe-duration" (func $duration))
 (core module $memory
  (memory (export "memory") 1)
  (global $next (mut i32) (i32.const 512))
  (func (export "reset") (param i32) i32.const 512 global.set $next)
  (func (export "realloc") (param i32 i32 i32 i32) (result i32)
   (local $ptr i32)
   global.get $next local.get 2 i32.const 1 i32.sub i32.add
   i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr
   local.get 3 i32.add global.set $next local.get $ptr))
 (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (alias core export $memory "realloc" (core func $realloc))
 (alias core export $memory "reset" (core func $reset))
 (core func $lower-ready (canon lower (func $ready)))
 (core func $lower-block (canon lower (func $block)))
 (core func $lower-poll (canon lower (func $poll) (memory $mem) (realloc $realloc)))
 (core func $lower-now (canon lower (func $now)))
 (core func $lower-resolution (canon lower (func $resolution)))
 (core func $lower-instant (canon lower (func $instant)))
 (core func $lower-duration (canon lower (func $duration)))
 (core func $create (canon lower (func $create)))
 (core func $drop (canon resource.drop $pollable))
 (core module $guest
  (import "env" "memory" (memory 1))
  (import "env" "ready" (func $ready (param i32) (result i32)))
  (import "env" "block" (func $block (param i32)))
  (import "env" "poll" (func $poll (param i32 i32 i32)))
  (import "env" "drop" (func $drop (param i32)))
  (import "env" "duration" (func $duration (param i64) (result i32)))
  (import "env" "instant" (func $instant (param i64) (result i32)))
  (func (export "probe") (result i32)
   (local $past i32) (local $future i32) (local $never i32)
   i64.const 0 call $instant local.set $past
   i64.const -1 call $instant local.tee $future call $ready if unreachable end
   i64.const -1 call $duration local.tee $never call $ready if unreachable end
   local.get $past call $block
   local.get $past call $ready i32.eqz if unreachable end
   i32.const 128 local.get $future i32.store
   i32.const 132 local.get $past i32.store
   i32.const 136 local.get $past i32.store
   i32.const 128 i32.const 3 i32.const 0 call $poll
   i32.const 4 i32.load i32.const 2 i32.ne if unreachable end
   i32.const 0 i32.load i32.load i32.const 1 i32.ne if unreachable end
   i32.const 0 i32.load i32.const 4 i32.add i32.load i32.const 2 i32.ne if unreachable end
   local.get $past call $drop local.get $future call $drop local.get $never call $drop i32.const 2)
  (func (export "churn") (param $count i32) (result i32)
   (local $at i32) (local $handle i32)
   block $end loop $loop
    local.get $at local.get $count i32.ge_u br_if $end
    i64.const 0 call $duration local.tee $handle call $block
    local.get $handle call $drop
    local.get $at i32.const 1 i32.add local.set $at br $loop
   end end local.get $at)
  (func (export "ready") (param i32) (result i32)
   (local $result i32) local.get 0 call $ready local.set $result
   local.get 0 call $drop local.get $result)
  (func (export "block") (param i32) local.get 0 call $block local.get 0 call $drop)
  (func (export "poll") (param $ptr i32) (param $len i32) (result i32)
   (local $at i32)
   local.get $ptr local.get $len i32.const 0 call $poll
   block $end loop $next
    local.get $at local.get $len i32.ge_u br_if $end
    local.get $ptr local.get $at i32.const 4 i32.mul i32.add i32.load call $drop
    local.get $at i32.const 1 i32.add local.set $at br $next
   end end i32.const 0))
 (core instance $env
  (export "memory" (memory $mem)) (export "ready" (func $lower-ready))
  (export "block" (func $lower-block)) (export "poll" (func $lower-poll)) (export "drop" (func $drop)) (export "duration" (func $lower-duration)) (export "instant" (func $lower-instant)))
 (core instance $guest (instantiate $guest (with "env" (instance $env))))
 (alias core export $guest "ready" (core func $guest-ready))
 (alias core export $guest "block" (core func $guest-block))
 (alias core export $guest "poll" (core func $guest-poll))
 (alias core export $guest "churn" (core func $guest-churn))
 (alias core export $guest "probe" (core func $guest-probe))
 (func (export "probe") (result u32) (canon lift (core func $guest-probe) (post-return $reset)))
 (func (export "churn") (param "count" u32) (result u32) (canon lift (core func $guest-churn)))
 (export "pollable" (type $pollable))
 (func (export "create") (result (own $pollable)) (canon lift (core func $create)))
 (type $input (list (borrow $pollable)))
 (type $output (list u32))
 (func (export "now") (result u64) (canon lift (core func $lower-now)))
 (func (export "resolution") (result u64) (canon lift (core func $lower-resolution)))
 (func (export "instant") (param "when" u64) (result (own $pollable)) (canon lift (core func $lower-instant)))
 (func (export "duration") (param "when" u64) (result (own $pollable)) (canon lift (core func $lower-duration)))
 (func (export "ready") (param "self" (borrow $pollable)) (result bool) (canon lift (core func $guest-ready)))
 (func (export "block") (param "self" (borrow $pollable)) (canon lift (core func $guest-block)))
 (func (export "poll") (param "in" $input) (result $output)
  (canon lift (core func $guest-poll) (memory $mem) (realloc $realloc) (post-return $reset))))

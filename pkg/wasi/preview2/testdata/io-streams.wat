(component
  (type $ty-wasi:io/error@0.2.12 (;0;)
    (instance
      (export (;0;) "error" (type (sub resource)))
      (type (;1;) (borrow 0))
      (type (;2;) (func (param "self" 1) (result string)))
      (export (;0;) "[method]error.to-debug-string" (func (type 2)))
    )
  )
  (import "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12 (;0;) (type $ty-wasi:io/error@0.2.12)))
  (type $ty-wasi:io/poll@0.2.12 (;1;)
    (instance
      (export (;0;) "pollable" (type (sub resource)))
      (type (;1;) (borrow 0))
      (type (;2;) (func (param "self" 1) (result bool)))
      (export (;0;) "[method]pollable.ready" (func (type 2)))
      (type (;3;) (func (param "self" 1)))
      (export (;1;) "[method]pollable.block" (func (type 3)))
      (type (;4;) (list 1))
      (type (;5;) (list u32))
      (type (;6;) (func (param "in" 4) (result 5)))
      (export (;2;) "poll" (func (type 6)))
    )
  )
  (import "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12 (;1;) (type $ty-wasi:io/poll@0.2.12)))
  (alias export $wasi:io/error@0.2.12 "error" (type $error (;2;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $pollable (;3;)))
  (type $ty-wasi:io/streams@0.2.12 (;4;)
    (instance
      (export (;0;) "input-stream" (type (sub resource)))
      (alias outer 1 $error (type (;1;)))
      (export (;2;) "error" (type (eq 1)))
      (type (;3;) (own 2))
      (type (;4;) (variant (case "last-operation-failed" 3) (case "closed")))
      (export (;5;) "stream-error" (type (eq 4)))
      (alias outer 1 $pollable (type (;6;)))
      (export (;7;) "pollable" (type (eq 6)))
      (export (;8;) "output-stream" (type (sub resource)))
      (type (;9;) (borrow 0))
      (type (;10;) (list u8))
      (type (;11;) (result 10 (error 5)))
      (type (;12;) (func (param "self" 9) (param "len" u64) (result 11)))
      (export (;0;) "[method]input-stream.read" (func (type 12)))
      (export (;1;) "[method]input-stream.blocking-read" (func (type 12)))
      (type (;13;) (result u64 (error 5)))
      (type (;14;) (func (param "self" 9) (param "len" u64) (result 13)))
      (export (;2;) "[method]input-stream.skip" (func (type 14)))
      (export (;3;) "[method]input-stream.blocking-skip" (func (type 14)))
      (type (;15;) (own 7))
      (type (;16;) (func (param "self" 9) (result 15)))
      (export (;4;) "[method]input-stream.subscribe" (func (type 16)))
      (type (;17;) (borrow 8))
      (type (;18;) (func (param "self" 17) (result 13)))
      (export (;5;) "[method]output-stream.check-write" (func (type 18)))
      (type (;19;) (result (error 5)))
      (type (;20;) (func (param "self" 17) (param "contents" 10) (result 19)))
      (export (;6;) "[method]output-stream.write" (func (type 20)))
      (export (;7;) "[method]output-stream.blocking-write-and-flush" (func (type 20)))
      (type (;21;) (func (param "self" 17) (result 19)))
      (export (;8;) "[method]output-stream.flush" (func (type 21)))
      (export (;9;) "[method]output-stream.blocking-flush" (func (type 21)))
      (type (;22;) (func (param "self" 17) (result 15)))
      (export (;10;) "[method]output-stream.subscribe" (func (type 22)))
      (type (;23;) (func (param "self" 17) (param "len" u64) (result 19)))
      (export (;11;) "[method]output-stream.write-zeroes" (func (type 23)))
      (export (;12;) "[method]output-stream.blocking-write-zeroes-and-flush" (func (type 23)))
      (type (;24;) (func (param "self" 17) (param "src" 9) (param "len" u64) (result 13)))
      (export (;13;) "[method]output-stream.splice" (func (type 24)))
      (export (;14;) "[method]output-stream.blocking-splice" (func (type 24)))
    )
  )
  (import "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12 (;2;) (type $ty-wasi:io/streams@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $input-stream (;5;)))
  (type $ty-wasi:cli/stdin@0.2.12 (;6;)
    (instance
      (alias outer 1 $input-stream (type (;0;)))
      (export (;1;) "input-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stdin" (func (type 3)))
    )
  )
  (import "wasi:cli/stdin@0.2.12" (instance $wasi:cli/stdin@0.2.12 (;3;) (type $ty-wasi:cli/stdin@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $output-stream (;7;)))
  (type $ty-wasi:cli/stdout@0.2.12 (;8;)
    (instance
      (alias outer 1 $output-stream (type (;0;)))
      (export (;1;) "output-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stdout" (func (type 3)))
    )
  )
  (import "wasi:cli/stdout@0.2.12" (instance $wasi:cli/stdout@0.2.12 (;4;) (type $ty-wasi:cli/stdout@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type9 output-stream" (@name "output-stream") (;9;)))
  (type $ty-wasi:cli/stderr@0.2.12 (;10;)
    (instance
      (alias outer 1 $"#type9 output-stream" (type (;0;)))
      (export (;1;) "output-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stderr" (func (type 3)))
    )
  )
  (import "wasi:cli/stderr@0.2.12" (instance $wasi:cli/stderr@0.2.12 (;5;) (type $ty-wasi:cli/stderr@0.2.12)))
  (type $ty-wasi:cli/terminal-input@0.2.12 (;11;)
    (instance
      (export (;0;) "terminal-input" (type (sub resource)))
    )
  )
  (import "wasi:cli/terminal-input@0.2.12" (instance $wasi:cli/terminal-input@0.2.12 (;6;) (type $ty-wasi:cli/terminal-input@0.2.12)))
  (type $ty-wasi:cli/terminal-output@0.2.12 (;12;)
    (instance
      (export (;0;) "terminal-output" (type (sub resource)))
    )
  )
  (import "wasi:cli/terminal-output@0.2.12" (instance $wasi:cli/terminal-output@0.2.12 (;7;) (type $ty-wasi:cli/terminal-output@0.2.12)))
  (alias export $wasi:cli/terminal-input@0.2.12 "terminal-input" (type $terminal-input (;13;)))
  (type $ty-wasi:cli/terminal-stdin@0.2.12 (;14;)
    (instance
      (alias outer 1 $terminal-input (type (;0;)))
      (export (;1;) "terminal-input" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (option 2))
      (type (;4;) (func (result 3)))
      (export (;0;) "get-terminal-stdin" (func (type 4)))
    )
  )
  (import "wasi:cli/terminal-stdin@0.2.12" (instance $wasi:cli/terminal-stdin@0.2.12 (;8;) (type $ty-wasi:cli/terminal-stdin@0.2.12)))
  (alias export $wasi:cli/terminal-output@0.2.12 "terminal-output" (type $terminal-output (;15;)))
  (type $ty-wasi:cli/terminal-stdout@0.2.12 (;16;)
    (instance
      (alias outer 1 $terminal-output (type (;0;)))
      (export (;1;) "terminal-output" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (option 2))
      (type (;4;) (func (result 3)))
      (export (;0;) "get-terminal-stdout" (func (type 4)))
    )
  )
  (import "wasi:cli/terminal-stdout@0.2.12" (instance $wasi:cli/terminal-stdout@0.2.12 (;9;) (type $ty-wasi:cli/terminal-stdout@0.2.12)))
  (alias export $wasi:cli/terminal-output@0.2.12 "terminal-output" (type $"#type17 terminal-output" (@name "terminal-output") (;17;)))
  (type $ty-wasi:cli/terminal-stderr@0.2.12 (;18;)
    (instance
      (alias outer 1 $"#type17 terminal-output" (type (;0;)))
      (export (;1;) "terminal-output" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (option 2))
      (type (;4;) (func (result 3)))
      (export (;0;) "get-terminal-stderr" (func (type 4)))
    )
  )
  (import "wasi:cli/terminal-stderr@0.2.12" (instance $wasi:cli/terminal-stderr@0.2.12 (;10;) (type $ty-wasi:cli/terminal-stderr@0.2.12)))
  (core module $main (;0;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32)))
    (type (;2;) (func (param i32) (result i32)))
    (type (;3;) (func (param i32 i32 i32)))
    (type (;4;) (func (param i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32 i32)))
    (type (;6;) (func (param i32 i32 i64 i32)))
    (type (;7;) (func (result i32)))
    (type (;8;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;9;) (func))
    (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func (;0;) (type 0)))
    (import "wasi:io/error@0.2.12" "[resource-drop]error" (func (;1;) (type 1)))
    (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func (;2;) (type 2)))
    (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func (;3;) (type 1)))
    (import "wasi:io/poll@0.2.12" "poll" (func (;4;) (type 3)))
    (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func (;5;) (type 1)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.read" (func (;6;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-read" (func (;7;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.skip" (func (;8;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-skip" (func (;9;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.subscribe" (func (;10;) (type 2)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.check-write" (func (;11;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.write" (func (;12;) (type 5)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-and-flush" (func (;13;) (type 5)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.flush" (func (;14;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-flush" (func (;15;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.subscribe" (func (;16;) (type 2)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.write-zeroes" (func (;17;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-zeroes-and-flush" (func (;18;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.splice" (func (;19;) (type 6)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-splice" (func (;20;) (type 6)))
    (import "wasi:io/streams@0.2.12" "[resource-drop]input-stream" (func (;21;) (type 1)))
    (import "wasi:io/streams@0.2.12" "[resource-drop]output-stream" (func (;22;) (type 1)))
    (import "wasi:cli/stdin@0.2.12" "get-stdin" (func (;23;) (type 7)))
    (import "wasi:cli/stdout@0.2.12" "get-stdout" (func (;24;) (type 7)))
    (import "wasi:cli/stderr@0.2.12" "get-stderr" (func (;25;) (type 7)))
    (import "wasi:cli/terminal-input@0.2.12" "[resource-drop]terminal-input" (func (;26;) (type 1)))
    (import "wasi:cli/terminal-output@0.2.12" "[resource-drop]terminal-output" (func (;27;) (type 1)))
    (import "wasi:cli/terminal-stdin@0.2.12" "get-terminal-stdin" (func (;28;) (type 1)))
    (import "wasi:cli/terminal-stdout@0.2.12" "get-terminal-stdout" (func (;29;) (type 1)))
    (import "wasi:cli/terminal-stderr@0.2.12" "get-terminal-stderr" (func (;30;) (type 1)))
    (memory (;0;) 1)
    (global $heap (;0;) (mut i32) i32.const 8192)
    (export "terminal-probe" (func 31))
    (export "cabi_post_terminal-probe" (func 32))
    (export "write-unchecked" (func 33))
    (export "cabi_post_write-unchecked" (func 34))
    (export "write-overpermit" (func 35))
    (export "cabi_post_write-overpermit" (func 36))
    (export "blocking-too-large" (func 37))
    (export "cabi_post_blocking-too-large" (func 38))
    (export "parent-drop" (func 39))
    (export "cabi_post_parent-drop" (func 40))
    (export "exercise" (func 41))
    (export "cabi_post_exercise" (func 42))
    (export "error-probe" (func 43))
    (export "cabi_post_error-probe" (func 44))
    (export "write-batch" (func 45))
    (export "cabi_post_write-batch" (func 46))
    (export "memory" (memory 0))
    (export "cabi_realloc" (func 47))
    (export "_initialize" (func 48))
    (func (;31;) (type 7) (result i32)
      (local $out i32)
      i32.const 64
      call 28
      i32.const 64
      i32.load8_u
      if ;; label = @1
        i32.const 68
        i32.load
        call 26
        local.get $out
        i32.const 1
        i32.add
        local.set $out
      end
      i32.const 64
      call 29
      i32.const 64
      i32.load8_u
      if ;; label = @1
        i32.const 68
        i32.load
        call 27
        local.get $out
        i32.const 1
        i32.add
        local.set $out
      end
      i32.const 64
      call 30
      i32.const 64
      i32.load8_u
      if ;; label = @1
        i32.const 68
        i32.load
        call 27
        local.get $out
        i32.const 1
        i32.add
        local.set $out
      end
      local.get $out
    )
    (func (;32;) (type 1) (param i32))
    (func (;33;) (type 7) (result i32)
      (local $out i32)
      call 24
      local.set $out
      local.get $out
      i32.const 128
      i32.const 3
      i32.const 0
      call 12
      i32.const 0
    )
    (func (;34;) (type 1) (param i32))
    (func (;35;) (type 7) (result i32)
      (local $out i32)
      call 24
      local.set $out
      local.get $out
      i32.const 0
      call 11
      i32.const 0
      call $ok
      local.get $out
      i64.const -1
      i32.const 0
      call 17
      i32.const 0
    )
    (func (;36;) (type 1) (param i32))
    (func (;37;) (type 7) (result i32)
      (local $out i32)
      call 24
      local.set $out
      local.get $out
      i64.const 4097
      i32.const 0
      call 18
      i32.const 0
    )
    (func (;38;) (type 1) (param i32))
    (func (;39;) (type 7) (result i32)
      (local $out i32)
      call 24
      local.set $out
      local.get $out
      call 16
      drop
      local.get $out
      call 22
      i32.const 0
    )
    (func (;40;) (type 1) (param i32))
    (func (;41;) (type 7) (result i32)
      (local $in i32) (local $out i32) (local $err i32) (local $poll i32)
      call 23
      local.set $in
      call 24
      local.set $out
      call 25
      local.set $err
      local.get $in
      call 10
      local.tee $poll
      call 3
      local.get $poll
      call 2
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $poll
      call 5
      local.get $in
      i64.const 0
      i32.const 0
      call 6
      i32.const 0
      call $ok
      i32.const 8
      i32.load
      if ;; label = @1
        unreachable
      end
      local.get $in
      i64.const 1
      i32.const 0
      call 8
      i32.const 0
      call $ok
      i32.const 8
      i64.load
      i64.const 1
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $in
      i64.const 2
      i32.const 0
      call 7
      i32.const 0
      call $ok
      i32.const 8
      i32.load
      i32.const 2
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 4
      i32.load
      i32.load16_u
      i32.const 25442
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $in
      i64.const 1
      i32.const 0
      call 9
      i32.const 0
      call $ok
      i32.const 8
      i64.load
      i64.const 1
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $out
      i32.const 0
      call 11
      i32.const 0
      call $ok
      i32.const 8
      i64.load
      i64.const 3
      i64.lt_u
      if ;; label = @1
        unreachable
      end
      local.get $out
      i32.const 128
      i32.const 3
      i32.const 0
      call 12
      i32.const 0
      call $ok
      local.get $out
      i32.const 0
      call 14
      i32.const 0
      call $ok
      local.get $out
      i32.const 0
      call 15
      i32.const 0
      call $ok
      local.get $out
      call 16
      local.tee $poll
      call 3
      local.get $poll
      call 2
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $poll
      call 5
      local.get $out
      i32.const 0
      call 11
      i32.const 0
      call $ok
      local.get $out
      i64.const 2
      i32.const 0
      call 17
      i32.const 0
      call $ok
      local.get $out
      i64.const 3
      i32.const 0
      call 18
      i32.const 0
      call $ok
      local.get $out
      i32.const 131
      i32.const 1
      i32.const 0
      call 13
      i32.const 0
      call $ok
      local.get $out
      local.get $in
      i64.const 2
      i32.const 0
      call 19
      i32.const 0
      call $ok
      i32.const 8
      i64.load
      i64.const 2
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $out
      local.get $in
      i64.const 1
      i32.const 0
      call 20
      i32.const 0
      call $ok
      i32.const 8
      i64.load
      i64.const 1
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $out
      i32.const 0
      call 15
      i32.const 0
      call $ok
      local.get $in
      i64.const 10
      i32.const 0
      call 6
      i32.const 0
      call $ok
      i32.const 8
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 4
      i32.load
      i32.load8_u
      i32.const 104
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $in
      i64.const 1
      i32.const 0
      call 7
      i32.const 0
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 4
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $in
      call 21
      local.get $out
      call 22
      local.get $err
      call 22
      i32.const 42
    )
    (func (;42;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;43;) (type 7) (result i32)
      (local $out i32) (local $error i32)
      call 24
      local.set $out
      local.get $out
      i32.const 131
      i32.const 1
      i32.const 0
      call 13
      i32.const 0
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 4
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 8
      i32.load
      local.set $error
      local.get $error
      i32.const 16
      call 0
      i32.const 20
      i32.load
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $error
      call 1
      local.get $out
      i32.const 0
      call 11
      i32.const 0
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 8
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $out
      call 22
      i32.const 1
    )
    (func (;44;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;45;) (type 2) (param $count i32) (result i32)
      (local $out i32) (local $at i32)
      call 24
      local.set $out
      block $end
        loop $next
          local.get $at
          local.get $count
          i32.ge_u
          br_if $end
          local.get $out
          i32.const 1024
          i32.const 4096
          i32.const 0
          call 13
          i32.const 0
          call $ok
          local.get $at
          i32.const 1
          i32.add
          local.set $at
          br $next
        end
      end
      local.get $out
      call 22
      local.get $at
    )
    (func (;46;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;47;) (type 8) (param i32 i32 i32 i32) (result i32)
      (local $ptr i32)
      global.get $heap
      local.get 2
      i32.const 1
      i32.sub
      i32.add
      i32.const 0
      local.get 2
      i32.sub
      i32.and
      local.tee $ptr
      local.get 3
      i32.add
      global.set $heap
      local.get $ptr
    )
    (func (;48;) (type 9))
    (func $ok (;49;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (data (;0;) (i32.const 128) "XYZ!")
    (data (;1;) (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core module $wit-component-shim-module (;1;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32)))
    (table (;0;) 18 18 funcref)
    (export "0" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "2" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "3" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "4" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "5" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "6" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "7" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "8" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "9" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "10" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "11" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "12" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "13" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "14" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "15" (func $indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin))
    (export "16" (func $indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout))
    (export "17" (func $indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr))
    (export "$imports" (table 0))
    (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 0
      call_indirect (type 0)
    )
    (func $indirect-wasi:io/poll@0.2.12-poll (;1;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 1
      call_indirect (type 1)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (;2;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 2
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (;3;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 3
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (;4;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 4
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (;5;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 5
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;6;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 6
      call_indirect (type 0)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (;7;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 7
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (;8;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 8
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;9;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 9
      call_indirect (type 0)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;10;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 10
      call_indirect (type 0)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (;11;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 11
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (;12;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 12
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (;13;) (type 4) (param i32 i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 13
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (;14;) (type 4) (param i32 i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 14
      call_indirect (type 4)
    )
    (func $indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin (;15;) (type 5) (param i32)
      local.get 0
      i32.const 15
      call_indirect (type 5)
    )
    (func $indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout (;16;) (type 5) (param i32)
      local.get 0
      i32.const 16
      call_indirect (type 5)
    )
    (func $indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr (;17;) (type 5) (param i32)
      local.get 0
      i32.const 17
      call_indirect (type 5)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias core export $wit-component-shim-instance "0" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type19 error" (@name "error") (;19;)))
  (core func $resource.drop (;1;) (canon resource.drop $"#type19 error"))
  (core instance $wasi:io/error@0.2.12 (;1;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $resource.drop))
  )
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;2;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;3;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "1" (core func $indirect-wasi:io/poll@0.2.12-poll (;4;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type20 pollable" (@name "pollable") (;20;)))
  (core func $"#core-func5 resource.drop" (@name "resource.drop") (;5;) (canon resource.drop $"#type20 pollable"))
  (core instance $wasi:io/poll@0.2.12 (;2;)
    (export "[method]pollable.ready" (func $"[method]pollable.ready"))
    (export "[method]pollable.block" (func $"[method]pollable.block"))
    (export "poll" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "[resource-drop]pollable" (func $"#core-func5 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "2" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (;6;)))
  (alias core export $wit-component-shim-instance "3" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (;7;)))
  (alias core export $wit-component-shim-instance "4" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (;8;)))
  (alias core export $wit-component-shim-instance "5" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (;9;)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.subscribe" (func $"[method]input-stream.subscribe" (;2;)))
  (core func $"[method]input-stream.subscribe" (;10;) (canon lower (func $"[method]input-stream.subscribe")))
  (alias core export $wit-component-shim-instance "6" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;11;)))
  (alias core export $wit-component-shim-instance "7" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (;12;)))
  (alias core export $wit-component-shim-instance "8" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (;13;)))
  (alias core export $wit-component-shim-instance "9" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;14;)))
  (alias core export $wit-component-shim-instance "10" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;15;)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.subscribe" (func $"[method]output-stream.subscribe" (;3;)))
  (core func $"[method]output-stream.subscribe" (;16;) (canon lower (func $"[method]output-stream.subscribe")))
  (alias core export $wit-component-shim-instance "11" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (;17;)))
  (alias core export $wit-component-shim-instance "12" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (;18;)))
  (alias core export $wit-component-shim-instance "13" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (;19;)))
  (alias core export $wit-component-shim-instance "14" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (;20;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type21 input-stream" (@name "input-stream") (;21;)))
  (core func $"#core-func21 resource.drop" (@name "resource.drop") (;21;) (canon resource.drop $"#type21 input-stream"))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type22 output-stream" (@name "output-stream") (;22;)))
  (core func $"#core-func22 resource.drop" (@name "resource.drop") (;22;) (canon resource.drop $"#type22 output-stream"))
  (core instance $wasi:io/streams@0.2.12 (;3;)
    (export "[method]input-stream.read" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "[method]input-stream.blocking-read" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "[method]input-stream.skip" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "[method]input-stream.blocking-skip" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "[method]input-stream.subscribe" (func $"[method]input-stream.subscribe"))
    (export "[method]output-stream.check-write" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "[method]output-stream.write" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "[method]output-stream.blocking-write-and-flush" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "[method]output-stream.flush" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "[method]output-stream.blocking-flush" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "[method]output-stream.subscribe" (func $"[method]output-stream.subscribe"))
    (export "[method]output-stream.write-zeroes" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "[method]output-stream.blocking-write-zeroes-and-flush" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "[method]output-stream.splice" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "[method]output-stream.blocking-splice" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "[resource-drop]input-stream" (func $"#core-func21 resource.drop"))
    (export "[resource-drop]output-stream" (func $"#core-func22 resource.drop"))
  )
  (alias export $wasi:cli/stdin@0.2.12 "get-stdin" (func $get-stdin (;4;)))
  (core func $get-stdin (;23;) (canon lower (func $get-stdin)))
  (core instance $wasi:cli/stdin@0.2.12 (;4;)
    (export "get-stdin" (func $get-stdin))
  )
  (alias export $wasi:cli/stdout@0.2.12 "get-stdout" (func $get-stdout (;5;)))
  (core func $get-stdout (;24;) (canon lower (func $get-stdout)))
  (core instance $wasi:cli/stdout@0.2.12 (;5;)
    (export "get-stdout" (func $get-stdout))
  )
  (alias export $wasi:cli/stderr@0.2.12 "get-stderr" (func $get-stderr (;6;)))
  (core func $get-stderr (;25;) (canon lower (func $get-stderr)))
  (core instance $wasi:cli/stderr@0.2.12 (;6;)
    (export "get-stderr" (func $get-stderr))
  )
  (alias export $wasi:cli/terminal-input@0.2.12 "terminal-input" (type $"#type23 terminal-input" (@name "terminal-input") (;23;)))
  (core func $"#core-func26 resource.drop" (@name "resource.drop") (;26;) (canon resource.drop $"#type23 terminal-input"))
  (core instance $wasi:cli/terminal-input@0.2.12 (;7;)
    (export "[resource-drop]terminal-input" (func $"#core-func26 resource.drop"))
  )
  (alias export $wasi:cli/terminal-output@0.2.12 "terminal-output" (type $"#type24 terminal-output" (@name "terminal-output") (;24;)))
  (core func $"#core-func27 resource.drop" (@name "resource.drop") (;27;) (canon resource.drop $"#type24 terminal-output"))
  (core instance $wasi:cli/terminal-output@0.2.12 (;8;)
    (export "[resource-drop]terminal-output" (func $"#core-func27 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "15" (core func $indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin (;28;)))
  (core instance $wasi:cli/terminal-stdin@0.2.12 (;9;)
    (export "get-terminal-stdin" (func $indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin))
  )
  (alias core export $wit-component-shim-instance "16" (core func $indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout (;29;)))
  (core instance $wasi:cli/terminal-stdout@0.2.12 (;10;)
    (export "get-terminal-stdout" (func $indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout))
  )
  (alias core export $wit-component-shim-instance "17" (core func $indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr (;30;)))
  (core instance $wasi:cli/terminal-stderr@0.2.12 (;11;)
    (export "get-terminal-stderr" (func $indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr))
  )
  (core instance $main (;12;) (instantiate $main
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12))
      (with "wasi:cli/stdin@0.2.12" (instance $wasi:cli/stdin@0.2.12))
      (with "wasi:cli/stdout@0.2.12" (instance $wasi:cli/stdout@0.2.12))
      (with "wasi:cli/stderr@0.2.12" (instance $wasi:cli/stderr@0.2.12))
      (with "wasi:cli/terminal-input@0.2.12" (instance $wasi:cli/terminal-input@0.2.12))
      (with "wasi:cli/terminal-output@0.2.12" (instance $wasi:cli/terminal-output@0.2.12))
      (with "wasi:cli/terminal-stdin@0.2.12" (instance $wasi:cli/terminal-stdin@0.2.12))
      (with "wasi:cli/terminal-stdout@0.2.12" (instance $wasi:cli/terminal-stdout@0.2.12))
      (with "wasi:cli/terminal-stderr@0.2.12" (instance $wasi:cli/terminal-stderr@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32)))
    (type (;6;) (func))
    (import "actual" "0" (func $0 (;0;) (type 0)))
    (import "actual" "1" (func $1 (;1;) (type 1)))
    (import "actual" "2" (func $2 (;2;) (type 2)))
    (import "actual" "3" (func $3 (;3;) (type 2)))
    (import "actual" "4" (func $4 (;4;) (type 2)))
    (import "actual" "5" (func $5 (;5;) (type 2)))
    (import "actual" "6" (func $6 (;6;) (type 0)))
    (import "actual" "7" (func $7 (;7;) (type 3)))
    (import "actual" "8" (func $8 (;8;) (type 3)))
    (import "actual" "9" (func $9 (;9;) (type 0)))
    (import "actual" "10" (func $10 (;10;) (type 0)))
    (import "actual" "11" (func $11 (;11;) (type 2)))
    (import "actual" "12" (func $12 (;12;) (type 2)))
    (import "actual" "13" (func $13 (;13;) (type 4)))
    (import "actual" "14" (func $14 (;14;) (type 4)))
    (import "actual" "15" (func $15 (;15;) (type 5)))
    (import "actual" "16" (func $16 (;16;) (type 5)))
    (import "actual" "17" (func $17 (;17;) (type 5)))
    (import "main" "_initialize" (func $_initialize (;18;) (type 6)))
    (import "shim" "$imports" (table (;0;) 18 18 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16 $17)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;7;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;31;)))
  (core func $"#core-func32 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;32;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;8;)))
  (core func $"#core-func33 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;33;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.read" (func $"[method]input-stream.read" (;9;)))
  (core func $"#core-func34 indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.read") (;34;) (canon lower (func $"[method]input-stream.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-read" (func $"[method]input-stream.blocking-read" (;10;)))
  (core func $"#core-func35 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read") (;35;) (canon lower (func $"[method]input-stream.blocking-read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.skip" (func $"[method]input-stream.skip" (;11;)))
  (core func $"#core-func36 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.skip") (;36;) (canon lower (func $"[method]input-stream.skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-skip" (func $"[method]input-stream.blocking-skip" (;12;)))
  (core func $"#core-func37 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip") (;37;) (canon lower (func $"[method]input-stream.blocking-skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.check-write" (func $"[method]output-stream.check-write" (;13;)))
  (core func $"#core-func38 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write") (;38;) (canon lower (func $"[method]output-stream.check-write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write" (func $"[method]output-stream.write" (;14;)))
  (core func $"#core-func39 indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write") (;39;) (canon lower (func $"[method]output-stream.write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-and-flush" (func $"[method]output-stream.blocking-write-and-flush" (;15;)))
  (core func $"#core-func40 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush") (;40;) (canon lower (func $"[method]output-stream.blocking-write-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.flush" (func $"[method]output-stream.flush" (;16;)))
  (core func $"#core-func41 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.flush") (;41;) (canon lower (func $"[method]output-stream.flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-flush" (func $"[method]output-stream.blocking-flush" (;17;)))
  (core func $"#core-func42 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush") (;42;) (canon lower (func $"[method]output-stream.blocking-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write-zeroes" (func $"[method]output-stream.write-zeroes" (;18;)))
  (core func $"#core-func43 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes") (;43;) (canon lower (func $"[method]output-stream.write-zeroes") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-zeroes-and-flush" (func $"[method]output-stream.blocking-write-zeroes-and-flush" (;19;)))
  (core func $"#core-func44 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush") (;44;) (canon lower (func $"[method]output-stream.blocking-write-zeroes-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.splice" (func $"[method]output-stream.splice" (;20;)))
  (core func $"#core-func45 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.splice") (;45;) (canon lower (func $"[method]output-stream.splice") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-splice" (func $"[method]output-stream.blocking-splice" (;21;)))
  (core func $"#core-func46 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice") (;46;) (canon lower (func $"[method]output-stream.blocking-splice") (memory $memory)))
  (alias export $wasi:cli/terminal-stdin@0.2.12 "get-terminal-stdin" (func $get-terminal-stdin (;22;)))
  (core func $"#core-func47 indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin" (@name "indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin") (;47;) (canon lower (func $get-terminal-stdin) (memory $memory)))
  (alias export $wasi:cli/terminal-stdout@0.2.12 "get-terminal-stdout" (func $get-terminal-stdout (;23;)))
  (core func $"#core-func48 indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout" (@name "indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout") (;48;) (canon lower (func $get-terminal-stdout) (memory $memory)))
  (alias export $wasi:cli/terminal-stderr@0.2.12 "get-terminal-stderr" (func $get-terminal-stderr (;24;)))
  (core func $"#core-func49 indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr" (@name "indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr") (;49;) (canon lower (func $get-terminal-stderr) (memory $memory)))
  (core instance $actual (;13;)
    (export "0" (func $"#core-func32 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $"#core-func33 indirect-wasi:io/poll@0.2.12-poll"))
    (export "2" (func $"#core-func34 indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "3" (func $"#core-func35 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "4" (func $"#core-func36 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "5" (func $"#core-func37 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "6" (func $"#core-func38 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "7" (func $"#core-func39 indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "8" (func $"#core-func40 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "9" (func $"#core-func41 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "10" (func $"#core-func42 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "11" (func $"#core-func43 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "12" (func $"#core-func44 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "13" (func $"#core-func45 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "14" (func $"#core-func46 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "15" (func $"#core-func47 indirect-wasi:cli/terminal-stdin@0.2.12-get-terminal-stdin"))
    (export "16" (func $"#core-func48 indirect-wasi:cli/terminal-stdout@0.2.12-get-terminal-stdout"))
    (export "17" (func $"#core-func49 indirect-wasi:cli/terminal-stderr@0.2.12-get-terminal-stderr"))
  )
  (core instance $fixup (;14;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;25;) (func (result u32)))
  (alias core export $main "terminal-probe" (core func $terminal-probe (;50;)))
  (alias core export $main "cabi_post_terminal-probe" (core func $cabi_post_terminal-probe (;51;)))
  (func $terminal-probe (;25;) (type 25) (canon lift (core func $terminal-probe) (post-return $cabi_post_terminal-probe)))
  (export $"#func26 terminal-probe" (@name "terminal-probe") (;26;) "terminal-probe" (func $terminal-probe))
  (alias core export $main "write-unchecked" (core func $write-unchecked (;52;)))
  (alias core export $main "cabi_post_write-unchecked" (core func $cabi_post_write-unchecked (;53;)))
  (func $write-unchecked (;27;) (type 25) (canon lift (core func $write-unchecked) (post-return $cabi_post_write-unchecked)))
  (export $"#func28 write-unchecked" (@name "write-unchecked") (;28;) "write-unchecked" (func $write-unchecked))
  (alias core export $main "write-overpermit" (core func $write-overpermit (;54;)))
  (alias core export $main "cabi_post_write-overpermit" (core func $cabi_post_write-overpermit (;55;)))
  (func $write-overpermit (;29;) (type 25) (canon lift (core func $write-overpermit) (post-return $cabi_post_write-overpermit)))
  (export $"#func30 write-overpermit" (@name "write-overpermit") (;30;) "write-overpermit" (func $write-overpermit))
  (alias core export $main "blocking-too-large" (core func $blocking-too-large (;56;)))
  (alias core export $main "cabi_post_blocking-too-large" (core func $cabi_post_blocking-too-large (;57;)))
  (func $blocking-too-large (;31;) (type 25) (canon lift (core func $blocking-too-large) (post-return $cabi_post_blocking-too-large)))
  (export $"#func32 blocking-too-large" (@name "blocking-too-large") (;32;) "blocking-too-large" (func $blocking-too-large))
  (alias core export $main "parent-drop" (core func $parent-drop (;58;)))
  (alias core export $main "cabi_post_parent-drop" (core func $cabi_post_parent-drop (;59;)))
  (func $parent-drop (;33;) (type 25) (canon lift (core func $parent-drop) (post-return $cabi_post_parent-drop)))
  (export $"#func34 parent-drop" (@name "parent-drop") (;34;) "parent-drop" (func $parent-drop))
  (alias core export $main "exercise" (core func $exercise (;60;)))
  (alias core export $main "cabi_post_exercise" (core func $cabi_post_exercise (;61;)))
  (func $exercise (;35;) (type 25) (canon lift (core func $exercise) (post-return $cabi_post_exercise)))
  (export $"#func36 exercise" (@name "exercise") (;36;) "exercise" (func $exercise))
  (alias core export $main "error-probe" (core func $error-probe (;62;)))
  (alias core export $main "cabi_post_error-probe" (core func $cabi_post_error-probe (;63;)))
  (func $error-probe (;37;) (type 25) (canon lift (core func $error-probe) (post-return $cabi_post_error-probe)))
  (export $"#func38 error-probe" (@name "error-probe") (;38;) "error-probe" (func $error-probe))
  (type (;26;) (func (param "count" u32) (result u32)))
  (alias core export $main "write-batch" (core func $write-batch (;64;)))
  (alias core export $main "cabi_post_write-batch" (core func $cabi_post_write-batch (;65;)))
  (func $write-batch (;39;) (type 26) (canon lift (core func $write-batch) (post-return $cabi_post_write-batch)))
  (export $"#func40 write-batch" (@name "write-batch") (;40;) "write-batch" (func $write-batch))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
)

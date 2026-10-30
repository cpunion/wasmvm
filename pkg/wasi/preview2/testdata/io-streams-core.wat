(module
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
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 128) "XYZ!")
 (data (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
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
  (func (;31;) (type 7) (result i32) (local $out i32) i32.const 64 call 28 i32.const 64 i32.load8_u if i32.const 68 i32.load call 26 local.get $out i32.const 1 i32.add local.set $out end i32.const 64 call 29 i32.const 64 i32.load8_u if i32.const 68 i32.load call 27 local.get $out i32.const 1 i32.add local.set $out end i32.const 64 call 30 i32.const 64 i32.load8_u if i32.const 68 i32.load call 27 local.get $out i32.const 1 i32.add local.set $out end local.get $out)
  (func (;32;) (type 1) (param i32))
  (func (;33;) (type 7) (result i32) (local $out i32) call 24 local.set $out local.get $out i32.const 128 i32.const 3 i32.const 0 call 12 i32.const 0)
  (func (;34;) (type 1) (param i32))
  (func (;35;) (type 7) (result i32) (local $out i32) call 24 local.set $out local.get $out i32.const 0 call 11 i32.const 0 call $ok local.get $out i64.const -1 i32.const 0 call 17 i32.const 0)
  (func (;36;) (type 1) (param i32))
  (func (;37;) (type 7) (result i32) (local $out i32) call 24 local.set $out local.get $out i64.const 4097 i32.const 0 call 18 i32.const 0)
  (func (;38;) (type 1) (param i32))
  (func (;39;) (type 7) (result i32) (local $out i32) call 24 local.set $out local.get $out call 16 drop local.get $out call 22 i32.const 0)
  (func (;40;) (type 1) (param i32))
  (func (;41;) (type 7) (result i32)
 (local $in i32) (local $out i32) (local $err i32) (local $poll i32)
 call 23 local.set $in call 24 local.set $out call 25 local.set $err
 local.get $in call 10 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end local.get $poll call 5
 local.get $in i64.const 0 i32.const 0 call 6 i32.const 0 call $ok
 i32.const 8 i32.load if unreachable end
 local.get $in i64.const 1 i32.const 0 call 8 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $in i64.const 2 i32.const 0 call 7 i32.const 0 call $ok
 i32.const 8 i32.load i32.const 2 i32.ne if unreachable end
 i32.const 4 i32.load i32.load16_u i32.const 25442 i32.ne if unreachable end
 local.get $in i64.const 1 i32.const 0 call 9 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $out i32.const 0 call 11 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 3 i64.lt_u if unreachable end
 local.get $out i32.const 128 i32.const 3 i32.const 0 call 12 i32.const 0 call $ok
 local.get $out i32.const 0 call 14 i32.const 0 call $ok
 local.get $out i32.const 0 call 15 i32.const 0 call $ok
 local.get $out call 16 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end local.get $poll call 5
 local.get $out i32.const 0 call 11 i32.const 0 call $ok
 local.get $out i64.const 2 i32.const 0 call 17 i32.const 0 call $ok
 local.get $out i64.const 3 i32.const 0 call 18 i32.const 0 call $ok
 local.get $out i32.const 131 i32.const 1 i32.const 0 call 13 i32.const 0 call $ok
 local.get $out local.get $in i64.const 2 i32.const 0 call 19 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 2 i64.ne if unreachable end
 local.get $out local.get $in i64.const 1 i32.const 0 call 20 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $out i32.const 0 call 15 i32.const 0 call $ok
 local.get $in i64.const 10 i32.const 0 call 6 i32.const 0 call $ok
 i32.const 8 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 4 i32.load i32.load8_u i32.const 104 i32.ne if unreachable end
 local.get $in i64.const 1 i32.const 0 call 7
 i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 4 i32.load8_u i32.const 1 i32.ne if unreachable end
 local.get $in call 21 local.get $out call 22 local.get $err call 22
 i32.const 42
 )
  (func (;42;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;43;) (type 7) (result i32)
 (local $out i32) (local $error i32)
 call 24 local.set $out
 local.get $out i32.const 131 i32.const 1 i32.const 0 call 13
 i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 4 i32.load8_u if unreachable end
 i32.const 8 i32.load local.set $error
 local.get $error i32.const 16 call 0
 i32.const 20 i32.load i32.eqz if unreachable end
 local.get $error call 1
 local.get $out i32.const 0 call 11
 i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 8 i32.load8_u i32.const 1 i32.ne if unreachable end
 local.get $out call 22 i32.const 1
 )
  (func (;44;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;45;) (type 2) (param $count i32) (result i32)
 (local $out i32) (local $at i32)
 call 24 local.set $out
 block $end loop $next
 local.get $at local.get $count i32.ge_u br_if $end
 local.get $out i32.const 1024 i32.const 4096 i32.const 0 call 13
 i32.const 0 call $ok
 local.get $at i32.const 1 i32.add local.set $at br $next
 end end
 local.get $out call 22 local.get $at
 )
  (func (;46;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;47;) (type 8) (param i32 i32 i32 i32) (result i32)
 (local $ptr i32)
 global.get $heap local.get 2 i32.const 1 i32.sub i32.add
 i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr
 local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;48;) (type 9))
  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\85\10\01A\02\01A&\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\00\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\01\02\03\00\00\05error\02\03\00\01\08pollable\01B(\02\03\02\01\02\04\00\05error\03\00\00\02\03\02\01\03\04\00\08pollable\03\00\02\01i\01\01q\02\15last-operation-failed\01\04\00\06closed\00\00\04\00\0cstream-error\03\00\05\04\00\0cinput-stream\03\01\04\00\0doutput-stream\03\01\01h\07\01p}\01j\01\0a\01\06\01@\02\04self\09\03lenw\00\0b\04\00\19[method]input-stream.read\01\0c\04\00\22[method]input-stream.blocking-read\01\0c\01j\01w\01\06\01@\02\04self\09\03lenw\00\0d\04\00\19[method]input-stream.skip\01\0e\04\00\22[method]input-stream.blocking-skip\01\0e\01i\03\01@\01\04self\09\00\0f\04\00\1e[method]input-stream.subscribe\01\10\01h\08\01@\01\04self\11\00\0d\04\00![method]output-stream.check-write\01\12\01j\00\01\06\01@\02\04self\11\08contents\0a\00\13\04\00\1b[method]output-stream.write\01\14\04\00.[method]output-stream.blocking-write-and-flush\01\14\01@\01\04self\11\00\13\04\00\1b[method]output-stream.flush\01\15\04\00$[method]output-stream.blocking-flush\01\15\01@\01\04self\11\00\0f\04\00\1f[method]output-stream.subscribe\01\16\01@\02\04self\11\03lenw\00\13\04\00\22[method]output-stream.write-zeroes\01\17\04\005[method]output-stream.blocking-write-zeroes-and-flush\01\17\01@\03\04self\11\03src\09\03lenw\00\0d\04\00\1c[method]output-stream.splice\01\18\04\00%[method]output-stream.blocking-splice\01\18\03\00\16wasi:io/streams@0.2.12\05\04\02\03\00\02\0cinput-stream\01B\05\02\03\02\01\05\04\00\0cinput-stream\03\00\00\01i\01\01@\00\00\02\04\00\09get-stdin\01\03\03\00\15wasi:cli/stdin@0.2.12\05\06\02\03\00\02\0doutput-stream\01B\05\02\03\02\01\07\04\00\0doutput-stream\03\00\00\01i\01\01@\00\00\02\04\00\0aget-stdout\01\03\03\00\16wasi:cli/stdout@0.2.12\05\08\01B\05\02\03\02\01\07\04\00\0doutput-stream\03\00\00\01i\01\01@\00\00\02\04\00\0aget-stderr\01\03\03\00\16wasi:cli/stderr@0.2.12\05\09\01B\01\04\00\0eterminal-input\03\01\03\00\1ewasi:cli/terminal-input@0.2.12\05\0a\01B\01\04\00\0fterminal-output\03\01\03\00\1fwasi:cli/terminal-output@0.2.12\05\0b\02\03\00\06\0eterminal-input\01B\06\02\03\02\01\0c\04\00\0eterminal-input\03\00\00\01i\01\01k\02\01@\00\00\03\04\00\12get-terminal-stdin\01\04\03\00\1ewasi:cli/terminal-stdin@0.2.12\05\0d\02\03\00\07\0fterminal-output\01B\06\02\03\02\01\0e\04\00\0fterminal-output\03\00\00\01i\01\01k\02\01@\00\00\03\04\00\13get-terminal-stdout\01\04\03\00\1fwasi:cli/terminal-stdout@0.2.12\05\0f\01B\06\02\03\02\01\0e\04\00\0fterminal-output\03\00\00\01i\01\01k\02\01@\00\00\03\04\00\13get-terminal-stderr\01\04\03\00\1fwasi:cli/terminal-stderr@0.2.12\05\10\01@\00\00y\04\00\0eterminal-probe\01\11\04\00\0fwrite-unchecked\01\11\04\00\10write-overpermit\01\11\04\00\12blocking-too-large\01\11\04\00\0bparent-drop\01\11\04\00\08exercise\01\11\04\00\0berror-probe\01\11\01@\01\05county\00y\04\00\0bwrite-batch\01\12\04\00\12wasmvm:io-tests/io\04\00\0b\08\01\00\02io\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

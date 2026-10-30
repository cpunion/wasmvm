(module
  (type (;0;) (func (param i32 i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32)))
  (type (;4;) (func (param i32 i64 i32)))
  (type (;5;) (func (param i32 i32 i32 i32)))
  (type (;6;) (func (param i32 i32 i64 i32)))
  (type (;7;) (func (param i32 i64 i64 i32 i32)))
  (type (;8;) (func (param i32 i32 i64 i32 i32 i64 i32 i32)))
  (type (;9;) (func (param i32 i64 i64 i32)))
  (type (;10;) (func (param i32 i32 i32 i64 i32)))
  (type (;11;) (func (param i32 i32 i32 i32 i32)))
  (type (;12;) (func (param i32 i32 i32 i32 i32 i64 i32 i32 i64 i32 i32)))
  (type (;13;) (func (param i32 i32 i32 i32 i32 i32 i32 i32)))
  (type (;14;) (func (param i32 i32 i32 i32 i32 i32 i32)))
  (type (;15;) (func (param i32 i32 i32 i32 i32 i32)))
  (type (;16;) (func (param i32 i32) (result i32)))
  (type (;17;) (func (result i32)))
  (type (;18;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;19;) (func))
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
  (import "wasi:clocks/wall-clock@0.2.12" "now" (func (;23;) (type 1)))
  (import "wasi:clocks/wall-clock@0.2.12" "resolution" (func (;24;) (type 1)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read-via-stream" (func (;25;) (type 4)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.write-via-stream" (func (;26;) (type 4)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.append-via-stream" (func (;27;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.advise" (func (;28;) (type 7)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.sync-data" (func (;29;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.get-flags" (func (;30;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.get-type" (func (;31;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-size" (func (;32;) (type 4)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-times" (func (;33;) (type 8)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read" (func (;34;) (type 9)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.write" (func (;35;) (type 10)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read-directory" (func (;36;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.sync" (func (;37;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.create-directory-at" (func (;38;) (type 5)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.stat" (func (;39;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.stat-at" (func (;40;) (type 11)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-times-at" (func (;41;) (type 12)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.link-at" (func (;42;) (type 13)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.open-at" (func (;43;) (type 14)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.readlink-at" (func (;44;) (type 5)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.remove-directory-at" (func (;45;) (type 5)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.rename-at" (func (;46;) (type 14)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.symlink-at" (func (;47;) (type 15)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.unlink-file-at" (func (;48;) (type 5)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.is-same-object" (func (;49;) (type 16)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.metadata-hash" (func (;50;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[method]descriptor.metadata-hash-at" (func (;51;) (type 11)))
  (import "wasi:filesystem/types@0.2.12" "[method]directory-entry-stream.read-directory-entry" (func (;52;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "filesystem-error-code" (func (;53;) (type 0)))
  (import "wasi:filesystem/types@0.2.12" "[resource-drop]descriptor" (func (;54;) (type 1)))
  (import "wasi:filesystem/types@0.2.12" "[resource-drop]directory-entry-stream" (func (;55;) (type 1)))
  (import "wasi:filesystem/preopens@0.2.12" "get-directories" (func (;56;) (type 1)))
  (memory (;0;) 1)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "alpha")
 (data (i32.const 1040) "beta")
 (data (i32.const 1056) "gamma")
 (data (i32.const 1072) "sym")
 (data (i32.const 1088) "sub")
 (data (i32.const 1104) "abcdEF!")
 (data (i32.const 1120) "bench.bin")
  (export "exercise" (func 57))
  (export "cabi_post_exercise" (func 58))
  (export "stat-batch" (func 59))
  (export "cabi_post_stat-batch" (func 60))
  (export "driver" (func 61))
  (export "cabi_post_driver" (func 62))
  (export "error-probe" (func 63))
  (export "cabi_post_error-probe" (func 64))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 65))
  (export "_initialize" (func 66))
  (func (;57;) (type 17) (result i32) (local $root i32) (local $file i32) (local $stream i32) (local $poll i32) (local $entries i32) (local $count i32)
i32.const 0 call 56 i32.const 4 i32.load i32.const 1 i32.ne if unreachable end i32.const 0 i32.load i32.load local.set $root
 local.get $root i32.const 64 call 30 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 1 i32.and i32.eqz if unreachable end
 local.get $root i32.const 64 call 31 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 3 i32.ne if unreachable end
 local.get $root i32.const 1088 i32.const 3 i32.const 64 call 38 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 1 i32.const 3 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 local.get $file i32.const 1104 i32.const 4 i64.const 0 i32.const 64 call 35 i32.const 64 call $ok
 i32.const 72 i64.load i64.const 4 i64.ne if unreachable end
 local.get $file i64.const 4 i64.const 0 i32.const 64 call 34 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 4 i32.ne if unreachable end
 i32.const 68 i32.load i32.load i32.const 1684234849 i32.ne if unreachable end
 local.get $file i64.const 8 i32.const 64 call 32 i32.const 64 call $ok
 local.get $file i32.const 64 call 39 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 8 i64.ne if unreachable end
 local.get $file i32.const 64 call 31 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 6 i32.ne if unreachable end
 local.get $file i32.const 64 call 30 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 3 i32.and i32.const 3 i32.ne if unreachable end
 local.get $file i64.const 0 i64.const 8 i32.const 0 i32.const 64 call 28 i32.const 64 call $ok
 local.get $file i32.const 64 call 29 i32.const 64 call $ok
 local.get $file i32.const 64 call 37 i32.const 64 call $ok
 local.get $file i32.const 1 i64.const 0 i32.const 0 i32.const 1 i64.const 0 i32.const 0 i32.const 64 call 33 i32.const 64 call $ok
 local.get $file i64.const 0 i32.const 64 call 25 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 10 local.tee $poll call 3 local.get $poll call 5
 local.get $stream i64.const 4 i32.const 64 call 7 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 4 i32.ne if unreachable end
 local.get $stream call 21
 local.get $file i64.const 4 i32.const 64 call 26 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 1108 i32.const 2 i32.const 64 call 13 i32.const 64 call $ok
 local.get $stream call 22
 local.get $file i32.const 64 call 27 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 1110 i32.const 1 i32.const 64 call 13 i32.const 64 call $ok
 local.get $stream call 22
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 64 call 40 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 9 i64.ne if unreachable end
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 1 i64.const 0 i32.const 0 i32.const 1 i64.const 0 i32.const 0 i32.const 64 call 41 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 local.get $root i32.const 1040 i32.const 4 i32.const 64 call 42 i32.const 64 call $ok
 local.get $root i32.const 1040 i32.const 4 local.get $root i32.const 1056 i32.const 5 i32.const 64 call 46 i32.const 64 call $ok
 local.get $root i32.const 1024 i32.const 5 i32.const 1072 i32.const 3 i32.const 64 call 47 i32.const 64 call $ok
 local.get $root i32.const 1072 i32.const 3 i32.const 64 call 44 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 5 i32.ne if unreachable end
 local.get $root i32.const 1 i32.const 1072 i32.const 3 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 local.get $file i32.const 68 i32.load call 49 i32.eqz if unreachable end
 i32.const 68 i32.load call 54
 local.get $file i32.const 64 call 50 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 128 call 51 i32.const 128 call $ok
 i32.const 72 i64.load i32.const 136 i64.load i64.ne if unreachable end
 i32.const 80 i64.load i32.const 144 i64.load i64.ne if unreachable end
 local.get $root i32.const 64 call 36 i32.const 64 call $ok
 i32.const 68 i32.load local.set $entries
 block $end loop $next
 local.get $entries i32.const 64 call 52 i32.const 64 call $ok
 i32.const 68 i32.load8_u i32.eqz br_if $end
 local.get $count i32.const 1 i32.add local.set $count br $next end end
 local.get $count i32.const 4 i32.lt_u if unreachable end
 local.get $entries call 55
 local.get $root i32.const 1056 i32.const 5 i32.const 64 call 48 i32.const 64 call $ok
 local.get $root i32.const 1072 i32.const 3 i32.const 64 call 48 i32.const 64 call $ok
 local.get $file call 54
 local.get $root i32.const 1024 i32.const 5 i32.const 64 call 48 i32.const 64 call $ok
 local.get $root i32.const 1088 i32.const 3 i32.const 64 call 45 i32.const 64 call $ok
 local.get $root call 54 i32.const 42

 )
  (func (;58;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;59;) (type 2) (param $count i32) (result i32) (local $root i32) (local $file i32) (local $at i32)
i32.const 0 call 56 i32.const 4 i32.load i32.const 1 i32.ne if unreachable end i32.const 0 i32.load i32.load local.set $root
 local.get $root i32.const 0 i32.const 1120 i32.const 9 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 block $end loop $next
 local.get $at local.get $count i32.ge_u br_if $end
 local.get $file i32.const 64 call 39 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 16 i64.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end
 local.get $file call 54 local.get $root call 54 local.get $at

 )
  (func (;60;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;61;) (type 17) (result i32) (local $root i32)
i32.const 0 call 56 i32.const 4 i32.load i32.const 1 i32.ne if unreachable end i32.const 0 i32.load i32.load local.set $root local.get $root call 54 i32.const 1
 )
  (func (;62;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;63;) (type 17) (result i32) (local $root i32) (local $file i32) (local $stream i32) (local $error i32)
i32.const 0 call 56 i32.const 4 i32.load i32.const 1 i32.ne if unreachable end i32.const 0 i32.load i32.load local.set $root
 local.get $root i32.const 0 i32.const 1120 i32.const 9 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 local.get $file i64.const 0 i32.const 64 call 25 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i64.const 1 i32.const 64 call 6
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u if unreachable end
 i32.const 72 i32.load local.set $error
 local.get $error i32.const 128 call 53
 i32.const 128 i32.load8_u i32.const 1 i32.ne if unreachable end
 local.get $error call 1 local.get $stream call 21
 local.get $file call 54 local.get $root call 54 i32.const 1

 )
  (func (;64;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;65;) (type 18) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;66;) (type 19))
  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\9b$\01A\02\01A\19\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\00\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\01\02\03\00\00\05error\02\03\00\01\08pollable\01B(\02\03\02\01\02\04\00\05error\03\00\00\02\03\02\01\03\04\00\08pollable\03\00\02\01i\01\01q\02\15last-operation-failed\01\04\00\06closed\00\00\04\00\0cstream-error\03\00\05\04\00\0cinput-stream\03\01\04\00\0doutput-stream\03\01\01h\07\01p}\01j\01\0a\01\06\01@\02\04self\09\03lenw\00\0b\04\00\19[method]input-stream.read\01\0c\04\00\22[method]input-stream.blocking-read\01\0c\01j\01w\01\06\01@\02\04self\09\03lenw\00\0d\04\00\19[method]input-stream.skip\01\0e\04\00\22[method]input-stream.blocking-skip\01\0e\01i\03\01@\01\04self\09\00\0f\04\00\1e[method]input-stream.subscribe\01\10\01h\08\01@\01\04self\11\00\0d\04\00![method]output-stream.check-write\01\12\01j\00\01\06\01@\02\04self\11\08contents\0a\00\13\04\00\1b[method]output-stream.write\01\14\04\00.[method]output-stream.blocking-write-and-flush\01\14\01@\01\04self\11\00\13\04\00\1b[method]output-stream.flush\01\15\04\00$[method]output-stream.blocking-flush\01\15\01@\01\04self\11\00\0f\04\00\1f[method]output-stream.subscribe\01\16\01@\02\04self\11\03lenw\00\13\04\00\22[method]output-stream.write-zeroes\01\17\04\005[method]output-stream.blocking-write-zeroes-and-flush\01\17\01@\03\04self\11\03src\09\03lenw\00\0d\04\00\1c[method]output-stream.splice\01\18\04\00%[method]output-stream.blocking-splice\01\18\03\00\16wasi:io/streams@0.2.12\05\04\01B\05\01r\02\07secondsw\0bnanosecondsy\04\00\08datetime\03\00\00\01@\00\00\01\04\00\03now\01\02\04\00\0aresolution\01\02\03\00\1dwasi:clocks/wall-clock@0.2.12\05\05\02\03\00\02\0cinput-stream\02\03\00\02\0doutput-stream\02\03\00\02\05error\02\03\00\03\08datetime\01Br\02\03\02\01\06\04\00\0cinput-stream\03\00\00\02\03\02\01\07\04\00\0doutput-stream\03\00\02\02\03\02\01\08\04\00\05error\03\00\04\02\03\02\01\09\04\00\08datetime\03\00\06\01w\04\00\08filesize\03\00\08\01m\08\07unknown\0cblock-device\10character-device\09directory\04fifo\0dsymbolic-link\0cregular-file\06socket\04\00\0fdescriptor-type\03\00\0a\01n\06\04read\05write\13file-integrity-sync\13data-integrity-sync\14requested-write-sync\10mutate-directory\04\00\10descriptor-flags\03\00\0c\01n\01\0esymlink-follow\04\00\0apath-flags\03\00\0e\01n\04\06create\09directory\09exclusive\08truncate\04\00\0aopen-flags\03\00\10\01w\04\00\0alink-count\03\00\12\01k\07\01r\06\04type\0b\0alink-count\13\04size\09\15data-access-timestamp\14\1bdata-modification-timestamp\14\17status-change-timestamp\14\04\00\0fdescriptor-stat\03\00\15\01q\03\09no-change\00\00\03now\00\00\09timestamp\01\07\00\04\00\0dnew-timestamp\03\00\17\01r\02\04type\0b\04names\04\00\0fdirectory-entry\03\00\19\01m%\06access\0bwould-block\07already\0ebad-descriptor\04busy\08deadlock\05quota\05exist\0efile-too-large\15illegal-byte-sequence\0bin-progress\0binterrupted\07invalid\02io\0cis-directory\04loop\0etoo-many-links\0cmessage-size\0dname-too-long\09no-device\08no-entry\07no-lock\13insufficient-memory\12insufficient-space\0dnot-directory\09not-empty\0fnot-recoverable\0bunsupported\06no-tty\0eno-such-device\08overflow\0dnot-permitted\04pipe\09read-only\0cinvalid-seek\0etext-file-busy\0ccross-device\04\00\0aerror-code\03\00\1b\01m\06\06normal\0asequential\06random\09will-need\09dont-need\08no-reuse\04\00\06advice\03\00\1d\01r\02\05lowerw\05upperw\04\00\13metadata-hash-value\03\00\1f\04\00\0adescriptor\03\01\04\00\16directory-entry-stream\03\01\01h!\01i\01\01j\01$\01\1c\01@\02\04self#\06offset\09\00%\04\00\22[method]descriptor.read-via-stream\01&\01i\03\01j\01'\01\1c\01@\02\04self#\06offset\09\00(\04\00#[method]descriptor.write-via-stream\01)\01@\01\04self#\00(\04\00$[method]descriptor.append-via-stream\01*\01j\00\01\1c\01@\04\04self#\06offset\09\06length\09\06advice\1e\00+\04\00\19[method]descriptor.advise\01,\01@\01\04self#\00+\04\00\1c[method]descriptor.sync-data\01-\01j\01\0d\01\1c\01@\01\04self#\00.\04\00\1c[method]descriptor.get-flags\01/\01j\01\0b\01\1c\01@\01\04self#\000\04\00\1b[method]descriptor.get-type\011\01@\02\04self#\04size\09\00+\04\00\1b[method]descriptor.set-size\012\01@\03\04self#\15data-access-timestamp\18\1bdata-modification-timestamp\18\00+\04\00\1c[method]descriptor.set-times\013\01p}\01o\024\7f\01j\015\01\1c\01@\03\04self#\06length\09\06offset\09\006\04\00\17[method]descriptor.read\017\01j\01\09\01\1c\01@\03\04self#\06buffer4\06offset\09\008\04\00\18[method]descriptor.write\019\01i\22\01j\01:\01\1c\01@\01\04self#\00;\04\00![method]descriptor.read-directory\01<\04\00\17[method]descriptor.sync\01-\01@\02\04self#\04paths\00+\04\00&[method]descriptor.create-directory-at\01=\01j\01\16\01\1c\01@\01\04self#\00>\04\00\17[method]descriptor.stat\01?\01@\03\04self#\0apath-flags\0f\04paths\00>\04\00\1a[method]descriptor.stat-at\01@\01@\05\04self#\0apath-flags\0f\04paths\15data-access-timestamp\18\1bdata-modification-timestamp\18\00+\04\00\1f[method]descriptor.set-times-at\01A\01@\05\04self#\0eold-path-flags\0f\08old-paths\0enew-descriptor#\08new-paths\00+\04\00\1a[method]descriptor.link-at\01B\01i!\01j\01\c3\00\01\1c\01@\05\04self#\0apath-flags\0f\04paths\0aopen-flags\11\05flags\0d\00\c4\00\04\00\1a[method]descriptor.open-at\01E\01j\01s\01\1c\01@\02\04self#\04paths\00\c6\00\04\00\1e[method]descriptor.readlink-at\01G\04\00&[method]descriptor.remove-directory-at\01=\01@\04\04self#\08old-paths\0enew-descriptor#\08new-paths\00+\04\00\1c[method]descriptor.rename-at\01H\01@\03\04self#\08old-paths\08new-paths\00+\04\00\1d[method]descriptor.symlink-at\01I\04\00![method]descriptor.unlink-file-at\01=\01@\02\04self#\05other#\00\7f\04\00![method]descriptor.is-same-object\01J\01j\01 \01\1c\01@\01\04self#\00\cb\00\04\00 [method]descriptor.metadata-hash\01L\01@\03\04self#\0apath-flags\0f\04paths\00\cb\00\04\00#[method]descriptor.metadata-hash-at\01M\01h\22\01k\1a\01j\01\cf\00\01\1c\01@\01\04self\ce\00\00\d0\00\04\003[method]directory-entry-stream.read-directory-entry\01Q\01h\05\01k\1c\01@\01\03err\d2\00\00\d3\00\04\00\15filesystem-error-code\01T\03\00\1cwasi:filesystem/types@0.2.12\05\0a\02\03\00\04\0adescriptor\01B\07\02\03\02\01\0b\04\00\0adescriptor\03\00\00\01i\01\01o\02\02s\01p\03\01@\00\00\04\04\00\0fget-directories\01\05\03\00\1fwasi:filesystem/preopens@0.2.12\05\0c\01@\00\00y\04\00\08exercise\01\0d\01@\01\05county\00y\04\00\0astat-batch\01\0e\04\00\06driver\01\0d\04\00\0berror-probe\01\0d\04\00\22wasmvm:filesystem-tests/filesystem\04\00\0b\10\01\00\0afilesystem\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

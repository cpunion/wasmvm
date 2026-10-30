(module
  (type (;0;) (func (param i32 i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32)))
  (type (;4;) (func (result i32)))
  (type (;5;) (func (param i32 i32 i32 i32)))
  (type (;6;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;7;) (func))
  (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func (;0;) (type 0)))
  (import "wasi:io/error@0.2.12" "[resource-drop]error" (func (;1;) (type 1)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func (;2;) (type 2)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func (;3;) (type 1)))
  (import "wasi:io/poll@0.2.12" "poll" (func (;4;) (type 3)))
  (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func (;5;) (type 1)))
  (import "wasi:sockets/network@0.2.12" "[resource-drop]network" (func (;6;) (type 1)))
  (import "wasi:sockets/instance-network@0.2.12" "instance-network" (func (;7;) (type 4)))
  (import "wasi:sockets/ip-name-lookup@0.2.12" "[method]resolve-address-stream.resolve-next-address" (func (;8;) (type 0)))
  (import "wasi:sockets/ip-name-lookup@0.2.12" "[method]resolve-address-stream.subscribe" (func (;9;) (type 2)))
  (import "wasi:sockets/ip-name-lookup@0.2.12" "resolve-addresses" (func (;10;) (type 5)))
  (import "wasi:sockets/ip-name-lookup@0.2.12" "[resource-drop]resolve-address-stream" (func (;11;) (type 1)))
  (memory (;0;) 1)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "127.0.0.1")
 (data (i32.const 1040) "::1")
  (export "exercise" (func 12))
  (export "cabi_post_exercise" (func 13))
  (export "lookup-batch" (func 14))
  (export "cabi_post_lookup-batch" (func 15))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 16))
  (export "_initialize" (func 17))
  (func (;12;) (type 4) (result i32) (local $network i32) (local $stream i32) (local $poll i32)
call 7 local.set $network
 local.get $network i32.const 1024 i32.const 9 i32.const 64 call 10 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 9 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end
 local.get $stream i32.const 64 call 8 i32.const 64 call $ok
 i32.const 66 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u if unreachable end
 i32.const 70 i32.load i32.const 16777343 i32.ne if unreachable end
 local.get $stream i32.const 64 call 8 i32.const 64 call $ok
 i32.const 66 i32.load8_u if unreachable end
 local.get $poll call 5 local.get $stream call 11

 local.get $network i32.const 1040 i32.const 3 i32.const 64 call 10 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 9 local.tee $poll call 3
 local.get $stream i32.const 64 call 8 i32.const 64 call $ok
 i32.const 66 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 84 i32.load16_u i32.const 1 i32.ne if unreachable end
 local.get $poll call 5 local.get $stream call 11 local.get $network call 6 i32.const 42

 )
  (func (;13;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;14;) (type 2) (param $count i32) (result i32) (local $network i32) (local $stream i32) (local $poll i32) (local $at i32)
call 7 local.set $network
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end

 local.get $network i32.const 1024 i32.const 9 i32.const 64 call 10 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 9 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end
 local.get $stream i32.const 64 call 8 i32.const 64 call $ok
 i32.const 66 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u if unreachable end
 i32.const 70 i32.load i32.const 16777343 i32.ne if unreachable end
 local.get $stream i32.const 64 call 8 i32.const 64 call $ok
 i32.const 66 i32.load8_u if unreachable end
 local.get $poll call 5 local.get $stream call 11
local.get $at i32.const 1 i32.add local.set $at br $next end end
 local.get $network call 6 local.get $at
 )
  (func (;15;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;16;) (type 6) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;17;) (type 7))
  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\f3\0b\01A\02\01A\13\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\00\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\01\02\03\00\00\05error\01B\13\02\03\02\01\02\04\00\05error\03\00\00\04\00\07network\03\01\01m\15\07unknown\0daccess-denied\0dnot-supported\10invalid-argument\0dout-of-memory\07timeout\14concurrency-conflict\0fnot-in-progress\0bwould-block\0dinvalid-state\10new-socket-limit\14address-not-bindable\0eaddress-in-use\12remote-unreachable\12connection-refused\10connection-reset\12connection-aborted\12datagram-too-large\11name-unresolvable\1atemporary-resolver-failure\1apermanent-resolver-failure\04\00\0aerror-code\03\00\03\01m\02\04ipv4\04ipv6\04\00\11ip-address-family\03\00\05\01o\04}}}}\04\00\0cipv4-address\03\00\07\01o\08{{{{{{{{\04\00\0cipv6-address\03\00\09\01q\02\04ipv4\01\08\00\04ipv6\01\0a\00\04\00\0aip-address\03\00\0b\01r\02\04port{\07address\08\04\00\13ipv4-socket-address\03\00\0d\01r\04\04port{\09flow-infoy\07address\0a\08scope-idy\04\00\13ipv6-socket-address\03\00\0f\01q\02\04ipv4\01\0e\00\04ipv6\01\10\00\04\00\11ip-socket-address\03\00\11\03\00\1bwasi:sockets/network@0.2.12\05\03\02\03\00\02\07network\01B\05\02\03\02\01\04\04\00\07network\03\00\00\01i\01\01@\00\00\02\04\00\10instance-network\01\03\03\00$wasi:sockets/instance-network@0.2.12\05\05\02\03\00\01\08pollable\02\03\00\02\0aerror-code\02\03\00\02\0aip-address\01B\16\02\03\02\01\06\04\00\08pollable\03\00\00\02\03\02\01\04\04\00\07network\03\00\02\02\03\02\01\07\04\00\0aerror-code\03\00\04\02\03\02\01\08\04\00\0aip-address\03\00\06\04\00\16resolve-address-stream\03\01\01h\08\01k\07\01j\01\0a\01\05\01@\01\04self\09\00\0b\04\003[method]resolve-address-stream.resolve-next-address\01\0c\01i\01\01@\01\04self\09\00\0d\04\00([method]resolve-address-stream.subscribe\01\0e\01h\03\01i\08\01j\01\10\01\05\01@\02\07network\0f\04names\00\11\04\00\11resolve-addresses\01\12\03\00\22wasi:sockets/ip-name-lookup@0.2.12\05\09\01@\00\00y\04\00\08exercise\01\0a\01@\01\05county\00y\04\00\0clookup-batch\01\0b\04\00\1cwasmvm:network-tests/network\04\00\0b\0d\01\00\07network\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

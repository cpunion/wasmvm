(module
  (type (;0;) (func (param i32 i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32)))
  (type (;4;) (func (result i32)))
  (type (;5;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
  (type (;6;) (func (param i32 i64 i32)))
  (type (;7;) (func (param i32 i32 i32 i32)))
  (type (;8;) (func (param i32 i32) (result i32)))
  (type (;9;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;10;) (func))
  (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func $error.error.to-debug-string (;0;) (type 0)))
  (import "wasi:io/error@0.2.12" "[resource-drop]error" (func $error.drop-error (;1;) (type 1)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func $poll.pollable.ready (;2;) (type 2)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func $poll.pollable.block (;3;) (type 1)))
  (import "wasi:io/poll@0.2.12" "poll" (func $poll.poll (;4;) (type 3)))
  (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func $poll.drop-pollable (;5;) (type 1)))
  (import "wasi:sockets/network@0.2.12" "network-error-code" (func $network.network-error-code (;6;) (type 0)))
  (import "wasi:sockets/network@0.2.12" "[resource-drop]network" (func $network.drop-network (;7;) (type 1)))
  (import "wasi:sockets/instance-network@0.2.12" "instance-network" (func $instance-network.instance-network (;8;) (type 4)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.start-bind" (func $udp.udp-socket.start-bind (;9;) (type 5)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.finish-bind" (func $udp.udp-socket.finish-bind (;10;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.stream" (func $udp.udp-socket.stream (;11;) (type 5)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.local-address" (func $udp.udp-socket.local-address (;12;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.remote-address" (func $udp.udp-socket.remote-address (;13;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.address-family" (func $udp.udp-socket.address-family (;14;) (type 2)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.unicast-hop-limit" (func $udp.udp-socket.unicast-hop-limit (;15;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-unicast-hop-limit" (func $udp.udp-socket.set-unicast-hop-limit (;16;) (type 3)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.receive-buffer-size" (func $udp.udp-socket.receive-buffer-size (;17;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-receive-buffer-size" (func $udp.udp-socket.set-receive-buffer-size (;18;) (type 6)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.send-buffer-size" (func $udp.udp-socket.send-buffer-size (;19;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-send-buffer-size" (func $udp.udp-socket.set-send-buffer-size (;20;) (type 6)))
  (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.subscribe" (func $udp.udp-socket.subscribe (;21;) (type 2)))
  (import "wasi:sockets/udp@0.2.12" "[method]incoming-datagram-stream.receive" (func $udp.incoming-datagram-stream.receive (;22;) (type 6)))
  (import "wasi:sockets/udp@0.2.12" "[method]incoming-datagram-stream.subscribe" (func $udp.incoming-datagram-stream.subscribe (;23;) (type 2)))
  (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.check-send" (func $udp.outgoing-datagram-stream.check-send (;24;) (type 0)))
  (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.send" (func $udp.outgoing-datagram-stream.send (;25;) (type 7)))
  (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.subscribe" (func $udp.outgoing-datagram-stream.subscribe (;26;) (type 2)))
  (import "wasi:sockets/udp@0.2.12" "[resource-drop]udp-socket" (func $udp.drop-udp-socket (;27;) (type 1)))
  (import "wasi:sockets/udp@0.2.12" "[resource-drop]incoming-datagram-stream" (func $udp.drop-incoming-datagram-stream (;28;) (type 1)))
  (import "wasi:sockets/udp@0.2.12" "[resource-drop]outgoing-datagram-stream" (func $udp.drop-outgoing-datagram-stream (;29;) (type 1)))
  (import "wasi:sockets/udp-create-socket@0.2.12" "create-udp-socket" (func $udp-create-socket.create-udp-socket (;30;) (type 0)))
  (memory (;0;) 32 64)
 (global $heap (mut i32) (i32.const 8192))
 (global $network (mut i32) (i32.const 0))
 (global $client (mut i32) (i32.const 0))
 (global $server (mut i32) (i32.const 0))
 (global $ci (mut i32) (i32.const 0))
 (global $co (mut i32) (i32.const 0))
 (global $si (mut i32) (i32.const 0))
 (global $so (mut i32) (i32.const 0))
 (global $opened (mut i32) (i32.const 0))
 (global $connected (mut i32) (i32.const 0))
 (data (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
  (export "exercise" (func 31))
  (export "cabi_post_exercise" (func 32))
  (export "setup" (func 33))
  (export "cabi_post_setup" (func 34))
  (export "transfer" (func 35))
  (export "cabi_post_transfer" (func 36))
  (export "teardown" (func 37))
  (export "cabi_post_teardown" (func 38))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 39))
  (export "_initialize" (func 40))
  (func (;31;) (type 2) (param i32) (result i32)
local.get 0 i32.const 0 call $open global.get $co i32.const 288 i32.const 0 call $send global.get $si i32.const 0 call $receive i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close local.get 0 i32.const 1 call $open i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close i32.const 42
 )
  (func (;32;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;33;) (type 8) (param i32 i32) (result i32)
local.get 0 local.get 1 call $open i32.const 1
 )
  (func (;34;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;35;) (type 2) (param i32) (result i32)
local.get 0 call $exchange
 )
  (func (;36;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;37;) (type 4) (result i32)
call $close i32.const 1
 )
  (func (;38;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;39;) (type 9) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;40;) (type 10))
(func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $open (param $family i32) (param $connected i32) (local $cp i32) (local $sp i32) (local $poll i32)
 global.get $opened if unreachable end
 local.get $connected global.set $connected
 call $instance-network.instance-network global.set $network
 local.get $family i32.const 64 call $udp-create-socket.create-udp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $client
 local.get $family i32.const 64 call $udp-create-socket.create-udp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $server
 global.get $client call $udp.udp-socket.address-family local.get $family i32.ne if unreachable end
 global.get $client i32.const 32 i32.const 64 call $udp.udp-socket.set-unicast-hop-limit i32.const 64 call $ok
 global.get $client i32.const 64 call $udp.udp-socket.unicast-hop-limit i32.const 64 call $ok
global.get $client i64.const 65536 i32.const 64 call $udp.udp-socket.set-receive-buffer-size i32.const 64 call $ok
 global.get $client i32.const 64 call $udp.udp-socket.receive-buffer-size i32.const 64 call $ok
global.get $client i64.const 65536 i32.const 64 call $udp.udp-socket.set-send-buffer-size i32.const 64 call $ok
 global.get $client i32.const 64 call $udp.udp-socket.send-buffer-size i32.const 64 call $ok

 global.get $client global.get $network local.get $family i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $udp.udp-socket.start-bind i32.const 64 call $ok
 global.get $client i32.const 64 call $udp.udp-socket.finish-bind i32.const 64 call $ok
 global.get $server global.get $network local.get $family i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $udp.udp-socket.start-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $udp.udp-socket.finish-bind i32.const 64 call $ok
 global.get $client call $udp.udp-socket.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable
 global.get $client i32.const 64 call $udp.udp-socket.local-address i32.const 64 call $ok
 i32.const 256 i32.const 68 i32.const 32 memory.copy i32.const 72 i32.load16_u local.tee $cp i32.eqz if unreachable end
 global.get $server i32.const 64 call $udp.udp-socket.local-address i32.const 64 call $ok
 i32.const 288 i32.const 68 i32.const 32 memory.copy i32.const 72 i32.load16_u local.tee $sp i32.eqz if unreachable end
 global.get $client local.get $connected local.get $family local.get $sp local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $udp.udp-socket.stream i32.const 64 call $ok
 i32.const 68 i32.load global.set $ci i32.const 72 i32.load global.set $co
 global.get $server local.get $connected local.get $family local.get $cp local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $udp.udp-socket.stream i32.const 64 call $ok
 i32.const 68 i32.load global.set $si i32.const 72 i32.load global.set $so
 local.get $connected if
 global.get $client i32.const 64 call $udp.udp-socket.remote-address i32.const 64 call $ok
 i32.const 72 i32.load16_u local.get $sp i32.ne if unreachable end end
 global.get $co call $udp.outgoing-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable
 global.get $si i64.const 0 i32.const 64 call $udp.incoming-datagram-stream.receive i32.const 64 call $ok
 i32.const 72 i32.load if unreachable end
 i32.const 1 global.set $opened)
 (func $send (param $stream i32) (param $peer i32) (param $length i32) (local $poll i32)
 block $permitted loop $again
 local.get $stream i32.const 64 call $udp.outgoing-datagram-stream.check-send i32.const 64 call $ok
 i32.const 72 i64.load i64.const 0 i64.gt_u br_if $permitted
 local.get $stream call $udp.outgoing-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $again end end
 i32.const 512 i32.const 1024 i32.store i32.const 516 local.get $length i32.store
 i32.const 520 i32.const 1 i32.store8
 i32.const 524 local.get $peer i32.const 32 memory.copy
 local.get $stream i32.const 512 i32.const 1 i32.const 64 call $udp.outgoing-datagram-stream.send i32.const 64 call $ok
 i32.const 72 i64.load i64.const 1 i64.ne if unreachable end)
 (func $receive (param $stream i32) (param $length i32) (local $poll i32) (local $item i32) (local $data i32)
 block $received loop $again
 local.get $stream i64.const 1 i32.const 64 call $udp.incoming-datagram-stream.receive i32.const 64 call $ok
 i32.const 72 i32.load br_if $received
 local.get $stream call $udp.incoming-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $again end end
 i32.const 72 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load local.tee $item i32.const 4 i32.add i32.load local.get $length i32.ne if unreachable end
 local.get $length if
 local.get $item i32.load local.tee $data i32.load8_u i32.const 48 i32.ne if unreachable end
 local.get $data local.get $length i32.add i32.const 1 i32.sub i32.load8_u i32.const 1024 local.get $length i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end end)
 (func $exchange (param $count i32) (result i32) (local $at i32)
 global.get $opened i32.eqz if unreachable end
 local.get $count i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end
 global.get $co i32.const 288 i32.const 4096 call $send global.get $si i32.const 4096 call $receive
 global.get $so i32.const 256 i32.const 4096 call $send global.get $ci i32.const 4096 call $receive
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at)
 (func $close
 global.get $opened i32.eqz if unreachable end
 global.get $ci call $udp.drop-incoming-datagram-stream global.get $co call $udp.drop-outgoing-datagram-stream
 global.get $si call $udp.drop-incoming-datagram-stream global.get $so call $udp.drop-outgoing-datagram-stream
 global.get $client call $udp.drop-udp-socket global.get $server call $udp.drop-udp-socket
 global.get $network call $network.drop-network i32.const 0 global.set $opened)

  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\af\16\01A\02\01A\1b\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\00\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\01\02\03\00\00\05error\01B\17\02\03\02\01\02\04\00\05error\03\00\00\04\00\07network\03\01\01m\15\07unknown\0daccess-denied\0dnot-supported\10invalid-argument\0dout-of-memory\07timeout\14concurrency-conflict\0fnot-in-progress\0bwould-block\0dinvalid-state\10new-socket-limit\14address-not-bindable\0eaddress-in-use\12remote-unreachable\12connection-refused\10connection-reset\12connection-aborted\12datagram-too-large\11name-unresolvable\1atemporary-resolver-failure\1apermanent-resolver-failure\04\00\0aerror-code\03\00\03\01m\02\04ipv4\04ipv6\04\00\11ip-address-family\03\00\05\01o\04}}}}\04\00\0cipv4-address\03\00\07\01o\08{{{{{{{{\04\00\0cipv6-address\03\00\09\01q\02\04ipv4\01\08\00\04ipv6\01\0a\00\04\00\0aip-address\03\00\0b\01r\02\04port{\07address\08\04\00\13ipv4-socket-address\03\00\0d\01r\04\04port{\09flow-infoy\07address\0a\08scope-idy\04\00\13ipv6-socket-address\03\00\0f\01q\02\04ipv4\01\0e\00\04ipv6\01\10\00\04\00\11ip-socket-address\03\00\11\01h\01\01k\04\01@\01\03err\13\00\14\04\00\12network-error-code\01\15\03\00\1bwasi:sockets/network@0.2.12\05\03\02\03\00\02\07network\01B\05\02\03\02\01\04\04\00\07network\03\00\00\01i\01\01@\00\00\02\04\00\10instance-network\01\03\03\00$wasi:sockets/instance-network@0.2.12\05\05\02\03\00\01\08pollable\02\03\00\02\0aerror-code\02\03\00\02\11ip-socket-address\02\03\00\02\11ip-address-family\01BD\02\03\02\01\06\04\00\08pollable\03\00\00\02\03\02\01\04\04\00\07network\03\00\02\02\03\02\01\07\04\00\0aerror-code\03\00\04\02\03\02\01\08\04\00\11ip-socket-address\03\00\06\02\03\02\01\09\04\00\11ip-address-family\03\00\08\01p}\01r\02\04data\0a\0eremote-address\07\04\00\11incoming-datagram\03\00\0b\01k\07\01r\02\04data\0a\0eremote-address\0d\04\00\11outgoing-datagram\03\00\0e\04\00\0audp-socket\03\01\04\00\18incoming-datagram-stream\03\01\04\00\18outgoing-datagram-stream\03\01\01h\10\01h\03\01j\00\01\05\01@\03\04self\13\07network\14\0dlocal-address\07\00\15\04\00\1d[method]udp-socket.start-bind\01\16\01@\01\04self\13\00\15\04\00\1e[method]udp-socket.finish-bind\01\17\01i\11\01i\12\01o\02\18\19\01j\01\1a\01\05\01@\02\04self\13\0eremote-address\0d\00\1b\04\00\19[method]udp-socket.stream\01\1c\01j\01\07\01\05\01@\01\04self\13\00\1d\04\00 [method]udp-socket.local-address\01\1e\04\00![method]udp-socket.remote-address\01\1e\01@\01\04self\13\00\09\04\00![method]udp-socket.address-family\01\1f\01j\01}\01\05\01@\01\04self\13\00 \04\00$[method]udp-socket.unicast-hop-limit\01!\01@\02\04self\13\05value}\00\15\04\00([method]udp-socket.set-unicast-hop-limit\01\22\01j\01w\01\05\01@\01\04self\13\00#\04\00&[method]udp-socket.receive-buffer-size\01$\01@\02\04self\13\05valuew\00\15\04\00*[method]udp-socket.set-receive-buffer-size\01%\04\00#[method]udp-socket.send-buffer-size\01$\04\00'[method]udp-socket.set-send-buffer-size\01%\01i\01\01@\01\04self\13\00&\04\00\1c[method]udp-socket.subscribe\01'\01h\11\01p\0c\01j\01)\01\05\01@\02\04self(\0bmax-resultsw\00*\04\00([method]incoming-datagram-stream.receive\01+\01@\01\04self(\00&\04\00*[method]incoming-datagram-stream.subscribe\01,\01h\12\01@\01\04self-\00#\04\00+[method]outgoing-datagram-stream.check-send\01.\01p\0f\01@\02\04self-\09datagrams/\00#\04\00%[method]outgoing-datagram-stream.send\010\01@\01\04self-\00&\04\00*[method]outgoing-datagram-stream.subscribe\011\03\00\17wasi:sockets/udp@0.2.12\05\0a\02\03\00\04\0audp-socket\01B\0c\02\03\02\01\04\04\00\07network\03\00\00\02\03\02\01\07\04\00\0aerror-code\03\00\02\02\03\02\01\09\04\00\11ip-address-family\03\00\04\02\03\02\01\0b\04\00\0audp-socket\03\00\06\01i\07\01j\01\08\01\03\01@\01\0eaddress-family\05\00\09\04\00\11create-udp-socket\01\0a\03\00%wasi:sockets/udp-create-socket@0.2.12\05\0c\01@\01\06familyy\00y\04\00\08exercise\01\0d\01@\02\06familyy\09connected\7f\00y\04\00\05setup\01\0e\01@\01\05county\00y\04\00\08transfer\01\0f\01@\00\00y\04\00\08teardown\01\10\04\00\14wasmvm:udp-tests/udp\04\00\0b\09\01\00\03udp\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

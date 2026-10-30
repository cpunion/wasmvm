(module
  (type (;0;) (func (param i32 i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32)))
  (type (;4;) (func (param i32 i64 i32)))
  (type (;5;) (func (param i32 i32 i32 i32)))
  (type (;6;) (func (param i32 i32 i64 i32)))
  (type (;7;) (func (result i32)))
  (type (;8;) (func (result i64)))
  (type (;9;) (func (param i64) (result i32)))
  (type (;10;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
  (type (;11;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;12;) (func))
  (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func $error.error.to-debug-string (;0;) (type 0)))
  (import "wasi:io/error@0.2.12" "[resource-drop]error" (func $error.drop-error (;1;) (type 1)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func $poll.pollable.ready (;2;) (type 2)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func $poll.pollable.block (;3;) (type 1)))
  (import "wasi:io/poll@0.2.12" "poll" (func $poll.poll (;4;) (type 3)))
  (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func $poll.drop-pollable (;5;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.read" (func $streams.input-stream.read (;6;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-read" (func $streams.input-stream.blocking-read (;7;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.skip" (func $streams.input-stream.skip (;8;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-skip" (func $streams.input-stream.blocking-skip (;9;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.subscribe" (func $streams.input-stream.subscribe (;10;) (type 2)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.check-write" (func $streams.output-stream.check-write (;11;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write" (func $streams.output-stream.write (;12;) (type 5)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-and-flush" (func $streams.output-stream.blocking-write-and-flush (;13;) (type 5)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.flush" (func $streams.output-stream.flush (;14;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-flush" (func $streams.output-stream.blocking-flush (;15;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.subscribe" (func $streams.output-stream.subscribe (;16;) (type 2)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write-zeroes" (func $streams.output-stream.write-zeroes (;17;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-zeroes-and-flush" (func $streams.output-stream.blocking-write-zeroes-and-flush (;18;) (type 4)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.splice" (func $streams.output-stream.splice (;19;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-splice" (func $streams.output-stream.blocking-splice (;20;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]input-stream" (func $streams.drop-input-stream (;21;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]output-stream" (func $streams.drop-output-stream (;22;) (type 1)))
  (import "wasi:sockets/network@0.2.12" "[resource-drop]network" (func $network.drop-network (;23;) (type 1)))
  (import "wasi:sockets/instance-network@0.2.12" "instance-network" (func $instance-network.instance-network (;24;) (type 7)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "now" (func $monotonic-clock.now (;25;) (type 8)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "resolution" (func $monotonic-clock.resolution (;26;) (type 8)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-instant" (func $monotonic-clock.subscribe-instant (;27;) (type 9)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-duration" (func $monotonic-clock.subscribe-duration (;28;) (type 9)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-bind" (func $tcp.start-bind (;29;) (type 10)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-bind" (func $tcp.finish-bind (;30;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-connect" (func $tcp.start-connect (;31;) (type 10)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-connect" (func $tcp.finish-connect (;32;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-listen" (func $tcp.start-listen (;33;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-listen" (func $tcp.finish-listen (;34;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.accept" (func $tcp.accept (;35;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.local-address" (func $tcp.local-address (;36;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.remote-address" (func $tcp.remote-address (;37;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.is-listening" (func $tcp.is-listening (;38;) (type 2)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.address-family" (func $tcp.address-family (;39;) (type 2)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-listen-backlog-size" (func $tcp.set-listen-backlog-size (;40;) (type 4)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-enabled" (func $tcp.keep-alive-enabled (;41;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-enabled" (func $tcp.set-keep-alive-enabled (;42;) (type 3)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-idle-time" (func $tcp.keep-alive-idle-time (;43;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-idle-time" (func $tcp.set-keep-alive-idle-time (;44;) (type 4)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-interval" (func $tcp.keep-alive-interval (;45;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-interval" (func $tcp.set-keep-alive-interval (;46;) (type 4)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-count" (func $tcp.keep-alive-count (;47;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-count" (func $tcp.set-keep-alive-count (;48;) (type 3)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.hop-limit" (func $tcp.hop-limit (;49;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-hop-limit" (func $tcp.set-hop-limit (;50;) (type 3)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.receive-buffer-size" (func $tcp.receive-buffer-size (;51;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-receive-buffer-size" (func $tcp.set-receive-buffer-size (;52;) (type 4)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.send-buffer-size" (func $tcp.send-buffer-size (;53;) (type 0)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-send-buffer-size" (func $tcp.set-send-buffer-size (;54;) (type 4)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.subscribe" (func $tcp.subscribe (;55;) (type 2)))
  (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.shutdown" (func $tcp.shutdown (;56;) (type 3)))
  (import "wasi:sockets/tcp@0.2.12" "[resource-drop]tcp-socket" (func $tcp.drop-tcp-socket (;57;) (type 1)))
  (import "wasi:sockets/tcp-create-socket@0.2.12" "create-tcp-socket" (func $tcp-create-socket.create-tcp-socket (;58;) (type 0)))
  (memory (;0;) 32 64)
 (global $heap (mut i32) (i32.const 8192))
 (global $network (mut i32) (i32.const 0))
 (global $server (mut i32) (i32.const 0))
 (global $client (mut i32) (i32.const 0))
 (global $child (mut i32) (i32.const 0))
 (global $ci (mut i32) (i32.const 0))
 (global $co (mut i32) (i32.const 0))
 (global $si (mut i32) (i32.const 0))
 (global $so (mut i32) (i32.const 0))
 (global $opened (mut i32) (i32.const 0))
 (data (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
  (export "create-probe" (func 59))
  (export "cabi_post_create-probe" (func 60))
  (export "setup" (func 61))
  (export "cabi_post_setup" (func 62))
  (export "teardown" (func 63))
  (export "cabi_post_teardown" (func 64))
  (export "exercise" (func 65))
  (export "cabi_post_exercise" (func 66))
  (export "transfer" (func 67))
  (export "cabi_post_transfer" (func 68))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 69))
  (export "_initialize" (func 70))
  (func (;59;) (type 2) (param i32) (result i32)
local.get 0 i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 i32.load8_u if (result i32) i32.const 68 i32.load8_u i32.const 1 i32.add else i32.const 68 i32.load call $tcp.drop-tcp-socket i32.const 0 end
 )
  (func (;60;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;61;) (type 2) (param i32) (result i32)
local.get 0 call $open i32.const 1
 )
  (func (;62;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;63;) (type 7) (result i32)
call $close i32.const 1
 )
  (func (;64;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;65;) (type 2) (param i32) (result i32)
local.get 0 call $open i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close i32.const 42
 )
  (func (;66;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;67;) (type 2) (param i32) (result i32)
local.get 0 call $exchange
 )
  (func (;68;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;69;) (type 11) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;70;) (type 12))
(func $open (param $family i32) (local $poll i32) (local $port i32)
 global.get $opened if unreachable end
 call $instance-network.instance-network global.set $network
 local.get $family i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $server
 local.get $family i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $client
 global.get $client call $tcp.address-family local.get $family i32.ne if unreachable end
 global.get $client call $tcp.is-listening if unreachable end
global.get $client i32.const 1 i32.const 64 call $tcp.set-keep-alive-enabled i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.keep-alive-enabled i32.const 64 call $ok
global.get $client i64.const 1000000000 i32.const 64 call $tcp.set-keep-alive-idle-time i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.keep-alive-idle-time i32.const 64 call $ok
global.get $client i64.const 1000000000 i32.const 64 call $tcp.set-keep-alive-interval i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.keep-alive-interval i32.const 64 call $ok
global.get $client i32.const 3 i32.const 64 call $tcp.set-keep-alive-count i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.keep-alive-count i32.const 64 call $ok
global.get $client i32.const 32 i32.const 64 call $tcp.set-hop-limit i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.hop-limit i32.const 64 call $ok
global.get $client i64.const 65536 i32.const 64 call $tcp.set-receive-buffer-size i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.receive-buffer-size i32.const 64 call $ok
global.get $client i64.const 65536 i32.const 64 call $tcp.set-send-buffer-size i32.const 64 call $ok
 global.get $client i32.const 64 call $tcp.send-buffer-size i32.const 64 call $ok

 global.get $server i64.const 128 i32.const 64 call $tcp.set-listen-backlog-size i32.const 64 call $ok
 global.get $server global.get $network local.get $family i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $tcp.start-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.finish-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.start-listen i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.finish-listen i32.const 64 call $ok
 global.get $server call $tcp.is-listening i32.eqz if unreachable end
 global.get $server i32.const 64 call $tcp.local-address i32.const 64 call $ok
 i32.const 72 i32.load16_u local.tee $port i32.eqz if unreachable end
 global.get $client global.get $network local.get $family local.get $port local.get $family if (result i32) i32.const 0 else i32.const 127 end i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 0 else i32.const 1 end i32.const 0 i32.const 0 i32.const 0 i32.const 0 local.get $family if (result i32) i32.const 1 else i32.const 0 end i32.const 0 i32.const 64 call $tcp.start-connect i32.const 64 call $ok
 block $connected loop $connect
 global.get $client i32.const 64 call $tcp.finish-connect
 i32.const 64 i32.load8_u i32.eqz br_if $connected
 i32.const 68 i32.load8_u i32.const 8 i32.ne if unreachable end
 global.get $client call $tcp.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $connect end end
 i32.const 68 i32.load global.set $ci i32.const 72 i32.load global.set $co
 block $accepted loop $accept
 global.get $server i32.const 64 call $tcp.accept
 i32.const 64 i32.load8_u i32.eqz br_if $accepted
 i32.const 68 i32.load8_u i32.const 8 i32.ne if unreachable end
 global.get $server call $tcp.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $accept end end
 i32.const 68 i32.load global.set $child i32.const 72 i32.load global.set $si i32.const 76 i32.load global.set $so
 global.get $client i32.const 64 call $tcp.remote-address i32.const 64 call $ok
 global.get $child i32.const 64 call $tcp.local-address i32.const 64 call $ok
 i32.const 1 global.set $opened)
(func $read-all (param $stream i32) (local $at i32) (local $n i32) (local $ptr i32)
 block $end loop $next local.get $at i32.const 4096 i32.ge_u br_if $end
 local.get $stream i32.const 4096 local.get $at i32.sub i64.extend_i32_u i32.const 64 call $streams.input-stream.blocking-read i32.const 64 call $ok
 i32.const 68 i32.load local.set $ptr i32.const 72 i32.load local.tee $n i32.eqz if unreachable end
 local.get $ptr i32.load8_u i32.const 1024 local.get $at i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $ptr local.get $n i32.add i32.const 1 i32.sub i32.load8_u i32.const 1024 local.get $at local.get $n i32.add i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $at local.get $n i32.add local.set $at br $next end end)
 (func $exchange (param $count i32) (result i32) (local $at i32)
 global.get $opened i32.eqz if unreachable end
 local.get $count i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end
 global.get $co i32.const 1024 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 global.get $si call $read-all
 global.get $so i32.const 1024 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 global.get $ci call $read-all
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at)
 (func $close
 global.get $opened i32.eqz if unreachable end
 global.get $client i32.const 2 i32.const 64 call $tcp.shutdown i32.const 64 call $ok
 global.get $child i32.const 2 i32.const 64 call $tcp.shutdown i32.const 64 call $ok
 global.get $ci call $streams.drop-input-stream global.get $co call $streams.drop-output-stream
 global.get $si call $streams.drop-input-stream global.get $so call $streams.drop-output-stream
 global.get $client call $tcp.drop-tcp-socket global.get $child call $tcp.drop-tcp-socket global.get $server call $tcp.drop-tcp-socket
 global.get $network call $network.drop-network i32.const 0 global.set $opened)
 (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)

  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\8d\22\01A\02\01A\22\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\00\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\01\02\03\00\00\05error\02\03\00\01\08pollable\01B(\02\03\02\01\02\04\00\05error\03\00\00\02\03\02\01\03\04\00\08pollable\03\00\02\01i\01\01q\02\15last-operation-failed\01\04\00\06closed\00\00\04\00\0cstream-error\03\00\05\04\00\0cinput-stream\03\01\04\00\0doutput-stream\03\01\01h\07\01p}\01j\01\0a\01\06\01@\02\04self\09\03lenw\00\0b\04\00\19[method]input-stream.read\01\0c\04\00\22[method]input-stream.blocking-read\01\0c\01j\01w\01\06\01@\02\04self\09\03lenw\00\0d\04\00\19[method]input-stream.skip\01\0e\04\00\22[method]input-stream.blocking-skip\01\0e\01i\03\01@\01\04self\09\00\0f\04\00\1e[method]input-stream.subscribe\01\10\01h\08\01@\01\04self\11\00\0d\04\00![method]output-stream.check-write\01\12\01j\00\01\06\01@\02\04self\11\08contents\0a\00\13\04\00\1b[method]output-stream.write\01\14\04\00.[method]output-stream.blocking-write-and-flush\01\14\01@\01\04self\11\00\13\04\00\1b[method]output-stream.flush\01\15\04\00$[method]output-stream.blocking-flush\01\15\01@\01\04self\11\00\0f\04\00\1f[method]output-stream.subscribe\01\16\01@\02\04self\11\03lenw\00\13\04\00\22[method]output-stream.write-zeroes\01\17\04\005[method]output-stream.blocking-write-zeroes-and-flush\01\17\01@\03\04self\11\03src\09\03lenw\00\0d\04\00\1c[method]output-stream.splice\01\18\04\00%[method]output-stream.blocking-splice\01\18\03\00\16wasi:io/streams@0.2.12\05\04\01B\13\02\03\02\01\02\04\00\05error\03\00\00\04\00\07network\03\01\01m\15\07unknown\0daccess-denied\0dnot-supported\10invalid-argument\0dout-of-memory\07timeout\14concurrency-conflict\0fnot-in-progress\0bwould-block\0dinvalid-state\10new-socket-limit\14address-not-bindable\0eaddress-in-use\12remote-unreachable\12connection-refused\10connection-reset\12connection-aborted\12datagram-too-large\11name-unresolvable\1atemporary-resolver-failure\1apermanent-resolver-failure\04\00\0aerror-code\03\00\03\01m\02\04ipv4\04ipv6\04\00\11ip-address-family\03\00\05\01o\04}}}}\04\00\0cipv4-address\03\00\07\01o\08{{{{{{{{\04\00\0cipv6-address\03\00\09\01q\02\04ipv4\01\08\00\04ipv6\01\0a\00\04\00\0aip-address\03\00\0b\01r\02\04port{\07address\08\04\00\13ipv4-socket-address\03\00\0d\01r\04\04port{\09flow-infoy\07address\0a\08scope-idy\04\00\13ipv6-socket-address\03\00\0f\01q\02\04ipv4\01\0e\00\04ipv6\01\10\00\04\00\11ip-socket-address\03\00\11\03\00\1bwasi:sockets/network@0.2.12\05\05\02\03\00\03\07network\01B\05\02\03\02\01\06\04\00\07network\03\00\00\01i\01\01@\00\00\02\04\00\10instance-network\01\03\03\00$wasi:sockets/instance-network@0.2.12\05\07\01B\0f\02\03\02\01\03\04\00\08pollable\03\00\00\01w\04\00\07instant\03\00\02\01w\04\00\08duration\03\00\04\01@\00\00\03\04\00\03now\01\06\01@\00\00\05\04\00\0aresolution\01\07\01i\01\01@\01\04when\03\00\08\04\00\11subscribe-instant\01\09\01@\01\04when\05\00\08\04\00\12subscribe-duration\01\0a\03\00\22wasi:clocks/monotonic-clock@0.2.12\05\08\02\03\00\02\0cinput-stream\02\03\00\02\0doutput-stream\02\03\00\05\08duration\02\03\00\03\0aerror-code\02\03\00\03\11ip-socket-address\02\03\00\03\11ip-address-family\01BT\02\03\02\01\09\04\00\0cinput-stream\03\00\00\02\03\02\01\0a\04\00\0doutput-stream\03\00\02\02\03\02\01\03\04\00\08pollable\03\00\04\02\03\02\01\0b\04\00\08duration\03\00\06\02\03\02\01\06\04\00\07network\03\00\08\02\03\02\01\0c\04\00\0aerror-code\03\00\0a\02\03\02\01\0d\04\00\11ip-socket-address\03\00\0c\02\03\02\01\0e\04\00\11ip-address-family\03\00\0e\01m\03\07receive\04send\04both\04\00\0dshutdown-type\03\00\10\04\00\0atcp-socket\03\01\01h\12\01h\09\01j\00\01\0b\01@\03\04self\13\07network\14\0dlocal-address\0d\00\15\04\00\1d[method]tcp-socket.start-bind\01\16\01@\01\04self\13\00\15\04\00\1e[method]tcp-socket.finish-bind\01\17\01@\03\04self\13\07network\14\0eremote-address\0d\00\15\04\00 [method]tcp-socket.start-connect\01\18\01i\01\01i\03\01o\02\19\1a\01j\01\1b\01\0b\01@\01\04self\13\00\1c\04\00![method]tcp-socket.finish-connect\01\1d\04\00\1f[method]tcp-socket.start-listen\01\17\04\00 [method]tcp-socket.finish-listen\01\17\01i\12\01o\03\1e\19\1a\01j\01\1f\01\0b\01@\01\04self\13\00 \04\00\19[method]tcp-socket.accept\01!\01j\01\0d\01\0b\01@\01\04self\13\00\22\04\00 [method]tcp-socket.local-address\01#\04\00![method]tcp-socket.remote-address\01#\01@\01\04self\13\00\7f\04\00\1f[method]tcp-socket.is-listening\01$\01@\01\04self\13\00\0f\04\00![method]tcp-socket.address-family\01%\01@\02\04self\13\05valuew\00\15\04\00*[method]tcp-socket.set-listen-backlog-size\01&\01j\01\7f\01\0b\01@\01\04self\13\00'\04\00%[method]tcp-socket.keep-alive-enabled\01(\01@\02\04self\13\05value\7f\00\15\04\00)[method]tcp-socket.set-keep-alive-enabled\01)\01j\01\07\01\0b\01@\01\04self\13\00*\04\00'[method]tcp-socket.keep-alive-idle-time\01+\01@\02\04self\13\05value\07\00\15\04\00+[method]tcp-socket.set-keep-alive-idle-time\01,\04\00&[method]tcp-socket.keep-alive-interval\01+\04\00*[method]tcp-socket.set-keep-alive-interval\01,\01j\01y\01\0b\01@\01\04self\13\00-\04\00#[method]tcp-socket.keep-alive-count\01.\01@\02\04self\13\05valuey\00\15\04\00'[method]tcp-socket.set-keep-alive-count\01/\01j\01}\01\0b\01@\01\04self\13\000\04\00\1c[method]tcp-socket.hop-limit\011\01@\02\04self\13\05value}\00\15\04\00 [method]tcp-socket.set-hop-limit\012\01j\01w\01\0b\01@\01\04self\13\003\04\00&[method]tcp-socket.receive-buffer-size\014\04\00*[method]tcp-socket.set-receive-buffer-size\01&\04\00#[method]tcp-socket.send-buffer-size\014\04\00'[method]tcp-socket.set-send-buffer-size\01&\01i\05\01@\01\04self\13\005\04\00\1c[method]tcp-socket.subscribe\016\01@\02\04self\13\0dshutdown-type\11\00\15\04\00\1b[method]tcp-socket.shutdown\017\03\00\17wasi:sockets/tcp@0.2.12\05\0f\02\03\00\06\0atcp-socket\01B\0c\02\03\02\01\06\04\00\07network\03\00\00\02\03\02\01\0c\04\00\0aerror-code\03\00\02\02\03\02\01\0e\04\00\11ip-address-family\03\00\04\02\03\02\01\10\04\00\0atcp-socket\03\00\06\01i\07\01j\01\08\01\03\01@\01\0eaddress-family\05\00\09\04\00\11create-tcp-socket\01\0a\03\00%wasi:sockets/tcp-create-socket@0.2.12\05\11\01@\01\06familyy\00y\04\00\0ccreate-probe\01\12\04\00\05setup\01\12\01@\00\00y\04\00\08teardown\01\13\04\00\08exercise\01\12\01@\01\05county\00y\04\00\08transfer\01\14\04\00\14wasmvm:tcp-tests/tcp\04\00\0b\09\01\00\03tcp\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

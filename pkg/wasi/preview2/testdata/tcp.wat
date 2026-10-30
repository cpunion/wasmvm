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
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type5 error" (@name "error") (;5;)))
  (type $ty-wasi:sockets/network@0.2.12 (;6;)
    (instance
      (alias outer 1 $"#type5 error" (type (;0;)))
      (export (;1;) "error" (type (eq 0)))
      (type (;2;) (enum "unknown" "access-denied" "not-supported" "invalid-argument" "out-of-memory" "timeout" "concurrency-conflict" "not-in-progress" "would-block" "invalid-state" "new-socket-limit" "address-not-bindable" "address-in-use" "remote-unreachable" "connection-refused" "connection-reset" "connection-aborted" "datagram-too-large" "name-unresolvable" "temporary-resolver-failure" "permanent-resolver-failure"))
      (export (;3;) "error-code" (type (eq 2)))
      (export (;4;) "network" (type (sub resource)))
      (type (;5;) (tuple u8 u8 u8 u8))
      (export (;6;) "ipv4-address" (type (eq 5)))
      (type (;7;) (record (field "port" u16) (field "address" 6)))
      (export (;8;) "ipv4-socket-address" (type (eq 7)))
      (type (;9;) (tuple u16 u16 u16 u16 u16 u16 u16 u16))
      (export (;10;) "ipv6-address" (type (eq 9)))
      (type (;11;) (record (field "port" u16) (field "flow-info" u32) (field "address" 10) (field "scope-id" u32)))
      (export (;12;) "ipv6-socket-address" (type (eq 11)))
      (type (;13;) (variant (case "ipv4" 8) (case "ipv6" 12)))
      (export (;14;) "ip-socket-address" (type (eq 13)))
      (type (;15;) (enum "ipv4" "ipv6"))
      (export (;16;) "ip-address-family" (type (eq 15)))
      (type (;17;) (borrow 1))
      (type (;18;) (option 3))
      (type (;19;) (func (param "err" 17) (result 18)))
      (export (;0;) "network-error-code" (func (type 19)))
    )
  )
  (import "wasi:sockets/network@0.2.12" (instance $wasi:sockets/network@0.2.12 (;3;) (type $ty-wasi:sockets/network@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $network (;7;)))
  (type $ty-wasi:sockets/instance-network@0.2.12 (;8;)
    (instance
      (alias outer 1 $network (type (;0;)))
      (export (;1;) "network" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "instance-network" (func (type 3)))
    )
  )
  (import "wasi:sockets/instance-network@0.2.12" (instance $wasi:sockets/instance-network@0.2.12 (;4;) (type $ty-wasi:sockets/instance-network@0.2.12)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type9 pollable" (@name "pollable") (;9;)))
  (type $ty-wasi:clocks/monotonic-clock@0.2.12 (;10;)
    (instance
      (type (;0;) u64)
      (export (;1;) "instant" (type (eq 0)))
      (type (;2;) u64)
      (export (;3;) "duration" (type (eq 2)))
      (alias outer 1 $"#type9 pollable" (type (;4;)))
      (export (;5;) "pollable" (type (eq 4)))
      (type (;6;) (func (result 1)))
      (export (;0;) "now" (func (type 6)))
      (type (;7;) (func (result 3)))
      (export (;1;) "resolution" (func (type 7)))
      (type (;8;) (own 5))
      (type (;9;) (func (param "when" 1) (result 8)))
      (export (;2;) "subscribe-instant" (func (type 9)))
      (type (;10;) (func (param "when" 3) (result 8)))
      (export (;3;) "subscribe-duration" (func (type 10)))
    )
  )
  (import "wasi:clocks/monotonic-clock@0.2.12" (instance $wasi:clocks/monotonic-clock@0.2.12 (;5;) (type $ty-wasi:clocks/monotonic-clock@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type11 network" (@name "network") (;11;)))
  (alias export $wasi:sockets/network@0.2.12 "ip-socket-address" (type $ip-socket-address (;12;)))
  (alias export $wasi:sockets/network@0.2.12 "error-code" (type $error-code (;13;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $input-stream (;14;)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $output-stream (;15;)))
  (alias export $wasi:sockets/network@0.2.12 "ip-address-family" (type $ip-address-family (;16;)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "duration" (type $duration (;17;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type18 pollable" (@name "pollable") (;18;)))
  (type $ty-wasi:sockets/tcp@0.2.12 (;19;)
    (instance
      (export (;0;) "tcp-socket" (type (sub resource)))
      (alias outer 1 $"#type11 network" (type (;1;)))
      (export (;2;) "network" (type (eq 1)))
      (alias outer 1 $ip-socket-address (type (;3;)))
      (export (;4;) "ip-socket-address" (type (eq 3)))
      (alias outer 1 $error-code (type (;5;)))
      (export (;6;) "error-code" (type (eq 5)))
      (alias outer 1 $input-stream (type (;7;)))
      (export (;8;) "input-stream" (type (eq 7)))
      (alias outer 1 $output-stream (type (;9;)))
      (export (;10;) "output-stream" (type (eq 9)))
      (alias outer 1 $ip-address-family (type (;11;)))
      (export (;12;) "ip-address-family" (type (eq 11)))
      (alias outer 1 $duration (type (;13;)))
      (export (;14;) "duration" (type (eq 13)))
      (alias outer 1 $"#type18 pollable" (type (;15;)))
      (export (;16;) "pollable" (type (eq 15)))
      (type (;17;) (enum "receive" "send" "both"))
      (export (;18;) "shutdown-type" (type (eq 17)))
      (type (;19;) (borrow 0))
      (type (;20;) (borrow 2))
      (type (;21;) (result (error 6)))
      (type (;22;) (func (param "self" 19) (param "network" 20) (param "local-address" 4) (result 21)))
      (export (;0;) "[method]tcp-socket.start-bind" (func (type 22)))
      (type (;23;) (func (param "self" 19) (result 21)))
      (export (;1;) "[method]tcp-socket.finish-bind" (func (type 23)))
      (type (;24;) (func (param "self" 19) (param "network" 20) (param "remote-address" 4) (result 21)))
      (export (;2;) "[method]tcp-socket.start-connect" (func (type 24)))
      (type (;25;) (own 8))
      (type (;26;) (own 10))
      (type (;27;) (tuple 25 26))
      (type (;28;) (result 27 (error 6)))
      (type (;29;) (func (param "self" 19) (result 28)))
      (export (;3;) "[method]tcp-socket.finish-connect" (func (type 29)))
      (export (;4;) "[method]tcp-socket.start-listen" (func (type 23)))
      (export (;5;) "[method]tcp-socket.finish-listen" (func (type 23)))
      (type (;30;) (own 0))
      (type (;31;) (tuple 30 25 26))
      (type (;32;) (result 31 (error 6)))
      (type (;33;) (func (param "self" 19) (result 32)))
      (export (;6;) "[method]tcp-socket.accept" (func (type 33)))
      (type (;34;) (result 4 (error 6)))
      (type (;35;) (func (param "self" 19) (result 34)))
      (export (;7;) "[method]tcp-socket.local-address" (func (type 35)))
      (export (;8;) "[method]tcp-socket.remote-address" (func (type 35)))
      (type (;36;) (func (param "self" 19) (result bool)))
      (export (;9;) "[method]tcp-socket.is-listening" (func (type 36)))
      (type (;37;) (func (param "self" 19) (result 12)))
      (export (;10;) "[method]tcp-socket.address-family" (func (type 37)))
      (type (;38;) (func (param "self" 19) (param "value" u64) (result 21)))
      (export (;11;) "[method]tcp-socket.set-listen-backlog-size" (func (type 38)))
      (type (;39;) (result bool (error 6)))
      (type (;40;) (func (param "self" 19) (result 39)))
      (export (;12;) "[method]tcp-socket.keep-alive-enabled" (func (type 40)))
      (type (;41;) (func (param "self" 19) (param "value" bool) (result 21)))
      (export (;13;) "[method]tcp-socket.set-keep-alive-enabled" (func (type 41)))
      (type (;42;) (result 14 (error 6)))
      (type (;43;) (func (param "self" 19) (result 42)))
      (export (;14;) "[method]tcp-socket.keep-alive-idle-time" (func (type 43)))
      (type (;44;) (func (param "self" 19) (param "value" 14) (result 21)))
      (export (;15;) "[method]tcp-socket.set-keep-alive-idle-time" (func (type 44)))
      (export (;16;) "[method]tcp-socket.keep-alive-interval" (func (type 43)))
      (export (;17;) "[method]tcp-socket.set-keep-alive-interval" (func (type 44)))
      (type (;45;) (result u32 (error 6)))
      (type (;46;) (func (param "self" 19) (result 45)))
      (export (;18;) "[method]tcp-socket.keep-alive-count" (func (type 46)))
      (type (;47;) (func (param "self" 19) (param "value" u32) (result 21)))
      (export (;19;) "[method]tcp-socket.set-keep-alive-count" (func (type 47)))
      (type (;48;) (result u8 (error 6)))
      (type (;49;) (func (param "self" 19) (result 48)))
      (export (;20;) "[method]tcp-socket.hop-limit" (func (type 49)))
      (type (;50;) (func (param "self" 19) (param "value" u8) (result 21)))
      (export (;21;) "[method]tcp-socket.set-hop-limit" (func (type 50)))
      (type (;51;) (result u64 (error 6)))
      (type (;52;) (func (param "self" 19) (result 51)))
      (export (;22;) "[method]tcp-socket.receive-buffer-size" (func (type 52)))
      (export (;23;) "[method]tcp-socket.set-receive-buffer-size" (func (type 38)))
      (export (;24;) "[method]tcp-socket.send-buffer-size" (func (type 52)))
      (export (;25;) "[method]tcp-socket.set-send-buffer-size" (func (type 38)))
      (type (;53;) (own 16))
      (type (;54;) (func (param "self" 19) (result 53)))
      (export (;26;) "[method]tcp-socket.subscribe" (func (type 54)))
      (type (;55;) (func (param "self" 19) (param "shutdown-type" 18) (result 21)))
      (export (;27;) "[method]tcp-socket.shutdown" (func (type 55)))
    )
  )
  (import "wasi:sockets/tcp@0.2.12" (instance $wasi:sockets/tcp@0.2.12 (;6;) (type $ty-wasi:sockets/tcp@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "ip-address-family" (type $"#type20 ip-address-family" (@name "ip-address-family") (;20;)))
  (alias export $wasi:sockets/tcp@0.2.12 "tcp-socket" (type $tcp-socket (;21;)))
  (alias export $wasi:sockets/network@0.2.12 "error-code" (type $"#type22 error-code" (@name "error-code") (;22;)))
  (type $ty-wasi:sockets/tcp-create-socket@0.2.12 (;23;)
    (instance
      (alias outer 1 $"#type20 ip-address-family" (type (;0;)))
      (export (;1;) "ip-address-family" (type (eq 0)))
      (alias outer 1 $tcp-socket (type (;2;)))
      (export (;3;) "tcp-socket" (type (eq 2)))
      (alias outer 1 $"#type22 error-code" (type (;4;)))
      (export (;5;) "error-code" (type (eq 4)))
      (type (;6;) (own 3))
      (type (;7;) (result 6 (error 5)))
      (type (;8;) (func (param "address-family" 1) (result 7)))
      (export (;0;) "create-tcp-socket" (func (type 8)))
    )
  )
  (import "wasi:sockets/tcp-create-socket@0.2.12" (instance $wasi:sockets/tcp-create-socket@0.2.12 (;7;) (type $ty-wasi:sockets/tcp-create-socket@0.2.12)))
  (core module $main (;0;)
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
    (import "wasi:sockets/network@0.2.12" "network-error-code" (func $network.network-error-code (;23;) (type 0)))
    (import "wasi:sockets/network@0.2.12" "[resource-drop]network" (func $network.drop-network (;24;) (type 1)))
    (import "wasi:sockets/instance-network@0.2.12" "instance-network" (func $instance-network.instance-network (;25;) (type 7)))
    (import "wasi:clocks/monotonic-clock@0.2.12" "now" (func $monotonic-clock.now (;26;) (type 8)))
    (import "wasi:clocks/monotonic-clock@0.2.12" "resolution" (func $monotonic-clock.resolution (;27;) (type 8)))
    (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-instant" (func $monotonic-clock.subscribe-instant (;28;) (type 9)))
    (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-duration" (func $monotonic-clock.subscribe-duration (;29;) (type 9)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-bind" (func $tcp.start-bind (;30;) (type 10)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-bind" (func $tcp.finish-bind (;31;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-connect" (func $tcp.start-connect (;32;) (type 10)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-connect" (func $tcp.finish-connect (;33;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.start-listen" (func $tcp.start-listen (;34;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.finish-listen" (func $tcp.finish-listen (;35;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.accept" (func $tcp.accept (;36;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.local-address" (func $tcp.local-address (;37;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.remote-address" (func $tcp.remote-address (;38;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.is-listening" (func $tcp.is-listening (;39;) (type 2)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.address-family" (func $tcp.address-family (;40;) (type 2)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-listen-backlog-size" (func $tcp.set-listen-backlog-size (;41;) (type 4)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-enabled" (func $tcp.keep-alive-enabled (;42;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-enabled" (func $tcp.set-keep-alive-enabled (;43;) (type 3)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-idle-time" (func $tcp.keep-alive-idle-time (;44;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-idle-time" (func $tcp.set-keep-alive-idle-time (;45;) (type 4)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-interval" (func $tcp.keep-alive-interval (;46;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-interval" (func $tcp.set-keep-alive-interval (;47;) (type 4)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.keep-alive-count" (func $tcp.keep-alive-count (;48;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-keep-alive-count" (func $tcp.set-keep-alive-count (;49;) (type 3)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.hop-limit" (func $tcp.hop-limit (;50;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-hop-limit" (func $tcp.set-hop-limit (;51;) (type 3)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.receive-buffer-size" (func $tcp.receive-buffer-size (;52;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-receive-buffer-size" (func $tcp.set-receive-buffer-size (;53;) (type 4)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.send-buffer-size" (func $tcp.send-buffer-size (;54;) (type 0)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.set-send-buffer-size" (func $tcp.set-send-buffer-size (;55;) (type 4)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.subscribe" (func $tcp.subscribe (;56;) (type 2)))
    (import "wasi:sockets/tcp@0.2.12" "[method]tcp-socket.shutdown" (func $tcp.shutdown (;57;) (type 3)))
    (import "wasi:sockets/tcp@0.2.12" "[resource-drop]tcp-socket" (func $tcp.drop-tcp-socket (;58;) (type 1)))
    (import "wasi:sockets/tcp-create-socket@0.2.12" "create-tcp-socket" (func $tcp-create-socket.create-tcp-socket (;59;) (type 0)))
    (memory (;0;) 32 64)
    (global $heap (;0;) (mut i32) i32.const 8192)
    (global $network (;1;) (mut i32) i32.const 0)
    (global $server (;2;) (mut i32) i32.const 0)
    (global $client (;3;) (mut i32) i32.const 0)
    (global $child (;4;) (mut i32) i32.const 0)
    (global $ci (;5;) (mut i32) i32.const 0)
    (global $co (;6;) (mut i32) i32.const 0)
    (global $si (;7;) (mut i32) i32.const 0)
    (global $so (;8;) (mut i32) i32.const 0)
    (global $opened (;9;) (mut i32) i32.const 0)
    (export "create-probe" (func 60))
    (export "cabi_post_create-probe" (func 61))
    (export "setup" (func 62))
    (export "cabi_post_setup" (func 63))
    (export "teardown" (func 64))
    (export "cabi_post_teardown" (func 65))
    (export "exercise" (func 66))
    (export "cabi_post_exercise" (func 67))
    (export "transfer" (func 68))
    (export "cabi_post_transfer" (func 69))
    (export "memory" (memory 0))
    (export "cabi_realloc" (func 70))
    (export "_initialize" (func 71))
    (func (;60;) (type 2) (param i32) (result i32)
      local.get 0
      i32.const 64
      call $tcp-create-socket.create-tcp-socket
      i32.const 64
      i32.load8_u
      if (result i32) ;; label = @1
        i32.const 68
        i32.load8_u
        i32.const 1
        i32.add
      else
        i32.const 68
        i32.load
        call $tcp.drop-tcp-socket
        i32.const 0
      end
    )
    (func (;61;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;62;) (type 2) (param i32) (result i32)
      local.get 0
      call $open
      i32.const 1
    )
    (func (;63;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;64;) (type 7) (result i32)
      call $close
      i32.const 1
    )
    (func (;65;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;66;) (type 2) (param i32) (result i32)
      local.get 0
      call $open
      i32.const 3
      call $exchange
      i32.const 3
      i32.ne
      if ;; label = @1
        unreachable
      end
      call $close
      i32.const 42
    )
    (func (;67;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;68;) (type 2) (param i32) (result i32)
      local.get 0
      call $exchange
    )
    (func (;69;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;70;) (type 11) (param i32 i32 i32 i32) (result i32)
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
    (func (;71;) (type 12))
    (func $open (;72;) (type 1) (param $family i32)
      (local $poll i32) (local $port i32)
      global.get $opened
      if ;; label = @1
        unreachable
      end
      call $instance-network.instance-network
      global.set $network
      local.get $family
      i32.const 64
      call $tcp-create-socket.create-tcp-socket
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $server
      local.get $family
      i32.const 64
      call $tcp-create-socket.create-tcp-socket
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $client
      global.get $client
      call $tcp.address-family
      local.get $family
      i32.ne
      if ;; label = @1
        unreachable
      end
      global.get $client
      call $tcp.is-listening
      if ;; label = @1
        unreachable
      end
      global.get $client
      i32.const 1
      i32.const 64
      call $tcp.set-keep-alive-enabled
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.keep-alive-enabled
      i32.const 64
      call $ok
      global.get $client
      i64.const 1000000000
      i32.const 64
      call $tcp.set-keep-alive-idle-time
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.keep-alive-idle-time
      i32.const 64
      call $ok
      global.get $client
      i64.const 1000000000
      i32.const 64
      call $tcp.set-keep-alive-interval
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.keep-alive-interval
      i32.const 64
      call $ok
      global.get $client
      i32.const 3
      i32.const 64
      call $tcp.set-keep-alive-count
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.keep-alive-count
      i32.const 64
      call $ok
      global.get $client
      i32.const 32
      i32.const 64
      call $tcp.set-hop-limit
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.hop-limit
      i32.const 64
      call $ok
      global.get $client
      i64.const 65536
      i32.const 64
      call $tcp.set-receive-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.receive-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i64.const 65536
      i32.const 64
      call $tcp.set-send-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $tcp.send-buffer-size
      i32.const 64
      call $ok
      global.get $server
      i64.const 128
      i32.const 64
      call $tcp.set-listen-backlog-size
      i32.const 64
      call $ok
      global.get $server
      global.get $network
      local.get $family
      i32.const 0
      local.get $family
      if (result i32) ;; label = @1
        i32.const 0
      else
        i32.const 127
      end
      i32.const 0
      i32.const 0
      local.get $family
      if (result i32) ;; label = @1
        i32.const 0
      else
        i32.const 1
      end
      i32.const 0
      i32.const 0
      i32.const 0
      i32.const 0
      local.get $family
      if (result i32) ;; label = @1
        i32.const 1
      else
        i32.const 0
      end
      i32.const 0
      i32.const 64
      call $tcp.start-bind
      i32.const 64
      call $ok
      global.get $server
      i32.const 64
      call $tcp.finish-bind
      i32.const 64
      call $ok
      global.get $server
      i32.const 64
      call $tcp.start-listen
      i32.const 64
      call $ok
      global.get $server
      i32.const 64
      call $tcp.finish-listen
      i32.const 64
      call $ok
      global.get $server
      call $tcp.is-listening
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $server
      i32.const 64
      call $tcp.local-address
      i32.const 64
      call $ok
      i32.const 72
      i32.load16_u
      local.tee $port
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $client
      global.get $network
      local.get $family
      local.get $port
      local.get $family
      if (result i32) ;; label = @1
        i32.const 0
      else
        i32.const 127
      end
      i32.const 0
      i32.const 0
      local.get $family
      if (result i32) ;; label = @1
        i32.const 0
      else
        i32.const 1
      end
      i32.const 0
      i32.const 0
      i32.const 0
      i32.const 0
      local.get $family
      if (result i32) ;; label = @1
        i32.const 1
      else
        i32.const 0
      end
      i32.const 0
      i32.const 64
      call $tcp.start-connect
      i32.const 64
      call $ok
      block $connected
        loop $connect
          global.get $client
          i32.const 64
          call $tcp.finish-connect
          i32.const 64
          i32.load8_u
          i32.eqz
          br_if $connected
          i32.const 68
          i32.load8_u
          i32.const 8
          i32.ne
          if ;; label = @3
            unreachable
          end
          global.get $client
          call $tcp.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $connect
        end
      end
      i32.const 68
      i32.load
      global.set $ci
      i32.const 72
      i32.load
      global.set $co
      block $accepted
        loop $accept
          global.get $server
          i32.const 64
          call $tcp.accept
          i32.const 64
          i32.load8_u
          i32.eqz
          br_if $accepted
          i32.const 68
          i32.load8_u
          i32.const 8
          i32.ne
          if ;; label = @3
            unreachable
          end
          global.get $server
          call $tcp.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $accept
        end
      end
      i32.const 68
      i32.load
      global.set $child
      i32.const 72
      i32.load
      global.set $si
      i32.const 76
      i32.load
      global.set $so
      global.get $client
      i32.const 64
      call $tcp.remote-address
      i32.const 64
      call $ok
      global.get $child
      i32.const 64
      call $tcp.local-address
      i32.const 64
      call $ok
      i32.const 1
      global.set $opened
    )
    (func $read-all (;73;) (type 1) (param $stream i32)
      (local $at i32) (local $n i32) (local $ptr i32)
      block $end
        loop $next
          local.get $at
          i32.const 4096
          i32.ge_u
          br_if $end
          local.get $stream
          i32.const 4096
          local.get $at
          i32.sub
          i64.extend_i32_u
          i32.const 64
          call $streams.input-stream.blocking-read
          i32.const 64
          call $ok
          i32.const 68
          i32.load
          local.set $ptr
          i32.const 72
          i32.load
          local.tee $n
          i32.eqz
          if ;; label = @3
            unreachable
          end
          local.get $ptr
          i32.load8_u
          i32.const 1024
          local.get $at
          i32.const 15
          i32.and
          i32.add
          i32.load8_u
          i32.ne
          if ;; label = @3
            unreachable
          end
          local.get $ptr
          local.get $n
          i32.add
          i32.const 1
          i32.sub
          i32.load8_u
          i32.const 1024
          local.get $at
          local.get $n
          i32.add
          i32.const 1
          i32.sub
          i32.const 15
          i32.and
          i32.add
          i32.load8_u
          i32.ne
          if ;; label = @3
            unreachable
          end
          local.get $at
          local.get $n
          i32.add
          local.set $at
          br $next
        end
      end
    )
    (func $exchange (;74;) (type 2) (param $count i32) (result i32)
      (local $at i32)
      global.get $opened
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $count
      i32.const 128
      i32.gt_u
      if ;; label = @1
        unreachable
      end
      block $end
        loop $next
          local.get $at
          local.get $count
          i32.ge_u
          br_if $end
          global.get $co
          i32.const 1024
          i32.const 4096
          i32.const 64
          call $streams.output-stream.blocking-write-and-flush
          i32.const 64
          call $ok
          global.get $si
          call $read-all
          global.get $so
          i32.const 1024
          i32.const 4096
          i32.const 64
          call $streams.output-stream.blocking-write-and-flush
          i32.const 64
          call $ok
          global.get $ci
          call $read-all
          local.get $at
          i32.const 1
          i32.add
          local.set $at
          br $next
        end
      end
      local.get $at
    )
    (func $close (;75;) (type 12)
      global.get $opened
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $client
      i32.const 2
      i32.const 64
      call $tcp.shutdown
      i32.const 64
      call $ok
      global.get $child
      i32.const 2
      i32.const 64
      call $tcp.shutdown
      i32.const 64
      call $ok
      global.get $ci
      call $streams.drop-input-stream
      global.get $co
      call $streams.drop-output-stream
      global.get $si
      call $streams.drop-input-stream
      global.get $so
      call $streams.drop-output-stream
      global.get $client
      call $tcp.drop-tcp-socket
      global.get $child
      call $tcp.drop-tcp-socket
      global.get $server
      call $tcp.drop-tcp-socket
      global.get $network
      call $network.drop-network
      i32.const 0
      global.set $opened
    )
    (func $ok (;76;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (data (;0;) (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
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
    (type (;5;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
    (table (;0;) 42 42 funcref)
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
    (export "15" (func $indirect-wasi:sockets/network@0.2.12-network-error-code))
    (export "16" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind"))
    (export "17" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind"))
    (export "18" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect"))
    (export "19" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect"))
    (export "20" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen"))
    (export "21" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen"))
    (export "22" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept"))
    (export "23" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address"))
    (export "24" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address"))
    (export "25" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size"))
    (export "26" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled"))
    (export "27" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled"))
    (export "28" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time"))
    (export "29" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time"))
    (export "30" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval"))
    (export "31" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval"))
    (export "32" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count"))
    (export "33" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count"))
    (export "34" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit"))
    (export "35" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit"))
    (export "36" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size"))
    (export "37" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size"))
    (export "38" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size"))
    (export "39" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size"))
    (export "40" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown"))
    (export "41" (func $indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket))
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
    (func $indirect-wasi:sockets/network@0.2.12-network-error-code (;15;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 15
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind" (;16;) (type 5) (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      local.get 8
      local.get 9
      local.get 10
      local.get 11
      local.get 12
      local.get 13
      local.get 14
      i32.const 16
      call_indirect (type 5)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind" (;17;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 17
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect" (;18;) (type 5) (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      local.get 8
      local.get 9
      local.get 10
      local.get 11
      local.get 12
      local.get 13
      local.get 14
      i32.const 18
      call_indirect (type 5)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect" (;19;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 19
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen" (;20;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 20
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen" (;21;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 21
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept" (;22;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 22
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address" (;23;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 23
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address" (;24;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 24
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size" (;25;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 25
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled" (;26;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 26
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled" (;27;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 27
      call_indirect (type 1)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time" (;28;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 28
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time" (;29;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 29
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval" (;30;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 30
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval" (;31;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 31
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count" (;32;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 32
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count" (;33;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 33
      call_indirect (type 1)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit" (;34;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 34
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit" (;35;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 35
      call_indirect (type 1)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size" (;36;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 36
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size" (;37;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 37
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size" (;38;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 38
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size" (;39;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 39
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown" (;40;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 40
      call_indirect (type 1)
    )
    (func $indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket (;41;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 41
      call_indirect (type 0)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias core export $wit-component-shim-instance "0" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type24 error" (@name "error") (;24;)))
  (core func $resource.drop (;1;) (canon resource.drop $"#type24 error"))
  (core instance $wasi:io/error@0.2.12 (;1;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $resource.drop))
  )
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;2;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;3;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "1" (core func $indirect-wasi:io/poll@0.2.12-poll (;4;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type25 pollable" (@name "pollable") (;25;)))
  (core func $"#core-func5 resource.drop" (@name "resource.drop") (;5;) (canon resource.drop $"#type25 pollable"))
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
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type26 input-stream" (@name "input-stream") (;26;)))
  (core func $"#core-func21 resource.drop" (@name "resource.drop") (;21;) (canon resource.drop $"#type26 input-stream"))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type27 output-stream" (@name "output-stream") (;27;)))
  (core func $"#core-func22 resource.drop" (@name "resource.drop") (;22;) (canon resource.drop $"#type27 output-stream"))
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
  (alias core export $wit-component-shim-instance "15" (core func $indirect-wasi:sockets/network@0.2.12-network-error-code (;23;)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type28 network" (@name "network") (;28;)))
  (core func $"#core-func24 resource.drop" (@name "resource.drop") (;24;) (canon resource.drop $"#type28 network"))
  (core instance $wasi:sockets/network@0.2.12 (;4;)
    (export "network-error-code" (func $indirect-wasi:sockets/network@0.2.12-network-error-code))
    (export "[resource-drop]network" (func $"#core-func24 resource.drop"))
  )
  (alias export $wasi:sockets/instance-network@0.2.12 "instance-network" (func $instance-network (;4;)))
  (core func $instance-network (;25;) (canon lower (func $instance-network)))
  (core instance $wasi:sockets/instance-network@0.2.12 (;5;)
    (export "instance-network" (func $instance-network))
  )
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "now" (func $now (;5;)))
  (core func $now (;26;) (canon lower (func $now)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "resolution" (func $resolution (;6;)))
  (core func $resolution (;27;) (canon lower (func $resolution)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "subscribe-instant" (func $subscribe-instant (;7;)))
  (core func $subscribe-instant (;28;) (canon lower (func $subscribe-instant)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "subscribe-duration" (func $subscribe-duration (;8;)))
  (core func $subscribe-duration (;29;) (canon lower (func $subscribe-duration)))
  (core instance $wasi:clocks/monotonic-clock@0.2.12 (;6;)
    (export "now" (func $now))
    (export "resolution" (func $resolution))
    (export "subscribe-instant" (func $subscribe-instant))
    (export "subscribe-duration" (func $subscribe-duration))
  )
  (alias core export $wit-component-shim-instance "16" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind" (;30;)))
  (alias core export $wit-component-shim-instance "17" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind" (;31;)))
  (alias core export $wit-component-shim-instance "18" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect" (;32;)))
  (alias core export $wit-component-shim-instance "19" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect" (;33;)))
  (alias core export $wit-component-shim-instance "20" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen" (;34;)))
  (alias core export $wit-component-shim-instance "21" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen" (;35;)))
  (alias core export $wit-component-shim-instance "22" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept" (;36;)))
  (alias core export $wit-component-shim-instance "23" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address" (;37;)))
  (alias core export $wit-component-shim-instance "24" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address" (;38;)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.is-listening" (func $"[method]tcp-socket.is-listening" (;9;)))
  (core func $"[method]tcp-socket.is-listening" (;39;) (canon lower (func $"[method]tcp-socket.is-listening")))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.address-family" (func $"[method]tcp-socket.address-family" (;10;)))
  (core func $"[method]tcp-socket.address-family" (;40;) (canon lower (func $"[method]tcp-socket.address-family")))
  (alias core export $wit-component-shim-instance "25" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size" (;41;)))
  (alias core export $wit-component-shim-instance "26" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled" (;42;)))
  (alias core export $wit-component-shim-instance "27" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled" (;43;)))
  (alias core export $wit-component-shim-instance "28" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time" (;44;)))
  (alias core export $wit-component-shim-instance "29" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time" (;45;)))
  (alias core export $wit-component-shim-instance "30" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval" (;46;)))
  (alias core export $wit-component-shim-instance "31" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval" (;47;)))
  (alias core export $wit-component-shim-instance "32" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count" (;48;)))
  (alias core export $wit-component-shim-instance "33" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count" (;49;)))
  (alias core export $wit-component-shim-instance "34" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit" (;50;)))
  (alias core export $wit-component-shim-instance "35" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit" (;51;)))
  (alias core export $wit-component-shim-instance "36" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size" (;52;)))
  (alias core export $wit-component-shim-instance "37" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size" (;53;)))
  (alias core export $wit-component-shim-instance "38" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size" (;54;)))
  (alias core export $wit-component-shim-instance "39" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size" (;55;)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.subscribe" (func $"[method]tcp-socket.subscribe" (;11;)))
  (core func $"[method]tcp-socket.subscribe" (;56;) (canon lower (func $"[method]tcp-socket.subscribe")))
  (alias core export $wit-component-shim-instance "40" (core func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown" (;57;)))
  (alias export $wasi:sockets/tcp@0.2.12 "tcp-socket" (type $"#type29 tcp-socket" (@name "tcp-socket") (;29;)))
  (core func $"#core-func58 resource.drop" (@name "resource.drop") (;58;) (canon resource.drop $"#type29 tcp-socket"))
  (core instance $wasi:sockets/tcp@0.2.12 (;7;)
    (export "[method]tcp-socket.start-bind" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind"))
    (export "[method]tcp-socket.finish-bind" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind"))
    (export "[method]tcp-socket.start-connect" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect"))
    (export "[method]tcp-socket.finish-connect" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect"))
    (export "[method]tcp-socket.start-listen" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen"))
    (export "[method]tcp-socket.finish-listen" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen"))
    (export "[method]tcp-socket.accept" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept"))
    (export "[method]tcp-socket.local-address" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address"))
    (export "[method]tcp-socket.remote-address" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address"))
    (export "[method]tcp-socket.is-listening" (func $"[method]tcp-socket.is-listening"))
    (export "[method]tcp-socket.address-family" (func $"[method]tcp-socket.address-family"))
    (export "[method]tcp-socket.set-listen-backlog-size" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size"))
    (export "[method]tcp-socket.keep-alive-enabled" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled"))
    (export "[method]tcp-socket.set-keep-alive-enabled" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled"))
    (export "[method]tcp-socket.keep-alive-idle-time" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time"))
    (export "[method]tcp-socket.set-keep-alive-idle-time" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time"))
    (export "[method]tcp-socket.keep-alive-interval" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval"))
    (export "[method]tcp-socket.set-keep-alive-interval" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval"))
    (export "[method]tcp-socket.keep-alive-count" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count"))
    (export "[method]tcp-socket.set-keep-alive-count" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count"))
    (export "[method]tcp-socket.hop-limit" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit"))
    (export "[method]tcp-socket.set-hop-limit" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit"))
    (export "[method]tcp-socket.receive-buffer-size" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size"))
    (export "[method]tcp-socket.set-receive-buffer-size" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size"))
    (export "[method]tcp-socket.send-buffer-size" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size"))
    (export "[method]tcp-socket.set-send-buffer-size" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size"))
    (export "[method]tcp-socket.subscribe" (func $"[method]tcp-socket.subscribe"))
    (export "[method]tcp-socket.shutdown" (func $"indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown"))
    (export "[resource-drop]tcp-socket" (func $"#core-func58 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "41" (core func $indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket (;59;)))
  (core instance $wasi:sockets/tcp-create-socket@0.2.12 (;8;)
    (export "create-tcp-socket" (func $indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket))
  )
  (core instance $main (;9;) (instantiate $main
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12))
      (with "wasi:sockets/network@0.2.12" (instance $wasi:sockets/network@0.2.12))
      (with "wasi:sockets/instance-network@0.2.12" (instance $wasi:sockets/instance-network@0.2.12))
      (with "wasi:clocks/monotonic-clock@0.2.12" (instance $wasi:clocks/monotonic-clock@0.2.12))
      (with "wasi:sockets/tcp@0.2.12" (instance $wasi:sockets/tcp@0.2.12))
      (with "wasi:sockets/tcp-create-socket@0.2.12" (instance $wasi:sockets/tcp-create-socket@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
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
    (import "actual" "15" (func $15 (;15;) (type 0)))
    (import "actual" "16" (func $16 (;16;) (type 5)))
    (import "actual" "17" (func $17 (;17;) (type 0)))
    (import "actual" "18" (func $18 (;18;) (type 5)))
    (import "actual" "19" (func $19 (;19;) (type 0)))
    (import "actual" "20" (func $20 (;20;) (type 0)))
    (import "actual" "21" (func $21 (;21;) (type 0)))
    (import "actual" "22" (func $22 (;22;) (type 0)))
    (import "actual" "23" (func $23 (;23;) (type 0)))
    (import "actual" "24" (func $24 (;24;) (type 0)))
    (import "actual" "25" (func $25 (;25;) (type 2)))
    (import "actual" "26" (func $26 (;26;) (type 0)))
    (import "actual" "27" (func $27 (;27;) (type 1)))
    (import "actual" "28" (func $28 (;28;) (type 0)))
    (import "actual" "29" (func $29 (;29;) (type 2)))
    (import "actual" "30" (func $30 (;30;) (type 0)))
    (import "actual" "31" (func $31 (;31;) (type 2)))
    (import "actual" "32" (func $32 (;32;) (type 0)))
    (import "actual" "33" (func $33 (;33;) (type 1)))
    (import "actual" "34" (func $34 (;34;) (type 0)))
    (import "actual" "35" (func $35 (;35;) (type 1)))
    (import "actual" "36" (func $36 (;36;) (type 0)))
    (import "actual" "37" (func $37 (;37;) (type 2)))
    (import "actual" "38" (func $38 (;38;) (type 0)))
    (import "actual" "39" (func $39 (;39;) (type 2)))
    (import "actual" "40" (func $40 (;40;) (type 1)))
    (import "actual" "41" (func $41 (;41;) (type 0)))
    (import "main" "_initialize" (func $_initialize (;42;) (type 6)))
    (import "shim" "$imports" (table (;0;) 42 42 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16 $17 $18 $19 $20 $21 $22 $23 $24 $25 $26 $27 $28 $29 $30 $31 $32 $33 $34 $35 $36 $37 $38 $39 $40 $41)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;12;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;60;)))
  (core func $"#core-func61 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;61;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;13;)))
  (core func $"#core-func62 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;62;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.read" (func $"[method]input-stream.read" (;14;)))
  (core func $"#core-func63 indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.read") (;63;) (canon lower (func $"[method]input-stream.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-read" (func $"[method]input-stream.blocking-read" (;15;)))
  (core func $"#core-func64 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read") (;64;) (canon lower (func $"[method]input-stream.blocking-read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.skip" (func $"[method]input-stream.skip" (;16;)))
  (core func $"#core-func65 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.skip") (;65;) (canon lower (func $"[method]input-stream.skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-skip" (func $"[method]input-stream.blocking-skip" (;17;)))
  (core func $"#core-func66 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip") (;66;) (canon lower (func $"[method]input-stream.blocking-skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.check-write" (func $"[method]output-stream.check-write" (;18;)))
  (core func $"#core-func67 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write") (;67;) (canon lower (func $"[method]output-stream.check-write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write" (func $"[method]output-stream.write" (;19;)))
  (core func $"#core-func68 indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write") (;68;) (canon lower (func $"[method]output-stream.write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-and-flush" (func $"[method]output-stream.blocking-write-and-flush" (;20;)))
  (core func $"#core-func69 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush") (;69;) (canon lower (func $"[method]output-stream.blocking-write-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.flush" (func $"[method]output-stream.flush" (;21;)))
  (core func $"#core-func70 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.flush") (;70;) (canon lower (func $"[method]output-stream.flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-flush" (func $"[method]output-stream.blocking-flush" (;22;)))
  (core func $"#core-func71 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush") (;71;) (canon lower (func $"[method]output-stream.blocking-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write-zeroes" (func $"[method]output-stream.write-zeroes" (;23;)))
  (core func $"#core-func72 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes") (;72;) (canon lower (func $"[method]output-stream.write-zeroes") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-zeroes-and-flush" (func $"[method]output-stream.blocking-write-zeroes-and-flush" (;24;)))
  (core func $"#core-func73 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush") (;73;) (canon lower (func $"[method]output-stream.blocking-write-zeroes-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.splice" (func $"[method]output-stream.splice" (;25;)))
  (core func $"#core-func74 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.splice") (;74;) (canon lower (func $"[method]output-stream.splice") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-splice" (func $"[method]output-stream.blocking-splice" (;26;)))
  (core func $"#core-func75 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice") (;75;) (canon lower (func $"[method]output-stream.blocking-splice") (memory $memory)))
  (alias export $wasi:sockets/network@0.2.12 "network-error-code" (func $network-error-code (;27;)))
  (core func $"#core-func76 indirect-wasi:sockets/network@0.2.12-network-error-code" (@name "indirect-wasi:sockets/network@0.2.12-network-error-code") (;76;) (canon lower (func $network-error-code) (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.start-bind" (func $"[method]tcp-socket.start-bind" (;28;)))
  (core func $"#core-func77 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind") (;77;) (canon lower (func $"[method]tcp-socket.start-bind") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.finish-bind" (func $"[method]tcp-socket.finish-bind" (;29;)))
  (core func $"#core-func78 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind") (;78;) (canon lower (func $"[method]tcp-socket.finish-bind") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.start-connect" (func $"[method]tcp-socket.start-connect" (;30;)))
  (core func $"#core-func79 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect") (;79;) (canon lower (func $"[method]tcp-socket.start-connect") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.finish-connect" (func $"[method]tcp-socket.finish-connect" (;31;)))
  (core func $"#core-func80 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect") (;80;) (canon lower (func $"[method]tcp-socket.finish-connect") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.start-listen" (func $"[method]tcp-socket.start-listen" (;32;)))
  (core func $"#core-func81 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen") (;81;) (canon lower (func $"[method]tcp-socket.start-listen") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.finish-listen" (func $"[method]tcp-socket.finish-listen" (;33;)))
  (core func $"#core-func82 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen") (;82;) (canon lower (func $"[method]tcp-socket.finish-listen") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.accept" (func $"[method]tcp-socket.accept" (;34;)))
  (core func $"#core-func83 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept") (;83;) (canon lower (func $"[method]tcp-socket.accept") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.local-address" (func $"[method]tcp-socket.local-address" (;35;)))
  (core func $"#core-func84 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address") (;84;) (canon lower (func $"[method]tcp-socket.local-address") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.remote-address" (func $"[method]tcp-socket.remote-address" (;36;)))
  (core func $"#core-func85 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address") (;85;) (canon lower (func $"[method]tcp-socket.remote-address") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-listen-backlog-size" (func $"[method]tcp-socket.set-listen-backlog-size" (;37;)))
  (core func $"#core-func86 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size") (;86;) (canon lower (func $"[method]tcp-socket.set-listen-backlog-size") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.keep-alive-enabled" (func $"[method]tcp-socket.keep-alive-enabled" (;38;)))
  (core func $"#core-func87 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled") (;87;) (canon lower (func $"[method]tcp-socket.keep-alive-enabled") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-keep-alive-enabled" (func $"[method]tcp-socket.set-keep-alive-enabled" (;39;)))
  (core func $"#core-func88 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled") (;88;) (canon lower (func $"[method]tcp-socket.set-keep-alive-enabled") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.keep-alive-idle-time" (func $"[method]tcp-socket.keep-alive-idle-time" (;40;)))
  (core func $"#core-func89 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time") (;89;) (canon lower (func $"[method]tcp-socket.keep-alive-idle-time") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-keep-alive-idle-time" (func $"[method]tcp-socket.set-keep-alive-idle-time" (;41;)))
  (core func $"#core-func90 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time") (;90;) (canon lower (func $"[method]tcp-socket.set-keep-alive-idle-time") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.keep-alive-interval" (func $"[method]tcp-socket.keep-alive-interval" (;42;)))
  (core func $"#core-func91 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval") (;91;) (canon lower (func $"[method]tcp-socket.keep-alive-interval") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-keep-alive-interval" (func $"[method]tcp-socket.set-keep-alive-interval" (;43;)))
  (core func $"#core-func92 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval") (;92;) (canon lower (func $"[method]tcp-socket.set-keep-alive-interval") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.keep-alive-count" (func $"[method]tcp-socket.keep-alive-count" (;44;)))
  (core func $"#core-func93 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count") (;93;) (canon lower (func $"[method]tcp-socket.keep-alive-count") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-keep-alive-count" (func $"[method]tcp-socket.set-keep-alive-count" (;45;)))
  (core func $"#core-func94 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count") (;94;) (canon lower (func $"[method]tcp-socket.set-keep-alive-count") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.hop-limit" (func $"[method]tcp-socket.hop-limit" (;46;)))
  (core func $"#core-func95 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit") (;95;) (canon lower (func $"[method]tcp-socket.hop-limit") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-hop-limit" (func $"[method]tcp-socket.set-hop-limit" (;47;)))
  (core func $"#core-func96 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit") (;96;) (canon lower (func $"[method]tcp-socket.set-hop-limit") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.receive-buffer-size" (func $"[method]tcp-socket.receive-buffer-size" (;48;)))
  (core func $"#core-func97 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size") (;97;) (canon lower (func $"[method]tcp-socket.receive-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-receive-buffer-size" (func $"[method]tcp-socket.set-receive-buffer-size" (;49;)))
  (core func $"#core-func98 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size") (;98;) (canon lower (func $"[method]tcp-socket.set-receive-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.send-buffer-size" (func $"[method]tcp-socket.send-buffer-size" (;50;)))
  (core func $"#core-func99 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size") (;99;) (canon lower (func $"[method]tcp-socket.send-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.set-send-buffer-size" (func $"[method]tcp-socket.set-send-buffer-size" (;51;)))
  (core func $"#core-func100 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size") (;100;) (canon lower (func $"[method]tcp-socket.set-send-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/tcp@0.2.12 "[method]tcp-socket.shutdown" (func $"[method]tcp-socket.shutdown" (;52;)))
  (core func $"#core-func101 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown" (@name "indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown") (;101;) (canon lower (func $"[method]tcp-socket.shutdown") (memory $memory)))
  (alias export $wasi:sockets/tcp-create-socket@0.2.12 "create-tcp-socket" (func $create-tcp-socket (;53;)))
  (core func $"#core-func102 indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket" (@name "indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket") (;102;) (canon lower (func $create-tcp-socket) (memory $memory)))
  (core instance $actual (;10;)
    (export "0" (func $"#core-func61 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $"#core-func62 indirect-wasi:io/poll@0.2.12-poll"))
    (export "2" (func $"#core-func63 indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "3" (func $"#core-func64 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "4" (func $"#core-func65 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "5" (func $"#core-func66 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "6" (func $"#core-func67 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "7" (func $"#core-func68 indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "8" (func $"#core-func69 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "9" (func $"#core-func70 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "10" (func $"#core-func71 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "11" (func $"#core-func72 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "12" (func $"#core-func73 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "13" (func $"#core-func74 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "14" (func $"#core-func75 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "15" (func $"#core-func76 indirect-wasi:sockets/network@0.2.12-network-error-code"))
    (export "16" (func $"#core-func77 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-bind"))
    (export "17" (func $"#core-func78 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-bind"))
    (export "18" (func $"#core-func79 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-connect"))
    (export "19" (func $"#core-func80 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-connect"))
    (export "20" (func $"#core-func81 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.start-listen"))
    (export "21" (func $"#core-func82 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.finish-listen"))
    (export "22" (func $"#core-func83 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.accept"))
    (export "23" (func $"#core-func84 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.local-address"))
    (export "24" (func $"#core-func85 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.remote-address"))
    (export "25" (func $"#core-func86 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-listen-backlog-size"))
    (export "26" (func $"#core-func87 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-enabled"))
    (export "27" (func $"#core-func88 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-enabled"))
    (export "28" (func $"#core-func89 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-idle-time"))
    (export "29" (func $"#core-func90 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-idle-time"))
    (export "30" (func $"#core-func91 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-interval"))
    (export "31" (func $"#core-func92 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-interval"))
    (export "32" (func $"#core-func93 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.keep-alive-count"))
    (export "33" (func $"#core-func94 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-keep-alive-count"))
    (export "34" (func $"#core-func95 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.hop-limit"))
    (export "35" (func $"#core-func96 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-hop-limit"))
    (export "36" (func $"#core-func97 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.receive-buffer-size"))
    (export "37" (func $"#core-func98 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-receive-buffer-size"))
    (export "38" (func $"#core-func99 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.send-buffer-size"))
    (export "39" (func $"#core-func100 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.set-send-buffer-size"))
    (export "40" (func $"#core-func101 indirect-wasi:sockets/tcp@0.2.12-[method]tcp-socket.shutdown"))
    (export "41" (func $"#core-func102 indirect-wasi:sockets/tcp-create-socket@0.2.12-create-tcp-socket"))
  )
  (core instance $fixup (;11;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;30;) (func (param "family" u32) (result u32)))
  (alias core export $main "create-probe" (core func $create-probe (;103;)))
  (alias core export $main "cabi_post_create-probe" (core func $cabi_post_create-probe (;104;)))
  (func $create-probe (;54;) (type 30) (canon lift (core func $create-probe) (post-return $cabi_post_create-probe)))
  (export $"#func55 create-probe" (@name "create-probe") (;55;) "create-probe" (func $create-probe))
  (alias core export $main "setup" (core func $setup (;105;)))
  (alias core export $main "cabi_post_setup" (core func $cabi_post_setup (;106;)))
  (func $setup (;56;) (type 30) (canon lift (core func $setup) (post-return $cabi_post_setup)))
  (export $"#func57 setup" (@name "setup") (;57;) "setup" (func $setup))
  (type (;31;) (func (result u32)))
  (alias core export $main "teardown" (core func $teardown (;107;)))
  (alias core export $main "cabi_post_teardown" (core func $cabi_post_teardown (;108;)))
  (func $teardown (;58;) (type 31) (canon lift (core func $teardown) (post-return $cabi_post_teardown)))
  (export $"#func59 teardown" (@name "teardown") (;59;) "teardown" (func $teardown))
  (alias core export $main "exercise" (core func $exercise (;109;)))
  (alias core export $main "cabi_post_exercise" (core func $cabi_post_exercise (;110;)))
  (func $exercise (;60;) (type 30) (canon lift (core func $exercise) (post-return $cabi_post_exercise)))
  (export $"#func61 exercise" (@name "exercise") (;61;) "exercise" (func $exercise))
  (type (;32;) (func (param "count" u32) (result u32)))
  (alias core export $main "transfer" (core func $transfer (;111;)))
  (alias core export $main "cabi_post_transfer" (core func $cabi_post_transfer (;112;)))
  (func $transfer (;62;) (type 32) (canon lift (core func $transfer) (post-return $cabi_post_transfer)))
  (export $"#func63 transfer" (@name "transfer") (;63;) "transfer" (func $transfer))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
)

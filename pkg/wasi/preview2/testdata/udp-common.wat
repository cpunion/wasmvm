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
  (type $ty-wasi:sockets/network@0.2.12 (;2;)
    (instance
      (export (;0;) "network" (type (sub resource)))
      (type (;1;) (tuple u8 u8 u8 u8))
      (export (;2;) "ipv4-address" (type (eq 1)))
      (type (;3;) (record (field "port" u16) (field "address" 2)))
      (export (;4;) "ipv4-socket-address" (type (eq 3)))
      (type (;5;) (tuple u16 u16 u16 u16 u16 u16 u16 u16))
      (export (;6;) "ipv6-address" (type (eq 5)))
      (type (;7;) (record (field "port" u16) (field "flow-info" u32) (field "address" 6) (field "scope-id" u32)))
      (export (;8;) "ipv6-socket-address" (type (eq 7)))
      (type (;9;) (variant (case "ipv4" 4) (case "ipv6" 8)))
      (export (;10;) "ip-socket-address" (type (eq 9)))
      (type (;11;) (enum "unknown" "access-denied" "not-supported" "invalid-argument" "out-of-memory" "timeout" "concurrency-conflict" "not-in-progress" "would-block" "invalid-state" "new-socket-limit" "address-not-bindable" "address-in-use" "remote-unreachable" "connection-refused" "connection-reset" "connection-aborted" "datagram-too-large" "name-unresolvable" "temporary-resolver-failure" "permanent-resolver-failure"))
      (export (;12;) "error-code" (type (eq 11)))
      (type (;13;) (enum "ipv4" "ipv6"))
      (export (;14;) "ip-address-family" (type (eq 13)))
    )
  )
  (import "wasi:sockets/network@0.2.12" (instance $wasi:sockets/network@0.2.12 (;2;) (type $ty-wasi:sockets/network@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $network (;3;)))
  (type $ty-wasi:sockets/instance-network@0.2.12 (;4;)
    (instance
      (alias outer 1 $network (type (;0;)))
      (export (;1;) "network" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "instance-network" (func (type 3)))
    )
  )
  (import "wasi:sockets/instance-network@0.2.12" (instance $wasi:sockets/instance-network@0.2.12 (;3;) (type $ty-wasi:sockets/instance-network@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type5 network" (@name "network") (;5;)))
  (alias export $wasi:sockets/network@0.2.12 "ip-socket-address" (type $ip-socket-address (;6;)))
  (alias export $wasi:sockets/network@0.2.12 "error-code" (type $error-code (;7;)))
  (alias export $wasi:sockets/network@0.2.12 "ip-address-family" (type $ip-address-family (;8;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $pollable (;9;)))
  (type $ty-wasi:sockets/udp@0.2.12 (;10;)
    (instance
      (export (;0;) "udp-socket" (type (sub resource)))
      (alias outer 1 $"#type5 network" (type (;1;)))
      (export (;2;) "network" (type (eq 1)))
      (alias outer 1 $ip-socket-address (type (;3;)))
      (export (;4;) "ip-socket-address" (type (eq 3)))
      (alias outer 1 $error-code (type (;5;)))
      (export (;6;) "error-code" (type (eq 5)))
      (export (;7;) "incoming-datagram-stream" (type (sub resource)))
      (export (;8;) "outgoing-datagram-stream" (type (sub resource)))
      (alias outer 1 $ip-address-family (type (;9;)))
      (export (;10;) "ip-address-family" (type (eq 9)))
      (alias outer 1 $pollable (type (;11;)))
      (export (;12;) "pollable" (type (eq 11)))
      (type (;13;) (list u8))
      (type (;14;) (record (field "data" 13) (field "remote-address" 4)))
      (export (;15;) "incoming-datagram" (type (eq 14)))
      (type (;16;) (option 4))
      (type (;17;) (record (field "data" 13) (field "remote-address" 16)))
      (export (;18;) "outgoing-datagram" (type (eq 17)))
      (type (;19;) (borrow 0))
      (type (;20;) (borrow 2))
      (type (;21;) (result (error 6)))
      (type (;22;) (func (param "self" 19) (param "network" 20) (param "local-address" 4) (result 21)))
      (export (;0;) "[method]udp-socket.start-bind" (func (type 22)))
      (type (;23;) (func (param "self" 19) (result 21)))
      (export (;1;) "[method]udp-socket.finish-bind" (func (type 23)))
      (type (;24;) (own 7))
      (type (;25;) (own 8))
      (type (;26;) (tuple 24 25))
      (type (;27;) (result 26 (error 6)))
      (type (;28;) (func (param "self" 19) (param "remote-address" 16) (result 27)))
      (export (;2;) "[method]udp-socket.stream" (func (type 28)))
      (type (;29;) (result 4 (error 6)))
      (type (;30;) (func (param "self" 19) (result 29)))
      (export (;3;) "[method]udp-socket.local-address" (func (type 30)))
      (export (;4;) "[method]udp-socket.remote-address" (func (type 30)))
      (type (;31;) (func (param "self" 19) (result 10)))
      (export (;5;) "[method]udp-socket.address-family" (func (type 31)))
      (type (;32;) (result u8 (error 6)))
      (type (;33;) (func (param "self" 19) (result 32)))
      (export (;6;) "[method]udp-socket.unicast-hop-limit" (func (type 33)))
      (type (;34;) (func (param "self" 19) (param "value" u8) (result 21)))
      (export (;7;) "[method]udp-socket.set-unicast-hop-limit" (func (type 34)))
      (type (;35;) (result u64 (error 6)))
      (type (;36;) (func (param "self" 19) (result 35)))
      (export (;8;) "[method]udp-socket.receive-buffer-size" (func (type 36)))
      (type (;37;) (func (param "self" 19) (param "value" u64) (result 21)))
      (export (;9;) "[method]udp-socket.set-receive-buffer-size" (func (type 37)))
      (export (;10;) "[method]udp-socket.send-buffer-size" (func (type 36)))
      (export (;11;) "[method]udp-socket.set-send-buffer-size" (func (type 37)))
      (type (;38;) (own 12))
      (type (;39;) (func (param "self" 19) (result 38)))
      (export (;12;) "[method]udp-socket.subscribe" (func (type 39)))
      (type (;40;) (borrow 7))
      (type (;41;) (list 15))
      (type (;42;) (result 41 (error 6)))
      (type (;43;) (func (param "self" 40) (param "max-results" u64) (result 42)))
      (export (;13;) "[method]incoming-datagram-stream.receive" (func (type 43)))
      (type (;44;) (func (param "self" 40) (result 38)))
      (export (;14;) "[method]incoming-datagram-stream.subscribe" (func (type 44)))
      (type (;45;) (borrow 8))
      (type (;46;) (func (param "self" 45) (result 35)))
      (export (;15;) "[method]outgoing-datagram-stream.check-send" (func (type 46)))
      (type (;47;) (list 18))
      (type (;48;) (func (param "self" 45) (param "datagrams" 47) (result 35)))
      (export (;16;) "[method]outgoing-datagram-stream.send" (func (type 48)))
      (type (;49;) (func (param "self" 45) (result 38)))
      (export (;17;) "[method]outgoing-datagram-stream.subscribe" (func (type 49)))
    )
  )
  (import "wasi:sockets/udp@0.2.12" (instance $wasi:sockets/udp@0.2.12 (;4;) (type $ty-wasi:sockets/udp@0.2.12)))
  (alias export $wasi:sockets/network@0.2.12 "ip-address-family" (type $"#type11 ip-address-family" (@name "ip-address-family") (;11;)))
  (alias export $wasi:sockets/udp@0.2.12 "udp-socket" (type $udp-socket (;12;)))
  (alias export $wasi:sockets/network@0.2.12 "error-code" (type $"#type13 error-code" (@name "error-code") (;13;)))
  (type $ty-wasi:sockets/udp-create-socket@0.2.12 (;14;)
    (instance
      (alias outer 1 $"#type11 ip-address-family" (type (;0;)))
      (export (;1;) "ip-address-family" (type (eq 0)))
      (alias outer 1 $udp-socket (type (;2;)))
      (export (;3;) "udp-socket" (type (eq 2)))
      (alias outer 1 $"#type13 error-code" (type (;4;)))
      (export (;5;) "error-code" (type (eq 4)))
      (type (;6;) (own 3))
      (type (;7;) (result 6 (error 5)))
      (type (;8;) (func (param "address-family" 1) (result 7)))
      (export (;0;) "create-udp-socket" (func (type 8)))
    )
  )
  (import "wasi:sockets/udp-create-socket@0.2.12" (instance $wasi:sockets/udp-create-socket@0.2.12 (;5;) (type $ty-wasi:sockets/udp-create-socket@0.2.12)))
  (core module $main (;0;)
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
    (import "wasi:sockets/network@0.2.12" "[resource-drop]network" (func $network.drop-network (;6;) (type 1)))
    (import "wasi:sockets/instance-network@0.2.12" "instance-network" (func $instance-network.instance-network (;7;) (type 4)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.start-bind" (func $udp.udp-socket.start-bind (;8;) (type 5)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.finish-bind" (func $udp.udp-socket.finish-bind (;9;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.stream" (func $udp.udp-socket.stream (;10;) (type 5)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.local-address" (func $udp.udp-socket.local-address (;11;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.remote-address" (func $udp.udp-socket.remote-address (;12;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.address-family" (func $udp.udp-socket.address-family (;13;) (type 2)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.unicast-hop-limit" (func $udp.udp-socket.unicast-hop-limit (;14;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-unicast-hop-limit" (func $udp.udp-socket.set-unicast-hop-limit (;15;) (type 3)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.receive-buffer-size" (func $udp.udp-socket.receive-buffer-size (;16;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-receive-buffer-size" (func $udp.udp-socket.set-receive-buffer-size (;17;) (type 6)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.send-buffer-size" (func $udp.udp-socket.send-buffer-size (;18;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.set-send-buffer-size" (func $udp.udp-socket.set-send-buffer-size (;19;) (type 6)))
    (import "wasi:sockets/udp@0.2.12" "[method]udp-socket.subscribe" (func $udp.udp-socket.subscribe (;20;) (type 2)))
    (import "wasi:sockets/udp@0.2.12" "[method]incoming-datagram-stream.receive" (func $udp.incoming-datagram-stream.receive (;21;) (type 6)))
    (import "wasi:sockets/udp@0.2.12" "[method]incoming-datagram-stream.subscribe" (func $udp.incoming-datagram-stream.subscribe (;22;) (type 2)))
    (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.check-send" (func $udp.outgoing-datagram-stream.check-send (;23;) (type 0)))
    (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.send" (func $udp.outgoing-datagram-stream.send (;24;) (type 7)))
    (import "wasi:sockets/udp@0.2.12" "[method]outgoing-datagram-stream.subscribe" (func $udp.outgoing-datagram-stream.subscribe (;25;) (type 2)))
    (import "wasi:sockets/udp@0.2.12" "[resource-drop]udp-socket" (func $udp.drop-udp-socket (;26;) (type 1)))
    (import "wasi:sockets/udp@0.2.12" "[resource-drop]incoming-datagram-stream" (func $udp.drop-incoming-datagram-stream (;27;) (type 1)))
    (import "wasi:sockets/udp@0.2.12" "[resource-drop]outgoing-datagram-stream" (func $udp.drop-outgoing-datagram-stream (;28;) (type 1)))
    (import "wasi:sockets/udp-create-socket@0.2.12" "create-udp-socket" (func $udp-create-socket.create-udp-socket (;29;) (type 0)))
    (memory (;0;) 32 64)
    (global $heap (;0;) (mut i32) i32.const 8192)
    (global $network (;1;) (mut i32) i32.const 0)
    (global $client (;2;) (mut i32) i32.const 0)
    (global $server (;3;) (mut i32) i32.const 0)
    (global $ci (;4;) (mut i32) i32.const 0)
    (global $co (;5;) (mut i32) i32.const 0)
    (global $si (;6;) (mut i32) i32.const 0)
    (global $so (;7;) (mut i32) i32.const 0)
    (global $opened (;8;) (mut i32) i32.const 0)
    (global $connected (;9;) (mut i32) i32.const 0)
    (export "exercise" (func 30))
    (export "cabi_post_exercise" (func 31))
    (export "setup" (func 32))
    (export "cabi_post_setup" (func 33))
    (export "transfer" (func 34))
    (export "cabi_post_transfer" (func 35))
    (export "teardown" (func 36))
    (export "cabi_post_teardown" (func 37))
    (export "memory" (memory 0))
    (export "cabi_realloc" (func 38))
    (export "_initialize" (func 39))
    (func (;30;) (type 2) (param i32) (result i32)
      local.get 0
      i32.const 0
      call $open
      global.get $co
      i32.const 288
      i32.const 0
      call $send
      global.get $si
      i32.const 0
      call $receive
      i32.const 3
      call $exchange
      i32.const 3
      i32.ne
      if ;; label = @1
        unreachable
      end
      call $close
      local.get 0
      i32.const 1
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
    (func (;31;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;32;) (type 8) (param i32 i32) (result i32)
      local.get 0
      local.get 1
      call $open
      i32.const 1
    )
    (func (;33;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;34;) (type 2) (param i32) (result i32)
      local.get 0
      call $exchange
    )
    (func (;35;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;36;) (type 4) (result i32)
      call $close
      i32.const 1
    )
    (func (;37;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;38;) (type 9) (param i32 i32 i32 i32) (result i32)
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
    (func (;39;) (type 10))
    (func $ok (;40;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (func $open (;41;) (type 0) (param $family i32) (param $connected i32)
      (local $cp i32) (local $sp i32) (local $poll i32)
      global.get $opened
      if ;; label = @1
        unreachable
      end
      local.get $connected
      global.set $connected
      call $instance-network.instance-network
      global.set $network
      local.get $family
      i32.const 64
      call $udp-create-socket.create-udp-socket
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $client
      local.get $family
      i32.const 64
      call $udp-create-socket.create-udp-socket
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $server
      global.get $client
      call $udp.udp-socket.address-family
      local.get $family
      i32.ne
      if ;; label = @1
        unreachable
      end
      global.get $client
      i32.const 32
      i32.const 64
      call $udp.udp-socket.set-unicast-hop-limit
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $udp.udp-socket.unicast-hop-limit
      i32.const 64
      call $ok
      global.get $client
      i64.const 65536
      i32.const 64
      call $udp.udp-socket.set-receive-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $udp.udp-socket.receive-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i64.const 65536
      i32.const 64
      call $udp.udp-socket.set-send-buffer-size
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $udp.udp-socket.send-buffer-size
      i32.const 64
      call $ok
      global.get $client
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
      call $udp.udp-socket.start-bind
      i32.const 64
      call $ok
      global.get $client
      i32.const 64
      call $udp.udp-socket.finish-bind
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
      call $udp.udp-socket.start-bind
      i32.const 64
      call $ok
      global.get $server
      i32.const 64
      call $udp.udp-socket.finish-bind
      i32.const 64
      call $ok
      global.get $client
      call $udp.udp-socket.subscribe
      local.tee $poll
      call $poll.pollable.block
      local.get $poll
      call $poll.drop-pollable
      global.get $client
      i32.const 64
      call $udp.udp-socket.local-address
      i32.const 64
      call $ok
      i32.const 256
      i32.const 68
      i32.const 32
      memory.copy
      i32.const 72
      i32.load16_u
      local.tee $cp
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $server
      i32.const 64
      call $udp.udp-socket.local-address
      i32.const 64
      call $ok
      i32.const 288
      i32.const 68
      i32.const 32
      memory.copy
      i32.const 72
      i32.load16_u
      local.tee $sp
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $client
      local.get $connected
      local.get $family
      local.get $sp
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
      call $udp.udp-socket.stream
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $ci
      i32.const 72
      i32.load
      global.set $co
      global.get $server
      local.get $connected
      local.get $family
      local.get $cp
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
      call $udp.udp-socket.stream
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      global.set $si
      i32.const 72
      i32.load
      global.set $so
      local.get $connected
      if ;; label = @1
        global.get $client
        i32.const 64
        call $udp.udp-socket.remote-address
        i32.const 64
        call $ok
        i32.const 72
        i32.load16_u
        local.get $sp
        i32.ne
        if ;; label = @2
          unreachable
        end
      end
      global.get $co
      call $udp.outgoing-datagram-stream.subscribe
      local.tee $poll
      call $poll.pollable.block
      local.get $poll
      call $poll.drop-pollable
      global.get $si
      i64.const 0
      i32.const 64
      call $udp.incoming-datagram-stream.receive
      i32.const 64
      call $ok
      i32.const 72
      i32.load
      if ;; label = @1
        unreachable
      end
      i32.const 1
      global.set $opened
    )
    (func $send (;42;) (type 3) (param $stream i32) (param $peer i32) (param $length i32)
      (local $poll i32)
      block $permitted
        loop $again
          local.get $stream
          i32.const 64
          call $udp.outgoing-datagram-stream.check-send
          i32.const 64
          call $ok
          i32.const 72
          i64.load
          i64.const 0
          i64.gt_u
          br_if $permitted
          local.get $stream
          call $udp.outgoing-datagram-stream.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $again
        end
      end
      i32.const 512
      i32.const 1024
      i32.store
      i32.const 516
      local.get $length
      i32.store
      i32.const 520
      i32.const 1
      i32.store8
      i32.const 524
      local.get $peer
      i32.const 32
      memory.copy
      local.get $stream
      i32.const 512
      i32.const 1
      i32.const 64
      call $udp.outgoing-datagram-stream.send
      i32.const 64
      call $ok
      i32.const 72
      i64.load
      i64.const 1
      i64.ne
      if ;; label = @1
        unreachable
      end
    )
    (func $receive (;43;) (type 0) (param $stream i32) (param $length i32)
      (local $poll i32) (local $item i32) (local $data i32)
      block $received
        loop $again
          local.get $stream
          i64.const 1
          i32.const 64
          call $udp.incoming-datagram-stream.receive
          i32.const 64
          call $ok
          i32.const 72
          i32.load
          br_if $received
          local.get $stream
          call $udp.incoming-datagram-stream.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $again
        end
      end
      i32.const 72
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 68
      i32.load
      local.tee $item
      i32.const 4
      i32.add
      i32.load
      local.get $length
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $length
      if ;; label = @1
        local.get $item
        i32.load
        local.tee $data
        i32.load8_u
        i32.const 48
        i32.ne
        if ;; label = @2
          unreachable
        end
        local.get $data
        local.get $length
        i32.add
        i32.const 1
        i32.sub
        i32.load8_u
        i32.const 1024
        local.get $length
        i32.const 1
        i32.sub
        i32.const 15
        i32.and
        i32.add
        i32.load8_u
        i32.ne
        if ;; label = @2
          unreachable
        end
      end
    )
    (func $exchange (;44;) (type 2) (param $count i32) (result i32)
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
          i32.const 288
          i32.const 4096
          call $send
          global.get $si
          i32.const 4096
          call $receive
          global.get $so
          i32.const 256
          i32.const 4096
          call $send
          global.get $ci
          i32.const 4096
          call $receive
          local.get $at
          i32.const 1
          i32.add
          local.set $at
          br $next
        end
      end
      local.get $at
    )
    (func $close (;45;) (type 10)
      global.get $opened
      i32.eqz
      if ;; label = @1
        unreachable
      end
      global.get $ci
      call $udp.drop-incoming-datagram-stream
      global.get $co
      call $udp.drop-outgoing-datagram-stream
      global.get $si
      call $udp.drop-incoming-datagram-stream
      global.get $so
      call $udp.drop-outgoing-datagram-stream
      global.get $client
      call $udp.drop-udp-socket
      global.get $server
      call $udp.drop-udp-socket
      global.get $network
      call $network.drop-network
      i32.const 0
      global.set $opened
    )
    (data (;0;) (i32.const 1024) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core module $wit-component-shim-module (;1;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
    (type (;3;) (func (param i32 i64 i32)))
    (type (;4;) (func (param i32 i32 i32 i32)))
    (table (;0;) 17 17 funcref)
    (export "0" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "2" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind"))
    (export "3" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind"))
    (export "4" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream"))
    (export "5" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address"))
    (export "6" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address"))
    (export "7" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit"))
    (export "8" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit"))
    (export "9" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size"))
    (export "10" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size"))
    (export "11" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size"))
    (export "12" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size"))
    (export "13" (func $"indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive"))
    (export "14" (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send"))
    (export "15" (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send"))
    (export "16" (func $indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket))
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
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind" (;2;) (type 2) (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)
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
      i32.const 2
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind" (;3;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 3
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream" (;4;) (type 2) (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)
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
      i32.const 4
      call_indirect (type 2)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address" (;5;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 5
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address" (;6;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 6
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit" (;7;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 7
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit" (;8;) (type 1) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 8
      call_indirect (type 1)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size" (;9;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 9
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size" (;10;) (type 3) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 10
      call_indirect (type 3)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size" (;11;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 11
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size" (;12;) (type 3) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 12
      call_indirect (type 3)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive" (;13;) (type 3) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 13
      call_indirect (type 3)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send" (;14;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 14
      call_indirect (type 0)
    )
    (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send" (;15;) (type 4) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 15
      call_indirect (type 4)
    )
    (func $indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket (;16;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 16
      call_indirect (type 0)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias core export $wit-component-shim-instance "0" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $error (;15;)))
  (core func $resource.drop (;1;) (canon resource.drop $error))
  (core instance $wasi:io/error@0.2.12 (;1;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $resource.drop))
  )
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;2;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;3;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "1" (core func $indirect-wasi:io/poll@0.2.12-poll (;4;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type16 pollable" (@name "pollable") (;16;)))
  (core func $"#core-func5 resource.drop" (@name "resource.drop") (;5;) (canon resource.drop $"#type16 pollable"))
  (core instance $wasi:io/poll@0.2.12 (;2;)
    (export "[method]pollable.ready" (func $"[method]pollable.ready"))
    (export "[method]pollable.block" (func $"[method]pollable.block"))
    (export "poll" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "[resource-drop]pollable" (func $"#core-func5 resource.drop"))
  )
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type17 network" (@name "network") (;17;)))
  (core func $"#core-func6 resource.drop" (@name "resource.drop") (;6;) (canon resource.drop $"#type17 network"))
  (core instance $wasi:sockets/network@0.2.12 (;3;)
    (export "[resource-drop]network" (func $"#core-func6 resource.drop"))
  )
  (alias export $wasi:sockets/instance-network@0.2.12 "instance-network" (func $instance-network (;2;)))
  (core func $instance-network (;7;) (canon lower (func $instance-network)))
  (core instance $wasi:sockets/instance-network@0.2.12 (;4;)
    (export "instance-network" (func $instance-network))
  )
  (alias core export $wit-component-shim-instance "2" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind" (;8;)))
  (alias core export $wit-component-shim-instance "3" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind" (;9;)))
  (alias core export $wit-component-shim-instance "4" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream" (;10;)))
  (alias core export $wit-component-shim-instance "5" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address" (;11;)))
  (alias core export $wit-component-shim-instance "6" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address" (;12;)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.address-family" (func $"[method]udp-socket.address-family" (;3;)))
  (core func $"[method]udp-socket.address-family" (;13;) (canon lower (func $"[method]udp-socket.address-family")))
  (alias core export $wit-component-shim-instance "7" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit" (;14;)))
  (alias core export $wit-component-shim-instance "8" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit" (;15;)))
  (alias core export $wit-component-shim-instance "9" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size" (;16;)))
  (alias core export $wit-component-shim-instance "10" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size" (;17;)))
  (alias core export $wit-component-shim-instance "11" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size" (;18;)))
  (alias core export $wit-component-shim-instance "12" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size" (;19;)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.subscribe" (func $"[method]udp-socket.subscribe" (;4;)))
  (core func $"[method]udp-socket.subscribe" (;20;) (canon lower (func $"[method]udp-socket.subscribe")))
  (alias core export $wit-component-shim-instance "13" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive" (;21;)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]incoming-datagram-stream.subscribe" (func $"[method]incoming-datagram-stream.subscribe" (;5;)))
  (core func $"[method]incoming-datagram-stream.subscribe" (;22;) (canon lower (func $"[method]incoming-datagram-stream.subscribe")))
  (alias core export $wit-component-shim-instance "14" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send" (;23;)))
  (alias core export $wit-component-shim-instance "15" (core func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send" (;24;)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]outgoing-datagram-stream.subscribe" (func $"[method]outgoing-datagram-stream.subscribe" (;6;)))
  (core func $"[method]outgoing-datagram-stream.subscribe" (;25;) (canon lower (func $"[method]outgoing-datagram-stream.subscribe")))
  (alias export $wasi:sockets/udp@0.2.12 "udp-socket" (type $"#type18 udp-socket" (@name "udp-socket") (;18;)))
  (core func $"#core-func26 resource.drop" (@name "resource.drop") (;26;) (canon resource.drop $"#type18 udp-socket"))
  (alias export $wasi:sockets/udp@0.2.12 "incoming-datagram-stream" (type $incoming-datagram-stream (;19;)))
  (core func $"#core-func27 resource.drop" (@name "resource.drop") (;27;) (canon resource.drop $incoming-datagram-stream))
  (alias export $wasi:sockets/udp@0.2.12 "outgoing-datagram-stream" (type $outgoing-datagram-stream (;20;)))
  (core func $"#core-func28 resource.drop" (@name "resource.drop") (;28;) (canon resource.drop $outgoing-datagram-stream))
  (core instance $wasi:sockets/udp@0.2.12 (;5;)
    (export "[method]udp-socket.start-bind" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind"))
    (export "[method]udp-socket.finish-bind" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind"))
    (export "[method]udp-socket.stream" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream"))
    (export "[method]udp-socket.local-address" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address"))
    (export "[method]udp-socket.remote-address" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address"))
    (export "[method]udp-socket.address-family" (func $"[method]udp-socket.address-family"))
    (export "[method]udp-socket.unicast-hop-limit" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit"))
    (export "[method]udp-socket.set-unicast-hop-limit" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit"))
    (export "[method]udp-socket.receive-buffer-size" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size"))
    (export "[method]udp-socket.set-receive-buffer-size" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size"))
    (export "[method]udp-socket.send-buffer-size" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size"))
    (export "[method]udp-socket.set-send-buffer-size" (func $"indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size"))
    (export "[method]udp-socket.subscribe" (func $"[method]udp-socket.subscribe"))
    (export "[method]incoming-datagram-stream.receive" (func $"indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive"))
    (export "[method]incoming-datagram-stream.subscribe" (func $"[method]incoming-datagram-stream.subscribe"))
    (export "[method]outgoing-datagram-stream.check-send" (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send"))
    (export "[method]outgoing-datagram-stream.send" (func $"indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send"))
    (export "[method]outgoing-datagram-stream.subscribe" (func $"[method]outgoing-datagram-stream.subscribe"))
    (export "[resource-drop]udp-socket" (func $"#core-func26 resource.drop"))
    (export "[resource-drop]incoming-datagram-stream" (func $"#core-func27 resource.drop"))
    (export "[resource-drop]outgoing-datagram-stream" (func $"#core-func28 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "16" (core func $indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket (;29;)))
  (core instance $wasi:sockets/udp-create-socket@0.2.12 (;6;)
    (export "create-udp-socket" (func $indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket))
  )
  (core instance $main (;7;) (instantiate $main
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:sockets/network@0.2.12" (instance $wasi:sockets/network@0.2.12))
      (with "wasi:sockets/instance-network@0.2.12" (instance $wasi:sockets/instance-network@0.2.12))
      (with "wasi:sockets/udp@0.2.12" (instance $wasi:sockets/udp@0.2.12))
      (with "wasi:sockets/udp-create-socket@0.2.12" (instance $wasi:sockets/udp-create-socket@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
    (type (;3;) (func (param i32 i64 i32)))
    (type (;4;) (func (param i32 i32 i32 i32)))
    (type (;5;) (func))
    (import "actual" "0" (func $0 (;0;) (type 0)))
    (import "actual" "1" (func $1 (;1;) (type 1)))
    (import "actual" "2" (func $2 (;2;) (type 2)))
    (import "actual" "3" (func $3 (;3;) (type 0)))
    (import "actual" "4" (func $4 (;4;) (type 2)))
    (import "actual" "5" (func $5 (;5;) (type 0)))
    (import "actual" "6" (func $6 (;6;) (type 0)))
    (import "actual" "7" (func $7 (;7;) (type 0)))
    (import "actual" "8" (func $8 (;8;) (type 1)))
    (import "actual" "9" (func $9 (;9;) (type 0)))
    (import "actual" "10" (func $10 (;10;) (type 3)))
    (import "actual" "11" (func $11 (;11;) (type 0)))
    (import "actual" "12" (func $12 (;12;) (type 3)))
    (import "actual" "13" (func $13 (;13;) (type 3)))
    (import "actual" "14" (func $14 (;14;) (type 0)))
    (import "actual" "15" (func $15 (;15;) (type 4)))
    (import "actual" "16" (func $16 (;16;) (type 0)))
    (import "main" "_initialize" (func $_initialize (;17;) (type 5)))
    (import "shim" "$imports" (table (;0;) 17 17 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;7;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;30;)))
  (core func $"#core-func31 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;31;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;8;)))
  (core func $"#core-func32 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;32;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.start-bind" (func $"[method]udp-socket.start-bind" (;9;)))
  (core func $"#core-func33 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind") (;33;) (canon lower (func $"[method]udp-socket.start-bind") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.finish-bind" (func $"[method]udp-socket.finish-bind" (;10;)))
  (core func $"#core-func34 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind") (;34;) (canon lower (func $"[method]udp-socket.finish-bind") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.stream" (func $"[method]udp-socket.stream" (;11;)))
  (core func $"#core-func35 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream") (;35;) (canon lower (func $"[method]udp-socket.stream") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.local-address" (func $"[method]udp-socket.local-address" (;12;)))
  (core func $"#core-func36 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address") (;36;) (canon lower (func $"[method]udp-socket.local-address") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.remote-address" (func $"[method]udp-socket.remote-address" (;13;)))
  (core func $"#core-func37 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address") (;37;) (canon lower (func $"[method]udp-socket.remote-address") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.unicast-hop-limit" (func $"[method]udp-socket.unicast-hop-limit" (;14;)))
  (core func $"#core-func38 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit") (;38;) (canon lower (func $"[method]udp-socket.unicast-hop-limit") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.set-unicast-hop-limit" (func $"[method]udp-socket.set-unicast-hop-limit" (;15;)))
  (core func $"#core-func39 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit") (;39;) (canon lower (func $"[method]udp-socket.set-unicast-hop-limit") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.receive-buffer-size" (func $"[method]udp-socket.receive-buffer-size" (;16;)))
  (core func $"#core-func40 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size") (;40;) (canon lower (func $"[method]udp-socket.receive-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.set-receive-buffer-size" (func $"[method]udp-socket.set-receive-buffer-size" (;17;)))
  (core func $"#core-func41 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size") (;41;) (canon lower (func $"[method]udp-socket.set-receive-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.send-buffer-size" (func $"[method]udp-socket.send-buffer-size" (;18;)))
  (core func $"#core-func42 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size") (;42;) (canon lower (func $"[method]udp-socket.send-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]udp-socket.set-send-buffer-size" (func $"[method]udp-socket.set-send-buffer-size" (;19;)))
  (core func $"#core-func43 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size" (@name "indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size") (;43;) (canon lower (func $"[method]udp-socket.set-send-buffer-size") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]incoming-datagram-stream.receive" (func $"[method]incoming-datagram-stream.receive" (;20;)))
  (core func $"#core-func44 indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive" (@name "indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive") (;44;) (canon lower (func $"[method]incoming-datagram-stream.receive") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]outgoing-datagram-stream.check-send" (func $"[method]outgoing-datagram-stream.check-send" (;21;)))
  (core func $"#core-func45 indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send" (@name "indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send") (;45;) (canon lower (func $"[method]outgoing-datagram-stream.check-send") (memory $memory)))
  (alias export $wasi:sockets/udp@0.2.12 "[method]outgoing-datagram-stream.send" (func $"[method]outgoing-datagram-stream.send" (;22;)))
  (core func $"#core-func46 indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send" (@name "indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send") (;46;) (canon lower (func $"[method]outgoing-datagram-stream.send") (memory $memory)))
  (alias export $wasi:sockets/udp-create-socket@0.2.12 "create-udp-socket" (func $create-udp-socket (;23;)))
  (core func $"#core-func47 indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket" (@name "indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket") (;47;) (canon lower (func $create-udp-socket) (memory $memory)))
  (core instance $actual (;8;)
    (export "0" (func $"#core-func31 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $"#core-func32 indirect-wasi:io/poll@0.2.12-poll"))
    (export "2" (func $"#core-func33 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.start-bind"))
    (export "3" (func $"#core-func34 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.finish-bind"))
    (export "4" (func $"#core-func35 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.stream"))
    (export "5" (func $"#core-func36 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.local-address"))
    (export "6" (func $"#core-func37 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.remote-address"))
    (export "7" (func $"#core-func38 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.unicast-hop-limit"))
    (export "8" (func $"#core-func39 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-unicast-hop-limit"))
    (export "9" (func $"#core-func40 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.receive-buffer-size"))
    (export "10" (func $"#core-func41 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-receive-buffer-size"))
    (export "11" (func $"#core-func42 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.send-buffer-size"))
    (export "12" (func $"#core-func43 indirect-wasi:sockets/udp@0.2.12-[method]udp-socket.set-send-buffer-size"))
    (export "13" (func $"#core-func44 indirect-wasi:sockets/udp@0.2.12-[method]incoming-datagram-stream.receive"))
    (export "14" (func $"#core-func45 indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.check-send"))
    (export "15" (func $"#core-func46 indirect-wasi:sockets/udp@0.2.12-[method]outgoing-datagram-stream.send"))
    (export "16" (func $"#core-func47 indirect-wasi:sockets/udp-create-socket@0.2.12-create-udp-socket"))
  )
  (core instance $fixup (;9;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;21;) (func (param "family" u32) (result u32)))
  (alias core export $main "exercise" (core func $exercise (;48;)))
  (alias core export $main "cabi_post_exercise" (core func $cabi_post_exercise (;49;)))
  (func $exercise (;24;) (type 21) (canon lift (core func $exercise) (post-return $cabi_post_exercise)))
  (export $"#func25 exercise" (@name "exercise") (;25;) "exercise" (func $exercise))
  (type (;22;) (func (param "family" u32) (param "connected" bool) (result u32)))
  (alias core export $main "setup" (core func $setup (;50;)))
  (alias core export $main "cabi_post_setup" (core func $cabi_post_setup (;51;)))
  (func $setup (;26;) (type 22) (canon lift (core func $setup) (post-return $cabi_post_setup)))
  (export $"#func27 setup" (@name "setup") (;27;) "setup" (func $setup))
  (type (;23;) (func (param "count" u32) (result u32)))
  (alias core export $main "transfer" (core func $transfer (;52;)))
  (alias core export $main "cabi_post_transfer" (core func $cabi_post_transfer (;53;)))
  (func $transfer (;28;) (type 23) (canon lift (core func $transfer) (post-return $cabi_post_transfer)))
  (export $"#func29 transfer" (@name "transfer") (;29;) "transfer" (func $transfer))
  (type (;24;) (func (result u32)))
  (alias core export $main "teardown" (core func $teardown (;54;)))
  (alias core export $main "cabi_post_teardown" (core func $cabi_post_teardown (;55;)))
  (func $teardown (;30;) (type 24) (canon lift (core func $teardown) (post-return $cabi_post_teardown)))
  (export $"#func31 teardown" (@name "teardown") (;31;) "teardown" (func $teardown))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
)

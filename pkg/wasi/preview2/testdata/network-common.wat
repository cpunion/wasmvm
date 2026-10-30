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
      (type (;3;) (tuple u16 u16 u16 u16 u16 u16 u16 u16))
      (export (;4;) "ipv6-address" (type (eq 3)))
      (type (;5;) (variant (case "ipv4" 2) (case "ipv6" 4)))
      (export (;6;) "ip-address" (type (eq 5)))
      (type (;7;) (enum "unknown" "access-denied" "not-supported" "invalid-argument" "out-of-memory" "timeout" "concurrency-conflict" "not-in-progress" "would-block" "invalid-state" "new-socket-limit" "address-not-bindable" "address-in-use" "remote-unreachable" "connection-refused" "connection-reset" "connection-aborted" "datagram-too-large" "name-unresolvable" "temporary-resolver-failure" "permanent-resolver-failure"))
      (export (;8;) "error-code" (type (eq 7)))
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
  (alias export $wasi:sockets/network@0.2.12 "ip-address" (type $ip-address (;5;)))
  (alias export $wasi:sockets/network@0.2.12 "error-code" (type $error-code (;6;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $pollable (;7;)))
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type8 network" (@name "network") (;8;)))
  (type $ty-wasi:sockets/ip-name-lookup@0.2.12 (;9;)
    (instance
      (export (;0;) "resolve-address-stream" (type (sub resource)))
      (alias outer 1 $ip-address (type (;1;)))
      (export (;2;) "ip-address" (type (eq 1)))
      (alias outer 1 $error-code (type (;3;)))
      (export (;4;) "error-code" (type (eq 3)))
      (alias outer 1 $pollable (type (;5;)))
      (export (;6;) "pollable" (type (eq 5)))
      (alias outer 1 $"#type8 network" (type (;7;)))
      (export (;8;) "network" (type (eq 7)))
      (type (;9;) (borrow 0))
      (type (;10;) (option 2))
      (type (;11;) (result 10 (error 4)))
      (type (;12;) (func (param "self" 9) (result 11)))
      (export (;0;) "[method]resolve-address-stream.resolve-next-address" (func (type 12)))
      (type (;13;) (own 6))
      (type (;14;) (func (param "self" 9) (result 13)))
      (export (;1;) "[method]resolve-address-stream.subscribe" (func (type 14)))
      (type (;15;) (borrow 8))
      (type (;16;) (own 0))
      (type (;17;) (result 16 (error 4)))
      (type (;18;) (func (param "network" 15) (param "name" string) (result 17)))
      (export (;2;) "resolve-addresses" (func (type 18)))
    )
  )
  (import "wasi:sockets/ip-name-lookup@0.2.12" (instance $wasi:sockets/ip-name-lookup@0.2.12 (;4;) (type $ty-wasi:sockets/ip-name-lookup@0.2.12)))
  (core module $main (;0;)
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
    (global $heap (;0;) (mut i32) i32.const 8192)
    (export "exercise" (func 12))
    (export "cabi_post_exercise" (func 13))
    (export "lookup-batch" (func 14))
    (export "cabi_post_lookup-batch" (func 15))
    (export "memory" (memory 0))
    (export "cabi_realloc" (func 16))
    (export "_initialize" (func 17))
    (func (;12;) (type 4) (result i32)
      (local $network i32) (local $stream i32) (local $poll i32)
      call 7
      local.set $network
      local.get $network
      i32.const 1024
      i32.const 9
      i32.const 64
      call 10
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      call 9
      local.tee $poll
      call 3
      local.get $poll
      call 2
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $stream
      i32.const 64
      call 8
      i32.const 64
      call $ok
      i32.const 66
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 68
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 70
      i32.load
      i32.const 16777343
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $stream
      i32.const 64
      call 8
      i32.const 64
      call $ok
      i32.const 66
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      local.get $poll
      call 5
      local.get $stream
      call 11
      local.get $network
      i32.const 1040
      i32.const 3
      i32.const 64
      call 10
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      call 9
      local.tee $poll
      call 3
      local.get $stream
      i32.const 64
      call 8
      i32.const 64
      call $ok
      i32.const 66
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 68
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 84
      i32.load16_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $poll
      call 5
      local.get $stream
      call 11
      local.get $network
      call 6
      i32.const 42
    )
    (func (;13;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;14;) (type 2) (param $count i32) (result i32)
      (local $network i32) (local $stream i32) (local $poll i32) (local $at i32)
      call 7
      local.set $network
      block $end
        loop $next
          local.get $at
          local.get $count
          i32.ge_u
          br_if $end
          local.get $network
          i32.const 1024
          i32.const 9
          i32.const 64
          call 10
          i32.const 64
          call $ok
          i32.const 68
          i32.load
          local.set $stream
          local.get $stream
          call 9
          local.tee $poll
          call 3
          local.get $poll
          call 2
          i32.eqz
          if ;; label = @3
            unreachable
          end
          local.get $stream
          i32.const 64
          call 8
          i32.const 64
          call $ok
          i32.const 66
          i32.load8_u
          i32.const 1
          i32.ne
          if ;; label = @3
            unreachable
          end
          i32.const 68
          i32.load8_u
          if ;; label = @3
            unreachable
          end
          i32.const 70
          i32.load
          i32.const 16777343
          i32.ne
          if ;; label = @3
            unreachable
          end
          local.get $stream
          i32.const 64
          call 8
          i32.const 64
          call $ok
          i32.const 66
          i32.load8_u
          if ;; label = @3
            unreachable
          end
          local.get $poll
          call 5
          local.get $stream
          call 11
          local.get $at
          i32.const 1
          i32.add
          local.set $at
          br $next
        end
      end
      local.get $network
      call 6
      local.get $at
    )
    (func (;15;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;16;) (type 6) (param i32 i32 i32 i32) (result i32)
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
    (func (;17;) (type 7))
    (func $ok (;18;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (data (;0;) (i32.const 1024) "127.0.0.1")
    (data (;1;) (i32.const 1040) "::1")
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core module $wit-component-shim-module (;1;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i32 i32 i32)))
    (table (;0;) 4 4 funcref)
    (export "0" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "2" (func $"indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address"))
    (export "3" (func $indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses))
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
    (func $"indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address" (;2;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 2
      call_indirect (type 0)
    )
    (func $indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses (;3;) (type 2) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 3
      call_indirect (type 2)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias core export $wit-component-shim-instance "0" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $error (;10;)))
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
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type11 pollable" (@name "pollable") (;11;)))
  (core func $"#core-func5 resource.drop" (@name "resource.drop") (;5;) (canon resource.drop $"#type11 pollable"))
  (core instance $wasi:io/poll@0.2.12 (;2;)
    (export "[method]pollable.ready" (func $"[method]pollable.ready"))
    (export "[method]pollable.block" (func $"[method]pollable.block"))
    (export "poll" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "[resource-drop]pollable" (func $"#core-func5 resource.drop"))
  )
  (alias export $wasi:sockets/network@0.2.12 "network" (type $"#type12 network" (@name "network") (;12;)))
  (core func $"#core-func6 resource.drop" (@name "resource.drop") (;6;) (canon resource.drop $"#type12 network"))
  (core instance $wasi:sockets/network@0.2.12 (;3;)
    (export "[resource-drop]network" (func $"#core-func6 resource.drop"))
  )
  (alias export $wasi:sockets/instance-network@0.2.12 "instance-network" (func $instance-network (;2;)))
  (core func $instance-network (;7;) (canon lower (func $instance-network)))
  (core instance $wasi:sockets/instance-network@0.2.12 (;4;)
    (export "instance-network" (func $instance-network))
  )
  (alias core export $wit-component-shim-instance "2" (core func $"indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address" (;8;)))
  (alias export $wasi:sockets/ip-name-lookup@0.2.12 "[method]resolve-address-stream.subscribe" (func $"[method]resolve-address-stream.subscribe" (;3;)))
  (core func $"[method]resolve-address-stream.subscribe" (;9;) (canon lower (func $"[method]resolve-address-stream.subscribe")))
  (alias core export $wit-component-shim-instance "3" (core func $indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses (;10;)))
  (alias export $wasi:sockets/ip-name-lookup@0.2.12 "resolve-address-stream" (type $resolve-address-stream (;13;)))
  (core func $"#core-func11 resource.drop" (@name "resource.drop") (;11;) (canon resource.drop $resolve-address-stream))
  (core instance $wasi:sockets/ip-name-lookup@0.2.12 (;5;)
    (export "[method]resolve-address-stream.resolve-next-address" (func $"indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address"))
    (export "[method]resolve-address-stream.subscribe" (func $"[method]resolve-address-stream.subscribe"))
    (export "resolve-addresses" (func $indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses))
    (export "[resource-drop]resolve-address-stream" (func $"#core-func11 resource.drop"))
  )
  (core instance $main (;6;) (instantiate $main
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:sockets/network@0.2.12" (instance $wasi:sockets/network@0.2.12))
      (with "wasi:sockets/instance-network@0.2.12" (instance $wasi:sockets/instance-network@0.2.12))
      (with "wasi:sockets/ip-name-lookup@0.2.12" (instance $wasi:sockets/ip-name-lookup@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i32 i32 i32)))
    (type (;3;) (func))
    (import "actual" "0" (func $0 (;0;) (type 0)))
    (import "actual" "1" (func $1 (;1;) (type 1)))
    (import "actual" "2" (func $2 (;2;) (type 0)))
    (import "actual" "3" (func $3 (;3;) (type 2)))
    (import "main" "_initialize" (func $_initialize (;4;) (type 3)))
    (import "shim" "$imports" (table (;0;) 4 4 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;4;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;12;)))
  (core func $"#core-func13 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;13;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;5;)))
  (core func $"#core-func14 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;14;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:sockets/ip-name-lookup@0.2.12 "[method]resolve-address-stream.resolve-next-address" (func $"[method]resolve-address-stream.resolve-next-address" (;6;)))
  (core func $"#core-func15 indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address" (@name "indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address") (;15;) (canon lower (func $"[method]resolve-address-stream.resolve-next-address") (memory $memory)))
  (alias export $wasi:sockets/ip-name-lookup@0.2.12 "resolve-addresses" (func $resolve-addresses (;7;)))
  (core func $"#core-func16 indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses" (@name "indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses") (;16;) (canon lower (func $resolve-addresses) (memory $memory) string-encoding=utf8))
  (core instance $actual (;7;)
    (export "0" (func $"#core-func13 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $"#core-func14 indirect-wasi:io/poll@0.2.12-poll"))
    (export "2" (func $"#core-func15 indirect-wasi:sockets/ip-name-lookup@0.2.12-[method]resolve-address-stream.resolve-next-address"))
    (export "3" (func $"#core-func16 indirect-wasi:sockets/ip-name-lookup@0.2.12-resolve-addresses"))
  )
  (core instance $fixup (;8;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;14;) (func (result u32)))
  (alias core export $main "exercise" (core func $exercise (;17;)))
  (alias core export $main "cabi_post_exercise" (core func $cabi_post_exercise (;18;)))
  (func $exercise (;8;) (type 14) (canon lift (core func $exercise) (post-return $cabi_post_exercise)))
  (export $"#func9 exercise" (@name "exercise") (;9;) "exercise" (func $exercise))
  (type (;15;) (func (param "count" u32) (result u32)))
  (alias core export $main "lookup-batch" (core func $lookup-batch (;19;)))
  (alias core export $main "cabi_post_lookup-batch" (core func $cabi_post_lookup-batch (;20;)))
  (func $lookup-batch (;10;) (type 15) (canon lift (core func $lookup-batch) (post-return $cabi_post_lookup-batch)))
  (export $"#func11 lookup-batch" (@name "lookup-batch") (;11;) "lookup-batch" (func $lookup-batch))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
)

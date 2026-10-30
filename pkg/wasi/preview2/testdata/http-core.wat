(module
  (type (;0;) (func (param i32) (result i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32 i32 i32)))
  (type (;3;) (func (result i64)))
  (type (;4;) (func (param i64) (result i32)))
  (type (;5;) (func (param i64 i32)))
  (type (;6;) (func (param i32 i32)))
  (type (;7;) (func (param i32 i64 i32)))
  (type (;8;) (func (param i32 i32 i32 i32)))
  (type (;9;) (func (param i32 i32 i64 i32)))
  (type (;10;) (func (result i32)))
  (type (;11;) (func (param i32 i32 i32) (result i32)))
  (type (;12;) (func (param i32 i32 i32 i32 i32 i32)))
  (type (;13;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;14;) (func (param i32 i32 i32 i32 i32) (result i32)))
  (type (;15;) (func (param i32 i32 i64) (result i32)))
  (type (;16;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
  (type (;17;) (func (param i32 i32) (result i32)))
  (type (;18;) (func))
  (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func $poll.pollable.ready (;0;) (type 0)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func $poll.pollable.block (;1;) (type 1)))
  (import "wasi:io/poll@0.2.12" "poll" (func $poll.poll (;2;) (type 2)))
  (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func $poll.drop-pollable (;3;) (type 1)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "now" (func $monotonic-clock.now (;4;) (type 3)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "resolution" (func $monotonic-clock.resolution (;5;) (type 3)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-instant" (func $monotonic-clock.subscribe-instant (;6;) (type 4)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-duration" (func $monotonic-clock.subscribe-duration (;7;) (type 4)))
  (import "wasi:clocks/wall-clock@0.2.12" "now" (func $wall-clock.now (;8;) (type 1)))
  (import "wasi:clocks/wall-clock@0.2.12" "resolution" (func $wall-clock.resolution (;9;) (type 1)))
  (import "wasi:random/random@0.2.12" "get-random-bytes" (func $random.get-random-bytes (;10;) (type 5)))
  (import "wasi:random/random@0.2.12" "get-random-u64" (func $random.get-random-u64 (;11;) (type 3)))
  (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func $error.error.to-debug-string (;12;) (type 6)))
  (import "wasi:io/error@0.2.12" "[resource-drop]error" (func $error.drop-error (;13;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.read" (func $streams.input-stream.read (;14;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-read" (func $streams.input-stream.blocking-read (;15;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.skip" (func $streams.input-stream.skip (;16;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-skip" (func $streams.input-stream.blocking-skip (;17;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.subscribe" (func $streams.input-stream.subscribe (;18;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.check-write" (func $streams.output-stream.check-write (;19;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write" (func $streams.output-stream.write (;20;) (type 8)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-and-flush" (func $streams.output-stream.blocking-write-and-flush (;21;) (type 8)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.flush" (func $streams.output-stream.flush (;22;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-flush" (func $streams.output-stream.blocking-flush (;23;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.subscribe" (func $streams.output-stream.subscribe (;24;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write-zeroes" (func $streams.output-stream.write-zeroes (;25;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-zeroes-and-flush" (func $streams.output-stream.blocking-write-zeroes-and-flush (;26;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.splice" (func $streams.output-stream.splice (;27;) (type 9)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-splice" (func $streams.output-stream.blocking-splice (;28;) (type 9)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]input-stream" (func $streams.drop-input-stream (;29;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]output-stream" (func $streams.drop-output-stream (;30;) (type 1)))
  (import "wasi:cli/stdout@0.2.12" "get-stdout" (func $stdout.get-stdout (;31;) (type 10)))
  (import "wasi:cli/stderr@0.2.12" "get-stderr" (func $stderr.get-stderr (;32;) (type 10)))
  (import "wasi:cli/stdin@0.2.12" "get-stdin" (func $stdin.get-stdin (;33;) (type 10)))
  (import "wasi:http/types@0.2.12" "[constructor]fields" (func $types.new-fields (;34;) (type 10)))
  (import "wasi:http/types@0.2.12" "[static]fields.from-list" (func $types.fields.from-list (;35;) (type 2)))
  (import "wasi:http/types@0.2.12" "[method]fields.get" (func $types.fields.get (;36;) (type 8)))
  (import "wasi:http/types@0.2.12" "[method]fields.has" (func $types.fields.has (;37;) (type 11)))
  (import "wasi:http/types@0.2.12" "[method]fields.set" (func $types.fields.set (;38;) (type 12)))
  (import "wasi:http/types@0.2.12" "[method]fields.delete" (func $types.fields.delete (;39;) (type 8)))
  (import "wasi:http/types@0.2.12" "[method]fields.append" (func $types.fields.append (;40;) (type 12)))
  (import "wasi:http/types@0.2.12" "[method]fields.entries" (func $types.fields.entries (;41;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]fields.clone" (func $types.fields.clone (;42;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.method" (func $types.incoming-request.method (;43;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.path-with-query" (func $types.incoming-request.path-with-query (;44;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.scheme" (func $types.incoming-request.scheme (;45;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.authority" (func $types.incoming-request.authority (;46;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.headers" (func $types.incoming-request.headers (;47;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.consume" (func $types.incoming-request.consume (;48;) (type 6)))
  (import "wasi:http/types@0.2.12" "[constructor]outgoing-request" (func $types.new-outgoing-request (;49;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.body" (func $types.outgoing-request.body (;50;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.method" (func $types.outgoing-request.method (;51;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-method" (func $types.outgoing-request.set-method (;52;) (type 13)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.path-with-query" (func $types.outgoing-request.path-with-query (;53;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-path-with-query" (func $types.outgoing-request.set-path-with-query (;54;) (type 13)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.scheme" (func $types.outgoing-request.scheme (;55;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-scheme" (func $types.outgoing-request.set-scheme (;56;) (type 14)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.authority" (func $types.outgoing-request.authority (;57;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-authority" (func $types.outgoing-request.set-authority (;58;) (type 13)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.headers" (func $types.outgoing-request.headers (;59;) (type 0)))
  (import "wasi:http/types@0.2.12" "[constructor]request-options" (func $types.new-request-options (;60;) (type 10)))
  (import "wasi:http/types@0.2.12" "[method]request-options.connect-timeout" (func $types.request-options.connect-timeout (;61;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-connect-timeout" (func $types.request-options.set-connect-timeout (;62;) (type 15)))
  (import "wasi:http/types@0.2.12" "[method]request-options.first-byte-timeout" (func $types.request-options.first-byte-timeout (;63;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-first-byte-timeout" (func $types.request-options.set-first-byte-timeout (;64;) (type 15)))
  (import "wasi:http/types@0.2.12" "[method]request-options.between-bytes-timeout" (func $types.request-options.between-bytes-timeout (;65;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-between-bytes-timeout" (func $types.request-options.set-between-bytes-timeout (;66;) (type 15)))
  (import "wasi:http/types@0.2.12" "[static]response-outparam.set" (func $types.response-outparam.set (;67;) (type 16)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.status" (func $types.incoming-response.status (;68;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.headers" (func $types.incoming-response.headers (;69;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.consume" (func $types.incoming-response.consume (;70;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]incoming-body.stream" (func $types.incoming-body.stream (;71;) (type 6)))
  (import "wasi:http/types@0.2.12" "[static]incoming-body.finish" (func $types.incoming-body.finish (;72;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-trailers.subscribe" (func $types.future-trailers.subscribe (;73;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-trailers.get" (func $types.future-trailers.get (;74;) (type 6)))
  (import "wasi:http/types@0.2.12" "[constructor]outgoing-response" (func $types.new-outgoing-response (;75;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.status-code" (func $types.outgoing-response.status-code (;76;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.set-status-code" (func $types.outgoing-response.set-status-code (;77;) (type 17)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.headers" (func $types.outgoing-response.headers (;78;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.body" (func $types.outgoing-response.body (;79;) (type 6)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-body.write" (func $types.outgoing-body.write (;80;) (type 6)))
  (import "wasi:http/types@0.2.12" "[static]outgoing-body.finish" (func $types.outgoing-body.finish (;81;) (type 8)))
  (import "wasi:http/types@0.2.12" "[method]future-incoming-response.subscribe" (func $types.future-incoming-response.subscribe (;82;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-incoming-response.get" (func $types.future-incoming-response.get (;83;) (type 6)))
  (import "wasi:http/types@0.2.12" "http-error-code" (func $types.http-error-code (;84;) (type 6)))
  (import "wasi:http/types@0.2.12" "[resource-drop]fields" (func $types.drop-fields (;85;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-request" (func $types.drop-incoming-request (;86;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-request" (func $types.drop-outgoing-request (;87;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]request-options" (func $types.drop-request-options (;88;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]response-outparam" (func $types.drop-response-outparam (;89;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-response" (func $types.drop-incoming-response (;90;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-body" (func $types.drop-incoming-body (;91;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]future-trailers" (func $types.drop-future-trailers (;92;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-response" (func $types.drop-outgoing-response (;93;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-body" (func $types.drop-outgoing-body (;94;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]future-incoming-response" (func $types.drop-future-incoming-response (;95;) (type 1)))
  (import "wasi:http/outgoing-handler@0.2.12" "handle" (func $outgoing-handler.handle (;96;) (type 8)))
  (memory (;0;) 64 128)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "x-wasmvm") (data (i32.const 1040) "value")
 (data (i32.const 1056) "/echo") (data (i32.const 1072) "x-request-trailer")
 (data (i32.const 1144) "trailer") (data (i32.const 1104) "done") (data (i32.const 1120) "x-reply-trailer")
 (data (i32.const 2048) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
  (export "probe" (func 97))
  (export "cabi_post_probe" (func 98))
  (export "request-batch" (func 99))
  (export "cabi_post_request-batch" (func 100))
  (export "wasi:http/incoming-handler@0.2.12#handle" (func 101))
  (export "cabi_post_wasi:http/incoming-handler@0.2.12#handle" (func 102))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 103))
  (export "_initialize" (func 104))
  (func (;97;) (type 11) (param i32 i32 i32) (result i32)
local.get 0 local.get 1 local.get 2 call $request
 )
  (func (;98;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;99;) (type 11) (param i32 i32 i32) (result i32) (local $at i32)
local.get 2 i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get 2 i32.ge_u br_if $end
 local.get 0 local.get 1 i32.const 0 call $request i32.const 42 i32.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at
 )
  (func (;100;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;101;) (type 6) (param i32 i32)
(local $headers i32) (local $body i32) (local $response i32) (local $output i32)
 local.get 0 i32.const 64 call $types.incoming-request.method
 local.get 0 i32.const 64 call $types.incoming-request.path-with-query
 local.get 0 i32.const 64 call $types.incoming-request.scheme
 local.get 0 i32.const 64 call $types.incoming-request.authority
 local.get 0 call $types.incoming-request.headers local.tee $headers call $types.drop-fields
 local.get 0 i32.const 64 call $types.incoming-request.consume i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get 0 call $types.drop-incoming-request
 local.get $body call $read-body i32.const 4096 i32.ne if unreachable end
 local.get $body call $types.incoming-body.finish i32.const 1072 i32.const 17 call $trailers
 call $types.new-fields call $types.new-outgoing-response local.set $response
 local.get $response i32.const 200 call $types.outgoing-response.set-status-code if unreachable end
 local.get $response call $types.outgoing-response.status-code i32.const 200 i32.ne if unreachable end
 local.get $response call $types.outgoing-response.headers local.tee $headers call $types.drop-fields
 local.get $response i32.const 64 call $types.outgoing-response.body i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get 1 i32.const 0 local.get $response i32.const 0 i64.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 call $types.response-outparam.set
 local.get $body i32.const 64 call $types.outgoing-body.write i32.const 64 call $ok i32.const 68 i32.load local.set $output
 local.get $output i32.const 2048 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 local.get $output call $streams.drop-output-stream
 local.get $body i32.const 1 i32.const 1120 i32.const 15 call $new-trailers i32.const 64 call $types.outgoing-body.finish i32.const 64 call $ok

 )
  (func (;102;) (type 18) i32.const 8192 global.set $heap)
  (func (;103;) (type 13) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;104;) (type 18))

 (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $trailers (param $future i32) (param $name i32) (param $len i32) (local $poll i32) (local $headers i32)
 block $ready loop $next
 local.get $future i32.const 64 call $types.future-trailers.get i32.const 64 i32.load8_u br_if $ready
 local.get $future call $types.future-trailers.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $next end end
 i32.const 72 i32.load8_u if unreachable end i32.const 80 i32.load8_u if unreachable end
 i32.const 88 i32.load8_u if
 i32.const 92 i32.load local.set $headers
 local.get $headers local.get $name local.get $len i32.const 64 call $types.fields.get
 i32.const 68 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 64 i32.load i32.const 4 i32.add i32.load i32.const 4 i32.ne if unreachable end
 local.get $headers call $types.drop-fields end
 local.get $future call $types.drop-future-trailers)
 (func $read-body (param $body i32) (result i32) (local $stream i32) (local $at i32) (local $n i32) (local $ptr i32)
 local.get $body i32.const 64 call $types.incoming-body.stream i32.const 64 call $ok i32.const 68 i32.load local.set $stream
 block $end loop $next
 local.get $stream i64.const 4096 i32.const 64 call $streams.input-stream.blocking-read
 i32.const 64 i32.load8_u if
 i32.const 68 i32.load8_u i32.const 1 i32.ne if unreachable end br $end end
 i32.const 68 i32.load local.set $ptr i32.const 72 i32.load local.tee $n i32.eqz if unreachable end
 local.get $ptr i32.load8_u i32.const 2048 local.get $at i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $ptr local.get $n i32.add i32.const 1 i32.sub i32.load8_u i32.const 2048 local.get $at local.get $n i32.add i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $at local.get $n i32.add local.tee $at i32.const 65536 i32.gt_u if unreachable end
 br $next end end
 local.get $stream call $streams.drop-input-stream local.get $at)
 (func $new-trailers (param $name i32) (param $len i32) (result i32) (local $fields i32)
 call $types.new-fields local.tee $fields local.get $name local.get $len i32.const 1104 i32.const 4 i32.const 64 call $types.fields.append i32.const 64 call $ok local.get $fields)
 (func $request (param $authority i32) (param $alen i32) (param $secure i32) (result i32)
 (local $headers i32) (local $request i32) (local $body i32) (local $stream i32) (local $options i32) (local $future i32) (local $poll i32) (local $response i32) (local $incoming i32)
 call $types.new-fields local.tee $headers i32.const 1024 i32.const 8 i32.const 1040 i32.const 5 i32.const 64 call $types.fields.append i32.const 64 call $ok
 local.get $headers i32.const 1144 i32.const 7 i32.const 1072 i32.const 17 i32.const 64 call $types.fields.append i32.const 64 call $ok
 local.get $headers call $types.new-outgoing-request local.set $request
 local.get $request i32.const 2 i32.const 0 i32.const 0 call $types.outgoing-request.set-method if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.method i32.const 64 i32.load8_u i32.const 2 i32.ne if unreachable end
 local.get $request i32.const 1 i32.const 1056 i32.const 5 call $types.outgoing-request.set-path-with-query if unreachable end
 local.get $request i32.const 1 local.get $secure i32.const 0 i32.const 0 call $types.outgoing-request.set-scheme if unreachable end
 local.get $request i32.const 1 local.get $authority local.get $alen call $types.outgoing-request.set-authority if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.path-with-query i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.scheme i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.authority i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request call $types.outgoing-request.headers local.set $headers
 local.get $headers i32.const 1024 i32.const 8 i32.const 1040 i32.const 5 i32.const 64 call $types.fields.append
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end i32.const 65 i32.load8_u i32.const 2 i32.ne if unreachable end
 local.get $headers call $types.drop-fields
 local.get $request i32.const 64 call $types.outgoing-request.body i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get $body i32.const 64 call $types.outgoing-body.write i32.const 64 call $ok i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 2048 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 local.get $stream call $streams.drop-output-stream
 local.get $body i32.const 1 i32.const 1072 i32.const 17 call $new-trailers i32.const 64 call $types.outgoing-body.finish i32.const 64 call $ok
 call $types.new-request-options local.set $options
 local.get $options i32.const 64 call $types.request-options.connect-timeout
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-connect-timeout if unreachable end
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-first-byte-timeout if unreachable end
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-between-bytes-timeout if unreachable end
 local.get $options i32.const 64 call $types.request-options.first-byte-timeout i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $options i32.const 64 call $types.request-options.between-bytes-timeout i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 1 local.get $options i32.const 64 call $outgoing-handler.handle i32.const 64 call $ok i32.const 72 i32.load local.set $future
 block $ready loop $next
 local.get $future i32.const 64 call $types.future-incoming-response.get i32.const 64 i32.load8_u br_if $ready
 local.get $future call $types.future-incoming-response.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $next end end
 i32.const 72 i32.load8_u if unreachable end i32.const 80 i32.load8_u if unreachable end i32.const 88 i32.load local.set $response
 local.get $future call $types.drop-future-incoming-response
 local.get $response call $types.incoming-response.status i32.const 200 i32.ne if unreachable end
 local.get $response call $types.incoming-response.headers local.tee $headers call $types.drop-fields
 local.get $response i32.const 64 call $types.incoming-response.consume i32.const 64 call $ok i32.const 68 i32.load local.set $incoming
 local.get $response call $types.drop-incoming-response
 local.get $incoming call $read-body i32.const 4096 i32.ne if unreachable end
 local.get $incoming call $types.incoming-body.finish i32.const 1120 i32.const 15 call $trailers
 i32.const 42)

  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\dc6\01A\02\01A'\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\00\02\03\00\00\08pollable\01B\0f\02\03\02\01\01\04\00\08pollable\03\00\00\01w\04\00\07instant\03\00\02\01w\04\00\08duration\03\00\04\01@\00\00\03\04\00\03now\01\06\01@\00\00\05\04\00\0aresolution\01\07\01i\01\01@\01\04when\03\00\08\04\00\11subscribe-instant\01\09\01@\01\04when\05\00\08\04\00\12subscribe-duration\01\0a\03\00\22wasi:clocks/monotonic-clock@0.2.12\05\02\01B\05\01r\02\07secondsw\0bnanosecondsy\04\00\08datetime\03\00\00\01@\00\00\01\04\00\03now\01\02\04\00\0aresolution\01\02\03\00\1dwasi:clocks/wall-clock@0.2.12\05\03\01B\05\01p}\01@\01\03lenw\00\00\04\00\10get-random-bytes\01\01\01@\00\00w\04\00\0eget-random-u64\01\02\03\00\19wasi:random/random@0.2.12\05\04\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\05\02\03\00\04\05error\01B(\02\03\02\01\06\04\00\05error\03\00\00\02\03\02\01\01\04\00\08pollable\03\00\02\01i\01\01q\02\15last-operation-failed\01\04\00\06closed\00\00\04\00\0cstream-error\03\00\05\04\00\0cinput-stream\03\01\04\00\0doutput-stream\03\01\01h\07\01p}\01j\01\0a\01\06\01@\02\04self\09\03lenw\00\0b\04\00\19[method]input-stream.read\01\0c\04\00\22[method]input-stream.blocking-read\01\0c\01j\01w\01\06\01@\02\04self\09\03lenw\00\0d\04\00\19[method]input-stream.skip\01\0e\04\00\22[method]input-stream.blocking-skip\01\0e\01i\03\01@\01\04self\09\00\0f\04\00\1e[method]input-stream.subscribe\01\10\01h\08\01@\01\04self\11\00\0d\04\00![method]output-stream.check-write\01\12\01j\00\01\06\01@\02\04self\11\08contents\0a\00\13\04\00\1b[method]output-stream.write\01\14\04\00.[method]output-stream.blocking-write-and-flush\01\14\01@\01\04self\11\00\13\04\00\1b[method]output-stream.flush\01\15\04\00$[method]output-stream.blocking-flush\01\15\01@\01\04self\11\00\0f\04\00\1f[method]output-stream.subscribe\01\16\01@\02\04self\11\03lenw\00\13\04\00\22[method]output-stream.write-zeroes\01\17\04\005[method]output-stream.blocking-write-zeroes-and-flush\01\17\01@\03\04self\11\03src\09\03lenw\00\0d\04\00\1c[method]output-stream.splice\01\18\04\00%[method]output-stream.blocking-splice\01\18\03\00\16wasi:io/streams@0.2.12\05\07\02\03\00\05\0doutput-stream\01B\05\02\03\02\01\08\04\00\0doutput-stream\03\00\00\01i\01\01@\00\00\02\04\00\0aget-stdout\01\03\03\00\16wasi:cli/stdout@0.2.12\05\09\01B\05\02\03\02\01\08\04\00\0doutput-stream\03\00\00\01i\01\01@\00\00\02\04\00\0aget-stderr\01\03\03\00\16wasi:cli/stderr@0.2.12\05\0a\02\03\00\05\0cinput-stream\01B\05\02\03\02\01\0b\04\00\0cinput-stream\03\00\00\01i\01\01@\00\00\02\04\00\09get-stdin\01\03\03\00\15wasi:cli/stdin@0.2.12\05\0c\02\03\00\01\08duration\01B\c1\01\02\03\02\01\0d\04\00\08duration\03\00\00\02\03\02\01\0b\04\00\0cinput-stream\03\00\02\02\03\02\01\08\04\00\0doutput-stream\03\00\04\02\03\02\01\06\04\00\08io-error\03\00\06\02\03\02\01\01\04\00\08pollable\03\00\08\01q\0a\03get\00\00\04head\00\00\04post\00\00\03put\00\00\06delete\00\00\07connect\00\00\07options\00\00\05trace\00\00\05patch\00\00\05other\01s\00\04\00\06method\03\00\0a\01q\03\04HTTP\00\00\05HTTPS\00\00\05other\01s\00\04\00\06scheme\03\00\0c\01ks\01k{\01r\02\05rcode\0e\09info-code\0f\04\00\11DNS-error-payload\03\00\10\01k}\01r\02\08alert-id\12\0dalert-message\0e\04\00\1aTLS-alert-received-payload\03\00\13\01ky\01r\02\0afield-name\0e\0afield-size\15\04\00\12field-size-payload\03\00\16\01kw\01k\17\01q'\0bDNS-timeout\00\00\09DNS-error\01\11\00\15destination-not-found\00\00\17destination-unavailable\00\00\19destination-IP-prohibited\00\00\19destination-IP-unroutable\00\00\12connection-refused\00\00\15connection-terminated\00\00\12connection-timeout\00\00\17connection-read-timeout\00\00\18connection-write-timeout\00\00\18connection-limit-reached\00\00\12TLS-protocol-error\00\00\15TLS-certificate-error\00\00\12TLS-alert-received\01\14\00\13HTTP-request-denied\00\00\1cHTTP-request-length-required\00\00\16HTTP-request-body-size\01\18\00\1bHTTP-request-method-invalid\00\00\18HTTP-request-URI-invalid\00\00\19HTTP-request-URI-too-long\00\00 HTTP-request-header-section-size\01\15\00\18HTTP-request-header-size\01\19\00!HTTP-request-trailer-section-size\01\15\00\19HTTP-request-trailer-size\01\17\00\18HTTP-response-incomplete\00\00!HTTP-response-header-section-size\01\15\00\19HTTP-response-header-size\01\17\00\17HTTP-response-body-size\01\18\00\22HTTP-response-trailer-section-size\01\15\00\1aHTTP-response-trailer-size\01\17\00\1dHTTP-response-transfer-coding\01\0e\00\1cHTTP-response-content-coding\01\0e\00\15HTTP-response-timeout\00\00\13HTTP-upgrade-failed\00\00\13HTTP-protocol-error\00\00\0dloop-detected\00\00\13configuration-error\00\00\0einternal-error\01\0e\00\04\00\0aerror-code\03\00\1a\01q\03\0einvalid-syntax\00\00\09forbidden\00\00\09immutable\00\00\04\00\0cheader-error\03\00\1c\01s\04\00\09field-key\03\00\1e\04\00\0afield-name\03\00\1f\01p}\04\00\0bfield-value\03\00!\04\00\06fields\03\01\04\00\07headers\03\00#\04\00\08trailers\03\00#\04\00\10incoming-request\03\01\04\00\10outgoing-request\03\01\04\00\0frequest-options\03\01\04\00\11response-outparam\03\01\01{\04\00\0bstatus-code\03\00*\04\00\11incoming-response\03\01\04\00\0dincoming-body\03\01\04\00\0ffuture-trailers\03\01\04\00\11outgoing-response\03\01\04\00\0doutgoing-body\03\01\04\00\18future-incoming-response\03\01\01i#\01@\00\002\04\00\13[constructor]fields\013\01o\02 \22\01p4\01j\012\01\1d\01@\01\07entries5\006\04\00\18[static]fields.from-list\017\01h#\01p\22\01@\02\04self8\04name \009\04\00\12[method]fields.get\01:\01@\02\04self8\04name \00\7f\04\00\12[method]fields.has\01;\01j\00\01\1d\01@\03\04self8\04name \05value9\00<\04\00\12[method]fields.set\01=\01@\02\04self8\04name \00<\04\00\15[method]fields.delete\01>\01@\03\04self8\04name \05value\22\00<\04\00\15[method]fields.append\01?\01@\01\04self8\005\04\00\16[method]fields.entries\01@\01@\01\04self8\002\04\00\14[method]fields.clone\01A\01h&\01@\01\04self\c2\00\00\0b\04\00\1f[method]incoming-request.method\01C\01@\01\04self\c2\00\00\0e\04\00([method]incoming-request.path-with-query\01D\01k\0d\01@\01\04self\c2\00\00\c5\00\04\00\1f[method]incoming-request.scheme\01F\04\00\22[method]incoming-request.authority\01D\01i$\01@\01\04self\c2\00\00\c7\00\04\00 [method]incoming-request.headers\01H\01i-\01j\01\c9\00\00\01@\01\04self\c2\00\00\ca\00\04\00 [method]incoming-request.consume\01K\01i'\01@\01\07headers\c7\00\00\cc\00\04\00\1d[constructor]outgoing-request\01M\01h'\01i0\01j\01\cf\00\00\01@\01\04self\ce\00\00\d0\00\04\00\1d[method]outgoing-request.body\01Q\01@\01\04self\ce\00\00\0b\04\00\1f[method]outgoing-request.method\01R\01j\00\00\01@\02\04self\ce\00\06method\0b\00\d3\00\04\00#[method]outgoing-request.set-method\01T\01@\01\04self\ce\00\00\0e\04\00([method]outgoing-request.path-with-query\01U\01@\02\04self\ce\00\0fpath-with-query\0e\00\d3\00\04\00,[method]outgoing-request.set-path-with-query\01V\01@\01\04self\ce\00\00\c5\00\04\00\1f[method]outgoing-request.scheme\01W\01@\02\04self\ce\00\06scheme\c5\00\00\d3\00\04\00#[method]outgoing-request.set-scheme\01X\04\00\22[method]outgoing-request.authority\01U\01@\02\04self\ce\00\09authority\0e\00\d3\00\04\00&[method]outgoing-request.set-authority\01Y\01@\01\04self\ce\00\00\c7\00\04\00 [method]outgoing-request.headers\01Z\01i(\01@\00\00\db\00\04\00\1c[constructor]request-options\01\5c\01h(\01k\01\01@\01\04self\dd\00\00\de\00\04\00'[method]request-options.connect-timeout\01_\01@\02\04self\dd\00\08duration\de\00\00\d3\00\04\00+[method]request-options.set-connect-timeout\01`\04\00*[method]request-options.first-byte-timeout\01_\04\00.[method]request-options.set-first-byte-timeout\01`\04\00-[method]request-options.between-bytes-timeout\01_\04\001[method]request-options.set-between-bytes-timeout\01`\01i)\01i/\01j\01\e2\00\01\1b\01@\02\05param\e1\00\08response\e3\00\01\00\04\00\1d[static]response-outparam.set\01d\01h,\01@\01\04self\e5\00\00+\04\00 [method]incoming-response.status\01f\01@\01\04self\e5\00\00\c7\00\04\00![method]incoming-response.headers\01g\01@\01\04self\e5\00\00\ca\00\04\00![method]incoming-response.consume\01h\01h-\01i\03\01j\01\ea\00\00\01@\01\04self\e9\00\00\eb\00\04\00\1c[method]incoming-body.stream\01l\01i.\01@\01\04this\c9\00\00\ed\00\04\00\1c[static]incoming-body.finish\01n\01h.\01i\09\01@\01\04self\ef\00\00\f0\00\04\00![method]future-trailers.subscribe\01q\01i%\01k\f2\00\01j\01\f3\00\01\1b\01j\01\f4\00\00\01k\f5\00\01@\01\04self\ef\00\00\f6\00\04\00\1b[method]future-trailers.get\01w\01@\01\07headers\c7\00\00\e2\00\04\00\1e[constructor]outgoing-response\01x\01h/\01@\01\04self\f9\00\00+\04\00%[method]outgoing-response.status-code\01z\01@\02\04self\f9\00\0bstatus-code+\00\d3\00\04\00)[method]outgoing-response.set-status-code\01{\01@\01\04self\f9\00\00\c7\00\04\00![method]outgoing-response.headers\01|\01@\01\04self\f9\00\00\d0\00\04\00\1e[method]outgoing-response.body\01}\01h0\01i\05\01j\01\ff\00\00\01@\01\04self\fe\00\00\80\01\04\00\1b[method]outgoing-body.write\01\81\01\01j\00\01\1b\01@\02\04this\cf\00\08trailers\f3\00\00\82\01\04\00\1c[static]outgoing-body.finish\01\83\01\01h1\01@\01\04self\84\01\00\f0\00\04\00*[method]future-incoming-response.subscribe\01\85\01\01i,\01j\01\86\01\01\1b\01j\01\87\01\00\01k\88\01\01@\01\04self\84\01\00\89\01\04\00$[method]future-incoming-response.get\01\8a\01\01h\07\01k\1b\01@\01\03err\8b\01\00\8c\01\04\00\0fhttp-error-code\01\8d\01\03\00\16wasi:http/types@0.2.12\05\0e\02\03\00\09\10outgoing-request\02\03\00\09\0frequest-options\02\03\00\09\18future-incoming-response\02\03\00\09\0aerror-code\01B\0f\02\03\02\01\0f\04\00\10outgoing-request\03\00\00\02\03\02\01\10\04\00\0frequest-options\03\00\02\02\03\02\01\11\04\00\18future-incoming-response\03\00\04\02\03\02\01\12\04\00\0aerror-code\03\00\06\01i\01\01i\03\01k\09\01i\05\01j\01\0b\01\07\01@\02\07request\08\07options\0a\00\0c\04\00\06handle\01\0d\03\00!wasi:http/outgoing-handler@0.2.12\05\13\01@\02\09authoritys\06secure\7f\00y\04\00\05probe\01\14\01@\02\09authoritys\05county\00y\04\00\0drequest-batch\01\15\02\03\00\09\10incoming-request\02\03\00\09\11response-outparam\01B\08\02\03\02\01\16\04\00\10incoming-request\03\00\00\02\03\02\01\17\04\00\11response-outparam\03\00\02\01i\01\01i\03\01@\02\07request\04\0cresponse-out\05\01\00\04\00\06handle\01\06\04\00!wasi:http/incoming-handler@0.2.12\05\18\04\00\16wasmvm:http-tests/http\04\00\0b\0a\01\00\04http\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

(module
  (type (;0;) (func (param i32) (result i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32 i32 i32)))
  (type (;3;) (func (result i64)))
  (type (;4;) (func (param i64) (result i32)))
  (type (;5;) (func (param i32 i32)))
  (type (;6;) (func (param i32 i64 i32)))
  (type (;7;) (func (param i32 i32 i32 i32)))
  (type (;8;) (func (param i32 i32 i64 i32)))
  (type (;9;) (func (result i32)))
  (type (;10;) (func (param i32 i32 i32) (result i32)))
  (type (;11;) (func (param i32 i32 i32 i32 i32 i32)))
  (type (;12;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;13;) (func (param i32 i32 i32 i32 i32) (result i32)))
  (type (;14;) (func (param i32 i32 i64) (result i32)))
  (type (;15;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
  (type (;16;) (func (param i32 i32) (result i32)))
  (type (;17;) (func))
  (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func $pollable.ready (;0;) (type 0)))
  (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func $pollable.block (;1;) (type 1)))
  (import "wasi:io/poll@0.2.12" "poll" (func $poll (;2;) (type 2)))
  (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func $drop-pollable (;3;) (type 1)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "now" (func $now (;4;) (type 3)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "resolution" (func $resolution (;5;) (type 3)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-instant" (func $subscribe-instant (;6;) (type 4)))
  (import "wasi:clocks/monotonic-clock@0.2.12" "subscribe-duration" (func $subscribe-duration (;7;) (type 4)))
  (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func $error.to-debug-string (;8;) (type 5)))
  (import "wasi:io/error@0.2.12" "[resource-drop]error" (func $drop-error (;9;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.read" (func $input-stream.read (;10;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-read" (func $input-stream.blocking-read (;11;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.skip" (func $input-stream.skip (;12;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-skip" (func $input-stream.blocking-skip (;13;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]input-stream.subscribe" (func $input-stream.subscribe (;14;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.check-write" (func $output-stream.check-write (;15;) (type 5)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write" (func $output-stream.write (;16;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-and-flush" (func $output-stream.blocking-write-and-flush (;17;) (type 7)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.flush" (func $output-stream.flush (;18;) (type 5)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-flush" (func $output-stream.blocking-flush (;19;) (type 5)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.subscribe" (func $output-stream.subscribe (;20;) (type 0)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.write-zeroes" (func $output-stream.write-zeroes (;21;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-zeroes-and-flush" (func $output-stream.blocking-write-zeroes-and-flush (;22;) (type 6)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.splice" (func $output-stream.splice (;23;) (type 8)))
  (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-splice" (func $output-stream.blocking-splice (;24;) (type 8)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]input-stream" (func $drop-input-stream (;25;) (type 1)))
  (import "wasi:io/streams@0.2.12" "[resource-drop]output-stream" (func $drop-output-stream (;26;) (type 1)))
  (import "wasi:http/types@0.2.12" "[constructor]fields" (func $new-fields (;27;) (type 9)))
  (import "wasi:http/types@0.2.12" "[static]fields.from-list" (func $fields.from-list (;28;) (type 2)))
  (import "wasi:http/types@0.2.12" "[method]fields.get" (func $fields.get (;29;) (type 7)))
  (import "wasi:http/types@0.2.12" "[method]fields.has" (func $fields.has (;30;) (type 10)))
  (import "wasi:http/types@0.2.12" "[method]fields.set" (func $fields.set (;31;) (type 11)))
  (import "wasi:http/types@0.2.12" "[method]fields.delete" (func $fields.delete (;32;) (type 7)))
  (import "wasi:http/types@0.2.12" "[method]fields.append" (func $fields.append (;33;) (type 11)))
  (import "wasi:http/types@0.2.12" "[method]fields.entries" (func $fields.entries (;34;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]fields.clone" (func $fields.clone (;35;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.method" (func $incoming-request.method (;36;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.path-with-query" (func $incoming-request.path-with-query (;37;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.scheme" (func $incoming-request.scheme (;38;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.authority" (func $incoming-request.authority (;39;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.headers" (func $incoming-request.headers (;40;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-request.consume" (func $incoming-request.consume (;41;) (type 5)))
  (import "wasi:http/types@0.2.12" "[constructor]outgoing-request" (func $new-outgoing-request (;42;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.body" (func $outgoing-request.body (;43;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.method" (func $outgoing-request.method (;44;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-method" (func $outgoing-request.set-method (;45;) (type 12)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.path-with-query" (func $outgoing-request.path-with-query (;46;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-path-with-query" (func $outgoing-request.set-path-with-query (;47;) (type 12)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.scheme" (func $outgoing-request.scheme (;48;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-scheme" (func $outgoing-request.set-scheme (;49;) (type 13)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.authority" (func $outgoing-request.authority (;50;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.set-authority" (func $outgoing-request.set-authority (;51;) (type 12)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-request.headers" (func $outgoing-request.headers (;52;) (type 0)))
  (import "wasi:http/types@0.2.12" "[constructor]request-options" (func $new-request-options (;53;) (type 9)))
  (import "wasi:http/types@0.2.12" "[method]request-options.connect-timeout" (func $request-options.connect-timeout (;54;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-connect-timeout" (func $request-options.set-connect-timeout (;55;) (type 14)))
  (import "wasi:http/types@0.2.12" "[method]request-options.first-byte-timeout" (func $request-options.first-byte-timeout (;56;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-first-byte-timeout" (func $request-options.set-first-byte-timeout (;57;) (type 14)))
  (import "wasi:http/types@0.2.12" "[method]request-options.between-bytes-timeout" (func $request-options.between-bytes-timeout (;58;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]request-options.set-between-bytes-timeout" (func $request-options.set-between-bytes-timeout (;59;) (type 14)))
  (import "wasi:http/types@0.2.12" "[static]response-outparam.set" (func $response-outparam.set (;60;) (type 15)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.status" (func $incoming-response.status (;61;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.headers" (func $incoming-response.headers (;62;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]incoming-response.consume" (func $incoming-response.consume (;63;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]incoming-body.stream" (func $incoming-body.stream (;64;) (type 5)))
  (import "wasi:http/types@0.2.12" "[static]incoming-body.finish" (func $incoming-body.finish (;65;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-trailers.subscribe" (func $future-trailers.subscribe (;66;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-trailers.get" (func $future-trailers.get (;67;) (type 5)))
  (import "wasi:http/types@0.2.12" "[constructor]outgoing-response" (func $new-outgoing-response (;68;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.status-code" (func $outgoing-response.status-code (;69;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.set-status-code" (func $outgoing-response.set-status-code (;70;) (type 16)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.headers" (func $outgoing-response.headers (;71;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-response.body" (func $outgoing-response.body (;72;) (type 5)))
  (import "wasi:http/types@0.2.12" "[method]outgoing-body.write" (func $outgoing-body.write (;73;) (type 5)))
  (import "wasi:http/types@0.2.12" "[static]outgoing-body.finish" (func $outgoing-body.finish (;74;) (type 7)))
  (import "wasi:http/types@0.2.12" "[method]future-incoming-response.subscribe" (func $future-incoming-response.subscribe (;75;) (type 0)))
  (import "wasi:http/types@0.2.12" "[method]future-incoming-response.get" (func $future-incoming-response.get (;76;) (type 5)))
  (import "wasi:http/types@0.2.12" "http-error-code" (func $http-error-code (;77;) (type 5)))
  (import "wasi:http/types@0.2.12" "[resource-drop]fields" (func $drop-fields (;78;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-request" (func $drop-incoming-request (;79;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-request" (func $drop-outgoing-request (;80;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]request-options" (func $drop-request-options (;81;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]response-outparam" (func $drop-response-outparam (;82;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-response" (func $drop-incoming-response (;83;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]incoming-body" (func $drop-incoming-body (;84;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]future-trailers" (func $drop-future-trailers (;85;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-response" (func $drop-outgoing-response (;86;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]outgoing-body" (func $drop-outgoing-body (;87;) (type 1)))
  (import "wasi:http/types@0.2.12" "[resource-drop]future-incoming-response" (func $drop-future-incoming-response (;88;) (type 1)))
  (memory (;0;) 1 4)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "x-key") (data (i32.const 1032) "one")
 (data (i32.const 1040) "X-KEY") (data (i32.const 1056) "bad name")

  (export "probe" (func 89))
  (export "cabi_post_probe" (func 90))
  (export "fields-batch" (func 91))
  (export "cabi_post_fields-batch" (func 92))
  (export "memory" (memory 0))
  (export "cabi_realloc" (func 93))
  (export "_initialize" (func 94))
  (func (;89;) (type 9) (result i32)
call $probe
 )
  (func (;90;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;91;) (type 0) (param i32) (result i32) (local $at i32)
local.get 0 i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get 0 i32.ge_u br_if $end
 call $probe i32.const 42 i32.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at
 )
  (func (;92;) (type 1) (param i32) i32.const 8192 global.set $heap)
  (func (;93;) (type 12) (param i32 i32 i32 i32) (result i32) (local $ptr i32)
global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr
 )
  (func (;94;) (type 17))
(func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $probe (result i32) (local $h i32) (local $copy i32) (local $other i32) (local $list i32)
 call $new-fields local.set $h
 local.get $h i32.const 1024 i32.const 5 i32.const 1032 i32.const 3 i32.const 64 call $fields.append i32.const 64 call $ok
 i32.const 256 i32.const 1032 i32.store i32.const 260 i32.const 3 i32.store
 i32.const 264 i32.const 0 i32.store i32.const 268 i32.const 0 i32.store
 local.get $h i32.const 1024 i32.const 5 i32.const 256 i32.const 2 i32.const 64 call $fields.set i32.const 64 call $ok
 local.get $h i32.const 1040 i32.const 5 call $fields.has i32.eqz if unreachable end
 local.get $h i32.const 1040 i32.const 5 i32.const 64 call $fields.get
 i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
 i32.const 64 i32.load local.tee $list i32.const 4 i32.add i32.load i32.const 3 i32.ne if unreachable end
 local.get $list i32.const 12 i32.add i32.load if unreachable end
 local.get $h i32.const 64 call $fields.entries
 i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
 local.get $h call $fields.clone local.set $copy
 local.get $copy i32.const 1024 i32.const 5 i32.const 64 call $fields.delete i32.const 64 call $ok
 local.get $copy i32.const 1040 i32.const 5 call $fields.has if unreachable end
 i32.const 288 i32.const 1024 i32.store i32.const 292 i32.const 5 i32.store
 i32.const 296 i32.const 1032 i32.store i32.const 300 i32.const 3 i32.store
 i32.const 288 i32.const 1 i32.const 64 call $fields.from-list i32.const 64 call $ok
 i32.const 68 i32.load local.set $other
 local.get $other i32.const 1056 i32.const 8 i32.const 1032 i32.const 3 i32.const 64 call $fields.append
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 65 i32.load8_u if unreachable end
 local.get $h call $drop-fields local.get $copy call $drop-fields local.get $other call $drop-fields
 i32.const 42)

  (@custom "component-type" (after code) "\00asm\0d\00\01\00\00\19\16wit-component-encoding\04\00\07\cc/\01A\02\01A\13\01B\0a\04\00\08pollable\03\01\01h\00\01@\01\04self\01\00\7f\04\00\16[method]pollable.ready\01\02\01@\01\04self\01\01\00\04\00\16[method]pollable.block\01\03\01p\01\01py\01@\01\02in\04\00\05\04\00\04poll\01\06\03\00\13wasi:io/poll@0.2.12\05\00\02\03\00\00\08pollable\01B\0f\02\03\02\01\01\04\00\08pollable\03\00\00\01w\04\00\07instant\03\00\02\01w\04\00\08duration\03\00\04\01@\00\00\03\04\00\03now\01\06\01@\00\00\05\04\00\0aresolution\01\07\01i\01\01@\01\04when\03\00\08\04\00\11subscribe-instant\01\09\01@\01\04when\05\00\08\04\00\12subscribe-duration\01\0a\03\00\22wasi:clocks/monotonic-clock@0.2.12\05\02\01B\04\04\00\05error\03\01\01h\00\01@\01\04self\01\00s\04\00\1d[method]error.to-debug-string\01\02\03\00\14wasi:io/error@0.2.12\05\03\02\03\00\02\05error\01B(\02\03\02\01\04\04\00\05error\03\00\00\02\03\02\01\01\04\00\08pollable\03\00\02\01i\01\01q\02\15last-operation-failed\01\04\00\06closed\00\00\04\00\0cstream-error\03\00\05\04\00\0cinput-stream\03\01\04\00\0doutput-stream\03\01\01h\07\01p}\01j\01\0a\01\06\01@\02\04self\09\03lenw\00\0b\04\00\19[method]input-stream.read\01\0c\04\00\22[method]input-stream.blocking-read\01\0c\01j\01w\01\06\01@\02\04self\09\03lenw\00\0d\04\00\19[method]input-stream.skip\01\0e\04\00\22[method]input-stream.blocking-skip\01\0e\01i\03\01@\01\04self\09\00\0f\04\00\1e[method]input-stream.subscribe\01\10\01h\08\01@\01\04self\11\00\0d\04\00![method]output-stream.check-write\01\12\01j\00\01\06\01@\02\04self\11\08contents\0a\00\13\04\00\1b[method]output-stream.write\01\14\04\00.[method]output-stream.blocking-write-and-flush\01\14\01@\01\04self\11\00\13\04\00\1b[method]output-stream.flush\01\15\04\00$[method]output-stream.blocking-flush\01\15\01@\01\04self\11\00\0f\04\00\1f[method]output-stream.subscribe\01\16\01@\02\04self\11\03lenw\00\13\04\00\22[method]output-stream.write-zeroes\01\17\04\005[method]output-stream.blocking-write-zeroes-and-flush\01\17\01@\03\04self\11\03src\09\03lenw\00\0d\04\00\1c[method]output-stream.splice\01\18\04\00%[method]output-stream.blocking-splice\01\18\03\00\16wasi:io/streams@0.2.12\05\05\02\03\00\01\08duration\02\03\00\03\0cinput-stream\02\03\00\03\0doutput-stream\01B\c1\01\02\03\02\01\06\04\00\08duration\03\00\00\02\03\02\01\07\04\00\0cinput-stream\03\00\02\02\03\02\01\08\04\00\0doutput-stream\03\00\04\02\03\02\01\04\04\00\08io-error\03\00\06\02\03\02\01\01\04\00\08pollable\03\00\08\01q\0a\03get\00\00\04head\00\00\04post\00\00\03put\00\00\06delete\00\00\07connect\00\00\07options\00\00\05trace\00\00\05patch\00\00\05other\01s\00\04\00\06method\03\00\0a\01q\03\04HTTP\00\00\05HTTPS\00\00\05other\01s\00\04\00\06scheme\03\00\0c\01ks\01k{\01r\02\05rcode\0e\09info-code\0f\04\00\11DNS-error-payload\03\00\10\01k}\01r\02\08alert-id\12\0dalert-message\0e\04\00\1aTLS-alert-received-payload\03\00\13\01ky\01r\02\0afield-name\0e\0afield-size\15\04\00\12field-size-payload\03\00\16\01kw\01k\17\01q'\0bDNS-timeout\00\00\09DNS-error\01\11\00\15destination-not-found\00\00\17destination-unavailable\00\00\19destination-IP-prohibited\00\00\19destination-IP-unroutable\00\00\12connection-refused\00\00\15connection-terminated\00\00\12connection-timeout\00\00\17connection-read-timeout\00\00\18connection-write-timeout\00\00\18connection-limit-reached\00\00\12TLS-protocol-error\00\00\15TLS-certificate-error\00\00\12TLS-alert-received\01\14\00\13HTTP-request-denied\00\00\1cHTTP-request-length-required\00\00\16HTTP-request-body-size\01\18\00\1bHTTP-request-method-invalid\00\00\18HTTP-request-URI-invalid\00\00\19HTTP-request-URI-too-long\00\00 HTTP-request-header-section-size\01\15\00\18HTTP-request-header-size\01\19\00!HTTP-request-trailer-section-size\01\15\00\19HTTP-request-trailer-size\01\17\00\18HTTP-response-incomplete\00\00!HTTP-response-header-section-size\01\15\00\19HTTP-response-header-size\01\17\00\17HTTP-response-body-size\01\18\00\22HTTP-response-trailer-section-size\01\15\00\1aHTTP-response-trailer-size\01\17\00\1dHTTP-response-transfer-coding\01\0e\00\1cHTTP-response-content-coding\01\0e\00\15HTTP-response-timeout\00\00\13HTTP-upgrade-failed\00\00\13HTTP-protocol-error\00\00\0dloop-detected\00\00\13configuration-error\00\00\0einternal-error\01\0e\00\04\00\0aerror-code\03\00\1a\01q\03\0einvalid-syntax\00\00\09forbidden\00\00\09immutable\00\00\04\00\0cheader-error\03\00\1c\01s\04\00\09field-key\03\00\1e\04\00\0afield-name\03\00\1f\01p}\04\00\0bfield-value\03\00!\04\00\06fields\03\01\04\00\07headers\03\00#\04\00\08trailers\03\00#\04\00\10incoming-request\03\01\04\00\10outgoing-request\03\01\04\00\0frequest-options\03\01\04\00\11response-outparam\03\01\01{\04\00\0bstatus-code\03\00*\04\00\11incoming-response\03\01\04\00\0dincoming-body\03\01\04\00\0ffuture-trailers\03\01\04\00\11outgoing-response\03\01\04\00\0doutgoing-body\03\01\04\00\18future-incoming-response\03\01\01i#\01@\00\002\04\00\13[constructor]fields\013\01o\02 \22\01p4\01j\012\01\1d\01@\01\07entries5\006\04\00\18[static]fields.from-list\017\01h#\01p\22\01@\02\04self8\04name \009\04\00\12[method]fields.get\01:\01@\02\04self8\04name \00\7f\04\00\12[method]fields.has\01;\01j\00\01\1d\01@\03\04self8\04name \05value9\00<\04\00\12[method]fields.set\01=\01@\02\04self8\04name \00<\04\00\15[method]fields.delete\01>\01@\03\04self8\04name \05value\22\00<\04\00\15[method]fields.append\01?\01@\01\04self8\005\04\00\16[method]fields.entries\01@\01@\01\04self8\002\04\00\14[method]fields.clone\01A\01h&\01@\01\04self\c2\00\00\0b\04\00\1f[method]incoming-request.method\01C\01@\01\04self\c2\00\00\0e\04\00([method]incoming-request.path-with-query\01D\01k\0d\01@\01\04self\c2\00\00\c5\00\04\00\1f[method]incoming-request.scheme\01F\04\00\22[method]incoming-request.authority\01D\01i$\01@\01\04self\c2\00\00\c7\00\04\00 [method]incoming-request.headers\01H\01i-\01j\01\c9\00\00\01@\01\04self\c2\00\00\ca\00\04\00 [method]incoming-request.consume\01K\01i'\01@\01\07headers\c7\00\00\cc\00\04\00\1d[constructor]outgoing-request\01M\01h'\01i0\01j\01\cf\00\00\01@\01\04self\ce\00\00\d0\00\04\00\1d[method]outgoing-request.body\01Q\01@\01\04self\ce\00\00\0b\04\00\1f[method]outgoing-request.method\01R\01j\00\00\01@\02\04self\ce\00\06method\0b\00\d3\00\04\00#[method]outgoing-request.set-method\01T\01@\01\04self\ce\00\00\0e\04\00([method]outgoing-request.path-with-query\01U\01@\02\04self\ce\00\0fpath-with-query\0e\00\d3\00\04\00,[method]outgoing-request.set-path-with-query\01V\01@\01\04self\ce\00\00\c5\00\04\00\1f[method]outgoing-request.scheme\01W\01@\02\04self\ce\00\06scheme\c5\00\00\d3\00\04\00#[method]outgoing-request.set-scheme\01X\04\00\22[method]outgoing-request.authority\01U\01@\02\04self\ce\00\09authority\0e\00\d3\00\04\00&[method]outgoing-request.set-authority\01Y\01@\01\04self\ce\00\00\c7\00\04\00 [method]outgoing-request.headers\01Z\01i(\01@\00\00\db\00\04\00\1c[constructor]request-options\01\5c\01h(\01k\01\01@\01\04self\dd\00\00\de\00\04\00'[method]request-options.connect-timeout\01_\01@\02\04self\dd\00\08duration\de\00\00\d3\00\04\00+[method]request-options.set-connect-timeout\01`\04\00*[method]request-options.first-byte-timeout\01_\04\00.[method]request-options.set-first-byte-timeout\01`\04\00-[method]request-options.between-bytes-timeout\01_\04\001[method]request-options.set-between-bytes-timeout\01`\01i)\01i/\01j\01\e2\00\01\1b\01@\02\05param\e1\00\08response\e3\00\01\00\04\00\1d[static]response-outparam.set\01d\01h,\01@\01\04self\e5\00\00+\04\00 [method]incoming-response.status\01f\01@\01\04self\e5\00\00\c7\00\04\00![method]incoming-response.headers\01g\01@\01\04self\e5\00\00\ca\00\04\00![method]incoming-response.consume\01h\01h-\01i\03\01j\01\ea\00\00\01@\01\04self\e9\00\00\eb\00\04\00\1c[method]incoming-body.stream\01l\01i.\01@\01\04this\c9\00\00\ed\00\04\00\1c[static]incoming-body.finish\01n\01h.\01i\09\01@\01\04self\ef\00\00\f0\00\04\00![method]future-trailers.subscribe\01q\01i%\01k\f2\00\01j\01\f3\00\01\1b\01j\01\f4\00\00\01k\f5\00\01@\01\04self\ef\00\00\f6\00\04\00\1b[method]future-trailers.get\01w\01@\01\07headers\c7\00\00\e2\00\04\00\1e[constructor]outgoing-response\01x\01h/\01@\01\04self\f9\00\00+\04\00%[method]outgoing-response.status-code\01z\01@\02\04self\f9\00\0bstatus-code+\00\d3\00\04\00)[method]outgoing-response.set-status-code\01{\01@\01\04self\f9\00\00\c7\00\04\00![method]outgoing-response.headers\01|\01@\01\04self\f9\00\00\d0\00\04\00\1e[method]outgoing-response.body\01}\01h0\01i\05\01j\01\ff\00\00\01@\01\04self\fe\00\00\80\01\04\00\1b[method]outgoing-body.write\01\81\01\01j\00\01\1b\01@\02\04this\cf\00\08trailers\f3\00\00\82\01\04\00\1c[static]outgoing-body.finish\01\83\01\01h1\01@\01\04self\84\01\00\f0\00\04\00*[method]future-incoming-response.subscribe\01\85\01\01i,\01j\01\86\01\01\1b\01j\01\87\01\00\01k\88\01\01@\01\04self\84\01\00\89\01\04\00$[method]future-incoming-response.get\01\8a\01\01h\07\01k\1b\01@\01\03err\8b\01\00\8c\01\04\00\0fhttp-error-code\01\8d\01\03\00\16wasi:http/types@0.2.12\05\09\01@\00\00y\04\00\05probe\01\0a\01@\01\05county\00y\04\00\0cfields-batch\01\0b\04\00#wasmvm:http-field-tests/http-fields\04\00\0b\11\01\00\0bhttp-fields\03\00\00\00/\09producers\01\0cprocessed-by\01\0dwit-component\070.258.0")
)

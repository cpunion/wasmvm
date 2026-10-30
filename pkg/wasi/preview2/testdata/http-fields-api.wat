(component
  (type $ty-wasi:io/poll@0.2.12 (;0;)
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
  (import "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12 (;0;) (type $ty-wasi:io/poll@0.2.12)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $pollable (;1;)))
  (type $ty-wasi:clocks/monotonic-clock@0.2.12 (;2;)
    (instance
      (type (;0;) u64)
      (export (;1;) "instant" (type (eq 0)))
      (type (;2;) u64)
      (export (;3;) "duration" (type (eq 2)))
      (alias outer 1 $pollable (type (;4;)))
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
  (import "wasi:clocks/monotonic-clock@0.2.12" (instance $wasi:clocks/monotonic-clock@0.2.12 (;1;) (type $ty-wasi:clocks/monotonic-clock@0.2.12)))
  (type $ty-wasi:io/error@0.2.12 (;3;)
    (instance
      (export (;0;) "error" (type (sub resource)))
      (type (;1;) (borrow 0))
      (type (;2;) (func (param "self" 1) (result string)))
      (export (;0;) "[method]error.to-debug-string" (func (type 2)))
    )
  )
  (import "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12 (;2;) (type $ty-wasi:io/error@0.2.12)))
  (alias export $wasi:io/error@0.2.12 "error" (type $error (;4;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type5 pollable" (@name "pollable") (;5;)))
  (type $ty-wasi:io/streams@0.2.12 (;6;)
    (instance
      (export (;0;) "input-stream" (type (sub resource)))
      (alias outer 1 $error (type (;1;)))
      (export (;2;) "error" (type (eq 1)))
      (type (;3;) (own 2))
      (type (;4;) (variant (case "last-operation-failed" 3) (case "closed")))
      (export (;5;) "stream-error" (type (eq 4)))
      (alias outer 1 $"#type5 pollable" (type (;6;)))
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
  (import "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12 (;3;) (type $ty-wasi:io/streams@0.2.12)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "duration" (type $duration (;7;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $input-stream (;8;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type9 pollable" (@name "pollable") (;9;)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $output-stream (;10;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type11 error" (@name "error") (;11;)))
  (type $ty-wasi:http/types@0.2.12 (;12;)
    (instance
      (export (;0;) "fields" (type (sub resource)))
      (type (;1;) string)
      (export (;2;) "field-key" (type (eq 1)))
      (export (;3;) "field-name" (type (eq 2)))
      (type (;4;) (list u8))
      (export (;5;) "field-value" (type (eq 4)))
      (type (;6;) (variant (case "invalid-syntax") (case "forbidden") (case "immutable")))
      (export (;7;) "header-error" (type (eq 6)))
      (export (;8;) "incoming-request" (type (sub resource)))
      (type (;9;) (variant (case "get") (case "head") (case "post") (case "put") (case "delete") (case "connect") (case "options") (case "trace") (case "patch") (case "other" string)))
      (export (;10;) "method" (type (eq 9)))
      (type (;11;) (variant (case "HTTP") (case "HTTPS") (case "other" string)))
      (export (;12;) "scheme" (type (eq 11)))
      (export (;13;) "headers" (type (eq 0)))
      (export (;14;) "incoming-body" (type (sub resource)))
      (export (;15;) "outgoing-request" (type (sub resource)))
      (export (;16;) "outgoing-body" (type (sub resource)))
      (export (;17;) "request-options" (type (sub resource)))
      (alias outer 1 $duration (type (;18;)))
      (export (;19;) "duration" (type (eq 18)))
      (export (;20;) "response-outparam" (type (sub resource)))
      (export (;21;) "outgoing-response" (type (sub resource)))
      (type (;22;) (option string))
      (type (;23;) (option u16))
      (type (;24;) (record (field "rcode" 22) (field "info-code" 23)))
      (export (;25;) "DNS-error-payload" (type (eq 24)))
      (type (;26;) (option u8))
      (type (;27;) (record (field "alert-id" 26) (field "alert-message" 22)))
      (export (;28;) "TLS-alert-received-payload" (type (eq 27)))
      (type (;29;) (option u32))
      (type (;30;) (record (field "field-name" 22) (field "field-size" 29)))
      (export (;31;) "field-size-payload" (type (eq 30)))
      (type (;32;) (option u64))
      (type (;33;) (option 31))
      (type (;34;) (variant (case "DNS-timeout") (case "DNS-error" 25) (case "destination-not-found") (case "destination-unavailable") (case "destination-IP-prohibited") (case "destination-IP-unroutable") (case "connection-refused") (case "connection-terminated") (case "connection-timeout") (case "connection-read-timeout") (case "connection-write-timeout") (case "connection-limit-reached") (case "TLS-protocol-error") (case "TLS-certificate-error") (case "TLS-alert-received" 28) (case "HTTP-request-denied") (case "HTTP-request-length-required") (case "HTTP-request-body-size" 32) (case "HTTP-request-method-invalid") (case "HTTP-request-URI-invalid") (case "HTTP-request-URI-too-long") (case "HTTP-request-header-section-size" 29) (case "HTTP-request-header-size" 33) (case "HTTP-request-trailer-section-size" 29) (case "HTTP-request-trailer-size" 31) (case "HTTP-response-incomplete") (case "HTTP-response-header-section-size" 29) (case "HTTP-response-header-size" 31) (case "HTTP-response-body-size" 32) (case "HTTP-response-trailer-section-size" 29) (case "HTTP-response-trailer-size" 31) (case "HTTP-response-transfer-coding" 22) (case "HTTP-response-content-coding" 22) (case "HTTP-response-timeout") (case "HTTP-upgrade-failed") (case "HTTP-protocol-error") (case "loop-detected") (case "configuration-error") (case "internal-error" 22)))
      (export (;35;) "error-code" (type (eq 34)))
      (export (;36;) "incoming-response" (type (sub resource)))
      (type (;37;) u16)
      (export (;38;) "status-code" (type (eq 37)))
      (alias outer 1 $input-stream (type (;39;)))
      (export (;40;) "input-stream" (type (eq 39)))
      (export (;41;) "future-trailers" (type (sub resource)))
      (alias outer 1 $"#type9 pollable" (type (;42;)))
      (export (;43;) "pollable" (type (eq 42)))
      (export (;44;) "trailers" (type (eq 0)))
      (alias outer 1 $output-stream (type (;45;)))
      (export (;46;) "output-stream" (type (eq 45)))
      (export (;47;) "future-incoming-response" (type (sub resource)))
      (alias outer 1 $"#type11 error" (type (;48;)))
      (export (;49;) "io-error" (type (eq 48)))
      (type (;50;) (own 0))
      (type (;51;) (func (result 50)))
      (export (;0;) "[constructor]fields" (func (type 51)))
      (type (;52;) (tuple 3 5))
      (type (;53;) (list 52))
      (type (;54;) (result 50 (error 7)))
      (type (;55;) (func (param "entries" 53) (result 54)))
      (export (;1;) "[static]fields.from-list" (func (type 55)))
      (type (;56;) (borrow 0))
      (type (;57;) (list 5))
      (type (;58;) (func (param "self" 56) (param "name" 3) (result 57)))
      (export (;2;) "[method]fields.get" (func (type 58)))
      (type (;59;) (func (param "self" 56) (param "name" 3) (result bool)))
      (export (;3;) "[method]fields.has" (func (type 59)))
      (type (;60;) (result (error 7)))
      (type (;61;) (func (param "self" 56) (param "name" 3) (param "value" 57) (result 60)))
      (export (;4;) "[method]fields.set" (func (type 61)))
      (type (;62;) (func (param "self" 56) (param "name" 3) (result 60)))
      (export (;5;) "[method]fields.delete" (func (type 62)))
      (type (;63;) (func (param "self" 56) (param "name" 3) (param "value" 5) (result 60)))
      (export (;6;) "[method]fields.append" (func (type 63)))
      (type (;64;) (func (param "self" 56) (result 53)))
      (export (;7;) "[method]fields.entries" (func (type 64)))
      (type (;65;) (func (param "self" 56) (result 50)))
      (export (;8;) "[method]fields.clone" (func (type 65)))
      (type (;66;) (borrow 8))
      (type (;67;) (func (param "self" 66) (result 10)))
      (export (;9;) "[method]incoming-request.method" (func (type 67)))
      (type (;68;) (func (param "self" 66) (result 22)))
      (export (;10;) "[method]incoming-request.path-with-query" (func (type 68)))
      (type (;69;) (option 12))
      (type (;70;) (func (param "self" 66) (result 69)))
      (export (;11;) "[method]incoming-request.scheme" (func (type 70)))
      (export (;12;) "[method]incoming-request.authority" (func (type 68)))
      (type (;71;) (own 13))
      (type (;72;) (func (param "self" 66) (result 71)))
      (export (;13;) "[method]incoming-request.headers" (func (type 72)))
      (type (;73;) (own 14))
      (type (;74;) (result 73))
      (type (;75;) (func (param "self" 66) (result 74)))
      (export (;14;) "[method]incoming-request.consume" (func (type 75)))
      (type (;76;) (own 15))
      (type (;77;) (func (param "headers" 71) (result 76)))
      (export (;15;) "[constructor]outgoing-request" (func (type 77)))
      (type (;78;) (borrow 15))
      (type (;79;) (own 16))
      (type (;80;) (result 79))
      (type (;81;) (func (param "self" 78) (result 80)))
      (export (;16;) "[method]outgoing-request.body" (func (type 81)))
      (type (;82;) (func (param "self" 78) (result 10)))
      (export (;17;) "[method]outgoing-request.method" (func (type 82)))
      (type (;83;) (result))
      (type (;84;) (func (param "self" 78) (param "method" 10) (result 83)))
      (export (;18;) "[method]outgoing-request.set-method" (func (type 84)))
      (type (;85;) (func (param "self" 78) (result 22)))
      (export (;19;) "[method]outgoing-request.path-with-query" (func (type 85)))
      (type (;86;) (func (param "self" 78) (param "path-with-query" 22) (result 83)))
      (export (;20;) "[method]outgoing-request.set-path-with-query" (func (type 86)))
      (type (;87;) (func (param "self" 78) (result 69)))
      (export (;21;) "[method]outgoing-request.scheme" (func (type 87)))
      (type (;88;) (func (param "self" 78) (param "scheme" 69) (result 83)))
      (export (;22;) "[method]outgoing-request.set-scheme" (func (type 88)))
      (export (;23;) "[method]outgoing-request.authority" (func (type 85)))
      (type (;89;) (func (param "self" 78) (param "authority" 22) (result 83)))
      (export (;24;) "[method]outgoing-request.set-authority" (func (type 89)))
      (type (;90;) (func (param "self" 78) (result 71)))
      (export (;25;) "[method]outgoing-request.headers" (func (type 90)))
      (type (;91;) (own 17))
      (type (;92;) (func (result 91)))
      (export (;26;) "[constructor]request-options" (func (type 92)))
      (type (;93;) (borrow 17))
      (type (;94;) (option 19))
      (type (;95;) (func (param "self" 93) (result 94)))
      (export (;27;) "[method]request-options.connect-timeout" (func (type 95)))
      (type (;96;) (func (param "self" 93) (param "duration" 94) (result 83)))
      (export (;28;) "[method]request-options.set-connect-timeout" (func (type 96)))
      (export (;29;) "[method]request-options.first-byte-timeout" (func (type 95)))
      (export (;30;) "[method]request-options.set-first-byte-timeout" (func (type 96)))
      (export (;31;) "[method]request-options.between-bytes-timeout" (func (type 95)))
      (export (;32;) "[method]request-options.set-between-bytes-timeout" (func (type 96)))
      (type (;97;) (own 20))
      (type (;98;) (own 21))
      (type (;99;) (result 98 (error 35)))
      (type (;100;) (func (param "param" 97) (param "response" 99)))
      (export (;33;) "[static]response-outparam.set" (func (type 100)))
      (type (;101;) (borrow 36))
      (type (;102;) (func (param "self" 101) (result 38)))
      (export (;34;) "[method]incoming-response.status" (func (type 102)))
      (type (;103;) (func (param "self" 101) (result 71)))
      (export (;35;) "[method]incoming-response.headers" (func (type 103)))
      (type (;104;) (func (param "self" 101) (result 74)))
      (export (;36;) "[method]incoming-response.consume" (func (type 104)))
      (type (;105;) (borrow 14))
      (type (;106;) (own 40))
      (type (;107;) (result 106))
      (type (;108;) (func (param "self" 105) (result 107)))
      (export (;37;) "[method]incoming-body.stream" (func (type 108)))
      (type (;109;) (own 41))
      (type (;110;) (func (param "this" 73) (result 109)))
      (export (;38;) "[static]incoming-body.finish" (func (type 110)))
      (type (;111;) (borrow 41))
      (type (;112;) (own 43))
      (type (;113;) (func (param "self" 111) (result 112)))
      (export (;39;) "[method]future-trailers.subscribe" (func (type 113)))
      (type (;114;) (own 44))
      (type (;115;) (option 114))
      (type (;116;) (result 115 (error 35)))
      (type (;117;) (result 116))
      (type (;118;) (option 117))
      (type (;119;) (func (param "self" 111) (result 118)))
      (export (;40;) "[method]future-trailers.get" (func (type 119)))
      (type (;120;) (func (param "headers" 71) (result 98)))
      (export (;41;) "[constructor]outgoing-response" (func (type 120)))
      (type (;121;) (borrow 21))
      (type (;122;) (func (param "self" 121) (result 38)))
      (export (;42;) "[method]outgoing-response.status-code" (func (type 122)))
      (type (;123;) (func (param "self" 121) (param "status-code" 38) (result 83)))
      (export (;43;) "[method]outgoing-response.set-status-code" (func (type 123)))
      (type (;124;) (func (param "self" 121) (result 71)))
      (export (;44;) "[method]outgoing-response.headers" (func (type 124)))
      (type (;125;) (func (param "self" 121) (result 80)))
      (export (;45;) "[method]outgoing-response.body" (func (type 125)))
      (type (;126;) (borrow 16))
      (type (;127;) (own 46))
      (type (;128;) (result 127))
      (type (;129;) (func (param "self" 126) (result 128)))
      (export (;46;) "[method]outgoing-body.write" (func (type 129)))
      (type (;130;) (result (error 35)))
      (type (;131;) (func (param "this" 79) (param "trailers" 115) (result 130)))
      (export (;47;) "[static]outgoing-body.finish" (func (type 131)))
      (type (;132;) (borrow 47))
      (type (;133;) (func (param "self" 132) (result 112)))
      (export (;48;) "[method]future-incoming-response.subscribe" (func (type 133)))
      (type (;134;) (own 36))
      (type (;135;) (result 134 (error 35)))
      (type (;136;) (result 135))
      (type (;137;) (option 136))
      (type (;138;) (func (param "self" 132) (result 137)))
      (export (;49;) "[method]future-incoming-response.get" (func (type 138)))
      (type (;139;) (borrow 49))
      (type (;140;) (option 35))
      (type (;141;) (func (param "err" 139) (result 140)))
      (export (;50;) "http-error-code" (func (type 141)))
    )
  )
  (import "wasi:http/types@0.2.12" (instance $wasi:http/types@0.2.12 (;4;) (type $ty-wasi:http/types@0.2.12)))
  (core module $main (;0;)
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
    (global $heap (;0;) (mut i32) i32.const 8192)
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
    (func (;90;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;91;) (type 0) (param i32) (result i32)
      (local $at i32)
      local.get 0
      i32.const 128
      i32.gt_u
      if ;; label = @1
        unreachable
      end
      block $end
        loop $next
          local.get $at
          local.get 0
          i32.ge_u
          br_if $end
          call $probe
          i32.const 42
          i32.ne
          if ;; label = @3
            unreachable
          end
          local.get $at
          i32.const 1
          i32.add
          local.set $at
          br $next
        end
      end
      local.get $at
    )
    (func (;92;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;93;) (type 12) (param i32 i32 i32 i32) (result i32)
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
    (func (;94;) (type 17))
    (func $ok (;95;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (func $probe (;96;) (type 9) (result i32)
      (local $h i32) (local $copy i32) (local $other i32) (local $list i32)
      call $new-fields
      local.set $h
      local.get $h
      i32.const 1024
      i32.const 5
      i32.const 1032
      i32.const 3
      i32.const 64
      call $fields.append
      i32.const 64
      call $ok
      i32.const 256
      i32.const 1032
      i32.store
      i32.const 260
      i32.const 3
      i32.store
      i32.const 264
      i32.const 0
      i32.store
      i32.const 268
      i32.const 0
      i32.store
      local.get $h
      i32.const 1024
      i32.const 5
      i32.const 256
      i32.const 2
      i32.const 64
      call $fields.set
      i32.const 64
      call $ok
      local.get $h
      i32.const 1040
      i32.const 5
      call $fields.has
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $h
      i32.const 1040
      i32.const 5
      i32.const 64
      call $fields.get
      i32.const 68
      i32.load
      i32.const 2
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 64
      i32.load
      local.tee $list
      i32.const 4
      i32.add
      i32.load
      i32.const 3
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $list
      i32.const 12
      i32.add
      i32.load
      if ;; label = @1
        unreachable
      end
      local.get $h
      i32.const 64
      call $fields.entries
      i32.const 68
      i32.load
      i32.const 2
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $h
      call $fields.clone
      local.set $copy
      local.get $copy
      i32.const 1024
      i32.const 5
      i32.const 64
      call $fields.delete
      i32.const 64
      call $ok
      local.get $copy
      i32.const 1040
      i32.const 5
      call $fields.has
      if ;; label = @1
        unreachable
      end
      i32.const 288
      i32.const 1024
      i32.store
      i32.const 292
      i32.const 5
      i32.store
      i32.const 296
      i32.const 1032
      i32.store
      i32.const 300
      i32.const 3
      i32.store
      i32.const 288
      i32.const 1
      i32.const 64
      call $fields.from-list
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $other
      local.get $other
      i32.const 1056
      i32.const 8
      i32.const 1032
      i32.const 3
      i32.const 64
      call $fields.append
      i32.const 64
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 65
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      local.get $h
      call $drop-fields
      local.get $copy
      call $drop-fields
      local.get $other
      call $drop-fields
      i32.const 42
    )
    (data (;0;) (i32.const 1024) "x-key")
    (data (;1;) (i32.const 1032) "one")
    (data (;2;) (i32.const 1040) "X-KEY")
    (data (;3;) (i32.const 1056) "bad name")
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core module $wit-component-shim-module (;1;)
    (type (;0;) (func (param i32 i32 i32)))
    (type (;1;) (func (param i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32) (result i32)))
    (type (;6;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;7;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;8;) (func (param i32 i32 i32 i32 i32) (result i32)))
    (type (;9;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
    (table (;0;) 48 48 funcref)
    (export "0" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "1" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
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
    (export "15" (func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list"))
    (export "16" (func $"indirect-wasi:http/types@0.2.12-[method]fields.get"))
    (export "17" (func $"indirect-wasi:http/types@0.2.12-[method]fields.has"))
    (export "18" (func $"indirect-wasi:http/types@0.2.12-[method]fields.set"))
    (export "19" (func $"indirect-wasi:http/types@0.2.12-[method]fields.delete"))
    (export "20" (func $"indirect-wasi:http/types@0.2.12-[method]fields.append"))
    (export "21" (func $"indirect-wasi:http/types@0.2.12-[method]fields.entries"))
    (export "22" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method"))
    (export "23" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query"))
    (export "24" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme"))
    (export "25" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority"))
    (export "26" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume"))
    (export "27" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body"))
    (export "28" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method"))
    (export "29" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method"))
    (export "30" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query"))
    (export "31" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query"))
    (export "32" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme"))
    (export "33" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme"))
    (export "34" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority"))
    (export "35" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority"))
    (export "36" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout"))
    (export "37" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout"))
    (export "38" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout"))
    (export "39" (func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set"))
    (export "40" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume"))
    (export "41" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream"))
    (export "42" (func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get"))
    (export "43" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body"))
    (export "44" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write"))
    (export "45" (func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish"))
    (export "46" (func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get"))
    (export "47" (func $indirect-wasi:http/types@0.2.12-http-error-code))
    (export "$imports" (table 0))
    (func $indirect-wasi:io/poll@0.2.12-poll (;0;) (type 0) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 0
      call_indirect (type 0)
    )
    (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;1;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
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
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;6;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 6
      call_indirect (type 1)
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
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;9;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 9
      call_indirect (type 1)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;10;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 10
      call_indirect (type 1)
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
    (func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list" (;15;) (type 0) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 15
      call_indirect (type 0)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.get" (;16;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 16
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.has" (;17;) (type 5) (param i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 17
      call_indirect (type 5)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.set" (;18;) (type 6) (param i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      i32.const 18
      call_indirect (type 6)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.delete" (;19;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 19
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.append" (;20;) (type 6) (param i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      i32.const 20
      call_indirect (type 6)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.entries" (;21;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 21
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (;22;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 22
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (;23;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 23
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (;24;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 24
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (;25;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 25
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (;26;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 26
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (;27;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 27
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (;28;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 28
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (;29;) (type 7) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 29
      call_indirect (type 7)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (;30;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 30
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (;31;) (type 7) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 31
      call_indirect (type 7)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (;32;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 32
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (;33;) (type 8) (param i32 i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 33
      call_indirect (type 8)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (;34;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 34
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (;35;) (type 7) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 35
      call_indirect (type 7)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (;36;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 36
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (;37;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 37
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (;38;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 38
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (;39;) (type 9) (param i32 i32 i32 i32 i64 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      local.get 8
      i32.const 39
      call_indirect (type 9)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (;40;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 40
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (;41;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 41
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (;42;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 42
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (;43;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 43
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (;44;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 44
      call_indirect (type 1)
    )
    (func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (;45;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 45
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (;46;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 46
      call_indirect (type 1)
    )
    (func $indirect-wasi:http/types@0.2.12-http-error-code (;47;) (type 1) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 47
      call_indirect (type 1)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;0;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;1;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "0" (core func $indirect-wasi:io/poll@0.2.12-poll (;2;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type13 pollable" (@name "pollable") (;13;)))
  (core func $resource.drop (;3;) (canon resource.drop $"#type13 pollable"))
  (core instance $wasi:io/poll@0.2.12 (;1;)
    (export "[method]pollable.ready" (func $"[method]pollable.ready"))
    (export "[method]pollable.block" (func $"[method]pollable.block"))
    (export "poll" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "[resource-drop]pollable" (func $resource.drop))
  )
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "now" (func $now (;2;)))
  (core func $now (;4;) (canon lower (func $now)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "resolution" (func $resolution (;3;)))
  (core func $resolution (;5;) (canon lower (func $resolution)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "subscribe-instant" (func $subscribe-instant (;4;)))
  (core func $subscribe-instant (;6;) (canon lower (func $subscribe-instant)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "subscribe-duration" (func $subscribe-duration (;5;)))
  (core func $subscribe-duration (;7;) (canon lower (func $subscribe-duration)))
  (core instance $wasi:clocks/monotonic-clock@0.2.12 (;2;)
    (export "now" (func $now))
    (export "resolution" (func $resolution))
    (export "subscribe-instant" (func $subscribe-instant))
    (export "subscribe-duration" (func $subscribe-duration))
  )
  (alias core export $wit-component-shim-instance "1" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;8;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type14 error" (@name "error") (;14;)))
  (core func $"#core-func9 resource.drop" (@name "resource.drop") (;9;) (canon resource.drop $"#type14 error"))
  (core instance $wasi:io/error@0.2.12 (;3;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $"#core-func9 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "2" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (;10;)))
  (alias core export $wit-component-shim-instance "3" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (;11;)))
  (alias core export $wit-component-shim-instance "4" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (;12;)))
  (alias core export $wit-component-shim-instance "5" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (;13;)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.subscribe" (func $"[method]input-stream.subscribe" (;6;)))
  (core func $"[method]input-stream.subscribe" (;14;) (canon lower (func $"[method]input-stream.subscribe")))
  (alias core export $wit-component-shim-instance "6" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;15;)))
  (alias core export $wit-component-shim-instance "7" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (;16;)))
  (alias core export $wit-component-shim-instance "8" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (;17;)))
  (alias core export $wit-component-shim-instance "9" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;18;)))
  (alias core export $wit-component-shim-instance "10" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;19;)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.subscribe" (func $"[method]output-stream.subscribe" (;7;)))
  (core func $"[method]output-stream.subscribe" (;20;) (canon lower (func $"[method]output-stream.subscribe")))
  (alias core export $wit-component-shim-instance "11" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (;21;)))
  (alias core export $wit-component-shim-instance "12" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (;22;)))
  (alias core export $wit-component-shim-instance "13" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (;23;)))
  (alias core export $wit-component-shim-instance "14" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (;24;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type15 input-stream" (@name "input-stream") (;15;)))
  (core func $"#core-func25 resource.drop" (@name "resource.drop") (;25;) (canon resource.drop $"#type15 input-stream"))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type16 output-stream" (@name "output-stream") (;16;)))
  (core func $"#core-func26 resource.drop" (@name "resource.drop") (;26;) (canon resource.drop $"#type16 output-stream"))
  (core instance $wasi:io/streams@0.2.12 (;4;)
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
    (export "[resource-drop]input-stream" (func $"#core-func25 resource.drop"))
    (export "[resource-drop]output-stream" (func $"#core-func26 resource.drop"))
  )
  (alias export $wasi:http/types@0.2.12 "[constructor]fields" (func $"[constructor]fields" (;8;)))
  (core func $"[constructor]fields" (;27;) (canon lower (func $"[constructor]fields")))
  (alias core export $wit-component-shim-instance "15" (core func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list" (;28;)))
  (alias core export $wit-component-shim-instance "16" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.get" (;29;)))
  (alias core export $wit-component-shim-instance "17" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.has" (;30;)))
  (alias core export $wit-component-shim-instance "18" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.set" (;31;)))
  (alias core export $wit-component-shim-instance "19" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.delete" (;32;)))
  (alias core export $wit-component-shim-instance "20" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.append" (;33;)))
  (alias core export $wit-component-shim-instance "21" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.entries" (;34;)))
  (alias export $wasi:http/types@0.2.12 "[method]fields.clone" (func $"[method]fields.clone" (;9;)))
  (core func $"[method]fields.clone" (;35;) (canon lower (func $"[method]fields.clone")))
  (alias core export $wit-component-shim-instance "22" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (;36;)))
  (alias core export $wit-component-shim-instance "23" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (;37;)))
  (alias core export $wit-component-shim-instance "24" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (;38;)))
  (alias core export $wit-component-shim-instance "25" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (;39;)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.headers" (func $"[method]incoming-request.headers" (;10;)))
  (core func $"[method]incoming-request.headers" (;40;) (canon lower (func $"[method]incoming-request.headers")))
  (alias core export $wit-component-shim-instance "26" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (;41;)))
  (alias export $wasi:http/types@0.2.12 "[constructor]outgoing-request" (func $"[constructor]outgoing-request" (;11;)))
  (core func $"[constructor]outgoing-request" (;42;) (canon lower (func $"[constructor]outgoing-request")))
  (alias core export $wit-component-shim-instance "27" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (;43;)))
  (alias core export $wit-component-shim-instance "28" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (;44;)))
  (alias core export $wit-component-shim-instance "29" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (;45;)))
  (alias core export $wit-component-shim-instance "30" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (;46;)))
  (alias core export $wit-component-shim-instance "31" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (;47;)))
  (alias core export $wit-component-shim-instance "32" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (;48;)))
  (alias core export $wit-component-shim-instance "33" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (;49;)))
  (alias core export $wit-component-shim-instance "34" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (;50;)))
  (alias core export $wit-component-shim-instance "35" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (;51;)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.headers" (func $"[method]outgoing-request.headers" (;12;)))
  (core func $"[method]outgoing-request.headers" (;52;) (canon lower (func $"[method]outgoing-request.headers")))
  (alias export $wasi:http/types@0.2.12 "[constructor]request-options" (func $"[constructor]request-options" (;13;)))
  (core func $"[constructor]request-options" (;53;) (canon lower (func $"[constructor]request-options")))
  (alias core export $wit-component-shim-instance "36" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (;54;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-connect-timeout" (func $"[method]request-options.set-connect-timeout" (;14;)))
  (core func $"[method]request-options.set-connect-timeout" (;55;) (canon lower (func $"[method]request-options.set-connect-timeout")))
  (alias core export $wit-component-shim-instance "37" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (;56;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-first-byte-timeout" (func $"[method]request-options.set-first-byte-timeout" (;15;)))
  (core func $"[method]request-options.set-first-byte-timeout" (;57;) (canon lower (func $"[method]request-options.set-first-byte-timeout")))
  (alias core export $wit-component-shim-instance "38" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (;58;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-between-bytes-timeout" (func $"[method]request-options.set-between-bytes-timeout" (;16;)))
  (core func $"[method]request-options.set-between-bytes-timeout" (;59;) (canon lower (func $"[method]request-options.set-between-bytes-timeout")))
  (alias core export $wit-component-shim-instance "39" (core func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (;60;)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.status" (func $"[method]incoming-response.status" (;17;)))
  (core func $"[method]incoming-response.status" (;61;) (canon lower (func $"[method]incoming-response.status")))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.headers" (func $"[method]incoming-response.headers" (;18;)))
  (core func $"[method]incoming-response.headers" (;62;) (canon lower (func $"[method]incoming-response.headers")))
  (alias core export $wit-component-shim-instance "40" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (;63;)))
  (alias core export $wit-component-shim-instance "41" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (;64;)))
  (alias export $wasi:http/types@0.2.12 "[static]incoming-body.finish" (func $"[static]incoming-body.finish" (;19;)))
  (core func $"[static]incoming-body.finish" (;65;) (canon lower (func $"[static]incoming-body.finish")))
  (alias export $wasi:http/types@0.2.12 "[method]future-trailers.subscribe" (func $"[method]future-trailers.subscribe" (;20;)))
  (core func $"[method]future-trailers.subscribe" (;66;) (canon lower (func $"[method]future-trailers.subscribe")))
  (alias core export $wit-component-shim-instance "42" (core func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (;67;)))
  (alias export $wasi:http/types@0.2.12 "[constructor]outgoing-response" (func $"[constructor]outgoing-response" (;21;)))
  (core func $"[constructor]outgoing-response" (;68;) (canon lower (func $"[constructor]outgoing-response")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.status-code" (func $"[method]outgoing-response.status-code" (;22;)))
  (core func $"[method]outgoing-response.status-code" (;69;) (canon lower (func $"[method]outgoing-response.status-code")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.set-status-code" (func $"[method]outgoing-response.set-status-code" (;23;)))
  (core func $"[method]outgoing-response.set-status-code" (;70;) (canon lower (func $"[method]outgoing-response.set-status-code")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.headers" (func $"[method]outgoing-response.headers" (;24;)))
  (core func $"[method]outgoing-response.headers" (;71;) (canon lower (func $"[method]outgoing-response.headers")))
  (alias core export $wit-component-shim-instance "43" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (;72;)))
  (alias core export $wit-component-shim-instance "44" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (;73;)))
  (alias core export $wit-component-shim-instance "45" (core func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (;74;)))
  (alias export $wasi:http/types@0.2.12 "[method]future-incoming-response.subscribe" (func $"[method]future-incoming-response.subscribe" (;25;)))
  (core func $"[method]future-incoming-response.subscribe" (;75;) (canon lower (func $"[method]future-incoming-response.subscribe")))
  (alias core export $wit-component-shim-instance "46" (core func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (;76;)))
  (alias core export $wit-component-shim-instance "47" (core func $indirect-wasi:http/types@0.2.12-http-error-code (;77;)))
  (alias export $wasi:http/types@0.2.12 "fields" (type $fields (;17;)))
  (core func $"#core-func78 resource.drop" (@name "resource.drop") (;78;) (canon resource.drop $fields))
  (alias export $wasi:http/types@0.2.12 "incoming-request" (type $incoming-request (;18;)))
  (core func $"#core-func79 resource.drop" (@name "resource.drop") (;79;) (canon resource.drop $incoming-request))
  (alias export $wasi:http/types@0.2.12 "outgoing-request" (type $outgoing-request (;19;)))
  (core func $"#core-func80 resource.drop" (@name "resource.drop") (;80;) (canon resource.drop $outgoing-request))
  (alias export $wasi:http/types@0.2.12 "request-options" (type $request-options (;20;)))
  (core func $"#core-func81 resource.drop" (@name "resource.drop") (;81;) (canon resource.drop $request-options))
  (alias export $wasi:http/types@0.2.12 "response-outparam" (type $response-outparam (;21;)))
  (core func $"#core-func82 resource.drop" (@name "resource.drop") (;82;) (canon resource.drop $response-outparam))
  (alias export $wasi:http/types@0.2.12 "incoming-response" (type $incoming-response (;22;)))
  (core func $"#core-func83 resource.drop" (@name "resource.drop") (;83;) (canon resource.drop $incoming-response))
  (alias export $wasi:http/types@0.2.12 "incoming-body" (type $incoming-body (;23;)))
  (core func $"#core-func84 resource.drop" (@name "resource.drop") (;84;) (canon resource.drop $incoming-body))
  (alias export $wasi:http/types@0.2.12 "future-trailers" (type $future-trailers (;24;)))
  (core func $"#core-func85 resource.drop" (@name "resource.drop") (;85;) (canon resource.drop $future-trailers))
  (alias export $wasi:http/types@0.2.12 "outgoing-response" (type $outgoing-response (;25;)))
  (core func $"#core-func86 resource.drop" (@name "resource.drop") (;86;) (canon resource.drop $outgoing-response))
  (alias export $wasi:http/types@0.2.12 "outgoing-body" (type $outgoing-body (;26;)))
  (core func $"#core-func87 resource.drop" (@name "resource.drop") (;87;) (canon resource.drop $outgoing-body))
  (alias export $wasi:http/types@0.2.12 "future-incoming-response" (type $future-incoming-response (;27;)))
  (core func $"#core-func88 resource.drop" (@name "resource.drop") (;88;) (canon resource.drop $future-incoming-response))
  (core instance $wasi:http/types@0.2.12 (;5;)
    (export "[constructor]fields" (func $"[constructor]fields"))
    (export "[static]fields.from-list" (func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list"))
    (export "[method]fields.get" (func $"indirect-wasi:http/types@0.2.12-[method]fields.get"))
    (export "[method]fields.has" (func $"indirect-wasi:http/types@0.2.12-[method]fields.has"))
    (export "[method]fields.set" (func $"indirect-wasi:http/types@0.2.12-[method]fields.set"))
    (export "[method]fields.delete" (func $"indirect-wasi:http/types@0.2.12-[method]fields.delete"))
    (export "[method]fields.append" (func $"indirect-wasi:http/types@0.2.12-[method]fields.append"))
    (export "[method]fields.entries" (func $"indirect-wasi:http/types@0.2.12-[method]fields.entries"))
    (export "[method]fields.clone" (func $"[method]fields.clone"))
    (export "[method]incoming-request.method" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method"))
    (export "[method]incoming-request.path-with-query" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query"))
    (export "[method]incoming-request.scheme" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme"))
    (export "[method]incoming-request.authority" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority"))
    (export "[method]incoming-request.headers" (func $"[method]incoming-request.headers"))
    (export "[method]incoming-request.consume" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume"))
    (export "[constructor]outgoing-request" (func $"[constructor]outgoing-request"))
    (export "[method]outgoing-request.body" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body"))
    (export "[method]outgoing-request.method" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method"))
    (export "[method]outgoing-request.set-method" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method"))
    (export "[method]outgoing-request.path-with-query" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query"))
    (export "[method]outgoing-request.set-path-with-query" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query"))
    (export "[method]outgoing-request.scheme" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme"))
    (export "[method]outgoing-request.set-scheme" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme"))
    (export "[method]outgoing-request.authority" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority"))
    (export "[method]outgoing-request.set-authority" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority"))
    (export "[method]outgoing-request.headers" (func $"[method]outgoing-request.headers"))
    (export "[constructor]request-options" (func $"[constructor]request-options"))
    (export "[method]request-options.connect-timeout" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout"))
    (export "[method]request-options.set-connect-timeout" (func $"[method]request-options.set-connect-timeout"))
    (export "[method]request-options.first-byte-timeout" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout"))
    (export "[method]request-options.set-first-byte-timeout" (func $"[method]request-options.set-first-byte-timeout"))
    (export "[method]request-options.between-bytes-timeout" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout"))
    (export "[method]request-options.set-between-bytes-timeout" (func $"[method]request-options.set-between-bytes-timeout"))
    (export "[static]response-outparam.set" (func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set"))
    (export "[method]incoming-response.status" (func $"[method]incoming-response.status"))
    (export "[method]incoming-response.headers" (func $"[method]incoming-response.headers"))
    (export "[method]incoming-response.consume" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume"))
    (export "[method]incoming-body.stream" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream"))
    (export "[static]incoming-body.finish" (func $"[static]incoming-body.finish"))
    (export "[method]future-trailers.subscribe" (func $"[method]future-trailers.subscribe"))
    (export "[method]future-trailers.get" (func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get"))
    (export "[constructor]outgoing-response" (func $"[constructor]outgoing-response"))
    (export "[method]outgoing-response.status-code" (func $"[method]outgoing-response.status-code"))
    (export "[method]outgoing-response.set-status-code" (func $"[method]outgoing-response.set-status-code"))
    (export "[method]outgoing-response.headers" (func $"[method]outgoing-response.headers"))
    (export "[method]outgoing-response.body" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body"))
    (export "[method]outgoing-body.write" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write"))
    (export "[static]outgoing-body.finish" (func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish"))
    (export "[method]future-incoming-response.subscribe" (func $"[method]future-incoming-response.subscribe"))
    (export "[method]future-incoming-response.get" (func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get"))
    (export "http-error-code" (func $indirect-wasi:http/types@0.2.12-http-error-code))
    (export "[resource-drop]fields" (func $"#core-func78 resource.drop"))
    (export "[resource-drop]incoming-request" (func $"#core-func79 resource.drop"))
    (export "[resource-drop]outgoing-request" (func $"#core-func80 resource.drop"))
    (export "[resource-drop]request-options" (func $"#core-func81 resource.drop"))
    (export "[resource-drop]response-outparam" (func $"#core-func82 resource.drop"))
    (export "[resource-drop]incoming-response" (func $"#core-func83 resource.drop"))
    (export "[resource-drop]incoming-body" (func $"#core-func84 resource.drop"))
    (export "[resource-drop]future-trailers" (func $"#core-func85 resource.drop"))
    (export "[resource-drop]outgoing-response" (func $"#core-func86 resource.drop"))
    (export "[resource-drop]outgoing-body" (func $"#core-func87 resource.drop"))
    (export "[resource-drop]future-incoming-response" (func $"#core-func88 resource.drop"))
  )
  (core instance $main (;6;) (instantiate $main
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:clocks/monotonic-clock@0.2.12" (instance $wasi:clocks/monotonic-clock@0.2.12))
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12))
      (with "wasi:http/types@0.2.12" (instance $wasi:http/types@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32 i32)))
    (type (;1;) (func (param i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32) (result i32)))
    (type (;6;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;7;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;8;) (func (param i32 i32 i32 i32 i32) (result i32)))
    (type (;9;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
    (type (;10;) (func))
    (import "actual" "0" (func $0 (;0;) (type 0)))
    (import "actual" "1" (func $1 (;1;) (type 1)))
    (import "actual" "2" (func $2 (;2;) (type 2)))
    (import "actual" "3" (func $3 (;3;) (type 2)))
    (import "actual" "4" (func $4 (;4;) (type 2)))
    (import "actual" "5" (func $5 (;5;) (type 2)))
    (import "actual" "6" (func $6 (;6;) (type 1)))
    (import "actual" "7" (func $7 (;7;) (type 3)))
    (import "actual" "8" (func $8 (;8;) (type 3)))
    (import "actual" "9" (func $9 (;9;) (type 1)))
    (import "actual" "10" (func $10 (;10;) (type 1)))
    (import "actual" "11" (func $11 (;11;) (type 2)))
    (import "actual" "12" (func $12 (;12;) (type 2)))
    (import "actual" "13" (func $13 (;13;) (type 4)))
    (import "actual" "14" (func $14 (;14;) (type 4)))
    (import "actual" "15" (func $15 (;15;) (type 0)))
    (import "actual" "16" (func $16 (;16;) (type 3)))
    (import "actual" "17" (func $17 (;17;) (type 5)))
    (import "actual" "18" (func $18 (;18;) (type 6)))
    (import "actual" "19" (func $19 (;19;) (type 3)))
    (import "actual" "20" (func $20 (;20;) (type 6)))
    (import "actual" "21" (func $21 (;21;) (type 1)))
    (import "actual" "22" (func $22 (;22;) (type 1)))
    (import "actual" "23" (func $23 (;23;) (type 1)))
    (import "actual" "24" (func $24 (;24;) (type 1)))
    (import "actual" "25" (func $25 (;25;) (type 1)))
    (import "actual" "26" (func $26 (;26;) (type 1)))
    (import "actual" "27" (func $27 (;27;) (type 1)))
    (import "actual" "28" (func $28 (;28;) (type 1)))
    (import "actual" "29" (func $29 (;29;) (type 7)))
    (import "actual" "30" (func $30 (;30;) (type 1)))
    (import "actual" "31" (func $31 (;31;) (type 7)))
    (import "actual" "32" (func $32 (;32;) (type 1)))
    (import "actual" "33" (func $33 (;33;) (type 8)))
    (import "actual" "34" (func $34 (;34;) (type 1)))
    (import "actual" "35" (func $35 (;35;) (type 7)))
    (import "actual" "36" (func $36 (;36;) (type 1)))
    (import "actual" "37" (func $37 (;37;) (type 1)))
    (import "actual" "38" (func $38 (;38;) (type 1)))
    (import "actual" "39" (func $39 (;39;) (type 9)))
    (import "actual" "40" (func $40 (;40;) (type 1)))
    (import "actual" "41" (func $41 (;41;) (type 1)))
    (import "actual" "42" (func $42 (;42;) (type 1)))
    (import "actual" "43" (func $43 (;43;) (type 1)))
    (import "actual" "44" (func $44 (;44;) (type 1)))
    (import "actual" "45" (func $45 (;45;) (type 3)))
    (import "actual" "46" (func $46 (;46;) (type 1)))
    (import "actual" "47" (func $47 (;47;) (type 1)))
    (import "main" "_initialize" (func $_initialize (;48;) (type 10)))
    (import "shim" "$imports" (table (;0;) 48 48 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16 $17 $18 $19 $20 $21 $22 $23 $24 $25 $26 $27 $28 $29 $30 $31 $32 $33 $34 $35 $36 $37 $38 $39 $40 $41 $42 $43 $44 $45 $46 $47)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;26;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;89;)))
  (core func $"#core-func90 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;90;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;27;)))
  (core func $"#core-func91 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;91;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.read" (func $"[method]input-stream.read" (;28;)))
  (core func $"#core-func92 indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.read") (;92;) (canon lower (func $"[method]input-stream.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-read" (func $"[method]input-stream.blocking-read" (;29;)))
  (core func $"#core-func93 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read") (;93;) (canon lower (func $"[method]input-stream.blocking-read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.skip" (func $"[method]input-stream.skip" (;30;)))
  (core func $"#core-func94 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.skip") (;94;) (canon lower (func $"[method]input-stream.skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-skip" (func $"[method]input-stream.blocking-skip" (;31;)))
  (core func $"#core-func95 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip") (;95;) (canon lower (func $"[method]input-stream.blocking-skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.check-write" (func $"[method]output-stream.check-write" (;32;)))
  (core func $"#core-func96 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write") (;96;) (canon lower (func $"[method]output-stream.check-write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write" (func $"[method]output-stream.write" (;33;)))
  (core func $"#core-func97 indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write") (;97;) (canon lower (func $"[method]output-stream.write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-and-flush" (func $"[method]output-stream.blocking-write-and-flush" (;34;)))
  (core func $"#core-func98 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush") (;98;) (canon lower (func $"[method]output-stream.blocking-write-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.flush" (func $"[method]output-stream.flush" (;35;)))
  (core func $"#core-func99 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.flush") (;99;) (canon lower (func $"[method]output-stream.flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-flush" (func $"[method]output-stream.blocking-flush" (;36;)))
  (core func $"#core-func100 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush") (;100;) (canon lower (func $"[method]output-stream.blocking-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write-zeroes" (func $"[method]output-stream.write-zeroes" (;37;)))
  (core func $"#core-func101 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes") (;101;) (canon lower (func $"[method]output-stream.write-zeroes") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-zeroes-and-flush" (func $"[method]output-stream.blocking-write-zeroes-and-flush" (;38;)))
  (core func $"#core-func102 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush") (;102;) (canon lower (func $"[method]output-stream.blocking-write-zeroes-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.splice" (func $"[method]output-stream.splice" (;39;)))
  (core func $"#core-func103 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.splice") (;103;) (canon lower (func $"[method]output-stream.splice") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-splice" (func $"[method]output-stream.blocking-splice" (;40;)))
  (core func $"#core-func104 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice") (;104;) (canon lower (func $"[method]output-stream.blocking-splice") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]fields.from-list" (func $"[static]fields.from-list" (;41;)))
  (core func $"#core-func105 indirect-wasi:http/types@0.2.12-[static]fields.from-list" (@name "indirect-wasi:http/types@0.2.12-[static]fields.from-list") (;105;) (canon lower (func $"[static]fields.from-list") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.get" (func $"[method]fields.get" (;42;)))
  (core func $"#core-func106 indirect-wasi:http/types@0.2.12-[method]fields.get" (@name "indirect-wasi:http/types@0.2.12-[method]fields.get") (;106;) (canon lower (func $"[method]fields.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.has" (func $"[method]fields.has" (;43;)))
  (core func $"#core-func107 indirect-wasi:http/types@0.2.12-[method]fields.has" (@name "indirect-wasi:http/types@0.2.12-[method]fields.has") (;107;) (canon lower (func $"[method]fields.has") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.set" (func $"[method]fields.set" (;44;)))
  (core func $"#core-func108 indirect-wasi:http/types@0.2.12-[method]fields.set" (@name "indirect-wasi:http/types@0.2.12-[method]fields.set") (;108;) (canon lower (func $"[method]fields.set") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.delete" (func $"[method]fields.delete" (;45;)))
  (core func $"#core-func109 indirect-wasi:http/types@0.2.12-[method]fields.delete" (@name "indirect-wasi:http/types@0.2.12-[method]fields.delete") (;109;) (canon lower (func $"[method]fields.delete") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.append" (func $"[method]fields.append" (;46;)))
  (core func $"#core-func110 indirect-wasi:http/types@0.2.12-[method]fields.append" (@name "indirect-wasi:http/types@0.2.12-[method]fields.append") (;110;) (canon lower (func $"[method]fields.append") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.entries" (func $"[method]fields.entries" (;47;)))
  (core func $"#core-func111 indirect-wasi:http/types@0.2.12-[method]fields.entries" (@name "indirect-wasi:http/types@0.2.12-[method]fields.entries") (;111;) (canon lower (func $"[method]fields.entries") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.method" (func $"[method]incoming-request.method" (;48;)))
  (core func $"#core-func112 indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.method") (;112;) (canon lower (func $"[method]incoming-request.method") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.path-with-query" (func $"[method]incoming-request.path-with-query" (;49;)))
  (core func $"#core-func113 indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query") (;113;) (canon lower (func $"[method]incoming-request.path-with-query") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.scheme" (func $"[method]incoming-request.scheme" (;50;)))
  (core func $"#core-func114 indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme") (;114;) (canon lower (func $"[method]incoming-request.scheme") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.authority" (func $"[method]incoming-request.authority" (;51;)))
  (core func $"#core-func115 indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.authority") (;115;) (canon lower (func $"[method]incoming-request.authority") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.consume" (func $"[method]incoming-request.consume" (;52;)))
  (core func $"#core-func116 indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.consume") (;116;) (canon lower (func $"[method]incoming-request.consume") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.body" (func $"[method]outgoing-request.body" (;53;)))
  (core func $"#core-func117 indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.body") (;117;) (canon lower (func $"[method]outgoing-request.body") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.method" (func $"[method]outgoing-request.method" (;54;)))
  (core func $"#core-func118 indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.method") (;118;) (canon lower (func $"[method]outgoing-request.method") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-method" (func $"[method]outgoing-request.set-method" (;55;)))
  (core func $"#core-func119 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method") (;119;) (canon lower (func $"[method]outgoing-request.set-method") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.path-with-query" (func $"[method]outgoing-request.path-with-query" (;56;)))
  (core func $"#core-func120 indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query") (;120;) (canon lower (func $"[method]outgoing-request.path-with-query") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-path-with-query" (func $"[method]outgoing-request.set-path-with-query" (;57;)))
  (core func $"#core-func121 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query") (;121;) (canon lower (func $"[method]outgoing-request.set-path-with-query") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.scheme" (func $"[method]outgoing-request.scheme" (;58;)))
  (core func $"#core-func122 indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme") (;122;) (canon lower (func $"[method]outgoing-request.scheme") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-scheme" (func $"[method]outgoing-request.set-scheme" (;59;)))
  (core func $"#core-func123 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme") (;123;) (canon lower (func $"[method]outgoing-request.set-scheme") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.authority" (func $"[method]outgoing-request.authority" (;60;)))
  (core func $"#core-func124 indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority") (;124;) (canon lower (func $"[method]outgoing-request.authority") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-authority" (func $"[method]outgoing-request.set-authority" (;61;)))
  (core func $"#core-func125 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority") (;125;) (canon lower (func $"[method]outgoing-request.set-authority") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.connect-timeout" (func $"[method]request-options.connect-timeout" (;62;)))
  (core func $"#core-func126 indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout") (;126;) (canon lower (func $"[method]request-options.connect-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.first-byte-timeout" (func $"[method]request-options.first-byte-timeout" (;63;)))
  (core func $"#core-func127 indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout") (;127;) (canon lower (func $"[method]request-options.first-byte-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.between-bytes-timeout" (func $"[method]request-options.between-bytes-timeout" (;64;)))
  (core func $"#core-func128 indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout") (;128;) (canon lower (func $"[method]request-options.between-bytes-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]response-outparam.set" (func $"[static]response-outparam.set" (;65;)))
  (core func $"#core-func129 indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (@name "indirect-wasi:http/types@0.2.12-[static]response-outparam.set") (;129;) (canon lower (func $"[static]response-outparam.set") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.consume" (func $"[method]incoming-response.consume" (;66;)))
  (core func $"#core-func130 indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-response.consume") (;130;) (canon lower (func $"[method]incoming-response.consume") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-body.stream" (func $"[method]incoming-body.stream" (;67;)))
  (core func $"#core-func131 indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-body.stream") (;131;) (canon lower (func $"[method]incoming-body.stream") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]future-trailers.get" (func $"[method]future-trailers.get" (;68;)))
  (core func $"#core-func132 indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (@name "indirect-wasi:http/types@0.2.12-[method]future-trailers.get") (;132;) (canon lower (func $"[method]future-trailers.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.body" (func $"[method]outgoing-response.body" (;69;)))
  (core func $"#core-func133 indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-response.body") (;133;) (canon lower (func $"[method]outgoing-response.body") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-body.write" (func $"[method]outgoing-body.write" (;70;)))
  (core func $"#core-func134 indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-body.write") (;134;) (canon lower (func $"[method]outgoing-body.write") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]outgoing-body.finish" (func $"[static]outgoing-body.finish" (;71;)))
  (core func $"#core-func135 indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (@name "indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish") (;135;) (canon lower (func $"[static]outgoing-body.finish") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]future-incoming-response.get" (func $"[method]future-incoming-response.get" (;72;)))
  (core func $"#core-func136 indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (@name "indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get") (;136;) (canon lower (func $"[method]future-incoming-response.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "http-error-code" (func $http-error-code (;73;)))
  (core func $"#core-func137 indirect-wasi:http/types@0.2.12-http-error-code" (@name "indirect-wasi:http/types@0.2.12-http-error-code") (;137;) (canon lower (func $http-error-code) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (core instance $actual (;7;)
    (export "0" (func $"#core-func90 indirect-wasi:io/poll@0.2.12-poll"))
    (export "1" (func $"#core-func91 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "2" (func $"#core-func92 indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "3" (func $"#core-func93 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "4" (func $"#core-func94 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "5" (func $"#core-func95 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "6" (func $"#core-func96 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "7" (func $"#core-func97 indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "8" (func $"#core-func98 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "9" (func $"#core-func99 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "10" (func $"#core-func100 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "11" (func $"#core-func101 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "12" (func $"#core-func102 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "13" (func $"#core-func103 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "14" (func $"#core-func104 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "15" (func $"#core-func105 indirect-wasi:http/types@0.2.12-[static]fields.from-list"))
    (export "16" (func $"#core-func106 indirect-wasi:http/types@0.2.12-[method]fields.get"))
    (export "17" (func $"#core-func107 indirect-wasi:http/types@0.2.12-[method]fields.has"))
    (export "18" (func $"#core-func108 indirect-wasi:http/types@0.2.12-[method]fields.set"))
    (export "19" (func $"#core-func109 indirect-wasi:http/types@0.2.12-[method]fields.delete"))
    (export "20" (func $"#core-func110 indirect-wasi:http/types@0.2.12-[method]fields.append"))
    (export "21" (func $"#core-func111 indirect-wasi:http/types@0.2.12-[method]fields.entries"))
    (export "22" (func $"#core-func112 indirect-wasi:http/types@0.2.12-[method]incoming-request.method"))
    (export "23" (func $"#core-func113 indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query"))
    (export "24" (func $"#core-func114 indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme"))
    (export "25" (func $"#core-func115 indirect-wasi:http/types@0.2.12-[method]incoming-request.authority"))
    (export "26" (func $"#core-func116 indirect-wasi:http/types@0.2.12-[method]incoming-request.consume"))
    (export "27" (func $"#core-func117 indirect-wasi:http/types@0.2.12-[method]outgoing-request.body"))
    (export "28" (func $"#core-func118 indirect-wasi:http/types@0.2.12-[method]outgoing-request.method"))
    (export "29" (func $"#core-func119 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method"))
    (export "30" (func $"#core-func120 indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query"))
    (export "31" (func $"#core-func121 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query"))
    (export "32" (func $"#core-func122 indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme"))
    (export "33" (func $"#core-func123 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme"))
    (export "34" (func $"#core-func124 indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority"))
    (export "35" (func $"#core-func125 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority"))
    (export "36" (func $"#core-func126 indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout"))
    (export "37" (func $"#core-func127 indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout"))
    (export "38" (func $"#core-func128 indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout"))
    (export "39" (func $"#core-func129 indirect-wasi:http/types@0.2.12-[static]response-outparam.set"))
    (export "40" (func $"#core-func130 indirect-wasi:http/types@0.2.12-[method]incoming-response.consume"))
    (export "41" (func $"#core-func131 indirect-wasi:http/types@0.2.12-[method]incoming-body.stream"))
    (export "42" (func $"#core-func132 indirect-wasi:http/types@0.2.12-[method]future-trailers.get"))
    (export "43" (func $"#core-func133 indirect-wasi:http/types@0.2.12-[method]outgoing-response.body"))
    (export "44" (func $"#core-func134 indirect-wasi:http/types@0.2.12-[method]outgoing-body.write"))
    (export "45" (func $"#core-func135 indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish"))
    (export "46" (func $"#core-func136 indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get"))
    (export "47" (func $"#core-func137 indirect-wasi:http/types@0.2.12-http-error-code"))
  )
  (core instance $fixup (;8;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;28;) (func (result u32)))
  (alias core export $main "probe" (core func $probe (;138;)))
  (alias core export $main "cabi_post_probe" (core func $cabi_post_probe (;139;)))
  (func $probe (;74;) (type 28) (canon lift (core func $probe) (post-return $cabi_post_probe)))
  (export $"#func75 probe" (@name "probe") (;75;) "probe" (func $probe))
  (type (;29;) (func (param "count" u32) (result u32)))
  (alias core export $main "fields-batch" (core func $fields-batch (;140;)))
  (alias core export $main "cabi_post_fields-batch" (core func $cabi_post_fields-batch (;141;)))
  (func $fields-batch (;76;) (type 29) (canon lift (core func $fields-batch) (post-return $cabi_post_fields-batch)))
  (export $"#func77 fields-batch" (@name "fields-batch") (;77;) "fields-batch" (func $fields-batch))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
 (export "types" (instance 4))
)

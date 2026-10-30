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
  (type $ty-wasi:clocks/wall-clock@0.2.12 (;3;)
    (instance
      (type (;0;) (record (field "seconds" u64) (field "nanoseconds" u32)))
      (export (;1;) "datetime" (type (eq 0)))
      (type (;2;) (func (result 1)))
      (export (;0;) "now" (func (type 2)))
      (export (;1;) "resolution" (func (type 2)))
    )
  )
  (import "wasi:clocks/wall-clock@0.2.12" (instance $wasi:clocks/wall-clock@0.2.12 (;2;) (type $ty-wasi:clocks/wall-clock@0.2.12)))
  (type $ty-wasi:random/random@0.2.12 (;4;)
    (instance
      (type (;0;) (list u8))
      (type (;1;) (func (param "len" u64) (result 0)))
      (export (;0;) "get-random-bytes" (func (type 1)))
      (type (;2;) (func (result u64)))
      (export (;1;) "get-random-u64" (func (type 2)))
    )
  )
  (import "wasi:random/random@0.2.12" (instance $wasi:random/random@0.2.12 (;3;) (type $ty-wasi:random/random@0.2.12)))
  (type $ty-wasi:io/error@0.2.12 (;5;)
    (instance
      (export (;0;) "error" (type (sub resource)))
      (type (;1;) (borrow 0))
      (type (;2;) (func (param "self" 1) (result string)))
      (export (;0;) "[method]error.to-debug-string" (func (type 2)))
    )
  )
  (import "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12 (;4;) (type $ty-wasi:io/error@0.2.12)))
  (alias export $wasi:io/error@0.2.12 "error" (type $error (;6;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type7 pollable" (@name "pollable") (;7;)))
  (type $ty-wasi:io/streams@0.2.12 (;8;)
    (instance
      (export (;0;) "input-stream" (type (sub resource)))
      (alias outer 1 $error (type (;1;)))
      (export (;2;) "error" (type (eq 1)))
      (type (;3;) (own 2))
      (type (;4;) (variant (case "last-operation-failed" 3) (case "closed")))
      (export (;5;) "stream-error" (type (eq 4)))
      (alias outer 1 $"#type7 pollable" (type (;6;)))
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
  (import "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12 (;5;) (type $ty-wasi:io/streams@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $output-stream (;9;)))
  (type $ty-wasi:cli/stdout@0.2.12 (;10;)
    (instance
      (alias outer 1 $output-stream (type (;0;)))
      (export (;1;) "output-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stdout" (func (type 3)))
    )
  )
  (import "wasi:cli/stdout@0.2.12" (instance $wasi:cli/stdout@0.2.12 (;6;) (type $ty-wasi:cli/stdout@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type11 output-stream" (@name "output-stream") (;11;)))
  (type $ty-wasi:cli/stderr@0.2.12 (;12;)
    (instance
      (alias outer 1 $"#type11 output-stream" (type (;0;)))
      (export (;1;) "output-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stderr" (func (type 3)))
    )
  )
  (import "wasi:cli/stderr@0.2.12" (instance $wasi:cli/stderr@0.2.12 (;7;) (type $ty-wasi:cli/stderr@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $input-stream (;13;)))
  (type $ty-wasi:cli/stdin@0.2.12 (;14;)
    (instance
      (alias outer 1 $input-stream (type (;0;)))
      (export (;1;) "input-stream" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (func (result 2)))
      (export (;0;) "get-stdin" (func (type 3)))
    )
  )
  (import "wasi:cli/stdin@0.2.12" (instance $wasi:cli/stdin@0.2.12 (;8;) (type $ty-wasi:cli/stdin@0.2.12)))
  (alias export $wasi:clocks/monotonic-clock@0.2.12 "duration" (type $duration (;15;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type16 input-stream" (@name "input-stream") (;16;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type17 pollable" (@name "pollable") (;17;)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type18 output-stream" (@name "output-stream") (;18;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type19 error" (@name "error") (;19;)))
  (type $ty-wasi:http/types@0.2.12 (;20;)
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
      (alias outer 1 $"#type16 input-stream" (type (;39;)))
      (export (;40;) "input-stream" (type (eq 39)))
      (export (;41;) "future-trailers" (type (sub resource)))
      (alias outer 1 $"#type17 pollable" (type (;42;)))
      (export (;43;) "pollable" (type (eq 42)))
      (export (;44;) "trailers" (type (eq 0)))
      (alias outer 1 $"#type18 output-stream" (type (;45;)))
      (export (;46;) "output-stream" (type (eq 45)))
      (export (;47;) "future-incoming-response" (type (sub resource)))
      (alias outer 1 $"#type19 error" (type (;48;)))
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
  (import "wasi:http/types@0.2.12" (instance $wasi:http/types@0.2.12 (;9;) (type $ty-wasi:http/types@0.2.12)))
  (alias export $wasi:http/types@0.2.12 "outgoing-request" (type $outgoing-request (;21;)))
  (alias export $wasi:http/types@0.2.12 "request-options" (type $request-options (;22;)))
  (alias export $wasi:http/types@0.2.12 "future-incoming-response" (type $future-incoming-response (;23;)))
  (alias export $wasi:http/types@0.2.12 "error-code" (type $error-code (;24;)))
  (type $ty-wasi:http/outgoing-handler@0.2.12 (;25;)
    (instance
      (alias outer 1 $outgoing-request (type (;0;)))
      (export (;1;) "outgoing-request" (type (eq 0)))
      (alias outer 1 $request-options (type (;2;)))
      (export (;3;) "request-options" (type (eq 2)))
      (alias outer 1 $future-incoming-response (type (;4;)))
      (export (;5;) "future-incoming-response" (type (eq 4)))
      (alias outer 1 $error-code (type (;6;)))
      (export (;7;) "error-code" (type (eq 6)))
      (type (;8;) (own 1))
      (type (;9;) (own 3))
      (type (;10;) (option 9))
      (type (;11;) (own 5))
      (type (;12;) (result 11 (error 7)))
      (type (;13;) (func (param "request" 8) (param "options" 10) (result 12)))
      (export (;0;) "handle" (func (type 13)))
    )
  )
  (import "wasi:http/outgoing-handler@0.2.12" (instance $wasi:http/outgoing-handler@0.2.12 (;10;) (type $ty-wasi:http/outgoing-handler@0.2.12)))
  (core module $main (;0;)
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
    (global $heap (;0;) (mut i32) i32.const 8192)
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
      local.get 0
      local.get 1
      local.get 2
      call $request
    )
    (func (;98;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;99;) (type 11) (param i32 i32 i32) (result i32)
      (local $at i32)
      local.get 2
      i32.const 128
      i32.gt_u
      if ;; label = @1
        unreachable
      end
      block $end
        loop $next
          local.get $at
          local.get 2
          i32.ge_u
          br_if $end
          local.get 0
          local.get 1
          i32.const 0
          call $request
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
    (func (;100;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;101;) (type 6) (param i32 i32)
      (local $headers i32) (local $body i32) (local $response i32) (local $output i32)
      local.get 0
      i32.const 64
      call $types.incoming-request.method
      local.get 0
      i32.const 64
      call $types.incoming-request.path-with-query
      local.get 0
      i32.const 64
      call $types.incoming-request.scheme
      local.get 0
      i32.const 64
      call $types.incoming-request.authority
      local.get 0
      call $types.incoming-request.headers
      local.tee $headers
      call $types.drop-fields
      local.get 0
      i32.const 64
      call $types.incoming-request.consume
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $body
      local.get 0
      call $types.drop-incoming-request
      local.get $body
      call $read-body
      i32.const 4096
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $body
      call $types.incoming-body.finish
      i32.const 1072
      i32.const 17
      call $trailers
      call $types.new-fields
      call $types.new-outgoing-response
      local.set $response
      local.get $response
      i32.const 200
      call $types.outgoing-response.set-status-code
      if ;; label = @1
        unreachable
      end
      local.get $response
      call $types.outgoing-response.status-code
      i32.const 200
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $response
      call $types.outgoing-response.headers
      local.tee $headers
      call $types.drop-fields
      local.get $response
      i32.const 64
      call $types.outgoing-response.body
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $body
      local.get 1
      i32.const 0
      local.get $response
      i32.const 0
      i64.const 0
      i32.const 0
      i32.const 0
      i32.const 0
      i32.const 0
      call $types.response-outparam.set
      local.get $body
      i32.const 64
      call $types.outgoing-body.write
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $output
      local.get $output
      i32.const 2048
      i32.const 4096
      i32.const 64
      call $streams.output-stream.blocking-write-and-flush
      i32.const 64
      call $ok
      local.get $output
      call $streams.drop-output-stream
      local.get $body
      i32.const 1
      i32.const 1120
      i32.const 15
      call $new-trailers
      i32.const 64
      call $types.outgoing-body.finish
      i32.const 64
      call $ok
    )
    (func (;102;) (type 18)
      i32.const 8192
      global.set $heap
    )
    (func (;103;) (type 13) (param i32 i32 i32 i32) (result i32)
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
    (func (;104;) (type 18))
    (func $ok (;105;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (func $trailers (;106;) (type 2) (param $future i32) (param $name i32) (param $len i32)
      (local $poll i32) (local $headers i32)
      block $ready
        loop $next
          local.get $future
          i32.const 64
          call $types.future-trailers.get
          i32.const 64
          i32.load8_u
          br_if $ready
          local.get $future
          call $types.future-trailers.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $next
        end
      end
      i32.const 72
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 80
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 88
      i32.load8_u
      if ;; label = @1
        i32.const 92
        i32.load
        local.set $headers
        local.get $headers
        local.get $name
        local.get $len
        i32.const 64
        call $types.fields.get
        i32.const 68
        i32.load
        i32.const 1
        i32.ne
        if ;; label = @2
          unreachable
        end
        i32.const 64
        i32.load
        i32.const 4
        i32.add
        i32.load
        i32.const 4
        i32.ne
        if ;; label = @2
          unreachable
        end
        local.get $headers
        call $types.drop-fields
      end
      local.get $future
      call $types.drop-future-trailers
    )
    (func $read-body (;107;) (type 0) (param $body i32) (result i32)
      (local $stream i32) (local $at i32) (local $n i32) (local $ptr i32)
      local.get $body
      i32.const 64
      call $types.incoming-body.stream
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      block $end
        loop $next
          local.get $stream
          i64.const 4096
          i32.const 64
          call $streams.input-stream.blocking-read
          i32.const 64
          i32.load8_u
          if ;; label = @3
            i32.const 68
            i32.load8_u
            i32.const 1
            i32.ne
            if ;; label = @4
              unreachable
            end
            br $end
          end
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
          i32.const 2048
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
          i32.const 2048
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
          local.tee $at
          i32.const 65536
          i32.gt_u
          if ;; label = @3
            unreachable
          end
          br $next
        end
      end
      local.get $stream
      call $streams.drop-input-stream
      local.get $at
    )
    (func $new-trailers (;108;) (type 17) (param $name i32) (param $len i32) (result i32)
      (local $fields i32)
      call $types.new-fields
      local.tee $fields
      local.get $name
      local.get $len
      i32.const 1104
      i32.const 4
      i32.const 64
      call $types.fields.append
      i32.const 64
      call $ok
      local.get $fields
    )
    (func $request (;109;) (type 11) (param $authority i32) (param $alen i32) (param $secure i32) (result i32)
      (local $headers i32) (local $request i32) (local $body i32) (local $stream i32) (local $options i32) (local $future i32) (local $poll i32) (local $response i32) (local $incoming i32)
      call $types.new-fields
      local.tee $headers
      i32.const 1024
      i32.const 8
      i32.const 1040
      i32.const 5
      i32.const 64
      call $types.fields.append
      i32.const 64
      call $ok
      local.get $headers
      i32.const 1144
      i32.const 7
      i32.const 1072
      i32.const 17
      i32.const 64
      call $types.fields.append
      i32.const 64
      call $ok
      local.get $headers
      call $types.new-outgoing-request
      local.set $request
      local.get $request
      i32.const 2
      i32.const 0
      i32.const 0
      call $types.outgoing-request.set-method
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 64
      call $types.outgoing-request.method
      i32.const 64
      i32.load8_u
      i32.const 2
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 1
      i32.const 1056
      i32.const 5
      call $types.outgoing-request.set-path-with-query
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 1
      local.get $secure
      i32.const 0
      i32.const 0
      call $types.outgoing-request.set-scheme
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 1
      local.get $authority
      local.get $alen
      call $types.outgoing-request.set-authority
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 64
      call $types.outgoing-request.path-with-query
      i32.const 64
      i32.load8_u
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 64
      call $types.outgoing-request.scheme
      i32.const 64
      i32.load8_u
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 64
      call $types.outgoing-request.authority
      i32.const 64
      i32.load8_u
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $request
      call $types.outgoing-request.headers
      local.set $headers
      local.get $headers
      i32.const 1024
      i32.const 8
      i32.const 1040
      i32.const 5
      i32.const 64
      call $types.fields.append
      i32.const 64
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 65
      i32.load8_u
      i32.const 2
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $headers
      call $types.drop-fields
      local.get $request
      i32.const 64
      call $types.outgoing-request.body
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $body
      local.get $body
      i32.const 64
      call $types.outgoing-body.write
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      i32.const 2048
      i32.const 4096
      i32.const 64
      call $streams.output-stream.blocking-write-and-flush
      i32.const 64
      call $ok
      local.get $stream
      call $streams.drop-output-stream
      local.get $body
      i32.const 1
      i32.const 1072
      i32.const 17
      call $new-trailers
      i32.const 64
      call $types.outgoing-body.finish
      i32.const 64
      call $ok
      call $types.new-request-options
      local.set $options
      local.get $options
      i32.const 64
      call $types.request-options.connect-timeout
      local.get $options
      i32.const 1
      i64.const 1000000000
      call $types.request-options.set-connect-timeout
      if ;; label = @1
        unreachable
      end
      local.get $options
      i32.const 1
      i64.const 1000000000
      call $types.request-options.set-first-byte-timeout
      if ;; label = @1
        unreachable
      end
      local.get $options
      i32.const 1
      i64.const 1000000000
      call $types.request-options.set-between-bytes-timeout
      if ;; label = @1
        unreachable
      end
      local.get $options
      i32.const 64
      call $types.request-options.first-byte-timeout
      i32.const 64
      i32.load8_u
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $options
      i32.const 64
      call $types.request-options.between-bytes-timeout
      i32.const 64
      i32.load8_u
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $request
      i32.const 1
      local.get $options
      i32.const 64
      call $outgoing-handler.handle
      i32.const 64
      call $ok
      i32.const 72
      i32.load
      local.set $future
      block $ready
        loop $next
          local.get $future
          i32.const 64
          call $types.future-incoming-response.get
          i32.const 64
          i32.load8_u
          br_if $ready
          local.get $future
          call $types.future-incoming-response.subscribe
          local.tee $poll
          call $poll.pollable.block
          local.get $poll
          call $poll.drop-pollable
          br $next
        end
      end
      i32.const 72
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 80
      i32.load8_u
      if ;; label = @1
        unreachable
      end
      i32.const 88
      i32.load
      local.set $response
      local.get $future
      call $types.drop-future-incoming-response
      local.get $response
      call $types.incoming-response.status
      i32.const 200
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $response
      call $types.incoming-response.headers
      local.tee $headers
      call $types.drop-fields
      local.get $response
      i32.const 64
      call $types.incoming-response.consume
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $incoming
      local.get $response
      call $types.drop-incoming-response
      local.get $incoming
      call $read-body
      i32.const 4096
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $incoming
      call $types.incoming-body.finish
      i32.const 1120
      i32.const 15
      call $trailers
      i32.const 42
    )
    (data (;0;) (i32.const 1024) "x-wasmvm")
    (data (;1;) (i32.const 1040) "value")
    (data (;2;) (i32.const 1056) "/echo")
    (data (;3;) (i32.const 1072) "x-request-trailer")
    (data (;4;) (i32.const 1144) "trailer")
    (data (;5;) (i32.const 1104) "done")
    (data (;6;) (i32.const 1120) "x-reply-trailer")
    (data (;7;) (i32.const 2048) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core module $wit-component-shim-module (;1;)
    (type (;0;) (func (param i32 i32 i32)))
    (type (;1;) (func (param i32)))
    (type (;2;) (func (param i64 i32)))
    (type (;3;) (func (param i32 i32)))
    (type (;4;) (func (param i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32 i32)))
    (type (;6;) (func (param i32 i32 i64 i32)))
    (type (;7;) (func (param i32 i32 i32) (result i32)))
    (type (;8;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;9;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;10;) (func (param i32 i32 i32 i32 i32) (result i32)))
    (type (;11;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
    (table (;0;) 52 52 funcref)
    (export "0" (func $indirect-wasi:io/poll@0.2.12-poll))
    (export "1" (func $indirect-wasi:clocks/wall-clock@0.2.12-now))
    (export "2" (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution))
    (export "3" (func $indirect-wasi:random/random@0.2.12-get-random-bytes))
    (export "4" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "5" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "6" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "7" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "8" (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "9" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "10" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "11" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "12" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "13" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "14" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "15" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "16" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "17" (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "18" (func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list"))
    (export "19" (func $"indirect-wasi:http/types@0.2.12-[method]fields.get"))
    (export "20" (func $"indirect-wasi:http/types@0.2.12-[method]fields.has"))
    (export "21" (func $"indirect-wasi:http/types@0.2.12-[method]fields.set"))
    (export "22" (func $"indirect-wasi:http/types@0.2.12-[method]fields.delete"))
    (export "23" (func $"indirect-wasi:http/types@0.2.12-[method]fields.append"))
    (export "24" (func $"indirect-wasi:http/types@0.2.12-[method]fields.entries"))
    (export "25" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method"))
    (export "26" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query"))
    (export "27" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme"))
    (export "28" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority"))
    (export "29" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume"))
    (export "30" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body"))
    (export "31" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method"))
    (export "32" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method"))
    (export "33" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query"))
    (export "34" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query"))
    (export "35" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme"))
    (export "36" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme"))
    (export "37" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority"))
    (export "38" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority"))
    (export "39" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout"))
    (export "40" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout"))
    (export "41" (func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout"))
    (export "42" (func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set"))
    (export "43" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume"))
    (export "44" (func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream"))
    (export "45" (func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get"))
    (export "46" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body"))
    (export "47" (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write"))
    (export "48" (func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish"))
    (export "49" (func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get"))
    (export "50" (func $indirect-wasi:http/types@0.2.12-http-error-code))
    (export "51" (func $indirect-wasi:http/outgoing-handler@0.2.12-handle))
    (export "$imports" (table 0))
    (func $indirect-wasi:io/poll@0.2.12-poll (;0;) (type 0) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 0
      call_indirect (type 0)
    )
    (func $indirect-wasi:clocks/wall-clock@0.2.12-now (;1;) (type 1) (param i32)
      local.get 0
      i32.const 1
      call_indirect (type 1)
    )
    (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution (;2;) (type 1) (param i32)
      local.get 0
      i32.const 2
      call_indirect (type 1)
    )
    (func $indirect-wasi:random/random@0.2.12-get-random-bytes (;3;) (type 2) (param i64 i32)
      local.get 0
      local.get 1
      i32.const 3
      call_indirect (type 2)
    )
    (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;4;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 4
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (;5;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 5
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (;6;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 6
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (;7;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 7
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (;8;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 8
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;9;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 9
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (;10;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 10
      call_indirect (type 5)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (;11;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 11
      call_indirect (type 5)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;12;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 12
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;13;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 13
      call_indirect (type 3)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (;14;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 14
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (;15;) (type 4) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 15
      call_indirect (type 4)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (;16;) (type 6) (param i32 i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 16
      call_indirect (type 6)
    )
    (func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (;17;) (type 6) (param i32 i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 17
      call_indirect (type 6)
    )
    (func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list" (;18;) (type 0) (param i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 18
      call_indirect (type 0)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.get" (;19;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 19
      call_indirect (type 5)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.has" (;20;) (type 7) (param i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 20
      call_indirect (type 7)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.set" (;21;) (type 8) (param i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      i32.const 21
      call_indirect (type 8)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.delete" (;22;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 22
      call_indirect (type 5)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.append" (;23;) (type 8) (param i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      i32.const 23
      call_indirect (type 8)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]fields.entries" (;24;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 24
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (;25;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 25
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (;26;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 26
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (;27;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 27
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (;28;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 28
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (;29;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 29
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (;30;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 30
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (;31;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 31
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (;32;) (type 9) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 32
      call_indirect (type 9)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (;33;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 33
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (;34;) (type 9) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 34
      call_indirect (type 9)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (;35;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 35
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (;36;) (type 10) (param i32 i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 36
      call_indirect (type 10)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (;37;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 37
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (;38;) (type 9) (param i32 i32 i32 i32) (result i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 38
      call_indirect (type 9)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (;39;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 39
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (;40;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 40
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (;41;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 41
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (;42;) (type 11) (param i32 i32 i32 i32 i64 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      local.get 8
      i32.const 42
      call_indirect (type 11)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (;43;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 43
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (;44;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 44
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (;45;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 45
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (;46;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 46
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (;47;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 47
      call_indirect (type 3)
    )
    (func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (;48;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 48
      call_indirect (type 5)
    )
    (func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (;49;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 49
      call_indirect (type 3)
    )
    (func $indirect-wasi:http/types@0.2.12-http-error-code (;50;) (type 3) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 50
      call_indirect (type 3)
    )
    (func $indirect-wasi:http/outgoing-handler@0.2.12-handle (;51;) (type 5) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 51
      call_indirect (type 5)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias export $wasi:http/types@0.2.12 "incoming-request" (type $incoming-request (;26;)))
  (alias export $wasi:http/types@0.2.12 "response-outparam" (type $response-outparam (;27;)))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;0;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;1;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "0" (core func $indirect-wasi:io/poll@0.2.12-poll (;2;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type28 pollable" (@name "pollable") (;28;)))
  (core func $resource.drop (;3;) (canon resource.drop $"#type28 pollable"))
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
  (alias core export $wit-component-shim-instance "1" (core func $indirect-wasi:clocks/wall-clock@0.2.12-now (;8;)))
  (alias core export $wit-component-shim-instance "2" (core func $indirect-wasi:clocks/wall-clock@0.2.12-resolution (;9;)))
  (core instance $wasi:clocks/wall-clock@0.2.12 (;3;)
    (export "now" (func $indirect-wasi:clocks/wall-clock@0.2.12-now))
    (export "resolution" (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution))
  )
  (alias core export $wit-component-shim-instance "3" (core func $indirect-wasi:random/random@0.2.12-get-random-bytes (;10;)))
  (alias export $wasi:random/random@0.2.12 "get-random-u64" (func $get-random-u64 (;6;)))
  (core func $get-random-u64 (;11;) (canon lower (func $get-random-u64)))
  (core instance $wasi:random/random@0.2.12 (;4;)
    (export "get-random-bytes" (func $indirect-wasi:random/random@0.2.12-get-random-bytes))
    (export "get-random-u64" (func $get-random-u64))
  )
  (alias core export $wit-component-shim-instance "4" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;12;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type29 error" (@name "error") (;29;)))
  (core func $"#core-func13 resource.drop" (@name "resource.drop") (;13;) (canon resource.drop $"#type29 error"))
  (core instance $wasi:io/error@0.2.12 (;5;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $"#core-func13 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "5" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (;14;)))
  (alias core export $wit-component-shim-instance "6" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (;15;)))
  (alias core export $wit-component-shim-instance "7" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (;16;)))
  (alias core export $wit-component-shim-instance "8" (core func $"indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (;17;)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.subscribe" (func $"[method]input-stream.subscribe" (;7;)))
  (core func $"[method]input-stream.subscribe" (;18;) (canon lower (func $"[method]input-stream.subscribe")))
  (alias core export $wit-component-shim-instance "9" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (;19;)))
  (alias core export $wit-component-shim-instance "10" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (;20;)))
  (alias core export $wit-component-shim-instance "11" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (;21;)))
  (alias core export $wit-component-shim-instance "12" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (;22;)))
  (alias core export $wit-component-shim-instance "13" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (;23;)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.subscribe" (func $"[method]output-stream.subscribe" (;8;)))
  (core func $"[method]output-stream.subscribe" (;24;) (canon lower (func $"[method]output-stream.subscribe")))
  (alias core export $wit-component-shim-instance "14" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (;25;)))
  (alias core export $wit-component-shim-instance "15" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (;26;)))
  (alias core export $wit-component-shim-instance "16" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (;27;)))
  (alias core export $wit-component-shim-instance "17" (core func $"indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (;28;)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type30 input-stream" (@name "input-stream") (;30;)))
  (core func $"#core-func29 resource.drop" (@name "resource.drop") (;29;) (canon resource.drop $"#type30 input-stream"))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type31 output-stream" (@name "output-stream") (;31;)))
  (core func $"#core-func30 resource.drop" (@name "resource.drop") (;30;) (canon resource.drop $"#type31 output-stream"))
  (core instance $wasi:io/streams@0.2.12 (;6;)
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
    (export "[resource-drop]input-stream" (func $"#core-func29 resource.drop"))
    (export "[resource-drop]output-stream" (func $"#core-func30 resource.drop"))
  )
  (alias export $wasi:cli/stdout@0.2.12 "get-stdout" (func $get-stdout (;9;)))
  (core func $get-stdout (;31;) (canon lower (func $get-stdout)))
  (core instance $wasi:cli/stdout@0.2.12 (;7;)
    (export "get-stdout" (func $get-stdout))
  )
  (alias export $wasi:cli/stderr@0.2.12 "get-stderr" (func $get-stderr (;10;)))
  (core func $get-stderr (;32;) (canon lower (func $get-stderr)))
  (core instance $wasi:cli/stderr@0.2.12 (;8;)
    (export "get-stderr" (func $get-stderr))
  )
  (alias export $wasi:cli/stdin@0.2.12 "get-stdin" (func $get-stdin (;11;)))
  (core func $get-stdin (;33;) (canon lower (func $get-stdin)))
  (core instance $wasi:cli/stdin@0.2.12 (;9;)
    (export "get-stdin" (func $get-stdin))
  )
  (alias export $wasi:http/types@0.2.12 "[constructor]fields" (func $"[constructor]fields" (;12;)))
  (core func $"[constructor]fields" (;34;) (canon lower (func $"[constructor]fields")))
  (alias core export $wit-component-shim-instance "18" (core func $"indirect-wasi:http/types@0.2.12-[static]fields.from-list" (;35;)))
  (alias core export $wit-component-shim-instance "19" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.get" (;36;)))
  (alias core export $wit-component-shim-instance "20" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.has" (;37;)))
  (alias core export $wit-component-shim-instance "21" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.set" (;38;)))
  (alias core export $wit-component-shim-instance "22" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.delete" (;39;)))
  (alias core export $wit-component-shim-instance "23" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.append" (;40;)))
  (alias core export $wit-component-shim-instance "24" (core func $"indirect-wasi:http/types@0.2.12-[method]fields.entries" (;41;)))
  (alias export $wasi:http/types@0.2.12 "[method]fields.clone" (func $"[method]fields.clone" (;13;)))
  (core func $"[method]fields.clone" (;42;) (canon lower (func $"[method]fields.clone")))
  (alias core export $wit-component-shim-instance "25" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (;43;)))
  (alias core export $wit-component-shim-instance "26" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (;44;)))
  (alias core export $wit-component-shim-instance "27" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (;45;)))
  (alias core export $wit-component-shim-instance "28" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (;46;)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.headers" (func $"[method]incoming-request.headers" (;14;)))
  (core func $"[method]incoming-request.headers" (;47;) (canon lower (func $"[method]incoming-request.headers")))
  (alias core export $wit-component-shim-instance "29" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (;48;)))
  (alias export $wasi:http/types@0.2.12 "[constructor]outgoing-request" (func $"[constructor]outgoing-request" (;15;)))
  (core func $"[constructor]outgoing-request" (;49;) (canon lower (func $"[constructor]outgoing-request")))
  (alias core export $wit-component-shim-instance "30" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (;50;)))
  (alias core export $wit-component-shim-instance "31" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (;51;)))
  (alias core export $wit-component-shim-instance "32" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (;52;)))
  (alias core export $wit-component-shim-instance "33" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (;53;)))
  (alias core export $wit-component-shim-instance "34" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (;54;)))
  (alias core export $wit-component-shim-instance "35" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (;55;)))
  (alias core export $wit-component-shim-instance "36" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (;56;)))
  (alias core export $wit-component-shim-instance "37" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (;57;)))
  (alias core export $wit-component-shim-instance "38" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (;58;)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.headers" (func $"[method]outgoing-request.headers" (;16;)))
  (core func $"[method]outgoing-request.headers" (;59;) (canon lower (func $"[method]outgoing-request.headers")))
  (alias export $wasi:http/types@0.2.12 "[constructor]request-options" (func $"[constructor]request-options" (;17;)))
  (core func $"[constructor]request-options" (;60;) (canon lower (func $"[constructor]request-options")))
  (alias core export $wit-component-shim-instance "39" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (;61;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-connect-timeout" (func $"[method]request-options.set-connect-timeout" (;18;)))
  (core func $"[method]request-options.set-connect-timeout" (;62;) (canon lower (func $"[method]request-options.set-connect-timeout")))
  (alias core export $wit-component-shim-instance "40" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (;63;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-first-byte-timeout" (func $"[method]request-options.set-first-byte-timeout" (;19;)))
  (core func $"[method]request-options.set-first-byte-timeout" (;64;) (canon lower (func $"[method]request-options.set-first-byte-timeout")))
  (alias core export $wit-component-shim-instance "41" (core func $"indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (;65;)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.set-between-bytes-timeout" (func $"[method]request-options.set-between-bytes-timeout" (;20;)))
  (core func $"[method]request-options.set-between-bytes-timeout" (;66;) (canon lower (func $"[method]request-options.set-between-bytes-timeout")))
  (alias core export $wit-component-shim-instance "42" (core func $"indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (;67;)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.status" (func $"[method]incoming-response.status" (;21;)))
  (core func $"[method]incoming-response.status" (;68;) (canon lower (func $"[method]incoming-response.status")))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.headers" (func $"[method]incoming-response.headers" (;22;)))
  (core func $"[method]incoming-response.headers" (;69;) (canon lower (func $"[method]incoming-response.headers")))
  (alias core export $wit-component-shim-instance "43" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (;70;)))
  (alias core export $wit-component-shim-instance "44" (core func $"indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (;71;)))
  (alias export $wasi:http/types@0.2.12 "[static]incoming-body.finish" (func $"[static]incoming-body.finish" (;23;)))
  (core func $"[static]incoming-body.finish" (;72;) (canon lower (func $"[static]incoming-body.finish")))
  (alias export $wasi:http/types@0.2.12 "[method]future-trailers.subscribe" (func $"[method]future-trailers.subscribe" (;24;)))
  (core func $"[method]future-trailers.subscribe" (;73;) (canon lower (func $"[method]future-trailers.subscribe")))
  (alias core export $wit-component-shim-instance "45" (core func $"indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (;74;)))
  (alias export $wasi:http/types@0.2.12 "[constructor]outgoing-response" (func $"[constructor]outgoing-response" (;25;)))
  (core func $"[constructor]outgoing-response" (;75;) (canon lower (func $"[constructor]outgoing-response")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.status-code" (func $"[method]outgoing-response.status-code" (;26;)))
  (core func $"[method]outgoing-response.status-code" (;76;) (canon lower (func $"[method]outgoing-response.status-code")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.set-status-code" (func $"[method]outgoing-response.set-status-code" (;27;)))
  (core func $"[method]outgoing-response.set-status-code" (;77;) (canon lower (func $"[method]outgoing-response.set-status-code")))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.headers" (func $"[method]outgoing-response.headers" (;28;)))
  (core func $"[method]outgoing-response.headers" (;78;) (canon lower (func $"[method]outgoing-response.headers")))
  (alias core export $wit-component-shim-instance "46" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (;79;)))
  (alias core export $wit-component-shim-instance "47" (core func $"indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (;80;)))
  (alias core export $wit-component-shim-instance "48" (core func $"indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (;81;)))
  (alias export $wasi:http/types@0.2.12 "[method]future-incoming-response.subscribe" (func $"[method]future-incoming-response.subscribe" (;29;)))
  (core func $"[method]future-incoming-response.subscribe" (;82;) (canon lower (func $"[method]future-incoming-response.subscribe")))
  (alias core export $wit-component-shim-instance "49" (core func $"indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (;83;)))
  (alias core export $wit-component-shim-instance "50" (core func $indirect-wasi:http/types@0.2.12-http-error-code (;84;)))
  (alias export $wasi:http/types@0.2.12 "fields" (type $fields (;32;)))
  (core func $"#core-func85 resource.drop" (@name "resource.drop") (;85;) (canon resource.drop $fields))
  (alias export $wasi:http/types@0.2.12 "incoming-request" (type $"#type33 incoming-request" (@name "incoming-request") (;33;)))
  (core func $"#core-func86 resource.drop" (@name "resource.drop") (;86;) (canon resource.drop $"#type33 incoming-request"))
  (alias export $wasi:http/types@0.2.12 "outgoing-request" (type $"#type34 outgoing-request" (@name "outgoing-request") (;34;)))
  (core func $"#core-func87 resource.drop" (@name "resource.drop") (;87;) (canon resource.drop $"#type34 outgoing-request"))
  (alias export $wasi:http/types@0.2.12 "request-options" (type $"#type35 request-options" (@name "request-options") (;35;)))
  (core func $"#core-func88 resource.drop" (@name "resource.drop") (;88;) (canon resource.drop $"#type35 request-options"))
  (alias export $wasi:http/types@0.2.12 "response-outparam" (type $"#type36 response-outparam" (@name "response-outparam") (;36;)))
  (core func $"#core-func89 resource.drop" (@name "resource.drop") (;89;) (canon resource.drop $"#type36 response-outparam"))
  (alias export $wasi:http/types@0.2.12 "incoming-response" (type $incoming-response (;37;)))
  (core func $"#core-func90 resource.drop" (@name "resource.drop") (;90;) (canon resource.drop $incoming-response))
  (alias export $wasi:http/types@0.2.12 "incoming-body" (type $incoming-body (;38;)))
  (core func $"#core-func91 resource.drop" (@name "resource.drop") (;91;) (canon resource.drop $incoming-body))
  (alias export $wasi:http/types@0.2.12 "future-trailers" (type $future-trailers (;39;)))
  (core func $"#core-func92 resource.drop" (@name "resource.drop") (;92;) (canon resource.drop $future-trailers))
  (alias export $wasi:http/types@0.2.12 "outgoing-response" (type $outgoing-response (;40;)))
  (core func $"#core-func93 resource.drop" (@name "resource.drop") (;93;) (canon resource.drop $outgoing-response))
  (alias export $wasi:http/types@0.2.12 "outgoing-body" (type $outgoing-body (;41;)))
  (core func $"#core-func94 resource.drop" (@name "resource.drop") (;94;) (canon resource.drop $outgoing-body))
  (alias export $wasi:http/types@0.2.12 "future-incoming-response" (type $"#type42 future-incoming-response" (@name "future-incoming-response") (;42;)))
  (core func $"#core-func95 resource.drop" (@name "resource.drop") (;95;) (canon resource.drop $"#type42 future-incoming-response"))
  (core instance $wasi:http/types@0.2.12 (;10;)
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
    (export "[resource-drop]fields" (func $"#core-func85 resource.drop"))
    (export "[resource-drop]incoming-request" (func $"#core-func86 resource.drop"))
    (export "[resource-drop]outgoing-request" (func $"#core-func87 resource.drop"))
    (export "[resource-drop]request-options" (func $"#core-func88 resource.drop"))
    (export "[resource-drop]response-outparam" (func $"#core-func89 resource.drop"))
    (export "[resource-drop]incoming-response" (func $"#core-func90 resource.drop"))
    (export "[resource-drop]incoming-body" (func $"#core-func91 resource.drop"))
    (export "[resource-drop]future-trailers" (func $"#core-func92 resource.drop"))
    (export "[resource-drop]outgoing-response" (func $"#core-func93 resource.drop"))
    (export "[resource-drop]outgoing-body" (func $"#core-func94 resource.drop"))
    (export "[resource-drop]future-incoming-response" (func $"#core-func95 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "51" (core func $indirect-wasi:http/outgoing-handler@0.2.12-handle (;96;)))
  (core instance $wasi:http/outgoing-handler@0.2.12 (;11;)
    (export "handle" (func $indirect-wasi:http/outgoing-handler@0.2.12-handle))
  )
  (core instance $main (;12;) (instantiate $main
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:clocks/monotonic-clock@0.2.12" (instance $wasi:clocks/monotonic-clock@0.2.12))
      (with "wasi:clocks/wall-clock@0.2.12" (instance $wasi:clocks/wall-clock@0.2.12))
      (with "wasi:random/random@0.2.12" (instance $wasi:random/random@0.2.12))
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12))
      (with "wasi:cli/stdout@0.2.12" (instance $wasi:cli/stdout@0.2.12))
      (with "wasi:cli/stderr@0.2.12" (instance $wasi:cli/stderr@0.2.12))
      (with "wasi:cli/stdin@0.2.12" (instance $wasi:cli/stdin@0.2.12))
      (with "wasi:http/types@0.2.12" (instance $wasi:http/types@0.2.12))
      (with "wasi:http/outgoing-handler@0.2.12" (instance $wasi:http/outgoing-handler@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32 i32)))
    (type (;1;) (func (param i32)))
    (type (;2;) (func (param i64 i32)))
    (type (;3;) (func (param i32 i32)))
    (type (;4;) (func (param i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32 i32)))
    (type (;6;) (func (param i32 i32 i64 i32)))
    (type (;7;) (func (param i32 i32 i32) (result i32)))
    (type (;8;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;9;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;10;) (func (param i32 i32 i32 i32 i32) (result i32)))
    (type (;11;) (func (param i32 i32 i32 i32 i64 i32 i32 i32 i32)))
    (type (;12;) (func))
    (import "actual" "0" (func $0 (;0;) (type 0)))
    (import "actual" "1" (func $1 (;1;) (type 1)))
    (import "actual" "2" (func $2 (;2;) (type 1)))
    (import "actual" "3" (func $3 (;3;) (type 2)))
    (import "actual" "4" (func $4 (;4;) (type 3)))
    (import "actual" "5" (func $5 (;5;) (type 4)))
    (import "actual" "6" (func $6 (;6;) (type 4)))
    (import "actual" "7" (func $7 (;7;) (type 4)))
    (import "actual" "8" (func $8 (;8;) (type 4)))
    (import "actual" "9" (func $9 (;9;) (type 3)))
    (import "actual" "10" (func $10 (;10;) (type 5)))
    (import "actual" "11" (func $11 (;11;) (type 5)))
    (import "actual" "12" (func $12 (;12;) (type 3)))
    (import "actual" "13" (func $13 (;13;) (type 3)))
    (import "actual" "14" (func $14 (;14;) (type 4)))
    (import "actual" "15" (func $15 (;15;) (type 4)))
    (import "actual" "16" (func $16 (;16;) (type 6)))
    (import "actual" "17" (func $17 (;17;) (type 6)))
    (import "actual" "18" (func $18 (;18;) (type 0)))
    (import "actual" "19" (func $19 (;19;) (type 5)))
    (import "actual" "20" (func $20 (;20;) (type 7)))
    (import "actual" "21" (func $21 (;21;) (type 8)))
    (import "actual" "22" (func $22 (;22;) (type 5)))
    (import "actual" "23" (func $23 (;23;) (type 8)))
    (import "actual" "24" (func $24 (;24;) (type 3)))
    (import "actual" "25" (func $25 (;25;) (type 3)))
    (import "actual" "26" (func $26 (;26;) (type 3)))
    (import "actual" "27" (func $27 (;27;) (type 3)))
    (import "actual" "28" (func $28 (;28;) (type 3)))
    (import "actual" "29" (func $29 (;29;) (type 3)))
    (import "actual" "30" (func $30 (;30;) (type 3)))
    (import "actual" "31" (func $31 (;31;) (type 3)))
    (import "actual" "32" (func $32 (;32;) (type 9)))
    (import "actual" "33" (func $33 (;33;) (type 3)))
    (import "actual" "34" (func $34 (;34;) (type 9)))
    (import "actual" "35" (func $35 (;35;) (type 3)))
    (import "actual" "36" (func $36 (;36;) (type 10)))
    (import "actual" "37" (func $37 (;37;) (type 3)))
    (import "actual" "38" (func $38 (;38;) (type 9)))
    (import "actual" "39" (func $39 (;39;) (type 3)))
    (import "actual" "40" (func $40 (;40;) (type 3)))
    (import "actual" "41" (func $41 (;41;) (type 3)))
    (import "actual" "42" (func $42 (;42;) (type 11)))
    (import "actual" "43" (func $43 (;43;) (type 3)))
    (import "actual" "44" (func $44 (;44;) (type 3)))
    (import "actual" "45" (func $45 (;45;) (type 3)))
    (import "actual" "46" (func $46 (;46;) (type 3)))
    (import "actual" "47" (func $47 (;47;) (type 3)))
    (import "actual" "48" (func $48 (;48;) (type 5)))
    (import "actual" "49" (func $49 (;49;) (type 3)))
    (import "actual" "50" (func $50 (;50;) (type 3)))
    (import "actual" "51" (func $51 (;51;) (type 5)))
    (import "main" "_initialize" (func $_initialize (;52;) (type 12)))
    (import "shim" "$imports" (table (;0;) 52 52 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16 $17 $18 $19 $20 $21 $22 $23 $24 $25 $26 $27 $28 $29 $30 $31 $32 $33 $34 $35 $36 $37 $38 $39 $40 $41 $42 $43 $44 $45 $46 $47 $48 $49 $50 $51)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;30;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;97;)))
  (core func $"#core-func98 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;98;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:clocks/wall-clock@0.2.12 "now" (func $"#func31 now" (@name "now") (;31;)))
  (core func $"#core-func99 indirect-wasi:clocks/wall-clock@0.2.12-now" (@name "indirect-wasi:clocks/wall-clock@0.2.12-now") (;99;) (canon lower (func $"#func31 now") (memory $memory)))
  (alias export $wasi:clocks/wall-clock@0.2.12 "resolution" (func $"#func32 resolution" (@name "resolution") (;32;)))
  (core func $"#core-func100 indirect-wasi:clocks/wall-clock@0.2.12-resolution" (@name "indirect-wasi:clocks/wall-clock@0.2.12-resolution") (;100;) (canon lower (func $"#func32 resolution") (memory $memory)))
  (alias export $wasi:random/random@0.2.12 "get-random-bytes" (func $get-random-bytes (;33;)))
  (core func $"#core-func101 indirect-wasi:random/random@0.2.12-get-random-bytes" (@name "indirect-wasi:random/random@0.2.12-get-random-bytes") (;101;) (canon lower (func $get-random-bytes) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;34;)))
  (core func $"#core-func102 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;102;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.read" (func $"[method]input-stream.read" (;35;)))
  (core func $"#core-func103 indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.read") (;103;) (canon lower (func $"[method]input-stream.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-read" (func $"[method]input-stream.blocking-read" (;36;)))
  (core func $"#core-func104 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read") (;104;) (canon lower (func $"[method]input-stream.blocking-read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.skip" (func $"[method]input-stream.skip" (;37;)))
  (core func $"#core-func105 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.skip") (;105;) (canon lower (func $"[method]input-stream.skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-skip" (func $"[method]input-stream.blocking-skip" (;38;)))
  (core func $"#core-func106 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip") (;106;) (canon lower (func $"[method]input-stream.blocking-skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.check-write" (func $"[method]output-stream.check-write" (;39;)))
  (core func $"#core-func107 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write") (;107;) (canon lower (func $"[method]output-stream.check-write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write" (func $"[method]output-stream.write" (;40;)))
  (core func $"#core-func108 indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write") (;108;) (canon lower (func $"[method]output-stream.write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-and-flush" (func $"[method]output-stream.blocking-write-and-flush" (;41;)))
  (core func $"#core-func109 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush") (;109;) (canon lower (func $"[method]output-stream.blocking-write-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.flush" (func $"[method]output-stream.flush" (;42;)))
  (core func $"#core-func110 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.flush") (;110;) (canon lower (func $"[method]output-stream.flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-flush" (func $"[method]output-stream.blocking-flush" (;43;)))
  (core func $"#core-func111 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush") (;111;) (canon lower (func $"[method]output-stream.blocking-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write-zeroes" (func $"[method]output-stream.write-zeroes" (;44;)))
  (core func $"#core-func112 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes") (;112;) (canon lower (func $"[method]output-stream.write-zeroes") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-zeroes-and-flush" (func $"[method]output-stream.blocking-write-zeroes-and-flush" (;45;)))
  (core func $"#core-func113 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush") (;113;) (canon lower (func $"[method]output-stream.blocking-write-zeroes-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.splice" (func $"[method]output-stream.splice" (;46;)))
  (core func $"#core-func114 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.splice") (;114;) (canon lower (func $"[method]output-stream.splice") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-splice" (func $"[method]output-stream.blocking-splice" (;47;)))
  (core func $"#core-func115 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice") (;115;) (canon lower (func $"[method]output-stream.blocking-splice") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]fields.from-list" (func $"[static]fields.from-list" (;48;)))
  (core func $"#core-func116 indirect-wasi:http/types@0.2.12-[static]fields.from-list" (@name "indirect-wasi:http/types@0.2.12-[static]fields.from-list") (;116;) (canon lower (func $"[static]fields.from-list") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.get" (func $"[method]fields.get" (;49;)))
  (core func $"#core-func117 indirect-wasi:http/types@0.2.12-[method]fields.get" (@name "indirect-wasi:http/types@0.2.12-[method]fields.get") (;117;) (canon lower (func $"[method]fields.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.has" (func $"[method]fields.has" (;50;)))
  (core func $"#core-func118 indirect-wasi:http/types@0.2.12-[method]fields.has" (@name "indirect-wasi:http/types@0.2.12-[method]fields.has") (;118;) (canon lower (func $"[method]fields.has") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.set" (func $"[method]fields.set" (;51;)))
  (core func $"#core-func119 indirect-wasi:http/types@0.2.12-[method]fields.set" (@name "indirect-wasi:http/types@0.2.12-[method]fields.set") (;119;) (canon lower (func $"[method]fields.set") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.delete" (func $"[method]fields.delete" (;52;)))
  (core func $"#core-func120 indirect-wasi:http/types@0.2.12-[method]fields.delete" (@name "indirect-wasi:http/types@0.2.12-[method]fields.delete") (;120;) (canon lower (func $"[method]fields.delete") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.append" (func $"[method]fields.append" (;53;)))
  (core func $"#core-func121 indirect-wasi:http/types@0.2.12-[method]fields.append" (@name "indirect-wasi:http/types@0.2.12-[method]fields.append") (;121;) (canon lower (func $"[method]fields.append") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]fields.entries" (func $"[method]fields.entries" (;54;)))
  (core func $"#core-func122 indirect-wasi:http/types@0.2.12-[method]fields.entries" (@name "indirect-wasi:http/types@0.2.12-[method]fields.entries") (;122;) (canon lower (func $"[method]fields.entries") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.method" (func $"[method]incoming-request.method" (;55;)))
  (core func $"#core-func123 indirect-wasi:http/types@0.2.12-[method]incoming-request.method" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.method") (;123;) (canon lower (func $"[method]incoming-request.method") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.path-with-query" (func $"[method]incoming-request.path-with-query" (;56;)))
  (core func $"#core-func124 indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query") (;124;) (canon lower (func $"[method]incoming-request.path-with-query") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.scheme" (func $"[method]incoming-request.scheme" (;57;)))
  (core func $"#core-func125 indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme") (;125;) (canon lower (func $"[method]incoming-request.scheme") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.authority" (func $"[method]incoming-request.authority" (;58;)))
  (core func $"#core-func126 indirect-wasi:http/types@0.2.12-[method]incoming-request.authority" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.authority") (;126;) (canon lower (func $"[method]incoming-request.authority") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-request.consume" (func $"[method]incoming-request.consume" (;59;)))
  (core func $"#core-func127 indirect-wasi:http/types@0.2.12-[method]incoming-request.consume" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-request.consume") (;127;) (canon lower (func $"[method]incoming-request.consume") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.body" (func $"[method]outgoing-request.body" (;60;)))
  (core func $"#core-func128 indirect-wasi:http/types@0.2.12-[method]outgoing-request.body" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.body") (;128;) (canon lower (func $"[method]outgoing-request.body") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.method" (func $"[method]outgoing-request.method" (;61;)))
  (core func $"#core-func129 indirect-wasi:http/types@0.2.12-[method]outgoing-request.method" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.method") (;129;) (canon lower (func $"[method]outgoing-request.method") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-method" (func $"[method]outgoing-request.set-method" (;62;)))
  (core func $"#core-func130 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method") (;130;) (canon lower (func $"[method]outgoing-request.set-method") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.path-with-query" (func $"[method]outgoing-request.path-with-query" (;63;)))
  (core func $"#core-func131 indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query") (;131;) (canon lower (func $"[method]outgoing-request.path-with-query") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-path-with-query" (func $"[method]outgoing-request.set-path-with-query" (;64;)))
  (core func $"#core-func132 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query") (;132;) (canon lower (func $"[method]outgoing-request.set-path-with-query") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.scheme" (func $"[method]outgoing-request.scheme" (;65;)))
  (core func $"#core-func133 indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme") (;133;) (canon lower (func $"[method]outgoing-request.scheme") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-scheme" (func $"[method]outgoing-request.set-scheme" (;66;)))
  (core func $"#core-func134 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme") (;134;) (canon lower (func $"[method]outgoing-request.set-scheme") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.authority" (func $"[method]outgoing-request.authority" (;67;)))
  (core func $"#core-func135 indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority") (;135;) (canon lower (func $"[method]outgoing-request.authority") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-request.set-authority" (func $"[method]outgoing-request.set-authority" (;68;)))
  (core func $"#core-func136 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority") (;136;) (canon lower (func $"[method]outgoing-request.set-authority") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.connect-timeout" (func $"[method]request-options.connect-timeout" (;69;)))
  (core func $"#core-func137 indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout") (;137;) (canon lower (func $"[method]request-options.connect-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.first-byte-timeout" (func $"[method]request-options.first-byte-timeout" (;70;)))
  (core func $"#core-func138 indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout") (;138;) (canon lower (func $"[method]request-options.first-byte-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]request-options.between-bytes-timeout" (func $"[method]request-options.between-bytes-timeout" (;71;)))
  (core func $"#core-func139 indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout" (@name "indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout") (;139;) (canon lower (func $"[method]request-options.between-bytes-timeout") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]response-outparam.set" (func $"[static]response-outparam.set" (;72;)))
  (core func $"#core-func140 indirect-wasi:http/types@0.2.12-[static]response-outparam.set" (@name "indirect-wasi:http/types@0.2.12-[static]response-outparam.set") (;140;) (canon lower (func $"[static]response-outparam.set") (memory $memory) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-response.consume" (func $"[method]incoming-response.consume" (;73;)))
  (core func $"#core-func141 indirect-wasi:http/types@0.2.12-[method]incoming-response.consume" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-response.consume") (;141;) (canon lower (func $"[method]incoming-response.consume") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]incoming-body.stream" (func $"[method]incoming-body.stream" (;74;)))
  (core func $"#core-func142 indirect-wasi:http/types@0.2.12-[method]incoming-body.stream" (@name "indirect-wasi:http/types@0.2.12-[method]incoming-body.stream") (;142;) (canon lower (func $"[method]incoming-body.stream") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]future-trailers.get" (func $"[method]future-trailers.get" (;75;)))
  (core func $"#core-func143 indirect-wasi:http/types@0.2.12-[method]future-trailers.get" (@name "indirect-wasi:http/types@0.2.12-[method]future-trailers.get") (;143;) (canon lower (func $"[method]future-trailers.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-response.body" (func $"[method]outgoing-response.body" (;76;)))
  (core func $"#core-func144 indirect-wasi:http/types@0.2.12-[method]outgoing-response.body" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-response.body") (;144;) (canon lower (func $"[method]outgoing-response.body") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[method]outgoing-body.write" (func $"[method]outgoing-body.write" (;77;)))
  (core func $"#core-func145 indirect-wasi:http/types@0.2.12-[method]outgoing-body.write" (@name "indirect-wasi:http/types@0.2.12-[method]outgoing-body.write") (;145;) (canon lower (func $"[method]outgoing-body.write") (memory $memory)))
  (alias export $wasi:http/types@0.2.12 "[static]outgoing-body.finish" (func $"[static]outgoing-body.finish" (;78;)))
  (core func $"#core-func146 indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish" (@name "indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish") (;146;) (canon lower (func $"[static]outgoing-body.finish") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "[method]future-incoming-response.get" (func $"[method]future-incoming-response.get" (;79;)))
  (core func $"#core-func147 indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get" (@name "indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get") (;147;) (canon lower (func $"[method]future-incoming-response.get") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/types@0.2.12 "http-error-code" (func $http-error-code (;80;)))
  (core func $"#core-func148 indirect-wasi:http/types@0.2.12-http-error-code" (@name "indirect-wasi:http/types@0.2.12-http-error-code") (;148;) (canon lower (func $http-error-code) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:http/outgoing-handler@0.2.12 "handle" (func $handle (;81;)))
  (core func $"#core-func149 indirect-wasi:http/outgoing-handler@0.2.12-handle" (@name "indirect-wasi:http/outgoing-handler@0.2.12-handle") (;149;) (canon lower (func $handle) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (core instance $actual (;13;)
    (export "0" (func $"#core-func98 indirect-wasi:io/poll@0.2.12-poll"))
    (export "1" (func $"#core-func99 indirect-wasi:clocks/wall-clock@0.2.12-now"))
    (export "2" (func $"#core-func100 indirect-wasi:clocks/wall-clock@0.2.12-resolution"))
    (export "3" (func $"#core-func101 indirect-wasi:random/random@0.2.12-get-random-bytes"))
    (export "4" (func $"#core-func102 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "5" (func $"#core-func103 indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "6" (func $"#core-func104 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "7" (func $"#core-func105 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "8" (func $"#core-func106 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "9" (func $"#core-func107 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "10" (func $"#core-func108 indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "11" (func $"#core-func109 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "12" (func $"#core-func110 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "13" (func $"#core-func111 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "14" (func $"#core-func112 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "15" (func $"#core-func113 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "16" (func $"#core-func114 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "17" (func $"#core-func115 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "18" (func $"#core-func116 indirect-wasi:http/types@0.2.12-[static]fields.from-list"))
    (export "19" (func $"#core-func117 indirect-wasi:http/types@0.2.12-[method]fields.get"))
    (export "20" (func $"#core-func118 indirect-wasi:http/types@0.2.12-[method]fields.has"))
    (export "21" (func $"#core-func119 indirect-wasi:http/types@0.2.12-[method]fields.set"))
    (export "22" (func $"#core-func120 indirect-wasi:http/types@0.2.12-[method]fields.delete"))
    (export "23" (func $"#core-func121 indirect-wasi:http/types@0.2.12-[method]fields.append"))
    (export "24" (func $"#core-func122 indirect-wasi:http/types@0.2.12-[method]fields.entries"))
    (export "25" (func $"#core-func123 indirect-wasi:http/types@0.2.12-[method]incoming-request.method"))
    (export "26" (func $"#core-func124 indirect-wasi:http/types@0.2.12-[method]incoming-request.path-with-query"))
    (export "27" (func $"#core-func125 indirect-wasi:http/types@0.2.12-[method]incoming-request.scheme"))
    (export "28" (func $"#core-func126 indirect-wasi:http/types@0.2.12-[method]incoming-request.authority"))
    (export "29" (func $"#core-func127 indirect-wasi:http/types@0.2.12-[method]incoming-request.consume"))
    (export "30" (func $"#core-func128 indirect-wasi:http/types@0.2.12-[method]outgoing-request.body"))
    (export "31" (func $"#core-func129 indirect-wasi:http/types@0.2.12-[method]outgoing-request.method"))
    (export "32" (func $"#core-func130 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-method"))
    (export "33" (func $"#core-func131 indirect-wasi:http/types@0.2.12-[method]outgoing-request.path-with-query"))
    (export "34" (func $"#core-func132 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-path-with-query"))
    (export "35" (func $"#core-func133 indirect-wasi:http/types@0.2.12-[method]outgoing-request.scheme"))
    (export "36" (func $"#core-func134 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-scheme"))
    (export "37" (func $"#core-func135 indirect-wasi:http/types@0.2.12-[method]outgoing-request.authority"))
    (export "38" (func $"#core-func136 indirect-wasi:http/types@0.2.12-[method]outgoing-request.set-authority"))
    (export "39" (func $"#core-func137 indirect-wasi:http/types@0.2.12-[method]request-options.connect-timeout"))
    (export "40" (func $"#core-func138 indirect-wasi:http/types@0.2.12-[method]request-options.first-byte-timeout"))
    (export "41" (func $"#core-func139 indirect-wasi:http/types@0.2.12-[method]request-options.between-bytes-timeout"))
    (export "42" (func $"#core-func140 indirect-wasi:http/types@0.2.12-[static]response-outparam.set"))
    (export "43" (func $"#core-func141 indirect-wasi:http/types@0.2.12-[method]incoming-response.consume"))
    (export "44" (func $"#core-func142 indirect-wasi:http/types@0.2.12-[method]incoming-body.stream"))
    (export "45" (func $"#core-func143 indirect-wasi:http/types@0.2.12-[method]future-trailers.get"))
    (export "46" (func $"#core-func144 indirect-wasi:http/types@0.2.12-[method]outgoing-response.body"))
    (export "47" (func $"#core-func145 indirect-wasi:http/types@0.2.12-[method]outgoing-body.write"))
    (export "48" (func $"#core-func146 indirect-wasi:http/types@0.2.12-[static]outgoing-body.finish"))
    (export "49" (func $"#core-func147 indirect-wasi:http/types@0.2.12-[method]future-incoming-response.get"))
    (export "50" (func $"#core-func148 indirect-wasi:http/types@0.2.12-http-error-code"))
    (export "51" (func $"#core-func149 indirect-wasi:http/outgoing-handler@0.2.12-handle"))
  )
  (core instance $fixup (;14;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;43;) (func (param "authority" string) (param "secure" bool) (result u32)))
  (alias core export $main "probe" (core func $probe (;150;)))
  (alias core export $main "cabi_post_probe" (core func $cabi_post_probe (;151;)))
  (func $probe (;82;) (type 43) (canon lift (core func $probe) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8 (post-return $cabi_post_probe)))
  (export $"#func83 probe" (@name "probe") (;83;) "probe" (func $probe))
  (type (;44;) (func (param "authority" string) (param "count" u32) (result u32)))
  (alias core export $main "request-batch" (core func $request-batch (;152;)))
  (alias core export $main "cabi_post_request-batch" (core func $cabi_post_request-batch (;153;)))
  (func $request-batch (;84;) (type 44) (canon lift (core func $request-batch) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8 (post-return $cabi_post_request-batch)))
  (export $"#func85 request-batch" (@name "request-batch") (;85;) "request-batch" (func $request-batch))
  (type (;45;) (own $incoming-request))
  (type (;46;) (own $response-outparam))
  (type (;47;) (func (param "request" 45) (param "response-out" 46)))
  (alias core export $main "wasi:http/incoming-handler@0.2.12#handle" (core func $wasi:http/incoming-handler@0.2.12#handle (;154;)))
  (alias core export $main "cabi_post_wasi:http/incoming-handler@0.2.12#handle" (core func $cabi_post_wasi:http/incoming-handler@0.2.12#handle (;155;)))
  (func $"#func86 handle" (@name "handle") (;86;) (type 47) (canon lift (core func $wasi:http/incoming-handler@0.2.12#handle) (post-return $cabi_post_wasi:http/incoming-handler@0.2.12#handle)))
  (component $wasi:http/incoming-handler@0.2.12-shim-component (;0;)
    (import "import-type-incoming-request" (type (;0;) (sub resource)))
    (import "import-type-response-outparam" (type (;1;) (sub resource)))
    (import "import-type-incoming-request0" (type (;2;) (eq 0)))
    (type (;3;) (own 2))
    (import "import-type-response-outparam0" (type (;4;) (eq 1)))
    (type (;5;) (own 4))
    (type (;6;) (func (param "request" 3) (param "response-out" 5)))
    (import "import-func-handle" (func (;0;) (type 6)))
    (export (;7;) "incoming-request" (type 0))
    (export (;8;) "response-outparam" (type 1))
    (type (;9;) (own 7))
    (type (;10;) (own 8))
    (type (;11;) (func (param "request" 9) (param "response-out" 10)))
    (export (;1;) "handle" (func 0) (func (type 11)))
  )
  (instance $wasi:http/incoming-handler@0.2.12-shim-instance (;11;) (instantiate $wasi:http/incoming-handler@0.2.12-shim-component
      (with "import-func-handle" (func $"#func86 handle"))
      (with "import-type-incoming-request" (type $incoming-request))
      (with "import-type-response-outparam" (type $response-outparam))
      (with "import-type-incoming-request0" (type $incoming-request))
      (with "import-type-response-outparam0" (type $response-outparam))
    )
  )
  (export $wasi:http/incoming-handler@0.2.12 (;12;) "wasi:http/incoming-handler@0.2.12" (instance $wasi:http/incoming-handler@0.2.12-shim-instance))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
 (export "types" (instance 9))
 (export "outgoing" (instance 10))
 (export "poll" (instance 0))
 (export "streams" (instance 5))
 (export "errors" (instance 4))
 (export "stdout" (instance 6))
)

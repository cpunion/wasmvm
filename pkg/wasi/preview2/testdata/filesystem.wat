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
  (type $ty-wasi:clocks/wall-clock@0.2.12 (;5;)
    (instance
      (type (;0;) (record (field "seconds" u64) (field "nanoseconds" u32)))
      (export (;1;) "datetime" (type (eq 0)))
      (type (;2;) (func (result 1)))
      (export (;0;) "now" (func (type 2)))
      (export (;1;) "resolution" (func (type 2)))
    )
  )
  (import "wasi:clocks/wall-clock@0.2.12" (instance $wasi:clocks/wall-clock@0.2.12 (;3;) (type $ty-wasi:clocks/wall-clock@0.2.12)))
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $input-stream (;6;)))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $output-stream (;7;)))
  (alias export $wasi:clocks/wall-clock@0.2.12 "datetime" (type $datetime (;8;)))
  (alias export $wasi:io/streams@0.2.12 "error" (type $"#type9 error" (@name "error") (;9;)))
  (type $ty-wasi:filesystem/types@0.2.12 (;10;)
    (instance
      (export (;0;) "descriptor" (type (sub resource)))
      (type (;1;) u64)
      (export (;2;) "filesize" (type (eq 1)))
      (alias outer 1 $input-stream (type (;3;)))
      (export (;4;) "input-stream" (type (eq 3)))
      (type (;5;) (enum "access" "would-block" "already" "bad-descriptor" "busy" "deadlock" "quota" "exist" "file-too-large" "illegal-byte-sequence" "in-progress" "interrupted" "invalid" "io" "is-directory" "loop" "too-many-links" "message-size" "name-too-long" "no-device" "no-entry" "no-lock" "insufficient-memory" "insufficient-space" "not-directory" "not-empty" "not-recoverable" "unsupported" "no-tty" "no-such-device" "overflow" "not-permitted" "pipe" "read-only" "invalid-seek" "text-file-busy" "cross-device"))
      (export (;6;) "error-code" (type (eq 5)))
      (alias outer 1 $output-stream (type (;7;)))
      (export (;8;) "output-stream" (type (eq 7)))
      (type (;9;) (enum "normal" "sequential" "random" "will-need" "dont-need" "no-reuse"))
      (export (;10;) "advice" (type (eq 9)))
      (type (;11;) (flags "read" "write" "file-integrity-sync" "data-integrity-sync" "requested-write-sync" "mutate-directory"))
      (export (;12;) "descriptor-flags" (type (eq 11)))
      (type (;13;) (enum "unknown" "block-device" "character-device" "directory" "fifo" "symbolic-link" "regular-file" "socket"))
      (export (;14;) "descriptor-type" (type (eq 13)))
      (alias outer 1 $datetime (type (;15;)))
      (export (;16;) "datetime" (type (eq 15)))
      (type (;17;) (variant (case "no-change") (case "now") (case "timestamp" 16)))
      (export (;18;) "new-timestamp" (type (eq 17)))
      (export (;19;) "directory-entry-stream" (type (sub resource)))
      (type (;20;) u64)
      (export (;21;) "link-count" (type (eq 20)))
      (type (;22;) (option 16))
      (type (;23;) (record (field "type" 14) (field "link-count" 21) (field "size" 2) (field "data-access-timestamp" 22) (field "data-modification-timestamp" 22) (field "status-change-timestamp" 22)))
      (export (;24;) "descriptor-stat" (type (eq 23)))
      (type (;25;) (flags "symlink-follow"))
      (export (;26;) "path-flags" (type (eq 25)))
      (type (;27;) (flags "create" "directory" "exclusive" "truncate"))
      (export (;28;) "open-flags" (type (eq 27)))
      (type (;29;) (record (field "lower" u64) (field "upper" u64)))
      (export (;30;) "metadata-hash-value" (type (eq 29)))
      (type (;31;) (record (field "type" 14) (field "name" string)))
      (export (;32;) "directory-entry" (type (eq 31)))
      (alias outer 1 $"#type9 error" (type (;33;)))
      (export (;34;) "error" (type (eq 33)))
      (type (;35;) (borrow 0))
      (type (;36;) (own 4))
      (type (;37;) (result 36 (error 6)))
      (type (;38;) (func (param "self" 35) (param "offset" 2) (result 37)))
      (export (;0;) "[method]descriptor.read-via-stream" (func (type 38)))
      (type (;39;) (own 8))
      (type (;40;) (result 39 (error 6)))
      (type (;41;) (func (param "self" 35) (param "offset" 2) (result 40)))
      (export (;1;) "[method]descriptor.write-via-stream" (func (type 41)))
      (type (;42;) (func (param "self" 35) (result 40)))
      (export (;2;) "[method]descriptor.append-via-stream" (func (type 42)))
      (type (;43;) (result (error 6)))
      (type (;44;) (func (param "self" 35) (param "offset" 2) (param "length" 2) (param "advice" 10) (result 43)))
      (export (;3;) "[method]descriptor.advise" (func (type 44)))
      (type (;45;) (func (param "self" 35) (result 43)))
      (export (;4;) "[method]descriptor.sync-data" (func (type 45)))
      (type (;46;) (result 12 (error 6)))
      (type (;47;) (func (param "self" 35) (result 46)))
      (export (;5;) "[method]descriptor.get-flags" (func (type 47)))
      (type (;48;) (result 14 (error 6)))
      (type (;49;) (func (param "self" 35) (result 48)))
      (export (;6;) "[method]descriptor.get-type" (func (type 49)))
      (type (;50;) (func (param "self" 35) (param "size" 2) (result 43)))
      (export (;7;) "[method]descriptor.set-size" (func (type 50)))
      (type (;51;) (func (param "self" 35) (param "data-access-timestamp" 18) (param "data-modification-timestamp" 18) (result 43)))
      (export (;8;) "[method]descriptor.set-times" (func (type 51)))
      (type (;52;) (list u8))
      (type (;53;) (tuple 52 bool))
      (type (;54;) (result 53 (error 6)))
      (type (;55;) (func (param "self" 35) (param "length" 2) (param "offset" 2) (result 54)))
      (export (;9;) "[method]descriptor.read" (func (type 55)))
      (type (;56;) (result 2 (error 6)))
      (type (;57;) (func (param "self" 35) (param "buffer" 52) (param "offset" 2) (result 56)))
      (export (;10;) "[method]descriptor.write" (func (type 57)))
      (type (;58;) (own 19))
      (type (;59;) (result 58 (error 6)))
      (type (;60;) (func (param "self" 35) (result 59)))
      (export (;11;) "[method]descriptor.read-directory" (func (type 60)))
      (export (;12;) "[method]descriptor.sync" (func (type 45)))
      (type (;61;) (func (param "self" 35) (param "path" string) (result 43)))
      (export (;13;) "[method]descriptor.create-directory-at" (func (type 61)))
      (type (;62;) (result 24 (error 6)))
      (type (;63;) (func (param "self" 35) (result 62)))
      (export (;14;) "[method]descriptor.stat" (func (type 63)))
      (type (;64;) (func (param "self" 35) (param "path-flags" 26) (param "path" string) (result 62)))
      (export (;15;) "[method]descriptor.stat-at" (func (type 64)))
      (type (;65;) (func (param "self" 35) (param "path-flags" 26) (param "path" string) (param "data-access-timestamp" 18) (param "data-modification-timestamp" 18) (result 43)))
      (export (;16;) "[method]descriptor.set-times-at" (func (type 65)))
      (type (;66;) (func (param "self" 35) (param "old-path-flags" 26) (param "old-path" string) (param "new-descriptor" 35) (param "new-path" string) (result 43)))
      (export (;17;) "[method]descriptor.link-at" (func (type 66)))
      (type (;67;) (own 0))
      (type (;68;) (result 67 (error 6)))
      (type (;69;) (func (param "self" 35) (param "path-flags" 26) (param "path" string) (param "open-flags" 28) (param "flags" 12) (result 68)))
      (export (;18;) "[method]descriptor.open-at" (func (type 69)))
      (type (;70;) (result string (error 6)))
      (type (;71;) (func (param "self" 35) (param "path" string) (result 70)))
      (export (;19;) "[method]descriptor.readlink-at" (func (type 71)))
      (export (;20;) "[method]descriptor.remove-directory-at" (func (type 61)))
      (type (;72;) (func (param "self" 35) (param "old-path" string) (param "new-descriptor" 35) (param "new-path" string) (result 43)))
      (export (;21;) "[method]descriptor.rename-at" (func (type 72)))
      (type (;73;) (func (param "self" 35) (param "old-path" string) (param "new-path" string) (result 43)))
      (export (;22;) "[method]descriptor.symlink-at" (func (type 73)))
      (export (;23;) "[method]descriptor.unlink-file-at" (func (type 61)))
      (type (;74;) (func (param "self" 35) (param "other" 35) (result bool)))
      (export (;24;) "[method]descriptor.is-same-object" (func (type 74)))
      (type (;75;) (result 30 (error 6)))
      (type (;76;) (func (param "self" 35) (result 75)))
      (export (;25;) "[method]descriptor.metadata-hash" (func (type 76)))
      (type (;77;) (func (param "self" 35) (param "path-flags" 26) (param "path" string) (result 75)))
      (export (;26;) "[method]descriptor.metadata-hash-at" (func (type 77)))
      (type (;78;) (borrow 19))
      (type (;79;) (option 32))
      (type (;80;) (result 79 (error 6)))
      (type (;81;) (func (param "self" 78) (result 80)))
      (export (;27;) "[method]directory-entry-stream.read-directory-entry" (func (type 81)))
      (type (;82;) (borrow 34))
      (type (;83;) (option 6))
      (type (;84;) (func (param "err" 82) (result 83)))
      (export (;28;) "filesystem-error-code" (func (type 84)))
    )
  )
  (import "wasi:filesystem/types@0.2.12" (instance $wasi:filesystem/types@0.2.12 (;4;) (type $ty-wasi:filesystem/types@0.2.12)))
  (alias export $wasi:filesystem/types@0.2.12 "descriptor" (type $descriptor (;11;)))
  (type $ty-wasi:filesystem/preopens@0.2.12 (;12;)
    (instance
      (alias outer 1 $descriptor (type (;0;)))
      (export (;1;) "descriptor" (type (eq 0)))
      (type (;2;) (own 1))
      (type (;3;) (tuple 2 string))
      (type (;4;) (list 3))
      (type (;5;) (func (result 4)))
      (export (;0;) "get-directories" (func (type 5)))
    )
  )
  (import "wasi:filesystem/preopens@0.2.12" (instance $wasi:filesystem/preopens@0.2.12 (;5;) (type $ty-wasi:filesystem/preopens@0.2.12)))
  (core module $main (;0;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32)))
    (type (;2;) (func (param i32) (result i32)))
    (type (;3;) (func (param i32 i32 i32)))
    (type (;4;) (func (param i32 i64 i32)))
    (type (;5;) (func (param i32 i32 i32 i32)))
    (type (;6;) (func (param i32 i32 i64 i32)))
    (type (;7;) (func (param i32 i64 i64 i32 i32)))
    (type (;8;) (func (param i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;9;) (func (param i32 i64 i64 i32)))
    (type (;10;) (func (param i32 i32 i32 i64 i32)))
    (type (;11;) (func (param i32 i32 i32 i32 i32)))
    (type (;12;) (func (param i32 i32 i32 i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;13;) (func (param i32 i32 i32 i32 i32 i32 i32 i32)))
    (type (;14;) (func (param i32 i32 i32 i32 i32 i32 i32)))
    (type (;15;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;16;) (func (param i32 i32) (result i32)))
    (type (;17;) (func (result i32)))
    (type (;18;) (func (param i32 i32 i32 i32) (result i32)))
    (type (;19;) (func))
    (import "wasi:io/error@0.2.12" "[method]error.to-debug-string" (func (;0;) (type 0)))
    (import "wasi:io/error@0.2.12" "[resource-drop]error" (func (;1;) (type 1)))
    (import "wasi:io/poll@0.2.12" "[method]pollable.ready" (func (;2;) (type 2)))
    (import "wasi:io/poll@0.2.12" "[method]pollable.block" (func (;3;) (type 1)))
    (import "wasi:io/poll@0.2.12" "poll" (func (;4;) (type 3)))
    (import "wasi:io/poll@0.2.12" "[resource-drop]pollable" (func (;5;) (type 1)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.read" (func (;6;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-read" (func (;7;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.skip" (func (;8;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.blocking-skip" (func (;9;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]input-stream.subscribe" (func (;10;) (type 2)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.check-write" (func (;11;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.write" (func (;12;) (type 5)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-and-flush" (func (;13;) (type 5)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.flush" (func (;14;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-flush" (func (;15;) (type 0)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.subscribe" (func (;16;) (type 2)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.write-zeroes" (func (;17;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-write-zeroes-and-flush" (func (;18;) (type 4)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.splice" (func (;19;) (type 6)))
    (import "wasi:io/streams@0.2.12" "[method]output-stream.blocking-splice" (func (;20;) (type 6)))
    (import "wasi:io/streams@0.2.12" "[resource-drop]input-stream" (func (;21;) (type 1)))
    (import "wasi:io/streams@0.2.12" "[resource-drop]output-stream" (func (;22;) (type 1)))
    (import "wasi:clocks/wall-clock@0.2.12" "now" (func (;23;) (type 1)))
    (import "wasi:clocks/wall-clock@0.2.12" "resolution" (func (;24;) (type 1)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read-via-stream" (func (;25;) (type 4)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.write-via-stream" (func (;26;) (type 4)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.append-via-stream" (func (;27;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.advise" (func (;28;) (type 7)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.sync-data" (func (;29;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.get-flags" (func (;30;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.get-type" (func (;31;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-size" (func (;32;) (type 4)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-times" (func (;33;) (type 8)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read" (func (;34;) (type 9)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.write" (func (;35;) (type 10)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.read-directory" (func (;36;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.sync" (func (;37;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.create-directory-at" (func (;38;) (type 5)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.stat" (func (;39;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.stat-at" (func (;40;) (type 11)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.set-times-at" (func (;41;) (type 12)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.link-at" (func (;42;) (type 13)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.open-at" (func (;43;) (type 14)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.readlink-at" (func (;44;) (type 5)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.remove-directory-at" (func (;45;) (type 5)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.rename-at" (func (;46;) (type 14)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.symlink-at" (func (;47;) (type 15)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.unlink-file-at" (func (;48;) (type 5)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.is-same-object" (func (;49;) (type 16)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.metadata-hash" (func (;50;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[method]descriptor.metadata-hash-at" (func (;51;) (type 11)))
    (import "wasi:filesystem/types@0.2.12" "[method]directory-entry-stream.read-directory-entry" (func (;52;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "filesystem-error-code" (func (;53;) (type 0)))
    (import "wasi:filesystem/types@0.2.12" "[resource-drop]descriptor" (func (;54;) (type 1)))
    (import "wasi:filesystem/types@0.2.12" "[resource-drop]directory-entry-stream" (func (;55;) (type 1)))
    (import "wasi:filesystem/preopens@0.2.12" "get-directories" (func (;56;) (type 1)))
    (memory (;0;) 1)
    (global $heap (;0;) (mut i32) i32.const 8192)
    (export "exercise" (func 57))
    (export "cabi_post_exercise" (func 58))
    (export "stat-batch" (func 59))
    (export "cabi_post_stat-batch" (func 60))
    (export "driver" (func 61))
    (export "cabi_post_driver" (func 62))
    (export "error-probe" (func 63))
    (export "cabi_post_error-probe" (func 64))
    (export "memory" (memory 0))
    (export "cabi_realloc" (func 65))
    (export "_initialize" (func 66))
    (func (;57;) (type 17) (result i32)
      (local $root i32) (local $file i32) (local $stream i32) (local $poll i32) (local $entries i32) (local $count i32)
      i32.const 0
      call 56
      i32.const 4
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 0
      i32.load
      i32.load
      local.set $root
      local.get $root
      i32.const 64
      call 30
      i32.const 64
      call $ok
      i32.const 65
      i32.load8_u
      i32.const 1
      i32.and
      i32.eqz
      if ;; label = @1
        unreachable
      end
      local.get $root
      i32.const 64
      call 31
      i32.const 64
      call $ok
      i32.const 65
      i32.load8_u
      i32.const 3
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $root
      i32.const 1088
      i32.const 3
      i32.const 64
      call 38
      i32.const 64
      call $ok
      local.get $root
      i32.const 0
      i32.const 1024
      i32.const 5
      i32.const 1
      i32.const 3
      i32.const 64
      call 43
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $file
      local.get $file
      i32.const 1104
      i32.const 4
      i64.const 0
      i32.const 64
      call 35
      i32.const 64
      call $ok
      i32.const 72
      i64.load
      i64.const 4
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $file
      i64.const 4
      i64.const 0
      i32.const 64
      call 34
      i32.const 64
      call $ok
      i32.const 72
      i32.load
      i32.const 4
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 68
      i32.load
      i32.load
      i32.const 1684234849
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $file
      i64.const 8
      i32.const 64
      call 32
      i32.const 64
      call $ok
      local.get $file
      i32.const 64
      call 39
      i32.const 64
      call $ok
      i32.const 88
      i64.load
      i64.const 8
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $file
      i32.const 64
      call 31
      i32.const 64
      call $ok
      i32.const 65
      i32.load8_u
      i32.const 6
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $file
      i32.const 64
      call 30
      i32.const 64
      call $ok
      i32.const 65
      i32.load8_u
      i32.const 3
      i32.and
      i32.const 3
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $file
      i64.const 0
      i64.const 8
      i32.const 0
      i32.const 64
      call 28
      i32.const 64
      call $ok
      local.get $file
      i32.const 64
      call 29
      i32.const 64
      call $ok
      local.get $file
      i32.const 64
      call 37
      i32.const 64
      call $ok
      local.get $file
      i32.const 1
      i64.const 0
      i32.const 0
      i32.const 1
      i64.const 0
      i32.const 0
      i32.const 64
      call 33
      i32.const 64
      call $ok
      local.get $file
      i64.const 0
      i32.const 64
      call 25
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      call 10
      local.tee $poll
      call 3
      local.get $poll
      call 5
      local.get $stream
      i64.const 4
      i32.const 64
      call 7
      i32.const 64
      call $ok
      i32.const 72
      i32.load
      i32.const 4
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $stream
      call 21
      local.get $file
      i64.const 4
      i32.const 64
      call 26
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      i32.const 1108
      i32.const 2
      i32.const 64
      call 13
      i32.const 64
      call $ok
      local.get $stream
      call 22
      local.get $file
      i32.const 64
      call 27
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      i32.const 1110
      i32.const 1
      i32.const 64
      call 13
      i32.const 64
      call $ok
      local.get $stream
      call 22
      local.get $root
      i32.const 0
      i32.const 1024
      i32.const 5
      i32.const 64
      call 40
      i32.const 64
      call $ok
      i32.const 88
      i64.load
      i64.const 9
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $root
      i32.const 0
      i32.const 1024
      i32.const 5
      i32.const 1
      i64.const 0
      i32.const 0
      i32.const 1
      i64.const 0
      i32.const 0
      i32.const 64
      call 41
      i32.const 64
      call $ok
      local.get $root
      i32.const 0
      i32.const 1024
      i32.const 5
      local.get $root
      i32.const 1040
      i32.const 4
      i32.const 64
      call 42
      i32.const 64
      call $ok
      local.get $root
      i32.const 1040
      i32.const 4
      local.get $root
      i32.const 1056
      i32.const 5
      i32.const 64
      call 46
      i32.const 64
      call $ok
      local.get $root
      i32.const 1024
      i32.const 5
      i32.const 1072
      i32.const 3
      i32.const 64
      call 47
      i32.const 64
      call $ok
      local.get $root
      i32.const 1072
      i32.const 3
      i32.const 64
      call 44
      i32.const 64
      call $ok
      i32.const 72
      i32.load
      i32.const 5
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $root
      i32.const 1
      i32.const 1072
      i32.const 3
      i32.const 0
      i32.const 1
      i32.const 64
      call 43
      i32.const 64
      call $ok
      local.get $file
      i32.const 68
      i32.load
      call 49
      i32.eqz
      if ;; label = @1
        unreachable
      end
      i32.const 68
      i32.load
      call 54
      local.get $file
      i32.const 64
      call 50
      i32.const 64
      call $ok
      local.get $root
      i32.const 0
      i32.const 1024
      i32.const 5
      i32.const 128
      call 51
      i32.const 128
      call $ok
      i32.const 72
      i64.load
      i32.const 136
      i64.load
      i64.ne
      if ;; label = @1
        unreachable
      end
      i32.const 80
      i64.load
      i32.const 144
      i64.load
      i64.ne
      if ;; label = @1
        unreachable
      end
      local.get $root
      i32.const 64
      call 36
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $entries
      block $end
        loop $next
          local.get $entries
          i32.const 64
          call 52
          i32.const 64
          call $ok
          i32.const 68
          i32.load8_u
          i32.eqz
          br_if $end
          local.get $count
          i32.const 1
          i32.add
          local.set $count
          br $next
        end
      end
      local.get $count
      i32.const 4
      i32.lt_u
      if ;; label = @1
        unreachable
      end
      local.get $entries
      call 55
      local.get $root
      i32.const 1056
      i32.const 5
      i32.const 64
      call 48
      i32.const 64
      call $ok
      local.get $root
      i32.const 1072
      i32.const 3
      i32.const 64
      call 48
      i32.const 64
      call $ok
      local.get $file
      call 54
      local.get $root
      i32.const 1024
      i32.const 5
      i32.const 64
      call 48
      i32.const 64
      call $ok
      local.get $root
      i32.const 1088
      i32.const 3
      i32.const 64
      call 45
      i32.const 64
      call $ok
      local.get $root
      call 54
      i32.const 42
    )
    (func (;58;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;59;) (type 2) (param $count i32) (result i32)
      (local $root i32) (local $file i32) (local $at i32)
      i32.const 0
      call 56
      i32.const 4
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 0
      i32.load
      i32.load
      local.set $root
      local.get $root
      i32.const 0
      i32.const 1120
      i32.const 9
      i32.const 0
      i32.const 1
      i32.const 64
      call 43
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $file
      block $end
        loop $next
          local.get $at
          local.get $count
          i32.ge_u
          br_if $end
          local.get $file
          i32.const 64
          call 39
          i32.const 64
          call $ok
          i32.const 88
          i64.load
          i64.const 16
          i64.ne
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
      local.get $file
      call 54
      local.get $root
      call 54
      local.get $at
    )
    (func (;60;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;61;) (type 17) (result i32)
      (local $root i32)
      i32.const 0
      call 56
      i32.const 4
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 0
      i32.load
      i32.load
      local.set $root
      local.get $root
      call 54
      i32.const 1
    )
    (func (;62;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;63;) (type 17) (result i32)
      (local $root i32) (local $file i32) (local $stream i32) (local $error i32)
      i32.const 0
      call 56
      i32.const 4
      i32.load
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      i32.const 0
      i32.load
      i32.load
      local.set $root
      local.get $root
      i32.const 0
      i32.const 1120
      i32.const 9
      i32.const 0
      i32.const 1
      i32.const 64
      call 43
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $file
      local.get $file
      i64.const 0
      i32.const 64
      call 25
      i32.const 64
      call $ok
      i32.const 68
      i32.load
      local.set $stream
      local.get $stream
      i64.const 1
      i32.const 64
      call 6
      i32.const 64
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
      i32.const 72
      i32.load
      local.set $error
      local.get $error
      i32.const 128
      call 53
      i32.const 128
      i32.load8_u
      i32.const 1
      i32.ne
      if ;; label = @1
        unreachable
      end
      local.get $error
      call 1
      local.get $stream
      call 21
      local.get $file
      call 54
      local.get $root
      call 54
      i32.const 1
    )
    (func (;64;) (type 1) (param i32)
      i32.const 8192
      global.set $heap
    )
    (func (;65;) (type 18) (param i32 i32 i32 i32) (result i32)
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
    (func (;66;) (type 19))
    (func $ok (;67;) (type 1) (param i32)
      local.get 0
      i32.load8_u
      if ;; label = @1
        unreachable
      end
    )
    (data (;0;) (i32.const 1024) "alpha")
    (data (;1;) (i32.const 1040) "beta")
    (data (;2;) (i32.const 1056) "gamma")
    (data (;3;) (i32.const 1072) "sym")
    (data (;4;) (i32.const 1088) "sub")
    (data (;5;) (i32.const 1104) "abcdEF!")
    (data (;6;) (i32.const 1120) "bench.bin")
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
    (type (;5;) (func (param i32)))
    (type (;6;) (func (param i32 i64 i64 i32 i32)))
    (type (;7;) (func (param i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;8;) (func (param i32 i64 i64 i32)))
    (type (;9;) (func (param i32 i32 i32 i64 i32)))
    (type (;10;) (func (param i32 i32 i32 i32 i32)))
    (type (;11;) (func (param i32 i32 i32 i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;12;) (func (param i32 i32 i32 i32 i32 i32 i32 i32)))
    (type (;13;) (func (param i32 i32 i32 i32 i32 i32 i32)))
    (type (;14;) (func (param i32 i32 i32 i32 i32 i32)))
    (table (;0;) 46 46 funcref)
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
    (export "15" (func $indirect-wasi:clocks/wall-clock@0.2.12-now))
    (export "16" (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution))
    (export "17" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream"))
    (export "18" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream"))
    (export "19" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream"))
    (export "20" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise"))
    (export "21" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data"))
    (export "22" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags"))
    (export "23" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type"))
    (export "24" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size"))
    (export "25" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times"))
    (export "26" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read"))
    (export "27" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write"))
    (export "28" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory"))
    (export "29" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync"))
    (export "30" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at"))
    (export "31" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat"))
    (export "32" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at"))
    (export "33" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at"))
    (export "34" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at"))
    (export "35" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at"))
    (export "36" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at"))
    (export "37" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at"))
    (export "38" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at"))
    (export "39" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at"))
    (export "40" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at"))
    (export "41" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash"))
    (export "42" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at"))
    (export "43" (func $"indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry"))
    (export "44" (func $indirect-wasi:filesystem/types@0.2.12-filesystem-error-code))
    (export "45" (func $indirect-wasi:filesystem/preopens@0.2.12-get-directories))
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
    (func $indirect-wasi:clocks/wall-clock@0.2.12-now (;15;) (type 5) (param i32)
      local.get 0
      i32.const 15
      call_indirect (type 5)
    )
    (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution (;16;) (type 5) (param i32)
      local.get 0
      i32.const 16
      call_indirect (type 5)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream" (;17;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 17
      call_indirect (type 2)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream" (;18;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 18
      call_indirect (type 2)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream" (;19;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 19
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise" (;20;) (type 6) (param i32 i64 i64 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 20
      call_indirect (type 6)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data" (;21;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 21
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags" (;22;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 22
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type" (;23;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 23
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size" (;24;) (type 2) (param i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      i32.const 24
      call_indirect (type 2)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times" (;25;) (type 7) (param i32 i32 i64 i32 i32 i64 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      i32.const 25
      call_indirect (type 7)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read" (;26;) (type 8) (param i32 i64 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 26
      call_indirect (type 8)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write" (;27;) (type 9) (param i32 i32 i32 i64 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 27
      call_indirect (type 9)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory" (;28;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 28
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync" (;29;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 29
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at" (;30;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 30
      call_indirect (type 3)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat" (;31;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 31
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at" (;32;) (type 10) (param i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 32
      call_indirect (type 10)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at" (;33;) (type 11) (param i32 i32 i32 i32 i32 i64 i32 i32 i64 i32 i32)
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
      i32.const 33
      call_indirect (type 11)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at" (;34;) (type 12) (param i32 i32 i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      local.get 7
      i32.const 34
      call_indirect (type 12)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at" (;35;) (type 13) (param i32 i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      i32.const 35
      call_indirect (type 13)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at" (;36;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 36
      call_indirect (type 3)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at" (;37;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 37
      call_indirect (type 3)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at" (;38;) (type 13) (param i32 i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      local.get 6
      i32.const 38
      call_indirect (type 13)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at" (;39;) (type 14) (param i32 i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      local.get 5
      i32.const 39
      call_indirect (type 14)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at" (;40;) (type 3) (param i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      i32.const 40
      call_indirect (type 3)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash" (;41;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 41
      call_indirect (type 0)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at" (;42;) (type 10) (param i32 i32 i32 i32 i32)
      local.get 0
      local.get 1
      local.get 2
      local.get 3
      local.get 4
      i32.const 42
      call_indirect (type 10)
    )
    (func $"indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry" (;43;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 43
      call_indirect (type 0)
    )
    (func $indirect-wasi:filesystem/types@0.2.12-filesystem-error-code (;44;) (type 0) (param i32 i32)
      local.get 0
      local.get 1
      i32.const 44
      call_indirect (type 0)
    )
    (func $indirect-wasi:filesystem/preopens@0.2.12-get-directories (;45;) (type 5) (param i32)
      local.get 0
      i32.const 45
      call_indirect (type 5)
    )
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (core instance $wit-component-shim-instance (;0;) (instantiate $wit-component-shim-module))
  (alias core export $wit-component-shim-instance "0" (core func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (;0;)))
  (alias export $wasi:io/error@0.2.12 "error" (type $"#type13 error" (@name "error") (;13;)))
  (core func $resource.drop (;1;) (canon resource.drop $"#type13 error"))
  (core instance $wasi:io/error@0.2.12 (;1;)
    (export "[method]error.to-debug-string" (func $"indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "[resource-drop]error" (func $resource.drop))
  )
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.ready" (func $"[method]pollable.ready" (;0;)))
  (core func $"[method]pollable.ready" (;2;) (canon lower (func $"[method]pollable.ready")))
  (alias export $wasi:io/poll@0.2.12 "[method]pollable.block" (func $"[method]pollable.block" (;1;)))
  (core func $"[method]pollable.block" (;3;) (canon lower (func $"[method]pollable.block")))
  (alias core export $wit-component-shim-instance "1" (core func $indirect-wasi:io/poll@0.2.12-poll (;4;)))
  (alias export $wasi:io/poll@0.2.12 "pollable" (type $"#type14 pollable" (@name "pollable") (;14;)))
  (core func $"#core-func5 resource.drop" (@name "resource.drop") (;5;) (canon resource.drop $"#type14 pollable"))
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
  (alias export $wasi:io/streams@0.2.12 "input-stream" (type $"#type15 input-stream" (@name "input-stream") (;15;)))
  (core func $"#core-func21 resource.drop" (@name "resource.drop") (;21;) (canon resource.drop $"#type15 input-stream"))
  (alias export $wasi:io/streams@0.2.12 "output-stream" (type $"#type16 output-stream" (@name "output-stream") (;16;)))
  (core func $"#core-func22 resource.drop" (@name "resource.drop") (;22;) (canon resource.drop $"#type16 output-stream"))
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
  (alias core export $wit-component-shim-instance "15" (core func $indirect-wasi:clocks/wall-clock@0.2.12-now (;23;)))
  (alias core export $wit-component-shim-instance "16" (core func $indirect-wasi:clocks/wall-clock@0.2.12-resolution (;24;)))
  (core instance $wasi:clocks/wall-clock@0.2.12 (;4;)
    (export "now" (func $indirect-wasi:clocks/wall-clock@0.2.12-now))
    (export "resolution" (func $indirect-wasi:clocks/wall-clock@0.2.12-resolution))
  )
  (alias core export $wit-component-shim-instance "17" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream" (;25;)))
  (alias core export $wit-component-shim-instance "18" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream" (;26;)))
  (alias core export $wit-component-shim-instance "19" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream" (;27;)))
  (alias core export $wit-component-shim-instance "20" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise" (;28;)))
  (alias core export $wit-component-shim-instance "21" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data" (;29;)))
  (alias core export $wit-component-shim-instance "22" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags" (;30;)))
  (alias core export $wit-component-shim-instance "23" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type" (;31;)))
  (alias core export $wit-component-shim-instance "24" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size" (;32;)))
  (alias core export $wit-component-shim-instance "25" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times" (;33;)))
  (alias core export $wit-component-shim-instance "26" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read" (;34;)))
  (alias core export $wit-component-shim-instance "27" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write" (;35;)))
  (alias core export $wit-component-shim-instance "28" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory" (;36;)))
  (alias core export $wit-component-shim-instance "29" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync" (;37;)))
  (alias core export $wit-component-shim-instance "30" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at" (;38;)))
  (alias core export $wit-component-shim-instance "31" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat" (;39;)))
  (alias core export $wit-component-shim-instance "32" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at" (;40;)))
  (alias core export $wit-component-shim-instance "33" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at" (;41;)))
  (alias core export $wit-component-shim-instance "34" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at" (;42;)))
  (alias core export $wit-component-shim-instance "35" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at" (;43;)))
  (alias core export $wit-component-shim-instance "36" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at" (;44;)))
  (alias core export $wit-component-shim-instance "37" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at" (;45;)))
  (alias core export $wit-component-shim-instance "38" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at" (;46;)))
  (alias core export $wit-component-shim-instance "39" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at" (;47;)))
  (alias core export $wit-component-shim-instance "40" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at" (;48;)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.is-same-object" (func $"[method]descriptor.is-same-object" (;4;)))
  (core func $"[method]descriptor.is-same-object" (;49;) (canon lower (func $"[method]descriptor.is-same-object")))
  (alias core export $wit-component-shim-instance "41" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash" (;50;)))
  (alias core export $wit-component-shim-instance "42" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at" (;51;)))
  (alias core export $wit-component-shim-instance "43" (core func $"indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry" (;52;)))
  (alias core export $wit-component-shim-instance "44" (core func $indirect-wasi:filesystem/types@0.2.12-filesystem-error-code (;53;)))
  (alias export $wasi:filesystem/types@0.2.12 "descriptor" (type $"#type17 descriptor" (@name "descriptor") (;17;)))
  (core func $"#core-func54 resource.drop" (@name "resource.drop") (;54;) (canon resource.drop $"#type17 descriptor"))
  (alias export $wasi:filesystem/types@0.2.12 "directory-entry-stream" (type $directory-entry-stream (;18;)))
  (core func $"#core-func55 resource.drop" (@name "resource.drop") (;55;) (canon resource.drop $directory-entry-stream))
  (core instance $wasi:filesystem/types@0.2.12 (;5;)
    (export "[method]descriptor.read-via-stream" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream"))
    (export "[method]descriptor.write-via-stream" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream"))
    (export "[method]descriptor.append-via-stream" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream"))
    (export "[method]descriptor.advise" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise"))
    (export "[method]descriptor.sync-data" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data"))
    (export "[method]descriptor.get-flags" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags"))
    (export "[method]descriptor.get-type" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type"))
    (export "[method]descriptor.set-size" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size"))
    (export "[method]descriptor.set-times" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times"))
    (export "[method]descriptor.read" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read"))
    (export "[method]descriptor.write" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write"))
    (export "[method]descriptor.read-directory" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory"))
    (export "[method]descriptor.sync" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync"))
    (export "[method]descriptor.create-directory-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at"))
    (export "[method]descriptor.stat" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat"))
    (export "[method]descriptor.stat-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at"))
    (export "[method]descriptor.set-times-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at"))
    (export "[method]descriptor.link-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at"))
    (export "[method]descriptor.open-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at"))
    (export "[method]descriptor.readlink-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at"))
    (export "[method]descriptor.remove-directory-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at"))
    (export "[method]descriptor.rename-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at"))
    (export "[method]descriptor.symlink-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at"))
    (export "[method]descriptor.unlink-file-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at"))
    (export "[method]descriptor.is-same-object" (func $"[method]descriptor.is-same-object"))
    (export "[method]descriptor.metadata-hash" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash"))
    (export "[method]descriptor.metadata-hash-at" (func $"indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at"))
    (export "[method]directory-entry-stream.read-directory-entry" (func $"indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry"))
    (export "filesystem-error-code" (func $indirect-wasi:filesystem/types@0.2.12-filesystem-error-code))
    (export "[resource-drop]descriptor" (func $"#core-func54 resource.drop"))
    (export "[resource-drop]directory-entry-stream" (func $"#core-func55 resource.drop"))
  )
  (alias core export $wit-component-shim-instance "45" (core func $indirect-wasi:filesystem/preopens@0.2.12-get-directories (;56;)))
  (core instance $wasi:filesystem/preopens@0.2.12 (;6;)
    (export "get-directories" (func $indirect-wasi:filesystem/preopens@0.2.12-get-directories))
  )
  (core instance $main (;7;) (instantiate $main
      (with "wasi:io/error@0.2.12" (instance $wasi:io/error@0.2.12))
      (with "wasi:io/poll@0.2.12" (instance $wasi:io/poll@0.2.12))
      (with "wasi:io/streams@0.2.12" (instance $wasi:io/streams@0.2.12))
      (with "wasi:clocks/wall-clock@0.2.12" (instance $wasi:clocks/wall-clock@0.2.12))
      (with "wasi:filesystem/types@0.2.12" (instance $wasi:filesystem/types@0.2.12))
      (with "wasi:filesystem/preopens@0.2.12" (instance $wasi:filesystem/preopens@0.2.12))
    )
  )
  (alias core export $main "memory" (core memory $memory (;0;)))
  (core module $wit-component-fixup (;2;)
    (type (;0;) (func (param i32 i32)))
    (type (;1;) (func (param i32 i32 i32)))
    (type (;2;) (func (param i32 i64 i32)))
    (type (;3;) (func (param i32 i32 i32 i32)))
    (type (;4;) (func (param i32 i32 i64 i32)))
    (type (;5;) (func (param i32)))
    (type (;6;) (func (param i32 i64 i64 i32 i32)))
    (type (;7;) (func (param i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;8;) (func (param i32 i64 i64 i32)))
    (type (;9;) (func (param i32 i32 i32 i64 i32)))
    (type (;10;) (func (param i32 i32 i32 i32 i32)))
    (type (;11;) (func (param i32 i32 i32 i32 i32 i64 i32 i32 i64 i32 i32)))
    (type (;12;) (func (param i32 i32 i32 i32 i32 i32 i32 i32)))
    (type (;13;) (func (param i32 i32 i32 i32 i32 i32 i32)))
    (type (;14;) (func (param i32 i32 i32 i32 i32 i32)))
    (type (;15;) (func))
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
    (import "actual" "15" (func $15 (;15;) (type 5)))
    (import "actual" "16" (func $16 (;16;) (type 5)))
    (import "actual" "17" (func $17 (;17;) (type 2)))
    (import "actual" "18" (func $18 (;18;) (type 2)))
    (import "actual" "19" (func $19 (;19;) (type 0)))
    (import "actual" "20" (func $20 (;20;) (type 6)))
    (import "actual" "21" (func $21 (;21;) (type 0)))
    (import "actual" "22" (func $22 (;22;) (type 0)))
    (import "actual" "23" (func $23 (;23;) (type 0)))
    (import "actual" "24" (func $24 (;24;) (type 2)))
    (import "actual" "25" (func $25 (;25;) (type 7)))
    (import "actual" "26" (func $26 (;26;) (type 8)))
    (import "actual" "27" (func $27 (;27;) (type 9)))
    (import "actual" "28" (func $28 (;28;) (type 0)))
    (import "actual" "29" (func $29 (;29;) (type 0)))
    (import "actual" "30" (func $30 (;30;) (type 3)))
    (import "actual" "31" (func $31 (;31;) (type 0)))
    (import "actual" "32" (func $32 (;32;) (type 10)))
    (import "actual" "33" (func $33 (;33;) (type 11)))
    (import "actual" "34" (func $34 (;34;) (type 12)))
    (import "actual" "35" (func $35 (;35;) (type 13)))
    (import "actual" "36" (func $36 (;36;) (type 3)))
    (import "actual" "37" (func $37 (;37;) (type 3)))
    (import "actual" "38" (func $38 (;38;) (type 13)))
    (import "actual" "39" (func $39 (;39;) (type 14)))
    (import "actual" "40" (func $40 (;40;) (type 3)))
    (import "actual" "41" (func $41 (;41;) (type 0)))
    (import "actual" "42" (func $42 (;42;) (type 10)))
    (import "actual" "43" (func $43 (;43;) (type 0)))
    (import "actual" "44" (func $44 (;44;) (type 0)))
    (import "actual" "45" (func $45 (;45;) (type 5)))
    (import "main" "_initialize" (func $_initialize (;46;) (type 15)))
    (import "shim" "$imports" (table (;0;) 46 46 funcref))
    (start $_initialize)
    (elem (;0;) (i32.const 0) func $0 $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15 $16 $17 $18 $19 $20 $21 $22 $23 $24 $25 $26 $27 $28 $29 $30 $31 $32 $33 $34 $35 $36 $37 $38 $39 $40 $41 $42 $43 $44 $45)
    (@producers
      (processed-by "wit-component" "0.258.0")
    )
  )
  (alias export $wasi:io/error@0.2.12 "[method]error.to-debug-string" (func $"[method]error.to-debug-string" (;5;)))
  (alias core export $main "cabi_realloc" (core func $cabi_realloc (;57;)))
  (core func $"#core-func58 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string" (@name "indirect-wasi:io/error@0.2.12-[method]error.to-debug-string") (;58;) (canon lower (func $"[method]error.to-debug-string") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:io/poll@0.2.12 "poll" (func $poll (;6;)))
  (core func $"#core-func59 indirect-wasi:io/poll@0.2.12-poll" (@name "indirect-wasi:io/poll@0.2.12-poll") (;59;) (canon lower (func $poll) (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.read" (func $"[method]input-stream.read" (;7;)))
  (core func $"#core-func60 indirect-wasi:io/streams@0.2.12-[method]input-stream.read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.read") (;60;) (canon lower (func $"[method]input-stream.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-read" (func $"[method]input-stream.blocking-read" (;8;)))
  (core func $"#core-func61 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read") (;61;) (canon lower (func $"[method]input-stream.blocking-read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.skip" (func $"[method]input-stream.skip" (;9;)))
  (core func $"#core-func62 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.skip") (;62;) (canon lower (func $"[method]input-stream.skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]input-stream.blocking-skip" (func $"[method]input-stream.blocking-skip" (;10;)))
  (core func $"#core-func63 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip" (@name "indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip") (;63;) (canon lower (func $"[method]input-stream.blocking-skip") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.check-write" (func $"[method]output-stream.check-write" (;11;)))
  (core func $"#core-func64 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write") (;64;) (canon lower (func $"[method]output-stream.check-write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write" (func $"[method]output-stream.write" (;12;)))
  (core func $"#core-func65 indirect-wasi:io/streams@0.2.12-[method]output-stream.write" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write") (;65;) (canon lower (func $"[method]output-stream.write") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-and-flush" (func $"[method]output-stream.blocking-write-and-flush" (;13;)))
  (core func $"#core-func66 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush") (;66;) (canon lower (func $"[method]output-stream.blocking-write-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.flush" (func $"[method]output-stream.flush" (;14;)))
  (core func $"#core-func67 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.flush") (;67;) (canon lower (func $"[method]output-stream.flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-flush" (func $"[method]output-stream.blocking-flush" (;15;)))
  (core func $"#core-func68 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush") (;68;) (canon lower (func $"[method]output-stream.blocking-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.write-zeroes" (func $"[method]output-stream.write-zeroes" (;16;)))
  (core func $"#core-func69 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes") (;69;) (canon lower (func $"[method]output-stream.write-zeroes") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-write-zeroes-and-flush" (func $"[method]output-stream.blocking-write-zeroes-and-flush" (;17;)))
  (core func $"#core-func70 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush") (;70;) (canon lower (func $"[method]output-stream.blocking-write-zeroes-and-flush") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.splice" (func $"[method]output-stream.splice" (;18;)))
  (core func $"#core-func71 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.splice") (;71;) (canon lower (func $"[method]output-stream.splice") (memory $memory)))
  (alias export $wasi:io/streams@0.2.12 "[method]output-stream.blocking-splice" (func $"[method]output-stream.blocking-splice" (;19;)))
  (core func $"#core-func72 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice" (@name "indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice") (;72;) (canon lower (func $"[method]output-stream.blocking-splice") (memory $memory)))
  (alias export $wasi:clocks/wall-clock@0.2.12 "now" (func $now (;20;)))
  (core func $"#core-func73 indirect-wasi:clocks/wall-clock@0.2.12-now" (@name "indirect-wasi:clocks/wall-clock@0.2.12-now") (;73;) (canon lower (func $now) (memory $memory)))
  (alias export $wasi:clocks/wall-clock@0.2.12 "resolution" (func $resolution (;21;)))
  (core func $"#core-func74 indirect-wasi:clocks/wall-clock@0.2.12-resolution" (@name "indirect-wasi:clocks/wall-clock@0.2.12-resolution") (;74;) (canon lower (func $resolution) (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.read-via-stream" (func $"[method]descriptor.read-via-stream" (;22;)))
  (core func $"#core-func75 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream") (;75;) (canon lower (func $"[method]descriptor.read-via-stream") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.write-via-stream" (func $"[method]descriptor.write-via-stream" (;23;)))
  (core func $"#core-func76 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream") (;76;) (canon lower (func $"[method]descriptor.write-via-stream") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.append-via-stream" (func $"[method]descriptor.append-via-stream" (;24;)))
  (core func $"#core-func77 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream") (;77;) (canon lower (func $"[method]descriptor.append-via-stream") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.advise" (func $"[method]descriptor.advise" (;25;)))
  (core func $"#core-func78 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise") (;78;) (canon lower (func $"[method]descriptor.advise") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.sync-data" (func $"[method]descriptor.sync-data" (;26;)))
  (core func $"#core-func79 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data") (;79;) (canon lower (func $"[method]descriptor.sync-data") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.get-flags" (func $"[method]descriptor.get-flags" (;27;)))
  (core func $"#core-func80 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags") (;80;) (canon lower (func $"[method]descriptor.get-flags") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.get-type" (func $"[method]descriptor.get-type" (;28;)))
  (core func $"#core-func81 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type") (;81;) (canon lower (func $"[method]descriptor.get-type") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.set-size" (func $"[method]descriptor.set-size" (;29;)))
  (core func $"#core-func82 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size") (;82;) (canon lower (func $"[method]descriptor.set-size") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.set-times" (func $"[method]descriptor.set-times" (;30;)))
  (core func $"#core-func83 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times") (;83;) (canon lower (func $"[method]descriptor.set-times") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.read" (func $"[method]descriptor.read" (;31;)))
  (core func $"#core-func84 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read") (;84;) (canon lower (func $"[method]descriptor.read") (memory $memory) (realloc $cabi_realloc)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.write" (func $"[method]descriptor.write" (;32;)))
  (core func $"#core-func85 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write") (;85;) (canon lower (func $"[method]descriptor.write") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.read-directory" (func $"[method]descriptor.read-directory" (;33;)))
  (core func $"#core-func86 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory") (;86;) (canon lower (func $"[method]descriptor.read-directory") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.sync" (func $"[method]descriptor.sync" (;34;)))
  (core func $"#core-func87 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync") (;87;) (canon lower (func $"[method]descriptor.sync") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.create-directory-at" (func $"[method]descriptor.create-directory-at" (;35;)))
  (core func $"#core-func88 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at") (;88;) (canon lower (func $"[method]descriptor.create-directory-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.stat" (func $"[method]descriptor.stat" (;36;)))
  (core func $"#core-func89 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat") (;89;) (canon lower (func $"[method]descriptor.stat") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.stat-at" (func $"[method]descriptor.stat-at" (;37;)))
  (core func $"#core-func90 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at") (;90;) (canon lower (func $"[method]descriptor.stat-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.set-times-at" (func $"[method]descriptor.set-times-at" (;38;)))
  (core func $"#core-func91 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at") (;91;) (canon lower (func $"[method]descriptor.set-times-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.link-at" (func $"[method]descriptor.link-at" (;39;)))
  (core func $"#core-func92 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at") (;92;) (canon lower (func $"[method]descriptor.link-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.open-at" (func $"[method]descriptor.open-at" (;40;)))
  (core func $"#core-func93 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at") (;93;) (canon lower (func $"[method]descriptor.open-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.readlink-at" (func $"[method]descriptor.readlink-at" (;41;)))
  (core func $"#core-func94 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at") (;94;) (canon lower (func $"[method]descriptor.readlink-at") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.remove-directory-at" (func $"[method]descriptor.remove-directory-at" (;42;)))
  (core func $"#core-func95 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at") (;95;) (canon lower (func $"[method]descriptor.remove-directory-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.rename-at" (func $"[method]descriptor.rename-at" (;43;)))
  (core func $"#core-func96 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at") (;96;) (canon lower (func $"[method]descriptor.rename-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.symlink-at" (func $"[method]descriptor.symlink-at" (;44;)))
  (core func $"#core-func97 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at") (;97;) (canon lower (func $"[method]descriptor.symlink-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.unlink-file-at" (func $"[method]descriptor.unlink-file-at" (;45;)))
  (core func $"#core-func98 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at") (;98;) (canon lower (func $"[method]descriptor.unlink-file-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.metadata-hash" (func $"[method]descriptor.metadata-hash" (;46;)))
  (core func $"#core-func99 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash") (;99;) (canon lower (func $"[method]descriptor.metadata-hash") (memory $memory)))
  (alias export $wasi:filesystem/types@0.2.12 "[method]descriptor.metadata-hash-at" (func $"[method]descriptor.metadata-hash-at" (;47;)))
  (core func $"#core-func100 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at" (@name "indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at") (;100;) (canon lower (func $"[method]descriptor.metadata-hash-at") (memory $memory) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "[method]directory-entry-stream.read-directory-entry" (func $"[method]directory-entry-stream.read-directory-entry" (;48;)))
  (core func $"#core-func101 indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry" (@name "indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry") (;101;) (canon lower (func $"[method]directory-entry-stream.read-directory-entry") (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (alias export $wasi:filesystem/types@0.2.12 "filesystem-error-code" (func $filesystem-error-code (;49;)))
  (core func $"#core-func102 indirect-wasi:filesystem/types@0.2.12-filesystem-error-code" (@name "indirect-wasi:filesystem/types@0.2.12-filesystem-error-code") (;102;) (canon lower (func $filesystem-error-code) (memory $memory)))
  (alias export $wasi:filesystem/preopens@0.2.12 "get-directories" (func $get-directories (;50;)))
  (core func $"#core-func103 indirect-wasi:filesystem/preopens@0.2.12-get-directories" (@name "indirect-wasi:filesystem/preopens@0.2.12-get-directories") (;103;) (canon lower (func $get-directories) (memory $memory) (realloc $cabi_realloc) string-encoding=utf8))
  (core instance $actual (;8;)
    (export "0" (func $"#core-func58 indirect-wasi:io/error@0.2.12-[method]error.to-debug-string"))
    (export "1" (func $"#core-func59 indirect-wasi:io/poll@0.2.12-poll"))
    (export "2" (func $"#core-func60 indirect-wasi:io/streams@0.2.12-[method]input-stream.read"))
    (export "3" (func $"#core-func61 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-read"))
    (export "4" (func $"#core-func62 indirect-wasi:io/streams@0.2.12-[method]input-stream.skip"))
    (export "5" (func $"#core-func63 indirect-wasi:io/streams@0.2.12-[method]input-stream.blocking-skip"))
    (export "6" (func $"#core-func64 indirect-wasi:io/streams@0.2.12-[method]output-stream.check-write"))
    (export "7" (func $"#core-func65 indirect-wasi:io/streams@0.2.12-[method]output-stream.write"))
    (export "8" (func $"#core-func66 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-and-flush"))
    (export "9" (func $"#core-func67 indirect-wasi:io/streams@0.2.12-[method]output-stream.flush"))
    (export "10" (func $"#core-func68 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-flush"))
    (export "11" (func $"#core-func69 indirect-wasi:io/streams@0.2.12-[method]output-stream.write-zeroes"))
    (export "12" (func $"#core-func70 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-write-zeroes-and-flush"))
    (export "13" (func $"#core-func71 indirect-wasi:io/streams@0.2.12-[method]output-stream.splice"))
    (export "14" (func $"#core-func72 indirect-wasi:io/streams@0.2.12-[method]output-stream.blocking-splice"))
    (export "15" (func $"#core-func73 indirect-wasi:clocks/wall-clock@0.2.12-now"))
    (export "16" (func $"#core-func74 indirect-wasi:clocks/wall-clock@0.2.12-resolution"))
    (export "17" (func $"#core-func75 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-via-stream"))
    (export "18" (func $"#core-func76 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write-via-stream"))
    (export "19" (func $"#core-func77 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.append-via-stream"))
    (export "20" (func $"#core-func78 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.advise"))
    (export "21" (func $"#core-func79 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync-data"))
    (export "22" (func $"#core-func80 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-flags"))
    (export "23" (func $"#core-func81 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.get-type"))
    (export "24" (func $"#core-func82 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-size"))
    (export "25" (func $"#core-func83 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times"))
    (export "26" (func $"#core-func84 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read"))
    (export "27" (func $"#core-func85 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.write"))
    (export "28" (func $"#core-func86 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.read-directory"))
    (export "29" (func $"#core-func87 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.sync"))
    (export "30" (func $"#core-func88 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.create-directory-at"))
    (export "31" (func $"#core-func89 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat"))
    (export "32" (func $"#core-func90 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.stat-at"))
    (export "33" (func $"#core-func91 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.set-times-at"))
    (export "34" (func $"#core-func92 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.link-at"))
    (export "35" (func $"#core-func93 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.open-at"))
    (export "36" (func $"#core-func94 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.readlink-at"))
    (export "37" (func $"#core-func95 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.remove-directory-at"))
    (export "38" (func $"#core-func96 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.rename-at"))
    (export "39" (func $"#core-func97 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.symlink-at"))
    (export "40" (func $"#core-func98 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.unlink-file-at"))
    (export "41" (func $"#core-func99 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash"))
    (export "42" (func $"#core-func100 indirect-wasi:filesystem/types@0.2.12-[method]descriptor.metadata-hash-at"))
    (export "43" (func $"#core-func101 indirect-wasi:filesystem/types@0.2.12-[method]directory-entry-stream.read-directory-entry"))
    (export "44" (func $"#core-func102 indirect-wasi:filesystem/types@0.2.12-filesystem-error-code"))
    (export "45" (func $"#core-func103 indirect-wasi:filesystem/preopens@0.2.12-get-directories"))
  )
  (core instance $fixup (;9;) (instantiate $wit-component-fixup
      (with "actual" (instance $actual))
      (with "main" (instance $main))
      (with "shim" (instance $wit-component-shim-instance))
    )
  )
  (type (;19;) (func (result u32)))
  (alias core export $main "exercise" (core func $exercise (;104;)))
  (alias core export $main "cabi_post_exercise" (core func $cabi_post_exercise (;105;)))
  (func $exercise (;51;) (type 19) (canon lift (core func $exercise) (post-return $cabi_post_exercise)))
  (export $"#func52 exercise" (@name "exercise") (;52;) "exercise" (func $exercise))
  (type (;20;) (func (param "count" u32) (result u32)))
  (alias core export $main "stat-batch" (core func $stat-batch (;106;)))
  (alias core export $main "cabi_post_stat-batch" (core func $cabi_post_stat-batch (;107;)))
  (func $stat-batch (;53;) (type 20) (canon lift (core func $stat-batch) (post-return $cabi_post_stat-batch)))
  (export $"#func54 stat-batch" (@name "stat-batch") (;54;) "stat-batch" (func $stat-batch))
  (alias core export $main "driver" (core func $driver (;108;)))
  (alias core export $main "cabi_post_driver" (core func $cabi_post_driver (;109;)))
  (func $driver (;55;) (type 19) (canon lift (core func $driver) (post-return $cabi_post_driver)))
  (export $"#func56 driver" (@name "driver") (;56;) "driver" (func $driver))
  (alias core export $main "error-probe" (core func $error-probe (;110;)))
  (alias core export $main "cabi_post_error-probe" (core func $cabi_post_error-probe (;111;)))
  (func $error-probe (;57;) (type 19) (canon lift (core func $error-probe) (post-return $cabi_post_error-probe)))
  (export $"#func58 error-probe" (@name "error-probe") (;58;) "error-probe" (func $error-probe))
  (@producers
    (processed-by "wit-component" "0.258.0")
  )
)

# P3 Store-local response collection: 2026-11-10

[Complete machine record](data/performance-p3-store-local-2026-11-10.json).

## Actual gaps and fixes

The streaming Go embedding adapter publishes a response while a Store worker drives the guest and an io.Pipe carries body chunks back to the application. The pinned Wasmtime adapter instead collects its body and trailers inside Store::run_concurrent. Those are different host topologies, even though both verify complete responses. This event compares both owned APIs using the same physical executable, without replacing the streaming baseline. It isolates the complete API choice, not a pure instruction-generation cost.

`HTTP.CallIncomingBuffered` materializes a complete response under one Store admission using the existing typed component scheduler. A positive explicit maximum bounds accepted body bytes; canonical transfers keep their existing IO/body limits. Headers, trailers, content-length validation, cancellation, timeout, quotas, independent returned storage, fuel and joined guest completion remain enforced. No extra Store worker or response pipe is needed on this path. The ordinary streaming API remains available for incremental or unbounded bodies.

A deterministic Store-contention regression also finds a preexisting correctness bug: publishing an incoming request before worker admission leaves a live HTTP resource if another call acquires the Store first. Both backends fail before the fix. `Function.CallWithResultPrepared` now prepares resources and executes the guest in the same lease, and rejected admission never calls preparation. Failure, panic, invalid argument and expired-Caller cleanup tests cover the generic API. The regression passes after the change.

Bounded input read-ahead from the separate [readiness study](performance-p3-read-ahead-2026-11-10.md) is present in both modes. Its instrumented 15 -> 2 external reads and 80 -> 41 task resumes are separate counter evidence, not counts measured in this event. A distinct pinned-compiler [zero-slice append issue](performance-llgo-zero-append-2026-11-10.md) is reported as [LLGo #3040](https://github.com/xgo-dev/llgo/issues/3040); no compiler patch or upgrade enters this comparison.

## Fixed comparison and limitations

M4 Max/macOS 27.0.1; source `d4c7045c2f747aa9d92f2e942734e162aef9134c`; Go 1.27.3 and LLGo `3d92a58` at `-O3 -lto=full`. One Go and one LLGo owned binary serve both API modes. Wasmtime 48.0.1 and ordinary native-Go application controls retain their qualified frozen binaries; shared admission/measurement code is byte-identical. Both Rust HTTP and composed fixtures, interpreter/native backends, one/four workers, 16 clients with four requests, 64 KiB, two warmups and seven measured batches. Every response byte, status 201, header and trailer is checked, with independent CRC32 2629362236.

Twenty randomized primary/A/A blocks, seed 2026111009; 8000 paired log-ratio bootstrap draws and pointwise 95% intervals. All 1600 throughput, 40 separate resource, 80 qualification and five actual loader-probe processes complete, checking 716800 warm responses. All owned build/test/probe sessions exited and the actual pre-start compiler/test process table was clear. There is no affinity, frequency or background isolation; GOMAXPROCS=4 is not a native CPU quota. No favorable repeat, source/profile/count sweep, cross-event pooling, family-wise claim or phase subtraction.

**Only 2/40 A/A controls resolve; no owned mode pair has both controls resolved.** This is advisory evidence, not a stable speed win, controlled ranking or release acceptance. Load changes from [4.89306640625, 5.39697265625, 6.5146484375] to [7.21484375, 10.0556640625, 8.63818359375]. All comparisons, including unresolved and adverse resource observations, remain in the record.

For four-worker LLGo native HTTP, streaming -> buffered observes 2841 -> 3489 req/s, complete API 1374.4 -> 1127.4 us and accounted allocation 846187 -> 561068 B/request (about 278 KiB less). The paired warm-time ratio is 0.8037, with unresolved controls. Between-batch threads fall 61 -> 57, while peak RSS rises 160.1 -> 176.2 MiB: lower cumulative allocation does not establish lower retained memory. Go native observes 12163 -> 13785 req/s. Same-event Wasmtime observes 58529 req/s and the bare LLGo application 8874 req/s.

The large gap remains: buffered LLGo native HTTP has a paired warm-time ratio of 16.47 to Wasmtime and 2.81 to its ordinary LLGo host control; neither pair resolves. Go buffered native observes 13785 req/s against LLGo 3489, while bare Go/LLGo applications observe 43821/8874. Host/runtime costs therefore need separate investigation; these ratios do not assign the entire deficit to either the Wasm compiler or LLGo. Allocation counters have different Go/LLGo scopes, and reference Go counters omit Rust/JIT allocation.

## Every owned API comparison

Warm ratios are buffered/streaming (smaller is faster); intervals are pointwise. No row has both A/A controls resolved. Throughput and bytes are medians of process reductions.

| Fixture / compiler / backend / workers | Streaming req/s | Buffered req/s | Warm ratio [95%] | Streaming B/req | Buffered B/req |
| --- | ---: | ---: | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 2495 | 2716 | 0.8423 [0.7703,0.9066] | 455528 | 375558 |
| world-rust-composed-http/go/interpreter/4 | 9244 | 9786 | 0.8857 [0.7971,0.9436] | 455502 | 375558 |
| world-rust-composed-http/go/native/1 | 3382 | 3764 | 0.8904 [0.8164,0.9654] | 451403 | 371428 |
| world-rust-composed-http/go/native/4 | 11381 | 12297 | 0.8993 [0.8082,0.9949] | 451377 | 371433 |
| world-rust-composed-http/llgo/interpreter/1 | 1661 | 2216 | 0.7505 [0.7300,0.7760] | 867521 | 582417 |
| world-rust-composed-http/llgo/interpreter/4 | 2804 | 3490 | 0.8200 [0.7837,0.8733] | 867539 | 582288 |
| world-rust-composed-http/llgo/native/1 | 1891 | 2663 | 0.7604 [0.6974,0.8481] | 863198 | 578139 |
| world-rust-composed-http/llgo/native/4 | 2737 | 3419 | 0.8792 [0.7789,1.0173] | 863193 | 578002 |
| world-rust-http/go/interpreter/1 | 2740 | 3006 | 0.9050 [0.8241,0.9837] | 439653 | 359681 |
| world-rust-http/go/interpreter/4 | 9995 | 10857 | 0.9743 [0.9075,1.0748] | 439621 | 359679 |
| world-rust-http/go/native/1 | 3777 | 4195 | 0.8808 [0.8260,0.9309] | 439797 | 359822 |
| world-rust-http/go/native/4 | 12163 | 13785 | 0.8766 [0.7946,0.9562] | 439769 | 359825 |
| world-rust-http/llgo/interpreter/1 | 1774 | 2466 | 0.6817 [0.5798,0.7898] | 845763 | 560686 |
| world-rust-http/llgo/interpreter/4 | 2908 | 3550 | 0.8376 [0.7021,0.9831] | 845867 | 560473 |
| world-rust-http/llgo/native/1 | 2008 | 2917 | 0.6523 [0.5110,0.8133] | 846220 | 561164 |
| world-rust-http/llgo/native/4 | 2841 | 3489 | 0.8037 [0.7364,0.8793] | 846187 | 561068 |

## All profiles and resource observations

| Profile | req/s | Complete API us | Request p99 us | Runtime B/req | Between-batch threads | Peak RSS MiB | A/A resolves |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| buffered/world-rust-composed-http/go/interpreter/1 | 2716 | 365.1 | 6698.2 | 375558 | 7 | 97.5 | yes |
| buffered/world-rust-composed-http/go/interpreter/4 | 9786 | 391.3 | 3410.1 | 375558 | 7 | 128.3 | no |
| buffered/world-rust-composed-http/go/native/1 | 3764 | 262.3 | 5032.9 | 371428 | 7 | 148.1 | no |
| buffered/world-rust-composed-http/go/native/4 | 12297 | 312.3 | 3051.1 | 371433 | 7 | 161.6 | no |
| buffered/world-rust-composed-http/llgo/interpreter/1 | 2216 | 444.9 | 10468.4 | 582417 | 42 | 150.0 | no |
| buffered/world-rust-composed-http/llgo/interpreter/4 | 3490 | 1126.8 | 7831.4 | 582288 | 57 | 180.3 | no |
| buffered/world-rust-composed-http/llgo/native/1 | 2663 | 367.2 | 9644.7 | 578139 | 42 | 180.8 | no |
| buffered/world-rust-composed-http/llgo/native/4 | 3419 | 1150.4 | 7924.1 | 578002 | 57 | 209.7 | no |
| buffered/world-rust-http/go/interpreter/1 | 3006 | 329.6 | 6065.8 | 359681 | 7 | 79.0 | no |
| buffered/world-rust-http/go/interpreter/4 | 10857 | 355.1 | 2897.7 | 359679 | 7 | 98.8 | no |
| buffered/world-rust-http/go/native/1 | 4195 | 235.0 | 4441.2 | 359822 | 7 | 117.8 | no |
| buffered/world-rust-http/go/native/4 | 13785 | 278.6 | 2571.7 | 359825 | 7 | 124.7 | no |
| buffered/world-rust-http/llgo/interpreter/1 | 2466 | 399.0 | 9153.6 | 560686 | 42 | 130.9 | no |
| buffered/world-rust-http/llgo/interpreter/4 | 3550 | 1107.9 | 7516.2 | 560473 | 57 | 140.8 | no |
| buffered/world-rust-http/llgo/native/1 | 2917 | 335.2 | 8940.8 | 561164 | 42 | 167.3 | no |
| buffered/world-rust-http/llgo/native/4 | 3489 | 1127.4 | 7770.4 | 561068 | 57 | 176.2 | no |
| host/native-go/go/host/1 | 40073 | 21.4 | 537.2 | 277914 | 6 | 17.3 | no |
| host/native-go/go/host/4 | 43821 | 78.8 | 584.0 | 277916 | 6 | 16.2 | no |
| host/native-go/llgo/host/1 | 8252 | 115.4 | 3795.4 | 732577 | 37 | 68.1 | no |
| host/native-go/llgo/host/4 | 8874 | 402.1 | 3487.1 | 732584 | 40 | 68.4 | no |
| reference/world-rust-composed-http/go/compiler/1 | 15002 | 64.3 | 1205.5 | 32 | 7 | 33.0 | no |
| reference/world-rust-composed-http/go/compiler/4 | 51304 | 73.3 | 382.1 | 32 | 7 | 33.2 | no |
| reference/world-rust-http/go/compiler/1 | 17406 | 55.0 | 1024.9 | 32 | 7 | 30.3 | no |
| reference/world-rust-http/go/compiler/4 | 58529 | 63.8 | 345.0 | 33 | 10 | 30.7 | no |
| streaming/world-rust-composed-http/go/interpreter/1 | 2495 | 398.3 | 7430.3 | 455528 | 7 | 103.9 | no |
| streaming/world-rust-composed-http/go/interpreter/4 | 9244 | 413.8 | 3115.1 | 455502 | 7 | 137.3 | no |
| streaming/world-rust-composed-http/go/native/1 | 3382 | 293.2 | 5592.2 | 451403 | 7 | 148.7 | yes |
| streaming/world-rust-composed-http/go/native/4 | 11381 | 337.6 | 3349.6 | 451377 | 7 | 158.8 | no |
| streaming/world-rust-composed-http/llgo/interpreter/1 | 1661 | 595.5 | 13114.4 | 867521 | 43 | 164.2 | no |
| streaming/world-rust-composed-http/llgo/interpreter/4 | 2804 | 1396.2 | 9075.0 | 867539 | 61 | 167.9 | no |
| streaming/world-rust-composed-http/llgo/native/1 | 1891 | 522.2 | 11923.9 | 863198 | 43 | 182.1 | no |
| streaming/world-rust-composed-http/llgo/native/4 | 2737 | 1427.1 | 9121.3 | 863193 | 61 | 209.8 | no |
| streaming/world-rust-http/go/interpreter/1 | 2740 | 362.5 | 6672.7 | 439653 | 7 | 81.6 | no |
| streaming/world-rust-http/go/interpreter/4 | 9995 | 386.8 | 3019.6 | 439621 | 8 | 103.0 | no |
| streaming/world-rust-http/go/native/1 | 3777 | 262.4 | 4891.8 | 439797 | 7 | 118.5 | no |
| streaming/world-rust-http/go/native/4 | 12163 | 316.0 | 2718.8 | 439769 | 7 | 120.5 | no |
| streaming/world-rust-http/llgo/interpreter/1 | 1774 | 556.4 | 12065.4 | 845763 | 43 | 140.5 | no |
| streaming/world-rust-http/llgo/interpreter/4 | 2908 | 1340.7 | 8036.9 | 845867 | 61 | 140.9 | no |
| streaming/world-rust-http/llgo/native/1 | 2008 | 486.9 | 10849.6 | 846220 | 43 | 168.5 | no |
| streaming/world-rust-http/llgo/native/4 | 2841 | 1374.4 | 8197.9 | 846187 | 61 | 160.1 | no |

The machine record retains all 40 profiles, 16 mode intervals, 64 paired reference intervals, per-process phase reductions and every resource snapshot. Readiness, load, engine setup, compile/validate, providers, instantiate, lookup, worker setup, first/warm batches, complete API, request latency, joined cleanup and process wall remain separate. Overlapping intervals and resource cohorts are never subtracted to invent pure costs. Request construction is outside the owned API timer; full validation remains inside request/batch timing. The reference constructs its request inside its complete bridge call.

## Validation and CI scope

Complete normal Go owned tests and vet pass. Exact CI-style cross-package coverage passes all 31 executable-package gates, minimum 90.86%. Affected component/P3 HTTP tests pass under the race detector; the P3 measurement module passes race/vet with 95.9% coverage. Actual pinned LLGo complete component, P3 and measurement package executions pass. Tests cover both backends, plain/composed fixtures, invalid UTF-8, partial chunks, empty/oversized bodies, independent retained storage, metadata/trailers, cancellation, timeout, forwarding, host panic, fuel and shared-World/two-Store isolation. Final supported-platform CI remains a separate merge requirement. Earlier full-P3-race diagnostic timeout and noisy read-ahead results remain explicitly recorded in their own study.

CI removes one duplicate LLGo p3host test build per platform from the threads shard. Its shared tests already execute in the owned P3 module; tagged Go quality coverage and standalone Go/LLGo host qualification remain. All 32 root packages, eight LLGo shards per platform, owned P3 execution and native-reference module execution remain. The full matrix still has 23 required jobs. This eliminates repeated linking without removing semantic inventory; the 5–10 minute job target still requires actual final-head evidence.

Next work targets remaining dynamic type/metadata allocation, host-runtime synchronization and native compiler gaps. LLGo #3034/#3035 remain upstream work, and #3040 is a separately reported exact allocation defect. Controlled comparisons on both targets and production-release acceptance remain open.

Raw result SHA256 `657fd76e83df9838116b45f9e85f7cacc60a432ddc5b620b6ccf3fe23109785e` in artifacts/p3-store-local-20261110. Source archives, binary/compiler/fixture hashes, actual linkage, compact source bundle and exact-source replay dependencies are retained.

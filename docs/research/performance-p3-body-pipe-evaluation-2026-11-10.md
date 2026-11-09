# P3 immutable-batch pipe evaluation: 2026-11-10

[Complete machine record](data/performance-p3-body-pipe-evaluation-2026-11-10.json).

**Not selected for production.** Base `825d588287b75f0563a72f1127be7d9cab34f92d`, prototype `209ed8556d166246969420dc5cbdd50a1e542450`, M4 Max/macOS27.0.1, Go1.27.3 and pinned LLGo3d92a58 `-O3 -lto=full`. The prototype replaces the private stdlib pipe with one immutable batch, retaining whole-batch backpressure and avoiding per-fragment writer wakeups and a mutable write copy. The standard-library transport is retained in subsequent production candidates.

LLGo HTTP native/4-worker throughput observes2101 ->2168req/s. Its warm-time ratio0.9660 has pointwise95% interval[0.9328,1.0004] and unresolved controls. The composed Go interpreter/1-worker case is adverse: ratio1.0137[1.0035,1.0241], with both A/A controls resolved. Four-worker Go interpreter is also adverse, ratio1.0338[1.0020,1.0878], with unresolved controls.10/40 controls resolve overall. The allocation decrease is about64KiB/request; it does not establish a timing win. These mixed results do not justify maintaining another pipe protocol for this contribution.

All1600 throughput and40 separate resource processes complete, with716800 warm responses,80 qualification and7 actual-loader probes. Both Rust service/composed fixtures, owned Go/LLGo interpreter/native, actual Wasmtime and bare host controls use the same driver. Fixed20 blocks,16 clients,4 requests/client,1/4 workers,2 warmups/7 measured batches,64KiB, seed2026111006,8000 paired-log bootstrap draws. Every response byte, status, header and trailer is checked. No outcome-based retiming, dropped adverse cell, phase subtraction or pooling with preceding events. Sources, binaries, fixtures, compilers and loaded libraries are checked.

Local complete Go tests, P3 race/vet and pinned LLGo complete P3/measurement tests pass on both backends;31 owned coverage gates pass, minimum90.86%. Tests cover partial/empty reads, whole-write backpressure, close/error precedence against io.Pipe, partial-close counts, blocked wakeups and concurrent readers/writers. Linux prototype validation was not run; the prototype is rejected before a contribution, not a completed supported-platform feature. Its source, tests, bundle and complete data remain recoverable.

The separate Go allocation/CPU diagnostic reports441310ns/op,674996B/op and2917allocations/op at GOMAXPROCS4. It includes request construction and complete-response validation, and CPU collection includes preparation/calibration. Darwin synchronization frames are not a pure per-request cost partition. It is not substituted for the fixed primary event. Allocation hotspots still include external read buffers, io.ReadAll and packed Canonical StoreBuffer staging. The next candidate uses bounded reuse and the existing immutable-memory primitive while retaining the standard pipe.

## All owned comparisons

| Profile | Base req/s | Prototype req/s | Warm ratio [95%] | Both A/A resolve | Base B/req | Prototype B/req |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 1562 | 1541 | 1.0137 [1.0035,1.0241] | yes | 755654 | 690251 |
| world-rust-composed-http/go/interpreter/4 | 6138 | 6106 | 1.0134 [0.9958,1.0372] | no | 755630 | 690227 |
| world-rust-composed-http/go/native/1 | 2058 | 2101 | 0.9680 [0.9498,0.9817] | no | 752051 | 686650 |
| world-rust-composed-http/go/native/4 | 7272 | 7369 | 0.9788 [0.9562,0.9968] | no | 752016 | 686614 |
| world-rust-composed-http/llgo/interpreter/1 | 1014 | 1108 | 0.9099 [0.8886,0.9364] | no | 1222196 | 1154138 |
| world-rust-composed-http/llgo/interpreter/4 | 2092 | 2177 | 0.9698 [0.9442,1.0032] | no | 1222384 | 1154346 |
| world-rust-composed-http/llgo/native/1 | 1173 | 1275 | 0.9078 [0.8822,0.9297] | no | 1218775 | 1150655 |
| world-rust-composed-http/llgo/native/4 | 2064 | 2163 | 0.9557 [0.9350,0.9751] | no | 1219017 | 1150949 |
| world-rust-http/go/interpreter/1 | 1651 | 1628 | 1.0118 [1.0028,1.0200] | no | 739789 | 674381 |
| world-rust-http/go/interpreter/4 | 6406 | 6303 | 1.0338 [1.0020,1.0878] | no | 739759 | 674349 |
| world-rust-http/go/native/1 | 2219 | 2237 | 0.9909 [0.9807,1.0008] | no | 740448 | 675041 |
| world-rust-http/go/native/4 | 7467 | 7844 | 0.9512 [0.9308,0.9697] | no | 740414 | 675005 |
| world-rust-http/llgo/interpreter/1 | 1060 | 1152 | 0.9109 [0.8775,0.9375] | no | 1200477 | 1132407 |
| world-rust-http/llgo/interpreter/4 | 2109 | 2169 | 0.9679 [0.9490,0.9889] | no | 1200642 | 1132572 |
| world-rust-http/llgo/native/1 | 1208 | 1330 | 0.9226 [0.8953,0.9613] | no | 1201785 | 1133678 |
| world-rust-http/llgo/native/4 | 2101 | 2168 | 0.9660 [0.9328,1.0004] | no | 1201994 | 1133898 |

## Every profile and control

| Profile | req/s | Complete API us | Runtime B/req | Between-batch threads | Peak RSS MiB | A/A95% | Resolved |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| base/world-rust-composed-http/go/interpreter/1 | 1562 | 637.647 | 755654 | 7 | 108.641 | [0.9954,1.0136] | yes |
| base/world-rust-composed-http/go/interpreter/4 | 6138 | 629.648 | 755630 | 7 | 149.312 | [0.9949,1.0216] | yes |
| base/world-rust-composed-http/go/native/1 | 2058 | 483.459 | 752051 | 7 | 132.328 | [0.9788,1.0089] | yes |
| base/world-rust-composed-http/go/native/4 | 7272 | 528.029 | 752016 | 7 | 160.141 | [0.9682,1.0131] | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 1014 | 979.114 | 1222196 | 43 | 168.734 | [0.9832,1.0090] | yes |
| base/world-rust-composed-http/llgo/interpreter/4 | 2092 | 1880.046 | 1222384 | 61 | 181.062 | [0.9924,1.0483] | no |
| base/world-rust-composed-http/llgo/native/1 | 1173 | 842.384 | 1218775 | 43 | 180.344 | [0.9645,1.0232] | no |
| base/world-rust-composed-http/llgo/native/4 | 2064 | 1908.087 | 1219017 | 61 | 229.891 | [0.9936,1.0998] | no |
| base/world-rust-http/go/interpreter/1 | 1651 | 603.244 | 739789 | 7 | 81.391 | [0.9970,1.0270] | no |
| base/world-rust-http/go/interpreter/4 | 6406 | 606.444 | 739759 | 7 | 101.438 | [0.9890,1.0590] | no |
| base/world-rust-http/go/native/1 | 2219 | 448.169 | 740448 | 8 | 109.156 | [1.0046,1.0265] | no |
| base/world-rust-http/go/native/4 | 7467 | 520.767 | 740414 | 7 | 126.688 | [0.9659,1.0056] | no |
| base/world-rust-http/llgo/interpreter/1 | 1060 | 933.152 | 1200477 | 43 | 140.172 | [0.9528,1.0145] | no |
| base/world-rust-http/llgo/interpreter/4 | 2109 | 1865.953 | 1200642 | 61 | 140.797 | [0.9953,1.0464] | no |
| base/world-rust-http/llgo/native/1 | 1208 | 813.890 | 1201785 | 43 | 168.750 | [0.9988,1.0611] | no |
| base/world-rust-http/llgo/native/4 | 2101 | 1876.306 | 1201994 | 61 | 175.766 | [0.9708,1.0268] | no |
| head/world-rust-composed-http/go/interpreter/1 | 1541 | 646.419 | 690251 | 7 | 104.969 | [0.9948,1.0158] | yes |
| head/world-rust-composed-http/go/interpreter/4 | 6106 | 633.437 | 690227 | 7 | 137.797 | [0.9685,1.0019] | no |
| head/world-rust-composed-http/go/native/1 | 2101 | 473.394 | 686650 | 7 | 148.781 | [1.0025,1.0298] | no |
| head/world-rust-composed-http/go/native/4 | 7369 | 523.437 | 686614 | 7 | 158.688 | [0.9692,1.0019] | no |
| head/world-rust-composed-http/llgo/interpreter/1 | 1108 | 894.196 | 1154138 | 43 | 168.828 | [0.9602,1.0100] | no |
| head/world-rust-composed-http/llgo/interpreter/4 | 2177 | 1808.698 | 1154346 | 61 | 181.078 | [0.9553,1.0114] | no |
| head/world-rust-composed-http/llgo/native/1 | 1275 | 777.895 | 1150655 | 43 | 199.859 | [0.9878,1.0184] | yes |
| head/world-rust-composed-http/llgo/native/4 | 2163 | 1814.272 | 1150949 | 61 | 201.234 | [0.9892,1.0823] | no |
| head/world-rust-http/go/interpreter/1 | 1628 | 611.566 | 674381 | 7 | 78.391 | [0.9902,1.0100] | yes |
| head/world-rust-http/go/interpreter/4 | 6303 | 615.478 | 674349 | 7 | 104.812 | [0.9215,0.9983] | no |
| head/world-rust-http/go/native/1 | 2237 | 444.364 | 675041 | 7 | 116.688 | [0.9820,1.0002] | yes |
| head/world-rust-http/go/native/4 | 7844 | 494.571 | 675005 | 7 | 120.594 | [0.9857,1.0269] | no |
| head/world-rust-http/llgo/interpreter/1 | 1152 | 861.375 | 1132407 | 43 | 140.531 | [0.9768,1.0081] | yes |
| head/world-rust-http/llgo/interpreter/4 | 2169 | 1819.311 | 1132572 | 61 | 141.219 | [0.9675,1.0131] | no |
| head/world-rust-http/llgo/native/1 | 1330 | 740.775 | 1133678 | 43 | 152.219 | [0.9548,1.0259] | no |
| head/world-rust-http/llgo/native/4 | 2168 | 1807.870 | 1133898 | 61 | 176.188 | [0.9654,1.0321] | no |
| host/native-go/go/host/1 | 40476 | 21.044 | 277914 | 6 | 16.672 | [0.9681,1.0131] | no |
| host/native-go/go/host/4 | 45145 | 75.677 | 277917 | 6 | 16.969 | [0.9756,1.0328] | no |
| host/native-go/llgo/host/1 | 8179 | 116.698 | 732572 | 37 | 68.109 | [0.9752,1.0355] | no |
| host/native-go/llgo/host/4 | 8965 | 399.419 | 732583 | 40 | 68.375 | [0.9575,1.0528] | no |
| reference/world-rust-composed-http/go/compiler/1 | 14972 | 64.368 | 32 | 8 | 32.375 | [0.9796,1.0126] | yes |
| reference/world-rust-composed-http/go/compiler/4 | 51519 | 72.820 | 32 | 8 | 33.266 | [0.9859,1.0430] | no |
| reference/world-rust-http/go/compiler/1 | 17396 | 55.136 | 32 | 7 | 29.719 | [0.9964,1.1710] | no |
| reference/world-rust-http/go/compiler/4 | 59543 | 62.862 | 39 | 7 | 30.734 | [1.0023,1.0444] | no |

All lifecycle intervals,1600 process reductions, resource snapshots and32 paired reference comparisons remain in the machine record; raw result SHA256 `3b263700bff35a930cb222daba45f999c0febd9bb18c3c1f5dc5741bd6226e1f` in `artifacts/p3-chunk-pipe-20261110`. Go managed/LLGo collector scopes differ, and Go counters omit Rust/JIT allocation. This uncontrolled workstation event closes no performance or release gate.

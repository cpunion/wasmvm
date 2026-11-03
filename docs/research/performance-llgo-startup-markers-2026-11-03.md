# LLGo startup: first goroutine and Darwin symbol metadata

The independent [startup report](performance-llgo-startup-2026-11-03.md)
now has a narrower reproduction for [LLGo #2973](https://github.com/xgo-dev/llgo/issues/2973).
On the pinned compiler, the first goroutine with the public runtime traceback
hook triggers expensive symbol metadata initialization. Darwin's default
DWARF/embedded-PCLN configuration suppresses entry sites and falls back to
dynamic symbol lookup. This is an upstream runtime/compiler investigation;
the VM's release profile remains unchanged.

Apple M4 Max/macOS arm64; Go 1.27.3; LLGo
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, `-O3 -lto=full`.
Each campaign has five shuffled serial fresh-process blocks, with all builds
completed before timing. All samples, including large first-launch values,
are retained. The shared host is not controlled performance acceptance.

## Markers and source controls

A C constructor records CLOCK_MONOTONIC_RAW. The first main statement takes
the same clock and publishes both timestamps through C. This distinguishes
launch-to-ready from the interval after that constructor. The constructor
need not precede other loader constructors; its interval is not a complete
loader profile. No intervals from different programs or campaigns are subtracted.

The initial common Cgo source with a blank `os` import gives Go 4.275 ms and
LLGo 46.322 ms launch-to-ready, with constructor-to-main medians 0.629 ms and
42.083 ms. A separate campaign changes LLGo's calls to direct C declarations:
Go Cgo 5.442 ms, LLGo Cgo 65.398 ms, LLGo direct 69.994 ms ready. The latter's
constructor-to-main is 63.430 ms. The gap therefore survives removal of normal
Cgo wrappers; these are different source variants, not an equivalent Go/LLGo ABI.

In a third campaign, the direct source without `os` reaches main in 0.092 ms
after the constructor. Adding a channel/goroutine handshake costs 0.293 ms
from the main marker through publication. In a fourth campaign:

| LLGo direct source | Constructor to main | Main marker to publish | Ready |
| --- | ---: | ---: | ---: |
| Blank public runtime import | 1.224 ms | 0.000042 ms | 7.571 ms |
| Same import plus first goroutine handshake | 1.552 ms | 38.266 ms | 49.320 ms |
| Blank os import | 67.553 ms | 0.000042 ms | 76.076 ms |

The handshake captures the first goroutine/channel operation together, not pure
thread creation. Separate debug-enabled runs show metadata initialization for
the public-runtime handshake and os variants, but none for runtime-only/no-os.
Intrusive one-shot LLDB stops also localize the largest interval to `os.init`
through main; their debugger overhead excludes them from timing comparisons.
The initial debugger output with ambiguous inline-frame names is retained;
the corrected trace records the requested breakpoint identity.

## Same-source metadata profile diagnostic

The final campaign keeps the direct os source identical. The default binary
is unchanged; two additional binaries change one explicit build option each.
These are diagnostic profile changes, not a VM benchmark or a proposed fix.

| Profile, in addition to -O3 -lto=full | Constructor to main | Ready |
| --- | ---: | ---: |
| Default DWARF, embedded PCLN | 67.004 ms | 79.441 ms |
| `-ldflags=-w` | 2.641 ms | 10.687 ms |
| `-pclntab=external`, DWARF retained | 1.930 ms | 12.271 ms |

Separate `LLGO_FUNCINFO_DEBUG=1` runs report one function frame with
`entries=dlsym` under the default profile and 900 with `entries=prebuilt` under
`-w`. The external profile produces its `.pclntab` sidecar and no such debug
line. All file hashes and raw samples remain available. These alternate
profiles implicate the metadata path; they do not isolate every constituent
cost, preserve every packaging/debugger contract, or establish an upstream fix.

## Source mechanism and upstream action

The pinned sources connect the observations:

1. Go 1.27.3's `os.newFile` calls `runtime.SetFinalizer`. LLGo's
   [`SetFinalizer` initializes its worker goroutine](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/runtime/internal/lib/runtime/mfinal.go#L59)
   on first use, before main in this os-importing source.
2. [`registerTraceback`](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/runtime/internal/runtime/traceback_threads_native.go)
   invokes the public runtime's creator hook on goroutine creation.
3. [`tracebackCreator`](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/runtime/internal/lib/runtime/traceback_threads_llgo.go#L15)
   eagerly calls `frameSymbol` while selecting a visible creator frame.
4. [`frameSymbolUncached` and metadata initialization](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/runtime/internal/lib/runtime/symtab.go)
   can build function and PC-line lookup tables and resolve missing entries
   through dynamic symbol lookup.
5. [`shouldEmitRuntimeEntrySites`](https://github.com/xgo-dev/llgo/blob/8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f/internal/build/funcinfo_table.go#L732)
   explicitly suppresses Darwin embedded-PCLN entry sites when DWARF is enabled
   to preserve LLDB's initial lexical scope. The prebuilt-table post-link rewrite
   is gated by the same predicate.

A separate one-shot LLDB creator breakpoint captures this exact call chain:
`os.init → os.newFile → runtime.SetFinalizer → initFinalizerState → NewProc
→ registerTraceback → tracebackCreator`. Optimized inline `AllocZ` frames are
also present. This verifies the trigger without treating the debugger's
1.1-second marker interval as startup performance.

Current-main source `043c98b0f6636de5cb2380fc0955399f7133ad0f` was read:
the [entry-site gate](https://github.com/xgo-dev/llgo/blob/043c98b0f6636de5cb2380fc0955399f7133ad0f/internal/build/funcinfo_table.go#L732)
and [eager creator symbolization](https://github.com/xgo-dev/llgo/blob/043c98b0f6636de5cb2380fc0955399f7133ad0f/runtime/internal/lib/runtime/traceback_threads_llgo.go#L15)
remain. Current main was not rebuilt or timed, so its performance status is not
asserted. The existing issue is updated rather than duplicated. A fix should
preserve accurate created-by frames, runtime stack APIs, LLDB scope and normal
embedded deployment, while avoiding full symbol resolution on first creation.
Disabling traceback capture or silently changing the benchmark profile is not
adopted.

[Runnable sources, commands, binary/compiler identities, every campaign and independent summary verification](evidence/performance-llgo-startup-markers-2026-11-03.json.gz)

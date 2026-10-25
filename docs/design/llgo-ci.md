# Go / LLGo 兼容性与 CI 设计

核查日期：2026-10-26。本文区分源码约束、已经观察到的工具行为与后续 VM 验收要求；探针通过不代表已实现 Wasm。

## 两种兼容性分别验收

| 目标 | 编译与执行路径 | 验收含义 |
| --- | --- | --- |
| LLGo 编译宿主 VM | `llgo test ./...`、以后 `llgo build ./cmd/wasmvm`，生成并运行本机程序 | 与官方 Go 编译器使用相同 VM 源码、相同语义测试与 fixture |
| VM 执行 LLGo 生成的 guest | `llgo build -target wasi -o fixture.wasm …`，随后由 VM 加载执行 | 满足该 guest 实际要求的 Core 指令与 WASI ABI |

第一项是项目从初始化就设置的 CI 门槛。第二项是解释器、WASI P1、线程、EH、SIMD 实现后的集成里程碑。宿主通过 LLGo 编译不会自动获得第二项能力；同样，执行官方 Go 的 `wasip1/wasm` 输出也不能证明可以执行 LLGo guest。

## 工具链与源码语言基线

- 源码设计以 Go 1.24 语言和标准库范围为保守基线，避免 Go 1.27 新语法。`go.mod` 声明的语言版本不等于 CI bootstrap 工具链版本，也不等于已在全部旧版工具链验收。
- 首批 CI 的两条路径都选择 Go **1.27.3** 的 GOROOT，确保标准库输入一致。LLGo **v1.0.8** 本身的 `go.mod` 要求 `go 1.27.0`，因此本次采用的 source 安装不能用 Go 1.24 bootstrap。
- 固定 LLGo revision：`c171abfa130108bc6451e01fcf15a8a3474eb0a7`（核查时 `v1.0.8` 指向此 commit）。固定 setup action：`4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91`。LLVM major 为 **22**；LLVM 和系统依赖的补丁版本仍由安装源决定，首期没有声称工具链完全可重现。
- 当前 workflow 先覆盖 Ubuntu 24.04 amd64；后续扩展 macOS arm64、Linux arm64，再验收 Windows。不要因为安装 action 支持某平台，就将该平台列为 VM 已支持。
- 独立增加 Go 1.24 的实际兼容性任务、锁定 LLVM 补丁版和工具链镜像，是正式声明最低宿主版本前的工作。当前锁定的 CI Go 版本是已执行环境的版本，源码基线仍需旧版本实测确认。

证据：[LLGo v1.0.8 go.mod](https://github.com/xgo-dev/llgo/blob/c171abfa130108bc6451e01fcf15a8a3474eb0a7/go.mod)、[v1.0.8 release](https://github.com/xgo-dev/llgo/releases/tag/v1.0.8)、[setup-llgo action 定义](https://github.com/xgo-dev/setup-llgo/blob/4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91/action.yml)、[setup-llgo 输入与平台说明](https://github.com/xgo-dev/setup-llgo/blob/4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91/README.md)。本次也通过 GitHub API 核对了 checkout/setup-go 的 v7 commit，并将 SHA 固定在 workflow 中。

## 可移植实现约束

1. 首期执行器、validator、内存和对象管理使用普通 Go；不依赖 gc 编译器的私有 ABI、汇编调用桥、`runtime` 私有符号或 `go:linkname`。确有平台差异时收敛到 `internal/platform`，使用 `//go:build llgo` / `!llgo`，保持上层 API 和测试一致。
2. 不把 Go 对象地址编码为 `uintptr` 长期保存，也不把 Go 指针藏进 `[]byte`、数值 value slot 或 guest linear memory。需要宿主 GC 保持可达的对象必须由真实类型化引用或显式 root registry 持有。若采用句柄管理 guest GC 对象，句柄表的释放、代际校验和根扫描属于 VM 的职责。
3. 不以 finalizer 执行时机、宿主 GC 回收精确时机或某个编译器的堆统计推导 guest 语义。LLGo 原生默认使用 conservative BDWGC，官方 Go 使用不同的运行时；guest WasmGC 对象生命周期、资源关闭和配额必须有明确边界。
4. 使用 `encoding/binary` 和显式位操作解释 little-endian guest memory；浮点位模式用 `math.Float32bits` / `Float64bits`。默认不以 `unsafe.Pointer` 强转 byte slice 做原子访问，避免对齐、宿主端序和 GC 根问题。无锁快速路径必须在参考内存实现与边界测试之后引入。
5. Wasm atomic 指令的语义由 VM 内存层实现；Go 的 `sync/atomic` 是候选宿主机制，不直接等同于完整 Wasm 内存模型。混合原子/普通访问、重叠宽度、`memory.grow`、wait/notify 注册与唤醒必须统一设计，保证 VM 本身没有 Go data race。
6. LLGo 原生 goroutine 与 OS thread 为 1:1，使用固定原生栈。限制 VM worker 数、guest thread 数和递归深度；解释器使用显式 frame stack，避免 guest 无限递归耗尽宿主栈。不能假定宿主 goroutine 可以无限廉价创建。

依据：[本地快照对应 README 的运行时与源码兼容说明](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/README.md)、[gc 专属实现差异分类](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/test/goroot/notapplicable.yaml)。以上实现约束是本项目设计决策，不表示普通 `unsafe` 或标准库 API 在 LLGo 中一概不可用。

## 当前 LLGo guest 的必要能力

本地 LLGo checkout 位于 `38bfd05ee102e36710a2df5bca47834a622b667c`（2026-10-16）。其 W32 文档约定如下；升级 guest 编译器时必须重新提取模块 imports、features 和 ABI：

| 能力 | 对新 VM 的影响 |
| --- | --- |
| Memory32 + WASI Preview 1 | 实现 `wasi_snapshot_preview1`；不能用 P2/Component Model 支持替代 |
| 共享 `env.memory` + `wasi.thread-spawn` | 支持共享线性内存、原子操作、线程创建 ABI；每个线程的 instance/stack/TLS 所有权明确 |
| 标准 Wasm EH | 支持非 legacy EH，验证 panic/recover、跨函数展开与线程隔离 |
| SIMD | 支持模块实际使用的 SIMD 指令；默认 runner 明确开启 SIMD |
| 线性内存中的 guest GC | 这是 LLGo 自己的 STW conservative collector；VM 不扫描其 Go 对象。不得与 WasmGC 的 struct/array/type/ref 支持混为一谈 |
| 线程退出 | worker 返回走 wasi-libc pthread cleanup/TLS destructor 路径；不要求 WASIX `thread_exit` |

现有文档明确将 WasmGC、W64、JSPI、WASI P2/components 列为独立工作。该版本的线性 GC 遇到无法到达 safepoint 的 C 调用，会在超时后放弃本次回收；这属于 guest runtime 的进展限制，新 VM 仍必须正确执行其内存和调度语义。

证据：[WASI threads 执行约定](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-wasi-threads.md)、[线程 GC 约定](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-threaded-gc.md)、[Wasm profile/ABI 提案](https://github.com/xgo-dev/llgo/blob/38bfd05ee102e36710a2df5bca47834a622b667c/doc/wasm-proposal.md)。这些来源描述 LLGo guest，本项目目前尚未实现或执行该 guest。

## CI 门槛

可执行配置见 [ci.yml](../../.github/workflows/ci.yml)。两个 compiler job 都实际执行 `test -count=1 -timeout=10m ./...`，任意一个失败都使 workflow 失败，不设置 `continue-on-error`。

| 检查 | 官方 Go | LLGo |
| --- | --- | --- |
| 共同 package tests、将来的 VM fixture | 必须运行 | 必须运行 |
| 核心纯 Go 约束 | `CGO_ENABLED=0 go test` | 默认 LLGo 原生环境运行；安装其 C runtime 依赖 |
| 格式、静态检查 | `gofmt`、`go vet` | 共享同一份源码检查结果 |
| 宿主 data race 检查 | `go test -race` | 暂不将其是否接受 `-race` 视为已验证的检测能力 |
| Wasm spec / WASI / LLGo guest corpus | 按能力里程碑逐步接入 | 同一 fixture、同一断言，不用全包 skip 隐藏差异 |

`CGO_ENABLED=0` 只作用于 Go 基础测试步骤，不应设置为整个 workflow 的环境变量。LLGo 的 LLVM、BDWGC 等依赖不影响“VM 源码不依赖 cgo 引擎”的目标。当前 scaffold 没有依赖 lockfile，故关闭 action cache；引入依赖后再按 `go.sum`、compiler revision、LLVM/OS/arch 设置缓存。

当前仓库兼容探针覆盖：类型化原子与 goroutine/锁同步、位模式和 little-endian、引用环跨显式 GC 保持可达、条件变量唤醒和 context 取消。后续 VM 测试必须覆盖共享内存增长、lost wakeup、取消传播、GC 根在 host call/栈扩容期间的有效性、EH 展开以及资源关闭。对 gc/LLGo 差异保留最小复现与 issue，不能通过跳过整个功能建立“兼容”结论。

## 本地验证事实与命令

2026-10-26 的独立临时探针已在两条命令下通过：

```sh
go test -v -count=1 ./...
llgo test -v -count=1 ./...
```

探针实际检查了 4 个 goroutine 通过 `atomic.Uint64` 累加到 4000、`sync.WaitGroup` 等待、little-endian uint64 往返、IEEE-754 负零位模式、环状对象在 `runtime.GC()` 后仍可达。探针目录为 `/var/folders/wb/8zzrvxd56294wxlxxbpws9440000gn/T/gwasm-llgo-probe-6zzrcjr4`；它是调查时的临时验证，项目自己的持续测试以仓库内容为准。

该临时调查环境：Go `go1.27.3 darwin/arm64`；`/Users/lijie/go/bin/llgo` 报 `llgo (devel) darwin/arm64`。`go version -m` 显示该二进制来自 `8ac217053d0337b7261cebc923bb6b17955bf49a` 且有 dirty 标记，`llgo env LLGO_ROOT` 却指向上述 2026-10-16 checkout。这只是调查环境的观察，不能作为固定 v1.0.8 CI 或完整 VM 的验收结果。

**新仓库实际测试环境不同**：在 `wasmvm` 工作目录中，shell 解析到 `/opt/homebrew/bin/llgo`，它是 Homebrew wrapper，实际执行 `/opt/homebrew/Cellar/llgo/1.0.8/libexec/bin/llgo`，报告 `llgo v1.0.8 darwin/arm64`，LLGO_ROOT 为同一 Cellar 安装。项目自身的五项兼容探针已用此安装和 Go 1.27.3 通过；Go race/vet 与 actionlint 也通过。验证工具链时必须记录工作目录、PATH 解析和实际 executable，不能沿用另一个 checkout 的环境。详细记录见 [validation.md](../validation.md)。CI 则从指定 revision 安装 compiler/runtime，并核对 action 输出的 revision。

在项目根目录的建议验证命令：

```sh
go version
llgo version
llgo env LLGO_ROOT GOROOT
# llgo 若为 wrapper，先找到实际二进制，再读取 Go build info。
CGO_ENABLED=0 go test -count=1 -timeout=10m ./...
llgo test -count=1 -timeout=10m ./...
go test -race -count=1 -timeout=10m ./...
```

远端 CI 尚需在仓库有 remote 并推送之后实际运行；仅编写 workflow 不等于远端任务已通过。

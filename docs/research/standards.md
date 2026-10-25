# WebAssembly / WASI 标准与提案调研

调研日期：**2026-10-26（Asia/Shanghai）**。这是本日实际读取的一手资料快照，不把旧路线图当成已经发生的发布，也不把提案成熟度当成某个引擎的默认支持情况。本文件中的优先级和工程方案是本项目的设计判断。

## 结论

建议以 **Core Wasm 3.0 为长期语义目标，以 WASI 0.1 + legacy wasi-threads 为首个实用兼容目标**，随后交付 WasmGC、Component Model / WASI 0.2，再实现已发布的 WASI 0.3。GC 和多线程从第一版数据结构中预留，但分阶段验收；初期支持“GC 模块”和“共享线性内存线程”两种能力，不承诺跨线程共享 WasmGC 对象。

截至调研日，GC、typed function references、tail call、最终版 exception handling、memory64、multi-memory 和 relaxed SIMD 已进入 Core 3.0。Threads 仍为 Phase 4；Shared-Everything Threads 为 Phase 1；Stack Switching 为 Phase 3。不能继续把 GC 写成实验性早期提案，也不能把 shared GC heap 写成已标准化线程能力。[完成提案表](https://github.com/WebAssembly/proposals/blob/main/finished-proposals.md)、[活动提案表](https://github.com/WebAssembly/proposals/blob/main/README.md)。

WASI 0.3.0 已于 **2026-06-11** 发布，最新查到的稳定补丁为 **0.3.2（2026-10-13）**；0.2 最新查到的补丁为 **0.2.12（2026-06-02）**。WASI 0.3 不应再列为“尚未发布”。[0.3.0 发布记录](https://github.com/WebAssembly/WASI/releases/tag/v0.3.0)、[0.3.2 发布记录](https://github.com/WebAssembly/WASI/releases/tag/v0.3.2)、[0.2.12 发布记录](https://github.com/WebAssembly/WASI/releases/tag/v0.2.12)。

## 1. “标准化”的三个不同层次

| 层次 | 本次观察 | 项目应如何记录 |
| --- | --- | --- |
| Core 提案阶段 | Phase 4 表示交由 WG 完成标准化；Phase 5 表示 WG 已达成完成共识，随后并入主规范 | FeatureSet 按具体特性开关，不靠一个“最新版本”布尔值 |
| Core 语义版本 | 官方 2025-09-17 宣布 Wasm 3.0 完成；当前 live spec 页标注 Release 3.0，2026-10-23 | 固定规范 / 测试集 commit，声明具体通过的特性集合 |
| W3C 发布状态 | 当前 TR 是 **Candidate Recommendation Draft，2026-10-23**；2019-12-05 的 1.0 是 Recommendation | 不把“Core 3.0 已完成”改写为“3.0 已成为 W3C Recommendation” |

来源：[阶段流程](https://github.com/WebAssembly/meetings/blob/main/process/phases.md)、[3.0 完成公告](https://webassembly.org/news/2025-09-17-wasm-3.0/)、[当前 Core 文本](https://webassembly.github.io/spec/core/)、[W3C 当前 TR](https://www.w3.org/TR/wasm-core/)、[W3C 发布系列页](https://www.w3.org/TR/wasm-core/all/)。TR 状态说明还明确其 living standard 维护方式；“Candidate Recommendation”与引擎是否适合生产使用是不同问题。

**另一个常见混淆：** Core 3.0 的模块二进制头仍是版本 `1`，不是把头改为 `3`；特性通过类型、指令、段及验证规则扩展。Component 二进制另有封装与版本规则，应由独立 decoder 处理。[Core 二进制模块格式](https://webassembly.github.io/spec/core/binary/modules.html#binary-module)、[Component binary format](https://github.com/WebAssembly/component-model/blob/main/design/mvp/Binary.md)。

## 2. Core / embedding 提案矩阵

以下覆盖官方 `WebAssembly/proposals` 在本次快照列出的完成、活动及 inactive 项目。Phase 0 的开放 issue 不等于受理的完整提案集合，因此只监控，不枚举为已承诺特性。

阶段来源固定为 [proposals commit `df6a2aca`](https://github.com/WebAssembly/proposals/tree/df6a2aca1cd9d779eb4f2f3e19dae8277f336015)，对应活动 / 完成表文件最近修改时间 2026-10-06。交叉核对 [官方 feature data](https://github.com/WebAssembly/website/blob/main/features.json)，本次读取显示文件修改于 2026-10-21。

### 2.1 已完成并进入 Core 或 embedding 规范的项目

下表“阶段”均为完成提案；版本和受影响规范见 [官方完成表](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/finished-proposals.md)。

| 特性 | 版本 / 层次 | 本项目取舍 |
| --- | --- | --- |
| MVP | Core 1.0 | 第一阶段执行与验证基础 |
| Import/Export of Mutable Globals | Core 1.0 | 第一阶段 |
| Non-trapping float-to-int | Core 2.0 | 第一阶段，注意 NaN 与饱和边界 |
| Sign-extension operators | Core 2.0 | 第一阶段 |
| Multi-value | Core 2.0 | 第一阶段，栈帧从一开始允许多返回值 |
| Reference Types | Core 2.0 | 第一阶段；`externref` 与 `funcref` 生命周期显式化 |
| Bulk memory operations | Core 2.0 | 第一阶段；支持 passive segments / data count |
| Fixed-width SIMD | Core 2.0 | 完成 Core 2.0 声明前实现；可先标量解释 |
| JS BigInt ↔ Wasm i64 | JS API 2.0 | 不属于独立 Go VM 的必要宿主 API |
| Tail call | Core 3.0 | 早期 IR 保留尾调用；不可用普通递归伪装 |
| Extended Constant Expressions | Core 3.0 | 与 initializer validator 一起实现 |
| Typed Function References | Core / JS API 3.0 | 与递归类型、GC 类型系统一起实现 |
| Garbage collection | Core / JS API 3.0 | 必做；单独的 WasmGC 里程碑 |
| Multiple memories | Core / JS API 3.0 | 从初期保存 memidx，避免单内存架构锁定 |
| Relaxed SIMD | Core 3.0 | SIMD 后实现，并明确确定性 profile |
| Custom Annotation Syntax | 文本格式 3.0 | WAT 工具链阶段；不阻塞二进制执行 |
| Branch Hinting | Core custom section 3.0 | 可识别并忽略，不改变执行语义 |
| Exception handling | Core / JS API 3.0 | 实现最终 `try_table` / `exnref` 形式 |
| JS String Builtins | Core / JS embedding 3.0 | 记录为 embedding 扩展，按需求另做 |
| Memory64 | Core 3.0 | 内部地址先用 `uint64`，独立开启执行支持 |

### 2.2 活动提案

每行的阶段均来自 [固定快照活动表](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/README.md)。项目链接提供该项原始定义；优先级为本项目判断。

| 提案 | Phase | 本项目取舍 |
| --- | ---: | --- |
| [JS Promise Integration](https://github.com/WebAssembly/js-promise-integration) | 5 | JS embedding 相关；独立 Go VM 无需模拟 Promise |
| [Web Content Security Policy](https://github.com/WebAssembly/content-security-policy) | 5 | Web embedding 相关 |
| [Threads](https://github.com/WebAssembly/threads) | 4 | 必做：shared memory、atomics、wait/notify |
| [Compact Import Section](https://github.com/WebAssembly/compact-import-section) | 4 | 后续 decoder 兼容；独立 feature gate |
| [Wide Arithmetic](https://github.com/WebAssembly/wide-arithmetic) | 4 | 后续数值扩展，结构上易隔离 |
| [ESM Integration](https://github.com/WebAssembly/esm-integration) | 3 | JS embedding 相关 |
| [Stack Switching](https://github.com/WebAssembly/stack-switching) | 3 | 栈帧设计预留 continuation；默认关闭 |
| [Custom Page Sizes](https://github.com/WebAssembly/custom-page-sizes) | 3 | 内部避免到处硬编码 page shift；实验开关 |
| [Custom Descriptors and JS Interop](https://github.com/WebAssembly/custom-descriptors) | 3 | GC 后跟进；不纳入首版 GC 声明 |
| [Relaxed dead code validation](https://github.com/WebAssembly/relaxed-dead-code-validation) | 2 | 暂不放宽 validator；按快照实验 |
| [Numeric Values in WAT Data Segments](https://github.com/WebAssembly/wat-numeric-values) | 2 | 文本工具层跟进 |
| [Extended Name Section](https://github.com/WebAssembly/extended-name-section) | 2 | 调试信息可选解析 |
| [Rounding Variants](https://github.com/WebAssembly/rounding-mode-control) | 2 | 实验数值扩展 |
| [Compilation Hints](https://github.com/WebAssembly/compilation-hints) | 2 | 解释器先忽略提示，JIT 时评估 |
| [JS Primitive Builtins](https://github.com/WebAssembly/js-primitive-builtins) | 2 | JS embedding 相关 |
| [Acquire-Release Atomics](https://github.com/WebAssembly/acquire-release-atomics) | 2 | 不与现有 SC atomics 混淆；后续实验 |
| [Multibyte Array Access](https://github.com/WebAssembly/multibyte-array-access) | 2 | WasmGC 后再评估 |
| [FP16](https://github.com/WebAssembly/half-precision) | 2 | SIMD / 数值扩展实验 |
| [Type Imports](https://github.com/WebAssembly/proposal-type-imports) | 1 | 类型注册表避免绑定单一模块；不承诺实现 |
| [Component Model](https://github.com/WebAssembly/component-model) | 1 | 必做其已发布 WASI 子集，不能以整体 Phase 1 为由忽略 |
| [C and C++ API](https://github.com/WebAssembly/wasm-c-api) | 1 | 首版提供 Go API；C ABI 适配以后再做 |
| [Flexible Vectors](https://github.com/WebAssembly/flexible-vectors) | 1 | 不作为 `v128` 首版范围 |
| [Memory control](https://github.com/WebAssembly/memory-control) | 1 | 当前资源配额属于 embedding，不假装提案实现 |
| [Reference-Typed Strings](https://github.com/WebAssembly/stringref) | 1 | 不等于 WasmGC；单独跟踪 |
| [Profiles](https://github.com/WebAssembly/profiles) | 1 | 更广泛 profile 提案；与 Core 3.0 已有确定性约束分开 |
| [Shared-Everything Threads](https://github.com/WebAssembly/shared-everything-threads) | 1 | 预留共享标记和 heap ownership，稳定范围外 |
| [Frozen Values](https://github.com/WebAssembly/frozen-values) | 1 | GC 后研究 |
| [More Array Constructors](https://github.com/WebAssembly/more-array-constructors) | 1 | GC 后研究 |
| [JIT Interface](https://github.com/WebAssembly/jit-interface) | 1 | 不作为本 VM 实现内部 JIT 的前置条件 |
| [Type Reflection for JS API](https://github.com/WebAssembly/js-types) | 1 | JS embedding 相关 |
| [JS Text Encoding Builtins](https://github.com/WebAssembly/js-text-encoding-builtins) | 1 | JS embedding 相关 |

Phase 5 但仍留在活动表的项目表示 WG 已完成，尚未并入主规范；这与“finished 表中的完成提案”位置差异应保留，不能机械地认为活动表全部不稳定。

### 2.3 Inactive / 被替代项目

下列清单来源：[固定快照 inactive 表](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/inactive-proposals.md)。

| 项目 | 状态 / 替代方向 | 项目决定 |
| --- | --- | --- |
| Unmanaged closures | Typed Function References 替代 | 不实现旧方案 |
| Conditional sections | 转向 Feature Detection，后者也撤回 | 不实现旧方案 |
| Feature Detection | 缺少共识而撤回 | 使用宿主 capability API |
| Module Linking | 转向 Component Model | 采用 Component 路线 |
| Interface Types | 转向 Component Model | 采用 WIT / Canonical ABI 路线 |
| Funclets | 转向 multiloop 方向 | 不采用旧编码 |
| Call Tags | inactive | 观察 |
| Constant Time | inactive | 不宣称恒定时间执行 |
| JS Customization for GC Objects | Custom Descriptors 替代 | 不实现旧方案 |
| Instrument and Tracing Technology | inactive | 使用自己的调试接口，不冒称提案兼容 |

另需显式记录 **legacy EH**：官方 feature data 将旧异常格式列为 inactive；最终 EH explainer 将旧文档单独链接。旧 `try/catch/rethrow/delegate` 不等于 Core 3.0 的 `try_table`、`throw_ref` 和 `exnref`。[最终 EH 文档](https://github.com/WebAssembly/exception-handling/blob/main/proposals/exception-handling/Exceptions.md)、[旧 EH 文档](https://github.com/WebAssembly/exception-handling/blob/main/proposals/exception-handling/legacy/Exceptions.md)。若 llgo 实际产物仍使用旧编码，做显式兼容 feature，分别测试，不混写“EH 支持”。

## 3. WASI 与 Component Model

### 3.1 已发布版本和 API 面

| 版本 | 当前地位 | 二进制 / ABI | 工程目标 |
| --- | --- | --- | --- |
| WASI 0.1 / Preview 1 | legacy、广泛部署 | core module，`wasi_snapshot_preview1`，WITX | 首个实际可运行目标 |
| WASI 0.2 / Preview 2 | 稳定，已被 0.3 接续；最新查到 0.2.12 | Component、WIT、资源 handles、Canonical ABI、`wasi:io` polling | 首个组件里程碑 |
| WASI 0.3 / Preview 3 | 当前稳定系列；最新查到 0.3.2 | 原生 `async func`、`future<T>`、`stream<T>`；移除 `wasi:io` 包 | 独立的异步组件里程碑 |
| WASI 1.0 | 没有查到已发布 1.0 | 不以未来语义作为当前验收标准 | 跟踪 |

来源：[0.1 官方说明](https://wasi.dev/releases/wasi-p1)、[0.2 官方说明](https://wasi.dev/releases/wasi-p2)、[0.3 官方说明](https://wasi.dev/releases/wasi-p3)、[WASI 发布页](https://github.com/WebAssembly/WASI/releases)。0.1 不提供通用 connect/listen 和 HTTP；需要完整网络能力时应实现对应 0.2/0.3 接口，而不是把自定义 socket 扩展命名为标准 WASI。

“WASI 支持”必须变成 **版本 + world/interface + 操作覆盖 + 测试结果**。建议依次验收 CLI、clocks/random、filesystem、poll/I/O、sockets，再验收 HTTP；只跑通 hello world 不构成整个版本的支持。0.2 的 `wasi:cli/command`、`wasi:http/proxy` 与 0.3 的 `wasi:cli/command`、`wasi:http/service` / `middleware` 分别处理。[0.2 API](https://wasi.dev/releases/wasi-p2)、[0.3 发布说明](https://github.com/WebAssembly/WASI/releases/tag/v0.3.0)。

### 3.2 Component Model Phase 1 与稳定子集并不矛盾

Component Model 总体仍处于提案流程 Phase 1，但 WASI 已发布版本依赖的子集具有工具链兼容承诺。按 WASI 版本固定编码和语义，避免直接把整个 Component main 分支当成稳定 ABI。[Component README](https://github.com/WebAssembly/component-model/blob/main/README.md)。

| WASI 起始版本 | 需要的 Component 特性 |
| --- | --- |
| 0.2.0 | 当时默认、非 gated 的 Component 特性 |
| 0.3.0 | 再加 async lift/lower、future、stream |
| 0.3.1 | 再加 `map<K,V>`、`implements` / `external-id` 注解 |
| 0.3.2 | 累积上述能力；没有从发布记录观察到新增采用项 |

来源：[正式采用记录](https://github.com/WebAssembly/WASI/blob/main/docs/ComponentModelFeatures.md)、[0.3.1 发布记录](https://github.com/WebAssembly/WASI/releases/tag/v0.3.1)。这些特性即使暂未出现在某个 WASI interface 内，也属于声明兼容对应版本的要求。

以下仍为 gated、不能混入稳定默认实现：threading built-ins、基于 Shared-Everything Threads 的 built-ins、stackful async lift、fixed-length lists、error-context、canonical interface names、Component memory64 等。[Component gated features](https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#gated-features)。尤其注意：**Core memory64 已完成，不代表 Component Canonical ABI 的 memory64 已稳定。**

### 3.3 WASI API 提案概览

阶段依据：[WASI proposals](https://github.com/WebAssembly/WASI/blob/main/docs/Proposals.md)，本次读取修改时间 2026-09-08。WASI 的 phase 流程和 Core 流程不是同一个发布承诺，不可仅比较数字。

| Phase | 官方列表中的 API / 扩展 | 本项目范围 |
| --- | --- | --- |
| 3 | Clocks、Random、Filesystem、Sockets、CLI、HTTP | 按 0.2 / 0.3 发布接口实现 |
| 2 | Timezone、HTTP Informational Outbound Response、I2C、Key-value Store、wasi-nn、Runtime Config、WebGPU、Messaging | 作为可选 host plugin 候选，不纳入初始标准承诺 |
| 1 | Blob Store、Crypto、GPIO、Distributed Lock、Logging、Observe、Parallel、Pattern Match、SPI、SQL、SQL Embed、Threads、TLS、URL、USB、OTel | 仅跟踪；成熟度、版本与稳定 WASI 主包分开 |
| 4 / 5 | 当前表无条目 | 不称 WASI 0.3 已经是 1.0 / WG 最终标准 |
| 0 | 当前仓库表没有正式列项 | 不将网页或旧路线图零散链接视作稳定 API |

### 3.4 本次发现的资料不同步

- [wasi.dev roadmap](https://wasi.dev/roadmap) 仍把 2026-10-13 的 0.3.2 标为 Planned，但 [GitHub 0.3.2](https://github.com/WebAssembly/WASI/releases/tag/v0.3.2) 已有非预发布 release 和 WIT artifact。本报告采用发布事实。
- [WASI API 列表](https://github.com/WebAssembly/WASI/blob/main/docs/Proposals.md) 的 Threads 链接使用 `wasi-native-threads`；本日访问该地址重定向到 [legacy wasi-threads](https://github.com/WebAssembly/wasi-threads)。不能由链接名称推导出独立且成熟的新线程 API。
- [wasi-testsuite README](https://github.com/WebAssembly/wasi-testsuite/blob/main/README.md) 仍有 “forthcoming Preview 3” 用语；这不能推翻已存在的 0.3 release。测试接入时读取实际测试目录和固定版本，不只抄 README。
- W3C 系列封面部分简介仍写 2.0，而当前规范正文为 3.0；本报告分别引用正文版本和 publication status，不从 URL 中的 `core-2` 推断语义版本。

## 4. GC、线程、异步的边界

### 4.1 三种“GC”需要分别命名

1. **VM 宿主 GC**：Go 或 llgo 对实现自身 Go 对象的内存管理。
2. **guest 线性内存 GC**：例如语言运行时在 Wasm linear memory 内自己维护的堆；VM 看到的是字节。
3. **WasmGC**：Core 指令 / 类型语义中的 struct、array、i31、递归类型、子类型和引用生命周期。

实现 1 不意味着实现 3；能运行带语言 GC 的 guest 也不代表能运行 WasmGC 模块。WasmGC 与 linear memory 是两个存储域。[Core 3.0 GC 说明](https://webassembly.org/news/2025-09-17-wasm-3.0/)、[GC proposal](https://github.com/WebAssembly/gc)。

工程建议：先实现非移动、精确标记清扫的 guest heap；引用通过可检查的 handle / slot 表表示，明确 store ownership 和 generation，避免将 Go 指针编码进 `uint64` 后依赖宿主保活。操作数栈、帧 locals、globals、tables、异常 payload、host handles 都属于 roots；尾调用、异常展开、host callback、取消和实例销毁都必须维护 root 生命周期。是否采用宿主对象承载部分数据可后续测量，但不能让 Go GC 的扫描规则替代 Wasm 类型和生命周期规则。

### 4.2 三种线程能力

| 能力 | 标准边界 | 项目实现路线 |
| --- | --- | --- |
| Core Threads | shared linear memory、SC atomics、wait/notify；创建 / join 由 embedder 决定 | 实现内存模型与等待队列，线程生命周期独立成层 |
| legacy wasi-threads | Preview 1 扩展；instance-per-thread，进入 `wasi_thread_start` | 显式可选 ABI，用于现有 llgo / C / Rust 产物 |
| Shared-Everything Threads | 共享函数 / 表 / globals / GC 引用、TLS、GC 原子操作等；仍 Phase 1 | 设计保留 shared 标记；稳定功能集暂不开放 |

来源：[Core Threads overview](https://github.com/WebAssembly/threads/blob/main/proposals/threads/Overview.md)、[legacy wasi-threads](https://github.com/WebAssembly/wasi-threads/blob/main/README.md)、[Shared-Everything overview](https://github.com/WebAssembly/shared-everything-threads/blob/main/proposals/shared-everything-threads/Overview.md)。

legacy `wasi-threads` 已有 wasi-libc pthreads 及 Wasmtime / WAMR / Wasmer / toywasm 实现经验，但它仍是 legacy Phase 1 扩展，不能因此称它为 WASI 0.2/0.3 的标准 pthreads。ABI 的 spawn 参数、返回值和 child entry 必须按固定版本文档测试。[ABI 定义](https://github.com/WebAssembly/wasi-threads/blob/main/wasi-threads.abi.md)、[实现列表](https://github.com/WebAssembly/wasi-threads#what-has-been-implemented)。

核心设计要求：

- 线程使用独立 execution context / stack，复用 immutable compiled module；共享内存与进程级 WASI context 不等于共享所有 globals、tables 或 GC refs。
- shared memory 的普通 load/store 与原子操作都必须经内存访问层，不能一部分直接访问 Go `[]byte`、另一部分用 atomic，从而引入宿主数据竞争。先实现正确的锁或字宽一致策略，之后再优化。
- 需要为重叠访问宽度、unaligned ordinary accesses、atomic alignment trap、越界、grow、wait 超时、notify 与取消竞态建立具体语义；不能把“用了 goroutine”作为实现线程标准的证据。
- wait 必须保证检查预期值与注册等待者的协议正确，避免丢失唤醒；退出能唤醒所有等待线程。`proc_exit` / 任一 thread trap 需要按 legacy 进程语义终止整个线程组；应在有界执行检查点和可取消 host 调用上兑现。
- child 实例建立不能随意重复执行父实例初始化逻辑；遵循 ABI 与 guest 约定，重点检查 active data segment、start section、stack/TLS global。需要专门的重新实例化 fixture。
- 在 Shared-Everything Threads 实验实现通过之前，跨线程传递非共享 WasmGC ref 必须被拒绝。并行运行互相隔离的 GC 实例仍然可行。

以上是本项目设计要求，不是对各引擎当前实现的性能判断。

### 4.3 WASI 0.3 async 不等于 OS 多线程或 Stack Switching

WASI 0.3 使用组件调用的异步任务、future、stream 语义；Core Stack Switching 提供 continuation 指令；Shared-Everything Threads 解决共享对象和线程模型。三者相关但互不等价。WASI 0.3 稳定采用记录未要求所有未稳定线程 / stackful 特性。[采用记录](https://github.com/WebAssembly/WASI/blob/main/docs/ComponentModelFeatures.md)、[Stack Switching explainer](https://github.com/WebAssembly/stack-switching/blob/main/proposals/stack-switching/Explainer.md)。

设计上应保留 `Execution` 的显式帧、暂停 / 恢复状态和 host-call 边界；组件异步调度后端可以先使用 Go 的同步机制实现，但外部契约必须按 Canonical ABI，而不是直接把 guest future 映射为任意 goroutine。

## 5. 现有实现给本项目的验证锚点

这部分只说明可用于交叉验证的来源；详细引擎调查另见同目录的实现调研。

| 领域 | 可查证实现 / 工具 | 使用方式 |
| --- | --- | --- |
| Core 3.0 | 官方 spec reference interpreter 与 testsuite；Wasmtime；浏览器引擎 | 固定 fixture，对比结果、trap、类型接受 / 拒绝 |
| Core threads | 浏览器、Wasmtime、WAMR 等 | 同步原语及内存语义回归；不能只比较有数据竞争程序单次输出 |
| legacy wasi-threads | wasi-libc + Wasmtime / WAMR / Wasmer / toywasm | 运行固定 SDK 构建的 pthread fixture |
| WASI 0.2 | Wasmtime、jco | 标准接口和组件 ABI 对照 |
| WASI 0.3 | 官方页面列 Wasmtime 46+、jco | 按确切 WIT 版本固定 producer 和 consumer |

来源：[官方引擎特性矩阵](https://webassembly.org/features/)、[spec 仓库](https://github.com/WebAssembly/spec)、[wasi-threads 实现说明](https://github.com/WebAssembly/wasi-threads#what-has-been-implemented)、[WASI 0.2](https://wasi.dev/releases/wasi-p2)、[WASI 0.3](https://wasi.dev/releases/wasi-p3)。特性矩阵可能把 flag、部分后端或仅解析实现写在单元格备注中；采用前需回到引擎源码 / 配置确认，不能抹掉这些条件。

## 6. 里程碑与可验收范围

下表是建议顺序，没有给尚未测量的工期或性能数字。每个阶段的同一测试语料都由 Go 和 llgo 编译出的 VM 执行。

| 阶段 | 交付范围 | 退出条件 |
| --- | --- | --- |
| S0：设计与编译器基线 | 包边界、FeatureSet、限制模型、兼容性探针、双编译器 CI | 文档可追溯；两个编译器执行 smoke，不能冒称已有 VM |
| S1：可用解释器 | MVP + 常用 2.0 特性、资源限额、基本 P1 CLI | decoder / validator 正反例；最小 WASI 程序；trap / exit / timeout 明确分型 |
| S2：P1 与共享内存线程 | P1 操作覆盖、Core Threads、可选 legacy wasi-threads | 官方 P1 测试、pthread 创建 / join / TLS / wait-notify / trap 和 exit 全线程终止；Go race 审计 |
| S3：WasmGC 与 3.0 控制流 | recursive types、subtyping、GC、typed refs、tail calls、最终 EH | spec GC / refs / EH 测试；深尾递归无宿主栈增长；跨模块类型、GC roots 与 callback 回归 |
| S4：补齐 Core 3.0 | SIMD / relaxed SIMD、memory64、multi-memory、extended const 等 | 明确列出完整或缺失能力；只有全套范围通过才宣称 Core 3.0 |
| S5：Component + WASI 0.2 | component decode / validate / instantiate、WIT 类型、Canonical ABI、resources、CLI 到 sockets/HTTP | 固定 0.2.12 WIT；跨语言组件；own/borrow/drop、字符串 / list 编码、错误路径 |
| S6：WASI 0.3 | async lift/lower、future/stream、取消 / 背压 / reentrancy；0.3.1 后累计特性 | 固定 0.3.2 WIT；CLI 与 HTTP service / middleware；异步任务 / 资源泄漏及双编译器一致性 |
| S7：实验提案 | stack switching、shared-everything、Component threads 等 | 独立 feature gate、固定提案 commit、独立兼容性报告 |

S2 与 S3 可在 S1 后并行；类型系统、引用布局与内存访问抽象必须在 S0/S1 先确定。性能优化 / JIT 应放在解释器语义和两种编译器一致性已有可靠证据之后。

## 7. 测试来源与更新规则

- Core 语义优先使用 [WebAssembly/spec](https://github.com/WebAssembly/spec) 和 [WebAssembly/testsuite](https://github.com/WebAssembly/testsuite)，Threads、Stack Switching 等未合入特性使用各提案测试目录。记录来源 commit、生成工具版本与许可证。
- WASI 使用 [WebAssembly/wasi-testsuite](https://github.com/WebAssembly/wasi-testsuite)，legacy threads 加 [专用测试](https://github.com/WebAssembly/wasi-threads/tree/main/test)。Component 使用 [Component test suite](https://github.com/WebAssembly/component-model/tree/main/test) 并补跨语言 golden fixtures。
- 对未完成特性做**显式拒绝**测试；暂不支持不是静默接受后执行错误。`unsupported feature`、malformed、validation failure、link failure、trap 与 WASI exit 应可区分。
- 合规报告记录 `compiler × OS/arch × feature profile × spec/WIT revision`。包含 unsupported、known failure、skip 原因与到期复查，不能用排除大量用例的通过率替代兼容声明。
- 每次准备升级 profile 时重新读取官方提案表、release tag、WIT 和对应测试 commit；不随 CI 自动跟随 main。研究记录中保留“观察日期”，后续修改追加变更原因。

# 现有实现与复用判断

调研日期：2026-10-26。以下是当日官方仓库、源码和文档的观察快照，**不是本项目复跑后的认证结果**。`main` 文档可能领先已发布版本，状态也可能随目标架构、编译后端及构建开关变化。正式选型必须锁定版本并执行 M0 对照实验。

## 结论

已有成熟的纯 Go Wasm runtime（wazero），也已出现声明支持 WasmGC / Core 3 的纯 Go 实现 Wago。不能以“Go 生态没有实现”为立项理由。本项目的可验证目标应是：**由 Go 和 llgo 编译相同 VM 代码，提供明确的 WasmGC、共享线性内存线程和 WASI 版本契约，并能用相同语料证明两种宿主编译器的行为一致**。

建议先完成 M0：评估 wazero 解释器、Wago 的 llgo 移植代价与可复用部分，再决定独立实现的边界。现阶段保留独立解释器设计；不预设长期 fork 整个现有引擎，也不把外部 Rust/C runtime 的 Go 包装当作“Go 编写的 VM”。

## 能力矩阵

“支持”只指列明的范围。`未证实` 表示本次没有找到可据以承诺的官方实现证据，不能据此推断全生态不存在。Core threads、WASI `thread-spawn`、Component Model 并发是三个不同层次。

| 实现 | 主体/Go 集成 | WasmGC | Core threads / 创建线程 ABI | WASI P1 | WASI P2 / Component Model | WASI P3 |
| --- | --- | --- | --- | --- | --- | --- |
| wazero | Go，无 cgo；解释器 + 原生代码后端 | main 尚无 GC feature；解释器 GC PR 开放中 | experimental CoreFeaturesThreads；内建 wasi-threads 未证实 | 有内建实现 | main 内建实现未证实；有第三方上层库 | 未证实 |
| Wago | Go，无 cgo；单遍原生代码后端；beta | 官方功能表声明支持，需按后端与对象边界复测 | Core threads 需开启；功能表明确 wasi-threads 不支持 | 独立实验插件 | 独立实验插件；支持 P2 command、资源与 Canonical ABI，网络默认拒绝 | 功能表列为计划；组件插件有实验异步机制 |
| wippyai/wasm-runtime | Go Component 层，建立在 wazero 上 | 不能超出底层引擎能力承诺 | 本次未单独确认 | 沿底层能力 | README 声明 WIT、Canonical ABI、P2、asyncify | 未证实 |
| pgavlin/warp、Life、wagon | Go 历史实现/工具包 | 未找到现代 GC 完整支持证据 | 未找到现代线程完整支持证据 | warp 有 wasi 包；其他须逐项确认 | 未证实 | 未证实 |
| Wasmtime | Rust；wasmtime-go 用 cgo 调 C API | 官方矩阵 Tier 1；Winch 存在覆盖例外 | Core threads Tier 2，有已知资源管理限制；另有 wasi-threads | 有 | 有；Component Model 默认启用 | 已有实现；wasi crate 源码仍标为实验、不稳定、不完整 |
| WAMR | C；Go language-binding 是 native runtime 绑定 | 可选构建；官方明确尚不完全符合 GC/Core 3 | shared memory、pthread、wasi-threads 均有实现 | 有 | 本次未证实原生组件执行/P2 | 未证实 |
| WasmEdge | C++；Go 通过 native API 绑定 | 0.16 起默认；解释器/AOT 有实现 | Core threads 可选开启；不能由此推导 wasi-threads | 有 WASI host | 0.17 起 Component **loader/validator**；不等于组件执行 | 未证实 |
| Wasmer | Rust；wasmer-go 用 cgo | 当前原生编译器 feature/validator 不启用 GC 指令 | threads、WASIX pthread、wasi-threads 有实现 | 有 | 未证实可据以承诺的原生 P2/Component 执行 | 未证实 |
| V8 / SpiderMonkey / JavaScriptCore | C++ 浏览器引擎；非 Go VM | 已发布 | shared memory + atomics；Web Worker 提供生命周期 | 由 embedding/shim 提供，不是 core 自动附带 | JS 工具层与原生引擎能力须分别评估 | 不据 core 支持作承诺 |

各行的具体证据和限制见下文；本表不是所有平台、所有 backend 的笛卡尔积认证。

## wazero：最有价值的 Go 基准与复用候选

官方仓库已由 `tetratelabs/wazero` 迁移至 `wazero/wazero`；模块路径仍为 `github.com/tetratelabs/wazero`。README 承诺 Core 1/2，一套解释器和一套原生编译后端，Apache-2.0。这里的“纯 Go”主要表示没有 cgo，并不等于没有汇编或运行时约定。[README](https://github.com/wazero/wazero/blob/main/README.md)

本次源码快照：`ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7`。稳定 feature 集合停留于 Core 2；实验集合已列出 threads、tail-call、extended-const、exception-handling、typed-function-references，不能继续使用“wazero 没有线程/EH”的旧结论。线程注释还明确 guest 原子操作与 host Memory API 能力有差别。[稳定 features](https://github.com/wazero/wazero/blob/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7/api/features.go)、[实验 features](https://github.com/wazero/wazero/blob/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7/experimental/features.go)

[GC PR #2494](https://github.com/wazero/wazero/pull/2494) 在调研时仍 open，声明实现递归类型、struct/array、相关 validator 和解释器指令，并利用宿主 Go GC；优化编译器不在该 PR 范围内。PR 不能当作主干或发布版已支持的证据。[Core 3 跟踪 #2426](https://github.com/wazero/wazero/issues/2426) 同样仍开放。

llgo 评估先走解释器：decoder、validator、WASI host、spec runner 和测试组织值得学习。原生后端存在 amd64/arm64 汇编入口，移植必须验证调用 ABI、栈切换、Go runtime 边界，不能仅靠替换编译命令。[amd64 ABI 入口](https://github.com/wazero/wazero/blob/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7/internal/engine/wazevo/backend/isa/amd64/abi_entry_amd64.s)

复用收益：成熟的 embedding API、WASI P1 行为经验、错误分类、测试语料。主要成本：GC 会穿透类型系统、值表示、验证器和执行器；大量核心实现位于 `internal`，正常导入无法直接复用，需要贡献上游、提取可复用库或有限代码导入。长期 fork 的维护成本必须纳入 M0。

## Wago：应在自研前认真比较的新实现

Wago 公开为 beta，采用 Go 编写的 decoder/validator 和单遍 Railshot 原生代码后端。WASI 与 Component Model 位于独立插件。许可证为 Apache-2.0；当前 go.mod 声明 Go 1.22，依赖 `golang.org/x/sys`。[README](https://github.com/wago-org/wago/blob/main/README.md)、[LICENSE](https://github.com/wago-org/wago/blob/main/LICENSE)、[go.mod](https://github.com/wago-org/wago/blob/main/go.mod)

本次代码树快照：`4de27acba12926e52b815e5845e035d1cf11edd0`。其[功能契约](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/FEATURES.md)声明 Core 3，包括 GC、typed refs、EH、memory64、multi-memory；Core threads 需显式开启。该表同时明确：shared-everything threads 不支持，stack switching 计划中，wasi-threads 不支持，P3 计划中。不能把其线程能力写成“已有 WASI pthread 支持”。

官网摘要与仓库契约有粒度差异，选型以具体源码和测试为准。例如 [WASI 插件 README](https://github.com/wago-org/wasi/blob/main/README.md)把 P1/P2 标为 experimental，P2 支持 `wasi:cli/command` 导入面，但网络调用返回 typed access-denied；legacy unstable 已移除。仓库中“0.3 breaking changes”指 **插件自身版本**，不是支持 WASI 0.3。[Component 插件](https://github.com/wago-org/component-model/blob/main/README.md)有 P2、嵌套组合、resources 和实验 tasks/futures/streams；这些机制仍不能直接等价为完整 WASI P3。

测试证据强于单纯宣传，但尚需复跑：[SPECTEST](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/SPECTEST.md)报告固定 Core 3 测试集 258 个文件、2,226 个模块、58,038 个断言，所列目标为 linux/amd64、linux/arm64、darwin/arm64。固定上游为 `WebAssembly/spec` 的 `wg-3.0` / `9d36019973201a19f9c9ebb0f10828b2fe2374aa`，WABT 1.0.41 配合 reference interpreter 处理 text 格式。文档保留历史阶段限制，不能把历史红表当现状，也不能用总数代替逐特性审查。[机器可读 baseline](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/tests/conformance/baselines/spec-v3-baseline.json)有总计，但 `files` 为空且各 family 计数为零；M0 应生成自己的逐文件/逐命令报告，复核汇总与真实执行一致。

llgo 的具体障碍不是 Go 语法，而是运行时契约：

- [标准 Go amd64 trampoline](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/src/core/runtime/trampoline_amd64.s)使用 Plan 9 汇编、`NOSPLIT`、外部栈，保存 `g` 所在寄存器并恢复 ABIInternal 的 X15 约定。其构建条件只有 `!tinygo`，尚无 llgo 分支证据。
- [TinyGo trampoline](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/src/core/runtime/trampoline_tinygo_amd64.go)另行生成机器码，并依赖 TinyGo 函数值内存表示和参数寄存器；TinyGo 路径提供可移植思路，不能直接作为 llgo 路径。
- [架构文档](https://github.com/wago-org/wago/blob/4de27acba12926e52b815e5845e035d1cf11edd0/ARCHITECTURE.md)还涉及 off-heap 栈、signal guard、W^X 映射、GC root maps 和宿主 collector 假设。即使汇编能被 llgo 处理，也要独立验证回调、GC 根、阻塞和取消。

因此 Wago 是潜在代码复用/移植候选和差分基准，不应被忽略；其公开数据尚不能证明 llgo 可用、线程与 GC 任意组合成立，或所有 WIT/WASI interfaces 都已完成。后续性能比较必须自己固定机器、版本、配置和工作负载，不能直接复用其相对 wazero 的宣传倍数。

## 其他 Go 项目

[wippyai/wasm-runtime](https://github.com/wippyai/wasm-runtime/blob/main/README.md)声明纯 Go Component Model、WIT、Canonical ABI、P2 host、asyncify、WAT 编译，底层依赖 wazero。它说明“Go 中的组件执行”已经有可调研实现；也说明组件层可以与 Core 引擎分离。README 自述不足以证明完整 conformance，更不能继承尚未发布的 wazero GC 能力。可在 M0 评估组件 decoding、资源表和 ABI 实现，导入前单独审查许可证及测试。

[pgavlin/warp](https://github.com/pgavlin/warp)有 Go interpreter/compiler、WAT/WAST、wasi 等目录，适合作为工具组织参考。它与 Wago 架构文档引用的 **wago-org/warp（C++）不是同一个项目**。[Life](https://github.com/perlin-network/life/blob/master/README.md)提供纯 Go 解释器和实验 AOT，README 的测试声明仍为 66/72。不能拿这些历史测试数量证明今天的 Core 3 / GC / Component Model 合规。[wagon](https://github.com/go-interpreter/wagon)亦列为历史参考，当前新特性不作承诺。

## 非 Go 引擎：参照实现与差分测试

**Wasmtime** 是现代规范与 Component Model 的首选对照。官方[proposal 矩阵](https://docs.wasmtime.dev/stability-wasm-proposals.html)将 GC、typed refs、EH 列为 Tier 1（部分 Winch 例外），threads 为 Tier 2；共享内存与 Store limits/pooling 的整合仍有限制。Stack switching 为 Tier 3 工作中，shared-everything threads 未实现。Component async 默认开启，但组件 threading 等功能仍单独 gating，不能将它们合为一项“多线程已完成”。

Wasmtime 的 [wasi crate 源码](https://github.com/bytecodealliance/wasmtime/blob/main/crates/wasi/src/lib.rs)分别提供 p1/p2/p3，且仍对 p3 写有实验、不稳定、不完整的限定。[运行组件文档](https://component-model.bytecodealliance.org/running-components/wasmtime.html)说明可以运行 P3 ABI；两者应综合理解为“已有可运行实现，支持边界仍须锁版本核实”。[wasmtime-go](https://github.com/bytecodealliance/wasmtime-go/blob/main/README.md)通过 cgo 调 C API，Go binding 的覆盖也不能自动等于 Rust API。

**WAMR** 已迁移到 `wasm-micro-runtime/wasm-micro-runtime`。[README](https://github.com/wasm-micro-runtime/wasm-micro-runtime/blob/main/README.md)明确包含 shared memory、pthread、wasi-threads 和 Go 绑定；[构建契约](https://github.com/wasm-micro-runtime/wasm-micro-runtime/blob/main/doc/build_wamr.md)却进一步限定 GC 默认关闭且尚不完全符合规范，EH 仍为 legacy 版本、仅 classic interpreter。应使用它检查 P1/thread-spawn 集成，不能把 README 的 feature 列表当完整 Core 3 认证。

**WasmEdge** 官方[proposal 表](https://wasmedge.org/docs/start/wasmedge/extensions/proposals/)分别列出解释器、AOT、默认开关：GC 从 0.16 默认；Core threads 需要开关；Component Model 自 0.17 仅 loader/validator。适合 GC/EH 的第二独立对照，但组件解析成功不能当作 P2/P3 运行成功。

**Wasmer** 的[feature 集合](https://github.com/wasmerio/wasmer/blob/main/lib/types/src/features.rs)和[原生编译器 validator](https://github.com/wasmerio/wasmer/blob/main/lib/compiler/src/compiler.rs)未启用 Wasm GC 指令；其中 `GC_TYPES` 不等于 `GC`。后端及 V8 集成可能另有差别，须分别测试。[WASIX FAQ](https://wasix.org/docs/explanation/faqs/)说明 wasi-threads 与 WASIX 线程 ABI 并存。WASIX 是扩展生态，不能写成 WASI P2/P3 的同义词。[wasmer-go](https://github.com/wasmerio/wasmer-go/blob/master/README.md)为 cgo/native library 包装。

**浏览器引擎** 证明 GC 的生态成熟度，而非 WASI host 能力：[V8](https://v8.dev/blog/holiday-season-2023)已发布 GC，[Mozilla](https://bugzilla.mozilla.org/show_bug.cgi?id=2013741)记录从 Firefox 120 起的 WasmGC 实现，[WebKit Safari 18.2](https://webkit.org/blog/16301/webkit-features-in-safari-18-2/)加入 GC。线程采用 Worker 与 SharedArrayBuffer 的宿主组合，[MDN 共享内存说明](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/SharedArrayBuffer#webassembly_shared_memory)明确 atomics 与共享内存的条件。浏览器上的 Component/WASI 通常还要工具和 host 层；本项目不计划把整个 JS 引擎嵌入 Go。

## M0：复用决策门与退出标准

M0 是架构实施前的实验，不要求现在就导入依赖或开发完整 VM。至少比较下列路径：

| 路径 | 能减少的工作 | 必须验证的成本 | 当前判断 |
| --- | --- | --- | --- |
| 扩展/移植 wazero 解释器 | Core 1/2、P1、宿主 API、成熟测试 | GC 类型/值/验证器改造；llgo 编译；internal 边界；补丁维护 | 优先实验 |
| 移植 Wago 或提取其模块 | Core 3 已有实现；GC/compiler/测试基础 | ABI、foreign stack、collector、signals、线程组合；beta API 变化 | 与 wazero 并行实验 |
| 新建可移植 Go 解释器，选择性复用 | 明确的 llgo 契约；GC/线程从类型模型设计 | decoder/validator/数值语义/测试覆盖投入最大 | 默认设计保留，M0 后确定范围 |
| 包装 Wasmtime/WAMR | 最快得到现成能力，可作为 oracle | Rust/C/FFI/toolchain、binding 覆盖，不满足主体 Go 目标 | 仅作测试/比较，不作为核心方案 |

M0 应交付一份可重现报告：

1. 固定候选 commit、许可证、Go/llgo 版本、平台；先用公共 API 执行 `add`、trap、宿主回调、P1 hello，再跑 GC 循环对象/强制回收、threads wait/notify、共享 memory.grow 的小语料。
2. 两个宿主编译器跑同一测试集；记录 build-only、实际执行、pass/fail/skip，而不是“能 go build 就兼容”。普通 Go 的 race/fuzz 结果不冒充 llgo 的 race/fuzz 结果。
3. 每条失败确定边界：Go 语言/stdlib、程序集成、宿主 ABI、GC 根、共享内存行为、未实现提案；给出 patch 规模和是否可上游化。llgo 内存管理应以真实工具链实现核对。
4. 同一预编译 fixture 对照 Wasmtime；P1 threads 另用 WAMR/Wasmer 验证。GC、threads、EH 单测通过后再测组合，避免由单项支持推出组合正确。
5. 复用官方 `WebAssembly/spec` / testsuite、WASI WIT 与测试集的固定版本。可借鉴 Wago 的 WAST 转换方式及 wazero 的 binary fixtures；不复制其 pass 数值、忽略列表或内部测试结果作为本项目结果。保留来源、许可证、工具版本和 fixture 生成过程。
6. 输出 ADR：若候选能以可维护补丁满足 Go+llgo、GC、线程和版本化 WASI 目标，则优先复用；若核心 ABI/GC/所有权模型需大面积重写，选择独立解释器并具体列出保留的代码/工具/测试资产。两种结果都保留 `wasmvm` 项目的独立 API 和一致性契约。

许可证边界：本次已核实 wazero/Wago 的 Apache-2.0、Life 的 MIT，warp 仓库标记 BSD-3-Clause，WAMR 为 Apache-2.0 WITH LLVM-exception。只阅读和借鉴设计并未导入代码。后续复制任何源码、测试或 fixture 时，必须逐项保留其许可证/NOTICE；不能因为引擎是宽松许可证就推导其第三方 corpus 也同许可。

## 尚未完成的验证

本轮没有编译或运行第三方引擎，没有对上述候选执行 llgo 兼容性认证，没有复跑 benchmark 或完整 conformance，也没有确认所有候选的最新发布 tag 与 main 一致。用户要求的当前阶段是调研与设计；这些实验是下一阶段决策门，不能在 README 中写成已有能力。

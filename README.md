# wasmvm

使用 Go 编写、同时由 Go 和 llgo 编译的 WebAssembly 虚拟机项目。

**当前处于调研与设计阶段，没有可执行 Wasm 的 VM 实现。** 仓库包含标准与实现调研、架构与目录设计、路线图，以及两种宿主编译器共用的兼容性探针和 CI。包目录中的 `doc.go` 只声明计划职责，不表示对应功能已支持。

## 目标

- 以纯 Go 可移植解释器为第一条执行路径，逐步覆盖 WebAssembly Core 3.0，包括 Wasm GC、标准异常处理、SIMD、Memory64 等。
- 支持 Core Threads 的共享线性内存、原子操作和 wait/notify；为现有 llgo 程序实现独立的 legacy `wasi-threads` 适配层。
- 按版本分别实现 WASI Preview 1、基于 Component Model 的 WASI 0.2，以及带原生异步的 WASI 0.3。
- 相同源代码、相同语义测试分别通过 `go test` 和 `llgo test`。VM 的宿主编译器兼容性与 llgo 生成的 guest Wasm 兼容性分开验收。

截至 **2026-10-26**，WASI 0.3 已发布；GC 已进入 Core 3.0；Core Threads 与共享 GC 对象的 Shared-Everything Threads 仍是不同提案。纯 Go 生态已有 wazero 和与本目标高度重合的 Wago，不能以“没有现有实现”为自研依据。

## 文档入口

| 文档 | 内容 |
| --- | --- |
| [标准调研](docs/research/standards.md) | 提案阶段、Core/WASI 版本、官方来源和过时资料辨析 |
| [现有实现](docs/research/implementations.md) | Go 与其他语言运行时能力矩阵、Wago/wazero 复用决策 |
| [总体架构](docs/design/architecture.md) | 执行模型、类型系统、GC、线程、WASI、安全边界和 API 草案 |
| [目录设计](docs/design/packages.md) | `pkg`/`internal` 分层、依赖方向、包职责与落地顺序 |
| [Go/llgo 与 CI](docs/design/llgo-ci.md) | 双编译器约束、固定工具链、宿主与 guest 测试区别 |
| [实施路线](docs/roadmap.md) | M0 复用评估、里程碑、验收与风险 |
| [验证记录](docs/validation.md) | 本次实际运行结果与尚未验证的范围 |

架构决策：以独立纯 Go 解释器作为当前设计基线，**M0 先验证 Wago/wazero 的复用价值，再决定正式实现来源**；暂不引入 JIT、LLVM 绑定或外部 VM 作为生产核心。llgo 用作宿主编译器，不等于 VM 需要依赖 LLVM API。

## 本地验证

```sh
go test -count=1 ./...
llgo test -count=1 ./...
# 同一入口也可以指定宿主编译器：
./scripts/check.sh go
./scripts/check.sh llgo
```

当前测试仅验证未来运行时需要的宿主行为：数值位模式、显式引用可达性、goroutine/锁/原子操作、条件变量及取消。通过测试不代表 Wasm、WASI 或线程规范符合性；后续验收矩阵见路线图。

项目路径：`github.com/cpunion/wasmvm`。尚未选择项目许可证；不在本阶段复制第三方实现代码，后续复用必须单独核对许可证与署名要求。

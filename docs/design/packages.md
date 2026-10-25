# 包与目录设计

状态：设计草案，2026-10-26。**只有标为“已建”的文件存在；树中其余路径是实施计划。** 实际公共包目前只有 `doc.go`，没有 VM API 或执行能力。

## 目录树

```text
wasmvm/
├── go.mod                         # 已建；github.com/cpunion/wasmvm
├── README.md                      # 已建；项目状态和入口
├── .github/workflows/ci.yml        # 已建；Go/llgo 同一套测试
├── scripts/
│   ├── check.sh                   # 已建；本地编译器选择入口
│   └── update-corpus.sh           # 计划；人工触发、固定版本、记录来源
├── docs/                          # 已建
│   ├── research/{standards,implementations}.md
│   ├── design/{architecture,packages,llgo-ci}.md
│   ├── roadmap.md
│   └── validation.md
├── cmd/
│   └── wasmvm/                    # 计划；run/inspect/validate CLI
├── pkg/
│   ├── README.md                  # 已建；公共包总览
│   ├── wasm/                      # 已建 doc.go；公开类型、值、FeatureSet
│   ├── vm/                        # 已建 doc.go；Engine/Store/Linker/Instance
│   ├── component/                 # 已建 doc.go；组件编译/链接/执行 API
│   └── wasi/
│       ├── preview1/              # 已建 doc.go；wasi_snapshot_preview1
│       ├── threads/               # 已建 doc.go；legacy wasi-threads 可选扩展
│       ├── preview2/              # 已建 doc.go；WASI 0.2 worlds
│       └── preview3/              # 已建 doc.go；WASI 0.3 worlds
├── internal/
│   ├── compat/                    # 已建；两编译器公共宿主探针
│   ├── binary/                    # 计划；有界 reader、LEB128
│   ├── core/
│   │   ├── module/                # 计划；内部 module/type/instruction IR
│   │   ├── decode/                # 计划；Core 二进制解码
│   │   ├── validate/              # 计划；控制栈、类型栈与 feature 验证
│   │   └── types/                 # 计划；递归类型组规范化、subtyping
│   ├── engine/
│   │   ├── state/                 # 计划；Store/Instance/frame 内部状态
│   │   ├── interp/                # 计划；可挂起解释器、EH、tail calls
│   │   ├── memory/                # 计划；私有/共享内存、atomics、waiters
│   │   ├── heap/                  # 计划；WasmGC、根、externref、句柄
│   │   └── schedule/              # 计划；任务预算、取消、线程生命周期
│   ├── component/
│   │   ├── decode/                # 计划；组件二进制
│   │   ├── validate/              # 计划；组件类型和组合规则
│   │   ├── canon/                 # 计划；lift/lower、post-return
│   │   ├── resource/              # 计划；own/borrow/drop 与句柄
│   │   └── async/                 # 计划；task/future/stream 状态机
│   ├── wasi/
│   │   ├── host/                  # 计划；公共适配器背后的宿主能力实现
│   │   ├── preview1/              # 计划；P1 ABI 实现
│   │   ├── threads/               # 计划；instance-per-thread ABI
│   │   ├── preview2/              # 计划；P2 bindings
│   │   └── preview3/              # 计划；P3 bindings
│   ├── platform/                 # 计划；OS/编译器差异，唯一隔离入口
│   └── spectest/                 # 计划；WAST assertions、fixture loader
├── tests/                         # 计划；跨包/跨语言集成测试
│   ├── spec/                     # Core/提案符合性
│   ├── wasi/                     # P1/P2/P3 与 capability 测试
│   ├── threads/                  # litmus、TLS、wait/notify、exit
│   ├── gc/                       # roots、环、跨调用/异常、配额
│   ├── llgo/                     # 固定 llgo guest 生产器
│   └── differential/             # 对照引擎共同子集
├── testdata/                      # 计划；小型可再生成 fixture
│   ├── manifest.json             # 来源、SHA、license、生成命令、profile
│   ├── core/
│   ├── components/
│   └── llgo/
├── examples/                      # 计划；embed/hostfunc/wasi/thread/gc
└── benchmarks/                    # 计划；冷启动、吞吐、内存、线程、GC
```

不预建空的每个内部目录，也不把没有实现的示例伪装成可运行程序。实施时按里程碑创建包；必要时先用较粗的内部包拆分，等类型边界明确再细分，避免包循环和纯转发层。

## 公共 API 边界

| 包 | 对外责任 | 不应该暴露 |
| --- | --- | --- |
| `pkg/wasm` | ValueType/Value、FuncType、FeatureSet、feature 描述 | 解码器临时状态、指针表示、GC slot 实现 |
| `pkg/vm` | 生命周期、Compile/Instantiate/Call、配置/限额、host imports、错误类别、受控 memory/ref 访问 | 内部指令码、Go runtime ABI、可长期持有的 shared `[]byte` |
| `pkg/component` | 组件加载/链接/调用、组件值、resource/task 的受控生命周期 | Canonical ABI scratch buffer、内部任务队列 |
| `pkg/wasi/preview1` | 注册 P1 host functions、args/env/stdio/preopens/poll 配置 | legacy threads 或自定义 socket 扩展 |
| `pkg/wasi/threads` | 对 P1 context 显式启用 legacy spawn/child-entry 支持 | “已支持 P2/P3 threads”的错误承诺 |
| `pkg/wasi/preview2` | 注册指定版本 WIT worlds、P2 resource/pollable 适配 | P1 内存布局细节 |
| `pkg/wasi/preview3` | 注册指定版本 P3 worlds、异步调用与流的宿主适配 | 用普通 channel 偷换 Canonical ABI 语义 |

`pkg` 是用户要求的公共入口约定，不把所有实现放进 `pkg`。CLI、测试与 SDK 都走这些入口；`internal` 允许重构而不制造兼容承诺。首版 API 标注实验性，待至少一种真实嵌入场景和完整生命周期测试后再稳定。

Host 能力接口放在最接近调用者的公共包；公共 Config 不使用无法由调用方命名的 `internal` 类型。共用 filesystem/clock/random 实现先留在 `internal/wasi/host`，只有实际需要跨版本直接共享公开配置时才考虑新增 `pkg/wasi` 公共包。

## 依赖方向和避免循环

```text
cmd / examples / tests
  ├─ pkg/wasi/preview1 ──────────────────────────┐
  ├─ pkg/wasi/threads → pkg/wasi/preview1       │
  └─ pkg/wasi/preview2,preview3 → pkg/component │
                                     ↓        ↓
                                  pkg/vm → pkg/wasm
                                     ↓
                         internal/engine + internal/core
```

内部图的方向是 `interp → state → memory/heap`，`interp → core/module/types`；decode/validate 只依赖静态 core 类型，不能依赖实例或解释器。schedule 通过窄接口驱动 Execution，不能要求 `state` 反向导入整个解释器。roots 由 `state` 实现 visitor 向 `heap` 提供，避免 `heap → state → heap` 循环。

`pkg/vm` 是组装与封装层。内部执行器不得导入 `pkg/vm`；它接收内部接口/状态。WASI 适配器可以经 `pkg/vm` 注册 host functions，因为 VM 不反向导入 WASI。CLI 负责选择适配器并组装。

底层定义若同时被公开 API 和内部使用，放在 `pkg/wasm` 或专门的内部底层包再由公共层转换；禁止依靠跨层 type alias 将整个内部结构泄漏出去。不同 package 的数量不是设计质量指标，先以依赖可解释、测试可隔离为准。

## 按阶段落地

1. M0：保留公共包 `doc.go` 和兼容性探针；评估 Wago/wazero；固定 corpus 清单和工具。
2. M1：`binary`、`core/{module,decode,validate}`、`engine/{state,interp,memory}`、最小 P1 host；实现 CLI 和 embed 示例。
3. M2：`engine/schedule`、共享 memory waiters、WASI threads；并行完善 `core/types`、`engine/heap`、最终 EH。
4. M3：补齐 Core 3.0 和更广泛测试；内部优化不扩张公开表面。
5. M4/M5：逐步创建 component 的 decoder/validator/canon/resource/async 与 P2/P3 bindings。

各阶段具体退出标准见[路线图](../roadmap.md)。

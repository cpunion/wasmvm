# 本轮验证记录

日期：2026-10-26。当前是设计仓库，测试只覆盖宿主编译器兼容性，不是 Wasm/WASI 符合性认证。

## 本地结果

工作目录：`/Users/lijie/source/goplus/wasmvm`，平台 darwin/arm64。

| 检查 | 结果 |
| --- | --- |
| `./scripts/check.sh go` | 通过；Go 1.27.3，`CGO_ENABLED=0` |
| `./scripts/check.sh llgo` | 通过；Homebrew LLGo v1.0.8 |
| `go test -race -count=1 -timeout=10m ./...` | 通过 |
| `go vet ./...` | 通过 |
| `actionlint .github/workflows/ci.yml` | 通过 |
| `gofmt` | 已格式化，无剩余差异 |

两编译器运行相同的 `internal/compat/compat_test.go` 五项测试：数值/NaN/负零位模式与小端编码、类型化引用环跨宿主 GC、goroutine 与 atomic/mutex 同步、条件变量唤醒、context 取消传播。各公共包只有 `doc.go`，`[no test files]` 不被计为功能验收。

LLGo 解析路径为 `/opt/homebrew/bin/llgo` wrapper，实际二进制为 `/opt/homebrew/Cellar/llgo/1.0.8/libexec/bin/llgo`；`LLGO_ROOT=/opt/homebrew/Cellar/llgo/1.0.8/libexec`，GOROOT 为 Homebrew Go 1.27.3。早期在 llgo checkout 中发现的 devel 二进制/源码组合与本次项目测试不同，详见 [llgo-ci.md](design/llgo-ci.md)。

## 远端 CI

**已实际通过**：[GitHub Actions run 40240658954](https://github.com/cpunion/wasmvm/actions/runs/40240658954)，验证提交 `4968459404a8fc67d6bf6eb27502aeb2a1dd56ea`。

| Job | 结果 | 耗时 |
| --- | --- | --- |
| `go / linux-amd64` | success；共同测试、gofmt/vet、race 全部通过 | 38 秒 |
| `llgo / linux-amd64` | success；固定 revision 安装与核对、共同测试通过 | 2 分 37 秒 |

两者运行同一 `test -count=1 -timeout=10m ./...`；没有 `continue-on-error` 或编译器专属测试跳过。当前 probe 在本地 darwin/arm64 和 CI linux/amd64 上均由 Go/llgo 实际运行。

该运行之后只补充本验证记录与文档链接说明；使用 `[skip ci]` 的文档提交不改动代码、测试或 workflow，也不重新安装工具链。以上结论对应所列验证提交的执行内容。

## 尚未执行

- 未编写或运行 VM 指令、WASI ABI、WasmGC、guest threads 或 Component Model 实现。
- 未跑官方 Wasm/WASI conformance、第三方引擎差分或性能基准。
- 未验证 llgo 生成的 guest，未移植 Wago/wazero。
- `go 1.24.0` 是保守源码语言声明，当前未使用 Go 1.24 工具链实测最低版本。
- 本地成功不代表全部 OS/arch 均兼容，也不代表宿主 race 检查证明了 Wasm memory model。

下一步按 [M0](roadmap.md) 固定候选引擎与测试语料，再作复用决策。

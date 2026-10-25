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

配置两个独立、必须成功的 job：`go / linux-amd64` 与 `llgo / linux-amd64`。两者运行同一 `test -count=1 -timeout=10m ./...`；Go 额外执行 race/vet，LLGo 从固定 revision 安装。创建远端并推送后记录实际运行链接与结论；配置通过 actionlint 不等于云端执行已成功。

## 尚未执行

- 未编写或运行 VM 指令、WASI ABI、WasmGC、guest threads 或 Component Model 实现。
- 未跑官方 Wasm/WASI conformance、第三方引擎差分或性能基准。
- 未验证 llgo 生成的 guest，未移植 Wago/wazero。
- `go 1.24.0` 是保守源码语言声明，当前未使用 Go 1.24 工具链实测最低版本。
- 本地成功不代表全部 OS/arch 均兼容，也不代表宿主 race 检查证明了 Wasm memory model。

下一步按 [M0](roadmap.md) 固定候选引擎与测试语料，再作复用决策。

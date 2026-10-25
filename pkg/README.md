# 公共包

当前只有包职责声明，没有 VM 功能实现。

| 路径 | 规划职责 |
| --- | --- |
| `wasm` | 公开值/类型与特性集合 |
| `vm` | 引擎、Store、模块、实例、host imports 和调用 API |
| `component` | Component Model 的嵌入接口 |
| `wasi/preview1` | WASI Preview 1 |
| `wasi/threads` | 独立可选的 legacy wasi-threads |
| `wasi/preview2` | WASI 0.2 |
| `wasi/preview3` | WASI 0.3 |

完整目录与依赖方向见 [packages.md](../docs/design/packages.md)；当前实现状态见 [README](../README.md)。

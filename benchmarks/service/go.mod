module github.com/cpunion/wasmvm/benchmarks/service

go 1.27.0

require (
	github.com/cpunion/wasmvm v0.0.0
	github.com/tetratelabs/wazero v1.12.0
)

require golang.org/x/sys v0.49.0 // indirect

replace github.com/cpunion/wasmvm => ../..

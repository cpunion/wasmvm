package vm_test

import (
	"fmt"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func ExampleEngine() {
	// (module (func (export "add") (param i32 i32) (result i32)
	//   local.get 0 local.get 1 i32.add))
	binary := []byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 7, 1, 96, 2, 127, 127, 1, 127, 3, 2, 1, 0, 7, 7, 1, 3, 97, 100, 100, 0, 0, 10, 9, 1, 7, 0, 32, 0, 32, 1, 106, 11}
	engine := vm.NewEngine()
	defer engine.Close()
	compiled, err := engine.Compile(nil, binary)
	if err != nil {
		panic(err)
	}
	defer compiled.Close()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		panic(err)
	}
	defer store.Close()
	instance, err := vm.NewLinker(store).Instantiate(nil, compiled, vm.CallOptions{})
	if err != nil {
		panic(err)
	}
	add, err := instance.Export("add")
	if err != nil {
		panic(err)
	}
	result, err := add.Call(nil, []vm.Value{vm.I32(17), vm.I32(25)}, vm.CallOptions{Metered: true, Fuel: 3})
	if err != nil {
		panic(err)
	}
	lo, _, err := result.Values[0].Bits()
	if err != nil {
		panic(err)
	}
	fmt.Println(lo, result.FuelRemaining)
	// Output: 42 0
}

func ExampleStore_NewFunction() {
	engine := vm.NewEngine()
	defer engine.Close()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		panic(err)
	}
	defer store.Close()
	double, err := store.NewFunction(vm.FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(caller *vm.Caller, args []vm.Value) ([]vm.Value, error) {
		lo, _, err := args[0].Bits()
		if err != nil {
			return nil, err
		}
		return []vm.Value{vm.I32(uint32(lo) * 2)}, nil
	})
	if err != nil {
		panic(err)
	}
	linker := vm.NewLinker(store)
	if err = linker.Define("host", "double", double); err != nil {
		panic(err)
	}
	// (module (import "host" "double" (func (param i32) (result i32)))
	//   (func (export "run") (param i32) (result i32) local.get 0 call 0))
	binary := []byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 6, 1, 96, 1, 127, 1, 127, 2, 15, 1, 4, 104, 111, 115, 116, 6, 100, 111, 117, 98, 108, 101, 0, 0, 3, 2, 1, 0, 7, 7, 1, 3, 114, 117, 110, 0, 1, 10, 8, 1, 6, 0, 32, 0, 16, 0, 11}
	compiled, err := engine.Compile(nil, binary)
	if err != nil {
		panic(err)
	}
	defer compiled.Close()
	instance, err := linker.Instantiate(nil, compiled, vm.CallOptions{})
	if err != nil {
		panic(err)
	}
	run, err := instance.Export("run")
	if err != nil {
		panic(err)
	}
	result, err := run.Call(nil, []vm.Value{vm.I32(21)}, vm.CallOptions{})
	if err != nil {
		panic(err)
	}
	lo, _, err := result.Values[0].Bits()
	if err != nil {
		panic(err)
	}
	fmt.Println(lo)
	// Output: 42
}

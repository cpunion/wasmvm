// Command native measures the declared compiler capabilities through one ABI.
// It does not expose or claim a complete native VM backend.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	fmt.Println(`{"ready":true}`)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	check(err)
	if line != "run\n" {
		panic("expected run acknowledgement")
	}
	tier := flag.String("tier", "unmetered", "unmetered or sliced")
	path := flag.String("wasm", "", "module")
	budget := flag.Uint64("budget", 64, "native slice budget")
	seed := flag.Uint64("seed", 305419896, "i32 seed")
	iterations := flag.Uint64("iterations", 250000, "positive i32 loop count")
	expected := flag.Uint64("expected", 956690092, "i32 checksum")
	flag.Parse()
	if *seed > 1<<32-1 || *iterations == 0 || *iterations > 1<<31-1 || *expected > 1<<32-1 {
		panic("invalid i32 arguments")
	}
	input, err := os.ReadFile(*path)
	check(err)
	start := time.Now()
	m, err := decode.Decode(input, decode.DefaultLimits())
	check(err)
	v, err := validate.Check(m, validate.DefaultLimits())
	check(err)
	i, err := lower.Compile(v, lower.DefaultLimits())
	check(err)
	s, err := slots.Compile(i, slots.DefaultLimits())
	check(err)
	front := time.Since(start).Nanoseconds()
	if len(s.Functions) != 1 || m.HasStart || len(m.Imports) != 0 || len(m.Memories) > 1 || len(m.Data) != 0 {
		panic("diagnostic requires one function, optional private memory, and no imports/data/start")
	}
	config := codegen.Config{Unmetered: true}
	if len(m.Memories) != 0 {
		config.MemoryMinimumPages = m.Memories[0].Min
	}
	if *tier == "sliced" {
		config.Unmetered = false
	} else if *tier != "unmetered" {
		panic("unknown tier")
	}
	if *budget == 0 && *tier == "sliced" {
		panic("zero scheduling budget")
	}
	start = time.Now()
	compiled, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, s.Functions[0], codegen.DefaultLimits(), config)
	check(err)
	code, err := native.Publish(compiled.Entry.Code, []uint32{0}, 64<<20)
	check(err)
	back := time.Since(start).Nanoseconds()
	defer code.Close()
	if compiled.Helpers != 0 || compiled.Calls != 0 || compiled.EdgeCopies != 0 {
		panic("unexpected runtime exit")
	}
	lease, err := code.Acquire()
	check(err)
	defer lease.Close()
	start = time.Now()
	frame := make([]byte, len(s.Functions[0].Slots)*16)
	c, err := native.NewContext(uint64(len(frame)), uint64(len(frame)), 1)
	check(err)
	defer c.Close()
	var memory *native.Memory
	if len(m.Memories) != 0 {
		declared := m.Memories[0]
		maximum := uint64(256)
		if declared.HasMax && uint64(declared.Max) < maximum {
			maximum = uint64(declared.Max)
		}
		memory, err = native.NewMemory(uint64(declared.Min)*65536, maximum*65536)
		check(err)
		defer memory.Close()
		check(c.BindMemory(memory))
	}
	setup := time.Since(start).Nanoseconds()
	binary.LittleEndian.PutUint64(frame, *seed)
	binary.LittleEndian.PutUint64(frame[16:], *iterations)
	var calls []int64
	var pollCounts []int
	var first int64
	for n := 0; n < 10; n++ {
		check(c.Write(0, frame))
		check(c.SetState(native.State{Depth: 1, Budget: *budget}))
		polls := 0
		start = time.Now()
		for {
			status, err := c.Enter(lease, 0)
			check(err)
			state, err := c.State()
			check(err)
			if status == native.Poll {
				polls++
				state.Budget = *budget
				check(c.SetState(state))
				continue
			}
			if status == native.GrowMemory {
				instruction := compiled.Function.Code[state.Payload[0]]
				var operand [16]byte
				check(c.Read(uint64(instruction.A)*16, operand[:]))
				size, err := memory.Size()
				check(err)
				result := uint32(size / 65536)
				if err = memory.Grow(size + uint64(binary.LittleEndian.Uint32(operand[:]))*65536); err != nil {
					result = ^uint32(0)
				}
				clear(operand[:])
				binary.LittleEndian.PutUint32(operand[:], result)
				check(c.Write(uint64(instruction.Dst)*16, operand[:]))
				continue
			}
			if status != native.Return {
				panic(fmt.Sprint(status, state))
			}
			in := compiled.Function.Code[state.Payload[0]]
			slot := compiled.Function.Args[in.Extra]
			var result [16]byte
			check(c.Read(uint64(slot)*16, result[:]))
			elapsed := time.Since(start).Nanoseconds()
			if binary.LittleEndian.Uint64(result[:8]) != *expected {
				panic("checksum")
			}
			if n == 0 {
				first = elapsed
			}
			if n >= 3 {
				calls = append(calls, elapsed)
				pollCounts = append(pollCounts, polls)
			}
			break
		}
	}
	check(json.NewEncoder(os.Stdout).Encode(map[string]any{"tier": *tier, "compiler": runtime.Compiler, "runtime": runtime.Version(), "arch": runtime.GOARCH, "wasmSHA256": fmt.Sprintf("%x", sha256.Sum256(input)), "budget": *budget, "iterations": *iterations, "seed": *seed, "codeBytes": len(compiled.Entry.Code), "frameBytes": len(frame), "boundsElided": compiled.BoundsElided, "frontendNs": front, "backendNs": back, "instanceNs": setup, "firstCallNs": first, "warmCallNs": calls, "pollsPerCall": pollCounts, "expected": *expected, "nativeCodeSHA256": fmt.Sprintf("%x", sha256.Sum256(compiled.Entry.Code))}))
}

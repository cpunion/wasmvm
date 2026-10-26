// Command native is a compiler-only architecture diagnostic for integer-mix.
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
	flag.Parse()
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
	if len(s.Functions) != 1 || m.HasStart || len(m.Imports) != 0 {
		panic("diagnostic requires one pure function")
	}
	compile := codegen.CompileUnmeteredFunction
	if *tier == "sliced" {
		compile = codegen.CompileFunction
	} else if *tier != "unmetered" {
		panic("unknown tier")
	}
	if *budget == 0 && *tier == "sliced" {
		panic("zero scheduling budget")
	}
	start = time.Now()
	compiled, err := compile(runtime.GOARCH, s.Functions[0], codegen.DefaultLimits())
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
	setup := time.Since(start).Nanoseconds()
	binary.LittleEndian.PutUint64(frame, 305419896)
	binary.LittleEndian.PutUint64(frame[16:], 250000)
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
			if status != native.Return {
				panic(fmt.Sprint(status, state))
			}
			in := compiled.Function.Code[state.Payload[0]]
			slot := compiled.Function.Args[in.Extra]
			var result [16]byte
			check(c.Read(uint64(slot)*16, result[:]))
			elapsed := time.Since(start).Nanoseconds()
			if binary.LittleEndian.Uint64(result[:8]) != 956690092 {
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
	check(json.NewEncoder(os.Stdout).Encode(map[string]any{"tier": *tier, "compiler": runtime.Compiler, "runtime": runtime.Version(), "arch": runtime.GOARCH, "wasmSHA256": fmt.Sprintf("%x", sha256.Sum256(input)), "budget": *budget, "iterations": 250000, "codeBytes": len(compiled.Entry.Code), "frameBytes": len(frame), "frontendNs": front, "backendNs": back, "instanceNs": setup, "firstCallNs": first, "warmCallNs": calls, "pollsPerCall": pollCounts, "expected": 956690092}))
}

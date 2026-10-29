package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	_ "unsafe"
)

const LLGoFiles = "_wrap/probe.c"

//go:linkname probePthreads C.wasmvm_probe_pthreads
func probePthreads() int32

//go:linkname probeTLS C.wasmvm_probe_tls
func probeTLS(value uint32) int32

//go:linkname registerTSD C.wasmvm_register_tsd
func registerTSD(value uint32) int32

//go:linkname destructorCount C.wasmvm_destructor_count
func destructorCount() uint32

//go:linkname probeSIMD C.wasmvm_probe_simd
func probeSIMD(seed uint32) uint32

type node struct {
	value int
	next  *node
}

//go:noinline
func recoverValue(value int) (result int) {
	defer func() { result = recover().(int) }()
	panic(value)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "exit":
			go func() { os.Exit(23) }()
			select {}
		case "panic":
			go func() { panic("expected-worker-panic") }()
			select {}
		}
	}
	if probePthreads() != 0 || destructorCount() != 3 || probeSIMD(17) != 154 {
		panic("C pthread, TLS destructor or SIMD")
	}
	const workers = 8
	ready := make(chan *node, workers)
	release := make(chan struct{})
	results := make(chan int, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			defer func() { results <- index + 100 }()
			if probeTLS(uint32(index+1)) != 0 || registerTSD(uint32(index+1)) != 0 {
				panic("per-thread TLS or TSD")
			}
			head := &node{value: index + 100}
			head.next = head
			ready <- head
			<-release
			if head.next != head || head.value != index+100 || recoverValue(index) != index {
				panic("worker GC roots or exception state")
			}
			data := bytes.Repeat([]byte{byte(index + 1)}, 32768)
			file, err := os.Create(fmt.Sprintf("/work/worker-%d", index))
			if err != nil {
				panic(err)
			}
			if _, err = file.Write(data); err != nil {
				panic(err)
			}
			if _, err = file.Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
			got, err := io.ReadAll(file)
			if err != nil || !bytes.Equal(got, data) {
				panic("shared process file I/O")
			}
			if err = file.Close(); err != nil {
				panic(err)
			}
			if index%2 == 0 {
				runtime.Goexit()
			}
		}(i)
	}
	heads := make([]*node, workers)
	for i := range heads {
		heads[i] = <-ready
	}
	for i := 0; i < 3; i++ {
		runtime.GC()
		for _, head := range heads {
			if head.next != head || head.value < 100 {
				panic("cross-thread collector handoff")
			}
		}
	}
	close(release)
	wait.Wait()
	sum := 0
	for i := 0; i < workers; i++ {
		sum += <-results
	}
	if sum != 828 {
		panic("worker results or Goexit defers")
	}
	deadline := time.Now().Add(5 * time.Second)
	for destructorCount() != 3+workers {
		if time.Now().After(deadline) {
			panic("pthread TLS destructors after Goexit/return")
		}
		time.Sleep(time.Millisecond)
	}
	if recoverValue(42) != 42 {
		panic("main exception state")
	}
	fmt.Println("llgo-threads-ok", sum)
}

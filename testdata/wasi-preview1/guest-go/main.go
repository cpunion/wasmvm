package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type node struct {
	next *node
	data [128]byte
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil || string(input) != "input" {
		panic("stdin")
	}
	if len(os.Args) != 2 || os.Args[1] != "argument" || os.Getenv("VM_TEST") != "guest" {
		panic("configuration")
	}
	result := make(chan int, 8)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			sum := 0
			for j := 0; j < 100; j++ {
				sum += i + j
				runtime.Gosched()
			}
			result <- sum
		}()
	}
	group.Wait()
	close(result)
	sum := 0
	for n := range result {
		sum += n
	}
	if sum != 42400 {
		panic("goroutine checksum")
	}
	root := &node{}
	root.next = root
	for i := 0; i < 1000; i++ {
		root = &node{next: root}
		root.data[0] = byte(i)
	}
	runtime.GC()
	if root.data[0] != 231 {
		panic("GC root")
	}
	recovered := false
	func() {
		defer func() {
			if recover() == "recover" {
				recovered = true
			}
		}()
		panic("recover")
	}()
	if !recovered {
		panic("recover failed")
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil || random == [32]byte{} {
		panic("random")
	}
	path := "/sandbox/roundtrip"
	data := bytes.Repeat([]byte("file I/O"), 4096)
	if err := os.WriteFile(path, data, 0600); err != nil {
		panic(err)
	}
	actual, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(actual, data) {
		panic("roundtrip")
	}
	if err := os.Remove(path); err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond)
	fmt.Println("go-p1-ok", sum)
}

package main

import (
	"runtime"
	"sync"
)

type node struct {
	value int
	next  *node
}

func recoverValue() (ok bool) {
	defer func() { ok = recover() == 42 }()
	panic(42)
}

func main() {
	a := &node{value: 42}
	a.next = a
	runtime.GC()
	if a.next != a || a.value != 42 {
		panic("GC root")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan int, 1)
	go func() { defer wg.Done(); done <- 42 }()
	wg.Wait()
	if <-done != 42 || !recoverValue() {
		panic("thread/EH")
	}
	println("m0-llgo-guest")
}

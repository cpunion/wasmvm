package main

import (
	"github.com/cpunion/wasmvm/internal/bench/phase"
	"os"
)

func main() { os.Exit(phase.RunCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

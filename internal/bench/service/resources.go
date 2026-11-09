package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type Resources struct {
	RSSBytes, PeakRSSBytes uint64
	OSThreads              int
	Scope                  string
}

// Use a separate resource cohort: ps can perturb later throughput samples.
// Snapshots are taken with clients quiescent. ps reports RSS in KiB on both
// targets; getrusage reports bytes on Darwin and KiB on Linux. PeakRSS includes
// preparation/JIT and is a process-lifetime high-water mark, not owned memory.
func resourceSnapshot() (Resources, error) {
	pid := strconv.Itoa(os.Getpid())
	rss, err := exec.Command("ps", "-o", "rss=", "-p", pid).Output()
	if err != nil {
		return Resources{}, err
	}
	resident, err := strconv.ParseUint(strings.TrimSpace(string(rss)), 10, 64)
	if err != nil {
		return Resources{}, err
	}
	command := []string{"-L", "-p", pid, "-o", "tid="}
	header := 0
	if runtime.GOOS == "darwin" {
		command, header = []string{"-M", "-p", pid}, 1
	}
	threads, err := exec.Command("ps", command...).Output()
	if err != nil {
		return Resources{}, err
	}
	count := len(strings.Split(strings.TrimSpace(string(threads)), "\n")) - header
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return Resources{}, err
	}
	peak := uint64(usage.Maxrss)
	if runtime.GOOS == "linux" {
		peak *= 1024
	}
	if resident == 0 || usage.Maxrss <= 0 || count <= 0 {
		return Resources{}, fmt.Errorf("invalid OS snapshot: rss=%d peak=%d threads=%d", resident, usage.Maxrss, count)
	}
	return Resources{RSSBytes: resident * 1024, PeakRSSBytes: peak, OSThreads: count,
		Scope: "ps instantaneous RSS/OS threads at idle batch boundaries; getrusage process-lifetime peak RSS. Instrumented process only; no throughput acceptance or peak thread claim."}, nil
}

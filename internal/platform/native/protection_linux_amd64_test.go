package native

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"unsafe"
)

func TestLinuxCodeIsRX(t *testing.T) {
	if !Available {
		return
	}
	code, _, _ := makeProbe(t, addProbe())
	maps, err := os.ReadFile("/proc/self/maps")
	if err != nil {
		t.Fatal(err)
	}
	address := uint64(uintptr(unsafe.Pointer(code.address)))
	for _, line := range strings.Split(string(maps), "\n") {
		var start, end uint64
		var flags string
		if _, err := fmt.Sscanf(line, "%x-%x %s", &start, &end, &flags); err != nil {
			continue
		}
		if address >= start && address < end {
			if flags != "r-xp" {
				t.Fatalf("published mapping: %s", line)
			}
			return
		}
	}
	t.Fatal("published code missing from mappings")
}

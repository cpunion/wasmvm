package probe

import (
	"context"
	_ "embed"
	wago "github.com/wago-org/wago"
	"strings"
	"testing"
)

//go:embed fixtures/core.wasm
var core []byte

//go:embed fixtures/host.wasm
var host []byte

func TestCore(t *testing.T) {
	r := wago.NewRuntime()
	defer r.Close()
	m, err := r.Compile(core)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	instance, err := r.Instantiate(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	defer instance.Close()
	got, err := instance.Invoke("add", wago.I32(20), wago.I32(22))
	if err != nil || len(got) != 1 || wago.AsI32(got[0]) != 42 {
		t.Fatalf("add=%v err=%v", got, err)
	}
	_, err = instance.Invoke("trap")
	if err == nil || !strings.Contains(err.Error(), "unreachable") {
		t.Fatalf("trap err=%v", err)
	}
}

func TestHostCallback(t *testing.T) {
	r := wago.NewRuntime()
	defer r.Close()
	m, err := r.Compile(host)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	imports := wago.NewImports()
	calls := 0
	imports.HostFunc("host", "double", func(v int32) int32 { calls++; return v * 2 })
	instance, err := r.Instantiate(context.Background(), m, wago.WithImports(imports))
	if err != nil {
		t.Fatal(err)
	}
	defer instance.Close()
	got, err := instance.Invoke("call_host", wago.I32(21))
	if err != nil || len(got) != 1 || wago.AsI32(got[0]) != 42 || calls != 1 {
		t.Fatalf("host=%v calls=%d err=%v", got, calls, err)
	}
}

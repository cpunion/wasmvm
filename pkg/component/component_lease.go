package component

import (
	"sync"

	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type componentLease struct {
	once  sync.Once
	owner *Compiled
	plan  *validate.Component
	cores map[*module.Module]*vm.CompiledModule
}

func (c *Compiled) releaseLocked() {
	c.refs--
	if c.refs != 0 {
		return
	}
	c.plan = nil
	for _, m := range c.cores {
		m.Close()
	}
	c.cores = nil
}

func (c *Compiled) acquire(e *Engine) (*componentLease, error) {
	if c == nil {
		return nil, vm.ErrReference
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.engine != e || e == nil {
		return nil, vm.ErrReference
	}
	if c.closed || c.plan == nil {
		return nil, wasm.ErrClosed
	}
	c.refs++
	return &componentLease{owner: c, plan: c.plan, cores: c.cores}, nil
}

func (l *componentLease) Close() {
	l.once.Do(func() {
		c := l.owner
		c.mu.Lock()
		defer c.mu.Unlock()
		c.releaseLocked()
		l.owner, l.plan, l.cores = nil, nil, nil
	})
}

//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
)

type subscription struct {
	user     uint64
	kind     byte
	clock    uint32
	end      uint64
	resource *resource
	err      error
}
type pollResult struct {
	index  int
	bytes  uint64
	hangup bool
	err    error
}

func (p *Process) poll(c *abiCall, a []uint64) error {
	n := a[2]
	if n == 0 || n > uint64(p.config.Limits.MaxSubscriptions) {
		return ErrInval
	}
	input, err := c.read(a[0], n*48)
	if err != nil {
		return err
	}
	if err := c.check(a[1], n*32); err != nil {
		return err
	}
	if err := c.check(a[3], 4); err != nil {
		return err
	}
	subs := make([]subscription, int(n))
	releases := make([]func(), 0, int(n))
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	for i := range subs {
		b := input[i*48:]
		s := &subs[i]
		s.user = binary.LittleEndian.Uint64(b)
		s.kind = b[8]
		switch s.kind {
		case 0:
			s.clock = binary.LittleEndian.Uint32(b[16:])
			s.end = binary.LittleEndian.Uint64(b[24:])
			flags := binary.LittleEndian.Uint16(b[40:])
			if flags > 1 {
				return ErrInval
			}
			now, err := p.clock.Now(s.clock)
			if err != nil {
				s.err = err
				continue
			}
			if flags == 0 {
				if s.end > ^uint64(0)-now {
					s.end = ^uint64(0)
				} else {
					s.end += now
				}
			}
		case 1, 2:
			rights := RightPollFdReadwrite
			if s.kind == 1 {
				rights |= RightFdRead
			} else {
				rights |= RightFdWrite
			}
			d, release, err := p.hold(binary.LittleEndian.Uint32(b[16:]), rights)
			if err != nil {
				s.err = err
				continue
			}
			releases = append(releases, release)
			s.resource = d.resource
		default:
			return ErrInval
		}
	}
	ready := make(map[int]pollResult)
	probe := func() error {
		for i, s := range subs {
			if _, ok := ready[i]; ok {
				continue
			}
			result, ok := p.probe(c.ctx, i, s)
			if ok {
				ready[i] = result
			}
		}
		return c.ctx.Err()
	}
	if err := probe(); err != nil {
		return err
	}
	if len(ready) == 0 {
		p.mu.Lock()
		if n > uint64(p.config.Limits.MaxSubscriptions)-p.pollWorkers {
			p.mu.Unlock()
			return ErrNomem
		}
		p.pollWorkers += n
		p.mu.Unlock()
		defer func() { p.mu.Lock(); p.pollWorkers -= n; p.mu.Unlock() }()
		ctx, cancel := context.WithCancel(c.ctx)
		results := make(chan pollResult, len(subs))
		var workers sync.WaitGroup
		for i, s := range subs {
			workers.Add(1)
			go func() {
				defer workers.Done()
				r := pollResult{index: i}
				if s.kind == 0 {
					r.err = p.clock.Wait(ctx, s.clock, s.end)
				} else {
					waitCtx, stop := context.WithCancel(ctx)
					after := context.AfterFunc(s.resource.ctx, stop)
					if s.resource.stream != nil {
						r.bytes, r.hangup, r.err = s.resource.stream.Poll(waitCtx, s.kind == 1)
					} else {
						r.bytes, r.hangup, r.err = host.Poll(waitCtx, host.Descriptor(s.resource.file), s.kind == 1)
					}
					after()
					stop()
				}
				results <- r
			}()
		}
		first := <-results
		cancel()
		workers.Wait()
		close(results)
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
		if errors.Is(first.err, context.Canceled) || errors.Is(first.err, context.DeadlineExceeded) {
			first.err = ErrBadf
		}
		ready[first.index] = first
		for r := range results {
			if !errors.Is(r.err, context.Canceled) && !errors.Is(r.err, context.DeadlineExceeded) {
				ready[r.index] = r
			}
		}
		if err := probe(); err != nil {
			return err
		}
	}
	output, err := c.buffer(uint64(len(ready)) * 32)
	if err != nil {
		return err
	}
	at := 0
	for i, s := range subs {
		r, ok := ready[i]
		if !ok {
			continue
		}
		b := output[at : at+32]
		binary.LittleEndian.PutUint64(b, s.user)
		binary.LittleEndian.PutUint16(b[8:], uint16(errno(r.err)))
		b[10] = s.kind
		binary.LittleEndian.PutUint64(b[16:], r.bytes)
		if r.hangup {
			binary.LittleEndian.PutUint16(b[24:], 1)
		}
		at += 32
	}
	if err := c.write(a[1], output); err != nil {
		return err
	}
	return c.u32(a[3], uint32(len(ready)))
}
func (p *Process) probe(ctx context.Context, index int, s subscription) (pollResult, bool) {
	r := pollResult{index: index, err: s.err}
	if s.err != nil {
		return r, true
	}
	if s.kind == 0 {
		now, err := p.clock.Now(s.clock)
		r.err = err
		return r, err != nil || now >= s.end
	}
	if err := s.resource.ctx.Err(); err != nil {
		r.err = ErrBadf
		return r, true
	}
	var ready bool
	if s.resource.stream != nil {
		r.bytes, r.hangup, ready, r.err = s.resource.stream.Ready(ctx, s.kind == 1)
	} else {
		r.bytes, r.hangup, ready, r.err = host.Ready(host.Descriptor(s.resource.file), s.kind == 1)
	}
	return r, ready || r.err != nil
}

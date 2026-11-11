package validate

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const maxValueComparisons = 64
const maxComparisonNominals = 4096

type comparisonKey struct {
	left, right *model.Type
	budget      uint32
}

type comparisonNominal struct{ left, right uint64 }

// ValueComparison is an immutable resource-aware comparison program. The
// original comparator compiles the reachable ordered nominal obligations and
// structural outcome for one exact node budget. A limit is an outcome after
// those obligations, not a cached error object or a skipped validation budget.
type ValueComparison struct {
	nominal []comparisonNominal
	equal   bool
	limited bool
}

func (p *ValueComparison) HasResources() bool { return len(p.nominal) != 0 }

// Match checks every current nominal identity, including repeats before a
// memoized graph edge. A missing resolver means unresolved identities. Only
// immutable structural work is reused; no runtime identity result is cached.
func (p *ValueComparison) Match(resources func(uint64, uint64) bool) (bool, error) {
	for _, pair := range p.nominal {
		if resources == nil || !resources(pair.left, pair.right) {
			return false, nil
		}
	}
	if p.limited {
		return false, limit("type comparison")
	}
	return p.equal, nil
}

// ValueComparisons owns bounded programs for immutable resolved graphs. Its
// caller serializes use. Budget changes select distinct programs; capacity
// exhaustion returns nil so the caller retains ordinary checked comparison.
type ValueComparisons struct {
	lastKey comparisonKey
	last    *ValueComparison
	entries map[comparisonKey]*ValueComparison
	nominal uint32
}

// Len reports the number of retained programs, including resource-free ones.
func (c *ValueComparisons) Len() int { return len(c.entries) }

func (c *ValueComparisons) Get(a, b *model.Type, budget uint32) *ValueComparison {
	key := comparisonKey{a, b, budget}
	if c.last != nil && c.lastKey == key {
		return c.last
	}
	if plan := c.entries[key]; plan != nil {
		c.lastKey, c.last = key, plan
		return plan
	}
	return c.compile(key)
}

// Keep recording callbacks and mutable build state entirely off the hit path.
func (c *ValueComparisons) compile(key comparisonKey) *ValueComparison {
	if len(c.entries) == maxValueComparisons {
		return nil
	}
	b := comparisonBuilder{remaining: maxComparisonNominals - c.nominal}
	e := equality{remaining: key.budget, resources: b.record}
	equal, err := e.value(key.left, key.right)
	if b.overflow || err != nil && !errors.Is(err, wasm.ErrLimit) {
		return nil
	}
	plan := &ValueComparison{nominal: b.nominal, equal: equal, limited: err != nil}
	if c.entries == nil {
		c.entries = make(map[comparisonKey]*ValueComparison)
	}
	c.entries[key] = plan
	c.nominal += uint32(len(b.nominal))
	c.lastKey, c.last = key, plan
	return plan
}

type comparisonBuilder struct {
	nominal   []comparisonNominal
	remaining uint32
	overflow  bool
}

func (b *comparisonBuilder) record(a, z uint64) bool {
	if b.remaining == 0 {
		b.overflow = true
	} else {
		b.remaining--
		b.nominal = append(b.nominal, comparisonNominal{a, z})
	}
	return true
}

// Clear releases programs and type roots only after the owning actor drains.
func (c *ValueComparisons) Clear() { *c = ValueComparisons{} }

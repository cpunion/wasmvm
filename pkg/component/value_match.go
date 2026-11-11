package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
)

func (in *Instance) matchValue(a, b *model.Type, left, right *namespace) (bool, error) {
	plan := in.comparisons.Get(a, b, in.config.Validate.MaxNodes)
	if plan == nil {
		return in.matchValueUncached(a, b, left, right)
	}
	if !plan.HasResources() {
		return plan.Match(nil)
	}
	return matchNominalPlan(plan, left, right)
}

func matchNominalPlan(plan *validate.ValueComparison, left, right *namespace) (bool, error) {
	return plan.Match(func(a, b uint64) bool {
		l, r := left.resources[a], right.resources[b]
		return l != nil && r != nil && l.identity == r.identity
	})
}

func (in *Instance) matchValueUncached(a, b *model.Type, left, right *namespace) (bool, error) {
	return validate.EqualValue(a, b, in.config.Validate.MaxNodes, func(a, b uint64) bool {
		l, r := left.resources[a], right.resources[b]
		return l != nil && r != nil && l.identity == r.identity
	})
}

package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
)

const maxSchemaMatches = 32
const maxSchemaNominalPairs = 4096

type nominalPair struct{ left, right uint64 }
type schemaMatch struct {
	left, right *model.Type
	budget      uint32
	nominal     []nominalPair
}

// Successful immutable structural comparisons retain their nominal obligations.
// Every use checks current namespace bindings. Errors and mismatches are retried
// through the original comparator. The Store actor owns this bounded cache.
func (in *Instance) matchSchema(a, b *model.Type, left, right *namespace) (bool, error) {
	for i := range in.schemaMatches {
		entry := &in.schemaMatches[i]
		if entry.left == a && entry.right == b && entry.budget == in.config.Validate.MaxNodes {
			for _, pair := range entry.nominal {
				l, r := left.resources[pair.left], right.resources[pair.right]
				if l == nil || r == nil || l.identity != r.identity {
					return false, nil
				}
			}
			return true, nil
		}
	}
	matches := func(a, b uint64) bool {
		l, r := left.resources[a], right.resources[b]
		return l != nil && r != nil && l.identity == r.identity
	}
	if len(in.schemaMatches) == maxSchemaMatches {
		return validate.EqualValue(a, b, in.config.Validate.MaxNodes, matches)
	}
	var nominal []nominalPair
	cacheable := true
	equal, err := validate.EqualValue(a, b, in.config.Validate.MaxNodes, func(a, b uint64) bool {
		if !matches(a, b) {
			return false
		}
		if len(nominal) < maxSchemaNominalPairs-in.schemaNominalPairs {
			nominal = append(nominal, nominalPair{a, b})
		} else {
			cacheable = false
		}
		return true
	})
	if equal && err == nil && cacheable {
		in.schemaMatches = append(in.schemaMatches, schemaMatch{a, b, in.config.Validate.MaxNodes, nominal})
		in.schemaNominalPairs += len(nominal)
	}
	return equal, err
}

package component

import "github.com/cpunion/wasmvm/internal/component/validate"

// lookupName preserves exact bindings and otherwise chooses the greatest
// stable version in the same compatibility track. Lexical name order breaks
// build-metadata ties deterministically, independently of Go map iteration.
// Imported host maps share the instantiation work budget; immutable validated
// export maps are already bounded and do not spend a lifetime lookup quota.
func lookupName[T any](items map[string]T, name string, charge func(uint64) error) (T, bool, error) {
	if value, ok := items[name]; ok {
		return value, true, nil
	}
	return lookupCompatibleName(items, name, charge)
}

// Keep version parsing/scanning out of the small exact-lookup wrapper.
func lookupCompatibleName[T any](items map[string]T, name string, charge func(uint64) error) (T, bool, error) {
	var value T
	key, _ := validate.CompatibleName(name)
	if key == "" {
		return value, false, nil
	}
	best := ""
	var version [3]uint64
	for candidate, item := range items {
		if charge != nil {
			if err := charge(1 + uint64(len(candidate))); err != nil {
				return value, false, err
			}
		}
		track, numbers := validate.CompatibleName(candidate)
		if track != key {
			continue
		}
		order := 0
		for i, n := range numbers {
			if n < version[i] {
				order = -1
				break
			}
			if n > version[i] {
				order = 1
				break
			}
		}
		if best == "" || order > 0 || order == 0 && candidate < best {
			best, version, value = candidate, numbers, item
		}
	}
	return value, best != "", nil
}

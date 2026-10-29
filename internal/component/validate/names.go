package validate

import (
	"strconv"
	"strings"

	"github.com/cpunion/wasmvm/internal/component/model"
)

type nameSet struct {
	keys    map[string]bool
	raw     map[string]Extern
	ordered []Extern
}

type parsedName struct {
	key, annotation, property, resource, version string
	interfacename                                bool
}

func normalize(name string) string { return strings.ReplaceAll(strings.ToLower(name), "-", "") }

func parseName(name string) (parsedName, error) {
	p := parsedName{key: normalize(name)}
	if strings.Contains(name, ":") {
		p.interfacename = true
		base, version, hasVersion := strings.Cut(name, "@")
		p.version = version
		if hasVersion && version == "" {
			return p, invalid("empty interface version")
		}
		namespace, projection, found := strings.Cut(base, ":")
		pkg, iface, slash := strings.Cut(projection, "/")
		if !found || !slash || namespace != strings.ToLower(namespace) || pkg != strings.ToLower(pkg) {
			return p, invalid("interface name syntax")
		}
		for _, part := range []string{namespace, pkg, iface} {
			if err := label(part, map[string]bool{}); err != nil {
				return p, err
			}
		}
		return p, nil
	}
	base := name
	for _, tag := range []string{"constructor", "method", "static", "get", "set"} {
		if strings.HasPrefix(base, "["+tag+"]") {
			p.annotation = tag
			base = strings.TrimPrefix(base, "["+tag+"]")
			break
		}
	}
	if p.annotation == "method" || p.annotation == "static" {
		for _, tag := range []string{"get", "set"} {
			if strings.HasPrefix(base, "["+tag+"]") {
				p.property, base = tag, strings.TrimPrefix(base, "["+tag+"]")
				break
			}
		}
		var found bool
		p.resource, base, found = strings.Cut(base, ".")
		if !found {
			return p, invalid("annotated name requires two labels")
		}
		if err := label(p.resource, map[string]bool{}); err != nil {
			return p, err
		}
	} else if p.annotation == "constructor" {
		p.resource = base
	} else if p.annotation == "get" || p.annotation == "set" {
		p.property = p.annotation
	}
	if err := label(base, map[string]bool{}); err != nil {
		return p, err
	}
	key := normalize(base)
	if p.resource != "" && p.annotation != "constructor" && normalize(p.resource) != key {
		key = normalize(p.resource) + "." + key
	}
	if p.annotation == "constructor" {
		key = "[constructor]" + key
	}
	if p.property == "set" {
		key = "[set]" + key
	}
	p.key = key
	return p, nil
}

func (n *nameSet) add(name model.Name, sort model.Sort, t *Type, pool *TypePool) error {
	p, err := parseName(name.Text)
	if err != nil {
		return err
	}
	if p.annotation != "" && sort != model.FunctionSort {
		return invalid("annotated external name requires function")
	}
	if err := checkAttributes(name, p, sort); err != nil {
		return err
	}
	if n.keys[p.key] {
		return invalid("conflicting external name")
	}
	if p.resource != "" {
		r, ok := n.raw[p.resource]
		if !ok || r.Sort != model.TypeSort || r.Type.Kind != model.Resource {
			return invalid("annotated function resource name not found")
		}
		if p.annotation == "method" {
			if len(t.Parameters) == 0 || t.Parameters[0].Name != "self" || t.Parameters[0].Type.Kind != model.Borrow || t.Parameters[0].Type.ResourceID != r.Type.Resource {
				return invalid("method receiver must borrow named resource")
			}
		}
		if p.annotation == "constructor" {
			result := t.Result
			if result != nil && result.Kind == model.Result {
				result = result.Cases[0].Type
			}
			if result == nil || result.Kind != model.Own || result.ResourceID != r.Type.Resource {
				return invalid("constructor result must own named resource")
			}
		}
	}
	if p.property != "" {
		if t.Kind == model.AsyncFunction {
			return invalid("async property function")
		}
		params := len(t.Parameters)
		if p.annotation == "method" {
			params--
		}
		if p.property == "get" {
			if params != 0 || propertyType(t.Result) == nil {
				return invalid("getter signature")
			}
		} else {
			getter, ok := n.raw[strings.Replace(name.Text, "[set]", "[get]", 1)]
			if !ok || params != 1 || t.Result != nil && (t.Result.Kind != model.Result || t.Result.Cases[0].Type != nil) {
				return invalid("setter signature or preceding getter")
			}
			budget := pool.limits.MaxNodes - pool.nodes
			c := equality{remaining: budget, visited: map[pair]bool{}}
			ok, err := c.value(t.Parameters[len(t.Parameters)-1].Type, propertyType(getter.Type.Result))
			if err != nil {
				return err
			}
			if !ok {
				return invalid("setter property type mismatch")
			}
			if err := pool.charge(uint64(budget - c.remaining)); err != nil {
				return err
			}
		}
	}
	if err := pool.charge(1 + uint64(len(name.Attributes))); err != nil {
		return err
	}
	if n.keys == nil {
		n.keys, n.raw = map[string]bool{}, map[string]Extern{}
	}
	name.Attributes = append([]model.Attribute(nil), name.Attributes...)
	e := Extern{Name: name, Sort: sort, Type: t}
	n.keys[p.key], n.raw[name.Text] = true, e
	n.ordered = append(n.ordered, e)
	return nil
}

func propertyType(t *model.Type) *model.Type {
	if t != nil && t.Kind == model.Result {
		return t.Cases[0].Type
	}
	return t
}

func checkAttributes(name model.Name, parsed parsedName, sort model.Sort) error {
	seen := byte(0)
	iface, suffix := name.Text, ""
	for _, a := range name.Attributes {
		if a.Kind > 2 || seen&(1<<a.Kind) != 0 {
			return invalid("invalid or duplicate external attribute")
		}
		seen |= 1 << a.Kind
		switch a.Kind {
		case 0:
			if sort != model.InstanceSort || parsed.interfacename || parsed.annotation != "" {
				return invalid("implements requires a plain instance name")
			}
			p, err := parseName(a.Value)
			if err != nil || !p.interfacename {
				return invalid("implements must name an interface")
			}
			iface = a.Value
		case 1:
			if sort != model.InstanceSort || a.Value == "" {
				return invalid("version suffix requires an instance and nonempty suffix")
			}
			suffix = a.Value
		}
	}
	if suffix != "" && seen&1 == 0 && !parsed.interfacename {
		return invalid("version suffix requires interface name or implements")
	}
	if !parsed.interfacename && seen&1 == 0 {
		return nil
	}
	_, version, hasVersion := strings.Cut(iface, "@")
	if !hasVersion {
		if suffix != "" {
			return invalid("version suffix requires interface version")
		}
		return nil
	}
	canonical, err := semver(version + suffix)
	if err != nil {
		return err
	}
	if suffix != "" && canonical != version {
		return invalid("noncanonical interface version with suffix")
	}
	return nil
}

// semver returns the canonical compatibility prefix of a valid SemVer 2.0
// version. Parsing and version-suffix checks do not perform numeric arithmetic.
func semver(version string) (string, error) {
	base, build, hasBuild := strings.Cut(version, "+")
	core, pre, hasPre := strings.Cut(base, "-")
	validIdentifiers := func(value string, prerelease bool) bool {
		for _, part := range strings.Split(value, ".") {
			if part == "" {
				return false
			}
			digits := true
			for i := 0; i < len(part); i++ {
				b := part[i]
				if b >= '0' && b <= '9' {
					continue
				}
				digits = false
				if b != '-' && (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') {
					return false
				}
			}
			if prerelease && digits && len(part) > 1 && part[0] == '0' {
				return false
			}
		}
		return true
	}
	if hasBuild && !validIdentifiers(build, false) || hasPre && !validIdentifiers(pre, true) {
		return "", invalid("interface version identifiers")
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return "", invalid("interface version core")
	}
	for _, p := range parts {
		if p == "" || len(p) > 1 && p[0] == '0' {
			return "", invalid("interface version number")
		}
		for i := range p {
			if p[i] < '0' || p[i] > '9' {
				return "", invalid("interface version number")
			}
		}
		if _, err := strconv.ParseUint(p, 10, 64); err != nil {
			return "", invalid("interface version number overflow")
		}
	}
	if hasPre {
		return base, nil
	}
	if parts[0] != "0" {
		return parts[0], nil
	}
	if parts[1] != "0" {
		return "0." + parts[1], nil
	}
	return core, nil
}

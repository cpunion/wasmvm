//go:build linux || darwin

package preview2

import "github.com/cpunion/wasmvm/pkg/component"

func (p *FileSystem) contracts() map[string]component.FunctionCheck {
	u64 := scalarShape(component.KindU64)
	u32 := scalarShape(component.KindU32)
	boolean := scalarShape(component.KindBool)
	text := scalarShape(component.KindString)
	bytes := elementShape(component.KindList, scalarShape(component.KindU8))
	labels := func(kind component.ValueKind, names ...string) *interfaceShape {
		return &interfaceShape{kind: kind, labels: names}
	}
	errCode := labels(component.KindEnum, fileSystemErrorLabels...)
	descFlags := labels(component.KindFlags, "read", "write", "file-integrity-sync", "data-integrity-sync", "requested-write-sync", "mutate-directory")
	pathFlags := labels(component.KindFlags, "symlink-follow")
	openFlags := labels(component.KindFlags, "create", "directory", "exclusive", "truncate")
	descType := labels(component.KindEnum, "unknown", "block-device", "character-device", "directory", "fifo", "symbolic-link", "regular-file", "socket")
	advice := labels(component.KindEnum, "normal", "sequential", "random", "will-need", "dont-need", "no-reuse")
	ownDesc := resourceShape(component.KindOwn, p.descriptorType)
	borrowDesc := resourceShape(component.KindBorrow, p.descriptorType)
	ownEntries := resourceShape(component.KindOwn, p.entriesType)
	borrowEntries := resourceShape(component.KindBorrow, p.entriesType)
	ownInput := resourceShape(component.KindOwn, p.io.inputType)
	ownOutput := resourceShape(component.KindOwn, p.io.outputType)
	borrowError := resourceShape(component.KindBorrow, p.io.errors.resource)
	datetime := aggregateShape(component.KindRecord, shapeField{"seconds", u64}, shapeField{"nanoseconds", u32})
	optionalTime := elementShape(component.KindOption, datetime)
	timestamp := aggregateShape(component.KindVariant, shapeField{"no-change", nil}, shapeField{"now", nil}, shapeField{"timestamp", datetime})
	stat := aggregateShape(component.KindRecord, shapeField{"type", descType}, shapeField{"link-count", u64}, shapeField{"size", u64}, shapeField{"data-access-timestamp", optionalTime}, shapeField{"data-modification-timestamp", optionalTime}, shapeField{"status-change-timestamp", optionalTime})
	hash := aggregateShape(component.KindRecord, shapeField{"lower", u64}, shapeField{"upper", u64})
	entry := aggregateShape(component.KindRecord, shapeField{"type", descType}, shapeField{"name", text})
	result := func(ok *interfaceShape) *interfaceShape {
		return aggregateShape(component.KindResult, shapeField{"", ok}, shapeField{"", errCode})
	}
	out := map[string]component.FunctionCheck{}
	add := func(name string, resultType *interfaceShape, args ...*interfaceShape) {
		parameters := append([]*interfaceShape{borrowDesc}, args...)
		out["[method]descriptor."+name] = shapeContract(parameters, result(resultType))
	}
	add("read-via-stream", ownInput, u64)
	add("write-via-stream", ownOutput, u64)
	add("append-via-stream", ownOutput)
	add("advise", nil, u64, u64, advice)
	add("sync-data", nil)
	add("get-flags", descFlags)
	add("get-type", descType)
	add("set-size", nil, u64)
	add("set-times", nil, timestamp, timestamp)
	add("read", aggregateShape(component.KindTuple, shapeField{"", bytes}, shapeField{"", boolean}), u64, u64)
	add("write", u64, bytes, u64)
	add("read-directory", ownEntries)
	add("sync", nil)
	add("create-directory-at", nil, text)
	add("stat", stat)
	add("stat-at", stat, pathFlags, text)
	add("set-times-at", nil, pathFlags, text, timestamp, timestamp)
	add("link-at", nil, pathFlags, text, borrowDesc, text)
	add("open-at", ownDesc, pathFlags, text, openFlags, descFlags)
	add("readlink-at", text, text)
	add("remove-directory-at", nil, text)
	add("rename-at", nil, text, borrowDesc, text)
	add("symlink-at", nil, text, text)
	add("unlink-file-at", nil, text)
	out["[method]descriptor.is-same-object"] = shapeContract([]*interfaceShape{borrowDesc, borrowDesc}, boolean)
	add("metadata-hash", hash)
	add("metadata-hash-at", hash, pathFlags, text)
	out["[method]directory-entry-stream.read-directory-entry"] = shapeContract([]*interfaceShape{borrowEntries}, result(elementShape(component.KindOption, entry)))
	out["filesystem-error-code"] = shapeContract([]*interfaceShape{borrowError}, elementShape(component.KindOption, errCode))
	out["get-directories"] = shapeContract(nil, elementShape(component.KindList, aggregateShape(component.KindTuple, shapeField{"", ownDesc}, shapeField{"", text})))
	return out
}

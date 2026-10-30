//go:build linux || darwin

package preview2

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

var fsDescriptorMethods = []string{"read-via-stream", "write-via-stream", "append-via-stream", "advise", "sync-data", "get-flags", "get-type", "set-size", "set-times", "read", "write", "read-directory", "sync", "create-directory-at", "stat", "stat-at", "set-times-at", "link-at", "open-at", "readlink-at", "remove-directory-at", "rename-at", "symlink-at", "unlink-file-at", "is-same-object", "metadata-hash", "metadata-hash-at"}

// Imports shares the IO owner's nominal stream/error domain. The filesystem
// owner grants only configured directories; it never inherits the host cwd.
func (p *FileSystem) Imports() component.Imports {
	if p == nil || p.descriptorType == nil {
		return component.Imports{}
	}
	out := p.io.Imports()
	contracts := p.contracts()
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	for _, name := range fsDescriptorMethods {
		key := "[method]descriptor." + name
		functions[key] = p.method(name)
		checks[key] = contracts[key]
	}
	const entries = "[method]directory-entry-stream.read-directory-entry"
	functions[entries] = p.readEntry
	functions["filesystem-error-code"] = p.errorCode
	checks[entries] = contracts[entries]
	checks["filesystem-error-code"] = contracts["filesystem-error-code"]
	out.Instances[FileSystemTypesInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"descriptor": p.descriptorType, "directory-entry-stream": p.entriesType, "input-stream": p.io.inputType, "output-stream": p.io.outputType, "error": p.io.errors.resource}, Functions: functions, FunctionChecks: checks}
	out.Instances[FileSystemPreopensInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"descriptor": p.descriptorType}, Functions: map[string]component.HostFunction{"get-directories": p.getDirectories}, FunctionChecks: map[string]component.FunctionCheck{"get-directories": contracts["get-directories"]}}
	return out
}
func (p *FileSystem) method(name string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(); err != nil {
			if name == "is-same-object" {
				return nil, err
			}
			return fsFailure(err), nil
		}
		defer p.active.Done()
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		d, err := p.descriptor(c, args[0].Resource)
		if err != nil {
			return nil, err
		}
		defer d.backing.release()
		value, hasValue, err := p.execute(c, d, name, args)
		if failed := p.check(c.Context()); failed != nil {
			return nil, failed
		}
		if name == "is-same-object" {
			return &value, err
		}
		if err != nil {
			return fsFailure(err), nil
		}
		if hasValue {
			return fsOK(&value), nil
		}
		return fsOK(nil), nil
	}
}
func (p *FileSystem) execute(c *component.Caller, d fsDescriptor, name string, args []component.Value) (value component.Value, hasValue bool, err error) {
	hasValue = true
	switch name {
	case "read-via-stream", "write-via-stream", "append-via-stream":
		offset := uint64(0)
		if name != "append-via-stream" {
			offset = args[1].Bits
		}
		value, err = p.stream(c, d, name, offset)
	case "get-flags":
		value.Bits = d.flags
	case "get-type", "stat", "metadata-hash":
		var st host.Stat
		st, err = host.Fstat(d.backing.file)
		if err == nil {
			switch name {
			case "get-type":
				value.Bits = uint64(fsType(st.Mode))
			case "stat":
				value, err = fsStatValue(st)
			default:
				value = fsHash(st)
			}
		}
	case "read":
		value, err = p.readAt(c.Context(), d, args[1].Bits, args[2].Bits)
	case "write":
		value, err = p.writeAt(c.Context(), d, args[1], args[2].Bits)
	case "read-directory":
		value, err = p.readDirectory(c, d)
	case "open-at":
		value, err = p.open(c, d, args[1].Bits, args[2].Text, args[3].Bits, args[4].Bits)
	case "stat-at", "metadata-hash-at":
		var st host.Stat
		st, err = p.statAt(d, args[1].Bits, args[2].Text)
		if err == nil {
			if name == "stat-at" {
				value, err = fsStatValue(st)
			} else {
				value = fsHash(st)
			}
		}
	case "is-same-object":
		var other fsDescriptor
		other, err = p.descriptor(c, args[1].Resource)
		if err != nil {
			break
		}
		defer other.backing.release()
		var a, b host.Stat
		a, err = host.Fstat(d.backing.file)
		if err == nil {
			b, err = host.Fstat(other.backing.file)
		}
		if err == nil && a.Dev == b.Dev && a.Inode == b.Inode {
			value.Bits = 1
		}
	case "readlink-at":
		var dir *host.Directory
		dir, err = fsDirectory(d, false)
		if err == nil {
			err = p.path(args[1].Text)
		}
		if err == nil {
			value.Text, err = dir.Readlink(args[1].Text)
		}
		if err == nil && strings.HasPrefix(value.Text, "/") {
			err = fsCode(FSNotPermitted)
		}
		if err == nil && !utf8.ValidString(value.Text) {
			err = fsCode(FSIllegalByteSequence)
		}
		if err == nil && uint64(len(value.Text)) > uint64(p.config.MaxPathBytes) {
			err = fsCode(FSNameTooLong)
		}
	default:
		hasValue = false
		switch name {
		case "advise":
			// Advice is a nonbinding hint. Validate representability even when the OS
			// does not offer an equivalent advisory call.
			if args[1].Bits > 1<<63-1 || args[2].Bits > uint64(1<<63-1)-args[1].Bits {
				err = fsCode(FSOverflow)
			}
		case "sync", "sync-data":
			if d.flags&(fsWrite|fsMutate) != 0 {
				d.backing.ioMu.Lock()
				err = d.backing.file.Sync()
				d.backing.ioMu.Unlock()
			}
		case "set-size":
			err = fsWritable(d)
			if err == nil && d.backing.directory != nil {
				err = fsCode(FSIsDirectory)
			}
			if err == nil && args[1].Bits > 1<<63-1 {
				err = fsCode(FSOverflow)
			}
			if err == nil {
				d.backing.ioMu.Lock()
				err = d.backing.file.Truncate(int64(args[1].Bits))
				d.backing.ioMu.Unlock()
			}
		case "set-times":
			err = fsWritable(d)
			if err == nil {
				var at, mt uint64
				var flags uint32
				at, mt, flags, err = fsTimestamps(args[1], args[2])
				if err == nil {
					err = host.SetTimes(d.backing.file, at, mt, flags)
				}
			}
		case "create-directory-at", "remove-directory-at", "unlink-file-at":
			err = p.mutatePath(d, name, args[1].Text)
		case "rename-at", "link-at":
			err = p.pair(c, d, name, args)
		case "set-times-at":
			err = p.setTimesAt(d, args)
		case "symlink-at":
			var dir *host.Directory
			dir, err = fsDirectory(d, true)
			if err == nil {
				err = p.path(args[1].Text)
			}
			if err == nil {
				err = p.path(args[2].Text)
			}
			if err == nil {
				err = dir.Symlink(args[1].Text, args[2].Text)
			}
		default:
			err = vm.ErrType
		}
	}
	return
}
func (p *FileSystem) getDirectories(c *component.Caller, _ []component.Value) (*component.Value, error) {
	if err := p.begin(); err != nil {
		return nil, err
	}
	defer p.active.Done()
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	p.mu.Lock()
	preopens := append([]fsPreopen(nil), p.preopens...)
	p.mu.Unlock()
	out := &component.Value{Elements: make([]component.Value, 0, len(preopens))}
	for _, entry := range preopens {
		value, err := p.publish(c, entry.descriptor)
		if err != nil {
			// Publication may fail after earlier handles were granted. Explicitly drop
			// those handles before returning a host error; no partial list escapes.
			for _, previous := range out.Elements {
				c.DropResource(previous.Elements[0].Resource)
			}
			return nil, err
		}
		out.Elements = append(out.Elements, component.Value{Elements: []component.Value{value, {Text: entry.name}}})
	}
	return out, nil
}
func (p *FileSystem) readEntry(c *component.Caller, args []component.Value) (*component.Value, error) {
	if err := p.begin(); err != nil {
		return fsFailure(err), nil
	}
	defer p.active.Done()
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	id, err := c.ResourceRep(p.entriesType, args[0].Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	entry := p.entries[id]
	p.mu.Unlock()
	if entry == nil {
		return nil, vm.ErrReference
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	value := component.Value{}
	if entry.at < len(entry.entries) {
		next := entry.entries[entry.at]
		entry.at++
		value.Case = 1
		value.Elements = []component.Value{{Elements: []component.Value{{Bits: uint64(fsType(next.Mode))}, {Text: next.Name}}}}
	}
	return fsOK(&value), nil
}
func (p *FileSystem) errorCode(c *component.Caller, args []component.Value) (*component.Value, error) {
	info, err := p.io.errors.Describe(c, args[0].Resource)
	if err != nil {
		return nil, err
	}
	value := &component.Value{}
	var cause *FileSystemError
	if errors.As(info.Cause, &cause) {
		value.Case = 1
		value.Elements = []component.Value{{Bits: uint64(cause.Code)}}
	}
	return value, nil
}

#!/usr/bin/env python3
"""Author value/start sections from the pinned Component Model binary grammar.

wasm-tools 1.258.0 does not implement these text constructs. Core/component
fragments below still use that pinned producer. These are authored tests, not
an independent conformance oracle.
"""
from pathlib import Path
import hashlib
import json
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'pkg/component/testdata'
PIN = 'd095e2f60ddae1f9a1e9b1e2fb547ec1114da677'


def u(n):
    out = bytearray()
    while n >= 128:
        out.append((n & 127) | 128)
        n >>= 7
    return bytes(out + bytes([n]))


def name(s):
    b = s.encode()
    return u(len(b)) + b


def vec(xs):
    return u(len(xs)) + b''.join(xs)


def section(tag, data):
    return bytes([tag]) + u(len(data)) + data


def definitions(tag, *xs):
    return section(tag, vec(xs))


def imp(n, typ):
    return b'\0' + name(n) + typ


def exp(n, sort, index):
    return b'\0' + name(n) + bytes([sort]) + u(index) + b'\0'


def val_import(n, typ):
    return definitions(10, imp(n, b'\x02\x01' + typ))


def start(f, args, result=True):
    return section(9, u(f) + vec([u(a) for a in args]) + u(int(result)))


def component(*parts):
    return b'\0asm\x0d\0\x01\0' + b''.join(parts)


def parse(text):
    with tempfile.TemporaryDirectory() as td:
        src, out = Path(td)/'input.wat', Path(td)/'output.wasm'
        src.write_text(text)
        subprocess.run(['wasm-tools', 'parse', str(src), '-o', str(out)], check=True)
        return out.read_bytes()


def main():
    assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
    resource = definitions(10, imp('r', b'\x03\x01'))
    own = definitions(7, b'\x69\0')
    borrow = definitions(7, b'\x68\0')
    export_r = definitions(11, exp('r', 3, 0))
    echo = definitions(7, b'\x40' + vec([name('value') + b'\x01']) + b'\0\x01')
    echo_import = definitions(10, imp('pass', b'\x01\x02'))
    generated = {}

    def emit(n, data, description):
        path = OUT/(n + '.wasm')
        path.write_bytes(data)
        generated[path.name] = {'sha256': hashlib.sha256(data).hexdigest(), 'description': description}

    emit('resource-value-import', component(resource, own, val_import('input', b'\x01'), export_r, definitions(11, exp('output', 2, 0))), 'Import own<r>; export the same affine occurrence.')
    emit('resource-value-start', component(resource, own, echo, echo_import, val_import('input', b'\x01'), start(0, [0]), export_r, definitions(11, exp('output', 2, 1))), 'Import own<r>; start host pass(own<r>); export its result.')
    emit('resource-value-host-start-failure', component(resource, own, echo, echo_import, val_import('input', b'\x01'), start(0, [0]), export_r, definitions(11, exp('output', 2, 1)), definitions(7, b'\x40\0\x01\0'), definitions(10, imp('fail', b'\x01\x04')), start(1, [], False)), 'A later start failure reclaims the moved host resource result.')
    emit('resource-value-pair', component(resource, own, definitions(7, b'\x6f'+vec([b'\x01', b'\x01'])), val_import('input', b'\x02'), export_r, definitions(11, exp('output', 2, 0))), 'Tuple own resources uses transactional duplicate detection.')
    emit('resource-value-empty', component(resource, own, definitions(7, b'\x70\x01'), val_import('input', b'\x02'), export_r, definitions(11, exp('output', 2, 0))), 'Even an empty list<own<r>> is an affine standalone value.')
    get = definitions(7, b'\x40' + vec([name('value') + b'\x01']) + b'\0\x79')
    emit('resource-value-borrow', component(resource, borrow, get, definitions(10, imp('get', b'\x01\x02')), val_import('input', b'\x01'), start(0, [0]), definitions(11, exp('answer', 2, 1))), 'Imported borrow lives through start and expires after initialization.')
    emit('resource-value-borrow-equality', component(resource, borrow, get, definitions(10, imp('get', b'\x01\x02')), val_import('first', b'\x01'), definitions(10, imp('same', b'\x02\0\0')), start(0, [0]), start(0, [1]), definitions(11, exp('first', 2, 2), exp('second', 2, 3))), 'Equality-bound borrows compare live nominal capabilities, not token addresses.')
    emit('resource-value-borrow-export', component(resource, borrow, val_import('input', b'\x01'), export_r, definitions(11, exp('output', 2, 0))), 'Invalid: a component cannot export a borrow value.')

    instance_type = b'\x42' + vec([b'\x04' + imp('r', b'\x03\x01'), b'\x01\x69\0', b'\x04' + imp('input', b'\x02\x01\x01')])
    emit('resource-value-nested', component(definitions(7, instance_type), definitions(10, imp('box', b'\x05\0')), definitions(11, exp('box', 5, 0))), 'A nested host interface exports a resource type and standalone own value.')

    # Child declarations/imports freshen resource binders. Aliases/projected views
    # must keep each runtime value occurrence and its original nominal scope.
    child = component(resource, own, val_import('input', b'\x01'), export_r, definitions(11, exp('output', 2, 0)))
    args = vec([name('r') + b'\x03\0', name('input') + b'\x02\0'])
    child_instance = definitions(5, b'\0\0' + args)
    alias = definitions(6, b'\x02\0\0' + name('output'))
    emit('resource-value-child', component(resource, own, val_import('input', b'\x01'), section(4, child), child_instance, alias, export_r, definitions(11, exp('output', 2, 1))), 'Move own<r> through child import/export and an alias.')

    # A real own backend creates the resource and invokes its guest destructor.
    guest = parse('''(component
      (core module $m (global $n (mut i32) (i32.const 0))
        (func (export "dtor") (param i32) global.get $n local.get 0 i32.add global.set $n)
        (func (export "total") (result i32) global.get $n))
      (core instance $m (instantiate $m))
      (alias core export $m "dtor" (core func $dtor))
      (alias core export $m "total" (core func $total))
      (type $r (resource (rep i32) (dtor (func $dtor))))
      (export "r" (type $r))
      (core func $new (canon resource.new $r))
      (core module $make (import "env" "new" (func $new (param i32) (result i32)))
        (func (export "make") (result i32) i32.const 42 call $new))
      (core instance $env (export "new" (func $new)))
      (core instance $make (instantiate $make (with "env" (instance $env))))
      (alias core export $make "make" (core func $make))
      (func (export "make") (result (own 1)) (canon lift (core func $make)))
      (func (export "total") (result u32) (canon lift (core func $total))))''')
    emit('resource-value-guest-start', guest + start(0, []) + definitions(11, exp('output', 2, 0)), 'Start native/interpreted guest make; export own result with its live destructor.')
    observed = parse('''(component
      (import "observe" (func $observe (param "rep" u32)))
      (core func $observe (canon lower (func $observe)))
      (type $r (resource (rep i32) (dtor (func $observe))))
      (export "r" (type $r))
      (core func $new (canon resource.new $r))
      (core module $make (import "env" "new" (func $new (param i32) (result i32)))
        (func (export "make") (result i32) i32.const 42 call $new))
      (core instance $env (export "new" (func $new)))
      (core instance $make (instantiate $make (with "env" (instance $env))))
      (alias core export $make "make" (core func $make))
      (func (export "make") (result (own 2)) (canon lift (core func $make))))''')
    printed = subprocess.check_output(['wasm-tools', 'print', '-'], input=observed).decode()
    assert '(type (;4;) (func (result 3)))' in printed, printed
    assert '(func (;1;) (type 4)' in printed, printed
    fail = definitions(7, b'\x40\0\x01\0') + definitions(10, imp('fail', b'\x01\x05'))
    emit('resource-value-start-failure', observed + start(1, []) + definitions(11, exp('output', 2, 0)) + fail + start(3, [], False), 'A later start failure runs the earlier unpublished resource guest destructor.')
    manifest = {'schemaVersion': 1, 'primarySource': f'https://github.com/WebAssembly/component-model/blob/{PIN}/design/mvp/Binary.md', 'producer': 'project-authored value/start sections; wasm-tools 1.258.0 Core fragments', 'independentOracle': False, 'generatorSHA256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'fixtures': generated}
    (OUT/'resource-values.json').write_text(json.dumps(manifest, indent=2)+'\n')


if __name__ == '__main__':
    main()

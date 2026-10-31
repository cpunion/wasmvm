#!/usr/bin/env python3
"""Build pinned host-component inputs and one authored value/start composition."""
from pathlib import Path
import hashlib
import importlib.util
import json
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'pkg/component/testdata'


def sections(data):
    at = 8
    while at < len(data):
        tag = data[at]
        at += 1
        size, shift = 0, 0
        while True:
            b = data[at]
            at += 1
            size |= (b & 127) << shift
            if not b & 128:
                break
            shift += 7
        yield tag, data[at:at+size]
        at += size


def main():
    assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
    generated = {}
    for src in sorted(OUT.glob('host-component*.wat')):
        dest = src.with_suffix('.wasm')
        subprocess.run(['wasm-tools', 'parse', str(src), '-o', str(dest)], check=True)
        subprocess.run(['wasm-tools', 'validate', str(dest)], check=True)
        generated[dest.name] = {'source': src.name, 'producerValidated': True}
    spec = importlib.util.spec_from_file_location('values', ROOT/'scripts/build_resource_value_fixtures.py')
    helper = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(helper)
    base = (OUT/'host-components-values-base.wasm').read_bytes()
    printed = subprocess.check_output(['wasm-tools', 'print', '-'], input=base)
    assert b'(func $make (;2;))' in printed
    assert b'(export (;4;) "r"' in printed
    members = [b'\0' + helper.name('r') + bytes([3, 3]),
               b'\0' + helper.name('output') + bytes([2, 0])]
    values_instance = helper.definitions(5, b'\x01' + helper.vec(members))
    out = base + helper.start(2, []) + values_instance + helper.definitions(11, helper.exp('values', 5, 1))
    (OUT/'host-components-values.wasm').write_bytes(out)
    generated['host-components-values.wasm'] = {'source': 'host-components-values-base.wat', 'textGenerated': False, 'producerValidated': True, 'validationFeatures': ['cm-values'], 'authored': 'Append start(make)->value and affine output export from pinned Binary.md.'}
    suffix = helper.parse('''(component
      (import "fail" (func $fail))
      (core func $fail (canon lower (func $fail)))
      (core module $m (import "env" "fail" (func $fail))
        (func $start call $fail) (start $start))
      (core instance $env (export "fail" (func $fail)))
      (core instance (instantiate $m (with "env" (instance $env)))))''')
    for tag, payload in sections(suffix):
        if tag in (0, 7, 10):
            continue
        if tag == 8:
            assert payload == bytes([1, 1, 0, 0, 0])
            payload = bytes([1, 1, 0, 1, 0])
        out += helper.section(tag, payload)
    (OUT/'host-components-start-failure.wasm').write_bytes(out)
    generated['host-components-start-failure.wasm'] = {'source': 'host-components-values-base.wat', 'textGenerated': False, 'producerValidated': True, 'validationFeatures': ['cm-values'], 'authored': 'Append the value start, then a real Core start whose host call fails.'}
    for name in ['host-components-values.wasm', 'host-components-start-failure.wasm']:
        subprocess.run(['wasm-tools', 'validate', '--features', 'cm-values', str(OUT/name)], check=True)
    hashes = {}
    for name, entry in generated.items():
        hashes[name] = hashlib.sha256((OUT/name).read_bytes()).hexdigest()
        source = entry['source']
        hashes[source] = hashlib.sha256((OUT/source).read_bytes()).hexdigest()
    producer = Path(__file__)
    helper_path = ROOT/'scripts/build_resource_value_fixtures.py'
    manifest = {'schemaVersion': 1, 'producer': 'wasm-tools 1.258.0 parse and validate; explicitly marked authored value/start extensions', 'sourceRevision': helper.PIN, 'generatorSHA256': hashlib.sha256(producer.read_bytes()).hexdigest(), 'helperSHA256': hashlib.sha256(helper_path.read_bytes()).hexdigest(), 'fixtures': generated, 'sha256': hashes}
    (OUT/'host-components.json').write_text(json.dumps(manifest, indent=2)+'\n')


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""Reproduce unmodified Wasmtime 48.0.1 Rust standard-library P2 commands."""
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / 'testdata/wasi-preview2/upstream-wasmtime'
OUT = ROOT / 'pkg/wasi/preview2/testdata'
MANIFEST = OUT / 'upstream-manifest.json'
WORK = ROOT / 'artifacts/components/upstream-producer'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(*args):
    subprocess.run([str(a) for a in args], check=True)


def main():
    pins = json.loads(MANIFEST.read_text())
    for name, expected in pins['sourcesSha256'].items():
        assert digest(SOURCE / name) == expected, name
    assert subprocess.check_output(['rustc', '+1.98.0', '--version'], text=True).startswith('rustc 1.98.0 ')
    assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
    WORK.mkdir(parents=True, exist_ok=True)
    binaries = {}
    for name in pins['tests']:
        raw, out = WORK / (name + '.wasm'), OUT / ('upstream-' + name + '.wasm')
        run('rustc', '+1.98.0', SOURCE / (name + '.rs'), '--edition', '2024',
            '--target', 'wasm32-wasip2', '-O', '-C', 'panic=abort',
            '--remap-path-prefix=' + str(SOURCE) + '=/wasmtime/crates/test-programs/src/bin', '-o', raw)
        run('wasm-tools', 'strip', '--all', raw, '-o', out)
        run('wasm-tools', 'validate', out)
        binaries[out.name] = digest(out)
    pins['flags'] = ['--edition=2024', '--target=wasm32-wasip2', '-O', '-Cpanic=abort', '--remap-path-prefix=SOURCE=/wasmtime/crates/test-programs/src/bin']
    pins['binarySha256'] = binaries
    pins['generatorSha256'] = digest(Path(__file__))
    MANIFEST.write_text(json.dumps(pins, indent=2) + '\n')
    print('Reproduced and validated', len(binaries), 'unmodified upstream commands')


if __name__ == '__main__':
    main()

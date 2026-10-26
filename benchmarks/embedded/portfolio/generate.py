#!/usr/bin/env python3
"""Build the authored Core 2.0 portfolio with exact producer versions."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import kernels

ROOT = Path(__file__).resolve().parent
EMBEDDED = ROOT.parent
FIXTURES = EMBEDDED / 'fixtures'


def run(command):
    return subprocess.check_output(command, text=True).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    if run(['wasm-tools', '--version']) != 'wasm-tools 1.258.0':
        raise ValueError('requires wasm-tools 1.258.0')
    clang = os.environ.get('PORTFOLIO_CLANG', 'clang')
    rustc = ['rustup', 'run', '1.98.0', 'rustc']
    clang_version = run([clang, '--version'])
    rust_version = run(rustc + ['--version'])
    if 'clang version 22.1.8' not in clang_version or rust_version != 'rustc 1.98.0 (88d9e12ae 2026-08-18)':
        raise ValueError('requires Clang 22.1.8 and Rust 1.98.0')
    entries = []
    for kind, count, family in kernels.CASES:
        wat, wasm = FIXTURES / (kind + '.wat'), FIXTURES / (kind + '.wasm')
        source = None
        if kind == 'image-filter-c':
            source = ROOT / 'image-filter.c'
            flags = ['--target=wasm32-unknown-unknown', '-O2', '-nostdlib', '-Wl,--no-entry', '-Wl,--export=run', '-Wl,--initial-memory=131072', '-Wl,--max-memory=131072', '-Wl,-z,stack-size=65536']
            command = [clang, *flags, str(source), '-o', str(wasm)]
            producer = {'version': clang_version, 'flags': flags, 'compilerSha256': digest(Path(shutil.which(clang)))}
        elif kind == 'packet-crc-rust':
            source = ROOT / 'packet-crc.rs'
            flags = ['--edition=2024', '--crate-type=cdylib', '--target=wasm32-unknown-unknown', '-Copt-level=2', '-Cpanic=abort', '-Ctarget-cpu=mvp', '-Clink-arg=--initial-memory=131072', '-Clink-arg=--max-memory=131072', '-Clink-arg=-zstack-size=65536']
            command = rustc + flags + [str(source), '-o', str(wasm)]
            producer = {'version': rust_version, 'flags': flags}
        else:
            wat.write_text(kernels.module(kind))
            command = ['wasm-tools', 'parse', str(wat), '-o', str(wasm)]
            producer = {'version': 'wasm-tools 1.258.0', 'flags': ['parse']}
        subprocess.run(command, check=True)
        wasm.chmod(0o644)
        subprocess.run(['wasm-tools', 'validate', str(wasm), '--features=wasm2'], check=True)
        if source:
            subprocess.run(['wasm-tools', 'print', str(wasm), '-o', str(wat)], check=True)
        pairs = ((0, 1), (1, 2), (kernels.MASK, 3), (0xdeadbeef, 17))
        entries.append({'id': kind, 'kind': kind, 'family': family, 'seed': 0x12345678,
                        'iterations': count, 'expectedChecksum': kernels.checksum(kind, 0x12345678, count),
                        'unit': 'ns/iteration', 'producer': producer,
                        'wat': 'fixtures/' + wat.name, 'wasm': 'fixtures/' + wasm.name,
                        'watSha256': digest(wat), 'wasmSha256': digest(wasm),
                        'source': 'portfolio/' + source.name if source else 'portfolio/kernels.py',
                        'sourceSha256': digest(source or ROOT / 'kernels.py'),
                        'validationCases': [{'seed': s, 'iterations': n, 'expectedChecksum': kernels.checksum(kind, s, n)} for s, n in pairs]})
    manifest = {'schemaVersion': 1, 'portfolioVersion': 'core2-v1', 'origin': 'authored in wasmvm; no third-party benchmark source',
                'profile': 'Core 2.0; no imports, WASI or start function; run(i32,i32)->i32',
                'generationCommand': 'python3 benchmarks/embedded/portfolio/generate.py', 'workloads': entries}
    (FIXTURES / 'portfolio.json').write_text(json.dumps(manifest, indent=2) + '\n')
    print('Generated', len(entries), 'Core 2.0 portfolio workloads')


if __name__ == '__main__':
    main()

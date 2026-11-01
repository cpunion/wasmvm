#!/usr/bin/env python3
"""Build pinned native-async C/Rust CLI and HTTP world fixtures."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[1]
SHA = '2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311'
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--work', type=Path, default=ROOT/'artifacts/components/preview3-world-producer')
p.add_argument('--archive', type=Path, default=ROOT/'artifacts/components/wasi-inputs/preview3/wasi-wit-0.3.2.tar.gz')
p.add_argument('--output', type=Path, default=ROOT/'pkg/wasi/preview3/testdata')
p.add_argument('--bindgen', type=Path, default=ROOT/'artifacts/components/wit-bindgen-0.54.0/bin/wit-bindgen')
p.add_argument('--sysroot', type=Path, default=Path('/opt/homebrew/opt/wasi-libc/share/wasi-sysroot'))
a = p.parse_args()
work, out = a.work.resolve(), a.output.resolve()
source = ROOT/'testdata/wasi-preview3/guests'
assert hashlib.sha256(a.archive.read_bytes()).hexdigest() == SHA
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
assert '0.54.0' in subprocess.check_output([str(a.bindgen), '--version'], text=True)
assert '22.1.8' in subprocess.check_output(['clang', '--version'], text=True)
assert 'rustc 1.98.0 ' in subprocess.check_output(['rustc', '+1.98.0', '--version'], text=True)
work.mkdir(parents=True, exist_ok=True)
out.mkdir(parents=True, exist_ok=True)
shutil.copytree(source, work/'guests', dirs_exist_ok=True, ignore=shutil.ignore_patterns('deps', 'target'))
wit = work/'guests/wit'
(wit/'deps').mkdir(exist_ok=True)
with tarfile.open(a.archive) as archive:
    # Use complete root packages; per-world deps can be reduced package views.
    for member in archive.getmembers():
        parts = Path(member.name).parts
        if len(parts) == 3 and parts[0] == 'wasi-wit-0.3.2' and member.name.endswith('.wit'):
            (wit/'deps'/parts[-1]).write_bytes(archive.extractfile(member).read())

def run(*command, **kwargs):
    subprocess.run([str(x) for x in command], check=True, **kwargs)

run(a.bindgen, 'c', wit, '--world', 'c-command', '--async=-all', '--out-dir', work/'c')
run('clang', '--target=wasm32-wasip1', f'--sysroot={a.sysroot}', '-O2', '-mexec-model=reactor', '-nodefaultlibs', '-I', work/'c', work/'guests/c/command.c', work/'c/c_command.c', work/'c/c_command_component_type.o', '-lc', '-Wl,--export-memory', '-o', work/'c.core.wasm')
run('wasm-tools', 'component', 'new', work/'c.core.wasm', '-o', out/'world-c-command.wasm')
env = os.environ.copy()
env['RUSTFLAGS'] = f'--remap-path-prefix={work}=/wasmvm-p3-guests'
for package, binary in [('rust', 'command'), ('rust-http', 'http'), ('rust-middleware', 'middleware')]:
    run('cargo', '+1.98.0', 'build', '--locked', '--offline', '--release', '--target', 'wasm32-unknown-unknown', '--manifest-path', work/f'guests/{package}/Cargo.toml', '--target-dir', work/'rust-target', env=env)
    run('wasm-tools', 'component', 'new', work/f'rust-target/wasm32-unknown-unknown/release/wasmvm_preview3_{binary}_guest.wasm', '-o', out/f'world-rust-{binary}.wasm')

# Independently preserve the whole stable import surface, including functions
# removed by language-compiler DCE. Only the CLI dummy export is executed.
for name, world in [('command', 'c-command'), ('service', 'rust-http'), ('middleware', 'rust-middleware')]:
    core = work/(name+'.core.wat')
    run('wasm-tools', 'component', 'embed', wit, '--world', world, '--dummy-names', 'legacy', '-t', '-o', core)
    if name == 'command':
        text = core.read_text()
        index = re.search(r'\(export "wasi:cli/run@0.3.2#run" \(func (\d+)\)\)', text).group(1)
        text, count = re.subn(r'(\(func \(;'+index+r';\)[^\n]*\n)\s+unreachable', r'\1    i32.const 0', text, count=1)
        assert count == 1
        core.write_text(text)
    run('wasm-tools', 'component', 'new', core, '-o', out/f'world-full-{name}.wasm')

run('wasm-tools', 'compose', out/'world-rust-middleware.wasm', '-d', out/'world-rust-http.wasm', '-o', out/'world-rust-composed-http.wasm')
names = ['world-c-command', 'world-rust-command', 'world-rust-http', 'world-rust-middleware', 'world-rust-composed-http', 'world-full-command', 'world-full-service', 'world-full-middleware']
for name in names:
    run('wasm-tools', 'validate', '--features', 'all', out/(name+'.wasm'))
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
files = {str(path.relative_to(source)): sha(path) for path in sorted(source.rglob('*')) if path.is_file()}
manifest = {
    'wasiVersion': '0.3.2', 'witArchiveSHA256': SHA,
    'producer': 'wasm-tools 1.258.0', 'bindingGenerator': 'wit-bindgen 0.54.0',
    'guestCompilers': ['clang 22.1.8 -O2 / wasi-libc 33', 'Rust 1.98.0 opt-level=2 panic=abort'],
    'bindings': {'c': 'sync ABI over native async WIT, native streams/futures/waitables', 'rust': 'native async exports/imports and bounded stream producers'},
    'unstableFeatures': [], 'generatorSHA256': sha(Path(__file__)),
    'sourcesSHA256': files,
    'binarySHA256': {name+'.wasm': sha(out/(name+'.wasm')) for name in names},
}
(out/'world-manifest.json').write_text(json.dumps(manifest, indent=2)+'\n')
print('Built', len(names), 'validated WASI 0.3.2 world fixtures')

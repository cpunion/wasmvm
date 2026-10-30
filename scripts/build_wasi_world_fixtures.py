#!/usr/bin/env python3
"""Build pinned C/Rust command guests and standard CLI world/contract fixtures."""
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
SHA = '2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive', type=Path, default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work', type=Path, default=ROOT/'artifacts/components/world-producer')
p.add_argument('--output', type=Path, default=ROOT/'pkg/wasi/preview2/testdata')
p.add_argument('--bindgen', type=Path, default=ROOT/'artifacts/components/wit-bindgen-0.54.0/bin/wit-bindgen')
p.add_argument('--sysroot', type=Path, default=Path('/opt/homebrew/opt/wasi-libc/share/wasi-sysroot'))
a = p.parse_args()
work, out = a.work.resolve(), a.output.resolve()
source = ROOT/'testdata/wasi-preview2/guests'
assert hashlib.sha256(a.archive.read_bytes()).hexdigest() == SHA
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
assert '0.54.0' in subprocess.check_output([str(a.bindgen), '--version'], text=True)
assert '22.1.8' in subprocess.check_output(['clang', '--version'], text=True)
work.mkdir(parents=True, exist_ok=True)
out.mkdir(parents=True, exist_ok=True)
shutil.copytree(source, work/'guests', dirs_exist_ok=True, ignore=shutil.ignore_patterns('deps', 'target'))
wit = work/'guests/wit'
(wit/'deps').mkdir(exist_ok=True)
with tarfile.open(a.archive) as archive:
    for member in archive.getmembers():
        if member.name.startswith('wasi-wit-0.2.12/cli/') and member.name.endswith('.wit'):
            (wit/'deps'/Path(member.name).name).write_bytes(archive.extractfile(member).read())

def run(*command, **kwargs):
    subprocess.run([str(x) for x in command], check=True, **kwargs)

run(a.bindgen, 'c', wit, '--world', 'c-command', '--out-dir', work/'c')
run('clang', '--target=wasm32-wasip1', f'--sysroot={a.sysroot}', '-O2', '-mexec-model=reactor', '-nodefaultlibs', '-I', work/'c', work/'guests/c/command.c', work/'c/c_command.c', work/'c/c_command_component_type.o', '-lc', '-Wl,--export-memory', '-o', work/'c.core.wasm')
run('wasm-tools', 'component', 'new', work/'c.core.wasm', '-o', out/'world-c-command.wasm')
env = os.environ.copy()
env['RUSTFLAGS'] = f'--remap-path-prefix={work}=/wasmvm-world-guests'
run('cargo', '+1.98.0', 'build', '--locked', '--release', '--target', 'wasm32-unknown-unknown', '--manifest-path', work/'guests/rust/Cargo.toml', '--target-dir', work/'rust-target', env=env)
run('wasm-tools', 'component', 'new', work/'rust-target/wasm32-unknown-unknown/release/wasmvm_wasi_command_guest.wasm', '-o', work/'rust-command.wasm')
run('wasm-tools', 'compose', work/'rust-command.wasm', '-d', out/'world-c-command.wasm', '-o', out/'world-composed-command.wasm')

# Keep the complete stable import surface, independently of compiler DCE.
with (wit/'world.wit').open('a') as f:
    f.write('\nworld full-command { include wasi:cli/command@0.2.12; }\n')
run('wasm-tools', 'component', 'embed', wit, '--world', 'full-command', '--dummy-names', 'legacy', '-t', '-o', work/'full-command.wat')
s = (work/'full-command.wat').read_text()
index = re.search(r'\(export "wasi:cli/run@0.2.12#run" \(func (\d+)\)\)', s).group(1)
s, n = re.subn(r'(\(func \(;'+index+r';\)[^\n]*\n)\s+unreachable', r'\1    i32.const 0', s, count=1)
assert n == 1
(work/'full-command.wat').write_text(s)
run('wasm-tools', 'parse', work/'full-command.wat', '-o', work/'full-command.core.wasm')
run('wasm-tools', 'component', 'new', work/'full-command.core.wasm', '-o', out/'world-full-command.wasm')

# Valid export mutations must reject before their observable host effect.
for mode in ['success', 'failure', 'trap', 'wrong-result', 'wrong-parameter', 'no-result', 'exit']:
    result = '(result (result))'
    core_result = '(result i32)'
    params = ''
    body = 'call $mark i32.const '+('1' if mode == 'failure' else '0')
    imports = '(import "mark" (func $mark)) (core func $mark (canon lower (func $mark)))'
    if mode == 'trap':
        body = 'call $mark unreachable'
    if mode == 'wrong-result':
        result = '(result u32)'
    if mode == 'wrong-parameter':
        params = '(param "value" u32)'
    if mode == 'no-result':
        result, core_result, body = '', '', 'call $mark'
    if mode == 'exit':
        imports = '''(import "wasi:cli/exit@0.2.12" (instance $exit
          (export "exit-with-code" (func (param "code" u8)))))
          (alias export $exit "exit-with-code" (func $exit))
          (core func $mark (canon lower (func $exit)))'''
        body = 'i32.const 255 call $mark unreachable'
    marker_params = '(param i32)' if mode == 'exit' else ''
    core_params = '(param i32)' if mode == 'wrong-parameter' else ''
    text = f'''(component
      {imports}
      (core module $m
        (import "host" "mark" (func $mark {marker_params}))
        (func (export "run") {core_params} {core_result} {body}))
      (core instance $host (export "mark" (func $mark)))
      (core instance $core (instantiate $m (with "host" (instance $host))))
      (func $run {params} {result} (canon lift (core func $core "run")))
      (instance $cli (export "run" (func $run)))
      (export "wasi:cli/run@0.2.12" (instance $cli)))'''
    path = work/f'{mode}.wat'
    path.write_text(text+'\n')
    run('wasm-tools', 'parse', path, '-o', out/f'world-command-{mode}.wasm')

bad_http = '''(component
  (import "wasi:http/types@0.2.12" (instance $types
    (export "incoming-request" (type (sub resource)))
    (export "response-outparam" (type (sub resource)))))
  (alias export $types "incoming-request" (type $request))
  (alias export $types "response-outparam" (type $outparam))
  (import "mark-incoming" (func $bad (param "request" (own $request))
    (param "outparam" (own $outparam)) (result u32)))
  (instance $incoming (export "handle" (func $bad)))
  (export "wasi:http/incoming-handler@0.2.12" (instance $incoming)))'''
(work/'bad-http.wat').write_text(bad_http+'\n')
run('wasm-tools', 'parse', work/'bad-http.wat', '-o', out/'world-invalid-incoming-result.wasm')
names = ['world-c-command', 'world-composed-command', 'world-full-command', 'world-invalid-incoming-result'] + ['world-command-'+x for x in ['success', 'failure', 'trap', 'wrong-result', 'wrong-parameter', 'no-result', 'exit']]
for name in names:
    run('wasm-tools', 'validate', '--features', 'all', out/f'{name}.wasm')
manifest = {
    'wasiVersion': '0.2.12', 'wasiArchiveSha256': SHA,
    'guestCompilers': ['clang 22.1.8 -O2 / wasi-libc 33', 'Rust 1.98.0 opt-level=2 panic=abort'],
    'bindingGenerator': 'wit-bindgen 0.54.0', 'componentProducer': 'wasm-tools 1.258.0',
    'unstableFeatures': [],
    'sourcesSha256': {str(f.relative_to(source)): hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(source.rglob('*')) if f.is_file() and 'deps' not in f.relative_to(source).parts},
    'binarySha256': {name+'.wasm': hashlib.sha256((out/(name+'.wasm')).read_bytes()).hexdigest() for name in names},
}
(out/'world-manifest.json').write_text(json.dumps(manifest, indent=2)+'\n')
print('Reproduced C/Rust composition, complete CLI imports and seven command controls')

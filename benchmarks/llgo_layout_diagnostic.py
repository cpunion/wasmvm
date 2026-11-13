"""Temporary ELF-layout ablation over the frozen PR169 comparison artifact."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import tarfile

import ci
import own_build
import routing_compare

SOURCE_RUN = 43448812030
HEAD = '95943842febc2d3571547c72f2d62bc79f232f88'
BASE = 'b546ea76663c2ee763ce303a7835e1f83096e59d'
PROFILES = {
    'fixed-page': ('-Wl,-Ttext=0x400000', 0x400000),
    'fixed-base-phase': ('-Wl,-Ttext=0x400030', 0x400030),
    'separate-code': ('-Wl,-z,separate-code', None),
}
SYMBOL = 'github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlots'


def load(folder):
    data = json.loads((folder / 'build.json').read_text())
    if (data['head'], data['base']) != (HEAD, BASE):
        raise ValueError('wrong frozen source revisions')
    for row in data['builds'].values():
        if own_build.digest(Path(row['binary'])) != row['binarySha256']:
            raise ValueError('frozen binary changed')
    return data


def inspect(binary, destination, expected):
    objdump = shutil.which('llvm-objdump-22') or shutil.which('llvm-objdump')
    nm = shutil.which('llvm-nm-22') or shutil.which('llvm-nm')
    sections = ci.capture([objdump, '--section-headers', str(binary)])
    match = re.search(r'^\s*\d+\s+\.text\s+[0-9a-fA-F]+\s+([0-9a-fA-F]+)', sections, re.M)
    if not match:
        raise ValueError('missing ELF text section')
    text = int(match[1], 16)
    if expected is not None and text != expected:
        raise ValueError('requested fixed ELF text address was not applied')
    if expected is None and text % 4096:
        raise ValueError('separate-code did not page-align text')
    row = next(line.split() for line in ci.capture([nm, '--print-size', str(binary)]).splitlines()
               if line.endswith(' ' + SYMBOL))
    address, size = int(row[0], 16), int(row[1], 16)
    assembly = ci.capture([objdump, '--disassemble-symbols=' + SYMBOL, '--no-show-raw-insn',
                           '--no-leading-addr', '--symbolize-operands', str(binary)])
    destination.write_text(assembly)
    return dict(text=text, dispatcher=address, size=size, textOffset=address-text,
                mod64=address % 64, sections=sections)


def build(folder, role):
    data = load(folder)
    selected = own_build.toolchains()['llgo']
    original = data['builds'][role + '-llgo']
    if selected['binarySha256'] != original['toolchain']['binarySha256']:
        raise ValueError('compiler executable differs from the frozen comparison')
    target = Path(original['binary']).parent
    if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != original['sourceHashes']:
        raise ValueError('frozen engine source changed')
    if os.environ.get('LDFLAGS'):
        raise ValueError('unexpected preexisting linker environment')
    probe = folder / ('layout-probe-' + role)
    probe.mkdir()
    fragment = dict(sourceRun=SOURCE_RUN, role=role, builds={}, runtimes={}, profiles=PROFILES)
    files = []
    for profile, (flag, expected) in PROFILES.items():
        name = profile + '-' + role + '-llgo'
        # Same bytes, newer mtime: request a link using LLGo's accepted cache
        # design. No package identity or implementation source is changed.
        os.utime(target / 'benchmarks/embedded/wasmvm/main.go', None)
        command = list(original['command'])
        command[0] = selected['command']
        binary = target / ('runner-' + profile)
        command[command.index('-o') + 1] = str(binary)
        env = dict(os.environ, CGO_ENABLED='1', GOWORK='off', GOTOOLCHAIN='local',
                   GOFLAGS='', GOEXPERIMENT='', LDFLAGS=flag)
        print('building', name, flag, flush=True)
        record = routing_compare.compile_runner(command, target, env)
        layout = inspect(binary, probe / (name + '.asm'), expected)
        fragment['builds'][name] = dict(original, **record, toolchain=selected,
            experimentalLinkerFlags=[flag], effectiveEnvironment={'LDFLAGS': flag},
            layoutDiagnostic=layout, frozenSourceRun=SOURCE_RUN)
        runtime = dict(data['runtimes'][role + '-llgo-interpreter'], command=[str(binary), '--mode', 'api'])
        fragment['runtimes'][name + '-interpreter'] = runtime
        fragment['runtimes']['aa-' + name + '-interpreter'] = dict(runtime)
        files.append(binary)
        print('verified', name, json.dumps(layout, sort_keys=True), flush=True)
    if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != original['sourceHashes']:
        raise ValueError('variant build changed engine source')
    manifest = folder / ('layout-' + role + '.json')
    manifest.write_text(json.dumps(fragment, indent=2) + '\n')
    with tarfile.open('llgo-layout-' + role + '.tar.gz', 'w:gz') as archive:
        for path in [manifest, probe, *files]:
            archive.add(path, arcname=str(path.relative_to(folder.parent)))


def merge(folder):
    data = load(folder)
    for role in ['base', 'head']:
        fragment = json.loads((folder / ('layout-' + role + '.json')).read_text())
        if fragment['sourceRun'] != SOURCE_RUN or fragment['role'] != role:
            raise ValueError('wrong variant provenance')
        data['builds'].update(fragment['builds'])
        data['runtimes'].update(fragment['runtimes'])
    data['layoutDiagnostic'] = dict(sourceRun=SOURCE_RUN, profiles=PROFILES,
        fixedBlocks=20, expectedObservations=640, expectedAlternateChecks=31,
        acceptance='Recover head against original fast default base; slowing both roles is not recovery.')
    (folder / 'build.json').write_text(json.dumps(data, indent=2) + '\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['build', 'merge'])
    parser.add_argument('--role', choices=['base', 'head'])
    parser.add_argument('--build-dir', type=Path, required=True)
    args = parser.parse_args()
    if args.command == 'build':
        if args.role is None:
            parser.error('--role is required for build')
        build(args.build_dir.resolve(), args.role)
    else:
        merge(args.build_dir.resolve())

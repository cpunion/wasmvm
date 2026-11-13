"""Temporary ELF-layout ablation over the frozen PR169 comparison artifact."""
import argparse
import difflib
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
    'function-32': ('-Wl,-mllvm,-align-all-functions=5', 32),
    'function-64': ('-Wl,-mllvm,-align-all-functions=6', 64),
    'loop-32': ('-Wl,-mllvm,-align-loops=32', 0),
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
    row = next(line.split() for line in ci.capture([nm, '--print-size', str(binary)]).splitlines()
               if line.endswith(' ' + SYMBOL))
    address, size = int(row[0], 16), int(row[1], 16)
    if expected and (text % expected or address % expected):
        raise ValueError('requested function alignment was not applied')
    assembly = ci.capture([objdump, '--disassemble-symbols=' + SYMBOL, '--no-show-raw-insn',
                           '--no-leading-addr', '--symbolize-operands', str(binary)])
    destination.write_text(assembly)
    return dict(text=text, dispatcher=address, size=size, textOffset=address-text,
                mod64=address % 64, sections=sections)


def build(folder, role):
    data = load(folder)
    selected = own_build.toolchains()['llgo']
    original = data['builds'][role + '-llgo']
    previous = original['toolchain']
    # Preserve both CLI hashes; require identical compiler/dependency/build
    # identities across fresh installations instead of assuming reproducibility.
    before = previous['goBuildInfo'].splitlines()[1:]
    after = selected['goBuildInfo'].splitlines()[1:]
    if before != after:
        print('\n'.join(difflib.unified_diff(before, after)), flush=True)
        raise ValueError('compiler dependency or build identity changed')
    print('compiler identities', json.dumps(dict(original=previous, actual=selected)), flush=True)
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
            layoutDiagnostic=layout, frozenSourceRun=SOURCE_RUN,
            frozenCompilerSha256=previous['binarySha256'])
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


def build_poll(folder, candidate, compiler):
    data = load(folder)
    selected = own_build.toolchains()[compiler]
    previous = data['builds']['head-' + compiler]['toolchain']
    if compiler == 'llgo':
        if previous['goBuildInfo'].splitlines()[1:] != selected['goBuildInfo'].splitlines()[1:]:
            raise ValueError('compiler dependency or build identity changed')
    elif selected['binarySha256'] != previous['binarySha256']:
        raise ValueError('Go executable differs from the frozen comparison')
    revision = ci.capture(['git', '-C', str(candidate), 'rev-parse', 'HEAD'])
    name = 'poll-' + compiler
    target = folder / name
    identity = own_build.prepare(candidate, candidate, revision, target)
    phase = target / 'internal/bench/phase/phase.go'
    phase.write_text(routing_compare.instrument_own(phase.read_text()))
    (phase.parent / 'provenance_gen.go').write_text('package phase\nfunc init(){sourceRevision=' + json.dumps(revision) + ';compilerRevision=' + json.dumps(selected['revision']) + '}\n')
    env = dict(os.environ, CGO_ENABLED='1', GOWORK='off', GOTOOLCHAIN='local', GOFLAGS='', GOEXPERIMENT='')
    if env.get('LDFLAGS'):
        raise ValueError('unexpected alignment or other linker override')
    command = [selected['command'], 'build', *selected['buildFlags'], '-o', str(target / 'runner'), './benchmarks/embedded/wasmvm']
    record = routing_compare.compile_runner(command, target, env)
    if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != identity['sourceHashes']:
        raise ValueError('candidate instrumentation changed engine sources')
    identity['harnessHashes'] = own_build.hashes(target, own_build.HARNESS)
    build = dict(record, **identity, toolchain=selected, instrumentation=phase.read_text(), frozenSourceRun=SOURCE_RUN)
    runtimes = {}
    for backend in ['interpreter', 'native']:
        runtime = dict(data['runtimes']['head-' + compiler + '-' + backend],
            command=[record['binary'], '--mode', 'api'], sourceRevision=revision)
        runtimes[name + '-' + backend] = runtime
        runtimes['aa-' + name + '-' + backend] = dict(runtime)
    fragment = folder / (name + '.json')
    fragment.write_text(json.dumps(dict(sourceRun=SOURCE_RUN, candidateRevision=revision,
        compiler=compiler, builds={name: build}, runtimes=runtimes), indent=2) + '\n')
    with tarfile.open(name + '.tar.gz', 'w:gz') as archive:
        paths = [fragment] + [target / p for p in (*own_build.ENGINE_PATHS, 'benchmarks/embedded/wasmvm', 'runner') if (target / p).exists()]
        for path in paths:
            archive.add(path, arcname=str(path.relative_to(folder.parent)))


def merge_poll(folder):
    data = load(folder)
    revisions = set()
    for compiler in ['go', 'llgo']:
        fragment = json.loads((folder / ('poll-' + compiler + '.json')).read_text())
        if fragment['sourceRun'] != SOURCE_RUN or fragment['compiler'] != compiler:
            raise ValueError('wrong cancellation candidate provenance')
        revisions.add(fragment['candidateRevision'])
        data['builds'].update(fragment['builds'])
        data['runtimes'].update(fragment['runtimes'])
    if len(revisions) != 1:
        raise ValueError('Go and LLGo candidate sources differ')
    data['cancellationDiagnostic'] = dict(sourceRun=SOURCE_RUN, candidateRevision=revisions.pop(),
        fixedBlocks=20, expectedObservations=560, expectedAlternateChecks=27,
        acceptance='Compare candidate against unchanged production head AND original fast LLGo base; no alignment flags.')
    (folder / 'build.json').write_text(json.dumps(data, indent=2) + '\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['build', 'merge', 'build-poll', 'merge-poll'])
    parser.add_argument('--role', choices=['base', 'head'])
    parser.add_argument('--build-dir', type=Path, required=True)
    parser.add_argument('--candidate-dir', type=Path)
    parser.add_argument('--compiler', choices=['go', 'llgo'])
    args = parser.parse_args()
    if args.command == 'build':
        if args.role is None:
            parser.error('--role is required for build')
        build(args.build_dir.resolve(), args.role)
    elif args.command == 'merge':
        merge(args.build_dir.resolve())
    elif args.command == 'build-poll':
        if args.candidate_dir is None or args.compiler is None:
            parser.error('--candidate-dir and --compiler are required')
        build_poll(args.build_dir.resolve(), args.candidate_dir.resolve(), args.compiler)
    else:
        merge_poll(args.build_dir.resolve())

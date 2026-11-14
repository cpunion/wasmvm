"""Temporary fixed-engine compiler upgrade comparison over frozen references."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile

import own_build
import routing_compare

ROOT = Path.cwd()
ENGINE = '688b40b67c9ce27607a52be2d3e394e2000c2e77'
OLD = '56fa9aedc957cf3c9656c1a70f82c5424ab3b1c5'
NEW = 'ac1292206d672c2aeb546780a1d585ad98edbc3d'
FLAGS = ['-O3', '-lto=full']
REFERENCES = ['wazero-interpreter', 'wazero-compiler', 'wasmtime-cranelift',
              'wasmer-cranelift', 'wasmedge-interpreter']


def libraries():
    paths = [str(ROOT / 'routing-build/host-libraries')]
    paths += sorted({str(p.parent) for p in (ROOT / 'routing-build/tools').rglob('*.so*') if p.is_file()})
    if os.environ.get('LD_LIBRARY_PATH'):
        paths.append(os.environ['LD_LIBRARY_PATH'])
    os.environ['LD_LIBRARY_PATH'] = ':'.join(paths)


def build():
    with tarfile.open('routing-build.tar.gz') as archive:
        archive.extractall(ROOT, filter='data')
    folder = ROOT / 'routing-build'
    data = json.loads((folder / 'build.json').read_text())
    with tarfile.open(ROOT / 'downloads/poll-llgo.tar.gz') as archive:
        archive.extractall(ROOT, filter='data')
    old = json.loads((folder / 'poll-llgo.json').read_text())
    with tarfile.open(ROOT / 'downloads/poll-go.tar.gz') as archive:
        go = json.load(archive.extractfile('routing-build/poll-go.json'))
        go_bytes = archive.extractfile('routing-build/poll-go/runner').read()
    assert old['candidateRevision'] == go['candidateRevision'] == ENGINE
    previous = old['builds']['poll-llgo']
    go_build = go['builds']['poll-go']
    assert previous['sourceHashes'] == go_build['sourceHashes']
    source = folder / 'poll-llgo'
    assert own_build.hashes(source, own_build.ENGINE_PATHS, own_build.HARNESS) == previous['sourceHashes']
    assert previous['toolchain']['revision'] == OLD and previous['toolchain']['buildFlags'] == FLAGS
    assert own_build.digest(source / 'runner') == previous['binarySha256']
    (source / 'runner').rename(source / 'runner-base-llgo')
    (source / 'runner-go').write_bytes(go_bytes)
    (source / 'runner-go').chmod(0o755)
    assert own_build.digest(source / 'runner-go') == go_build['binarySha256']
    cli = shutil.which('llgo')
    version = subprocess.check_output([cli, 'version'], text=True)
    assert NEW[:12] in version and os.environ['WASMVM_LLGO_REVISION'] == NEW
    assert subprocess.check_output(['go', 'env', 'GOVERSION'], text=True).strip() == 'go1.27.3'
    assert subprocess.check_output(['llvm-config', '--version'], text=True).strip() == '22.1.8'
    generated = source / 'internal/bench/phase/provenance_gen.go'
    generated.write_text('package phase\nfunc init(){sourceRevision=' + json.dumps(ENGINE) + ';compilerRevision=' + json.dumps(NEW) + '}\n')
    libraries()
    environment = dict(os.environ, CGO_ENABLED='1', GOWORK='off', GOTOOLCHAIN='local', GOFLAGS='', GOEXPERIMENT='')
    assert not environment.get('LDFLAGS')
    command = [cli, 'build', *FLAGS, '-o', str(source / 'runner'), './benchmarks/embedded/wasmvm']
    record = routing_compare.compile_runner(command, source, environment)
    assert own_build.hashes(source, own_build.ENGINE_PATHS, own_build.HARNESS) == previous['sourceHashes']
    latest = dict(record, sourceRevision=ENGINE, sourceHashes=previous['sourceHashes'],
                  harnessHashes=own_build.hashes(source, own_build.HARNESS),
                  toolchain=dict(revision=NEW, buildFlags=FLAGS, version=version, command=cli, goos='linux', goarch='amd64',
                      binarySha256=own_build.digest(Path(cli)),
                      goBuildInfo=subprocess.check_output(['go', 'version', '-m', cli], text=True)))
    ref_paths = {data['runtimes'][name]['command'][0] for name in REFERENCES}
    ref_paths.add(data['builds']['native-c']['binary'])
    data['builds'] = {name: row for name, row in data['builds'].items() if row.get('binary') in ref_paths}
    assert 'native-c' in data['builds']
    base = dict(previous, binary=str(source / 'runner-base-llgo'), frozenBuildRun=43561639936)
    frozen_go = dict(go_build, binary=str(source / 'runner-go'), frozenBuildRun=43561639936)
    data['builds'].update({'base-llgo': base, 'head-llgo': latest, 'base-go': frozen_go, 'head-go': frozen_go})
    runtimes = {name: data['runtimes'][name] for name in REFERENCES}
    for role, binary, compiler in [('base', base['binary'], OLD), ('head', latest['binary'], NEW)]:
        for backend in ['interpreter', 'native']:
            row = dict(old['runtimes']['poll-llgo-' + backend], command=[binary, '--mode', 'api'], compilerRevision=compiler)
            runtimes[role + '-llgo-' + backend] = row
    for role, backend in [('base', 'native'), ('head', 'native'), ('head', 'interpreter')]:
        runtimes[role + '-go-' + backend] = dict(go['runtimes']['poll-go-' + backend], command=[frozen_go['binary'], '--mode', 'api'])
    for name, row in list(runtimes.items()):
        runtimes['aa-' + name] = dict(row)
    data['sourceArtifact'] = dict(run=43448812030, base=data['base'], head=data['head'],
        linkage=data.pop('linkage', None), scope='Historical frozen SDK/reference provenance; old owned roles are retired.')
    data.update(base=ENGINE, head=ENGINE, runtimes=runtimes,
        compilerUpgrade=dict(old=OLD, new=NEW, flags=FLAGS, engine=ENGINE, buildRun=43561639936,
            expectedObservations=500, alternateChecks=24, controls=12,
            scope='Only the host compiler/runtime and its identity change. Go native base/head labels share one frozen control binary. No source optimization, layout override or GC override.'))
    assert len(runtimes) == 24 and sum(name.startswith('aa-') for name in runtimes) == 12
    (folder / 'upgrade.json').write_text(json.dumps(data, indent=2) + '\n')
    with tarfile.open('compiler-upgrade.tar.gz', 'w:gz') as archive:
        archive.add(folder / 'upgrade.json', arcname='routing-build/upgrade.json')
        for relative in (*own_build.ENGINE_PATHS, 'benchmarks/embedded/wasmvm', 'runner', 'runner-base-llgo', 'runner-go'):
            path = source / relative
            if path.exists():
                archive.add(path, arcname='routing-build/poll-llgo/' + relative)


def measure(case, group):
    folder = ROOT / 'routing-build'
    with tarfile.open('routing-build.tar.gz') as archive:
        archive.extractall(ROOT, filter='data')
    with tarfile.open('compiler-upgrade.tar.gz') as archive:
        archive.extractall(ROOT, filter='data')
    libraries()
    data = json.loads((folder / 'upgrade.json').read_text())
    assert data['compilerUpgrade']['engine'] == ENGINE
    assert case in {row['id'] for row in data['fixtures']}
    (folder / 'build.json').write_text(json.dumps(data, indent=2) + '\n')
    routing_compare.measure(argparse.Namespace(head_dir=ROOT / 'head', build_dir=folder,
        output_dir=ROOT / 'upgrade-results', group=group, case_id=case, blocks=20,
        gc_sustained=False, threads_sustained=False))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['build', 'measure'])
    parser.add_argument('--case')
    parser.add_argument('--group')
    args = parser.parse_args()
    if args.command == 'build':
        build()
    else:
        if not args.case or not args.group:
            parser.error('measure requires --case and --group')
        measure(args.case, args.group)

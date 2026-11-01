#!/usr/bin/env python3
"""Pin all upstream component scripts and execute the declared runtime profile."""
import argparse
from collections import Counter
import gzip
import hashlib
import io
import json
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile

from core2_corpus import digest, extract, inventory, relative_path

ROOT = Path(__file__).resolve().parents[1]
RECORDS = ROOT/'testdata/component-runtime'
REVISION = 'd095e2f60ddae1f9a1e9b1e2fb547ec1114da677'
SOURCE_SHA = 'dbd68be264dbe55110ed489377aa84281dada28c1da416d31f86b5623e69d2d2'
LIMITATIONS = {'async/big-interleaving-test': 'stream.forward', 'async/forward': 'stream.forward',
               'values/post-return': 'stream.forward', 'validation/version-suffix': 'version-suffix'}
ADDITIONAL_PRODUCER = 'wasm-tools 1.262.0 (e22db44da 2026-10-13)'
PRODUCER_SHA = '4a5b127f48340fbd453708283d190f9b1dbbe5820571eae5cbae0a205a095a00'


def select_commands(directory, commands, producer='wasm-tools'):
    """Classify optional threading from independently printed canonical opcodes."""
    modules, instances, current = {}, {}, []
    selected, deferred = [], {}
    for index, command in enumerate(commands):
        kind = command['type']
        features = []
        if 'filename' in command:
            result = subprocess.run([producer, 'print', directory/command['filename']],
                                    check=True, capture_output=True, text=True)
            features = sorted(set(re.findall(r'\(canon thread\.([a-z-]+)', result.stdout)) - {'yield'})
        if kind in ('module', 'module_definition'):
            modules[command.get('name', '')] = features
            if kind == 'module':
                current = features
                instances[command.get('name', '')] = features
        elif kind == 'module_instance':
            features = modules[command['module']]
            current = features
            instances[command.get('instance', '')] = features
        elif kind in ('assert_return', 'assert_trap', 'action'):
            name = command['action'].get('module')
            features = instances[name] if name else current
        elif kind not in ('assert_invalid', 'assert_uninstantiable'):
            raise ValueError('unknown runtime command: '+kind)
        if features:
            deferred[str(index)] = {'profile': 'component-threading', 'builtins': features}
        else:
            selected.append(index)
    return selected, deferred


def regenerate(source, destination, additional):
    assert digest(source) == SOURCE_SHA, 'changed upstream archive'
    assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
    assert digest(additional) == PRODUCER_SHA
    assert subprocess.check_output([additional, '--version'], text=True).strip() == ADDITIONAL_PRODUCER
    with tempfile.TemporaryDirectory() as temp:
        work = Path(temp)
        prefix = 'component-model-'+REVISION+'/'
        def select(name):
            if not name.startswith(prefix): return None
            name = name[len(prefix):]
            if name in ('LICENSE', 'LICENSE-APACHE'): return name
            if name.startswith('test/') and name.endswith('.wast'): return 'source/'+name[5:]
            return None
        extract(source, work, select)
        sources = sorted((work/'source').rglob('*.wast'))
        assert len(sources) == 75
        scripts, resolved = {}, {}
        for source_file in sources:
            name = source_file.relative_to(work/'source').with_suffix('').as_posix()
            if name.startswith(('validation/', 'binary/')) and name not in LIMITATIONS:
                continue  # Complete grammar/type scripts already have a separate corpus.
            directory = work/'runtime'/name
            directory.mkdir(parents=True)
            result = subprocess.run(['wasm-tools', 'json-from-wast', source_file,
                '--wasm-dir', directory, '-o', directory/'script.json'], capture_output=True, text=True)
            producer = 'wasm-tools'
            if name in LIMITATIONS:
                assert result.returncode != 0 and LIMITATIONS[name] in result.stderr
                resolved[name] = {'priorProducer': 'wasm-tools 1.258.0', 'token': LIMITATIONS[name],
                                 'diagnostic': result.stderr.replace(str(work)+'/', ''),
                                 'replacementProducer': ADDITIONAL_PRODUCER}
                # The producer may leave incomplete output; never admit those bytes.
                for file in directory.iterdir(): file.unlink()
                producer = str(additional)
                result = subprocess.run([producer, 'json-from-wast', source_file,
                    '--wasm-dir', directory, '-o', directory/'script.json'], capture_output=True, text=True)
            assert result.returncode == 0, result.stderr
            path = directory/'script.json'
            script = json.loads(path.read_text())
            script['source_filename'] = source_file.relative_to(work).as_posix()
            path.write_text(json.dumps(script, indent=2)+'\n')
            selected, deferred = select_commands(directory, script['commands'], producer)
            counts = Counter(script['commands'][i]['type'] for i in selected)
            scripts[name] = {'selectedCommands': selected, 'deferredCommands': deferred,
                             'commands': len(script['commands']), 'selectedCounts': dict(sorted(counts.items()))}
        # Deterministic local archive; sources and optional-profile inputs stay intact.
        hashes = inventory(work)
        with io.BytesIO() as buffer:
            with gzip.GzipFile(fileobj=buffer, mode='wb', filename='', mtime=0) as compressed:
                with tarfile.open(fileobj=compressed, mode='w') as archive:
                    for name in sorted(hashes):
                        data = (work/name).read_bytes()
                        info = tarfile.TarInfo(name); info.size = len(data); info.mode = 0o644
                        archive.addfile(info, io.BytesIO(data))
            data = buffer.getvalue()
        RECORDS.mkdir(exist_ok=True)
        (RECORDS/'inputs.tar.gz').write_bytes(data)
        manifest = {'revision': REVISION, 'sourceArchiveSHA256': SOURCE_SHA,
            'producer': 'wasm-tools 1.258.0', 'sourceScripts': 75,
            'archiveSHA256': hashlib.sha256(data).hexdigest(),
            'generatorSHA256': digest(Path(__file__)), 'sha256': hashes,
            'scripts': scripts, 'producerLimitations': {}, 'resolvedProducerLimitations': resolved,
            'additionalProducer': {'version': ADDITIONAL_PRODUCER, 'binarySHA256': PRODUCER_SHA,
                'archiveSHA256': '801b237643cc0f4ab27ebd1745e7863e23bf194a038bc6ca8ce5618212e43669',
                'url': 'https://github.com/bytecodealliance/wasm-tools/releases/download/v1.262.0/wasm-tools-1.262.0-aarch64-macos.tar.gz'},
            'separateValidationCorpus': '../component-binary/manifest.json'}
        (RECORDS/'manifest.json').write_text(json.dumps(manifest, indent=2)+'\n')
    return prepare(destination)


def prepare(destination=None):
    destination = Path(destination or ROOT/'artifacts/components/runtime-corpus')
    manifest = json.loads((RECORDS/'manifest.json').read_text())
    assert manifest['revision'] == REVISION and manifest['producer'] == 'wasm-tools 1.258.0'
    assert manifest['generatorSHA256'] == digest(Path(__file__))
    assert manifest['archiveSHA256'] == digest(RECORDS/'inputs.tar.gz')
    expected = manifest['sha256']
    destination.mkdir(parents=True, exist_ok=True)
    with tarfile.open(RECORDS/'inputs.tar.gz') as archive:
        members = archive.getmembers()
        assert len(members) == len(expected) and {m.name for m in members} == set(expected)
        for member in members:
            relative_path(member.name)
            assert member.isfile() and hashlib.sha256(archive.extractfile(member).read()).hexdigest() == expected[member.name]
        archive.extractall(destination, filter='data')
    assert inventory(destination) == expected
    selected = deferred = 0
    for name, row in manifest['scripts'].items():
        commands = json.loads((destination/'runtime'/name/'script.json').read_text())['commands']
        assert len(commands) == row['commands']
        assert set(row['selectedCommands']).isdisjoint(int(i) for i in row['deferredCommands'])
        assert set(row['selectedCommands']) | {int(i) for i in row['deferredCommands']} == set(range(len(commands)))
        assert dict(Counter(commands[i]['type'] for i in row['selectedCommands'])) == row['selectedCounts']
        selected += len(row['selectedCommands']); deferred += len(row['deferredCommands'])
    print(f'Component runtime: {selected} selected commands; {deferred} optional-threading commands; '
          f'{len(manifest["producerLimitations"])} explicit producer limitations; 75 original scripts retained')
    return manifest


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--regenerate', action='store_true')
    parser.add_argument('--archive', type=Path, default=ROOT/'artifacts/components/component-model.tar.gz')
    parser.add_argument('--destination', type=Path)
    parser.add_argument('--additional-producer', type=Path, default=ROOT/'artifacts/components/wasm-tools-1.262.0/wasm-tools-1.262.0-aarch64-macos/wasm-tools')
    args = parser.parse_args()
    regenerate(args.archive, args.destination, args.additional_producer) if args.regenerate else prepare(args.destination)

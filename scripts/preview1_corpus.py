#!/usr/bin/env python3
"""Prepare and verify the offline, pinned complete applicable P1 guest corpus."""
from collections import Counter
import hashlib
import json
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parent.parent
RECORDS = ROOT / 'testdata/wasi-preview1'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(output, records=RECORDS):
    manifest = json.loads((records / 'manifest.json').read_text())
    if (manifest['sourceRevision'] != '58b6206a46013fb53d23990e5a650584a5361043'
            or manifest['prebuiltRevision'] != '9dfd4867f12498bd6411a57739c24b4880668683'
            or len(manifest['tests']) != 72):
        raise ValueError('P1 inventory identity changed')
    paths = [entry['path'] for entry in manifest['tests']]
    expected = {p.relative_to(output).as_posix() for lang in ('c', 'rust', 'assemblyscript')
                for p in (output / lang).glob('*.wasm')}
    if (len(set(paths)) != 72 or set(paths) != expected
            or Counter(Path(path).parts[0] for path in paths)
            != {'c': 14, 'rust': 46, 'assemblyscript': 12}):
        raise ValueError('P1 complete program inventory changed')
    actual = {p.relative_to(output).as_posix(): digest(p)
              for p in sorted(output.rglob('*')) if p.is_file()}
    if actual != manifest['sha256']:
        raise ValueError('P1 corpus file inventory/checksum mismatch')
    return manifest


def prepare(output=ROOT / 'artifacts/preview1/fixtures', records=RECORDS):
    pins = json.loads((records / 'pins.json').read_text())['fixtures']
    archive = records / pins['file']
    if archive.stat().st_size != pins['size'] or digest(archive) != pins['sha256']:
        raise ValueError('P1 input archive checksum mismatch')
    if not output.exists():
        output.mkdir(parents=True)
        with tarfile.open(archive) as stream:
            stream.extractall(output, filter='data')
    manifest = verify(output, records)
    print(f"Verified {len(manifest['tests'])} P1 programs and {len(manifest['sha256'])} asset hashes")
    return manifest


if __name__ == '__main__':
    prepare()

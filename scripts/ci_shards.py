#!/usr/bin/env python3
"""Check that LLGo shards run every repository test package exactly once."""
import argparse
from collections import Counter
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
INPUTS = {'core2', 'core3', 'core_extensions', 'preview1', 'wasi_threads', 'component'}


def validate(shards, root=ROOT):
    expected = {
        './' + str(path.parent.relative_to(root))
        for directory in ('internal', 'pkg', 'cmd')
        for path in (root / directory).rglob('*_test.go')
        if 'testdata' not in path.relative_to(root).parts
    }
    packages = []
    names = set()
    for shard in shards:
        name = shard['name']
        if not re.fullmatch('[a-z]+', name) or name in names:
            raise ValueError('invalid or duplicate shard name')
        names.add(name)
        paths = shard['packages'].split()
        if not paths or any(not re.fullmatch(r'\./(?:internal|pkg|cmd)/[a-z0-9_/-]+', p) for p in paths):
            raise ValueError('invalid package list')
        if not set(shard['inputs'].split()) <= INPUTS or not isinstance(shard['cli'], bool):
            raise ValueError('invalid shard input/CLI configuration')
        packages.extend(paths)
    duplicate = [p for p, count in Counter(packages).items() if count != 1]
    if duplicate or set(packages) != expected:
        raise ValueError(f'package coverage mismatch: duplicates={duplicate}, '
                         f'missing={sorted(expected - set(packages))}, '
                         f'extra={sorted(set(packages) - expected)}')
    if sum(s['cli'] for s in shards) != 1:
        raise ValueError('exactly one shard must execute the CLI')
    return len(packages)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--github-output', type=Path)
    args = parser.parse_args()
    shards = json.loads((ROOT / '.github/llgo-shards.json').read_text())
    count = validate(shards)
    if args.github_output:
        with args.github_output.open('a') as output:
            output.write('shards=' + json.dumps(shards, separators=(',', ':')) + '\n')
    print(f'Validated {count} test packages, once each across {len(shards)} LLGo shards')


if __name__ == '__main__':
    main()

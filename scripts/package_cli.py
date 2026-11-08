#!/usr/bin/env python3
"""Package an exact-checkout CLI and verify the extracted executable.

This creates CI candidate artifacts, never tags or publishes a release. The CI
build step supplies the binary; hashes/provenance are not reproducible-build or
third-party security certification.
"""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parent.parent
GO_VERSION = 'go1.27.3'
LLGO_REVISION = '38de1e218b8146ba73f9c6f2e60e3c31f29d96b6'


def capture(command):
    return subprocess.check_output(command, cwd=ROOT, text=True, stderr=subprocess.STDOUT, timeout=60).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def json_stream(text):
    decoder = json.JSONDecoder()
    while text.strip():
        value, end = decoder.raw_decode(text.lstrip())
        yield value
        text = text.lstrip()[end:]


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + '\n')


def archive_folder(stage, output, timestamp):
    # Canonical tar/gzip metadata; compiler-produced binaries may still differ
    # between rebuilds. Do not promise bit-for-bit toolchain reproducibility.
    with output.open('wb') as raw, gzip.GzipFile(filename='', fileobj=raw, mode='wb', mtime=0) as zipped:
        with tarfile.open(fileobj=zipped, mode='w') as archive:
            for path in sorted(stage.rglob('*')):
                if not path.is_file():
                    continue
                info = archive.gettarinfo(str(path), arcname=str(path.relative_to(stage)))
                info.uid = info.gid = 0
                info.uname = info.gname = ''
                info.mtime = timestamp
                info.mode = 0o755 if path == stage / 'bin/wasmvm' else 0o644
                archive.addfile(info, io.BytesIO(path.read_bytes()))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('executable', type=Path)
    parser.add_argument('--compiler', choices=('go', 'llgo'), required=True)
    parser.add_argument('--native', choices=('available', 'unavailable'), required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if capture(['git', 'status', '--porcelain', '--untracked-files=all']):
        parser.error('candidate packages require a clean tracked checkout')
    revision = capture(['git', 'rev-parse', 'HEAD'])
    tree = capture(['git', 'rev-parse', 'HEAD^{tree}'])
    go_version = capture(['go', 'version'])
    if go_version.split()[2] != GO_VERSION:
        parser.error('the candidate requires pinned ' + GO_VERSION)
    goos, goarch = capture(['go', 'env', 'GOOS', 'GOARCH']).splitlines()
    host = ('darwin' if sys.platform == 'darwin' else sys.platform, platform.machine())
    if (goos, goarch) not in [('linux', 'amd64'), ('darwin', 'arm64')] or host != (goos, {'amd64': 'x86_64', 'arm64': 'arm64'}[goarch]):
        parser.error('run packaging on its supported native target, without cross-compilation')
    compiler = shutil.which(args.compiler)
    if not compiler:
        parser.error('compiler executable not found')
    compiler_version = capture([compiler, 'version'])
    flags = []
    if args.compiler == 'llgo':
        if LLGO_REVISION[:12] not in compiler_version or args.native != 'available':
            parser.error('LLGo requires the pinned native release profile')
        flags = ['-O3', '-lto=full']
    executable = args.executable.resolve()
    build_info = capture(['go', 'version', '-m', str(executable)]) if args.compiler == 'go' else None
    if build_info is not None:
        expected_cgo = 'CGO_ENABLED=' + ('1' if args.native == 'available' else '0')
        if GO_VERSION not in build_info.splitlines()[0] or expected_cgo not in build_info or 'github.com/cpunion/wasmvm/cmd/wasmvm' not in build_info or 'vcs.revision=' + revision not in build_info or 'vcs.modified=false' not in build_info:
            parser.error('binary build information does not match the declared Go/native profile')
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    name = 'wasmvm-' + revision[:12] + '-' + goos + '-' + goarch + '-' + args.compiler + '-' + ('native' if args.native == 'available' else 'portable')
    archive_path = output / (name + '.tar.gz')
    if archive_path.exists():
        parser.error('candidate output already exists: ' + str(archive_path))
    packages = list(json_stream(capture(['go', 'list', '-deps', '-json', './cmd/wasmvm'])))
    modules = {p['Module']['Path']: p['Module'] for p in packages if 'Module' in p}
    modules = [modules[name] for name in sorted(modules)]
    links = capture(['otool', '-L', str(executable)] if goos == 'darwin' else ['ldd', str(executable)]) if args.native == 'available' else 'Portable Go interpreter; CGO_ENABLED=0; native backend unavailable.'
    loader = capture(['otool', '-l', str(executable)]) if goos == 'darwin' else capture(['readelf', '--version-info', str(executable)])
    versions = re.findall(r'Name: GLIBC_([0-9.]+)', loader)
    glibc = max(versions, key=lambda s: tuple(map(int, s.split('.')))) if versions else None
    with tempfile.TemporaryDirectory(prefix='wasmvm-package-') as temporary:
        stage = Path(temporary) / 'stage'
        (stage / 'bin').mkdir(parents=True)
        shutil.copyfile(executable, stage / 'bin/wasmvm')
        (stage / 'bin/wasmvm').chmod(0o755)
        shutil.copyfile(ROOT / 'LICENSE', stage / 'LICENSE')
        licenses = stage / 'licenses'
        licenses.mkdir()
        goroot = Path(capture(['go', 'env', 'GOROOT']))
        go_license = next((p for p in [goroot / 'LICENSE', goroot.parent / 'LICENSE'] if p.is_file()), None)
        if go_license is None:
            raise ValueError('missing Go distribution license')
        shutil.copyfile(go_license, licenses / 'Go-LICENSE')
        # Retain additional notices for actually imported vendored/generated
        # standard-library packages, alongside the Go distribution license.
        notices = set()
        for package in packages:
            if not package.get('Standard'):
                continue
            directory = Path(package['Dir'])
            while directory.is_relative_to(goroot) and directory != goroot:
                for filename in ['LICENSE', 'NOTICE']:
                    if (directory / filename).is_file():
                        notices.add(directory / filename)
                directory = directory.parent
        for path in sorted(notices):
            shutil.copyfile(path, licenses / ('Go-' + path.relative_to(goroot).as_posix().replace('/', '_')))
        dependencies = []
        for module in modules:
            if module.get('Main'):
                continue
            directory = Path(module['Dir'])
            dependency = {k: module[k] for k in ['Path', 'Version', 'Sum']}
            dependency['licenseFiles'] = []
            for filename in ['LICENSE', 'PATENTS', 'AUTHORS', 'CONTRIBUTORS', 'NOTICE']:
                source = directory / filename
                if source.is_file():
                    target = licenses / (module['Path'].replace('/', '_') + '-' + filename)
                    shutil.copyfile(source, target)
                    dependency['licenseFiles'].append(target.relative_to(stage).as_posix())
            if not any(p.endswith('-LICENSE') for p in dependency['licenseFiles']):
                raise ValueError('missing dependency license: ' + module['Path'])
            dependencies.append(dependency)
        if args.compiler == 'llgo':
            llgo_root = Path(capture([compiler, 'env', 'LLGO_ROOT']))
            for path in sorted(llgo_root.rglob('LICENSE')):
                target = licenses / ('LLGo-' + path.relative_to(llgo_root).as_posix().replace('/', '_'))
                shutil.copyfile(path, target)
        manifest = dict(formatVersion=1, kind='CI candidate; performance/release acceptance remains open',
                        sourceRevision=revision, sourceTree=tree, goos=goos, goarch=goarch, native=args.native,
                        compiler=args.compiler, compilerVersion=compiler_version, compilerSHA256=digest(Path(compiler)),
                        compilerRevision=LLGO_REVISION if args.compiler == 'llgo' else GO_VERSION,
                        buildFlags=flags, buildEnvironment={'CGO_ENABLED': '1' if args.native == 'available' else '0'},
                        binarySHA256=digest(executable), buildHost=platform.platform(), goVersion=go_version, binaryGoBuildInfo=build_info,
                        dynamicLibraries=links, glibcSymbolFloor=glibc, dependencies=dependencies,
                        licenseSHA256={p.relative_to(stage).as_posix(): digest(p) for p in [stage / 'LICENSE', *licenses.iterdir()]})
        (stage / 'loader.txt').write_text(loader + '\n')
        (stage / 'README.txt').write_text('wasmvm CI candidate\n\nSource: ' + revision + '\nTree: ' + tree + '\n\nRun bin/wasmvm --help. This archive is experimental and unversioned; performance and release acceptance remain open.\n\nNative backend: ' + args.native + '\nDynamic runtime libraries (not bundled):\n' + links + '\n\nBuild-host information is provenance, not an OS compatibility promise. See loader.txt and manifest.json for actual loader/version requirements. Install the listed external libraries before running. LLVM/LLGo compilers are build tools, not bundled runtime dependencies.\n\nThe project uses MIT. Go, module and LLGo notices retain their own licenses in licenses/. No competitor engine, conformance corpus or third-party SDK is bundled.\n')
        write_json(stage / 'manifest.json', manifest)
        archive_folder(stage, archive_path, int(capture(['git', 'show', '-s', '--format=%ct', 'HEAD'])))
        extracted = Path(temporary) / 'extracted'
        with tarfile.open(archive_path) as archive:
            archive.extractall(extracted, filter='data')
        assert digest(extracted / 'bin/wasmvm') == manifest['binarySHA256']
        assert os.access(extracted / 'bin/wasmvm', os.X_OK)
        for relative, sha in manifest['licenseSHA256'].items():
            assert digest(extracted / relative) == sha
        report = output / (name + '.verification.json')
        subprocess.run([sys.executable, str(ROOT / 'scripts/cli_smoke.py'), str(extracted / 'bin/wasmvm'),
                        '--native', args.native, '--report', str(report)], cwd=ROOT, check=True, timeout=300)
        verified = json.loads(report.read_text())
        assert verified['passed'] and verified['executableSHA256'] == manifest['binarySHA256']
        verified.update(sourceRevision=revision, sourceTree=tree, archiveSHA256=digest(archive_path), archive=archive_path.name)
        write_json(report, verified)
        write_json(output / (name + '.manifest.json'), manifest)
    with (output / 'SHA256SUMS').open('a') as sums:
        for path in [archive_path, report, output / (name + '.manifest.json')]:
            sums.write(digest(path) + '  ' + path.name + '\n')
    print('Verified extracted candidate: ' + archive_path.name)


if __name__ == '__main__':
    main()

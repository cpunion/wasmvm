"""Temporary Go layout ablation; no diagnostic linker flag is promoted."""
import argparse
import copy
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tarfile

ROOT = Path.cwd()
OUT = ROOT / 'regression-profile'
OUT.mkdir()
sys.path.insert(0, str(ROOT / 'control/benchmarks'))
import own_build
import routing_compare

assert os.environ['REGRESSION_COMPILER'] == 'go'
revision = os.environ['REGRESSION_CANDIDATE_REVISION']
assert revision == '1f659b690dd13e0bfc3918594e148511f7c28c9f'
with tarfile.open('routing-build.tar.gz') as archive:
    archive.extractall(ROOT, filter='data')
data = json.loads((ROOT / 'routing-build/build.json').read_text())
os.environ['LD_LIBRARY_PATH'] = str(ROOT / 'routing-build/host-libraries')
records = {}
for role in ['predecessor', 'candidate']:
    destination = ROOT if role == 'candidate' else ROOT / 'predecessor'
    with tarfile.open(ROOT / 'downloads' / role / 'poll-go.tar.gz') as archive:
        archive.extractall(destination, filter='data')
    folder = destination / 'routing-build'
    fragment = json.loads((folder / 'poll-go.json').read_text())
    build = fragment['builds']['poll-go']
    binary = folder / 'poll-go/runner'
    assert own_build.digest(binary) == build['binarySha256']
    assert own_build.hashes(binary.parent, own_build.ENGINE_PATHS, own_build.HARNESS) == build['sourceHashes']
    records[role] = (fragment, dict(build, binary=str(binary)))
base, head = records['predecessor'], records['candidate']
assert base[0]['candidateRevision'] == '688b40b67c9ce27607a52be2d3e394e2000c2e77'
assert head[0]['candidateRevision'] == revision
go = shutil.which('go')
assert own_build.digest(Path(go)) == head[1]['toolchain']['binarySha256']
assert base[1]['toolchain']['binarySha256'] == head[1]['toolchain']['binarySha256']
source = Path(head[1]['binary']).parent
shifted = source / 'runner-shift32'
readelf = shutil.which('llvm-readelf-22') or shutil.which('llvm-readelf')
assert readelf
sections = subprocess.check_output([readelf, '--sections', '--wide', head[1]['binary']], text=True)
text_address = int(re.search(r'\.text\s+PROGBITS\s+([0-9a-f]+)', sections)[1], 16)
script = OUT / 'text-placement.ld'
script.write_text('SECTIONS { .text ' + hex(text_address + 32) + ' : SUBALIGN(32) { *(.text .text.*) } } INSERT AFTER .plt;\n')
flag = '-ldflags=-extldflags=-Wl,-T,' + str(script)
command = [go, 'build', flag, '-o', str(shifted), './benchmarks/embedded/wasmvm']
environment = dict(os.environ, CGO_ENABLED='1', GOWORK='off', GOTOOLCHAIN='local', GOFLAGS='', GOEXPERIMENT='')
assert not environment.get('LDFLAGS')
subprocess.run(command, cwd=source, env=environment, check=True, timeout=180)
assert own_build.hashes(source, own_build.ENGINE_PATHS, own_build.HARNESS) == head[1]['sourceHashes']
symbol = 'github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop'
objdump = shutil.which('llvm-objdump-22') or shutil.which('llvm-objdump')
assert objdump
static = {}
normalized = {}
for role, binary in [('base', Path(base[1]['binary'])), ('head', Path(head[1]['binary'])), ('shift', shifted)]:
    asm = subprocess.check_output([objdump, '--disassemble-symbols=' + symbol, '--no-show-raw-insn', '--x86-asm-syntax=intel', str(binary)], text=True)
    (OUT / (role + '.asm')).write_text(asm)
    rows = re.findall(r'^\s*([0-9a-f]+):\s+(.*)$', asm, re.M)
    assert rows
    static[role] = dict(entry=int(rows[0][0], 16), binarySha256=own_build.digest(binary), instructions=len(rows))
    normalized[role] = [re.sub(r'# 0x[0-9a-f]+', '# <reloc>', re.sub(r'\[rip ([+-]) 0x[0-9a-f]+\]', r'[rip \1 <reloc>]', re.sub(r'0x[0-9a-f]+(?= <)', '<reloc>', text))) for _, text in rows]
    shutil.copyfile(binary, OUT / (role + '-runner'))
(OUT / 'placement-precondition.json').write_text(json.dumps(dict(flag=flag, static=static), indent=2) + '\n')
assert static['shift']['entry'] - static['head']['entry'] == 32, static
assert normalized['shift'] == normalized['head'], 'placement changed loop instructions/registers/control'
derived = copy.deepcopy(head[1])
derived.update(binary=str(shifted), binarySha256=own_build.digest(shifted), buildCommand=command)
derived['toolchain']['buildFlags'] = [flag]
data['builds'] = {'base-go': base[1], 'head-go': head[1], 'shift-go': derived, 'native-c': data['builds']['native-c']}
runtimes = {name: data['runtimes'][name] for name in ['wazero-interpreter', 'wasmedge-interpreter']}
for role, fragment, binary in [('base', base[0], base[1]['binary']), ('head', head[0], head[1]['binary']), ('shift', head[0], str(shifted))]:
    for backend in (['interpreter'] if role == 'shift' else ['interpreter', 'native']):
        runtimes[role + '-go-' + backend] = dict(fragment['runtimes']['poll-go-' + backend], command=[binary, '--mode', 'api'])
for name, row in list(runtimes.items()):
    runtimes['aa-' + name] = dict(row)
data['sourceArtifact'] = dict(run=43448812030, base=data['base'], head=data['head'], linkage=data.pop('linkage', None), scope='Historical SDK/reference identities; retired owned binaries are not measured.')
data.update(base=base[0]['candidateRevision'], head=revision, runtimes=runtimes,
            goLayoutDiagnostic=dict(flag=flag, static=static, loopInstructionsEqual=True, expectedObservations=300, alternateChecks=14, controls=7,
                scope='Identical candidate sources; explicit text placement and32-byte input-section alignment are diagnostic, never production. Loop instructions/registers/control must match; all variants/references retained.'))
(ROOT / 'routing-build/build.json').write_text(json.dumps(data, indent=2) + '\n')
(OUT / 'static-provenance.json').write_text(json.dumps(data['goLayoutDiagnostic'], indent=2) + '\n')
routing_compare.measure(argparse.Namespace(head_dir=ROOT / 'head', build_dir=ROOT / 'routing-build', output_dir=OUT / 'measurements',
                                         group='memory-simd', case_id='memory/dependent-address', blocks=20, gc_sustained=False, threads_sustained=False))

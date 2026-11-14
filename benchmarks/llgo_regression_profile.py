"""Profile frozen native LLGo regressions; instrumented counts are not timings."""
import hashlib
import json
import os
import re
from pathlib import Path
import subprocess
import tarfile

ROOT = Path.cwd()
OUT = ROOT / 'regression-profile'
OUT.mkdir()
BIN = ROOT / 'regression-input'
BIN.mkdir()
SYMBOL = 'github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop'
compiler = os.environ.get('REGRESSION_COMPILER', 'llgo')
assert compiler in ('llgo', 'go')
artifact_role = 'poll-' + compiler
ROLES = {'counter': ('5e5e6a44b2062d61bfa569f384232483763d3380', 43537347012),
         'dense': ('26b788258ac6c280639001ecda0af28159cf8068', 43543803052),
         'predecessor': ('688b40b67c9ce27607a52be2d3e394e2000c2e77', 43561639936)}
if compiler == 'go':
    assert os.environ.get('REGRESSION_CANDIDATE_RUN')
    ROLES = {'counter': ROLES['counter'],
             'predecessor': ('688b40b67c9ce27607a52be2d3e394e2000c2e77', 43561639936)}
    # Avoid asynchronous scheduler signals perturbing exact instruction counts.
    # This is a count-only diagnostic setting, never production timing policy.
    os.environ['GODEBUG'] = 'asyncpreemptoff=1'
if candidate_run := os.environ.get('REGRESSION_CANDIDATE_RUN'):
    revision = os.environ['REGRESSION_CANDIDATE_REVISION']
    assert re.fullmatch(r'[0-9a-f]{40}', revision)
    assert candidate_run.isdigit()
    ROLES.pop('dense', None)
    ROLES['candidate'] = revision, int(candidate_run)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def save(name, value):
    (OUT / name).write_text(json.dumps(value, indent=2) + '\n')


# Retain only exact measured executables/manifests and required host libraries.
with tarfile.open('routing-build.tar.gz') as archive:
    members = [m for m in archive.getmembers() if m.name.startswith('routing-build/host-libraries/')]
    archive.extractall(ROOT, members=members, filter='data')
os.environ['LD_LIBRARY_PATH'] = str(ROOT / 'routing-build/host-libraries')
os.environ['GOMAXPROCS'] = '1'
os.sched_setaffinity(0, {min(os.sched_getaffinity(0))})
builds = {}
for role, (revision, run) in ROLES.items():
    with tarfile.open(ROOT / 'downloads' / role / (artifact_role + '.tar.gz')) as archive:
        names = archive.getnames()
        manifest = json.load(archive.extractfile(next(n for n in names if n.endswith('/' + artifact_role + '.json'))))
        row = manifest['builds'][artifact_role]
        assert manifest['candidateRevision'] == revision
        binary = BIN / (role + '-runner')
        binary.write_bytes(archive.extractfile(next(n for n in names if n.endswith('/' + artifact_role + '/runner'))).read())
        binary.chmod(0o755)
        assert sha(binary) == row['binarySha256']
        builds[role] = {'source': revision, 'buildRun': run, 'binary': str(binary),
                        'binarySha256': sha(binary), 'toolchain': row['toolchain']}
save('provenance.json', builds)

# The fixture and oracle are unchanged from the original frozen comparison.
import sys
sys.path.insert(0, str(ROOT / 'head/benchmarks/native'))
from memory_fixtures import checksum
count, seed = 100000, 0x12345678
expected = checksum('dependent-address', seed, count)
fixture = ROOT / 'head/benchmarks/native/memory-fixtures/dependent-address.wasm'
plan = {'iterations': count, 'seed': seed, 'expected': expected, 'samples': 1, 'warmups': 0,
        'compiler': compiler, 'GODEBUG': os.environ.get('GODEBUG', ''),
        'fixtureSha256': sha(fixture), 'roles': [*ROLES, 'counter-aa'],
        'cpu': subprocess.check_output(['lscpu'], text=True), 'affinity': list(os.sched_getaffinity(0)),
        'valgrind': subprocess.check_output(['valgrind', '--version'], text=True).strip(),
        'collection': SYMBOL,
        'scope': 'Frozen native x86-64 host processes under Callgrind. Actual dynamic instruction/memory-operation counts, simulated cache/branch counters; never wall-time acceptance. Collection toggles only in the shared portable loop, covering first and one measured invocation. Identical-binary counter alias checks count reproducibility. Normal GC, checks and polling retained.'}
save('plan.json', plan)
rows = []
for label in plan['roles']:
    role = 'counter' if label == 'counter-aa' else label
    target = OUT / (label + '.callgrind')
    command = ['valgrind', '--tool=callgrind', '--collect-atstart=no', '--toggle-collect=' + SYMBOL,
               '--dump-instr=yes', '--cache-sim=yes', '--branch-sim=yes', '--callgrind-out-file=' + str(target),
               builds[role]['binary'], '--mode', 'api', '--backend', 'interpreter', '--wasm', str(fixture),
               '--iterations', str(count), '--seed', str(seed), '--expected', str(expected), '--samples', '1', '--warmups', '0']
    row = {'role': label, 'command': command}
    try:
        result = subprocess.run(command, input='run\n', capture_output=True, text=True, timeout=100)
        (OUT / (label + '.stdout')).write_text(result.stdout)
        (OUT / (label + '.stderr')).write_text(result.stderr)
        assert result.returncode == 0, result.stderr[-3000:]
        record = json.loads(result.stdout.splitlines()[-1])
        assert record['sourceRevision'] == ROLES[role][0] and record['result'] == expected
        assert record['warmResults'] == [expected] and record['warmupResults'] == []
        lines = target.read_text().splitlines()
        events = next(s for s in lines if s.startswith('events:')).split()[1:]
        totals = [int(v) for v in next(s for s in lines if s.startswith('summary:')).split()[1:]]
        row['events'] = dict(zip(events, totals))
        assert row['events']['Ir'] > 1000000, 'loop symbol collection did not start'
        row['record'] = record
        row['status'] = 'complete'
    except Exception as error:
        row.update(status='failed', error=str(error))
        rows.append(row)
        save('results.json', rows)
        raise
    rows.append(row)
    save('results.json', rows)
    print(label, json.dumps(row['events']), flush=True)
assert rows[0]['events']['Ir'] == rows[-1]['events']['Ir'], 'identical-binary instruction counts differ'

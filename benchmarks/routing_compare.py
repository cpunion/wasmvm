"""Opt-in hosted source comparison; independent groups reuse one pinned build.

Affinity and A/A calibration improve diagnostics, not the physical isolation of
a hosted VM. These results never close controlled milestone acceptance.
"""
import argparse
import importlib.util
import json
import math
import os
from pathlib import Path
import platform
import random
import re
import shutil
import statistics
import subprocess
import sys
import time

import ci
import interleaved
import own_build
import phase_ci
import survey
import survey_modern
from applications.fixtures import load as applications
from applications.oracle import checksum as application_oracle

POLICY = survey_modern.POLICY
BLOCKS = 5
SEED = 2026111019
GROUPS = {
    'core': ('core/call-control', 'core/integer-loop', 'core/function-calls', 'core/memory-load-store'),
    'memory-simd': ('memory/narrow-unaligned', 'memory/dependent-address', 'memory/float64-unaligned',
                    'runtime/simd-memory-full', 'runtime/simd-memory-widen', 'runtime/simd-memory-lanes'),
    'gc': ('gc/i31-roundtrip', 'gc/struct-access', 'gc/struct-allocation', 'gc/struct-indirect-access'),
    'threads': ('threads/private-memory', 'threads/shared-memory', 'threads/atomic-rmw', 'threads/notify-empty',
                'threads/wait-mismatch', 'threads/wait-timeout'),
    'applications': ('applications/sha256-clang', 'applications/sha256-rust'),
}


def save(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def instrument_own(source):
    marker = 'config.Backend = args.backend'
    if source.count(marker) != 1 or source.count('unmetered; bounded cancellation; checked Core 2.0') != 1:
        raise ValueError('unrecognized common phase instrumentation')
    return source.replace(marker, marker + '\n config.Features = wasm.SupportedFeatures', 1).replace(
        'unmetered; bounded cancellation; checked Core 2.0', POLICY)


def reference_source(name, source):
    if name == 'wazero':
        return source.replace('"github.com/tetratelabs/wazero/api"',
                              '"github.com/tetratelabs/wazero/api"\n"github.com/tetratelabs/wazero/experimental"').replace(
            'runtime := wazero.NewRuntimeWithConfig(ctx, config)',
            'config = config.WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads)\n runtime := wazero.NewRuntimeWithConfig(ctx, config)')
    patches = {
        'wasmtime': ('wasmtime_config_strategy_set(config, WASMTIME_STRATEGY_CRANELIFT);',
                     '\n wasmtime_config_wasm_threads_set(config,true);\n wasmtime_config_shared_memory_set(config,true);'),
        'wasmer': ('wasm_config_set_backend(config, CRANELIFT);',
                   '\n wasmer_features_t *features=wasmer_features_new();\n if(!wasmer_features_threads(features,true))last_error("enable threads");\n wasm_config_set_features(config,features);'),
        'wasmedge': ('WasmEdge_ConfigureSetRunMode(config, WasmEdge_RunMode_Interpreter);',
                     '\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_Threads);\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_FunctionReferences);\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_GC);'),
    }
    marker, addition = patches[name]
    if source.count(marker) != 1:
        raise ValueError('unrecognized reference feature boundary')
    return source.replace(marker, marker + addition)


def fixtures(root):
    wanted = {name for names in GROUPS.values() for name in names}
    result = [f for f in survey.fixtures(root) if f['id'] in wanted]
    modern = json.loads((root / 'benchmarks/modern/manifest.json').read_text())
    for f in modern['workloads']:
        for ext in ['wasm', 'wat']:
            if own_build.digest(root / 'benchmarks/modern' / f[ext]) != f[ext + 'Sha256']:
                raise ValueError('modern fixture identity mismatch')
        result.append(dict(f, path=str(root / 'benchmarks/modern' / f['wasm']), unit='ns/source iteration'))
    result.extend(applications(root / 'benchmarks/applications'))
    if len(result) != len(wanted) or {f['id'] for f in result} != wanted:
        raise ValueError('routing portfolio has missing or duplicate cases')
    for f in result:
        if f['family'] not in ['gc', 'threads', 'application']:
            f['iterations'] = 1 if f['id'] == 'core/call-control' else 1_000_000
        f['expectedChecksum'] = oracle(root, f, f['seed'], f['iterations'])
    return result


def oracle(root, fixture, seed, count):
    family, kind = fixture['id'].split('/')
    if family == 'applications':
        return application_oracle(seed, count)
    if family == 'threads':
        return (seed + count * (0 if kind == 'notify-empty' else 2 if kind == 'wait-timeout' else 1)) & 0xffffffff
    if family == 'gc':
        x, objects = seed, [seed, seed ^ 0xdeadbeef]
        for i in range(count):
            if kind == 'struct-indirect-access':
                objects[x & 1] = (x * 1664525 + 1013904223) & 0xffffffff
                x = objects[(x >> 1) & 1] ^ i
            else:
                x = (x * 1664525 + 1013904223) & (0x7fffffff if kind == 'i31-roundtrip' else 0xffffffff)
        return x
    if family == 'core':
        spec = importlib.util.spec_from_file_location('core_oracle', root / 'benchmarks/embedded/generate.py')
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module.checksum('integer-mix' if kind == 'integer-loop' else kind, seed, count)
    sys.path.insert(0, str(root / 'benchmarks/native'))
    from memory_fixtures import checksum as memory_checksum
    from runtime_fixtures import checksum as vector_checksum
    return (memory_checksum if family == 'memory' else vector_checksum)(kind, seed, count)


def compile_runner(command, directory, environment):
    start = time.monotonic()
    subprocess.run(command, cwd=directory, env=environment, check=True, timeout=300)
    binary = Path(command[command.index('-o') + 1])
    return dict(command=command, binary=str(binary), binarySha256=own_build.digest(binary),
                elapsedSeconds=time.monotonic() - start)


def linked_paths(text):
    if 'not found' in text:
        raise ValueError('a runtime shared library is missing')
    return [(name, Path(path)) for name, path in re.findall(r'^\s*(\S+)\s+=>\s+(/\S+)', text, re.M)]


def bundle_host_libraries(output, builds):
    if platform.system() != 'Linux':
        return dict(status='not-needed-for-local-dry-build')
    target = output / 'host-libraries'
    target.mkdir()
    result = {}
    system = ('libc.so', 'libm.so', 'libpthread.so', 'libdl.so', 'librt.so', 'libresolv.so',
              'libstdc++.so', 'libgcc_s.so', 'libutil.so', 'libnsl.so')
    for name, row in builds.items():
        probe = subprocess.run(['ldd', row['binary']], capture_output=True, text=True, check=False)
        text = probe.stdout + probe.stderr
        if probe.returncode:
            if probe.returncode == 1 and ('not a dynamic executable' in text or 'statically linked' in text):
                result[name] = dict(ldd=text, copied=[], status='static executable')
                continue
            raise ValueError('unable to inspect runtime linkage: ' + name)
        result[name] = dict(ldd=text, copied=[])
        for soname, source in linked_paths(text):
            if soname.startswith(system) or source.resolve().is_relative_to(output):
                continue
            destination = target / soname
            if destination.exists() and own_build.digest(destination) != own_build.digest(source):
                raise ValueError('conflicting runtime library: ' + soname)
            if not destination.exists():
                shutil.copyfile(source, destination)
            result[name]['copied'].append(dict(source=str(source), destination=str(destination), sha256=own_build.digest(destination)))
    return result


def build(args):
    head, base, out, tools = [p.resolve() for p in [args.head_dir, args.base_dir, args.output_dir, args.tools_dir]]
    out.mkdir(parents=True, exist_ok=True)
    if (out / 'build.json').exists():
        raise ValueError('do not replace an existing build')
    selected = own_build.toolchains()
    env = dict(os.environ, CGO_ENABLED='1', GOWORK='off', GOTOOLCHAIN='local', GOFLAGS='', GOEXPERIMENT='')
    builds, runtimes = {}, {}
    for role, checkout in [('base', base), ('head', head)]:
        revision = ci.capture(['git', '-C', str(checkout), 'rev-parse', 'HEAD'])
        for compiler, toolchain in selected.items():
            name, target = role + '-' + compiler, out / (role + '-' + compiler)
            identity = own_build.prepare(head, checkout, revision, target)
            if identity['status'] != 'available':
                raise ValueError('required engine baseline is unavailable')
            phase = target / 'internal/bench/phase/phase.go'
            phase.write_text(instrument_own(phase.read_text()))
            (phase.parent / 'provenance_gen.go').write_text('package phase\nfunc init(){sourceRevision=' + json.dumps(revision) + ';compilerRevision=' + json.dumps(toolchain['revision']) + '}\n')
            record = compile_runner([toolchain['command'], 'build', *toolchain['buildFlags'], '-o', str(target / 'runner'), './benchmarks/embedded/wasmvm'], target, env)
            if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != identity['sourceHashes']:
                raise ValueError('instrumentation changed engine sources')
            builds[name] = dict(record, **identity, toolchain=toolchain, instrumentation=phase.read_text())
            for backend in ['interpreter', 'native']:
                runtimes[name + '-' + backend] = dict(command=[record['binary'], '--mode', 'api'], backend=backend,
                    engine='wasmvm', sourceRevision=revision, compiler='gc' if compiler == 'go' else 'llgo',
                    compilerRevision=toolchain['revision'], mode='api', goos=toolchain['goos'], goarch=toolchain['goarch'], policy=POLICY)
            print('built', name, flush=True)
    for role in ['base', 'head']:
        for backend in ['interpreter', 'native']:
            runtimes['aa-' + role + '-llgo-' + backend] = dict(runtimes[role + '-llgo-' + backend])
    # Temporary diagnostic: fixed text VMA isolates metadata-driven movement;
    # the standard separate-code linker flag is the production candidate.
    # All other source/compiler/optimization/policy settings remain unchanged.
    for profile, flag in [('layout', '-Wl,-Ttext=0x400000'), ('separate', '-Wl,-z,separate-code')]:
        for role in ['base', 'head']:
            original = builds[role + '-llgo']
            name = profile + '-' + role + '-llgo'
            target = Path(original['binary']).parent
            command = list(original['command'])
            command[command.index('-o') + 1] = str(target / ('runner-' + profile))
            command.insert(command.index('-o'), '-ldflags=-extldflags=' + flag)
            record = compile_runner(command, target, env)
            builds[name] = dict(original, **record)
            builds[name]['experimentalLinkerFlags'] = [flag]
            n = name + '-interpreter'
            runtimes[n] = dict(runtimes[role + '-llgo-interpreter'], command=[record['binary'], '--mode', 'api'])
            runtimes['aa-' + n] = dict(runtimes[n])
    layout_dir = out / 'layout-probe'
    layout_dir.mkdir()
    objdump = shutil.which('llvm-objdump-22') or shutil.which('llvm-objdump')
    nm = shutil.which('llvm-nm-22') or shutil.which('llvm-nm')
    if not objdump or not nm:
        raise ValueError('LLVM object inspection tools are required for the layout diagnostic')
    for name in ['base-llgo', 'head-llgo', 'layout-base-llgo', 'layout-head-llgo', 'separate-base-llgo', 'separate-head-llgo']:
        binary = builds[name]['binary']
        sections = ci.capture([objdump, '--section-headers', binary])
        match = re.search(r'^\s*\d+\s+\.text\s+[0-9a-fA-F]+\s+([0-9a-fA-F]+)', sections, re.M)
        if not match:
            raise ValueError('missing ELF text section')
        text = int(match[1], 16)
        if name.startswith('layout-') and text != 0x400000:
            raise ValueError('diagnostic text VMA was not applied')
        if name.startswith('separate-') and text % 4096 != 0:
            raise ValueError('separate-code did not page-align the text section')
        symbol = 'github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlots'
        row = next(line.split() for line in ci.capture([nm, '--print-size', binary]).splitlines() if line.endswith(' ' + symbol))
        address, size = int(row[0], 16), int(row[1], 16)
        assembly = ci.capture([objdump, '--disassemble-symbols=' + symbol, '--no-show-raw-insn', '--no-leading-addr', '--symbolize-operands', binary])
        (layout_dir / (name + '.asm')).write_text(assembly)
        builds[name]['layoutDiagnostic'] = dict(text=text, dispatcher=address, size=size, textOffset=address-text, mod64=address%64, sections=sections)

    revision = ci.capture(['git', '-C', str(head), 'rev-parse', 'HEAD'])
    target = out / 'wazero'
    own_build.archive(head, revision, target, ('benchmarks/embedded/wazero',))
    target = target / 'benchmarks/embedded/wazero'
    path = target / 'main.go'
    path.write_text(reference_source('wazero', path.read_text()))
    builds['wazero'] = compile_runner([selected['go']['command'], 'build', '-trimpath', '-o', str(target / 'runner'), '.'], target, env)
    builds['wazero']['dependency'] = json.loads(ci.capture(['go', 'list', '-m', '-json', ci.MODULE], cwd=target))
    if builds['wazero']['dependency']['Version'] != 'v1.12.0' or builds['wazero']['dependency'].get('Replace'):
        raise ValueError('wazero reference pin differs')
    for backend in ['interpreter', 'compiler']:
        runtimes['wazero-' + backend] = dict(command=[builds['wazero']['binary']], backend=backend)
    source_dir = out / 'reference-source'
    source_dir.mkdir()
    shutil.copyfile(head / 'benchmarks/embedded/c/common.h', source_dir / 'common.h')
    cc = shutil.which('clang-22') or shutil.which('clang')
    if not cc or '22.1.8' not in ci.capture([cc, '--version']):
        raise ValueError('pinned Clang 22.1.8 is required')
    for name in ['wasmtime', 'wasmer', 'wasmedge']:
        suffix = '.dylib' if platform.system() == 'Darwin' else '.so'
        libraries = sorted(tools.rglob('lib' + name + suffix))
        if len(libraries) != 1:
            raise ValueError('missing or ambiguous reference library: ' + name)
        library = libraries[0]
        sdk = library.parent.parent
        path = source_dir / (name + '.c')
        path.write_text(reference_source(name, (head / 'benchmarks/embedded/c' / path.name).read_text()))
        command = [cc, '-std=c11', '-D_POSIX_C_SOURCE=200809L', '-O2', '-Wall', '-Wextra', '-Werror', '-I', str(sdk / 'include'), str(path), '-o', str(out / (name + '-runner')), '-L', str(library.parent), '-Wl,-rpath,' + str(library.parent), '-l' + name]
        builds[name] = dict(compile_runner(command, head, env), sourceSha256=own_build.digest(path), sdkLibrary=str(library), sdkLibrarySha256=own_build.digest(library))
        runtimes[name + ('-interpreter' if name == 'wasmedge' else '-cranelift')] = dict(command=[builds[name]['binary']], backend='interpreter' if name == 'wasmedge' else 'compiler')
    for name, source in [('native-c', head / 'benchmarks/native/c/scalar_baseline.c'), ('native-application', head / 'benchmarks/applications/sha256.c')]:
        builds[name] = dict(compile_runner([cc, '-O3', '-std=c11', '-D_POSIX_C_SOURCE=200809L', '-Wall', '-Wextra', '-Werror', str(source), '-o', str(out / name)], head, env), sourceSha256=own_build.digest(source))
    linkage = bundle_host_libraries(out, builds)
    save(out / 'build.json', dict(schemaVersion=1, head=builds['head-go']['sourceRevision'], base=builds['base-go']['sourceRevision'], builds=builds, runtimes=runtimes, fixtures=fixtures(head), protocol=dict(blocks=BLOCKS, samples=7, warmups=2, seed=SEED), installation=json.loads((tools / 'installation.json').read_text()), clang=ci.capture([cc, '--version']), linkage=linkage))


def warm_values(case, statistic=None):
    """Reduce each fresh process separately; preserve pauses in allocation calls."""
    if statistic is None:
        statistic = statistics.mean if case.get('warmStatistic') == 'mean-including-pauses-v1' else statistics.median
    rows = case['observations'] + case.get('nativeC', [])
    return {name: statistics.median(statistic(row['record']['warmCallNs'])
                                   for row in rows if row['runtime'] == name)
            for name in {row['runtime'] for row in rows}}


def calibration(case, runtimes, statistic=None, blocks=BLOCKS):
    if statistic is None:
        statistic = statistics.mean if case.get('warmStatistic') == 'mean-including-pauses-v1' else statistics.median
    result = []
    for name in runtimes:
        if not name.startswith('aa-'):
            continue
        rates = {}
        for row in case['observations']:
            if row['runtime'] in [name, name[3:]]:
                rates[row['runtime'], row['block']] = statistic(row['record']['warmCallNs'])
        if any(rates[name, i] <= 0 or rates[name[3:], i] <= 0 for i in range(blocks)):
            result.append(dict(runtime=name, resolved=False, reason='unresolved timer control'))
            continue
        logs = [math.log(rates[name, i] / rates[name[3:], i]) for i in range(blocks)]
        rng = random.Random(SEED)
        draws = sorted(math.exp(statistics.mean(rng.choices(logs, k=blocks))) for _ in range(8000))
        lower, upper = draws[200], draws[-201]
        result.append(dict(runtime=name, lower=lower, upper=upper, width=upper-lower,
                           resolved=0.975 <= lower <= upper <= 1.025))
    return result



def measurement_fixture(root, fixture, sustained_gc=False, sustained_threads=False):
    """Keep original counts separate from explicitly named sustained cohorts."""
    result = dict(fixture)
    if sustained_gc and sustained_threads:
        raise ValueError("select only one sustained cohort")
    if sustained_gc:
        if result['family'] != 'gc':
            raise ValueError('sustained GC requires a GC fixture')
        result['iterations'] = 300_000 if result['id'] == 'gc/struct-allocation' else 1_000_000
        result['expectedChecksum'] = oracle(root, result, result['seed'], result['iterations'])
    if sustained_threads:
        if result['family'] != 'threads':
            raise ValueError('sustained Threads requires a Threads fixture')
        result['iterations'] = 100_000
        result['expectedChecksum'] = oracle(root, result, result['seed'], result['iterations'])
    return result


def measure(args):
    if not 2 <= args.blocks <= 60 or args.case_id and args.case_id not in GROUPS[args.group]:
        raise ValueError('select 2–60 fixed blocks and a case within the selected group')
    sustained_gc = getattr(args, 'gc_sustained', False)
    sustained_threads = getattr(args, 'threads_sustained', False)
    if sustained_threads and (args.group != 'threads' or sustained_gc):
        raise ValueError('--threads-sustained requires only --group threads')
    if sustained_gc and args.group != 'gc':
        raise ValueError('--gc-sustained requires --group gc')
    root, folder, out = [p.resolve() for p in [args.head_dir, args.build_dir, args.output_dir]]
    out.mkdir(parents=True, exist_ok=True)
    if (out / 'results.json').exists():
        raise ValueError('do not replace measured attempts')
    b = json.loads((folder / 'build.json').read_text())
    if args.case_id:
        for role in ['base', 'head']:
            name = role + '-go-native'
            b['runtimes']['aa-' + name] = dict(b['runtimes'][name])
    if b['head'] != ci.capture(['git', '-C', str(root), 'rev-parse', 'HEAD']):
        raise ValueError('candidate source differs from the built artifact')
    if not hasattr(os, 'sched_setaffinity'):
        raise ValueError('this hosted pilot requires Linux affinity')
    cpu = min(os.sched_getaffinity(0))
    os.sched_setaffinity(0, {cpu})
    os.environ['GOMAXPROCS'] = '1'
    for row in b['builds'].values():
        if own_build.digest(Path(row['binary'])) != row['binarySha256']:
            raise ValueError('built binary identity changed')
        if 'sourceHashes' in row and own_build.hashes(Path(row['binary']).parent, own_build.ENGINE_PATHS, own_build.HARNESS) != row['sourceHashes']:
            raise ValueError('archived engine identity changed')
    for path, digest in b.get('nativeRuntime', {}).get('libraryHashes', {}).items():
        if own_build.digest(Path(path)) != digest:
            raise ValueError('native runtime library identity changed')
    report = dict(status='running', classification='Hosted advisory; controlled milestone acceptance remains open', build=b, group=args.group, cohort='gc-sustained-v1' if sustained_gc else 'threads-sustained-v2' if sustained_threads else 'original22-v1', host=dict(platform=platform.platform(), cpu=ci.capture(['lscpu']), affinity=sorted(os.sched_getaffinity(0)), GOMAXPROCS='1', loadBefore=os.getloadavg(), powerPolicy='Hosted VM default; physical frequency/neighbor isolation not controlled', cachePolicy='Fresh processes; disk compilation cache disabled; filesystem cache not flushed'), cases=[], checks=[], unsupported=[])
    report['protocol'] = dict(blocks=args.blocks, samples=7, warmups=2, seed=SEED, case=args.case_id,
                              controls=[name for name in b['runtimes'] if name.startswith('aa-')])
    if args.case_id:
        report['cohort'] = 'focused-case-v1'
    save(out / 'plan.json', report['protocol'])
    rng = random.Random(SEED)
    try:
        for f in b['fixtures']:
            if f['id'] not in GROUPS[args.group]:
                continue
            if args.case_id and f['id'] != args.case_id:
                continue
            f = measurement_fixture(root, f, sustained_gc, sustained_threads)
            if own_build.digest(Path(f['path'])) != f['wasmSha256']:
                raise ValueError('fixture identity changed')
            commands = {n: r['command'] + ['--backend', r['backend'], '--wasm', f['path'], '--iterations', str(f['iterations']), '--seed', str(f['seed']), '--expected', str(f['expectedChecksum']), '--samples', '7', '--warmups', '2'] for n, r in b['runtimes'].items()}
            if f['family'] == 'gc':
                for name in ['wazero-compiler', 'wazero-interpreter', 'wasmer-cranelift']:
                    command = commands.pop(name)
                    try:
                        phase_ci.invoke(command, 120)
                    except RuntimeError as error:
                        if not survey_modern.expected_gc_rejection(name, error):
                            raise
                        report['unsupported'].append(dict(runtime=name, workload=f['id'], error=str(error)))
                    else:
                        raise ValueError('reference capability changed; update matrix explicitly')
            for name, command in commands.items():
                alt = list(command)
                count = 1 if f['id'] == 'core/call-control' else 17
                expected = oracle(root, f, 0xffffffff, count)
                for flag, value in [('--iterations', count), ('--seed', 0xffffffff), ('--expected', expected)]:
                    alt[alt.index(flag)+1] = str(value)
                row = dict(phase_ci.invoke(alt, 120), runtime=name, workload=f['id'], wasmSha256=f['wasmSha256'])
                report['checks'].append(row)
                phase_ci.validate_observation(row, b['runtimes'][name], dict(f, seed=0xffffffff, iterations=count, expectedChecksum=expected))
            kind = f['id'].split('/')[1]
            if f['family'] == 'application':
                commands['native-c'] = [b['builds']['native-application']['binary'], str(f['seed']), str(f['iterations']), str(f['expectedChecksum'])]
            elif kind in ['integer-loop', 'function-calls', 'memory-load-store', 'narrow-unaligned', 'dependent-address', 'float64-unaligned', 'simd-memory-full', 'simd-memory-widen', 'simd-memory-lanes']:
                commands['native-c'] = [b['builds']['native-c']['binary'], kind, str(f['seed']), str(f['iterations']), str(f['expectedChecksum'])]
            if 'native-c' in commands:
                commands['native-c'] = b.get('nativeRuntime', {}).get('prefix', []) + commands['native-c']
            case = dict(fixture=f, observations=[], nativeC=[], sampleOrders=[], warmStatistic='mean-including-pauses-v1' if f['id'] == 'gc/struct-allocation' else 'median-v1')
            report['cases'].append(case)
            for block in range(args.blocks):
                rows, orders = interleaved.invoke(commands, 180, rng)
                case['sampleOrders'].append(orders)
                for name, row in rows.items():
                    row = dict(row, runtime=name, block=block, wasmSha256=f['wasmSha256'])
                    if name == 'native-c':
                        if row['exitCode'] or row['record']['checksum'] != f['expectedChecksum'] or row['record']['iterations'] != f['iterations'] or len(row['record']['warmCallNs']) != 7:
                            raise ValueError('native C result/protocol differs')
                        case['nativeC'].append(row)
                    else:
                        phase_ci.validate_observation(row, b['runtimes'][name], f)
                        case['observations'].append(row)
                save(out / 'results.json', report)
            case['warmStatistic'] = 'mean-including-pauses-v1' if f['id'] == 'gc/struct-allocation' else 'median-v1'
            case['calibration'] = calibration(case, b['runtimes'], blocks=args.blocks)
            if f['id'] == 'gc/struct-allocation':
                case['medianCalibration'] = calibration(case, b['runtimes'], statistics.median, args.blocks)
                case['warmMedianNs'] = warm_values(case, statistics.median)
                case['warmMeanNs'] = warm_values(case, statistics.mean)
                case['processWarmMeanNs'] = [dict(runtime=row['runtime'], block=row['block'],
                                                 meanNs=statistics.mean(row['record']['warmCallNs']))
                                             for row in case['observations']]
                save(out / 'results.json', report)
            print('measured', f['id'], flush=True)
        report['status'] = 'complete'
    except BaseException as error:
        report.update(status='failed', failure=repr(error))
        raise
    finally:
        report['host']['loadAfter'] = os.getloadavg()
        save(out / 'results.json', report)
        lines = ['## Runtime routing comparison: ' + args.group + (' / sustained GC (300000 allocation, 1000000 other iterations)' if sustained_gc else ' / sustained Threads (100000 iterations)' if sustained_threads else ' / focused ' + args.case_id if args.case_id else ' / original22'), '', f'**Hosted advisory; controlled milestone acceptance remains open.** {args.blocks} fresh blocks, two warmups/seven interleaved samples; pinned source/compiler/reference identities, one inherited CPU affinity and GOMAXPROCS=1. Physical CPU power and neighbor contention remain uncontrolled. Each marginal 95% paired A/A interval must lie within ±2.5% for this diagnostic calibration; this is not the controlled native feature equivalence gate.', '', 'Base `' + b['base'] + '`; head `' + b['head'] + '`.', '', '| Workload / statistic | Go native head/base | LLGo native head/base | Go interpreter head/base | LLGo interpreter head/base | LLGo native / Wasmtime | / Native C | A/A within ±2.5% |', '| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |']
        for case in report['cases']:
            values = warm_values(case)
            def ratio(a, c):
                return f'{values[a]/values[c]:.3f}×' if a in values and c in values else 'unavailable'
            cells = [ratio('head-' + co + '-' + be, 'base-' + co + '-' + be) for be in ['native', 'interpreter'] for co in ['go', 'llgo']]
            cells += [ratio('head-llgo-native', 'wasmtime-cranelift'), ratio('head-llgo-native', 'native-c')]
            resolved = bool(case.get('calibration')) and all(c['resolved'] for c in case['calibration']) and len(case['calibration']) == len(report['protocol']['controls'])
            lines.append('| ' + case['fixture']['id'] + (' / process mean' if case.get('warmStatistic') == 'mean-including-pauses-v1' else ' / process median') + ' | ' + ' | '.join(cells) + ' | ' + ('yes, advisory only' if resolved else 'inconclusive') + ' |')
            if case.get('warmStatistic') == 'mean-including-pauses-v1':
                values = warm_values(case, statistics.median)
                cells = [ratio('head-' + co + '-' + be, 'base-' + co + '-' + be) for be in ['native', 'interpreter'] for co in ['go', 'llgo']]
                cells += [ratio('head-llgo-native', 'wasmtime-cranelift'), ratio('head-llgo-native', 'native-c')]
                lines.append('| ' + case['fixture']['id'] + ' / median diagnostic | ' + ' | '.join(cells) + ' | separate median controls in JSON |')
        lines += ['', f'Allocation throughput uses the median of {args.blocks} fresh-process means, including every measured warm call and pause within each process. Other cases retain process medians. Allocation median diagnostics and both control reductions remain explicit; this statistic correction changes no counts, timers or samples.', '', 'Status: **' + report['status'] + '**. Complete phases, all runtimes including Wasmer/wazero/WasmEdge, unsupported GC probes, orders, controls and raw observations are retained. No timer/phase subtraction or cross-job pooling; smaller diagnostic components cannot replace complete API outcomes.']
        (out / 'report.md').write_text('\n'.join(lines) + '\n')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='action', required=True)
    builder = sub.add_parser('build')
    for name in ['head-dir', 'base-dir', 'tools-dir', 'output-dir']:
        builder.add_argument('--' + name, type=Path, required=True)
    runner = sub.add_parser('measure')
    for name in ['head-dir', 'build-dir', 'output-dir']:
        runner.add_argument('--' + name, type=Path, required=True)
    runner.add_argument('--group', choices=GROUPS, required=True)
    runner.add_argument('--case', dest='case_id', help='Exact case in the group; adds identical-binary Go native controls')
    runner.add_argument('--blocks', type=int, default=BLOCKS, help='Predeclared fixed process blocks (2–60; default portfolio is unchanged)')
    runner.add_argument('--threads-sustained', action='store_true', help='Separate fixed 100000-iteration Threads cohort; original22 is unchanged')
    runner.add_argument('--gc-sustained', action='store_true', help='Separate fixed long GC cohort; never replaces the original22 cases')
    args = parser.parse_args()
    (build if args.action == 'build' else measure)(args)


if __name__ == '__main__':
    main()

"""Owned host-callback and fresh-preparation baselines, separate from pure calls."""
import hashlib
import json
from pathlib import Path
import random
import statistics

import own_build

ROUNDS, SAMPLES, WARMUPS = 3, 7, 2
HOST_CASES = {'host-control', 'host-loop'}
PREPARATION_CASES = {'small-module','many-functions','large-functions','deep-control','many-locals',
                     'many-imports','tables-globals','data-64k','invalid-stack','excessive-locals','malformed'}


def digest(path): return hashlib.sha256(path.read_bytes()).hexdigest()


def fixtures(head):
    root = head / 'benchmarks/embedded'
    result = json.loads((root / 'subsystems/manifest.json').read_text())
    if result.get('version') != 'subsystems-v1' or result.get('producer') != 'wasm-tools 1.258.0':
        raise ValueError('unexpected subsystem fixture producer/version')
    if result['generatorSha256'] != digest(root / 'subsystems/generate.py'):
        raise ValueError('subsystem generator checksum mismatch')
    for group, expected in (('host', HOST_CASES), ('preparation', PREPARATION_CASES)):
        if len(result[group]) != len(expected) or {x['id'] for x in result[group]} != expected:
            raise ValueError('incomplete subsystem fixture matrix')
        for case in result[group]:
            path = (root / case['wasm']).resolve()
            if not path.is_relative_to((root / 'subsystems').resolve()) or digest(path) != case['wasmSha256'] or path.stat().st_size != case['inputBytes']:
                raise ValueError('subsystem fixture path/hash/size mismatch')
            source = path.with_suffix('.wat')
            if source.exists() and digest(source) != case['sourceSha256']:
                raise ValueError('subsystem WAT source hash mismatch')
    return result


def configurations(builds):
    result = []
    for ref in ('base','head'):
        for compiler in ('go','llgo'):
            b = builds[f'{ref}-{compiler}']
            if ref == 'base' and b['status'] == 'baseline-unavailable': continue
            if b['status'] != 'available': raise ValueError('subsystem engine unavailable')
            result.append({'id':f'{ref}-{compiler}', 'binary':b['binary'], 'sourceRevision':b['sourceRevision'],
                'compiler':'gc' if compiler=='go' else 'llgo', 'compilerRevision':b['toolchain']['revision'],
                'goos':b['toolchain']['goos'], 'goarch':b['toolchain']['goarch']})
    return result


def duration(value):
    if type(value) is not int or value < 0: raise ValueError('invalid subsystem duration')


def validate(item, runtime, case):
    r = item['record']
    expected = {key:runtime[key] for key in ('sourceRevision','compiler','compilerRevision','goos','goarch')}
    expected.update(schemaVersion=1,engine='wasmvm',runtimeVersion=own_build.GO_VERSION,
                    mode='host' if item['series']=='host' else 'prepare')
    if item['exitCode'] != 0 or item['wasmSha256'] != case['wasmSha256'] or any(r.get(k)!=v for k,v in expected.items()):
        raise ValueError('subsystem engine/fixture identity mismatch')
    duration(item['processReadyNs']); duration(item['processWallNs'])
    if item['series']=='host':
        if r.get('backend')!='interpreter' or r.get('policy')!='unmetered; bounded cancellation; checked Core 2.0':
            raise ValueError('host callback policy mismatch')
        for key, value in (('seed',case['seed']),('iterations',case['iterations']),('samples',SAMPLES),('warmups',WARMUPS),('result',case['expectedChecksum']),('hostCalls',case['iterations']*(1+SAMPLES+WARMUPS))):
            if type(r.get(key)) is not int or r[key]!=value: raise ValueError('host callback count/checksum mismatch')
        expected_phases = {'load','engineSetup','compileValidate','instantiate','exportLookup','firstCall','cleanup'}
        if set(r['phasesNs'])!=expected_phases: raise ValueError('incomplete host phase intervals')
        for key,count in (('warmCallNs',SAMPLES),('warmupCallNs',WARMUPS),('timerPairNs',SAMPLES)):
            if len(r[key])!=count: raise ValueError('incomplete host samples')
            for ns in r[key]: duration(ns)
        if r['warmResults']!=[case['expectedChecksum']]*SAMPLES or r['warmupResults']!=[case['expectedChecksum']]*WARMUPS:
            raise ValueError('host warm checksum mismatch')
        for ns in r['phasesNs'].values(): duration(ns)
        duration(r['hostBindingNs'])
        measured = sum(r['phasesNs'].values()) + r['hostBindingNs'] + sum(r['warmCallNs']) + sum(r['warmupCallNs'])
    else:
        if type(r.get('inputBytes')) is not int or r['inputBytes']!=case['inputBytes'] or len(r['samples'])!=SAMPLES:
            raise ValueError('preparation sample/input mismatch')
        duration(r['loadNs'])
        measured=r['loadNs']
        for sample in r['samples']:
            if sample['outcome']!=case['outcome'] or type(sample['definedFunctions']) is not int or sample['definedFunctions']!=case['definedFunctions'] or type(sample['imports']) is not int or sample['imports']!=case['imports']:
                raise ValueError('preparation outcome/module shape mismatch')
            wanted={'engineSetup','compileValidate','cleanup'}
            if case['outcome']=='valid': wanted|={'hostBinding','instantiate'}
            if set(sample['phasesNs'])!=wanted: raise ValueError('incomplete or invented preparation interval')
            for ns in sample['phasesNs'].values(): duration(ns)
            measured+=sum(sample['phasesNs'].values())
    if item['processReadyNs']+measured>item['processWallNs']:
        raise ValueError('subsystem intervals overlap process boundary')


def run(head, builds, timeout, seed, invoke, output):
    output.update(version='subsystems-v1',rounds=ROUNDS,samples=SAMPLES,warmups=WARMUPS,
                  fixtures=fixtures(head),runtimes=configurations(builds),observations=[])
    cases=[(runtime,group,case) for runtime in output['runtimes'] for group in ('host','preparation') for case in output['fixtures'][group]]
    rng=random.Random(seed)
    for number in range(ROUNDS):
        rng.shuffle(cases)
        for runtime, group, case in cases:
            path=str(head/'benchmarks/embedded'/case['wasm'])
            if group=='host':
                command=[runtime['binary'],'--mode','host','--wasm',path,'--seed',str(case['seed']),'--iterations',str(case['iterations']),'--expected',str(case['expectedChecksum']),'--samples',str(SAMPLES),'--warmups',str(WARMUPS)]
            else:
                command=[runtime['binary'],'prepare','--wasm',path,'--expect',case['outcome'],'--samples',str(SAMPLES)]
            output['activeCase']={'runtime':runtime['id'],'series':group,'case':case['id'],'round':number,'command':command}
            observation={**invoke(command,timeout),'runtime':runtime['id'],'series':group,'case':case['id'],'round':number,'wasmSha256':case['wasmSha256']}
            validate(observation,runtime,case)
            output['observations'].append(observation)
            del output['activeCase']
        print(f'Owned subsystem round {number+1}/{ROUNDS} complete',flush=True)


def summarize(series, builds):
    if series.get('version')!='subsystems-v1' or series.get('rounds')!=ROUNDS or series.get('samples')!=SAMPLES or series.get('warmups')!=WARMUPS or series['runtimes']!=configurations(builds):
        raise ValueError('subsystem configuration mismatch')
    runtimes={r['id']:r for r in series['runtimes']}
    fixtures_={(group,c['id']):c for group in ('host','preparation') for c in series['fixtures'][group]}
    for group, expected in (('host',HOST_CASES),('preparation',PREPARATION_CASES)):
        if len(series['fixtures'][group])!=len(expected) or {name for g,name in fixtures_ if g==group}!=expected:
            raise ValueError('subsystem fixture matrix incomplete')
    expected={(r,g,c,n) for r in runtimes for g,c in fixtures_ for n in range(ROUNDS)}
    seen=set()
    for item in series['observations']:
        key=(item['runtime'],item['series'],item['case'],item['round'])
        if type(item['round']) is not int or key not in expected or key in seen: raise ValueError('duplicate/unexpected subsystem observation')
        validate(item,runtimes[item['runtime']],fixtures_[item['series'],item['case']])
        seen.add(key)
    if seen!=expected: raise ValueError('incomplete subsystem process matrix')
    rows=[]
    for runtime in runtimes:
        for (group,name),case in fixtures_.items():
            chosen=[x for x in series['observations'] if x['runtime']==runtime and x['series']==group and x['case']==name]
            row={'runtime':runtime,'series':group,'case':name,'outcome':case['outcome']}
            if group=='host':
                row['callNs']=statistics.median(statistics.median(x['record']['warmCallNs']) for x in chosen)
                row['normalizedNs']=row['callNs']/case['iterations']
                row['hostBindingNs']=statistics.median(x['record']['hostBindingNs'] for x in chosen)
            else:
                phases=chosen[0]['record']['samples'][0]['phasesNs']
                row['phasesNs']={p:statistics.median(statistics.median(s['phasesNs'][p] for s in x['record']['samples']) for x in chosen) for p in phases}
            rows.append(row)
    return rows


def markdown(series, builds):
    rows=summarize(series,builds)
    lines=['<details><summary>Owned host callbacks and fresh preparation shapes</summary>','',
           'These separate diagnostics measure exact own base/head builds under Go and LLGo. Competitor throughput remains above; this subsystem series has no competitor rows. Host-call iteration includes the guest loop, VM boundary, callback arithmetic and marshalling. Each preparation sample uses a fresh Engine/Store and invokes no guest export. Invalid/limit rejection is labeled separately.','',
           '| Runtime | Host case | ns/callback iteration | Bind (us) |','| --- | --- | ---: | ---: |']
    for r in rows:
        if r['series']=='host': lines.append(f"| {r['runtime']} | {r['case']} | {r['normalizedNs']:.2f} | {r['hostBindingNs']/1000:.2f} |")
    lines+=['','| Runtime | Preparation case | Outcome | Setup (us) | Compile/reject (us) | Bind (us) | Instantiate (us) | Cleanup (us) |',
            '| --- | --- | --- | ---: | ---: | ---: | ---: | ---: |']
    for r in rows:
        if r['series']!='preparation': continue
        values=' | '.join(f"{r['phasesNs'][p]/1000:.2f}" if p in r['phasesNs'] else 'not applicable' for p in ('engineSetup','compileValidate','hostBinding','instantiate','cleanup'))
        lines.append(f"| {r['runtime']} | {r['case']} | {r['outcome']} | {values} |")
    lines+=['','Three independent processes, seven samples per process. Per-process medians are the aggregation unit. Readiness, load, resource usage and every sample remain in JSON. These diagnostics do not certify zero allocations; LLGo BDWGC allocation-event counts are unavailable (llgo#2923).','', '</details>','']
    return '\n'.join(lines)

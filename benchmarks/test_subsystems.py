"""Integrity checks for the separate owned subsystem measurement series."""
import copy
import importlib.util
from pathlib import Path
import unittest

import own_build
import subsystems

PROJECT=Path(__file__).resolve().parent.parent


def complete_series(builds):
    s={'version':'subsystems-v1','rounds':3,'samples':7,'warmups':2,'fixtures':subsystems.fixtures(PROJECT),
       'runtimes':subsystems.configurations(builds),'observations':[]}
    for runtime in s['runtimes']:
        for group in ('host','preparation'):
            for case in s['fixtures'][group]:
                for n in range(3):
                    r={k:runtime[k] for k in ('sourceRevision','compiler','compilerRevision','goos','goarch')}
                    r.update(schemaVersion=1,engine='wasmvm',runtimeVersion=own_build.GO_VERSION,mode='host' if group=='host' else 'prepare')
                    if group=='host':
                        r.update(backend='interpreter',policy='unmetered; bounded cancellation; checked Core 2.0',seed=case['seed'],iterations=case['iterations'],samples=7,warmups=2,
                                 result=case['expectedChecksum'],hostCalls=case['iterations']*10,phasesNs={p:1 for p in ('load','engineSetup','compileValidate','instantiate','exportLookup','firstCall','cleanup')},
                                 hostBindingNs=1,warmCallNs=[10]*7,warmupCallNs=[20]*2,timerPairNs=[0]*7,warmResults=[case['expectedChecksum']]*7,warmupResults=[case['expectedChecksum']]*2)
                    else:
                        phases={'engineSetup':1,'compileValidate':2,'cleanup':1}
                        if case['outcome']=='valid':phases.update(hostBinding=1,instantiate=1)
                        r.update(inputBytes=case['inputBytes'],loadNs=1,samples=[{'phasesNs':phases.copy(),'outcome':case['outcome'],'definedFunctions':case['definedFunctions'],'imports':case['imports'],'usage':{}} for _ in range(7)])
                    s['observations'].append({'runtime':runtime['id'],'series':group,'case':case['id'],'round':n,'wasmSha256':case['wasmSha256'],'exitCode':0,'processReadyNs':10,'processWallNs':1000,'record':r})
    return s


class SubsystemTests(unittest.TestCase):
    def report(self):
        from test_phase_ci import complete_report
        return complete_report()

    def test_complete_counts_and_separate_outcomes(self):
        r=self.report()
        s=r['subsystems']
        self.assertEqual(len(s['observations']),156)
        rows=subsystems.summarize(s,r['ownBuilds'])
        self.assertEqual(len(rows),52)
        self.assertEqual(next(x for x in rows if x['case']=='host-loop')['normalizedNs'],0.01)
        self.assertNotIn('instantiate',next(x for x in rows if x['case']=='invalid-stack')['phasesNs'])
        output=subsystems.markdown(s,r['ownBuilds'])
        self.assertIn('no competitor rows',output)
        self.assertIn('not applicable',output)

    def test_missing_corrupted_and_invented_observations_fail(self):
        r=self.report()
        def host(s):return next(x for x in s['observations'] if x['series']=='host')
        def prep(s):return next(x for x in s['observations'] if x['case']=='invalid-stack')
        mutations={
            'missing':lambda s:s['observations'].pop(),
            'duplicate':lambda s:s['observations'].append(s['observations'][0]),
            'runtime':lambda s:s['runtimes'][0].update(sourceRevision='f'*40),
            'round':lambda s:host(s).update(round=True),
            'hash':lambda s:host(s).update(wasmSha256='f'*64),
            'compiler':lambda s:host(s)['record'].update(compiler='foreign'),
            'callback-count':lambda s:host(s)['record'].update(hostCalls=0),
            'callback-result':lambda s:host(s)['record']['warmResults'].__setitem__(0,0),
            'duration':lambda s:host(s)['record']['warmCallNs'].__setitem__(0,-1),
            'overlap':lambda s:host(s).update(processWallNs=1),
            'prepare-outcome':lambda s:prep(s)['record']['samples'][0].update(outcome='valid'),
            'invented-instantiation':lambda s:prep(s)['record']['samples'][0]['phasesNs'].update(instantiate=1),
            'module-shape':lambda s:prep(s)['record']['samples'][0].update(definedFunctions=1),
            'incomplete-preparation':lambda s:prep(s)['record']['samples'].pop(),
            'invalid-load':lambda s:prep(s)['record'].update(loadNs=True),
        }
        for name,change in mutations.items():
            with self.subTest(case=name),self.assertRaises(ValueError):
                s=copy.deepcopy(r['subsystems']);change(s);subsystems.summarize(s,r['ownBuilds'])

    def test_host_oracles_use_the_same_independent_source_definition(self):
        path=PROJECT/'benchmarks/embedded/subsystems/generate.py'
        spec=importlib.util.spec_from_file_location('subsystem_generator',path)
        module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
        fixtures=subsystems.fixtures(PROJECT)
        self.assertEqual(module.checksum(0,1),0x9e3779b9)
        for case in fixtures['host']:
            self.assertEqual(module.checksum(case['seed'],case['iterations']),case['expectedChecksum'])
            for vector in case['validationCases']:
                self.assertEqual(module.checksum(vector['seed'],vector['iterations']),vector['expectedChecksum'])


if __name__=='__main__':unittest.main()

"""Keep SIMD diagnostics honest about fixture identity and incomplete evidence."""
from pathlib import Path
import unittest
from unittest.mock import patch

import native_simd
import phase_ci


class FixedSIMDTests(unittest.TestCase):
    def builds(self):
        return {f'{ref}-{compiler}': {'binary': f'{ref}-{compiler}', 'sourceRevision': ref,
                'toolchain': {'revision': 'go1.27.3' if compiler == 'go' else 'llgo-pin',
                              'goos': 'linux', 'goarch': 'amd64'}}
                for ref in ('base', 'head') for compiler in ('go', 'llgo')}

    def references(self):
        return [{'id': name, 'command': [name]} for name in
                ('head-compiler', 'wasmtime-cranelift', 'wasmer-cranelift')]

    def row(self, command):
        name = command[0]
        def option(key):
            return command[command.index(key)+1]
        expected, count = int(option('--expected')), int(option('--iterations'))
        own = name in self.builds()
        record = {'schemaVersion': 1, 'backend': 'native' if own else 'compiler',
                  'iterations': count, 'seed': int(option('--seed')), 'samples': int(option('--samples')),
                  'warmups': int(option('--warmups')), 'result': expected,
                  'phasesNs': {key: 1 for key in phase_ci.PHASES},
                  'warmCallNs': [100]*int(option('--samples')),
                  'warmupCallNs': [100]*int(option('--warmups')),
                  'timerPairNs': [1]*int(option('--samples')),
                  'warmResults': [expected]*int(option('--samples')),
                  'warmupResults': [expected]*int(option('--warmups'))}
        if own:
            compiler = name.split('-')[-1]
            record.update(engine='wasmvm', sourceRevision=name.split('-')[0],
                          compiler='gc' if compiler == 'go' else 'llgo',
                          compilerRevision='go1.27.3' if compiler == 'go' else 'llgo-pin',
                          runtimeVersion='go1.27.3', mode='api', goos='linux', goarch='amd64',
                          policy='unmetered; bounded cancellation; checked Core 2.0', nativeCPUFeatures=3)
        return {'record': record, 'exitCode': 0, 'processReadyNs': 1, 'processWallNs': 1000}

    def case(self):
        return ({'id': 'fixture', 'path': 'fixture.wasm', 'wasmSha256': 'fixture-hash',
                 'seed': 7, 'iterations': 10, 'expectedChecksum': 17}, lambda seed, count: seed+count)

    def test_pinned_complete_fixture_groups_and_hashes(self):
        cases = native_simd.fixtures(Path(__file__).resolve().parent.parent)
        self.assertEqual(len(cases), 41)
        self.assertEqual(len({f['id'] for f, _ in cases}), 41)
        self.assertEqual(sum(f['id'].startswith('regression-') for f, _ in cases), 8)
        with patch('native_simd.hashlib.sha256') as digest:
            digest.return_value.hexdigest.return_value = 'corrupt'
            with self.assertRaisesRegex(ValueError, 'hash mismatch'):
                native_simd.fixtures(Path(__file__).resolve().parent.parent)

    def test_complete_diagnostics_never_authorize_merging(self):
        def interleave(commands, timeout, rng):
            return {name: self.row(cmd) for name, cmd in commands.items()}, [list(commands)]
        with patch('native_simd.fixtures', return_value=[self.case()]), \
             patch('native_simd.phase_ci.invoke', side_effect=lambda cmd, timeout: self.row(cmd)), \
             patch('native_simd.interleaved.invoke', side_effect=interleave):
            report = native_simd.run(Path('.'), self.builds(), self.references())
        self.assertEqual(report['status'], 'diagnostic')
        self.assertEqual(len(report['checks']), 28)
        self.assertEqual(len(report['cases'][0]['observations']), 21)
        text = native_simd.markdown(report)
        self.assertIn('merge acceptance remains open', text)
        self.assertIn('Missing A/A calibration', text)

    def test_failed_checksum_preserves_partial_evidence(self):
        calls = 0
        def invoke(cmd, timeout):
            nonlocal calls
            calls += 1
            row = self.row(cmd)
            if calls == 2:
                row['record']['warmResults'] = [0]
            return row
        with patch('native_simd.fixtures', return_value=[self.case()]), \
             patch('native_simd.phase_ci.invoke', side_effect=invoke):
            report = native_simd.run(Path('.'), self.builds(), self.references())
        self.assertEqual(report['status'], 'failed')
        self.assertEqual(len(report['checks']), 1)
        self.assertIn('oracle mismatch', report['failure'])
        self.assertIn('Measurement failed', native_simd.markdown(report))


if __name__ == '__main__':
    unittest.main()

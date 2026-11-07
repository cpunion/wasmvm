from pathlib import Path
import unittest
from unittest.mock import patch

import routing_compare as routing


class RoutingComparisonTests(unittest.TestCase):
    def test_missing_runtime_dependencies_fail_instead_of_reaching_timing(self):
        self.assertEqual([('libgc.so.1', Path('/lib/libgc.so.1'))], routing.linked_paths(' libgc.so.1 => /lib/libgc.so.1 (0x1)\n'))
        with self.assertRaises(ValueError):
            routing.linked_paths(' libgc.so.1 => not found\n')

    def test_portfolio_retains_every_family_and_both_compiled_producers(self):
        root = Path(__file__).resolve().parent.parent
        fixtures = routing.fixtures(root)
        self.assertEqual(22, len(fixtures))
        self.assertEqual({f['id'] for f in fixtures}, {name for group in routing.GROUPS.values() for name in group})
        self.assertEqual(22, sum(map(len, routing.GROUPS.values())))
        self.assertEqual(3307539256, next(f['expectedChecksum'] for f in fixtures if f['id'] == 'applications/sha256-clang'))
        self.assertEqual({'sha256.c', 'sha256.rs'}, {f['producer']['source'] for f in fixtures if f['family'] == 'application'})

    def test_sustained_gc_has_fixed_counts_without_mutating_original_cases(self):
        root = Path(__file__).resolve().parent.parent
        for kind in ['struct-allocation', 'struct-access', 'struct-indirect-access', 'i31-roundtrip']:
            original = dict(id='gc/' + kind, family='gc', iterations=10_000,
                            seed=17, expectedChecksum=99, wasmSha256='fixture')
            with patch.object(routing, 'oracle', return_value=123) as oracle:
                measured = routing.measurement_fixture(root, original, True)
                count = 300_000 if kind == 'struct-allocation' else 1_000_000
                self.assertEqual(count, measured['iterations'])
                self.assertEqual(123, measured['expectedChecksum'])
                oracle.assert_called_once_with(root, measured, 17, count)
            self.assertEqual(10_000, original['iterations'])
            self.assertEqual(99, original['expectedChecksum'])
            self.assertEqual(original, routing.measurement_fixture(root, original))
            self.assertEqual('fixture', measured['wasmSha256'])
        with self.assertRaises(ValueError):
            routing.measurement_fixture(root, dict(family='core'), True)

    def test_common_features_are_instrumentation_only_and_boundary_is_exact(self):
        source = 'config.Backend = args.backend\n"unmetered; bounded cancellation; checked Core 2.0"'
        patched = routing.instrument_own(source)
        self.assertIn('wasm.SupportedFeatures', patched)
        self.assertIn(routing.POLICY, patched)
        for bad in ['', source + source, patched]:
            with self.assertRaises(ValueError):
                routing.instrument_own(bad)

    def test_reference_feature_patches_reject_missing_boundaries(self):
        for name in ['wasmtime', 'wasmer', 'wasmedge']:
            with self.assertRaises(ValueError):
                routing.reference_source(name, '')
        root = Path(__file__).resolve().parent.parent
        for name in ['wasmtime', 'wasmer', 'wasmedge']:
            source = (root / 'benchmarks/embedded/c' / (name + '.c')).read_text()
            self.assertNotEqual(source, routing.reference_source(name, source))

    def test_allocation_means_keep_slow_calls_and_calibrate_the_same_statistic(self):
        rows = [dict(runtime=name, block=i, record={'warmCallNs': values})
                for name, values in [('head-llgo-native', [1] * 6 + [701]),
                                     ('aa-head-llgo-native', [101] * 7)]
                for i in range(routing.BLOCKS)]
        case = dict(observations=rows)
        runtimes = {'head-llgo-native': {}, 'aa-head-llgo-native': {}}
        self.assertEqual(1, routing.warm_values(case)['head-llgo-native'])
        self.assertFalse(routing.calibration(case, runtimes)[0]['resolved'])
        case['warmStatistic'] = 'mean-including-pauses-v1'
        self.assertEqual(101, routing.warm_values(case)['head-llgo-native'])
        self.assertTrue(routing.calibration(case, runtimes)[0]['resolved'])
        self.assertEqual([1] * 6 + [701], rows[0]['record']['warmCallNs'])
        self.assertEqual(routing.BLOCKS * 2, len(rows))

    def test_calibration_rejects_consistent_bias_noise_and_unresolved_clocks(self):
        def case(values):
            return {'observations': [dict(runtime=name, block=i, record={'warmCallNs': [v] * 7})
                    for name, series in [('head-llgo-native', [100] * routing.BLOCKS), ('aa-head-llgo-native', values)]
                    for i, v in enumerate(series)]}
        runtimes = {'head-llgo-native': {}, 'aa-head-llgo-native': {}}
        self.assertTrue(routing.calibration(case([100] * routing.BLOCKS), runtimes)[0]['resolved'])
        for values in [[120] * routing.BLOCKS, [40, 150, 70, 140, 110], [0] * routing.BLOCKS]:
            self.assertFalse(routing.calibration(case(values), runtimes)[0]['resolved'])


if __name__ == '__main__':
    unittest.main()

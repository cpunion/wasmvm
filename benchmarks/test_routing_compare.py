from pathlib import Path
import unittest

import routing_compare as routing


class RoutingComparisonTests(unittest.TestCase):
    def test_portfolio_retains_every_family_and_both_compiled_producers(self):
        root = Path(__file__).resolve().parent.parent
        fixtures = routing.fixtures(root)
        self.assertEqual(22, len(fixtures))
        self.assertEqual({f['id'] for f in fixtures}, {name for group in routing.GROUPS.values() for name in group})
        self.assertEqual(22, sum(map(len, routing.GROUPS.values())))
        self.assertEqual(3307539256, next(f['expectedChecksum'] for f in fixtures if f['id'] == 'applications/sha256-clang'))
        self.assertEqual({'sha256.c', 'sha256.rs'}, {f['producer']['source'] for f in fixtures if f['family'] == 'application'})

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

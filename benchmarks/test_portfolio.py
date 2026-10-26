"""Check authored Core 2.0 portfolio provenance and independent checksums."""
import hashlib
import importlib.util
import json
from pathlib import Path
import sys
import unittest
import zlib

import phase_ci
from test_embedded_fixtures import binary_sections

ROOT = Path(__file__).resolve().parent / 'embedded'
SPEC = importlib.util.spec_from_file_location('portfolio_kernels', ROOT / 'portfolio/kernels.py')
KERNELS = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(KERNELS)


class PortfolioTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads((ROOT / 'fixtures/portfolio.json').read_text())

    def test_complete_versioned_families_and_oracles(self):
        m = self.manifest
        self.assertEqual(m['portfolioVersion'], 'core2-v1')
        self.assertEqual([(w['id'], w['iterations'], w['family']) for w in m['workloads']], list(KERNELS.CASES))
        self.assertEqual(len(phase_ci.load_workloads(ROOT.parents[1])['workloads']), 14)
        for w in m['workloads']:
            with self.subTest(workload=w['id']):
                self.assertEqual(w['expectedChecksum'], KERNELS.checksum(w['id'], w['seed'], w['iterations']))
                for v in w['validationCases']:
                    self.assertEqual(v['expectedChecksum'], KERNELS.checksum(w['id'], v['seed'], v['iterations']))
                self.assertNotEqual(KERNELS.checksum(w['id'], 0, 1), KERNELS.checksum(w['id'], 1, 1))
                self.assertNotEqual(KERNELS.checksum(w['id'], 7, 1), KERNELS.checksum(w['id'], 7, 2))
                for count in (0, -1, 2**31, True):
                    with self.assertRaises(ValueError): KERNELS.checksum(w['id'], 0, count)
                for seed in (-1, 2**32, True):
                    with self.assertRaises(ValueError): KERNELS.checksum(w['id'], seed, 1)

    def test_every_binary_source_producer_and_pure_boundary(self):
        for w in self.manifest['workloads']:
            for kind in ('wasm', 'wat', 'source'):
                path = ROOT / w[kind]
                self.assertEqual(hashlib.sha256(path.read_bytes()).hexdigest(), w[kind+'Sha256'])
            sections = binary_sections((ROOT / w['wasm']).read_bytes())
            self.assertNotIn(2, sections)
            self.assertNotIn(8, sections)
            self.assertTrue(w['producer']['version'])
            if w['id'] in KERNELS.BODIES:
                self.assertEqual((ROOT / w['wat']).read_text(), KERNELS.module(w['id']))
                self.assertEqual(w['producer']['version'], 'wasm-tools 1.258.0')
        applications = self.manifest['workloads'][-2:]
        self.assertIn('clang version 22.1.8', applications[0]['producer']['version'])
        self.assertEqual(applications[1]['producer']['version'], 'rustc 1.98.0 (88d9e12ae 2026-08-18)')

    def test_crc_oracle_against_independent_zlib(self):
        for seed in (0, 1, 0xffffffff, 0xdeadbeef):
            for count in (1, 3, 17):
                result = seed
                for i in range(count):
                    state = (seed + i) & 0xffffffff
                    packet = bytearray()
                    for _ in range(256):
                        state = (state * 1664525 + 1013904223) & 0xffffffff
                        packet.append(state >> 24)
                    result = KERNELS.rot(result, 3) ^ zlib.crc32(packet)
                self.assertEqual(KERNELS.checksum('packet-crc-rust', seed, count), result)


if __name__ == '__main__':
    unittest.main()

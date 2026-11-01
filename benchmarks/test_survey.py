"""Protect fixture identity, normalization and unsupported-feature reporting."""
from pathlib import Path
import importlib.util
import json
import tempfile
import unittest
from unittest.mock import patch

import survey
import survey_modern


class SurveyTests(unittest.TestCase):
    def test_sdk_library_layouts_and_missing_dependency(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "lib64").mkdir()
            (root / "lib64/libwasmedge.so").touch()
            self.assertEqual(survey.sdk_library_directory(root, "wasmedge", "Linux"), root / "lib64")
            (root / "lib").mkdir()
            (root / "lib/libwasmtime.dylib").touch()
            self.assertEqual(survey.sdk_library_directory(root, "wasmtime", "Darwin"), root / "lib")
            with self.assertRaisesRegex(FileNotFoundError, "SDK library is missing"):
                survey.sdk_library_directory(root, "wasmer", "Linux")

    def test_modern_manifest_has_independent_oracles(self):
        root = Path(__file__).resolve().parent / "modern"
        spec = importlib.util.spec_from_file_location("modern_oracle", root / "verify.py")
        oracle = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(oracle)
        fixtures = json.loads((root / "manifest.json").read_text())["workloads"]
        self.assertEqual(len(fixtures), 10)
        self.assertEqual(sum(f["requiresGC"] for f in fixtures), 4)
        self.assertEqual(sum(f["requiresThreads"] for f in fixtures), 5)
        for f in fixtures:
            self.assertEqual(oracle.checksum(f["id"], f["seed"], f["iterations"]), f["expectedChecksum"])
            self.assertEqual(oracle.checksum(f["id"], 0xffffffff, 17), f["alternateExpectedChecksum"])

    def test_complete_portfolio_and_corrupt_input(self):
        root = Path(__file__).resolve().parent.parent
        fixtures = survey.fixtures(root)
        self.assertEqual(len(fixtures), 68)
        self.assertEqual(len({f["id"] for f in fixtures}), 68)
        self.assertEqual(len({f["family"] for f in fixtures}), 8)
        with patch("survey.digest", return_value="corrupt"):
            with self.assertRaisesRegex(ValueError, "fixture hash mismatch"):
                survey.fixtures(root)

    def test_process_medians_and_nonloop_controls(self):
        observations = [dict(runtime="r", processReadyNs=10,
                             record=dict(warmCallNs=values, phasesNs={"compileValidate": 20}))
                        for values in ([1, 2, 10000], [3, 4, 10000], [5, 6, 10000])]
        cases = [dict(fixture=dict(id=unit, family="f", unit=unit, iterations=2), observations=observations)
                 for unit in ("ns/call", "ns/source iteration")]
        rows = survey.summarize(dict(cases=cases, runtimes={"r": {}}))
        self.assertEqual([r["warm"] for r in rows], [4, 2])
        self.assertEqual([r["processes"] for r in rows], [3, 3])

    def test_gc_rejection_does_not_hide_crashes_or_unrelated_failures(self):
        accepted = [
            ("wazero-interpreter", "embedded runner exit 1: compile/validate: invalid instruction 0xfb"),
            ("wazero-compiler", "embedded runner exit 1: compile/validate: invalid byte: 0x5f != 0x60"),
            ("wasmer-cranelift", "embedded runner exit 1: Validation error: gc support is not enabled"),
            ("wasmer-cranelift", "embedded runner exit 1: Validation error: struct indexed types not supported without the gc feature"),
        ]
        for name, message in accepted:
            self.assertTrue(survey_modern.expected_gc_rejection(name, RuntimeError(message)))
        for message in ("embedded runner exit -11: invalid instruction 0xfb",
                        "embedded runner exit 1: missing run export",
                        "embedded runner exit 1: compile/validate: resource limit",
                        "timeout"):
            self.assertFalse(survey_modern.expected_gc_rejection("wazero-compiler", RuntimeError(message)))
        self.assertFalse(survey_modern.expected_gc_rejection("wasmtime-cranelift", RuntimeError(accepted[2][1])))


if __name__ == "__main__":
    unittest.main()

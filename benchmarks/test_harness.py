"""Checks for result integrity and statistics; runtime execution is separate."""

import subprocess
import unittest
from unittest.mock import patch

import generate
import run


class HarnessTests(unittest.TestCase):
    def test_nearest_rank_and_median(self):
        actual = run.summary([7, 1, 6, 2, 5, 3, 4])
        self.assertEqual(actual, {"samples": 7, "minMs": 1, "medianMs": 4, "p95Ms": 7})

    def test_reject_invalid_samples(self):
        for samples in ([], [-1], [float("nan")], [float("inf")]):
            with self.subTest(samples=samples), self.assertRaises(ValueError):
                run.summary(samples)

    def test_fixture_integrity(self):
        self.assertEqual(len(run.load_workloads()["workloads"]), 6)

    def test_wrong_stdout_and_failed_exit_are_rejected(self):
        runtime = {"id": "test", "command": ["fake"]}
        workload = {"id": "test", "wasm": "fixtures/test.wasm", "expectedStdout": "expected\n"}
        for code, output in ((0, b"wrong\n"), (1, b"expected\n")):
            result = subprocess.CompletedProcess([], code, output, b"")
            with patch("run.subprocess.run", return_value=result), self.assertRaises(RuntimeError):
                run.measure(runtime, workload, False, 0, 1)

    def test_success_records_observed_output(self):
        runtime = {"id": "test", "command": ["fake"]}
        workload = {"id": "test", "wasm": "fixtures/test.wasm", "expectedStdout": "expected\n"}
        result = subprocess.CompletedProcess([], 0, b"expected\n", b"warning")
        with patch("run.subprocess.run", return_value=result):
            observed = run.measure(runtime, workload, False, 2, 1)
        self.assertEqual(observed["stdout"], "expected\n")
        self.assertEqual(observed["stderr"], "warning")
        self.assertEqual(observed["round"], 2)

    def test_checksum_oracle_and_output(self):
        self.assertEqual(generate.rotate_left(0x80000001, 1), 3)
        for identifier, _, expected, helper, body, _ in generate.definitions():
            wat, output = generate.module(identifier, expected, helper, body)
            self.assertIn(f"call $work i32.const {expected} i32.ne", wat)
            self.assertEqual(output, f"{identifier}:{expected}\n")


if __name__ == "__main__":
    unittest.main()

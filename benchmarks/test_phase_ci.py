"""Check stage boundaries, process handshakes, and complete phase evidence."""

import copy
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

import ci
import phase_ci as phase


PROJECT = Path(__file__).resolve().parent.parent


def complete_report():
    fixtures = phase.load_workloads(PROJECT)
    runtimes = [{"id": name, "backend": "interpreter" if name.endswith("interpreter") else "compiler"}
                for name in ci.CONFIGURATIONS]
    report = {"status": "complete", "baseSha": "a" * 40, "headSha": "b" * 40,
              "builds": {name: {"version": "v1.12.0"} for name in ("base", "head")},
              "fixtures": fixtures, "runtimes": runtimes, "observations": []}
    for runtime in runtimes:
        for workload in fixtures["workloads"]:
            for number in range(phase.ROUNDS):
                record = {"schemaVersion": 1, "backend": runtime["backend"], "iterations": workload["iterations"],
                          "seed": workload["seed"], "result": workload["expectedChecksum"],
                          "samples": phase.SAMPLES, "warmups": phase.WARMUPS,
                          "phasesNs": {name: 100 for name in phase.PHASES},
                          "warmCallNs": [10, 20, 30, 40, 50, 60, 70],
                          "warmupCallNs": [1000, 2000], "timerPairNs": [1] * phase.SAMPLES,
                          "warmResults": [workload["expectedChecksum"]] * phase.SAMPLES,
                          "warmupResults": [workload["expectedChecksum"]] * phase.WARMUPS}
                report["observations"].append({"runtime": runtime["id"], "workload": workload["id"], "round": number,
                    "record": record, "processReadyNs": 1000, "processWallNs": 10000,
                    "exitCode": 0, "wasmSha256": workload["wasmSha256"]})
    return report


class PhaseIntegrityTests(unittest.TestCase):
    def test_complete_matrix_uses_only_warm_api_call_samples(self):
        report = complete_report()
        result = phase.summarize(report)
        self.assertEqual(len(report["observations"]), 126)
        self.assertEqual(len(result), 42)
        for item in result:
            self.assertEqual(item["warmCallNs"], 40)
            self.assertEqual(item["firstCallNs"], 100)
            self.assertEqual(item["processReadyNs"], 1000)
            self.assertEqual(item["timerPairNs"], 1)
        control = next(item for item in result if item["workload"] == "call-control")
        self.assertEqual(control["normalizedNs"], 40)
        self.assertEqual(control["unit"], "ns/call")
        sustained = next(item for item in result if item["workload"] == "integer-sustained")
        self.assertEqual(sustained["normalizedNs"], 40 / 5_000_000)

    def test_reject_incomplete_duplicate_and_bad_measurements(self):
        original = complete_report()
        changes = {
            "missing-process": lambda r: r["observations"].pop(),
            "duplicate": lambda r: r["observations"].append(r["observations"][0]),
            "missing-runtime": lambda r: r["runtimes"].pop(),
            "bad-round": lambda r: r["observations"][0].update(round=True),
            "bad-exit": lambda r: r["observations"][0].update(exitCode=1),
            "bad-hash": lambda r: r["observations"][0].update(wasmSha256="0" * 64),
            "bad-first": lambda r: r["observations"][0]["record"].update(result=1),
            "bad-seed": lambda r: r["observations"][0]["record"].update(seed=True),
            "bad-warm": lambda r: r["observations"][0]["record"]["warmResults"].__setitem__(1, 1),
            "bad-warmup": lambda r: r["observations"][0]["record"]["warmupResults"].pop(),
            "missing-phase": lambda r: r["observations"][0]["record"]["phasesNs"].pop("load"),
            "missing-clock": lambda r: r["observations"][0]["record"]["timerPairNs"].pop(),
            "negative-time": lambda r: r["observations"][0]["record"]["warmCallNs"].__setitem__(1, -1),
            "float-time": lambda r: r["observations"][0]["record"]["warmCallNs"].__setitem__(1, float("nan")),
            "bool-time": lambda r: r["observations"][0]["record"]["phasesNs"].update(load=True),
            "overlap": lambda r: r["observations"][0]["record"]["phasesNs"].update(cleanup=20000),
            "late-ready": lambda r: r["observations"][0].update(processReadyNs=10001),
            "startup-overlap": lambda r: r["observations"][0].update(processReadyNs=9000),
        }
        for name, change in changes.items():
            with self.subTest(name=name), self.assertRaises(ValueError):
                report = copy.deepcopy(original)
                change(report)
                phase.summarize(report)

    def test_optional_parse_and_validate_are_not_double_counted(self):
        report = complete_report()
        item = report["observations"][0]
        item["record"]["preparationPartsNs"] = {"parse": 30, "validate": 70}
        phase.summarize(report)
        item["record"]["preparationPartsNs"]["parse"] = 31
        with self.assertRaisesRegex(ValueError, "disagree"):
            phase.summarize(report)

    def test_report_units_scope_and_no_startup_subtraction(self):
        report = complete_report()
        output = phase.markdown(report)
        self.assertIn("ns/iteration", output)
        self.assertIn("ns/call", output)
        self.assertIn("same engine version", output)
        self.assertIn("API dispatch", output)
        self.assertIn("Process ready", output)
        self.assertIn("Wasmtime Cranelift", output)
        self.assertIn("WasmEdge interpreter", output)
        self.assertLess(len(output.encode()), 40000)
        report["status"] = "failed"
        report["failure"] = "untrusted | row"
        failed = phase.markdown(report)
        self.assertNotIn("| Workload", failed)
        self.assertNotIn("untrusted", failed)

    def test_fixture_corruption_is_rejected(self):
        import shutil
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            shutil.copytree(PROJECT / "benchmarks/embedded/fixtures", root / "benchmarks/embedded/fixtures")
            phase.load_workloads(root)
            path = root / "benchmarks/embedded/fixtures/call-control.wasm"
            path.write_bytes(b"corrupt")
            with self.assertRaisesRegex(ValueError, "hash mismatch"):
                phase.load_workloads(root)


class ProcessProtocolTests(unittest.TestCase):
    def test_parent_releases_child_after_observing_readiness(self):
        code = 'import json,sys; print(\'{"ready":true}\',flush=True); assert sys.stdin.readline()=="run\\n"; print(json.dumps({"received":"run"}))'
        result = phase.invoke([sys.executable, "-c", code], 5)
        self.assertEqual(result["record"], {"received": "run"})
        self.assertLessEqual(result["processReadyNs"], result["processWallNs"])
        self.assertEqual(result["exitCode"], 0)

    def test_bad_protocol_missing_result_failed_exit_and_deadline(self):
        programs = [
            ('print("wrong",flush=True)', ValueError),
            ('print(\'{"ready":true}\',flush=True)', ValueError),
            ('import sys; print("failed",file=sys.stderr); sys.exit(1)', RuntimeError),
            ('import sys; print(\'{"ready":true}\',flush=True); sys.stdin.readline(); print("{}"); print("extra")', ValueError),
            ('import time; time.sleep(3)', TimeoutError),
        ]
        for code, error in programs:
            with self.subTest(code=code), self.assertRaises(error):
                phase.invoke([sys.executable, "-c", code], 0.5)

    def test_stderr_is_drained_and_output_is_bounded(self):
        code = 'import sys; print(\'{"ready":true}\',flush=True); sys.stdin.readline(); sys.stderr.write("x"*70000); print("{}")'
        result = phase.invoke([sys.executable, "-c", code], 5)
        self.assertEqual(len(result["stderr"]), 70000)
        with patch.object(phase, "MAX_OUTPUT", 1000), self.assertRaisesRegex(ValueError, "output limit"):
            phase.invoke([sys.executable, "-c", code], 5)

    def test_failed_main_preserves_phase_report(self):
        with tempfile.TemporaryDirectory() as temporary:
            arguments = ["phase_ci.py", "--base-dir", str(PROJECT), "--head-dir", str(PROJECT), "--tools-dir", temporary,
                         "--base-sha", "a" * 40, "--head-sha", "b" * 40, "--output-dir", temporary]
            with patch("sys.argv", arguments), patch("ci.verify_checkout", side_effect=ValueError("wrong input")), self.assertRaises(ValueError):
                phase.main()
            output = Path(temporary)
            data = json.loads((output / "results.json").read_text())
            self.assertEqual(data["status"], "failed")
            self.assertIn("wrong input", data["failure"])
            self.assertNotIn("| Workload", (output / "report.md").read_text())


if __name__ == "__main__":
    unittest.main()

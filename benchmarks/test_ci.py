"""Regression checks for benchmark integrity, comparison math, and failure reports."""

import copy
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import ci


PROJECT = Path(__file__).resolve().parent.parent


def complete_report():
    fixtures = ci.load_workloads(PROJECT)
    report = {"status": "complete", "baseSha": "a" * 40, "headSha": "b" * 40,
              "builds": {"base": {"version": "v1.12.0"}, "head": {"version": "v1.12.0"}},
              "runtimes": [{"id": identifier, "command": ["fake", identifier]} for identifier in ci.CONFIGURATIONS],
              "fixtures": fixtures, "observations": []}
    for runtime in report["runtimes"]:
        for workload in fixtures["workloads"]:
            for warmup, count in ((True, ci.WARMUPS), (False, ci.SAMPLES)):
                for round_number in range(count):
                    # Warmup timings differ deliberately: summaries must exclude them.
                    elapsed = 1000 if warmup else round_number + 1
                    report["observations"].append({"runtime": runtime["id"], "workload": workload["id"],
                                                   "warmup": warmup, "round": round_number,
                                                   "elapsedMs": elapsed, "exitCode": 0,
                                                   "stdout": workload["expectedStdout"],
                                                   "wasmSha256": workload["wasmSha256"]})
    return report


class ComparisonTests(unittest.TestCase):
    def test_comparison_formula_and_invalid_medians(self):
        self.assertAlmostEqual(ci.median_delta(10, 11), 10)
        self.assertAlmostEqual(ci.median_delta(10, 8), -20)
        self.assertEqual(ci.median_delta(10, 10), 0)
        for base, head in ((0, 1), (1, 0), (-1, 1), (float("nan"), 1), (1, float("inf"))):
            with self.subTest(base=base, head=head), self.assertRaises(ValueError):
                ci.median_delta(base, head)

    def test_complete_matrix_and_excluded_warmups(self):
        report = complete_report()
        self.assertEqual(len(report["observations"]), 378)
        result = ci.summarize(report)
        self.assertEqual(len(result), 42)
        for item in result:
            self.assertEqual(item["samples"], 7)
            self.assertEqual(item["medianMs"], 4)
            self.assertEqual(item["p95Ms"], 7)
            self.assertEqual(item["minMs"], 1)

    def test_truncated_duplicate_unknown_and_incorrect_observations_fail(self):
        original = complete_report()
        mutations = {
            "truncated": lambda report: report["observations"].pop(),
            "duplicate": lambda report: report["observations"].append(report["observations"][0]),
            "unknown-runtime": lambda report: report["observations"][0].update(runtime="unknown"),
            "wrong-round": lambda report: report["observations"][0].update(round=8),
            "wrong-warmup-type": lambda report: report["observations"][0].update(warmup=1),
            "wrong-round-type": lambda report: report["observations"][0].update(round=True),
            "exit": lambda report: report["observations"][0].update(exitCode=1),
            "stdout": lambda report: report["observations"][0].update(stdout="incorrect\n"),
            "fixture": lambda report: report["observations"][0].update(wasmSha256="0" * 64),
            "negative-time": lambda report: report["observations"][0].update(elapsedMs=-1),
            "zero-time": lambda report: report["observations"][0].update(elapsedMs=0),
            "nan-time": lambda report: report["observations"][0].update(elapsedMs=float("nan")),
            "infinite-time": lambda report: report["observations"][0].update(elapsedMs=float("inf")),
            "boolean-time": lambda report: report["observations"][0].update(elapsedMs=True),
            "missing-config": lambda report: report["runtimes"].pop(),
            "duplicate-config": lambda report: report["runtimes"].append(report["runtimes"][0]),
            "missing-workload": lambda report: report["fixtures"]["workloads"].pop(),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name), self.assertRaises(ValueError):
                report = copy.deepcopy(original)
                mutate(report)
                ci.summarize(report)

    def test_markdown_reports_scope_both_baselines_and_competitors(self):
        report = complete_report()
        output = ci.markdown(report)
        self.assertIn(report["baseSha"], output)
        self.assertIn(report["headSha"], output)
        self.assertIn("version is unchanged", output)
        self.assertIn("not production wasmvm code", output)
        self.assertIn("Base ms | Head ms", output)
        self.assertIn("Wasmtime ms | Wasmer ms | WasmEdge", output)
        self.assertIn("not a statistical significance test or a speed gate", output)
        self.assertLess(len(output.encode()), 50000)
        report["builds"]["head"]["version"] = "v1.13.0"
        self.assertNotIn("version is unchanged", ci.markdown(report))
        report["observations"].pop()
        with self.assertRaises(ValueError):
            ci.markdown(report)

    def test_failure_markdown_does_not_present_partial_or_inject_output(self):
        report = {"status": "failed", "baseSha": "a" * 40, "headSha": "b" * 40,
                  "failure": "attacker\n| fake result |", "observations": []}
        output = ci.markdown(report)
        self.assertIn("no performance comparison is accepted", output)
        self.assertNotIn("attacker", output)
        self.assertNotIn("Base ms | Head ms", output)


class InputTests(unittest.TestCase):
    def test_fixture_bytes_paths_and_matrix_are_validated(self):
        with tempfile.TemporaryDirectory() as temporary:
            head = Path(temporary)
            shutil.copytree(PROJECT / "benchmarks/fixtures", head / "benchmarks/fixtures")
            manifest_path = head / "benchmarks/fixtures/manifest.json"
            original = json.loads(manifest_path.read_text())
            self.assertEqual(len(ci.load_workloads(head)["workloads"]), 6)
            for kind in ("wat", "wasm"):
                workload = original["workloads"][0]
                fixture = head / "benchmarks" / workload[kind]
                saved = fixture.read_bytes()
                fixture.write_bytes(b"corrupt")
                with self.subTest(kind=kind), self.assertRaisesRegex(ValueError, "hash mismatch"):
                    ci.load_workloads(head)
                fixture.write_bytes(saved)
            for mutate in (
                lambda manifest: manifest["workloads"].pop(),
                lambda manifest: manifest["workloads"][0].update(wat="../outside.wat"),
                lambda manifest: manifest["workloads"][0].update(expectedStdout="missing newline"),
                lambda manifest: manifest.update(schemaVersion=2),
                lambda manifest: manifest.update(origin=""),
            ):
                altered = copy.deepcopy(original)
                mutate(altered)
                manifest_path.write_text(json.dumps(altered))
                with self.assertRaises(ValueError):
                    ci.load_workloads(head)

    def test_candidate_selection_pins_and_rejects_missing_or_replaced_module(self):
        required = {"Require": [{"Path": ci.MODULE, "Version": "v1.12.0"}]}
        with patch("ci.capture", return_value=json.dumps(required)):
            selected = ci.candidate_version(PROJECT, "go")
        self.assertEqual(selected["version"], "v1.12.0")
        self.assertEqual(selected["moduleFileSha256"], ci.sha256(PROJECT / "experiments/wazero/go.mod"))
        invalid = [{"Require": []}, {"Require": [{"Path": ci.MODULE, "Version": "latest"}]},
                   {**required, "Replace": [{"Old": {"Path": ci.MODULE}, "New": {"Path": "../fork"}}]}]
        for metadata in invalid:
            with patch("ci.capture", return_value=json.dumps(metadata)), self.assertRaises(ValueError):
                ci.candidate_version(PROJECT, "go")
        with tempfile.TemporaryDirectory() as temporary, self.assertRaisesRegex(ValueError, "no implicit baseline"):
            ci.candidate_version(Path(temporary), "go")

    def test_checkout_identity_and_tracked_changes(self):
        with self.assertRaisesRegex(ValueError, "full 40-character"):
            ci.verify_checkout(PROJECT, "main")
        with patch("ci.capture", return_value="a" * 40), self.assertRaisesRegex(ValueError, "expected"):
            ci.verify_checkout(PROJECT, "b" * 40)
        with patch("ci.capture", return_value="a" * 40), patch("ci.subprocess.run") as run:
            ci.verify_checkout(PROJECT, "a" * 40)
            self.assertIn("--quiet", run.call_args.args[0])
            self.assertTrue(run.call_args.kwargs["check"])
        with patch("ci.capture", return_value="a" * 40), patch("ci.subprocess.run", side_effect=subprocess.CalledProcessError(1, "git")), self.assertRaises(subprocess.CalledProcessError):
            ci.verify_checkout(PROJECT, "a" * 40)

    def test_competitor_pins_and_backend_flags(self):
        builds = {name: {"binary": f"/{name}/wazero", "version": "v1.12.0"} for name in ("base", "head")}
        versions = {"wasmtime": "wasmtime-cli 48.0.1 (commit date)",
                    "wasmer": "wasmer 7.3.0", "wasmedge": "/tools/wasmedge version 0.17.1\nother details"}
        with patch("ci.shutil.which", side_effect=lambda name: "/tools/" + name), patch("ci.capture", side_effect=lambda args: versions[Path(args[0]).name]), patch("ci.sha256", return_value="x"):
            configurations = ci.runtime_commands(builds)
            self.assertEqual({item["id"] for item in configurations}, set(ci.CONFIGURATIONS))
            for item in configurations:
                if item["id"].endswith("-interpreter") and item["project"] == "wazero":
                    self.assertIn("-interpreter", item["command"])
                if item["id"] == "wasmedge-interpreter":
                    self.assertEqual(item["command"][-2:], ["--run-mode", "interpreter"])
            versions["wasmtime"] = "wasmtime-cli 48.0.10"
            with self.assertRaisesRegex(ValueError, "version mismatch"):
                ci.runtime_commands(builds)
        with patch("ci.shutil.which", return_value=None), self.assertRaisesRegex(ValueError, "absent"):
            ci.runtime_commands(builds)


class ExecutionTests(unittest.TestCase):
    def test_measurement_asserts_exit_stdout_and_common_head_fixture(self):
        workload = ci.load_workloads(PROJECT)["workloads"][0]
        runtime = {"id": "fake", "command": ["tool"]}
        for code, stdout in ((1, workload["expectedStdout"].encode()), (0, b"wrong\n")):
            result = subprocess.CompletedProcess([], code, stdout, b"error")
            with patch("ci.subprocess.run", return_value=result), self.assertRaisesRegex(RuntimeError, "correctness failure"):
                ci.measure(runtime, workload, PROJECT, False, 0, 30)
        result = subprocess.CompletedProcess([], 0, workload["expectedStdout"].encode(), b"warning")
        with patch("ci.subprocess.run", return_value=result) as run:
            item = ci.measure(runtime, workload, PROJECT, False, 0, 30)
        self.assertEqual(run.call_args.args[0][-1], str(PROJECT / "benchmarks" / workload["wasm"]))
        self.assertEqual(item["wasmSha256"], workload["wasmSha256"])
        self.assertEqual(item["stderr"], "warning")
        with patch("ci.subprocess.run", side_effect=subprocess.TimeoutExpired("tool", 30)), self.assertRaises(subprocess.TimeoutExpired):
            ci.measure(runtime, workload, PROJECT, False, 0, 30)

    def test_schedule_is_complete_seeded_and_interleaved(self):
        original = complete_report()
        lookup = {(item["runtime"], item["workload"], item["warmup"], item["round"]): item for item in original["observations"]}
        orders = []
        for seed in (10, 10, 11):
            report = copy.deepcopy(original)
            report["observations"] = []
            with patch("ci.measure", side_effect=lambda runtime, workload, head, warmup, round_number, timeout: lookup[(runtime["id"], workload["id"], warmup, round_number)]), patch("builtins.print"):
                ci.execute(report, PROJECT, 30, seed)
            self.assertEqual(report["status"], "complete")
            ci.validate_matrix(report)
            order = [(item["runtime"], item["workload"]) for item in report["observations"]]
            for start in range(0, len(order), 42):
                self.assertEqual(len(set(order[start:start + 42])), 42)
            orders.append(order)
        self.assertEqual(orders[0], orders[1])
        self.assertNotEqual(orders[0], orders[2])

    def test_main_preserves_failure_artifact_without_performance_table(self):
        with tempfile.TemporaryDirectory() as temporary:
            arguments = ["ci.py", "--base-dir", str(PROJECT), "--head-dir", str(PROJECT),
                         "--base-sha", "a" * 40, "--head-sha", "b" * 40, "--output-dir", temporary]
            with patch("ci.verify_checkout", side_effect=ValueError("wrong checkout")), patch("sys.argv", arguments), self.assertRaisesRegex(ValueError, "wrong checkout"):
                ci.main()
            output = Path(temporary)
            report = json.loads((output / "results.json").read_text())
            self.assertEqual(report["status"], "failed")
            self.assertEqual(report["baseSha"], "a" * 40)
            self.assertEqual(report["headSha"], "b" * 40)
            self.assertIn("wrong checkout", report["failure"])
            self.assertNotIn("Base ms | Head ms", (output / "report.md").read_text())


if __name__ == "__main__":
    unittest.main()

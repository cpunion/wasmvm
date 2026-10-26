import copy
import unittest

import own_cli
from test_phase_ci import complete_report


class ProductionCLIIntegrityTests(unittest.TestCase):
    def test_complete_checked_cli_matrix(self):
        report = complete_report()
        rows = own_cli.summarize(report["productionCLI"], report["ownBuilds"], report["fixtures"])
        self.assertEqual(len(rows), 24)
        self.assertTrue(all(row["medianNs"] == 10000 for row in rows))
        text = own_cli.markdown(report["productionCLI"], report["ownBuilds"], report["fixtures"])
        self.assertIn("not subtracted", text)
        self.assertIn("0.0100", text)

    def test_no_missing_or_corrupt_cli_measurements(self):
        original = complete_report()
        mutations = {
            "identity": lambda s: s["runtimes"][0].update(sourceRevision="f" * 40),
            "missing": lambda s: s["observations"].pop(),
            "duplicate": lambda s: s["observations"].append(s["observations"][0]),
            "round": lambda s: s["observations"][0].update(round=True),
            "wall": lambda s: s["observations"][0].update(processWallNs=-1),
            "bits": lambda s: s["observations"][0]["result"]["results"][0].update(bits="0x00000000"),
            "backend": lambda s: s["observations"][0]["result"].update(backend="compiler"),
            "fixture": lambda s: s["observations"][0].update(wasmSha256="0" * 64),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name), self.assertRaises(ValueError):
                report = copy.deepcopy(original)
                mutate(report["productionCLI"])
                own_cli.summarize(report["productionCLI"], report["ownBuilds"], report["fixtures"])

    def test_cli_absence_does_not_replace_the_available_embedding_engine(self):
        report = complete_report()
        for compiler in ("go", "llgo"):
            report["ownBuilds"][f"base-{compiler}"]["cli"] = {"status": "baseline-unavailable"}
        series = report["productionCLI"]
        series["runtimes"] = [r for r in series["runtimes"] if r["id"].startswith("head-")]
        series["observations"] = [r for r in series["observations"] if r["runtime"].startswith("head-")]
        self.assertEqual(len(own_cli.summarize(series, report["ownBuilds"], report["fixtures"])), 12)
        self.assertIn("unavailable", own_cli.markdown(series, report["ownBuilds"], report["fixtures"]))
        report["ownBuilds"]["head-go"]["cli"]["status"] = "baseline-unavailable"
        with self.assertRaises(ValueError):
            own_cli.configurations(report["ownBuilds"])


if __name__ == "__main__":
    unittest.main()

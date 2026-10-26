import unittest

import native_compare as native


def observations(factor):
    rows = []
    names = [f"{ref}-{c}" for ref in ("base", "head") for c in ("go", "llgo")] + ["head-go-control", "head-llgo-control"] + list(native.REFERENCES)
    for n in range(native.ROUNDS):
        for name in names:
            if name.startswith("base-") and n>=native.BASELINE_ROUNDS: continue
            value = 10000 * (n + 1)  # Paired host variation must cancel.
            if name.startswith("head-") and name not in native.REFERENCES:
                value *= factor[n % len(factor)]
            rows.append({"id": name, "round": n, "record": {"warmCallNs": [value]*7, "iterations": 1}})
    return rows


class NativeGateTests(unittest.TestCase):
    def test_win_loss_and_unresolved_are_distinct(self):
        for factors, status in (([.8], "faster"), ([1.001], "parity"), ([1.2], "slower"), ([.9, 1.1], "inconclusive")):
            result = native.compare(observations(factors))
            self.assertEqual(len(result), 8)
            self.assertEqual({r["status"] for r in result if not r["reference"].endswith("-control")}, {status})

    def test_missing_duplicate_and_zero_samples_do_not_pass(self):
        rows = observations([.8])
        for invalid in (rows[:-1], rows + rows[:1]):
            with self.assertRaises(ValueError): native.compare(invalid)
        rows[0]["record"]["warmCallNs"] = [0]*7
        with self.assertRaises(ValueError): native.compare(rows)

    def test_one_fast_compiler_cannot_hide_the_other(self):
        rows = observations([.8])
        for r in rows:
            if r["id"] in ("head-llgo", "head-llgo-control"):
                r["record"]["warmCallNs"] = [30000*(r["round"]+1)]*7
        result = native.compare(rows)
        self.assertEqual({r["status"] for r in result if r["compiler"]=="go" and not r["reference"].endswith("-control")}, {"faster"})
        self.assertEqual({r["status"] for r in result if r["compiler"]=="llgo" and not r["reference"].endswith("-control")}, {"slower"})

    def test_interval_spanning_faster_and_parity_has_no_rejection_gap(self):
        result = native.compare(observations([.99, 1.0]))
        comparisons = [r for r in result if not r["reference"].endswith("-control")]
        for r in comparisons:
            self.assertLess(r["lower"], 1-native.PARITY)
            self.assertGreater(r["upper"], 1-native.PARITY)
            self.assertLessEqual(r["upper"], 1+native.PARITY)
            self.assertEqual(r["status"], "parity-or-faster")
        self.assertTrue(all(r["status"] in native.ACCEPTED for r in result))

    def test_possible_slowdown_still_blocks_with_a_fast_lower_bound(self):
        comparisons = native.compare(observations([.97, 1.03]))
        for r in comparisons:
            if not r["reference"].endswith("-control"):
                self.assertLess(r["lower"], 1-native.PARITY)
                self.assertGreater(r["upper"], 1+native.PARITY)
                self.assertEqual(r["status"], "inconclusive")

    def test_noisy_identical_binary_blocks_equivalence(self):
        rows = observations([1.0])
        for r in rows:
            if r["id"].endswith("-control"):
                r["record"]["warmCallNs"] = [x*(.99 if r["round"]%2 else 1.01) for x in r["record"]["warmCallNs"]]
        result = native.compare(rows)
        self.assertEqual({r["status"] for r in result if r["reference"].endswith("-control")}, {"inconclusive"})

    def test_longer_candidates_preserve_bytes_and_seed(self):
        fixture = {"id": "integer-sustained", "iterations": 5000000, "seed": 305419896,
                   "expectedChecksum": 3219525462, "wasmSha256": "owned-by-manifest", "wasm": "integer-mix.wasm"}
        candidate = native.candidate_fixture(fixture)
        self.assertEqual(candidate["iterations"], 25000000)
        self.assertEqual(candidate["expectedChecksum"], 3139208734)
        for key in ("id", "seed", "wasmSha256", "wasm"):
            self.assertEqual(candidate[key], fixture[key])
        self.assertEqual(fixture["iterations"], 5000000)
        with self.assertRaises(ValueError):
            native.candidate_fixture({**fixture, "seed": 1})

    def test_report_normalizes_descriptive_baseline_separately(self):
        rows = observations([1.0])
        for row in rows:
            count = 5000000 if row["id"].startswith("base-") else 25000000
            row["record"].update(iterations=count, warmCallNs=[2*count]*7,
                                 frontendNs=1000, backendNs=2000, firstCallNs=3*count)
        report = {"observations": rows, "fixture": {"iterations": 25000000},
                  "baselineFixture": {"iterations": 5000000}, "rounds": native.ROUNDS,
                  "comparisons": native.compare(rows), "status": "pass"}
        output = native.markdown(report)
        self.assertIn("| base-go | 2.000 | 0.003 | 15.000 |", output)
        self.assertIn("| head-go | 2.000 | 0.003 | 75.000 |", output)
        self.assertIn("head/base median ratio: 1.0000", output)
        self.assertIn("25,000,000 iterations", output)
        self.assertIn("baselines retain 5,000,000", output)
        self.assertTrue(all(r["status"] in native.ACCEPTED for r in report["comparisons"]))

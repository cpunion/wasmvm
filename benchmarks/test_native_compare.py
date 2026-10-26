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
            rows.append({"id": name, "round": n, "record": {"warmCallNs": [value]*7}})
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

    def test_noisy_identical_binary_blocks_equivalence(self):
        rows = observations([1.0])
        for r in rows:
            if r["id"].endswith("-control"):
                r["record"]["warmCallNs"] = [x*(.99 if r["round"]%2 else 1.01) for x in r["record"]["warmCallNs"]]
        result = native.compare(rows)
        self.assertEqual({r["status"] for r in result if r["reference"].endswith("-control")}, {"inconclusive"})

import math
import unittest

import comparison_format as fmt


class PercentageComparisonTests(unittest.TestCase):
    def test_time_reduction_and_speedup_have_different_denominators(self):
        stats = dict(ratio=0.5, ci95=[0.4, 0.6], controlsResolve=True)
        text = fmt.comparison(stats)
        self.assertIn("Δ -50.00%", text)
        self.assertIn("speed +100.00% [+66.67%, +150.00%]", text)
        self.assertTrue(text.endswith("; pass"))
        self.assertEqual("-50.00%", fmt.observed(200, 100))
        self.assertEqual("+100.00%", fmt.observed(50, 100))

    def test_throughput_direction_and_unresolved_controls_survive(self):
        text = fmt.comparison(dict(ratio=2, ci95=[1.5, 3]), higher_is_better=True)
        self.assertIn("speed +100.00% [+50.00%, +200.00%]", text)
        self.assertTrue(text.endswith("; unresolved"))
        self.assertEqual("+100.00%", fmt.observed(200, 100, higher_is_better=True))
        text = fmt.comparison(dict(ratio=0.99, ci95=[0.9, 1.1], controlsResolve=True))
        self.assertIn("[-9.09%, +11.11%]", text)

    def test_counters_do_not_claim_execution_speed(self):
        text = fmt.comparison(dict(ratio=0.5, ci95=[0.4, 0.6]), speed=False)
        self.assertIn("Δ -50.00%", text)
        self.assertNotIn("speed", text)
        self.assertEqual("+0.00%", fmt.observed(100, 100))
        self.assertEqual("+0.00%", fmt.signed(-0.0001))

    def test_missing_or_invalid_data_cannot_turn_into_a_numeric_delta(self):
        for bad in [0, -1, math.inf, -math.inf, math.nan]:
            with self.assertRaises(ValueError):
                fmt.percent(bad)
            for value, reference in [(bad, 1), (1, bad)]:
                with self.assertRaises(ValueError):
                    fmt.observed(value, reference)
        for low, high in [(1.1, 0.9), (0.1, 0.5), (2, 3), (0, 2)]:
            with self.assertRaises(ValueError):
                fmt.comparison(dict(ratio=1, ci95=[low, high]))


if __name__ == "__main__":
    unittest.main()

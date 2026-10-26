import contextlib
import io
import unittest

import check_coverage


class CoverageTests(unittest.TestCase):
    def check(self, profile, minimum=90, required=()):
        with contextlib.redirect_stdout(io.StringIO()):
            return check_coverage.check(profile, minimum, required)

    def test_package_failure_cannot_hide_behind_aggregate(self):
        profile = "mode: set\na/one.go:1.1,2.1 99 1\nb/two.go:1.1,2.1 1 0\n"
        self.assertFalse(self.check(profile))
        self.assertTrue(self.check(profile, 0))

    def test_exact_threshold_and_required_package(self):
        profile = "mode: count\na/one.go:1.1,2.1 9 5\na/two.go:1.1,2.1 1 0\n"
        self.assertTrue(self.check(profile, 90, ["a"]))
        self.assertFalse(self.check(profile, 90.01))
        with self.assertRaisesRegex(ValueError, "missing required coverage: b"):
            self.check(profile, required=["a", "b"])

    def test_duplicate_block_counts_statements_once(self):
        profile = "mode: atomic\na/one.go:1.1,2.1 4 0\na/one.go:1.1,2.1 4 8\n"
        self.assertEqual(check_coverage.package_coverage(profile), {"a": (4, 4)})
        with self.assertRaisesRegex(ValueError, "inconsistent duplicate"):
            self.check(profile + "a/one.go:1.1,2.1 5 1\n")

    def test_empty_or_corrupt_profiles_are_errors(self):
        for profile in [
            "", "mode: unknown\n", "mode: set\n", "mode: set\ninvalid\n",
            "mode: set\nfile.go:1.1,2.1 1 1\n",
            "mode: set\na/one.go:1.1,2.1 1 -1\n",
            "mode: set\na/one.go:1.1,2.1 0 1\n",
        ]:
            with self.subTest(profile=profile), self.assertRaises(ValueError):
                self.check(profile)

    def test_zero_statement_blocks_do_not_invent_package_coverage(self):
        profile = "mode: set\na/doc.go:1.1,1.1 0 0\nb/impl.go:1.1,2.1 1 1\n"
        self.assertEqual(check_coverage.package_coverage(profile), {"b": (1, 1)})
        with self.assertRaisesRegex(ValueError, "missing required coverage: a"):
            self.check(profile, required=["a"])


if __name__ == "__main__":
    unittest.main()

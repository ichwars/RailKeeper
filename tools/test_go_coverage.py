import unittest

from check_go_coverage import failures, summarize


class GoCoverageChecks(unittest.TestCase):
    def test_cross_package_blocks_count_once_and_merge_execution(self):
        report = summarize("mode: set\np/internal/rules/a.go:1.1,3.2 3 0\np/internal/rules/a.go:1.1,3.2 3 1\np/internal/rules/b.go:1.1,2.2 1 0\n", "p/internal/")
        self.assertEqual(report["total"], {"statements": 4, "covered": 3, "percent": 75})
        self.assertFalse(failures(report, {"total": 75, "modules": {"rules": 75}}))

    def test_missing_module_and_regression_fail(self):
        report = summarize("mode: count\np/internal/rules/a.go:1.1,3.2 2 0\n", "p/internal/")
        self.assertEqual(len(failures(report, {"total": 74, "modules": {"rules": 80, "storage": 70}})), 3)

    def test_empty_and_outside_profiles_are_not_successful(self):
        with self.assertRaises(ValueError):
            summarize("", "p/internal/")
        with self.assertRaises(ValueError):
            summarize("mode: set\np/command/a.go:1.1,2.2 1 1\n", "p/internal/")
        self.assertTrue(failures(summarize("mode: set\n", "p/internal/"), {"total": 74, "modules": {"rules": 80}}))

    def test_inconsistent_blocks_and_negative_counts_are_rejected(self):
        for profile in ("mode: set\np/internal/rules/a.go:1.1,2.2 1 0\np/internal/rules/a.go:1.1,2.2 2 1\n", "mode: set\np/internal/rules/a.go:1.1,2.2 1 -1\n"):
            with self.assertRaises(ValueError):
                summarize(profile, "p/internal/")


if __name__ == "__main__":
    unittest.main()

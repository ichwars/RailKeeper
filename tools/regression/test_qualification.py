import unittest

from qualification import read_candidate


class LocalVersionClient:
    def __init__(self, build):
        self.build = build

    def request(self, method, path):
        assert method == "GET" and path == "/api/v1/version"
        return {"version": "0.1.22", "status": "update_available", "build": self.build}


class CandidateQualification(unittest.TestCase):
    def test_source_and_compiler_remain_separate_from_update_status(self):
        for status in ("release", "development"):
            client = LocalVersionClient({"commit": "a" * 40, "shortCommit": "a" * 12, "status": status, "goVersion": "go-test", "platform": "windows/amd64"})
            result = read_candidate(client, "a" * 40, "go-test")
            self.assertEqual(result["build"]["status"], status)
            self.assertEqual(result["version"], "0.1.22")

    def test_modified_unknown_and_mismatched_sources_fail_qualification(self):
        for status, commit in (("modified", "a" * 40), ("unknown", "a" * 40), ("development", "b" * 40)):
            with self.assertRaises(AssertionError):
                read_candidate(LocalVersionClient({"commit": commit, "status": status}), "a" * 40)

    def test_wrong_compiler_fails_even_for_the_same_commit(self):
        client = LocalVersionClient({"commit": "a" * 40, "status": "development", "goVersion": "go-other", "platform": "linux/test"})
        with self.assertRaises(AssertionError):
            read_candidate(client, "a" * 40, "go-test")

    def test_legacy_responses_stay_readable_but_are_not_exact_candidates(self):
        self.assertIsNone(read_candidate(LocalVersionClient(None))["build"])
        with self.assertRaises(AssertionError):
            read_candidate(LocalVersionClient(None), "a" * 40)
        with self.assertRaises(AssertionError):
            read_candidate(LocalVersionClient({"commit": "a" * 40, "status": "development"}), "a" * 40)


if __name__ == "__main__":
    unittest.main()

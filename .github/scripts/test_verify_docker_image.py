import unittest

from verify_docker_image import verify, SPDX, SLSA

REPOSITORY = "https://github.com/ichwars/RailKeeper"
COMMIT = "d" * 40


def digest(character):
    return f"sha256:{character * 64}"


class DockerEvidenceTests(unittest.TestCase):
    def setUp(self):
        self.descriptors = []
        self.qualification = {}
        self.sbom = {}
        self.provenance = {}
        self.manifests = {}
        for arch, char in (("amd64", "a"), ("arm64", "b")):
            platform = f"linux/{arch}"
            manifest_digest = digest(char)
            attestation_digest = digest("1" if arch == "amd64" else "2")
            self.descriptors.extend([
                {"digest": manifest_digest, "platform": {"os": "linux", "architecture": arch}},
                {"digest": attestation_digest, "annotations": {
                    "vnd.docker.reference.type": "attestation-manifest",
                    "vnd.docker.reference.digest": manifest_digest,
                }},
            ])
            self.qualification[arch] = {
                "build": {"platform": platform, "commit": COMMIT},
                "indexDigest": digest("3"),
                "platformDigest": manifest_digest,
            }
            key = platform + "/v8" if arch == "arm64" else platform
            self.sbom[key] = {"SPDX": {"spdxVersion": "SPDX-2.3", "packages": [{"name": "busybox"}]}}
            self.provenance[key] = {"SLSA": {"buildDefinition": {
                "resolvedDependencies": [{"uri": f"git+{REPOSITORY}.git#{COMMIT}",
                                          "digest": {"sha1": COMMIT}}],
            }}}
            self.manifests[attestation_digest] = {"layers": [
                {"mediaType": "application/vnd.in-toto+json", "annotations": {"in-toto.io/predicate-type": kind}}
                for kind in (SPDX, SLSA)
            ]}

    def check(self):
        return verify({"manifests": self.descriptors}, self.qualification,
                      self.sbom, self.provenance, self.manifests.__getitem__, REPOSITORY, COMMIT)

    def test_preserves_the_tested_platforms_and_evidence(self):
        report = self.check()
        self.assertEqual(report["linux/amd64"]["digest"], digest("a"))
        self.assertEqual(report["linux/arm64"]["sbomPackages"], 1)

    def test_rejects_an_untested_manifest(self):
        self.qualification["arm64"]["platformDigest"] = digest("c")
        with self.assertRaisesRegex(ValueError, "not the tested manifest"):
            self.check()

    def test_rejects_attestation_reassigned_to_another_platform(self):
        self.descriptors[3]["annotations"]["vnd.docker.reference.digest"] = digest("a")
        with self.assertRaisesRegex(ValueError, "references do not match"):
            self.check()

    def test_rejects_an_sbom_without_provenance(self):
        self.manifests[digest("2")]["layers"].pop()
        with self.assertRaisesRegex(ValueError, "lacks SBOM or SLSA provenance"):
            self.check()

    def test_rejects_the_wrong_git_commit(self):
        definition = self.provenance["linux/arm64/v8"]["SLSA"]["buildDefinition"]
        definition["resolvedDependencies"][0]["digest"] = {"sha1": "e" * 40}
        with self.assertRaisesRegex(ValueError, "expected Git source and commit"):
            self.check()


if __name__ == "__main__":
    unittest.main(verbosity=2)

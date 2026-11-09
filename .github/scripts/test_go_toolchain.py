"""Reject compiler and analyzer drift without downloading a toolchain."""
import tempfile
import unittest
from pathlib import Path

from check_go_toolchain import check


class ToolchainAlignmentTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / "backend").mkdir()
        (self.root / ".github/workflows").mkdir(parents=True)
        self.write(".go-version", "1.27.3\n")
        self.write("backend/go.mod", "module example\n\ngo 1.27.3\n")
        self.write("Dockerfile", "FROM golang:1.27.3-alpine AS backend-build\n")
        for name in ("ci.yml", "windows-standalone.yml", "codeql.yml"):
            self.write(f".github/workflows/{name}", "go-version-file: .go-version\n")
        self.write(".github/workflows/ci.yml", "go-version-file: .go-version\ngo-version-file: backend/go.mod\nversion: v2.14.0\n")

    def write(self, name, body):
        (self.root / name).write_text(body)

    def test_same_version_and_qualified_analyzer_are_required(self):
        self.assertIn("Go 1.27.3 is identical", check(self.root))

    def test_separately_pinned_older_and_newer_module_versions_fail(self):
        for version in ("1.26.10", "1.27.4"):
            with self.subTest(version=version):
                self.write("backend/go.mod", f"module example\ngo {version}\n")
                with self.assertRaisesRegex(SystemExit, "same Go compiler"):
                    check(self.root)

    def test_docker_drift_fails(self):
        self.write("Dockerfile", "FROM golang:1.26.10-alpine AS backend-build\n")
        with self.assertRaisesRegex(SystemExit, "Docker release compiler differs"):
            check(self.root)

    def test_implicit_native_compiler_selection_fails(self):
        self.write(".github/workflows/windows-standalone.yml", "go-version: stable\n")
        with self.assertRaisesRegex(SystemExit, "must use the pinned release compiler"):
            check(self.root)

    def test_old_analyzer_cannot_reintroduce_the_previous_split(self):
        self.write(".github/workflows/ci.yml", "go-version-file: .go-version\ngo-version-file: backend/go.mod\nversion: v2.11.4\n")
        with self.assertRaisesRegex(SystemExit, "qualified golangci-lint v2.14.0"):
            check(self.root)


if __name__ == "__main__":
    unittest.main()

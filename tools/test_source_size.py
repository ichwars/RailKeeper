import tempfile
import unittest
from pathlib import Path

from check_source_size import check


class SourceSizeTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / "backend/internal/api").mkdir(parents=True)
        (self.root / "frontend/src/shared/i18n").mkdir(parents=True)

    def write(self, name, body):
        path = self.root / name
        path.write_text(body)

    def test_both_languages_fail_when_an_executable_module_grows(self):
        self.write("backend/internal/api/handler.go", "package api\n" * 500)
        self.write("frontend/src/shared/request.ts", "export {};\n" * 500)
        result = check(self.root)
        self.assertEqual(len(result["violations"]), 2)
        self.assertTrue(all("500 lines exceed 499" in item for item in result["violations"]))

    def test_boundary_and_test_files_keep_the_fixed_production_scope(self):
        self.write("backend/internal/api/handler.go", "package api\n" * 499)
        self.write("backend/internal/api/handler_test.go", "test\n" * 700)
        self.write("frontend/src/shared/request.test.ts", "test\n" * 700)
        result = check(self.root)
        self.assertEqual(result["violations"], [])
        self.assertEqual(len(result["files"]), 1)

    def test_dictionary_exception_rejects_hidden_executable_code(self):
        name = "frontend/src/shared/i18n/de.ts"
        self.write(name, 'export const deTranslations: Record<string, string> = {"greeting": "Guten Morgen"};\n')
        self.assertEqual(check(self.root)["violations"], [])
        self.write(name, 'export const deTranslations: Record<string, string> = {"greeting": fetch("/api/private")};\n')
        self.assertEqual(len(check(self.root)["violations"]), 1)


if __name__ == "__main__":
    unittest.main()

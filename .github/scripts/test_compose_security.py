"""Check resolved deployment settings without building or starting a container."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class ComposeSecurityTests(unittest.TestCase):
    def environment(self, **settings):
        clean = {key: value for key, value in os.environ.items()
                 if not key.startswith(("RAILKEEPER_", "COMPOSE_"))}
        with tempfile.TemporaryDirectory() as directory:
            fixture = Path(directory) / "deployment.env"
            fixture.write_text("".join(f"{key}={value}\n" for key, value in settings.items()), encoding="utf-8")
            result = subprocess.run(["docker", "compose", "--env-file", str(fixture),
                                     "--project-name", "railkeeper-config-test", "-f", str(ROOT / "docker-compose.yml"),
                                     "config", "--format", "json"],
                                    env=clean, capture_output=True, text=True, timeout=30, check=True)
        return json.loads(result.stdout)["services"]["railkeeper"]["environment"]

    def test_default_local_http_keeps_auto_without_a_public_url(self):
        settings = self.environment()
        self.assertEqual(settings["RAILKEEPER_COOKIE_SECURE"], "auto")
        self.assertEqual(settings["RAILKEEPER_PUBLIC_URL"], "")

    def test_public_urls_reach_the_container_for_auto_cookie_resolution(self):
        for url in ("https://railkeeper.example.test", "http://127.0.0.1:8080"):
            with self.subTest(url=url):
                settings = self.environment(RAILKEEPER_PUBLIC_URL=url)
                self.assertEqual(settings["RAILKEEPER_PUBLIC_URL"], url)
                self.assertEqual(settings["RAILKEEPER_COOKIE_SECURE"], "auto")

    def test_explicit_cookie_overrides_are_preserved(self):
        for policy in ("true", "false"):
            with self.subTest(policy=policy):
                settings = self.environment(RAILKEEPER_PUBLIC_URL="https://railkeeper.example.test",
                                            RAILKEEPER_COOKIE_SECURE=policy)
                self.assertEqual(settings["RAILKEEPER_COOKIE_SECURE"], policy)
                self.assertEqual(settings["RAILKEEPER_PUBLIC_URL"], "https://railkeeper.example.test")


if __name__ == "__main__":
    unittest.main()

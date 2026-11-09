import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, mkdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { after, test } from "node:test";
import { validateReleaseDocs } from "./validate-release-docs.mjs";

const roots = [];
after(async () => { await Promise.all(roots.map((root) => rm(root, { recursive: true, force: true }))); });
const png = Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aRz8AAAAASUVORK5CYII=", "base64");
const versions = { stable: "0.2.0", development: "main" };
async function fixture() {
  const root = await mkdtemp(join(tmpdir(), "railkeeper-release-docs-"));
  roots.push(root);
  const put = async (path, data) => {
    const target = join(root, path);
    await mkdir(dirname(target), { recursive: true });
    await writeFile(target, data);
  };
  await put("backend/cmd/railkeeper/main.go", 'var (\n version = "0.2.0"\n)');
  for (const path of ["CHANGELOG.md", "CHANGELOG.de.md"]) {
    await put(path, "# Changelog\n\n## Unreleased\n\n## [0.2.0] - 2026-11-09\n\n### Added\n\n- Retained data.\n\n## [0.1.22] - 2026-11-01\n\n- Previous.\n");
  }
  await put("docs/releases/v0.2.0.md", "# v0.2.0\n\n## Deutsch\n\nErhalten.\n\n## English\n\nRetained.\n");
  const sources = { schemaVersion: 2, applicationVersion: "0.2.0", buildStatus: "release",
    applicationCommit: "e".repeat(40), languages: ["de", "en"], captures: {} };
  for (const language of sources.languages) {
    for (const file of ["overview.png", "account-two-factor-setup.png"]) {
      const path = `${language}/${file}`;
      await put("docs/site/assets/screenshots/" + path, png);
      sources.captures[path] = { sha256: createHash("sha256").update(png).digest("hex"),
        applicationVersion: sources.applicationVersion, applicationCommit: sources.applicationCommit,
        width: 1, height: 1, capturedAt: "2026-11-09T18:00:00.000Z", privacyMask: file.startsWith("account-") };
    }
  }
  return { root, put, sources, check: () => validateReleaseDocs(root, versions, sources) };
}

test("accepts a complete bilingual release and matching capture bytes", async () => {
  const f = await fixture();
  assert.deepEqual(await f.check(), []);
});

test("rejects an absent or empty current release changelog even with older entries", async () => {
  const f = await fixture();
  await f.put("CHANGELOG.md", "## [0.1.22] - 2026-11-01\n\n- Old release.\n");
  await f.put("CHANGELOG.de.md", "## [0.2.0] - 2026-11-09\n\n## [0.1.22] - 2026-11-01\n\n- Alt.\n");
  const errors = await f.check();
  assert.ok(errors.some((error) => error.startsWith("CHANGELOG.md:")));
  assert.ok(errors.some((error) => error.startsWith("CHANGELOG.de.md:")));
});

test("rejects diverging changelog dates, runtime version and monolingual release notes", async () => {
  const f = await fixture();
  await f.put("CHANGELOG.md", "## [0.2.0] - 2026-11-08\n\n- Retained.\n");
  await f.put("backend/cmd/railkeeper/main.go", 'var version = "0.2.1"');
  await f.put("docs/releases/v0.2.0.md", "## Deutsch\n\nNur Deutsch.\n");
  const errors = await f.check();
  assert.ok(errors.some((error) => error.includes("release dates differ")));
  assert.ok(errors.some((error) => error.includes("differs from runtime")));
  assert.ok(errors.some((error) => error.includes("German and English release notes")));
});

test("rejects stale screenshot release provenance", async () => {
  const f = await fixture();
  f.sources.applicationVersion = "0.1.22";
  f.sources.buildStatus = "development";
  assert.ok((await f.check()).some((error) => error.includes("current stable release")));
});

test("rejects an older capture hidden inside a current release manifest", async () => {
  const f = await fixture();
  f.sources.captures["de/overview.png"].applicationVersion = "0.1.22";
  f.sources.captures["en/overview.png"].applicationCommit = "a".repeat(40);
  const errors = await f.check();
  assert.equal(errors.filter((error) => error.includes("different application version or commit")).length, 2);
});

test("rejects changed screenshot bytes and dimensions", async () => {
  const f = await fixture();
  f.sources.captures["de/overview.png"].sha256 = "0".repeat(64);
  f.sources.captures["en/overview.png"].width = 1440;
  const errors = await f.check();
  assert.ok(errors.some((error) => error.includes("SHA256 differs")));
  assert.ok(errors.some((error) => error.includes("dimensions differ")));
});

test("rejects missing, extra, unpaired and unmasked capture evidence", async () => {
  const f = await fixture();
  delete f.sources.captures["de/overview.png"];
  f.sources.captures["de/account-two-factor-setup.png"].privacyMask = false;
  f.sources.captures["en/overview.png"].capturedAt = "unknown";
  f.sources.captures["en/deleted.png"] = {};
  await f.put("docs/site/assets/screenshots/de/new.png", png);
  const errors = await f.check();
  for (const reason of ["missing screenshot provenance", "privacy mask", "capture timestamp", "missing screenshot", "image sets differ"]) {
    assert.ok(errors.some((error) => error.includes(reason)), reason);
  }
});

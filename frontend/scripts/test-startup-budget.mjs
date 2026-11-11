import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

const script = fileURLToPath(new URL("./check-startup-budget.mjs", import.meta.url));

test("only initial assets count and exceeding either gzip budget produces a warning without failing", async t => {
  const dist = await mkdtemp(join(tmpdir(), "startup-budget-"));
  t.after(() => rm(dist, { force: true, recursive: true }));
  const assets = join(dist, "assets");
  await mkdir(assets);
  await writeFile(join(dist, "index.html"), `
    <script type="module" src="/assets/main.js"></script>
    <link rel="modulepreload" href="/assets/shared.js">
    <link rel="stylesheet" href="/assets/main.css">
  `);
  // Fixed high-entropy data stays above either limit after gzip.
  const largeAsset = Buffer.concat(Array.from({ length: 9_600 }, (_, index) =>
    createHash("sha256").update(String(index)).digest()));
  await Promise.all([
    writeFile(join(assets, "main.js"), "export {}"),
    writeFile(join(assets, "shared.js"), "export {}"),
    writeFile(join(assets, "main.css"), "html {}"),
    writeFile(join(assets, "later.js"), largeAsset),
  ]);

  const run = () => {
    const result = spawnSync(process.execPath, [script, dist], {
      encoding: "utf8", env: { ...process.env, GITHUB_ACTIONS: "true" },
    });
    assert.ifError(result.error);
    return result;
  };
  const withinBudget = run();
  assert.equal(withinBudget.status, 0, withinBudget.stderr);
  assert.equal(withinBudget.stderr, "");

  await Promise.all([
    writeFile(join(assets, "shared.js"), largeAsset),
    writeFile(join(assets, "main.css"), largeAsset),
  ]);
  const overBudget = run();
  assert.equal(overBudget.status, 0, overBudget.stderr);
  assert.match(overBudget.stderr, /::warning title=Frontend performance budget::Initial JS .* exceeds the 250 KiB budget/);
  assert.match(overBudget.stderr, /::warning title=Frontend performance budget::Initial CSS .* exceeds the 75 KiB budget/);
});

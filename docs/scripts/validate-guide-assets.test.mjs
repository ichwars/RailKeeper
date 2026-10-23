import assert from "node:assert/strict";
import { test } from "node:test";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { validateGuideAssets } from "./validate-guide-assets.mjs";

async function fixture(run) {
  const root = await mkdtemp(join(tmpdir(), "railkeeper-guide-assets-"));
  const files = ["guide/index.md", "de/guide/index.md"];
  for (const language of ["en", "de"]) {
    const path = join(root, "assets/screenshots", language, "example.png");
    await mkdir(dirname(path), { recursive: true });
    await writeFile(path, "demo image fixture");
    const page = join(root, language === "de" ? files[1] : files[0]);
    await mkdir(dirname(page), { recursive: true });
    await writeFile(page, `![${language} interface](${language === "de" ? "../../" : "../"}assets/screenshots/${language}/example.png)`);
  }
  try { await run(root, files); } finally { await rm(root, { recursive: true, force: true }); }
}

test("accepts localized paired screenshots with coverage evidence", async () => {
  await fixture(async (root, files) => {
    const topics = [{ id: "demo", labelDE: "Beispiel", labelEN: "Example", reviewEvidence: "demo-ui",
      reviewedAt: "2026-10-23", screenshot: "example" }];
    assert.deepEqual(await validateGuideAssets(root, files, topics), []);
  });
});

test("rejects a missing image, empty alt text and mismatched translations", async () => {
  await fixture(async (root, files) => {
    await writeFile(join(root, files[1]), "![](../../assets/screenshots/en/missing.png)");
    const errors = await validateGuideAssets(root, files, []);
    assert(errors.some((message) => message.includes("alternative text")));
    assert(errors.some((message) => message.includes("missing or out-of-tree image")));
    assert(errors.some((message) => message.includes("must use de interface")));
    assert(errors.some((message) => message.includes("differ from German")));
  });
});

test("rejects image paths outside the handbook and missing demo evidence", async () => {
  await fixture(async (root, files) => {
    await writeFile(join(root, files[0]), "![Secret](../../credentials.png)");
    const errors = await validateGuideAssets(root, files, [{ id: "demo", reviewEvidence: "demo-ui" }]);
    assert(errors.some((message) => message.includes("out-of-tree")));
    assert(errors.some((message) => message.includes("localized interface picture")));
    assert(errors.some((message) => message.includes("missing reader labels")));
    assert(errors.some((message) => message.includes("missing review evidence")));
    assert(errors.some((message) => message.includes("missing de demo evidence")));
  });
});

test("validates localized GuideFigure sources and alternative text", async () => {
  await fixture(async (root, files) => {
    for (const file of files) {
      await writeFile(join(root, file), '<GuideFigure name="example" alt="An illustrated workflow" caption="Demo" />');
    }
    assert.deepEqual(await validateGuideAssets(root, files, []), []);
    await writeFile(join(root, files[0]), '<GuideFigure name="missing" alt="" caption="Demo" />');
    const errors = await validateGuideAssets(root, files, []);
    assert(errors.some((message) => message.includes("alternative text")));
    assert(errors.some((message) => message.includes("missing or out-of-tree image")));
  });
});

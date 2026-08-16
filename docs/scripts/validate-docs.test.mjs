import assert from "node:assert/strict";
import { after, test } from "node:test";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";

import { parseFrontmatter, validateCoverage, validateDocumentTree } from "./validate-docs.mjs";

const versions = {
  stable: "0.1.18",
  development: "main",
};

const temporaryRoots = [];

after(async () => {
  await Promise.all(temporaryRoots.map((root) => rm(root, { force: true, recursive: true })));
});

async function fixture(files) {
  const root = await mkdtemp(join(tmpdir(), "railkeeper-docs-"));
  temporaryRoots.push(root);

  await Promise.all(
    Object.entries(files).map(async ([relativePath, source]) => {
      const absolutePath = join(root, relativePath);
      await mkdir(dirname(absolutePath), { recursive: true });
      await writeFile(absolutePath, source, "utf8");
    }),
  );

  return root;
}

function page(audience, status, reviewedVersion, overrides = {}) {
  const values = {
    title: "Test",
    description: "Test page.",
    audience,
    status,
    reviewedVersion,
    lastReviewed: "2026-08-15",
    ...overrides,
  };

  return `---
title: ${values.title}
description: ${values.description}
audience: ${values.audience}
status: ${values.status}
reviewedVersion: ${values.reviewedVersion}
lastReviewed: ${values.lastReviewed}
---

# Test
`;
}

test("accepts a complete matching language pair", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.18"),
    "de/index.md": page("reference", "stable", "0.1.18"),
  });
  assert.deepEqual(await validateDocumentTree(root, versions), []);
});

test("reports a missing German counterpart", async () => {
  const root = await fixture({ "guide/index.md": page("user", "stable", "0.1.18") });
  assert.deepEqual(await validateDocumentTree(root, versions), [
    "guide/index.md: missing counterpart de/guide/index.md",
  ]);
});

test("reports a missing English counterpart", async () => {
  const root = await fixture({
    "de/administration/index.md": page("admin", "stable", "0.1.18"),
  });
  assert.deepEqual(await validateDocumentTree(root, versions), [
    "de/administration/index.md: missing counterpart administration/index.md",
  ]);
});

test("reports missing and mismatched metadata", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.18"),
    "de/index.md": page("developer", "development", "main"),
  });
  const errors = await validateDocumentTree(root, versions);
  assert(errors.includes("index.md: audience differs from de/index.md"));
  assert(errors.includes("index.md: status differs from de/index.md"));
  assert(errors.includes("index.md: reviewedVersion differs from de/index.md"));
});

test("enforces the canonical version for each status", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.14"),
    "de/index.md": page("reference", "stable", "0.1.14"),
  });
  assert((await validateDocumentTree(root, versions)).some((error) =>
    error.includes("reviewedVersion must be 0.1.18"),
  ));
});

test("rejects invalid metadata and unfinished markers", async () => {
  const invalid = page("reader", "draft", "next", { lastReviewed: "15.08.2026" }).replace(
    "# Test",
    "# Test\n\nTBD",
  );
  const root = await fixture({ "index.md": invalid, "de/index.md": invalid });
  const errors = await validateDocumentTree(root, versions);

  assert(errors.includes("index.md: audience must be one of admin, developer, reference, user"));
  assert(errors.includes("index.md: status must be one of development, stable"));
  assert(errors.includes("index.md: lastReviewed must use YYYY-MM-DD"));
  assert(errors.includes("index.md: contains unfinished marker TBD"));
});

test("requires every contract field", async () => {
  const incomplete = `---
title: Test
description: Test page.
---

# Test
`;
  const root = await fixture({ "index.md": incomplete, "de/index.md": incomplete });
  const errors = await validateDocumentTree(root, versions);

  for (const field of ["audience", "status", "reviewedVersion", "lastReviewed"]) {
    assert(errors.includes(`index.md: missing ${field}`));
  }
});

test("frontmatter parsing rejects duplicate keys and unterminated blocks", () => {
  assert.throws(
    () => parseFrontmatter("---\naudience: user\naudience: admin\n---\n"),
    /duplicate frontmatter key audience/,
  );
  assert.throws(() => parseFrontmatter("---\naudience: user\n"), /unterminated frontmatter/);
});

test("frontmatter parsing ignores nested VitePress theme data", () => {
  assert.deepEqual(
    parseFrontmatter(`---
layout: home
audience: reference
hero:
  name: RailKeeper
  actions:
    - theme: brand
---
`),
    {
      layout: "home",
      audience: "reference",
      hero: "",
    },
  );
});

test("coverage validation reports every unmapped source surface", async () => {
  const contentRoot = await fixture({});
  const inventory = {
    frontendRoutes: ["/unmapped"],
    translationKeys: ["settings.newArea.title"],
    apiRoutes: [{ access: "Viewer", method: "GET", path: "/api/v1/unmapped" }],
    environmentVariables: ["RAILKEEPER_UNMAPPED"],
  };
  const manifest = {
    schemaVersion: 1,
    topics: [
      {
        id: "vehicles",
        audience: "user",
        status: "documented",
        englishPath: "guide/vehicles/index.md",
        germanPath: "de/guide/vehicles/index.md",
      },
    ],
    owners: {
      frontendRoutes: {},
      translationPrefixes: {},
      apiPrefixes: {},
      environmentVariables: {},
    },
  };

  const errors = validateCoverage(inventory, manifest, contentRoot);
  assert(errors.includes("frontend route /unmapped is not covered"));
  assert(errors.includes("translation key settings.newArea.title is not covered"));
  assert(errors.includes("API route GET /api/v1/unmapped is not covered"));
  assert(errors.includes("environment variable RAILKEEPER_UNMAPPED is not covered"));
  assert(
    errors.includes(
      "coverage topic vehicles references missing English page guide/vehicles/index.md",
    ),
  );
});

test("coverage validation accepts one owner for every source item", async () => {
  const contentRoot = await fixture({});
  const inventory = {
    frontendRoutes: ["/vehicles"],
    translationKeys: ["vehicles.cv.title"],
    apiRoutes: [{ access: "Viewer", method: "GET", path: "/api/v1/vehicles" }],
    environmentVariables: ["RAILKEEPER_ADDR"],
  };
  const manifest = {
    schemaVersion: 1,
    topics: [
      {
        id: "vehicles",
        audience: "user",
        status: "planned",
        englishPath: "guide/vehicles/index.md",
        germanPath: "de/guide/vehicles/index.md",
      },
    ],
    owners: {
      frontendRoutes: { "/vehicles": "vehicles" },
      translationPrefixes: { vehicles: "vehicles", "vehicles.cv": "vehicles" },
      apiPrefixes: { "/api/v1/vehicles": "vehicles" },
      environmentVariables: { RAILKEEPER_ADDR: "vehicles" },
    },
  };

  assert.deepEqual(validateCoverage(inventory, manifest, contentRoot), []);
});

test("coverage validation rejects invalid topics and unknown owner references", async () => {
  const contentRoot = await fixture({});
  const manifest = {
    schemaVersion: 1,
    topics: [
      {
        id: "duplicate",
        audience: "reader",
        status: "draft",
        englishPath: "guide/index.md",
        germanPath: "de/guide/index.md",
      },
      {
        id: "duplicate",
        audience: "user",
        status: "planned",
        englishPath: "guide/index.md",
        germanPath: "de/guide/index.md",
      },
    ],
    owners: {
      frontendRoutes: { "/vehicles": "missing" },
      translationPrefixes: {},
      apiPrefixes: {},
      environmentVariables: {},
    },
  };
  const errors = validateCoverage(
    { frontendRoutes: [], translationKeys: [], apiRoutes: [], environmentVariables: [] },
    manifest,
    contentRoot,
  );

  assert(errors.includes("coverage topic duplicate has invalid audience reader"));
  assert(errors.includes("coverage topic duplicate has invalid status draft"));
  assert(errors.includes("duplicate coverage topic duplicate"));
  assert(errors.includes("frontend route owner /vehicles references unknown topic missing"));
});

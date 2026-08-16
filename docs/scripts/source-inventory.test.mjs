import assert from "node:assert/strict";
import { after, test } from "node:test";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";

import {
  buildSourceInventory,
  extractApiRoutes,
  extractEnvironmentVariables,
  extractFrontendRoutes,
  extractOpenApiPaths,
  extractTranslationKeys,
} from "./source-inventory.mjs";

const temporaryRoots = [];

after(async () => {
  await Promise.all(temporaryRoots.map((root) => rm(root, { force: true, recursive: true })));
});

async function fixture(files) {
  const root = await mkdtemp(join(tmpdir(), "railkeeper-inventory-"));
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

test("extracts visible frontend routes", () => {
  assert.deepEqual(extractFrontendRoutes('startsWith("/vehicles")\nhref: "/overview"'), [
    "/overview",
    "/vehicles",
  ]);
});

test("extracts full translation keys", () => {
  assert.deepEqual(extractTranslationKeys('  "vehicles.cv.title": "CV",'), [
    "vehicles.cv.title",
  ]);
});

test("extracts API route specifications", () => {
  assert.deepEqual(
    extractApiRoutes('{http.MethodGet, "/api/v1/vehicles", routeAccessViewer, handler, nil}'),
    [{ access: "Viewer", method: "GET", path: "/api/v1/vehicles" }],
  );
});

test("extracts only OpenAPI operations below path keys", () => {
  assert.deepEqual(extractOpenApiPaths("paths:\n  /vehicles:\n    get:\n    post:\n"), [
    "GET /vehicles",
    "POST /vehicles",
  ]);

  assert.deepEqual(
    extractOpenApiPaths("components:\n  schemas:\n    Vehicle:\n      properties:\n        get:\n"),
    [],
  );
});

test("extracts RailKeeper environment variables", () => {
  assert.deepEqual(extractEnvironmentVariables('env("RAILKEEPER_ADDR", ":8080")'), [
    "RAILKEEPER_ADDR",
  ]);
});

test("builds a deterministic source inventory when translations match", async () => {
  const translations = 'export const translations = {\n  "nav.overview": "Overview",\n};\n';
  const root = await fixture({
    "frontend/src/app/App.tsx": 'startsWith("/vehicles")',
    "frontend/src/app/Shell.tsx": 'href: "/overview"',
    "frontend/src/shared/i18n/en.ts": translations,
    "frontend/src/shared/i18n/de.ts": translations,
    "backend/internal/api/routes.go":
      '{http.MethodGet, "/api/v1/vehicles", routeAccessViewer, handler, nil}',
    "backend/config.go": 'env("RAILKEEPER_ADDR", ":8080")',
    "openapi/railkeeper.yaml": "paths:\n  /vehicles:\n    get:\n",
    "README.md": "# RailKeeper",
    "README.de.md": "# RailKeeper",
    "docs/architecture.md": "# Architecture",
    "docs/releases/v0.1.0.md": "# Release",
  });

  const first = await buildSourceInventory(root);
  const second = await buildSourceInventory(root);

  assert.deepEqual(first, second);
  assert.deepEqual(first.translationKeys, ["nav.overview"]);
  assert.deepEqual(first.translationNamespaces, ["nav"]);
  assert.deepEqual(first.legacyDocuments, [
    "README.de.md",
    "README.md",
    "docs/architecture.md",
    "docs/releases/v0.1.0.md",
  ]);
});

test("rejects mismatched English and German translation keys", async () => {
  const root = await fixture({
    "frontend/src/app/App.tsx": "",
    "frontend/src/app/Shell.tsx": "",
    "frontend/src/shared/i18n/en.ts": '"nav.overview": "Overview",',
    "frontend/src/shared/i18n/de.ts": '"nav.vehicles": "Fahrzeuge",',
    "backend/internal/api/routes.go": "",
    "openapi/railkeeper.yaml": "paths:\n",
  });

  await assert.rejects(
    () => buildSourceInventory(root),
    /translation keys differ: missing in German: nav\.overview; missing in English: nav\.vehicles/,
  );
});

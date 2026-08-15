import { execFile } from "node:child_process";
import { readdir, readFile } from "node:fs/promises";
import { basename, extname, relative, resolve, sep } from "node:path";
import { promisify } from "node:util";
import { fileURLToPath, pathToFileURL } from "node:url";

const execFileAsync = promisify(execFile);
const excludedSegments = new Set([
  ".cache",
  ".git",
  ".superpowers",
  "data",
  "dist",
  "node_modules",
]);
const textExtensions = new Set([
  ".bat",
  ".cmd",
  ".css",
  ".env",
  ".go",
  ".html",
  ".js",
  ".json",
  ".md",
  ".mjs",
  ".ps1",
  ".sh",
  ".sql",
  ".toml",
  ".ts",
  ".tsx",
  ".txt",
  ".yaml",
  ".yml",
]);
const extensionlessTextFiles = new Set([
  ".gitattributes",
  ".gitignore",
  "Dockerfile",
  "LICENSE",
  "Makefile",
]);

function sortUnique(values) {
  return [...new Set(values)].sort();
}

export function extractFrontendRoutes(source) {
  const routes = [];
  const patterns = [
    /startsWith\(\s*["'](\/[^"']*)["']\s*\)/g,
    /\bhref\s*:\s*["'](\/[^"']*)["']/g,
    /\bhref\s*=\s*["'](\/[^"']*)["']/g,
  ];

  for (const pattern of patterns) {
    for (const match of source.matchAll(pattern)) {
      routes.push(match[1]);
    }
  }

  return sortUnique(routes);
}

export function extractTranslationKeys(source) {
  return sortUnique([...source.matchAll(/^\s*"([^"]+)"\s*:/gm)].map((match) => match[1]));
}

export function extractApiRoutes(source) {
  const routes = [];
  const routePattern =
    /\{\s*http\.Method([A-Za-z]+)\s*,\s*"([^"]+)"\s*,\s*routeAccess([A-Za-z]+)\s*,/g;

  for (const match of source.matchAll(routePattern)) {
    routes.push({
      access: match[3],
      method: match[1].toUpperCase(),
      path: match[2],
    });
  }

  return routes.sort((left, right) => {
    const leftKey = `${left.method} ${left.path} ${left.access}`;
    const rightKey = `${right.method} ${right.path} ${right.access}`;
    return leftKey < rightKey ? -1 : leftKey > rightKey ? 1 : 0;
  });
}

export function extractOpenApiPaths(source) {
  const operations = [];
  const methodPattern = /^    (delete|get|head|options|patch|post|put|trace):\s*(?:#.*)?$/i;
  let currentPath = "";
  let inPaths = false;

  for (const line of source.replace(/\r\n/g, "\n").split("\n")) {
    if (!inPaths) {
      inPaths = /^paths:\s*(?:#.*)?$/.test(line);
      continue;
    }
    if (/^[^\s#][^:]*:/.test(line)) {
      break;
    }

    const pathMatch = line.match(/^  (\/[^:]*):\s*(?:#.*)?$/);
    if (pathMatch) {
      currentPath = pathMatch[1];
      continue;
    }

    const methodMatch = currentPath ? line.match(methodPattern) : null;
    if (methodMatch) {
      operations.push(`${methodMatch[1].toUpperCase()} ${currentPath}`);
    }
  }

  return sortUnique(operations);
}

export function extractEnvironmentVariables(source) {
  return sortUnique([...source.matchAll(/\bRAILKEEPER_[A-Z0-9][A-Z0-9_]*\b/g)].map((match) => match[0]));
}

function excludedPath(relativePath) {
  const segments = relativePath.split("/");
  if (segments.some((segment) => excludedSegments.has(segment))) {
    return true;
  }

  return relativePath.startsWith("docs/scripts/") || relativePath.startsWith("docs/superpowers/");
}

function isTextFile(relativePath) {
  return (
    textExtensions.has(extname(relativePath).toLowerCase()) ||
    extensionlessTextFiles.has(basename(relativePath))
  );
}

async function walkFiles(root, directory, files) {
  const entries = await readdir(directory, { withFileTypes: true });
  entries.sort((left, right) => left.name.localeCompare(right.name, "en"));

  for (const entry of entries) {
    const absolutePath = resolve(directory, entry.name);
    const relativePath = relative(root, absolutePath).split(sep).join("/");
    if (excludedPath(relativePath)) {
      continue;
    }
    if (entry.isDirectory()) {
      await walkFiles(root, absolutePath, files);
    } else if (entry.isFile()) {
      files.push(relativePath);
    }
  }
}

async function repositoryFiles(root) {
  try {
    const { stdout } = await execFileAsync("git", ["ls-files", "-z"], {
      cwd: root,
      encoding: "utf8",
      maxBuffer: 10 * 1024 * 1024,
    });
    return sortUnique(stdout.split("\0").filter(Boolean).filter((path) => !excludedPath(path)));
  } catch {
    const files = [];
    await walkFiles(root, root, files);
    return sortUnique(files);
  }
}

async function readRepositoryFile(root, relativePath) {
  return readFile(resolve(root, ...relativePath.split("/")), "utf8");
}

function translationDifference(english, german) {
  const englishSet = new Set(english);
  const germanSet = new Set(german);
  const missingGerman = english.filter((key) => !germanSet.has(key));
  const missingEnglish = german.filter((key) => !englishSet.has(key));

  if (missingGerman.length === 0 && missingEnglish.length === 0) {
    return "";
  }

  return [
    `missing in German: ${missingGerman.join(", ") || "none"}`,
    `missing in English: ${missingEnglish.join(", ") || "none"}`,
  ].join("; ");
}

export async function buildSourceInventory(repositoryRoot) {
  const root = resolve(repositoryRoot);
  const [appSource, shellSource, englishSource, germanSource, apiSource, openApiSource, files] =
    await Promise.all([
      readRepositoryFile(root, "frontend/src/app/App.tsx"),
      readRepositoryFile(root, "frontend/src/app/Shell.tsx"),
      readRepositoryFile(root, "frontend/src/shared/i18n/en.ts"),
      readRepositoryFile(root, "frontend/src/shared/i18n/de.ts"),
      readRepositoryFile(root, "backend/internal/api/routes.go"),
      readRepositoryFile(root, "openapi/railkeeper.yaml"),
      repositoryFiles(root),
    ]);

  const englishKeys = extractTranslationKeys(englishSource);
  const germanKeys = extractTranslationKeys(germanSource);
  const difference = translationDifference(englishKeys, germanKeys);
  if (difference) {
    throw new Error(`translation keys differ: ${difference}`);
  }

  const environmentVariables = [];
  for (const path of files.filter(isTextFile)) {
    const source = await readRepositoryFile(root, path);
    environmentVariables.push(...extractEnvironmentVariables(source));
  }

  const legacyDocuments = files.filter(
    (path) =>
      /^README(?:\.[^/]+)?\.md$/.test(path) ||
      /^docs\/[^/]+\.md$/.test(path) ||
      /^docs\/releases\/[^/]+\.md$/.test(path),
  );

  return {
    schemaVersion: 1,
    frontendRoutes: sortUnique([
      ...extractFrontendRoutes(appSource),
      ...extractFrontendRoutes(shellSource),
    ]),
    translationKeys: englishKeys,
    translationNamespaces: sortUnique(englishKeys.map((key) => key.split(".", 1)[0])),
    apiRoutes: extractApiRoutes(apiSource),
    openApiOperations: extractOpenApiPaths(openApiSource),
    environmentVariables: sortUnique(environmentVariables),
    legacyDocuments: sortUnique(legacyDocuments),
  };
}

async function runCli() {
  const repositoryRoot = resolve(fileURLToPath(new URL("../..", import.meta.url)));
  const inventory = await buildSourceInventory(repositoryRoot);
  process.stdout.write(`${JSON.stringify(inventory, null, 2)}\n`);
}

const invokedPath = process.argv[1] ? pathToFileURL(resolve(process.argv[1])).href : "";
if (invokedPath === import.meta.url) {
  await runCli();
}

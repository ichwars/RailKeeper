import { readdir, readFile } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const requiredFields = ["audience", "status", "reviewedVersion", "lastReviewed"];
const allowedAudiences = ["admin", "developer", "reference", "user"];
const allowedStatuses = ["development", "stable"];
const pairFields = ["audience", "status", "reviewedVersion", "lastReviewed"];
const unfinishedMarkerPattern = /\b(TODO|TBD|FIXME)\b/;

function frontmatterParts(source) {
  const normalized = source.replace(/^\uFEFF/, "").replace(/\r\n/g, "\n");
  if (!normalized.startsWith("---\n")) {
    throw new Error("missing frontmatter");
  }

  const end = normalized.indexOf("\n---\n", 4);
  if (end === -1) {
    throw new Error("unterminated frontmatter");
  }

  return {
    body: normalized.slice(end + 5),
    frontmatter: normalized.slice(4, end),
  };
}

export function parseFrontmatter(source) {
  const { frontmatter } = frontmatterParts(source);
  const metadata = {};

  for (const line of frontmatter.split("\n")) {
    if (line.trim() === "") {
      continue;
    }
    if (/^\s/.test(line)) {
      continue;
    }

    const separator = line.indexOf(":");
    if (separator <= 0) {
      throw new Error(`invalid frontmatter line ${line}`);
    }

    const key = line.slice(0, separator).trim();
    const value = line.slice(separator + 1).trim();
    if (!/^[A-Za-z][A-Za-z0-9]*$/.test(key)) {
      throw new Error(`invalid frontmatter key ${key}`);
    }
    if (Object.hasOwn(metadata, key)) {
      throw new Error(`duplicate frontmatter key ${key}`);
    }

    metadata[key] = value;
  }

  return metadata;
}

async function collectMarkdownFiles(root, directory, files) {
  const entries = await readdir(directory, { withFileTypes: true });
  entries.sort((left, right) => left.name.localeCompare(right.name, "en"));

  for (const entry of entries) {
    const absolutePath = resolve(directory, entry.name);
    if (entry.isDirectory()) {
      await collectMarkdownFiles(root, absolutePath, files);
    } else if (entry.isFile() && entry.name.endsWith(".md")) {
      files.push(relative(root, absolutePath).split(sep).join("/"));
    }
  }
}

export async function markdownFiles(root) {
  const files = [];
  await collectMarkdownFiles(resolve(root), resolve(root), files);
  return files.sort((left, right) => left.localeCompare(right, "en"));
}

function validateMetadata(path, metadata, versions, errors) {
  for (const field of requiredFields) {
    if (!metadata[field]) {
      errors.push(`${path}: missing ${field}`);
    }
  }

  if (metadata.audience && !allowedAudiences.includes(metadata.audience)) {
    errors.push(`${path}: audience must be one of ${allowedAudiences.join(", ")}`);
  }
  if (metadata.status && !allowedStatuses.includes(metadata.status)) {
    errors.push(`${path}: status must be one of ${allowedStatuses.join(", ")}`);
  }
  if (metadata.lastReviewed && !/^\d{4}-\d{2}-\d{2}$/.test(metadata.lastReviewed)) {
    errors.push(`${path}: lastReviewed must use YYYY-MM-DD`);
  }

  const canonicalVersion = versions[metadata.status];
  if (canonicalVersion && metadata.reviewedVersion !== canonicalVersion) {
    errors.push(`${path}: reviewedVersion must be ${canonicalVersion} for ${metadata.status}`);
  }
}

export async function validateDocumentTree(root, versions) {
  const resolvedRoot = resolve(root);
  const files = await markdownFiles(resolvedRoot);
  const fileSet = new Set(files);
  const documents = new Map();
  const errors = [];

  for (const path of files) {
    const source = await readFile(resolve(resolvedRoot, ...path.split("/")), "utf8");
    try {
      const metadata = parseFrontmatter(source);
      documents.set(path, metadata);
      validateMetadata(path, metadata, versions, errors);

      const { body } = frontmatterParts(source);
      const marker = body.match(unfinishedMarkerPattern)?.[1];
      if (marker) {
        errors.push(`${path}: contains unfinished marker ${marker}`);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      errors.push(`${path}: ${message}`);
    }

    const counterpart = path.startsWith("de/") ? path.slice(3) : `de/${path}`;
    if (!fileSet.has(counterpart)) {
      errors.push(`${path}: missing counterpart ${counterpart}`);
    }
  }

  for (const path of files.filter((candidate) => !candidate.startsWith("de/"))) {
    const counterpart = `de/${path}`;
    const english = documents.get(path);
    const german = documents.get(counterpart);
    if (!english || !german) {
      continue;
    }

    for (const field of pairFields) {
      if (english[field] !== german[field]) {
        errors.push(`${path}: ${field} differs from ${counterpart}`);
      }
    }
  }

  return errors.sort((left, right) => left.localeCompare(right, "en"));
}

async function runCli() {
  const scriptDirectory = dirname(fileURLToPath(import.meta.url));
  const contentRoot = resolve(scriptDirectory, "../site");
  const versions = JSON.parse(await readFile(resolve(scriptDirectory, "../versions.json"), "utf8"));
  const errors = await validateDocumentTree(contentRoot, versions);

  for (const error of errors) {
    console.error(error);
  }
  if (errors.length > 0) {
    process.exitCode = 1;
  }
}

const invokedPath = process.argv[1] ? pathToFileURL(resolve(process.argv[1])).href : "";
if (invokedPath === import.meta.url) {
  await runCli();
}

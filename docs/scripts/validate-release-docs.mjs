import { createHash } from "node:crypto";
import { readFile, readdir } from "node:fs/promises";
import { resolve } from "node:path";

const languages = ["de", "en"];
const pngSignature = Buffer.from("89504e470d0a1a0a", "hex");

export async function validateReleaseDocs(repositoryRoot, versions, sources) {
  const errors = [];
  const read = async (path) => {
    try {
      return await readFile(resolve(repositoryRoot, path), "utf8");
    } catch {
      errors.push(`${path}: missing release documentation`);
      return "";
    }
  };
  const version = versions.stable;
  const runtime = await read("backend/cmd/railkeeper/main.go");
  const runtimeVersion = runtime.match(/^\s*version\s*=\s*"([^"]+)"/m)?.[1];
  if (!version || runtimeVersion !== version) {
    errors.push(`docs/versions.json: stable ${version} differs from runtime ${runtimeVersion}`);
  }
  const dates = [];
  for (const path of ["CHANGELOG.de.md", "CHANGELOG.md"]) {
    const source = await read(path);
    const entry = [...source.matchAll(/^## \[([^\]]+)\] - (\d{4}-\d{2}-\d{2})\n([\s\S]*?)(?=^## |$(?![\s\S]))/gm)]
      .filter((match) => match[1] === version);
    if (entry.length !== 1 || !/^\s*- \S/m.test(entry[0]?.[3] ?? "")) {
      errors.push(`${path}: needs one substantive [${version}] release entry`);
    } else {
      dates.push(entry[0][2]);
    }
  }
  if (dates.length === 2 && dates[0] !== dates[1]) {
    errors.push("changelogs: German and English release dates differ");
  }
  const release = await read(`docs/releases/v${version}.md`);
  if (!/^## Deutsch\s*$/m.test(release) || !/^## English\s*$/m.test(release)) {
    errors.push(`docs/releases/v${version}.md: needs German and English release notes`);
  }
  if (sources.schemaVersion !== 2 || sources.applicationVersion !== version ||
      sources.buildStatus !== "release" || !/^[a-f0-9]{40}$/.test(sources.applicationCommit ?? "")) {
    errors.push("screenshot-sources.json: needs provenance for the current stable release");
  }
  if (JSON.stringify(sources.languages) !== JSON.stringify(languages)) {
    errors.push("screenshot-sources.json: languages must be de and en");
  }
  const captures = sources.captures ?? {};
  const expected = new Set();
  const names = new Map();
  for (const language of languages) {
    const directory = resolve(repositoryRoot, "docs/site/assets/screenshots", language);
    let files = [];
    try {
      files = (await readdir(directory)).filter((file) => file.endsWith(".png")).sort();
    } catch {
      errors.push(`screenshots/${language}: missing image directory`);
    }
    if (files.length === 0) errors.push(`screenshots/${language}: no captured images`);
    names.set(language, files);
    for (const file of files) {
      const path = `${language}/${file}`;
      expected.add(path);
      const capture = captures[path];
      if (!capture) {
        errors.push(`${path}: missing screenshot provenance`);
        continue;
      }
      if (capture.applicationVersion !== version || capture.applicationCommit !== sources.applicationCommit) {
        errors.push(`${path}: capture belongs to a different application version or commit`);
      }
      if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/.test(capture.capturedAt ?? "") ||
          !Number.isFinite(Date.parse(capture.capturedAt))) {
        errors.push(`${path}: needs an actual capture timestamp`);
      }
      const image = await readFile(resolve(directory, file));
      const validPng = image.length >= 24 && image.subarray(0, 8).equals(pngSignature);
      if (!validPng || image.readUInt32BE(16) !== capture.width || image.readUInt32BE(20) !== capture.height) {
        errors.push(`${path}: PNG dimensions differ from capture provenance`);
      }
      if (createHash("sha256").update(image).digest("hex") !== capture.sha256) {
        errors.push(`${path}: image SHA256 differs from capture provenance`);
      }
      if (capture.simulation && !["command-station", "article-search"].includes(capture.simulation)) {
        errors.push(`${path}: unknown external simulation type`);
      }
      if (file === "account-two-factor-setup.png" && capture.privacyMask !== true) {
        errors.push(`${path}: two-factor setup requires a capture-time privacy mask`);
      }
    }
  }
  if (JSON.stringify(names.get("de")) !== JSON.stringify(names.get("en"))) {
    errors.push("screenshots: German and English image sets differ");
  }
  for (const path of Object.keys(captures)) {
    if (!expected.has(path)) errors.push(`${path}: provenance references a missing screenshot`);
  }
  return errors.sort();
}

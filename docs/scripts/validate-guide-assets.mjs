import { existsSync } from "node:fs";
import { readFile } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";

export async function validateGuideAssets(contentRoot, files, topics) {
  const errors = [];
  const imageNames = new Map();
  const downloadNames = new Map();
  const root = resolve(contentRoot);
  for (const file of files) {
    const source = await readFile(resolve(root, file), "utf8");
    const names = new Set();
    const images = [...source.matchAll(/!\[([^\]]*)\]\(([^\s)]+)(?:\s+"[^"]*")?\)/g)];
    for (const figure of source.matchAll(/<GuideFigure\b[^>]*\/>/g)) {
      const name = figure[0].match(/\bname="([^"]*)"/)?.[1] || "";
      const alt = figure[0].match(/\balt="([^"]*)"/)?.[1] || "";
      const language = file.startsWith("de/") ? "de" : "en";
      const target = relative(dirname(resolve(root, file)), resolve(root, "assets/screenshots", language, `${name}.png`));
      images.push([figure[0], alt, target.split(sep).join("/")]);
    }
    for (const [, alt, target] of images) {
      if (!alt.trim()) errors.push(`${file}: image needs descriptive alternative text`);
      if (/^https?:\/\//.test(target)) continue;
      const path = resolve(dirname(resolve(root, file)), decodeURIComponent(target));
      if (!path.startsWith(`${root}${sep}`) || !existsSync(path)) {
        errors.push(`${file}: missing or out-of-tree image ${target}`);
      }
      if (target.includes("assets/screenshots/")) {
        const language = file.startsWith("de/") ? "de" : "en";
        if (!target.includes(`assets/screenshots/${language}/`)) {
          errors.push(`${file}: screenshot must use ${language} interface`);
        }
        names.add(target.split("/").at(-1));
      }
    }
    if (/^(de\/)?guide\//.test(file) && names.size === 0) {
      errors.push(`${file}: user guide needs a localized interface picture`);
    }
    imageNames.set(file, [...names].sort().join(","));
    const downloads = new Set();
    for (const link of source.matchAll(/<GuideDownload\b([^>]*)>([\s\S]*?)<\/GuideDownload>/g)) {
      const name = link[1].match(/\bfile="([^"]*)"/)?.[1] || "";
      const label = link[2].trim();
      if (!label || label.includes("<") || label.includes(">")) {
        errors.push(`${file}: download needs plain descriptive link text`);
      }
      if (!/^[a-z0-9-]+\.(csv|json|txt|png)$/.test(name) ||
        !existsSync(resolve(root, "assets/examples", name))) {
        errors.push(`${file}: missing or invalid example download ${name}`);
      }
      downloads.add(name);
    }
    downloadNames.set(file, [...downloads].sort().join(","));
  }
  for (const file of files.filter((path) => !path.startsWith("de/"))) {
    if (downloadNames.get(file) !== downloadNames.get(`de/${file}`)) {
      errors.push(`${file}: example downloads differ from German counterpart`);
    }
    if (imageNames.get(file) !== imageNames.get(`de/${file}`)) {
      errors.push(`${file}: screenshots differ from German counterpart`);
    }
  }
  for (const topic of topics) {
    if (!topic.labelDE?.trim() || !topic.labelEN?.trim()) {
      errors.push(`coverage topic ${topic.id}: missing reader labels`);
    }
    if (!["demo-ui", "source"].includes(topic.reviewEvidence) || !/^\d{4}-\d{2}-\d{2}$/.test(topic.reviewedAt || "")) {
      errors.push(`coverage topic ${topic.id}: missing review evidence or date`);
    }
    if (topic.reviewEvidence === "demo-ui") {
      for (const language of ["de", "en"]) {
        if (!topic.screenshot || !existsSync(resolve(root, "assets/screenshots", language, `${topic.screenshot}.png`))) {
          errors.push(`coverage topic ${topic.id}: missing ${language} demo evidence`);
        }
      }
    }
  }
  return errors.sort();
}

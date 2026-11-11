import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { gzipSync } from "node:zlib";

const output = resolve(process.argv[2] ?? fileURLToPath(new URL("../dist", import.meta.url)));
const html = await readFile(resolve(output, "index.html"), "utf8");
const assetTags = html.matchAll(/<(?:script|link)\b[^>]*\s(?:src|href)="\/(assets\/[^"?#]+\.(?:js|css))"[^>]*>/g);
const assets = new Set(Array.from(assetTags, ([, path]) => path));
const budgets = { js: 250 * 1024, css: 75 * 1024 };
const sizes = { js: 0, css: 0 };

for (const asset of assets) {
  const type = asset.endsWith(".js") ? "js" : "css";
  sizes[type] += gzipSync(await readFile(resolve(output, asset))).length;
}

function warn(message) {
  const annotation = process.env.GITHUB_ACTIONS === "true" ? "::warning title=Frontend performance budget::" : "";
  console.warn(`${annotation}${message}`);
}

for (const type of ["js", "css"]) {
  const label = type.toUpperCase();
  const limit = budgets[type] / 1024;
  const size = (sizes[type] / 1024).toFixed(1);
  if (sizes[type] === 0) {
    warn(`No initial ${label} assets found in index.html`);
  } else if (sizes[type] > budgets[type]) {
    warn(`Initial ${label} uses ${size} KiB gzip and exceeds the ${limit} KiB budget`);
  } else {
    console.log(`Initial ${label} uses ${size} of ${limit} KiB gzip`);
  }
}

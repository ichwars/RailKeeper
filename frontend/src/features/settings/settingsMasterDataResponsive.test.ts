import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

describe("master-data table responsiveness", () => {
  it("lets the table scroll inside the bounded settings panel", () => {
    const css = readFileSync(resolve(process.cwd(), "src/styles/settings.css"), "utf8");

    expect(css).toMatch(/\.master-data-panel\s*\{[^}]*min-width:\s*0/s);
    expect(css).toMatch(/\.master-data-create\s*\{[^}]*min-width:\s*0[^}]*overflow-x:\s*auto/s);
    expect(css).toMatch(/\.master-data-table\s*\{[^}]*min-width:\s*0[^}]*width:\s*100%/s);
    expect(css).toMatch(/\.master-data-table \.master-data-col-actions\s*\{[^}]*position:\s*sticky[^}]*right:\s*0/s);
    expect(css).toMatch(/\.master-data-table table\s*\{[^}]*overflow:\s*visible/s);
  });
});

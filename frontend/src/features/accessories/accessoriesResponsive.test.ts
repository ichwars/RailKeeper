import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const responsiveCss = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");

describe("article editor responsive tabs", () => {
  it("restores horizontally reachable tabs in the later mobile override stylesheet", () => {
    expect(responsiveCss).toMatch(/\.article-editor-tabs\s*\{[^}]*overflow-x:\s*auto/s);
    expect(responsiveCss).toMatch(/\.article-editor-tabs button\s*\{[^}]*flex:\s*0 0 auto/s);
  });
});

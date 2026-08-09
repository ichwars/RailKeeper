import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const responsiveCss = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");
const accessoriesCss = readFileSync(resolve(process.cwd(), "src/styles/accessories.css"), "utf8");

describe("article editor responsive tabs", () => {
  it("restores horizontally reachable tabs in the later mobile override stylesheet", () => {
    expect(responsiveCss).toMatch(/\.article-editor-tabs\s*\{[^}]*overflow-x:\s*auto/s);
    expect(responsiveCss).toMatch(/\.article-editor-tabs button\s*\{[^}]*flex:\s*0 0 auto/s);
  });

  it("keeps the eleven-column article table horizontally reachable at every width", () => {
    expect(accessoriesCss).toMatch(/\.article-table\s*\{[^}]*min-width:\s*14[0-9]{2}px/s);
    expect(accessoriesCss).toMatch(/\.article-table th\.actions-cell,[\s\S]*?width:\s*9[0-9]px/s);
    expect(accessoriesCss).toMatch(/\.article-table-wrap\s*\{[^}]*overflow-x:\s*auto/s);
    expect(accessoriesCss).toMatch(/\.article-table \.select-cell\s*\{[^}]*width:\s*4[0-9]px/s);
  });

  it("keeps stacked article filters compact on narrow screens", () => {
    expect(accessoriesCss).toMatch(
      /@media\s*\(max-width:\s*560px\)[\s\S]*?\.article-filter-select,[\s\S]*?\.article-location-filter\s*\{[^}]*flex:\s*0 0 auto/s,
    );
  });

  it("keeps long subject labels shrinkable and stacks the subject grid at narrow widths", () => {
    expect(accessoriesCss).toMatch(/\.article-subject-grid\s*>\s*\.app-field\s*\{[^}]*min-width:\s*0/s);
    expect(accessoriesCss).toMatch(/\.article-subject-grid[^}]*overflow-wrap:\s*anywhere/s);
    expect(accessoriesCss).toMatch(/@media\s*\(max-width:\s*560px\)[\s\S]*\.article-editor-grid,[^}]*grid-template-columns:\s*1fr/s);
  });

  it("uses a two-column allocation grid on desktop and stacks it below 920 pixels", () => {
    expect(accessoriesCss).toMatch(
      /\.accessory-allocation-form\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s,
    );
    expect(accessoriesCss).toMatch(
      /@media\s*\(max-width:\s*920px\)[\s\S]*\.accessory-allocation-form\s*\{[^}]*grid-template-columns:\s*1fr/s,
    );
  });
});

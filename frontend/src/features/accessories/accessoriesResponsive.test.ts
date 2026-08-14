import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const responsiveCss = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");
const accessoriesCss = readFileSync(resolve(process.cwd(), "src/styles/accessories.css"), "utf8");
const articleOverviewCss = readFileSync(resolve(process.cwd(), "src/styles/article-overview.css"), "utf8");
const appCss = readFileSync(resolve(process.cwd(), "src/app/styles.css"), "utf8");
const vehicleDialogsCss = readFileSync(resolve(process.cwd(), "src/styles/vehicle-dialogs.css"), "utf8");

function zIndex(css: string, selector: string) {
  const match = css.match(new RegExp(`${selector}\\s*\\{[^}]*z-index:\\s*(\\d+)`, "s"));
  return match ? Number(match[1]) : -1;
}

describe("article editor responsive tabs", () => {
  it("restores horizontally reachable tabs in the later mobile override stylesheet", () => {
    expect(responsiveCss).toMatch(/\.article-editor-tabs\s*\{[^}]*overflow-x:\s*auto/s);
    expect(responsiveCss).toMatch(/\.article-editor-tabs button\s*\{[^}]*flex:\s*0 0 auto/s);
  });

  it("keeps the configurable article table horizontally reachable at every width", () => {
    expect(accessoriesCss).toMatch(/\.article-table\s*\{[^}]*min-width:\s*var\(--article-table-min-width\)/s);
    expect(accessoriesCss).toMatch(/\.article-table th\.actions-cell,[\s\S]*?width:\s*13[0-9]px/s);
    expect(accessoriesCss).toMatch(/\.article-table-wrap\s*\{[^}]*overflow-x:\s*auto/s);
    expect(accessoriesCss).toMatch(/\.article-table \.select-cell\s*\{[^}]*width:\s*4[0-9]px/s);
    expect(accessoriesCss).not.toMatch(/\.article-table th:nth-child/);
  });

  it("loads focused overview styles and switches presentations at the mobile breakpoint", () => {
    expect(appCss).toContain('@import "../styles/article-overview.css";');
    expect(articleOverviewCss).toMatch(/\.article-mobile-list\s*\{[^}]*display:\s*none/s);
    expect(articleOverviewCss).toMatch(
      /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-desktop-content\s*\{[^}]*display:\s*none/s
    );
    expect(articleOverviewCss).toMatch(
      /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-mobile-list\s*\{[^}]*display:\s*grid/s
    );
    expect(articleOverviewCss).toMatch(
      /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-view-tools\s*\{[^}]*display:\s*none/s
    );
  });

  it("keeps long card metadata labels separated in a two-column grid", () => {
    expect(articleOverviewCss).toMatch(
      /\.article-card-body dl\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s
    );
    expect(articleOverviewCss).toMatch(
      /\.article-card-body dl div:nth-child\(3\)\s*\{[^}]*grid-column:\s*1 \/ -1/s
    );
  });

  it("positions the column picker and removes obsolete action-menu styles", () => {
    expect(articleOverviewCss).toMatch(/\.article-column-picker\s*\{[^}]*position:\s*relative/s);
    expect(articleOverviewCss).toMatch(
      /\.article-column-picker-popover\s*\{[^}]*position:\s*absolute/s
    );
    expect(articleOverviewCss).not.toMatch(/\.article-action-menu/);
    expect(accessoriesCss).not.toMatch(/\.article-action-menu/);
    expect(accessoriesCss).not.toMatch(/\.article-overflow/);
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

  it("uses the compact weighted stock grid and stacks transfer fields on narrow screens", () => {
    expect(accessoriesCss).toMatch(
      /\.article-stock-commands\s*\{[^}]*grid-template-columns:\s*minmax\(240px,\s*1fr\)\s*minmax\(0,\s*2fr\)/s,
    );
    expect(accessoriesCss).toMatch(
      /\.article-transfer-fields\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s,
    );
    expect(accessoriesCss).toMatch(
      /@media\s*\(max-width:\s*560px\)[\s\S]*\.article-transfer-fields\s*\{[^}]*grid-template-columns:\s*1fr/s,
    );
    expect(accessoriesCss).not.toMatch(/\.article-stock-form\s*\{[^}]*height:\s*100%/s);
    expect(accessoriesCss).not.toContain(".article-stock-form > .primary-button");
  });

  it("stacks article and barcode search dialogs above the article editor", () => {
    const editorLayer = zIndex(accessoriesCss, "\\.article-editor-layer");

    expect(zIndex(vehicleDialogsCss, "\\.article-search-layer")).toBeGreaterThan(editorLayer);
    expect(zIndex(vehicleDialogsCss, "\\.barcode-search-layer")).toBeGreaterThan(editorLayer);
  });
});

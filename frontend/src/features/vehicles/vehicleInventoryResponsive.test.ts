import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/styles/vehicle-inventory.css"), "utf8");
const appCss = readFileSync(resolve(process.cwd(), "src/app/styles.css"), "utf8");

describe("vehicle inventory column layout", () => {
  it("imports a focused vehicle inventory stylesheet", () => {
    expect(appCss).toContain('@import "../styles/vehicle-inventory.css";');
  });

  it("bounds the long picker and uses a fixed mobile sheet", () => {
    expect(css).toMatch(/\.vehicle-column-picker-popover\s*\{[^}]*width:\s*min\(680px,\s*calc\(100vw\s*-\s*32px\)\)/s);
    expect(css).toMatch(/@media\s*\(max-width:\s*720px\)[\s\S]*\.vehicle-column-picker-popover\s*\{[^}]*position:\s*fixed[^}]*inset:\s*auto\s+12px\s+12px/s);
  });

  it("uses an expandable compact mobile card with a narrow fallback", () => {
    expect(css).toMatch(/\.vehicle-mobile-toggle\s*\{[^}]*grid-template-columns:\s*64px\s+minmax\(0,\s*1fr\)\s+24px/s);
    expect(css).toMatch(/\.vehicle-mobile-fields\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s);
    expect(css).toMatch(/@media\s*\(max-width:\s*420px\)[\s\S]*\.vehicle-mobile-fields\s*\{[^}]*grid-template-columns:\s*1fr/s);
  });

  it("keeps open mobile quick menus above following cards", () => {
    expect(css).toMatch(/\.vehicle-mobile-item:has\(\.quick-menu\)\s*\{[^}]*z-index:\s*5/s);
    expect(css).toMatch(/\.vehicle-mobile-item\s*\{[^}]*overflow:\s*visible/s);
    expect(css).toMatch(/\.quick-menu\.open-above\s*\{[^}]*top:\s*auto[^}]*bottom:\s*calc\(100%\s*\+\s*6px\)/s);
  });

  it("clips long data cells without clipping the desktop action menu", () => {
    expect(css).toMatch(/\.vehicle-inventory-table td:not\(\.actions-cell\)\s*\{[^}]*overflow:\s*hidden/s);
    expect(css).toMatch(/\.vehicle-inventory-table td\.actions-cell\s*\{[^}]*overflow:\s*visible/s);
    expect(css).toMatch(/\.vehicle-inventory-table tbody tr:has\(\.quick-menu\)\s*\{[^}]*z-index:\s*5/s);
  });
});

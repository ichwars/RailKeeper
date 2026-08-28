import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/styles/vehicle-inventory.css"), "utf8");
const appCss = readFileSync(resolve(process.cwd(), "src/app/styles.css"), "utf8");
const responsiveCss = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");
const inventoryResponsiveCss = readFileSync(
  resolve(process.cwd(), "src/styles/vehicle-inventory-responsive.css"),
  "utf8",
);
const inventoryPanelSource = readFileSync(
  resolve(process.cwd(), "src/features/vehicles/VehicleInventoryPanel.tsx"),
  "utf8",
);

describe("vehicle inventory column layout", () => {
  it("imports a focused vehicle inventory stylesheet", () => {
    expect(appCss).toContain('@import "../styles/vehicle-inventory.css";');
    expect(appCss).toContain('@import "../styles/vehicle-inventory-responsive.css";');
    expect(appCss.indexOf('@import "../styles/overrides-responsive.css";')).toBeLessThan(
      appCss.indexOf('@import "../styles/vehicle-inventory-responsive.css";'),
    );
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

	it("uses a one-column set hierarchy with accessible action targets", () => {
		expect(css).toMatch(/\.vehicle-mobile-set-card\s*\{[^}]*grid-template-columns:\s*1fr/s);
		expect(css).toMatch(/\.vehicle-mobile-set-actions button\s*\{[^}]*min-height:\s*44px/s);
		expect(css).toMatch(/\.vehicle-mobile-set-actions button\s*\{[^}]*min-width:\s*44px/s);
	});

  it("keeps mobile quick menus inside the viewport", () => {
    expect(css).toMatch(/\.quick-menu\.quick-menu-floating\s*\{[^}]*position:\s*fixed/s);
    expect(css).toMatch(/\.quick-menu\.quick-menu-floating\s*\{[^}]*max-width:\s*calc\(100vw\s*-\s*16px\)/s);
    expect(css).toMatch(/\.quick-menu\.quick-menu-floating\s*\{[^}]*overflow-y:\s*auto/s);
  });

  it("clips long data cells without clipping the desktop action menu", () => {
    expect(css).toMatch(/\.vehicle-inventory-table td:not\(\.actions-cell\)\s*\{[^}]*overflow:\s*hidden/s);
    expect(css).toMatch(/\.vehicle-inventory-table td\.actions-cell\s*\{[^}]*overflow:\s*visible/s);
    expect(css).toMatch(/\.vehicle-inventory-table tbody tr:has\(\.quick-menu\)\s*\{[^}]*z-index:\s*5/s);
  });

  it("fills the desktop wrapper but preserves configured overflow widths", () => {
    expect(css).toMatch(/\.vehicle-inventory-table\s*\{[^}]*width:\s*max\(100%,\s*var\(--vehicle-table-min-width\)\)[^}]*min-width:\s*var\(--vehicle-table-min-width\)[^}]*table-layout:\s*fixed/s);
    expect(css).toMatch(/\.vehicle-inventory-table th:not\(\.select-cell\):not\(\.actions-cell\)\s*\{[^}]*position:\s*relative/s);
    expect(css).not.toMatch(/--vehicle-data-column-count,\s*8/);
  });

  it("keeps the next appointment card wide until the status row stacks", () => {
    expect(inventoryResponsiveCss).toMatch(
      /@media\s*\(max-width:\s*640px\)[\s\S]*?\.inventory-status-row\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)[\s\S]*?\.inventory-status-card\.wide\s*\{[^}]*grid-column:\s*1\s*\/\s*-1/s,
    );
    expect(inventoryResponsiveCss).toMatch(
      /@media\s*\(max-width:\s*420px\)[\s\S]*?\.inventory-status-row\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/s,
    );
  });

  it("places four compact status cards before the wide next appointment card", () => {
    expect(inventoryPanelSource.indexOf('t("vehicles.imageCare")')).toBeLessThan(
      inventoryPanelSource.indexOf('t("vehicles.nextAppointment")'),
    );
  });

  it("contains the mobile filter scroller without clipping its controls", () => {
    expect(inventoryResponsiveCss).toMatch(
      /@media\s*\(max-width:\s*640px\)[\s\S]*?\.inventory-filter-row\s*\{[^}]*min-width:\s*0[^}]*max-width:\s*100%[^}]*overflow-x:\s*auto[^}]*overscroll-behavior-x:\s*contain[^}]*scrollbar-gutter:\s*stable[^}]*scroll-padding-inline:\s*2px/s,
    );
  });

  it("keeps inventory breakpoints out of the shared responsive layer", () => {
    [
      ".inventory-head",
      ".inventory-status-row",
      ".inventory-toolbar",
      ".inventory-filter-row",
      ".inventory-mobile-list",
      ".inventory-desktop-content",
      ".inventory-card {",
      ".inventory-card-media",
    ].forEach((selector) => expect(responsiveCss).not.toContain(selector));
  });
});

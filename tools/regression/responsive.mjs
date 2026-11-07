import assert from "node:assert/strict";
import { assertPersonalViews } from "./personal-views.mjs";

// Exercise real controls: invisible native switch inputs must stay inside the scrollable table,
// and the visible maintenance selector must never cover the image-upload action.
export async function assertResponsiveWorkspaces(page, { base, username, width, de, screenshotPrefix }) {
  await assertInventoryFilters(page, { width, de, screenshotPrefix });
  await page.goto(base + "/settings", { waitUntil: "domcontentloaded" });
  await page.locator(".settings-dashboard-grid").waitFor();
  await assertPersonalViews(page, { base, username, width, de, screenshotPrefix });
  await page.evaluate(() => document.fonts.ready);
  await assertViewportFits(page, "general settings");
  const toggle = page.locator('.settings-inline-table input[type="checkbox"]').first();
  await toggle.focus();
  await page.keyboard.press("Tab");
  await page.keyboard.press("Shift+Tab");
  assert.equal(await toggle.evaluate(input => document.activeElement === input), true);
  assert.equal(await toggle.evaluate(input => getComputedStyle(input.nextElementSibling).outlineStyle), "solid");
  await assertViewportFits(page, "focused inventory-number switch");
  await page.screenshot({ path: screenshotPrefix + "-settings.png", fullPage: true });

  // Open the existing table action before inspecting the same dialog at the requested width.
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(base + "/vehicles", { waitUntil: "domcontentloaded" });
  const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: "Referenzlok 1", exact: true }) });
  await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  const dialog = page.getByRole("dialog");
  await dialog.locator("nav.modal-tabs").getByRole("button", { name: "Uploads", exact: true }).click();
  await page.setViewportSize({ width, height: 1000 });
  await page.evaluate(() => document.fonts.ready);
  const controls = await page.locator(".uploads-tab .upload-head").first().evaluate(head => {
    const selector = head.querySelector(".upload-maintenance-select .app-select-trigger");
    const upload = head.querySelector(".primary-button");
    const a = selector.getBoundingClientRect(), b = upload.getBoundingClientRect();
    const bounds = head.getBoundingClientRect();
    return {
      overlap: Math.max(0, Math.min(a.right, b.right) - Math.max(a.left, b.left)) *
        Math.max(0, Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top)),
      contained: [a, b].every(r => r.left >= bounds.left - 1 && r.right <= bounds.right + 1),
      selectorHeight: a.height, uploadHeight: b.height
    };
  });
  assert.equal(controls.overlap, 0, "maintenance selector overlaps the image-upload button");
  assert.equal(controls.contained, true, "upload actions extend past the visible section");
  if (width <= 600) {
    assert.ok(controls.selectorHeight >= 44, "mobile maintenance selector is too small");
    assert.ok(controls.uploadHeight >= 44, "mobile image-upload button is too small");
  }
  await assertViewportFits(page, "vehicle uploads");
  await page.screenshot({ path: screenshotPrefix + "-uploads.png", fullPage: true });
  await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
}

async function assertInventoryFilters(page, { width, de, screenshotPrefix }) {
  const more = page.locator(".inventory-more-toggle");
  await more.waitFor();
  await assertInventoryControlGeometry(page, width);
  await more.click();
  assert.equal(await more.getAttribute("aria-expanded"), "true");
  await assertInventoryControlGeometry(page, width);
  await page.screenshot({ path: screenshotPrefix + "-filters-open.png", fullPage: true });

  const quality = page.getByRole("button", { name: de ? "Datenprüfung" : "Data checks", exact: true });
  await quality.click();
  const option = page.getByRole("option").nth(1);
  const selectedLabel = await option.innerText();
  await option.click();
  await more.click();
  assert.equal(await more.getAttribute("aria-expanded"), "false");
  assert.equal(await quality.count(), 0);
  const chip = page.getByRole("button", {
    name: (de ? "Filter entfernen: " : "Remove filter: ") + selectedLabel, exact: true
  });
  assert.equal(await chip.isVisible(), true, "collapsed active filter must remain visible");
  await assertInventoryControlGeometry(page, width);
  await chip.click();
  assert.equal(await more.evaluate(button => button === document.activeElement), true,
    "clearing a disappearing chip must retain keyboard focus");
  assert.equal(await chip.count(), 0);

  const pickerTrigger = page.getByRole("button", { name: de ? "Tabellenspalten auswählen" : "Choose table columns", exact: true });
  await pickerTrigger.click();
  const picker = page.getByRole("dialog", { name: de ? "Fahrzeugspalten konfigurieren" : "Configure vehicle columns", exact: true });
  const compact = picker.getByRole("button", { name: de ? "Kompakte Spalten" : "Compact columns", exact: true });
  await compact.scrollIntoViewIfNeeded();
  const bounds = await compact.boundingBox();
  assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1,
    "compact column action must be inside the viewport");
  assert.equal(await compact.evaluate(button => {
    const r = button.getBoundingClientRect();
    return button.contains(document.elementFromPoint(r.x + r.width / 2, r.y + r.height / 2));
  }), true, "compact column action must be reachable above other layers");
  if (width === 1280) {
    const [saved] = await Promise.all([
      page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
      compact.click()
    ]);
    assert.equal(saved.status(), 200);
    assert.equal(await page.locator(".vehicle-inventory-table col[data-column]").count(), 6);
  }
  await pickerTrigger.focus();
  await page.keyboard.press("Escape");
  await picker.waitFor({ state: "hidden" });
  assert.equal(await pickerTrigger.evaluate(button => button === document.activeElement), true);
  if (width >= 1024) {
    const table = page.getByRole("region", { name: de ? "Fahrzeugtabelle, seitlich scrollbar" : "Vehicle table, scrolls sideways", exact: true });
    await table.focus();
    assert.equal(await table.evaluate(element => element === document.activeElement), true);
    const maximumScroll = await table.evaluate(element => element.scrollWidth - element.clientWidth);
    if (maximumScroll > 0) {
      await page.keyboard.press("ArrowRight");
      await page.waitForFunction(() => document.querySelector(".inventory-table-scroll-region").scrollLeft > 0);
      await table.evaluate(element => { element.scrollLeft = 0; });
    }
  }
  await assertViewportFits(page, "visible inventory filters and column controls");
  await page.screenshot({ path: screenshotPrefix + "-filters-closed.png", fullPage: true });
}

async function assertInventoryControlGeometry(page, width) {
  const fits = await page.locator(".inventory-filters").evaluate(root => root.scrollWidth <= root.clientWidth + 1);
  assert.equal(fits, true, "filter controls must reflow within their own horizontal bounds");
  const buttons = await page.locator(".inventory-filters button").all();
  assert.ok(buttons.length >= 11, "inventory quick filters and field actions must all remain present");
  for (const button of buttons) {
    // The app scrolls its main pane vertically; reach each control before checking clipping.
    await button.scrollIntoViewIfNeeded();
    const control = await button.evaluate(button => {
    const r = button.getBoundingClientRect();
    const outside = [];
    for (let ancestor = button.parentElement; ancestor; ancestor = ancestor.parentElement) {
      const style = getComputedStyle(ancestor), bounds = ancestor.getBoundingClientRect();
      if (["hidden", "clip", "auto", "scroll"].includes(style.overflowX) &&
          (r.left < bounds.left + ancestor.clientLeft - 1 ||
           r.right > bounds.left + ancestor.clientLeft + ancestor.clientWidth + 1)) outside.push({ tag: ancestor.tagName, class: ancestor.className, axis: "x" });
      if (["hidden", "clip"].includes(style.overflowY) &&
          (r.top < bounds.top + ancestor.clientTop - 1 ||
           r.bottom > bounds.top + ancestor.clientTop + ancestor.clientHeight + 1)) outside.push({ tag: ancestor.tagName, class: ancestor.className, axis: "y" });
    }
    return { label: button.getAttribute("aria-label") || button.innerText, x: r.x, width: r.width, height: r.height, outside };
    });
    assert.ok(control.width > 0 && control.height > 0 && control.x >= -1 && control.x + control.width <= width + 1,
      `inventory control outside visible viewport: ${control.label}`);
    assert.deepEqual(control.outside, [], `clipped inventory control: ${control.label}`);
    if (width <= 900) assert.ok(control.height >= 44, `small mobile inventory control: ${control.label}`);
  }
}

async function assertViewportFits(page, area) {
  const dimensions = await page.evaluate(() => ({ viewport: innerWidth, content: document.documentElement.scrollWidth }));
  assert.ok(dimensions.content <= dimensions.viewport + 1,
    `${area}: content ${dimensions.content}px exceeds viewport ${dimensions.viewport}px`);
}

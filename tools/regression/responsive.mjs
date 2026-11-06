import assert from "node:assert/strict";

// Exercise real controls: invisible native switch inputs must stay inside the scrollable table,
// and the visible maintenance selector must never cover the image-upload action.
export async function assertResponsiveWorkspaces(page, { base, width, de, screenshotPrefix }) {
  await page.goto(base + "/settings", { waitUntil: "networkidle" });
  await page.locator(".settings-dashboard-grid").waitFor();
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
  await page.goto(base + "/vehicles", { waitUntil: "networkidle" });
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

async function assertViewportFits(page, area) {
  const dimensions = await page.evaluate(() => ({ viewport: innerWidth, content: document.documentElement.scrollWidth }));
  assert.ok(dimensions.content <= dimensions.viewport + 1,
    `${area}: content ${dimensions.content}px exceeds viewport ${dimensions.viewport}px`);
}

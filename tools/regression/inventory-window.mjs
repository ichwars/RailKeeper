import assert from "node:assert/strict";

const windowedTable = ".vehicle-inventory-table[aria-rowcount]";
const mountedRows = `${windowedTable} tbody tr[data-inventory-row]`;

async function scrollInventory(page, fraction) {
  await page.locator("main").evaluate((main, value) => {
    main.scrollTop = (main.scrollHeight - main.clientHeight) * value;
  }, fraction);
}

async function checkTabAtWindowEdge(page, reverse) {
  await page.locator("main").evaluate(main => main.focus({ preventScroll: true }));
  await scrollInventory(page, reverse ? 0.68 : 0.38);
  await page.waitForFunction(({ selector, table }) => {
    const rows = Array.from(document.querySelectorAll(selector));
    const first = Number(rows[0]?.getAttribute("aria-rowindex"));
    const last = Number(rows.at(-1)?.getAttribute("aria-rowindex"));
    const total = Number(document.querySelector(table)?.getAttribute("aria-rowcount"));
    return first > 2 && last < total;
  }, { selector: mountedRows, table: windowedTable });
  await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
  const index = await page.evaluate(({ selector, reverse }) => {
    const rows = [...document.querySelectorAll(selector)];
    const row = reverse ? rows[0] : rows.at(-1);
    const stops = [...row.querySelectorAll("button,input")]
      .filter(element => element.tabIndex >= 0 && !element.disabled && element.getClientRects().length > 0);
    const destination = reverse ? stops[0] : stops.at(-1);
    destination.focus({ preventScroll: true });
    return Number(row.getAttribute("aria-rowindex"));
  }, { selector: mountedRows, reverse });
  const next = index + (reverse ? -1 : 1);
  assert.equal(await page.locator(`${mountedRows}[aria-rowindex="${next}"]`).count(), 0,
    "The adjacent logical row must start outside the rendered window");
  await page.keyboard.press(reverse ? "Shift+Tab" : "Tab");
  assert.equal(await page.evaluate(() => Number(document.activeElement.closest("tr")?.getAttribute("aria-rowindex"))),
    next, "Tab must reach the adjacent logical row across the window edge");
}

export async function checkInventoryWindow(page, imageCount) {
  await page.setViewportSize({ width: 1440, height: 900 });
  const fast = page.getByRole("button", { name: "Schnelle Tabellenansicht", exact: true });
  const full = page.getByRole("button", { name: "Tabellenansicht", exact: true });
  const windowed = page.locator(windowedTable);
  const first = page.getByRole("row").filter({ hasText: "Performance vehicle 0001" }).getByRole("checkbox");
  const last = page.getByRole("row").filter({ hasText: `Performance vehicle ${String(imageCount).padStart(4, "0")}` })
    .getByRole("checkbox");
  const tableSet = page.locator(".vehicle-set-hierarchy-toggle");

  await full.waitFor();
  assert.equal(await full.getAttribute("aria-pressed"), "true", "Full table is the default");
  await fast.press("Space");
  await windowed.waitFor();
  assert.equal(await fast.getAttribute("aria-pressed"), "true");
  assert.equal(await windowed.getAttribute("aria-rowcount"), String(imageCount + 4));
  assert.ok(await page.locator(mountedRows).count() < imageCount / 2, "Render only the window and overscan");
  assert.ok(await first.isChecked(), "Switching table mode must preserve the existing selection");

  await first.focus();
  await scrollInventory(page, 1);
  await tableSet.waitFor();
  assert.ok(await first.evaluate(input => input === document.activeElement), "Scrolling away must preserve row focus");
  assert.equal(await tableSet.getAttribute("aria-expanded"), "true", "Expanded sets must survive virtualization");
  const members = page.locator(".vehicle-set-child-row .inventory-name-link");
  assert.deepEqual(await members.allTextContents(), ["Performance set A", "Performance set B"]);
  await last.press("Space");
  assert.ok(await last.isChecked());
  await page.locator(".vehicle-set-inventory-row").getByRole("checkbox").press("Space");
  const reportButton = page.getByRole("button", { name: "Report erstellen", exact: true });
  await reportButton.click();
  const report = page.getByRole("dialog", { name: "Report erstellen" });
  await report.getByRole("radio", { name: "Markierung (4)", exact: true }).waitFor();
  await report.getByRole("button", { name: "Abbrechen", exact: true }).click();

  await checkTabAtWindowEdge(page, false);
  await checkTabAtWindowEdge(page, true);

  await full.press("Space");
  await page.locator(".vehicle-inventory-table:not([aria-rowcount])").waitFor();
  assert.equal(await page.locator(".vehicle-inventory-table tbody tr").count(), imageCount + 3);
  assert.ok(await first.isChecked() && await last.isChecked(), "Full view must retain selections across the window");
  assert.equal(await tableSet.getAttribute("aria-expanded"), "true");
  await fast.press("Space");
  await page.reload();
  await windowed.waitFor();
  assert.equal(await fast.getAttribute("aria-pressed"), "true", "Reload must restore the opt-in desktop mode");

  await page.setViewportSize({ width: 820, height: 900 });
  await page.locator(".inventory-mobile-list").waitFor();
  assert.equal(await fast.count(), 0, "The fast table is a desktop control");
  assert.equal(await windowed.count(), 0);
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
  await page.setViewportSize({ width: 2580, height: 900 });
  await windowed.waitFor();
  assert.equal(await fast.getAttribute("aria-pressed"), "true");
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
  console.log("Optional inventory window, cross-window selection, sets, focus and full table fallback passed");
}

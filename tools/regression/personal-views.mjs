import assert from "node:assert/strict";

export async function assertPersonalViews(page, { base, username, width, de, screenshotPrefix }) {
  if (width !== 1280 && width !== 390) return;
  const key = `railkeeper.settings.sidebarPrefs:${username}`;
  const previous = await page.evaluate(key => JSON.parse(localStorage.getItem(key) || "null"), key);
  const choice = page.getByRole("button", { name: de ? "Sammlung" : "Collection", exact: true });
  const [saved] = await Promise.all([
    page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
    choice.click()
  ]);
  assert.equal(saved.status(), 200);
  assert.equal(await choice.getAttribute("aria-pressed"), "true");
  const current = await page.evaluate(key => JSON.parse(localStorage.getItem(key)), key);
  if (previous) assert.deepEqual(current.order, previous.order, "preset must retain custom ordering");
  assert.deepEqual(current.hidden, ["exhibition", "importExport", "digitalCenters"]);

  if (width <= 900) await page.getByRole("button", { name: de ? "Menü öffnen" : "Open menu", exact: true }).click();
  const more = page.locator(".nav-more > summary");
  await more.focus();
  await page.keyboard.press("Enter");
  assert.equal(await page.locator(".nav-more").evaluate(details => details.open), true,
    "additional functions must open using a keyboard");
  const link = page.locator(".nav-more").getByRole("link", { name: de ? "Digitalzentralen" : "Command stations", exact: true });
  await link.scrollIntoViewIfNeeded();
  const target = await link.boundingBox();
  assert.ok(target && target.x >= 0 && target.x + target.width <= width + 1);
  if (width <= 900) assert.ok(target.height >= 44);
  await page.screenshot({ path: screenshotPrefix + "-collection-navigation.png" });
  await link.click();
  await page.waitForURL("**/digital-centers");
  await page.getByRole("heading", { name: de ? "Digitalzentralen" : "Command stations", exact: true }).waitFor();
  await page.goto(base + "/settings", { waitUntil: "domcontentloaded" });
  await page.locator(".settings-dashboard-grid").waitFor();
  const all = page.getByRole("button", { name: de ? "Alle Funktionen" : "All functions", exact: true });
  const [restored] = await Promise.all([
    page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
    all.click()
  ]);
  assert.equal(restored.status(), 200);
  assert.equal(await all.getAttribute("aria-pressed"), "true");
  assert.equal(await page.locator(".nav-more").count(), 0);
  if (width === 390) {
    const failSave = route => route.request().method() === "PUT" ? route.abort("failed") : route.continue();
    await page.route("**/api/v1/profile/settings", failSave);
    try {
      await page.getByRole("button", { name: de ? "Sammlung" : "Collection", exact: true }).click();
      await page.getByRole("alert").filter({ hasText: de ? "lokal aktiv" : "active locally" }).waitFor();
    } finally {
      await page.unroute("**/api/v1/profile/settings", failSave);
    }
    const [retried] = await Promise.all([
      page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
      page.getByRole("button", { name: de ? "Ansicht erneut speichern" : "Retry saving view", exact: true }).click()
    ]);
    assert.equal(retried.status(), 200);
    const [finalSave] = await Promise.all([
      page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
      all.click()
    ]);
    assert.equal(finalSave.status(), 200);
  }
  await all.scrollIntoViewIfNeeded();
  await page.screenshot({ path: screenshotPrefix + "-personal-view.png" });
  if (width === 1280) await assertQueuedProfileNavigation(page, { base, key, de });
}

async function assertQueuedProfileNavigation(page, { base, key, de }) {
  const profileURL = base + "/api/v1/profile/settings";
  const isWrite = request => request.url() === profileURL && request.method() === "PUT";
  const [customized] = await Promise.all([
    page.waitForResponse(response => isWrite(response.request())),
    page.getByRole("button", { name: de ? "Zubehör ausblenden" : "Hide Accessories", exact: true }).click()
  ]);
  assert.equal(customized.status(), 200);
  let release;
  const held = new Promise(resolve => { release = resolve; });
  let intercepted = false;
  const handler = async route => {
    if (!intercepted && isWrite(route.request()) && route.request().postDataJSON()?.settings?.[key]) {
      intercepted = true;
      await held;
    }
    await route.continue();
  };
  await page.route("**/api/v1/profile/settings", handler);
  try {
    await Promise.all([
      page.waitForRequest(isWrite),
      page.getByRole("button", { name: de ? "Sammlung" : "Collection", exact: true }).click()
    ]);
    await page.getByRole("button", { name: de ? "Alle Funktionen" : "All functions", exact: true }).click();
    const desired = await page.evaluate(key => localStorage.getItem(key), key);
    await page.evaluate(() => { document.documentElement.dataset.profileQueueProbe = "pending"; });
    await page.locator("#main-navigation").getByRole("link", {
      name: de ? "Fahrzeugbestand" : "Vehicle inventory", exact: true
    }).click();
    await page.waitForURL("**/vehicles");
    await page.locator(".inventory-filter-row").waitFor();
    assert.equal(await page.evaluate(() => document.documentElement.dataset.profileQueueProbe), "pending",
      "sidebar navigation must preserve the document containing pending saves");
    const finished = page.waitForResponse(response => isWrite(response.request()) &&
      response.request().postDataJSON()?.settings?.[key] === desired);
    release();
    assert.equal((await finished).status(), 200);
    const stored = await page.request.get(profileURL);
    assert.equal(stored.status(), 200);
    assert.equal((await stored.json()).settings[key], desired, "latest queued choice must win");
    await page.locator("#main-navigation").getByRole("link", { name: de ? "Einstellungen" : "Settings", exact: true }).click();
    await page.locator(".settings-dashboard-grid").waitFor();
    const synchronized = page.waitForResponse(response => response.url() === profileURL && response.request().method() === "GET");
    await page.reload({ waitUntil: "domcontentloaded" });
    assert.equal((await synchronized).status(), 200);
    await page.locator(".settings-dashboard-grid").waitFor();
    await page.waitForFunction(({ key, desired }) => localStorage.getItem(key) === desired, { key, desired });
  } finally {
    release();
    await page.unroute("**/api/v1/profile/settings", handler);
  }
}

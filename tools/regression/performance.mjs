import assert from "node:assert/strict";
import { chromium } from "playwright";
import { blockExternalRequests, closeBrowser, closeContext } from "./browser-network.mjs";

const catalogPath = "/api/v1/decoder-library";
const catalogChunk = /^\/assets\/decoderLibrary-[^/]+\.js$/;
const inventoryChunk = /^\/assets\/VehiclesView-[^/]+\.js$/;
const desktop = ".inventory-desktop-content";
const mobile = ".inventory-mobile-list";

async function openInventoryFromOverview(page, base, requests) {
  await page.goto(`${base}/overview`);
  await page.getByRole("heading", { name: "Übersicht", exact: true }).first().waitFor();
  assert.ok(!requests.has(catalogPath), "Overview must not request the decoder catalog API");
  assert.ok(![...requests].some(path => catalogChunk.test(path) || inventoryChunk.test(path)),
    "Overview must not download the decoder catalog or vehicle inventory chunks");

  const [response] = await Promise.all([
    page.waitForResponse(response => new URL(response.url()).pathname === catalogPath),
    page.getByRole("link", { name: "Fahrzeugbestand", exact: true }).click(),
  ]);
  assert.equal(response.status(), 200, "Inventory must load the decoder catalog successfully");
  assert.ok([...requests].some(path => catalogChunk.test(path)), "Inventory must load the decoder catalog chunk");
  assert.ok([...requests].some(path => inventoryChunk.test(path)), "Inventory must load its route chunk");
}

async function checkLazyImages(page, requests, imageCount, compact) {
  const active = page.locator(compact ? mobile : desktop);
  await active.waitFor();
  assert.equal(await page.locator(compact ? desktop : mobile).count(), 0, "Only the active layout may mount");
  const images = active.locator(compact ? ".vehicle-mobile-image img" : "img.inventory-thumb");
  await images.first().waitFor();
  assert.equal(await images.count(), imageCount);
  assert.equal(await images.and(page.locator('img[loading="lazy"]')).count(), imageCount,
    "Every inventory preview must opt into native lazy loading");

  const last = images.last();
  const lastPath = new URL(await last.getAttribute("src"), page.url()).pathname;
  const imageRequests = () => [...requests].filter(path => /^\/api\/v1\/vehicles\/[^/]+\/images\//.test(path)).length;
  await images.first().evaluate(image => image.decode());
  const beforeScroll = imageRequests();
  assert.ok(beforeScroll > 0 && beforeScroll < imageCount, `Initially requested ${beforeScroll} of ${imageCount} previews`);
  assert.ok(!requests.has(lastPath), "The last preview must wait until the user scrolls");
  await last.scrollIntoViewIfNeeded();
  await last.evaluate(image => image.decode());
  assert.ok(requests.has(lastPath), "Scrolling to the last preview must request and display it");
  console.log(`${compact ? "Mobile" : "Desktop"} previews: ${beforeScroll}/${imageCount} before scrolling, ${imageRequests()}/${imageCount} after`);
}

async function checkResponsiveState(page) {
  const firstVehicle = page.getByRole("row").filter({ hasText: "Performance vehicle 0001" });
  const checkbox = firstVehicle.getByRole("checkbox");
  await checkbox.press("Space");
  assert.ok(await checkbox.isChecked(), "The keyboard must select the vehicle");
  assert.ok(await checkbox.evaluate(input => input === document.activeElement), "Selection must retain keyboard focus");

  const tableSet = page.locator(".vehicle-set-hierarchy-toggle");
  await tableSet.press("Space");
  assert.equal(await tableSet.getAttribute("aria-expanded"), "true");
  await page.getByRole("row").filter({ hasText: "Performance set A" }).waitFor();

  await page.setViewportSize({ width: 900, height: 900 });
  await page.locator(mobile).waitFor();
  assert.equal(await page.locator(desktop).count(), 0);
  const mobileVehicle = page.locator(".vehicle-mobile-toggle").filter({ hasText: "Performance vehicle 0001" });
  const mobileSet = page.locator(".vehicle-mobile-set-toggle");
  await mobileVehicle.click();
  await mobileSet.click();
  await page.locator(mobile).getByText("Performance set A", { exact: true }).first().waitFor();

  await page.setViewportSize({ width: 901, height: 900 });
  await page.locator(desktop).waitFor();
  assert.equal(await page.locator(mobile).count(), 0);
  assert.ok(await checkbox.isChecked(), "Vehicle selection must survive a layout switch");
  assert.equal(await tableSet.getAttribute("aria-expanded"), "true", "Desktop set expansion must survive a layout switch");

  await page.setViewportSize({ width: 390, height: 900 });
  await page.locator(mobile).waitFor();
  assert.equal(await page.locator(desktop).count(), 0);
  assert.equal(await mobileVehicle.getAttribute("aria-expanded"), "true", "Mobile vehicle expansion must survive a layout switch");
  assert.equal(await mobileSet.getAttribute("aria-expanded"), "true", "Mobile set expansion must survive a layout switch");
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), "Mobile layout must fit the viewport");
}

async function checkViewport(browser, { base, credentials, imageCount }, width) {
  const context = await browser.newContext({ viewport: { width, height: 900 }, serviceWorkers: "block" });
  try {
    await blockExternalRequests(context, new URL(base));
    await context.addInitScript(() => localStorage.setItem("railkeeper.settings.language", "de"));
    const login = await context.request.post(`${base}/api/v1/auth/login`, { headers: { Origin: base }, data: credentials });
    assert.equal(login.status(), 200);
    const page = await context.newPage();
    const requests = new Set();
    const errors = [];
    page.on("request", request => requests.add(new URL(request.url()).pathname));
    page.on("pageerror", error => errors.push(error.message));
    const compact = width === 390;
    if (compact) {
      await page.goto(`${base}/vehicles`);
    } else {
      await openInventoryFromOverview(page, base, requests);
    }
    await checkLazyImages(page, requests, imageCount, compact);
    if (!compact) await checkResponsiveState(page);
    assert.deepEqual(errors, [], "No browser runtime errors");
  } finally {
    await closeContext(context);
  }
}

let input = "";
for await (const chunk of process.stdin) input += chunk;
const config = JSON.parse(input);
assert.equal(new URL(config.base).hostname, "127.0.0.1", "Only the synthetic loopback server is supported");
const browser = await chromium.launch();
try {
  await checkViewport(browser, config, 1440);
  await checkViewport(browser, config, 390);
  console.log("Browser performance checks passed");
} finally {
  await closeBrowser(browser);
}

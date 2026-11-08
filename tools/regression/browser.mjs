import { blockExternalRequests, closeBrowser, closeContext } from "./browser-network.mjs";
import assert from "node:assert/strict";
import { mkdir, writeFile } from "node:fs/promises";
import { chromium } from "playwright";
import { assertSaveErrorRecovery } from "./quality-save-errors.mjs";
import { assertInventoryToolbar } from "./personal-entry.mjs";
import { assertResponsiveWorkspaces } from "./responsive.mjs";

let inputText = "";
for await (const part of process.stdin) inputText += part;
const input = JSON.parse(inputText);
const origin = new URL(input.base);
assert.equal(origin.hostname, "127.0.0.1");
assert.equal(origin.protocol, "http:");
await mkdir(input.output, { recursive: true });
const browser = await chromium.launch({ headless: true });
const checks = [];
try {
  for (const language of ["de", "en"]) {
    const de = language === "de";
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
    await blockExternalRequests(context, origin);
    await context.addInitScript(({ language, origin }) => {
      if (window.location.origin !== origin) return;
      localStorage.setItem("railkeeper.settings.language", language);
      if (!localStorage.getItem("railkeeper.theme")) localStorage.setItem("railkeeper.theme", "light");
    }, { language, origin: origin.origin });
    const page = await context.newPage();
    page.setDefaultTimeout(10000);
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto(input.base + "/vehicles");
    await page.getByLabel(de ? "Benutzername" : "Username", { exact: true }).fill(input.credentials.username);
    await page.getByLabel(de ? "Passwort" : "Password", { exact: true }).fill(input.credentials.password);
    await page.getByRole("button", { name: de ? "Anmelden" : "Sign in", exact: true }).click();
    await page.waitForURL("**/overview");
    await assertInventoryToolbar(page, input, de, checks);
    await assertSaveErrorRecovery(page, input, de, checks);
    const workspaces = [
      ["/overview", de ? "Übersicht" : "Overview"],
      ["/vehicles", de ? "Bestand" : "Inventory"],
      ["/accessories", de ? "Artikel" : "Articles"],
      ["/exhibition", de ? "Ausstellung" : "Exhibition"],
      ["/import-export", "Import/Export"],
      ["/digital-centers", de ? "Digitalzentralen" : "Command stations"],
      ["/settings", de ? "Allgemein" : "General"]
    ];
    for (const [path, title] of workspaces) {
      await page.goto(input.base + path, { waitUntil: "domcontentloaded" });
      await page.getByRole("heading", { name: title, exact: true }).first().waitFor();
      assert.equal(await page.getByLabel(de ? "Benutzername" : "Username", { exact: true }).count(), 0);
      checks.push(language + path);
    }
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: "Referenzlok 1", exact: true }) });
    const edit = () => row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    await edit();
    let dialog = page.getByRole("dialog");
    const localImageUrl = `/api/v1/vehicles/${input.fixture.vehicleId}/images/${input.fixture.imageId}/file`;
    await dialog.getByRole("button", { name: "Uploads", exact: true }).click();
    assert.equal(await dialog.getByRole("link", { name: de ? "Quelle öffnen" : "Open source", exact: true })
      .getAttribute("href"), localImageUrl);
    await dialog.locator("img").first().click();
    const preview = page.locator(".image-preview-layer");
    await preview.waitFor();
    assert.equal(await preview.locator("a.image-title-link").getAttribute("href"), localImageUrl);
    await preview.getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).click();
    await preview.waitFor({ state: "hidden" });
    await dialog.getByRole("button", { name: de ? "Modell" : "Model", exact: true }).click();
    checks.push(language + ":local-image-source-actions");
    const label = de ? /^Bezeichnung/ : /^Designation/;
    await dialog.getByRole("textbox", { name: label }).fill("UNSAVED-REFERENCE");
    await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await dialog.waitFor({ state: "hidden" });
    assert.equal(await page.getByText("UNSAVED-REFERENCE", { exact: true }).count(), 0);
    await edit();
    dialog = page.getByRole("dialog");
    assert.equal(await dialog.getByRole("textbox", { name: label }).inputValue(), "Referenzlok 1");
    await dialog.getByRole("textbox", { name: label }).fill("Referenzlok gespeichert");
    const saved = page.waitForResponse((response) => response.request().method() === "PUT"
      && response.url() === input.base + "/api/v1/vehicles/" + input.fixture.vehicleId);
    await dialog.getByRole("button", { name: de ? "Änderungen speichern" : "Save changes", exact: true }).click();
    assert.equal((await saved).status(), 200);
    // Saving intentionally keeps the editor open for further work.
    await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await dialog.waitFor({ state: "hidden" });
    await page.getByRole("cell", { name: "Referenzlok gespeichert", exact: true }).waitFor();
    const changed = page.getByRole("row").filter({ has: page.getByRole("cell", { name: "Referenzlok gespeichert", exact: true }) });
    await changed.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    await page.getByRole("dialog").getByRole("textbox", { name: label }).fill("Referenzlok 1");
    const restored = page.waitForResponse((response) => response.request().method() === "PUT"
      && response.url() === input.base + "/api/v1/vehicles/" + input.fixture.vehicleId);
    await page.getByRole("dialog").getByRole("button", { name: de ? "Änderungen speichern" : "Save changes", exact: true }).click();
    assert.equal((await restored).status(), 200);
    await page.getByRole("dialog").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });
    checks.push(language + ":edit-cancel-save-reopen");
    // 1024 and 640 CSS pixels also cover a 1280px display at 125% and 200% zoom.
    for (const width of [2580, 1440, 1280, 1024, 820, 640, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate((theme) => localStorage.setItem("railkeeper.theme", theme), theme);
        await page.reload({ waitUntil: "domcontentloaded" });
        await page.getByText("Referenzlok 1", { exact: true }).filter({ visible: true }).first().waitFor();
        const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
        assert.equal(overflow, false, `${language}/${theme}/${width}: page overflow`);
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}.png` });
        checks.push(`${language}:${theme}:${width}`);
        await assertResponsiveWorkspaces(page, {
          base: input.base, username: input.credentials.username, width, de,
          screenshotPrefix: `${input.output}/${language}-${theme}-${width}`
        });
        checks.push(`${language}:${theme}:${width}:filters-settings-and-uploads`);
      }
    }
    assert.deepEqual(errors, []);
    await closeContext(context);
  }
  await writeFile(input.output + "/report.json", JSON.stringify({ passed: true, checks }, null, 2) + "\n");
} finally {
  await closeBrowser(browser);
}

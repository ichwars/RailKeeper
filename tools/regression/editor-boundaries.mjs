import assert from "node:assert/strict";
import { mkdir, writeFile } from "node:fs/promises";
import { chromium } from "playwright";
import { assertMediaDraftWorkflow } from "./media-draft.mjs";
import { assertCVDraftWorkflow } from "./cv-draft.mjs";
import { assertMediaActionsWorkflow } from "./media-actions.mjs";
import { referenceActor } from "./browser-actor.mjs";
import { assertCVImportGuards } from "./cv-import-guards.mjs";
import { assertTransferImportRecovery } from "./transfer-import-recovery.mjs";
import { assertStationAdoptionGuards } from "./station-adoption-guards.mjs";
import { assertArticleReviewGuards } from "./article-review-guards.mjs";
import { assertLegacyCVAdoption } from "./legacy-cv-adoption.mjs";
import { assertCreationRetryGuards } from "./creation-retry-guards.mjs";
import { assertCatalogRevisionWorkflow } from "./catalog-editor.mjs";
import { assertSetCatalogWorkflow } from "./catalog-set-editor.mjs";

let inputText = "";
for await (const part of process.stdin) inputText += part;
const input = JSON.parse(inputText), origin = new URL(input.base);
assert.equal(origin.hostname, "127.0.0.1");
assert.equal(origin.protocol, "http:");
await mkdir(input.output, { recursive: true });
const browser = await chromium.launch({ headless: true });
const checks = [];
try {
  for (const language of ["de", "en"]) {
    const de = language === "de";
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
    await context.route("**/*", route => {
      const url = new URL(route.request().url());
      return url.origin === origin.origin || ["data:", "blob:"].includes(url.protocol)
        ? route.continue() : route.abort();
    });
    await context.addInitScript(language => localStorage.setItem("railkeeper.settings.language", language), language);
    const page = await context.newPage(), errors = [];
    page.setDefaultTimeout(10000);
    page.on("pageerror", error => errors.push(error.message));
    await page.goto(input.base + "/vehicles");
    await page.getByLabel(de ? "Benutzername" : "Username", { exact: true }).fill(input.credentials.username);
    await page.getByLabel(de ? "Passwort" : "Password", { exact: true }).fill(input.credentials.password);
    await page.getByRole("button", { name: de ? "Anmelden" : "Sign in", exact: true }).click();
    await page.waitForURL("**/overview");
    const detailURL = input.base + "/api/v1/vehicles/" + input.fixture.vehicleId;
    const readDetail = async () => {
      const response = await context.request.get(detailURL);
      assert.equal(response.status(), 200);
      return response.json();
    };
    const before = await readDetail();
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const edit = async () => {
      const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
      await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    };
    await edit();
    const dialog = page.getByRole("dialog");
    const tab = async name => dialog.getByRole("button", { name, exact: true }).click();
    const designation = () => dialog.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
    await designation().fill("UNSAVED-CORE");
    await tab(de ? "Steuerung" : "Control");
    const functionName = (key = "F0") => dialog.getByRole("textbox", {
      name: `${key} ${de ? "Funktionsname" : "Function name"}`, exact: true
    });
    await functionName().fill("UNSAVED-FUNCTION");
    await tab("CV");
    const raw = () => dialog.locator(".cv-subtabs button").nth(1).click();
    const manual = async () => {
      const details = dialog.locator("details.cv-manual-editor");
      if (!(await details.getAttribute("open")) && !(await details.evaluate(element => element.open))) {
        await details.locator("summary").click();
      }
      return details.getByRole("textbox", { name: de ? "Beschreibung" : "Description", exact: true });
    };
    await raw();
    await (await manual()).fill("UNSAVED-CV-FORM");
    await dialog.locator('.cv-tab input[type="file"]').nth(1).setInputFiles({
      name: "synthetic-functions.xml", mimeType: "text/xml", buffer: Buffer.from(
        '<decoder><function key="F0" name="Do not overwrite"/><function key="F2" name="Project function"/></decoder>')
    });
    await dialog.locator(".cv-subtabs button").nth(2).click();
    const adoptFunctions = dialog.getByRole("button", {
      name: de ? "In Funktionsentwurf übernehmen" : "Add to function draft", exact: true
    });
    await adoptFunctions.waitFor();
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        await adoptFunctions.scrollIntoViewIfNeeded();
        const bounds = await adoptFunctions.boundingBox();
        assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1,
          `${language}/${theme}/${width}: function-preview action outside viewport`);
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-function-preview.png` });
      }
    }
    await page.setViewportSize({ width: 1440, height: 1000 });
    await adoptFunctions.click();
    assert.equal(await functionName().inputValue(), "UNSAVED-FUNCTION");
    assert.equal(await functionName("F2").inputValue(), "Project function");
    await tab("Uploads");
    const imageTitle = () => dialog.getByRole("textbox", { name: de ? "Bildbeschreibung" : "Image description", exact: true }).first();
    const originalAttachment = () => dialog.locator(".attachment-row").filter({ hasText: "anleitung.txt" });
    await imageTitle().fill("UNSAVED-IMAGE");
    await originalAttachment().getByPlaceholder(de ? "Bemerkung" : "Note", { exact: true }).fill("UNSAVED-ATTACHMENT");
    const uploaded = page.waitForResponse(response => response.request().method() === "POST"
      && response.url() === detailURL + "/attachments");
    await dialog.locator(".upload-section").nth(1).locator('input[type="file"]').setInputFiles({
      name: `confirmed-${language}.txt`, mimeType: "text/plain", buffer: Buffer.from("Confirmed side action.\n")
    });
    const uploadConfirmation = page.getByRole("dialog", { name: de ? "Dateien jetzt speichern?" : "Save selected files now?", exact: true });
    await uploadConfirmation.getByRole("button", { name: de ? "Ausgewählte Dateien speichern" : "Save selected files", exact: true }).click();
    assert.equal((await uploaded).status(), 201);
    const uploadResult = page.getByRole("dialog", { name: de ? "Dateien: Fortschritt und Ergebnis" : "Files: progress and results", exact: true });
    await uploadResult.getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).last().click();
    await dialog.getByRole("status").filter({ hasText: de ? "Bereits gespeichert" : "Already saved" }).waitFor();
    await dialog.getByText(`confirmed-${language}.txt`, { exact: true }).waitFor();
    assert.equal(await imageTitle().inputValue(), "UNSAVED-IMAGE");
    assert.equal(await originalAttachment().getByPlaceholder(de ? "Bemerkung" : "Note", { exact: true }).inputValue(), "UNSAVED-ATTACHMENT");
    await tab(de ? "Modell" : "Model");
    assert.equal(await designation().inputValue(), "UNSAVED-CORE");
    await tab(de ? "Steuerung" : "Control");
    assert.equal(await functionName().inputValue(), "UNSAVED-FUNCTION");
    await tab("CV");
    await raw();
    assert.equal(await (await manual()).inputValue(), "UNSAVED-CV-FORM");
    await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await dialog.waitFor({ state: "hidden" });
    const after = await readDetail();
    const newAttachments = after.attachments.filter(value => !before.attachments.some(old => old.id === value.id));
    assert.equal(newAttachments.length, 1);
    assert.equal(newAttachments[0].originalName, `confirmed-${language}.txt`);
    assert.deepEqual({ ...after, attachments: before.attachments }, before,
      "Cancel must preserve only the already confirmed upload, with all draft fields unchanged on the server");
    await edit();
    assert.equal(await designation().inputValue(), before.name);
    await tab(de ? "Steuerung" : "Control");
    assert.equal(await functionName().inputValue(), before.functions.find(value => value.functionKey === "F0").name);
    await tab("Uploads");
    assert.equal(await imageTitle().inputValue(), before.images[0].title);
    assert.equal(await originalAttachment().getByPlaceholder(de ? "Bemerkung" : "Note", { exact: true }).inputValue(),
      before.attachments.find(value => value.id === input.fixture.attachmentId).description || "");
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        const bounds = await dialog.locator(".vehicle-save-state").boundingBox();
        assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1);
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}.png` });
      }
    }
    await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    assert.deepEqual(errors, []);
    checks.push(language + ":confirmed-upload-preserves-five-drafts-and-cancel");
    checks.push(language + ":decoder-file-supplements-shared-function-draft-without-overwriting");
    const reference = await referenceActor(input);
    try {
      checks.push(...await assertCVDraftWorkflow(page, input, de, reference));
      checks.push(...await assertCVImportGuards(page, input, de, reference));
      checks.push(...await assertMediaDraftWorkflow(page, input, de, reference));
      checks.push(...await assertMediaActionsWorkflow(page, input, de, reference));
      checks.push(...await assertTransferImportRecovery(page, input, de, reference));
      checks.push(...await assertStationAdoptionGuards(page, input, de, reference));
      checks.push(...await assertArticleReviewGuards(page, input, de, reference));
      checks.push(...await assertLegacyCVAdoption(page, input, de, reference));
      checks.push(...await assertCreationRetryGuards(page, input, de, reference));
      checks.push(...await assertCatalogRevisionWorkflow(page, input, de, reference));
      checks.push(...await assertSetCatalogWorkflow(page, input, de, reference));
    } finally { await reference.actor.dispose(); }
    assert.deepEqual(errors, []);
    await context.close();
  }
  await writeFile(input.output + "/report.json", JSON.stringify({ passed: true, checks }, null, 2) + "\n");
} finally {
  await browser.close();
}

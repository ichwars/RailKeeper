import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";

export async function assertCatalogRevisionWorkflow(page, input, de, { actor, csrfToken, origin }) {
  const language = de ? "de" : "en", checks = [];
  const path = "/api/v1/vehicles/" + encodeURIComponent(input.fixture.vehicleId);
  const catalogPath = "/api/v1/catalog/vehicle/" + encodeURIComponent(input.fixture.vehicleId);
  const read = async route => {
    const response = await actor.get(route); assert.equal(response.status(), 200); return response.json();
  };
  const before = await read(path), initialCatalog = await read(catalogPath);
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
  await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  const editor = page.locator(".vehicle-edit-modal"), panel = editor.locator(".vehicle-catalog-panel");
  const tab = name => editor.getByRole("button", { name, exact: true }).click();
  const designation = () => editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
  const functionName = () => editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true });
  const manualDescription = async () => {
    await editor.locator(".cv-subtabs button").nth(1).click();
    const details = editor.locator("details.cv-manual-editor");
    if (!await details.evaluate(element => element.open)) await details.locator("summary").click();
    return details.getByRole("textbox", { name: de ? "Beschreibung" : "Description", exact: true });
  };
  const coreDraft = `UNSAVED-CATALOG-CORE-${language}`, functionDraft = `UNSAVED-CATALOG-FUNCTION-${language}`;
  const cvDraft = `UNSAVED-CATALOG-CV-${language}`, savedName = `CATALOG-CONFIRMED-${language}`;
  const imageDraft = `UNSAVED-CATALOG-IMAGE-${language}`;
  const imageTitle = () => editor.getByRole("textbox", { name: de ? "Bildbeschreibung" : "Image description", exact: true }).first();
  await designation().fill(coreDraft);
  await tab(de ? "Steuerung" : "Control"); await functionName().fill(functionDraft);
  await tab("CV"); await editor.locator(".cv-subtabs button").nth(1).click();
  await (await manualDescription()).fill(cvDraft);
  await tab("Uploads"); await imageTitle().fill(imageDraft);
  await tab(de ? "Modell" : "Model");
  await panel.locator("summary").first().click();
  await panel.getByRole("button", { name: de ? "Neue Revision entwerfen" : "Draft a new revision", exact: true }).click();
  const sourceName = () => panel.getByRole("textbox", { name: de ? "Artikelwert: Bezeichnung" : "Article value: Designation", exact: true });
  await sourceName().fill(savedName);
  await panel.getByRole("checkbox", { name: de ? "Übernehmen: Bezeichnung" : "Adopt: Designation", exact: true }).check();
  const note = `Synthetic confirmed catalog revision ${language}`;
  await panel.getByRole("textbox", { name: de ? "Notiz zur neuen Revision" : "Note for the new revision", exact: true }).fill(note);
  await tab("CV"); assert.equal(await (await manualDescription()).inputValue(), cvDraft);
  await tab(de ? "Modell" : "Model");
  assert.equal(await sourceName().inputValue(), savedName);
  assert.equal(await panel.getByRole("textbox", { name: de ? "Notiz zur neuen Revision" : "Note for the new revision", exact: true }).inputValue(), note);
  checks.push(`${language}:catalog-draft-survives-tab-switch`);
  let release = () => undefined, started;
  const held = new Promise(resolve => { started = resolve; });
  const hold = new Promise(resolve => { release = resolve; });
  const endpoint = origin.origin + catalogPath + "/revisions", detailEndpoint = origin.origin + path;
  let writes = 0, blockRefresh = false;
  const observe = request => { if (request.method() === "POST" && request.url() === endpoint) writes++; };
  page.on("request", observe);
  await page.route(endpoint, async route => {
    const response = await route.fetch(); assert.equal(response.status(), 201); started(); await hold;
    await route.fulfill({ response });
  });
  await page.route(detailEndpoint, route => blockRefresh && route.request().method() === "GET"
    ? route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ error: "synthetic_view_unavailable", message: "Synthetic saved-view failure" }) })
    : route.continue());
  try {
    await panel.getByRole("button", { name: de ? "Revision erstellen und Auswahl speichern" : "Create revision and save selection", exact: true }).click();
    await held;
    await tab("CV"); await (await manualDescription()).fill(cvDraft + "-latest");
    blockRefresh = true; release();
    await tab(de ? "Modell" : "Model");
    const retry = panel.getByRole("button", { name: de ? "Gespeicherte Ansicht erneut laden" : "Reload the saved view", exact: true });
    await retry.waitFor();
    await panel.getByText(de ? "Der Artikelstand ist gespeichert. Die Bestandsansicht konnte noch nicht aktualisiert werden."
      : "The article revision is saved. The inventory view could not be refreshed yet.", { exact: true }).waitFor();
    blockRefresh = false; await retry.click();
    await editor.locator(".vehicle-refresh-conflicts").waitFor();
    assert.equal(writes, 1, "Retrying a saved view must not repeat the confirmed catalog write");
    assert.equal(await designation().inputValue(), coreDraft);
    assert.ok((await editor.locator(".vehicle-refresh-conflicts").textContent()).includes(savedName));
    await tab(de ? "Steuerung" : "Control"); assert.equal(await functionName().inputValue(), functionDraft);
    await tab("CV"); assert.equal(await (await manualDescription()).inputValue(), cvDraft + "-latest");
    await tab("Uploads"); assert.equal(await imageTitle().inputValue(), imageDraft);
    checks.push(`${language}:catalog-confirmation-preserves-core-function-cv-and-media-drafts`, `${language}:catalog-view-retry-does-not-repeat-write`);
    await tab(de ? "Modell" : "Model");
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        await panel.locator("summary").first().scrollIntoViewIfNeeded();
        const box = await panel.boundingBox();
        assert.ok(box && box.x >= 0 && box.x + box.width <= width + 1, "Catalog panel outside editor");
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), "Catalog editor page overflow");
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/catalog-${language}-${theme}-${width}.png` });
      }
    }
    await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    const after = await read(path);
    const withoutClock = ({ revision, updatedAt, ...value }) => value;
    assert.deepEqual(withoutClock(after), { ...withoutClock(before), name: savedName }, "Cancel persisted an unconfirmed draft or lost inventory data");
    const catalog = await read(catalogPath);
    assert.ok(catalog.revisions.some(revision => revision.note === note), "Confirmed immutable revision disappeared after cancel");
    checks.push(`${language}:catalog-confirmation-remains-after-cancel`);
    const restore = await actor.post(catalogPath + "/revisions", { headers: { "X-CSRF-Token": csrfToken }, data: {
      id: randomBytes(16).toString("hex"), expectedRevision: catalog.revision, fields: initialCatalog.currentFields,
      note: `Synthetic explicit restore of original name ${language}`, selectedFields: ["name"]
    } });
    assert.equal(restore.status(), 201);
    assert.deepEqual(withoutClock(await read(path)), withoutClock(before), "Restoring the name changed unrelated inventory");
  } finally {
    release(); page.off("request", observe); await page.unroute(endpoint); await page.unroute(detailEndpoint);
  }
  return checks;
}

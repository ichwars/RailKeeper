import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";

export async function assertSetCatalogWorkflow(page, input, de, { actor, csrfToken, origin }) {
  const language = de ? "de" : "en", checks = [];
  const path = "/api/v1/vehicle-sets/" + encodeURIComponent(input.fixture.setId);
  const catalogPath = "/api/v1/catalog/set/" + encodeURIComponent(input.fixture.setId);
  const read = async route => {
    const response = await actor.get(route); assert.equal(response.status(), 200); return response.json();
  };
  const before = await read(path), initialCatalog = await read(catalogPath);
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  const row = page.locator(".vehicle-set-inventory-row").filter({ hasText: before.inventoryNumber });
  await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  const editor = page.locator(".vehicle-set-editor-dialog"), panel = editor.locator(".vehicle-catalog-panel");
  const general = () => editor.getByRole("tab", { name: de ? "Allgemein" : "General", exact: true }).click();
  const name = () => editor.getByRole("textbox", { name: de ? "Bezeichnung" : "Designation", exact: true });
  const price = () => editor.getByRole("textbox", { name: de ? "Preis" : "Price", exact: true });
  const save = () => editor.locator(".modal-actions").getByRole("button", { name: de ? "Speichern" : "Save", exact: true });
  const localName = `UNSAVED-SET-${language}`, savedName = `CONFIRMED-SET-${language}`;
  await name().fill(localName); await price().fill("321.00");
  await panel.locator("summary").first().click();
  await panel.getByRole("button", { name: de ? "Neue Revision entwerfen" : "Draft a new revision", exact: true }).click();
  await panel.getByRole("textbox", { name: de ? "Artikelwert: Bezeichnung" : "Article value: Designation", exact: true }).fill(savedName);
  await panel.getByRole("checkbox", { name: de ? "Übernehmen: Bezeichnung" : "Adopt: Designation", exact: true }).check();
  const note = () => panel.getByRole("textbox", { name: de ? "Notiz zur neuen Revision" : "Note for the new revision", exact: true });
  await note().fill(`Synthetic set catalog revision ${language}`);
  await editor.getByRole("tab", { name: "Upload", exact: true }).click(); await general();
  assert.equal(await note().inputValue(), `Synthetic set catalog revision ${language}`);
  checks.push(`${language}:set-catalog-draft-survives-tab-switch`);
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
    await held; assert.equal(await save().isDisabled(), true);
    await price().fill("432.00");
    blockRefresh = true; release();
    const retry = panel.getByRole("button", { name: de ? "Gespeicherte Ansicht erneut laden" : "Reload the saved view", exact: true });
    await retry.waitFor();
    await panel.getByText(de ? "Der Artikelstand ist gespeichert. Die Bestandsansicht konnte noch nicht aktualisiert werden."
      : "The article revision is saved. The inventory view could not be refreshed yet.", { exact: true }).waitFor();
    assert.equal(await save().isDisabled(), true);
    blockRefresh = false; await retry.click();
    await editor.locator(".vehicle-refresh-conflicts").waitFor();
    assert.equal(writes, 1); assert.equal(await name().inputValue(), localName); assert.equal(await price().inputValue(), "432.00");
    assert.equal(await save().isDisabled(), true);
    checks.push(`${language}:set-catalog-refresh-preserves-drafts-and-blocks-conflicting-save`);
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        await panel.locator("summary").first().scrollIntoViewIfNeeded();
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), "Set dialog page overflow");
        const footer = await editor.locator(".modal-actions").boundingBox();
        assert.ok(footer && footer.y + footer.height <= 1001 && footer.x >= 0 && footer.x + footer.width <= width + 1);
        if (width <= 820) {
          for (const button of await editor.locator(".modal-actions button, .vehicle-refresh-actions button").all()) {
            const box = await button.boundingBox();
            assert.ok(box && box.height >= 44,
              `${language}/${theme}/${width}: set action ${await button.textContent()} below touch target: ${box?.height}`);
          }
        }
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/catalog-set-${language}-${theme}-${width}.png` });
      }
    }
    await editor.getByRole("button", { name: de ? "Meine Änderungen für diese Felder behalten" : "Keep my changes for these fields", exact: true }).click();
    assert.equal(await save().isEnabled(), true);
    assert.equal(await name().inputValue(), localName);
    await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    const after = await read(path);
    const expected = { ...before, name: savedName, updatedAt: after.updatedAt,
      members: before.members.map(member => ({ ...member, vehicleSetName: savedName,
        vehicleSet: member.vehicleSet ? { ...member.vehicleSet, name: savedName } : member.vehicleSet })) };
    assert.deepEqual(after, expected, "Set catalog confirmation changed members, images or unconfirmed ownership fields");
    checks.push(`${language}:set-catalog-cancel-keeps-only-confirmed-change`);
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    await page.locator(".vehicle-set-inventory-row").filter({ hasText: before.inventoryNumber })
      .locator(".vehicle-set-name-cell button").click();
    const summary = page.locator(".vehicle-set-summary-dialog");
    await summary.locator(".vehicle-catalog-panel summary").first().click();
    await summary.getByText(savedName, { exact: true }).first().waitFor();
    assert.equal(await summary.getByRole("button", { name: de ? "Neue Revision entwerfen" : "Draft a new revision", exact: true }).count(), 0);
    await summary.locator(".modal-actions").getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).click();
    checks.push(`${language}:set-summary-catalog-is-read-only`);
    const catalog = await read(catalogPath);
    const restore = await actor.post(catalogPath + "/revisions", { headers: { "X-CSRF-Token": csrfToken }, data: {
      id: randomBytes(16).toString("hex"), expectedRevision: catalog.revision, fields: initialCatalog.currentFields,
      note: `Synthetic restore of original set name ${language}`, selectedFields: ["name"]
    } });
    assert.equal(restore.status(), 201);
    const restored = await read(path);
    assert.deepEqual(restored, { ...before, updatedAt: restored.updatedAt });
  } finally {
    release(); page.off("request", observe); await page.unroute(endpoint); await page.unroute(detailEndpoint);
  }
  return checks;
}

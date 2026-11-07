import assert from "node:assert/strict";

export async function assertCVImportGuards(page, input, de, reference) {
  const { actor, csrfToken, origin } = reference, language = de ? "de" : "en";
  const path = "/api/v1/vehicles/" + encodeURIComponent(input.fixture.vehicleId), headers = { "X-CSRF-Token": csrfToken };
  const profile = `Confirmed CV import ${language}`;
  const read = async () => { const response = await actor.get(path); assert.equal(response.status(), 200); return response.json(); };
  const create = async (number, value) => {
    const response = await actor.post(path + "/cv-values", { headers, data: { cvNumber: number, value, decoderProfile: profile } });
    assert.equal(response.status(), 201); return response.json();
  };
  const first = await create(1, 3), second = await create(2, 6), before = await read();
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  const edit = async () => {
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
    await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  };
  await edit();
  const editor = page.locator(".vehicle-edit-modal"), tab = name => editor.getByRole("button", { name, exact: true }).click();
  await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).fill("CV-IMPORT-UNSAVED");
  await tab(de ? "Steuerung" : "Control");
  await editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).fill("CV-IMPORT-FUNCTION");
  await tab("CV"); await editor.locator(".cv-subtabs button").nth(1).click();
  const details = editor.locator(".cv-manual-editor");
  if (!(await details.evaluate(element => element.open))) await details.locator("summary").click();
  const manual = () => details.getByRole("textbox", { name: de ? "Beschreibung" : "Description", exact: true });
  await manual().fill("CV-IMPORT-MANUAL");
  const choose = rows => editor.locator('.cv-tab input[type="file"]').first().setInputFiles({
    name: `confirmed-import-${language}.json`, mimeType: "application/json", buffer: Buffer.from(JSON.stringify({ cvValues: rows }))
  });
  const preview = editor.locator(".cv-import-preview"), table = preview.locator(".cv-import-table");
  const apply = () => preview.getByRole("button", { name: de ? "Ausgewählte Felder übernehmen" : "Apply selected fields", exact: true });
  const compare = () => preview.getByRole("button", { name: de ? "Erneut vergleichen" : "Compare again", exact: true });
  await choose([1, 2, 3, 4].map(number => ({ cvNumber: number, value: [4, 9, 10, 11][number - 1], decoderProfile: profile })));
  await table.getByRole("checkbox", { name: de ? "CV 4 übernehmen" : "Apply CV 4", exact: true }).uncheck();
  const changed = await actor.put(path + "/cv-values/" + second.id, { headers, data: {
    cvNumber: 2, value: 8, decoderProfile: profile, expectedRevision: second.revision
  } }); assert.equal(changed.status(), 200);
  const concurrent = await read(), url = origin.origin + path + "/cv-values/import";
  const rejected = page.waitForResponse(response => response.url() === url && response.request().method() === "POST");
  await apply().click(); assert.equal((await rejected).status(), 409);
  await preview.getByRole("alert").waitFor(); assert.equal(await apply().isDisabled(), true);
  assert.deepEqual(await read(), concurrent, "A stale CV import must roll back every selected value and history");
  for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 1000 });
    await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
    await preview.scrollIntoViewIfNeeded();
    const bounds = await preview.boundingBox();
    assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
    await page.keyboard.press("Tab"); await compare().focus();
    assert.ok(await compare().evaluate(button => document.activeElement === button && getComputedStyle(button).outlineStyle !== "none"));
    if (width === 390) {
      const size = await compare().boundingBox(); assert.ok(size && size.width >= 44 && size.height >= 44);
    }
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-cv-import-conflict.png` });
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
  await compare().click();
  await preview.getByRole("alert").waitFor({ state: "hidden" });
  assert.deepEqual(await read(), concurrent, "Comparing again must not write");
  assert.equal(await table.getByRole("checkbox", { name: de ? "CV 4 übernehmen" : "Apply CV 4", exact: true }).isChecked(), false);
  const accepted = page.waitForResponse(response => response.url() === url && response.request().method() === "POST");
  await apply().click(); assert.equal((await accepted).status(), 200);
  await preview.waitFor({ state: "hidden" });
  await page.waitForFunction(() => !document.querySelector(".modal-actions button.primary-button")?.disabled);
  const stored = await read(), imported = stored.cvValues.filter(value => value.decoderProfile === profile);
  assert.equal(imported.length, 3); assert.equal(imported.find(value => value.cvNumber === 1).id, first.id);
  assert.equal(imported.find(value => value.cvNumber === 2).id, second.id);
  assert.deepEqual({ ...stored, cvValues: before.cvValues }, before, "CV import changed another vehicle field or record");
  assert.equal(await manual().inputValue(), "CV-IMPORT-MANUAL");
  await tab(de ? "Modell" : "Model");
  assert.equal(await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), "CV-IMPORT-UNSAVED");
  await tab(de ? "Steuerung" : "Control");
  assert.equal(await editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).inputValue(), "CV-IMPORT-FUNCTION");
  await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  assert.deepEqual(await read(), stored);

  await edit(); await tab("CV"); await editor.locator(".cv-subtabs button").nth(1).click();
  let requests = 0;
  await page.route(url, async route => {
    if (route.request().method() !== "POST") return route.continue();
    requests++;
    const response = await route.fetch(); assert.equal(response.status(), 200);
    await route.abort("failed");
  });
  await choose([{ cvNumber: 5, value: 12, decoderProfile: profile }]);
  await apply().click(); await preview.getByRole("alert").waitFor();
  assert.equal(await apply().isDisabled(), true);
  const afterUnknown = await read();
  assert.equal(afterUnknown.cvValues.filter(value => value.decoderProfile === profile && value.cvNumber === 5).length, 1);
  await compare().click(); await preview.getByRole("alert").waitFor({ state: "hidden" });
  assert.equal(await table.getByRole("checkbox").first().isChecked(), false);
  assert.equal(await apply().isDisabled(), true); assert.equal(requests, 1);
  await page.unroute(url);
  await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  assert.deepEqual(await read(), afterUnknown, "A lost-response comparison must not duplicate CV IDs or histories");
  return [language + ":stale-import-rolls-back-complete-selection", language + ":fresh-import-keeps-unselected-identities",
    language + ":cv-import-preserves-other-drafts-and-cancel", language + ":lost-cv-import-response-does-not-repeat"];
}

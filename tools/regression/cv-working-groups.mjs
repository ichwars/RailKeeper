import assert from "node:assert/strict";

export async function assertCVWorkingGroups(page, input, de, { actor, csrfToken }) {
  const language = de ? "de" : "en", checks = [], headers = { "X-CSRF-Token": csrfToken };
  const read = async path => { const result = await actor.get(path); assert.equal(result.status(), 200); return result.json(); };
  const post = async (path, data, status = 201) => { const result = await actor.post(path, { headers, data }); assert.equal(result.status(), status, await result.text()); return result.json(); };
  const vehicle = await post("/api/v1/vehicles", { inventoryNumber: `SYNTHETIC-CV-${language}`, name: `CV working groups ${language}`,
    manufacturer: "Synthetic", articleNumber: `CV-${language}`, category: "Lokomotive", gattung: "Diesellokomotive", gauge: "H0", digital: true,
    decoderType: "Unbekannter Decoder", condition: "ready", qrCodeEnabled: true });
  const path = `/api/v1/vehicles/${vehicle.id}`, endpoint = path + "/decoder-configurations";
  const profileA = "Arbeitsgruppe A", profileB = "Arbeitsgruppe B";
  const original = await post(path + "/cv-values", { cvNumber: 1, value: 3, decoderProfile: profileA });
  await post(path + "/cv-values", { cvNumber: 1, value: 7, decoderProfile: profileB });
  const before = await read(path), initial = await read(endpoint);
  const groupA = initial.workspaces.find(group => group.profile === profileA);
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  await page.getByRole("row").filter({ has: page.getByRole("cell", { name: vehicle.name, exact: true }) })
    .getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  const editor = page.locator(".vehicle-edit-modal");
  await page.route(input.base + endpoint, route => route.fulfill({ status: 503, contentType: "application/json",
    body: JSON.stringify({ error: "Synthetic context unavailable" }) }), { times: 1 });
  await editor.getByRole("button", { name: "CV", exact: true }).click();
  const picker = editor.locator(".cv-working-group-picker"), form = editor.locator(".cv-manual-editor");
  await picker.getByRole("alert").waitFor();
  await picker.getByRole("button", { name: de ? "CV-Gruppen neu laden" : "Reload CV groups", exact: true }).click();
  await editor.locator(".cv-subtabs button").nth(1).click();
  const selection = picker.getByRole("button", { name: de ? "Voreinstellung für neue CV-Eingaben und Dateiübernahmen" : "Default for new CV entries and file imports", exact: true });
  const confirm = picker.getByRole("button", { name: de ? "CV-Zielgruppe jetzt bestätigen" : "Confirm target CV group now", exact: true });
  const choose = async profile => { await selection.click(); await page.getByRole("option", { name: profile, exact: true }).click(); };
  const activate = async profile => { await choose(profile); await confirm.click();
    await picker.locator(".cv-working-group-current").filter({ hasText: `${de ? "Bestätigte CV-Zielgruppe" : "Confirmed target CV group"}: ${profile}` }).waitFor(); };
  await selection.waitFor(); await page.waitForFunction(() => !document.querySelector(".cv-working-group-picker .app-select-trigger")?.disabled);
  await form.locator("summary").click();
  const profileInput = form.getByRole("textbox", { name: de ? "Decoder-Profil" : "Decoder profile", exact: true });
  await choose(profileA); assert.equal(await profileInput.inputValue(), "");
  await selection.focus(); await page.keyboard.press("Tab");
  assert.equal(await confirm.evaluate(element => element === document.activeElement), true);
  checks.push(`${language}:context-failure-retry-and-keyboard-confirmation-focus`);
  await confirm.click(); await page.waitForFunction(profile => document.querySelector('.cv-manual-editor input[placeholder]')?.value === profile, profileA);
  await form.getByRole("spinbutton", { name: "CV-Nr.", exact: true }).fill("25");
  await form.getByRole("spinbutton", { name: de ? "Wert" : "Value", exact: true }).fill("9");
  await form.getByRole("button", { name: de ? "Im CV-Entwurf vormerken" : "Stage in CV draft", exact: true }).click();
  await choose(profileB); assert.equal(await profileInput.inputValue(), profileA);
  await confirm.click(); await page.waitForFunction(profile => document.querySelector('.cv-manual-editor input[placeholder]')?.value === profile, profileB);
  const draft = editor.locator(".cv-draft-review"); await draft.locator("strong").filter({ hasText: profileA }).waitFor();
  const applyDraft = draft.getByRole("button", { name: de ? "CV-Entwurf speichern" : "Save CV draft", exact: true });
  await editor.locator(".cv-values-table tbody tr").filter({ has: page.getByRole("cell", { name: /^CV 25/ }) })
    .getByRole("button", { name: de ? "CV bearbeiten" : "Edit CV", exact: true }).click();
  assert.equal(await profileInput.inputValue(), profileA);
  await applyDraft.click(); await draft.waitFor({ state: "hidden" });
  if (!await form.evaluate(element => element.open)) await form.locator("summary").click();
  assert.equal(await profileInput.inputValue(), profileA);
  await form.getByRole("spinbutton", { name: "CV-Nr.", exact: true }).fill("28");
  await form.getByRole("button", { name: de ? "Im CV-Entwurf vormerken" : "Stage in CV draft", exact: true }).click();
  await applyDraft.click(); await draft.waitFor({ state: "hidden" });
  let current = await read(path);
  assert.equal(current.cvValues.find(value => value.cvNumber === 25).decoderProfile, profileA);
  assert.equal(current.cvValues.find(value => value.cvNumber === 28).decoderProfile, profileA);
  assert.deepEqual(current.cvValues.filter(value => value.cvNumber === 1), before.cvValues);
  checks.push(`${language}:explicit-default-and-staged-target-retained`);
  checks.push(`${language}:committed-temporary-row-retains-open-nondefault-target`);

  await activate(profileA);
  await editor.locator('input[type="file"][accept=".json,.csv,.txt"]').setInputFiles({ name: "original-profile.json", mimeType: "application/json",
    buffer: Buffer.from(JSON.stringify([{ cvNumber: 26, value: 8, decoderProfile: "Original source profile", workspaceId: "file-cannot-authorize", expectedWorkspaceRevision: "forged" }])) });
  const preview = editor.locator(".cv-import-preview"); await preview.getByText(/Original source profile/).waitFor();
  await activate(profileB);
  await preview.getByRole("button", { name: de ? "Ausgewählte Felder übernehmen" : "Apply selected fields", exact: true }).click();
  await preview.waitFor({ state: "hidden" }); current = await read(path);
  assert.equal(current.cvValues.find(value => value.cvNumber === 26).decoderProfile, profileA);
  checks.push(`${language}:source-profile-visible-and-prepared-import-not-redirected`);

  await editor.locator(".cv-values-table tbody tr").filter({ has: page.getByRole("cell", { name: /^CV 1$/ }) })
    .filter({ has: page.getByRole("cell", { name: profileA, exact: true }) })
    .getByRole("button", { name: de ? "CV bearbeiten" : "Edit CV", exact: true }).click();
  await form.getByRole("spinbutton", { name: de ? "Wert" : "Value", exact: true }).fill("5");
  await activate(profileA); await activate(profileB); assert.equal(await profileInput.inputValue(), profileA);
  await form.getByRole("button", { name: de ? "Im CV-Entwurf vormerken" : "Stage in CV draft", exact: true }).click();
  const state = await read(endpoint);
  const unitId = crypto.randomUUID().replaceAll("-", ""), installationId = crypto.randomUUID().replaceAll("-", "");
  let changed = await post(endpoint + "/instances", { id: unitId, expectedRevision: state.revision, label: "Synthetic replacement",
    manufacturer: "", model: "", serialNumber: "TEST-CV-001", notes: "" });
  changed = await post(endpoint + "/installations", { id: installationId, expectedRevision: changed.revision,
    decoderId: unitId, installedAt: null, notes: "" });
  const assignment = await actor.put(endpoint + `/workspaces/${groupA.id}/assignment`, { headers,
    data: { expectedRevision: changed.revision, installationId, snapshotId: crypto.randomUUID().replaceAll("-", "") } });
  assert.equal(assignment.status(), 200);
  await applyDraft.click(); await draft.getByRole("alert").waitFor();
  assert.equal((await read(path)).cvValues.find(value => value.id === original.id).value, 3);
  await picker.getByRole("button", { name: de ? "CV-Gruppen neu laden" : "Reload CV groups", exact: true }).click();
  await draft.getByRole("button", { name: de ? "Diesen CV als neue Entwurfsbasis verwenden" : "Use this CV as the new draft baseline", exact: true }).click();
  await page.waitForFunction(() => !document.querySelector(".cv-draft-actions .primary-button")?.disabled);
  await applyDraft.click(); await draft.waitFor({ state: "hidden" });
  current = await read(path); assert.equal(current.cvValues.find(value => value.id === original.id).value, 5);
  assert.deepEqual(current.cvValues.filter(value => value.decoderProfile === profileB), before.cvValues.filter(value => value.decoderProfile === profileB));
  checks.push(`${language}:concurrent-hardware-change-blocked-and-explicitly-recompared`);

  await activate(profileA);
  await editor.locator(".cv-subtabs button").nth(2).click();
  const fileName = `confirmed-source-${language}.csv`, sourceBytes = Buffer.from("CV;Value\n27;9\n");
  await editor.locator('input[type="file"][accept*=".esux"]').setInputFiles({ name: fileName, mimeType: "text/csv", buffer: sourceBytes });
  const filePreview = editor.locator(".cv-file-preview"); await filePreview.waitFor();
  await activate(profileB);
  await editor.locator(".cv-upload-options summary").click();
  await editor.getByPlaceholder(de ? "Bemerkung für neue Dateien" : "Note for new files", { exact: true }).fill("Retained source note");
  const fileContext = await read(endpoint);
  const unassign = await actor.put(endpoint + `/workspaces/${groupA.id}/assignment`, { headers,
    data: { expectedRevision: fileContext.revision, installationId: null, snapshotId: crypto.randomUUID().replaceAll("-", "") } });
  assert.equal(unassign.status(), 200);
  await filePreview.getByRole("button", { name: de ? "Dateien speichern" : "Save files", exact: true }).click();
  const upload = page.locator(".vehicle-media-action-layer");
  const reviewFile = upload.getByRole("button", { name: de ? "Dateiziel erneut vergleichen" : "Compare file target again", exact: true });
  await reviewFile.waitFor();
  assert.equal((await read(path)).cvFiles?.length || 0, 0);
  assert.equal(await upload.getByRole("button", { name: de ? "Offene Dateien fortsetzen" : "Continue remaining files", exact: true }).isDisabled(), true);
  await reviewFile.click();
  const resume = upload.getByRole("button", { name: de ? "Offene Dateien fortsetzen" : "Continue remaining files", exact: true });
  await page.waitForFunction(() => !document.querySelector(".vehicle-media-action-buttons .primary-button")?.disabled);
  await upload.getByText(/TEST-CV-001/).waitFor();
  assert.equal((await read(path)).cvFiles?.length || 0, 0);
  await page.screenshot({ path: `${input.output}/cv-file-recompared-${language}.png` });
  const storedResponse = page.waitForResponse(response => response.url() === input.base + path + "/cv-files" && response.request().method() === "POST" && response.status() === 201);
  await resume.click(); await storedResponse;
  await page.waitForFunction(() => !document.querySelector(".vehicle-media-action-buttons .secondary-button")?.disabled);
  const storedFile = (await read(path)).cvFiles.find(file => file.originalName === fileName);
  assert.equal(storedFile.decoderProfile, profileA); assert.equal(storedFile.description, "Retained source note");
  const downloaded = await actor.get(path + `/cv-files/${storedFile.id}/download`); assert.equal(downloaded.status(), 200);
  assert.deepEqual(await downloaded.body(), sourceBytes);
  await upload.locator(".vehicle-media-action-buttons").getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).click();
  await editor.locator(".attachment-row").filter({ hasText: fileName }).getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
  const deletion = page.locator(".vehicle-media-action-layer");
  await deletion.getByText(new RegExp(profileA)).waitFor();
  await deletion.getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
  await deletion.waitFor({ state: "hidden" });
  assert.equal((await read(path)).cvFiles?.length || 0, 0);
  checks.push(`${language}:file-scope-frozen-conflict-reviewed-separately-original-bytes-and-deletion`);
  await activate(profileA);
  for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
    await page.setViewportSize({ width, height: 1000 });
    await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
    await picker.scrollIntoViewIfNeeded();
    assert.equal(await editor.locator(".modal-body").evaluate(element => element.scrollWidth > element.clientWidth + 1), false);
    for (const button of await picker.getByRole("button").all()) {
      const box = await button.boundingBox(); assert.ok(box && box.height >= 44 && box.x >= 0 && box.x + box.width <= width + 1);
    }
    await page.screenshot({ path: `${input.output}/cv-working-groups-${language}-${theme}-${width}.png` });
    checks.push(`${language}:${theme}:${width}:working-group-selection`);
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
  await editor.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  return checks;
}

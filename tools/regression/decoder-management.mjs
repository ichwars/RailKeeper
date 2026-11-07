import assert from "node:assert/strict";

export async function assertDecoderManagement(page, input, de, { actor, csrfToken }) {
  const language = de ? "de" : "en", checks = [];
  const path = `/api/v1/vehicles/${encodeURIComponent(input.fixture.vehicleId)}`, endpoint = path + "/decoder-configurations";
  const read = async route => { const response = await actor.get(route); assert.equal(response.status(), 200); return response.json(); };
  const before = await read(path), initial = await read(endpoint);
  const base = before.cvValues.find(value => value.cvNumber === 1 && value.decoderProfile === "Legacy-Profil");
  assert.ok(base);
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
  await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  const editor = page.locator(".vehicle-edit-modal"), panel = editor.locator(".decoder-configurations"), command = panel.locator(".decoder-command");
  const tab = name => editor.getByRole("button", { name, exact: true }).click();
  const designation = () => editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
  await designation().fill("UNSAVED-DECODER-CORE");
  await tab(de ? "Steuerung" : "Control");
  await editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).fill("UNSAVED-DECODER-FUNCTION");
  await tab("CV"); await editor.locator(".cv-subtabs button").nth(1).click();
  await editor.locator(".cv-values-table tbody tr").filter({ has: page.getByRole("cell", { name: /^CV 1$/ }) })
    .filter({ has: page.getByRole("cell", { name: "Legacy-Profil", exact: true }) })
    .getByRole("button", { name: de ? "CV bearbeiten" : "Edit CV", exact: true }).click();
  await editor.locator(".cv-manual-editor").getByRole("spinbutton", { name: de ? "Wert" : "Value", exact: true }).fill(String(base.value + 1));
  await editor.getByRole("button", { name: de ? "Im CV-Entwurf vormerken" : "Stage in CV draft", exact: true }).click();
  await editor.locator(".cv-subtabs").getByRole("button", { name: de ? "Decoder und Arbeitsgruppen" : "Decoders and working groups", exact: true }).click();
  const newUnit = () => panel.getByRole("button", { name: de ? "Decoderexemplar erfassen" : "Record decoder unit", exact: true });
  await newUnit().waitFor();
  await page.waitForFunction(() => !document.querySelector(".decoder-configurations-toolbar button:nth-child(2)")?.disabled);
  const field = name => command.getByRole("textbox", { name, exact: true });
  const commit = async (name, status = 200) => {
    const response = page.waitForResponse(response => response.url().startsWith(input.base + endpoint) && response.request().method() !== "GET");
    await command.getByRole("button", { name, exact: true }).click();
    const result = await response; assert.equal(result.status(), status);
    if (status < 300) await command.waitFor({ state: "hidden" });
    return result.json();
  };
  const unitA = `SYNTHETIC-UNIT-${language}-A`, unitB = `SYNTHETIC-UNIT-${language}-B`;
  for (const label of [unitA, unitB]) {
    await newUnit().click();
    await field(de ? "Bezeichnung des Exemplars" : "Unit designation").fill(label);
    await field(de ? "Hersteller" : "Manufacturer").fill("ESU");
    await field(de ? "Modell" : "Model").fill("LokSound 5");
    await commit(de ? "Exemplar jetzt speichern" : "Save unit now", 201);
  }
  const pool = panel.locator(".decoder-instance-pool"); await pool.locator("summary").click();
  for (const label of [unitA, unitB]) {
    await pool.locator(".decoder-instance").filter({ has: page.getByText(label, { exact: true }) })
      .getByRole("button", { name: de ? "Einbau erfassen" : "Record installation", exact: true }).click();
    assert.equal(await field(de ? "Einbau" : "Installed").inputValue(), "");
    if (label === unitB) await field(de ? "Einbau" : "Installed").fill("2026-10-30");
    await commit(de ? "Einbau jetzt speichern" : "Save installation now", 201);
  }
  let current = await read(endpoint);
  const decoderA = current.instances.find(value => value.label === unitA), decoderB = current.instances.find(value => value.label === unitB);
  assert.notEqual(decoderA.id, decoderB.id);
  const installationA = current.installations.find(value => value.decoderId === decoderA.id && value.state === "installed");
  assert.equal(installationA.installedAt, null);
  assert.equal(current.installations.filter(value => value.vehicleId === before.id && value.state === "installed").length,
    initial.installations.filter(value => value.vehicleId === before.id && value.state === "installed").length + 2);
  const group = panel.locator(".decoder-working-group").filter({ has: page.getByText("Legacy-Profil", { exact: true }) });
  await group.getByRole("button", { name: de ? "Zuordnung prüfen" : "Review assignment", exact: true }).click();
  await command.getByRole("button", { name: de ? "Einbau für diese Arbeitsgruppe" : "Installation for this working group", exact: true }).click();
  await page.getByRole("option", { name: new RegExp(unitA) }).click();
  await commit(de ? "Teilstand erfassen und Zuordnung speichern" : "Record partial state and save assignment");
  await group.getByRole("button", { name: de ? "Teilstand erfassen" : "Record partial state", exact: true }).click();
  const note = `SYNTHETIC-CHECKPOINT-${language}`;
  await field(de ? "Notiz zur Erfassung" : "Recording note").fill(note);
  current = await commit(de ? "Gespeicherten CV-Teilbestand jetzt erfassen" : "Record saved partial CV state now", 201);
  const summary = current.snapshots.find(value => value.note === note);
  const frozen = await read(`/api/v1/decoder-configurations/snapshots/${summary.id}`);
  assert.equal(frozen.hardware.decoderId, decoderA.id); assert.equal(frozen.hardware.installedAt, null);
  assert.equal(frozen.payload.values.find(value => value.id === base.id).value, base.value);
  const installation = panel.locator(".decoder-installation").filter({ has: page.getByText(unitA, { exact: true }) });
  await installation.getByRole("button", { name: de ? "Exemplar bearbeiten" : "Edit unit", exact: true }).click();
  await field(de ? "Modell" : "Model").fill("Corrected description");
  await commit(de ? "Exemplarangaben jetzt speichern" : "Save unit details now");
  await installation.getByRole("button", { name: de ? "Ausbau erfassen" : "Record removal", exact: true }).click();
  await commit(de ? "Ausbau jetzt speichern" : "Save removal now");
  current = await read(endpoint);
  assert.equal(current.installations.find(value => value.id === installationA.id).removedAt, null);
  await installation.getByRole("button", { name: de ? "Datumsangaben berichtigen" : "Correct dates", exact: true }).click();
  await field(de ? "Einbau" : "Installed").fill("2026-10-01");
  await field(de ? "Ausbau" : "Removed").fill("2026-11-07");
  await commit(de ? "Datumsangaben jetzt speichern" : "Save dates now");
  assert.deepEqual(await read(`/api/v1/decoder-configurations/snapshots/${summary.id}`), frozen);
  const corrected = (await read(endpoint)).installations.find(value => value.id === installationA.id);
  assert.equal(corrected.installedAt, "2026-10-01"); assert.equal(corrected.removedAt, "2026-11-07"); assert.equal(corrected.state, "removed");
  checks.push(`${language}:two-identical-models-distinct-units-unknown-dates-and-frozen-checkpoint`);

  await panel.locator(".decoder-installation").filter({ has: page.getByText(unitB, { exact: true }) })
    .getByRole("button", { name: de ? "Exemplar bearbeiten" : "Edit unit", exact: true }).click();
  await field(de ? "Hersteller" : "Manufacturer").fill("Local brand");
  const concurrent = await read(endpoint), previousUnit = concurrent.instances.find(value => value.id === decoderB.id);
  const changed = await actor.put(`${endpoint}/instances/${decoderB.id}`, { headers: { "X-CSRF-Token": csrfToken },
    data: { id: decoderB.id, expectedRevision: concurrent.revision, label: previousUnit.label, model: previousUnit.model,
      manufacturer: "Remote brand", serialNumber: "Remote serial", notes: "Remote note retained" } });
  assert.equal(changed.status(), 200);
  await commit(de ? "Exemplarangaben jetzt speichern" : "Save unit details now", 409);
  await command.getByRole("button", { name: de ? "Aktuellen Stand laden und prüfen" : "Load and review current state", exact: true }).click();
  const conflict = command.locator(".decoder-command-conflict");
  await conflict.getByText("Remote brand", { exact: true }).waitFor();
  assert.equal(await field(de ? "Hersteller" : "Manufacturer").inputValue(), "Local brand");
  assert.equal(await field(de ? "Seriennummer" : "Serial number").inputValue(), "Remote serial");
  assert.equal(await field(de ? "Notizen" : "Notes").inputValue(), "Remote note retained");
  assert.equal(await command.getByRole("button", { name: de ? "Exemplarangaben jetzt speichern" : "Save unit details now", exact: true }).isDisabled(), true);
  await page.setViewportSize({ width: 390, height: 1000 }); await conflict.scrollIntoViewIfNeeded();
  await page.screenshot({ path: `${input.output}/decoder-conflict-${language}.png` });
  await conflict.getByRole("button", { name: de ? "Meine Eingabe behalten" : "Keep my input", exact: true }).click();
  await commit(de ? "Exemplarangaben jetzt speichern" : "Save unit details now");
  const reconciled = (await read(endpoint)).instances.find(value => value.id === decoderB.id);
  assert.equal(reconciled.manufacturer, "Local brand"); assert.equal(reconciled.serialNumber, "Remote serial");
  assert.equal(reconciled.notes, "Remote note retained"); await page.setViewportSize({ width: 1440, height: 1000 });
  checks.push(`${language}:concurrent-field-choice-preserves-unmodified-serial-and-notes`);

  await panel.getByRole("button", { name: de ? "Arbeitsgruppe anlegen" : "Create working group", exact: true }).click();
  await field(de ? "Bezeichnung der Arbeitsgruppe" : "Working group designation").fill(`SYNTHETIC-GROUP-${language}`);
  let writes = 0;
  const workspaceEndpoint = input.base + endpoint + "/workspaces";
  await page.route(workspaceEndpoint, async route => {
    writes++; const response = await route.fetch(); assert.equal(response.status(), 201); await route.abort("failed");
  });
  try {
    await command.getByRole("button", { name: de ? "Arbeitsgruppe jetzt anlegen" : "Create working group now", exact: true }).click();
    await panel.getByText(de ? "Es fehlt eine eindeutige Speicherbestätigung. Lade den aktuellen Stand, bevor du erneut bestätigst."
      : "There is no definite save acknowledgment. Load the current state before confirming again.", { exact: true }).waitFor();
    assert.equal(await command.getByRole("button", { name: de ? "Arbeitsgruppe jetzt anlegen" : "Create working group now", exact: true }).isDisabled(), true);
    await command.getByRole("button", { name: de ? "Aktuellen Stand laden und prüfen" : "Load and review current state", exact: true }).click();
    await command.waitFor({ state: "hidden" }); assert.equal(writes, 1);
  } finally { await page.unroute(workspaceEndpoint); }
  const createdGroup = (await read(endpoint)).workspaces.find(value => value.profile === `SYNTHETIC-GROUP-${language}`);
  assert.equal(createdGroup.valueCount, 0); assert.equal(createdGroup.installationId, null);
  checks.push(`${language}:lost-working-group-response-reconciled-without-replay`);

  for (const width of [2580, 1440, 820, 390]) {
    for (const theme of ["light", "dark"]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
      await page.evaluate(() => document.fonts.ready);
      await panel.locator(".decoder-configurations-toolbar").scrollIntoViewIfNeeded();
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
      assert.equal(await editor.locator(".modal-body").evaluate(element => element.scrollWidth > element.clientWidth + 1), false);
      await page.screenshot({ path: `${input.output}/management-${language}-${theme}-${width}.png` });
      checks.push(`${language}:${theme}:${width}:decoder-management`);
    }
  }
  await newUnit().click();
  await field(de ? "Bezeichnung des Exemplars" : "Unit designation").fill("Lange individuelle Decoderbezeichnung für getrennte Exemplare desselben Modells");
  await field(de ? "Notizen" : "Notes").fill("Uncommitted unit input retained across tabs.");
  for (const width of [2580, 1440, 820, 390]) {
    for (const theme of ["light", "dark"]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
      await command.scrollIntoViewIfNeeded();
      await page.screenshot({ path: `${input.output}/decoder-command-${language}-${theme}-${width}.png` });
      for (const button of await command.getByRole("button").all()) {
        const bounds = await button.boundingBox();
        const valid = bounds && bounds.height >= 44 && bounds.x >= 0 && bounds.x + bounds.width <= width + 1;
        if (!valid) {
          const style = await button.evaluate(element => { const value = getComputedStyle(element);
          return { minHeight: value.minHeight, minBlockSize: value.minBlockSize, height: value.height, transform: value.transform, zoom: value.zoom,
            rules: Array.from(document.styleSheets).flatMap(sheet => Array.from(sheet.cssRules)).filter(rule => rule.selectorText &&
              element.matches(rule.selectorText) && (rule.style.minHeight || rule.style.minBlockSize)).map(rule => rule.cssText).slice(-6) };
        });
          assert.fail(`${language}/${theme}/${width}: ${await button.textContent()} ${JSON.stringify({ bounds, style })}`);
        }
      }
    }
  }
  await tab(de ? "Modell" : "Model"); assert.equal(await designation().inputValue(), "UNSAVED-DECODER-CORE");
  await tab(de ? "Steuerung" : "Control");
  assert.equal(await editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).inputValue(), "UNSAVED-DECODER-FUNCTION");
  await tab("CV");
  await editor.locator(".cv-subtabs").getByRole("button", { name: de ? "Decoder und Arbeitsgruppen" : "Decoders and working groups", exact: true }).click();
  await field(de ? "Notizen" : "Notes").waitFor();
  assert.equal(await field(de ? "Notizen" : "Notes").inputValue(), "Uncommitted unit input retained across tabs.");
  await command.getByRole("button", { name: de ? "Diese Aktion verwerfen" : "Discard this action", exact: true }).click();
  await editor.locator(".cv-subtabs button").nth(1).click();
  await editor.locator(".cv-draft-review").waitFor();
  await editor.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  assert.deepEqual(await read(path), before);
  assert.equal((await read(endpoint)).installations.find(value => value.id === installationA.id).state, "removed");
  checks.push(`${language}:decoder-actions-preserve-cv-function-core-and-command-drafts`);
  return checks;
}

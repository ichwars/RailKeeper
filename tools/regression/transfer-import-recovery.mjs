import assert from "node:assert/strict";

export async function assertTransferImportRecovery(page, input, de, reference) {
  const { actor, csrfToken, origin } = reference, language = de ? "de" : "en";
  const headers = { "X-CSRF-Token": csrfToken }, name = `Reply recovery ${language}`;
  const inventory = `RK-REPLY-${language.toUpperCase()}`;
  const created = await actor.post("/api/v1/data-transfer/profiles", { headers, data: {
    name, direction: "import", format: "csv", areas: ["vehicles"], options: {}, enabled: true
  } }); assert.equal(created.status(), 201);
  await page.goto(input.base + "/import-export", { waitUntil: "domcontentloaded" });
  await page.getByRole("button", { name: de ? "Neuer Import" : "New import", exact: true }).click();
  const dialog = page.getByRole("dialog", { name: de ? "Import prüfen" : "Review import", exact: true });
  await dialog.getByRole("button", { name: de ? "Importprofil" : "Import profile", exact: true }).click();
  await page.getByRole("option", { name, exact: true }).click();
  const fileButton = dialog.getByRole("button", { name: de ? "Importdatei" : "Import file", exact: true });
  const [chooser] = await Promise.all([page.waitForEvent("filechooser"), fileButton.click()]);
  const [uploaded] = await Promise.all([
    page.waitForResponse(response => response.request().method() === "POST" && response.url().endsWith("/upload")),
    chooser.setFiles({
    name: `reply-recovery-${language}.csv`, mimeType: "text/csv", buffer: Buffer.from(
      "Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n" +
      `${inventory};Roco;Reply recovery ${language};H0;Lokomotive;Diesellokomotive\n`)
    })
  ]);
  const preview = await uploaded.json(); assert.equal(preview.job.state, "ready");
  await dialog.getByRole("button", { name: de ? "Weiter zur Prüfung" : "Continue to review", exact: true }).click();
  const jobURL = origin.origin + "/api/v1/data-transfer/jobs/" + preview.job.id;
  let confirms = 0, failedRead = false;
  await page.route(jobURL, async route => {
    if (route.request().method() === "GET" && !failedRead) { failedRead = true; return route.abort("failed"); }
    return route.continue();
  });
  await page.route(jobURL + "/confirm", async route => {
    confirms++;
    const response = await route.fetch(); assert.equal(response.status(), 200);
    return route.abort("failed");
  });
  await dialog.getByRole("button", { name: de ? "1 Datensatz importieren" : "Import 1 record", exact: true }).click();
  await dialog.getByRole("alert").filter({ hasText: de ? "noch unklar" : "still uncertain" }).waitFor();
  assert.equal(await fileButton.isDisabled(), true);
  assert.equal(await dialog.getByRole("button", { name: de ? "Auftrag abbrechen" : "Cancel job", exact: true }).count(), 0);
  const check = dialog.getByRole("button", { name: de ? "Auftragsstatus prüfen" : "Check job status", exact: true });
  for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 1000 });
    await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
    await page.keyboard.press("Tab"); await check.focus();
    assert.ok(await check.evaluate(button => document.activeElement === button && getComputedStyle(button).outlineStyle !== "none"));
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
    if (width === 390) { const box = await check.boundingBox(); assert.ok(box && box.width >= 44 && box.height >= 44); }
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-transfer-unknown.png` });
  }
  const status = await actor.get(jobURL); assert.equal(status.status(), 200);
  const completed = await status.json(); assert.equal(completed.job.state, "completed");
  await check.click(); await dialog.getByRole("status").filter({ hasText: de ? "bereits abgeschlossen" : "already completed" }).waitFor();
  assert.equal(confirms, 1); assert.equal(await check.count(), 0);
  const vehicles = await actor.get("/api/v1/vehicles"); assert.equal(vehicles.status(), 200);
  const saved = (await vehicles.json()).filter(vehicle => vehicle.inventoryNumber === inventory);
  assert.equal(saved.length, 1);
  for (const revision of [preview.job.revision, completed.job.revision]) {
    const replay = await actor.post(jobURL + "/confirm", { headers, data: { confirm: true, expectedRevision: revision } });
    assert.equal(replay.status(), 409);
  }
  const after = await actor.get(jobURL); assert.deepEqual(await after.json(), completed);
  const afterVehicles = await actor.get("/api/v1/vehicles");
  assert.deepEqual((await afterVehicles.json()).filter(vehicle => vehicle.inventoryNumber === inventory), saved);
  await page.unroute(jobURL); await page.unroute(jobURL + "/confirm");
  await dialog.getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).last().click();
  await page.setViewportSize({ width: 1440, height: 1000 });
  return [language + ":lost-transfer-reply-only-reads-existing-job", language + ":completed-transfer-rejects-replay-without-duplicate"];
}

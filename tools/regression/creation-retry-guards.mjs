import assert from "node:assert/strict";

export async function assertCreationRetryGuards(page, input, de, reference) {
  const { actor, origin } = reference, language = de ? "de" : "en", checks = [];
  const list = async () => { const response = await actor.get("/api/v1/vehicles"); assert.equal(response.status(), 200); return response.json(); };
  const original = (await list()).find(vehicle => vehicle.id === input.fixture.vehicleId);
  for (const kind of ["single", "set"]) {
    const name = `Creation retry ${kind} ${language}`, before = await list();
    const wizard = page.locator(".vehicle-create-wizard");
    const submit = () => wizard.locator('.vehicle-wizard-actions button[type="submit"]');
    const choose = async (label, value) => {
      await wizard.locator("label").filter({ hasText: label }).locator(".app-select-trigger").click();
      await page.getByRole("option", { name: value, exact: true }).click();
    };
    const open = async () => {
      await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
      await page.locator(".new-vehicle-button").click();
      if (kind === "set") await wizard.getByRole("radio").nth(1).click();
      await choose(de ? "Hersteller" : "Manufacturer", original.manufacturer);
      await wizard.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).fill(name);
      await choose(de ? "Spurweite" : "Gauge", original.gauge);
      await choose(de ? "Kategorie" : "Category", original.category);
      await submit().click(); await submit().click();
      if (kind === "set") await wizard.locator('.vehicle-set-details-tabbar [role="tab"]').first().click();
      await choose(de ? "Gattung" : "Subtype", original.gattung);
    };
    const endpoint = origin.origin + (kind === "set" ? "/api/v1/vehicle-sets" : "/api/v1/vehicles");
    const identities = [], results = [];
    let heldReady, releaseHeld;
    const held = new Promise(resolve => { heldReady = resolve; }), released = new Promise(resolve => { releaseHeld = resolve; });
    await page.route(endpoint, async route => {
      if (route.request().method() !== "POST") return route.continue();
      const body = route.request().postDataJSON(); assert.match(body.creationId, /^[0-9a-f]{32}$/);
      identities.push(body.creationId);
      const response = await route.fetch();
      if (identities.length === 2) {
        assert.equal(body.creationId, identities[0]); assert.equal(response.status(), 409);
        assert.equal((await response.json()).error, "vehicle_creation_conflict");
        return route.fulfill({ response });
      }
      assert.equal(response.status(), 201);
      const created = await response.json(); assert.equal(created.id, body.creationId); results.push(created);
      if (identities.length === 1) return route.abort("failed");
      if (identities.length === 4) { heldReady(); await released; }
      assert.notEqual(body.creationId, identities[0]); return route.fulfill({ response });
    });
    try {
      await open(); await submit().click();
      await wizard.locator(".form-message").filter({ hasText: de ? "Antwort beim Anlegen ist unklar" : "creation reply is uncertain" }).waitFor();
      const afterFirst = await list(), createdIDs = kind === "set" ? results[0].members.map(member => member.id) : [results[0].id];
      assert.equal(afterFirst.length, before.length + createdIDs.length);
      assert.deepEqual(afterFirst.filter(vehicle => !createdIDs.includes(vehicle.id)), before);
      await submit().click();
      await wizard.locator(".form-message").filter({ hasText: de ? "Anlegevorgang ist bereits gespeichert" : "creation is already stored" }).waitFor();
      assert.deepEqual(await list(), afterFirst);
      await wizard.locator(".vehicle-create-head button").click();
      await open();
      const accepted = page.waitForResponse(response => response.url() === endpoint && response.request().method() === "POST");
      await submit().click(); assert.equal((await accepted).status(), 201);
      await wizard.waitFor({ state: "hidden" });
      const afterNew = await list(), newIDs = kind === "set" ? results[1].members.map(member => member.id) : [results[1].id];
      assert.equal(afterNew.length, afterFirst.length + newIDs.length);
      assert.deepEqual(afterNew.filter(vehicle => !newIDs.includes(vehicle.id)), afterFirst);
      assert.equal(identities.length, 3);
      await page.locator(".vehicle-edit-modal .modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
      await open(); await submit().click(); await held;
      await wizard.locator(".vehicle-create-head button").click();
      const originalResponse = await actor.get("/api/v1/vehicles/" + original.id), originalDetail = await originalResponse.json();
      const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: original.name, exact: true }) });
      await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
      const editor = page.locator(".vehicle-edit-modal");
      const designation = editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
      const retained = `Retained after ${kind} reply ${language}`; await designation.fill(retained);
      let staleReads = 0;
      const observe = request => { if (request.method() === "GET" && request.url() === origin.origin + "/api/v1/vehicles/" + results[2].id) staleReads++; };
      page.on("request", observe);
      try {
        const completed = page.waitForResponse(response => response.url() === endpoint && response.request().method() === "POST");
        releaseHeld(); assert.equal((await completed).status(), 201);
        await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
        assert.equal(await designation.inputValue(), retained); assert.equal(staleReads, 0);
        assert.equal(await editor.locator(".modal-actions .primary-button").isEnabled(), true);
        const unchanged = await actor.get("/api/v1/vehicles/" + original.id); assert.deepEqual(await unchanged.json(), originalDetail);
      } finally { page.off("request", observe); }
      await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    } finally { releaseHeld(); await page.unroute(endpoint); }
    checks.push(`${language}:${kind}-lost-creation-reply-retains-one-result`, `${language}:${kind}-creation-retry-preserves-all-inventory`,
      `${language}:${kind}-new-identical-draft-creates-a-distinct-owned-result`);
    checks.push(`${language}:${kind}-late-create-reply-preserves-another-editor`, `${language}:${kind}-late-create-starts-no-ancillary-requests`);
  }
  return checks;
}

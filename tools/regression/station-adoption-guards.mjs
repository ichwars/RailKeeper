import assert from "node:assert/strict";

export async function assertStationAdoptionGuards(page, input, de, reference) {
  const { actor, origin } = reference, language = de ? "de" : "en", checks = [];
  const readList = async () => { const response = await actor.get("/api/v1/vehicles"); assert.equal(response.status(), 200); return response.json(); };
  for (const provider of ["ecos", "cs3"]) {
    const name = `Mapped ${provider} ${language}`, before = await readList();
    const externalMapping = { provider, externalId: de ? "9001" : "9002", externalName: name, syncStatus: "linked" };
    await page.evaluate(draft => sessionStorage.setItem("railkeeper.ecosVehicleDraft", JSON.stringify(draft)), {
      source: provider, mode: "create", sourceSummary: { objectId: Number(externalMapping.externalId), name, address: "3", protocol: "DCC", profile: "" },
      vehicle: { name, manufacturer: "Roco", gauge: "H0", category: "Lokomotive", gattung: "Diesellokomotive", digital: true },
      importedKeys: ["name", "manufacturer", "gauge", "category", "gattung", "digital"],
      externalMapping, cvValues: [], functionValues: [], unclearFields: []
    });
    await page.goto(input.base + "/vehicles?source=" + provider, { waitUntil: "domcontentloaded" });
    const wizard = page.locator(".vehicle-create-wizard"), submit = () => wizard.locator('.vehicle-wizard-actions button[type="submit"]');
    await submit().click(); await submit().click();
    const endpoint = origin.origin + "/api/v1/vehicles";
    let requests = 0, creationId;
    await page.route(endpoint, async route => {
      if (route.request().method() !== "POST") return route.continue();
      requests++; const body = route.request().postDataJSON(); assert.deepEqual(body.externalMapping, externalMapping);
      assert.match(body.creationId, /^[0-9a-f]{32}$/);
      if (creationId) assert.equal(body.creationId, creationId); else creationId = body.creationId;
      if (requests > 1) return route.continue();
      const response = await route.fetch(); assert.equal(response.status(), 201);
      const vehicle = await response.json(); assert.equal(vehicle.externalMappings[0].vehicleId, vehicle.id);
      return route.abort("failed");
    });
    await submit().click(); await wizard.locator(".form-message").waitFor();
    await page.waitForFunction(() => !document.querySelector('.vehicle-wizard-actions button[type="submit"]')?.disabled);
    const saved = (await readList()).filter(vehicle => vehicle.name === name);
    assert.equal(saved.length, 1);
    const replay = page.waitForResponse(response => response.url() === endpoint && response.request().method() === "POST");
    await submit().click(); assert.equal((await replay).status(), 409);
    await wizard.locator(".form-message").filter({ hasText: de ? "Anlegevorgang ist bereits gespeichert" : "creation is already stored" }).waitFor();
    const after = await readList(); assert.deepEqual(after.filter(vehicle => vehicle.name === name), saved);
    assert.deepEqual(after.filter(vehicle => vehicle.id !== saved[0].id), before, "Source replay changed other inventory");
    assert.equal(requests, 2);
    if (provider === "cs3") for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
      await page.setViewportSize({ width, height: width === 390 ? 844 : 1000 });
      await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
      await page.keyboard.press("Tab"); await submit().focus();
      assert.ok(await submit().evaluate(button => document.activeElement === button && getComputedStyle(button).outlineStyle !== "none"));
      const contrast = await submit().evaluate(button => {
        const context = document.createElement("canvas").getContext("2d"), style = getComputedStyle(button);
        const light = color => {
          context.fillStyle = color; context.fillRect(0, 0, 1, 1);
          const rgb = [...context.getImageData(0, 0, 1, 1).data].slice(0, 3).map(value => {
            const channel = value / 255;
            return channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
          });
          return rgb[0] * 0.2126 + rgb[1] * 0.7152 + rgb[2] * 0.0722;
        };
        const foreground = light(style.color), background = light(style.backgroundColor);
        return (Math.max(foreground, background) + 0.05) / (Math.min(foreground, background) + 0.05);
      });
      assert.ok(contrast >= 4.5, `${language}/${theme}/${width}: create action contrast ${contrast}`);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
      if (width === 390) { const box = await submit().boundingBox(); assert.ok(box && box.width >= 44 && box.height >= 44); }
      await page.evaluate(() => document.fonts.ready);
      await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-station-replay.png` });
    }
    await page.unroute(endpoint);
    await wizard.locator(".vehicle-create-head button").click();
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/overview", { waitUntil: "domcontentloaded" });
    checks.push(`${language}:${provider}-lost-create-reply-retains-one-mapped-vehicle`, `${language}:${provider}-source-replay-rolls-back-other-data`);
  }
  return checks;
}

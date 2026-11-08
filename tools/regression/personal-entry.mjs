import assert from "node:assert/strict";

export async function assertFirstPersonalView(page, input, checks) {
  const key = `railkeeper.settings.sidebarPrefs:${input.credentials.username}`;
  const initial = await (await page.request.get(input.base + "/api/v1/profile/settings")).json();
  assert.equal(initial.settings[key], undefined, "first setup unexpectedly has saved navigation");
  for (const language of ["de", "en"]) for (const theme of ["light", "dark"]) {
    const de = language === "de";
    await page.evaluate(({ language, theme }) => {
      localStorage.setItem("railkeeper.settings.language", language);
      window.dispatchEvent(new Event("railkeeper-language-changed"));
      document.documentElement.dataset.theme = theme;
    }, { language, theme });
    for (const width of [390, 820, 1440, 2580]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.getByRole("heading", { name: de ? "Dein Einstieg in RailKeeper" : "Your start in RailKeeper" }).waitFor();
      const choice = page.getByRole("radio", { name: de ? /^Sammlung/ : /^Collection/ });
      assert.equal(await choice.isChecked(), true);
      await choice.focus();
      assert.equal(await choice.evaluate(element => element === document.activeElement), true);
      await page.keyboard.press("ArrowRight");
      assert.equal(await page.getByRole("radio", { name: de ? /^Alle Funktionen/ : /^All functions/ }).isChecked(), true);
      await page.keyboard.press("ArrowLeft");
      assert.equal(await choice.isChecked(), true);
      const start = page.getByRole("button", { name: de ? "Mit dieser Ansicht starten" : "Start with this view", exact: true });
      await start.waitFor();
      assert.equal(await start.isEnabled(), true);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true, "welcome overflows");
      for (const button of await page.getByRole("button").all()) {
        const box = await button.boundingBox();
        assert.ok(box && box.x >= 0 && box.x + box.width <= width + 1 && box.height >= 44, "welcome action is clipped");
      }
      await page.screenshot({ path: `${input.output}/personal-entry-${language}-${theme}-${width}.png` });
      checks.push(`${language}:${theme}:${width}:setup-choice-radio-keyboard-no-overflow`);
    }
  }
  const unchanged = await (await page.request.get(input.base + "/api/v1/profile/settings")).json();
  assert.deepEqual(unchanged, initial, "offering the choice wrote preferences before confirmation");
  await page.evaluate(() => {
    localStorage.setItem("railkeeper.settings.language", "de");
    window.dispatchEvent(new Event("railkeeper-language-changed"));
  });
  const [saved] = await Promise.all([
    page.waitForResponse(response => response.url().endsWith("/api/v1/profile/settings") && response.request().method() === "PUT"),
    page.getByRole("button", { name: "Mit dieser Ansicht starten", exact: true }).click()
  ]);
  assert.equal(saved.status(), 200);
  await page.locator(".layout").waitFor();
  const preference = JSON.parse((await (await page.request.get(input.base + "/api/v1/profile/settings")).json()).settings[key]);
  assert.deepEqual(preference.hidden, ["exhibition", "importExport", "digitalCenters"]);
  const more = page.locator(".nav-more > summary");
  await more.focus();
  await page.keyboard.press("Enter");
  for (const href of ["/exhibition", "/import-export", "/digital-centers"]) {
    assert.equal(await page.locator(`.nav-more a[href="${href}"]`).isVisible(), true, "specialist function disappeared");
  }
  checks.push("first-setup-explicit-collection-save-all-specialist-functions-retained");
}

export async function assertInventoryToolbar(page, input, de, checks) {
  const language = de ? "de" : "en";
  for (const theme of ["light", "dark"]) for (const width of [390, 820, 1440, 2580]) {
    await page.setViewportSize({ width, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    await page.locator(".inventory-toolbar .inventory-labeled-action").first().waitFor();
    await page.evaluate(theme => { document.documentElement.dataset.theme = theme; }, theme);
    const actions = page.locator(".inventory-toolbar .inventory-labeled-action");
    assert.equal(await actions.count(), 5);
    for (const button of await actions.all()) {
      assert.ok((await button.innerText()).trim(), "action has no visible label");
      await button.scrollIntoViewIfNeeded();
      await button.focus();
      assert.equal(await button.evaluate(element => element === document.activeElement), true);
      const box = await button.boundingBox();
      assert.ok(box && box.x >= 0 && box.x + box.width <= width + 1, "toolbar action clipped");
      if (width <= 900) assert.ok(box.height >= 44, "small touch target");
    }
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true, "inventory page overflow");
    await page.screenshot({ path: `${input.output}/inventory-tools-${language}-${theme}-${width}.png` });
    checks.push(`${language}:${theme}:${width}:five-visible-toolbar-actions-keyboard-no-overflow`);
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
}

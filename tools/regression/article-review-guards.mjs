import assert from "node:assert/strict";

export async function assertArticleReviewGuards(page, input, de, reference) {
  const { actor, origin } = reference, language = de ? "de" : "en";
  const read = async id => { const response = await actor.get("/api/v1/vehicles/" + id); assert.equal(response.status(), 200); return response.json(); };
  const before = await read(input.fixture.vehicleId), other = await read(input.fixture.secondVehicleId);
  const endpoint = origin.origin + "/api/v1/article-search";
  const first = { source: "manufacturer", title: "Article variant A", url: "https://example.invalid/catalog/reference", snippet: "Fixture", score: 95,
    fields: { ean: { label: "EAN", value: "4011111111111", confidence: 1 }, name: { label: "Name", value: "Unselected name", confidence: 1 } } };
  const second = { ...first, title: "Article variant B", fields: { ...first.fields, ean: { label: "EAN", value: "4022222222222", confidence: 1 } } };
  const response = { query: "Synthetic article comparison", results: [first, second] };
  const open = async (vehicle, navigate = true) => {
    if (navigate) await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: vehicle.name, exact: true }) });
    await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  };
  const editor = page.locator(".vehicle-edit-modal"), search = page.locator(".article-search-layer");
  const run = () => editor.getByRole("button", { name: de ? "Artikeldaten suchen" : "Search article data", exact: true }).click();
  await page.route(endpoint, route => route.fulfill({ json: response }));
  await open(before); await run();
  const card = search.locator(".article-result-card").filter({ hasText: "Article variant B" });
  await card.getByRole("checkbox", { name: /EAN/ }).check();
  for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 1000 });
    await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
    await card.scrollIntoViewIfNeeded();
    await page.keyboard.press("Tab"); await card.getByRole("checkbox", { name: /EAN/ }).focus();
    assert.equal(await card.getByRole("checkbox", { name: /EAN/ }).evaluate(element => document.activeElement === element), true);
    const action = card.getByRole("button", { name: de ? "Ausgewählte Felder übernehmen" : "Apply selected fields", exact: true });
    const contrast = await action.evaluate(button => {
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
      return { image: style.backgroundImage, ratio: (Math.max(foreground, background) + 0.05) / (Math.min(foreground, background) + 0.05) };
    });
    assert.equal(contrast.image, "none"); assert.ok(contrast.ratio >= 4.5);
    if (width === 390) {
      const label = await card.getByRole("checkbox", { name: /EAN/ }).evaluate(element => {
        const bounds = element.closest("label").getBoundingClientRect(); return { width: bounds.width, height: bounds.height };
      });
      assert.ok(label.width >= 44 && label.height >= 44);
      const table = card.locator(".article-field-table-scroll");
      assert.ok(await table.evaluate(element => element.scrollWidth > element.clientWidth));
      await table.focus(); await page.keyboard.press("ArrowRight");
      await page.waitForFunction(() => [...document.querySelectorAll(".article-field-table-scroll")].some(element => element.scrollLeft > 0));
      await table.evaluate(element => element.scrollLeft = 0);
      const bounds = await action.boundingBox(); assert.ok(bounds && bounds.width >= 44 && bounds.height >= 44);
    }
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-article-selection.png` });
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
  await card.getByRole("button", { name: de ? "Ausgewählte Felder übernehmen" : "Apply selected fields", exact: true }).click();
  await search.waitFor({ state: "hidden" });
  assert.equal(await editor.getByRole("textbox", { name: /EAN/ }).inputValue(), second.fields.ean.value);
  assert.equal(await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), before.name);
  assert.deepEqual(await read(before.id), before, "Article adoption must remain a local draft");
  await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  await page.unroute(endpoint);
  let held, release;
  const waiting = new Promise(resolve => { release = resolve; });
  await page.route(endpoint, async route => { held = route; await waiting; await route.fulfill({ json: response }); });
  try {
    await open(before); await run();
    await search.getByText(de ? "Externe Suchtreffer und verfügbare Detailseiten werden geprüft. Bis zum Ergebnis werden keine Daten übernommen."
      : "External search results and available detail pages are being checked. No data is applied before the result is ready.", { exact: true }).waitFor();
    await search.getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).click();
    await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await open(other, false); assert.ok(held);
    const finished = page.waitForResponse(item => item.url() === endpoint);
    release(); await finished;
    assert.equal(await search.count(), 0);
    assert.equal(await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), other.name);
    await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    assert.deepEqual(await read(before.id), before); assert.deepEqual(await read(other.id), other);
  } finally { release(); await page.unroute(endpoint); }
  return [language + ":article-exact-result-with-repeated-url", language + ":article-selection-and-cancel-write-no-data",
    language + ":late-article-response-does-not-affect-another-vehicle"];
}

import assert from "node:assert/strict";

export async function assertSaveErrorRecovery(page, input, de, checks) {
  const language = de ? "de" : "en", settingsURL = input.base + "/api/v1/system/digital-settings";
  const initial = await (await page.request.get(settingsURL)).json();
  const session = await (await page.request.get(input.base + "/api/v1/auth/session")).json();
  const headers = { "X-CSRF-Token": session.csrfToken, Origin: input.base };
  const reject = route => route.request().method() === "PUT" ? route.fulfill({ status: 503,
    contentType: "application/json", body: JSON.stringify({ error: "synthetic_save_failure", message: "Synthetic save failure" }) }) : route.continue();
  try {
    for (const theme of ["light", "dark"]) for (const width of [390, 820, 1440, 2580]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.goto(input.base + "/settings?tab=digital", { waitUntil: "domcontentloaded" });
      await page.locator(".digital-provider-options").waitFor();
      await page.locator(".digital-provider-options button").first().click();
      await page.getByRole("button", { name: de ? "Konfiguration ändern" : "Change configuration", exact: true }).first().click();
      const dialog = page.getByRole("dialog");
      await page.evaluate(theme => { document.documentElement.dataset.theme = theme; }, theme);
      await dialog.getByLabel(de ? "ECoS-IP oder Hostname" : "ECoS IP or hostname", { exact: true }).fill("synthetic.invalid");
      const port = dialog.getByLabel("Port", { exact: true });
      await port.fill(""); assert.equal(await port.inputValue(), "", "clearing a port restored its default");
      await port.pressSequentially("3456"); assert.equal(await port.inputValue(), "3456");
      const metadata = await page.evaluate(() => localStorage.getItem("railkeeper.digital.configurationMeta"));
      await page.route(settingsURL, reject);
      await dialog.getByRole("button", { name: de ? "Speichern" : "Save", exact: true }).click();
      await dialog.getByRole("alert").filter({ hasText: "Synthetic save failure" }).waitFor();
      assert.equal(await port.inputValue(), "3456", "save failure lost port draft");
      assert.equal(await page.evaluate(() => localStorage.getItem("railkeeper.digital.configurationMeta")), metadata);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true);
      await port.focus(); assert.equal(await port.evaluate(element => element === document.activeElement), true);
      await page.screenshot({ path: `${input.output}/configuration-error-${language}-${theme}-${width}.png` });
      await page.unroute(settingsURL, reject);
      const [saved] = await Promise.all([
        page.waitForResponse(response => response.url() === settingsURL && response.request().method() === "PUT"),
        dialog.getByRole("button", { name: de ? "Speichern" : "Save", exact: true }).click()
      ]);
      assert.equal(saved.status(), 200); await dialog.waitFor({ state: "hidden" });
      assert.equal((await (await page.request.get(settingsURL)).json()).ecos.port, "3456");
      checks.push(`${language}:${theme}:${width}:port-empty-failed-save-retained-explicit-retry`);
    }
  } finally {
    await page.unroute(settingsURL, reject);
    const restored = await page.request.put(settingsURL, { headers, data: initial });
    assert.equal(restored.status(), 200);
    assert.deepEqual(await restored.json(), initial);
  }

  const workspaceURL = input.base + `/api/v1/exhibition-lists/${input.fixture.eventId}/workspace`;
  const before = await (await page.request.get(workspaceURL)).json();
  const entriesURL = input.base + `/api/v1/exhibition-lists/${input.fixture.eventId}/entries`;
  const rejectEntry = route => route.fulfill({ status: 409, contentType: "application/json",
    body: JSON.stringify({ error: "synthetic_entry_failure", message: "Synthetic entry failure" }) });
  for (const theme of ["light", "dark"]) for (const width of [390, 820, 1440, 2580]) {
    await page.setViewportSize({ width, height: 1000 });
    await page.goto(input.base + "/exhibition", { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: de ? "Eintrag hinzufügen" : "Add entry", exact: true }).click();
    const dialog = page.getByRole("dialog");
    await page.evaluate(theme => { document.documentElement.dataset.theme = theme; }, theme);
    await dialog.getByLabel(de ? "Besitzer" : "Owner", { exact: true }).fill("Retained owner");
    await dialog.getByLabel(de ? "Lok-Bezeichnung" : "Locomotive name", { exact: true }).fill("Retained guest");
    await page.route(entriesURL, rejectEntry);
    await dialog.getByRole("button", { name: de ? "Speichern" : "Save", exact: true }).click();
    await dialog.getByRole("alert").filter({ hasText: "Synthetic entry failure" }).waitFor();
    assert.equal(await dialog.getByLabel(de ? "Lok-Bezeichnung" : "Locomotive name", { exact: true }).inputValue(), "Retained guest");
    assert.equal(await page.getByRole("alert").count(), 1, "error feedback is duplicated behind the dialog");
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true);
    await page.screenshot({ path: `${input.output}/exhibition-error-${language}-${theme}-${width}.png` });
    await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await dialog.waitFor({ state: "hidden" }); await page.unroute(entriesURL, rejectEntry);
    assert.deepEqual((await (await page.request.get(workspaceURL)).json()).entries, before.entries);
    checks.push(`${language}:${theme}:${width}:entry-rejection-visible-draft-retained-no-write`);
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
}

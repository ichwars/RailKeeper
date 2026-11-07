import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import { readFile } from "node:fs/promises";

export async function assertDecoderArchive(page, input, de, { actor, csrfToken }) {
  const language = de ? "de" : "en", checks = [];
  const path = `/api/v1/vehicles/${encodeURIComponent(input.fixture.vehicleId)}`;
  const read = async endpoint => {
    const response = await actor.get(endpoint); assert.equal(response.status(), 200); return response.json();
  };
  const before = await read(path), context = await read(path + "/decoder-configurations");
  const workspace = context.workspaces.find(value => value.profile === "Legacy-Profil");
  assert.ok(workspace);
  const id = randomBytes(16).toString("hex"), note = `SYNTHETIC-ARCHIVE-${language}`;
  const capture = await actor.post(`${path}/decoder-configurations/workspaces/${workspace.id}/checkpoints`, {
    headers: { "X-CSRF-Token": csrfToken }, data: { id, expectedRevision: context.revision, note }
  });
  assert.equal(capture.status(), 201);
  const snapshot = await read(`/api/v1/decoder-configurations/snapshots/${id}`);
  const writes = [], observe = request => {
    if (request.url().includes("/api/v1/") && !["GET", "HEAD"].includes(request.method())) writes.push(request.url());
  };
  page.on("request", observe);
  try {
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const open = page.getByRole("button", { name: de ? "CV-Archiv" : "CV archive", exact: true });
    await open.click();
    const dialog = page.locator(".decoder-archive-layer");
    const close = dialog.getByRole("button", { name: de ? "Schließen" : "Close", exact: true });
    await close.waitFor();
    assert.equal(await close.evaluate(element => element === document.activeElement), true);
    await dialog.getByRole("searchbox", { name: de ? "Teilstände durchsuchen" : "Search recorded states", exact: true }).fill(note);
    await dialog.getByText(note, { exact: true }).waitFor();
    assert.ok(await dialog.getByText("CV31=0 / CV32=0", { exact: true }).count());
    await dialog.locator("details").filter({ hasText: /3 → 5/ }).first().locator("summary").click();
    await dialog.getByText(/3 → 5/).first().waitFor();
    const source = dialog.getByRole("link", { name: de ? "Originaldatei herunterladen" : "Download original file", exact: true }).first();
    const original = await actor.get(await source.getAttribute("href"));
    assert.equal(original.status(), 200);
    assert.equal((await original.body()).toString(), "CV;Value\n1;3\n8;151\n");
    const downloaded = page.waitForEvent("download");
    await dialog.getByRole("button", { name: de ? "Teilstand als JSON herunterladen" : "Download recorded state as JSON", exact: true }).click();
    assert.deepEqual(JSON.parse(await readFile(await (await downloaded).path(), "utf8")), snapshot);
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        await page.evaluate(() => document.fonts.ready);
        const bounds = await dialog.locator(".decoder-archive-dialog").boundingBox();
        assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1, `${language}/${theme}/${width}: archive outside viewport`);
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1), false);
        assert.equal(await dialog.locator(".modal-body").evaluate(element => element.scrollWidth > element.clientWidth + 1), false);
        for (const control of await dialog.locator("button, input, a").all()) {
          const size = await control.boundingBox();
          assert.ok(size && size.height >= 44, `${language}/${theme}/${width}: archive action below 44px`);
        }
        await dialog.locator(".modal-body").evaluate(element => element.scrollTop = 0);
        await dialog.locator(".table-wrap").evaluate(element => element.scrollTop = 0);
        await page.screenshot({ path: `${input.output}/archive-${language}-${theme}-${width}.png` });
        checks.push(`${language}:${theme}:${width}:archive`);
      }
    }
    await close.focus(); await page.keyboard.press("Shift+Tab");
    assert.equal(await dialog.evaluate(element => element.contains(document.activeElement)), true);
    await page.keyboard.press("Tab");
    assert.equal(await close.evaluate(element => element === document.activeElement), true);
    await dialog.locator(".decoder-archive-intro").click();
    await page.keyboard.press("Shift+Tab");
    assert.equal(await dialog.evaluate(element => element.contains(document.activeElement)), true, "Clicking prose must not release the modal focus trap");
    await page.keyboard.press("Escape"); await dialog.waitFor({ state: "hidden" });
    assert.equal(await open.evaluate(element => element === document.activeElement), true);

    await page.setViewportSize({ width: 1440, height: 1000 });
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
    await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    const editor = page.getByRole("dialog");
    await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).fill("UNSAVED-ARCHIVE-CORE");
    await editor.getByRole("button", { name: "CV", exact: true }).click();
    await editor.locator(".cv-subtabs").getByRole("button", { name: de ? "CV-Archiv" : "CV archive", exact: true }).click();
    await editor.getByRole("searchbox", { name: de ? "Teilstände durchsuchen" : "Search recorded states", exact: true }).fill(note);
    await page.keyboard.press("Enter"); await editor.getByText(note, { exact: true }).waitFor();
    await editor.getByRole("button", { name: de ? "Modell" : "Model", exact: true }).click();
    assert.equal(await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), "UNSAVED-ARCHIVE-CORE");
    await editor.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    assert.deepEqual(await read(path), before);
    assert.deepEqual(writes, []);
    checks.push(`${language}:archive-original-json-focus-and-vehicle-draft-preserved`);

    let mode = "loading", release;
    const pending = new Promise(resolve => { release = resolve; });
    const endpoint = input.base + "/api/v1/decoder-configurations/snapshots";
    await page.route(endpoint, async route => {
      if (mode === "loading") await pending;
      if (mode === "error") return route.fulfill({ status: 503, contentType: "application/json", body: '{"error":"synthetic_unavailable"}' });
      if (mode === "empty") return route.fulfill({ status: 200, contentType: "application/json", body: "[]" });
      return route.continue();
    });
    try {
      await page.setViewportSize({ width: 390, height: 1000 });
      await open.click();
      await dialog.getByText(de ? "Gespeicherte Teilstände werden geladen …" : "Loading recorded states …", { exact: true }).waitFor();
      await page.screenshot({ path: `${input.output}/archive-${language}-loading.png` });
      mode = "error"; release();
      await dialog.getByRole("alert").waitFor();
      await page.screenshot({ path: `${input.output}/archive-${language}-error.png` });
      mode = "empty";
      await dialog.getByRole("button", { name: de ? "Neu laden" : "Reload", exact: true }).click();
      await dialog.getByText(de ? "Noch keine CV-Teilstände vorhanden." : "No recorded CV states yet.", { exact: true }).waitFor();
      await page.screenshot({ path: `${input.output}/archive-${language}-empty.png` });
      mode = "live";
      await dialog.getByRole("button", { name: de ? "Neu laden" : "Reload", exact: true }).click();
      await dialog.locator(".decoder-archive-record").first().waitFor();
      await close.click();
      checks.push(`${language}:archive-loading-error-empty-recovery`);
    } finally { release(); await page.unroute(endpoint); }
  } finally { page.off("request", observe); }
  return checks;
}

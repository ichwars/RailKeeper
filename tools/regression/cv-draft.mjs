import assert from "node:assert/strict";
import { referenceActor } from "./browser-actor.mjs";

const inputFields = ["cvNumber", "value", "decoderProfile", "indexCV31", "indexCV32", "description", "category",
  "protocol", "sourceFileId", "sourceLine", "sourceExportedAt"];
const editable = value => Object.fromEntries(inputFields.map(field => [field, value[field]]));

export async function assertCVDraftWorkflow(page, input, de, reference) {
  const language = de ? "de" : "en";
  const { actor, csrfToken, origin } = reference || await referenceActor(input);
  try {
    const path = "/api/v1/vehicles/" + encodeURIComponent(input.fixture.vehicleId);
    const read = async () => {
      const response = await actor.get(path);
      assert.equal(response.status(), 200);
      return response.json();
    };
    const before = await read();
    const base = before.cvValues.find(value => value.cvNumber === 1 && value.decoderProfile === "Legacy-Profil");
    const indexed = before.cvValues.find(value => value.cvNumber === 257 && value.decoderProfile === "Legacy-Profil"
      && value.indexCV31 === 0 && value.indexCV32 === 0);
    assert.ok(base && indexed);
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const edit = async () => {
      const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
      await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    };
    await edit();
    const dialog = page.getByRole("dialog");
    const tab = name => dialog.getByRole("button", { name, exact: true }).click();
    const raw = () => dialog.locator(".cv-subtabs button").nth(1).click();
    const saveDraft = () => dialog.getByRole("button", { name: de ? "CV-Entwurf speichern" : "Save CV draft", exact: true });
    const cancel = () => dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true });
    const stage = async (value, nextValue) => {
      let row = dialog.locator(".cv-values-table tbody tr").filter({ has: page.getByRole("cell", { name: new RegExp(`^CV ${value.cvNumber}(?:\\s|$)`) }) })
        .filter({ has: page.getByRole("cell", { name: value.decoderProfile, exact: true }) });
      if (value.indexCV31 != null) row = row.filter({ hasText: `CV31=${value.indexCV31} / CV32=${value.indexCV32}` });
      else row = row.filter({ hasNotText: "CV31=" });
      await row.getByRole("button", { name: de ? "CV bearbeiten" : "Edit CV", exact: true }).click();
      await dialog.locator(".cv-manual-editor").getByRole("spinbutton", { name: de ? "Wert" : "Value", exact: true }).fill(String(nextValue));
      await dialog.getByRole("button", { name: de ? "Im CV-Entwurf vormerken" : "Stage in CV draft", exact: true }).click();
    };
    const frames = async suffix => {
      for (const width of [2580, 1440, 820, 390]) {
        for (const theme of ["light", "dark"]) {
          await page.setViewportSize({ width, height: 1000 });
          await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
          await dialog.locator(".cv-draft-review").scrollIntoViewIfNeeded();
          const dimensions = await page.evaluate(() => ({ width: window.innerWidth, document: document.documentElement.scrollWidth }));
          assert.ok(dimensions.document <= dimensions.width + 1, `CV draft ${language}/${theme}/${width}: page overflow`);
          const bounds = await dialog.locator(".cv-draft-review").boundingBox();
          assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1);
          if (width === 390) {
            for (const button of await dialog.locator(".cv-draft-actions button, .cv-draft-row-actions button").all()) {
              const size = await button.boundingBox();
              assert.ok(size && size.width >= 44 && size.height >= 44, "CV draft mobile action below 44px");
              assert.ok(size.height <= 88, "CV draft action label is squeezed into too many lines");
            }
            const footer = dialog.locator(".modal-actions"), message = footer.locator(".form-message");
            if (await message.count()) {
              const note = await message.boundingBox();
              assert.ok(note && note.width >= width * 0.8, "CV error must have its own readable footer row");
              for (const button of await footer.getByRole("button").all()) {
                const bounds = await button.boundingBox();
                assert.ok(bounds && note.y + note.height <= bounds.y + 1 && bounds.height >= 44);
              }
            }
          }
          await page.evaluate(() => document.fonts.ready);
          await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-${suffix}.png` });
        }
      }
      await page.setViewportSize({ width: 1440, height: 1000 });
    };

    // Preparing and canceling a change must not send any CV write request.
    const writes = [];
    const recordWrite = req => { if (req.url().includes(path + "/cv-values") && req.method() !== "GET") writes.push(req.method()); };
    page.on("request", recordWrite);
    await tab("CV"); await raw(); await stage(base, base.value + 1);
    assert.equal(await dialog.locator(".cv-draft-review tbody tr").count(), 1);
    await cancel().click(); await dialog.waitFor({ state: "hidden" });
    assert.deepEqual(writes, []);
    assert.deepEqual((await read()).cvValues, before.cvValues);
    page.off("request", recordWrite);

    await edit();
    await dialog.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).fill("UNSAVED-CORE-DURING-CV");
    await tab(de ? "Steuerung" : "Control");
    await dialog.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).fill("UNSAVED-FUNCTION-DURING-CV");
    await tab("CV"); await raw();
    // The first row would be valid. The later, concurrently changed row must roll it back too.
    await stage(indexed, indexed.value + 1); await stage(base, base.value + 1);
    await frames("cv-draft");
    await saveDraft().focus();
    await page.keyboard.press("Tab"); await page.keyboard.press("Shift+Tab");
    const focus = await saveDraft().evaluate(button => {
      const style = getComputedStyle(button);
      return { active: document.activeElement === button,
        visible: (style.outlineStyle !== "none" && parseFloat(style.outlineWidth) > 0) || style.boxShadow !== "none" };
    });
    assert.ok(focus.active && focus.visible, "CV save keyboard focus must be visible");
    const changed = await actor.put(path + "/cv-values/" + encodeURIComponent(base.id), {
      headers: { "X-CSRF-Token": csrfToken },
      data: { ...editable(base), value: base.value + 3, description: "Concurrent note " + language, expectedRevision: base.revision }
    });
    assert.equal(changed.status(), 200);
    const concurrent = await read();

    let release, entered;
    const held = new Promise(resolve => { release = resolve; });
    const started = new Promise(resolve => { entered = resolve; });
    const endpoint = origin.origin + path + "/cv-values/changes";
    await page.context().route(endpoint, async route => { entered(); await held; await route.continue(); }, { times: 1 });
    const rejected = page.waitForResponse(response => response.url() === endpoint && response.request().method() === "POST");
    await saveDraft().click(); await started;
    assert.equal(await cancel().isDisabled(), true);
    assert.equal(await dialog.getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).isDisabled(), true);
    await page.keyboard.press("Escape"); assert.equal(await dialog.isVisible(), true);
    release(); assert.equal((await rejected).status(), 409);
    await dialog.locator(".cv-draft-review [role=alert]").waitFor();
    assert.equal(await dialog.locator(".cv-draft-review tbody tr").count(), 2);
    assert.deepEqual((await read()).cvValues, concurrent.cvValues, "A stale later row must roll back the first row and its history");
    await frames("cv-conflict");
    const conflictRow = dialog.locator(".cv-draft-review tbody tr").filter({ hasText: /CV 1 ·/ });
    await conflictRow.getByRole("button", { name: de ? "Diesen CV als neue Entwurfsbasis verwenden" : "Use this CV as the new draft baseline", exact: true }).click();
    // Comparing reloads values, the working group and the aggregate asynchronously.
    // Start the write response budget only after that explicit comparison completed.
    await dialog.locator(".cv-draft-review [role=alert]").waitFor({ state: "hidden" });
    await page.waitForFunction(() => {
      const button = document.querySelector(".cv-draft-actions button.primary-button");
      return button instanceof HTMLButtonElement && !button.disabled;
    });
    const saved = page.waitForResponse(response => response.url() === endpoint && response.request().method() === "POST");
    await saveDraft().click(); assert.equal((await saved).status(), 200);
    await dialog.locator(".cv-draft-review").waitFor({ state: "hidden" });
    await dialog.locator(".cv-draft-result").waitFor();
    await tab(de ? "Modell" : "Model");
    assert.equal(await dialog.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), "UNSAVED-CORE-DURING-CV");
    await tab(de ? "Steuerung" : "Control");
    assert.equal(await dialog.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).inputValue(), "UNSAVED-FUNCTION-DURING-CV");
    await cancel().click(); await dialog.waitFor({ state: "hidden" });
    const after = await read();
    for (const old of before.cvValues) {
      const value = after.cvValues.find(item => item.id === old.id);
      if (old.id === base.id || old.id === indexed.id) {
        assert.equal(value.value, old.value + 1);
        assert.equal(value.history.length, (old.history?.length || 0) + (old.id === base.id ? 2 : 1));
        assert.equal(value.sourceLine, old.sourceLine);
        assert.equal(value.sourceExportedAt, old.sourceExportedAt);
        if (old.id === base.id) assert.equal(value.description, "Concurrent note " + language);
      } else assert.deepEqual(value, old, "Unselected CV identity changed");
    }
    assert.deepEqual({ ...after, cvValues: before.cvValues }, before, "Manual CV save changed another vehicle field/record");
    return [language + ":cv-stage-cancel-without-write", language + ":cv-stale-batch-rollback-and-explicit-rebase",
      language + ":cv-save-preserves-unselected-data-and-other-drafts", language + ":cv-draft-bounds-focus-and-busy-state"];
  } finally { if (!reference) await actor.dispose(); }
}

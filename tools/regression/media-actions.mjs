import assert from "node:assert/strict";
import { referenceActor } from "./browser-actor.mjs";

export async function assertMediaActionsWorkflow(page, input, de, reference) {
  const { actor, csrfToken, origin } = reference || await referenceActor(input), language = de ? "de" : "en";
  const path = "/api/v1/vehicles/" + encodeURIComponent(input.fixture.vehicleId);
  const read = async () => { const response = await actor.get(path); assert.equal(response.status(), 200); return response.json(); };
  const headers = { "X-CSRF-Token": csrfToken };
  const popup = () => page.locator(".vehicle-media-action-layer");
  const close = async () => {
    await popup().getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).last().click();
    await popup().waitFor({ state: "hidden" });
  };
  try {
    const before = await read();
    let expectedCVs = before.cvValues;
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
    await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    const editor = page.locator(".vehicle-edit-modal");
    const tab = name => editor.getByRole("button", { name, exact: true }).click();
    await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).fill("UNSAVED-MEDIA-ACTIONS");
    await tab("CV"); await editor.locator(".cv-subtabs button").nth(2).click();
    await editor.locator(".cv-upload-options summary").click();
    const profile = () => editor.getByPlaceholder(de ? "Decoder-Profil für neue Dateien" : "Decoder profile for new files", { exact: true });
    const fileNote = () => editor.getByPlaceholder(de ? "Bemerkung für neue Dateien" : "Note for new files", { exact: true });
    await profile().fill("RETAIN-FILE-PROFILE"); await fileNote().fill("RETAIN-FILE-NOTE");
    await tab("Uploads");
    const attachmentInput = editor.locator(".upload-section").nth(1).locator('input[type="file"]');
    const imageInput = editor.locator(".upload-section").first().locator('input[type="file"]');
    const confirm = () => popup().getByRole("button", { name: de ? "Ausgewählte Dateien speichern" : "Save selected files", exact: true }).click();
    const resume = () => popup().getByRole("button", { name: de ? "Offene Dateien fortsetzen" : "Continue remaining files", exact: true });
    const names = ["partial-first", "partial-too-large", "partial-third"].map(name => `${name}-${language}.txt`);
    const chosen = names.map((name, index) => ({ name, mimeType: "text/plain", buffer: index === 1
      ? Buffer.alloc(2 * 1024 * 1024 + 1, 65) : Buffer.from(name) }));
    await attachmentInput.setInputFiles(chosen);
    await popup().waitFor();
    assert.deepEqual(await read(), before, "File selection must not write before confirmation");
    assert.ok(await editor.evaluate(element => element.closest("[role=dialog]").hasAttribute("inert")));
    await popup().getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    assert.deepEqual(await read(), before, "Cancelling confirmation must not write");
    await attachmentInput.setInputFiles(chosen);
    await confirm();
    await popup().getByText(de ? "Abgelehnt, kann erneut versucht werden" : "Rejected, may be retried", { exact: true }).waitFor();
    let current = await read();
    assert.equal(current.attachments.length, before.attachments.length + 1);
    assert.equal(current.attachments.filter(file => file.originalName === names[0]).length, 1);
    assert.deepEqual({ ...current, attachments: before.attachments }, before);
    for (const width of [2580, 1440, 820, 390]) for (const theme of ["light", "dark"]) {
      const height = width === 390 ? 844 : 1000;
      await page.setViewportSize({ width, height });
      await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
      const bounds = await popup().locator("section.vehicle-media-action-dialog").boundingBox();
      assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width && bounds.y >= 0 && bounds.y + bounds.height <= height + 1);
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
      const content = popup().getByRole("region"); await page.keyboard.press("Tab"); await content.focus();
      assert.ok(await content.evaluate(element => document.activeElement === element && getComputedStyle(element).outlineStyle !== "none"));
      if (width === 390) for (const button of await popup().getByRole("button").all()) {
        const size = await button.boundingBox(); assert.ok(size && size.width >= 44 && size.height >= 44);
      }
      const contrasts = await popup().evaluate(layer => {
        const context = document.createElement("canvas").getContext("2d");
        if (!context) throw new Error("Canvas color check unavailable");
        const luminance = rgb => rgb.slice(0, 3).map(channel => {
          const value = channel / 255;
          return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
        }).reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
        return [...layer.querySelectorAll("p, li strong, li span, footer button")].map(element => {
          context.fillStyle = "white"; context.fillRect(0, 0, 1, 1);
          const parents = []; for (let value = element; value; value = value.parentElement) parents.push(value);
          for (const value of parents.reverse()) {
            context.fillStyle = getComputedStyle(value).backgroundColor; context.fillRect(0, 0, 1, 1);
          }
          const background = luminance([...context.getImageData(0, 0, 1, 1).data]);
          context.fillStyle = getComputedStyle(element).color; context.fillRect(0, 0, 1, 1);
          const foreground = luminance([...context.getImageData(0, 0, 1, 1).data]);
          return (Math.max(background, foreground) + 0.05) / (Math.min(background, foreground) + 0.05);
        });
      });
      assert.ok(contrasts.every(value => value >= 4.5), `Media dialog text contrast below 4.5:1: ${contrasts}`);
      await page.evaluate(() => document.fonts.ready);
      await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-media-partial.png` });
    }
    await popup().getByRole("button", { name: de ? `${names[1]} überspringen` : `Skip ${names[1]}`, exact: true }).click();
    await resume().click();
    await popup().getByText(de ? "Gespeichert" : "Saved", { exact: true }).nth(1).waitFor();
    current = await read();
    assert.equal(current.attachments.length, before.attachments.length + 2);
    for (const name of [names[0], names[2]]) assert.equal(current.attachments.filter(file => file.originalName === name).length, 1);
    await close();
    await page.setViewportSize({ width: 1440, height: 1000 });

    const unknownName = `unknown-response-${language}.txt`, uploadURL = origin.origin + path + "/attachments";
    let unknownRequests = 0;
    await page.route(uploadURL, async route => {
      if (route.request().method() !== "POST") return route.continue();
      unknownRequests++;
      const response = await route.fetch(); assert.equal(response.status(), 201);
      await route.abort("failed");
    });
    await attachmentInput.setInputFiles({ name: unknownName, mimeType: "text/plain", buffer: Buffer.from("Stored, reply lost") });
    await confirm();
    await popup().getByText(de ? "Antwort unklar, zuerst Dateiliste prüfen" : "Response unknown, check the file list first", { exact: true }).waitFor();
    assert.equal(await resume().count(), 0);
    assert.equal((await read()).attachments.filter(file => file.originalName === unknownName).length, 1);
    await close(); await page.unroute(uploadURL); assert.equal(unknownRequests, 1);

    // Another writer changes an acknowledged upload before the client receives its receipt.
    const imageURL = origin.origin + path + "/images", png = await (await actor.get(path + "/images/" + before.images[0].id + "/file")).body();
    let imageRequests = 0, storedId = "", actorSnapshot = before;
    await page.route(imageURL, async route => {
      if (route.request().method() !== "POST") return route.continue();
      imageRequests++;
      const response = await route.fetch();
      if (imageRequests === 1) {
        assert.equal(response.status(), 201); const stored = await response.json(); storedId = stored.id;
        const snapshot = await read();
        const changed = await actor.put(path, { headers, data: { ...snapshot, expectedRevision: snapshot.revision,
          expectedImagesRevision: snapshot.imagesRevision, images: snapshot.images.map(image => ({ ...image,
            isPrimary: image.id === storedId, ...(image.id === storedId ? { title: "EXTERNAL-UPLOAD-TITLE" } : {}) })) } });
        assert.equal(changed.status(), 200);
        actorSnapshot = await read();
      } else assert.equal(response.status(), 409);
      await route.fulfill({ response });
    });
    await imageInput.setInputFiles([1, 2].map(index => ({ name: `guarded-${language}-${index}.png`, mimeType: "image/png", buffer: png })));
    await confirm();
    await popup().getByRole("alert").filter({ hasText: de ? "Der Bildbestand" : "The image collection" }).waitFor();
    assert.equal(await resume().isDisabled(), true); assert.equal(imageRequests, 2);
    current = await read(); assert.equal(current.images.length, before.images.length + 1);
    assert.equal(current.images.find(image => image.isPrimary).id, storedId);
    assert.equal(current.images.find(image => image.id === storedId).title, "EXTERNAL-UPLOAD-TITLE");
    await close(); await page.unroute(imageURL);
    assert.ok(await editor.getByRole("textbox", { name: de ? "Bildbeschreibung" : "Image description", exact: true })
      .evaluateAll(inputs => inputs.some(input => input.value === "EXTERNAL-UPLOAD-TITLE")));
    await tab("CV"); await editor.locator(".cv-subtabs button").nth(2).click();
    assert.equal(await profile().inputValue(), "RETAIN-FILE-PROFILE"); assert.equal(await fileNote().inputValue(), "RETAIN-FILE-NOTE");
    await tab("Uploads");
    const firstAttachment = current.attachments.find(file => file.originalName === names[0]);
    const attachmentRow = () => editor.locator(".attachment-row").filter({ hasText: names[0] });
    await attachmentRow().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    assert.ok((await read()).attachments.some(file => file.id === firstAttachment.id));
    await attachmentRow().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    const change = await actor.put(path + "/attachments/" + firstAttachment.id, { headers, data: {
      expectedRevision: firstAttachment.revision, description: "NEWER-DELETE-NOTE", category: firstAttachment.category || "",
      maintenanceId: firstAttachment.maintenanceId || ""
    } }); assert.equal(change.status(), 200);
    await popup().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().getByRole("alert").waitFor();
    assert.equal(await popup().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).isDisabled(), true);
    assert.ok((await read()).attachments.some(file => file.id === firstAttachment.id));
    await popup().getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await tab("CV"); await editor.locator(".cv-subtabs button").nth(2).click();
    const originalName = `confirmed-source-${language}.xml`;
    await editor.locator('.cv-tab input[type="file"]').nth(1).setInputFiles({
      name: originalName, mimeType: "text/xml", buffer: Buffer.from('<decoder name="Source trial"/>')
    });
    await editor.locator(".cv-file-preview").getByRole("button", { name: de ? "Dateien speichern" : "Save files", exact: true }).click();
    await popup().getByText(de ? "Gespeichert" : "Saved", { exact: true }).waitFor();
    await close();
    const file = (await read()).cvFiles.find(file => file.originalName === originalName);
    assert.ok(file); assert.equal(file.decoderProfile, "RETAIN-FILE-PROFILE"); assert.equal(file.description, "RETAIN-FILE-NOTE");
    const fileRow = () => editor.locator(".cv-files-view .attachment-row").filter({ hasText: originalName });
    await fileRow().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().getByText(de ? /Aktuell verknüpfte CV-Werte: 0/ : /Currently linked CV values: 0/).waitFor();
    const created = await actor.post(path + "/cv-values", { headers, data: {
      cvNumber: 991, value: 17, decoderProfile: `Media source ${language}`, description: "Preserve historical source",
      sourceFileId: file.id, sourceLine: 42, sourceExportedAt: "2025-01-02T03:04:05"
    } }); assert.equal(created.status(), 201);
    const cv = await created.json(); expectedCVs = (await read()).cvValues;
    const history = expectedCVs.find(value => value.id === cv.id).history;
    await popup().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().getByRole("alert").waitFor();
    assert.ok((await read()).cvFiles.some(value => value.id === file.id));
    await popup().getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    await fileRow().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().getByText(de ? /Aktuell verknüpfte CV-Werte: 1/ : /Currently linked CV values: 1/).waitFor();
    await popup().getByRole("button", { name: de ? "Löschen" : "Delete", exact: true }).click();
    await popup().waitFor({ state: "hidden" });
    assert.deepEqual((await read()).cvValues, expectedCVs, "Deleting an original must preserve CV IDs, sources and metadata");
    assert.deepEqual((await read()).cvValues.find(value => value.id === cv.id).history, history);
    assert.equal((await actor.get(path + "/cv-files/" + file.id + "/download")).status(), 404);
    await tab(de ? "Modell" : "Model");
    assert.equal(await editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ }).inputValue(), "UNSAVED-MEDIA-ACTIONS");
    await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
    current = await read();
    assert.deepEqual({ ...current, images: before.images, imagesRevision: before.imagesRevision, attachments: before.attachments },
      { ...before, revision: actorSnapshot.revision, updatedAt: actorSnapshot.updatedAt, cvValues: expectedCVs },
      "Confirmed media operations must preserve every unrelated field and draft");
    return [language + ":media-selection-cancel-and-bounded-partial-resume",
      language + ":lost-upload-response-never-resubmitted",
      language + ":image-receipt-chain-rejects-concurrent-change-and-keeps-decoder-fields",
      language + ":stale-deletion-requires-new-confirmation",
      language + ":confirmed-decoder-original-uses-captured-profile-and-note",
      language + ":source-usage-guards-deletion-and-preserves-cv-history"];
  } finally { if (!reference) await actor.dispose(); }
}

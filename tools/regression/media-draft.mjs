import assert from "node:assert/strict";
import { referenceActor } from "./browser-actor.mjs";

export async function assertMediaDraftWorkflow(page, input, de) {
  const { actor, csrfToken, origin } = await referenceActor(input), language = de ? "de" : "en";
  const path = "/api/v1/vehicles/" + encodeURIComponent(input.fixture.vehicleId);
  try {
    const read = async () => {
      const response = await actor.get(path);
      assert.equal(response.status(), 200);
      return response.json();
    };
    const put = (suffix, data) => actor.put(path + suffix, { data, headers: { "X-CSRF-Token": csrfToken } });
    const before = await read(), image = before.images[0];
    const attachment = before.attachments.find(value => value.id === input.fixture.attachmentId);
    assert.ok(image && attachment);
    const fileBytes = async () => {
      const imageFile = await actor.get(path + "/images/" + image.id + "/file");
      const attachmentFile = await actor.get(path + "/attachments/" + attachment.id + "/download");
      assert.equal(imageFile.status(), 200); assert.equal(attachmentFile.status(), 200);
      return { image: await imageFile.body(), attachment: await attachmentFile.body() };
    };
    const payloads = await fileBytes();
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: before.name, exact: true }) });
    await row.getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    const dialog = page.getByRole("dialog"), tab = name => dialog.getByRole("button", { name, exact: true }).click();
    const designation = () => dialog.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
    const save = dialog.locator(".modal-actions button.primary-button");
    const cancel = dialog.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true });
    const ownName = `Confirmed core ${language}`, ownFunction = `Confirmed function ${language}`;
    const ownTitle = `Eigener langer Bildtitel ${language}`, savedTitle = `Neuer gespeicherter Bildtitel ${language}`;
    const ownNote = `Eigene Anhangsbemerkung ${language}`, savedNote = `Neuere Anhangsbemerkung ${language}`;
    await designation().fill(ownName);
    await tab(de ? "Steuerung" : "Control");
    await dialog.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true }).fill(ownFunction);
    await tab("Uploads");
    const imageTitle = () => dialog.getByRole("textbox", { name: de ? "Bildbeschreibung" : "Image description", exact: true }).first();
    const attachmentRow = () => dialog.locator(".attachment-row").filter({ hasText: attachment.originalName });
    const note = () => attachmentRow().getByPlaceholder(de ? "Bemerkung" : "Note", { exact: true });
    await imageTitle().fill(ownTitle); await note().fill(ownNote);
    const coreWriter = await put("", { ...before, images: undefined, name: `Newer core ${language}`, expectedRevision: before.revision });
    assert.equal(coreWriter.status(), 200);
    const afterCoreActor = await read();
    let rejected = page.waitForResponse(response => response.request().method() === "PUT" && response.url() === origin.origin + path);
    await save.click(); assert.equal((await rejected).status(), 409);
    const coreConflict = dialog.locator(".vehicle-save-state .vehicle-refresh-conflicts");
    await coreConflict.waitFor();
    assert.equal(await save.isDisabled(), true);
    assert.deepEqual(await read(), afterCoreActor, "A stale core request must not change any records");
    await tab(de ? "Modell" : "Model"); assert.equal(await designation().inputValue(), ownName);
    await coreConflict.getByRole("button", { name: de ? "Meine Änderungen für diese Felder behalten" : "Keep my changes for these fields", exact: true }).click();
    assert.equal(await save.isEnabled(), true);
    const imageWriter = await put("", { ...afterCoreActor,
      expectedRevision: afterCoreActor.revision, expectedImagesRevision: afterCoreActor.imagesRevision,
      images: afterCoreActor.images.map(value => value.id === image.id ? { ...value, title: savedTitle } : value)
    });
    assert.equal(imageWriter.status(), 200);
    const attachmentWriter = await put("/attachments/" + attachment.id, {
      description: savedNote, category: "Anleitung", maintenanceId: attachment.maintenanceId || "",
      expectedRevision: attachment.revision
    });
    assert.equal(attachmentWriter.status(), 200);
    const afterMediaActor = await read();
    rejected = page.waitForResponse(response => response.request().method() === "PUT" && response.url() === origin.origin + path);
    await save.click(); assert.equal((await rejected).status(), 409);
    const review = dialog.locator(".vehicle-media-conflicts");
    await review.waitFor();
    assert.equal(await review.locator("article").count(), 2);
    assert.equal(await save.isDisabled(), true);
    await dialog.locator(".modal-actions .form-message").waitFor();
    assert.deepEqual(await read(), afterMediaActor, "A stale media request must not partially save core/functions/images");
    await tab("Uploads");
    assert.equal(await imageTitle().inputValue(), ownTitle); assert.equal(await note().inputValue(), ownNote);
    for (const width of [2580, 1440, 820, 390]) {
      for (const theme of ["light", "dark"]) {
        await page.setViewportSize({ width, height: 1000 });
        await page.evaluate(theme => document.documentElement.dataset.theme = theme, theme);
        await review.scrollIntoViewIfNeeded();
        const bounds = await review.boundingBox();
        assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1);
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
        const first = review.getByRole("button").first();
        await page.keyboard.press("Tab"); await first.focus();
        assert.ok(await first.evaluate(button => document.activeElement === button && getComputedStyle(button).outlineStyle !== "none"));
        if (width === 390) {
          const footerNote = await dialog.locator(".modal-actions .form-message").boundingBox();
          assert.ok(footerNote && footerNote.width >= width * 0.8, "Conflict text needs a readable footer row");
        }
        if (width === 390) for (const button of await review.getByRole("button").all()) {
          const size = await button.boundingBox();
          assert.ok(size && size.width >= 44 && size.height >= 44 && size.height <= 88);
        }
        const contrasts = await review.evaluate(panel => {
          const canvas = document.createElement("canvas"); canvas.width = canvas.height = 1;
          const context = canvas.getContext("2d");
          if (!context) throw new Error("Canvas color check unavailable");
          const luminance = rgb => rgb.slice(0, 3).map(channel => {
            const v = channel / 255; return v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
          }).reduce((value, channel, index) => value + channel * [0.2126, 0.7152, 0.0722][index], 0);
          const elements = [...panel.querySelectorAll("p, dd span:first-child"),
            document.querySelector(".modal-actions .form-message")].filter(Boolean);
          return elements.map(element => {
            context.fillStyle = "white"; context.fillRect(0, 0, 1, 1);
            const parents = []; for (let item = element; item; item = item.parentElement) parents.push(item);
            for (const item of parents.reverse()) { context.fillStyle = getComputedStyle(item).backgroundColor; context.fillRect(0, 0, 1, 1); }
            const background = luminance([...context.getImageData(0, 0, 1, 1).data]);
            context.fillStyle = getComputedStyle(element).color; context.fillRect(0, 0, 1, 1);
            const foreground = luminance([...context.getImageData(0, 0, 1, 1).data]);
            return (Math.max(background, foreground) + 0.05) / (Math.min(background, foreground) + 0.05);
          });
        });
        assert.ok(contrasts.every(value => value >= 4.5), `Media conflict text contrast below 4.5:1: ${contrasts}`);
        await page.evaluate(() => document.fonts.ready);
        await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-media-conflict.png` });
      }
    }
    await page.setViewportSize({ width: 1440, height: 1000 });
    const keep = de ? "Meine Änderungen für diese Felder behalten" : "Keep my changes for these fields";
    await review.locator("article").filter({ hasText: ownTitle }).getByRole("button", { name: keep, exact: true }).click();
    await review.locator("article").filter({ hasText: attachment.originalName }).getByRole("button", { name: keep, exact: true }).click();
    await review.waitFor({ state: "hidden" });
    const attachmentSaved = page.waitForResponse(response => response.request().method() === "PUT"
      && response.url() === origin.origin + path + "/attachments/" + attachment.id);
    await attachmentRow().getByRole("button", { name: de ? "Speichern" : "Save", exact: true }).click();
    assert.equal((await attachmentSaved).status(), 200);
    await save.waitFor({ state: "visible" });
    await page.waitForFunction(() => !document.querySelector(".modal-actions button.primary-button")?.disabled);
    const afterAttachmentSave = await read();
    assert.equal(afterAttachmentSave.attachments.find(value => value.id === attachment.id).description, ownNote);
    assert.equal(afterAttachmentSave.attachments.find(value => value.id === attachment.id).category, "Anleitung");
    assert.equal(await imageTitle().inputValue(), ownTitle);
    const successful = page.waitForResponse(response => response.request().method() === "PUT" && response.url() === origin.origin + path);
    await save.click(); assert.equal((await successful).status(), 200);
    await page.waitForFunction(() => !document.querySelector(".modal-actions button.primary-button")?.disabled);
    const after = await read();
    assert.equal(after.name, ownName);
    assert.equal(after.images.find(value => value.id === image.id).title, ownTitle);
    assert.equal(after.functions.find(value => value.functionKey === "F0").name, ownFunction);
    const omit = (value, fields) => Object.fromEntries(Object.entries(value).filter(([field]) => !fields.includes(field)));
    const intended = ["name", "images", "functions", "revision", "imagesRevision", "updatedAt"];
    assert.deepEqual(omit(after, intended), omit(afterAttachmentSave, intended), "All unrelated vehicle/attachment/CV fields must stay exact");
    assert.deepEqual(after.images.map(value => omit(value, ["title", "updatedAt", "revision"])),
      before.images.map(value => omit(value, ["title", "updatedAt", "revision"])));
    assert.deepEqual(after.images.filter(value => value.id !== image.id).map(value => value.title),
      before.images.filter(value => value.id !== image.id).map(value => value.title));
    assert.deepEqual(after.functions.filter(value => value.functionKey !== "F0"), before.functions.filter(value => value.functionKey !== "F0"));
    assert.deepEqual(omit(after.functions.find(value => value.functionKey === "F0"), ["name", "updatedAt", "revision"]),
      omit(before.functions.find(value => value.functionKey === "F0"), ["name", "updatedAt", "revision"]));
    assert.deepEqual(await fileBytes(), payloads);
    await cancel.click(); await dialog.waitFor({ state: "hidden" });
    assert.deepEqual(await read(), after, "Cancel after confirmed saves must preserve those completed actions");
    return [`${language}:stale-core-and-media-requests-roll-back`, `${language}:reviewed-media-save-preserves-unselected-data-and-bytes`];
  } finally { await actor.dispose(); }
}

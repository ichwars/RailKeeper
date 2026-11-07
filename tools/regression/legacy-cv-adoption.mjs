import assert from "node:assert/strict";
import { writeFile } from "node:fs/promises";
import { createHash } from "node:crypto";

export async function assertLegacyCVAdoption(page, input, de, reference) {
  const { actor, csrfToken, origin } = reference, language = de ? "de" : "en", profile = `Legacy batch ${language}`;
  const path = "/api/v1/vehicles/" + input.fixture.vehicleId, headers = { "X-CSRF-Token": csrfToken };
  const read = async () => { const response = await actor.get(path); assert.equal(response.status(), 200); return response.json(); };
  const create = async value => { const response = await actor.post(path + "/cv-values", { headers, data: value }); assert.equal(response.status(), 201, await response.text()); return response.json(); };
  const first = await create({ cvNumber: 257, value: 3, decoderProfile: profile, indexCV31: 0, indexCV32: 0 });
  const unselected = await create({ cvNumber: 799, value: 7, decoderProfile: profile, description: "Keep unselected" });
  const before = await read(), mapping = before.externalMappings.find(value => value.provider === "ecos");
  assert.ok(mapping);
  const imageBytes = async vehicle => Object.fromEntries(await Promise.all(vehicle.images.map(async image => {
    const response = await actor.get(image.url); assert.equal(response.status(), 200);
    return [image.id, createHash("sha256").update(await response.body()).digest("hex")];
  })));
  const bytesBefore = await imageBytes(before);
  await page.goto(input.base + "/overview", { waitUntil: "domcontentloaded" });
  const values = [{ cvNumber: 257, value: 4, decoderProfile: profile, indexCV31: 0, indexCV32: 0 },
    { cvNumber: 701, value: 5, decoderProfile: profile, sourceExportedAt: "2026-08-20T08:00:00" }];
  await page.evaluate(draft => sessionStorage.setItem("railkeeper.ecosVehicleDraft", JSON.stringify(draft)), {
    source: "ecos", mode: "update", targetVehicleId: before.id,
    sourceSummary: { objectId: Number(mapping.externalId), name: "Legacy fixture", address: "3", protocol: "DCC", profile },
    vehicle: {}, importedKeys: [], externalMapping: { provider: mapping.provider, externalId: mapping.externalId,
      externalName: mapping.externalName || "", externalAddress: mapping.externalAddress || "", externalProtocol: mapping.externalProtocol || "", syncStatus: mapping.syncStatus },
    cvValues: values, functionValues: [], unclearFields: []
  });
  await page.goto(input.base + "/vehicles?source=ecos", { waitUntil: "domcontentloaded" });
  const editor = page.locator(".vehicle-edit-modal"), save = editor.locator(".modal-actions button.primary-button");
  await save.waitFor(); let requests = 0;
  await page.route(origin.origin + path + "/cv-values/import", async route => {
    requests++;
    assert.equal(route.request().postDataJSON().values.length, 2);
    const response = await route.fetch(); assert.equal(response.status(), 200);
    return route.abort("failed");
  });
  await save.click();
  await editor.getByText(de ? /Die Antwort ist unklar/ : /The response is unknown/).waitFor();
  const stored = await read(), selected = stored.cvValues.filter(value => value.decoderProfile === profile);
  assert.equal(selected.length, 3); assert.equal(selected.find(value => value.cvNumber === 257).id, first.id);
  assert.deepEqual(selected.find(value => value.cvNumber === 799), unselected);
  assert.equal(selected.find(value => value.cvNumber === 701).sourceExportedAt, values[1].sourceExportedAt);
  const completed = page.waitForResponse(response => response.url() === origin.origin + path && response.request().method() === "GET");
  await save.click(); await completed;
  await page.waitForFunction(() => !document.querySelector(".modal-actions button.primary-button")?.disabled);
  assert.equal(await editor.locator(".form-message").count(), 0, "The reviewed source retry did not complete");
  assert.equal(requests, 1, "The acknowledged CV content must not be submitted again");
  const after = await read();
  assert.deepEqual(after.cvValues, stored.cvValues, "Retry changed CV identities, metadata, histories or unselected values");
  const business = vehicle => {
    const { revision, updatedAt, externalMappings, cvValues, imagesRevision, images, ...fields } = vehicle;
    // The confirmed core save re-submits image metadata and advances its timestamps/revisions.
    return { ...fields, images: images.map(({ revision, updatedAt, ...image }) => image) };
  };
  const fieldsBefore = business(before), fieldsAfter = business(after);
  const changed = Object.keys(fieldsBefore).filter(key => JSON.stringify(fieldsBefore[key]) !== JSON.stringify(fieldsAfter[key]));
  if (changed.length) await writeFile(`${input.output}/${language}-legacy-business-diff.json`, JSON.stringify({ changed,
    before: Object.fromEntries(changed.map(key => [key, fieldsBefore[key]])),
    after: Object.fromEntries(changed.map(key => [key, fieldsAfter[key]])) }, null, 2));
  assert.deepEqual(changed, [], "Legacy CV adoption changed another business field");
  assert.deepEqual(await imageBytes(after), bytesBefore, "Legacy CV adoption changed stored image bytes");
  await page.unroute(origin.origin + path + "/cv-values/import");
  await editor.locator(".modal-actions").getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  assert.deepEqual((await read()).cvValues, stored.cvValues);
  return [language + ":legacy-source-cvs-use-one-guarded-batch", language + ":lost-legacy-cv-reply-retains-unselected-data-without-replay"];
}

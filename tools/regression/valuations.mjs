import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";

export async function assertValuations(page, input, de, { actor, csrfToken }) {
  const language = de ? "de" : "en", checks = [];
  const headers = { "X-CSRF-Token": csrfToken };
  const read = async path => {
    const response = await actor.get(path);
    assert.equal(response.status(), 200);
    return response.json();
  };
  const created = await actor.post("/api/v1/vehicles", { headers, data: {
    inventoryNumber: `SYNTHETIC-VALUATION-${language}`, name: `Manual valuation ${language}`,
    manufacturer: "Synthetic", gauge: "H0", category: "Lokomotive", gattung: "Diesellok", digital: true
  } });
  assert.equal(created.status(), 201);
  const vehicle = await created.json(), path = `/api/v1/vehicles/${vehicle.id}`, endpoint = path + "/valuations";
  const editor = page.locator(".vehicle-edit-modal"), panel = editor.locator(".vehicle-valuations");
  const tab = name => editor.locator("nav.modal-tabs").getByRole("button", { name, exact: true }).click();
  const row = () => page.getByRole("row").filter({ has: page.getByRole("cell", { name: vehicle.name, exact: true }) });
  const open = async () => {
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    await row().getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
  };
  const field = (german, english) => panel.getByLabel(new RegExp(`^${de ? german : english}`));
  const amount = () => field("Betrag", "Amount");
  const source = () => field("Herkunft der Bewertung", "Valuation provenance");
  const save = () => panel.getByRole("button", { name: de ? "Bewertung jetzt speichern" : "Save valuation now", exact: true }).click();
  const acknowledged = () => panel.getByRole("status").filter({ hasText: de ? "ist gespeichert" : "is saved" }).waitFor();
  const close = () => editor.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  const writeResponse = () => page.waitForResponse(response => response.url() === input.base + endpoint && response.request().method() === "POST");
  try {
    const storedFunction = await actor.put(path + "/functions/F0", { headers, data: {
      name: "Stored F0", functionType: "standard", mode: "dauer", directionDependent: null,
      inverted: false, flagsVersion: 2, notes: "Synthetic function draft reference"
    } });
    assert.equal(storedFunction.status(), 200);
    const before = await read(path);
    await open();
    const designation = () => editor.getByRole("textbox", { name: de ? /^Bezeichnung/ : /^Designation/ });
    await designation().fill("UNSAVED-VALUATION-CORE");
    await tab(de ? "Steuerung" : "Control");
    const functionName = () => editor.getByRole("textbox", { name: de ? "F0 Funktionsname" : "F0 Function name", exact: true });
    await functionName().fill("UNSAVED-VALUATION-FUNCTION");
    await tab(de ? "Wertentwicklung" : "Value history");
    await amount().fill("100,123456");
    await field("Bewertungsdatum", "Valuation date").fill("2026-01-01");
    await source().fill("Manual original estimate");
    await amount().press("Enter");
    assert.equal((await read(endpoint)).total, 0, "Enter must not publish the outer vehicle form");
    const [firstResponse] = await Promise.all([writeResponse(), save()]);
    assert.equal(firstResponse.status(), 201);
    const first = await firstResponse.json();
    await acknowledged();
    assert.equal(first.amount, "100.123456");
    assert.equal(first.valuedOn, "2026-01-01");
    assert.notEqual(first.recordedAt.slice(0, 10), first.valuedOn);
    assert.deepEqual(await read(path), before, "valuation save changed another vehicle area");
    await tab(de ? "Modell" : "Model");
    assert.equal(await designation().inputValue(), "UNSAVED-VALUATION-CORE");
    await tab(de ? "Steuerung" : "Control");
    assert.equal(await functionName().inputValue(), "UNSAVED-VALUATION-FUNCTION");
    await tab(de ? "Wertentwicklung" : "Value history");
    await source().fill("UNSAVED-VALUATION-INPUT");
    await tab(de ? "Modell" : "Model");
    await tab(de ? "Wertentwicklung" : "Value history");
    assert.equal(await source().inputValue(), "UNSAVED-VALUATION-INPUT");
    checks.push(`${language}:valuation-save-and-tabs-preserve-independent-inputs`);

    await panel.getByRole("button", { name: de ? "Bewertung berichtigen" : "Correct valuation", exact: true }).click();
    await amount().fill("125.00"); await source().fill("My reviewed correction");
    const peer = await actor.post(endpoint, { headers, data: {
      id: randomBytes(16).toString("hex"), replacesId: first.id, amount: "120.00", currency: "CHF",
      valuedOn: "2026-01-01", source: "Another session's correction"
    } });
    assert.equal(peer.status(), 201);
    const [conflict] = await Promise.all([writeResponse(), save()]);
    assert.equal(conflict.status(), 409);
    assert.equal(await amount().inputValue(), "125.00");
    assert.equal(await source().inputValue(), "My reviewed correction");
    await panel.getByRole("button", { name: de ? "Neu laden" : "Reload", exact: true }).click();
    await panel.getByRole("button", {
      name: de ? "Meine Eingabe als Berichtigung dieses Eintrags vorbereiten" : "Prepare my input as a correction of this entry", exact: true
    }).click();
    assert.equal(await amount().inputValue(), "125.00");
    assert.equal((await read(endpoint)).total, 2, "comparison itself must not write");
    const [corrected] = await Promise.all([writeResponse(), save()]);
    assert.equal(corrected.status(), 201);
    assert.equal((await corrected.json()).replacesId, (await peer.json()).id);
    await acknowledged();
    checks.push(`${language}:valuation-conflict-explicit-comparison-retains-input`);

    await amount().fill("0.000001");
    await field("Währungscode", "Currency code").fill("CHF");
    await field("Bewertungsdatum", "Valuation date").fill("2026-02-01");
    await source().fill("Lost acknowledgment, original source ä");
    let committed;
    const loseResponse = async route => {
      const response = await route.fetch();
      assert.equal(response.status(), 201);
      committed = await response.json();
      await route.abort("failed");
    };
    await page.route(input.base + endpoint, loseResponse);
    try {
      await save();
      await panel.getByRole("alert").filter({ hasText: de ? "Speicherbestätigung" : "save acknowledgment" }).waitFor();
    } finally { await page.unroute(input.base + endpoint, loseResponse); }
    assert.ok(committed);
    assert.equal(await amount().isDisabled(), true);
    const count = (await read(endpoint)).total;
    await close();
    await open(); await tab(de ? "Wertentwicklung" : "Value history");
    assert.equal(await amount().isDisabled(), true, "dismissal lost the uncertain request identity");
    assert.equal(await source().inputValue(), "Lost acknowledgment, original source ä");
    await page.reload({ waitUntil: "domcontentloaded" });
    await row().getByRole("button", { name: de ? "Bearbeiten" : "Edit", exact: true }).click();
    await tab(de ? "Wertentwicklung" : "Value history");
    assert.equal(await amount().isDisabled(), true, "document reload lost the uncertain request identity");
    await panel.getByRole("button", { name: de ? "Gespeicherten Stand prüfen" : "Check saved state", exact: true }).click();
    await panel.getByRole("status").filter({ hasText: de ? "ist gespeichert" : "is saved" }).waitFor();
    assert.equal((await read(endpoint)).total, count, "reading a lost acknowledgment duplicated the valuation");
    checks.push(`${language}:valuation-lost-response-confirmed-without-second-write`);

    const micro = panel.locator(".valuation-history li").filter({ hasText: "0.000001 CHF" });
    await micro.getByRole("button", { name: de ? "Bewertung zurückziehen" : "Withdraw valuation", exact: true }).click();
    const reason = field("Begründung für das Zurückziehen", "Reason for withdrawal");
    assert.equal(await reason.inputValue(), "");
    await reason.fill("Estimate entered for the wrong date");
    const [withdrawn] = await Promise.all([writeResponse(), panel.getByRole("button", {
      name: de ? "Bewertung jetzt zurückziehen" : "Withdraw valuation now", exact: true
    }).click()]);
    assert.equal(withdrawn.status(), 201);
    await acknowledged();
    const history = await read(endpoint + "/export");
    assert.equal(history.entries.length, 5);
    assert.equal(history.entries.find(entry => entry.id === first.id).amount, first.amount);
    assert.equal(history.entries.find(entry => entry.id === first.id).recordedAt, first.recordedAt);
    const summary = await read("/api/v1/valuations/summary?asOf=2026-02-15");
    assert.deepEqual(summary.totals, [{ currency: "EUR", amount: "125.00", vehicles: 1 }]);
    assert.equal((await read("/api/v1/valuations/summary?asOf=2025-12-31")).valuedVehicles, 0);
    assert.deepEqual(await read(path), before);
    await close();
    checks.push(`${language}:valuation-withdrawal-export-dates-and-coverage`);

    for (const theme of ["light", "dark"]) for (const width of [1280, 390]) {
      await page.evaluate(theme => localStorage.setItem("railkeeper.theme", theme), theme);
      await open(); await tab(de ? "Wertentwicklung" : "Value history");
      await panel.locator(".valuation-history li").filter({ hasText: "125.00 EUR" })
        .getByRole("button", { name: de ? "Bewertung berichtigen" : "Correct valuation", exact: true }).click();
      await page.setViewportSize({ width, height: 1000 });
      await page.evaluate(() => document.fonts.ready);
      const activeTab = await editor.locator("nav.modal-tabs button.active").boundingBox();
      assert.ok(activeTab && activeTab.x >= 0 && activeTab.x + activeTab.width <= width + 1,
        "the active valuation tab must remain visible after resizing");
      assert.equal(await panel.evaluate(root => root.scrollWidth <= root.clientWidth + 1), true);
      for (const control of await panel.locator("input,textarea,button,a").all()) {
        await control.scrollIntoViewIfNeeded();
        const bounds = await control.boundingBox();
        assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1, "valuation control outside viewport");
        if (width === 390) assert.ok(bounds.height >= 44,
          `mobile valuation target too small: ${await control.evaluate(element => element.tagName + ":" + element.getAttribute("type"))} height=${bounds.height}`);
      }
      await editor.locator(".modal-body").evaluate(body => { body.scrollTop = 0; });
      await page.screenshot({ path: `${input.output}/${language}-${theme}-${width}-valuations.png` });
      await close();
      checks.push(`${language}:${theme}:${width}:valuation-controls`);
    }
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
    await row().getByRole("button", { name: de ? "Ansehen" : "View", exact: true }).click();
    const readonly = page.locator(".vehicle-read-view .vehicle-valuations");
    await readonly.getByText("125.00 EUR", { exact: true }).waitFor();
    assert.equal(await readonly.locator("input,textarea").count(), 0);
    await page.locator(".vehicle-read-modal-shell").getByRole("button", { name: de ? "Schließen" : "Close", exact: true }).click();
    checks.push(`${language}:valuation-readonly-history`);
    await page.goto(input.base + "/overview", { waitUntil: "domcontentloaded" });
    const card = page.locator(".manual-valuation-summary");
    await card.getByRole("button", { name: de ? "Manuell bewertete Sammlung" : "Manually valued collection", exact: true }).click();
    const cutoff = card.getByLabel(de ? "Stichtag" : "Valuation cutoff", { exact: true });
    const changed = date => page.waitForResponse(response => response.url() === input.base + "/api/v1/valuations/summary?asOf=" + date);
    const setCutoff = async date => { await cutoff.fill(date); await cutoff.press("Tab"); };
    await Promise.all([changed("2025-12-31"), setCutoff("2025-12-31")]);
    await card.getByText(de ? "Für diesen Stichtag liegen keine gültigen Bewertungen vor." : "There are no valid valuations at this date.", { exact: true }).waitFor();
    await Promise.all([changed("2026-02-15"), setCutoff("2026-02-15")]);
    await card.getByText("125.00", { exact: true }).waitFor();
    assert.equal(await card.getByText("EUR", { exact: true }).count(), 1);
    assert.equal(await card.getByText("CHF", { exact: true }).count(), 0);
    for (const width of [1440, 390]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.locator(".main").evaluate(main => { main.scrollTop = main.scrollHeight; });
      const content = await card.boundingBox(), report = await page.locator(".feedback-button").boundingBox();
      const main = await page.locator(".main").boundingBox();
      assert.ok(main);
      const overlapWidth = Math.min(content.x + content.width, main.x + main.width, report.x + report.width)
        - Math.max(content.x, main.x, report.x);
      const overlapHeight = Math.min(content.y + content.height, main.y + main.height, report.y + report.height)
        - Math.max(content.y, main.y, report.y);
      assert.ok(overlapWidth <= 0 || overlapHeight <= 0,
        "The fixed report action must not cover the visible valuation summary");
    }
    checks.push(`${language}:valuation-dashboard-cutoff-shows-unknown-past-and-exact-currency`);
  } finally {
    const removed = await actor.delete(path, { headers });
    assert.equal(removed.status(), 204);
  }
  return checks;
}

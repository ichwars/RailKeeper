import assert from "node:assert/strict";
import { assertMobileActionTargets } from "./touch-targets.mjs";

export async function assertLoginRecoveryLayout(page, { base, width, de, screenshotPrefix }) {
  await page.goto(base + "/settings?tab=auth", { waitUntil: "domcontentloaded" });
  const panel = page.locator(".user-management-card");
  const create = panel.getByRole("button", { name: de ? "Neuer Benutzer" : "New user", exact: true });
  const recover = panel.getByRole("button", { name: de ? "Anmeldung freigeben" : "Unlock login", exact: true }).first();
  await recover.waitFor();
  await assertMobileActionTargets(page);
  for (const button of [create, recover]) {
    await button.scrollIntoViewIfNeeded();
    const bounds = await button.boundingBox();
    assert.ok(bounds && bounds.x >= 0 && bounds.x + bounds.width <= width + 1,
      "account actions must remain inside the viewport");
    if (width <= 600) assert.ok(bounds.height >= 44, "mobile account target must be at least 44px");
    await button.focus();
    assert.equal(await button.evaluate(element => document.activeElement === element), true);
  }
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true);
  await page.screenshot({ path: screenshotPrefix + "-login-recovery.png" });
  await recover.click();
  const dialog = page.getByRole("dialog", { name: de ? "Anmeldung freigeben" : "Unlock login", exact: true });
  await dialog.waitFor();
  await assertMobileActionTargets(page);
  await page.screenshot({ path: screenshotPrefix + "-login-recovery-confirm.png" });
  await dialog.getByRole("button", { name: de ? "Abbrechen" : "Cancel", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
}

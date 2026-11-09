import assert from "node:assert/strict";
import { referenceActor } from "./browser-actor.mjs";

export async function assertLogoutRecovery(page, input, de) {
  const language = de ? "de" : "en", endpoint = input.base + "/api/v1/auth/logout";
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(input.base + "/vehicles", { waitUntil: "domcontentloaded" });
  const logout = () => page.locator(".sidebar-footer").getByRole("button", { name: de ? /^Abmelden/ : /^Sign out/ });
  const login = () => page.getByLabel(de ? "Benutzername" : "Username", { exact: true });

  await page.route(endpoint, route => route.abort("failed"), { times: 1 });
  await logout().click();
  await page.locator(".form-message.error[role=alert]").waitFor();
  assert.equal(await login().count(), 0);
  assert.equal((await page.context().request.get(input.base + "/api/v1/auth/session")).status(), 200);

  await page.route(endpoint, route => route.fulfill({ status: 403, contentType: "application/json",
    body: JSON.stringify({ error: "csrf_required", message: "Synthetic stale CSRF token" }) }), { times: 1 });
  const [revalidated] = await Promise.all([
    page.waitForResponse(response => response.url() === input.base + "/api/v1/auth/session"), logout().click()
  ]);
  assert.equal(revalidated.status(), 200);
  await page.getByRole("alert").filter({ hasText: "Synthetic stale CSRF token" }).waitFor();
  assert.equal(await login().count(), 0);
  await logout().click();
  await login().waitFor();

  // Select the new browser session by its creation, without revoking the baseline
  // client's separate session or reading any installation-local credentials.
  const { actor, csrfToken } = await referenceActor(input);
  try {
    const before = await actor.get("/api/v1/sessions?limit=200");
    assert.equal(before.status(), 200);
    const known = new Set((await before.json()).map(session => session.id));
    await login().fill(input.credentials.username);
    await page.getByLabel(de ? "Passwort" : "Password", { exact: true }).fill(input.credentials.password);
    await page.getByRole("button", { name: de ? "Anmelden" : "Sign in", exact: true }).click();
    await page.waitForURL("**/overview");
    const after = await actor.get("/api/v1/sessions?limit=200");
    assert.equal(after.status(), 200);
    const current = (await after.json()).filter(session => !known.has(session.id)
      && session.username === input.credentials.username && session.active);
    assert.equal(current.length, 1);
    const revoked = await actor.put("/api/v1/sessions/" + encodeURIComponent(current[0].id) + "/revoke", {
      headers: { "X-CSRF-Token": csrfToken }
    });
    assert.equal(revoked.status(), 204);
    const documentOrigin = await page.evaluate(() => performance.timeOrigin);
    const [rejected] = await Promise.all([
      page.waitForResponse(response => response.url() === endpoint), logout().click()
    ]);
    assert.equal(rejected.status(), 403);
    assert.equal((await rejected.json()).error, "csrf_required");
    await login().waitFor();
    assert.equal(await page.evaluate(() => performance.timeOrigin), documentOrigin);
  } finally { await actor.dispose(); }
  return [`${language}:logout-network-failure-preserves-workspace`, `${language}:logout-valid-session-refreshes-csrf-for-retry`,
    `${language}:revoked-session-returns-to-login-without-reload`];
}

import assert from "node:assert/strict";

export async function assertLogoutRecovery(page, input, de, sessionId) {
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
  // Revoke only this already identified browser session. The baseline client's
  // separate session and the production login budgets remain untouched.
  const session = await revalidated.json();
  const revoked = await page.context().request.put(input.base + "/api/v1/sessions/" + encodeURIComponent(sessionId) + "/revoke", {
    headers: { "X-CSRF-Token": session.csrfToken, Origin: input.base }
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
  return [`${language}:logout-network-failure-preserves-workspace`, `${language}:logout-valid-session-revalidation-retains-workspace`,
    `${language}:revoked-session-returns-to-login-without-reload`];
}

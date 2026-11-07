import assert from "node:assert/strict";
import { createHmac } from "node:crypto";
import { mkdir } from "node:fs/promises";
import { chromium } from "playwright";

let text = "";
for await (const part of process.stdin) text += part;
const input = JSON.parse(text);
const origin = new URL(input.base);
assert.equal(origin.hostname, "127.0.0.1");
assert.equal(origin.protocol, "http:");
await mkdir(input.output, { recursive: true });

function totp(secret, step) {
  let bits = 0, value = 0;
  const bytes = [];
  for (const character of secret) {
    const digit = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567".indexOf(character);
    assert.ok(digit >= 0);
    value = (value << 5) | digit;
    bits += 5;
    if (bits >= 8) { bits -= 8; bytes.push((value >>> bits) & 255); }
  }
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(step));
  const digest = createHmac("sha1", Buffer.from(bytes)).update(counter).digest();
  return ((digest.readUInt32BE(digest[19] & 15) & 0x7fffffff) % 1000000).toString().padStart(6, "0");
}

async function capture(page, prefix, focusedLabel) {
  for (const theme of ["light", "dark"]) {
    await page.evaluate((theme) => { document.documentElement.dataset.theme = theme; }, theme);
    for (const width of [2580, 1440, 820]) {
      await page.setViewportSize({ width, height: 1100 });
      const field = page.getByLabel(focusedLabel, { exact: true });
      await field.focus();
      assert.ok(await field.evaluate((element) => element === document.activeElement));
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), "page overflow");
      await page.screenshot({ path: `${input.output}/${prefix}-${theme}-${width}.png`,
        mask: [page.locator('input[type="password"]'), page.locator('[autocomplete="one-time-code"]')] });
    }
  }
}

assert.equal(totp("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", 1), "287082");

const browser = await chromium.launch({ headless: true });
const checks = [], errors = [];
let finalState;
try {
  async function context(language, storageState) {
    const actor = await browser.newContext({ storageState });
    await actor.route("**/*", (route) => {
      const target = new URL(route.request().url());
      return target.origin === origin.origin || ["data:", "blob:"].includes(target.protocol)
        ? route.continue() : route.abort();
    });
    await actor.addInitScript((language) => {
      localStorage.setItem("railkeeper.settings.language", language);
      localStorage.setItem("railkeeper.theme", "light");
    }, language);
    const page = await actor.newPage();
    page.setDefaultTimeout(10000);
    page.on("pageerror", (error) => errors.push(error.message));
    return { actor, page };
  }

  if (input.mode === "existing") {
    const { actor, page } = await context("de", input.sessionState);
    await page.goto(input.base + "/vehicles");
    await page.getByRole("heading", { name: "Bestand", exact: true }).waitFor();
    const response = await page.request.get(input.base + "/api/v1/auth/session");
    assert.equal(response.status(), 200);
    assert.equal((await response.json()).username, input.credentials.username);
    checks.push("existing-session-and-enabled-two-factor-survive-process-restart");
    await actor.close();
  } else {
    for (const language of ["de", "en"]) {
      const { actor, page } = await context(language);
      await page.goto(input.base);
      await page.getByRole("heading", { name: language === "de" ? "Ersteinrichtung" : "First Setup" }).waitFor();
      await capture(page, `setup-${language}`, language === "de" ? "Einrichtungscode" : "Setup code");
      checks.push(`${language}-setup-label-focus-themes-and-responsive-widths`);
      await actor.close();
    }
    const { actor, page } = await context("de");
    await page.goto(input.base);
    await page.getByLabel("Einrichtungscode", { exact: true }).fill(input.setupCode);
    await page.getByLabel("Benutzername", { exact: true }).fill(input.credentials.username);
    await page.getByLabel("E-Mail-Adresse", { exact: true }).fill("setup-owner@example.test");
    await page.getByLabel("Passwort", { exact: true }).fill(input.credentials.password);
    await page.getByLabel("Passwort wiederholen", { exact: true }).fill(input.credentials.password);
    const created = page.waitForResponse((response) => response.url().endsWith("/api/v1/setup/admin"));
    await page.getByRole("button", { name: "Admin erstellen", exact: true }).click();
    assert.equal((await created).status(), 201);
    await page.getByRole("button", { name: "Anmelden", exact: true }).waitFor();
    const repeated = await page.request.post(input.base + "/api/v1/setup/admin", {
      data: { ...input.credentials, email: "setup-owner@example.test", setupCode: input.setupCode }
    });
    assert.equal(repeated.status(), 409);
    checks.push("browser-first-admin-consumes-local-code-once");
    await page.getByLabel("Benutzername", { exact: true }).fill(input.credentials.username);
    await page.getByLabel("Passwort", { exact: true }).fill(input.credentials.password);
    await page.getByRole("button", { name: "Anmelden", exact: true }).click();
    await page.waitForURL("**/overview");
    const session = await (await page.request.get(input.base + "/api/v1/auth/session")).json();
    const headers = { "X-CSRF-Token": session.csrfToken };
    const preparation = await page.request.post(input.base + "/api/v1/auth/two-factor/setup", { headers });
    assert.equal(preparation.status(), 200);
    const { secret } = await preparation.json();
    const activationStep = Math.floor(Date.now() / 30000);
    const activationCode = totp(secret, activationStep);
    const enabled = await page.request.post(input.base + "/api/v1/auth/two-factor/enable", {
      headers, data: { code: activationCode }
    });
    assert.equal(enabled.status(), 200);
    await actor.close();

    for (const language of ["de", "en"]) {
      const de = language === "de";
      const { actor, page } = await context(language);
      await page.goto(input.base);
      await page.getByLabel(de ? "Benutzername" : "Username", { exact: true }).fill(input.credentials.username);
      await page.getByLabel(de ? "Passwort" : "Password", { exact: true }).fill(input.credentials.password);
      await page.getByRole("button", { name: de ? "Anmelden" : "Sign in", exact: true }).click();
      const codeLabel = de ? "Zwei-Faktor-Code" : "Two-factor code";
      await page.getByLabel(codeLabel, { exact: true }).waitFor();
      await capture(page, `two-factor-${language}`, codeLabel);
      if (de) {
        await page.getByLabel(codeLabel, { exact: true }).fill(activationCode);
        const rejected = page.waitForResponse((response) => response.url().endsWith("/api/v1/auth/login"));
        await page.getByRole("button", { name: "Anmelden", exact: true }).click();
        assert.equal((await rejected).status(), 401);
        assert.equal(await page.getByLabel("Benutzername", { exact: true }).inputValue(), input.credentials.username);
        checks.push("used-activation-code-rejected-with-inputs-retained");
      } else {
        const freshStep = Math.max(activationStep + 1, Math.floor(Date.now() / 30000));
        await page.getByLabel(codeLabel, { exact: true }).fill(totp(secret, freshStep));
        await page.getByRole("button", { name: "Sign in", exact: true }).click();
        await page.waitForURL("**/overview");
        finalState = await actor.storageState();
        checks.push("fresh-totp-sign-in-succeeds-with-existing-password");
      }
      checks.push(`${language}-two-factor-label-hint-focus-themes-and-responsive-widths`);
      await actor.close();
    }
  }
  assert.deepEqual(errors, []);
  // The restart orchestrator keeps this synthetic session in memory, never in reports.
  process.stdout.write(JSON.stringify({ checks, sessionState: finalState }));
} finally {
  await browser.close();
}

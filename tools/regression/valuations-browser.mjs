import { blockExternalRequests, closeBrowser, closeContext } from "./browser-network.mjs";
import assert from "node:assert/strict";
import { mkdir, writeFile } from "node:fs/promises";
import { chromium } from "playwright";
import { referenceActor } from "./browser-actor.mjs";
import { assertValuations } from "./valuations.mjs";
import { assertMobileActionTargets } from "./touch-targets.mjs";

let raw = "";
for await (const part of process.stdin) raw += part;
const input = JSON.parse(raw), origin = new URL(input.base);
assert.equal(origin.hostname, "127.0.0.1");
assert.equal(origin.protocol, "http:");
await mkdir(input.output, { recursive: true });
const browser = await chromium.launch({ headless: true }), reference = await referenceActor(input), checks = [];
try {
  for (const language of ["de", "en"]) {
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
    await blockExternalRequests(context, origin);
    await context.addInitScript(language => localStorage.setItem("railkeeper.settings.language", language), language);
    const page = await context.newPage(), errors = [], de = language === "de";
    page.setDefaultTimeout(10000);
    page.on("pageerror", error => errors.push(error.message));
    await page.goto(input.base + "/vehicles");
    const forgot = page.getByRole("button", { name: de ? "Passwort vergessen?" : "Forgot password?", exact: true });
    await forgot.waitFor();
    for (const width of [320, 390, 820]) {
      await page.setViewportSize({ width, height: 1000 });
      await assertMobileActionTargets(page);
      const geometry = await page.evaluate(() => ({ content: document.documentElement.scrollWidth, viewport: innerWidth }));
      assert.ok(geometry.content <= geometry.viewport + 1, `Login content exceeds viewport: ${JSON.stringify(geometry)}`);
    }
    for (let tab = 0; tab < 12 && !await forgot.evaluate(node => node === document.activeElement); tab++) {
      await page.keyboard.press("Tab");
    }
    assert.equal(await forgot.evaluate(node => node === document.activeElement), true);
    assert.equal(await forgot.evaluate(node => getComputedStyle(node).outlineStyle), "solid");
    checks.push(`${language}:mobile-login-actions-and-keyboard-focus`);
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.getByLabel(de ? "Benutzername" : "Username", { exact: true }).fill(input.credentials.username);
    await page.getByLabel(de ? "Passwort" : "Password", { exact: true }).fill(input.credentials.password);
    await page.getByRole("button", { name: de ? "Anmelden" : "Sign in", exact: true }).click();
    await page.waitForURL("**/overview");
    await page.setViewportSize({ width: 320, height: 1000 });
    await page.goto(input.base + "/settings", { waitUntil: "networkidle" });
    await page.evaluate(() => document.fonts.ready);
    await assertMobileActionTargets(page);
    const settingsFits = await page.locator(".main").evaluate(node => node.scrollWidth <= node.clientWidth + 1);
    assert.equal(settingsFits, true, "Mobile settings must wrap their title and action groups at 320 pixels");
    checks.push(`${language}:mobile-settings-action-layout`);
    checks.push(...await assertValuations(page, input, de, reference));
    assert.deepEqual(errors, []);
    await closeContext(context);
  }
  await writeFile(input.output + "/report.json", JSON.stringify({ passed: true, checks }, null, 2) + "\n");
  process.stdout.write(JSON.stringify({ checks }) + "\n");
} catch (failure) {
  for (const [index, context] of browser.contexts().entries()) {
    for (const [pageIndex, page] of context.pages().entries()) {
      await page.screenshot({ path: `${input.output}/failure-${index}-${pageIndex}.png` }).catch(() => undefined);
    }
  }
  const secrets = [input.credentials.password, reference.csrfToken];
  const actorState = await reference.actor.storageState();
  secrets.push(...actorState.cookies.map(cookie => cookie.value));
  for (const context of browser.contexts()) secrets.push(...(await context.cookies()).map(cookie => cookie.value));
  let message = String(failure.stack || failure);
  for (const value of secrets.filter(Boolean)) message = message.replaceAll(value, "[redacted]");
  message = message.replace(/(cookie|x-csrf-token|set-cookie):[^\n]*/gi, "$1: [redacted]");
  process.stderr.write(message + "\n");
  process.exitCode = 1;
} finally {
  await reference.actor.dispose();
  await closeBrowser(browser);
}

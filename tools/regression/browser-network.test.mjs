import assert from "node:assert/strict";
import { createServer } from "node:http";
import { once } from "node:events";
import test from "node:test";
import { chromium } from "playwright";
import { blockExternalRequests, closeBrowser, closeContext } from "./browser-network.mjs";

test("browser request ownership and teardown retain the external network boundary", { timeout: 60000 }, async () => {
  let localWrites = 0, foreignRequests = 0;
  const local = createServer((request, response) => {
    if (request.method === "PUT") localWrites++;
    response.setHeader("Content-Type", request.url === "/" ? "text/html" : "application/json");
    response.end(request.url === "/" ? "<!doctype html><title>Local fixture</title>" : JSON.stringify({ saved: true }));
  });
  const foreign = createServer((_request, response) => { foreignRequests++; response.end("must not be requested"); });
  local.listen(0, "127.0.0.1"); foreign.listen(0, "127.0.0.1");
  await Promise.all([once(local, "listening"), once(foreign, "listening")]);
  const origin = new URL(`http://127.0.0.1:${local.address().port}`);
  const foreignURL = `http://127.0.0.1:${foreign.address().port}/foreign`;
  const browser = await chromium.launch({ headless: true });
  try {
    const context = await browser.newContext();
    await blockExternalRequests(context, origin);
    const page = await context.newPage();
    await page.goto(origin.href);
    const endpoint = new URL("/probe", origin).href;
    for (let round = 0; round < 20; round++) {
      const rejectWrite = route => route.request().method() === "PUT"
        ? route.fulfill({ status: 503, body: "synthetic rejection" }) : route.continue();
      await page.route(endpoint, rejectWrite);
      const results = await page.evaluate(async () => Promise.all([
        ...Array.from({ length: 4 }, () => fetch("/probe").then(response => response.status)),
        fetch("/probe", { method: "PUT" }).then(response => response.status)
      ]));
      assert.deepEqual(results, [200, 200, 200, 200, 503]);
      assert.equal(localWrites, round, "rejected writes reached the local server");
      await page.unroute(endpoint, rejectWrite);
      assert.equal(await page.evaluate(() => fetch("/probe", { method: "PUT" }).then(response => response.status)), 200);
    }
    assert.equal(await page.evaluate(url => fetch(url).then(() => "allowed", () => "blocked"), foreignURL), "blocked");
    assert.equal(foreignRequests, 0, "network protection contacted another origin");
    assert.equal(await page.evaluate(() => fetch("data:text/plain,local-data").then(response => response.text())), "local-data");
    assert.equal(await page.evaluate(async () => {
      const url = URL.createObjectURL(new Blob(["local-blob"]));
      try { return await (await fetch(url)).text(); } finally { URL.revokeObjectURL(url); }
    }), "local-blob");
    await closeContext(context);

    const pendingContext = await browser.newContext();
    await blockExternalRequests(pendingContext, origin);
    const pendingPage = await pendingContext.newPage();
    await pendingPage.goto(origin.href);
    let release, started;
    const held = new Promise(resolve => { release = resolve; });
    const active = new Promise(resolve => { started = resolve; });
    let handlerFinished = false, contextClosed = false;
    await pendingPage.route(new URL("/pending", origin).href, async route => {
      started(); await held; await route.continue(); handlerFinished = true;
    });
    const request = pendingPage.evaluate(() => fetch("/pending").then(response => response.status))
      .then(status => assert.equal(status, 200), error => assert.match(error.message, /closed/));
    await active;
    const closing = closeContext(pendingContext).then(() => { contextClosed = true; });
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(contextClosed, false, "teardown closed a context with an active route");
    release(); await closing; await request;
    assert.equal(handlerFinished, true);
  } finally {
    await closeBrowser(browser);
    local.closeAllConnections(); foreign.closeAllConnections();
    await Promise.all([new Promise(resolve => local.close(resolve)), new Promise(resolve => foreign.close(resolve))]);
  }
});

// Leave same-origin requests to the real server or a test's specific page route.
// A second context handler must not continue a request that the page already handled.
export async function blockExternalRequests(context, origin) {
  await context.route(url => url.origin !== origin.origin && !["data:", "blob:"].includes(url.protocol),
    route => route.abort());
}

export async function closeContext(context) {
  for (const page of context.pages()) await page.unrouteAll({ behavior: "wait" });
  await context.unrouteAll({ behavior: "wait" });
  await context.close();
}

export async function closeBrowser(browser) {
  for (const context of browser.contexts()) await closeContext(context);
  await browser.close();
}

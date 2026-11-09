import assert from "node:assert/strict";

export async function assertMobileActionTargets(page) {
  if ((await page.evaluate(() => innerWidth)) > 900) return;
  const undersized = await page.locator(
    "button:not(.visually-hidden-action):not(.visually-hidden):not(.sr-only), " +
    "a:is(.primary-button, .secondary-button, .icon-button, .table-icon-link, .auth-reset-link, .overview-open-action), " +
    ".overview-priority-list a, .overview-vehicle-list a, .overview-maintenance-list a, " +
    ".transfer-artifact-links a, .sidebar-nav a, .sidebar-footer-actions a"
  ).evaluateAll(nodes => nodes.filter(node => node.checkVisibility({
    checkOpacity: true, checkVisibilityCSS: true
  })).flatMap(node => {
    const bounds = node.getBoundingClientRect();
    if (!bounds.width || !bounds.height || (bounds.width >= 43.9 && bounds.height >= 43.9)) return [];
    return [{ label: node.getAttribute("aria-label") || node.textContent.trim().slice(0, 80),
      className: node.className, width: bounds.width, height: bounds.height }];
  }));
  assert.deepEqual(undersized, [], "Mobile actions must retain a 44-pixel target in both dimensions");
}

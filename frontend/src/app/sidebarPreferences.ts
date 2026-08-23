import type { AppView } from "./App";

export const sidebarOrderChangedEvent = "railkeeper-sidebar-order-changed";
export const legacySidebarOrderKey = "railkeeper.settings.sidebarOrder";
export const sidebarPrefsBaseKey = "railkeeper.settings.sidebarPrefs";
export const defaultSidebarOrder: AppView[] = [
  "overview",
  "vehicles",
  "accessories",
  "exhibition",
  "importExport",
  "digitalCenters",
  "settings"
];

export type SidebarPrefs = {
  order: AppView[];
  hidden: AppView[];
};

export function sidebarPrefsKey(username: string) {
  return `${sidebarPrefsBaseKey}:${username || "local"}`;
}

export function normalizeSidebarOrder(order: AppView[]) {
  const normalized = order.filter(
    (view, index): view is AppView => defaultSidebarOrder.includes(view) && order.indexOf(view) === index
  );

  defaultSidebarOrder.forEach((view, defaultIndex) => {
    if (normalized.includes(view)) return;
    const successor = defaultSidebarOrder
      .slice(defaultIndex + 1)
      .find((candidate) => normalized.includes(candidate));
    const insertionIndex = successor ? normalized.indexOf(successor) : normalized.length;
    normalized.splice(insertionIndex, 0, view);
  });

  return normalized;
}

function normalizeSidebarHidden(hidden: AppView[]) {
  return hidden.filter(
    (view, index): view is AppView =>
      defaultSidebarOrder.includes(view) && view !== "settings" && hidden.indexOf(view) === index
  );
}

export function normalizeSidebarPrefs(prefs: Partial<SidebarPrefs>): SidebarPrefs {
  return {
    order: normalizeSidebarOrder(Array.isArray(prefs.order) ? prefs.order : []),
    hidden: normalizeSidebarHidden(Array.isArray(prefs.hidden) ? prefs.hidden : [])
  };
}

export function readSidebarPrefs(username: string): SidebarPrefs {
  try {
    const stored = JSON.parse(
      window.localStorage.getItem(sidebarPrefsKey(username)) || "null"
    ) as Partial<SidebarPrefs> | null;
    if (stored) {
      return normalizeSidebarPrefs(stored);
    }
  } catch {
    return { order: defaultSidebarOrder, hidden: [] };
  }

  try {
    const legacyOrder = JSON.parse(
      window.localStorage.getItem(legacySidebarOrderKey) || "[]"
    ) as AppView[];
    return { order: normalizeSidebarOrder(legacyOrder), hidden: [] };
  } catch {
    return { order: defaultSidebarOrder, hidden: [] };
  }
}

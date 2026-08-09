import type { AppView } from "./App";

const appViews = [
  "overview",
  "vehicles",
  "accessories",
  "layouts",
  "exhibition",
  "importExport",
  "settings"
] as const;

const temporarilyDisabledViews = new Set<AppView>();

function isAppView(value: string): value is AppView {
  return appViews.some((view) => view === value);
}

export function isViewTemporarilyDisabled(view: AppView): boolean {
  return temporarilyDisabledViews.has(view);
}

export function availableStartView(value: string): AppView {
  if (value === "inventory") return "vehicles";
  if (!isAppView(value) || isViewTemporarilyDisabled(value)) return "overview";
  return value;
}
